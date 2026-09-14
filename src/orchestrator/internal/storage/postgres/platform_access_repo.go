package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PlatformAccessRepo реализует domain.PlatformAccessRepo поверх PostgreSQL.
type PlatformAccessRepo struct {
	pool *pgxpool.Pool
}

// NewPlatformAccessRepo создаёт новый репозиторий заявок на платформенные серверы.
func NewPlatformAccessRepo(pool *pgxpool.Pool) *PlatformAccessRepo {
	return &PlatformAccessRepo{pool: pool}
}

// SaveGrant сохраняет или обновляет квоту доступа к платформенным серверам для игры.
func (r *PlatformAccessRepo) SaveGrant(ctx context.Context, g *domain.PlatformGrant) (*domain.PlatformGrant, error) {
	if g.MaxInstances <= 0 {
		g.MaxInstances = 2
	}
	now := time.Now()

	const q = `
		INSERT INTO platform_grants (
			game_id, max_instances, max_total_cpu_millis, max_total_memory_mb,
			max_instance_cpu_millis, max_instance_memory_mb, is_active, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, true, $7, $7)
		ON CONFLICT (game_id) DO UPDATE
		SET max_instances = EXCLUDED.max_instances,
		    max_total_cpu_millis = EXCLUDED.max_total_cpu_millis,
		    max_total_memory_mb = EXCLUDED.max_total_memory_mb,
		    max_instance_cpu_millis = EXCLUDED.max_instance_cpu_millis,
		    max_instance_memory_mb = EXCLUDED.max_instance_memory_mb,
		    is_active = true,
		    updated_at = EXCLUDED.updated_at
		RETURNING game_id, max_instances, max_total_cpu_millis, max_total_memory_mb,
		          max_instance_cpu_millis, max_instance_memory_mb, is_active, created_at, updated_at
	`

	res := &domain.PlatformGrant{}
	err := r.pool.QueryRow(ctx, q,
		g.GameID, g.MaxInstances, g.MaxTotalCPUMillis, g.MaxTotalMemoryMB,
		g.MaxInstanceCPUMillis, g.MaxInstanceMemoryMB, now,
	).Scan(
		&res.GameID,
		&res.MaxInstances,
		&res.MaxTotalCPUMillis,
		&res.MaxTotalMemoryMB,
		&res.MaxInstanceCPUMillis,
		&res.MaxInstanceMemoryMB,
		&res.IsActive,
		&res.CreatedAt,
		&res.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("postgres.PlatformAccessRepo.SaveGrant: %w", err)
	}
	return res, nil
}

// RevokeGrant отзывает доступ к платформенным мощностям для игры.
func (r *PlatformAccessRepo) RevokeGrant(ctx context.Context, gameID int64) error {
	const q = `
		UPDATE platform_grants
		SET is_active = false, updated_at = NOW()
		WHERE game_id = $1
	`
	_, err := r.pool.Exec(ctx, q, gameID)
	if err != nil {
		return fmt.Errorf("postgres.PlatformAccessRepo.RevokeGrant: %w", err)
	}
	return nil
}

// GetGrant возвращает текущий грант проекта.
func (r *PlatformAccessRepo) GetGrant(ctx context.Context, gameID int64) (*domain.PlatformGrant, error) {
	const q = `
		SELECT game_id, max_instances, max_total_cpu_millis, max_total_memory_mb,
		       max_instance_cpu_millis, max_instance_memory_mb, is_active, created_at, updated_at
		FROM platform_grants
		WHERE game_id = $1
	`
	grant := &domain.PlatformGrant{}
	err := r.pool.QueryRow(ctx, q, gameID).Scan(
		&grant.GameID,
		&grant.MaxInstances,
		&grant.MaxTotalCPUMillis,
		&grant.MaxTotalMemoryMB,
		&grant.MaxInstanceCPUMillis,
		&grant.MaxInstanceMemoryMB,
		&grant.IsActive,
		&grant.CreatedAt,
		&grant.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("postgres.PlatformAccessRepo.GetGrant: %w", err)
	}
	return grant, nil
}

// HasApprovedAccess проверяет, имеет ли проект активный доступ к серверам платформы, и возвращает грант.
func (r *PlatformAccessRepo) HasApprovedAccess(ctx context.Context, gameID int64) (bool, *domain.PlatformGrant, error) {
	grant, err := r.GetGrant(ctx, gameID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return false, nil, nil
		}
		return false, nil, err
	}
	if !grant.IsActive {
		return false, nil, nil
	}
	return true, grant, nil
}

// CountPlatformInstances подсчитывает число запущенных/стартующих инстансов проекта на платформенных нодах.
func (r *PlatformAccessRepo) CountPlatformInstances(ctx context.Context, gameID int64) (int32, error) {
	count, _, _, err := r.GetActivePlatformUsage(ctx, gameID)
	return count, err
}

// GetActivePlatformUsage подсчитывает число запущенных/стартующих инстансов проекта на платформенных нодах
// и суммарно выделенные им ресурсы (CPU millis и memory bytes).
func (r *PlatformAccessRepo) GetActivePlatformUsage(ctx context.Context, gameID int64) (count int32, allocatedCPUMillis uint32, allocatedMemoryBytes uint64, err error) {
	const q = `
		SELECT COUNT(*),
		       COALESCE(SUM(i.allocated_cpu_millis), 0),
		       COALESCE(SUM(i.allocated_memory_bytes), 0)
		FROM instances i
		JOIN nodes n ON i.node_id = n.id
		WHERE i.game_id = $1 AND n.is_platform = true AND i.status IN (1, 2)
	`
	var (
		c    int64
		cpu  int64
		mBytes int64
	)
	err = r.pool.QueryRow(ctx, q, gameID).Scan(&c, &cpu, &mBytes)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("postgres.PlatformAccessRepo.GetActivePlatformUsage: %w", err)
	}
	return int32(c), uint32(cpu), uint64(mBytes), nil
}
