// Package app координирует инициализацию всех компонентов сервиса project-manager.
package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"

	"google.golang.org/grpc"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
	"github.com/Be4Die/game-developer-hub/project-manager/internal/infrastructure/config"
	"github.com/Be4Die/game-developer-hub/project-manager/internal/infrastructure/valkey"
	"github.com/Be4Die/game-developer-hub/project-manager/internal/service"
	"github.com/Be4Die/game-developer-hub/project-manager/internal/storage/deployment"
	"github.com/Be4Die/game-developer-hub/project-manager/internal/storage/filesystem"
	"github.com/Be4Die/game-developer-hub/project-manager/internal/storage/postgres"
	grpctransport "github.com/Be4Die/game-developer-hub/project-manager/internal/transport/grpc"
	pb "github.com/Be4Die/game-developer-hub/protos/project_manager/v1"
	"github.com/jackc/pgx/v5/pgxpool"
)

// App координирует все компоненты сервиса.
type App struct {
	log        *slog.Logger
	cfg        *config.Config
	gRPCServer *grpc.Server
	pool       *pgxpool.Pool
	once       sync.Once
}

// New создаёт и инициализирует все компоненты сервиса.
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

	if err := postgres.InitSchema(context.Background(), pool); err != nil {
		return nil, fmt.Errorf("init schema: %w", err)
	}

	// ─── Valkey Locker ──────────────────────────────────────────
	var locker domain.Locker
	if cfg.Valkey.Addr != "" {
		valkeyLocker, err := valkey.NewLocker(cfg.Valkey.Addr, cfg.Valkey.Password, cfg.Valkey.DB)
		if err != nil {
			log.Warn("failed to connect to valkey, falling back to no-op locker", slog.String("error", err.Error()))
			locker = valkey.NewNoOpLocker()
		} else {
			locker = valkeyLocker
			log.Info("connected to valkey for distributed locks", slog.String("addr", cfg.Valkey.Addr))
		}
	} else {
		locker = valkey.NewNoOpLocker()
	}

	// ─── Репозитории ────────────────────────────────────────────
	projectRepo := postgres.NewProjectRepo(pool)
	draftRepo := postgres.NewDraftRepo(pool)
	buildRepo := postgres.NewBuildRepo(pool)
	moderationRepo := postgres.NewModerationRepo(pool)
	releaseRepo := postgres.NewReleaseRepo(pool)
	deploymentRepo := postgres.NewDeploymentRepo(pool)

	// ─── Хранилища и Драйверы развертывания ─────────────────────
	buildStorage := filesystem.NewBuildStorage(cfg.Storage.ProjectsPath)
	mediaStorage := filesystem.NewMediaStorage(cfg.Storage.ProjectsPath)

	var deployer domain.Deployer
	if cfg.Deployment.Mode == "agent" {
		agentDeployer, err := deployment.NewAgentDeployer(
			cfg.Deployment.AgentEndpoint,
			cfg.Deployment.AgentAPIKey,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create agent deployer: %w", err)
		}
		deployer = agentDeployer
		log.Info("using remote agent deployer", slog.String("endpoint", cfg.Deployment.AgentEndpoint))
	} else {
		deployer = deployment.NewLocalDeployer(
			cfg.Deployment.GamesBasePath,
			cfg.Deployment.URLPrefix,
		)
		log.Info("using local deployer", slog.String("games_path", cfg.Deployment.GamesBasePath))
	}

	// ─── Сервисы ────────────────────────────────────────────────
	projectService := service.NewProjectService(
		projectRepo,
		draftRepo,
		buildRepo,
		moderationRepo,
		releaseRepo,
		deploymentRepo,
		buildStorage,
		mediaStorage,
		deployer,
		locker,
		cfg.Storage.MaxBuildVersions,
	)

	moderationService := service.NewModerationService(
		projectRepo,
		draftRepo,
		buildRepo,
		moderationRepo,
		releaseRepo,
		deploymentRepo,
		buildStorage,
		deployer,
	)

	// ─── gRPC-транспорт ─────────────────────────────────────────
	projectHandler := grpctransport.NewProjectHandler(projectService)
	moderationHandler := grpctransport.NewModerationHandler(moderationService, projectService)

	// ─── Аутентификация ─────────────────────────────────────────
	authInterceptor, err := grpctransport.NewJWTAuth(cfg.JWT.Secret, cfg.JWT.Issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT auth: %w", err)
	}

	// ─── Создание gRPC-сервера ──────────────────────────────────
	gRPCServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.Unary()),
		grpc.StreamInterceptor(authInterceptor.Stream()),
	)

	pb.RegisterProjectServiceServer(gRPCServer, projectHandler)
	pb.RegisterModerationServiceServer(gRPCServer, moderationHandler)

	log.Info("all components initialized")

	return &App{
		log:        log,
		cfg:        cfg,
		gRPCServer: gRPCServer,
		pool:       pool,
	}, nil
}

// MustRun запускает gRPC-сервер. Блокирует вызов.
func (a *App) MustRun() {
	addr := fmt.Sprintf(":%d", a.cfg.GRPC.Port)
	a.log.Info("gRPC server listening", slog.String("addr", addr))

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
	a.once.Do(func() {
		a.log.Info("shutting down gRPC server")
		a.gRPCServer.GracefulStop()
		a.pool.Close()
		a.log.Info("application stopped")
	})
}
