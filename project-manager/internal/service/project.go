// Package service содержит бизнес-логику подсистемы проектов.
package service

import (
	"context"
	"fmt"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
)

// ProjectService бизнес-логика управления проектами.
type ProjectService struct {
	projectRepo  domain.ProjectRepo
	buildRepo    domain.ProjectBuildRepo
	mediaStorage domain.ProjectMediaStorage
	maxVersions  int
}

// NewProjectService создаёт сервис.
func NewProjectService(
	projectRepo domain.ProjectRepo,
	buildRepo domain.ProjectBuildRepo,
	mediaStorage domain.ProjectMediaStorage,
	maxVersions int,
) *ProjectService {
	if maxVersions <= 0 {
		maxVersions = 5
	}
	return &ProjectService{
		projectRepo:  projectRepo,
		buildRepo:    buildRepo,
		mediaStorage: mediaStorage,
		maxVersions:  maxVersions,
	}
}

// CreateProject создаёт новый проект.
func (s *ProjectService) CreateProject(ctx context.Context, ownerID, titleRu, titleEn string) (*domain.Project, error) {
	p := &domain.Project{
		OwnerID: ownerID,
		TitleRu: titleRu,
		TitleEn: titleEn,
		Status:  domain.ProjectStatusDraft,
	}
	id, err := s.projectRepo.Create(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("ProjectService.CreateProject: %w", err)
	}
	p.ID = id
	return p, nil
}

// GetProject возвращает проект по ID.
func (s *ProjectService) GetProject(ctx context.Context, projectID int64) (*domain.Project, error) {
	p, err := s.projectRepo.Get(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("ProjectService.GetProject: %w", err)
	}
	return p, nil
}

// ListProjects возвращает проекты пользователя.
func (s *ProjectService) ListProjects(ctx context.Context, ownerID string) ([]*domain.Project, error) {
	projects, err := s.projectRepo.ListByOwner(ctx, ownerID, 100, 0)
	if err != nil {
		return nil, fmt.Errorf("ProjectService.ListProjects: %w", err)
	}
	return projects, nil
}

// UpdateMeta обновляет метаданные проекта.
func (s *ProjectService) UpdateMeta(ctx context.Context, projectID int64, ownerID string, meta ProjectMeta) error {
	p, err := s.projectRepo.Get(ctx, projectID)
	if err != nil {
		return fmt.Errorf("ProjectService.UpdateMeta: %w", err)
	}
	if p.OwnerID != ownerID {
		return domain.ErrForbidden
	}
	p.TitleRu = meta.TitleRu
	p.TitleEn = meta.TitleEn
	p.SeoRu = meta.SeoRu
	p.SeoEn = meta.SeoEn
	p.About = meta.About
	p.ActiveBuildVersion = meta.ActiveBuildVersion
	return s.projectRepo.Update(ctx, p)
}

// DeleteProject удаляет проект вместе с файлами.
func (s *ProjectService) DeleteProject(ctx context.Context, projectID int64, ownerID string) error {
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
	_ = s.mediaStorage.DeleteProjectDir(projectID)
	return nil
}

