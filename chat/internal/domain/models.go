package domain

import (
	"context"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	chatv1 "github.com/Be4Die/game-developer-hub/protos/chat/v1"
)

type ChatType int32

const (
	ChatTypeUnspecified ChatType = 0
	ChatTypeDirect      ChatType = 1
	ChatTypeTicket      ChatType = 2
	ChatTypeProject     ChatType = 3
)

func (c ChatType) ToProto() chatv1.ChatType {
	return chatv1.ChatType(c)
}

func ChatTypeFromProto(p chatv1.ChatType) ChatType {
	return ChatType(p)
}

type Chat struct {
	ID             string
	Type           ChatType
	Title          string
	ParticipantIDs []string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (c *Chat) ToProto() *chatv1.Chat {
	return &chatv1.Chat{
		Id:             c.ID,
		Type:           c.Type.ToProto(),
		Title:          c.Title,
		ParticipantIds: c.ParticipantIDs,
		CreatedAt:      timestamppb.New(c.CreatedAt),
		UpdatedAt:      timestamppb.New(c.UpdatedAt),
	}
}

type Message struct {
	ID        string
	ChatID    string
	AuthorID  string
	Content   string
	CreatedAt time.Time
}

func (m *Message) ToProto() *chatv1.Message {
	return &chatv1.Message{
		Id:        m.ID,
		ChatId:    m.ChatID,
		AuthorId:  m.AuthorID,
		Content:   m.Content,
		CreatedAt: timestamppb.New(m.CreatedAt),
	}
}

type ChatRepository interface {
	Create(ctx context.Context, chat *Chat) error
	GetByID(ctx context.Context, id string) (*Chat, error)
	ListByUser(ctx context.Context, userID string, limit, offset int) ([]*Chat, error)
	AddParticipant(ctx context.Context, chatID, userID string) error
	RemoveParticipant(ctx context.Context, chatID, userID string) error
}

type MessageRepository interface {
	Create(ctx context.Context, msg *Message) error
	ListByChat(ctx context.Context, chatID string, limit, offset int) ([]*Message, error)
}
