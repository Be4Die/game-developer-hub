-- init.sql — инициализация схемы БД project-manager.

-- ─────────────────────────────────────────────────────────────────────────────
-- Таблица projects — проекты игр разработчиков
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS projects (
    id          BIGSERIAL PRIMARY KEY,
    owner_id    TEXT NOT NULL,                    -- UUID пользователя из SSO
    title_ru    TEXT NOT NULL DEFAULT '',
    title_en    TEXT NOT NULL DEFAULT '',
    seo_ru      TEXT NOT NULL DEFAULT '',
    seo_en      TEXT NOT NULL DEFAULT '',
    about       TEXT NOT NULL DEFAULT '',
    status      SMALLINT NOT NULL DEFAULT 1,     -- 1=draft, 2=pending, 3=published, 4=rejected
    icon_path   TEXT NOT NULL DEFAULT '',
    cover_path  TEXT NOT NULL DEFAULT '',
    video_path            TEXT NOT NULL DEFAULT '',
    active_build_version  TEXT NOT NULL DEFAULT '',
    created_at            TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE projects IS 'Проекты игр (черновики и опубликованные)';
COMMENT ON COLUMN projects.status IS '1=draft, 2=pending, 3=published, 4=rejected';

CREATE INDEX IF NOT EXISTS idx_projects_owner ON projects(owner_id);
CREATE INDEX IF NOT EXISTS idx_projects_status ON projects(status);

-- ─────────────────────────────────────────────────────────────────────────────
-- Таблица project_builds — клиентские билды проектов
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS project_builds (
    id          BIGSERIAL PRIMARY KEY,
    project_id  BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    version     TEXT NOT NULL,
    file_path   TEXT NOT NULL,
    file_size   BIGINT NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, version)
);

COMMENT ON TABLE project_builds IS 'Клиентские билды проектов (макс 5 на проект)';

CREATE INDEX IF NOT EXISTS idx_project_builds_project ON project_builds(project_id, created_at DESC);

-- ─────────────────────────────────────────────────────────────────────────────
-- Функция для обновления updated_at
-- ─────────────────────────────────────────────────────────────────────────────

CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_projects_updated_at
    BEFORE UPDATE ON projects
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
