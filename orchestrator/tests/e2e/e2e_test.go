// Package e2e содержит end-to-end тесты для оркестратора.
//
// В отличие от интеграционных тестов (httptest с моками), e2e используют
// реальные сервисы: PostgreSQL, Valkey и game-server-node в контейнерах.
// Orchestrator запускается in-process как HTTP сервер на случайном порту.
//
// Перед запуском:
//  1. Соберите образ game-server-node: task node:build
//  2. Убедитесь что Docker Desktop запущен
//
// Запуск:
//
//	go test -tags=e2e ./tests/e2e/...
package e2e

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/config"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/service"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/storage/filesystem"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/storage/postgres"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/storage/valkey"
	orchhttp "github.com/Be4Die/game-developer-hub/orchestrator/internal/transport/http"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	testcontainerspostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	testcontainersredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

const (
	e2eTestTimeout = 60 * time.Second
	containerWait  = 500 * time.Millisecond
	nodeImageTag   = "game-server-node:latest"
	testAPIKey     = "dev-api-key-for-local-testing"
	testImageTag   = "alpine:3.18"
)

// e2eTestEnv хранит все компоненты тестового окружения.
type e2eTestEnv struct {
	pgContainer    *testcontainerspostgres.PostgresContainer
	redisContainer *testcontainersredis.RedisContainer
	pool           *pgxpool.Pool
	redisClient    *redis.Client
	nodeRepo       *postgres.NodeRepo
	instanceRepo   *postgres.InstanceRepo
	buildStorage   *postgres.BuildStorage
	buildFS        *filesystem.BuildStorageFS
	nodeState      *valkey.NodeStateStore
	instanceState  *valkey.InstanceStateStore
	httpServer     *httptest.Server
	baseURL        string
	log            *slog.Logger
}

