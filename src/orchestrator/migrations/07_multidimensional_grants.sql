-- Расширение таблицы platform_grants многомерными квотами
ALTER TABLE platform_grants ADD COLUMN IF NOT EXISTS max_total_cpu_millis INT NOT NULL DEFAULT 0;
ALTER TABLE platform_grants ADD COLUMN IF NOT EXISTS max_total_memory_mb BIGINT NOT NULL DEFAULT 0;
ALTER TABLE platform_grants ADD COLUMN IF NOT EXISTS max_instance_cpu_millis INT NOT NULL DEFAULT 0;
ALTER TABLE platform_grants ADD COLUMN IF NOT EXISTS max_instance_memory_mb BIGINT NOT NULL DEFAULT 0;

-- Фиксация выделенных ресурсов инстанса в таблице instances
ALTER TABLE instances ADD COLUMN IF NOT EXISTS allocated_cpu_millis INT NOT NULL DEFAULT 0;
ALTER TABLE instances ADD COLUMN IF NOT EXISTS allocated_memory_bytes BIGINT NOT NULL DEFAULT 0;
