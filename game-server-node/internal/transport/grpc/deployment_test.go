package grpc

import (
	"bytes"
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
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Заглушка для Docker-а (чтобы не поднимать реальный контейнер)
type fakeRuntime struct {
	hostPort uint32
}

func (f *fakeRuntime) LoadImage(ctx context.Context, tag string, data io.Reader) error { return nil }
func (f *fakeRuntime) BuildImage(ctx context.Context, imageTag string, internalPort uint32, archive io.Reader) error {
	return nil
}
func (f *fakeRuntime) CreateContainer(ctx context.Context, opts domain.ContainerOpts) (string, error) {
	f.hostPort = opts.HostPort
	return "cid", nil
}
func (f *fakeRuntime) StartContainer(ctx context.Context, id string) error { return nil }
func (f *fakeRuntime) GetHostPort(ctx context.Context, containerID string, internalPort uint32) (uint32, error) {
	return f.hostPort, nil
}
func (f *fakeRuntime) StopContainer(ctx context.Context, id string, t time.Duration) error {
	return nil
}
func (f *fakeRuntime) RemoveContainer(ctx context.Context, id string) error { return nil }
func (f *fakeRuntime) ContainerLogs(ctx context.Context, id string, follow bool) (io.ReadCloser, error) {
	// Return empty reader to avoid nil pointer panic
	return io.NopCloser(bytes.NewReader(nil)), nil
}
func (f *fakeRuntime) ContainerStats(ctx context.Context, id string) (domain.ResourcesUsage, error) {
	return domain.ResourcesUsage{}, nil
}

func TestDeploymentHandler_StartInstance(t *testing.T) {
	// Arrange
	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	storage := memory.NewStorage()

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
	svc := service.NewDeploymentService(log, memory.NewStorage(), &fakeRuntime{})
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

func TestDeploymentHandler_StopInstance_Success(t *testing.T) {
	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	storage := memory.NewStorage()

	_ = storage.RecordInstance(ctx, domain.Instance{
		ID:          1,
		ContainerID: "container-abc",
		Status:      domain.InstanceStatusRunning,
	})

	svc := service.NewDeploymentService(log, storage, &fakeRuntime{})
	handler := NewDeploymentHandler(svc)

	req := &pb.StopInstanceRequest{
		InstanceId:     1,
		TimeoutSeconds: 10,
	}

	resp, err := handler.StopInstance(ctx, req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp == nil {
		t.Errorf("expected non-nil response")
	}
}

func TestDeploymentHandler_StartInstance_Handler(t *testing.T) {
	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	storage := memory.NewStorage()

	svc := service.NewDeploymentService(log, storage, &fakeRuntime{})
	handler := NewDeploymentHandler(svc)

	// Pre-load image
	_ = svc.LoadImage(ctx, 1, "test-game:v1", nil)

	req := &pb.StartInstanceRequest{
		GameId:       1,
		Name:         "TestInstance",
		Protocol:     pb.Protocol_PROTOCOL_UDP,
		InternalPort: 7777,
		PortAllocation: &pb.PortAllocation{
			Strategy: &pb.PortAllocation_Exact{Exact: 27015},
		},
		MaxPlayers: 16,
	}

	resp, err := handler.StartInstance(ctx, req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.InstanceId <= 0 {
		t.Errorf("expected positive instance ID, got %d", resp.InstanceId)
	}
	if resp.HostPort != 27015 {
		t.Errorf("expected host port 27015, got %d", resp.HostPort)
	}
}

func TestDeploymentHandler_LoadImage(t *testing.T) {
	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	storage := memory.NewStorage()
	svc := service.NewDeploymentService(log, storage, &fakeRuntime{})
	handler := NewDeploymentHandler(svc)

	mockStream := &mockLoadImageStream{
		ctx: ctx,
		metadata: &pb.ImageMetadata{
			GameId:   42,
			ImageTag: "my-game:latest",
		},
	}

	err := handler.LoadImage(mockStream)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeploymentHandler_StreamLogs(t *testing.T) {
	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	storage := memory.NewStorage()

	// Pre-populate with an instance
	_ = storage.RecordInstance(ctx, domain.Instance{
		ID:          1,
		ContainerID: "container-abc",
	})

	svc := service.NewDeploymentService(log, storage, &fakeRuntime{})
	handler := NewDeploymentHandler(svc)

	req := &pb.StreamLogsRequest{InstanceId: 1}
	stream := &mockLogStream{ctx: ctx}

	err := handler.StreamLogs(req, stream)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeploymentHandler_StreamLogs_NotFound(t *testing.T) {
	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	storage := memory.NewStorage()

	svc := service.NewDeploymentService(log, storage, &fakeRuntime{})
	handler := NewDeploymentHandler(svc)

	req := &pb.StreamLogsRequest{InstanceId: 999}
	stream := &mockLogStream{ctx: ctx}

	err := handler.StreamLogs(req, stream)

	st, _ := status.FromError(err)
	if st.Code() != codes.NotFound {
		t.Errorf("expected NotFound error, got %s", st.Code())
	}
}

// mockLoadImageStream implements pb.DeploymentService_LoadImageServer
type mockLoadImageStream struct {
	grpc.ServerStream
	ctx      context.Context
	metadata *pb.ImageMetadata
	response *pb.LoadImageResponse
}

func (m *mockLoadImageStream) Context() context.Context {
	return m.ctx
}

func (m *mockLoadImageStream) Recv() (*pb.LoadImageRequest, error) {
	if m.metadata != nil {
		// First call returns metadata
		meta := m.metadata
		m.metadata = nil // Clear so next call returns EOF
		return &pb.LoadImageRequest{
			Payload: &pb.LoadImageRequest_Metadata{
				Metadata: meta,
			},
		}, nil
	}
	// Subsequent calls return EOF (no chunks)
	return nil, io.EOF
}

func (m *mockLoadImageStream) SendAndClose(resp *pb.LoadImageResponse) error {
	m.response = resp
	return nil
}

// mockLogStream implements pb.DeploymentService_StreamLogsServer
type mockLogStream struct {
	grpc.ServerStream
	ctx      context.Context
	sentLogs []*pb.StreamLogsResponse
}

func (m *mockLogStream) Context() context.Context {
	return m.ctx
}

func (m *mockLogStream) Send(resp *pb.StreamLogsResponse) error {
	m.sentLogs = append(m.sentLogs, resp)
	return nil
}
