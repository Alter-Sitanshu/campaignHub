ALTER TABLE submissions
ADD COLUMN IF NOT EXISTS last_synced_at TIMESTAMPTZ DEFAULT now();

ALTER TABLE submissions
ADD COLUMN IF NOT EXISTS sync_frequency_minutes INTEGER DEFAULT 5; -- adaptive!

CREATE INDEX IF NOT EXISTS idx_active_submissions
ON submissions (last_synced_at);