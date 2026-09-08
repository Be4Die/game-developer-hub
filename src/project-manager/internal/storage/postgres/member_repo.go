package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MemberRepo реализует domain.MemberRepo для таблицы project_members.
type MemberRepo struct {
	pool *pgxpool.Pool
}

// NewMemberRepo создаёт репозиторий участников проектов.
func NewMemberRepo(pool *pgxpool.Pool) *MemberRepo {
	return &MemberRepo{pool: pool}
}

// Add добавляет участника в проект или обновляет права при повторном добавлении.
func (r *MemberRepo) Add(ctx context.Context, m *domain.Member) error {
	const query = `
		INSERT INTO project_members (project_id, user_id, user_email, user_name, permissions)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (project_id, user_id) DO UPDATE
		SET permissions = EXCLUDED.permissions,
		    user_email = CASE WHEN EXCLUDED.user_email <> '' THEN EXCLUDED.user_email ELSE project_members.user_email END,
		    user_name = CASE WHEN EXCLUDED.user_name <> '' THEN EXCLUDED.user_name ELSE project_members.user_name END,
		    updated_at = NOW()
		RETURNING id, created_at, updated_at
	`
	err := r.pool.QueryRow(ctx, query, m.ProjectID, m.UserID, m.UserEmail, m.UserName, m.Permissions).
		Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return fmt.Errorf("postgres.MemberRepo.Add: %w", err)
	}
	return nil
}

// Get возвращает участника проекта по projectID и userID.
func (r *MemberRepo) Get(ctx context.Context, projectID int64, userID string) (*domain.Member, error) {
	const query = `
		SELECT id, project_id, user_id, user_email, user_name, permissions, created_at, updated_at
		FROM project_members
		WHERE project_id = $1 AND user_id = $2
	`
	var m domain.Member
	err := r.pool.QueryRow(ctx, query, projectID, userID).Scan(
		&m.ID, &m.ProjectID, &m.UserID, &m.UserEmail, &m.UserName, &m.Permissions, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("postgres.MemberRepo.Get: %w", err)
	}
	return &m, nil
}

// ListByProject возвращает всех участников указанного проекта.
func (r *MemberRepo) ListByProject(ctx context.Context, projectID int64) ([]*domain.Member, error) {
	const query = `
		SELECT id, project_id, user_id, user_email, user_name, permissions, created_at, updated_at
		FROM project_members
		WHERE project_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("postgres.MemberRepo.ListByProject: %w", err)
	}
	defer rows.Close()

	var members []*domain.Member
	for rows.Next() {
		var m domain.Member
		if err := rows.Scan(
			&m.ID, &m.ProjectID, &m.UserID, &m.UserEmail, &m.UserName, &m.Permissions, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("postgres.MemberRepo.ListByProject scan: %w", err)
		}
		members = append(members, &m)
	}
	return members, nil
}

// ListByUser возвращает записи участия для конкретного пользователя.
func (r *MemberRepo) ListByUser(ctx context.Context, userID string) ([]*domain.Member, error) {
	const query = `
		SELECT id, project_id, user_id, user_email, user_name, permissions, created_at, updated_at
		FROM project_members
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("postgres.MemberRepo.ListByUser: %w", err)
	}
	defer rows.Close()

	var members []*domain.Member
	for rows.Next() {
		var m domain.Member
		if err := rows.Scan(
			&m.ID, &m.ProjectID, &m.UserID, &m.UserEmail, &m.UserName, &m.Permissions, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("postgres.MemberRepo.ListByUser scan: %w", err)
		}
		members = append(members, &m)
	}
	return members, nil
}

// UpdatePermissions обновляет права участника проекта.
func (r *MemberRepo) UpdatePermissions(ctx context.Context, projectID int64, userID string, permissions []string) error {
	const query = `
		UPDATE project_members
		SET permissions = $1, updated_at = NOW()
		WHERE project_id = $2 AND user_id = $3
	`
	cmd, err := r.pool.Exec(ctx, query, permissions, projectID, userID)
	if err != nil {
		return fmt.Errorf("postgres.MemberRepo.UpdatePermissions: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// Delete удаляет участника из проекта.
func (r *MemberRepo) Delete(ctx context.Context, projectID int64, userID string) error {
	const query = `DELETE FROM project_members WHERE project_id = $1 AND user_id = $2`
	cmd, err := r.pool.Exec(ctx, query, projectID, userID)
	if err != nil {
		return fmt.Errorf("postgres.MemberRepo.Delete: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// IsMember проверяет, является ли пользователь участником проекта.
func (r *MemberRepo) IsMember(ctx context.Context, projectID int64, userID string) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM project_members WHERE project_id = $1 AND user_id = $2)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, projectID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("postgres.MemberRepo.IsMember: %w", err)
	}
	return exists, nil
}
