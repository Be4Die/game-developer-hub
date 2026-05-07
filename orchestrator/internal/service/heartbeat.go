package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/infrastructure/config"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
)

// HeartbeatService — фоновый сервис мониторинга жизнеспособности нод.
// Запускается как горутина, периодически опрашивает ноды через gRPC Heartbeat.
type HeartbeatService struct {
	nodeRepo          domain.NodeRepo
	nodeState         domain.NodeStateStore
	instanceRepo      domain.InstanceRepo
	instanceState     domain.InstanceStateStore
	nodeClient        domain.NodeClient
	checkInterval     time.Duration
	inactivityTimeout time.Duration
	log               *slog.Logger
}

// NewHeartbeatService создаёт сервис мониторинга нод.
func NewHeartbeatService(
	nodeRepo domain.NodeRepo,
	nodeState domain.NodeStateStore,
	instanceRepo domain.InstanceRepo,
	instanceState domain.InstanceStateStore,
	nodeClient domain.NodeClient,
	hb config.NodeHeartbeatCfg,
	log *slog.Logger,
) *HeartbeatService {
	return &HeartbeatService{
		nodeRepo:          nodeRepo,
		nodeState:         nodeState,
		instanceRepo:      instanceRepo,
		instanceState:     instanceState,
		nodeClient:        nodeClient,
		checkInterval:     hb.CheckInterval,
		inactivityTimeout: hb.InactivityTimeout,
		log:               log,
	}
}

// Run запускает цикл мониторинга. Блокирует вызывающую горутину до отмены контекста.
func (s *HeartbeatService) Run(ctx context.Context) {
	s.log.Info("heartbeat service started",
		slog.Duration("check_interval", s.checkInterval),
		slog.Duration("inactivity_timeout", s.inactivityTimeout),
	)

	ticker := time.NewTicker(s.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.log.Info("heartbeat service stopped")
			return
		case <-ticker.C:
			s.checkAllNodes(ctx)
		}
	}
}

func (s *HeartbeatService) checkAllNodes(ctx context.Context) {
	nodes, err := s.nodeRepo.List(ctx, nil)
	if err != nil {
		s.log.Error("failed to list nodes for heartbeat", slog.String("error", err.Error()))
		return
	}

	s.log.Debug("checking nodes for heartbeat", slog.Int("count", len(nodes)))

	for _, node := range nodes {
		if node.Status == domain.NodeStatusMaintenance {
			s.log.Debug("skipping node in maintenance", slog.Int64("node_id", node.ID))
			continue
		}

		if err := s.checkNode(ctx, node); err != nil {
			s.log.Warn("node heartbeat failed",
				slog.Int64("node_id", node.ID),
				slog.String("address", node.Address),
				slog.String("error", err.Error()),
			)
		}
	}
}

func (s *HeartbeatService) checkNode(ctx context.Context, node *domain.Node) error {
	s.log.Debug("checking node heartbeat",
		slog.Int64("node_id", node.ID),
		slog.String("address", node.Address),
	)

	// Попытка gRPC Heartbeat.
	result, err := s.nodeClient.Heartbeat(ctx, node.Address, node.APIToken)
	if err != nil {
		// Проверяем таймаут неактивности.
		timeSincePing := time.Since(node.LastPingAt)
		if timeSincePing > s.inactivityTimeout && node.Status == domain.NodeStatusOnline {
			s.log.Error("node timed out, marking as offline",
				slog.Int64("node_id", node.ID),
				slog.String("address", node.Address),
				slog.Duration("since_last_ping", timeSincePing),
			)
			_ = s.markNodeOffline(ctx, node)
		}
		// Even if heartbeat fails, try to sync instance statuses if node might still be alive.
		// This keeps KV statuses fresh during transient network issues.
		_ = s.syncInstanceStatuses(ctx, node)
		return fmt.Errorf("heartbeat rpc: %w", err)
	}

	// Extract usage and active instance count from heartbeat result
	usage := result.Usage
	activeCount := result.ActiveInstanceCount

	s.log.Debug("node heartbeat successful",
		slog.Int64("node_id", node.ID),
		slog.Int("active_instances", int(activeCount)),
	)

	// Обновляем last_ping_at в PG.
	if err := s.nodeRepo.UpdateLastPing(ctx, node.ID); err != nil {
		s.log.Warn("failed to update last ping",
			slog.Int64("node_id", node.ID),
			slog.String("error", err.Error()),
		)
	}

	// Update node state in KV
	if usage != nil {
		if err := s.nodeState.UpdateHeartbeat(ctx, node.ID, usage); err != nil {
			s.log.Warn("failed to update node state in KV",
				slog.Int64("node_id", node.ID),
				slog.String("error", err.Error()),
			)
		}
	}

	// Store active instance count returned by the node
	if activeCount > 0 {
		if err := s.nodeState.SetActiveInstanceCount(ctx, node.ID, activeCount); err != nil {
			s.log.Warn("failed to update node instance count in KV",
				slog.Int64("node_id", node.ID),
				slog.Int("count", int(activeCount)),
				slog.String("error", err.Error()),
			)
		}
	}

	// Sync instance statuses from node (refreshes KV TTL).
	_ = s.syncInstanceStatuses(ctx, node)

	// Reconcile DB instances with node reality.
	// Marks Running/Starting instances as Crashed if node no longer reports them.
	_ = s.reconcileInstances(ctx, node)

	// Обновляем статус ноды в PG если она была offline.
	if node.Status == domain.NodeStatusOffline {
		now := time.Now()
		node.Status = domain.NodeStatusOnline
		node.LastPingAt = now
		node.UpdatedAt = now
		if err := s.nodeRepo.Update(ctx, node); err != nil {
			s.log.Warn("failed to update node status to online",
				slog.Int64("node_id", node.ID),
				slog.String("error", err.Error()),
			)
		}
		s.log.Info("node came back online", slog.Int64("node_id", node.ID))
	}

	return nil
}

