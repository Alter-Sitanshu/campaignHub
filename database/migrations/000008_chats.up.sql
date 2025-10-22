CREATE TABLE IF NOT EXISTS conv_status (
    id int primary key,
    name varchar(10) not null
);

CREATE TABLE IF NOT EXISTS messages (
    id varchar(36) primary key,
    conversation_id varchar(36) not null,
    sender_id varchar(36) not null,
    -- I know the tags will always be 3 chars only
    message_type varchar(3) not null check (message_type IN ('txt', 'pdf', 'img')),
    content text not null,
    is_read boolean DEFAULT FALSE,
    created_at timestamptz DEFAULT now(),

    CONSTRAINT fk_conversation_id FOREIGN KEY (conversation_id) REFERENCES conversations(id)
);

CREATE TABLE IF NOT EXISTS conversations (
    id varchar(36) primary key,
    participant_one varchar(36) not null,
    participant_two varchar(36) not null,
    type varchar(10) not null check (type IN ('direct', 'campaign')),
    campaign_id varchar(36),  -- can be null if type != campaign
    status int DEFAULT 1, -- default status code for active is 1
    created_at timestamptz DEFAULT now(),
    last_message_at timestamptz DEFAULT now(),

    CONSTRAINT fk_conv_status FOREIGN KEY (status) REFERENCES conv_status (id)
);

-- Populate the status table for conversations
INSERT INTO conv_status (id, name)
VALUES (1, 'open'), (0, 'closed');