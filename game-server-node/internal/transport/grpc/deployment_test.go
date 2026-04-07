package grpc

import (
	"context"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/game-server-node/internal/domain"
	"github.com/Be4Die/game-developer-hub/game-server-node/internal/service"
	"github.com/Be4Die/game-developer-hub/game-server-node/internal/storage/memory"
	pb "github.com/Be4Die/game-developer-hub/protos/game_server_node/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Заглушка для Docker-а (чтобы не поднимать реальный контейнер)
type fakeRuntime struct{}

func (f *fakeRuntime) LoadImage(ctx context.Context, tag string, data io.Reader) error { return nil }
func (f *fakeRuntime) CreateContainer(ctx context.Context, opts domain.ContainerOpts) (string, error) {
	return "cid", nil
}
func (f *fakeRuntime) StartContainer(ctx context.Context, id string) error { return nil }
func (f *fakeRuntime) StopContainer(ctx context.Context, id string, t time.Duration) error {
	return nil
}
func (f *fakeRuntime) RemoveContainer(ctx context.Context, id string) error { return nil }
func (f *fakeRuntime) ContainerLogs(ctx context.Context, id string, follow bool) (io.ReadCloser, error) {
	return nil, nil
}
func (f *fakeRuntime) ContainerStats(ctx context.Context, id string) (domain.ResourcesUsage, error) {
	return domain.ResourcesUsage{}, nil
}

func TestDeploymentHandler_StartInstance(t *testing.T) {
	// Arrange
	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	storage := memory.NewMemoryInstanceStorage()

	// Собираем DeploymentService с in-memory базой и фейковым докером
	svc := service.NewDeploymentService(log, storage, &fakeRuntime{})
	handler := NewDeploymentHandler(svc)

	// Симулируем предварительную загрузку образа
	_ = svc.LoadImage(ctx, 1, "test-image:v1", nil)

	// Формируем gRPC запрос от клиента
	req := &pb.StartInstanceRequest{
		GameId:       1,
		Name:         "gRPC-Lobby",
		Protocol:     pb.Protocol_PROTOCOL_TCP,
		InternalPort: 8080,
		PortAllocation: &pb.PortAllocation{
			Strategy: &pb.PortAllocation_Exact{Exact: 9090},
		},
		MaxPlayers: 16,
	}

	// Act
	resp, err := handler.StartInstance(ctx, req)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.InstanceId <= 0 {
		t.Errorf("expected positive instance ID, got %d", resp.InstanceId)
	}

	if resp.HostPort != 9090 {
		t.Errorf("expected host port 9090, got %d", resp.HostPort)
	}
}

func TestDeploymentHandler_StopInstance_NotFound(t *testing.T) {
	// Проверяем, что хэндлер правильно возвращает 404, если инстанса нет
	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	svc := service.NewDeploymentService(log, memory.NewMemoryInstanceStorage(), &fakeRuntime{})
	handler := NewDeploymentHandler(svc)

	req := &pb.StopInstanceRequest{
		InstanceId:     999, // Не существует
		TimeoutSeconds: 5,
	}

	_, err := handler.StopInstance(ctx, req)
	st, _ := status.FromError(err)

	if st.Code() != codes.NotFound {
		t.Errorf("expected NotFound error, got %s", st.Code())
	}
}
