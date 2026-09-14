-- 04_request_types.sql
-- Расширение таблицы moderation_requests для поддержки различных типов заявок (в т.ч. серверов платформы).

ALTER TABLE moderation_requests ADD COLUMN IF NOT EXISTS request_type SMALLINT NOT NULL DEFAULT 1;
-- 1 = REQUEST_TYPE_PROJECT_PUBLICATION (Публикация проекта)
-- 2 = REQUEST_TYPE_SERVER_ACCESS (Доступ к серверам платформы)

ALTER TABLE moderation_requests ADD COLUMN IF NOT EXISTS reason TEXT NOT NULL DEFAULT '';
ALTER TABLE moderation_requests ADD COLUMN IF NOT EXISTS max_instances INT NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_mod_requests_type ON moderation_requests(request_type);
