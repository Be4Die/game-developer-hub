package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/Be4Die/game-developer-hub/chat/internal/domain"
)

type ChatService struct {
	log              *slog.Logger
	conversationRepo domain.ConversationRepository
	messageRepo      domain.MessageRepository
}

func NewChatService(log *slog.Logger, conversationRepo domain.ConversationRepository, messageRepo domain.MessageRepository) *ChatService {
	return &ChatService{
		log:              log,
		conversationRepo: conversationRepo,
		messageRepo:      messageRepo,
	}
}

func (s *ChatService) Log() *slog.Logger {
	return s.log
}

func (s *ChatService) SendMessage(ctx context.Context, conversationID, senderID, senderName, senderRole, content string) (*domain.Message, error) {
	msg := &domain.Message{
		ConversationID: conversationID,
		SenderID:       senderID,
		SenderName:     senderName,
		SenderRole:     senderRole,
		Content:        content,
		IsRead:         false,
	}

	if err := s.messageRepo.Create(ctx, msg); err != nil {
		s.log.Error("failed to create message", slog.String("error", err.Error()), slog.String("senderRole", msg.SenderRole))
		return nil, err
	}

	s.log.Info("message created", slog.String("id", msg.ID), slog.String("senderRole", msg.SenderRole))

	conv, err := s.conversationRepo.GetByID(ctx, conversationID)
	if err == nil {
		conv.LastMessage = content
		conv.LastMessageAt = msg.CreatedAt
		// Увеличиваем счетчик непрочитанных для получателя
		if conv.UserID != senderID {
			conv.UnreadCount++
		}
		s.conversationRepo.Update(ctx, conv)
	}

	return msg, nil
}

func (s *ChatService) GetConversations(ctx context.Context, userID string) ([]domain.Conversation, error) {
	conversations, err := s.conversationRepo.ListByUser(ctx, userID)
	if err != nil {
		s.log.Error("failed to get conversations", slog.String("error", err.Error()))
		return nil, err
	}
	return conversations, nil
}

func (s *ChatService) GetMessages(ctx context.Context, conversationID string, limit, offset int) ([]domain.Message, int, error) {
	if limit <= 0 {
		limit = 50
	}
	messages, total, err := s.messageRepo.ListByConversation(ctx, conversationID, limit, offset)
	if err != nil {
		s.log.Error("failed to get messages", slog.String("error", err.Error()))
		return nil, 0, err
	}
	return messages, total, nil
}

func (s *ChatService) MarkAsRead(ctx context.Context, conversationID, userID string) error {
	if err := s.messageRepo.MarkAsRead(ctx, conversationID, userID); err != nil {
		s.log.Error("failed to mark as read", slog.String("error", err.Error()))
		return err
	}
	conv, err := s.conversationRepo.GetByID(ctx, conversationID)
	if err == nil {
		conv.UnreadCount = 0
		s.conversationRepo.Update(ctx, conv)
	}
	return nil
}

func (s *ChatService) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	count, err := s.messageRepo.CountUnread(ctx, userID)
	if err != nil {
		s.log.Error("failed to get unread count", slog.String("error", err.Error()))
		return 0, err
	}
	return count, nil
}

func (s *ChatService) CreateConversation(ctx context.Context, userID, userName, participantID, participantName string) (*domain.Conversation, error) {
	existing, err := s.conversationRepo.GetByParticipants(ctx, userID, participantID)
	if err == nil && existing != nil {
		return existing, nil
	}

	conv := &domain.Conversation{
		UserID:          userID,
		ParticipantID:   participantID,
		ParticipantName: participantName,
		LastMessage:     "",
		LastMessageAt:   time.Now(),
		UnreadCount:     0,
	}

	if err := s.conversationRepo.Create(ctx, conv); err != nil {
		s.log.Error("failed to create conversation", slog.String("error", err.Error()))
		return nil, err
	}
	return conv, nil
}
