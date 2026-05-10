package service

import (
	"context"
	"fmt"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
)

// GamePolicyService управляет политиками оркестрации серверов по проектам.
type GamePolicyService struct {
	policyRepo domain.GamePolicyRepo
}

// NewGamePolicyService создаёт сервис политик.
func NewGamePolicyService(policyRepo domain.GamePolicyRepo) *GamePolicyService {
	return &GamePolicyService{policyRepo: policyRepo}
}

// Get возвращает политику игры. Если не найдена — возвращает политику по умолчанию (disabled).
func (s *GamePolicyService) Get(ctx context.Context, gameID int64) (*domain.GamePolicy, error) {
	policy, err := s.policyRepo.Get(ctx, gameID)
	if err != nil {
		if err == domain.ErrNotFound {
			return defaultPolicy(gameID), nil
		}
		return nil, fmt.Errorf("GamePolicyService.Get: %w", err)
	}
	return policy, nil
}

// Set создаёт или обновляет политику игры.
func (s *GamePolicyService) Set(ctx context.Context, policy *domain.GamePolicy) (*domain.GamePolicy, error) {
	if err := s.policyRepo.Set(ctx, policy); err != nil {
		return nil, fmt.Errorf("GamePolicyService.Set: %w", err)
	}
	return policy, nil
}

// ListAll возвращает все сохранённые политики.
func (s *GamePolicyService) ListAll(ctx context.Context) ([]*domain.GamePolicy, error) {
	policies, err := s.policyRepo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("GamePolicyService.ListAll: %w", err)
	}
	return policies, nil
}

// defaultPolicy возвращает консервативную политику по умолчанию.
func defaultPolicy(gameID int64) *domain.GamePolicy {
	return &domain.GamePolicy{
		GameID:                gameID,
		OwnerID:               "",
		Mode:                  domain.OrchestrationModeDisabled,
		TargetInstances:       1,
		AutoRestart:           false,
		ScaleToZeroTimeout:    10,
		DefaultBuildVersion:   "latest",
		MaxPlayersPerInstance: 100,
		MaxInstancesPerGame:   1,
		ScaleBehavior:         domain.ScaleBehaviorSpawn,
		NodePreference:        "auto",
	}
}
