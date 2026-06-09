CREATE TABLE IF NOT EXISTS game_moderations (
    id BIGSERIAL PRIMARY KEY,
    game_id BIGINT NOT NULL UNIQUE,
    developer_id VARCHAR(255) NOT NULL,
    game_name VARCHAR(255) NOT NULL,
    game_description TEXT NOT NULL,
    moderator_id VARCHAR(255) DEFAULT '',
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    rejection_reason TEXT DEFAULT '',
    submitted_at TIMESTAMP NOT NULL DEFAULT NOW(),
    reviewed_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_game_moderations_status ON game_moderations(status);
CREATE INDEX IF NOT EXISTS idx_game_moderations_game_id ON game_moderations(game_id);
