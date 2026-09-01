// Package app координирует инициализацию всех компонентов сервиса game-deployment-agent.
package app

import (
	"fmt"
	"log/slog"
	"net"
	"sync"

	"github.com/Be4Die/game-developer-hub/game-deployment-agent/internal/infrastructure/config"
	"github.com/Be4Die/game-developer-hub/game-deployment-agent/internal/service"
	transGrpc "github.com/Be4Die/game-developer-hub/game-deployment-agent/internal/transport/grpc"
	pb "github.com/Be4Die/game-developer-hub/protos/game_deployment_agent/v1"
	"google.golang.org/grpc"
)

// App представляет запущенный микросервис game-deployment-agent.
type App struct {
	log        *slog.Logger
	cfg        *config.Config
	grpcServer *grpc.Server
	once       sync.Once
}

// New инициализирует сервисы и gRPC сервер агента развертывания.
func New(log *slog.Logger, cfg *config.Config) (*App, error) {
	deploySvc := service.NewDeploymentService(
		cfg.Deployment.GamesBasePath,
		cfg.Deployment.URLPrefix,
		cfg.Deployment.MaxUnpackedSizeMB,
		cfg.Deployment.MaxFilesCount,
	)

	unaryAuth, streamAuth := transGrpc.APIKeyInterceptor(cfg.Server.APIKey)

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(unaryAuth),
		grpc.ChainStreamInterceptor(streamAuth),
	)

	handler := transGrpc.NewDeploymentHandler(deploySvc, "1.0.0")
	pb.RegisterWebGameDeploymentServiceServer(grpcServer, handler)

	log.Info("all components initialized")

	return &App{
		log:        log,
		cfg:        cfg,
		grpcServer: grpcServer,
	}, nil
}

// MustRun запускает gRPC сервер агента. Блокирует вызов.
func (a *App) MustRun() {
	addr := fmt.Sprintf(":%d", a.cfg.Server.Port)
	a.log.Info("gRPC server listening",
		slog.String("addr", addr),
		slog.String("games_base_path", a.cfg.Deployment.GamesBasePath),
	)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		a.log.Error("failed to listen", slog.String("error", err.Error()))
		panic(err)
	}

	if err := a.grpcServer.Serve(lis); err != nil && err != grpc.ErrServerStopped {
		a.log.Error("gRPC server failed", slog.String("error", err.Error()))
		panic(err)
	}
}

// MustStop выполняет graceful shutdown.
func (a *App) MustStop() {
	a.once.Do(func() {
		a.log.Info("shutting down gRPC server")
		a.grpcServer.GracefulStop()
		a.log.Info("application stopped")
	})
}
