package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

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
		INSERT INTO projects (owner_id, status, is_online)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	var id int64
	err := r.pool.QueryRow(ctx, query, p.OwnerID, p.Status, p.IsOnline).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("postgres.ProjectRepo.Create: %w", err)
	}
	return id, nil
}

// Get загружает проект по его идентификатору. Возвращает ErrNotFound при отсутствии.
func (r *ProjectRepo) Get(ctx context.Context, id int64) (*domain.Project, error) {
	const query = `
		SELECT p.id, p.owner_id, p.status, p.is_online, p.created_at, p.updated_at,
		       d.title_ru, d.title_en, d.seo_ru, d.seo_en, d.about_ru, d.about_en,
		       d.icon_path, d.cover_path, d.video_path, d.active_build_version,
		       d.dev_url, d.is_online, d.updated_at
		FROM projects p
		LEFT JOIN project_drafts d ON d.project_id = p.id
		WHERE p.id = $1
	`
	var (
		p                    domain.Project
		draft                domain.Draft
		titleRu, titleEn     *string
		seoRu, seoEn         *string
		aboutRu, aboutEn     *string
		iconPath, coverPath  *string
		videoPath, activeVer *string
		devURL               *string
		draftIsOnline        *bool
		draftUpdatedAt       *time.Time
	)

	row := r.pool.QueryRow(ctx, query, id)
	err := row.Scan(
		&p.ID, &p.OwnerID, &p.Status, &p.IsOnline, &p.CreatedAt, &p.UpdatedAt,
		&titleRu, &titleEn, &seoRu, &seoEn, &aboutRu, &aboutEn,
		&iconPath, &coverPath, &videoPath, &activeVer,
		&devURL, &draftIsOnline, &draftUpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("postgres.ProjectRepo.Get: %w", err)
	}

	if draftUpdatedAt != nil || titleRu != nil {
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
		if aboutRu != nil {
			draft.AboutRu = *aboutRu
		}
		if aboutEn != nil {
			draft.AboutEn = *aboutEn
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
		if draftIsOnline != nil {
			draft.IsOnline = *draftIsOnline
		}
		if draftUpdatedAt != nil {
			draft.UpdatedAt = *draftUpdatedAt
		}
		p.Draft = &draft
	}

	return &p, nil
}

// ListByOwner возвращает список проектов пользователя с пагинацией.
func (r *ProjectRepo) ListByOwner(ctx context.Context, ownerID string, limit, offset int) ([]*domain.Project, error) {
	const query = `
		SELECT p.id, p.owner_id, p.status, p.is_online, p.created_at, p.updated_at,
		       d.title_ru, d.title_en, d.seo_ru, d.seo_en, d.about_ru, d.about_en,
		       d.icon_path, d.cover_path, d.video_path, d.active_build_version,
		       d.dev_url, d.is_online, d.updated_at
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
			p                    domain.Project
			draft                domain.Draft
			titleRu, titleEn     *string
			seoRu, seoEn         *string
			aboutRu, aboutEn     *string
			iconPath, coverPath  *string
			videoPath, activeVer *string
			devURL               *string
			draftIsOnline        *bool
			draftUpdatedAt       *time.Time
		)
		if err := rows.Scan(
			&p.ID, &p.OwnerID, &p.Status, &p.IsOnline, &p.CreatedAt, &p.UpdatedAt,
			&titleRu, &titleEn, &seoRu, &seoEn, &aboutRu, &aboutEn,
			&iconPath, &coverPath, &videoPath, &activeVer,
			&devURL, &draftIsOnline, &draftUpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("postgres.ProjectRepo.ListByOwner scan: %w", err)
		}

		if draftUpdatedAt != nil || titleRu != nil {
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
			if aboutRu != nil {
				draft.AboutRu = *aboutRu
			}
			if aboutEn != nil {
				draft.AboutEn = *aboutEn
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
			if draftIsOnline != nil {
				draft.IsOnline = *draftIsOnline
			}
			if draftUpdatedAt != nil {
				draft.UpdatedAt = *draftUpdatedAt
			}
			p.Draft = &draft
		}

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

// ListForUser возвращает список проектов, к которым пользователь имеет доступ (владелец или участник).
func (r *ProjectRepo) ListForUser(ctx context.Context, userID string, limit, offset int) ([]*domain.Project, error) {
	const query = `
		SELECT p.id, p.owner_id, p.status, p.is_online, p.created_at, p.updated_at,
		       d.title_ru, d.title_en, d.seo_ru, d.seo_en, d.about_ru, d.about_en,
		       d.icon_path, d.cover_path, d.video_path, d.active_build_version,
		       d.dev_url, d.is_online, d.updated_at,
		       COALESCE(pm.permissions, '{}')
		FROM projects p
		LEFT JOIN project_drafts d ON d.project_id = p.id
		LEFT JOIN project_members pm ON pm.project_id = p.id AND pm.user_id = $1
		WHERE p.owner_id = $1 OR pm.user_id = $1
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("postgres.ProjectRepo.ListForUser: %w", err)
	}
	defer rows.Close()

	var projects []*domain.Project
	for rows.Next() {
		var (
			p                    domain.Project
			draft                domain.Draft
			titleRu, titleEn     *string
			seoRu, seoEn         *string
			aboutRu, aboutEn     *string
			iconPath, coverPath  *string
			videoPath, activeVer *string
			devURL               *string
			draftIsOnline        *bool
			draftUpdatedAt       *time.Time
			memberPerms          []string
		)
		if err := rows.Scan(
			&p.ID, &p.OwnerID, &p.Status, &p.IsOnline, &p.CreatedAt, &p.UpdatedAt,
			&titleRu, &titleEn, &seoRu, &seoEn, &aboutRu, &aboutEn,
			&iconPath, &coverPath, &videoPath, &activeVer,
			&devURL, &draftIsOnline, &draftUpdatedAt,
			&memberPerms,
		); err != nil {
			return nil, fmt.Errorf("postgres.ProjectRepo.ListForUser scan: %w", err)
		}

		if draftUpdatedAt != nil || titleRu != nil {
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
			if aboutRu != nil {
				draft.AboutRu = *aboutRu
			}
			if aboutEn != nil {
				draft.AboutEn = *aboutEn
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
			if draftIsOnline != nil {
				draft.IsOnline = *draftIsOnline
			}
			if draftUpdatedAt != nil {
				draft.UpdatedAt = *draftUpdatedAt
			}
			p.Draft = &draft
		}

		p.IsOwner = (p.OwnerID == userID)
		if p.IsOwner {
			p.CurrentUserPermissions = domain.AllPermissions()
		} else {
			p.CurrentUserPermissions = memberPerms
		}

		projects = append(projects, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres.ProjectRepo.ListForUser rows: %w", err)
	}

	return projects, nil
}

// CountForUser возвращает общее количество доступных пользователю проектов.
func (r *ProjectRepo) CountForUser(ctx context.Context, userID string) (int, error) {
	const query = `
		SELECT COUNT(DISTINCT p.id)
		FROM projects p
		LEFT JOIN project_members pm ON pm.project_id = p.id AND pm.user_id = $1
		WHERE p.owner_id = $1 OR pm.user_id = $1
	`
	var count int
	err := r.pool.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("postgres.ProjectRepo.CountForUser: %w", err)
	}
	return count, nil
}

// ListPublished возвращает список опубликованных проектов с пагинацией.
func (r *ProjectRepo) ListPublished(ctx context.Context, limit, offset int) ([]*domain.Project, error) {
	const query = `
		SELECT p.id, p.owner_id, p.status, p.is_online, p.created_at, p.updated_at,
		       d.title_ru, d.title_en, d.seo_ru, d.seo_en, d.about_ru, d.about_en,
		       d.icon_path, d.cover_path, d.video_path, d.active_build_version,
		       d.dev_url, d.is_online, d.updated_at
		FROM projects p
		LEFT JOIN project_drafts d ON d.project_id = p.id
		WHERE p.status = 3
		ORDER BY p.updated_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("postgres.ProjectRepo.ListPublished: %w", err)
	}
	defer rows.Close()

	var projects []*domain.Project
	for rows.Next() {
		var (
			p                    domain.Project
			draft                domain.Draft
			titleRu, titleEn     *string
			seoRu, seoEn         *string
			aboutRu, aboutEn     *string
			iconPath, coverPath  *string
			videoPath, activeVer *string
			devURL               *string
			draftIsOnline        *bool
			draftUpdatedAt       *time.Time
		)
		if err := rows.Scan(
			&p.ID, &p.OwnerID, &p.Status, &p.IsOnline, &p.CreatedAt, &p.UpdatedAt,
			&titleRu, &titleEn, &seoRu, &seoEn, &aboutRu, &aboutEn,
			&iconPath, &coverPath, &videoPath, &activeVer,
			&devURL, &draftIsOnline, &draftUpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("postgres.ProjectRepo.ListPublished scan: %w", err)
		}

		if draftUpdatedAt != nil || titleRu != nil {
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
			if aboutRu != nil {
				draft.AboutRu = *aboutRu
			}
			if aboutEn != nil {
				draft.AboutEn = *aboutEn
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
			if draftIsOnline != nil {
				draft.IsOnline = *draftIsOnline
			}
			if draftUpdatedAt != nil {
				draft.UpdatedAt = *draftUpdatedAt
			}
			p.Draft = &draft
		}

		projects = append(projects, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres.ProjectRepo.ListPublished rows: %w", err)
	}

	return projects, nil
}

// CountPublished возвращает общее количество опубликованных проектов.
func (r *ProjectRepo) CountPublished(ctx context.Context) (int, error) {
	const query = `SELECT COUNT(*) FROM projects WHERE status = 3`
	var count int
	err := r.pool.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("postgres.ProjectRepo.CountPublished: %w", err)
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

// UpdateIsOnline обновляет признак онлайн-игры базовой сущности проекта.
func (r *ProjectRepo) UpdateIsOnline(ctx context.Context, id int64, isOnline bool) error {
	const query = `UPDATE projects SET is_online = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, isOnline, id)
	if err != nil {
		return fmt.Errorf("postgres.ProjectRepo.UpdateIsOnline: %w", err)
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
