// Package service содержит бизнес-логику управления игровыми серверами.
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Be4Die/game-developer-hub/game-server-node/internal/domain"
)

// DeploymentService управляет развёртыванием игровых инстансов.
// Безопасен для конкурентного использования.
type DeploymentService struct {
	log          *slog.Logger
	storage      domain.InstanceStorage
	runtime      domain.ContainerRuntime
	imageMapPath string
	nodeID       string

	// Simple ID generator. In production — use UUID or database sequence.
	nextID atomic.Int64

	// game_id → image_tag mapping, populated by LoadImage.
	images   map[int64]string
	imagesMu sync.RWMutex
}

// NewDeploymentService создаёт сервис для управления развёртыванием.
func NewDeploymentService(
	log *slog.Logger,
	storage domain.InstanceStorage,
	runtime domain.ContainerRuntime,
	imageMapPath string,
	nodeID string,
) *DeploymentService {
	svc := &DeploymentService{
		log:          log,
		storage:      storage,
		runtime:      runtime,
		imageMapPath: imageMapPath,
		nodeID:       nodeID,
		images:       make(map[int64]string),
	}

	// Load persisted image registry (if any).
	_ = svc.loadImageRegistry()

	return svc
}

// loadImageRegistry загружает маппинг gameID→imageTag из JSON-файла.
// При отсутствии файла или ошибке парсинга логирует предупреждение и продолжает с пустым маппингом.
func (s *DeploymentService) loadImageRegistry() error {
	if s.imageMapPath == "" {
		// Нет пути — пропускаем
		return nil
	}

	data, err := os.ReadFile(s.imageMapPath)
	if err != nil {
		if os.IsNotExist(err) {
			s.log.Info("image registry not found, starting with empty registry",
				slog.String("path", s.imageMapPath),
			)
			return nil
		}
		return fmt.Errorf("read image registry: %w", err)
	}

	registry := make(map[int64]string)
	if err := json.Unmarshal(data, &registry); err != nil {
		s.log.Warn("failed to parse image registry, starting with empty registry",
			slog.String("path", s.imageMapPath),
			slog.String("error", err.Error()),
		)
		return nil
	}

	s.imagesMu.Lock()
	s.images = registry
	s.imagesMu.Unlock()

	s.log.Info("image registry loaded",
		slog.Int("entries", len(registry)),
		slog.String("path", s.imageMapPath),
	)
	return nil
}

// saveImageRegistry сохраняет текущий маппинг gameID→imageTag в JSON-файл.
func (s *DeploymentService) saveImageRegistry() error {
	if s.imageMapPath == "" {
		return nil
	}

	s.imagesMu.RLock()
	data, err := json.MarshalIndent(s.images, "", "  ")
	s.imagesMu.RUnlock()

	if err != nil {
		return fmt.Errorf("marshal image registry: %w", err)
	}

	// Убедимся, что директория существует
	if err := os.MkdirAll(filepath.Dir(s.imageMapPath), 0o750); err != nil {
		return fmt.Errorf("create data directory: %w", err)
	}

	if err := os.WriteFile(s.imageMapPath, data, 0o600); err != nil {
		return fmt.Errorf("write image registry: %w", err)
	}

	return nil
}

// BuildImage собирает Docker-образ из исходного архива на стороне ноды.
func (s *DeploymentService) BuildImage(ctx context.Context, gameID int64, imageTag string, internalPort uint32, archive io.Reader) error {
	const op = "DeploymentService.BuildImage"

	s.log.Info("building image",
		slog.String("op", op),
		slog.Int64("game_id", gameID),
		slog.String("image_tag", imageTag),
		slog.Uint64("internal_port", uint64(internalPort)),
	)

	if err := s.runtime.BuildImage(ctx, imageTag, internalPort, archive); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// Remember which image belongs to which game.
	s.imagesMu.Lock()
	s.images[gameID] = imageTag
	s.imagesMu.Unlock()

	// Persist the registry.
	if err := s.saveImageRegistry(); err != nil {
		s.log.Warn("failed to save image registry",
			slog.String("op", op),
			slog.String("error", err.Error()),
		)
	}

	s.log.Info("image built",
		slog.String("op", op),
		slog.String("image_tag", imageTag),
	)
	return nil
}

