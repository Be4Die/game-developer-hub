package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
	"github.com/Be4Die/game-developer-hub/project-manager/internal/infrastructure/valkey"
)

// ProjectService реализует бизнес-логику управления проектами, черновиками и сборками.
type ProjectService struct {
	projectRepo      domain.ProjectRepo
	draftRepo        domain.DraftRepo
	buildRepo        domain.BuildRepo
	releaseRepo      domain.ReleaseRepo
	deploymentRepo   domain.DeploymentRepo
	moderationClient domain.ModerationClient
	buildStorage     domain.BuildStorage
	mediaStorage     domain.MediaStorage
	deployer         domain.Deployer
	locker           domain.Locker
	maxVersions      int
}

// NewProjectService создаёт экземпляр ProjectService.
func NewProjectService(
	projectRepo domain.ProjectRepo,
	draftRepo domain.DraftRepo,
	buildRepo domain.BuildRepo,
	releaseRepo domain.ReleaseRepo,
	deploymentRepo domain.DeploymentRepo,
	moderationClient domain.ModerationClient,
	buildStorage domain.BuildStorage,
	mediaStorage domain.MediaStorage,
	deployer domain.Deployer,
	locker domain.Locker,
	maxVersions int,
) *ProjectService {
	if maxVersions <= 0 {
		maxVersions = 5
	}
	if locker == nil {
		locker = valkey.NewNoOpLocker()
	}
	return &ProjectService{
		projectRepo:      projectRepo,
		draftRepo:        draftRepo,
		buildRepo:        buildRepo,
		releaseRepo:      releaseRepo,
		deploymentRepo:   deploymentRepo,
		moderationClient: moderationClient,
		buildStorage:     buildStorage,
		mediaStorage:     mediaStorage,
		deployer:         deployer,
		locker:           locker,
		maxVersions:      maxVersions,
	}
}

// CreateProject создаёт новый игровой проект и инициализирует для него пустой черновик.
func (s *ProjectService) CreateProject(ctx context.Context, ownerID, titleRu, titleEn string) (*domain.Project, error) {
	if ownerID == "" {
		return nil, domain.ErrForbidden
	}

	p := &domain.Project{
		OwnerID: ownerID,
		Status:  domain.ProjectStatusDraft,
	}

	id, err := s.projectRepo.Create(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("ProjectService.CreateProject: %w", err)
	}
	p.ID = id

	draft := &domain.Draft{
		ProjectID: id,
		TitleRu:   titleRu,
		TitleEn:   titleEn,
	}
	if err := s.draftRepo.Create(ctx, draft); err != nil {
		_ = s.projectRepo.Delete(ctx, id)
		return nil, fmt.Errorf("ProjectService.CreateProject init draft: %w", err)
	}
	p.Draft = draft

	return p, nil
}

// GetProject возвращает проект по его идентификатору вместе с актуальным черновиком и активным релизом.
func (s *ProjectService) GetProject(ctx context.Context, projectID int64) (*domain.Project, error) {
	p, err := s.projectRepo.Get(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("ProjectService.GetProject: %w", err)
	}

	if d, err := s.draftRepo.Get(ctx, projectID); err == nil {
		p.Draft = d
	}

	if rel, err := s.releaseRepo.GetActive(ctx, projectID); err == nil {
		p.Release = rel
	}

	return p, nil
}

// ListProjects возвращает постраничный список проектов указанного владельца и общее количество.
func (s *ProjectService) ListProjects(ctx context.Context, ownerID string, limit, offset int) ([]*domain.Project, int, error) {
	projects, err := s.projectRepo.ListByOwner(ctx, ownerID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("ProjectService.ListProjects: %w", err)
	}

	total, err := s.projectRepo.CountByOwner(ctx, ownerID)
	if err != nil {
		total = len(projects)
	}

	for _, p := range projects {
		if d, err := s.draftRepo.Get(ctx, p.ID); err == nil {
			p.Draft = d
		}
		if rel, err := s.releaseRepo.GetActive(ctx, p.ID); err == nil {
			p.Release = rel
		}
	}

	return projects, total, nil
}

