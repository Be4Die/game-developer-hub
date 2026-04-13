package service

import (
	"context"
	"fmt"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
)

// DiscoveryService предоставляет данные для подключения к игровым серверам.
type DiscoveryService struct {
	instanceRepo  domain.InstanceRepo
	instanceState domain.InstanceStateStore
	nodeRepo      domain.NodeRepo
}

// NewDiscoveryService создаёт сервис обнаружения серверов.
func NewDiscoveryService(
	instanceRepo domain.InstanceRepo,
	instanceState domain.InstanceStateStore,
	nodeRepo domain.NodeRepo,
) *DiscoveryService {
	return &DiscoveryService{
		instanceRepo:  instanceRepo,
		instanceState: instanceState,
		nodeRepo:      nodeRepo,
	}
}

// DiscoverServers возвращает список доступных серверов для подключения.
// Серверы сортируются по принципу least-loaded (меньше игроков — выше приоритет).
func (s *DiscoveryService) DiscoverServers(ctx context.Context, gameID int64) ([]domain.ServerEndpoint, error) {
	// Получаем все running-инстансы игры.
	status := domain.InstanceStatusRunning
	instances, err := s.instanceRepo.ListByGame(ctx, gameID, &status)
	if err != nil {
		return nil, fmt.Errorf("DiscoveryService.DiscoverServers: list instances: %w", err)
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

		endpoints = append(endpoints, domain.ServerEndpoint{
			InstanceID:  inst.ID,
			Address:     node.Address,
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
