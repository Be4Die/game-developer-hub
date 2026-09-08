-- 02_shared_access.sql — таблицы для системы общего доступа к проектам

CREATE TABLE IF NOT EXISTS project_members (
    id          BIGSERIAL PRIMARY KEY,
    project_id  BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id     TEXT NOT NULL,
    user_email  TEXT NOT NULL DEFAULT '',
    user_name   TEXT NOT NULL DEFAULT '',
    permissions TEXT[] NOT NULL DEFAULT '{}',
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_project_members_user ON project_members(user_id);
CREATE INDEX IF NOT EXISTS idx_project_members_project ON project_members(project_id);

DROP TRIGGER IF EXISTS trigger_project_members_updated_at ON project_members;
CREATE TRIGGER trigger_project_members_updated_at
    BEFORE UPDATE ON project_members
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TABLE IF NOT EXISTS project_invitations (
    id             BIGSERIAL PRIMARY KEY,
    project_id     BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    inviter_id     TEXT NOT NULL,
    inviter_email  TEXT NOT NULL DEFAULT '',
    inviter_name   TEXT NOT NULL DEFAULT '',
    invitee_id     TEXT NOT NULL,
    invitee_email  TEXT NOT NULL DEFAULT '',
    permissions    TEXT[] NOT NULL DEFAULT '{}',
    status         SMALLINT NOT NULL DEFAULT 1, -- 1=pending, 2=accepted, 3=declined, 4=canceled
    created_at     TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_project_invitations_invitee ON project_invitations(invitee_id, status);
CREATE INDEX IF NOT EXISTS idx_project_invitations_project ON project_invitations(project_id);

DROP TRIGGER IF EXISTS trigger_project_invitations_updated_at ON project_invitations;
CREATE TRIGGER trigger_project_invitations_updated_at
    BEFORE UPDATE ON project_invitations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TABLE IF NOT EXISTS user_access_blocks (
    id                 BIGSERIAL PRIMARY KEY,
    user_id            TEXT NOT NULL,
    blocked_user_id    TEXT NOT NULL,
    blocked_user_email TEXT NOT NULL DEFAULT '',
    blocked_user_name  TEXT NOT NULL DEFAULT '',
    created_at         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, blocked_user_id)
);

CREATE INDEX IF NOT EXISTS idx_user_access_blocks_user ON user_access_blocks(user_id);
CREATE INDEX IF NOT EXISTS idx_user_access_blocks_blocked ON user_access_blocks(blocked_user_id);
