package service

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestProjectService(t *testing.T) (*ProjectService, *mockProjectRepo, *mockReleaseRepo) {
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

	return svc, pRepo, rRepo
}

func TestUnit_ProjectService_CreateProject(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	svc, _, _ := setupTestProjectService(t)

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

	svc, _, _ := setupTestProjectService(t)

	p, _ := svc.CreateProject(ctx, "user-123", "Игра", "Game")

	meta := domain.DraftMeta{
		TitleRu: "Обновленная игра",
		TitleEn: "Updated Game",
		AboutRu: "Описание игры",
		AboutEn: "Game description",
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
	if updated.Draft.TitleRu != "Обновленная игра" || updated.Draft.AboutRu != "Описание игры" || updated.Draft.AboutEn != "Game description" {
		t.Errorf("draft not updated properly: %+v", updated.Draft)
	}

	// Проверка прав доступа
	err = svc.UpdateDraft(ctx, p.ID, "another-user", meta)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden for another user, got: %v", err)
	}

	// Проверка валидации лимита длины (название > 50)
	invalidMeta := meta
	invalidMeta.TitleRu = "123456789012345678901234567890123456789012345678901" // 51 chars
	if err := svc.UpdateDraft(ctx, p.ID, "user-123", invalidMeta); !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for TitleRu > 50 chars, got: %v", err)
	}
}

func TestUnit_ProjectService_UploadBuildStream(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	svc, _, _ := setupTestProjectService(t)

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

	svc, pRepo, _ := setupTestProjectService(t)

	p, _ := svc.CreateProject(ctx, "user-123", "Игра", "Game")

	// 1. Попытка отправки неполного черновика (нет описания и билда)
	_, err := svc.SubmitForModeration(ctx, p.ID, "user-123")
	if !errors.Is(err, domain.ErrDraftNotReady) {
		t.Errorf("expected ErrDraftNotReady, got: %v", err)
	}

	// 2. Заполняем черновик и загружаем билд
	_ = svc.UpdateDraft(ctx, p.ID, "user-123", domain.DraftMeta{
		TitleRu: "Игра",
		AboutRu: "Описание игры",
		AboutEn: "Game description",
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

	svc, pRepo, rRepo := setupTestProjectService(t)

	p, _ := svc.CreateProject(ctx, "user-123", "Игра", "Game")
	_ = svc.UpdateDraft(ctx, p.ID, "user-123", domain.DraftMeta{
		TitleRu: "Игра",
		AboutRu: "Описание игры",
		AboutEn: "Game description",
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

	svc, pRepo, _ := setupTestProjectService(t)

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

func TestUnit_ProjectService_ListProjectsAndBuilds(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	svc, _, _ := setupTestProjectService(t)

	p1, _ := svc.CreateProject(ctx, "user-abc", "Игра 1", "Game 1")
	p2, _ := svc.CreateProject(ctx, "user-abc", "Игра 2", "Game 2")
	_, _ = svc.CreateProject(ctx, "user-other", "Игра 3", "Game 3")

	// List user-abc projects
	list, total, err := svc.ListProjects(ctx, "user-abc", 10, 0)
	if err != nil {
		t.Fatalf("list projects failed: %v", err)
	}
	if total != 2 || len(list) != 2 {
		t.Errorf("expected 2 projects, got total=%d, len=%d", total, len(list))
	}

	// Upload builds to p1
	_, _, _ = svc.UploadBuildStream(ctx, p1.ID, "user-abc", "1.0.0", bytes.NewReader([]byte("b1")))
	_, _, _ = svc.UploadBuildStream(ctx, p1.ID, "user-abc", "1.0.1", bytes.NewReader([]byte("b2")))

	builds, err := svc.ListBuilds(ctx, p1.ID, "user-abc")
	if err != nil {
		t.Fatalf("list builds failed: %v", err)
	}
	if len(builds) != 2 {
		t.Errorf("expected 2 builds, got %d", len(builds))
	}

	// List builds for other user should be forbidden
	_, err = svc.ListBuilds(ctx, p1.ID, "user-intruder")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden for non-owner listing builds, got: %v", err)
	}

	_ = p2
}

func TestUnit_ProjectService_MediaManagement(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	svc, _, _ := setupTestProjectService(t)

	p, _ := svc.CreateProject(ctx, "user-123", "Игра", "Game")

	// 1. Upload Icon
	iconPath, err := svc.UploadMediaStream(ctx, p.ID, "user-123", "icon", bytes.NewReader([]byte("icon-png")))
	if err != nil {
		t.Fatalf("upload icon failed: %v", err)
	}
	if iconPath == "" {
		t.Errorf("expected non-empty icon path")
	}

	// 2. Forbidden upload by non-owner
	_, err = svc.UploadMediaStream(ctx, p.ID, "user-intruder", "icon", bytes.NewReader([]byte("icon-png")))
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden on media upload by non-owner, got: %v", err)
	}

	// 3. Publish and Unpublish
	_ = svc.UpdateDraft(ctx, p.ID, "user-123", domain.DraftMeta{
		TitleRu: "Игра",
		AboutRu: "Описание",
	})
	_, _, _ = svc.UploadBuildStream(ctx, p.ID, "user-123", "1.0.0", bytes.NewReader([]byte("zip")))
	rel, err := svc.PublishRelease(ctx, p.ID, "1.0.0", "mod")
	require.NoError(t, err)
	assert.Equal(t, "1.0.0", rel.Version)

	pub, err := svc.GetPublished(ctx, p.ID)
	require.NoError(t, err)
	assert.Equal(t, "1.0.0", pub.Version)

	// Unpublish by non-owner -> forbidden
	err = svc.Unpublish(ctx, p.ID, "user-intruder")
	assert.ErrorIs(t, err, domain.ErrForbidden)

	// Unpublish by owner
	err = svc.Unpublish(ctx, p.ID, "user-123")
	require.NoError(t, err)
}

func TestUnit_ProjectService_DeleteProjectAndBuild(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	svc, pRepo, _ := setupTestProjectService(t)

	p, _ := svc.CreateProject(ctx, "user-123", "Игра", "Game")
	_, _, _ = svc.UploadBuildStream(ctx, p.ID, "user-123", "1.0.0", bytes.NewReader([]byte("b1")))

	// Delete build by non-owner
	err := svc.DeleteBuild(ctx, p.ID, "user-intruder", "1.0.0")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden for non-owner deleting build, got: %v", err)
	}

	// Delete build by owner
	err = svc.DeleteBuild(ctx, p.ID, "user-123", "1.0.0")
	if err != nil {
		t.Fatalf("delete build failed: %v", err)
	}

	// Delete project by non-owner
	err = svc.DeleteProject(ctx, p.ID, "user-intruder")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden for non-owner deleting project, got: %v", err)
	}

	// Delete project by owner
	err = svc.DeleteProject(ctx, p.ID, "user-123")
	if err != nil {
		t.Fatalf("delete project failed: %v", err)
	}

	// Verify project deleted
	_, err = pRepo.Get(ctx, p.ID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound for deleted project, got: %v", err)
	}
}
