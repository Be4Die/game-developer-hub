package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Be4Die/game-developer-hub/moderation/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ModerationRepository struct {
	pool *pgxpool.Pool
}

func NewModerationRepository(pool *pgxpool.Pool) *ModerationRepository {
	return &ModerationRepository{pool: pool}
}

func (r *ModerationRepository) Create(ctx context.Context, m *domain.GameModeration) error {
	query := `
		INSERT INTO game_moderations (
			game_id, developer_id, game_name, game_description,
			status, submitted_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	return r.pool.QueryRow(ctx, query,
		m.GameID, m.DeveloperID, m.GameName, m.GameDescription,
		m.Status, m.SubmittedAt,
	).Scan(&m.ID)
}

func (r *ModerationRepository) GetByGameID(ctx context.Context, gameID int64) (*domain.GameModeration, error) {
	query := `
		SELECT id, game_id, developer_id, game_name, game_description,
		       moderator_id, status, rejection_reason, submitted_at, reviewed_at
		FROM game_moderations
		WHERE game_id = $1
		ORDER BY submitted_at DESC
		LIMIT 1
	`
	var m domain.GameModeration
	err := r.pool.QueryRow(ctx, query, gameID).Scan(
		&m.ID, &m.GameID, &m.DeveloperID, &m.GameName, &m.GameDescription,
		&m.ModeratorID, &m.Status, &m.RejectionReason, &m.SubmittedAt, &m.ReviewedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrModerationNotFound
		}
		return nil, fmt.Errorf("get moderation by game_id: %w", err)
	}
	return &m, nil
}

func (r *ModerationRepository) GetPending(ctx context.Context, limit, offset int) ([]domain.GameModeration, int, error) {
	countQuery := `SELECT COUNT(*) FROM game_moderations WHERE status = 'pending'`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count pending: %w", err)
	}

	query := `
		SELECT id, game_id, developer_id, game_name, game_description,
		       moderator_id, status, rejection_reason, submitted_at, reviewed_at
		FROM game_moderations
		WHERE status = 'pending'
		ORDER BY submitted_at ASC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("get pending: %w", err)
	}
	defer rows.Close()

	var moderations []domain.GameModeration
	for rows.Next() {
		var m domain.GameModeration
		if err := rows.Scan(
			&m.ID, &m.GameID, &m.DeveloperID, &m.GameName, &m.GameDescription,
			&m.ModeratorID, &m.Status, &m.RejectionReason, &m.SubmittedAt, &m.ReviewedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan row: %w", err)
		}
		moderations = append(moderations, m)
	}
	return moderations, total, nil
}

func (r *ModerationRepository) Update(ctx context.Context, m *domain.GameModeration) error {
	query := `
		UPDATE game_moderations SET
			moderator_id = $2,
			status = $3,
			rejection_reason = $4,
			reviewed_at = $5
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query,
		m.ID, m.ModeratorID, m.Status, m.RejectionReason, m.ReviewedAt,
	)
	if err != nil {
		return fmt.Errorf("update moderation: %w", err)
	}
	return nil
}

func InitSchema(ctx context.Context, pool *pgxpool.Pool) error {
	schema := `
		CREATE TABLE IF NOT EXISTS game_moderations (
			id BIGSERIAL PRIMARY KEY,
			game_id BIGINT NOT NULL UNIQUE,
			developer_id VARCHAR(255) NOT NULL,
			game_name VARCHAR(255) NOT NULL,
			game_description TEXT NOT NULL,
			moderator_id VARCHAR(255) DEFAULT '',
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			rejection_reason TEXT DEFAULT '',
			submitted_at TIMESTAMP NOT NULL DEFAULT NOW(),
			reviewed_at TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_game_moderations_status ON game_moderations(status);
		CREATE INDEX IF NOT EXISTS idx_game_moderations_game_id ON game_moderations(game_id);
	`
	_, err := pool.Exec(ctx, schema)
	return err
}
