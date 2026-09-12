-- 02_chat_attachments.sql — таблица для хранения метаданных вложений чата модерации

CREATE TABLE IF NOT EXISTS moderation_attachments (
    id            TEXT PRIMARY KEY,                     -- UUID строкой (gen_random_uuid()::text или клиентский UUID)
    project_id    BIGINT NOT NULL,
    message_id    BIGINT REFERENCES moderation_messages(id) ON DELETE SET NULL,
    uploader_id   TEXT NOT NULL,
    uploader_role SMALLINT NOT NULL DEFAULT 1,          -- 1=developer, 2=moderator, 3=admin
    file_name     TEXT NOT NULL,                        -- Исходное имя файла (например, "screenshot.png")
    file_size     BIGINT NOT NULL,                      -- Размер файла в байтах
    mime_type     TEXT NOT NULL,                        -- "image/png", "image/jpeg", "image/webp", "video/mp4", "video/webm"
    storage_path  TEXT NOT NULL DEFAULT '',             -- Путь в S3 или файловой системе
    is_purged     BOOLEAN NOT NULL DEFAULT FALSE,       -- Флаг очистки файла после публикации проекта
    created_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE moderation_attachments IS 'Метаданные загруженных фото и видео вложений для чата модерации проекта';
COMMENT ON COLUMN moderation_attachments.uploader_role IS '1=developer, 2=moderator, 3=admin';
COMMENT ON COLUMN moderation_attachments.is_purged IS 'true если физический файл удален после одобрения проекта';

CREATE INDEX IF NOT EXISTS idx_mod_attachments_project ON moderation_attachments(project_id);
CREATE INDEX IF NOT EXISTS idx_mod_attachments_message ON moderation_attachments(message_id) WHERE message_id IS NOT NULL;
