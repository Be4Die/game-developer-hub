package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Be4Die/game-developer-hub/moderation/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MessageRepo реализует domain.MessageRepo поверх PostgreSQL.
type MessageRepo struct {
	pool *pgxpool.Pool
}

// NewMessageRepo создаёт новый экземпляр MessageRepo.
func NewMessageRepo(pool *pgxpool.Pool) *MessageRepo {
	return &MessageRepo{pool: pool}
}

// Create сохраняет сообщение или системное событие в чат проекта.
func (r *MessageRepo) Create(ctx context.Context, msg *domain.ChatMessage) (int64, error) {
	payloadJSON, err := json.Marshal(msg.Payload)
	if err != nil {
		payloadJSON = []byte("{}")
	}

	query := `
		INSERT INTO moderation_messages (
			project_id, request_id, sender_id, sender_role,
			message_type, content, payload, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, NOW()
		)
		RETURNING id, created_at
	`

	var id int64
	var createdAt time.Time

	err = r.pool.QueryRow(
		ctx, query,
		msg.ProjectID,
		msg.RequestID,
		msg.SenderID,
		int16(msg.SenderRole),
		int16(msg.MessageType),
		msg.Content,
		payloadJSON,
	).Scan(&id, &createdAt)
	if err != nil {
		return 0, fmt.Errorf("MessageRepo.Create: %w", err)
	}

	msg.ID = id
	msg.CreatedAt = createdAt

	return id, nil
}

// ListByProject возвращает сообщения чата проекта с сортировкой по времени (от старых к новым).
func (r *MessageRepo) ListByProject(ctx context.Context, projectID int64, limit, offset int) ([]*domain.ChatMessage, int, error) {
	countQuery := `SELECT COUNT(*) FROM moderation_messages WHERE project_id = $1`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, projectID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("MessageRepo.ListByProject count: %w", err)
	}

	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT
			id, project_id, request_id, sender_id, sender_role,
			message_type, content, payload, created_at
		FROM moderation_messages
		WHERE project_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, projectID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("MessageRepo.ListByProject select: %w", err)
	}
	defer rows.Close()

	var messages []*domain.ChatMessage
	for rows.Next() {
		msg := &domain.ChatMessage{}
		var payloadRaw []byte
		var senderRoleInt, msgTypeInt int16

		err := rows.Scan(
			&msg.ID,
			&msg.ProjectID,
			&msg.RequestID,
			&msg.SenderID,
			&senderRoleInt,
			&msgTypeInt,
			&msg.Content,
			&payloadRaw,
			&msg.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("MessageRepo.ListByProject scan: %w", err)
		}

		msg.SenderRole = domain.SenderRole(senderRoleInt)
		msg.MessageType = domain.MessageType(msgTypeInt)
		if len(payloadRaw) > 0 {
			_ = json.Unmarshal(payloadRaw, &msg.Payload)
		}

		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("MessageRepo.ListByProject rows: %w", err)
	}

	return messages, total, nil
}

// ListActiveChats возвращает список проектов, отсортированных по дате последнего сообщения.
func (r *MessageRepo) ListActiveChats(ctx context.Context, limit, offset int) ([]*domain.ChatSummary, int, error) {
	countQuery := `SELECT COUNT(DISTINCT project_id) FROM moderation_messages`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("MessageRepo.ListActiveChats count: %w", err)
	}

	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT m.id, m.project_id, m.request_id, m.sender_id, m.sender_role,
			   m.message_type, m.content, m.payload, m.created_at
		FROM (
			SELECT DISTINCT ON (project_id) *
			FROM moderation_messages
			ORDER BY project_id, created_at DESC, id DESC
		) m
		ORDER BY m.created_at DESC, m.id DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("MessageRepo.ListActiveChats select: %w", err)
	}
	defer rows.Close()

	var chats []*domain.ChatSummary
	for rows.Next() {
		msg := &domain.ChatMessage{}
		var payloadRaw []byte
		var senderRoleInt, msgTypeInt int16

		err := rows.Scan(
			&msg.ID,
			&msg.ProjectID,
			&msg.RequestID,
			&msg.SenderID,
			&senderRoleInt,
			&msgTypeInt,
			&msg.Content,
			&payloadRaw,
			&msg.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("MessageRepo.ListActiveChats scan: %w", err)
		}

		msg.SenderRole = domain.SenderRole(senderRoleInt)
		msg.MessageType = domain.MessageType(msgTypeInt)
		if len(payloadRaw) > 0 {
			_ = json.Unmarshal(payloadRaw, &msg.Payload)
		}

		chats = append(chats, &domain.ChatSummary{
			ProjectID:   msg.ProjectID,
			LastMessage: msg,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("MessageRepo.ListActiveChats rows: %w", err)
	}

	return chats, total, nil
}
