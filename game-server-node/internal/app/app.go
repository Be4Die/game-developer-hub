// Package app координирует инициализацию всех компонентов приложения.
package app

import (
	"fmt"
	"log/slog"
	"net"

	"github.com/Be4Die/game-developer-hub/game-server-node/internal/config"
	"github.com/Be4Die/game-developer-hub/game-server-node/internal/runtime/docker"
	"github.com/Be4Die/game-developer-hub/game-server-node/internal/service"
	"github.com/Be4Die/game-developer-hub/game-server-node/internal/storage/memory"
	grpctransport "github.com/Be4Die/game-developer-hub/game-server-node/internal/transport/grpc"
	pb "github.com/Be4Die/game-developer-hub/protos/game_server_node/v1"
	"google.golang.org/grpc"
)

// App координирует запуск gRPC-сервера со всеми зависимостями.
type App struct {
	log        *slog.Logger
	config     *config.Config
	gRPCServer *grpc.Server
}

// New создаёт приложение со всеми инициализированными компонентами.
// Возвращает ошибку если Docker-демон недоступен.
func New(log *slog.Logger, cfg *config.Config) (*App, error) {
	// Initialize infrastructure.
	storage := memory.NewStorage()

	runtime, err := docker.New(log)
	if err != nil {
		return nil, fmt.Errorf("app.New: init docker runtime: %w", err)
	}

	// Initialize services.
	discoverySvc := service.NewDiscoveryService(storage, runtime, cfg)
	deploymentSvc := service.NewDeploymentService(log, storage, runtime)

	// Initialize transport layer.
	discoveryHandler := grpctransport.NewDiscoveryHandler(discoverySvc)
	deploymentHandler := grpctransport.NewDeploymentHandler(deploymentSvc)

	// Configure and register gRPC server.
	gRPCServer := grpc.NewServer()
	pb.RegisterDiscoveryServiceServer(gRPCServer, discoveryHandler)
	pb.RegisterDeploymentServiceServer(gRPCServer, deploymentHandler)

	return &App{
		log:        log,
		config:     cfg,
		gRPCServer: gRPCServer,
	}, nil
}

// MustRun запускает gRPC-сервер и паникует при ошибке.
func (a *App) MustRun() {
	if err := a.runGRPCServer(); err != nil {
		panic(err)
	}
}

// runGRPCServer слушает TCP-порт и обслуживает gRPC-запросы.
func (a *App) runGRPCServer() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.config.GRPC.Port))
	if err != nil {
		return fmt.Errorf("app.runGRPCServer: %w", err)
	}

	a.log.Info("grpc server started", slog.String("addr", l.Addr().String()))

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("app.runGRPCServer: %w", err)
	}

	return nil
}

// MustStop gracefully останавливает gRPC-сервер.
func (a *App) MustStop() {
	a.log.Info("stopping gRPC server", slog.Int("port", a.config.GRPC.Port))
	a.gRPCServer.GracefulStop()
}
