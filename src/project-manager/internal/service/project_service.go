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
	memberRepo       domain.MemberRepo
	invitationRepo   domain.InvitationRepo
	blockRepo        domain.BlockRepo
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
	memberRepo domain.MemberRepo,
	invitationRepo domain.InvitationRepo,
	blockRepo domain.BlockRepo,
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
		memberRepo:       memberRepo,
		invitationRepo:   invitationRepo,
		blockRepo:        blockRepo,
		moderationClient: moderationClient,
		buildStorage:     buildStorage,
		mediaStorage:     mediaStorage,
		deployer:         deployer,
		locker:           locker,
		maxVersions:      maxVersions,
	}
}

// CreateProject создаёт новый игровой проект и инициализирует для него пустой черновик.
func (s *ProjectService) CreateProject(ctx context.Context, ownerID, titleRu, titleEn string, isOnline bool) (*domain.Project, error) {
	if ownerID == "" {
		return nil, domain.ErrForbidden
	}
	if len([]rune(titleRu)) > 50 || len([]rune(titleEn)) > 50 {
		return nil, domain.ErrInvalidInput
	}

	p := &domain.Project{
		OwnerID:  ownerID,
		Status:   domain.ProjectStatusDraft,
		IsOnline: isOnline,
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
		IsOnline:  isOnline,
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

// CheckAccess проверяет доступ пользователя к проекту и возвращает сущность проекта с вычисленными правами.
func (s *ProjectService) CheckAccess(ctx context.Context, projectID int64, userID string, requiredPerm string) (*domain.Project, error) {
	p, err := s.projectRepo.Get(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("ProjectService.CheckAccess: %w", err)
	}

	if p.OwnerID == userID {
		p.IsOwner = true
		p.CurrentUserPermissions = domain.AllPermissions()
		return p, nil
	}

	if s.memberRepo == nil {
		return nil, domain.ErrForbidden
	}

	m, err := s.memberRepo.Get(ctx, projectID, userID)
	if err != nil {
		return nil, domain.ErrForbidden
	}

	if requiredPerm != "" && !m.HasPermission(requiredPerm) {
		return nil, domain.ErrForbidden
	}

	p.IsOwner = false
	p.CurrentUserPermissions = m.Permissions
	return p, nil
}

// GetProjectForUser загружает проект и вычисляет права текущего пользователя (или модератора/администратора).
func (s *ProjectService) GetProjectForUser(ctx context.Context, projectID int64, userID string, isStaff bool) (*domain.Project, error) {
	p, err := s.projectRepo.Get(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("ProjectService.GetProjectForUser: %w", err)
	}

	if d, err := s.draftRepo.Get(ctx, projectID); err == nil {
		p.Draft = d
	}

	if rel, err := s.releaseRepo.GetActive(ctx, projectID); err == nil {
		p.Release = rel
	}

	if isStaff {
		p.IsOwner = (p.OwnerID == userID)
		p.CurrentUserPermissions = domain.AllPermissions()
		return p, nil
	}

	if p.OwnerID == userID {
		p.IsOwner = true
		p.CurrentUserPermissions = domain.AllPermissions()
		return p, nil
	}

	if s.memberRepo != nil {
		if m, err := s.memberRepo.Get(ctx, projectID, userID); err == nil {
			p.IsOwner = false
			p.CurrentUserPermissions = m.Permissions
			return p, nil
		}
	}

	return nil, domain.ErrForbidden
}

// ListProjects возвращает постраничный список доступных пользователю проектов (собственных и совместных) и общее количество.
func (s *ProjectService) ListProjects(ctx context.Context, userID string, limit, offset int) ([]*domain.Project, int, error) {
	projects, err := s.projectRepo.ListForUser(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("ProjectService.ListProjects: %w", err)
	}

	total, err := s.projectRepo.CountForUser(ctx, userID)
	if err != nil {
		total = len(projects)
	}

	for _, p := range projects {
		if p.OwnerID == userID {
			p.IsOwner = true
			p.CurrentUserPermissions = domain.AllPermissions()
		} else {
			p.IsOwner = false
			if len(p.CurrentUserPermissions) == 0 && s.memberRepo != nil {
				if m, err := s.memberRepo.Get(ctx, p.ID, userID); err == nil {
					p.CurrentUserPermissions = m.Permissions
				}
			}
		}
		if d, err := s.draftRepo.Get(ctx, p.ID); err == nil {
			p.Draft = d
		}
		if rel, err := s.releaseRepo.GetActive(ctx, p.ID); err == nil {
			p.Release = rel
		}
	}

	return projects, total, nil
}

// ListPublishedProjects возвращает список всех опубликованных проектов со связанными активными релизами.
func (s *ProjectService) ListPublishedProjects(ctx context.Context, limit, offset int) ([]*domain.Project, int, error) {
	projects, err := s.projectRepo.ListPublished(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("ProjectService.ListPublishedProjects: %w", err)
	}

	total, err := s.projectRepo.CountPublished(ctx)
	if err != nil {
		total = len(projects)
	}

	for _, p := range projects {
		if rel, err := s.releaseRepo.GetActive(ctx, p.ID); err == nil {
			p.Release = rel
		}
	}

	return projects, total, nil
}

// UpdateDraft обновляет метаданные черновика проекта.
func (s *ProjectService) UpdateDraft(ctx context.Context, projectID int64, userID string, meta domain.DraftMeta) error {
	if len([]rune(meta.TitleRu)) > 50 || len([]rune(meta.TitleEn)) > 50 {
		return domain.ErrInvalidInput
	}
	if len([]rune(meta.SeoRu)) > 180 || len([]rune(meta.SeoEn)) > 180 {
		return domain.ErrInvalidInput
	}
	if len([]rune(meta.AboutRu)) > 800 || len([]rune(meta.AboutEn)) > 800 {
		return domain.ErrInvalidInput
	}

	unlock, err := s.locker.Acquire(ctx, fmt.Sprintf("project:%d", projectID), 1*time.Minute)
	if err != nil {
		return err
	}
	defer unlock()

	if _, err := s.CheckAccess(ctx, projectID, userID, domain.PermEditInfo); err != nil {
		return err
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
	if meta.AboutRu != "" {
		draft.AboutRu = meta.AboutRu
	}
	if meta.AboutEn != "" {
		draft.AboutEn = meta.AboutEn
	}
	if meta.ActiveBuildVersion != "" {
		draft.ActiveBuildVersion = meta.ActiveBuildVersion
	}
	if meta.IsOnline != nil {
		draft.IsOnline = *meta.IsOnline
		_ = s.projectRepo.UpdateIsOnline(ctx, projectID, *meta.IsOnline)
		_ = s.deployer.UpdateCSP(ctx, projectID, "dev", *meta.IsOnline, nil)
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

	_, err = s.CheckAccess(ctx, projectID, ownerID, domain.PermUploadBuild)
	if err != nil {
		return nil, "", err
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
	_ = s.draftRepo.UpdateActiveBuild(ctx, projectID, version, deployRes.URL)
	if d, err := s.draftRepo.Get(ctx, projectID); err == nil {
		_ = s.deployer.UpdateCSP(ctx, projectID, "dev", d.IsOnline, nil)
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
	if _, err := s.CheckAccess(ctx, projectID, ownerID, domain.PermUploadBuild); err != nil {
		return nil, err
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

	if _, err := s.CheckAccess(ctx, projectID, ownerID, domain.PermUploadBuild); err != nil {
		return err
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

	if _, err := s.CheckAccess(ctx, projectID, ownerID, domain.PermUploadMedia); err != nil {
		return "", err
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

	p, err := s.CheckAccess(ctx, projectID, ownerID, domain.PermSubmitModeration)
	if err != nil {
		return 0, err
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
		OwnerID:            p.OwnerID,
		TitleRu:            draft.TitleRu,
		TitleEn:            draft.TitleEn,
		AboutRu:            draft.AboutRu,
		AboutEn:            draft.AboutEn,
		SeoRu:              draft.SeoRu,
		SeoEn:              draft.SeoEn,
		IconPath:           draft.IconPath,
		CoverPath:          draft.CoverPath,
		VideoPath:          draft.VideoPath,
		ActiveBuildVersion: draft.ActiveBuildVersion,
		DevURL:             draft.DevURL,
		IsOnline:           draft.IsOnline,
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

	iconPath := draft.IconPath
	if iconPath != "" {
		if snapped, err := s.mediaStorage.SnapshotMediaForRelease(ctx, projectID, version, iconPath, "icon"); err == nil && snapped != "" {
			iconPath = snapped
		}
	}

	coverPath := draft.CoverPath
	if coverPath != "" {
		if snapped, err := s.mediaStorage.SnapshotMediaForRelease(ctx, projectID, version, coverPath, "cover"); err == nil && snapped != "" {
			coverPath = snapped
		}
	}

	videoPath := draft.VideoPath
	if videoPath != "" {
		if snapped, err := s.mediaStorage.SnapshotMediaForRelease(ctx, projectID, version, videoPath, "video"); err == nil && snapped != "" {
			videoPath = snapped
		}
	}

	release := &domain.Release{
		ProjectID:   projectID,
		Version:     version,
		TitleRu:     draft.TitleRu,
		TitleEn:     draft.TitleEn,
		AboutRu:     draft.AboutRu,
		AboutEn:     draft.AboutEn,
		SeoRu:       draft.SeoRu,
		SeoEn:       draft.SeoEn,
		IconPath:    iconPath,
		CoverPath:   coverPath,
		VideoPath:   videoPath,
		ProdURL:     deployRes.URL,
		IsOnline:    draft.IsOnline,
		IsActive:    true,
		PublishedBy: approvedBy,
	}

	relID, err := s.releaseRepo.Create(ctx, release)
	if err != nil {
		return nil, fmt.Errorf("ProjectService.PublishRelease create release: %w", err)
	}
	release.ID = relID

	_ = s.deployer.UpdateCSP(ctx, projectID, "prod", draft.IsOnline, nil)
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
func (s *ProjectService) Unpublish(ctx context.Context, projectID int64, ownerID string, isStaff ...bool) error {
	unlock, err := s.locker.Acquire(ctx, fmt.Sprintf("project:%d", projectID), 1*time.Minute)
	if err != nil {
		return err
	}
	defer unlock()

	staff := len(isStaff) > 0 && isStaff[0]
	if !staff {
		if _, err := s.CheckAccess(ctx, projectID, ownerID, domain.PermSubmitModeration); err != nil {
			return err
		}
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

// ─── Совместный доступ и управление участниками ─────────────────────────────────────

// SendInvitation отправляет приглашение к совместной разработке.
func (s *ProjectService) SendInvitation(
	ctx context.Context,
	projectID int64,
	inviterID, inviterEmail, inviterName string,
	inviteeID, inviteeEmail string,
	permissions []string,
) (*domain.Invitation, error) {
	if s.invitationRepo == nil || s.memberRepo == nil || s.blockRepo == nil {
		return nil, domain.ErrInvalidInput
	}
	p, err := s.projectRepo.Get(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("ProjectService.SendInvitation get project: %w", err)
	}
	if p.OwnerID != inviterID {
		return nil, domain.ErrForbidden
	}
	if inviterID == inviteeID {
		return nil, domain.ErrCannotInviteSelf
	}

	// Проверка системного пользователя: нельзя приглашать модераторов и администраторов
	isSystem, err := s.invitationRepo.IsSystemUser(ctx, inviteeID, inviteeEmail)
	if err != nil {
		return nil, fmt.Errorf("ProjectService.SendInvitation check system user: %w", err)
	}
	if isSystem {
		return nil, domain.ErrCannotInviteSystemUser
	}

	// 1. Проверка блокировки: заблокировал ли получатель отправителя?
	isBlocked, err := s.blockRepo.IsBlocked(ctx, inviteeID, inviterID)
	if err != nil {
		return nil, fmt.Errorf("ProjectService.SendInvitation check block: %w", err)
	}
	if isBlocked {
		return nil, domain.ErrUserBlocked
	}

	// 2. Проверка: уже участник?
	isMember, err := s.memberRepo.IsMember(ctx, projectID, inviteeID)
	if err != nil {
		return nil, fmt.Errorf("ProjectService.SendInvitation check member: %w", err)
	}
	if isMember {
		return nil, domain.ErrAlreadyMember
	}

	// 3. Проверка: уже есть открытое приглашение?
	if _, err := s.invitationRepo.GetPending(ctx, projectID, inviteeID); err == nil {
		return nil, domain.ErrAlreadyInvited
	}

	// 4. Фильтрация разрешений
	var validPerms []string
	for _, perm := range permissions {
		if domain.IsValidPermission(perm) {
			validPerms = append(validPerms, perm)
		}
	}
	if len(validPerms) == 0 {
		return nil, domain.ErrInvalidInput
	}

	inv := &domain.Invitation{
		ProjectID:    projectID,
		InviterID:    inviterID,
		InviterEmail: inviterEmail,
		InviterName:  inviterName,
		InviteeID:    inviteeID,
		InviteeEmail: inviteeEmail,
		Permissions:  validPerms,
		Status:       domain.InvitationStatusPending,
	}

	id, err := s.invitationRepo.Create(ctx, inv)
	if err != nil {
		return nil, fmt.Errorf("ProjectService.SendInvitation create: %w", err)
	}
	inv.ID = id
	inv.ProjectTitle = fmt.Sprintf("Проект #%d", projectID)
	if d, err := s.draftRepo.Get(ctx, projectID); err == nil {
		if d.TitleRu != "" {
			inv.ProjectTitle = d.TitleRu
		} else if d.TitleEn != "" {
			inv.ProjectTitle = d.TitleEn
		}
		inv.ProjectIcon = d.IconPath
	}
	return inv, nil
}

// ListIncomingInvitations возвращает входящие активные приглашения пользователя.
func (s *ProjectService) ListIncomingInvitations(ctx context.Context, userID string) ([]*domain.Invitation, error) {
	if s.invitationRepo == nil {
		return nil, nil
	}
	return s.invitationRepo.ListIncoming(ctx, userID)
}

// ListOutgoingInvitations возвращает исходящие приглашения владельца проекта.
func (s *ProjectService) ListOutgoingInvitations(ctx context.Context, userID string, projectID int64) ([]*domain.Invitation, error) {
	if s.invitationRepo == nil {
		return nil, nil
	}
	if projectID != 0 {
		p, err := s.projectRepo.Get(ctx, projectID)
		if err != nil {
			return nil, err
		}
		if p.OwnerID != userID {
			return nil, domain.ErrForbidden
		}
	}
	return s.invitationRepo.ListOutgoing(ctx, userID, projectID)
}

// RespondInvitation принимает или отклоняет входящее приглашение.
func (s *ProjectService) RespondInvitation(ctx context.Context, invitationID int64, userID string, accept bool) error {
	if s.invitationRepo == nil || s.memberRepo == nil {
		return domain.ErrInvalidInput
	}
	inv, err := s.invitationRepo.Get(ctx, invitationID)
	if err != nil {
		return fmt.Errorf("ProjectService.RespondInvitation get: %w", err)
	}
	if inv.InviteeID != userID {
		return domain.ErrForbidden
	}
	if inv.Status != domain.InvitationStatusPending {
		return domain.ErrInvitationClosed
	}

	if accept {
		if err := s.invitationRepo.UpdateStatus(ctx, invitationID, domain.InvitationStatusAccepted); err != nil {
			return fmt.Errorf("ProjectService.RespondInvitation update status: %w", err)
		}
		m := &domain.Member{
			ProjectID:   inv.ProjectID,
			UserID:      inv.InviteeID,
			UserEmail:   inv.InviteeEmail,
			Permissions: inv.Permissions,
		}
		if err := s.memberRepo.Add(ctx, m); err != nil {
			return fmt.Errorf("ProjectService.RespondInvitation add member: %w", err)
		}
		return nil
	}

	return s.invitationRepo.UpdateStatus(ctx, invitationID, domain.InvitationStatusDeclined)
}

// CancelInvitation отменяет отправленное приглашение (только отправитель или владелец проекта).
func (s *ProjectService) CancelInvitation(ctx context.Context, invitationID int64, userID string) error {
	if s.invitationRepo == nil {
		return domain.ErrInvalidInput
	}
	inv, err := s.invitationRepo.Get(ctx, invitationID)
	if err != nil {
		return fmt.Errorf("ProjectService.CancelInvitation get: %w", err)
	}
	if inv.InviterID != userID {
		p, err := s.projectRepo.Get(ctx, inv.ProjectID)
		if err != nil || p.OwnerID != userID {
			return domain.ErrForbidden
		}
	}
	if inv.Status != domain.InvitationStatusPending {
		return domain.ErrInvitationClosed
	}
	return s.invitationRepo.UpdateStatus(ctx, invitationID, domain.InvitationStatusCanceled)
}

// ListProjectMembers возвращает список участников проекта и ID владельца.
func (s *ProjectService) ListProjectMembers(ctx context.Context, projectID int64, userID string) ([]*domain.Member, string, error) {
	if s.memberRepo == nil {
		return nil, "", nil
	}
	p, err := s.CheckAccess(ctx, projectID, userID, "")
	if err != nil {
		return nil, "", err
	}
	members, err := s.memberRepo.ListByProject(ctx, projectID)
	if err != nil {
		return nil, "", fmt.Errorf("ProjectService.ListProjectMembers: %w", err)
	}
	return members, p.OwnerID, nil
}

// UpdateMemberPermissions обновляет разрешения участника (доступно только владельцу проекта).
func (s *ProjectService) UpdateMemberPermissions(ctx context.Context, projectID int64, ownerID, memberUserID string, permissions []string) (*domain.Member, error) {
	if s.memberRepo == nil {
		return nil, domain.ErrInvalidInput
	}
	p, err := s.projectRepo.Get(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("ProjectService.UpdateMemberPermissions get: %w", err)
	}
	if p.OwnerID != ownerID {
		return nil, domain.ErrForbidden
	}

	var validPerms []string
	for _, perm := range permissions {
		if domain.IsValidPermission(perm) {
			validPerms = append(validPerms, perm)
		}
	}

	if err := s.memberRepo.UpdatePermissions(ctx, projectID, memberUserID, validPerms); err != nil {
		return nil, fmt.Errorf("ProjectService.UpdateMemberPermissions update: %w", err)
	}
	return s.memberRepo.Get(ctx, projectID, memberUserID)
}

// RemoveMember удаляет участника из проекта (доступно только владельцу проекта).
func (s *ProjectService) RemoveMember(ctx context.Context, projectID int64, ownerID, memberUserID string) error {
	if s.memberRepo == nil {
		return domain.ErrInvalidInput
	}
	p, err := s.projectRepo.Get(ctx, projectID)
	if err != nil {
		return fmt.Errorf("ProjectService.RemoveMember get: %w", err)
	}
	if p.OwnerID != ownerID {
		return domain.ErrForbidden
	}
	return s.memberRepo.Delete(ctx, projectID, memberUserID)
}

// LeaveProject позволяет участнику добровольно покинуть проект.
func (s *ProjectService) LeaveProject(ctx context.Context, projectID int64, userID string) error {
	if s.memberRepo == nil {
		return domain.ErrInvalidInput
	}
	p, err := s.projectRepo.Get(ctx, projectID)
	if err != nil {
		return fmt.Errorf("ProjectService.LeaveProject get: %w", err)
	}
	if p.OwnerID == userID {
		return domain.ErrForbidden // Владелец не может покинуть собственный проект
	}
	return s.memberRepo.Delete(ctx, projectID, userID)
}

// ListSharedProjects возвращает все проекты, к которым пользователю предоставлен доступ в качестве участника.
func (s *ProjectService) ListSharedProjects(ctx context.Context, userID string) ([]*domain.SharedProject, error) {
	if s.memberRepo == nil {
		return nil, nil
	}
	memberships, err := s.memberRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ProjectService.ListSharedProjects list memberships: %w", err)
	}

	var res []*domain.SharedProject
	for _, m := range memberships {
		p, err := s.projectRepo.Get(ctx, m.ProjectID)
		if err != nil {
			continue
		}
		p.IsOwner = false
		p.CurrentUserPermissions = m.Permissions
		if d, err := s.draftRepo.Get(ctx, p.ID); err == nil {
			p.Draft = d
		}
		if rel, err := s.releaseRepo.GetActive(ctx, p.ID); err == nil {
			p.Release = rel
		}

		res = append(res, &domain.SharedProject{
			Project:     p,
			Permissions: m.Permissions,
			JoinedAt:    m.CreatedAt,
		})
	}
	return res, nil
}

// BlockUser добавляет пользователя в черный список и отменяет висящие приглашения от него.
func (s *ProjectService) BlockUser(ctx context.Context, userID, blockedUserID, blockedEmail, blockedName string) error {
	if s.blockRepo == nil {
		return domain.ErrInvalidInput
	}
	if userID == blockedUserID {
		return domain.ErrInvalidInput
	}

	block := &domain.UserBlock{
		UserID:           userID,
		BlockedUserID:    blockedUserID,
		BlockedUserEmail: blockedEmail,
		BlockedUserName:  blockedName,
	}
	if err := s.blockRepo.Block(ctx, block); err != nil {
		return fmt.Errorf("ProjectService.BlockUser: %w", err)
	}

	// Отменяем любые висящие приглашения от заблокированного пользователя
	if s.invitationRepo != nil {
		_ = s.invitationRepo.CancelAllPendingBetween(ctx, blockedUserID, userID)
	}
	return nil
}

// UnblockUser удаляет пользователя из черного списка.
func (s *ProjectService) UnblockUser(ctx context.Context, userID, blockedUserID string) error {
	if s.blockRepo == nil {
		return domain.ErrInvalidInput
	}
	return s.blockRepo.Unblock(ctx, userID, blockedUserID)
}

// ListBlockedUsers возвращает список заблокированных пользователей.
func (s *ProjectService) ListBlockedUsers(ctx context.Context, userID string) ([]*domain.UserBlock, error) {
	if s.blockRepo == nil {
		return nil, nil
	}
	return s.blockRepo.ListBlocked(ctx, userID)
}
