// Package service реализует бизнес-логику управления запросами на модерацию и чатом проекта.
package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/Be4Die/game-developer-hub/moderation/internal/domain"
)

// ModerationService координирует операции модерации, чата и взаимодействия с сервисом управления проектами.
type ModerationService struct {
	requestRepo   domain.RequestRepo
	messageRepo   domain.MessageRepo
	projectClient domain.ProjectClient
}

// NewModerationService создаёт новый экземпляр ModerationService.
func NewModerationService(
	requestRepo domain.RequestRepo,
	messageRepo domain.MessageRepo,
	projectClient domain.ProjectClient,
) *ModerationService {
	return &ModerationService{
		requestRepo:   requestRepo,
		messageRepo:   messageRepo,
		projectClient: projectClient,
	}
}

// SubmitDraft создает новый запрос на модерацию со снимком метаданных и фиксирует системное событие в чате.
func (s *ModerationService) SubmitDraft(ctx context.Context, projectID int64, ownerID string, snapshot domain.ProjectSnapshot) (*domain.ModerationRequest, error) {
	req := &domain.ModerationRequest{
		ProjectID: projectID,
		OwnerID:   ownerID,
		Status:    domain.RequestStatusPending,
		Snapshot:  snapshot,
	}

	id, err := s.requestRepo.Create(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ModerationService.SubmitDraft create request: %w", err)
	}
	req.ID = id

	// Запись системного события в чат проекта
	sysMsg := &domain.ChatMessage{
		ProjectID:   projectID,
		RequestID:   &id,
		SenderID:    "system",
		SenderRole:  domain.SenderRoleSystem,
		MessageType: domain.MessageTypeSubmitted,
		Content:     "Черновик отправлен на модерацию",
		Payload: map[string]any{
			"active_build_version": snapshot.ActiveBuildVersion,
			"dev_url":              snapshot.DevURL,
		},
	}
	_, _ = s.messageRepo.Create(ctx, sysMsg)

	return req, nil
}

// ListRequests возвращает постраничный список запросов согласно фильтрам.
func (s *ModerationService) ListRequests(ctx context.Context, filter domain.RequestFilter) ([]*domain.ModerationRequest, int, error) {
	return s.requestRepo.List(ctx, filter)
}

// GetRequest возвращает запрос на модерацию по его ID.
func (s *ModerationService) GetRequest(ctx context.Context, requestID int64) (*domain.ModerationRequest, error) {
	return s.requestRepo.Get(ctx, requestID)
}

// GetLatestRequestByProject возвращает последнюю заявку по проекту.
func (s *ModerationService) GetLatestRequestByProject(ctx context.Context, projectID int64) (*domain.ModerationRequest, error) {
	return s.requestRepo.GetLatestByProject(ctx, projectID)
}

// ClaimRequest закрепляет запрос за модератором и переводит в статус проверки.
func (s *ModerationService) ClaimRequest(ctx context.Context, requestID int64, moderatorID string) (*domain.ModerationRequest, error) {
	if err := s.requestRepo.Claim(ctx, requestID, moderatorID); err != nil {
		return nil, fmt.Errorf("ModerationService.ClaimRequest: %w", err)
	}

	req, err := s.requestRepo.Get(ctx, requestID)
	if err != nil {
		return nil, fmt.Errorf("ModerationService.ClaimRequest get: %w", err)
	}

	// Фиксация системного события в чате
	sysMsg := &domain.ChatMessage{
		ProjectID:   req.ProjectID,
		RequestID:   &req.ID,
		SenderID:    moderatorID,
		SenderRole:  domain.SenderRoleSystem,
		MessageType: domain.MessageTypeStatusChanged,
		Content:     "Модератор взял проект на проверку",
	}
	_, _ = s.messageRepo.Create(ctx, sysMsg)

	return req, nil
}