// UploadBuild загружает билд с ротацией (FIFO, maxVersions).
func (s *ProjectService) UploadBuild(ctx context.Context, projectID int64, ownerID, version string, data []byte) error {
	p, err := s.projectRepo.Get(ctx, projectID)
	if err != nil {
		return fmt.Errorf("ProjectService.UploadBuild: %w", err)
	}
	if p.OwnerID != ownerID {
		return domain.ErrForbidden
	}
	// Проверка уникальности версии
	_, err = s.buildRepo.Get(ctx, projectID, version)
	if err == nil {
		return domain.ErrAlreadyExists
	}
	if err != domain.ErrNotFound {
		return fmt.Errorf("ProjectService.UploadBuild check: %w", err)
	}

	// Сохраняем файл
	path, err := s.mediaStorage.SaveBuild(projectID, version, data)
	if err != nil {
		return fmt.Errorf("ProjectService.UploadBuild save: %w", err)
	}

	// Записываем в БД
	b := &domain.ProjectBuild{
		ProjectID: projectID,
		Version:   version,
		FilePath:  path,
		FileSize:  int64(len(data)),
	}
	if err := s.buildRepo.Create(ctx, b); err != nil {
		// Откат файла
		_ = s.mediaStorage.DeleteBuild(projectID, version)
		return fmt.Errorf("ProjectService.UploadBuild db: %w", err)
	}

	// Обновляем активную версию билда
	p.ActiveBuildVersion = version
	if err := s.projectRepo.Update(ctx, p); err != nil {
		// Не критично, но логируем
		// В реальном сервисе здесь стоит добавить structured logging
	}

	// Ротация: удаляем старые билды, оставляя maxVersions
	builds, err := s.buildRepo.ListByProject(ctx, projectID, s.maxVersions+1)
	if err != nil {
		return fmt.Errorf("ProjectService.UploadBuild list: %w", err)
	}
	if len(builds) > s.maxVersions {
		for _, old := range builds[s.maxVersions:] {
			_ = s.mediaStorage.DeleteBuild(old.ProjectID, old.Version)
			_ = s.buildRepo.Delete(ctx, old.ID)
		}
	}

	return nil
}

// ListBuilds возвращает билды проекта.
func (s *ProjectService) ListBuilds(ctx context.Context, projectID int64, ownerID string) ([]*domain.ProjectBuild, error) {
	p, err := s.projectRepo.Get(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("ProjectService.ListBuilds: %w", err)
	}
	if p.OwnerID != ownerID {
		return nil, domain.ErrForbidden
	}
	builds, err := s.buildRepo.ListByProject(ctx, projectID, s.maxVersions)
	if err != nil {
		return nil, fmt.Errorf("ProjectService.ListBuilds: %w", err)
	}
	return builds, nil
}

// DeleteBuild удаляет конкретный билд.
func (s *ProjectService) DeleteBuild(ctx context.Context, projectID int64, ownerID, version string) error {
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
	if err := s.mediaStorage.DeleteBuild(projectID, version); err != nil {
		return fmt.Errorf("ProjectService.DeleteBuild fs: %w", err)
	}
	if err := s.buildRepo.Delete(ctx, b.ID); err != nil {
		return fmt.Errorf("ProjectService.DeleteBuild db: %w", err)
	}
	// Если удалённый билд был активным — сбрасываем активную версию
	if p.ActiveBuildVersion == version {
		p.ActiveBuildVersion = ""
		_ = s.projectRepo.Update(ctx, p)
	}
	return nil
}

// UploadMedia загружает промо-материал (icon, cover, video).
func (s *ProjectService) UploadMedia(ctx context.Context, projectID int64, ownerID, mediaType string, data []byte) error {
	p, err := s.projectRepo.Get(ctx, projectID)
	if err != nil {
		return fmt.Errorf("ProjectService.UploadMedia: %w", err)
	}
	if p.OwnerID != ownerID {
		return domain.ErrForbidden
	}

	var path string
	switch mediaType {
	case "icon":
		path, err = s.mediaStorage.SaveIcon(projectID, data)
		p.IconPath = path
	case "cover":
		path, err = s.mediaStorage.SaveCover(projectID, data)
		p.CoverPath = path
	case "video":
		path, err = s.mediaStorage.SaveVideo(projectID, data)
		p.VideoPath = path
	default:
		return domain.ErrInvalidInput
	}
	if err != nil {
		return fmt.Errorf("ProjectService.UploadMedia save: %w", err)
	}
	return s.projectRepo.Update(ctx, p)
}

// ProjectMeta метаданные для обновления.
type ProjectMeta struct {
	TitleRu            string
	TitleEn            string
	SeoRu              string
	SeoEn              string
	About              string
	ActiveBuildVersion string
}
