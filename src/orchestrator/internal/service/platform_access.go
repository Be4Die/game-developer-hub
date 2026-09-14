package service

import (
	"context"
	"fmt"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
)

// PlatformAccessService управляет выдачей и отзывом квот на платформенные серверы.
type PlatformAccessService struct {
	repo domain.PlatformAccessRepo
}

// NewPlatformAccessService создаёт новый сервис управления платформенными квотами.
func NewPlatformAccessService(repo domain.PlatformAccessRepo) *PlatformAccessService {
	return &PlatformAccessService{repo: repo}
}

// GrantAccess выдает или обновляет квоту доступа к платформенным мощностям для проекта.
func (s *PlatformAccessService) GrantAccess(
	ctx context.Context,
	projectID int64,
	maxInstances int32,
	maxTotalCPU uint32,
	maxTotalMemoryMB uint64,
	maxInstanceCPU uint32,
	maxInstanceMemoryMB uint64,
) (*domain.PlatformGrant, error) {
	if projectID <= 0 {
		return nil, fmt.Errorf("invalid project_id: %d", projectID)
	}
	if maxInstances <= 0 {
		maxInstances = 2
	}
	return s.repo.SaveGrant(ctx, &domain.PlatformGrant{
		GameID:               projectID,
		MaxInstances:         maxInstances,
		MaxTotalCPUMillis:    maxTotalCPU,
		MaxTotalMemoryMB:     maxTotalMemoryMB,
		MaxInstanceCPUMillis: maxInstanceCPU,
		MaxInstanceMemoryMB:  maxInstanceMemoryMB,
		IsActive:             true,
	})
}

// RevokeAccess отзывает доступ к платформенным серверам для проекта.
func (s *PlatformAccessService) RevokeAccess(ctx context.Context, projectID int64) error {
	if projectID <= 0 {
		return fmt.Errorf("invalid project_id: %d", projectID)
	}
	return s.repo.RevokeGrant(ctx, projectID)
}

// GetGrant возвращает грант проекта и текущее число активных инстансов.
func (s *PlatformAccessService) GetGrant(ctx context.Context, projectID int64) (*domain.PlatformGrant, int32, error) {
	if projectID <= 0 {
		return nil, 0, fmt.Errorf("invalid project_id: %d", projectID)
	}
	grant, err := s.repo.GetGrant(ctx, projectID)
	if err != nil {
		return nil, 0, err
	}
	activeCount, _ := s.repo.CountPlatformInstances(ctx, projectID)
	return grant, activeCount, nil
}
