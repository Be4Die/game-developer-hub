// Package app координирует инициализацию всех компонентов сервиса project-manager.
package app

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"

	"google.golang.org/grpc"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
	"github.com/Be4Die/game-developer-hub/project-manager/internal/infrastructure/client"
	"github.com/Be4Die/game-developer-hub/project-manager/internal/infrastructure/config"
	"github.com/Be4Die/game-developer-hub/project-manager/internal/infrastructure/valkey"
	"github.com/Be4Die/game-developer-hub/project-manager/internal/service"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/storage/deployment"
	"github.com/Be4Die/game-developer-hub/project-manager/internal/storage/filesystem"
	"github.com/Be4Die/game-developer-hub/project-manager/internal/storage/postgres"
	s3storage "github.com/Be4Die/game-developer-hub/project-manager/internal/storage/s3"
	grpctransport "github.com/Be4Die/game-developer-hub/project-manager/internal/transport/grpc"
	pb "github.com/Be4Die/game-developer-hub/protos/project_manager/v1"
	"github.com/jackc/pgx/v5/pgxpool"
)

// App координирует все компоненты сервиса.
type App struct {
	log        *slog.Logger
	cfg        *config.Config
	gRPCServer *grpc.Server
	pool       *pgxpool.Pool
	once       sync.Once
}

