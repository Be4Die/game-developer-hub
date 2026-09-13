-- ─────────────────────────────────────────────────────────────────────────────
-- Миграция: Добавление сетевого режима ноды (Direct / Platform Proxy) и домена
-- ─────────────────────────────────────────────────────────────────────────────

ALTER TABLE nodes ADD COLUMN IF NOT EXISTS ingress_mode SMALLINT NOT NULL DEFAULT 1;
ALTER TABLE nodes ADD COLUMN IF NOT EXISTS custom_domain TEXT NOT NULL DEFAULT '';

COMMENT ON COLUMN nodes.ingress_mode IS '1=platform_proxy, 2=direct';
