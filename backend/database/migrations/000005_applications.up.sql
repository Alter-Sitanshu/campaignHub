CREATE TABLE IF NOT EXISTS appl_status(
    id INT PRIMARY KEY,
    name VARCHAR(10) NOT NULL
);

CREATE TABLE applications(
    id VARCHAR(36) PRIMARY KEY,
    campaign_id VARCHAR(36) NOT NULL,
    creator_id VARCHAR(36) NOT NULL,
    status INT DEFAULT 3, --DEFAULT CODE 2 Signifies pending status
    created_at TIMESTAMPTZ DEFAULT now(),

    CONSTRAINT applications_campaign_creator_unique UNIQUE (campaign_id, creator_id),
    CONSTRAINT fk_application_status FOREIGN KEY (status) REFERENCES appl_status(id),
    CONSTRAINT fk_application_creator FOREIGN KEY (creator_id) REFERENCES users(id),
    CONSTRAINT fk_application_campaign FOREIGN KEY (campaign_id) REFERENCES campaigns(id)
);

CREATE INDEX idx_application_creator ON applications(campaign_id, creator_id);

INSERT INTO appl_status
VALUES (0, 'rejected'), (1, 'approved'), (2, 'pending'); 