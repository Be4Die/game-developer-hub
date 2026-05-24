package postgres

import (
	"context"
	"fmt"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProjectBuildRepo реализация domain.ProjectBuildRepo.
type ProjectBuildRepo struct {
	pool *pgxpool.Pool
}

// NewProjectBuildRepo создаёт репозиторий билдов.
func NewProjectBuildRepo(pool *pgxpool.Pool) *ProjectBuildRepo {
	return &ProjectBuildRepo{pool: pool}
}

// Create создаёт запись о билде.
func (r *ProjectBuildRepo) Create(ctx context.Context, b *domain.ProjectBuild) error {
	const query = `
		INSERT INTO project_builds (project_id, version, file_path, file_size)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.pool.Exec(ctx, query, b.ProjectID, b.Version, b.FilePath, b.FileSize)
	if err != nil {
		return fmt.Errorf("postgres.ProjectBuildRepo.Create: %w", err)
	}
	return nil
}

// ListByProject возвращает билды проекта, отсортированные по дате (новые сначала).
func (r *ProjectBuildRepo) ListByProject(ctx context.Context, projectID int64, limit int) ([]*domain.ProjectBuild, error) {
	const query = `
		SELECT id, project_id, version, file_path, file_size, created_at
		FROM project_builds
		WHERE project_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	rows, err := r.pool.Query(ctx, query, projectID, limit)
	if err != nil {
		return nil, fmt.Errorf("postgres.ProjectBuildRepo.ListByProject: %w", err)
	}
	defer rows.Close()

	var builds []*domain.ProjectBuild
	for rows.Next() {
		b, err := scanBuild(rows)
		if err != nil {
			return nil, fmt.Errorf("postgres.ProjectBuildRepo.ListByProject scan: %w", err)
		}
		builds = append(builds, b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres.ProjectBuildRepo.ListByProject rows: %w", err)
	}
	return builds, nil
}

// Get возвращает билд по project_id + version.
func (r *ProjectBuildRepo) Get(ctx context.Context, projectID int64, version string) (*domain.ProjectBuild, error) {
	const query = `
		SELECT id, project_id, version, file_path, file_size, created_at
		FROM project_builds
		WHERE project_id = $1 AND version = $2
	`
	row := r.pool.QueryRow(ctx, query, projectID, version)
	b, err := scanBuild(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("postgres.ProjectBuildRepo.Get: %w", err)
	}
	return b, nil
}

// Delete удаляет билд по ID.
func (r *ProjectBuildRepo) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM project_builds WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("postgres.ProjectBuildRepo.Delete: %w", err)
	}
	return nil
}

// DeleteOldest удаляет самые старые билды, оставляя `keep` последних.
func (r *ProjectBuildRepo) DeleteOldest(ctx context.Context, projectID int64, keep int) error {
	const query = `
		DELETE FROM project_builds
		WHERE id IN (
			SELECT id FROM project_builds
			WHERE project_id = $1
			ORDER BY created_at DESC
			OFFSET $2
		)
	`
	_, err := r.pool.Exec(ctx, query, projectID, keep)
	if err != nil {
		return fmt.Errorf("postgres.ProjectBuildRepo.DeleteOldest: %w", err)
	}
	return nil
}

type buildScanner interface {
	Scan(dest ...any) error
}

func scanBuild(s buildScanner) (*domain.ProjectBuild, error) {
	var b domain.ProjectBuild
	err := s.Scan(&b.ID, &b.ProjectID, &b.Version, &b.FilePath, &b.FileSize, &b.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &b, nil
}
