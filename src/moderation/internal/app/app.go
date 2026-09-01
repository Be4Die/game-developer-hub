// Package app координирует инициализацию и запуск сервиса moderation.
package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"

	"github.com/Be4Die/game-developer-hub/moderation/internal/domain"
	"github.com/Be4Die/game-developer-hub/moderation/internal/infrastructure/client"
	"github.com/Be4Die/game-developer-hub/moderation/internal/infrastructure/config"
	"github.com/Be4Die/game-developer-hub/moderation/internal/service"
	"github.com/Be4Die/game-developer-hub/moderation/internal/storage/postgres"
	grpctransport "github.com/Be4Die/game-developer-hub/moderation/internal/transport/grpc"
	pb "github.com/Be4Die/game-developer-hub/protos/moderation/v1"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

// App координирует все компоненты сервиса модерации.
type App struct {
	log        *slog.Logger
	cfg        *config.Config
	gRPCServer *grpc.Server
	pool       *pgxpool.Pool
	pmClient   *client.ProjectGRPCClient
	once       sync.Once
}

// New создаёт и инициализирует все зависимости сервиса.
func New(log *slog.Logger, cfg *config.Config) (*App, error) {
	// ─── PostgreSQL ─────────────────────────────────────────────
	pool, err := pgxpool.New(context.Background(), cfg.DB.DSN())
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	log.Info("connected to postgres", slog.String("host", cfg.DB.Host))

	// ─── Репозитории ────────────────────────────────────────────
	requestRepo := postgres.NewRequestRepo(pool)
	messageRepo := postgres.NewMessageRepo(pool)

	// ─── Клиент Project Manager ─────────────────────────────────
	var projectClient domain.ProjectClient
	var pmGRPCClient *client.ProjectGRPCClient

	if cfg.ProjectManager.Addr != "" {
		grpcClient, err := client.NewProjectGRPCClient(cfg.ProjectManager.Addr)
		if err != nil {
			log.Warn("failed to connect to project-manager, using stub client", slog.String("error", err.Error()))
			projectClient = client.NewProjectStubClient()
		} else {
			projectClient = grpcClient
			pmGRPCClient = grpcClient
			log.Info("connected to project-manager", slog.String("addr", cfg.ProjectManager.Addr))
		}
	} else {
		projectClient = client.NewProjectStubClient()
		log.Info("using in-memory stub project client")
	}

	// ─── Сервис ─────────────────────────────────────────────────
	moderationService := service.NewModerationService(
		requestRepo,
		messageRepo,
		projectClient,
	)

	// ─── gRPC-транспорт ─────────────────────────────────────────
	moderationHandler := grpctransport.NewModerationHandler(moderationService)

	// ─── JWT Interceptor ────────────────────────────────────────
	authInterceptor, err := grpctransport.NewJWTAuth(cfg.JWT.Secret, cfg.JWT.Issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT auth: %w", err)
	}

	gRPCServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.Unary()),
		grpc.StreamInterceptor(authInterceptor.Stream()),
	)

	pb.RegisterModerationServiceServer(gRPCServer, moderationHandler)

	log.Info("all moderation components initialized")

	return &App{
		log:        log,
		cfg:        cfg,
		gRPCServer: gRPCServer,
		pool:       pool,
		pmClient:   pmGRPCClient,
	}, nil
}

// MustRun запускает gRPC-сервер. Блокирует вызов.
func (a *App) MustRun() {
	addr := fmt.Sprintf(":%d", a.cfg.GRPC.Port)
	a.log.Info("moderation gRPC server listening", slog.String("addr", addr))

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

// MustStop выполняет graceful shutdown сервиса.
func (a *App) MustStop() {
	a.once.Do(func() {
		a.log.Info("shutting down moderation service")
		a.gRPCServer.GracefulStop()
		if a.pmClient != nil {
			_ = a.pmClient.Close()
		}
		a.pool.Close()
		a.log.Info("moderation service stopped")
	})
}
