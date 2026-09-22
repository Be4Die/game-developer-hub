-- 04_moderation_locking.sql — флаг нахождения на модерации и блокировки изменений
ALTER TABLE projects ADD COLUMN IF NOT EXISTS is_under_review BOOLEAN NOT NULL DEFAULT FALSE;

COMMENT ON COLUMN projects.is_under_review IS 'Флаг нахождения проекта на модерации (блокирует изменения черновика, билдов и товаров)';

CREATE INDEX IF NOT EXISTS idx_projects_is_under_review ON projects(is_under_review);
