package app

import (
	"context"

	"github.com/Be4Die/game-developer-hub/chat/internal/domain"
	"github.com/google/uuid"
)

type ChatService struct {
	chatRepo    domain.ChatRepository
	messageRepo domain.MessageRepository
}

func NewChatService(chatRepo domain.ChatRepository, messageRepo domain.MessageRepository) *ChatService {
	return &ChatService{
		chatRepo:    chatRepo,
		messageRepo: messageRepo,
	}
}

func (s *ChatService) CreateChat(ctx context.Context, userID string, chatType domain.ChatType, title string, participantIDs []string) (*domain.Chat, error) {
	if len(participantIDs) == 0 {
		participantIDs = []string{userID}
	}

	found := false
	for _, id := range participantIDs {
		if id == userID {
			found = true
			break
		}
	}
	if !found {
		participantIDs = append(participantIDs, userID)
	}

	chat := &domain.Chat{
		ID:             uuid.New().String(),
		Type:           chatType,
		Title:          title,
		ParticipantIDs: participantIDs,
	}

	if err := s.chatRepo.Create(ctx, chat); err != nil {
		return nil, err
	}

	return chat, nil
}

func (s *ChatService) GetChat(ctx context.Context, userID, chatID string) (*domain.Chat, error) {
	chat, err := s.chatRepo.GetByID(ctx, chatID)
	if err != nil {
		return nil, err
	}

	if !containsUser(chat.ParticipantIDs, userID) {
		return nil, domain.ErrUnauthorized
	}

	return chat, nil
}

func (s *ChatService) ListChats(ctx context.Context, userID string, limit, offset int) ([]*domain.Chat, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.chatRepo.ListByUser(ctx, userID, limit, offset)
}

func (s *ChatService) SendMessage(ctx context.Context, userID, chatID, content string) (*domain.Message, error) {
	chat, err := s.chatRepo.GetByID(ctx, chatID)
	if err != nil {
		return nil, err
	}

	if !containsUser(chat.ParticipantIDs, userID) {
		return nil, domain.ErrUnauthorized
	}

	msg := &domain.Message{
		ID:        uuid.New().String(),
		ChatID:    chatID,
		AuthorID:  userID,
		Content:   content,
	}

	if err := s.messageRepo.Create(ctx, msg); err != nil {
		return nil, err
	}

	return msg, nil
}

func (s *ChatService) GetMessages(ctx context.Context, userID, chatID string, limit, offset int) ([]*domain.Message, error) {
	chat, err := s.chatRepo.GetByID(ctx, chatID)
	if err != nil {
		return nil, err
	}

	if !containsUser(chat.ParticipantIDs, userID) {
		return nil, domain.ErrUnauthorized
	}

	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	return s.messageRepo.ListByChat(ctx, chatID, limit, offset)
}

func (s *ChatService) AddParticipant(ctx context.Context, userID, chatID, newUserID string) (*domain.Chat, error) {
	chat, err := s.chatRepo.GetByID(ctx, chatID)
	if err != nil {
		return nil, err
	}

	if !containsUser(chat.ParticipantIDs, userID) {
		return nil, domain.ErrUnauthorized
	}

	if err := s.chatRepo.AddParticipant(ctx, chatID, newUserID); err != nil {
		return nil, err
	}

	return s.chatRepo.GetByID(ctx, chatID)
}

func (s *ChatService) RemoveParticipant(ctx context.Context, userID, chatID, targetUserID string) (*domain.Chat, error) {
	chat, err := s.chatRepo.GetByID(ctx, chatID)
	if err != nil {
		return nil, err
	}

	if !containsUser(chat.ParticipantIDs, userID) {
		return nil, domain.ErrUnauthorized
	}

	if err := s.chatRepo.RemoveParticipant(ctx, chatID, targetUserID); err != nil {
		return nil, err
	}

	return s.chatRepo.GetByID(ctx, chatID)
}

func containsUser(userIDs []string, userID string) bool {
	for _, id := range userIDs {
		if id == userID {
			return true
		}
	}
	return false
}