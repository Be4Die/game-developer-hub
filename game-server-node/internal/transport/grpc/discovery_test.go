package grpc

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/game-server-node/internal/config"
	"github.com/Be4Die/game-developer-hub/game-server-node/internal/domain"
	"github.com/Be4Die/game-developer-hub/game-server-node/internal/service"
	"github.com/Be4Die/game-developer-hub/game-server-node/internal/storage/memory"
	pb "github.com/Be4Die/game-developer-hub/protos/game_server_node/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestDiscoveryHandler_GetInstance(t *testing.T) {
	// 1. Arrange
	ctx := context.Background()
	storage := memory.NewMemoryInstanceStorage()

	// Предзаполняем базу
	_ = storage.RecordInstance(ctx, domain.Instance{
		ID:     1,
		Name:   "Test-Lobby",
		GameID: 100,
	})

	// Собираем реальный сервис с in-memory хранилищем
	cfg := &config.Config{}
	svc := service.NewDiscoveryService(storage, nil, cfg)
	handler := NewDiscoveryHandler(svc)

	// 2. Act - Успешный сценарий
	reqSuccess := &pb.GetInstanceRequest{InstanceId: 1}
	resp, err := handler.GetInstance(ctx, reqSuccess)

	// 3. Assert - Успех
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Instance.Name != "Test-Lobby" {
		t.Errorf("expected name 'Test-Lobby', got '%s'", resp.Instance.Name)
	}

	// 4. Act - Ошибка (Инстанс не найден)
	reqFail := &pb.GetInstanceRequest{InstanceId: 99}
	_, err = handler.GetInstance(ctx, reqFail)

	// 5. Assert - Проверка трансляции ошибки gRPC
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}
	if st.Code() != codes.NotFound {
		t.Errorf("expected code NotFound, got %s", st.Code())
	}
}

func TestDiscoveryHandler_ListInstances(t *testing.T) {
	ctx := context.Background()
	storage := memory.NewMemoryInstanceStorage()

	_ = storage.RecordInstance(ctx, domain.Instance{
		ID:       1,
		Name:     "server-1",
		GameID:   10,
		Port:     8080,
		Protocol: domain.ProtocolTCP,
		Status:   domain.InstanceStatusRunning,
	})
	_ = storage.RecordInstance(ctx, domain.Instance{
		ID:       2,
		Name:     "server-2",
		GameID:   10,
		Port:     8081,
		Protocol: domain.ProtocolUDP,
		Status:   domain.InstanceStatusStopped,
	})

	cfg := &config.Config{}
	svc := service.NewDiscoveryService(storage, nil, cfg)
	handler := NewDiscoveryHandler(svc)

	req := &pb.ListInstancesRequest{}
	resp, err := handler.ListInstances(ctx, req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Instances) != 2 {
		t.Errorf("expected 2 instances, got %d", len(resp.Instances))
	}
}

func TestDiscoveryHandler_ListInstances_Empty(t *testing.T) {
	ctx := context.Background()
	storage := memory.NewMemoryInstanceStorage()

	cfg := &config.Config{}
	svc := service.NewDiscoveryService(storage, nil, cfg)
	handler := NewDiscoveryHandler(svc)

	req := &pb.ListInstancesRequest{}
	resp, err := handler.ListInstances(ctx, req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Instances) != 0 {
		t.Errorf("expected 0 instances, got %d", len(resp.Instances))
	}
}

func TestDiscoveryHandler_ListInstancesByGame(t *testing.T) {
	ctx := context.Background()
	storage := memory.NewMemoryInstanceStorage()

	_ = storage.RecordInstance(ctx, domain.Instance{ID: 1, GameID: 10, Name: "game1-a"})
	_ = storage.RecordInstance(ctx, domain.Instance{ID: 2, GameID: 10, Name: "game1-b"})
	_ = storage.RecordInstance(ctx, domain.Instance{ID: 3, GameID: 20, Name: "game2-a"})

	cfg := &config.Config{}
	svc := service.NewDiscoveryService(storage, nil, cfg)
	handler := NewDiscoveryHandler(svc)

	req := &pb.ListInstancesByGameRequest{GameId: 10}
	resp, err := handler.ListInstancesByGame(ctx, req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Instances) != 2 {
		t.Errorf("expected 2 instances for game 10, got %d", len(resp.Instances))
	}
}

func TestDiscoveryHandler_ListInstancesByGame_NoMatches(t *testing.T) {
	ctx := context.Background()
	storage := memory.NewMemoryInstanceStorage()
	_ = storage.RecordInstance(ctx, domain.Instance{ID: 1, GameID: 10})

	cfg := &config.Config{}
	svc := service.NewDiscoveryService(storage, nil, cfg)
	handler := NewDiscoveryHandler(svc)

	req := &pb.ListInstancesByGameRequest{GameId: 99}
	resp, err := handler.ListInstancesByGame(ctx, req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Instances) != 0 {
		t.Errorf("expected 0 instances for game 99, got %d", len(resp.Instances))
	}
}