// Approve утверждает проект, публикует сборку в продуктивное окружение и фиксирует статус.
// При сбое публикации запрос НЕ отклоняется, а остается на проверке с возвратом ошибки.
func (s *ModerationService) Approve(ctx context.Context, projectID int64, moderatorID, comment string) (*domain.ModerationRequest, string, error) {
	req, err := s.requestRepo.GetLatestByProject(ctx, projectID)
	if err != nil {
		return nil, "", fmt.Errorf("ModerationService.Approve get latest request: %w", err)
	}

	if req.Status == domain.RequestStatusApproved {
		return nil, "", domain.ErrInvalidStatus
	}

	// 1. Физическая публикация сборки через Project Manager
	prodURL, err := s.projectClient.PublishRelease(ctx, req.ProjectID, req.Snapshot.ActiveBuildVersion, moderatorID, comment)
	if err != nil {
		// Ошибка инфраструктуры деплоя — не отклоняем проект, сохраняем в IN_REVIEW
		return nil, "", fmt.Errorf("%w: %w", domain.ErrDeployFailed, err)
	}

	// 2. Фиксация вердикта
	if err := s.requestRepo.Resolve(ctx, req.ID, domain.RequestStatusApproved, "", moderatorID); err != nil {
		return nil, "", fmt.Errorf("ModerationService.Approve resolve: %w", err)
	}

	// 3. Системное сообщение в чат
	sysMsg := &domain.ChatMessage{
		ProjectID:   projectID,
		RequestID:   &req.ID,
		SenderID:    moderatorID,
		SenderRole:  domain.SenderRoleSystem,
		MessageType: domain.MessageTypeApproved,
		Content:     "Проект одобрен и опубликован в каталоге",
		Payload: map[string]any{
			"comment":  comment,
			"prod_url": prodURL,
		},
	}
	_, _ = s.messageRepo.Create(ctx, sysMsg)

	if strings.TrimSpace(comment) != "" {
		commentMsg := &domain.ChatMessage{
			ProjectID:   projectID,
			RequestID:   &req.ID,
			SenderID:    moderatorID,
			SenderRole:  domain.SenderRoleModerator,
			MessageType: domain.MessageTypeText,
			Content:     comment,
		}
		_, _ = s.messageRepo.Create(ctx, commentMsg)
	}

	req.Status = domain.RequestStatusApproved
	req.ModeratorID = moderatorID

	return req, prodURL, nil
}

// Reject отклоняет проект с обязательным указанием причины и возвращает черновик разработчику на доработку.
func (s *ModerationService) Reject(ctx context.Context, projectID int64, moderatorID, reason string) (*domain.ModerationRequest, error) {
	cleanReason := strings.TrimSpace(reason)
	if cleanReason == "" {
		return nil, domain.ErrEmptyReason
	}

	req, err := s.requestRepo.GetLatestByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("ModerationService.Reject get latest request: %w", err)
	}

	if req.Status == domain.RequestStatusApproved {
		return nil, domain.ErrInvalidStatus
	}

	// 1. Возврат черновика в Project Manager
	if err := s.projectClient.RejectDraft(ctx, projectID, cleanReason); err != nil {
		return nil, fmt.Errorf("ModerationService.Reject projectClient: %w", err)
	}

	// 2. Фиксация вердикта
	if err := s.requestRepo.Resolve(ctx, req.ID, domain.RequestStatusRejected, cleanReason, moderatorID); err != nil {
		return nil, fmt.Errorf("ModerationService.Reject resolve: %w", err)
	}

	// 3. Системное сообщение в чат
	sysMsg := &domain.ChatMessage{
		ProjectID:   projectID,
		RequestID:   &req.ID,
		SenderID:    moderatorID,
		SenderRole:  domain.SenderRoleSystem,
		MessageType: domain.MessageTypeRejected,
		Content:     fmt.Sprintf("Заявка отклонена: %s", cleanReason),
		Payload: map[string]any{
			"reason": cleanReason,
		},
	}
	_, _ = s.messageRepo.Create(ctx, sysMsg)

	req.Status = domain.RequestStatusRejected
	req.RejectionReason = cleanReason
	req.ModeratorID = moderatorID

	return req, nil
}

