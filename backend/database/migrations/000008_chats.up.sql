CREATE TABLE IF NOT EXISTS conversations (
    id varchar(36) primary key,
    participant_one varchar(36) not null,
    participant_two varchar(36) not null,
    type varchar(10) not null check (type IN ('direct', 'campaign')),
    campaign_id varchar(36),  -- can be null if type != campaign
    status varchar(6) DEFAULT 'active' check (status IN ('active', 'closed')), -- default status code for active is 1
    created_at timestamptz DEFAULT now(),
    last_message_at timestamptz DEFAULT now()
);

CREATE TABLE IF NOT EXISTS messages (
    client_id varchar(30) not null,
    id varchar(36) primary key,
    conversation_id varchar(36) not null,
    sender_id varchar(36) not null,
    -- I know the tags will always be 3 chars only
    message_type varchar(3) not null check (message_type IN ('txt', 'pdf', 'img')),
    content text not null,
    is_read boolean DEFAULT FALSE,
    created_at timestamptz DEFAULT now(),
    seq BIGSERIAL,

    CONSTRAINT fk_conversation_id FOREIGN KEY (conversation_id) REFERENCES conversations(id)
);

-- INDEX on created_ar and seq for faster retrieval
CREATE INDEX idx_conversation_created_at_seq ON messages (conversation_id, created_at, seq);

CREATE UNIQUE INDEX uniq_direct_conversation
ON conversations (
    LEAST(participant_one, participant_two),
    GREATEST(participant_one, participant_two)
)
WHERE type = 'direct';
