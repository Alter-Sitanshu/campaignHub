ALTER TABLE platform_links
DROP CONSTRAINT fk_platform_user;

ALTER TABLE platform_links
ADD CONSTRAINT fk_platform_user
FOREIGN KEY (userid)
REFERENCES users (id);

ALTER TABLE submissions
DROP CONSTRAINT fk_submission_user;

ALTER TABLE submissions
ADD CONSTRAINT fk_submission_user
FOREIGN KEY (creator_id)
REFERENCES users (id);
