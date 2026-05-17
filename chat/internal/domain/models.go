package domain

import (
	"context"
	"time"
)

type Conversation struct {
	ID            string
	UserID        string
	ParticipantID string
	ParticipantName string
	LastMessage   string
	LastMessageAt time.Time
	UnreadCount   int
}

type Message struct {
	ID             string
	ConversationID string
	SenderID       string
	SenderName     string
	SenderRole     string
	Content        string
	CreatedAt      time.Time
	IsRead         bool
}

type ConversationRepository interface {
	Create(ctx context.Context, conv *Conversation) error
	GetByID(ctx context.Context, id string) (*Conversation, error)
	GetByParticipants(ctx context.Context, userID1, userID2 string) (*Conversation, error)
	ListByUser(ctx context.Context, userID string) ([]Conversation, error)
	Update(ctx context.Context, conv *Conversation) error
	Delete(ctx context.Context, id string) error
}

type MessageRepository interface {
	Create(ctx context.Context, msg *Message) error
	GetByID(ctx context.Context, id string) (*Message, error)
	ListByConversation(ctx context.Context, conversationID string, limit, offset int) ([]Message, int, error)
	MarkAsRead(ctx context.Context, conversationID, userID string) error
	CountUnread(ctx context.Context, userID string) (int, error)
}
