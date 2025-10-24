ALTER TABLE submissions
DROP CONSTRAINT IF EXISTS chk_like_count_positive;

-- Drop indexes
DROP INDEX IF EXISTS idx_submissions_video_title;
DROP INDEX IF EXISTS idx_submissions_platform_video;
DROP INDEX IF EXISTS idx_submissions_video_status;

-- Drop video metadata columns
ALTER TABLE submissions
DROP COLUMN IF EXISTS video_title,
DROP COLUMN IF EXISTS video_platform,
DROP COLUMN IF EXISTS platform_video_id,
DROP COLUMN IF EXISTS thumbnail_url,
DROP COLUMN IF EXISTS like_count,
DROP COLUMN IF EXISTS uploaded_at,
DROP COLUMN IF EXISTS video_status;

