package service

import (
	"context"
	"fmt"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
)

// ModerationService реализует бизнес-логику проверки игр модератором и публикации в прод.
type ModerationService struct {
	projectRepo    domain.ProjectRepo
	draftRepo      domain.DraftRepo
	buildRepo      domain.BuildRepo
	moderationRepo domain.ModerationRepo
	releaseRepo    domain.ReleaseRepo
	deploymentRepo domain.DeploymentRepo
	buildStorage   domain.BuildStorage
	deployer       domain.Deployer
}

// NewModerationService создаёт экземпляр ModerationService.
func NewModerationService(
	projectRepo domain.ProjectRepo,
	draftRepo domain.DraftRepo,
	buildRepo domain.BuildRepo,
	moderationRepo domain.ModerationRepo,
	releaseRepo domain.ReleaseRepo,
	deploymentRepo domain.DeploymentRepo,
	buildStorage domain.BuildStorage,
	deployer domain.Deployer,
) *ModerationService {
	return &ModerationService{
		projectRepo:    projectRepo,
		draftRepo:      draftRepo,
		buildRepo:      buildRepo,
		moderationRepo: moderationRepo,
		releaseRepo:    releaseRepo,
		deploymentRepo: deploymentRepo,
		buildStorage:   buildStorage,
		deployer:       deployer,
	}
}

// ListTickets возвращает список заявок модерации.
func (s *ModerationService) ListTickets(ctx context.Context, status *domain.ModerationStatus, limit, offset int) ([]*domain.ModerationTicket, error) {
	return s.moderationRepo.ListTickets(ctx, status, limit, offset)
}

// GetTicket возвращает информацию о тикете модерации.
func (s *ModerationService) GetTicket(ctx context.Context, ticketID int64) (*domain.ModerationTicket, error) {
	return s.moderationRepo.GetTicket(ctx, ticketID)
}

// GetTicketByProject возвращает последний тикет проекта.
func (s *ModerationService) GetTicketByProject(ctx context.Context, projectID int64) (*domain.ModerationTicket, error) {
	return s.moderationRepo.GetLatestTicketByProject(ctx, projectID)
}

// Approve утверждает проект: развёртывает в Prod, создаёт Release и переводит статус в Published.
func (s *ModerationService) Approve(ctx context.Context, projectID int64, moderatorID, comment string) (*domain.Release, error) {
	draft, err := s.draftRepo.Get(ctx, projectID)
	if err != nil {
		return nil, domain.ErrNotFound
	}
	if draft.ActiveBuildVersion == "" {
		return nil, domain.ErrNoActiveBuild
	}

	// Находим активный тикет
	ticket, err := s.moderationRepo.GetLatestTicketByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("ModerationService.Approve get ticket: %w", err)
	}

	archivePath := s.buildStorage.GetArchivePath(projectID, draft.ActiveBuildVersion)

	// 1. Развертывание в Prod
	deployRes, err := s.deployer.DeployProd(ctx, projectID, draft.ActiveBuildVersion, archivePath)
	if err != nil {
		_ = s.deploymentRepo.Create(ctx, &domain.DeploymentRecord{
			ProjectID:    projectID,
			Environment:  domain.DeploymentEnvProd,
			Version:      draft.ActiveBuildVersion,
			Status:       domain.DeploymentStatusFailed,
			ErrorMessage: err.Error(),
		})
		return nil, fmt.Errorf("ModerationService.Approve deploy prod: %w", err)
	}

	// 2. Создание Release
	rel := &domain.Release{
		ProjectID:   projectID,
		Version:     draft.ActiveBuildVersion,
		TitleRu:     draft.TitleRu,
		TitleEn:     draft.TitleEn,
		SeoRu:       draft.SeoRu,
		SeoEn:       draft.SeoEn,
		About:       draft.About,
		IconPath:    draft.IconPath,
		CoverPath:   draft.CoverPath,
		VideoPath:   draft.VideoPath,
		ProdURL:     deployRes.URL,
		PublishedBy: moderatorID,
	}

	if _, err := s.releaseRepo.Create(ctx, rel); err != nil {
		return nil, fmt.Errorf("ModerationService.Approve create release: %w", err)
	}

	// 3. Обновление статусов
	_ = s.projectRepo.UpdateStatus(ctx, projectID, domain.ProjectStatusPublished)
	_ = s.moderationRepo.ResolveTicket(ctx, ticket.ID, domain.ModerationStatusApproved, comment, moderatorID)

	_ = s.deploymentRepo.Create(ctx, &domain.DeploymentRecord{
		ProjectID:   projectID,
		Environment: domain.DeploymentEnvProd,
		Version:     draft.ActiveBuildVersion,
		Status:      domain.DeploymentStatusSuccess,
	})

	return rel, nil
}

// Reject отклоняет проект модератором с указанием причин.
func (s *ModerationService) Reject(ctx context.Context, projectID int64, moderatorID, reason string) (*domain.ModerationTicket, error) {
	ticket, err := s.moderationRepo.GetLatestTicketByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("ModerationService.Reject get ticket: %w", err)
	}

	if err := s.moderationRepo.ResolveTicket(ctx, ticket.ID, domain.ModerationStatusRejected, reason, moderatorID); err != nil {
		return nil, fmt.Errorf("ModerationService.Reject resolve ticket: %w", err)
	}

	_ = s.projectRepo.UpdateStatus(ctx, projectID, domain.ProjectStatusRejected)

	ticket.Status = domain.ModerationStatusRejected
	ticket.RejectionReason = reason
	ticket.ModeratorID = moderatorID

	return ticket, nil
}
