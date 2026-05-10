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

// DiscoverServers возвращает список доступных серверов для подключения.
// Серверы сортируются по принципу least-loaded (меньше игроков — выше приоритет).
// Если нет запущенных инстансов и политика игры требует автозапуска,
// инициирует асинхронный запуск инстанса (для scale-to-zero и keep-alive).
func (s *DiscoveryService) DiscoverServers(ctx context.Context, gameID int64) ([]domain.ServerEndpoint, error) {
	// Получаем все running-инстансы игры.
	status := domain.InstanceStatusRunning
	instances, err := s.instanceRepo.ListByGame(ctx, gameID, &status)
	if err != nil {
		return nil, fmt.Errorf("DiscoveryService.DiscoverServers: list instances: %w", err)
	}

	// Если нет running инстансов — проверяем политику и пытаемся запустить.
	if len(instances) == 0 {
		policy, err := s.policyService.Get(ctx, gameID)
		if err == nil && policy.IsAuto() {
			// Асинхронный запуск, чтобы не блокировать ответ.
			go s.autoStartInstance(context.Background(), gameID, policy)
		}
	}

	endpoints := make([]domain.ServerEndpoint, 0, len(instances))
	for _, inst := range instances {
		// Обогащаем из KV.
		playerCount, _ := s.instanceState.GetPlayerCount(ctx, inst.ID)

		node, err := s.nodeRepo.GetByID(ctx, inst.NodeID)
		if err != nil {
			// Если нода не найдена — пропускаем инстанс.
			continue
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
	}

	// Сортировка по least-loaded (наименьшее количество игроков первым).
	sortByPlayerCount(endpoints)

	return endpoints, nil
}

// autoStartInstance запускает инстанс для игры на основе политики.
// Вызывается асинхронно из DiscoverServers.
func (s *DiscoveryService) autoStartInstance(ctx context.Context, gameID int64, policy *domain.GamePolicy) {
	// Проверяем лимит инстансов из политики (все статусы).
	all, _ := s.instanceRepo.ListByGame(ctx, gameID, nil)
	if int32(len(all)) >= policy.MaxInstancesPerGame {
		return // Лимит достигнут.
	}

	buildVersion := policy.DefaultBuildVersion
	if buildVersion == "latest" || buildVersion == "" {
		// Пытаемся найти последний билд.
		builds, err := s.buildRepo.ListByGame(ctx, gameID, 1)
		if err == nil && len(builds) > 0 {
			buildVersion = builds[0].Version
		} else {
			return // Нет билдов для запуска.
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
