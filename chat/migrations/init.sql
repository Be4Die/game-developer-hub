-- Chat migrations

CREATE TABLE IF NOT EXISTS chat_conversations (
    id VARCHAR(36) PRIMARY KEY,
    participant_id VARCHAR(36) NOT NULL,
    participant_name VARCHAR(255) NOT NULL,
    last_message TEXT,
    last_message_at TIMESTAMP NOT NULL DEFAULT NOW(),
    unread_count INT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS chat_messages (
    id VARCHAR(36) PRIMARY KEY,
    conversation_id VARCHAR(36) NOT NULL,
    sender_id VARCHAR(36) NOT NULL,
    sender_name VARCHAR(255) NOT NULL,
    sender_role VARCHAR(50) NOT NULL DEFAULT '',
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    is_read BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX IF NOT EXISTS idx_conversations_participant ON chat_conversations(participant_id);
CREATE INDEX IF NOT EXISTS idx_messages_conversation ON chat_messages(conversation_id);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON chat_messages(created_at);
