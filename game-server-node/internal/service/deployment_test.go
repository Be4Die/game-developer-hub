package service

import (
	"context"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/game-server-node/internal/domain"
	"github.com/Be4Die/game-developer-hub/game-server-node/internal/storage/memory"
)

// stubRuntime — это "фейковый" Docker для тестов.
// Он реализует интерфейс domain.ContainerRuntime.
type stubRuntime struct {
	createdContainerID string
	started            bool
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
	return nil
}

func (s *stubRuntime) RemoveContainer(ctx context.Context, containerID string) error {
	return nil
}

func (s *stubRuntime) ContainerLogs(ctx context.Context, containerID string, follow bool) (io.ReadCloser, error) {
	return nil, nil
}

func (s *stubRuntime) ContainerStats(ctx context.Context, containerID string) (domain.ResourcesUsage, error) {
	return domain.ResourcesUsage{}, nil
}

func TestDeploymentService_StartInstance(t *testing.T) {
	// Arrange
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	storage := memory.NewMemoryInstanceStorage()
	runtime := &stubRuntime{} // Используем нашу заглушку вместо реального Docker!

	svc := NewDeploymentService(log, storage, runtime)
	ctx := context.Background()

	// "Загружаем" образ в память сервиса
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

	// Act
	instanceID, hostPort, err := svc.StartInstance(ctx, opts)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if instanceID == 0 {
		t.Errorf("expected valid instanceID, got 0")
	}

	if hostPort != 27015 {
		t.Errorf("expected hostPort 27015, got %d", hostPort)
	}

	// Проверяем, что заглушка Docker-а была вызвана
	if !runtime.started {
		t.Errorf("expected container to be started in runtime")
	}

	// Проверяем, что инстанс сохранился в базу
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

func TestDeploymentService_resolvePort(t *testing.T) {
	ctx := context.Background()
	storage := memory.NewMemoryInstanceStorage()

	// Занимаем порты 30000 и 30001
	_ = storage.RecordInstance(ctx, domain.Instance{ID: 1, Port: 30000, Status: domain.InstanceStatusRunning})
	_ = storage.RecordInstance(ctx, domain.Instance{ID: 2, Port: 30001, Status: domain.InstanceStatusStarting})
	// Порт 30002 занят, но инстанс остановлен (значит порт можно переиспользовать)
	_ = storage.RecordInstance(ctx, domain.Instance{ID: 3, Port: 30002, Status: domain.InstanceStatusStopped})

	svc := NewDeploymentService(nil, storage, &stubRuntime{})

	tests := []struct {
		name          string
		strategy      domain.PortStrategy
		expectedPort  uint32
		expectedError error
	}{
		{
			name:         "Exact port",
			strategy:     domain.PortStrategy{Exact: 27015},
			expectedPort: 27015,
		},
		{
			name:         "Any port",
			strategy:     domain.PortStrategy{Any: true},
			expectedPort: 0,
		},
		{
			name: "Range - picks first available (skips 30000, 30001, picks 30002)",
			strategy: domain.PortStrategy{
				Range: &domain.PortRange{Min: 30000, Max: 30005},
			},
			expectedPort: 30002,
		},
		{
			name: "Range - no ports available",
			strategy: domain.PortStrategy{
				Range: &domain.PortRange{Min: 30000, Max: 30001},
			},
			expectedError: domain.ErrNoAvailablePort,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port, err := svc.resolvePort(ctx, tt.strategy)
			if err != tt.expectedError {
				t.Errorf("expected error %v, got %v", tt.expectedError, err)
			}
			if port != tt.expectedPort {
				t.Errorf("expected port %d, got %d", tt.expectedPort, port)
			}
		})
	}
}
