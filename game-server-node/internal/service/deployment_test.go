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

	// Error injection
	createErr error
	startErr  error
}

func (s *stubRuntime) LoadImage(ctx context.Context, imageTag string, data io.Reader) error {
	return nil
}

func (s *stubRuntime) CreateContainer(ctx context.Context, opts domain.ContainerOpts) (string, error) {
	if s.createErr != nil {
		return "", s.createErr
	}
	s.createdContainerID = "stub-docker-id-123"
	return s.createdContainerID, nil
}

func (s *stubRuntime) StartContainer(ctx context.Context, containerID string) error {
	if s.startErr != nil {
		return s.startErr
	}
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
	return io.NopCloser(nil), nil
}

func (s *stubRuntime) ContainerStats(ctx context.Context, containerID string) (domain.ResourcesUsage, error) {
	return domain.ResourcesUsage{}, nil
}

func TestDeploymentService_StartInstance(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	storage := memory.NewStorage()
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
	storage := memory.NewStorage()
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
	storage := memory.NewStorage()

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
	storage := memory.NewStorage() // Empty storage
	svc := NewDeploymentService(log, storage, &stubRuntime{})

	err := svc.StopInstance(context.Background(), 99, 5*time.Second)

	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDeploymentService_ResolvePort_Exact(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	storage := memory.NewStorage()
	svc := NewDeploymentService(log, storage, &stubRuntime{})
	ctx := context.Background()

	port, err := svc.resolvePort(ctx, domain.PortStrategy{Exact: 12345})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if port != 12345 {
		t.Errorf("expected port 12345, got %d", port)
	}
}

func TestDeploymentService_ResolvePort_Any(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	storage := memory.NewStorage()
	svc := NewDeploymentService(log, storage, &stubRuntime{})
	ctx := context.Background()

	port, err := svc.resolvePort(ctx, domain.PortStrategy{Any: true})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if port != 0 {
		t.Errorf("expected port 0 (OS-assigned), got %d", port)
	}
}

func TestDeploymentService_ResolvePort_Range_FreePort(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	storage := memory.NewStorage()
	ctx := context.Background()

	// Занимаем порт 30001
	_ = storage.RecordInstance(ctx, domain.Instance{
		ID:     1,
		Port:   30001,
		Status: domain.InstanceStatusRunning,
	})

	svc := NewDeploymentService(log, storage, &stubRuntime{})

	port, err := svc.resolvePort(ctx, domain.PortStrategy{
		Range: &domain.PortRange{Min: 30000, Max: 30005},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Порт 30000 должен быть свободен
	if port != 30000 {
		t.Errorf("expected port 30000, got %d", port)
	}
}

func TestDeploymentService_ResolvePort_Range_AllOccupied(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	storage := memory.NewStorage()
	ctx := context.Background()

	// Занимаем все порты в диапазоне
	_ = storage.RecordInstance(ctx, domain.Instance{ID: 1, Port: 40000, Status: domain.InstanceStatusRunning})
	_ = storage.RecordInstance(ctx, domain.Instance{ID: 2, Port: 40001, Status: domain.InstanceStatusRunning})
	_ = storage.RecordInstance(ctx, domain.Instance{ID: 3, Port: 40002, Status: domain.InstanceStatusRunning})

	svc := NewDeploymentService(log, storage, &stubRuntime{})

	_, err := svc.resolvePort(ctx, domain.PortStrategy{
		Range: &domain.PortRange{Min: 40000, Max: 40002},
	})

	if !errors.Is(err, domain.ErrNoAvailablePort) {
		t.Errorf("expected ErrNoAvailablePort, got %v", err)
	}
}

func TestDeploymentService_ResolvePort_Range_MinGreaterThanMax(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	storage := memory.NewStorage()
	svc := NewDeploymentService(log, storage, &stubRuntime{})
	ctx := context.Background()

	_, err := svc.resolvePort(ctx, domain.PortStrategy{
		Range: &domain.PortRange{Min: 50000, Max: 40000},
	})

	if !errors.Is(err, domain.ErrNoAvailablePort) {
		t.Errorf("expected ErrNoAvailablePort, got %v", err)
	}
}

func TestDeploymentService_ResolvePort_Default(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	storage := memory.NewStorage()
	svc := NewDeploymentService(log, storage, &stubRuntime{})
	ctx := context.Background()

	// Пустая стратегия — должен вернуть 0 (OS-assigned)
	port, err := svc.resolvePort(ctx, domain.PortStrategy{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if port != 0 {
		t.Errorf("expected port 0, got %d", port)
	}
}

func TestDeploymentService_ResolvePort_Range_SkipsStoppedInstances(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	storage := memory.NewStorage()
	ctx := context.Background()

	// Остановленный инстанс не должен занимать порт
	_ = storage.RecordInstance(ctx, domain.Instance{
		ID:     1,
		Port:   60000,
		Status: domain.InstanceStatusStopped,
	})

	svc := NewDeploymentService(log, storage, &stubRuntime{})

	port, err := svc.resolvePort(ctx, domain.PortStrategy{
		Range: &domain.PortRange{Min: 60000, Max: 60005},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if port != 60000 {
		t.Errorf("expected port 60000 (stopped instance should not block), got %d", port)
	}
}

func TestDeploymentService_StreamLogs_Success(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	storage := memory.NewStorage()
	ctx := context.Background()

	_ = storage.RecordInstance(ctx, domain.Instance{
		ID:          1,
		ContainerID: "container-abc",
	})

	runtime := &stubRuntime{}
	svc := NewDeploymentService(log, storage, runtime)

	logs, err := svc.StreamLogs(ctx, 1, false)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if logs == nil {
		t.Errorf("expected non-nil logs reader")
	}
}

func TestDeploymentService_StreamLogs_InstanceNotFound(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	storage := memory.NewStorage()
	svc := NewDeploymentService(log, storage, &stubRuntime{})

	_, err := svc.StreamLogs(context.Background(), 999, false)

	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDeploymentService_StartInstance_NoImage(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	storage := memory.NewStorage()
	svc := NewDeploymentService(log, storage, &stubRuntime{})
	ctx := context.Background()

	// Не загружаем image для gameID
	opts := StartInstanceOpts{
		GameID:       99,
		Name:         "Test",
		Protocol:     domain.ProtocolTCP,
		InternalPort: 8080,
		PortStrategy: domain.PortStrategy{Exact: 9090},
	}

	_, _, err := svc.StartInstance(ctx, opts)

	if err == nil {
		t.Fatalf("expected error when no image loaded, got nil")
	}
}

func TestDeploymentService_StartInstance_CreateContainerError(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	storage := memory.NewStorage()
	runtime := &stubRuntime{
		createErr: errors.New("docker daemon error"),
	}
	svc := NewDeploymentService(log, storage, runtime)
	ctx := context.Background()

	_ = svc.LoadImage(ctx, 1, "test:v1", nil)

	opts := StartInstanceOpts{
		GameID:       1,
		Name:         "Test",
		Protocol:     domain.ProtocolTCP,
		InternalPort: 8080,
		PortStrategy: domain.PortStrategy{Exact: 9090},
	}

	_, _, err := svc.StartInstance(ctx, opts)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestDeploymentService_StartInstance_StartContainerError(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	storage := memory.NewStorage()
	runtime := &stubRuntime{
		startErr: errors.New("container failed to start"),
	}
	svc := NewDeploymentService(log, storage, runtime)
	ctx := context.Background()

	_ = svc.LoadImage(ctx, 1, "test:v1", nil)

	opts := StartInstanceOpts{
		GameID:       1,
		Name:         "Test",
		Protocol:     domain.ProtocolTCP,
		InternalPort: 8080,
		PortStrategy: domain.PortStrategy{Exact: 9090},
	}

	_, _, err := svc.StartInstance(ctx, opts)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	// Контейнер должен быть удалён после ошибки
	if !runtime.removed {
		t.Errorf("expected container to be removed after start failure")
	}
}
