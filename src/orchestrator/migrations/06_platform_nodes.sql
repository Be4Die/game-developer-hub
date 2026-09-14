-- 06_platform_nodes.sql — добавление флага платформенной ноды и таблицы активных квот (грантов) проектов

ALTER TABLE nodes ADD COLUMN IF NOT EXISTS is_platform BOOLEAN NOT NULL DEFAULT false;

COMMENT ON COLUMN nodes.is_platform IS 'Флаг платформенной ноды (пул общих серверов платформы)';

CREATE TABLE IF NOT EXISTS platform_grants (
    id BIGSERIAL PRIMARY KEY,
    game_id BIGINT NOT NULL UNIQUE,
    max_instances INTEGER NOT NULL DEFAULT 5,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE platform_grants IS 'Активные разрешения (квоты) проектов на использование мощностей платформы';
CREATE INDEX IF NOT EXISTS idx_platform_grants_game ON platform_grants(game_id);
