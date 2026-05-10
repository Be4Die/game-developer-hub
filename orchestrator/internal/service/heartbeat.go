package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/infrastructure/config"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
)

// instanceOrchestrator описывает методы InstanceService, нужные HeartbeatService.
type instanceOrchestrator interface {
	StartInstance(ctx context.Context, params StartInstanceParams) (*domain.Instance, error)
	RestartInstance(ctx context.Context, ownerID string, gameID, instanceID int64) (*domain.Instance, error)
	StopInstance(ctx context.Context, ownerID string, gameID, instanceID int64, timeoutSec uint32) (*domain.Instance, error)
}

// HeartbeatService — фоновый сервис мониторинга жизнеспособности нод.
// Запускается как горутина, периодически опрашивает ноды через gRPC Heartbeat.
// Также применяет политики оркестрации: авто-рестарт crashed-инстансов.
type HeartbeatService struct {
	nodeRepo          domain.NodeRepo
	nodeState         domain.NodeStateStore
	instanceRepo      domain.InstanceRepo
	instanceState     domain.InstanceStateStore
	nodeClient        domain.NodeClient
	buildRepo         domain.BuildStorage
	policyService     *GamePolicyService
	instanceSvc       instanceOrchestrator
	queueSvc          *QueueService
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
	buildRepo domain.BuildStorage,
	policyService *GamePolicyService,
	instanceSvc instanceOrchestrator,
	queueSvc *QueueService,
	hb config.NodeHeartbeatCfg,
	log *slog.Logger,
) *HeartbeatService {
	return &HeartbeatService{
		nodeRepo:          nodeRepo,
		nodeState:         nodeState,
		instanceRepo:      instanceRepo,
		instanceState:     instanceState,
		nodeClient:        nodeClient,
		buildRepo:         buildRepo,
		policyService:     policyService,
		instanceSvc:       instanceSvc,
		queueSvc:          queueSvc,
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

	// Применяем политики оркестрации после проверки всех нод.
	s.EnforcePolicies(ctx)
	s.enforceScaleToZero(ctx)
	s.enforceScaleUp(ctx)
	s.processQueues(ctx)
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
		if inst.PlayerCount != nil {
			if err := s.instanceState.SetPlayerCount(ctx, inst.ID, *inst.PlayerCount); err != nil {
				s.log.Debug("failed to update player count in KV",
					slog.Int64("instance_id", inst.ID),
					slog.String("error", err.Error()),
				)
			}
		}
	}

	return nil
}

// reconcileInstances сравнивает инстансы в БД с реальностью на ноде.
// Инстансы со статусом Running/Starting, которых нет на ноде, помечаются как Crashed.
// Если политика игры требует auto_restart — перезапускает crashed инстанс.
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

			// Проверяем политику и перезапускаем если нужно.
			go s.maybeAutoRestart(context.Background(), inst)
		}
	}

	return nil
}

// maybeAutoRestart проверяет политику игры и перезапускает инстанс при auto_restart.
// Сначала пробует быстрый docker restart; если контейнер уже удалён — создаёт новый инстанс.
func (s *HeartbeatService) maybeAutoRestart(ctx context.Context, inst *domain.Instance) {
	policy, err := s.policyService.Get(ctx, inst.GameID)
	if err != nil {
		s.log.Warn("failed to get policy for auto-restart",
			slog.Int64("game_id", inst.GameID),
			slog.String("error", err.Error()),
		)
		return
	}

	if !policy.AutoRestart {
		return
	}

	// Сначала пробуем быстрый docker restart (работает если контейнер ещё на ноде).
	_, err = s.instanceSvc.RestartInstance(ctx, "", inst.GameID, inst.ID)
	if err == nil {
		s.log.Info("instance auto-restarted",
			slog.Int64("instance_id", inst.ID),
			slog.Int64("game_id", inst.GameID),
		)
		return
	}

	// Если restart не сработал (контейнер удалён при краше), создаём новый инстанс.
	s.log.Warn("auto-restart (restart) failed, starting new instance",
		slog.Int64("instance_id", inst.ID),
		slog.Int64("game_id", inst.GameID),
		slog.String("error", err.Error()),
	)
	s.autoStartInstance(ctx, inst.GameID, policy)
}