// LoadImage загружает контейнер и связывает его с gameID.
func (s *DeploymentService) LoadImage(ctx context.Context, gameID int64, imageTag string, data io.Reader) error {
	const op = "DeploymentService.LoadImage"

	s.log.Info("loading image",
		slog.String("op", op),
		slog.Int64("game_id", gameID),
		slog.String("image_tag", imageTag),
	)

	if err := s.runtime.LoadImage(ctx, imageTag, data); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// Remember which image belongs to which game.
	s.imagesMu.Lock()
	s.images[gameID] = imageTag
	s.imagesMu.Unlock()

	// Persist the registry.
	if err := s.saveImageRegistry(); err != nil {
		s.log.Warn("failed to save image registry",
			slog.String("op", op),
			slog.String("error", err.Error()),
		)
	}

	s.log.Info("image loaded",
		slog.String("op", op),
		slog.String("image_tag", imageTag),
	)
	return nil
}

// StartInstanceOpts задаёт параметры запуска инстанса.
type StartInstanceOpts struct {
	GameID           int64
	InstanceID       int64 // Если != 0 — использовать этот ID вместо автогенерации
	Name             string
	Protocol         domain.Protocol
	InternalPort     uint32
	PortStrategy     domain.PortStrategy // ← было HostPort uint32
	MaxPlayers       uint32
	DeveloperPayload map[string]string
	EnvVars          map[string]string
	Args             []string
	CPUMillis        *uint32
	MemoryBytes      *uint64
}

// StartInstance создаёт и запускает новый игровой инстанс.
// Возвращает ID инстанса и выделенный порт.
func (s *DeploymentService) StartInstance(ctx context.Context, opts StartInstanceOpts) (int64, uint32, error) {
	const op = "DeploymentService.StartInstance"

	// Find image for this game.
	s.imagesMu.RLock()
	imageTag, ok := s.images[opts.GameID]
	s.imagesMu.RUnlock()

	if !ok {
		return 0, 0, fmt.Errorf("%s: no image loaded for game %d", op, opts.GameID)
	}

	hostPort, err := s.resolvePort(ctx, opts.PortStrategy)
	if err != nil {
		return 0, 0, fmt.Errorf("%s: resolve port: %w", op, err)
	}

	// Create container (does not start yet).
	containerID, err := s.runtime.CreateContainer(ctx, domain.ContainerOpts{
		ImageTag:     imageTag,
		InternalPort: opts.InternalPort,
		HostPort:     hostPort,
		EnvVars:      opts.EnvVars,
		Args:         opts.Args,
		CPUMillis:    opts.CPUMillis,
		MemoryBytes:  opts.MemoryBytes,
		Labels: map[string]string{
			"managed_by": "game-server-node",
			"node_id":    s.nodeID,
			"game_id":    fmt.Sprintf("%d", opts.GameID),
		},
	})
	if err != nil {
		return 0, 0, fmt.Errorf("%s: create container: %w", op, err)
	}

	// Start container. If fails — cleanup.
	if err := s.runtime.StartContainer(ctx, containerID); err != nil {
		_ = s.runtime.RemoveContainer(ctx, containerID)
		return 0, 0, fmt.Errorf("%s: start container: %w", op, err)
	}

	// После запуска контейнера получаем реальный хост-порт ( dynamic ports may be assigned).
	actualHostPort, err := s.runtime.GetHostPort(ctx, containerID, opts.InternalPort)
	if err != nil {
		// Останавливаем и удаляем контейнер, т.к. без порта инстанс нерабочий.
		_ = s.runtime.StopContainer(ctx, containerID, 5*time.Second)
		_ = s.runtime.RemoveContainer(ctx, containerID)
		return 0, 0, fmt.Errorf("%s: get host port: %w", op, err)
	}

	// Save to storage. If fails — cleanup steps (stop+remove).
	// Если InstanceID передан — используем его, иначе генерируем свой.
	var id int64
	if opts.InstanceID != 0 {
		id = opts.InstanceID
	} else {
		id = s.nextID.Add(1)
	}

	instance := domain.Instance{
		ID:               id,
		ContainerID:      containerID,
		ImageTag:         imageTag,
		Name:             opts.Name,
		GameID:           opts.GameID,
		Port:             actualHostPort,
		Protocol:         opts.Protocol,
		Status:           domain.InstanceStatusRunning,
		MaxPlayers:       opts.MaxPlayers,
		DeveloperPayload: opts.DeveloperPayload,
		StartedAt:        time.Now(),
	}

	if err := s.storage.RecordInstance(ctx, instance); err != nil {
		_ = s.runtime.StopContainer(ctx, containerID, 5*time.Second)
		_ = s.runtime.RemoveContainer(ctx, containerID)
		return 0, 0, fmt.Errorf("%s: save instance: %w", op, err)
	}

	s.log.Info("instance started",
		slog.String("op", op),
		slog.Int64("instance_id", id),
		slog.String("container_id", containerID),
		slog.Uint64("host_port", uint64(actualHostPort)),
	)

	return id, actualHostPort, nil
}

