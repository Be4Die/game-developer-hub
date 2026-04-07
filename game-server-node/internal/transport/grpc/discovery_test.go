package grpc

import (
	"context"
	"testing"

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
