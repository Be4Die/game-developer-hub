package postgres

import (
	"context"
	"time"

	"github.com/Be4Die/game-developer-hub/chat/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type conversationRepo struct {
	pool *pgxpool.Pool
}

func NewConversationRepository(pool *pgxpool.Pool) domain.ConversationRepository {
	return &conversationRepo{pool: pool}
}

func (r *conversationRepo) Create(ctx context.Context, conv *domain.Conversation) error {
	conv.ID = uuid.New().String()
	query := `
		INSERT INTO chat_conversations (id, user_id, user_name, participant_id, participant_name, last_message, last_message_at, unread_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.pool.Exec(ctx, query, conv.ID, conv.UserID, conv.UserName, conv.ParticipantID, conv.ParticipantName, conv.LastMessage, conv.LastMessageAt, conv.UnreadCount)
	return err
}

func (r *conversationRepo) GetByID(ctx context.Context, id string) (*domain.Conversation, error) {
	query := `
		SELECT id, user_id, user_name, participant_id, participant_name, last_message, last_message_at, unread_count
		FROM chat_conversations WHERE id = $1
	`
	var conv domain.Conversation
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&conv.ID, &conv.UserID, &conv.UserName, &conv.ParticipantID, &conv.ParticipantName, &conv.LastMessage, &conv.LastMessageAt, &conv.UnreadCount,
	)
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

func (r *conversationRepo) GetByParticipants(ctx context.Context, userID1, userID2 string) (*domain.Conversation, error) {
	query := `
		SELECT id, user_id, user_name, participant_id, participant_name, last_message, last_message_at, unread_count
		FROM chat_conversations 
		WHERE user_id = $1 AND participant_id = $2
		LIMIT 1
	`
	var conv domain.Conversation
	err := r.pool.QueryRow(ctx, query, userID1, userID2).Scan(
		&conv.ID, &conv.UserID, &conv.UserName, &conv.ParticipantID, &conv.ParticipantName, &conv.LastMessage, &conv.LastMessageAt, &conv.UnreadCount,
	)
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

func (r *conversationRepo) ListByUser(ctx context.Context, userID string) ([]domain.Conversation, error) {
	query := `
		SELECT id, user_id, user_name, participant_id, participant_name, last_message, last_message_at, unread_count
		FROM chat_conversations WHERE user_id = $1 OR participant_id = $1
		ORDER BY last_message_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []domain.Conversation
	for rows.Next() {
		var conv domain.Conversation
		if err := rows.Scan(&conv.ID, &conv.UserID, &conv.UserName, &conv.ParticipantID, &conv.ParticipantName, &conv.LastMessage, &conv.LastMessageAt, &conv.UnreadCount); err != nil {
			return nil, err
		}
		conversations = append(conversations, conv)
	}
	return conversations, nil
}

func (r *conversationRepo) Update(ctx context.Context, conv *domain.Conversation) error {
	query := `
		UPDATE chat_conversations SET last_message = $1, last_message_at = $2, unread_count = $3
		WHERE id = $4
	`
	_, err := r.pool.Exec(ctx, query, conv.LastMessage, conv.LastMessageAt, conv.UnreadCount, conv.ID)
	return err
}

func (r *conversationRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM chat_conversations WHERE id = $1", id)
	return err
}

type messageRepo struct {
	pool *pgxpool.Pool
}

func NewMessageRepository(pool *pgxpool.Pool) domain.MessageRepository {
	return &messageRepo{pool: pool}
}

func (r *messageRepo) Create(ctx context.Context, msg *domain.Message) error {
	msg.ID = uuid.New().String()
	msg.CreatedAt = time.Now()
	query := `
		INSERT INTO chat_messages (id, conversation_id, sender_id, sender_name, sender_role, content, created_at, is_read)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.pool.Exec(ctx, query, msg.ID, msg.ConversationID, msg.SenderID, msg.SenderName, msg.SenderRole, msg.Content, msg.CreatedAt, msg.IsRead)
	return err
}

func (r *messageRepo) GetByID(ctx context.Context, id string) (*domain.Message, error) {
	query := `
		SELECT id, conversation_id, sender_id, sender_name, sender_role, content, created_at, is_read
		FROM chat_messages WHERE id = $1
	`
	var msg domain.Message
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&msg.ID, &msg.ConversationID, &msg.SenderID, &msg.SenderName, &msg.SenderRole, &msg.Content, &msg.CreatedAt, &msg.IsRead,
	)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *messageRepo) ListByConversation(ctx context.Context, conversationID string, limit, offset int) ([]domain.Message, int, error) {
	countQuery := "SELECT COUNT(*) FROM chat_messages WHERE conversation_id = $1"
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, conversationID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, conversation_id, sender_id, sender_name, sender_role, content, created_at, is_read
		FROM chat_messages WHERE conversation_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, conversationID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var messages []domain.Message
	for rows.Next() {
		var msg domain.Message
		if err := rows.Scan(&msg.ID, &msg.ConversationID, &msg.SenderID, &msg.SenderName, &msg.SenderRole, &msg.Content, &msg.CreatedAt, &msg.IsRead); err != nil {
			return nil, 0, err
		}
		messages = append(messages, msg)
	}
	return messages, total, nil
}

func (r *messageRepo) MarkAsRead(ctx context.Context, conversationID, userID string) error {
	query := `
		UPDATE chat_messages SET is_read = true
		WHERE conversation_id = $1 AND sender_id != $2 AND is_read = false
	`
	_, err := r.pool.Exec(ctx, query, conversationID, userID)
	return err
}

func (r *messageRepo) CountUnread(ctx context.Context, userID string) (int, error) {
	query := `
		SELECT COUNT(*) FROM chat_messages m
		JOIN chat_conversations c ON m.conversation_id = c.id
		WHERE c.participant_id = $1 AND m.sender_id != $1 AND m.is_read = false
	`
	var count int
	err := r.pool.QueryRow(ctx, query, userID).Scan(&count)
	return count, err
}

func InitSchema(ctx context.Context, pool *pgxpool.Pool) error {
	schema := `
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
			content TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			is_read BOOLEAN NOT NULL DEFAULT false
		);

		CREATE INDEX IF NOT EXISTS idx_conversations_participant ON chat_conversations(participant_id);
		CREATE INDEX IF NOT EXISTS idx_messages_conversation ON chat_messages(conversation_id);
		CREATE INDEX IF NOT EXISTS idx_messages_created_at ON chat_messages(created_at);
	`
	_, err := pool.Exec(ctx, schema)
	if err != nil {
		return err
	}

	// Migration: add user_name column if not exists
	_, err = pool.Exec(ctx, "ALTER TABLE chat_conversations ADD COLUMN IF NOT EXISTS user_name VARCHAR(255) NOT NULL DEFAULT ''")
	if err != nil {
		return err
	}

	// Migration: add user_id column if not exists
	_, err = pool.Exec(ctx, "ALTER TABLE chat_conversations ADD COLUMN IF NOT EXISTS user_id VARCHAR(36) NOT NULL DEFAULT ''")
	if err != nil {
		return err
	}

	// Migration: add sender_role column if not exists
	_, err = pool.Exec(ctx, "ALTER TABLE chat_messages ADD COLUMN IF NOT EXISTS sender_role VARCHAR(50) NOT NULL DEFAULT ''")
	if err != nil {
		return err
	}

	// Create indexes after migration
	_, err = pool.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_conversations_user ON chat_conversations(user_id)")
	if err != nil {
		return err
	}

	return nil
}