// mockRuntime для тестов GetInstanceUsage
type mockRuntime struct {
	usageToReturn domain.ResourcesUsage
	errToReturn   error
}

func (m *mockRuntime) LoadImage(ctx context.Context, imageTag string, data io.Reader) error {
	return nil
}
func (m *mockRuntime) CreateContainer(ctx context.Context, opts domain.ContainerOpts) (string, error) {
	return "", nil
}
func (m *mockRuntime) StartContainer(ctx context.Context, containerID string) error {
	return nil
}
func (m *mockRuntime) StopContainer(ctx context.Context, containerID string, timeout time.Duration) error {
	return nil
}
func (m *mockRuntime) RemoveContainer(ctx context.Context, containerID string) error {
	return nil
}
func (m *mockRuntime) ContainerLogs(ctx context.Context, containerID string, follow bool) (io.ReadCloser, error) {
	return nil, nil
}
func (m *mockRuntime) ContainerStats(ctx context.Context, containerID string) (domain.ResourcesUsage, error) {
	if m.errToReturn != nil {
		return domain.ResourcesUsage{}, m.errToReturn
	}
	return m.usageToReturn, nil
}

func TestDiscoveryHandler_GetInstanceUsage(t *testing.T) {
	ctx := context.Background()
	storage := memory.NewMemoryInstanceStorage()

	_ = storage.RecordInstance(ctx, domain.Instance{
		ID:          1,
		ContainerID: "container-123",
	})

	cfg := &config.Config{}
	runtime := &mockRuntime{
		usageToReturn: domain.ResourcesUsage{
			CPU:    25.5,
			Memory: 512000,
		},
	}
	svc := service.NewDiscoveryService(storage, runtime, cfg)
	handler := NewDiscoveryHandler(svc)

	req := &pb.GetInstanceUsageRequest{InstanceId: 1}
	resp, err := handler.GetInstanceUsage(ctx, req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.InstanceId != 1 {
		t.Errorf("expected instance ID 1, got %d", resp.InstanceId)
	}
	if resp.Usage.CpuUsagePercent != 25.5 {
		t.Errorf("expected CPU 25.5, got %f", resp.Usage.CpuUsagePercent)
	}
	if resp.Usage.MemoryUsedBytes != 512000 {
		t.Errorf("expected memory 512000, got %d", resp.Usage.MemoryUsedBytes)
	}
}

func TestDiscoveryHandler_GetInstanceUsage_NotFound(t *testing.T) {
	ctx := context.Background()
	storage := memory.NewMemoryInstanceStorage()

	cfg := &config.Config{}
	svc := service.NewDiscoveryService(storage, &mockRuntime{}, cfg)
	handler := NewDiscoveryHandler(svc)

	req := &pb.GetInstanceUsageRequest{InstanceId: 999}
	_, err := handler.GetInstanceUsage(ctx, req)

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}
	if st.Code() != codes.NotFound {
		t.Errorf("expected code NotFound, got %s", st.Code())
	}
}

func TestDiscoveryHandler_GetNodeInfo(t *testing.T) {
	ctx := context.Background()
	storage := memory.NewMemoryInstanceStorage()

	cfg := &config.Config{
		Node: config.NodeConfig{
			Region:  "us-east-1",
			Version: "v2.0.0",
		},
	}

	svc := service.NewDiscoveryService(storage, nil, cfg)
	handler := NewDiscoveryHandler(svc)

	req := &pb.GetNodeInfoRequest{}
	resp, err := handler.GetNodeInfo(ctx, req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Region != "us-east-1" {
		t.Errorf("expected region us-east-1, got %s", resp.Region)
	}
	if resp.AgentVersion != "v2.0.0" {
		t.Errorf("expected version v2.0.0, got %s", resp.AgentVersion)
	}
}

func TestDiscoveryHandler_Heartbeat(t *testing.T) {
	ctx := context.Background()
	storage := memory.NewMemoryInstanceStorage()

	_ = storage.RecordInstance(ctx, domain.Instance{ID: 1, Status: domain.InstanceStatusRunning})
	_ = storage.RecordInstance(ctx, domain.Instance{ID: 2, Status: domain.InstanceStatusStarting})
	_ = storage.RecordInstance(ctx, domain.Instance{ID: 3, Status: domain.InstanceStatusStopped})

	cfg := &config.Config{}
	svc := service.NewDiscoveryService(storage, nil, cfg)
	handler := NewDiscoveryHandler(svc)

	req := &pb.HeartbeatRequest{}
	resp, err := handler.Heartbeat(ctx, req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.ActiveInstanceCount != 2 {
		t.Errorf("expected 2 active instances, got %d", resp.ActiveInstanceCount)
	}
}
