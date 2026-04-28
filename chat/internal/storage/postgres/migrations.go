package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func (r *ChatRepository) InitSchema(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS chats (
		id VARCHAR(36) PRIMARY KEY,
		type INTEGER NOT NULL DEFAULT 0,
		title VARCHAR(255) NOT NULL DEFAULT '',
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS chat_participants (
		chat_id VARCHAR(36) NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
		user_id VARCHAR(36) NOT NULL,
		PRIMARY KEY (chat_id, user_id)
	);

	CREATE TABLE IF NOT EXISTS messages (
		id VARCHAR(36) PRIMARY KEY,
		chat_id VARCHAR(36) NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
		author_id VARCHAR(36) NOT NULL,
		content TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_chat_participants_user ON chat_participants(user_id);
	CREATE INDEX IF NOT EXISTS idx_messages_chat_created ON messages(chat_id, created_at);
	`

	_, err := r.db.Exec(ctx, schema)
	return err
}

func (r *MessageRepository) InitSchema(ctx context.Context) error {
	return nil
}

func RunMigrations(ctx context.Context, db *pgxpool.Pool) error {
	schema := `
	CREATE TABLE IF NOT EXISTS chats (
		id VARCHAR(36) PRIMARY KEY,
		type INTEGER NOT NULL DEFAULT 0,
		title VARCHAR(255) NOT NULL DEFAULT '',
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS chat_participants (
		chat_id VARCHAR(36) NOT NULL,
		user_id VARCHAR(36) NOT NULL,
		PRIMARY KEY (chat_id, user_id)
	);

	CREATE TABLE IF NOT EXISTS messages (
		id VARCHAR(36) PRIMARY KEY,
		chat_id VARCHAR(36) NOT NULL,
		author_id VARCHAR(36) NOT NULL,
		content TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_chat_participants_user ON chat_participants(user_id);
	CREATE INDEX IF NOT EXISTS idx_messages_chat_created ON messages(chat_id, created_at);
	`
	_, err := db.Exec(ctx, schema)
	return err
}

func DBNamesFromEnv() []string {
	return []string{"chat"}
}

func DSNFromEnv(dbName string) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		"postgres", "postgres", "localhost", "5432", dbName)
}
