package service

import (
	"context"
	"fmt"
	"net"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
)

// instanceStarter описывает методы InstanceService, нужные DiscoveryService.
type instanceStarter interface {
	StartInstance(ctx context.Context, params StartInstanceParams) (*domain.Instance, error)
}

// DiscoveryService предоставляет данные для подключения к игровым серверам.
type DiscoveryService struct {
	instanceRepo  domain.InstanceRepo
	instanceState domain.InstanceStateStore
	nodeRepo      domain.NodeRepo
	buildRepo     domain.BuildStorage
	policyService *GamePolicyService
	instanceSvc   instanceStarter
}

// NewDiscoveryService создаёт сервис обнаружения серверов.
func NewDiscoveryService(
	instanceRepo domain.InstanceRepo,
	instanceState domain.InstanceStateStore,
	nodeRepo domain.NodeRepo,
	buildRepo domain.BuildStorage,
	policyService *GamePolicyService,
	instanceSvc instanceStarter,
) *DiscoveryService {
	return &DiscoveryService{
		instanceRepo:  instanceRepo,
		instanceState: instanceState,
		nodeRepo:      nodeRepo,
		buildRepo:     buildRepo,
		policyService: policyService,
		instanceSvc:   instanceSvc,
	}
}

// DiscoverServers возвращает список доступных серверов для подключения
// вместе со статусом, который помогает клиенту понять текущую ситуацию.
//
// Сценарии:
//   - READY — есть running-инстансы со свободными слотами.
//   - STARTING — нет свободных running, но инстансы стартуют (или запущен асинхронный auto-start).
//   - CAPACITY_REACHED — все инстансы заполнены, новых запустить нельзя (лимит или queue-режим).
//   - UNAVAILABLE — авто-оркестрация выключена или не настроена.
func (s *DiscoveryService) DiscoverServers(ctx context.Context, gameID int64) (*domain.DiscoveryResult, error) {
	// 1. Получаем running-инстансы.
	runningStatus := domain.InstanceStatusRunning
	running, err := s.instanceRepo.ListByGame(ctx, gameID, &runningStatus)
	if err != nil {
		return nil, fmt.Errorf("DiscoveryService.DiscoverServers: list running instances: %w", err)
	}

	// 2. Получаем starting-инстансы.
	startingStatus := domain.InstanceStatusStarting
	starting, _ := s.instanceRepo.ListByGame(ctx, gameID, &startingStatus)

	// 3. Формируем endpoints из running и проверяем наличие свободных слотов.
	endpoints := make([]domain.ServerEndpoint, 0, len(running))
	hasAvailable := false

	for _, inst := range running {
		playerCount, _ := s.instanceState.GetPlayerCount(ctx, inst.ID)

		node, err := s.nodeRepo.GetByID(ctx, inst.NodeID)
		if err != nil {
			continue // Нода не найдена — пропускаем инстанс.
		}

		host, _, err := net.SplitHostPort(node.Address)
		if err != nil {
			host = node.Address
		}

		endpoints = append(endpoints, domain.ServerEndpoint{
			InstanceID:  inst.ID,
			Address:     host,
			Port:        inst.HostPort,
			Protocol:    inst.Protocol,
			PlayerCount: &playerCount,
			MaxPlayers:  inst.MaxPlayers,
		})

		if playerCount < inst.MaxPlayers {
			hasAvailable = true
		}
	}

	// 4. Есть доступные running → READY.
	if hasAvailable {
		sortByPlayerCount(endpoints)
		return &domain.DiscoveryResult{
			Status:  domain.DiscoveryStatusReady,
			Servers: endpoints,
		}, nil
	}

	// 5. Нет доступных running, но есть starting → STARTING.
	if len(starting) > 0 {
		return &domain.DiscoveryResult{
			Status:  domain.DiscoveryStatusStarting,
			Servers: endpoints,
			Message: "Server instances are starting, please wait",
		}, nil
	}

	// 6. Нет running и нет starting. Проверяем политику.
	policy, err := s.policyService.Get(ctx, gameID)
	if err != nil || policy == nil {
		return &domain.DiscoveryResult{
			Status:  domain.DiscoveryStatusUnavailable,
			Servers: endpoints,
			Message: "Game is not configured for auto-orchestration",
		}, nil
	}

	if !policy.IsAuto() {
		return &domain.DiscoveryResult{
			Status:  domain.DiscoveryStatusUnavailable,
			Servers: endpoints,
			Message: "Auto-start is disabled for this game",
		}, nil
	}

	// 7. Политика auto. Проверяем, можем ли запустить новый инстанс.
	canStart, reason := s.canAutoStart(ctx, gameID, policy)
	if !canStart {
		return &domain.DiscoveryResult{
			Status:  domain.DiscoveryStatusCapacityReached,
			Servers: endpoints,
			Message: reason,
		}, nil
	}

	// 8. Если running есть, но все заполнены — решаем по scale_behavior.
	if len(running) > 0 && policy.ScaleBehavior == domain.ScaleBehaviorQueue {
		return &domain.DiscoveryResult{
			Status:  domain.DiscoveryStatusCapacityReached,
			Servers: endpoints,
			Message: "All servers are full. Waiting in queue.",
		}, nil
	}

	// 9. Запускаем асинхронно.
	go s.autoStartInstance(context.Background(), gameID, policy)

	return &domain.DiscoveryResult{
		Status:  domain.DiscoveryStatusStarting,
		Servers: endpoints,
		Message: "Spinning up a new server instance",
	}, nil
}