// UpdateDraft обновляет метаданные черновика проекта.
func (s *ProjectService) UpdateDraft(ctx context.Context, projectID int64, ownerID string, meta domain.DraftMeta) error {
	unlock, err := s.locker.Acquire(ctx, fmt.Sprintf("project:%d", projectID), 1*time.Minute)
	if err != nil {
		return err
	}
	defer unlock()

	p, err := s.projectRepo.Get(ctx, projectID)
	if err != nil {
		return fmt.Errorf("ProjectService.UpdateDraft: %w", err)
	}
	if p.OwnerID != ownerID {
		return domain.ErrForbidden
	}

	draft, err := s.draftRepo.Get(ctx, projectID)
	if err != nil {
		draft = &domain.Draft{ProjectID: projectID}
	}

	if meta.TitleRu != "" {
		draft.TitleRu = meta.TitleRu
	}
	if meta.TitleEn != "" {
		draft.TitleEn = meta.TitleEn
	}
	if meta.SeoRu != "" {
		draft.SeoRu = meta.SeoRu
	}
	if meta.SeoEn != "" {
		draft.SeoEn = meta.SeoEn
	}
	if meta.About != "" {
		draft.About = meta.About
	}
	if meta.ActiveBuildVersion != "" {
		draft.ActiveBuildVersion = meta.ActiveBuildVersion
	}

	return s.draftRepo.Update(ctx, draft)
}

// DeleteProject удаляет проект, все его файлы, историю и деплои.
func (s *ProjectService) DeleteProject(ctx context.Context, projectID int64, ownerID string) error {
	unlock, err := s.locker.Acquire(ctx, fmt.Sprintf("project:%d", projectID), 1*time.Minute)
	if err != nil {
		return err
	}
	defer unlock()

	p, err := s.projectRepo.Get(ctx, projectID)
	if err != nil {
		return fmt.Errorf("ProjectService.DeleteProject: %w", err)
	}
	if p.OwnerID != ownerID {
		return domain.ErrForbidden
	}

	if err := s.projectRepo.Delete(ctx, projectID); err != nil {
		return fmt.Errorf("ProjectService.DeleteProject repo: %w", err)
	}

	_ = s.buildStorage.DeleteProject(projectID)
	_ = s.deployer.DeleteProject(ctx, projectID)
	_ = s.deployer.UndeployProd(ctx, projectID)

	return nil
}

// UploadBuildStream принимает поток архива сборки, сохраняет на диск, валидирует и авто-деплоит в Dev.
func (s *ProjectService) UploadBuildStream(ctx context.Context, projectID int64, ownerID, version string, archiveReader io.Reader) (*domain.Build, string, error) {
	unlock, err := s.locker.Acquire(ctx, fmt.Sprintf("project:%d", projectID), 2*time.Minute)
	if err != nil {
		return nil, "", err
	}
	defer unlock()

	p, err := s.projectRepo.Get(ctx, projectID)
	if err != nil {
		return nil, "", fmt.Errorf("ProjectService.UploadBuild: %w", err)
	}
	if p.OwnerID != ownerID {
		return nil, "", domain.ErrForbidden
	}

	// Проверка на дубликат версии
	if _, err := s.buildRepo.Get(ctx, projectID, version); err == nil {
		return nil, "", domain.ErrAlreadyExists
	}

	// 1. Потоковое сохранение архива на диск
	archivePath, sizeBytes, err := s.buildStorage.SaveArchiveStream(ctx, projectID, version, archiveReader)
	if err != nil {
		return nil, "", fmt.Errorf("ProjectService.UploadBuild save archive: %w", err)
	}

	// 2. Автоматическое развертывание в Dev-окружение (с распаковкой и валидацией)
	deployRes, err := s.deployer.DeployDev(ctx, projectID, version, archivePath)
	if err != nil {
		_ = s.buildStorage.DeleteBuild(projectID, version)
		return nil, "", fmt.Errorf("ProjectService.UploadBuild deploy dev: %w", err)
	}

	// 3. Запись сборки в БД
	b := &domain.Build{
		ProjectID:    projectID,
		Version:      version,
		FilePath:     archivePath,
		FileSize:     sizeBytes,
		IsUnpacked:   true,
		UnpackedPath: deployRes.UnpackedPath,
	}
	if err := s.buildRepo.Create(ctx, b); err != nil {
		_ = s.buildStorage.DeleteBuild(projectID, version)
		_ = s.deployer.DeleteVersion(ctx, projectID, version)
		return nil, "", fmt.Errorf("ProjectService.UploadBuild db create: %w", err)
	}

	// 4. Обновление активной сборки в черновике
	if err := s.draftRepo.UpdateActiveBuild(ctx, projectID, version, deployRes.URL); err != nil {
		// non-critical
	}

	// 5. Аудит развертывания
	_ = s.deploymentRepo.Create(ctx, &domain.DeploymentRecord{
		ProjectID:   projectID,
		Environment: domain.DeploymentEnvDev,
		Version:     version,
		Status:      domain.DeploymentStatusSuccess,
	})

	// 6. Ротация старых версий (FIFO) с полной очисткой архивов и распакованных файлов
	builds, err := s.buildRepo.ListByProject(ctx, projectID, s.maxVersions+1)
	if err == nil && len(builds) > s.maxVersions {
		for _, old := range builds[s.maxVersions:] {
			_ = s.buildStorage.DeleteBuild(old.ProjectID, old.Version)
			_ = s.deployer.DeleteVersion(ctx, old.ProjectID, old.Version)
			_ = s.buildRepo.Delete(ctx, old.ID)
		}
	}

	return b, deployRes.URL, nil
}

