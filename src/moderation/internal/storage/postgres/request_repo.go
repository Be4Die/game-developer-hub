// Package postgres реализует адаптеры постоянного хранения данных в PostgreSQL для сервиса moderation.
package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Be4Die/game-developer-hub/moderation/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RequestRepo реализует domain.RequestRepo поверх PostgreSQL.
type RequestRepo struct {
	pool *pgxpool.Pool
}

// NewRequestRepo создаёт новый экземпляр RequestRepo.
func NewRequestRepo(pool *pgxpool.Pool) *RequestRepo {
	return &RequestRepo{pool: pool}
}

// Create создает новый запрос на модерацию.
func (r *RequestRepo) Create(ctx context.Context, req *domain.ModerationRequest) (int64, error) {
	snapshotJSON, err := json.Marshal(req.Snapshot)
	if err != nil {
		return 0, fmt.Errorf("marshal snapshot: %w", err)
	}

	query := `
		INSERT INTO moderation_requests (
			project_id, owner_id, moderator_id, status, snapshot_meta,
			rejection_reason, submitted_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, NOW(), NOW(), NOW()
		)
		RETURNING id, submitted_at, created_at, updated_at
	`

	var id int64
	var submittedAt, createdAt, updatedAt time.Time
	err = r.pool.QueryRow(
		ctx, query,
		req.ProjectID,
		req.OwnerID,
		req.ModeratorID,
		int16(req.Status),
		snapshotJSON,
		req.RejectionReason,
	).Scan(&id, &submittedAt, &createdAt, &updatedAt)
	if err != nil {
		return 0, fmt.Errorf("RequestRepo.Create: %w", err)
	}

	req.ID = id
	req.SubmittedAt = submittedAt
	req.CreatedAt = createdAt
	req.UpdatedAt = updatedAt

	return id, nil
}

// Get возвращает запрос на модерацию по его первичному ключу.
func (r *RequestRepo) Get(ctx context.Context, id int64) (*domain.ModerationRequest, error) {
	query := `
		SELECT
			id, project_id, owner_id, moderator_id, status, snapshot_meta,
			rejection_reason, submitted_at, started_review_at, resolved_at,
			created_at, updated_at
		FROM moderation_requests
		WHERE id = $1
	`

	req := &domain.ModerationRequest{}
	var snapshotRaw []byte
	var statusInt int16

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&req.ID,
		&req.ProjectID,
		&req.OwnerID,
		&req.ModeratorID,
		&statusInt,
		&snapshotRaw,
		&req.RejectionReason,
		&req.SubmittedAt,
		&req.StartedReviewAt,
		&req.ResolvedAt,
		&req.CreatedAt,
		&req.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("RequestRepo.Get: %w", err)
	}

	req.Status = domain.RequestStatus(statusInt)
	if len(snapshotRaw) > 0 {
		_ = json.Unmarshal(snapshotRaw, &req.Snapshot)
	}

	return req, nil
}

// GetLatestByProject возвращает последнюю отправленную заявку по ID проекта.
func (r *RequestRepo) GetLatestByProject(ctx context.Context, projectID int64) (*domain.ModerationRequest, error) {
	query := `
		SELECT
			id, project_id, owner_id, moderator_id, status, snapshot_meta,
			rejection_reason, submitted_at, started_review_at, resolved_at,
			created_at, updated_at
		FROM moderation_requests
		WHERE project_id = $1
		ORDER BY submitted_at DESC
		LIMIT 1
	`

	req := &domain.ModerationRequest{}
	var snapshotRaw []byte
	var statusInt int16

	err := r.pool.QueryRow(ctx, query, projectID).Scan(
		&req.ID,
		&req.ProjectID,
		&req.OwnerID,
		&req.ModeratorID,
		&statusInt,
		&snapshotRaw,
		&req.RejectionReason,
		&req.SubmittedAt,
		&req.StartedReviewAt,
		&req.ResolvedAt,
		&req.CreatedAt,
		&req.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("RequestRepo.GetLatestByProject: %w", err)
	}

	req.Status = domain.RequestStatus(statusInt)
	if len(snapshotRaw) > 0 {
		_ = json.Unmarshal(snapshotRaw, &req.Snapshot)
	}

	return req, nil
}

