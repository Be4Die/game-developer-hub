package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/moderation/internal/domain"
)

func setupTestModerationService(t *testing.T) (*ModerationService, *mockRequestRepo, *mockMessageRepo, *mockProjectClient) {
	t.Helper()
	reqRepo := newMockRequestRepo()
	msgRepo := newMockMessageRepo()
	pmClient := newMockProjectClient()

	svc := NewModerationService(reqRepo, msgRepo, pmClient)
	return svc, reqRepo, msgRepo, pmClient
}

func TestUnit_ModerationService_SubmitDraft(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	svc, reqRepo, msgRepo, _ := setupTestModerationService(t)

	snapshot := domain.ProjectSnapshot{
		ProjectID:          100,
		TitleRu:            "Супер Игра",
		TitleEn:            "Super Game",
		AboutRu:            "Описание игры",
		AboutEn:            "Game description",
		ActiveBuildVersion: "1.0.0",
		DevURL:             "/games/100/dev/index.html",
	}

	req, err := svc.SubmitDraft(ctx, 100, "dev-1", snapshot)
	if err != nil {
		t.Fatalf("expected submit draft success, got error: %v", err)
	}

	if req.ID == 0 {
		t.Errorf("expected non-zero request id")
	}
	if req.Status != domain.RequestStatusPending {
		t.Errorf("expected status PENDING, got: %v", req.Status)
	}

	// Проверяем наличие в репозитории
	saved, err := reqRepo.Get(ctx, req.ID)
	if err != nil || saved.ProjectID != 100 {
		t.Errorf("expected saved request in repo, got: %v, req: %+v", err, saved)
	}

	// Проверяем системное сообщение в чате
	messages, total, err := msgRepo.ListByProject(ctx, 100, 10, 0)
	if err != nil || total != 1 {
		t.Fatalf("expected 1 chat message, got: %d (err: %v)", total, err)
	}
	if messages[0].MessageType != domain.MessageTypeSubmitted {
		t.Errorf("expected MessageTypeSubmitted, got: %v", messages[0].MessageType)
	}
}

func TestUnit_ModerationService_ClaimRequest(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	svc, _, _, _ := setupTestModerationService(t)

	snapshot := domain.ProjectSnapshot{
		ProjectID:          100,
		TitleRu:            "Супер Игра",
		ActiveBuildVersion: "1.0.0",
	}

	req, err := svc.SubmitDraft(ctx, 100, "dev-1", snapshot)
	if err != nil {
		t.Fatalf("submit draft failed: %v", err)
	}

	// 1. Успешный захват модератором
	claimed, err := svc.ClaimRequest(ctx, req.ID, "mod-1")
	if err != nil {
		t.Fatalf("expected claim success, got: %v", err)
	}
	if claimed.Status != domain.RequestStatusInReview {
		t.Errorf("expected status IN_REVIEW, got: %v", claimed.Status)
	}
	if claimed.ModeratorID != "mod-1" {
		t.Errorf("expected moderator_id mod-1, got: %s", claimed.ModeratorID)
	}

	// 2. Попытка захвата другим модератором должна вернуть ErrAlreadyClaimed
	_, err = svc.ClaimRequest(ctx, req.ID, "mod-2")
	if !errors.Is(err, domain.ErrAlreadyClaimed) {
		t.Errorf("expected ErrAlreadyClaimed, got: %v", err)
	}
}

func TestUnit_ModerationService_Approve_Success(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	svc, _, msgRepo, pmClient := setupTestModerationService(t)

	snapshot := domain.ProjectSnapshot{
		ProjectID:          100,
		TitleRu:            "Супер Игра",
		ActiveBuildVersion: "1.0.0",
		DevURL:             "/games/100/dev/index.html",
	}

	_, _ = svc.SubmitDraft(ctx, 100, "dev-1", snapshot)

	approvedReq, prodURL, err := svc.Approve(ctx, 100, "mod-1", "Игра проверена, багов нет")
	if err != nil {
		t.Fatalf("expected approve success, got error: %v", err)
	}

	if approvedReq.Status != domain.RequestStatusApproved {
		t.Errorf("expected status APPROVED, got: %v", approvedReq.Status)
	}
	if prodURL != "/games/100/prod/index.html" {
		t.Errorf("unexpected prodURL: %s", prodURL)
	}

	// Проверяем, что в project-manager зафиксирован релиз
	if pmClient.releases[100] == "" {
		t.Errorf("expected release in project manager client")
	}

	// Проверяем сообщения чата: 1 submit, 1 approve system msg, 1 moderator comment msg
	messages, total, _ := msgRepo.ListByProject(ctx, 100, 10, 0)
	if total != 3 {
		t.Errorf("expected 3 chat messages, got: %d", total)
	}
	if messages[1].MessageType != domain.MessageTypeApproved {
		t.Errorf("expected MessageTypeApproved, got: %v", messages[1].MessageType)
	}
	if messages[2].Content != "Игра проверена, багов нет" {
		t.Errorf("expected comment in chat, got: %s", messages[2].Content)
	}
}