// UploadBuild выполняет загрузку сборки из среза байтов.
func (s *ProjectService) UploadBuild(ctx context.Context, projectID int64, ownerID, version string, data []byte) (*domain.Build, string, error) {
	return s.UploadBuildStream(ctx, projectID, ownerID, version, bytes.NewReader(data))
}

// ListBuilds возвращает список сборок проекта.
func (s *ProjectService) ListBuilds(ctx context.Context, projectID int64, ownerID string) ([]*domain.Build, error) {
	p, err := s.projectRepo.Get(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("ProjectService.ListBuilds: %w", err)
	}
	if p.OwnerID != ownerID {
		return nil, domain.ErrForbidden
	}
	return s.buildRepo.ListByProject(ctx, projectID, s.maxVersions)
}

// DeleteBuild удаляет конкретную версию сборки и её распакованную директорию.
func (s *ProjectService) DeleteBuild(ctx context.Context, projectID int64, ownerID, version string) error {
	unlock, err := s.locker.Acquire(ctx, fmt.Sprintf("project:%d", projectID), 1*time.Minute)
	if err != nil {
		return err
	}
	defer unlock()

	p, err := s.projectRepo.Get(ctx, projectID)
	if err != nil {
		return fmt.Errorf("ProjectService.DeleteBuild: %w", err)
	}
	if p.OwnerID != ownerID {
		return domain.ErrForbidden
	}

	b, err := s.buildRepo.Get(ctx, projectID, version)
	if err != nil {
		return fmt.Errorf("ProjectService.DeleteBuild get: %w", err)
	}

	_ = s.buildStorage.DeleteBuild(projectID, version)
	_ = s.deployer.DeleteVersion(ctx, projectID, version)
	_ = s.buildRepo.Delete(ctx, b.ID)

	if d, err := s.draftRepo.Get(ctx, projectID); err == nil && d.ActiveBuildVersion == version {
		_ = s.draftRepo.UpdateActiveBuild(ctx, projectID, "", "")
	}

	return nil
}

// UploadMediaStream загружает промо-материал (icon, cover, video) через поток.
func (s *ProjectService) UploadMediaStream(ctx context.Context, projectID int64, ownerID, mediaType string, reader io.Reader) (string, error) {
	unlock, err := s.locker.Acquire(ctx, fmt.Sprintf("project:%d", projectID), 1*time.Minute)
	if err != nil {
		return "", err
	}
	defer unlock()

	p, err := s.projectRepo.Get(ctx, projectID)
	if err != nil {
		return "", fmt.Errorf("ProjectService.UploadMedia: %w", err)
	}
	if p.OwnerID != ownerID {
		return "", domain.ErrForbidden
	}

	filePath, err := s.mediaStorage.SaveMediaStream(ctx, projectID, mediaType, reader)
	if err != nil {
		return "", fmt.Errorf("ProjectService.UploadMedia save: %w", err)
	}

	if err := s.draftRepo.UpdateMedia(ctx, projectID, mediaType, filePath); err != nil {
		return "", fmt.Errorf("ProjectService.UploadMedia update draft: %w", err)
	}

	return filePath, nil
}

// UploadMedia загружает промо-материал из среза байтов.
func (s *ProjectService) UploadMedia(ctx context.Context, projectID int64, ownerID, mediaType string, data []byte) (string, error) {
	return s.UploadMediaStream(ctx, projectID, ownerID, mediaType, bytes.NewReader(data))
}

