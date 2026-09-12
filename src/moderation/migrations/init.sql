-- init.sql — инициализация схемы БД сервиса moderation.

-- ─────────────────────────────────────────────────────────────────────────────
-- Таблица moderation_requests — запросы на модерацию игровых проектов
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS moderation_requests (
    id                    BIGSERIAL PRIMARY KEY,
    project_id            BIGINT NOT NULL,
    owner_id              TEXT NOT NULL,
    moderator_id          TEXT NOT NULL DEFAULT '',
    status                SMALLINT NOT NULL DEFAULT 1, -- 1=pending, 2=in_review, 3=approved, 4=rejected, 5=cancelled
    snapshot_meta         JSONB NOT NULL DEFAULT '{}'::jsonb,
    rejection_reason      TEXT NOT NULL DEFAULT '',
    submitted_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    started_review_at     TIMESTAMP WITH TIME ZONE,
    resolved_at           TIMESTAMP WITH TIME ZONE,
    created_at            TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE moderation_requests IS 'Запросы на модерацию игровых проектов со слепком данных черновика';
COMMENT ON COLUMN moderation_requests.status IS '1=pending, 2=in_review, 3=approved, 4=rejected, 5=cancelled';

CREATE INDEX IF NOT EXISTS idx_mod_requests_project ON moderation_requests(project_id, submitted_at DESC);
CREATE INDEX IF NOT EXISTS idx_mod_requests_status ON moderation_requests(status, submitted_at ASC);
CREATE INDEX IF NOT EXISTS idx_mod_requests_moderator ON moderation_requests(moderator_id) WHERE moderator_id != '';
CREATE INDEX IF NOT EXISTS idx_mod_requests_owner ON moderation_requests(owner_id);

-- ─────────────────────────────────────────────────────────────────────────────
-- Таблица moderation_messages — сообщения и системные события в чате проекта
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS moderation_messages (
    id            BIGSERIAL PRIMARY KEY,
    project_id    BIGINT NOT NULL,
    request_id    BIGINT REFERENCES moderation_requests(id) ON DELETE SET NULL,
    sender_id     TEXT NOT NULL,                        -- ID пользователя или 'system'
    sender_role   SMALLINT NOT NULL DEFAULT 1,          -- 1=developer, 2=moderator, 3=system
    message_type  SMALLINT NOT NULL DEFAULT 1,          -- 1=text, 2=submitted, 3=status_changed, 4=approved, 5=rejected
    content       TEXT NOT NULL DEFAULT '',
    payload       JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE moderation_messages IS 'Сообщения чата проекта и аудит системных событий модерации';
COMMENT ON COLUMN moderation_messages.sender_role IS '1=developer, 2=moderator, 3=system';
COMMENT ON COLUMN moderation_messages.message_type IS '1=text, 2=submitted, 3=status_changed, 4=approved, 5=rejected';

CREATE INDEX IF NOT EXISTS idx_mod_messages_project ON moderation_messages(project_id, created_at ASC);

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

DROP TRIGGER IF EXISTS trigger_moderation_requests_updated_at ON moderation_requests;
CREATE TRIGGER trigger_moderation_requests_updated_at
    BEFORE UPDATE ON moderation_requests
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ─────────────────────────────────────────────────────────────────────────────
-- Таблица moderation_attachments — вложения (фото и видео) в чате проекта
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS moderation_attachments (
    id            TEXT PRIMARY KEY,                     -- UUID строкой
    project_id    BIGINT NOT NULL,
    message_id    BIGINT REFERENCES moderation_messages(id) ON DELETE SET NULL,
    uploader_id   TEXT NOT NULL,
    uploader_role SMALLINT NOT NULL DEFAULT 1,          -- 1=developer, 2=moderator, 3=admin
    file_name     TEXT NOT NULL,                        -- Исходное имя файла
    file_size     BIGINT NOT NULL,                      -- Размер файла в байтах
    mime_type     TEXT NOT NULL,                        -- "image/png", "image/jpeg", "video/mp4", etc.
    storage_path  TEXT NOT NULL DEFAULT '',             -- Путь в S3 или файловой системе
    is_purged     BOOLEAN NOT NULL DEFAULT FALSE,       -- Флаг очистки файла после публикации проекта
    created_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE moderation_attachments IS 'Метаданные загруженных фото и видео вложений для чата модерации проекта';
COMMENT ON COLUMN moderation_attachments.uploader_role IS '1=developer, 2=moderator, 3=admin';
COMMENT ON COLUMN moderation_attachments.is_purged IS 'true если физический файл удален после одобрения проекта';

CREATE INDEX IF NOT EXISTS idx_mod_attachments_project ON moderation_attachments(project_id);
CREATE INDEX IF NOT EXISTS idx_mod_attachments_message ON moderation_attachments(message_id) WHERE message_id IS NOT NULL;

-- ─────────────────────────────────────────────────────────────────────────────
-- Таблица moderation_snapshots — неизменяемые сжатые аудит-снимки решений
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS moderation_snapshots (
    id                BIGSERIAL PRIMARY KEY,
    request_id        BIGINT UNIQUE NOT NULL REFERENCES moderation_requests(id) ON DELETE CASCADE,
    project_id        BIGINT NOT NULL,
    status            SMALLINT NOT NULL,                     -- 3=approved, 4=rejected, 5=cancelled
    format_version    INT NOT NULL DEFAULT 1,
    snapshot_json     JSONB NOT NULL DEFAULT '{}'::jsonb,     -- Структурированный JSON для быстрого чтения через API
    snapshot_blob     BYTEA NOT NULL DEFAULT ''::bytea,      -- Сжатый бинарный блоб (gzip) для компактного долговременного хранения
    created_at        TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE moderation_snapshots IS 'Неизменяемые сжатые аудит-снимки состояния проекта, медиа (с превью и sha256) и переписки на момент вердикта';
COMMENT ON COLUMN moderation_snapshots.status IS '3=approved, 4=rejected, 5=cancelled';
COMMENT ON COLUMN moderation_snapshots.snapshot_blob IS 'Gzip-сжатый бинарный срез полного состояния (JSON -> gzip -> bytea)';

CREATE INDEX IF NOT EXISTS idx_mod_snapshots_project ON moderation_snapshots(project_id);
CREATE INDEX IF NOT EXISTS idx_mod_snapshots_request ON moderation_snapshots(request_id);

