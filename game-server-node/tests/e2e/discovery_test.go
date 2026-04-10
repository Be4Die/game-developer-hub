package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	pb "github.com/Be4Die/game-developer-hub/protos/game_server_node/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// E2E_01: GetNodeInfo — проверка характеристик ноды.
func TestE2E_Discovery_GetNodeInfo(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	tc := setupServerContainer(t)

	resp, err := tc.discovery.GetNodeInfo(ctx, &pb.GetNodeInfoRequest{})
	if err != nil {
		t.Fatalf("GetNodeInfo RPC failed: %v", err)
	}

	if resp.Region == "" {
		t.Error("expected non-empty region")
	}
	if resp.AgentVersion == "" {
		t.Error("expected non-empty agent_version")
	}

	// CPU и память зависят от виртуализации Docker Desktop.
	if resp.CpuCores > 0 || resp.TotalMemoryBytes > 0 {
		t.Logf("Resources: cpu=%d cores, mem=%d bytes", resp.CpuCores, resp.TotalMemoryBytes)
	} else {
		t.Log("Resources: stub values returned (Docker Desktop virtualization)")
	}

	t.Logf("Node: region=%s, version=%s, cpu=%d, mem=%d bytes",
		resp.Region, resp.AgentVersion, resp.CpuCores, resp.TotalMemoryBytes)
}

// E2E_02: Heartbeat — проверка загрузки ноды.
func TestE2E_Discovery_Heartbeat(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	tc := setupServerContainer(t)

	resp, err := tc.discovery.Heartbeat(ctx, &pb.HeartbeatRequest{})
	if err != nil {
		t.Fatalf("Heartbeat RPC failed: %v", err)
	}

	t.Logf("Heartbeat: cpu=%.2f%%, mem=%d bytes, active_instances=%d",
		resp.Usage.CpuUsagePercent, resp.Usage.MemoryUsedBytes, resp.ActiveInstanceCount)
}

// E2E_03: ListInstances — пустой список → добавление → список с данными.
func TestE2E_Discovery_ListInstances(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	tc := setupServerContainer(t)

	// 1. Пустой список
	resp, err := tc.discovery.ListInstances(ctx, &pb.ListInstancesRequest{})
	if err != nil {
		t.Fatalf("ListInstances (empty) failed: %v", err)
	}
	initialCount := len(resp.Instances)

	// 2. Запускаем инстанс
	_ = startTestInstance(ctx, t, tc, 1, "list-test-server")

	// 3. Проверяем что список увеличился
	resp, err = tc.discovery.ListInstances(ctx, &pb.ListInstancesRequest{})
	if err != nil {
		t.Fatalf("ListInstances (after start) failed: %v", err)
	}

	if len(resp.Instances) != initialCount+1 {
		t.Errorf("expected %d instances, got %d", initialCount+1, len(resp.Instances))
	}
}

// E2E_04: ListInstancesByGame — фильтрация по game_id.
func TestE2E_Discovery_ListInstancesByGame(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	tc := setupServerContainer(t)

	const gameID1 = 10001
	const gameID2 = 10002

	// Запускаем 2 инстанса для game 1, 1 для game 2
	for i := 0; i < 2; i++ {
		_ = startTestInstanceWithGameID(ctx, t, tc, gameID1, fmt.Sprintf("game1-srv-%d", i))
	}
	_ = startTestInstanceWithGameID(ctx, t, tc, gameID2, "game2-srv")

	// Проверяем фильтрацию
	resp, err := tc.discovery.ListInstancesByGame(ctx, &pb.ListInstancesByGameRequest{GameId: gameID1})
	if err != nil {
		t.Fatalf("ListInstancesByGame(game=%d) failed: %v", gameID1, err)
	}

	if len(resp.Instances) != 2 {
		t.Errorf("expected 2 instances for game %d, got %d", gameID1, len(resp.Instances))
	}

	resp, err = tc.discovery.ListInstancesByGame(ctx, &pb.ListInstancesByGameRequest{GameId: gameID2})
	if err != nil {
		t.Fatalf("ListInstancesByGame(game=%d) failed: %v", gameID2, err)
	}
	if len(resp.Instances) != 1 {
		t.Errorf("expected 1 instance for game %d, got %d", gameID2, len(resp.Instances))
	}

	// Несуществующая игра
	resp, err = tc.discovery.ListInstancesByGame(ctx, &pb.ListInstancesByGameRequest{GameId: 999999})
	if err != nil {
		t.Fatalf("ListInstancesByGame(game=999999) failed: %v", err)
	}
	if len(resp.Instances) != 0 {
		t.Errorf("expected 0 instances for game 999999, got %d", len(resp.Instances))
	}
}

// E2E_05: GetInstance — существующий и несуществующий.
func TestE2E_Discovery_GetInstance(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	tc := setupServerContainer(t)

	startResp := startTestInstance(ctx, t, tc, 1, "get-instance-test")

	// Получение существующего
	resp, err := tc.discovery.GetInstance(ctx, &pb.GetInstanceRequest{
		InstanceId: startResp.InstanceId,
	})
	if err != nil {
		t.Fatalf("GetInstance failed: %v", err)
	}

	if resp.Instance.Name != "get-instance-test" {
		t.Errorf("expected name 'get-instance-test', got '%s'", resp.Instance.Name)
	}
	if resp.Instance.Status != pb.InstanceStatus_INSTANCE_STATUS_RUNNING {
		t.Errorf("expected status RUNNING, got %s", resp.Instance.Status)
	}
	if resp.Instance.GameId != 1 {
		t.Errorf("expected game_id 1, got %d", resp.Instance.GameId)
	}

	// Получение несуществующего
	_, err = tc.discovery.GetInstance(ctx, &pb.GetInstanceRequest{InstanceId: 999999})
	if err == nil {
		t.Fatal("expected NotFound error for non-existent instance")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}
	if st.Code() != codes.NotFound {
		t.Errorf("expected code NotFound, got %s", st.Code())
	}
}

// E2E_06: GetInstanceUsage — метрики запущенного контейнера.
func TestE2E_Discovery_GetInstanceUsage(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	tc := setupServerContainer(t)

	startResp := startTestInstance(ctx, t, tc, 1, "usage-test")

	time.Sleep(containerWait)

	resp, err := tc.discovery.GetInstanceUsage(ctx, &pb.GetInstanceUsageRequest{
		InstanceId: startResp.InstanceId,
	})
	if err != nil {
		t.Fatalf("GetInstanceUsage failed: %v", err)
	}

	if resp.InstanceId != startResp.InstanceId {
		t.Errorf("expected instance_id %d, got %d", startResp.InstanceId, resp.InstanceId)
	}

	t.Logf("Usage: CPU=%.2f%%, Mem=%d bytes, Disk=%d bytes, Net=%d bytes/s",
		resp.Usage.CpuUsagePercent,
		resp.Usage.MemoryUsedBytes,
		resp.Usage.DiskUsedBytes,
		resp.Usage.NetworkBytesPerSec,
	)

	// Метрики несуществующего инстанса
	_, err = tc.discovery.GetInstanceUsage(ctx, &pb.GetInstanceUsageRequest{InstanceId: 999999})
	if err == nil {
		t.Fatal("expected error for non-existent instance usage")
	}
}