// EnforcePolicies применяет политики оркестрации.
// Для игр в режиме keep_alive проверяет наличие target_instances и запускает недостающие.
// Также обрабатывает scale_to_zero для поддержки target_instances.
// Может вызываться как из heartbeat-цикла, так и при старте сервера.
func (s *HeartbeatService) EnforcePolicies(ctx context.Context) {
	policies, err := s.policyService.ListAll(ctx)
	if err != nil {
		s.log.Warn("enforcePolicies: failed to list policies", slog.String("error", err.Error()))
		return
	}

	for _, policy := range policies {
		if !policy.IsAuto() {
			continue
		}

		// Считаем ВСЕ инстансы (включая stopped), чтобы отличить ручную остановку
		// от отсутствия инстансов. Если total >= target — значит разработчик
		// остановил вручную, система не поднимает новые.
		allInstances, err := s.instanceRepo.ListByGame(ctx, policy.GameID, nil)
		if err != nil {
			s.log.Warn("enforcePolicies: failed to list instances",
				slog.Int64("game_id", policy.GameID),
				slog.String("error", err.Error()),
			)
			continue
		}

		totalCount := len(allInstances)
		if totalCount >= int(policy.TargetInstances) {
			// Достаточно инстансов (включая остановленные вручную).
			continue
		}

		// Проверяем лимит max_instances_per_game по ВСЕМ инстансам.
		if int32(totalCount) >= policy.MaxInstancesPerGame {
			s.log.Info("policy: max_instances_per_game reached",
				slog.Int64("game_id", policy.GameID),
				slog.Int("total", totalCount),
				slog.Int("target", int(policy.TargetInstances)),
			)
			continue
		}

		needed := int(policy.TargetInstances) - totalCount
		s.log.Info("policy: starting missing instances",
			slog.Int64("game_id", policy.GameID),
			slog.String("mode", policy.Mode.String()),
			slog.Int("needed", needed),
		)
		for i := 0; i < needed; i++ {
			go s.autoStartInstance(context.Background(), policy.GameID, policy)
		}
	}
}