func TestUnit_ModerationService_Approve_DeployFailure(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	svc, reqRepo, _, pmClient := setupTestModerationService(t)

	snapshot := domain.ProjectSnapshot{
		ProjectID:          100,
		TitleRu:            "Супер Игра",
		ActiveBuildVersion: "1.0.0",
	}

	req, _ := svc.SubmitDraft(ctx, 100, "dev-1", snapshot)

	// Эмулируем сбой инфраструктуры деплоера
	pmClient.deployFails = true

	_, _, err := svc.Approve(ctx, 100, "mod-1", "Одобряю")
	if !errors.Is(err, domain.ErrDeployFailed) {
		t.Fatalf("expected ErrDeployFailed, got: %v", err)
	}

	// Критически важно: статус заявки НЕ должен стать REJECTED, так как разработчик не виноват
	currentReq, _ := reqRepo.Get(ctx, req.ID)
	if currentReq.Status == domain.RequestStatusRejected {
		t.Errorf("request must NOT be rejected on deployment infrastructure failure")
	}
}

func TestUnit_ModerationService_Reject_Success(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	svc, _, msgRepo, pmClient := setupTestModerationService(t)

	snapshot := domain.ProjectSnapshot{
		ProjectID:          100,
		TitleRu:            "Супер Игра",
		ActiveBuildVersion: "1.0.0",
	}

	_, _ = svc.SubmitDraft(ctx, 100, "dev-1", snapshot)

	rejectedReq, err := svc.Reject(ctx, 100, "mod-1", "Игра не запускается в Firefox")
	if err != nil {
		t.Fatalf("expected reject success, got error: %v", err)
	}

	if rejectedReq.Status != domain.RequestStatusRejected {
		t.Errorf("expected status REJECTED, got: %v", rejectedReq.Status)
	}
	if rejectedReq.RejectionReason != "Игра не запускается в Firefox" {
		t.Errorf("expected reason to be saved, got: %s", rejectedReq.RejectionReason)
	}

	// Проверяем уведомление project-manager
	if pmClient.rejectedMods[100] != "Игра не запускается в Firefox" {
		t.Errorf("expected project-manager to receive reject with reason")
	}

	// Проверяем сообщения чата
	messages, total, _ := msgRepo.ListByProject(ctx, 100, 10, 0)
	if total != 2 {
		t.Errorf("expected 2 chat messages, got: %d", total)
	}
	if messages[1].MessageType != domain.MessageTypeRejected {
		t.Errorf("expected MessageTypeRejected, got: %v", messages[1].MessageType)
	}
}

func TestUnit_ModerationService_Reject_EmptyReason(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	svc, _, _, _ := setupTestModerationService(t)

	snapshot := domain.ProjectSnapshot{
		ProjectID:          100,
		TitleRu:            "Супер Игра",
		ActiveBuildVersion: "1.0.0",
	}

	_, _ = svc.SubmitDraft(ctx, 100, "dev-1", snapshot)

	_, err := svc.Reject(ctx, 100, "mod-1", "   ")
	if !errors.Is(err, domain.ErrEmptyReason) {
		t.Errorf("expected ErrEmptyReason on empty reason, got: %v", err)
	}
}

func TestUnit_ModerationService_SendMessage_And_ListMessages(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	svc, _, _, _ := setupTestModerationService(t)

	// Отправка пустого сообщения
	_, err := svc.SendMessage(ctx, 100, "dev-1", domain.SenderRoleDeveloper, "  ")
	if !errors.Is(err, domain.ErrEmptyMessage) {
		t.Errorf("expected ErrEmptyMessage, got: %v", err)
	}

	// Отправка валидных сообщений
	msg1, err := svc.SendMessage(ctx, 100, "dev-1", domain.SenderRoleDeveloper, "Здравствуйте, проверите игру?")
	if err != nil || msg1.ID == 0 {
		t.Fatalf("expected send message 1 success: %v", err)
	}

	msg2, err := svc.SendMessage(ctx, 100, "mod-1", domain.SenderRoleModerator, "Здравствуйте! Проверяю.")
	if err != nil || msg2.ID == 0 {
		t.Fatalf("expected send message 2 success: %v", err)
	}

	// Листинг с пагинацией
	msgs, total, err := svc.ListMessages(ctx, 100, 10, 0)
	if err != nil || total != 2 {
		t.Fatalf("expected 2 messages, got: %d (err: %v)", total, err)
	}
	if msgs[0].Content != "Здравствуйте, проверите игру?" {
		t.Errorf("unexpected content msg 0: %s", msgs[0].Content)
	}
	if msgs[1].Content != "Здравствуйте! Проверяю." {
		t.Errorf("unexpected content msg 1: %s", msgs[1].Content)
	}
}
