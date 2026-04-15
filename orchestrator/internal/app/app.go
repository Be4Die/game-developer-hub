// Package app координирует инициализацию всех компонентов приложения.
package app

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/client/grpcnode"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/config"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/service"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/storage/filesystem"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/storage/postgres"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/storage/valkey"
	orchhttp "github.com/Be4Die/game-developer-hub/orchestrator/internal/transport/http"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// App координирует все компоненты оркестратора.
type App struct {
	log              *slog.Logger
	cfg              *config.Config
	srv              *http.Server
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

	// ─── HTTP-транспорт ─────────────────────────────────────────
	buildHandler := orchhttp.NewBuildHandler(buildPipeline)
	instanceHandler := orchhttp.NewInstanceHandler(instanceService, cfg.Limits.MaxLogTailLines)
	discoveryHandler := orchhttp.NewDiscoveryHandler(discoveryService)
	nodeHandler := orchhttp.NewNodeHandler(nodeService)
	healthHandler := orchhttp.NewHealthHandler("1.0.0")

	router := orchhttp.NewRouter(
		buildHandler, instanceHandler, discoveryHandler, nodeHandler, healthHandler, log,
	)

	srv := &http.Server{
		Addr:         cfg.HTTP.Addr(),
		Handler:      router,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	log.Info("all components initialized")

	return &App{
		log:              log,
		cfg:              cfg,
		srv:              srv,
		grpcClient:       nodeClient,
		pool:             pool,
		valkey:           valkeyClient,
		heartbeatService: heartbeatService,
	}, nil
}

// MustRun запускает HTTP-сервер и фоновые процессы. Блокирует вызов.
func (a *App) MustRun() {
	// Запускаем heartbeat-сервис.
	hbCtx, hbCancel := context.WithCancel(context.Background()) //nolint:gosec
	a.hbCancel = hbCancel

	go func() {
		a.log.Info("heartbeat service started")
		a.heartbeatService.Run(hbCtx)
		a.log.Info("heartbeat service stopped")
	}()

	a.log.Info("http server listening", slog.String("addr", a.srv.Addr))
	if err := a.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		a.log.Error("http server failed", slog.String("error", err.Error()))
	}
}

// MustStop выполняет graceful shutdown HTTP-сервера и фоновых процессов.
func (a *App) MustStop() {
	a.once.Do(func() {
		a.log.Info("shutting down http server")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := a.srv.Shutdown(ctx); err != nil {
			a.log.Error("http server shutdown error", slog.String("error", err.Error()))
		}

		if a.hbCancel != nil {
			a.hbCancel()
		}

		a.grpcClient.Close()
		_ = a.valkey.Close()
		a.pool.Close()

		a.log.Info("application stopped")
	})
}
