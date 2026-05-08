package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GamePolicyRepo реализует domain.GamePolicyRepo поверх PostgreSQL.
type GamePolicyRepo struct {
	pool *pgxpool.Pool
}

// NewGamePolicyRepo создаёт хранилище политик.
func NewGamePolicyRepo(pool *pgxpool.Pool) *GamePolicyRepo {
	return &GamePolicyRepo{pool: pool}
}

// Get возвращает политику игры. Возвращает ErrNotFound при отсутствии.
func (r *GamePolicyRepo) Get(ctx context.Context, gameID int64) (*domain.GamePolicy, error) {
	const q = `
		SELECT game_id, mode, target_instances, auto_restart, scale_to_zero_timeout,
		       default_build_version, max_players_per_instance, max_instances_per_game,
		       scale_behavior, node_preference, created_at, updated_at
		FROM game_policies WHERE game_id = $1
	`

	row := r.pool.QueryRow(ctx, q, gameID)
	return scanGamePolicy(row)
}

// Set создаёт или обновляет политику игры (UPSERT).
func (r *GamePolicyRepo) Set(ctx context.Context, policy *domain.GamePolicy) error {
	const q = `
		INSERT INTO game_policies (game_id, mode, target_instances, auto_restart, scale_to_zero_timeout,
		                           default_build_version, max_players_per_instance, max_instances_per_game,
		                           scale_behavior, node_preference, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		ON CONFLICT (game_id) DO UPDATE SET
			mode = EXCLUDED.mode,
			target_instances = EXCLUDED.target_instances,
			auto_restart = EXCLUDED.auto_restart,
			scale_to_zero_timeout = EXCLUDED.scale_to_zero_timeout,
			default_build_version = EXCLUDED.default_build_version,
			max_players_per_instance = EXCLUDED.max_players_per_instance,
			max_instances_per_game = EXCLUDED.max_instances_per_game,
			scale_behavior = EXCLUDED.scale_behavior,
			node_preference = EXCLUDED.node_preference,
			updated_at = NOW()
	`

	_, err := r.pool.Exec(ctx, q,
		policy.GameID, policy.Mode, policy.TargetInstances, policy.AutoRestart, policy.ScaleToZeroTimeout,
		policy.DefaultBuildVersion, policy.MaxPlayersPerInstance, policy.MaxInstancesPerGame,
		policy.ScaleBehavior, policy.NodePreference,
	)
	if err != nil {
		return fmt.Errorf("postgres.GamePolicyRepo.Set: %w", err)
	}

	return nil
}

// Delete удаляет политику игры. Возвращает ErrNotFound при отсутствии.
func (r *GamePolicyRepo) Delete(ctx context.Context, gameID int64) error {
	const q = `DELETE FROM game_policies WHERE game_id = $1`

	tag, err := r.pool.Exec(ctx, q, gameID)
	if err != nil {
		return fmt.Errorf("postgres.GamePolicyRepo.Delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

type gamePolicyScanner interface {
	Scan(dest ...any) error
}

func scanGamePolicy(s gamePolicyScanner) (*domain.GamePolicy, error) {
	p := &domain.GamePolicy{}
	err := s.Scan(
		&p.GameID, &p.Mode, &p.TargetInstances, &p.AutoRestart, &p.ScaleToZeroTimeout,
		&p.DefaultBuildVersion, &p.MaxPlayersPerInstance, &p.MaxInstancesPerGame,
		&p.ScaleBehavior, &p.NodePreference, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("postgres.scanGamePolicy: %w", err)
	}

	return p, nil
}
