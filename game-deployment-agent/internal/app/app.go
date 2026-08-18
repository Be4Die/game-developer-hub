package app

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/Be4Die/game-developer-hub/game-deployment-agent/internal/infrastructure/config"
	"github.com/Be4Die/game-developer-hub/game-deployment-agent/internal/service"
	transGrpc "github.com/Be4Die/game-developer-hub/game-deployment-agent/internal/transport/grpc"
	pb "github.com/Be4Die/game-developer-hub/protos/game_deployment_agent/v1"
	"google.golang.org/grpc"
)

// App представляет запущенный микросервис game-deployment-agent.
type App struct {
	cfg        *config.Config
	grpcServer *grpc.Server
}

// New инициализирует сервисы и gRPC сервер агента развертывания.
func New(cfg *config.Config) (*App, error) {
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

	return &App{
		cfg:        cfg,
		grpcServer: grpcServer,
	}, nil
}

// Run запускает gRPC сервер агента и слушает системные сигналы для graceful shutdown.
func (a *App) Run(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", a.cfg.Server.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}

	errChan := make(chan error, 1)
	go func() {
		log.Printf("[game-deployment-agent] gRPC server listening on %s (base path: %s)", addr, a.cfg.Deployment.GamesBasePath)
		if err := a.grpcServer.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			errChan <- fmt.Errorf("grpc serve: %w", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		log.Println("[game-deployment-agent] context cancelled, shutting down...")
	case sig := <-sigChan:
		log.Printf("[game-deployment-agent] received signal %v, shutting down...", sig)
	case err := <-errChan:
		return err
	}

	a.grpcServer.GracefulStop()
	log.Println("[game-deployment-agent] stopped gracefully")
	return nil
}
