//go:build integration

package integration

import (
	"context"
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
	svc := service.NewModerationService(reqRepo, msgRepo, pmClient)

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

func createModTables(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()

	ddl := []string{
		`CREATE TABLE IF NOT EXISTS moderation_requests (
			id                    BIGSERIAL PRIMARY KEY,
			project_id            BIGINT NOT NULL,
			owner_id              TEXT NOT NULL,
			moderator_id          TEXT NOT NULL DEFAULT '',
			status                SMALLINT NOT NULL DEFAULT 1,
			snapshot_meta         JSONB NOT NULL DEFAULT '{}'::jsonb,
			rejection_reason      TEXT NOT NULL DEFAULT '',
			submitted_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			started_review_at     TIMESTAMP WITH TIME ZONE,
			resolved_at           TIMESTAMP WITH TIME ZONE,
			created_at            TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at            TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS moderation_messages (
			id            BIGSERIAL PRIMARY KEY,
			project_id    BIGINT NOT NULL,
			request_id    BIGINT REFERENCES moderation_requests(id) ON DELETE SET NULL,
			sender_id     TEXT NOT NULL,
			sender_role   SMALLINT NOT NULL DEFAULT 1,
			message_type  SMALLINT NOT NULL DEFAULT 1,
			content       TEXT NOT NULL DEFAULT '',
			payload       JSONB NOT NULL DEFAULT '{}'::jsonb,
			created_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)`,
	}

	for _, stmt := range ddl {
		if _, err := pool.Exec(ctx, stmt); err != nil {
			t.Fatalf("failed to execute DDL: %v", err)
		}
	}
}

func (e *ModTestEnv) cleanup(t *testing.T) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, _ = e.pool.Exec(ctx, "TRUNCATE TABLE moderation_messages, moderation_requests CASCADE")
}