// List возвращает постраничный список запросов согласно переданным фильтрам.
func (r *RequestRepo) List(ctx context.Context, filter domain.RequestFilter) ([]*domain.ModerationRequest, int, error) {
	var whereConditions []string
	var args []any
	argIdx := 1

	if filter.Status != nil && *filter.Status != domain.RequestStatusUnspecified {
		whereConditions = append(whereConditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, int16(*filter.Status))
		argIdx++
	}

	if filter.ModeratorID != nil && *filter.ModeratorID != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("moderator_id = $%d", argIdx))
		args = append(args, *filter.ModeratorID)
		argIdx++
	}

	whereSQL := ""
	if len(whereConditions) > 0 {
		whereSQL = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Подсчёт общего числа записей
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM moderation_requests %s", whereSQL)
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("RequestRepo.List count: %w", err)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
		SELECT
			id, project_id, owner_id, moderator_id, status, snapshot_meta,
			rejection_reason, submitted_at, started_review_at, resolved_at,
			created_at, updated_at
		FROM moderation_requests
		%s
		ORDER BY submitted_at ASC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argIdx, argIdx+1)

	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("RequestRepo.List select: %w", err)
	}
	defer rows.Close()

	var requests []*domain.ModerationRequest
	for rows.Next() {
		req := &domain.ModerationRequest{}
		var snapshotRaw []byte
		var statusInt int16

		err := rows.Scan(
			&req.ID,
			&req.ProjectID,
			&req.OwnerID,
			&req.ModeratorID,
			&statusInt,
			&snapshotRaw,
			&req.RejectionReason,
			&req.SubmittedAt,
			&req.StartedReviewAt,
			&req.ResolvedAt,
			&req.CreatedAt,
			&req.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("RequestRepo.List scan: %w", err)
		}

		req.Status = domain.RequestStatus(statusInt)
		if len(snapshotRaw) > 0 {
			_ = json.Unmarshal(snapshotRaw, &req.Snapshot)
		}

		requests = append(requests, req)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("RequestRepo.List rows: %w", err)
	}

	return requests, total, nil
}

// Claim закрепляет заявку за модератором и переводит в статус IN_REVIEW.
func (r *RequestRepo) Claim(ctx context.Context, id int64, moderatorID string) error {
	query := `
		UPDATE moderation_requests
		SET
			moderator_id = $2,
			status = $3,
			started_review_at = COALESCE(started_review_at, NOW()),
			updated_at = NOW()
		WHERE id = $1 AND (moderator_id = '' OR moderator_id = $2) AND status IN (1, 2)
	`

	tag, err := r.pool.Exec(ctx, query, id, moderatorID, int16(domain.RequestStatusInReview))
	if err != nil {
		return fmt.Errorf("RequestRepo.Claim: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrAlreadyClaimed
	}

	return nil
}

// Resolve переводит заявку в финальный статус (APPROVED или REJECTED).
func (r *RequestRepo) Resolve(ctx context.Context, id int64, status domain.RequestStatus, reason, moderatorID string) error {
	query := `
		UPDATE moderation_requests
		SET
			status = $2,
			rejection_reason = $3,
			moderator_id = CASE WHEN moderator_id = '' THEN $4 ELSE moderator_id END,
			resolved_at = NOW(),
			updated_at = NOW()
		WHERE id = $1
	`

	tag, err := r.pool.Exec(ctx, query, id, int16(status), reason, moderatorID)
	if err != nil {
		return fmt.Errorf("RequestRepo.Resolve: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
