package service

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
)

func setupTestProjectService(t *testing.T) (*ProjectService, *mockProjectRepo, *mockDraftRepo, *mockBuildRepo, *mockModerationClient, *mockReleaseRepo) {
	t.Helper()
	pRepo := newMockProjectRepo()
	dRepo := newMockDraftRepo()
	bRepo := newMockBuildRepo()
	mClient := newMockModerationClient()
	rRepo := newMockReleaseRepo()
	depRepo := newMockDeploymentRepo()
	bStorage := &mockBuildStorage{}
	mStorage := &mockMediaStorage{}
	deployer := &mockDeployer{}

	svc := NewProjectService(
		pRepo, dRepo, bRepo, rRepo, depRepo, mClient,
		bStorage, mStorage, deployer, nil, 5,
	)

	return svc, pRepo, dRepo, bRepo, mClient, rRepo
}

func TestUnit_ProjectService_CreateProject(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	svc, _, _, _, _, _ := setupTestProjectService(t)

	p, err := svc.CreateProject(ctx, "user-123", "Тестовая игра", "Test Game")
	if err != nil {
		t.Fatalf("expected create success, got error: %v", err)
	}

	if p.ID == 0 {
		t.Errorf("expected non-zero project id")
	}
	if p.OwnerID != "user-123" {
		t.Errorf("expected owner_id user-123, got: %s", p.OwnerID)
	}
	if p.Draft == nil || p.Draft.TitleRu != "Тестовая игра" {
		t.Errorf("expected draft with TitleRu 'Тестовая игра', got: %+v", p.Draft)
	}
}

func TestUnit_ProjectService_UpdateDraft(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	svc, _, _, _, _, _ := setupTestProjectService(t)

	p, _ := svc.CreateProject(ctx, "user-123", "Игра", "Game")

	meta := domain.DraftMeta{
		TitleRu: "Обновленная игра",
		TitleEn: "Updated Game",
		About:   "Описание игры",
		SeoRu:   "seo ru",
	}

	err := svc.UpdateDraft(ctx, p.ID, "user-123", meta)
	if err != nil {
		t.Fatalf("expected update success, got: %v", err)
	}

	updated, err := svc.GetProject(ctx, p.ID)
	if err != nil {
		t.Fatalf("get project error: %v", err)
	}
	if updated.Draft.TitleRu != "Обновленная игра" || updated.Draft.About != "Описание игры" {
		t.Errorf("draft not updated properly: %+v", updated.Draft)
	}

	// Проверка прав доступа
	err = svc.UpdateDraft(ctx, p.ID, "another-user", meta)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden for another user, got: %v", err)
	}
}

func TestUnit_ProjectService_UploadBuildStream(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	svc, _, _, _, _, _ := setupTestProjectService(t)

	p, _ := svc.CreateProject(ctx, "user-123", "Игра", "Game")

	archiveData := []byte("dummy zip content")
	b, devURL, err := svc.UploadBuildStream(ctx, p.ID, "user-123", "1.0.0", bytes.NewReader(archiveData))
	if err != nil {
		t.Fatalf("upload build failed: %v", err)
	}

	if b.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got: %s", b.Version)
	}
	if devURL != "/games/1/dev/index.html" {
		t.Errorf("unexpected devURL: %s", devURL)
	}

	// Повторная загрузка той же версии должна вернуть ErrAlreadyExists
	_, _, err = svc.UploadBuildStream(ctx, p.ID, "user-123", "1.0.0", bytes.NewReader(archiveData))
	if !errors.Is(err, domain.ErrAlreadyExists) {
		t.Errorf("expected ErrAlreadyExists on duplicate version, got: %v", err)
	}
}

func TestUnit_ProjectService_SubmitForModeration(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	svc, pRepo, _, _, _, _ := setupTestProjectService(t)

	p, _ := svc.CreateProject(ctx, "user-123", "Игра", "Game")

	// 1. Попытка отправки неполного черновика (нет описания и билда)
	_, err := svc.SubmitForModeration(ctx, p.ID, "user-123")
	if !errors.Is(err, domain.ErrDraftNotReady) {
		t.Errorf("expected ErrDraftNotReady, got: %v", err)
	}

	// 2. Заполняем черновик и загружаем билд
	_ = svc.UpdateDraft(ctx, p.ID, "user-123", domain.DraftMeta{
		TitleRu: "Игра",
		About:   "Описание игры",
	})
	_, _, _ = svc.UploadBuildStream(ctx, p.ID, "user-123", "1.0.0", bytes.NewReader([]byte("zip")))

	// 3. Отправляем на модерацию
	reqID, err := svc.SubmitForModeration(ctx, p.ID, "user-123")
	if err != nil {
		t.Fatalf("expected successful submit, got: %v", err)
	}
	if reqID == 0 {
		t.Errorf("expected non-zero request id, got: %d", reqID)
	}

	// Проверяем статус проекта
	updatedProj, _ := pRepo.Get(ctx, p.ID)
	if updatedProj.Status != domain.ProjectStatusPending {
		t.Errorf("expected project status Pending, got: %v", updatedProj.Status)
	}

	// 4. Повторная отправка должна вернуть ErrAlreadyInModeration
	_, err = svc.SubmitForModeration(ctx, p.ID, "user-123")
	if !errors.Is(err, domain.ErrAlreadyInModeration) {
		t.Errorf("expected ErrAlreadyInModeration, got: %v", err)
	}
}

func TestUnit_ProjectService_PublishRelease(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	svc, pRepo, _, _, _, rRepo := setupTestProjectService(t)

	p, _ := svc.CreateProject(ctx, "user-123", "Игра", "Game")
	_ = svc.UpdateDraft(ctx, p.ID, "user-123", domain.DraftMeta{
		TitleRu: "Игра",
		About:   "Описание игры",
	})
	_, _, _ = svc.UploadBuildStream(ctx, p.ID, "user-123", "1.0.0", bytes.NewReader([]byte("zip")))

	rel, err := svc.PublishRelease(ctx, p.ID, "1.0.0", "mod-999")
	if err != nil {
		t.Fatalf("expected publish release success, got error: %v", err)
	}
	if rel.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got: %s", rel.Version)
	}
	if rel.PublishedBy != "mod-999" {
		t.Errorf("expected published by mod-999, got: %s", rel.PublishedBy)
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
}

func TestUnit_ProjectService_RejectDraft(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	svc, pRepo, _, _, _, _ := setupTestProjectService(t)

	p, _ := svc.CreateProject(ctx, "user-123", "Игра", "Game")
	_ = pRepo.UpdateStatus(ctx, p.ID, domain.ProjectStatusPending)

	err := svc.RejectDraft(ctx, p.ID)
	if err != nil {
		t.Fatalf("expected reject success, got: %v", err)
	}

	updatedProj, _ := pRepo.Get(ctx, p.ID)
	if updatedProj.Status != domain.ProjectStatusDraft {
		t.Errorf("expected project status Draft after reject, got: %v", updatedProj.Status)
	}
}