// autoStartInstance запускает инстанс для игры на основе политики.
func (s *HeartbeatService) autoStartInstance(ctx context.Context, gameID int64, policy *domain.GamePolicy) {
	// Проверяем лимит инстансов из политики (все статусы).
	all, _ := s.instanceRepo.ListByGame(ctx, gameID, nil)
	if int32(len(all)) >= policy.MaxInstancesPerGame {
		s.log.Warn("autoStartInstance: max_instances_per_game reached",
			slog.Int64("game_id", gameID),
			slog.Int("current", len(all)),
			slog.Int("max", int(policy.MaxInstancesPerGame)),
		)
		return
	}

	buildVersion := policy.DefaultBuildVersion
	if buildVersion == "latest" || buildVersion == "" {
		// Пытаемся найти последний билд.
		builds, err := s.buildRepo.ListByGame(ctx, gameID, 1)
		if err == nil && len(builds) > 0 {
			buildVersion = builds[0].Version
		} else {
			s.log.Warn("autoStartInstance: no builds found for game",
				slog.Int64("game_id", gameID),
			)
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

	_, err := s.instanceSvc.StartInstance(ctx, params)
	if err != nil {
		s.log.Warn("autoStartInstance failed",
			slog.Int64("game_id", gameID),
			slog.String("error", err.Error()),
		)
	}
}

// enforceScaleToZero останавливает running-инстансы с 0 игроками
// при превышении scale_to_zero_timeout (только для режима scale_to_zero).
func (s *HeartbeatService) enforceScaleToZero(ctx context.Context) {
	policies, err := s.policyService.ListAll(ctx)
	if err != nil {
		s.log.Warn("enforceScaleToZero: failed to list policies", slog.String("error", err.Error()))
		return
	}

	for _, policy := range policies {
		if policy.Mode != domain.OrchestrationModeScaleToZero {
			continue
		}
		if policy.ScaleToZeroTimeout <= 0 {
			continue
		}

		status := domain.InstanceStatusRunning
		instances, err := s.instanceRepo.ListByGame(ctx, policy.GameID, &status)
		if err != nil {
			s.log.Warn("enforceScaleToZero: failed to list instances",
				slog.Int64("game_id", policy.GameID),
				slog.String("error", err.Error()),
			)
			continue
		}

		threshold := time.Duration(policy.ScaleToZeroTimeout) * time.Minute
		for _, inst := range instances {
			pc, _ := s.instanceState.GetPlayerCount(ctx, inst.ID)
			if pc > 0 {
				// Есть игроки — сбрасываем zero-timer если был.
				_ = s.instanceState.DeleteZeroPlayersSince(ctx, inst.ID)
				continue
			}

			zeroSince, err := s.instanceState.GetZeroPlayersSince(ctx, inst.ID)
			if err != nil {
				// Первый раз видим 0 игроков — запоминаем.
				_ = s.instanceState.SetZeroPlayersSince(ctx, inst.ID, time.Now())
				continue
			}

			if time.Since(zeroSince) >= threshold {
				// Таймаут превышен — останавливаем инстанс.
				s.log.Info("scale_to_zero: stopping idle instance",
					slog.Int64("instance_id", inst.ID),
					slog.Int64("game_id", policy.GameID),
					slog.Duration("idle", time.Since(zeroSince)),
				)
				_, _ = s.instanceSvc.StopInstance(ctx, policy.OwnerID, policy.GameID, inst.ID, 0)
				_ = s.instanceState.DeleteZeroPlayersSince(ctx, inst.ID)
			}
		}
	}
}

// enforceScaleUp поднимает дополнительный инстанс, если running-инстанс заполнен
// (player_count >= max_players_per_instance) и scale_behavior = spawn.
func (s *HeartbeatService) enforceScaleUp(ctx context.Context) {
	policies, err := s.policyService.ListAll(ctx)
	if err != nil {
		s.log.Warn("enforceScaleUp: failed to list policies", slog.String("error", err.Error()))
		return
	}

	for _, policy := range policies {
		if !policy.IsAuto() || policy.ScaleBehavior != domain.ScaleBehaviorSpawn {
			continue
		}

		status := domain.InstanceStatusRunning
		instances, err := s.instanceRepo.ListByGame(ctx, policy.GameID, &status)
		if err != nil {
			s.log.Warn("enforceScaleUp: failed to list instances",
				slog.Int64("game_id", policy.GameID),
				slog.String("error", err.Error()),
			)
			continue
		}

		// Считаем текущее общее количество инстансов (все статусы) для лимита.
		all, _ := s.instanceRepo.ListByGame(ctx, policy.GameID, nil)
		if int32(len(all)) >= policy.MaxInstancesPerGame {
			continue
		}

		for _, inst := range instances {
			pc, _ := s.instanceState.GetPlayerCount(ctx, inst.ID)
			if pc >= uint32(policy.MaxPlayersPerInstance) {
				s.log.Info("scale_up: instance full, spawning new",
					slog.Int64("instance_id", inst.ID),
					slog.Int64("game_id", policy.GameID),
					slog.Uint64("players", uint64(pc)),
					slog.Int("max", int(policy.MaxPlayersPerInstance)),
				)
				go s.autoStartInstance(context.Background(), policy.GameID, policy)
				break // Одно масштабирование за цикл.
			}
		}
	}
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
		if err := s.instanceRepo.Update(ctx, inst); err != nil {
			s.log.Warn("failed to update instance status during markNodeOffline",
				slog.Int64("instance_id", inst.ID),
				slog.String("error", err.Error()),
			)
			continue
		}
		_ = s.instanceState.SetStatus(ctx, inst.ID, domain.InstanceStatusCrashed)
		s.log.Info("instance marked as crashed (node offline)",
			slog.Int64("instance_id", inst.ID),
			slog.Int64("node_id", node.ID),
		)

		// Проверяем политику и перезапускаем если нужно.
		go s.maybeAutoRestart(context.Background(), inst)
	}

	return nil
}

// processQueues обрабатывает очереди игроков для всех игр с активной политикой.
// Вызывается после каждого цикла heartbeat.
func (s *HeartbeatService) processQueues(ctx context.Context) {
	if s.queueSvc == nil {
		return
	}

	policies, err := s.policyService.ListAll(ctx)
	if err != nil {
		s.log.Warn("processQueues: failed to list policies", slog.String("error", err.Error()))
		return
	}

	for _, policy := range policies {
		if policy.ScaleBehavior != domain.ScaleBehaviorQueue {
			continue
		}

		if err := s.queueSvc.ProcessQueue(ctx, policy.GameID); err != nil {
			s.log.Debug("processQueues: failed",
				slog.Int64("game_id", policy.GameID),
				slog.String("error", err.Error()),
			)
		}
	}
}