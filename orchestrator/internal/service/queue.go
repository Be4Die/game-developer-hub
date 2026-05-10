package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
)

// QueueService управляет очередью игроков.
type QueueService struct {
	store          domain.QueueStore
	eventRepo      domain.QueueEventRepo
	policySvc      *GamePolicyService
	instanceRepo   domain.InstanceRepo
	instanceState  domain.InstanceStateStore
	nodeRepo       domain.NodeRepo
	log            *slog.Logger
}

// NewQueueService создаёт сервис очереди.
func NewQueueService(
	store domain.QueueStore,
	eventRepo domain.QueueEventRepo,
	policySvc *GamePolicyService,
	instanceRepo domain.InstanceRepo,
	instanceState domain.InstanceStateStore,
	nodeRepo domain.NodeRepo,
	log *slog.Logger,
) *QueueService {
	return &QueueService{
		store:         store,
		eventRepo:     eventRepo,
		policySvc:     policySvc,
		instanceRepo:  instanceRepo,
		instanceState: instanceState,
		nodeRepo:      nodeRepo,
		log:           log,
	}
}

// Join добавляет игрока в очередь.
func (s *QueueService) Join(ctx context.Context, gameID int64, playerID, mode string) (*QueueStatusResult, error) {
	// Очистка просроченных перед join
	policy, err := s.policySvc.Get(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("QueueService.Join: get policy: %w", err)
	}
	_, _ = s.CleanupExpired(ctx, gameID, policy)

	if err := s.store.Join(ctx, gameID, playerID, mode); err != nil {
		return nil, fmt.Errorf("QueueService.Join: %w", err)
	}

	if err := s.eventRepo.Log(ctx, gameID, playerID, domain.QueueEventJoin, 0, 0); err != nil {
		s.log.Warn("queue event log failed", slog.String("error", err.Error()))
	}

	pos, total, err := s.store.GetPosition(ctx, gameID, playerID)
	if err != nil {
		return nil, fmt.Errorf("QueueService.Join: get position: %w", err)
	}

	return &QueueStatusResult{
		Status:               domain.QueueStatusWaiting,
		Position:             int32(pos),
		TotalInQueue:         int32(total),
		EstimatedWaitSeconds: s.estimateWait(pos),
	}, nil
}

// Heartbeat обновляет heartbeat и возвращает текущий статус.
func (s *QueueService) Heartbeat(ctx context.Context, gameID int64, playerID string) (*QueueStatusResult, error) {
	policy, err := s.policySvc.Get(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("QueueService.Heartbeat: get policy: %w", err)
	}

	// Проверяем, не истекла ли резервация
	if endpoint, expiresAt, err := s.store.GetReservation(ctx, gameID, playerID); err == nil {
		if time.Now().After(expiresAt) {
			// Reservation истёк — выкидываем
			_ = s.store.Leave(ctx, gameID, playerID)
			_ = s.eventRepo.Log(ctx, gameID, playerID, domain.QueueEventTimeout, 0, 0)
			return &QueueStatusResult{Status: domain.QueueStatusExpired}, nil
		}
		return &QueueStatusResult{
			Status:               domain.QueueStatusReserved,
			ReservedEndpoint:     endpoint,
			ReservedUntil:        expiresAt,
		}, nil
	}

	// Обычный heartbeat
	if err := s.store.Heartbeat(ctx, gameID, playerID); err != nil {
		if err == domain.ErrNotFound {
			return &QueueStatusResult{Status: domain.QueueStatusExpired}, nil
		}
		return nil, fmt.Errorf("QueueService.Heartbeat: %w", err)
	}

	// Проверяем, не превысил ли max_wait
	entry, err := s.store.ListQueue(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("QueueService.Heartbeat: list queue: %w", err)
	}

	for _, e := range entry {
		if e.PlayerID == playerID {
			if time.Since(e.JoinTime) > time.Duration(policy.QueueMaxWaitSec)*time.Second {
				_ = s.store.Leave(ctx, gameID, playerID)
				_ = s.eventRepo.Log(ctx, gameID, playerID, domain.QueueEventTimeout, 0, int(time.Since(e.JoinTime).Seconds()))
				return &QueueStatusResult{Status: domain.QueueStatusExpired}, nil
			}
			break
		}
	}

	pos, total, err := s.store.GetPosition(ctx, gameID, playerID)
	if err != nil {
		return nil, fmt.Errorf("QueueService.Heartbeat: get position: %w", err)
	}

	return &QueueStatusResult{
		Status:               domain.QueueStatusWaiting,
		Position:             int32(pos),
		TotalInQueue:         int32(total),
		EstimatedWaitSeconds: s.estimateWait(pos),
	}, nil
}

// Leave удаляет игрока из очереди.
func (s *QueueService) Leave(ctx context.Context, gameID int64, playerID string) error {
	if err := s.store.Leave(ctx, gameID, playerID); err != nil {
		return fmt.Errorf("QueueService.Leave: %w", err)
	}
	_ = s.eventRepo.Log(ctx, gameID, playerID, domain.QueueEventLeave, 0, 0)
	return nil
}

