package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ReleaseRepo реализует domain.ReleaseRepo для работы с таблицей project_releases.
type ReleaseRepo struct {
	pool *pgxpool.Pool
}

// NewReleaseRepo создаёт новый экземпляр репозитория релизов.
func NewReleaseRepo(pool *pgxpool.Pool) *ReleaseRepo {
	return &ReleaseRepo{pool: pool}
}

// Create сохраняет запись о новом опубликованном релизе.
func (r *ReleaseRepo) Create(ctx context.Context, rel *domain.Release) (int64, error) {
	// Сначала деактивируем все предыдущие активные релизы проекта
	const deactivateQuery = `UPDATE project_releases SET is_active = false, unpublish_at = NOW() WHERE project_id = $1 AND is_active = true`
	if _, err := r.pool.Exec(ctx, deactivateQuery, rel.ProjectID); err != nil {
		return 0, fmt.Errorf("postgres.ReleaseRepo.Create deactivate old: %w", err)
	}

	const query = `
		INSERT INTO project_releases (
			project_id, version, title_ru, title_en, seo_ru, seo_en,
			about, icon_path, cover_path, video_path, prod_url, is_active, published_by
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, true, $12)
		RETURNING id, published_at
	`
	var id int64
	err := r.pool.QueryRow(ctx, query,
		rel.ProjectID, rel.Version, rel.TitleRu, rel.TitleEn, rel.SeoRu, rel.SeoEn,
		rel.About, rel.IconPath, rel.CoverPath, rel.VideoPath, rel.ProdURL, rel.PublishedBy,
	).Scan(&id, &rel.PublishedAt)
	if err != nil {
		return 0, fmt.Errorf("postgres.ReleaseRepo.Create: %w", err)
	}
	rel.ID = id
	rel.IsActive = true
	return id, nil
}

// GetActive загружает текущий активный релиз проекта.
func (r *ReleaseRepo) GetActive(ctx context.Context, projectID int64) (*domain.Release, error) {
	const query = `
		SELECT id, project_id, version, title_ru, title_en, seo_ru, seo_en,
		       about, icon_path, cover_path, video_path, prod_url, is_active,
		       published_by, published_at, unpublish_at
		FROM project_releases
		WHERE project_id = $1 AND is_active = true
		LIMIT 1
	`
	var rel domain.Release
	err := r.pool.QueryRow(ctx, query, projectID).Scan(
		&rel.ID, &rel.ProjectID, &rel.Version, &rel.TitleRu, &rel.TitleEn, &rel.SeoRu, &rel.SeoEn,
		&rel.About, &rel.IconPath, &rel.CoverPath, &rel.VideoPath, &rel.ProdURL, &rel.IsActive,
		&rel.PublishedBy, &rel.PublishedAt, &rel.UnpublishAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("postgres.ReleaseRepo.GetActive: %w", err)
	}
	return &rel, nil
}

// ListByProject возвращает историю всех релизов проекта.
func (r *ReleaseRepo) ListByProject(ctx context.Context, projectID int64) ([]*domain.Release, error) {
	const query = `
		SELECT id, project_id, version, title_ru, title_en, seo_ru, seo_en,
		       about, icon_path, cover_path, video_path, prod_url, is_active,
		       published_by, published_at, unpublish_at
		FROM project_releases
		WHERE project_id = $1
		ORDER BY published_at DESC
	`
	rows, err := r.pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("postgres.ReleaseRepo.ListByProject: %w", err)
	}
	defer rows.Close()

	var releases []*domain.Release
	for rows.Next() {
		var rel domain.Release
		if err := rows.Scan(
			&rel.ID, &rel.ProjectID, &rel.Version, &rel.TitleRu, &rel.TitleEn, &rel.SeoRu, &rel.SeoEn,
			&rel.About, &rel.IconPath, &rel.CoverPath, &rel.VideoPath, &rel.ProdURL, &rel.IsActive,
			&rel.PublishedBy, &rel.PublishedAt, &rel.UnpublishAt,
		); err != nil {
			return nil, fmt.Errorf("postgres.ReleaseRepo.ListByProject scan: %w", err)
		}
		releases = append(releases, &rel)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres.ReleaseRepo.ListByProject rows: %w", err)
	}
	return releases, nil
}

// Deactivate снимает текущий активный релиз с публикации.
func (r *ReleaseRepo) Deactivate(ctx context.Context, projectID int64) error {
	const query = `
		UPDATE project_releases
		SET is_active = false, unpublish_at = $1
		WHERE project_id = $2 AND is_active = true
	`
	now := time.Now()
	_, err := r.pool.Exec(ctx, query, now, projectID)
	if err != nil {
		return fmt.Errorf("postgres.ReleaseRepo.Deactivate: %w", err)
	}
	return nil
}
