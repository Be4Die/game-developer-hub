package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// InvitationRepo реализует domain.InvitationRepo для таблицы project_invitations.
type InvitationRepo struct {
	pool *pgxpool.Pool
}

// NewInvitationRepo создаёт репозиторий приглашений.
func NewInvitationRepo(pool *pgxpool.Pool) *InvitationRepo {
	return &InvitationRepo{pool: pool}
}

// Create сохраняет новое приглашение.
func (r *InvitationRepo) Create(ctx context.Context, inv *domain.Invitation) (int64, error) {
	const query = `
		INSERT INTO project_invitations (
			project_id, inviter_id, inviter_email, inviter_name,
			invitee_id, invitee_email, permissions, status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`
	status := inv.Status
	if status == 0 {
		status = domain.InvitationStatusPending
	}
	var id int64
	err := r.pool.QueryRow(
		ctx, query,
		inv.ProjectID, inv.InviterID, inv.InviterEmail, inv.InviterName,
		inv.InviteeID, inv.InviteeEmail, inv.Permissions, status,
	).Scan(&id, &inv.CreatedAt, &inv.UpdatedAt)
	if err != nil {
		return 0, fmt.Errorf("postgres.InvitationRepo.Create: %w", err)
	}
	inv.ID = id
	inv.Status = status
	return id, nil
}

// Get возвращает приглашение по ID.
func (r *InvitationRepo) Get(ctx context.Context, id int64) (*domain.Invitation, error) {
	const query = `
		SELECT i.id, i.project_id, i.inviter_id, i.inviter_email, i.inviter_name,
		       i.invitee_id, i.invitee_email, i.permissions, i.status, i.created_at, i.updated_at,
		       COALESCE(NULLIF(d.title_ru, ''), NULLIF(d.title_en, ''), 'Проект #' || i.project_id::text) AS project_title,
		       COALESCE(d.icon_path, '') AS project_icon
		FROM project_invitations i
		LEFT JOIN project_drafts d ON d.project_id = i.project_id
		WHERE i.id = $1
	`
	var inv domain.Invitation
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&inv.ID, &inv.ProjectID, &inv.InviterID, &inv.InviterEmail, &inv.InviterName,
		&inv.InviteeID, &inv.InviteeEmail, &inv.Permissions, &inv.Status, &inv.CreatedAt, &inv.UpdatedAt,
		&inv.ProjectTitle, &inv.ProjectIcon,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("postgres.InvitationRepo.Get: %w", err)
	}
	return &inv, nil
}


// GetPending возвращает активное (PENDING) приглашение для пользователя на проект.
func (r *InvitationRepo) GetPending(ctx context.Context, projectID int64, inviteeID string) (*domain.Invitation, error) {
	const query = `
		SELECT i.id, i.project_id, i.inviter_id, i.inviter_email, i.inviter_name,
		       i.invitee_id, i.invitee_email, i.permissions, i.status, i.created_at, i.updated_at,
		       COALESCE(NULLIF(d.title_ru, ''), NULLIF(d.title_en, ''), 'Проект #' || i.project_id::text) AS project_title,
		       COALESCE(d.icon_path, '') AS project_icon
		FROM project_invitations i
		LEFT JOIN project_drafts d ON d.project_id = i.project_id
		WHERE i.project_id = $1 AND i.invitee_id = $2 AND i.status = 1
		LIMIT 1
	`
	var inv domain.Invitation
	err := r.pool.QueryRow(ctx, query, projectID, inviteeID).Scan(
		&inv.ID, &inv.ProjectID, &inv.InviterID, &inv.InviterEmail, &inv.InviterName,
		&inv.InviteeID, &inv.InviteeEmail, &inv.Permissions, &inv.Status, &inv.CreatedAt, &inv.UpdatedAt,
		&inv.ProjectTitle, &inv.ProjectIcon,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("postgres.InvitationRepo.GetPending: %w", err)
	}
	return &inv, nil
}

