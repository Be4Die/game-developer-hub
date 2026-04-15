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
//   - Orchestrator: in-process через httptest.NewServer с реальным gRPC-клиентом к ноде
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
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/client/grpcnode"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/config"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/service"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/storage/filesystem"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/storage/postgres"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/storage/valkey"
	orchhttp "github.com/Be4Die/game-developer-hub/orchestrator/internal/transport/http"
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
	e2eAPIKey      = "dev-api-key-for-local-testing"
	e2eTestTimeout = 60 * time.Second
)

// e2eTestEnv хранит все компоненты тестового окружения.
type e2eTestEnv struct {
	pgContainer    *testcontainerspostgres.PostgresContainer
	redisContainer *testcontainersredis.RedisContainer
	nodeContainer  testcontainers.Container
	pool           *pgxpool.Pool
	redisClient    *redis.Client
	nodeRepo       *postgres.NodeRepo
	instanceRepo   *postgres.InstanceRepo
	buildStorage   *postgres.BuildStorage
	buildFS        domain.BuildStorageFS
	nodeState      *valkey.NodeStateStore
	instanceState  *valkey.InstanceStateStore
	httpServer     *httptest.Server
	baseURL        string
	log            *slog.Logger
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
			"NODE_API_KEY": e2eAPIKey,
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
	nodeClient := grpcnode.New(grpcCfg)
	t.Cleanup(func() { nodeClient.Close() })

	// ─── Сервисы ────────────────────────────────────────────────
	limits := config.LimitsConfig{
		MaxBuildsPerGame:    10,
		MaxInstancesPerGame: 5,
		MaxLogTailLines:     5000,
		MaxBuildSizeBytes:   2 * 1024 * 1024 * 1024,
	}

	buildPipeline := service.NewBuildPipeline(
		buildStorage, buildFS, nodeClient, nodeRepo, nodeState, limits,
	)
	instanceService := service.NewInstanceService(
		instanceRepo, instanceState, buildStorage, nodeRepo, nodeState, nodeClient, limits,
	)
	discoveryService := service.NewDiscoveryService(
		instanceRepo, instanceState, nodeRepo,
	)
	nodeService := service.NewNodeService(
		nodeRepo, nodeState, instanceRepo, instanceState, nodeClient,
	)

	// ─── HTTP-сервер ────────────────────────────────────────────
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	buildHandler := orchhttp.NewBuildHandler(buildPipeline)
	instanceHandler := orchhttp.NewInstanceHandler(instanceService, limits.MaxLogTailLines)
	discoveryHandler := orchhttp.NewDiscoveryHandler(discoveryService)
	nodeHandler := orchhttp.NewNodeHandler(nodeService)
	healthHandler := orchhttp.NewHealthHandler("e2e-1.0.0")

	router := orchhttp.NewRouter(
		buildHandler, instanceHandler, discoveryHandler, nodeHandler, healthHandler, log,
	)

	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)

	t.Cleanup(func() {
		redisClient.Close()
		pool.Close()
	})

	// Регистрируем ноду через API для E2E-тестов.
	nodeAddr := nodeAddress
	token := e2eAPIKey

	regResp, regErr := http.Post(srv.URL+"/nodes", "application/json",
		strings.NewReader(fmt.Sprintf(`{"address":"%s","token":"%s","region":"e2e"}`, nodeAddr, token)),
	)
	if regErr == nil {
		defer regResp.Body.Close()
		t.Logf("Node registered via API, status=%d", regResp.StatusCode)
	}

	return &e2eTestEnv{
		pgContainer:    pgContainer,
		redisContainer: redisContainer,
		nodeContainer:  nodeContainer,
		pool:           pool,
		redisClient:    redisClient,
		nodeRepo:       nodeRepo,
		instanceRepo:   instanceRepo,
		buildStorage:   buildStorage,
		buildFS:        buildFS,
		nodeState:      nodeState,
		instanceState:  instanceState,
		httpServer:     srv,
		baseURL:        srv.URL,
		log:            log,
	}
}

// createE2ETables создаёт таблицы в PostgreSQL.
func createE2ETables(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()

	migrations := []string{
		`CREATE TABLE IF NOT EXISTS nodes (
			id            BIGSERIAL PRIMARY KEY,
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

// TestMain — точка входа для всех e2e тестов.
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