// SubmitForModeration проверяет готовность черновика и отправляет снимок в подсистему модерации.
func (s *ProjectService) SubmitForModeration(ctx context.Context, projectID int64, ownerID string) (int64, error) {
	unlock, err := s.locker.Acquire(ctx, fmt.Sprintf("project:%d", projectID), 1*time.Minute)
	if err != nil {
		return 0, err
	}
	defer unlock()

	p, err := s.projectRepo.Get(ctx, projectID)
	if err != nil {
		return 0, fmt.Errorf("ProjectService.SubmitForModeration: %w", err)
	}
	if p.OwnerID != ownerID {
		return 0, domain.ErrForbidden
	}
	if p.Status == domain.ProjectStatusPending {
		return 0, domain.ErrAlreadyInModeration
	}

	draft, err := s.draftRepo.Get(ctx, projectID)
	if err != nil {
		return 0, domain.ErrDraftNotReady
	}
	p.Draft = draft

	if err := draft.IsReadyForModeration(); err != nil {
		return 0, err
	}

	snapshot := &domain.ProjectSnapshot{
		ProjectID:          projectID,
		OwnerID:            ownerID,
		TitleRu:            draft.TitleRu,
		TitleEn:            draft.TitleEn,
		About:              draft.About,
		SeoRu:              draft.SeoRu,
		SeoEn:              draft.SeoEn,
		IconPath:           draft.IconPath,
		CoverPath:          draft.CoverPath,
		VideoPath:          draft.VideoPath,
		ActiveBuildVersion: draft.ActiveBuildVersion,
		DevURL:             draft.DevURL,
	}

	requestID, err := s.moderationClient.SubmitDraft(ctx, snapshot)
	if err != nil {
		return 0, fmt.Errorf("ProjectService.SubmitForModeration submit draft: %w", err)
	}

	_ = s.projectRepo.UpdateStatus(ctx, projectID, domain.ProjectStatusPending)

	return requestID, nil
}

// PublishRelease публикует одобренную версию игры в продуктивное окружение.
func (s *ProjectService) PublishRelease(ctx context.Context, projectID int64, version, approvedBy string) (*domain.Release, error) {
	unlock, err := s.locker.Acquire(ctx, fmt.Sprintf("project:%d", projectID), 2*time.Minute)
	if err != nil {
		return nil, err
	}
	defer unlock()

	archivePath := s.buildStorage.GetArchivePath(projectID, version)
	deployRes, err := s.deployer.DeployProd(ctx, projectID, version, archivePath)
	if err != nil {
		return nil, fmt.Errorf("ProjectService.PublishRelease deploy prod: %w", err)
	}

	draft, err := s.draftRepo.Get(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("ProjectService.PublishRelease get draft: %w", err)
	}

	_ = s.releaseRepo.Deactivate(ctx, projectID)

	release := &domain.Release{
		ProjectID:   projectID,
		Version:     version,
		TitleRu:     draft.TitleRu,
		TitleEn:     draft.TitleEn,
		About:       draft.About,
		IconPath:    draft.IconPath,
		CoverPath:   draft.CoverPath,
		VideoPath:   draft.VideoPath,
		ProdURL:     deployRes.URL,
		IsActive:    true,
		PublishedBy: approvedBy,
	}

	relID, err := s.releaseRepo.Create(ctx, release)
	if err != nil {
		return nil, fmt.Errorf("ProjectService.PublishRelease create release: %w", err)
	}
	release.ID = relID

	_ = s.projectRepo.UpdateStatus(ctx, projectID, domain.ProjectStatusPublished)
	_ = s.deploymentRepo.Create(ctx, &domain.DeploymentRecord{
		ProjectID:   projectID,
		Environment: domain.DeploymentEnvProd,
		Version:     version,
		Status:      domain.DeploymentStatusSuccess,
	})

	return release, nil
}

// RejectDraft возвращает проект в статус черновика при отклонении модератором.
func (s *ProjectService) RejectDraft(ctx context.Context, projectID int64) error {
	unlock, err := s.locker.Acquire(ctx, fmt.Sprintf("project:%d", projectID), 1*time.Minute)
	if err != nil {
		return err
	}
	defer unlock()

	return s.projectRepo.UpdateStatus(ctx, projectID, domain.ProjectStatusDraft)
}

// GetPublished возвращает активный опубликованный релиз проекта.
func (s *ProjectService) GetPublished(ctx context.Context, projectID int64) (*domain.Release, error) {
	return s.releaseRepo.GetActive(ctx, projectID)
}

// Unpublish снимает игру с публикации в продуктивном окружении.
func (s *ProjectService) Unpublish(ctx context.Context, projectID int64, ownerID string) error {
	unlock, err := s.locker.Acquire(ctx, fmt.Sprintf("project:%d", projectID), 1*time.Minute)
	if err != nil {
		return err
	}
	defer unlock()

	p, err := s.projectRepo.Get(ctx, projectID)
	if err != nil {
		return fmt.Errorf("ProjectService.Unpublish: %w", err)
	}
	if p.OwnerID != ownerID {
		return domain.ErrForbidden
	}

	if err := s.deployer.UndeployProd(ctx, projectID); err != nil {
		return fmt.Errorf("ProjectService.Unpublish undeploy: %w", err)
	}

	if err := s.releaseRepo.Deactivate(ctx, projectID); err != nil {
		return fmt.Errorf("ProjectService.Unpublish deactivate: %w", err)
	}

	_ = s.projectRepo.UpdateStatus(ctx, projectID, domain.ProjectStatusDraft)

	return nil
}