// SendMessage отправляет пользовательское или модераторское текстовое сообщение в чат проекта.
func (s *ModerationService) SendMessage(ctx context.Context, projectID int64, senderID string, senderRole domain.SenderRole, content string) (*domain.ChatMessage, error) {
	cleanContent := strings.TrimSpace(content)
	if cleanContent == "" {
		return nil, domain.ErrEmptyMessage
	}

	var reqID *int64
	if latestReq, err := s.requestRepo.GetLatestByProject(ctx, projectID); err == nil {
		reqID = &latestReq.ID
	}

	msg := &domain.ChatMessage{
		ProjectID:   projectID,
		RequestID:   reqID,
		SenderID:    senderID,
		SenderRole:  senderRole,
		MessageType: domain.MessageTypeText,
		Content:     cleanContent,
	}

	id, err := s.messageRepo.Create(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("ModerationService.SendMessage create: %w", err)
	}
	msg.ID = id

	return msg, nil
}

// ListMessages возвращает историю сообщений чата проекта.
func (s *ModerationService) ListMessages(ctx context.Context, projectID int64, limit, offset int) ([]*domain.ChatMessage, int, error) {
	return s.messageRepo.ListByProject(ctx, projectID, limit, offset)
}

// ListActiveChats возвращает активные чаты (последние сообщения по каждому проекту).
func (s *ModerationService) ListActiveChats(ctx context.Context, limit, offset int) ([]*domain.ChatSummary, int, error) {
	return s.messageRepo.ListActiveChats(ctx, limit, offset)
}

// CloseDialog закрывает диалог по проекту и отправляет системное сообщение о решении вопроса.
func (s *ModerationService) CloseDialog(ctx context.Context, projectID int64, moderatorID, comment string) (*domain.ChatMessage, error) {
	var reqID *int64
	if latestReq, err := s.requestRepo.GetLatestByProject(ctx, projectID); err == nil {
		reqID = &latestReq.ID
	}

	content := "Модератор закрыл диалог. Все вопросы решены."
	if strings.TrimSpace(comment) != "" {
		content = fmt.Sprintf("Диалог закрыт: %s", strings.TrimSpace(comment))
	}

	sysMsg := &domain.ChatMessage{
		ProjectID:   projectID,
		RequestID:   reqID,
		SenderID:    moderatorID,
		SenderRole:  domain.SenderRoleSystem,
		MessageType: domain.MessageTypeStatusChanged,
		Content:     content,
		Payload: map[string]any{
			"dialog_status": "closed",
			"closed_by":     moderatorID,
		},
	}

	id, err := s.messageRepo.Create(ctx, sysMsg)
	if err != nil {
		return nil, fmt.Errorf("ModerationService.CloseDialog create: %w", err)
	}
	sysMsg.ID = id

	return sysMsg, nil
}

// GetModeratorStats возвращает агрегированную статистику по конкретному модератору.
func (s *ModerationService) GetModeratorStats(ctx context.Context, moderatorID string) (*domain.ModeratorStats, error) {
	if strings.TrimSpace(moderatorID) == "" {
		return nil, fmt.Errorf("moderator id is required")
	}
	return s.requestRepo.GetModeratorStats(ctx, strings.TrimSpace(moderatorID))
}

// ListModeratorsStats возвращает сводную статистику по всем модераторам.
func (s *ModerationService) ListModeratorsStats(ctx context.Context) ([]*domain.ModeratorStats, error) {
	return s.requestRepo.ListModeratorsStats(ctx)
}

// ListModeratorActivity возвращает постраничный журнал действий конкретного модератора.
func (s *ModerationService) ListModeratorActivity(ctx context.Context, moderatorID string, actionType string, limit, offset int) ([]*domain.ModeratorActivityItem, int, error) {
	if strings.TrimSpace(moderatorID) == "" {
		return nil, 0, fmt.Errorf("moderator id is required")
	}
	return s.requestRepo.ListModeratorActivity(ctx, strings.TrimSpace(moderatorID), strings.TrimSpace(actionType), limit, offset)
}


