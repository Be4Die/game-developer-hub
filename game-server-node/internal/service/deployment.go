package service

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Be4Die/game-developer-hub/game-server-node/internal/domain"
)

type DeploymentService struct {
	log     *slog.Logger
	storage domain.InstanceStorage
	runtime domain.ContainerRuntime

	// Simple ID generator. In production — use UUID or database sequence.
	nextID atomic.Int64

	// game_id → image_tag mapping, populated by LoadImage.
	images   map[int64]string
	imagesMu sync.RWMutex
}

func NewDeploymentService(
	log *slog.Logger,
	storage domain.InstanceStorage,
	runtime domain.ContainerRuntime,
) *DeploymentService {
	return &DeploymentService{
		log:     log,
		storage: storage,
		runtime: runtime,
		images:  make(map[int64]string),
	}
}

// LoadImage loads a container image and associates it with a game.
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

	s.log.Info("image loaded",
		slog.String("op", op),
		slog.String("image_tag", imageTag),
	)
	return nil
}

type StartInstanceOpts struct {
	GameID           int64
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

// StartInstance creates and starts a new game server instance.
func (s *DeploymentService) StartInstance(ctx context.Context, opts StartInstanceOpts) (int64, uint32, error) {
	const op = "DeploymentService.StartInstance"

	// 0. Find image for this game.
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

	// 1. Create container.
	containerID, err := s.runtime.CreateContainer(ctx, domain.ContainerOpts{
		ImageTag:     imageTag,
		InternalPort: opts.InternalPort,
		HostPort:     hostPort,
		EnvVars:      opts.EnvVars,
		Args:         opts.Args,
		CPUMillis:    opts.CPUMillis,
		MemoryBytes:  opts.MemoryBytes,
	})
	if err != nil {
		return 0, 0, fmt.Errorf("%s: create container: %w", op, err)
	}

	// 2. Start container. If fails — cleanup step 1.
	if err := s.runtime.StartContainer(ctx, containerID); err != nil {
		_ = s.runtime.RemoveContainer(ctx, containerID)
		return 0, 0, fmt.Errorf("%s: start container: %w", op, err)
	}

	// 3. Save to storage. If fails — cleanup steps 1+2.
	id := s.nextID.Add(1)

	instance := domain.Instance{
		ID:               id,
		ContainerID:      containerID,
		ImageTag:         imageTag,
		Name:             opts.Name,
		GameID:           opts.GameID,
		Port:             hostPort,
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
	)

	return id, hostPort, nil
}

// StopInstance stops a running instance and removes its container.
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

	// Mark as stopped.
	instance.Status = domain.InstanceStatusStopped
	_ = s.storage.RecordInstance(ctx, *instance)

	s.log.Info("instance stopped",
		slog.String("op", op),
		slog.Int64("instance_id", instanceID),
	)

	return nil
}

// StreamLogs returns a stream of container logs.
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

// resolvePort picks a concrete host port based on the strategy.
func (s *DeploymentService) resolvePort(ctx context.Context, strategy domain.PortStrategy) (uint32, error) {
	switch {
	case strategy.Exact != 0:
		// Client wants a specific port — use it as-is.
		return strategy.Exact, nil

	case strategy.Any:
		// Let the OS pick a free port.
		// Port 0 tells Docker to assign a random available port.
		return 0, nil

	case strategy.Range != nil:
		// Find a port in range that's not used by other instances.
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
		// No strategy specified — let OS decide.
		return 0, nil
	}
}
