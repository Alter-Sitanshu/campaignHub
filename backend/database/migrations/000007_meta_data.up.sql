-- Add video metadata columns
ALTER TABLE submissions
ADD COLUMN video_title VARCHAR(500),
ADD COLUMN video_platform VARCHAR(20) DEFAULT 'youtube',
ADD COLUMN platform_video_id VARCHAR(100),
ADD COLUMN thumbnail_url TEXT,
ADD COLUMN like_count INTEGER DEFAULT 0,
ADD COLUMN uploaded_at TIMESTAMP,
ADD COLUMN video_status VARCHAR(20) DEFAULT 'available'; -- available, deleted, private
-- Create indexes for common queries
CREATE INDEX idx_submissions_video_title ON submissions USING gin(to_tsvector('english', video_title));
CREATE INDEX idx_submissions_platform_video ON submissions(video_platform, platform_video_id);
CREATE INDEX idx_submissions_video_status ON submissions(video_status) WHERE video_status != 'available';

-- Add check constraint
ALTER TABLE submissions
ADD CONSTRAINT chk_like_count_positive CHECK (like_count >= 0);