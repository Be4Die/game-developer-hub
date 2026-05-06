// Package app координирует инициализацию всех компонентов приложения.
package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/Be4Die/game-developer-hub/game-server-node/internal/infrastructure/config"
	"github.com/Be4Die/game-developer-hub/game-server-node/internal/infrastructure/runtime/docker"
	"github.com/Be4Die/game-developer-hub/game-server-node/internal/infrastructure/sysinfo"
	"github.com/Be4Die/game-developer-hub/game-server-node/internal/service"
	"github.com/Be4Die/game-developer-hub/game-server-node/internal/storage/memory"
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
	deploymentSvc   *service.DeploymentService
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

	// Ensure data directory exists.
	dataDir := cfg.Node.DataDir
	if err := os.MkdirAll(dataDir, 0o750); err != nil {
		return nil, fmt.Errorf("app.New: create data dir: %w", err)
	}

	imageMapPath := filepath.Join(dataDir, "images.json")

	// Ensure node ID exists.
	nodeIDPath := filepath.Join(dataDir, "node_id")
	nodeID, err := ensureNodeID(nodeIDPath)
	if err != nil {
		return nil, fmt.Errorf("app.New: ensure node ID: %w", err)
	}

	// Initialize services.
	discoverySvc := service.NewDiscoveryService(storage, runtime, cfg)
	deploymentSvc := service.NewDeploymentService(log, storage, runtime, imageMapPath, nodeID)

	// Initialize transport layer.
	discoveryHandler := grpctransport.NewDiscoveryHandler(discoverySvc)
	deploymentHandler := grpctransport.NewDeploymentHandler(deploymentSvc)

	// Configure auth interceptor.
	authInterceptor := grpctransport.NewAPIKeyAuth(cfg.APIKey)

	// Configure and register gRPC server.
	gRPCServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.Unary()),
		grpc.StreamInterceptor(authInterceptor.Stream()),
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
		deploymentSvc:   deploymentSvc,
	}, nil
}

// MustRun запускает gRPC-сервер и паникует при ошибке.
// В режиме auto-discovery сначала выполняет анонсирование ноды.
func (a *App) MustRun() {
	ctx := context.Background()
	// Cleanup any orphan containers from previous runs.
	if err := a.deploymentSvc.CleanupOrphans(ctx); err != nil {
		a.log.Warn("failed to cleanup orphan containers", slog.String("error", err.Error()))
	}
	// В режиме auto-discovery анонсируем ноду перед запуском сервера.
	if a.config.Orchestrator.Mode == "auto-discovery" && a.announcementSvc != nil {
		// Collect active container IDs for sync.
		activeContainerIDs := a.deploymentSvc.GetActiveContainerIDs(ctx)
		result, err := a.announcementSvc.AnnounceWithRetry(ctx, activeContainerIDs)
		if err != nil {
			panic(fmt.Sprintf("failed to announce node: %v", err))
		}
		a.log.Info("node registered with orchestrator",
			slog.Int64("node_id", result.NodeID),
		)
		fmt.Printf("═══════════════════════════════════════════════════════\n")
		fmt.Printf("  Node announced. Authorization key: %s\n", a.config.APIKey)
		fmt.Printf("  Use this key to authorize the node in the dashboard.\n")
		fmt.Printf("═══════════════════════════════════════════════════\n")
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

// ensureNodeID читает node_id из файла или создаёт новый, если файл отсутствует.
func ensureNodeID(path string) (string, error) {
	if _, err := os.Stat(path); err == nil {
		// Файл существует — читаем.
		data, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("read node ID: %w", err)
		}
		id := strings.TrimSpace(string(data))
		if id == "" {
			return "", fmt.Errorf("node ID file is empty")
		}
		return id, nil
	} else if os.IsNotExist(err) {
		// Создаём новый ID.
		bytes := make([]byte, 16)
		if _, err := rand.Read(bytes); err != nil {
			return "", fmt.Errorf("generate random node ID: %w", err)
		}
		id := hex.EncodeToString(bytes)

		// Убедимся, что директория существует.
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			return "", fmt.Errorf("create data directory: %w", err)
		}
		if err := os.WriteFile(path, []byte(id), 0o600); err != nil {
			return "", fmt.Errorf("write node ID: %w", err)
		}
		return id, nil
	} else {
		return "", fmt.Errorf("stat node ID file: %w", err)
	}
}

// MustStop gracefully остаанавливает gRPC-сервер и все инстансы.
func (a *App) MustStop() {
	a.log.Info("stopping gRPC server", slog.Int("port", a.config.GRPC.Port))
	a.gRPCServer.GracefulStop()

	// Stop all managed instances.
	ctx := context.Background()
	if err := a.deploymentSvc.StopAllInstances(ctx); err != nil {
		a.log.Warn("failed to stop all instances", slog.String("error", err.Error()))
	}
	// Cleanup any stray containers that might remain.
	if err := a.deploymentSvc.CleanupOrphans(ctx); err != nil {
		a.log.Warn("failed to cleanup orphan containers", slog.String("error", err.Error()))
	}
}
