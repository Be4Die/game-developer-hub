//go:build e2e

// Package e2e содержит end-to-end тесты для оркестратора.
//
// В отличие от интеграционных тестов (mock NodeClient), e2e используют
// реальный game-server-node контейнер с настоящим gRPC и Docker-сокетом.
//
// Архитектура:
//   - PostgreSQL: testcontainers (реальный)
//   - Valkey: testcontainers (реальный)
//   - game-server-node: testcontainers с docker.sock mount (реальный)
//   - Orchestrator: in-process gRPC server
//
// Перед запуском:
//  1. Соберите образ game-server-node: task node:build
//  2. Убедитесь что Docker запущен
//
// Запуск:
//
//	go test -tags=e2e ./tests/e2e/...
package e2e

import (
	"context"
	"log/slog"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/infrastructure/client/grpcnode"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/infrastructure/config"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/service"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/storage/filesystem"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/storage/postgres"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/storage/valkey"
	grpctransport "github.com/Be4Die/game-developer-hub/orchestrator/internal/transport/grpc"
	pb "github.com/Be4Die/game-developer-hub/protos/orchestrator/v1"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/moby/moby/api/types/container"
	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	testcontainerspostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	testcontainersredis "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	nodeImageTag   = "game-server-node:latest"
	e2eJWTSecret   = "dev-jwt-secret-for-local-testing"
	e2eIssuer      = "sso"
	e2eTestUserID  = "test-user-001"
	e2eTestTimeout = 60 * time.Second
)

// e2eTestEnv хранит все компоненты тестового окружения.
type e2eTestEnv struct {
	pgContainer     *testcontainerspostgres.PostgresContainer
	redisContainer  *testcontainersredis.RedisContainer
	nodeContainer   testcontainers.Container
	pool            *pgxpool.Pool
	redisClient     *redis.Client
	nodeRepo        *postgres.NodeRepo
	instanceRepo    *postgres.InstanceRepo
	buildStorage    *postgres.BuildStorage
	buildFS         domain.BuildStorageFS
	nodeState       *valkey.NodeStateStore
	instanceState   *valkey.InstanceStateStore
	grpcServer      *grpc.Server
	grpcConn        *grpc.ClientConn
	buildClient     pb.BuildServiceClient
	instanceClient  pb.InstanceServiceClient
	discoveryClient pb.DiscoveryServiceClient
	nodeClient      pb.NodeServiceClient
	healthClient    pb.HealthServiceClient
	log             *slog.Logger
}

