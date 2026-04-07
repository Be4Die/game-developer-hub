package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/game-server-node/internal/domain"
	"github.com/Be4Die/game-developer-hub/game-server-node/internal/storage/memory"
)

// stubRuntime is a fake Docker client for tests.
type stubRuntime struct {
	createdContainerID string
	started            bool

	// Fields to verify StopContainer behavior
	stopContainerID string
	stopTimeout     time.Duration
	stopErr         error // If set, StopContainer will return this error
	removed         bool
}

func (s *stubRuntime) LoadImage(ctx context.Context, imageTag string, data io.Reader) error {
	return nil
}

func (s *stubRuntime) CreateContainer(ctx context.Context, opts domain.ContainerOpts) (string, error) {
	s.createdContainerID = "stub-docker-id-123"
	return s.createdContainerID, nil
}

func (s *stubRuntime) StartContainer(ctx context.Context, containerID string) error {
	s.started = true
	return nil
}

func (s *stubRuntime) StopContainer(ctx context.Context, containerID string, timeout time.Duration) error {
	if s.stopErr != nil {
		return s.stopErr
	}
	s.stopContainerID = containerID
	s.stopTimeout = timeout
	return nil
}

func (s *stubRuntime) RemoveContainer(ctx context.Context, containerID string) error {
	s.removed = true
	return nil
}

func (s *stubRuntime) ContainerLogs(ctx context.Context, containerID string, follow bool) (io.ReadCloser, error) {
	return nil, nil
}

func (s *stubRuntime) ContainerStats(ctx context.Context, containerID string) (domain.ResourcesUsage, error) {
	return domain.ResourcesUsage{}, nil
}

func TestDeploymentService_StartInstance(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	storage := memory.NewMemoryInstanceStorage()
	runtime := &stubRuntime{}

	svc := NewDeploymentService(log, storage, runtime)
	ctx := context.Background()

	err := svc.LoadImage(ctx, 42, "test-game:v1", nil)
	if err != nil {
		t.Fatalf("unexpected error in LoadImage: %v", err)
	}

	opts := StartInstanceOpts{
		GameID:       42,
		Name:         "Test Match",
		Protocol:     domain.ProtocolUDP,
		InternalPort: 7777,
		PortStrategy: domain.PortStrategy{Exact: 27015},
		MaxPlayers:   10,
	}

	instanceID, hostPort, err := svc.StartInstance(ctx, opts)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if instanceID == 0 {
		t.Errorf("expected valid instanceID, got 0")
	}

	if hostPort != 27015 {
		t.Errorf("expected hostPort 27015, got %d", hostPort)
	}

	if !runtime.started {
		t.Errorf("expected container to be started in runtime")
	}

	savedInstance, err := storage.GetInstanceByID(ctx, instanceID)
	if err != nil {
		t.Fatalf("expected instance to be in storage, got error: %v", err)
	}

	if savedInstance.Status != domain.InstanceStatusRunning {
		t.Errorf("expected status Running, got %v", savedInstance.Status)
	}

	if savedInstance.ContainerID != "stub-docker-id-123" {
		t.Errorf("expected container ID from stub, got %s", savedInstance.ContainerID)
	}
}

func TestDeploymentService_StopInstance_Success(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	storage := memory.NewMemoryInstanceStorage()
	runtime := &stubRuntime{}
	svc := NewDeploymentService(log, storage, runtime)
	ctx := context.Background()

	// Pre-populate storage with a running instance
	_ = storage.RecordInstance(ctx, domain.Instance{
		ID:          1,
		ContainerID: "docker-abc",
		Status:      domain.InstanceStatusRunning,
	})

	err := svc.StopInstance(ctx, 1, 5*time.Second)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify runtime was called correctly
	if runtime.stopContainerID != "docker-abc" {
		t.Errorf("expected StopContainer to be called with 'docker-abc', got %s", runtime.stopContainerID)
	}
	if runtime.stopTimeout != 5*time.Second {
		t.Errorf("expected 5s timeout, got %v", runtime.stopTimeout)
	}
	if !runtime.removed {
		t.Errorf("expected RemoveContainer to be called")
	}

	// Verify final status in DB
	saved, _ := storage.GetInstanceByID(ctx, 1)
	if saved.Status != domain.InstanceStatusStopped {
		t.Errorf("expected status Stopped, got %v", saved.Status)
	}
}

func TestDeploymentService_StopInstance_RuntimeError(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	storage := memory.NewMemoryInstanceStorage()

	// Configure runtime to fail on stop
	runtime := &stubRuntime{
		stopErr: errors.New("docker daemon not responding"),
	}

	svc := NewDeploymentService(log, storage, runtime)
	ctx := context.Background()

	_ = storage.RecordInstance(ctx, domain.Instance{
		ID:          1,
		ContainerID: "docker-abc",
		Status:      domain.InstanceStatusRunning,
	})

	err := svc.StopInstance(ctx, 1, 5*time.Second)

	if err == nil {
		t.Fatalf("expected error from runtime, got nil")
	}

	// Verify status changed to Crashed due to Docker error
	saved, _ := storage.GetInstanceByID(ctx, 1)
	if saved.Status != domain.InstanceStatusCrashed {
		t.Errorf("expected status Crashed, got %v", saved.Status)
	}
}

func TestDeploymentService_StopInstance_NotFound(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	storage := memory.NewMemoryInstanceStorage() // Empty storage
	svc := NewDeploymentService(log, storage, &stubRuntime{})

	err := svc.StopInstance(context.Background(), 99, 5*time.Second)

	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
