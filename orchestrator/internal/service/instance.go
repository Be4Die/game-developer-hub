package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/config"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
)

// InstanceService управляет жизненным циклом экземпляров игровых серверов.
type InstanceService struct {
	instanceRepo  domain.InstanceRepo
	instanceState domain.InstanceStateStore
	buildRepo     domain.BuildStorage
	nodeRepo      domain.NodeRepo
	nodeState     domain.NodeStateStore
	nodeClient    domain.NodeClient
	limits        config.LimitsConfig
}

// NewInstanceService создаёт сервис управления инстансами.
func NewInstanceService(
	instanceRepo domain.InstanceRepo,
	instanceState domain.InstanceStateStore,
	buildRepo domain.BuildStorage,
	nodeRepo domain.NodeRepo,
	nodeState domain.NodeStateStore,
	nodeClient domain.NodeClient,
	limits config.LimitsConfig,
) *InstanceService {
	return &InstanceService{
		instanceRepo:  instanceRepo,
		instanceState: instanceState,
		buildRepo:     buildRepo,
		nodeRepo:      nodeRepo,
		nodeState:     nodeState,
		nodeClient:    nodeClient,
		limits:        limits,
	}
}

// StartInstanceParams содержит параметры запуска инстанса.
type StartInstanceParams struct {
	OwnerID          string
	GameID           int64
	BuildVersion     string
	Name             string
	PortAllocation   domain.PortAllocation
	ResourceLimits   *domain.ResourceLimits
	EnvVars          map[string]string
	Args             []string
	DeveloperPayload map[string]string
	MaxPlayers       *uint32
}

