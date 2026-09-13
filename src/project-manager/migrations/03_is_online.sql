-- 03_is_online.sql — признак онлайн-игры (is_online)
ALTER TABLE project_drafts ADD COLUMN IF NOT EXISTS is_online BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE project_releases ADD COLUMN IF NOT EXISTS is_online BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE projects ADD COLUMN IF NOT EXISTS is_online BOOLEAN NOT NULL DEFAULT FALSE;

COMMENT ON COLUMN project_drafts.is_online IS 'Флаг онлайн-игры с выделенными серверами и оркестрацией';
COMMENT ON COLUMN project_releases.is_online IS 'Снимок статуса онлайн-игры на момент публикации релиза';
COMMENT ON COLUMN projects.is_online IS 'Признак онлайн-игры проекта';