// StopInstance останавливает инстанс и удаляет его контейнер.
// При успешной остановке запись удаляется из storage; при ошибке - статус меняется на Crashed.
func (s *DeploymentService) StopInstance(ctx context.Context, instanceID int64, timeout time.Duration) error {
	const op = "DeploymentService.StopInstance"

	instance, err := s.storage.GetInstanceByID(ctx, instanceID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// Mark as stopping.
	instance.Status = domain.InstanceStatusStopping
	_ = s.storage.RecordInstance(ctx, *instance)

	// Stop the container.
	if err := s.runtime.StopContainer(ctx, instance.ContainerID, timeout); err != nil {
		instance.Status = domain.InstanceStatusCrashed
		_ = s.storage.RecordInstance(ctx, *instance)
		return fmt.Errorf("%s: %w", op, err)
	}

	// Remove the container (cleanup).
	if err := s.runtime.RemoveContainer(ctx, instance.ContainerID); err != nil {
		s.log.Warn("failed to remove container",
			slog.String("op", op),
			slog.String("error", err.Error()),
		)
	}

	// Delete the instance record from storage - instance is fully stopped.
	if err := s.storage.DeleteInstance(ctx, instanceID); err != nil {
		s.log.Warn("failed to delete instance from storage after stop",
			slog.String("op", op),
			slog.Int64("instance_id", instanceID),
			slog.String("error", err.Error()),
		)
	}

	s.log.Info("instance stopped",
		slog.String("op", op),
		slog.Int64("instance_id", instanceID),
	)

	return nil
}

// DeleteInstance удаляет инстанс и его контейнер без graceful остановки.
func (s *DeploymentService) DeleteInstance(ctx context.Context, instanceID int64) error {
	const op = "DeploymentService.DeleteInstance"

	instance, err := s.storage.GetInstanceByID(ctx, instanceID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// Force stop the container (0 timeout = immediate kill).
	if err := s.runtime.StopContainer(ctx, instance.ContainerID, 0); err != nil {
		s.log.Warn("failed to stop container during delete",
			slog.String("op", op),
			slog.Int64("instance_id", instanceID),
			slog.String("error", err.Error()),
		)
	}

	// Remove the container.
	if err := s.runtime.RemoveContainer(ctx, instance.ContainerID); err != nil {
		s.log.Warn("failed to remove container during delete",
			slog.String("op", op),
			slog.String("error", err.Error()),
		)
	}

	// Delete the instance record from storage.
	if err := s.storage.DeleteInstance(ctx, instanceID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	s.log.Info("instance deleted",
		slog.String("op", op),
		slog.Int64("instance_id", instanceID),
	)

	return nil
}

// StreamLogs возвращает поток логов контейнера инстанса.
func (s *DeploymentService) StreamLogs(ctx context.Context, instanceID int64, follow bool) (io.ReadCloser, error) {
	const op = "DeploymentService.StreamLogs"

	instance, err := s.storage.GetInstanceByID(ctx, instanceID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	logs, err := s.runtime.ContainerLogs(ctx, instance.ContainerID, follow)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return logs, nil
}

// resolvePort подбирает порт на хосте согласно стратегии.
func (s *DeploymentService) resolvePort(ctx context.Context, strategy domain.PortStrategy) (uint32, error) {
	switch {
	case strategy.Exact != 0:
		// Exact port requested — use as-is.
		return strategy.Exact, nil

	case strategy.Any:
		// Let OS pick — port 0 tells Docker to assign random available.
		return 0, nil

	case strategy.Range != nil:
		// Find free port in range not used by active instances.
		instances, err := s.storage.GetAllInstances(ctx)
		if err != nil {
			return 0, err
		}

		usedPorts := make(map[uint32]bool, len(instances))
		for _, inst := range instances {
			if inst.Status == domain.InstanceStatusRunning ||
				inst.Status == domain.InstanceStatusStarting {
				usedPorts[inst.Port] = true
			}
		}

		for port := strategy.Range.Min; port <= strategy.Range.Max; port++ {
			if !usedPorts[port] {
				return port, nil
			}
		}
		return 0, domain.ErrNoAvailablePort

	default:
		// No strategy — let OS decide.
		return 0, nil
	}
}

// StopAllInstances останавливает все инстансы, управляемые этой нодой.
// Вызывается при graceful shutdown.
func (s *DeploymentService) StopAllInstances(ctx context.Context) error {
	const op = "DeploymentService.StopAllInstances"

	instances, err := s.storage.GetAllInstances(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	for _, inst := range instances {
		// Пропускаем уже остановленные или критические.
		if inst.Status == domain.InstanceStatusStopped || inst.Status == domain.InstanceStatusCrashed {
			continue
		}
		if err := s.StopInstance(ctx, inst.ID, 10*time.Second); err != nil {
			s.log.Warn("failed to stop instance during shutdown",
				slog.Int64("instance_id", inst.ID),
				slog.String("error", err.Error()),
			)
		}
	}

	return nil
}

// CleanupOrphans удаляет контейнеры, принадлежащие ноде, но отсутствующие в storage.
// Вызывается при старте (остатки от краша) и при остановке (на всякий случай).
func (s *DeploymentService) CleanupOrphans(ctx context.Context) error {
	const op = "DeploymentService.CleanupOrphans"
	s.log.Info("checking for orphan containers", slog.String("op", op))

	// Получаем все контейнеры хоста.
	containers, err := s.runtime.ListContainers(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// Собираем множество известных containerID из storage.
	knownContainers := make(map[string]struct{})
	instances, err := s.storage.GetAllInstances(ctx)
	if err != nil {
		s.log.Warn("failed to get all instances for orphan check", slog.String("error", err.Error()))
		// Продолжаем без этого списка — будем удалять только по labels, но проверим known later.
	} else {
		for _, inst := range instances {
			knownContainers[inst.ContainerID] = struct{}{}
		}
	}

	toRemove := make([]string, 0)
	for _, c := range containers {
		// Проверяем, что контейнер управляется этой нодой.
		if c.Labels["managed_by"] != "game-server-node" {
			continue
		}
		if c.Labels["node_id"] != s.nodeID {
			continue
		}
		// Если контейнер известен (есть в storage), пропускаем.
		if _, ok := knownContainers[c.ID]; ok {
			continue
		}
		toRemove = append(toRemove, c.ID)
	}

	// Останавливаем и удаляем каждый орфан.
	for _, cid := range toRemove {
		s.log.Info("removing orphan container", slog.String("container_id", cid[:12]))
		if err := s.runtime.StopContainer(ctx, cid, 10*time.Second); err != nil {
			s.log.Warn("failed to stop orphan container",
				slog.String("container_id", cid[:12]),
				slog.String("error", err.Error()),
			)
		}
		if err := s.runtime.RemoveContainer(ctx, cid); err != nil {
			s.log.Warn("failed to remove orphan container",
				slog.String("container_id", cid[:12]),
				slog.String("error", err.Error()),
			)
		} else {
			s.log.Info("orphan container removed", slog.String("container_id", cid[:12]))
		}
	}

	s.log.Info("orphan cleanup finished", slog.Int("removed", len(toRemove)))
	return nil
}

// GetActiveContainerIDs возвращает список ID контейнеров, управляемых этой нодой
// (с labels managed_by=game-server-node и node_id=s.nodeID).
func (s *DeploymentService) GetActiveContainerIDs(ctx context.Context) []string {
	const op = "DeploymentService.GetActiveContainerIDs"
	containers, err := s.runtime.ListContainers(ctx)
	if err != nil {
		s.log.Warn("failed to list containers", slog.String("op", op), slog.String("error", err.Error()))
		return nil
	}
	var ids []string
	for _, c := range containers {
		if c.Labels["managed_by"] == "game-server-node" && c.Labels["node_id"] == s.nodeID {
			ids = append(ids, c.ID)
		}
	}
	return ids
}
