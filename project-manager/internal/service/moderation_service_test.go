package service

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
)

func setupTestModerationService(t *testing.T) (*ModerationService, *ProjectService, *mockProjectRepo, *mockModerationRepo, *mockReleaseRepo) {
	t.Helper()
	pRepo := newMockProjectRepo()
	dRepo := newMockDraftRepo()
	bRepo := newMockBuildRepo()
	mRepo := newMockModerationRepo()
	rRepo := newMockReleaseRepo()
	depRepo := newMockDeploymentRepo()
	bStorage := &mockBuildStorage{}
	mStorage := &mockMediaStorage{}
	deployer := &mockDeployer{}

	projectSvc := NewProjectService(
		pRepo, dRepo, bRepo, mRepo, rRepo, depRepo,
		bStorage, mStorage, deployer, nil, 5,
	)

	moderationSvc := NewModerationService(
		pRepo, dRepo, bRepo, mRepo, rRepo, depRepo,
		bStorage, deployer,
	)

	return moderationSvc, projectSvc, pRepo, mRepo, rRepo
}

func TestUnit_ModerationService_Approve(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	modSvc, projSvc, pRepo, mRepo, rRepo := setupTestModerationService(t)

	// Создаем и отправляем проект на модерацию
	p, _ := projSvc.CreateProject(ctx, "user-1", "Супер Игра", "Super Game")
	_ = projSvc.UpdateDraft(ctx, p.ID, "user-1", domain.DraftMeta{
		TitleRu: "Супер Игра",
		About:   "Описание",
	})
	_, _, _ = projSvc.UploadBuildStream(ctx, p.ID, "user-1", "1.0.0", bytes.NewReader([]byte("zip")))
	_, _ = projSvc.SubmitForModeration(ctx, p.ID, "user-1")

	// Модератор одобряет проект
	rel, err := modSvc.Approve(ctx, p.ID, "mod-999", "Отличная игра, одобрено!")
	if err != nil {
		t.Fatalf("expected approve success, got error: %v", err)
	}

	if rel.Version != "1.0.0" {
		t.Errorf("expected release version 1.0.0, got: %s", rel.Version)
	}
	if rel.ProdURL != "/games/1/prod/index.html" {
		t.Errorf("unexpected prod url: %s", rel.ProdURL)
	}

	// Проверяем статус проекта
	updatedProj, _ := pRepo.Get(ctx, p.ID)
	if updatedProj.Status != domain.ProjectStatusPublished {
		t.Errorf("expected project status Published, got: %v", updatedProj.Status)
	}

	// Проверяем активный релиз
	activeRel, err := rRepo.GetActive(ctx, p.ID)
	if err != nil || !activeRel.IsActive {
		t.Errorf("expected active release in repo: %v", err)
	}

	// Проверяем тикет
	ticket, _ := mRepo.GetLatestTicketByProject(ctx, p.ID)
	if ticket.Status != domain.ModerationStatusApproved {
		t.Errorf("expected ticket status Approved, got: %v", ticket.Status)
	}
}

func TestUnit_ModerationService_Reject(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	modSvc, projSvc, pRepo, _, _ := setupTestModerationService(t)

	p, _ := projSvc.CreateProject(ctx, "user-1", "Игра с багом", "Buggy Game")
	_ = projSvc.UpdateDraft(ctx, p.ID, "user-1", domain.DraftMeta{
		TitleRu: "Игра с багом",
		About:   "Описание",
	})
	_, _, _ = projSvc.UploadBuildStream(ctx, p.ID, "user-1", "1.0.0", bytes.NewReader([]byte("zip")))
	_, _ = projSvc.SubmitForModeration(ctx, p.ID, "user-1")

	ticket, err := modSvc.Reject(ctx, p.ID, "mod-999", "Игра крашится при запуске")
	if err != nil {
		t.Fatalf("expected reject success, got: %v", err)
	}

	if ticket.Status != domain.ModerationStatusRejected {
		t.Errorf("expected ticket status Rejected, got: %v", ticket.Status)
	}
	if ticket.RejectionReason != "Игра крашится при запуске" {
		t.Errorf("unexpected rejection reason: %s", ticket.RejectionReason)
	}

	updatedProj, _ := pRepo.Get(ctx, p.ID)
	if updatedProj.Status != domain.ProjectStatusRejected {
		t.Errorf("expected project status Rejected, got: %v", updatedProj.Status)
	}
}
