-- migrations/002_queue.sql — поддержка очередей игроков
-- Применить: psql -U postgres -d orchestrator -f orchestrator/migrations/002_queue.sql

-- ═════════════════════════════════════════════════════════════════════════════
-- Расширение политик оркестрации: параметры очереди
-- ═════════════════════════════════════════════════════════════════════════════

ALTER TABLE game_policies
    ADD COLUMN IF NOT EXISTS queue_reservation_seconds INTEGER NOT NULL DEFAULT 30,
    ADD COLUMN IF NOT EXISTS queue_max_wait_seconds    INTEGER NOT NULL DEFAULT 300,
    ADD COLUMN IF NOT EXISTS queue_heartbeat_timeout   INTEGER NOT NULL DEFAULT 15;

COMMENT ON COLUMN game_policies.queue_reservation_seconds IS 'Секунды на подключение после резервации слота';
COMMENT ON COLUMN game_policies.queue_max_wait_seconds    IS 'Макс. время ожидания в очереди до авто-отмены';
COMMENT ON COLUMN game_policies.queue_heartbeat_timeout   IS 'Выкидывание из очереди при отсутствии heartbeat (сек)';

-- ═════════════════════════════════════════════════════════════════════════════
-- Таблица queue_events — аудит-лог очереди
-- ═════════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS queue_events (
    id           BIGSERIAL PRIMARY KEY,
    game_id      BIGINT NOT NULL,
    player_id    TEXT NOT NULL,
    event_type   SMALLINT NOT NULL, -- 1=join, 2=reserved, 3=connected, 4=timeout, 5=leave, 6=cancel
    instance_id  BIGINT,
    wait_seconds INT,
    created_at   TIMESTAMP NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE queue_events IS 'Аудит-лог событий очереди игроков';
COMMENT ON COLUMN queue_events.event_type IS '1=join, 2=reserved, 3=connected, 4=timeout, 5=leave, 6=cancel';

CREATE INDEX IF NOT EXISTS idx_queue_events_game ON queue_events(game_id, created_at);
CREATE INDEX IF NOT EXISTS idx_queue_events_player ON queue_events(game_id, player_id, created_at);
