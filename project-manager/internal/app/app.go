// Package app координирует инициализацию всех компонентов.
package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"

	"google.golang.org/grpc"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/infrastructure/config"
	"github.com/Be4Die/game-developer-hub/project-manager/internal/service"
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

// New создаёт и инициализирует все компоненты.
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

	// ─── Хранилища ──────────────────────────────────────────────
	projectRepo := postgres.NewProjectRepo(pool)
	buildRepo := postgres.NewProjectBuildRepo(pool)
	mediaStorage := filesystem.NewProjectStorage(cfg.Storage.ProjectsPath)

	// ─── Сервисы ────────────────────────────────────────────────
	projectService := service.NewProjectService(
		projectRepo, buildRepo, mediaStorage, cfg.Storage.MaxBuildVersions,
	)

	// ─── gRPC-транспорт ─────────────────────────────────────────
	projectHandler := grpctransport.NewProjectHandler(projectService)

	// ─── Аутентификация ─────────────────────────────────────────
	authInterceptor, err := grpctransport.NewJWTAuth(cfg.JWT.Secret, cfg.JWT.Issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT auth: %w", err)
	}

	// ─── Создание gRPC-сервера ──────────────────────────────────
	const maxMsgSize = 128 * 1024 * 1024 // 128MB для загрузки билдов
	gRPCServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.Unary()),
		grpc.StreamInterceptor(authInterceptor.Stream()),
		grpc.MaxRecvMsgSize(maxMsgSize),
		grpc.MaxSendMsgSize(maxMsgSize),
	)

	pb.RegisterProjectServiceServer(gRPCServer, projectHandler)

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
