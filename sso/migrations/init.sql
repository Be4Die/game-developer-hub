-- SSO Database Schema
-- Usage: docker exec -i sso-postgres psql -U postgres -d sso < migrations/init.sql

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash BYTEA NOT NULL,
    display_name  VARCHAR(255) NOT NULL DEFAULT '',
    role          SMALLINT NOT NULL DEFAULT 1,  -- 1=developer, 2=moderator, 3=admin
    status        SMALLINT NOT NULL DEFAULT 1,  -- 1=active, 2=suspended, 3=deleted
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Sessions table
CREATE TABLE IF NOT EXISTS sessions (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_agent       VARCHAR(512) NOT NULL DEFAULT '',
    ip_address       VARCHAR(45) NOT NULL DEFAULT '',
    refresh_token_hash VARCHAR(128) NOT NULL DEFAULT '',
    created_at       TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_used_at     TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at       TIMESTAMP WITH TIME ZONE NOT NULL,
    revoked          BOOLEAN NOT NULL DEFAULT FALSE,
    revoked_at       TIMESTAMP WITH TIME ZONE
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_sessions_refresh_token ON sessions(refresh_token_hash);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);

-- Comments
COMMENT ON TABLE users IS 'Пользователи платформы';
COMMENT ON COLUMN users.role IS '1=developer, 2=moderator, 3=admin';
COMMENT ON COLUMN users.status IS '1=active, 2=suspended, 3=deleted';
COMMENT ON TABLE sessions IS 'Активные сессии пользователей';

-- Function for automatic updated_at (must be created before triggers)
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers
DROP TRIGGER IF EXISTS trigger_users_updated_at ON users;
CREATE TRIGGER trigger_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

DROP TRIGGER IF EXISTS trigger_sessions_updated_at ON sessions;
CREATE TRIGGER trigger_sessions_updated_at
    BEFORE UPDATE ON sessions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();
