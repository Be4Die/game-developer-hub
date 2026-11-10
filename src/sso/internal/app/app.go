// Package app координирует инициализацию всех компонентов SSO-сервиса.
package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"google.golang.org/grpc"

	pb "github.com/Be4Die/game-developer-hub/protos/sso/v1"
	"github.com/Be4Die/game-developer-hub/sso/internal/domain"
	bcryptpkg "github.com/Be4Die/game-developer-hub/sso/internal/infrastructure/bcrypt"
	"github.com/Be4Die/game-developer-hub/sso/internal/infrastructure/config"
	jwtpkg "github.com/Be4Die/game-developer-hub/sso/internal/infrastructure/jwt"
	smtppkg "github.com/Be4Die/game-developer-hub/sso/internal/infrastructure/smtp"
	"github.com/Be4Die/game-developer-hub/sso/internal/service"
	pg "github.com/Be4Die/game-developer-hub/sso/internal/storage/postgres"
	"github.com/Be4Die/game-developer-hub/sso/internal/storage/valkey"
	grpctransport "github.com/Be4Die/game-developer-hub/sso/internal/transport/grpc"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// App — основной координатор SSO-сервиса.
type App struct {
	log        *slog.Logger
	config     *config.Config
	gRPCServer *grpc.Server
	pool       *pgxpool.Pool
	valkey     *redis.Client
}

// New создаёт SSO-сервис со всеми зависимостями.
func New(log *slog.Logger, cfg *config.Config) (*App, error) {
	// 1. Инициализация PostgreSQL.
	pool, err := pgxpool.New(context.Background(), cfg.DB.DSN())
	if err != nil {
		return nil, fmt.Errorf("app.New: create pgx pool: %w", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("app.New: ping postgres: %w", err)
	}
	log.Info("connected to postgres", slog.String("host", cfg.DB.Host))

	// 2. Инициализация Valkey.
	valkeyClient := redis.NewClient(&redis.Options{
		Addr:     cfg.KV.Addr,
		Password: cfg.KV.Password,
		DB:       cfg.KV.DB,
	})
	if err := valkeyClient.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("app.New: ping valkey: %w", err)
	}
	log.Info("connected to valkey", slog.String("addr", cfg.KV.Addr))

	// 3. Инициализация хранилищ.
	userRepo := pg.NewUserRepository(pool)
	sessionRepo := pg.NewSessionRepository(pool)
	emailStore := valkey.NewEmailVerificationStore(valkeyClient, cfg.Email.VerificationCodeTTL)
	resetStore := valkey.NewPasswordResetStore(valkeyClient, cfg.Email.ResetTokenTTL)
	sessionCache := valkey.NewSessionCache(valkeyClient, cfg.KV.KeyTTL)

	// 4. Инициализация крипто-провайдеров.
	passwordHasher, err := bcryptpkg.NewPasswordHasher(bcryptpkg.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("app.New: create password hasher: %w", err)
	}
	tokenManager, err := jwtpkg.NewTokenManager(cfg.JWT.Secret, cfg.JWT.AccessTokenTTL, cfg.JWT.Issuer)
	if err != nil {
		return nil, fmt.Errorf("app.New: create token manager: %w", err)
	}

	// 4.1. Обеспечение наличия администратора.
	if cfg.Admin.Password != "" {
		adminInit := service.NewAdminInitializer(
			log, userRepo, passwordHasher,
			cfg.Admin.Email, cfg.Admin.Password, cfg.Admin.DisplayName,
		)
		if err := adminInit.EnsureAdmin(context.Background()); err != nil {
			log.Warn("failed to ensure admin user", slog.String("error", err.Error()))
		}
	}

	// 5. Инициализация сервисов.
	var emailSender domain.EmailSender
	if cfg.Email.Driver == config.EmailDriverSMTP {
		smtpSender, err := smtppkg.NewSender(smtppkg.Config{
			Host:     cfg.SMTP.Host,
			Port:     cfg.SMTP.Port,
			Username: cfg.SMTP.Username,
			Password: cfg.SMTP.Password,
			From:     cfg.SMTP.From,
			TLS:      cfg.SMTP.TLS,
		})
		if err != nil {
			return nil, fmt.Errorf("app.New: create smtp sender: %w", err)
		}
		emailSender = smtpSender
		log.Info("using SMTP email sender", slog.String("host", cfg.SMTP.Host), slog.Int("port", cfg.SMTP.Port))
	} else {
		emailSender = service.NewStubEmailSender(log)
		log.Warn("using STUB email sender: verification codes are written to the log, do not use in production")
	}

	// Ограничение перебора: 10 неверных паролей или 5 неверных кодов за 15 минут на один адрес.
	loginLimiter := valkey.NewAttemptLimiter(valkeyClient, "login", 10, 15*time.Minute)
	verifyLimiter := valkey.NewAttemptLimiter(valkeyClient, "verify", 5, 15*time.Minute)

	authService := service.NewAuthService(
		log, userRepo, sessionRepo, tokenManager, passwordHasher,
		emailStore, resetStore, emailSender, cfg.JWT.RefreshTokenTTL,
	).WithAttemptLimiters(loginLimiter, verifyLimiter)
	userService := service.NewUserService(log, userRepo, passwordHasher, sessionRepo)
	tokenService := service.NewTokenService(log, sessionRepo, sessionCache, tokenManager)

	// 6. Инициализация обработчиков gRPC.
	authHandler := grpctransport.NewAuthHandler(authService)
	userHandler := grpctransport.NewUserHandler(userService)
	tokenHandler := grpctransport.NewTokenHandler(tokenService)

	// 7. Настройка аутентификации.
	// Chain interceptor: сначала API key (для service-to-service), затем JWT (для user endpoints).
	authInterceptor := grpctransport.NewAPIKeyAuth(cfg.APIKey)
	jwtInterceptor := grpctransport.NewJWTAuth(tokenManager)

	// 8. Создание и регистрация gRPC-сервера.
	gRPCServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(grpctransport.NewErrorSanitizer(log).Unary(), authInterceptor.Unary(), jwtInterceptor.Unary()),
	)
	pb.RegisterAuthServiceServer(gRPCServer, authHandler)
	pb.RegisterUserServiceServer(gRPCServer, userHandler)
	pb.RegisterTokenServiceServer(gRPCServer, tokenHandler)

	return &App{
		log:        log,
		config:     cfg,
		gRPCServer: gRPCServer,
		pool:       pool,
		valkey:     valkeyClient,
	}, nil
}

// MustRun запускает gRPC-сервер.
func (a *App) MustRun() {
	addr := fmt.Sprintf(":%d", a.config.GRPC.Port)
	a.log.Info("starting gRPC server", slog.String("addr", addr))

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		a.log.Error("failed to listen", slog.String("error", err.Error()))
		panic(err)
	}

	if err := a.gRPCServer.Serve(lis); err != nil {
		a.log.Error("gRPC server failed", slog.String("error", err.Error()))
		panic(err)
	}
}

// MustStop выполняет graceful shutdown.
func (a *App) MustStop() {
	a.log.Info("stopping gRPC server")
	a.gRPCServer.GracefulStop()
	_ = a.valkey.Close()
	a.pool.Close()
}
