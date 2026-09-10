-- 03_service_backups.sql — Резервные копии управляемых сервисов (Postgres, MySQL, Redis, Volumes)

CREATE TABLE IF NOT EXISTS node_service_backups (
    id                BIGSERIAL PRIMARY KEY,
    backup_id         VARCHAR(100) NOT NULL UNIQUE,
    node_id           BIGINT NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    service_id        BIGINT NOT NULL REFERENCES node_services(id) ON DELETE CASCADE,
    service_name      VARCHAR(100) NOT NULL,
    service_type      SMALLINT NOT NULL, -- 1=postgres, 2=redis, 3=mysql, 5=volume
    file_name         VARCHAR(255) NOT NULL,
    size_bytes        BIGINT NOT NULL DEFAULT 0,
    checksum          VARCHAR(64) NOT NULL DEFAULT '',
    backup_type       SMALLINT NOT NULL DEFAULT 1, -- 1=manual, 2=uploaded, 3=scheduled
    status            SMALLINT NOT NULL DEFAULT 2, -- 1=creating, 2=ready, 3=failed, 4=restoring
    created_at        TIMESTAMP NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE node_service_backups IS 'Резервные копии управляемых сервисов и баз данных';
COMMENT ON COLUMN node_service_backups.backup_type IS '1=manual, 2=uploaded, 3=scheduled';
COMMENT ON COLUMN node_service_backups.status IS '1=creating, 2=ready, 3=failed, 4=restoring';

CREATE INDEX IF NOT EXISTS idx_node_service_backups_node_service ON node_service_backups(node_id, service_name);
CREATE INDEX IF NOT EXISTS idx_node_service_backups_created ON node_service_backups(created_at DESC);
