package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// QueueEventRepo реализует domain.QueueEventRepo поверх PostgreSQL.
type QueueEventRepo struct {
	pool *pgxpool.Pool
}

// NewQueueEventRepo создаёт хранилище аудит-лога очереди.
func NewQueueEventRepo(pool *pgxpool.Pool) *QueueEventRepo {
	return &QueueEventRepo{pool: pool}
}

// Log записывает событие в аудит-лог.
func (r *QueueEventRepo) Log(ctx context.Context, gameID int64, playerID string, eventType domain.QueueEventType, instanceID int64, waitSeconds int) error {
	const q = `
		INSERT INTO queue_events (game_id, player_id, event_type, instance_id, wait_seconds, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`
	_, err := r.pool.Exec(ctx, q, gameID, playerID, eventType, instanceID, waitSeconds)
	if err != nil {
		return fmt.Errorf("postgres.QueueEventRepo.Log: %w", err)
	}
	return nil
}

// ListEvents возвращает события очереди игры.
func (r *QueueEventRepo) ListEvents(ctx context.Context, gameID int64, limit int) ([]*domain.QueueEvent, error) {
	const q = `
		SELECT id, game_id, player_id, event_type, instance_id, wait_seconds, created_at
		FROM queue_events
		WHERE game_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	rows, err := r.pool.Query(ctx, q, gameID, limit)
	if err != nil {
		return nil, fmt.Errorf("postgres.QueueEventRepo.ListEvents: %w", err)
	}
	defer rows.Close()

	var events []*domain.QueueEvent
	for rows.Next() {
		e := &domain.QueueEvent{}
		if err := rows.Scan(&e.ID, &e.GameID, &e.PlayerID, &e.EventType, &e.InstanceID, &e.WaitSeconds, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("postgres.QueueEventRepo.ListEvents: scan: %w", err)
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres.QueueEventRepo.ListEvents: rows: %w", err)
	}

	return events, nil
}

// GetStats возвращает агрегированную статистику очереди за период.
func (r *QueueEventRepo) GetStats(ctx context.Context, gameID int64, since time.Time) (*domain.QueueStats, error) {
	const q = `
		SELECT
			COUNT(*) FILTER (WHERE event_type = 1) AS total_joined,
			COUNT(*) FILTER (WHERE event_type = 2) AS total_reserved,
			COUNT(*) FILTER (WHERE event_type = 3) AS total_connected,
			COUNT(*) FILTER (WHERE event_type = 4) AS total_timeouts,
			COALESCE(AVG(wait_seconds) FILTER (WHERE event_type = 3), 0)::bigint AS avg_wait
		FROM queue_events
		WHERE game_id = $1 AND created_at >= $2
	`
	stats := &domain.QueueStats{}
	if err := r.pool.QueryRow(ctx, q, gameID, since).Scan(
		&stats.TotalJoined, &stats.TotalReserved, &stats.TotalConnected, &stats.TotalTimeouts, &stats.AvgWaitSeconds,
	); err != nil {
		return nil, fmt.Errorf("postgres.QueueEventRepo.GetStats: %w", err)
	}
	return stats, nil
}
