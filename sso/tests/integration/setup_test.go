//go:build integration

// Package integration содержит интеграционные тесты для SSO-сервиса.
//
// Эти тесты проверяют реальное взаимодействие с PostgreSQL и Valkey:
// - UserRepository через реальный PostgreSQL
// - SessionRepository через реальный PostgreSQL
// - EmailVerificationStore / PasswordResetStore / SessionCache через реальный Valkey
// - Полные потоки: register → verify email → login → refresh → logout
//
// Запуск:
//
//	go test -tags=integration ./tests/integration/...
package integration

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	bcryptpkg "github.com/Be4Die/game-developer-hub/sso/internal/infrastructure/bcrypt"
	jwtpkg "github.com/Be4Die/game-developer-hub/sso/internal/infrastructure/jwt"
	"github.com/Be4Die/game-developer-hub/sso/internal/service"
	pg "github.com/Be4Die/game-developer-hub/sso/internal/storage/postgres"
	"github.com/Be4Die/game-developer-hub/sso/internal/storage/valkey"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	testcontainerspostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	testcontainersredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

// IntegrationTestEnv хранит компоненты тестового окружения.
type IntegrationTestEnv struct {
	pgContainer    *testcontainerspostgres.PostgresContainer
	redisContainer *testcontainersredis.RedisContainer
	pool           *pgxpool.Pool
	redisClient    *redis.Client
	userRepo       *pg.UserRepository
	sessionRepo    *pg.SessionRepository
	emailStore     *valkey.EmailVerificationStore
	resetStore     *valkey.PasswordResetStore
	sessionCache   *valkey.SessionCache
	passwordHasher *bcryptpkg.PasswordHasher
	tokenManager   *jwtpkg.TokenManager
	authService    *service.AuthService
	userService    *service.UserService
	tokenService   *service.TokenService
	log            *slog.Logger
}

// setupIntegration запускает PostgreSQL и Valkey контейнеры, создаёт схему и инициализирует сервисы.
func setupIntegration(t *testing.T) *IntegrationTestEnv {
	t.Helper()
	ctx := context.Background()

	// ─── PostgreSQL ─────────────────────────────────────────────
	pgContainer, err := testcontainerspostgres.Run(ctx,
		"postgres:17-alpine",
		testcontainerspostgres.WithDatabase("sso"),
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
	createTables(t, pool)

	// ─── Репозитории и хранилища ────────────────────────────────
	userRepo := pg.NewUserRepository(pool)
	sessionRepo := pg.NewSessionRepository(pool)
	keyTTL := 10 * time.Minute
	emailStore := valkey.NewEmailVerificationStore(redisClient, keyTTL)
	resetStore := valkey.NewPasswordResetStore(redisClient, keyTTL)
	sessionCache := valkey.NewSessionCache(redisClient, keyTTL)

	// ─── Крипто-провайдеры ──────────────────────────────────────
	passwordHasher, err := bcryptpkg.NewPasswordHasher(bcryptpkg.DefaultCost)
	if err != nil {
		t.Fatalf("failed to create password hasher: %v", err)
	}
	tokenManager, err := jwtpkg.NewTokenManager("test-jwt-secret-for-integration", 15*time.Minute, "test-issuer")
	if err != nil {
		t.Fatalf("failed to create token manager: %v", err)
	}

	// ─── Сервисы ────────────────────────────────────────────────
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	emailSender := service.NewStubEmailSender(log)

	authService := service.NewAuthService(
		log, userRepo, sessionRepo, tokenManager, passwordHasher,
		emailStore, resetStore, emailSender, 24*time.Hour,
	)
	userService := service.NewUserService(log, userRepo, passwordHasher)
	tokenService := service.NewTokenService(log, sessionRepo, sessionCache, tokenManager)

	t.Cleanup(func() {
		_ = redisClient.Close()
		pool.Close()
		_ = pgContainer.Terminate(ctx)
		_ = redisContainer.Terminate(ctx)
	})

	return &IntegrationTestEnv{
		pgContainer:    pgContainer,
		redisContainer: redisContainer,
		pool:           pool,
		redisClient:    redisClient,
		userRepo:       userRepo,
		sessionRepo:    sessionRepo,
		emailStore:     emailStore,
		resetStore:     resetStore,
		sessionCache:   sessionCache,
		passwordHasher: passwordHasher,
		tokenManager:   tokenManager,
		authService:    authService,
		userService:    userService,
		tokenService:   tokenService,
		log:            log,
	}
}

// createTables создаёт таблицы в PostgreSQL.
func createTables(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()

	migrations := []string{
		`CREATE OR REPLACE FUNCTION update_updated_at()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = NOW();
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql`,

		`CREATE TABLE IF NOT EXISTS users (
			id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email         VARCHAR(255) NOT NULL UNIQUE,
			password_hash BYTEA NOT NULL,
			display_name  VARCHAR(255) NOT NULL DEFAULT '',
			role          SMALLINT NOT NULL DEFAULT 1,
			status        SMALLINT NOT NULL DEFAULT 1,
			email_verified BOOLEAN NOT NULL DEFAULT FALSE,
			created_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)`,

		`CREATE TABLE IF NOT EXISTS sessions (
			id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id          UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			user_agent       VARCHAR(512) NOT NULL DEFAULT '',
			ip_address       VARCHAR(45) NOT NULL DEFAULT '',
			refresh_token_hash VARCHAR(128) NOT NULL DEFAULT '',
			created_at       TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			last_used_at     TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			expires_at       TIMESTAMP WITH TIME ZONE NOT NULL,
			revoked          BOOLEAN NOT NULL DEFAULT FALSE,
			revoked_at       TIMESTAMP WITH TIME ZONE
		)`,

		`CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_refresh_token ON sessions(refresh_token_hash)`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
		`CREATE INDEX IF NOT EXISTS idx_users_status ON users(status)`,

		`CREATE TRIGGER IF NOT EXISTS trigger_users_updated_at
			BEFORE UPDATE ON users
			FOR EACH ROW
			EXECUTE FUNCTION update_updated_at()`,
	}

	for _, migration := range migrations {
		if _, err := pool.Exec(ctx, migration); err != nil {
			t.Fatalf("failed to execute migration: %v\nSQL: %s", err, migration)
		}
	}
}

// cleanupTables удаляет все данные из таблиц и flush'ит Valkey.
func (env *IntegrationTestEnv) CleanupTables(t *testing.T) {
	t.Helper()
	ctx := context.Background()

	_, err := env.pool.Exec(ctx, "TRUNCATE sessions, users RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatalf("failed to clean tables: %v", err)
	}

	if err := env.redisClient.FlushDB(ctx).Err(); err != nil {
		t.Fatalf("failed to flush valkey: %v", err)
	}
}

// TestMain — точка входа для integration тестов.
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