// Status возвращает статус без обновления heartbeat (read-only).
func (s *QueueService) Status(ctx context.Context, gameID int64, playerID string) (*QueueStatusResult, error) {
	// Проверяем резервацию
	if endpoint, expiresAt, err := s.store.GetReservation(ctx, gameID, playerID); err == nil {
		if time.Now().After(expiresAt) {
			_ = s.store.Leave(ctx, gameID, playerID)
			return &QueueStatusResult{Status: domain.QueueStatusExpired}, nil
		}
		return &QueueStatusResult{
			Status:           domain.QueueStatusReserved,
			ReservedEndpoint: endpoint,
			ReservedUntil:    expiresAt,
		}, nil
	}

	pos, total, err := s.store.GetPosition(ctx, gameID, playerID)
	if err != nil {
		if err == domain.ErrNotFound {
			return &QueueStatusResult{Status: domain.QueueStatusExpired}, nil
		}
		return nil, fmt.Errorf("QueueService.Status: %w", err)
	}

	return &QueueStatusResult{
		Status:               domain.QueueStatusWaiting,
		Position:             int32(pos),
		TotalInQueue:         int32(total),
		EstimatedWaitSeconds: s.estimateWait(pos),
	}, nil
}

// ProcessQueue вызывается при освобождении слота.
// Резервирует слот для первого игрока в очереди.
func (s *QueueService) ProcessQueue(ctx context.Context, gameID int64) error {
	policy, err := s.policySvc.Get(ctx, gameID)
	if err != nil {
		return fmt.Errorf("QueueService.ProcessQueue: get policy: %w", err)
	}

	// Очистка просроченных
	expired, _ := s.CleanupExpired(ctx, gameID, policy)
	if len(expired) > 0 {
		s.log.Debug("queue cleanup expired", slog.Int("count", len(expired)), slog.Int64("game_id", gameID))
	}

	// Проверяем, есть ли очередь
	count, err := s.store.Count(ctx, gameID)
	if err != nil || count == 0 {
		return nil // очередь пуста
	}

	// Ищем running инстанс со свободными слотами
	status := domain.InstanceStatusRunning
	instances, err := s.instanceRepo.ListByGame(ctx, gameID, &status)
	if err != nil {
		return fmt.Errorf("QueueService.ProcessQueue: list instances: %w", err)
	}

	var available *domain.Instance
	for _, inst := range instances {
		pc, _ := s.instanceState.GetPlayerCount(ctx, inst.ID)
		if pc < inst.MaxPlayers {
			available = inst
			break
		}
	}

	if available == nil {
		// Нет доступных инстансов — ждём scale up
		return nil
	}

	// Формируем endpoint
	node, err := s.nodeRepo.GetByID(ctx, available.NodeID)
	if err != nil {
		return fmt.Errorf("QueueService.ProcessQueue: get node: %w", err)
	}

	// Используем server_address из инстанса если есть, иначе из ноды
	host := available.ServerAddress
	if host == "" {
		host = node.Address
	}

	endpoint := &domain.ServerEndpoint{
		InstanceID: available.ID,
		Address:    host,
		Port:       available.HostPort,
		Protocol:   available.Protocol,
		MaxPlayers: available.MaxPlayers,
	}

	// Резервируем для первого игрока
	playerID, err := s.store.Reserve(ctx, gameID, endpoint, time.Duration(policy.QueueReservationSec)*time.Second)
	if err != nil {
		if err == domain.ErrNotFound {
			return nil // очередь опустела
		}
		return fmt.Errorf("QueueService.ProcessQueue: reserve: %w", err)
	}

	_ = s.eventRepo.Log(ctx, gameID, playerID, domain.QueueEventReserved, available.ID, 0)
	s.log.Info("queue slot reserved",
		slog.Int64("game_id", gameID),
		slog.String("player_id", playerID),
		slog.Int64("instance_id", available.ID),
	)

	return nil
}

// CleanupExpired удаляет игроков с просроченным heartbeat.
func (s *QueueService) CleanupExpired(ctx context.Context, gameID int64, policy *domain.GamePolicy) ([]string, error) {
	if policy == nil {
		var err error
		policy, err = s.policySvc.Get(ctx, gameID)
		if err != nil {
			return nil, fmt.Errorf("QueueService.CleanupExpired: get policy: %w", err)
		}
	}
	expired, err := s.store.CleanupExpired(ctx, gameID, time.Duration(policy.QueueHeartbeatTimeout)*time.Second)
	if err != nil {
		return nil, fmt.Errorf("QueueService.CleanupExpired: %w", err)
	}
	for _, playerID := range expired {
		_ = s.eventRepo.Log(ctx, gameID, playerID, domain.QueueEventTimeout, 0, 0)
	}
	return expired, nil
}

// estimateWait оценивает время ожидания (грубая эвристика).
func (s *QueueService) estimateWait(position int64) int32 {
	// Примерно 20 сек на игрока впереди (можно улучшить по статистике)
	return int32(position * 20)
}