// setupE2E запускает PostgreSQL, Valkey, game-server-node и orchestrator in-process.
func setupE2E(t *testing.T) *e2eTestEnv {
	t.Helper()
	ctx := context.Background()

	// ─── PostgreSQL ─────────────────────────────────────────────
	pgContainer, err := testcontainerspostgres.Run(ctx,
		"postgres:17-alpine",
		testcontainerspostgres.WithDatabase("orchestrator"),
		testcontainerspostgres.WithUsername("postgres"),
		testcontainerspostgres.WithPassword("postgres"),
	)
	if err != nil {
		t.Skipf("PostgreSQL container not available: %v", err)
	}

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	// ─── Valkey ─────────────────────────────────────────────────
	redisContainer, err := testcontainersredis.Run(ctx, "valkey/valkey:8-alpine")
	if err != nil {
		t.Skipf("Valkey container not available: %v", err)
	}

	redisConnStr, err := redisContainer.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("failed to get redis connection string: %v", err)
	}
	redisAddr := strings.TrimPrefix(redisConnStr, "redis://")

	// ─── game-server-node ───────────────────────────────────────
	nodeReq := testcontainers.ContainerRequest{
		Image:        nodeImageTag,
		ExposedPorts: []string{"44044/tcp"},
		Env: map[string]string{
			"CONFIG_PATH":  "/app/config/local.yaml",
			"NODE_API_KEY": "test-node-token",
		},
		WaitingFor: wait.ForListeningPort("44044/tcp").WithStartupTimeout(15 * time.Second),
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.Binds = append(hc.Binds, "/var/run/docker.sock:/var/run/docker.sock")
		},
	}

	nodeContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: nodeReq,
		Started:          true,
	})
	if err != nil {
		t.Skipf("game-server-node container not available (проверьте Docker и образ game-server-node:latest): %v", err)
	}

	t.Cleanup(func() { _ = nodeContainer.Terminate(context.Background()) })

	nodePort, err := nodeContainer.MappedPort(ctx, "44044")
	if err != nil {
		t.Fatalf("failed to get node port: %v", err)
	}
	nodeHost, err := nodeContainer.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get node host: %v", err)
	}
	nodeAddress := nodeHost + ":" + nodePort.Port()

	t.Logf("game-server-node started at %s", nodeAddress)

	// ─── Подключение к PostgreSQL с retry ───────────────────────
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}
	for i := range 10 {
		if err := pool.Ping(ctx); err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
		if i == 9 {
			t.Fatalf("failed to ping postgres after 10 retries: %v", err)
		}
	}

	// ─── Подключение к Valkey ───────────────────────────────────
	redisClient := redis.NewClient(&redis.Options{Addr: redisAddr})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		t.Fatalf("failed to ping valkey: %v", err)
	}

	// ─── Создание таблиц ────────────────────────────────────────
	createE2ETables(t, pool)

	// ─── Репозитории и хранилища ────────────────────────────────
	nodeRepo := postgres.NewNodeRepo(pool)
	instanceRepo := postgres.NewInstanceRepo(pool)
	buildStorage := postgres.NewBuildStorage(pool)
	keyTTL := 45 * time.Second
	nodeState := valkey.NewNodeStateStore(redisClient, keyTTL)
	instanceState := valkey.NewInstanceStateStore(redisClient, keyTTL)
	buildFS := filesystem.NewBuildStorageFS(t.TempDir())

	// ─── gRPC-клиент к реальной ноде ────────────────────────────
	grpcCfg := config.GRPCClientConfig{
		Timeout:           30 * time.Second,
		ConnectTimeout:    10 * time.Second,
		KeepAliveTime:     30 * time.Second,
		KeepAliveTimeout:  10 * time.Second,
		MaxMessageSize:    16 * 1024 * 1024,
		EnableCompression: true,
	}
	orchNodeClient := grpcnode.New(grpcCfg)
	t.Cleanup(func() { orchNodeClient.Close() })

	// ─── Сервисы ────────────────────────────────────────────────
	limits := config.LimitsConfig{
		MaxBuildsPerGame:    10,
		MaxInstancesPerGame: 5,
		MaxLogTailLines:     5000,
		MaxBuildSizeBytes:   2 * 1024 * 1024 * 1024,
	}

	buildPipeline := service.NewBuildPipeline(
		buildStorage, buildFS, orchNodeClient, nodeRepo, nodeState, limits,
	)
	instanceService := service.NewInstanceService(
		instanceRepo, instanceState, buildStorage, nodeRepo, nodeState, orchNodeClient, limits,
	)
	discoveryService := service.NewDiscoveryService(
		instanceRepo, instanceState, nodeRepo,
	)
	nodeService := service.NewNodeService(
		nodeRepo, nodeState, instanceRepo, instanceState, orchNodeClient,
	)

	// ─── gRPC-сервер ────────────────────────────────────────────
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	buildHandler := grpctransport.NewBuildHandler(buildPipeline)
	instanceHandler := grpctransport.NewInstanceHandler(instanceService, limits.MaxLogTailLines)
	discoveryHandler := grpctransport.NewDiscoveryHandler(discoveryService)
	nodeHandler := grpctransport.NewNodeHandler(nodeService)
	healthHandler := grpctransport.NewHealthHandler("e2e-1.0.0")

	authInterceptor, err := grpctransport.NewJWTAuth(e2eJWTSecret, e2eIssuer)
	if err != nil {
		t.Fatalf("failed to create auth interceptor: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.Unary()),
		grpc.StreamInterceptor(authInterceptor.Stream()),
	)

	pb.RegisterBuildServiceServer(grpcServer, buildHandler)
	pb.RegisterInstanceServiceServer(grpcServer, instanceHandler)
	pb.RegisterDiscoveryServiceServer(grpcServer, discoveryHandler)
	pb.RegisterNodeServiceServer(grpcServer, nodeHandler)
	pb.RegisterHealthServiceServer(grpcServer, healthHandler)

	// Запуск на случайном порту.
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Error("gRPC server failed", slog.String("error", err.Error()))
		}
	}()

	t.Cleanup(func() {
		grpcServer.GracefulStop()
		redisClient.Close()
		pool.Close()
	})

	// ─── gRPC-клиент к серверу ──────────────────────────────────
	conn, err := grpc.NewClient(lis.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("failed to connect to gRPC server: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	// Регистрируем ноду через gRPC API.
	nsClient := pb.NewNodeServiceClient(conn)
	_, err = nsClient.Register(withJWT(ctx, e2eJWTSecret, e2eIssuer), &pb.NodeServiceRegisterRequest{
		Mode: &pb.NodeServiceRegisterRequest_Manual{
			Manual: &pb.RegisterNodeManual{
				Address: nodeAddress,
				Token:   "test-node-token",
				Region:  ptrStr("e2e"),
			},
		},
	})
	if err != nil {
		t.Logf("Node registration via gRPC failed: %v", err)
	} else {
		t.Logf("Node registered via gRPC")
	}

	return &e2eTestEnv{
		pgContainer:     pgContainer,
		redisContainer:  redisContainer,
		nodeContainer:   nodeContainer,
		pool:            pool,
		redisClient:     redisClient,
		nodeRepo:        nodeRepo,
		instanceRepo:    instanceRepo,
		buildStorage:    buildStorage,
		buildFS:         buildFS,
		nodeState:       nodeState,
		instanceState:   instanceState,
		grpcServer:      grpcServer,
		grpcConn:        conn,
		buildClient:     pb.NewBuildServiceClient(conn),
		instanceClient:  pb.NewInstanceServiceClient(conn),
		discoveryClient: pb.NewDiscoveryServiceClient(conn),
		nodeClient:      nsClient,
		healthClient:    pb.NewHealthServiceClient(conn),
		log:             log,
	}
}

// createE2ETables создаёт таблицы в PostgreSQL.
func createE2ETables(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()

	migrations := []string{
		`CREATE TABLE IF NOT EXISTS nodes (
			id            BIGSERIAL PRIMARY KEY,
			owner_id      TEXT NOT NULL DEFAULT '',
			address       TEXT NOT NULL UNIQUE,
			token_hash    BYTEA NOT NULL,
			api_token     TEXT NOT NULL DEFAULT '',
			region        TEXT,
			status        SMALLINT NOT NULL DEFAULT 1,
			cpu_cores     INTEGER NOT NULL DEFAULT 0,
			total_memory  BIGINT NOT NULL DEFAULT 0,
			total_disk    BIGINT NOT NULL DEFAULT 0,
			agent_version TEXT NOT NULL DEFAULT '',
			last_ping_at  TIMESTAMP NOT NULL DEFAULT NOW(),
			created_at    TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at    TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS server_builds (
			id             BIGSERIAL PRIMARY KEY,
			owner_id       TEXT NOT NULL DEFAULT '',
			game_id        BIGINT NOT NULL,
			uploaded_by    BIGINT NOT NULL DEFAULT 0,
			version        TEXT NOT NULL,
			image_tag      TEXT NOT NULL,
			protocol       SMALLINT NOT NULL,
			internal_port  INTEGER NOT NULL,
			max_players    INTEGER NOT NULL,
			file_url       TEXT NOT NULL,
			file_size      BIGINT NOT NULL,
			created_at     TIMESTAMP NOT NULL DEFAULT NOW(),
			UNIQUE (game_id, version)
		)`,
		`CREATE TABLE IF NOT EXISTS instances (
			id               BIGSERIAL PRIMARY KEY,
			owner_id         TEXT NOT NULL DEFAULT '',
			node_id          BIGINT NOT NULL REFERENCES nodes(id),
			server_build_id  BIGINT NOT NULL REFERENCES server_builds(id),
			game_id          BIGINT NOT NULL,
			name             TEXT NOT NULL,
			build_version    TEXT NOT NULL,
			protocol         SMALLINT NOT NULL,
			host_port        INTEGER NOT NULL,
			internal_port    INTEGER NOT NULL,
			status           SMALLINT NOT NULL DEFAULT 1,
			max_players      INTEGER NOT NULL,
			developer_payload JSONB,
			server_address   TEXT NOT NULL,
			started_at       TIMESTAMP NOT NULL,
			created_at       TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at       TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
	}

	for _, migration := range migrations {
		if _, err := pool.Exec(ctx, migration); err != nil {
			t.Fatalf("failed to execute migration: %v\nSQL: %s", err, migration)
		}
	}
}

// cleanupTables удаляет все данные из таблиц.
func (env *e2eTestEnv) cleanupTables(t *testing.T) {
	t.Helper()
	ctx := context.Background()

	stmt := "TRUNCATE instances, server_builds, nodes RESTART IDENTITY CASCADE"
	if _, err := env.pool.Exec(ctx, stmt); err != nil {
		t.Fatalf("failed to clean tables: %v", err)
	}

	if err := env.redisClient.FlushDB(ctx).Err(); err != nil {
		t.Fatalf("failed to flush valkey: %v", err)
	}
}

// ─── Helpers ────────────────────────────────────────────────────

func ptrStr(s string) *string {
	return &s
}

// generateTestJWT создаёт JWT-токен для тестов.
func generateTestJWT(secret, issuer, userID string) string {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   userID,
		"sid":   "test-session",
		"email": "test@example.com",
		"role":  1,
		"iss":   issuer,
		"iat":   now.Unix(),
		"exp":   now.Add(time.Hour).Unix(),
	})
	s, _ := token.SignedString([]byte(secret))
	return s
}

// withJWT adds JWT to outgoing metadata and returns the context.
func withJWT(ctx context.Context, secret, issuer string) context.Context {
	tokenStr := generateTestJWT(secret, issuer, e2eTestUserID)
	md := metadata.New(map[string]string{"authorization": "Bearer " + tokenStr})
	return metadata.NewOutgoingContext(ctx, md)
}
