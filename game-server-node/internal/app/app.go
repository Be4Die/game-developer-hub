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

type App struct {
	log        *slog.Logger
	config     *config.Config
	gRPCServer *grpc.Server
}

// New creates and wires all dependencies.
// Returns error if any infrastructure component fails to initialize.
func New(log *slog.Logger, cfg *config.Config) (*App, error) {
	// 1. Infrastructure.
	storage := memory.NewMemoryInstanceStorage()

	runtime, err := docker.New(log)
	if err != nil {
		return nil, fmt.Errorf("app.New: init docker runtime: %w", err)
	}

	// 2. Services.
	discoverySvc := service.NewDiscoveryService(storage, runtime, cfg)
	deploymentSvc := service.NewDeploymentService(log, storage, runtime)

	// 3. Transport.
	discoveryHandler := grpctransport.NewDiscoveryHandler(discoverySvc)
	deploymentHandler := grpctransport.NewDeploymentHandler(deploymentSvc)

	// 4. gRPC server.
	gRPCServer := grpc.NewServer()
	pb.RegisterDiscoveryServiceServer(gRPCServer, discoveryHandler)
	pb.RegisterDeploymentServiceServer(gRPCServer, deploymentHandler)

	return &App{
		log:        log,
		config:     cfg,
		gRPCServer: gRPCServer,
	}, nil
}

func (a *App) MustRun() {
	if err := a.runGRPCServer(); err != nil {
		panic(err)
	}
}

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

func (a *App) MustStop() {
	a.log.Info("stopping gRPC server", slog.Int("port", a.config.GRPC.Port))
	a.gRPCServer.GracefulStop()
}
