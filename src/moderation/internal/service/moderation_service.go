// Package service реализует бизнес-логику управления запросами на модерацию и чатом проекта.
package service

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Be4Die/game-developer-hub/moderation/internal/domain"
)

// ModerationService координирует операции модерации, чата и взаимодействия с сервисом управления проектами.
type ModerationService struct {
	requestRepo    domain.RequestRepo
	messageRepo    domain.MessageRepo
	attachmentRepo domain.AttachmentRepo
	snapshotRepo   domain.SnapshotRepo
	projectClient  domain.ProjectClient
	storagePath    string
}

// NewModerationService создаёт новый экземпляр ModerationService.
func NewModerationService(
	requestRepo domain.RequestRepo,
	messageRepo domain.MessageRepo,
	attachmentRepo domain.AttachmentRepo,
	projectClient domain.ProjectClient,
	snapshotRepos ...domain.SnapshotRepo,
) *ModerationService {
	var snapRepo domain.SnapshotRepo
	if len(snapshotRepos) > 0 {
		snapRepo = snapshotRepos[0]
	}
	return &ModerationService{
		requestRepo:    requestRepo,
		messageRepo:    messageRepo,
		attachmentRepo: attachmentRepo,
		snapshotRepo:   snapRepo,
		projectClient:  projectClient,
	}
}

// SetStoragePath задает корневую директорию для чтения медиа-материалов проектов.
func (s *ModerationService) SetStoragePath(path string) {
	s.storagePath = path
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

	// Создание неизменяемого аудит-снимка решения перед очисткой временных файлов
	_, _ = s.buildAndSaveSnapshot(ctx, req, domain.RequestStatusApproved, moderatorID, "", nil, comment, prodURL)

	if s.attachmentRepo != nil {
		_, _, _ = s.attachmentRepo.PurgeByProjectID(ctx, projectID)
	}

	req.Status = domain.RequestStatusApproved
	req.ModeratorID = moderatorID

	return req, prodURL, nil
}

