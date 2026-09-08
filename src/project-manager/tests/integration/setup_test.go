//go:build integration

package integration

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/infrastructure/valkey"
	"github.com/Be4Die/game-developer-hub/project-manager/internal/service"
	"github.com/Be4Die/game-developer-hub/project-manager/internal/storage/deployment"
	"github.com/Be4Die/game-developer-hub/project-manager/internal/storage/filesystem"
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
	buildStorage   *filesystem.BuildStorage
	mediaStorage   *filesystem.MediaStorage
	deployer       *deployment.LocalDeployer
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

	tmpDir := t.TempDir()
	bStorage := filesystem.NewBuildStorage(filepath.Join(tmpDir, "projects"))
	mStorage := filesystem.NewMediaStorage(filepath.Join(tmpDir, "projects"))
	deployer := deployment.NewLocalDeployer(filepath.Join(tmpDir, "games"), "/games")
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

func createTables(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()

	ddl := []string{
		`CREATE TABLE IF NOT EXISTS projects (
			id          BIGSERIAL PRIMARY KEY,
			owner_id    TEXT NOT NULL,
			status      SMALLINT NOT NULL DEFAULT 1,
			created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS project_drafts (
			project_id            BIGINT PRIMARY KEY REFERENCES projects(id) ON DELETE CASCADE,
			title_ru              VARCHAR(50) NOT NULL DEFAULT '',
			title_en              VARCHAR(50) NOT NULL DEFAULT '',
			seo_ru                VARCHAR(180) NOT NULL DEFAULT '',
			seo_en                VARCHAR(180) NOT NULL DEFAULT '',
			about_ru              VARCHAR(800) NOT NULL DEFAULT '',
			about_en              VARCHAR(800) NOT NULL DEFAULT '',
			icon_path             TEXT NOT NULL DEFAULT '',
			cover_path            TEXT NOT NULL DEFAULT '',
			video_path            TEXT NOT NULL DEFAULT '',
			active_build_version  TEXT NOT NULL DEFAULT '',
			dev_url               TEXT NOT NULL DEFAULT '',
			updated_at            TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS project_builds (
			id            BIGSERIAL PRIMARY KEY,
			project_id    BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			version       TEXT NOT NULL,
			file_path     TEXT NOT NULL,
			file_size     BIGINT NOT NULL DEFAULT 0,
			is_unpacked   BOOLEAN NOT NULL DEFAULT FALSE,
			unpacked_path TEXT NOT NULL DEFAULT '',
			created_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			UNIQUE (project_id, version)
		)`,
		`CREATE TABLE IF NOT EXISTS project_releases (
			id                    BIGSERIAL PRIMARY KEY,
			project_id            BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			version               TEXT NOT NULL,
			title_ru              VARCHAR(50) NOT NULL DEFAULT '',
			title_en              VARCHAR(50) NOT NULL DEFAULT '',
			seo_ru                VARCHAR(180) NOT NULL DEFAULT '',
			seo_en                VARCHAR(180) NOT NULL DEFAULT '',
			about_ru              VARCHAR(800) NOT NULL DEFAULT '',
			about_en              VARCHAR(800) NOT NULL DEFAULT '',
			icon_path             TEXT NOT NULL DEFAULT '',
			cover_path            TEXT NOT NULL DEFAULT '',
			video_path            TEXT NOT NULL DEFAULT '',
			prod_url              TEXT NOT NULL DEFAULT '',
			is_active             BOOLEAN NOT NULL DEFAULT TRUE,
			published_by          TEXT NOT NULL DEFAULT '',
			published_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			unpublish_at          TIMESTAMP WITH TIME ZONE
		)`,
		`CREATE TABLE IF NOT EXISTS deployments (
			id            BIGSERIAL PRIMARY KEY,
			project_id    BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			environment   SMALLINT NOT NULL DEFAULT 1,
			version       TEXT NOT NULL DEFAULT '',
			status        SMALLINT NOT NULL DEFAULT 1,
			error_message TEXT NOT NULL DEFAULT '',
			deployed_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)`,
	}

	for _, stmt := range ddl {
		if _, err := pool.Exec(ctx, stmt); err != nil {
			t.Fatalf("failed to execute DDL: %v", err)
		}
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