// StartInstance запускает новый экземпляр игрового сервера на доступной ноде.
// Проверяет лимит инстансов на игру, выбирает ноду с наименьшей загрузкой,
// запускает контейнер через gRPC и регистрирует метаданные.
// При ошибке сохраняет целостность: откатывает запуск на ноде, если не удалось сохранить в PG.
// Возвращает ErrNoAvailableNode при отсутствии свободных нод, ErrNotFound при отсутствии билда.
func (s *InstanceService) StartInstance(ctx context.Context, params StartInstanceParams) (*domain.Instance, error) {
	// Шаг 1: проверка лимита инстансов на игру.
	count, err := s.instanceRepo.CountByGame(ctx, params.GameID)
	if err != nil {
		return nil, fmt.Errorf("InstanceService.StartInstance: count instances: %w", err)
	}
	if count >= s.limits.MaxInstancesPerGame {
		return nil, fmt.Errorf("InstanceService.StartInstance: max instances limit reached (%d)", s.limits.MaxInstancesPerGame)
	}

	// Шаг 2: загрузка билда.
	build, err := s.buildRepo.GetByVersion(ctx, params.GameID, params.BuildVersion)
	if err != nil {
		return nil, fmt.Errorf("InstanceService.StartInstance: get build: %w", err)
	}

	// Шаг 3: выбор ноды.
	node, err := s.selectNodeForInstance(ctx, build)
	if err != nil {
		return nil, fmt.Errorf("InstanceService.StartInstance: select node: %w", err)
	}

	// Обновляем адрес ноды для записи.
	nodeAddr := node.Address

	// Шаг 4: запуск инстанса на ноде через gRPC.
	maxPlayers := build.MaxPlayers
	if params.MaxPlayers != nil {
		maxPlayers = *params.MaxPlayers
	}

	startReq := domain.StartInstanceRequest{
		GameID:           params.GameID,
		Name:             params.Name,
		Protocol:         build.Protocol,
		InternalPort:     build.InternalPort,
		PortAllocation:   params.PortAllocation,
		MaxPlayers:       maxPlayers,
		DeveloperPayload: params.DeveloperPayload,
		EnvVars:          params.EnvVars,
		Args:             params.Args,
		ResourceLimits:   params.ResourceLimits,
	}

	result, err := s.nodeClient.StartInstance(ctx, nodeAddr, node.APIToken, startReq)
	if err != nil {
		return nil, fmt.Errorf("InstanceService.StartInstance: node StartInstance: %w", err)
	}

	// Шаг 5: запись метаданных в PG.
	now := time.Now()
	instance := &domain.Instance{
		ID:               result.InstanceID,
		OwnerID:          params.OwnerID,
		NodeID:           node.ID,
		ServerBuildID:    build.ID,
		GameID:           params.GameID,
		Name:             params.Name,
		BuildVersion:     params.BuildVersion,
		Protocol:         build.Protocol,
		HostPort:         result.HostPort,
		InternalPort:     build.InternalPort,
		Status:           domain.InstanceStatusStarting,
		MaxPlayers:       maxPlayers,
		DeveloperPayload: params.DeveloperPayload,
		ServerAddress:    nodeAddr,
		StartedAt:        now,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := s.instanceRepo.Create(ctx, instance); err != nil {
		// Откат: останавливаем инстанс на ноде.
		_ = s.nodeClient.StopInstance(ctx, nodeAddr, node.APIToken, result.InstanceID, 0)
		return nil, fmt.Errorf("InstanceService.StartInstance: save instance: %w", err)
	}

	// Шаг 6: инициализация состояния в KV.
	if err := s.instanceState.SetStatus(ctx, instance.ID, domain.InstanceStatusRunning); err != nil {
		return nil, fmt.Errorf("InstanceService.StartInstance: set KV status: %w", err)
	}

	// Обновляем счётчик активных инстансов на ноде.
	if err := s.incrementNodeInstanceCount(ctx, node.ID, 1); err != nil {
		return nil, fmt.Errorf("InstanceService.StartInstance: update node count: %w", err)
	}

	// Финализируем статус.
	instance.Status = domain.InstanceStatusRunning
	if err := s.instanceRepo.Update(ctx, instance); err != nil {
		return nil, fmt.Errorf("InstanceService.StartInstance: update status to running: %w", err)
	}

	return instance, nil
}

// StopInstance останавливает экземпляр игрового сервера, обновляет статус в PostgreSQL
// и удаляет состояние из KV. Уменьшает счётчик активных инстансов на ноде.
// Возвращает ошибку, если instanceID не принадлежит указанной игре или владельцу.
func (s *InstanceService) StopInstance(ctx context.Context, ownerID string, gameID, instanceID int64, timeoutSec uint32) (*domain.Instance, error) {
	instance, err := s.instanceRepo.GetByID(ctx, instanceID)
	if err != nil {
		return nil, fmt.Errorf("InstanceService.StopInstance: get instance: %w", err)
	}
	if instance.GameID != gameID {
		return nil, fmt.Errorf("InstanceService.StopInstance: instance %d does not belong to game %d", instanceID, gameID)
	}
	if ownerID != "" && instance.OwnerID != ownerID {
		return nil, domain.ErrForbidden
	}

	// Определяем ноду для gRPC-вызова.
	node, err := s.nodeRepo.GetByID(ctx, instance.NodeID)
	if err != nil {
		return nil, fmt.Errorf("InstanceService.StopInstance: get node: %w", err)
	}

	// gRPC остановка.
	if err := s.nodeClient.StopInstance(ctx, node.Address, node.APIToken, instanceID, timeoutSec); err != nil {
		return nil, fmt.Errorf("InstanceService.StopInstance: node StopInstance: %w", err)
	}

	// Обновление статуса в PG.
	instance.Status = domain.InstanceStatusStopped
	instance.UpdatedAt = time.Now()
	if err := s.instanceRepo.Update(ctx, instance); err != nil {
		return nil, fmt.Errorf("InstanceService.StopInstance: update status: %w", err)
	}

	// Удаление из KV.
	if err := s.instanceState.Delete(ctx, instanceID); err != nil {
		return nil, fmt.Errorf("InstanceService.StopInstance: delete KV state: %w", err)
	}

	// Уменьшаем счётчик на ноде.
	if err := s.incrementNodeInstanceCount(ctx, node.ID, -1); err != nil {
		return nil, fmt.Errorf("InstanceService.StopInstance: update node count: %w", err)
	}

	return instance, nil
}

// ListInstances возвращает список инстансов пользователя с обогащением из KV.
func (s *InstanceService) ListInstances(ctx context.Context, ownerID string, gameID int64, status *domain.InstanceStatus) ([]*EnrichedInstance, error) {
	instances, err := s.instanceRepo.ListByGame(ctx, gameID, status)
	if err != nil {
		return nil, fmt.Errorf("InstanceService.ListInstances: %w", err)
	}

	result := make([]*EnrichedInstance, 0, len(instances))
	for _, inst := range instances {
		if ownerID != "" && inst.OwnerID != ownerID {
			continue
		}

		enriched := &EnrichedInstance{Instance: inst}

		// Обогащение из KV.
		st, err := s.instanceState.GetStatus(ctx, inst.ID)
		if err == nil {
			enriched.Status = st
		}

		pc, err := s.instanceState.GetPlayerCount(ctx, inst.ID)
		if err == nil {
			enriched.PlayerCount = &pc
		}

		result = append(result, enriched)
	}

	return result, nil
}

// GetInstance возвращает инстанс с обогащением из KV.
func (s *InstanceService) GetInstance(ctx context.Context, ownerID string, gameID, instanceID int64) (*EnrichedInstance, error) {
	instance, err := s.instanceRepo.GetByID(ctx, instanceID)
	if err != nil {
		return nil, fmt.Errorf("InstanceService.GetInstance: %w", err)
	}
	if instance.GameID != gameID {
		return nil, fmt.Errorf("InstanceService.GetInstance: instance %d does not belong to game %d", instanceID, gameID)
	}
	if ownerID != "" && instance.OwnerID != ownerID {
		return nil, domain.ErrForbidden
	}

	enriched := &EnrichedInstance{Instance: instance}

	st, err := s.instanceState.GetStatus(ctx, instanceID)
	if err == nil {
		enriched.Status = st
	}

	pc, err := s.instanceState.GetPlayerCount(ctx, instanceID)
	if err == nil {
		enriched.PlayerCount = &pc
	}

	return enriched, nil
}

// StreamInstanceLogs возвращает поток журналов инстанса.
// Caller обязан закрыть stream после использования.
func (s *InstanceService) StreamInstanceLogs(ctx context.Context, ownerID string, gameID, instanceID int64, req domain.StreamLogsRequest) (domain.LogStream, error) {
	instance, err := s.instanceRepo.GetByID(ctx, instanceID)
	if err != nil {
		return nil, fmt.Errorf("InstanceService.StreamInstanceLogs: get instance: %w", err)
	}
	if instance.GameID != gameID {
		return nil, fmt.Errorf("InstanceService.StreamInstanceLogs: instance %d does not belong to game %d", instanceID, gameID)
	}
	if ownerID != "" && instance.OwnerID != ownerID {
		return nil, domain.ErrForbidden
	}

	node, err := s.nodeRepo.GetByID(ctx, instance.NodeID)
	if err != nil {
		return nil, fmt.Errorf("InstanceService.StreamInstanceLogs: get node: %w", err)
	}

	stream, err := s.nodeClient.StreamLogs(ctx, node.Address, node.APIToken, req)
	if err != nil {
		return nil, fmt.Errorf("InstanceService.StreamInstanceLogs: node StreamLogs: %w", err)
	}

	return stream, nil
}

// GetInstanceUsage возвращает метрики потребления ресурсов инстанса.
// Сначала пробует KV (быстро), fallback — gRPC на ноду.
func (s *InstanceService) GetInstanceUsage(ctx context.Context, ownerID string, gameID, instanceID int64) (*domain.ResourceUsage, error) {
	// Пробуем KV.
	usage, err := s.instanceState.GetUsage(ctx, instanceID)
	if err == nil {
		return usage, nil
	}

	// Fallback: gRPC на ноду.
	instance, err := s.instanceRepo.GetByID(ctx, instanceID)
	if err != nil {
		return nil, fmt.Errorf("InstanceService.GetInstanceUsage: get instance: %w", err)
	}
	if instance.GameID != gameID {
		return nil, fmt.Errorf("InstanceService.GetInstanceUsage: instance %d does not belong to game %d", instanceID, gameID)
	}
	if ownerID != "" && instance.OwnerID != ownerID {
		return nil, domain.ErrForbidden
	}

	node, err := s.nodeRepo.GetByID(ctx, instance.NodeID)
	if err != nil {
		return nil, fmt.Errorf("InstanceService.GetInstanceUsage: get node: %w", err)
	}

	usage, err = s.nodeClient.GetInstanceUsage(ctx, node.Address, node.APIToken, instanceID)
	if err != nil {
		return nil, fmt.Errorf("InstanceService.GetInstanceUsage: node GetInstanceUsage: %w", err)
	}

	return usage, nil
}

// EnrichedInstance — инстанс с данными из KV.
type EnrichedInstance struct {
	*domain.Instance
	Status      domain.InstanceStatus // из KV
	PlayerCount *uint32               // из KV
}

// selectNodeForInstance выбирает ноду с наименьшей загрузкой.
func (s *InstanceService) selectNodeForInstance(ctx context.Context, _ *domain.ServerBuild) (*domain.Node, error) {
	nodes, err := s.nodeRepo.List(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("selectNodeForInstance: list nodes: %w", err)
	}

	var best *domain.Node
	bestLoad := ^uint32(0)

	for _, n := range nodes {
		if n.Status != domain.NodeStatusOnline {
			continue
		}

		load, err := s.nodeState.GetActiveInstanceCount(ctx, n.ID)
		if err != nil {
			load = 0
		}

		if load < bestLoad {
			bestLoad = load
			best = n
		}
	}

	if best == nil {
		return nil, domain.ErrNoAvailableNode
	}

	return best, nil
}

// incrementNodeInstanceCount изменяет счётчик активных инстансов на ноде.
func (s *InstanceService) incrementNodeInstanceCount(ctx context.Context, nodeID int64, delta int) error {
	current, err := s.nodeState.GetActiveInstanceCount(ctx, nodeID)
	if err != nil {
		current = 0
	}

	var newVal uint32
	if delta < 0 {
		d := uint32(-delta) //nolint:gosec // delta проверяется на < 0
		if d > current {
			newVal = 0
		} else {
			newVal = current - d
		}
	} else {
		newVal = current + uint32(delta) //nolint:gosec // delta >= 0
	}

	return s.nodeState.SetActiveInstanceCount(ctx, nodeID, newVal)
}

// LogStreamReader адаптер: domain.LogStream → io.Reader (для SSE).
type LogStreamReader struct {
	Stream domain.LogStream
}

// Read читает следующую журнальную запись в формате JSON.
func (r *LogStreamReader) Read(p []byte) (int, error) {
	entry, err := r.Stream.Recv()
	if err != nil {
		return 0, err
	}

	// Простой формат: "timestamp source message\n"
	line := fmt.Sprintf("%s [%s] %s\n",
		entry.Timestamp.Format(time.RFC3339),
		entry.Source,
		entry.Message,
	)

	return copy(p, line), nil
}
