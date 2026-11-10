//go:build integration

package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/moderation/internal/service"
	pg "github.com/Be4Die/game-developer-hub/moderation/internal/storage/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	testcontainerspostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

type ModTestEnv struct {
	pgContainer *testcontainerspostgres.PostgresContainer
	pool        *pgxpool.Pool
	requestRepo *pg.RequestRepo
	messageRepo *pg.MessageRepo
	projectCl   *inMemoryProjectClient
	svc         *service.ModerationService
}

func setupModIntegration(t *testing.T) *ModTestEnv {
	t.Helper()
	ctx := context.Background()

	// ─── PostgreSQL ─────────────────────────────────────────────
	pgContainer, err := testcontainerspostgres.Run(ctx,
		"postgres:17-alpine",
		testcontainerspostgres.WithDatabase("moderation_test"),
		testcontainerspostgres.WithUsername("postgres"),
		testcontainerspostgres.WithPassword("postgres"),
		testcontainerspostgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = pgContainer.Terminate(ctx)
		t.Fatalf("failed to get connection string: %v", err)
	}

	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		_ = pgContainer.Terminate(ctx)
		t.Fatalf("failed to parse pgx config: %v", err)
	}
	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		_ = pgContainer.Terminate(ctx)
		t.Fatalf("failed to connect to postgres: %v", err)
	}

	createModTables(t, ctx, pool)

	reqRepo := pg.NewRequestRepo(pool)
	msgRepo := pg.NewMessageRepo(pool)
	pmClient := newInMemoryProjectClient()
	svc := service.NewModerationService(reqRepo, msgRepo, pg.NewAttachmentRepo(pool), pmClient, pg.NewSnapshotRepo(pool))

	env := &ModTestEnv{
		pgContainer: pgContainer,
		pool:        pool,
		requestRepo: reqRepo,
		messageRepo: msgRepo,
		projectCl:   pmClient,
		svc:         svc,
	}

	t.Cleanup(func() {
		pool.Close()
		_ = pgContainer.Terminate(context.Background())
	})

	return env
}

// createModTables накатывает схему сервиса из migrations/init.sql — ту же, что применяется
// в docker-compose, чтобы тесты репозиториев проверяли реальную схему, а не её копию.
func createModTables(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	schema, err := os.ReadFile(filepath.Join("..", "..", "migrations", "init.sql"))
	if err != nil {
		t.Fatalf("failed to read schema: %v", err)
	}
	if _, err := pool.Exec(ctx, string(schema)); err != nil {
		t.Fatalf("failed to apply schema: %v", err)
	}
}

func (e *ModTestEnv) cleanup(t *testing.T) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, _ = e.pool.Exec(ctx, "TRUNCATE TABLE moderation_messages, moderation_requests CASCADE")
}
