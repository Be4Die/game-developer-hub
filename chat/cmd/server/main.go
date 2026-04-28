package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"github.com/Be4Die/game-developer-hub/chat/internal/app"
	"github.com/Be4Die/game-developer-hub/chat/internal/config"
	"github.com/Be4Die/game-developer-hub/chat/internal/storage/postgres"
	grpcTransport "github.com/Be4Die/game-developer-hub/chat/internal/transport/grpc"
	pb "github.com/Be4Die/game-developer-hub/protos/chat/v1"
	ssopb "github.com/Be4Die/game-developer-hub/protos/sso/v1"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := postgres.New(ctx, cfg.Database.DSN())
	if err != nil {
		return err
	}
	defer db.Close()

	if err := postgres.RunMigrations(ctx, db); err != nil {
		return err
	}

	chatRepo := postgres.NewChatRepository(db)
	messageRepo := postgres.NewMessageRepository(db)

	chatService := app.NewChatService(chatRepo, messageRepo)

	// Подключаемся к SSO для валидации токенов
	ssoConn, err := grpc.NewClient(cfg.SSO.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer ssoConn.Close()
	ssoClient := ssopb.NewTokenServiceClient(ssoConn)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpcTransport.AuthInterceptor(ssoClient)),
	)
	pb.RegisterChatServiceServer(grpcServer, grpcTransport.NewChatServer(chatService))
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", cfg.GRPC.Addr)
	if err != nil {
		return err
	}

	go func() {
		log.Printf("gRPC server listening on %s", lis.Addr())
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("gRPC server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down gRPC server...")
	grpcServer.GracefulStop()
	return nil
}
