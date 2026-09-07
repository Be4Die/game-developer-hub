-- 02_node_roles_and_services.sql — Роли нод и управляемые сервисы хранения данных

-- 1. Добавление роли ноды: 1=mixed, 2=compute, 3=storage
ALTER TABLE nodes ADD COLUMN IF NOT EXISTS role SMALLINT NOT NULL DEFAULT 1;
COMMENT ON COLUMN nodes.role IS '1=mixed, 2=compute, 3=storage';

-- 2. Таблица управляемых сервисов баз данных и хранилищ (Postgres, Redis, MySQL, MinIO)
CREATE TABLE IF NOT EXISTS node_services (
    id                BIGSERIAL PRIMARY KEY,
    node_id           BIGINT NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    owner_id          TEXT NOT NULL,
    allowed_game_ids  BIGINT[] NOT NULL DEFAULT '{}',
    service_type      SMALLINT NOT NULL, -- 1=postgres, 2=redis, 3=mysql, 4=minio
    name              VARCHAR(100) NOT NULL,
    container_id      VARCHAR(255) NOT NULL DEFAULT '',
    host_port         INTEGER NOT NULL,
    connection_uri    TEXT NOT NULL,
    credentials       JSONB NOT NULL DEFAULT '{}',
    status            SMALLINT NOT NULL DEFAULT 1, -- 1=starting, 2=running, 3=stopped, 4=error
    volume_path       TEXT NOT NULL,
    created_at        TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (node_id, name)
);

COMMENT ON TABLE node_services IS 'Управляемые сервисы баз данных и кэшей на нодах (Postgres, Redis, MySQL, MinIO)';
COMMENT ON COLUMN node_services.service_type IS '1=postgres, 2=redis, 3=mysql, 4=minio';
COMMENT ON COLUMN node_services.status IS '1=starting, 2=running, 3=stopped, 4=error';

CREATE INDEX IF NOT EXISTS idx_node_services_node ON node_services(node_id);
CREATE INDEX IF NOT EXISTS idx_node_services_owner ON node_services(owner_id);

CREATE TRIGGER trigger_node_services_updated_at
    BEFORE UPDATE ON node_services
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
