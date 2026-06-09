// Package postgres реализует репозитории проектов на PostgreSQL.
package postgres

import (
	"context"
	"fmt"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProjectRepo реализация domain.ProjectRepo.
type ProjectRepo struct {
	pool *pgxpool.Pool
}

// NewProjectRepo создаёт репозиторий проектов.
func NewProjectRepo(pool *pgxpool.Pool) *ProjectRepo {
	return &ProjectRepo{pool: pool}
}

// Create создаёт новый проект.
func (r *ProjectRepo) Create(ctx context.Context, p *domain.Project) (int64, error) {
	const query = `
		INSERT INTO projects (owner_id, title_ru, title_en, seo_ru, seo_en, about, status, icon_path, cover_path, video_path, active_build_version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`
	var id int64
	err := r.pool.QueryRow(ctx, query,
		p.OwnerID, p.TitleRu, p.TitleEn, p.SeoRu, p.SeoEn, p.About,
		p.Status, p.IconPath, p.CoverPath, p.VideoPath, p.ActiveBuildVersion,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("postgres.ProjectRepo.Create: %w", err)
	}
	return id, nil
}

// Get возвращает проект по ID.
func (r *ProjectRepo) Get(ctx context.Context, id int64) (*domain.Project, error) {
	const query = `
		SELECT id, owner_id, title_ru, title_en, seo_ru, seo_en, about, status,
		       icon_path, cover_path, video_path, active_build_version, created_at, updated_at
		FROM projects
		WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, query, id)
	p, err := scanProject(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("postgres.ProjectRepo.Get: %w", err)
	}
	return p, nil
}

// ListByOwner возвращает проекты пользователя.
func (r *ProjectRepo) ListByOwner(ctx context.Context, ownerID string, limit, offset int) ([]*domain.Project, error) {
	const query = `
		SELECT id, owner_id, title_ru, title_en, seo_ru, seo_en, about, status,
		       icon_path, cover_path, video_path, active_build_version, created_at, updated_at
		FROM projects
		WHERE owner_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, ownerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("postgres.ProjectRepo.ListByOwner: %w", err)
	}
	defer rows.Close()

	var projects []*domain.Project
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			return nil, fmt.Errorf("postgres.ProjectRepo.ListByOwner scan: %w", err)
		}
		projects = append(projects, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres.ProjectRepo.ListByOwner rows: %w", err)
	}
	return projects, nil
}

// Update обновляет проект.
func (r *ProjectRepo) Update(ctx context.Context, p *domain.Project) error {
	const query = `
		UPDATE projects
		SET title_ru = $1, title_en = $2, seo_ru = $3, seo_en = $4, about = $5,
		    status = $6, icon_path = $7, cover_path = $8, video_path = $9, active_build_version = $10
		WHERE id = $11
	`
	_, err := r.pool.Exec(ctx, query,
		p.TitleRu, p.TitleEn, p.SeoRu, p.SeoEn, p.About,
		p.Status, p.IconPath, p.CoverPath, p.VideoPath, p.ActiveBuildVersion, p.ID,
	)
	if err != nil {
		return fmt.Errorf("postgres.ProjectRepo.Update: %w", err)
	}
	return nil
}

// Delete удаляет проект.
func (r *ProjectRepo) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM projects WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("postgres.ProjectRepo.Delete: %w", err)
	}
	return nil
}

// scanner-интерфейс для переиспользования.
type projectScanner interface {
	Scan(dest ...any) error
}

// InitSchema создаёт таблицы project-manager, если их ещё нет (для существующих volume Postgres).
func InitSchema(ctx context.Context, pool *pgxpool.Pool) error {
	schema := `
		CREATE TABLE IF NOT EXISTS projects (
			id          BIGSERIAL PRIMARY KEY,
			owner_id    TEXT NOT NULL,
			title_ru    TEXT NOT NULL DEFAULT '',
			title_en    TEXT NOT NULL DEFAULT '',
			seo_ru      TEXT NOT NULL DEFAULT '',
			seo_en      TEXT NOT NULL DEFAULT '',
			about       TEXT NOT NULL DEFAULT '',
			status      SMALLINT NOT NULL DEFAULT 1,
			icon_path   TEXT NOT NULL DEFAULT '',
			cover_path  TEXT NOT NULL DEFAULT '',
			video_path  TEXT NOT NULL DEFAULT '',
			active_build_version TEXT NOT NULL DEFAULT '',
			created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at  TIMESTAMP NOT NULL DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_projects_owner ON projects(owner_id);
		CREATE INDEX IF NOT EXISTS idx_projects_status ON projects(status);

		CREATE TABLE IF NOT EXISTS project_builds (
			id          BIGSERIAL PRIMARY KEY,
			project_id  BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			version     TEXT NOT NULL,
			file_path   TEXT NOT NULL,
			file_size   BIGINT NOT NULL,
			created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
			UNIQUE (project_id, version)
		);
		CREATE INDEX IF NOT EXISTS idx_project_builds_project ON project_builds(project_id, created_at DESC);
	`
	_, err := pool.Exec(ctx, schema)
	return err
}

func scanProject(s projectScanner) (*domain.Project, error) {
	var p domain.Project
	err := s.Scan(
		&p.ID, &p.OwnerID, &p.TitleRu, &p.TitleEn, &p.SeoRu, &p.SeoEn, &p.About, &p.Status,
		&p.IconPath, &p.CoverPath, &p.VideoPath, &p.ActiveBuildVersion, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}
