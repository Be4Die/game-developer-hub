package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	pb "github.com/Be4Die/game-developer-hub/protos/game_server_node/v1"
)

// TestGRPC_FullLifecycle_StartAndStop проверяет полный цикл через gRPC клиент-сервер.
func TestGRPC_FullLifecycle_StartAndStop(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	if err := env.deploymentSvc.LoadImage(ctx, 1, imageTag, nil); err != nil {
		t.Fatalf("LoadImage failed: %v", err)
	}

	startReq := &pb.StartInstanceRequest{
		GameId:       1,
		Name:         "grpc-test-server",
		Protocol:     pb.Protocol_PROTOCOL_TCP,
		InternalPort: 8080,
		PortAllocation: &pb.PortAllocation{
			Strategy: &pb.PortAllocation_Any{Any: true},
		},
		MaxPlayers: 10,
		Args:       []string{"sleep", "60"},
	}

	startResp, err := env.deploymentClient.StartInstance(ctx, startReq)
	if err != nil {
		t.Fatalf("StartInstance RPC failed: %v", err)
	}

	if startResp.InstanceId <= 0 {
		t.Errorf("expected positive instance ID, got %d", startResp.InstanceId)
	}
	t.Logf("gRPC StartInstance: instance_id=%d, host_port=%d", startResp.InstanceId, startResp.HostPort)

	getReq := &pb.GetInstanceRequest{InstanceId: startResp.InstanceId}
	getResp, err := env.discoveryClient.GetInstance(ctx, getReq)
	if err != nil {
		t.Fatalf("GetInstance RPC failed: %v", err)
	}

	if getResp.Instance.Name != "grpc-test-server" {
		t.Errorf("expected name 'grpc-test-server', got '%s'", getResp.Instance.Name)
	}
	if getResp.Instance.Status != pb.InstanceStatus_INSTANCE_STATUS_RUNNING {
		t.Errorf("expected status RUNNING, got %s", getResp.Instance.Status)
	}

	listResp, err := env.discoveryClient.ListInstances(ctx, &pb.ListInstancesRequest{})
	if err != nil {
		t.Fatalf("ListInstances RPC failed: %v", err)
	}

	if len(listResp.Instances) < 1 {
		t.Errorf("expected at least 1 instance, got %d", len(listResp.Instances))
	}

	stopResp, err := env.deploymentClient.StopInstance(ctx, &pb.StopInstanceRequest{
		InstanceId:     startResp.InstanceId,
		TimeoutSeconds: 5,
	})
	if err != nil {
		t.Fatalf("StopInstance RPC failed: %v", err)
	}
	_ = stopResp

	getResp2, err := env.discoveryClient.GetInstance(ctx, getReq)
	if err != nil {
		t.Fatalf("GetInstance after stop failed: %v", err)
	}

	if getResp2.Instance.Status != pb.InstanceStatus_INSTANCE_STATUS_STOPPED {
		t.Errorf("expected status STOPPED, got %s", getResp2.Instance.Status)
	}
}

// TestGRPC_Heartbeat проверяет heartbeat через gRPC.
func TestGRPC_Heartbeat(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	resp, err := env.discoveryClient.Heartbeat(ctx, &pb.HeartbeatRequest{})
	if err != nil {
		t.Fatalf("Heartbeat RPC failed: %v", err)
	}

	t.Logf("Heartbeat: active_instances=%d, cpu=%.2f%%, memory=%d bytes",
		resp.ActiveInstanceCount, resp.Usage.CpuUsagePercent, resp.Usage.MemoryUsedBytes)
}

// TestGRPC_GetNodeInfo проверяет GetNodeInfo через gRPC.
func TestGRPC_GetNodeInfo(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	resp, err := env.discoveryClient.GetNodeInfo(ctx, &pb.GetNodeInfoRequest{})
	if err != nil {
		t.Fatalf("GetNodeInfo RPC failed: %v", err)
	}

	if resp.Region != "test-region" {
		t.Errorf("expected region 'test-region', got '%s'", resp.Region)
	}
	if resp.AgentVersion != "test-0.0.1" {
		t.Errorf("expected version 'test-0.0.1', got '%s'", resp.AgentVersion)
	}

	t.Logf("GetNodeInfo: region=%s, version=%s, cpu_cores=%d, total_memory=%d bytes",
		resp.Region, resp.AgentVersion, resp.CpuCores, resp.TotalMemoryBytes)
}

