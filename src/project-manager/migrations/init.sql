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
    title_ru              VARCHAR(50) NOT NULL DEFAULT '',
    title_en              VARCHAR(50) NOT NULL DEFAULT '',
    seo_ru                VARCHAR(180) NOT NULL DEFAULT '',
    seo_en                VARCHAR(180) NOT NULL DEFAULT '',
    about_ru              VARCHAR(800) NOT NULL DEFAULT '',
    about_en              VARCHAR(800) NOT NULL DEFAULT '',
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
-- Таблица project_releases — опубликованные версии игр (для игроков)
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS project_releases (
    id                    BIGSERIAL PRIMARY KEY,
    project_id            BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    version               TEXT NOT NULL,
    title_ru              VARCHAR(50) NOT NULL DEFAULT '',
    title_en              VARCHAR(50) NOT NULL DEFAULT '',
    seo_ru                VARCHAR(180) NOT NULL DEFAULT '',
    seo_en                VARCHAR(180) NOT NULL DEFAULT '',
    about_ru              VARCHAR(800) NOT NULL DEFAULT '',
    about_en              VARCHAR(800) NOT NULL DEFAULT '',
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

-- ─────────────────────────────────────────────────────────────────────────────
-- Таблица project_members — участники проектов (общий доступ)
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS project_members (
    id          BIGSERIAL PRIMARY KEY,
    project_id  BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id     TEXT NOT NULL,
    user_email  TEXT NOT NULL DEFAULT '',
    user_name   TEXT NOT NULL DEFAULT '',
    permissions TEXT[] NOT NULL DEFAULT '{}',
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, user_id)
);

COMMENT ON TABLE project_members IS 'Участники проекта с правами доступа';
CREATE INDEX IF NOT EXISTS idx_project_members_user ON project_members(user_id);
CREATE INDEX IF NOT EXISTS idx_project_members_project ON project_members(project_id);

DROP TRIGGER IF EXISTS trigger_project_members_updated_at ON project_members;
CREATE TRIGGER trigger_project_members_updated_at
    BEFORE UPDATE ON project_members
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ─────────────────────────────────────────────────────────────────────────────
-- Таблица project_invitations — приглашения в проект
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS project_invitations (
    id             BIGSERIAL PRIMARY KEY,
    project_id     BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    inviter_id     TEXT NOT NULL,
    inviter_email  TEXT NOT NULL DEFAULT '',
    inviter_name   TEXT NOT NULL DEFAULT '',
    invitee_id     TEXT NOT NULL,
    invitee_email  TEXT NOT NULL DEFAULT '',
    permissions    TEXT[] NOT NULL DEFAULT '{}',
    status         SMALLINT NOT NULL DEFAULT 1, -- 1=pending, 2=accepted, 3=declined, 4=canceled
    created_at     TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE project_invitations IS 'Приглашения разработчикам на доступ к проектам';
COMMENT ON COLUMN project_invitations.status IS '1=pending, 2=accepted, 3=declined, 4=canceled';

CREATE INDEX IF NOT EXISTS idx_project_invitations_invitee ON project_invitations(invitee_id, status);
CREATE INDEX IF NOT EXISTS idx_project_invitations_project ON project_invitations(project_id);

DROP TRIGGER IF EXISTS trigger_project_invitations_updated_at ON project_invitations;
CREATE TRIGGER trigger_project_invitations_updated_at
    BEFORE UPDATE ON project_invitations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ─────────────────────────────────────────────────────────────────────────────
-- Таблица user_access_blocks — блокировки пользователей (защита от спама)
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS user_access_blocks (
    id                 BIGSERIAL PRIMARY KEY,
    user_id            TEXT NOT NULL,
    blocked_user_id    TEXT NOT NULL,
    blocked_user_email TEXT NOT NULL DEFAULT '',
    blocked_user_name  TEXT NOT NULL DEFAULT '',
    created_at         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, blocked_user_id)
);

COMMENT ON TABLE user_access_blocks IS 'Черный список пользователей для блокировки спама приглашениями';
CREATE INDEX IF NOT EXISTS idx_user_access_blocks_user ON user_access_blocks(user_id);
CREATE INDEX IF NOT EXISTS idx_user_access_blocks_blocked ON user_access_blocks(blocked_user_id);