// RestoreInstanceStatuses восстанавливает статусы инстансов из БД в KV при старте.
// Вызывается один раз при инициализации.
func (s *HeartbeatService) RestoreInstanceStatuses(ctx context.Context) error {
	instances, err := s.instanceRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("RestoreInstanceStatuses: list instances: %w", err)
	}

	for _, inst := range instances {
		// Восстанавливаем статус в KV только для активных инстансов
		if inst.Status == domain.InstanceStatusRunning || inst.Status == domain.InstanceStatusStarting {
			if err := s.instanceState.SetStatus(ctx, inst.ID, inst.Status); err != nil {
				s.log.Warn("failed to restore instance status",
					slog.Int64("instance_id", inst.ID),
					slog.String("error", err.Error()),
				)
			}
		}
	}

	s.log.Info("instance statuses restored", slog.Int("count", len(instances)))
	return nil
}

// syncInstanceStatuses synchronizes instance statuses from node to KV store.
// Called after successful or failed heartbeat to ensure KV TTL stays fresh.
func (s *HeartbeatService) syncInstanceStatuses(ctx context.Context, node *domain.Node) error {
	instances, err := s.nodeClient.ListInstances(ctx, node.Address, node.APIToken)
	if err != nil {
		s.log.Debug("failed to list instances during sync",
			slog.Int64("node_id", node.ID),
			slog.String("error", err.Error()),
		)
		return err
	}

	s.log.Debug("syncing instance statuses",
		slog.Int64("node_id", node.ID),
		slog.Int("count", len(instances)),
	)

	for _, inst := range instances {
		if err := s.instanceState.SetStatus(ctx, inst.ID, inst.Status); err != nil {
			s.log.Debug("failed to update instance status in KV",
				slog.Int64("instance_id", inst.ID),
				slog.String("error", err.Error()),
			)
		}
	}

	return nil
}

// reconcileInstances сравнивает инстансы в БД с реальностью на ноде.
// Инстансы со статусом Running/Starting, которых нет на ноде, помечаются как Crashed.
func (s *HeartbeatService) reconcileInstances(ctx context.Context, node *domain.Node) error {
	const op = "HeartbeatService.reconcileInstances"

	// Получаем список инстансов с ноды.
	nodeInstances, err := s.nodeClient.ListInstances(ctx, node.Address, node.APIToken)
	if err != nil {
		return fmt.Errorf("%s: list from node: %w", op, err)
	}

	nodeInstanceSet := make(map[int64]struct{}, len(nodeInstances))
	for _, inst := range nodeInstances {
		nodeInstanceSet[inst.ID] = struct{}{}
	}

	// Получаем ожидаемые инстансы из БД для этой ноды.
	dbInstances, err := s.instanceRepo.ListByNode(ctx, node.ID)
	if err != nil {
		return fmt.Errorf("%s: list from db: %w", op, err)
	}

	now := time.Now()
	for _, inst := range dbInstances {
		if inst.Status != domain.InstanceStatusRunning && inst.Status != domain.InstanceStatusStarting {
			continue
		}
		if _, ok := nodeInstanceSet[inst.ID]; !ok {
			inst.Status = domain.InstanceStatusCrashed
			inst.UpdatedAt = now
			if err := s.instanceRepo.Update(ctx, inst); err != nil {
				s.log.Warn("failed to update instance status during reconcile",
					slog.Int64("instance_id", inst.ID),
					slog.String("error", err.Error()),
				)
				continue
			}
			_ = s.instanceState.SetStatus(ctx, inst.ID, domain.InstanceStatusCrashed)
			s.log.Info("instance marked as crashed during reconcile",
				slog.Int64("instance_id", inst.ID),
				slog.Int64("node_id", node.ID),
			)
		}
	}

	return nil
}

// markNodeOffline переводит ноду в offline и все её инстансы в crashed.
func (s *HeartbeatService) markNodeOffline(ctx context.Context, node *domain.Node) error {
	// Обновляем статус ноды.
	now := time.Now()
	node.Status = domain.NodeStatusOffline
	node.UpdatedAt = now
	if err := s.nodeRepo.Update(ctx, node); err != nil {
		return fmt.Errorf("markNodeOffline: update node: %w", err)
	}

	// Переводим инстансы в crashed.
	instances, err := s.instanceRepo.ListByNode(ctx, node.ID)
	if err != nil {
		return fmt.Errorf("markNodeOffline: list instances: %w", err)
	}

	for _, inst := range instances {
		inst.Status = domain.InstanceStatusCrashed
		inst.UpdatedAt = now
		_ = s.instanceRepo.Update(ctx, inst)
		_ = s.instanceState.SetStatus(ctx, inst.ID, domain.InstanceStatusCrashed)
	}

	return nil
}