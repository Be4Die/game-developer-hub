package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ModerationRepo реализует domain.ModerationRepo для работы с таблицей moderation_tickets.
type ModerationRepo struct {
	pool *pgxpool.Pool
}

// NewModerationRepo создаёт новый экземпляр репозитория модерации.
func NewModerationRepo(pool *pgxpool.Pool) *ModerationRepo {
	return &ModerationRepo{pool: pool}
}

// CreateTicket создаёт новый тикет модерации в базе данных.
func (r *ModerationRepo) CreateTicket(ctx context.Context, t *domain.ModerationTicket) (int64, error) {
	const query = `
		INSERT INTO moderation_tickets (
			project_id, owner_id, game_title, game_description, status,
			snapshot_meta, rejection_reason, moderator_id, dev_url, active_build_version
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, submitted_at
	`
	var id int64
	err := r.pool.QueryRow(ctx, query,
		t.ProjectID, t.OwnerID, t.GameTitle, t.GameDescription, t.Status,
		t.SnapshotMeta, t.RejectionReason, t.ModeratorID, t.DevURL, t.ActiveBuildVersion,
	).Scan(&id, &t.SubmittedAt)
	if err != nil {
		return 0, fmt.Errorf("postgres.ModerationRepo.CreateTicket: %w", err)
	}
	t.ID = id
	return id, nil
}

// GetTicket загружает тикет модерации по его ID.
func (r *ModerationRepo) GetTicket(ctx context.Context, id int64) (*domain.ModerationTicket, error) {
	const query = `
		SELECT id, project_id, owner_id, game_title, game_description, status,
		       snapshot_meta, rejection_reason, moderator_id, dev_url, active_build_version,
		       submitted_at, resolved_at
		FROM moderation_tickets
		WHERE id = $1
	`
	var t domain.ModerationTicket
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.ProjectID, &t.OwnerID, &t.GameTitle, &t.GameDescription, &t.Status,
		&t.SnapshotMeta, &t.RejectionReason, &t.ModeratorID, &t.DevURL, &t.ActiveBuildVersion,
		&t.SubmittedAt, &t.ResolvedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("postgres.ModerationRepo.GetTicket: %w", err)
	}
	return &t, nil
}

// GetLatestTicketByProject возвращает последний созданный тикет проекта.
func (r *ModerationRepo) GetLatestTicketByProject(ctx context.Context, projectID int64) (*domain.ModerationTicket, error) {
	const query = `
		SELECT id, project_id, owner_id, game_title, game_description, status,
		       snapshot_meta, rejection_reason, moderator_id, dev_url, active_build_version,
		       submitted_at, resolved_at
		FROM moderation_tickets
		WHERE project_id = $1
		ORDER BY submitted_at DESC
		LIMIT 1
	`
	var t domain.ModerationTicket
	err := r.pool.QueryRow(ctx, query, projectID).Scan(
		&t.ID, &t.ProjectID, &t.OwnerID, &t.GameTitle, &t.GameDescription, &t.Status,
		&t.SnapshotMeta, &t.RejectionReason, &t.ModeratorID, &t.DevURL, &t.ActiveBuildVersion,
		&t.SubmittedAt, &t.ResolvedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("postgres.ModerationRepo.GetLatestTicketByProject: %w", err)
	}
	return &t, nil
}

// ListTickets возвращает список тикетов модерации с опциональной фильтрацией по статусу.
func (r *ModerationRepo) ListTickets(ctx context.Context, status *domain.ModerationStatus, limit, offset int) ([]*domain.ModerationTicket, error) {
	if limit <= 0 {
		limit = 50
	}

	var (
		rows pgx.Rows
		err  error
	)

	if status != nil {
		const query = `
			SELECT id, project_id, owner_id, game_title, game_description, status,
			       snapshot_meta, rejection_reason, moderator_id, dev_url, active_build_version,
			       submitted_at, resolved_at
			FROM moderation_tickets
			WHERE status = $1
			ORDER BY submitted_at ASC
			LIMIT $2 OFFSET $3
		`
		rows, err = r.pool.Query(ctx, query, *status, limit, offset)
	} else {
		const query = `
			SELECT id, project_id, owner_id, game_title, game_description, status,
			       snapshot_meta, rejection_reason, moderator_id, dev_url, active_build_version,
			       submitted_at, resolved_at
			FROM moderation_tickets
			ORDER BY submitted_at DESC
			LIMIT $1 OFFSET $2
		`
		rows, err = r.pool.Query(ctx, query, limit, offset)
	}

	if err != nil {
		return nil, fmt.Errorf("postgres.ModerationRepo.ListTickets: %w", err)
	}
	defer rows.Close()

	var tickets []*domain.ModerationTicket
	for rows.Next() {
		var t domain.ModerationTicket
		if err := rows.Scan(
			&t.ID, &t.ProjectID, &t.OwnerID, &t.GameTitle, &t.GameDescription, &t.Status,
			&t.SnapshotMeta, &t.RejectionReason, &t.ModeratorID, &t.DevURL, &t.ActiveBuildVersion,
			&t.SubmittedAt, &t.ResolvedAt,
		); err != nil {
			return nil, fmt.Errorf("postgres.ModerationRepo.ListTickets scan: %w", err)
		}
		tickets = append(tickets, &t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres.ModerationRepo.ListTickets rows: %w", err)
	}

	return tickets, nil
}

// ResolveTicket фиксирует вердикт модератора по тикету.
func (r *ModerationRepo) ResolveTicket(ctx context.Context, id int64, status domain.ModerationStatus, rejectionReason, moderatorID string) error {
	const query = `
		UPDATE moderation_tickets
		SET status = $1, rejection_reason = $2, moderator_id = $3, resolved_at = $4
		WHERE id = $5
	`
	now := time.Now()
	_, err := r.pool.Exec(ctx, query, status, rejectionReason, moderatorID, now, id)
	if err != nil {
		return fmt.Errorf("postgres.ModerationRepo.ResolveTicket: %w", err)
	}
	return nil
}