// Reject отклоняет проект с обязательным указанием причины или списка нарушений и возвращает черновик разработчику на доработку.
func (s *ModerationService) Reject(ctx context.Context, projectID int64, moderatorID, reason string, violations []*domain.ViolationItem) (*domain.ModerationRequest, error) {
	cleanReason := strings.TrimSpace(reason)
	if cleanReason == "" && len(violations) == 0 {
		return nil, domain.ErrEmptyReason
	}

	req, err := s.requestRepo.GetLatestByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("ModerationService.Reject get latest request: %w", err)
	}

	if req.Status == domain.RequestStatusApproved {
		return nil, domain.ErrInvalidStatus
	}

	// Формирование итогового текстового описания вердикта (Markdown для content)
	var mdBuilder strings.Builder
	if cleanReason != "" {
		mdBuilder.WriteString(cleanReason)
	} else {
		mdBuilder.WriteString("Обнаружены нарушения правил публикации игры:")
	}

	var allAttachmentIDs []string
	if len(violations) > 0 {
		mdBuilder.WriteString("\n\n")
		for i, v := range violations {
			if strings.TrimSpace(v.RuleCode) != "" && strings.TrimSpace(v.RuleTitle) != "" {
				mdBuilder.WriteString(fmt.Sprintf("%d. Нарушение правила %s \"%s\": %s\n", i+1, v.RuleCode, v.RuleTitle, v.Description))
			} else if strings.TrimSpace(v.RuleCode) != "" {
				mdBuilder.WriteString(fmt.Sprintf("%d. Нарушение правила %s: %s\n", i+1, v.RuleCode, v.Description))
			} else {
				mdBuilder.WriteString(fmt.Sprintf("%d. Замечание: %s\n", i+1, v.Description))
			}
			for _, attID := range v.AttachmentIDs {
				if attID != "" {
					allAttachmentIDs = append(allAttachmentIDs, attID)
				}
			}
		}
	}
	content := strings.TrimSpace(mdBuilder.String())
	effectiveReason := cleanReason
	if effectiveReason == "" {
		effectiveReason = content
	}

	// 1. Возврат черновика в Project Manager
	if err := s.projectClient.RejectDraft(ctx, projectID, effectiveReason); err != nil {
		return nil, fmt.Errorf("ModerationService.Reject projectClient: %w", err)
	}

	// 2. Фиксация вердикта
	if err := s.requestRepo.Resolve(ctx, req.ID, domain.RequestStatusRejected, effectiveReason, moderatorID); err != nil {
		return nil, fmt.Errorf("ModerationService.Reject resolve: %w", err)
	}

	// 3. Загрузка вложений для структурированного JSON
	payload := map[string]any{
		"reason": effectiveReason,
	}
	if len(violations) > 0 {
		payload["type"] = "moderation_verdict"
		if s.attachmentRepo != nil && len(allAttachmentIDs) > 0 {
			if atts, err := s.attachmentRepo.GetByIDs(ctx, allAttachmentIDs); err == nil {
				attMap := make(map[string]*domain.Attachment)
				for _, a := range atts {
					attMap[a.ID] = a
				}
				for _, v := range violations {
					for _, attID := range v.AttachmentIDs {
						if a, ok := attMap[attID]; ok {
							v.Attachments = append(v.Attachments, a)
						}
					}
				}
			}
		}
		payload["violations"] = violations
	}

	// 4. Системное сообщение в чат
	sysMsg := &domain.ChatMessage{
		ProjectID:   projectID,
		RequestID:   &req.ID,
		SenderID:    moderatorID,
		SenderRole:  domain.SenderRoleSystem,
		MessageType: domain.MessageTypeRejected,
		Content:     fmt.Sprintf("Заявка отклонена: %s", content),
		Payload:     payload,
	}
	msgID, _ := s.messageRepo.Create(ctx, sysMsg)
	if msgID > 0 && s.attachmentRepo != nil && len(allAttachmentIDs) > 0 {
		_ = s.attachmentRepo.BindToMessage(ctx, allAttachmentIDs, msgID)
		if atts, err := s.attachmentRepo.GetByIDs(ctx, allAttachmentIDs); err == nil {
			sysMsg.Attachments = atts
		}
	}

	req.Status = domain.RequestStatusRejected
	req.RejectionReason = effectiveReason
	req.ModeratorID = moderatorID

	// Создание неизменяемого аудит-снимка решения
	_, _ = s.buildAndSaveSnapshot(ctx, req, domain.RequestStatusRejected, moderatorID, effectiveReason, violations, "", "")

	return req, nil
}


// SendMessage отправляет пользовательское или модераторское сообщение в чат проекта.
func (s *ModerationService) SendMessage(ctx context.Context, projectID int64, senderID string, senderRole domain.SenderRole, content string, attachmentIDs []string, payload map[string]any) (*domain.ChatMessage, error) {
	cleanContent := strings.TrimSpace(content)
	if cleanContent == "" && len(attachmentIDs) == 0 {
		return nil, domain.ErrEmptyMessage
	}

	var reqID *int64
	if latestReq, err := s.requestRepo.GetLatestByProject(ctx, projectID); err == nil {
		reqID = &latestReq.ID
	}

	if payload == nil {
		payload = make(map[string]any)
	}

	msg := &domain.ChatMessage{
		ProjectID:   projectID,
		RequestID:   reqID,
		SenderID:    senderID,
		SenderRole:  senderRole,
		MessageType: domain.MessageTypeText,
		Content:     cleanContent,
		Payload:     payload,
	}

	id, err := s.messageRepo.Create(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("ModerationService.SendMessage create: %w", err)
	}
	msg.ID = id

	if len(attachmentIDs) > 0 && s.attachmentRepo != nil {
		_ = s.attachmentRepo.BindToMessage(ctx, attachmentIDs, id)
		if atts, err := s.attachmentRepo.GetByIDs(ctx, attachmentIDs); err == nil {
			msg.Attachments = atts
		}
	}

	return msg, nil
}

