package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Be4Die/game-developer-hub/moderation/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SnapshotRepo реализует domain.SnapshotRepo поверх PostgreSQL.
type SnapshotRepo struct {
	pool *pgxpool.Pool
}

// NewSnapshotRepo создаёт новый экземпляр SnapshotRepo.
func NewSnapshotRepo(pool *pgxpool.Pool) *SnapshotRepo {
	return &SnapshotRepo{pool: pool}
}

// Save сохраняет или обновляет неизменяемый аудит-снимок.
func (r *SnapshotRepo) Save(ctx context.Context, s *domain.ModerationSnapshot) error {
	query := `
		INSERT INTO moderation_snapshots (
			request_id, project_id, status, format_version, snapshot_json, snapshot_blob, created_at
		) VALUES (
			$1, $2, $3, $4, $5::jsonb, $6, NOW()
		)
		ON CONFLICT (request_id) DO UPDATE SET
			status = EXCLUDED.status,
			format_version = EXCLUDED.format_version,
			snapshot_json = EXCLUDED.snapshot_json,
			snapshot_blob = EXCLUDED.snapshot_blob
		RETURNING id, created_at
	`

	jsonStr := s.SnapshotJSON
	if jsonStr == "" {
		jsonStr = "{}"
	}

	var id int64
	var createdAt time.Time
	err := r.pool.QueryRow(
		ctx, query,
		s.RequestID,
		s.ProjectID,
		int16(s.Status),
		s.FormatVersion,
		jsonStr,
		s.SnapshotBlob,
	).Scan(&id, &createdAt)
	if err != nil {
		return fmt.Errorf("SnapshotRepo.Save: %w", err)
	}

	s.ID = id
	s.CreatedAt = createdAt
	return nil
}

// GetByRequestID возвращает аудит-снимок по ID заявки на модерацию.
func (r *SnapshotRepo) GetByRequestID(ctx context.Context, requestID int64) (*domain.ModerationSnapshot, error) {
	query := `
		SELECT id, request_id, project_id, status, format_version, snapshot_json::text, snapshot_blob, created_at
		FROM moderation_snapshots
		WHERE request_id = $1
	`

	var s domain.ModerationSnapshot
	var statusInt int16
	err := r.pool.QueryRow(ctx, query, requestID).Scan(
		&s.ID,
		&s.RequestID,
		&s.ProjectID,
		&statusInt,
		&s.FormatVersion,
		&s.SnapshotJSON,
		&s.SnapshotBlob,
		&s.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("SnapshotRepo.GetByRequestID: %w", err)
	}

	s.Status = domain.RequestStatus(statusInt)
	return &s, nil
}

// GetByProjectID возвращает последний аудит-снимок по ID проекта.
func (r *SnapshotRepo) GetByProjectID(ctx context.Context, projectID int64) (*domain.ModerationSnapshot, error) {
	query := `
		SELECT id, request_id, project_id, status, format_version, snapshot_json::text, snapshot_blob, created_at
		FROM moderation_snapshots
		WHERE project_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var s domain.ModerationSnapshot
	var statusInt int16
	err := r.pool.QueryRow(ctx, query, projectID).Scan(
		&s.ID,
		&s.RequestID,
		&s.ProjectID,
		&statusInt,
		&s.FormatVersion,
		&s.SnapshotJSON,
		&s.SnapshotBlob,
		&s.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("SnapshotRepo.GetByProjectID: %w", err)
	}

	s.Status = domain.RequestStatus(statusInt)
	return &s, nil
}
