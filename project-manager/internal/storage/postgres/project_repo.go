package postgres

import (
	"context"
	"fmt"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProjectRepo реализует domain.ProjectRepo для работы с таблицей projects.
type ProjectRepo struct {
	pool *pgxpool.Pool
}

// NewProjectRepo создаёт новый экземпляр репозитория проектов.
func NewProjectRepo(pool *pgxpool.Pool) *ProjectRepo {
	return &ProjectRepo{pool: pool}
}

// Create создаёт новый проект в базе данных и возвращает его ID.
func (r *ProjectRepo) Create(ctx context.Context, p *domain.Project) (int64, error) {
	const query = `
		INSERT INTO projects (owner_id, status)
		VALUES ($1, $2)
		RETURNING id
	`
	var id int64
	err := r.pool.QueryRow(ctx, query, p.OwnerID, p.Status).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("postgres.ProjectRepo.Create: %w", err)
	}
	return id, nil
}

// Get загружает проект по его идентификатору. Возвращает ErrNotFound при отсутствии.
func (r *ProjectRepo) Get(ctx context.Context, id int64) (*domain.Project, error) {
	const query = `
		SELECT p.id, p.owner_id, p.status, p.created_at, p.updated_at,
		       d.title_ru, d.title_en, d.seo_ru, d.seo_en, d.about,
		       d.icon_path, d.cover_path, d.video_path, d.active_build_version,
		       d.dev_url, d.updated_at
		FROM projects p
		LEFT JOIN project_drafts d ON d.project_id = p.id
		WHERE p.id = $1
	`
	var (
		p                   domain.Project
		draft               domain.Draft
		titleRu, titleEn    *string
		seoRu, seoEn, about *string
		iconPath, coverPath *string
		videoPath, activeVer *string
		devURL              *string
		draftUpdatedAt      *context.Context // placeholder
	)
	_ = draftUpdatedAt

	row := r.pool.QueryRow(ctx, query, id)
	err := row.Scan(
		&p.ID, &p.OwnerID, &p.Status, &p.CreatedAt, &p.UpdatedAt,
		&titleRu, &titleEn, &seoRu, &seoEn, &about,
		&iconPath, &coverPath, &videoPath, &activeVer,
		&devURL, &draft.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("postgres.ProjectRepo.Get: %w", err)
	}

	draft.ProjectID = p.ID
	if titleRu != nil {
		draft.TitleRu = *titleRu
	}
	if titleEn != nil {
		draft.TitleEn = *titleEn
	}
	if seoRu != nil {
		draft.SeoRu = *seoRu
	}
	if seoEn != nil {
		draft.SeoEn = *seoEn
	}
	if about != nil {
		draft.About = *about
	}
	if iconPath != nil {
		draft.IconPath = *iconPath
	}
	if coverPath != nil {
		draft.CoverPath = *coverPath
	}
	if videoPath != nil {
		draft.VideoPath = *videoPath
	}
	if activeVer != nil {
		draft.ActiveBuildVersion = *activeVer
	}
	if devURL != nil {
		draft.DevURL = *devURL
	}
	p.Draft = &draft

	return &p, nil
}

// ListByOwner возвращает список проектов пользователя с пагинацией.
func (r *ProjectRepo) ListByOwner(ctx context.Context, ownerID string, limit, offset int) ([]*domain.Project, error) {
	const query = `
		SELECT p.id, p.owner_id, p.status, p.created_at, p.updated_at,
		       d.title_ru, d.title_en, d.seo_ru, d.seo_en, d.about,
		       d.icon_path, d.cover_path, d.video_path, d.active_build_version,
		       d.dev_url, d.updated_at
		FROM projects p
		LEFT JOIN project_drafts d ON d.project_id = p.id
		WHERE p.owner_id = $1
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, ownerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("postgres.ProjectRepo.ListByOwner: %w", err)
	}
	defer rows.Close()

	var projects []*domain.Project
	for rows.Next() {
		var (
			p                   domain.Project
			draft               domain.Draft
			titleRu, titleEn    *string
			seoRu, seoEn, about *string
			iconPath, coverPath *string
			videoPath, activeVer *string
			devURL              *string
		)
		if err := rows.Scan(
			&p.ID, &p.OwnerID, &p.Status, &p.CreatedAt, &p.UpdatedAt,
			&titleRu, &titleEn, &seoRu, &seoEn, &about,
			&iconPath, &coverPath, &videoPath, &activeVer,
			&devURL, &draft.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("postgres.ProjectRepo.ListByOwner scan: %w", err)
		}

		draft.ProjectID = p.ID
		if titleRu != nil {
			draft.TitleRu = *titleRu
		}
		if titleEn != nil {
			draft.TitleEn = *titleEn
		}
		if seoRu != nil {
			draft.SeoRu = *seoRu
		}
		if seoEn != nil {
			draft.SeoEn = *seoEn
		}
		if about != nil {
			draft.About = *about
		}
		if iconPath != nil {
			draft.IconPath = *iconPath
		}
		if coverPath != nil {
			draft.CoverPath = *coverPath
		}
		if videoPath != nil {
			draft.VideoPath = *videoPath
		}
		if activeVer != nil {
			draft.ActiveBuildVersion = *activeVer
		}
		if devURL != nil {
			draft.DevURL = *devURL
		}
		p.Draft = &draft

		projects = append(projects, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres.ProjectRepo.ListByOwner rows: %w", err)
	}

	return projects, nil
}

// CountByOwner возвращает общее количество проектов пользователя.
func (r *ProjectRepo) CountByOwner(ctx context.Context, ownerID string) (int, error) {
	const query = `SELECT COUNT(*) FROM projects WHERE owner_id = $1`
	var count int
	err := r.pool.QueryRow(ctx, query, ownerID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("postgres.ProjectRepo.CountByOwner: %w", err)
	}
	return count, nil
}

// UpdateStatus обновляет статус жизненного цикла проекта.
func (r *ProjectRepo) UpdateStatus(ctx context.Context, id int64, status domain.ProjectStatus) error {
	const query = `UPDATE projects SET status = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("postgres.ProjectRepo.UpdateStatus: %w", err)
	}
	return nil
}

// Delete удаляет проект и связанные каскадные записи.
func (r *ProjectRepo) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM projects WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("postgres.ProjectRepo.Delete: %w", err)
	}
	return nil
}
