//go:build integration

package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
	"github.com/Be4Die/game-developer-hub/project-manager/internal/infrastructure/valkey"
	"github.com/Be4Die/game-developer-hub/project-manager/internal/service"
	pg "github.com/Be4Die/game-developer-hub/project-manager/internal/storage/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	testcontainerspostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	testcontainersredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

type TestEnv struct {
	pgContainer    *testcontainerspostgres.PostgresContainer
	redisContainer *testcontainersredis.RedisContainer
	pool           *pgxpool.Pool
	locker         *valkey.Locker
	projectRepo    *pg.ProjectRepo
	draftRepo      *pg.DraftRepo
	buildRepo      *pg.BuildRepo
	releaseRepo    *pg.ReleaseRepo
	deploymentRepo *pg.DeploymentRepo
	projectSvc     *service.ProjectService
	buildStorage   domain.BuildStorage
	mediaStorage   domain.MediaStorage
	deployer       domain.Deployer
}

func setupIntegration(t *testing.T) *TestEnv {
	t.Helper()
	ctx := context.Background()

	// ─── PostgreSQL ─────────────────────────────────────────────
	pgContainer, err := testcontainerspostgres.Run(ctx,
		"postgres:17-alpine",
		testcontainerspostgres.WithDatabase("project_manager_test"),
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

	createTables(t, ctx, pool)

	// ─── Valkey ─────────────────────────────────────────────────
	redisContainer, err := testcontainersredis.Run(ctx, "valkey/valkey:8-alpine")
	if err != nil {
		pool.Close()
		_ = pgContainer.Terminate(ctx)
		t.Fatalf("failed to start valkey container: %v", err)
	}

	redisURI, err := redisContainer.ConnectionString(ctx)
	if err != nil {
		pool.Close()
		_ = redisContainer.Terminate(ctx)
		_ = pgContainer.Terminate(ctx)
		t.Fatalf("failed to get valkey URI: %v", err)
	}

	redisAddr := redisURI
	if len(redisURI) > 8 && redisURI[:8] == "redis://" {
		redisAddr = redisURI[8:]
	}

	locker, err := valkey.NewLocker(redisAddr, "", 0)
	if err != nil {
		pool.Close()
		_ = redisContainer.Terminate(ctx)
		_ = pgContainer.Terminate(ctx)
		t.Fatalf("failed to init valkey locker: %v", err)
	}

	// Repositories
	pRepo := pg.NewProjectRepo(pool)
	dRepo := pg.NewDraftRepo(pool)
	bRepo := pg.NewBuildRepo(pool)
	rRepo := pg.NewReleaseRepo(pool)
	depRepo := pg.NewDeploymentRepo(pool)
	memRepo := pg.NewMemberRepo(pool)
	invRepo := pg.NewInvitationRepo(pool)
	blockRepo := pg.NewBlockRepo(pool)

	bStorage := newMockBuildStorage()
	mStorage := newMockMediaStorage()
	deployer := newMockDeployer()
	mClient := newMockModerationClient()

	projSvc := service.NewProjectService(
		pRepo, dRepo, bRepo, rRepo, depRepo, memRepo, invRepo, blockRepo, mClient,
		bStorage, mStorage, deployer, locker, 5,
	)

	env := &TestEnv{
		pgContainer:    pgContainer,
		redisContainer: redisContainer,
		pool:           pool,
		locker:         locker,
		projectRepo:    pRepo,
		draftRepo:      dRepo,
		buildRepo:      bRepo,
		releaseRepo:    rRepo,
		deploymentRepo: depRepo,
		projectSvc:     projSvc,
		buildStorage:   bStorage,
		mediaStorage:   mStorage,
		deployer:       deployer,
	}

	t.Cleanup(func() {
		_ = locker.Close()
		pool.Close()
		_ = redisContainer.Terminate(context.Background())
		_ = pgContainer.Terminate(context.Background())
	})

	return env
}

// createTables накатывает схему сервиса из migrations/init.sql — ту же, что применяется
// в docker-compose, чтобы тесты репозиториев проверяли реальную схему, а не её копию.
func createTables(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	schema, err := os.ReadFile(filepath.Join("..", "..", "migrations", "init.sql"))
	if err != nil {
		t.Fatalf("failed to read schema: %v", err)
	}
	if _, err := pool.Exec(ctx, string(schema)); err != nil {
		t.Fatalf("failed to apply schema: %v", err)
	}
}

func (e *TestEnv) cleanup(t *testing.T) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tables := []string{"deployments", "project_releases", "project_builds", "project_drafts", "projects"}
	for _, tbl := range tables {
		_, _ = e.pool.Exec(ctx, "TRUNCATE TABLE "+tbl+" CASCADE")
	}
}
