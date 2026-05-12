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
		SELECT game_id, owner_id, mode, target_instances, auto_restart, scale_to_zero_timeout,
		       default_build_version, max_players_per_instance, max_instances_per_game,
		       scale_behavior, node_preference, queue_location, queue_scale_up_threshold,
		       queue_reservation_seconds, queue_max_wait_seconds, queue_heartbeat_timeout,
		       created_at, updated_at
		FROM game_policies WHERE game_id = $1
	`

	row := r.pool.QueryRow(ctx, q, gameID)
	return scanGamePolicy(row)
}

// Set создаёт или обновляет политику игры (UPSERT).
func (r *GamePolicyRepo) Set(ctx context.Context, policy *domain.GamePolicy) error {
	const q = `
		INSERT INTO game_policies (game_id, owner_id, mode, target_instances, auto_restart, scale_to_zero_timeout,
		                           default_build_version, max_players_per_instance, max_instances_per_game,
		                           scale_behavior, node_preference, queue_location, queue_scale_up_threshold,
		                           queue_reservation_seconds, queue_max_wait_seconds, queue_heartbeat_timeout,
		                           created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, NOW(), NOW())
		ON CONFLICT (game_id) DO UPDATE SET
			owner_id = EXCLUDED.owner_id,
			mode = EXCLUDED.mode,
			target_instances = EXCLUDED.target_instances,
			auto_restart = EXCLUDED.auto_restart,
			scale_to_zero_timeout = EXCLUDED.scale_to_zero_timeout,
			default_build_version = EXCLUDED.default_build_version,
			max_players_per_instance = EXCLUDED.max_players_per_instance,
			max_instances_per_game = EXCLUDED.max_instances_per_game,
			scale_behavior = EXCLUDED.scale_behavior,
			node_preference = EXCLUDED.node_preference,
			queue_location = EXCLUDED.queue_location,
			queue_scale_up_threshold = EXCLUDED.queue_scale_up_threshold,
			queue_reservation_seconds = EXCLUDED.queue_reservation_seconds,
			queue_max_wait_seconds = EXCLUDED.queue_max_wait_seconds,
			queue_heartbeat_timeout = EXCLUDED.queue_heartbeat_timeout,
			updated_at = NOW()
	`

	_, err := r.pool.Exec(ctx, q,
		policy.GameID, policy.OwnerID, policy.Mode, policy.TargetInstances, policy.AutoRestart, policy.ScaleToZeroTimeout,
		policy.DefaultBuildVersion, policy.MaxPlayersPerInstance, policy.MaxInstancesPerGame,
		policy.ScaleBehavior, policy.NodePreference, policy.QueueLocation, policy.QueueScaleUpThreshold,
		policy.QueueReservationSec, policy.QueueMaxWaitSec, policy.QueueHeartbeatTimeout,
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

// ListAll возвращает все сохранённые политики.
func (r *GamePolicyRepo) ListAll(ctx context.Context) ([]*domain.GamePolicy, error) {
	const q = `
		SELECT game_id, owner_id, mode, target_instances, auto_restart, scale_to_zero_timeout,
		       default_build_version, max_players_per_instance, max_instances_per_game,
		       scale_behavior, node_preference, queue_location, queue_scale_up_threshold,
		       queue_reservation_seconds, queue_max_wait_seconds, queue_heartbeat_timeout,
		       created_at, updated_at
		FROM game_policies
		ORDER BY game_id
	`

	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("postgres.GamePolicyRepo.ListAll: %w", err)
	}
	defer rows.Close()

	var policies []*domain.GamePolicy
	for rows.Next() {
		p, err := scanGamePolicy(rows)
		if err != nil {
			return nil, fmt.Errorf("postgres.GamePolicyRepo.ListAll: scan: %w", err)
		}
		policies = append(policies, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres.GamePolicyRepo.ListAll: rows error: %w", err)
	}

	return policies, nil
}

type gamePolicyScanner interface {
	Scan(dest ...any) error
}

func scanGamePolicy(s gamePolicyScanner) (*domain.GamePolicy, error) {
	p := &domain.GamePolicy{}
	err := s.Scan(
		&p.GameID, &p.OwnerID, &p.Mode, &p.TargetInstances, &p.AutoRestart, &p.ScaleToZeroTimeout,
		&p.DefaultBuildVersion, &p.MaxPlayersPerInstance, &p.MaxInstancesPerGame,
		&p.ScaleBehavior, &p.NodePreference, &p.QueueLocation, &p.QueueScaleUpThreshold,
		&p.QueueReservationSec, &p.QueueMaxWaitSec, &p.QueueHeartbeatTimeout,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("postgres.scanGamePolicy: %w", err)
	}

	return p, nil
}
