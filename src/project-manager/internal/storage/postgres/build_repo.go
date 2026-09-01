package postgres

import (
	"context"
	"fmt"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BuildRepo реализует domain.BuildRepo для работы с таблицей project_builds.
type BuildRepo struct {
	pool *pgxpool.Pool
}

// NewBuildRepo создаёт новый экземпляр репозитория сборок.
func NewBuildRepo(pool *pgxpool.Pool) *BuildRepo {
	return &BuildRepo{pool: pool}
}

// Create сохраняет запись о сборке в базе данных.
func (r *BuildRepo) Create(ctx context.Context, b *domain.Build) error {
	const query = `
		INSERT INTO project_builds (project_id, version, file_path, file_size, is_unpacked, unpacked_path)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`
	err := r.pool.QueryRow(ctx, query,
		b.ProjectID, b.Version, b.FilePath, b.FileSize, b.IsUnpacked, b.UnpackedPath,
	).Scan(&b.ID, &b.CreatedAt)
	if err != nil {
		return fmt.Errorf("postgres.BuildRepo.Create: %w", err)
	}
	return nil
}

// Get загружает информацию о конкретной версии сборки проекта.
func (r *BuildRepo) Get(ctx context.Context, projectID int64, version string) (*domain.Build, error) {
	const query = `
		SELECT id, project_id, version, file_path, file_size, is_unpacked, unpacked_path, created_at
		FROM project_builds
		WHERE project_id = $1 AND version = $2
	`
	var b domain.Build
	err := r.pool.QueryRow(ctx, query, projectID, version).Scan(
		&b.ID, &b.ProjectID, &b.Version, &b.FilePath, &b.FileSize, &b.IsUnpacked, &b.UnpackedPath, &b.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("postgres.BuildRepo.Get: %w", err)
	}
	return &b, nil
}

// ListByProject возвращает список сборок проекта, отсортированный от новых к старым.
func (r *BuildRepo) ListByProject(ctx context.Context, projectID int64, limit int) ([]*domain.Build, error) {
	if limit <= 0 {
		limit = 10
	}
	const query = `
		SELECT id, project_id, version, file_path, file_size, is_unpacked, unpacked_path, created_at
		FROM project_builds
		WHERE project_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	rows, err := r.pool.Query(ctx, query, projectID, limit)
	if err != nil {
		return nil, fmt.Errorf("postgres.BuildRepo.ListByProject: %w", err)
	}
	defer rows.Close()

	var builds []*domain.Build
	for rows.Next() {
		var b domain.Build
		if err := rows.Scan(
			&b.ID, &b.ProjectID, &b.Version, &b.FilePath, &b.FileSize, &b.IsUnpacked, &b.UnpackedPath, &b.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("postgres.BuildRepo.ListByProject scan: %w", err)
		}
		builds = append(builds, &b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres.BuildRepo.ListByProject rows: %w", err)
	}

	return builds, nil
}

// MarkUnpacked помечает сборку как успешно распакованную с указанием директории.
func (r *BuildRepo) MarkUnpacked(ctx context.Context, projectID int64, version, unpackedPath string) error {
	const query = `
		UPDATE project_builds
		SET is_unpacked = true, unpacked_path = $1
		WHERE project_id = $2 AND version = $3
	`
	_, err := r.pool.Exec(ctx, query, unpackedPath, projectID, version)
	if err != nil {
		return fmt.Errorf("postgres.BuildRepo.MarkUnpacked: %w", err)
	}
	return nil
}

// Delete удаляет запись о сборке по её ID.
func (r *BuildRepo) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM project_builds WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("postgres.BuildRepo.Delete: %w", err)
	}
	return nil
}
