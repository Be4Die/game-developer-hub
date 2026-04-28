// Package app координирует инициализацию всех компонентов приложения.
package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"

	"google.golang.org/grpc"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/infrastructure/client/grpcnode"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/infrastructure/config"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/service"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/storage/filesystem"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/storage/postgres"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/storage/valkey"
	grpctransport "github.com/Be4Die/game-developer-hub/orchestrator/internal/transport/grpc"
	pb "github.com/Be4Die/game-developer-hub/protos/orchestrator/v1"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// App координирует все компоненты оркестратора.
type App struct {
	log              *slog.Logger
	cfg              *config.Config
	gRPCServer       *grpc.Server
	grpcClient       *grpcnode.Client
	pool             *pgxpool.Pool
	valkey           *redis.Client
	heartbeatService *service.HeartbeatService
	hbCancel         context.CancelFunc
	once             sync.Once
}

// New создаёт и инициализирует все компоненты приложения.
func New(log *slog.Logger, cfg *config.Config) (*App, error) {
	// ─── PostgreSQL ─────────────────────────────────────────────
	pool, err := pgxpool.New(context.Background(), cfg.DB.DSN())
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(context.Background()); err != nil {
		return nil, err
	}
	log.Info("connected to postgres", slog.String("host", cfg.DB.Host))

	// ─── Valkey ─────────────────────────────────────────────────
	valkeyClient := redis.NewClient(&redis.Options{
		Addr:     cfg.KV.Addr,
		Password: cfg.KV.Password,
		DB:       cfg.KV.DB,
	})
	if err := valkeyClient.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	log.Info("connected to valkey", slog.String("addr", cfg.KV.Addr))

	// ─── gRPC-клиент к нодам ────────────────────────────────────
	nodeClient := grpcnode.New(cfg.GRPCClient)

	// ─── Хранилища (PostgreSQL) ─────────────────────────────────
	nodeRepo := postgres.NewNodeRepo(pool)
	instanceRepo := postgres.NewInstanceRepo(pool)
	buildRepo := postgres.NewBuildStorage(pool)

	// ─── Хранилища (Valkey) ─────────────────────────────────────
	nodeState := valkey.NewNodeStateStore(valkeyClient, cfg.KV.KeyTTL)
	instanceState := valkey.NewInstanceStateStore(valkeyClient, cfg.KV.KeyTTL)

	// ─── Хранилище файлов ───────────────────────────────────────
	buildFS := filesystem.NewBuildStorageFS(cfg.Storage.BuildsPath)

	// ─── Сервисы ────────────────────────────────────────────────
	buildPipeline := service.NewBuildPipeline(
		buildRepo, buildFS, nodeClient, nodeRepo, nodeState, cfg.Limits,
	)

	instanceService := service.NewInstanceService(
		instanceRepo, instanceState, buildRepo, nodeRepo, nodeState, nodeClient, cfg.Limits,
	)

	discoveryService := service.NewDiscoveryService(
		instanceRepo, instanceState, nodeRepo,
	)

	nodeService := service.NewNodeService(
		nodeRepo, nodeState, instanceRepo, instanceState, nodeClient,
	)

	heartbeatService := service.NewHeartbeatService(
		nodeRepo, nodeState, instanceRepo, instanceState, nodeClient,
		cfg.NodeHeartbeat, log,
	)

	// ─── gRPC-транспорт ─────────────────────────────────────────
	buildHandler := grpctransport.NewBuildHandler(buildPipeline)
	instanceHandler := grpctransport.NewInstanceHandler(instanceService, cfg.Limits.MaxLogTailLines)
	discoveryHandler := grpctransport.NewDiscoveryHandler(discoveryService)
	nodeHandler := grpctransport.NewNodeHandler(nodeService)
	healthHandler := grpctransport.NewHealthHandler("1.0.0")

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

	pb.RegisterBuildServiceServer(gRPCServer, buildHandler)
	pb.RegisterInstanceServiceServer(gRPCServer, instanceHandler)
	pb.RegisterDiscoveryServiceServer(gRPCServer, discoveryHandler)
	pb.RegisterNodeServiceServer(gRPCServer, nodeHandler)
	pb.RegisterHealthServiceServer(gRPCServer, healthHandler)

	log.Info("all components initialized")

	return &App{
		log:              log,
		cfg:              cfg,
		gRPCServer:       gRPCServer,
		grpcClient:       nodeClient,
		pool:             pool,
		valkey:           valkeyClient,
		heartbeatService: heartbeatService,
	}, nil
}

// MustRun запускает gRPC-сервер и фоновые процессы. Блокирует вызов.
func (a *App) MustRun() {
	// Запускаем heartbeat-сервис.
	hbCtx, hbCancel := context.WithCancel(context.Background()) //nolint:gosec // hbCancel вызывается в MustStop
	a.hbCancel = hbCancel

	go func() {
		a.log.Info("heartbeat service started")
		a.heartbeatService.Run(hbCtx)
		a.log.Info("heartbeat service stopped")
	}()

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

// MustStop выполняет graceful shutdown gRPC-сервера и фоновых процессов.
func (a *App) MustStop() {
	a.once.Do(func() {
		a.log.Info("shutting down gRPC server")

		a.gRPCServer.GracefulStop()

		if a.hbCancel != nil {
			a.hbCancel()
		}

		a.grpcClient.Close()
		_ = a.valkey.Close()
		a.pool.Close()

		a.log.Info("application stopped")
	})
}
