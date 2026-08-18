package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// InitSchema инициализирует таблицы и индексы project-manager в базе данных.
func InitSchema(ctx context.Context, pool *pgxpool.Pool) error {
	const schema = `
		CREATE TABLE IF NOT EXISTS projects (
			id          BIGSERIAL PRIMARY KEY,
			owner_id    TEXT NOT NULL,
			status      SMALLINT NOT NULL DEFAULT 1,
			created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_projects_owner ON projects(owner_id);
		CREATE INDEX IF NOT EXISTS idx_projects_status ON projects(status);

		CREATE TABLE IF NOT EXISTS project_drafts (
			project_id            BIGINT PRIMARY KEY REFERENCES projects(id) ON DELETE CASCADE,
			title_ru              TEXT NOT NULL DEFAULT '',
			title_en              TEXT NOT NULL DEFAULT '',
			seo_ru                TEXT NOT NULL DEFAULT '',
			seo_en                TEXT NOT NULL DEFAULT '',
			about                 TEXT NOT NULL DEFAULT '',
			icon_path             TEXT NOT NULL DEFAULT '',
			cover_path            TEXT NOT NULL DEFAULT '',
			video_path            TEXT NOT NULL DEFAULT '',
			active_build_version  TEXT NOT NULL DEFAULT '',
			dev_url               TEXT NOT NULL DEFAULT '',
			updated_at            TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS project_builds (
			id            BIGSERIAL PRIMARY KEY,
			project_id    BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			version       TEXT NOT NULL,
			file_path     TEXT NOT NULL,
			file_size     BIGINT NOT NULL DEFAULT 0,
			is_unpacked   BOOLEAN NOT NULL DEFAULT FALSE,
			unpacked_path TEXT NOT NULL DEFAULT '',
			created_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			UNIQUE (project_id, version)
		);
		CREATE INDEX IF NOT EXISTS idx_project_builds_project ON project_builds(project_id, created_at DESC);

		CREATE TABLE IF NOT EXISTS moderation_tickets (
			id                    BIGSERIAL PRIMARY KEY,
			project_id            BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			owner_id              TEXT NOT NULL,
			game_title            TEXT NOT NULL DEFAULT '',
			game_description      TEXT NOT NULL DEFAULT '',
			status                SMALLINT NOT NULL DEFAULT 1,
			snapshot_meta         JSONB NOT NULL DEFAULT '{}'::jsonb,
			rejection_reason      TEXT NOT NULL DEFAULT '',
			moderator_id          TEXT NOT NULL DEFAULT '',
			dev_url               TEXT NOT NULL DEFAULT '',
			active_build_version  TEXT NOT NULL DEFAULT '',
			submitted_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			resolved_at           TIMESTAMP WITH TIME ZONE
		);
		CREATE INDEX IF NOT EXISTS idx_moderation_tickets_status ON moderation_tickets(status, submitted_at DESC);
		CREATE INDEX IF NOT EXISTS idx_moderation_tickets_project ON moderation_tickets(project_id);

		CREATE TABLE IF NOT EXISTS project_releases (
			id                    BIGSERIAL PRIMARY KEY,
			project_id            BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			version               TEXT NOT NULL,
			title_ru              TEXT NOT NULL DEFAULT '',
			title_en              TEXT NOT NULL DEFAULT '',
			seo_ru                TEXT NOT NULL DEFAULT '',
			seo_en                TEXT NOT NULL DEFAULT '',
			about                 TEXT NOT NULL DEFAULT '',
			icon_path             TEXT NOT NULL DEFAULT '',
			cover_path            TEXT NOT NULL DEFAULT '',
			video_path            TEXT NOT NULL DEFAULT '',
			prod_url              TEXT NOT NULL DEFAULT '',
			is_active             BOOLEAN NOT NULL DEFAULT TRUE,
			published_by          TEXT NOT NULL DEFAULT '',
			published_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			unpublish_at          TIMESTAMP WITH TIME ZONE
		);
		CREATE INDEX IF NOT EXISTS idx_project_releases_project ON project_releases(project_id, is_active);

		CREATE TABLE IF NOT EXISTS deployments (
			id            BIGSERIAL PRIMARY KEY,
			project_id    BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			environment   SMALLINT NOT NULL DEFAULT 1,
			version       TEXT NOT NULL DEFAULT '',
			status        SMALLINT NOT NULL DEFAULT 1,
			error_message TEXT NOT NULL DEFAULT '',
			deployed_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_deployments_project ON deployments(project_id, deployed_at DESC);
	`
	_, err := pool.Exec(ctx, schema)
	return err
}
