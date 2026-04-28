package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/Be4Die/game-developer-hub/chat/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChatRepository struct {
	db *pgxpool.Pool
}

func NewChatRepository(db *pgxpool.Pool) *ChatRepository {
	return &ChatRepository{db: db}
}

func (r *ChatRepository) Create(ctx context.Context, chat *domain.Chat) error {
	query := `
		INSERT INTO chats (id, type, title, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
	`
	_, err := r.db.Exec(ctx, query, chat.ID, chat.Type, chat.Title)
	if err != nil {
		return err
	}

	if len(chat.ParticipantIDs) > 0 {
		if err := r.addParticipants(ctx, chat.ID, chat.ParticipantIDs); err != nil {
			return err
		}
	}

	return nil
}

func (r *ChatRepository) addParticipants(ctx context.Context, chatID string, userIDs []string) error {
	if len(userIDs) == 0 {
		return nil
	}

	values := make([]string, 0, len(userIDs))
	args := make([]interface{}, 0, len(userIDs)+1)
	args = append(args, chatID)
	for i, id := range userIDs {
		values = append(values, fmt.Sprintf("($1, $%d)", i+2))
		args = append(args, id)
	}

	query := fmt.Sprintf("INSERT INTO chat_participants (chat_id, user_id) VALUES %s", strings.Join(values, ", "))
	_, err := r.db.Exec(ctx, query, args...)
	return err
}

func (r *ChatRepository) GetByID(ctx context.Context, id string) (*domain.Chat, error) {
	query := `
		SELECT c.id, c.type, c.title, c.created_at, c.updated_at,
		       COALESCE(array_agg(cp.user_id) FILTER (WHERE cp.user_id IS NOT NULL), ARRAY[]::text[]) as participant_ids
		FROM chats c
		LEFT JOIN chat_participants cp ON c.id = cp.chat_id
		WHERE c.id = $1
		GROUP BY c.id
	`

	row := r.db.QueryRow(ctx, query, id)

	var chat domain.Chat
	var participantIDs []string

	err := row.Scan(&chat.ID, &chat.Type, &chat.Title, &chat.CreatedAt, &chat.UpdatedAt, &participantIDs)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrChatNotFound
		}
		return nil, err
	}

	chat.ParticipantIDs = participantIDs
	return &chat, nil
}

func (r *ChatRepository) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*domain.Chat, error) {
	query := `
		SELECT c.id, c.type, c.title, c.created_at, c.updated_at,
		       COALESCE(array_agg(cp.user_id) FILTER (WHERE cp.user_id IS NOT NULL), ARRAY[]::text[]) as participant_ids
		FROM chats c
		INNER JOIN chat_participants cp ON c.id = cp.chat_id
		WHERE cp.user_id = $1
		GROUP BY c.id
		ORDER BY c.updated_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []*domain.Chat
	for rows.Next() {
		var chat domain.Chat
		var participantIDs []string

		err := rows.Scan(&chat.ID, &chat.Type, &chat.Title, &chat.CreatedAt, &chat.UpdatedAt, &participantIDs)
		if err != nil {
			return nil, err
		}

		chat.ParticipantIDs = participantIDs
		chats = append(chats, &chat)
	}

	return chats, rows.Err()
}

func (r *ChatRepository) AddParticipant(ctx context.Context, chatID, userID string) error {
	query := `INSERT INTO chat_participants (chat_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.db.Exec(ctx, query, chatID, userID)
	return err
}

func (r *ChatRepository) RemoveParticipant(ctx context.Context, chatID, userID string) error {
	query := `DELETE FROM chat_participants WHERE chat_id = $1 AND user_id = $2`
	_, err := r.db.Exec(ctx, query, chatID, userID)
	return err
}

type MessageRepository struct {
	db *pgxpool.Pool
}

func NewMessageRepository(db *pgxpool.Pool) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Create(ctx context.Context, msg *domain.Message) error {
	query := `
		INSERT INTO messages (id, chat_id, author_id, content, created_at)
		VALUES ($1, $2, $3, $4, NOW())
	`
	_, err := r.db.Exec(ctx, query, msg.ID, msg.ChatID, msg.AuthorID, msg.Content)
	return err
}

func (r *MessageRepository) ListByChat(ctx context.Context, chatID string, limit, offset int) ([]*domain.Message, error) {
	query := `
		SELECT id, chat_id, author_id, content, created_at
		FROM messages
		WHERE chat_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, chatID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*domain.Message
	for rows.Next() {
		var msg domain.Message
		err := rows.Scan(&msg.ID, &msg.ChatID, &msg.AuthorID, &msg.Content, &msg.CreatedAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, &msg)
	}

	return messages, rows.Err()
}

func New(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return pool, nil
}
