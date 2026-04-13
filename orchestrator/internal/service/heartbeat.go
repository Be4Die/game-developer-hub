package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/config"
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

	for _, node := range nodes {
		if node.Status == domain.NodeStatusMaintenance {
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
	// Попытка gRPC Heartbeat.
	usage, err := s.nodeClient.Heartbeat(ctx, node.Address)
	if err != nil {
		// Проверяем таймаут неактивности.
		timeSincePing := time.Since(node.LastPingAt)
		if timeSincePing > s.inactivityTimeout && node.Status == domain.NodeStatusOnline {
			s.log.Error("node timed out, marking as offline",
				slog.Int64("node_id", node.ID),
				slog.String("address", node.Address),
				slog.Duration("since_last_ping", timeSincePing),
			)

			if markErr := s.markNodeOffline(ctx, node); markErr != nil {
				return errors.Join(
					fmt.Errorf("checkNode: mark offline: %w", markErr),
					fmt.Errorf("checkNode: heartbeat: %w", err),
				)
			}
		}
		return fmt.Errorf("heartbeat rpc: %w", err)
	}

	// Обновляем last_ping_at в PG.
	if err := s.nodeRepo.UpdateLastPing(ctx, node.ID); err != nil {
		s.log.Warn("failed to update last ping",
			slog.Int64("node_id", node.ID),
			slog.String("error", err.Error()),
		)
	}

	// Обновляем состояние в KV.
	if err := s.nodeState.UpdateHeartbeat(ctx, node.ID, usage); err != nil {
		s.log.Warn("failed to update node state in KV",
			slog.Int64("node_id", node.ID),
			slog.String("error", err.Error()),
		)
	}

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
