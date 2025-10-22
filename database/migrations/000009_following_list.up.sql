CREATE TABLE IF NOT EXISTS following_list (
    user_id varchar(36) not null,
    brand_id varchar(36) not null,
    created_at timestamptz DEFAULT now(),

    CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_brand_id FOREIGN KEY (brand_id) REFERENCES brands(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, brand_id)
);

CREATE INDEX idx_user_brand ON following_list (user_id, brand_id);