// setupE2E поднимает PostgreSQL, Valkey и запускает orchestrator in-process.
func setupE2E(t *testing.T) *e2eTestEnv {
	t.Helper()
	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))

	// PostgreSQL.
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

	// Valkey.
	redisContainer, err := testcontainersredis.Run(ctx, "valkey/valkey:8-alpine")
	if err != nil {
		t.Skipf("Valkey container not available: %v", err)
	}

	redisConnStr, err := redisContainer.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("failed to get redis connection string: %v", err)
	}
	redisAddr := strings.TrimPrefix(redisConnStr, "redis://")

	// Подключение к PostgreSQL с retry.
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

	// Подключение к Valkey.
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		t.Fatalf("failed to ping valkey: %v", err)
	}

	// Создание таблиц.
	createE2ETables(t, pool)

	// Репозитории.
	nodeRepo := postgres.NewNodeRepo(pool)
	instanceRepo := postgres.NewInstanceRepo(pool)
	buildStorage := postgres.NewBuildStorage(pool)

	// Хранилища состояний.
	keyTTL := 45 * time.Second
	nodeState := valkey.NewNodeStateStore(redisClient, keyTTL)
	instanceState := valkey.NewInstanceStateStore(redisClient, keyTTL)

	// Файловое хранилище (временная директория).
	buildFS := filesystem.NewBuildStorageFS(t.TempDir())

	// Конфиг лимитов.
	limits := config.LimitsConfig{
		MaxBuildsPerGame:    10,
		MaxInstancesPerGame: 50,
		MaxLogTailLines:     5000,
		MaxBuildSizeBytes:   2147483648,
	}

	// TODO: для полного e2e с game-server-node контейнером нужен реальный gRPC client.
	// Пока тестируем с mock NodeClient (имитирует успешные gRPC вызовы).
	nodeClient := &e2eMockNodeClient{}

	// Сервисы.
	buildPipeline := service.NewBuildPipeline(
		buildStorage, buildFS, nodeClient, nodeRepo, nodeState, limits,
	)

	instanceService := service.NewInstanceService(
		instanceRepo, instanceState, buildStorage,
		nodeRepo, nodeState, nodeClient, limits,
	)

	discoveryService := service.NewDiscoveryService(
		instanceRepo, instanceState, nodeRepo,
	)

	nodeService := service.NewNodeService(
		nodeRepo, nodeState, instanceRepo, instanceState, nodeClient,
	)

	// Handlers.
	buildHandler := orchhttp.NewBuildHandler(buildPipeline)
	instanceHandler := orchhttp.NewInstanceHandler(instanceService, limits.MaxLogTailLines)
	discoveryHandler := orchhttp.NewDiscoveryHandler(discoveryService)
	nodeHandler := orchhttp.NewNodeHandler(nodeService)
	healthHandler := orchhttp.NewHealthHandler("e2e-test-1.0.0")

	router := orchhttp.NewRouter(
		buildHandler, instanceHandler, discoveryHandler, nodeHandler, healthHandler, log,
	)

	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)

	t.Cleanup(func() {
		redisClient.Close()
		pool.Close()
	})

	return &e2eTestEnv{
		pgContainer:    pgContainer,
		redisContainer: redisContainer,
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

// cleanupDB удаляет все данные из БД и KV.
func (env *e2eTestEnv) cleanupDB(t *testing.T) {
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

// createE2ETables создаёт таблицы в PostgreSQL.
func createE2ETables(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()

	migrations := []string{
		`CREATE TABLE IF NOT EXISTS nodes (
			id            BIGSERIAL PRIMARY KEY,
			address       TEXT NOT NULL UNIQUE,
			token_hash    BYTEA NOT NULL,
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

// e2eMockNodeClient — mock NodeClient для e2e тестов (реализует domain.NodeClient).
// Позволяет тестировать сценарии без реального game-server-node контейнера.
type e2eMockNodeClient struct {
	nextInstanceID int64
}

func (m *e2eMockNodeClient) nextID() int64 {
	m.nextInstanceID++
	return m.nextInstanceID
}

func (m *e2eMockNodeClient) LoadImage(ctx context.Context, nodeAddress string, metadata domain.ImageMetadata, chunks io.Reader) (*domain.ImageLoadResult, error) {
	return &domain.ImageLoadResult{ImageTag: metadata.ImageTag, SizeBytes: 1000}, nil
}
func (m *e2eMockNodeClient) StartInstance(ctx context.Context, nodeAddress string, req domain.StartInstanceRequest) (*domain.StartInstanceResult, error) {
	return &domain.StartInstanceResult{InstanceID: m.nextID(), HostPort: 7001}, nil
}
func (m *e2eMockNodeClient) StopInstance(ctx context.Context, nodeAddress string, instanceID int64, timeoutSec uint32) error {
	return nil
}
func (m *e2eMockNodeClient) StreamLogs(ctx context.Context, nodeAddress string, req domain.StreamLogsRequest) (domain.LogStream, error) {
	return nil, nil
}
func (m *e2eMockNodeClient) GetNodeInfo(ctx context.Context, nodeAddress string) (*domain.NodeInfo, error) {
	return &domain.NodeInfo{
		Region:           "test-region",
		CPUCores:         4,
		TotalMemoryBytes: 8000000000,
		TotalDiskBytes:   250000000000,
		AgentVersion:     "e2e-test-1.0.0",
	}, nil
}
func (m *e2eMockNodeClient) Heartbeat(ctx context.Context, nodeAddress string) (*domain.ResourceUsage, error) {
	return &domain.ResourceUsage{CPUUsagePercent: 10.0}, nil
}
func (m *e2eMockNodeClient) ListInstances(ctx context.Context, nodeAddress string) ([]*domain.Instance, error) {
	return nil, nil
}
func (m *e2eMockNodeClient) GetInstance(ctx context.Context, nodeAddress string, instanceID int64) (*domain.Instance, error) {
	return nil, nil
}
func (m *e2eMockNodeClient) GetInstanceUsage(ctx context.Context, nodeAddress string, instanceID int64) (*domain.ResourceUsage, error) {
	return &domain.ResourceUsage{CPUUsagePercent: 5.0}, nil
}

// findFreePort возвращает свободный TCP порт.
func findFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

// TestMain — точка входа для всех e2e тестов.
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