// New создаёт и инициализирует все компоненты сервиса.
func New(log *slog.Logger, cfg *config.Config) (*App, error) {
	// ─── PostgreSQL ─────────────────────────────────────────────
	pool, err := pgxpool.New(context.Background(), cfg.DB.DSN())
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	log.Info("connected to postgres", slog.String("host", cfg.DB.Host))

	// ─── Valkey Locker ──────────────────────────────────────────
	var locker domain.Locker
	if cfg.Valkey.Addr != "" {
		valkeyLocker, err := valkey.NewLocker(cfg.Valkey.Addr, cfg.Valkey.Password, cfg.Valkey.DB)
		if err != nil {
			log.Warn("failed to connect to valkey, falling back to no-op locker", slog.String("error", err.Error()))
			locker = valkey.NewNoOpLocker()
		} else {
			locker = valkeyLocker
			log.Info("connected to valkey for distributed locks", slog.String("addr", cfg.Valkey.Addr))
		}
	} else {
		locker = valkey.NewNoOpLocker()
	}

	// ─── Репозитории ────────────────────────────────────────────
	projectRepo := postgres.NewProjectRepo(pool)
	draftRepo := postgres.NewDraftRepo(pool)
	buildRepo := postgres.NewBuildRepo(pool)
	releaseRepo := postgres.NewReleaseRepo(pool)
	deploymentRepo := postgres.NewDeploymentRepo(pool)
	memberRepo := postgres.NewMemberRepo(pool)
	invitationRepo := postgres.NewInvitationRepo(pool)
	blockRepo := postgres.NewBlockRepo(pool)

	// ─── Хранилища и Драйверы развертывания ─────────────────────
	var (
		buildStorage domain.BuildStorage
		mediaStorage domain.MediaStorage
		deployer     domain.Deployer
		s3Client     *s3storage.Client
	)

	if cfg.Storage.Driver == "s3" || cfg.Deployment.Mode == "s3" {
		client, err := s3storage.NewClient(context.Background(), s3storage.Config{
			Endpoint:     cfg.Storage.S3.Endpoint,
			AccessKey:    cfg.Storage.S3.AccessKey,
			SecretKey:    cfg.Storage.S3.SecretKey,
			UseSSL:       cfg.Storage.S3.UseSSL,
			Region:       cfg.Storage.S3.Region,
			GamesBucket:  cfg.Storage.S3.GamesBucket,
			MediaBucket:  cfg.Storage.S3.MediaBucket,
			BuildsBucket: cfg.Storage.S3.BuildsBucket,
		}, log)
		if err != nil {
			return nil, fmt.Errorf("failed to create S3 client: %w", err)
		}
		s3Client = client

		if err := s3Client.EnsureBuckets(context.Background(),
			cfg.Storage.S3.GamesBucket,
			cfg.Storage.S3.MediaBucket,
			cfg.Storage.S3.BuildsBucket,
		); err != nil {
			log.Warn("failed to ensure S3 buckets (storage might still be starting)", slog.String("error", err.Error()))
		}
	}

	// Выбор хранилища билдов и промо-медиа
	if cfg.Storage.Driver == "s3" && s3Client != nil {
		buildStorage = s3storage.NewBuildStorage(s3Client, cfg.Storage.S3.BuildsBucket)
		mediaStorage = s3storage.NewMediaStorage(s3Client, cfg.Storage.S3.MediaBucket)
		log.Info("using S3 build and media storage", slog.String("endpoint", cfg.Storage.S3.Endpoint))
	} else {
		buildStorage = filesystem.NewBuildStorage(cfg.Storage.ProjectsPath)
		mediaStorage = filesystem.NewMediaStorage(cfg.Storage.ProjectsPath)
		log.Info("using filesystem build and media storage", slog.String("path", cfg.Storage.ProjectsPath))
	}

	// Выбор режима публикации игр (local, agent, s3)
	switch cfg.Deployment.Mode {
	case "s3":
		if s3Client == nil {
			return nil, fmt.Errorf("deployment.mode is 's3' but S3 client was not initialized")
		}
		deployer = s3storage.NewDeployer(s3Client, cfg.Storage.S3.GamesBucket, cfg.Storage.S3.BuildsBucket, cfg.Deployment.URLPrefix)
		log.Info("using S3 deployer for web games", slog.String("endpoint", cfg.Storage.S3.Endpoint))
	case "agent":
		agentDeployer, err := deployment.NewAgentDeployer(
			cfg.Deployment.AgentEndpoint,
			cfg.Deployment.AgentAPIKey,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create agent deployer: %w", err)
		}
		if s3Storage, ok := buildStorage.(interface {
			GetObject(ctx context.Context, projectID int64, version string) (io.ReadCloser, error)
		}); ok {
			agentDeployer.SetArchiveOpener(func(ctx context.Context, projectID int64, version, _ string) (io.ReadCloser, error) {
				return s3Storage.GetObject(ctx, projectID, version)
			})
		}
		deployer = agentDeployer
		log.Info("using remote agent deployer", slog.String("endpoint", cfg.Deployment.AgentEndpoint))
	case "local":
		fallthrough
	default:
		deployer = deployment.NewLocalDeployer(
			cfg.Deployment.GamesBasePath,
			cfg.Deployment.URLPrefix,
		)
		log.Info("using local deployer", slog.String("games_path", cfg.Deployment.GamesBasePath))
	}

	// ─── Клиент сервиса модерации (gRPC или Stub) ────────────────
	var moderationClient domain.ModerationClient
	if cfg.Moderation.Addr != "" {
		grpcModClient, err := client.NewGRPCModerationClient(cfg.Moderation.Addr)
		if err != nil {
			log.Warn("failed to connect to moderation service, using stub", slog.String("error", err.Error()))
			moderationClient = client.NewStubModerationClient()
		} else {
			moderationClient = grpcModClient
			log.Info("connected to moderation service", slog.String("addr", cfg.Moderation.Addr))
		}
	} else {
		moderationClient = client.NewStubModerationClient()
		log.Info("using in-memory stub moderation client")
	}

	// ─── Сервисы ────────────────────────────────────────────────
	projectService := service.NewProjectService(
		projectRepo,
		draftRepo,
		buildRepo,
		releaseRepo,
		deploymentRepo,
		memberRepo,
		invitationRepo,
		blockRepo,
		moderationClient,
		buildStorage,
		mediaStorage,
		deployer,
		locker,
		cfg.Storage.MaxBuildVersions,
	)

	// ─── gRPC-транспорт ─────────────────────────────────────────
	projectHandler := grpctransport.NewProjectHandler(projectService)

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

	pb.RegisterProjectServiceServer(gRPCServer, projectHandler)

	log.Info("all components initialized")

	return &App{
		log:        log,
		cfg:        cfg,
		gRPCServer: gRPCServer,
		pool:       pool,
	}, nil
}

// MustRun запускает gRPC-сервер. Блокирует вызов.
func (a *App) MustRun() {
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

// MustStop выполняет graceful shutdown.
func (a *App) MustStop() {
	a.once.Do(func() {
		a.log.Info("shutting down gRPC server")
		a.gRPCServer.GracefulStop()
		a.pool.Close()
		a.log.Info("application stopped")
	})
}
