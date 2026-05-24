package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"google.golang.org/grpc"

	pb "github.com/Be4Die/game-developer-hub/protos/chat/v1"
	"github.com/Be4Die/game-developer-hub/chat/internal/infrastructure/config"
	grpctransport "github.com/Be4Die/game-developer-hub/chat/internal/transport/grpc"
	pg "github.com/Be4Die/game-developer-hub/chat/internal/storage/postgres"
	"github.com/Be4Die/game-developer-hub/chat/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	log        *slog.Logger
	config     *config.Config
	gRPCServer *grpc.Server
	pool       *pgxpool.Pool
}

func New(log *slog.Logger, cfg *config.Config) (*App, error) {
	pool, err := pgxpool.New(context.Background(), cfg.DB.DSN())
	if err != nil {
		return nil, fmt.Errorf("app.New: create pgx pool: %w", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("app.New: ping postgres: %w", err)
	}
	log.Info("connected to postgres", slog.String("host", cfg.DB.Host))

	if err := pg.InitSchema(context.Background(), pool); err != nil {
		return nil, fmt.Errorf("app.New: init schema: %w", err)
	}

	convRepo := pg.NewConversationRepository(pool)
	msgRepo := pg.NewMessageRepository(pool)

	chatService := service.NewChatService(log, convRepo, msgRepo)
	chatHandler := grpctransport.NewChatHandler(chatService)

	gRPCServer := grpc.NewServer()
	pb.RegisterChatServiceServer(gRPCServer, chatHandler)

	return &App{
		log:        log,
		config:     cfg,
		gRPCServer: gRPCServer,
		pool:       pool,
	}, nil
}

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

func (a *App) MustStop() {
	a.log.Info("stopping gRPC server")
	a.gRPCServer.GracefulStop()
	a.pool.Close()
}