// canAutoStart выполняет быструю синхронную проверку возможности запуска
// без сетевых вызовов (лимит инстансов и наличие билда).
func (s *DiscoveryService) canAutoStart(ctx context.Context, gameID int64, policy *domain.GamePolicy) (bool, string) {
	// Проверка лимита инстансов.
	all, _ := s.instanceRepo.ListByGame(ctx, gameID, nil)
	if int32(len(all)) >= policy.MaxInstancesPerGame {
		return false, "Maximum instance limit reached for this game"
	}

	// Проверка наличия билда.
	buildVersion := policy.DefaultBuildVersion
	if buildVersion == "latest" || buildVersion == "" {
		builds, err := s.buildRepo.ListByGame(ctx, gameID, 1)
		if err != nil || len(builds) == 0 {
			return false, "No server builds available for this game"
		}
	} else {
		_, err := s.buildRepo.GetByVersion(ctx, gameID, buildVersion)
		if err != nil {
			return false, "Default build version not found"
		}
	}

	return true, ""
}

// autoStartInstance запускает инстанс для игры на основе политики.
// Вызывается асинхронно из DiscoverServers.
func (s *DiscoveryService) autoStartInstance(ctx context.Context, gameID int64, policy *domain.GamePolicy) {
	// Дополнительная проверка лимита (race condition).
	all, _ := s.instanceRepo.ListByGame(ctx, gameID, nil)
	if int32(len(all)) >= policy.MaxInstancesPerGame {
		return
	}

	buildVersion := policy.DefaultBuildVersion
	if buildVersion == "latest" || buildVersion == "" {
		builds, err := s.buildRepo.ListByGame(ctx, gameID, 1)
		if err == nil && len(builds) > 0 {
			buildVersion = builds[0].Version
		} else {
			return
		}
	}

	params := StartInstanceParams{
		GameID:         gameID,
		OwnerID:        policy.OwnerID,
		BuildVersion:   buildVersion,
		NodePreference: policy.NodePreference,
	}
	if policy.MaxPlayersPerInstance > 0 {
		mp := uint32(policy.MaxPlayersPerInstance)
		params.MaxPlayers = &mp
	}

	_, _ = s.instanceSvc.StartInstance(ctx, params)
}

// sortByPlayerCount сортирует эндпоинты по возрастанию player_count.
func sortByPlayerCount(endpoints []domain.ServerEndpoint) {
	for i := 0; i < len(endpoints); i++ {
		for j := i + 1; j < len(endpoints); j++ {
			a := uint32(0)
			if endpoints[i].PlayerCount != nil {
				a = *endpoints[i].PlayerCount
			}
			b := uint32(0)
			if endpoints[j].PlayerCount != nil {
				b = *endpoints[j].PlayerCount
			}
			if a > b {
				endpoints[i], endpoints[j] = endpoints[j], endpoints[i]
			}
		}
	}
}
