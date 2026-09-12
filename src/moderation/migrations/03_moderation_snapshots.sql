-- 03_moderation_snapshots.sql — таблица неизменяемых аудит-снимков модерации

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
