-- init.sql — инициализация схемы БД оркестратора.
-- Запускается при первом развёртывании или для пересоздания БД.
--
-- Использование:
--   psql -U postgres -d orchestrator -f migrations/init.sql
--   docker exec -i orchestrator-postgres psql -U postgres -d orchestrator < migrations/init.sql

-- ─────────────────────────────────────────────────────────────────────────────
-- Таблица nodes — вычислительные узлы (game-server-node)
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS nodes (
    id            BIGSERIAL PRIMARY KEY,
    owner_id      TEXT NOT NULL DEFAULT '',           -- ID владельца (пользователь)
    address       TEXT NOT NULL UNIQUE,              -- gRPC-адрес (host:port)
    token_hash    BYTEA NOT NULL,                    -- хеш авторизационного токена
    api_token     TEXT NOT NULL DEFAULT '',          -- plaintext токен для gRPC-запросов
    region        TEXT,                              -- опциональный регион
    status        SMALLINT NOT NULL DEFAULT 1,       -- 1=unauthorized, 2=online, 3=offline, 4=maintenance
    cpu_cores     INTEGER NOT NULL DEFAULT 0,        -- количество CPU ядер
    total_memory  BIGINT NOT NULL DEFAULT 0,         -- объём оперативной памяти (bytes)
    total_disk    BIGINT NOT NULL DEFAULT 0,         -- объём диска (bytes)
    agent_version TEXT NOT NULL DEFAULT '',          -- версия агента ноды
    last_ping_at  TIMESTAMP NOT NULL DEFAULT NOW(),  -- время последнего heartbeat
    created_at    TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE nodes IS 'Реестр вычислительных узлов (game-server-node)';
COMMENT ON COLUMN nodes.status IS '1=unauthorized, 2=online, 3=offline, 4=maintenance';

CREATE INDEX IF NOT EXISTS idx_nodes_status ON nodes(status);
CREATE INDEX IF NOT EXISTS idx_nodes_address ON nodes(address);
CREATE INDEX IF NOT EXISTS idx_nodes_last_ping ON nodes(last_ping_at);

-- ─────────────────────────────────────────────────────────────────────────────
-- Таблица server_builds — серверные билды игр
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS server_builds (
    id             BIGSERIAL PRIMARY KEY,
    owner_id       TEXT NOT NULL DEFAULT '',          -- ID владельца (пользователь)
    game_id        BIGINT NOT NULL,                  -- ID игры
    uploaded_by    BIGINT NOT NULL DEFAULT 0,        -- ID пользователя (0=неизвестно)
    version        TEXT NOT NULL,                    -- версия билда (semver)
    image_tag      TEXT NOT NULL,                    -- Docker-тег образа
    protocol       SMALLINT NOT NULL,                -- 1=tcp, 2=udp, 3=websocket, 4=webrtc
    internal_port  INTEGER NOT NULL,                 -- порт внутри контейнера
    max_players    INTEGER NOT NULL,                 -- макс. игроков
    file_url       TEXT NOT NULL,                    -- путь к файлу в хранилище
    file_size      BIGINT NOT NULL,                  -- размер файла (bytes)
    created_at     TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (game_id, version)
);

COMMENT ON TABLE server_builds IS 'Метаданные серверных билдов игр';
COMMENT ON COLUMN server_builds.protocol IS '1=tcp, 2=udp, 3=websocket, 4=webrtc';

CREATE INDEX IF NOT EXISTS idx_server_builds_game ON server_builds(game_id);
CREATE INDEX IF NOT EXISTS idx_server_builds_created ON server_builds(created_at DESC);

-- ─────────────────────────────────────────────────────────────────────────────
-- Таблица instances — экземпляры игровых серверов
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS instances (
    id               BIGSERIAL PRIMARY KEY,
    owner_id         TEXT NOT NULL DEFAULT '',       -- ID владельца (пользователь)
    node_id          BIGINT NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    server_build_id  BIGINT NOT NULL REFERENCES server_builds(id) ON DELETE RESTRICT,
    game_id          BIGINT NOT NULL,                  -- денормализовано для быстрых запросов
    name             TEXT NOT NULL,                    -- имя инстанса
    build_version    TEXT NOT NULL,                    -- версия билда
    protocol         SMALLINT NOT NULL,                -- 1=tcp, 2=udp, 3=websocket, 4=webrtc
    host_port        INTEGER NOT NULL,                 -- порт на хосте
    internal_port    INTEGER NOT NULL,                 -- порт внутри контейнера
    status           SMALLINT NOT NULL DEFAULT 1,      -- 1=starting, 2=running, 3=stopping, 4=stopped, 5=crashed
    max_players      INTEGER NOT NULL,
    developer_payload JSONB,                           -- произвольные данные разработчика
    server_address   TEXT NOT NULL,                    -- адрес для клиентов
    started_at       TIMESTAMP NOT NULL,
    created_at       TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMP NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE instances IS 'Экземпляры игровых серверов';
COMMENT ON COLUMN instances.status IS '1=starting, 2=running, 3=stopping, 4=stopped, 5=crashed';
COMMENT ON COLUMN instances.protocol IS '1=tcp, 2=udp, 3=websocket, 4=webrtc';

CREATE INDEX IF NOT EXISTS idx_instances_game ON instances(game_id);
CREATE INDEX IF NOT EXISTS idx_instances_node ON instances(node_id);
CREATE INDEX IF NOT EXISTS idx_instances_status ON instances(status);
CREATE INDEX IF NOT EXISTS idx_instances_build ON instances(server_build_id);
CREATE INDEX IF NOT EXISTS idx_instances_game_status ON instances(game_id, status);

-- ─────────────────────────────────────────────────────────────────────────────
-- Функция для обновления updated_at
-- ─────────────────────────────────────────────────────────────────────────────

CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_nodes_updated_at
    BEFORE UPDATE ON nodes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trigger_instances_updated_at
    BEFORE UPDATE ON instances
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ─────────────────────────────────────────────────────────────────────────────
-- Очистка и пересоздание последовательностей (для тестов)
-- ─────────────────────────────────────────────────────────────────────────────

-- Выполняйте только если нужно сбросить БД:
-- TRUNCATE instances, server_builds, nodes RESTART IDENTITY CASCADE;
