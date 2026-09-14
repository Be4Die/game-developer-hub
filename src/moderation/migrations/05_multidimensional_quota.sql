ALTER TABLE moderation_requests ADD COLUMN IF NOT EXISTS max_total_cpu_millis INT NOT NULL DEFAULT 0;
ALTER TABLE moderation_requests ADD COLUMN IF NOT EXISTS max_total_memory_mb BIGINT NOT NULL DEFAULT 0;
ALTER TABLE moderation_requests ADD COLUMN IF NOT EXISTS max_instance_cpu_millis INT NOT NULL DEFAULT 0;
ALTER TABLE moderation_requests ADD COLUMN IF NOT EXISTS max_instance_memory_mb BIGINT NOT NULL DEFAULT 0;
ALTER TABLE moderation_requests ADD COLUMN IF NOT EXISTS moderator_comment TEXT NOT NULL DEFAULT '';
