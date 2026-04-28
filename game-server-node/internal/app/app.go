// Package app координирует инициализацию всех компонентов приложения.
package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"github.com/Be4Die/game-developer-hub/game-server-node/internal/infrastructure/config"
	"github.com/Be4Die/game-developer-hub/game-server-node/internal/infrastructure/runtime/docker"
	"github.com/Be4Die/game-developer-hub/game-server-node/internal/service"
	"github.com/Be4Die/game-developer-hub/game-server-node/internal/storage/memory"
	"github.com/Be4Die/game-developer-hub/game-server-node/internal/infrastructure/sysinfo"
	grpctransport "github.com/Be4Die/game-developer-hub/game-server-node/internal/transport/grpc"
	pb "github.com/Be4Die/game-developer-hub/protos/game_server_node/v1"
	"google.golang.org/grpc"
)

// App координирует запуск gRPC-сервера со всеми зависимостями.
type App struct {
	log             *slog.Logger
	config          *config.Config
	gRPCServer      *grpc.Server
	announcementSvc *service.AnnouncementService
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

	// Configure auth interceptor.
	authInterceptor := grpctransport.NewAPIKeyAuth(cfg.APIKey)

	// Configure and register gRPC server.
	gRPCServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.Unary()),
	)
	pb.RegisterDiscoveryServiceServer(gRPCServer, discoveryHandler)
	pb.RegisterDeploymentServiceServer(gRPCServer, deploymentHandler)

	// Initialize announcement service if in auto-discovery mode.
	var announcementSvc *service.AnnouncementService
	if cfg.Orchestrator.Mode == "auto-discovery" {
		sysProvider := sysinfo.NewProvider(cfg.Node.EthName)
		announcementSvc = service.NewAnnouncementService(log, cfg, sysProvider)
	}

	return &App{
		log:             log,
		config:          cfg,
		gRPCServer:      gRPCServer,
		announcementSvc: announcementSvc,
	}, nil
}

// MustRun запускает gRPC-сервер и паникует при ошибке.
// В режиме auto-discovery сначала выполняет анонсирование ноды.
func (a *App) MustRun() {
	// В режиме auto-discovery анонсируем ноду перед запуском сервера.
	if a.config.Orchestrator.Mode == "auto-discovery" && a.announcementSvc != nil {
		ctx := context.Background()
		result, err := a.announcementSvc.AnnounceWithRetry(ctx)
		if err != nil {
			panic(fmt.Sprintf("failed to announce node: %v", err))
		}
		a.log.Info("node registered with orchestrator",
			slog.Int64("node_id", result.NodeID),
		)
		fmt.Printf("═══════════════════════════════════════════════════════\n")
		fmt.Printf("  Node announced. Authorization key: %s\n", a.config.APIKey)
		fmt.Printf("  Use this key to authorize the node in the dashboard.\n")
		fmt.Printf("═══════════════════════════════════════════════════════\n")
	}

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
