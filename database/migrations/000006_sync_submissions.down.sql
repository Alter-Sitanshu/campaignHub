ALTER TABLE submissions
DROP COLUMN IF EXISTS last_synced_at;

ALTER TABLE submissions
DROP COLUMN IF EXISTS sync_frequency_minutes;

DROP INDEX IF EXISTS idx_active_submissions;