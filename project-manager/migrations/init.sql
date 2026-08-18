-- init.sql — инициализация схемы БД project-manager.

-- ─────────────────────────────────────────────────────────────────────────────
-- Таблица projects — базовые сущности игровых проектов
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS projects (
    id          BIGSERIAL PRIMARY KEY,
    owner_id    TEXT NOT NULL,                                       -- ID владельца (пользователь SSO)
    status      SMALLINT NOT NULL DEFAULT 1,                        -- 1=draft, 2=pending, 3=published, 4=rejected
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE projects IS 'Проекты игр разработчиков';
COMMENT ON COLUMN projects.status IS '1=draft, 2=pending, 3=published, 4=rejected';

CREATE INDEX IF NOT EXISTS idx_projects_owner ON projects(owner_id);
CREATE INDEX IF NOT EXISTS idx_projects_status ON projects(status);

-- ─────────────────────────────────────────────────────────────────────────────
-- Таблица project_drafts — рабочее состояние черновика проекта
-- ─────────────────────────────────────────────────────────────────────────────

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

COMMENT ON TABLE project_drafts IS 'Рабочие изменяемые черновики проектов';

-- ─────────────────────────────────────────────────────────────────────────────
-- Таблица project_builds — клиентские билды проектов
-- ─────────────────────────────────────────────────────────────────────────────

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

COMMENT ON TABLE project_builds IS 'Клиентские билды проектов (хранятся до N последних версий)';

CREATE INDEX IF NOT EXISTS idx_project_builds_project ON project_builds(project_id, created_at DESC);

-- ─────────────────────────────────────────────────────────────────────────────
-- Таблица moderation_tickets — история и очередь заявок на модерацию
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS moderation_tickets (
    id                    BIGSERIAL PRIMARY KEY,
    project_id            BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    owner_id              TEXT NOT NULL,
    game_title            TEXT NOT NULL DEFAULT '',
    game_description      TEXT NOT NULL DEFAULT '',
    status                SMALLINT NOT NULL DEFAULT 1,  -- 1=pending, 2=approved, 3=rejected
    snapshot_meta         JSONB NOT NULL DEFAULT '{}'::jsonb,
    rejection_reason      TEXT NOT NULL DEFAULT '',
    moderator_id          TEXT NOT NULL DEFAULT '',
    dev_url               TEXT NOT NULL DEFAULT '',
    active_build_version  TEXT NOT NULL DEFAULT '',
    submitted_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    resolved_at           TIMESTAMP WITH TIME ZONE
);

COMMENT ON TABLE moderation_tickets IS 'Тикеты модерации игровых проектов';
COMMENT ON COLUMN moderation_tickets.status IS '1=pending, 2=approved, 3=rejected';

CREATE INDEX IF NOT EXISTS idx_moderation_tickets_status ON moderation_tickets(status, submitted_at DESC);
CREATE INDEX IF NOT EXISTS idx_moderation_tickets_project ON moderation_tickets(project_id);

-- ─────────────────────────────────────────────────────────────────────────────
-- Таблица project_releases — опубликованные версии игр (для игроков)
-- ─────────────────────────────────────────────────────────────────────────────

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

COMMENT ON TABLE project_releases IS 'Снапшоты опубликованных релизов игр в продуктивном окружении';

CREATE INDEX IF NOT EXISTS idx_project_releases_project ON project_releases(project_id, is_active);

-- ─────────────────────────────────────────────────────────────────────────────
-- Таблица deployments — аудит-лог развертываний в Dev/Prod
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS deployments (
    id            BIGSERIAL PRIMARY KEY,
    project_id    BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    environment   SMALLINT NOT NULL DEFAULT 1, -- 1=dev, 2=prod
    version       TEXT NOT NULL DEFAULT '',
    status        SMALLINT NOT NULL DEFAULT 1, -- 1=pending, 2=success, 3=failed
    error_message TEXT NOT NULL DEFAULT '',
    deployed_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE deployments IS 'История операций развертывания веб-сборок в dev и prod окружения';
COMMENT ON COLUMN deployments.environment IS '1=dev, 2=prod';
COMMENT ON COLUMN deployments.status IS '1=pending, 2=success, 3=failed';

CREATE INDEX IF NOT EXISTS idx_deployments_project ON deployments(project_id, deployed_at DESC);

-- ─────────────────────────────────────────────────────────────────────────────
-- Функции и триггеры обновления updated_at
-- ─────────────────────────────────────────────────────────────────────────────

CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_projects_updated_at ON projects;
CREATE TRIGGER trigger_projects_updated_at
    BEFORE UPDATE ON projects
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

DROP TRIGGER IF EXISTS trigger_project_drafts_updated_at ON project_drafts;
CREATE TRIGGER trigger_project_drafts_updated_at
    BEFORE UPDATE ON project_drafts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
