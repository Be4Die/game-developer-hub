package service

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/game-server-node/internal/config"
	"github.com/Be4Die/game-developer-hub/game-server-node/internal/domain"
	"github.com/Be4Die/game-developer-hub/game-server-node/internal/storage/memory"
)

// stubSysProvider is a mock for system metrics.
type stubSysProvider struct {
	maxToReturn   domain.ResourcesMax
	usageToReturn domain.ResourcesUsage
}

func (s *stubSysProvider) GetMax() (domain.ResourcesMax, error) {
	return s.maxToReturn, nil
}

func (s *stubSysProvider) GetUsage() (domain.ResourcesUsage, error) {
	return s.usageToReturn, nil
}

// stubDiscoveryRuntime is a minimal Docker mock to test GetInstanceUsage.
type stubDiscoveryRuntime struct {
	expectedContainerID string
	usageToReturn       domain.ResourcesUsage
	errToReturn         error
}

func (s *stubDiscoveryRuntime) LoadImage(ctx context.Context, imageTag string, data io.Reader) error {
	return nil
}
func (s *stubDiscoveryRuntime) CreateContainer(ctx context.Context, opts domain.ContainerOpts) (string, error) {
	return "", nil
}
func (s *stubDiscoveryRuntime) StartContainer(ctx context.Context, containerID string) error {
	return nil
}
func (s *stubDiscoveryRuntime) StopContainer(ctx context.Context, containerID string, timeout time.Duration) error {
	return nil
}
func (s *stubDiscoveryRuntime) RemoveContainer(ctx context.Context, containerID string) error {
	return nil
}
func (s *stubDiscoveryRuntime) ContainerLogs(ctx context.Context, containerID string, follow bool) (io.ReadCloser, error) {
	return nil, nil
}

func (s *stubDiscoveryRuntime) ContainerStats(ctx context.Context, containerID string) (domain.ResourcesUsage, error) {
	if s.errToReturn != nil {
		return domain.ResourcesUsage{}, s.errToReturn
	}
	if containerID != s.expectedContainerID {
		return domain.ResourcesUsage{}, errors.New("wrong container ID")
	}
	return s.usageToReturn, nil
}

func TestDiscoveryService_Heartbeat(t *testing.T) {
	ctx := context.Background()

	storage := memory.NewMemoryInstanceStorage()
	_ = storage.RecordInstance(ctx, domain.Instance{ID: 1, Status: domain.InstanceStatusRunning})
	_ = storage.RecordInstance(ctx, domain.Instance{ID: 2, Status: domain.InstanceStatusStarting})
	_ = storage.RecordInstance(ctx, domain.Instance{ID: 3, Status: domain.InstanceStatusStopped})
	_ = storage.RecordInstance(ctx, domain.Instance{ID: 4, Status: domain.InstanceStatusCrashed})

	mockSys := &stubSysProvider{
		usageToReturn: domain.ResourcesUsage{
			CPU:     42.5,
			Memory:  1024 * 1024 * 500,
			Network: 1000,
		},
	}

	cfg := &config.Config{
		Node: config.NodeConfig{
			Region:  "test-region",
			Version: "1.0.0",
		},
	}

	svc := NewDiscoveryService(storage, nil, cfg)
	svc.sysProvider = mockSys

	result, err := svc.Heartbeat(ctx)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.ActiveInstanceCount != 2 {
		t.Errorf("expected 2 active instances, got %d", result.ActiveInstanceCount)
	}

	if result.Usage.CPU != 42.5 {
		t.Errorf("expected CPU 42.5, got %f", result.Usage.CPU)
	}
}

func TestDiscoveryService_GetNode(t *testing.T) {
	mockSys := &stubSysProvider{
		maxToReturn: domain.ResourcesMax{
			CPUCores: 8,
		},
	}

	cfg := &config.Config{
		Node: config.NodeConfig{
			Region:  "eu-central",
			Version: "v1.2.3",
		},
	}

	svc := NewDiscoveryService(nil, nil, cfg)
	svc.sysProvider = mockSys

	testTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	svc.startedAt = testTime

	node, err := svc.GetNode()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Region != "eu-central" {
		t.Errorf("expected region eu-central, got %s", node.Region)
	}

	if node.Version != "v1.2.3" {
		t.Errorf("expected version v1.2.3, got %s", node.Version)
	}

	if node.Resources.CPUCores != 8 {
		t.Errorf("expected 8 CPU cores, got %d", node.Resources.CPUCores)
	}

	if !node.StartedAt.Equal(testTime) {
		t.Errorf("expected time %v, got %v", testTime, node.StartedAt)
	}
}

func TestDiscoveryService_GetInstanceUsage_Success(t *testing.T) {
	ctx := context.Background()
	storage := memory.NewMemoryInstanceStorage()

	_ = storage.RecordInstance(ctx, domain.Instance{
		ID:          1,
		ContainerID: "docker-123",
	})

	expectedUsage := domain.ResourcesUsage{
		CPU:    15.5,
		Memory: 256000,
	}

	mockRuntime := &stubDiscoveryRuntime{
		expectedContainerID: "docker-123",
		usageToReturn:       expectedUsage,
	}

	cfg := &config.Config{}
	svc := NewDiscoveryService(storage, mockRuntime, cfg)

	usage, err := svc.GetInstanceUsage(ctx, 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if usage.CPU != expectedUsage.CPU {
		t.Errorf("expected CPU %f, got %f", expectedUsage.CPU, usage.CPU)
	}
	if usage.Memory != expectedUsage.Memory {
		t.Errorf("expected Memory %d, got %d", expectedUsage.Memory, usage.Memory)
	}
}

func TestDiscoveryService_GetInstanceUsage_NotFound(t *testing.T) {
	ctx := context.Background()
	storage := memory.NewMemoryInstanceStorage() // Empty DB

	cfg := &config.Config{}
	svc := NewDiscoveryService(storage, &stubDiscoveryRuntime{}, cfg)

	_, err := svc.GetInstanceUsage(ctx, 99)

	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
