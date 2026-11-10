//go:build integration

// Package integration содержит интеграционные тесты оркестратора.
// Тесты запускают реальные PostgreSQL и Valkey контейнеры через testcontainers.
// Запуск: go test -tags=integration ./tests/integration/...
package integration

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/storage/postgres"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/storage/valkey"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	testcontainerspostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	testcontainersredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

// integrationTestEnv хранит все компоненты тестового окружения.
type integrationTestEnv struct {
	pgContainer    *testcontainerspostgres.PostgresContainer
	redisContainer *testcontainersredis.RedisContainer
	pool           *pgxpool.Pool
	redisClient    *redis.Client
	nodeRepo       *postgres.NodeRepo
	instanceRepo   *postgres.InstanceRepo
	buildStorage   *postgres.BuildStorage
	nodeState      *valkey.NodeStateStore
	instanceState  *valkey.InstanceStateStore
	log            *slog.Logger
}

// setupIntegration запускает PostgreSQL и Valkey контейнеры, создаёт pool и репозитории.
func setupIntegration(t *testing.T) *integrationTestEnv {
	t.Helper()
	ctx := context.Background()

	// PostgreSQL.
	pgContainer, err := testcontainerspostgres.Run(ctx,
		"postgres:17-alpine",
		testcontainerspostgres.WithDatabase("orchestrator"),
		testcontainerspostgres.WithUsername("postgres"),
		testcontainerspostgres.WithPassword("postgres"),
		testcontainerspostgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Skipf("PostgreSQL container not available: %v", err)
	}

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	// Valkey (Redis-compatible).
	redisContainer, err := testcontainersredis.Run(ctx, "valkey/valkey:8-alpine")
	if err != nil {
		t.Skipf("Valkey container not available: %v", err)
	}

	redisConnStr, err := redisContainer.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("failed to get redis connection string: %v", err)
	}
	// ConnectionString returns "redis://host:port" — extract host:port
	redisAddr := strings.TrimPrefix(redisConnStr, "redis://")

	// Подключение к PostgreSQL с retry.
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}
	// PostgreSQL может быть ещё не готов принимать подключения — retry.
	for i := range 10 {
		if err = pool.Ping(ctx); err == nil {
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
	createTables(t, pool)

	// Репозитории.
	nodeRepo := postgres.NewNodeRepo(pool)
	instanceRepo := postgres.NewInstanceRepo(pool)
	buildStorage := postgres.NewBuildStorage(pool)

	// Хранилища состояний.
	keyTTL := 45 * time.Second
	nodeState := valkey.NewNodeStateStore(redisClient, keyTTL)
	instanceState := valkey.NewInstanceStateStore(redisClient, keyTTL)

	t.Cleanup(func() {
		redisClient.Close()
		pool.Close()
	})

	return &integrationTestEnv{
		pgContainer:    pgContainer,
		redisContainer: redisContainer,
		pool:           pool,
		redisClient:    redisClient,
		nodeRepo:       nodeRepo,
		instanceRepo:   instanceRepo,
		buildStorage:   buildStorage,
		nodeState:      nodeState,
		instanceState:  instanceState,
		log:            slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}
}

// createTables накатывает схему сервиса из migrations/init.sql — ту же, что применяется
// в docker-compose, чтобы тесты репозиториев проверяли реальную схему, а не её копию.
func createTables(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	schema, err := os.ReadFile(filepath.Join("..", "..", "migrations", "init.sql"))
	if err != nil {
		t.Fatalf("failed to read schema: %v", err)
	}
	if _, err := pool.Exec(ctx, string(schema)); err != nil {
		t.Fatalf("failed to apply schema: %v", err)
	}
}

// cleanupTables удаляет все данные из таблиц перед каждым тестом.
func (env *integrationTestEnv) cleanupTables(t *testing.T) {
	t.Helper()
	ctx := context.Background()

	// TRUNCATE with RESTART IDENTITY to reset sequences and respect FK order.
	stmt := "TRUNCATE instances, server_builds, nodes RESTART IDENTITY CASCADE"
	if _, err := env.pool.Exec(ctx, stmt); err != nil {
		t.Fatalf("failed to clean tables: %v", err)
	}
}

// cleanupKeys удаляет все ключи в Valkey.
func (env *integrationTestEnv) cleanupKeys(t *testing.T) {
	t.Helper()
	ctx := context.Background()

	if err := env.redisClient.FlushDB(ctx).Err(); err != nil {
		t.Fatalf("failed to flush valkey: %v", err)
	}
}

// Cleanup удаляет все данные из БД и KV (вызывается перед каждым тестом).
func (env *integrationTestEnv) Cleanup(t *testing.T) {
	t.Helper()
	env.cleanupTables(t)
	env.cleanupKeys(t)
}

// ─── Helper: открыть SQL connection (для raw SQL тестов) ───────────────────

func openSQL(t *testing.T, env *integrationTestEnv) *sql.DB {
	t.Helper()

	ctx := context.Background()
	host, err := env.pgContainer.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get pg host: %v", err)
	}
	port, err := env.pgContainer.MappedPort(ctx, "5432/tcp")
	if err != nil {
		t.Fatalf("failed to get pg port: %v", err)
	}

	// port.Port() returns string like "5432"
	portStr := port.Port()

	dsn := fmt.Sprintf("host=%s port=%s user=postgres password=postgres dbname=orchestrator sslmode=disable",
		host, portStr)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("failed to open sql: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}