// TestGRPC_ListInstancesByGame проверяет фильтрацию по game_id через gRPC.
func TestGRPC_ListInstancesByGame(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	if err := env.deploymentSvc.LoadImage(ctx, 1, imageTag, nil); err != nil {
		t.Fatalf("LoadImage failed: %v", err)
	}
	if err := env.deploymentSvc.LoadImage(ctx, 2, imageTag, nil); err != nil {
		t.Fatalf("LoadImage failed: %v", err)
	}

	var game1IDs []int64
	for i := 0; i < 2; i++ {
		resp, err := env.deploymentClient.StartInstance(ctx, &pb.StartInstanceRequest{
			GameId:       1,
			Name:         fmt.Sprintf("game1-server-%d", i),
			Protocol:     pb.Protocol_PROTOCOL_TCP,
			InternalPort: 8080,
			PortAllocation: &pb.PortAllocation{
				Strategy: &pb.PortAllocation_Any{Any: true},
			},
			MaxPlayers: 10,
			Args:       []string{"sleep", "60"},
		})
		if err != nil {
			t.Fatalf("StartInstance for game 1 failed: %v", err)
		}
		game1IDs = append(game1IDs, resp.InstanceId)
	}

	resp2, err := env.deploymentClient.StartInstance(ctx, &pb.StartInstanceRequest{
		GameId:       2,
		Name:         "game2-server",
		Protocol:     pb.Protocol_PROTOCOL_TCP,
		InternalPort: 8080,
		PortAllocation: &pb.PortAllocation{
			Strategy: &pb.PortAllocation_Any{Any: true},
		},
		MaxPlayers: 10,
		Args:       []string{"sleep", "60"},
	})
	if err != nil {
		t.Fatalf("StartInstance for game 2 failed: %v", err)
	}
	defer func() {
		_, _ = env.deploymentClient.StopInstance(ctx, &pb.StopInstanceRequest{InstanceId: resp2.InstanceId, TimeoutSeconds: 3})
	}()

	listByGame, err := env.discoveryClient.ListInstancesByGame(ctx, &pb.ListInstancesByGameRequest{GameId: 1})
	if err != nil {
		t.Fatalf("ListInstancesByGame RPC failed: %v", err)
	}

	if len(listByGame.Instances) != 2 {
		t.Errorf("expected 2 instances for game 1, got %d", len(listByGame.Instances))
	}

	for _, id := range game1IDs {
		_, _ = env.deploymentClient.StopInstance(ctx, &pb.StopInstanceRequest{InstanceId: id, TimeoutSeconds: 3})
	}
}

// TestGRPC_StopNonExistentInstance проверяет корректную обработку ошибки через gRPC.
func TestGRPC_StopNonExistentInstance(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	_, err := env.deploymentClient.StopInstance(ctx, &pb.StopInstanceRequest{
		InstanceId:     99999,
		TimeoutSeconds: 5,
	})
	if err == nil {
		t.Fatal("expected error when stopping non-existent instance via gRPC")
	}

	t.Logf("got expected gRPC error: %v", err)
}

// TestGRPC_GetInstanceUsage проверяет получение метрик через gRPC.
func TestGRPC_GetInstanceUsage(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	if err := env.deploymentSvc.LoadImage(ctx, 1, imageTag, nil); err != nil {
		t.Fatalf("LoadImage failed: %v", err)
	}

	startResp, err := env.deploymentClient.StartInstance(ctx, &pb.StartInstanceRequest{
		GameId:       1,
		Name:         "usage-test-server",
		Protocol:     pb.Protocol_PROTOCOL_TCP,
		InternalPort: 8080,
		PortAllocation: &pb.PortAllocation{
			Strategy: &pb.PortAllocation_Any{Any: true},
		},
		MaxPlayers: 10,
		Args:       []string{"sleep", "60"},
	})
	if err != nil {
		t.Fatalf("StartInstance failed: %v", err)
	}
	defer func() {
		_, _ = env.deploymentClient.StopInstance(ctx, &pb.StopInstanceRequest{InstanceId: startResp.InstanceId, TimeoutSeconds: 3})
	}()

	time.Sleep(500 * time.Millisecond)

	usageResp, err := env.discoveryClient.GetInstanceUsage(ctx, &pb.GetInstanceUsageRequest{
		InstanceId: startResp.InstanceId,
	})
	if err != nil {
		t.Fatalf("GetInstanceUsage RPC failed: %v", err)
	}

	if usageResp.InstanceId != startResp.InstanceId {
		t.Errorf("expected instance ID %d, got %d", startResp.InstanceId, usageResp.InstanceId)
	}

	t.Logf("Instance usage: CPU=%.2f%%, Memory=%d bytes, Disk=%d, Network=%d",
		usageResp.Usage.CpuUsagePercent,
		usageResp.Usage.MemoryUsedBytes,
		usageResp.Usage.DiskUsedBytes,
		usageResp.Usage.NetworkBytesPerSec,
	)
}
