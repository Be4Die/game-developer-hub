-- ─────────────────────────────────────────────────────────────────────────────
-- Миграция: Добавление настроек планировщика бэкапов
-- ─────────────────────────────────────────────────────────────────────────────

ALTER TABLE nodes ADD COLUMN backups_enabled BOOLEAN NOT NULL DEFAULT true;
ALTER TABLE node_services ADD COLUMN auto_backup_enabled BOOLEAN NOT NULL DEFAULT false;

