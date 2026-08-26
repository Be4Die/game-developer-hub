package postgres

import (
	"context"
	"fmt"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DraftRepo реализует domain.DraftRepo для работы с таблицей project_drafts.
type DraftRepo struct {
	pool *pgxpool.Pool
}

// NewDraftRepo создаёт новый экземпляр репозитория черновиков.
func NewDraftRepo(pool *pgxpool.Pool) *DraftRepo {
	return &DraftRepo{pool: pool}
}

// Create сохраняет новую запись черновика в базе данных.
func (r *DraftRepo) Create(ctx context.Context, d *domain.Draft) error {
	const query = `
		INSERT INTO project_drafts (project_id, title_ru, title_en, seo_ru, seo_en, about_ru, about_en, icon_path, cover_path, video_path, active_build_version, dev_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.pool.Exec(ctx, query,
		d.ProjectID, d.TitleRu, d.TitleEn, d.SeoRu, d.SeoEn, d.AboutRu, d.AboutEn,
		d.IconPath, d.CoverPath, d.VideoPath, d.ActiveBuildVersion, d.DevURL,
	)
	if err != nil {
		return fmt.Errorf("postgres.DraftRepo.Create: %w", err)
	}
	return nil
}

// Get загружает черновик проекта по ID проекта. Возвращает ErrNotFound при отсутствии.
func (r *DraftRepo) Get(ctx context.Context, projectID int64) (*domain.Draft, error) {
	const query = `
		SELECT project_id, title_ru, title_en, seo_ru, seo_en, about_ru, about_en,
		       icon_path, cover_path, video_path, active_build_version, dev_url, updated_at
		FROM project_drafts
		WHERE project_id = $1
	`
	var d domain.Draft
	err := r.pool.QueryRow(ctx, query, projectID).Scan(
		&d.ProjectID, &d.TitleRu, &d.TitleEn, &d.SeoRu, &d.SeoEn, &d.AboutRu, &d.AboutEn,
		&d.IconPath, &d.CoverPath, &d.VideoPath, &d.ActiveBuildVersion, &d.DevURL, &d.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("postgres.DraftRepo.Get: %w", err)
	}
	return &d, nil
}

// Update обновляет метаданные черновика.
func (r *DraftRepo) Update(ctx context.Context, d *domain.Draft) error {
	const query = `
		UPDATE project_drafts
		SET title_ru = $1, title_en = $2, seo_ru = $3, seo_en = $4, about_ru = $5, about_en = $6, active_build_version = $7
		WHERE project_id = $8
	`
	_, err := r.pool.Exec(ctx, query,
		d.TitleRu, d.TitleEn, d.SeoRu, d.SeoEn, d.AboutRu, d.AboutEn, d.ActiveBuildVersion, d.ProjectID,
	)
	if err != nil {
		return fmt.Errorf("postgres.DraftRepo.Update: %w", err)
	}
	return nil
}

// UpdateActiveBuild обновляет активную версию сборки и dev-ссылку черновика.
func (r *DraftRepo) UpdateActiveBuild(ctx context.Context, projectID int64, version, devURL string) error {
	const query = `
		UPDATE project_drafts
		SET active_build_version = $1, dev_url = $2
		WHERE project_id = $3
	`
	_, err := r.pool.Exec(ctx, query, version, devURL, projectID)
	if err != nil {
		return fmt.Errorf("postgres.DraftRepo.UpdateActiveBuild: %w", err)
	}
	return nil
}

// UpdateMedia обновляет путь к медиафайлу в черновике.
func (r *DraftRepo) UpdateMedia(ctx context.Context, projectID int64, mediaType, path string) error {
	var query string
	switch mediaType {
	case "icon":
		query = `UPDATE project_drafts SET icon_path = $1 WHERE project_id = $2`
	case "cover":
		query = `UPDATE project_drafts SET cover_path = $1 WHERE project_id = $2`
	case "video":
		query = `UPDATE project_drafts SET video_path = $1 WHERE project_id = $2`
	default:
		return domain.ErrInvalidInput
	}

	_, err := r.pool.Exec(ctx, query, path, projectID)
	if err != nil {
		return fmt.Errorf("postgres.DraftRepo.UpdateMedia: %w", err)
	}
	return nil
}
