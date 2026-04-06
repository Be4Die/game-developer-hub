package service

import (
	"context"
	"time"

	"github.com/Be4Die/game-developer-hub/game-server-node/internal/config"
	"github.com/Be4Die/game-developer-hub/game-server-node/internal/domain"
	"github.com/Be4Die/game-developer-hub/game-server-node/internal/sysinfo"
)

type HeartbeatResult struct {
	Usage               domain.ResourcesUsage
	ActiveInstanceCount uint32
}

type DiscoveryService struct {
	storage     domain.InstanceStorage
	runtime     domain.ContainerRuntime
	sysProvider sysinfo.Provider
	config      *config.NodeConfig
	startedAt   time.Time
}

func NewDiscoveryService(
	storage domain.InstanceStorage,
	runtime domain.ContainerRuntime,
	config *config.Config,
) *DiscoveryService {
	return &DiscoveryService{
		storage:     storage,
		runtime:     runtime,
		config:      &config.Node,
		sysProvider: sysinfo.NewProvider(config.Node.EthName),
		startedAt:   time.Now(),
	}
}

func (d *DiscoveryService) GetNode() (*domain.Node, error) {
	res, err := d.sysProvider.GetMax()
	if err != nil {
		return nil, err
	}

	return &domain.Node{
		Version:   d.config.Version,
		Region:    d.config.Region,
		Resources: res,
		StartedAt: d.startedAt,
	}, nil
}

func (d *DiscoveryService) Heartbeat(ctx context.Context) (*HeartbeatResult, error) {
	usage, err := d.sysProvider.GetUsage()
	if err != nil {
		return nil, err
	}

	instances, err := d.storage.GetAllInstances(ctx)
	if err != nil {
		return nil, err
	}

	var active uint32
	for _, inst := range instances {
		if inst.Status == domain.InstanceStatusRunning ||
			inst.Status == domain.InstanceStatusStarting {
			active++
		}
	}

	return &HeartbeatResult{
		Usage:               usage,
		ActiveInstanceCount: active,
	}, nil
}

func (d *DiscoveryService) GetAllInstances(ctx context.Context) ([]domain.Instance, error) {
	return d.storage.GetAllInstances(ctx)
}

func (d *DiscoveryService) GetInstance(ctx context.Context, id int64) (*domain.Instance, error) {
	return d.storage.GetInstanceByID(ctx, id)
}

func (d *DiscoveryService) GetInstancesByGameID(ctx context.Context, gameID int64) ([]domain.Instance, error) {
	return d.storage.GetInstancesByGameID(ctx, gameID)
}

func (d *DiscoveryService) GetInstanceUsage(ctx context.Context, instanceID int64) (*domain.ResourcesUsage, error) {
	instance, err := d.storage.GetInstanceByID(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	usage, err := d.runtime.ContainerStats(ctx, instance.ContainerID)
	if err != nil {
		return nil, err
	}

	return &usage, nil
}