// ListIncoming возвращает входящие активные (PENDING) приглашения для пользователя.
func (r *InvitationRepo) ListIncoming(ctx context.Context, inviteeID string) ([]*domain.Invitation, error) {
	const query = `
		SELECT i.id, i.project_id, i.inviter_id, i.inviter_email, i.inviter_name,
		       i.invitee_id, i.invitee_email, i.permissions, i.status, i.created_at, i.updated_at,
		       COALESCE(NULLIF(d.title_ru, ''), NULLIF(d.title_en, ''), 'Проект #' || i.project_id::text) AS project_title,
		       COALESCE(d.icon_path, '') AS project_icon
		FROM project_invitations i
		LEFT JOIN project_drafts d ON d.project_id = i.project_id
		WHERE i.invitee_id = $1 AND i.status = 1
		ORDER BY i.created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, inviteeID)
	if err != nil {
		return nil, fmt.Errorf("postgres.InvitationRepo.ListIncoming: %w", err)
	}
	defer rows.Close()

	var res []*domain.Invitation
	for rows.Next() {
		var inv domain.Invitation
		if err := rows.Scan(
			&inv.ID, &inv.ProjectID, &inv.InviterID, &inv.InviterEmail, &inv.InviterName,
			&inv.InviteeID, &inv.InviteeEmail, &inv.Permissions, &inv.Status, &inv.CreatedAt, &inv.UpdatedAt,
			&inv.ProjectTitle, &inv.ProjectIcon,
		); err != nil {
			return nil, fmt.Errorf("postgres.InvitationRepo.ListIncoming scan: %w", err)
		}
		res = append(res, &inv)
	}
	return res, nil
}

// ListOutgoing возвращает исходящие приглашения, отправленные пользователем.
func (r *InvitationRepo) ListOutgoing(ctx context.Context, inviterID string, projectID int64) ([]*domain.Invitation, error) {
	query := `
		SELECT i.id, i.project_id, i.inviter_id, i.inviter_email, i.inviter_name,
		       i.invitee_id, i.invitee_email, i.permissions, i.status, i.created_at, i.updated_at,
		       COALESCE(NULLIF(d.title_ru, ''), NULLIF(d.title_en, ''), 'Проект #' || i.project_id::text) AS project_title,
		       COALESCE(d.icon_path, '') AS project_icon
		FROM project_invitations i
		LEFT JOIN project_drafts d ON d.project_id = i.project_id
		WHERE i.inviter_id = $1 AND ($2 = 0 OR i.project_id = $2)
		ORDER BY i.created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, inviterID, projectID)
	if err != nil {
		return nil, fmt.Errorf("postgres.InvitationRepo.ListOutgoing: %w", err)
	}
	defer rows.Close()

	var res []*domain.Invitation
	for rows.Next() {
		var inv domain.Invitation
		if err := rows.Scan(
			&inv.ID, &inv.ProjectID, &inv.InviterID, &inv.InviterEmail, &inv.InviterName,
			&inv.InviteeID, &inv.InviteeEmail, &inv.Permissions, &inv.Status, &inv.CreatedAt, &inv.UpdatedAt,
			&inv.ProjectTitle, &inv.ProjectIcon,
		); err != nil {
			return nil, fmt.Errorf("postgres.InvitationRepo.ListOutgoing scan: %w", err)
		}
		res = append(res, &inv)
	}
	return res, nil
}

// UpdateStatus обновляет статус приглашения (accepted, declined, canceled).
func (r *InvitationRepo) UpdateStatus(ctx context.Context, id int64, status domain.InvitationStatus) error {
	const query = `UPDATE project_invitations SET status = $1, updated_at = NOW() WHERE id = $2`
	cmd, err := r.pool.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("postgres.InvitationRepo.UpdateStatus: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// CancelAllPendingBetween отменяет все висящие инвайты между отправителем и получателем (например при блокировке).
func (r *InvitationRepo) CancelAllPendingBetween(ctx context.Context, inviterID, inviteeID string) error {
	const query = `
		UPDATE project_invitations
		SET status = 4, updated_at = NOW()
		WHERE inviter_id = $1 AND invitee_id = $2 AND status = 1
	`
	_, err := r.pool.Exec(ctx, query, inviterID, inviteeID)
	if err != nil {
		return fmt.Errorf("postgres.InvitationRepo.CancelAllPendingBetween: %w", err)
	}
	return nil
}