// ListMessages возвращает историю сообщений чата проекта с вложениями.
func (s *ModerationService) ListMessages(ctx context.Context, projectID int64, limit, offset int) ([]*domain.ChatMessage, int, error) {
	messages, total, err := s.messageRepo.ListByProject(ctx, projectID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	if len(messages) > 0 && s.attachmentRepo != nil {
		msgIDs := make([]int64, len(messages))
		for i, m := range messages {
			msgIDs[i] = m.ID
		}
		attMap, err := s.attachmentRepo.ListByMessageIDs(ctx, msgIDs)
		if err == nil {
			for _, m := range messages {
				if atts, ok := attMap[m.ID]; ok {
					m.Attachments = atts
				}
			}
		}
	}

	return messages, total, nil
}

// RegisterAttachment регистрирует загруженное медиа-вложение в системе.
func (s *ModerationService) RegisterAttachment(ctx context.Context, att *domain.Attachment) error {
	if s.attachmentRepo == nil {
		return nil
	}
	return s.attachmentRepo.Create(ctx, att)
}

// GetAttachment возвращает метаданные вложения по ID и ID проекта.
func (s *ModerationService) GetAttachment(ctx context.Context, id string, projectID int64) (*domain.Attachment, error) {
	if s.attachmentRepo == nil {
		return nil, domain.ErrNotFound
	}
	return s.attachmentRepo.Get(ctx, id, projectID)
}

// PurgeProjectMedia помечает медиафайлы проекта как очищенные и возвращает список путей файлов для удаления.
func (s *ModerationService) PurgeProjectMedia(ctx context.Context, projectID int64) (int, []string, error) {
	if s.attachmentRepo == nil {
		return 0, nil, nil
	}
	return s.attachmentRepo.PurgeByProjectID(ctx, projectID)
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

// GetSnapshot возвращает неизменяемый аудит-снимок по ID заявки на модерацию.
func (s *ModerationService) GetSnapshot(ctx context.Context, requestID int64) (*domain.ModerationSnapshot, error) {
	if s.snapshotRepo != nil {
		snap, err := s.snapshotRepo.GetByRequestID(ctx, requestID)
		if err == nil && snap != nil {
			return snap, nil
		}
	}

	// Fallback для ранее созданных заявок: синтезируем снимок на лету
	req, err := s.requestRepo.Get(ctx, requestID)
	if err != nil {
		return nil, fmt.Errorf("ModerationService.GetSnapshot request not found: %w", err)
	}

	return s.buildAndSaveSnapshot(ctx, req, req.Status, req.ModeratorID, req.RejectionReason, nil, "", "")
}

// buildAndSaveSnapshot формирует полное состояние снимка, упаковывает в JSON и сжимает gzip в бинарный блоб.
func (s *ModerationService) buildAndSaveSnapshot(
	ctx context.Context,
	req *domain.ModerationRequest,
	status domain.RequestStatus,
	moderatorID, reason string,
	violations []*domain.ViolationItem,
	comment, prodURL string,
) (*domain.ModerationSnapshot, error) {
	if s.snapshotRepo == nil {
		return nil, nil
	}

	// 1. Извлечение сообщений чата и вложений
	messages, _, _ := s.messageRepo.ListByProject(ctx, req.ProjectID, 500, 0)
	var snapshotMessages []*domain.SnapshotMessageItem
	if len(messages) > 0 && s.attachmentRepo != nil {
		msgIDs := make([]int64, len(messages))
		for i, m := range messages {
			msgIDs[i] = m.ID
		}
		attMap, _ := s.attachmentRepo.ListByMessageIDs(ctx, msgIDs)
		for _, m := range messages {
			var attItems []*domain.SnapshotAttachmentItem
			if atts, ok := attMap[m.ID]; ok {
				for _, a := range atts {
					attItem := &domain.SnapshotAttachmentItem{
						ID:          a.ID,
						FileName:    a.FileName,
						FileSize:    a.FileSize,
						MimeType:    a.MimeType,
						URL:         a.StoragePath,
						DownloadURL: a.StoragePath,
					}
					s.enrichAttachmentFootprint(req.ProjectID, attItem, a.StoragePath)
					attItems = append(attItems, attItem)
				}
			}
			senderName := m.SenderID
			if m.SenderRole == domain.SenderRoleSystem {
				senderName = "Система"
			}
			snapshotMessages = append(snapshotMessages, &domain.SnapshotMessageItem{
				ID:          m.ID,
				SenderID:    m.SenderID,
				SenderRole:  m.SenderRole,
				SenderName:  senderName,
				MessageType: m.MessageType,
				Content:     m.Content,
				CreatedAt:   m.CreatedAt.Format(time.RFC3339),
				Payload:     m.Payload,
				Attachments: attItems,
			})
		}
	}

	// 2. Сбор нарушений регламента
	var snapshotViolations []*domain.SnapshotViolationItem
	for _, v := range violations {
		var attItems []*domain.SnapshotAttachmentItem
		for _, a := range v.Attachments {
			attItem := &domain.SnapshotAttachmentItem{
				ID:          a.ID,
				FileName:    a.FileName,
				FileSize:    a.FileSize,
				MimeType:    a.MimeType,
				URL:         a.StoragePath,
				DownloadURL: a.StoragePath,
			}
			s.enrichAttachmentFootprint(req.ProjectID, attItem, a.StoragePath)
			attItems = append(attItems, attItem)
		}
		snapshotViolations = append(snapshotViolations, &domain.SnapshotViolationItem{
			RuleCode:    v.RuleCode,
			RuleTitle:   v.RuleTitle,
			Description: v.Description,
			Attachments: attItems,
		})
	}

	// 3. Вычисление времени проверки
	now := time.Now().UTC()
	durationStr := "—"
	var startedReviewStr string
	if req.StartedReviewAt != nil {
		startedReviewStr = req.StartedReviewAt.Format(time.RFC3339)
		d := now.Sub(*req.StartedReviewAt)
		if d >= 24*time.Hour {
			days := int(d.Hours() / 24)
			hours := int(d.Hours()) % 24
			if hours > 0 {
				durationStr = fmt.Sprintf("%d д %d ч", days, hours)
			} else {
				durationStr = fmt.Sprintf("%d д", days)
			}
		} else if d >= time.Hour {
			hours := int(d.Hours())
			mins := int(d.Minutes()) % 60
			if mins > 0 {
				durationStr = fmt.Sprintf("%d ч %d мин", hours, mins)
			} else {
				durationStr = fmt.Sprintf("%d ч", hours)
			}
		} else if d > 0 {
			durationStr = fmt.Sprintf("%d мин", int(d.Minutes()))
		}
	}

	var resolvedStr string
	if req.ResolvedAt != nil {
		resolvedStr = req.ResolvedAt.Format(time.RFC3339)
	} else {
		resolvedStr = now.Format(time.RFC3339)
	}

	devName := req.OwnerID

	// 4. Обогащение медиа-материалов визуальными отпечатками и контрольными суммами
	iconItem := s.resolveMediaItem(req.ProjectID, "icon", req.Snapshot.IconPath)
	coverItem := s.resolveMediaItem(req.ProjectID, "cover", req.Snapshot.CoverPath)
	videoItem := s.resolveMediaItem(req.ProjectID, "video", req.Snapshot.VideoPath)

	payload := domain.SnapshotPayload{
		Project: domain.SnapshotProjectData{
			ID:            req.ProjectID,
			OwnerID:       req.OwnerID,
			DeveloperName: devName,
			TitleRu:       req.Snapshot.TitleRu,
			TitleEn:       req.Snapshot.TitleEn,
			SeoRu:         req.Snapshot.SeoRu,
			SeoEn:         req.Snapshot.SeoEn,
			AboutRu:       req.Snapshot.AboutRu,
			AboutEn:       req.Snapshot.AboutEn,
			BuildVersion:  req.Snapshot.ActiveBuildVersion,
			DevURL:        req.Snapshot.DevURL,
			ProdURL:       prodURL,
			IsOnline:      req.Snapshot.IsOnline,
		},
		Media: domain.SnapshotMediaData{
			Icon:  iconItem,
			Cover: coverItem,
			Video: videoItem,
		},
		Verdict: domain.SnapshotVerdictData{
			Status:          status,
			ModeratorID:     moderatorID,
			ModeratorName:   moderatorID,
			SubmittedAt:     req.SubmittedAt.Format(time.RFC3339),
			StartedReviewAt: startedReviewStr,
			ResolvedAt:      resolvedStr,
			ReviewDuration:  durationStr,
			RejectionReason: reason,
			Comment:         comment,
			ProdURL:         prodURL,
			Violations:      snapshotViolations,
		},
		ChatTranscript: snapshotMessages,
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal snapshot payload: %w", err)
	}

	// Gzip сжатие в бинарный блоб
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write(jsonBytes); err != nil {
		return nil, fmt.Errorf("gzip compress snapshot: %w", err)
	}
	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("gzip close: %w", err)
	}

	snap := &domain.ModerationSnapshot{
		RequestID:     req.ID,
		ProjectID:     req.ProjectID,
		Status:        status,
		FormatVersion: 1,
		SnapshotJSON:  string(jsonBytes),
		SnapshotBlob:  buf.Bytes(),
	}

	if err := s.snapshotRepo.Save(ctx, snap); err != nil {
		return nil, fmt.Errorf("save snapshot: %w", err)
	}
	return snap, nil
}

// resolveMediaItem строит визуальный слепок медиа-файла с диска или fallback на переданный путь.
func (s *ModerationService) resolveMediaItem(projectID int64, mediaType, rawPath string) domain.SnapshotMediaItem {
	item := domain.SnapshotMediaItem{
		FileName:    mediaType + ".png",
		MimeType:    "image/png",
		OriginalURL: rawPath,
	}
	if mediaType == "video" {
		item.FileName = "video.mp4"
		item.MimeType = "video/mp4"
	}

	if s.storagePath == "" {
		return item
	}

	candidatePaths := []string{
		filepath.Join(s.storagePath, fmt.Sprintf("%d", projectID), item.FileName),
		filepath.Join(s.storagePath, strings.TrimPrefix(rawPath, "/media/projects/")),
		filepath.Join(s.storagePath, strings.TrimPrefix(rawPath, "/media/")),
		filepath.Join(s.storagePath, strings.TrimPrefix(rawPath, "/data/projects/")),
	}

	for _, cp := range candidatePaths {
		if fp, err := GenerateMediaFootprint(cp, 260); err == nil && fp != nil {
			item.FileName = fp.FileName
			item.FileSize = fp.FileSize
			item.MimeType = fp.MimeType
			item.Sha256 = fp.Sha256
			item.ThumbnailData = fp.ThumbnailData
			item.Width = fp.Width
			item.Height = fp.Height
			break
		}
	}
	return item
}

// enrichAttachmentFootprint обогащает метаданные вложения превью и sha256 хэшем из файла на диске.
func (s *ModerationService) enrichAttachmentFootprint(projectID int64, attItem *domain.SnapshotAttachmentItem, rawPath string) {
	if s.storagePath == "" || rawPath == "" {
		return
	}
	candidatePaths := []string{
		filepath.Join(s.storagePath, "chat_attachments", fmt.Sprintf("%d", projectID), attItem.FileName),
		filepath.Join(s.storagePath, strings.TrimPrefix(rawPath, "/")),
		filepath.Join(os.TempDir(), "chat_attachments", fmt.Sprintf("%d", projectID), attItem.FileName),
	}
	for _, cp := range candidatePaths {
		if fp, err := GenerateMediaFootprint(cp, 260); err == nil && fp != nil {
			attItem.Sha256 = fp.Sha256
			if fp.FileSize > 0 {
				attItem.FileSize = fp.FileSize
			}
			if fp.ThumbnailData != "" {
				attItem.ThumbnailData = fp.ThumbnailData
			}
			break
		}
	}
}



