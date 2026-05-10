//go:build e2e

package e2e

import (
	"context"
	"testing"
	"time"

	pb "github.com/Be4Die/game-developer-hub/protos/orchestrator/v1"
)

// E2E_Queue_01: Игрок встаёт в очередь, получает position и heartbeat работает.
func TestE2E_Queue_JoinAndHeartbeat(t *testing.T) {
	env := setupE2E(t)
	env.cleanupTables(t)

	ctx := withJWT(context.Background(), e2eJWTSecret, e2eIssuer)
	gameID := int64(42)

	// Устанавливаем политику с queue
	_, err := env.policyClient.Set(ctx, &pb.GamePolicyServiceSetRequest{
		GameId:                gameID,
		Mode:                  pb.OrchestrationMode_ORCHESTRATION_MODE_KEEP_ALIVE,
		TargetInstances:       1,
		MaxInstancesPerGame:   1,
		MaxPlayersPerInstance: 10,
		ScaleBehavior:         pb.ScaleBehavior_SCALE_BEHAVIOR_QUEUE,
		QueueReservationSeconds: 30,
		QueueHeartbeatTimeout:   15,
		QueueMaxWaitSeconds:     300,
	})
	if err != nil {
		t.Fatalf("set policy failed: %v", err)
	}

	// Join queue
	joinResp, err := env.queueClient.QueueServiceJoin(ctx, &pb.QueueServiceJoinRequest{
		GameId:   gameID,
		PlayerId: "player-1",
	})
	if err != nil {
		t.Fatalf("join failed: %v", err)
	}
	if joinResp.Status != pb.QueueStatus_QUEUE_STATUS_WAITING {
		t.Errorf("status = %v, want WAITING", joinResp.Status)
	}
	if joinResp.Position != 1 {
		t.Errorf("position = %d, want 1", joinResp.Position)
	}
	if joinResp.TotalInQueue != 1 {
		t.Errorf("total = %d, want 1", joinResp.TotalInQueue)
	}

	// Heartbeat
	hbResp, err := env.queueClient.QueueServiceHeartbeat(ctx, &pb.QueueServiceHeartbeatRequest{
		GameId:   gameID,
		PlayerId: "player-1",
	})
	if err != nil {
		t.Fatalf("heartbeat failed: %v", err)
	}
	if hbResp.Status != pb.QueueStatus_QUEUE_STATUS_WAITING {
		t.Errorf("heartbeat status = %v, want WAITING", hbResp.Status)
	}
	if hbResp.Position != 1 {
		t.Errorf("heartbeat position = %d, want 1", hbResp.Position)
	}

	// Второй игрок join
	joinResp2, err := env.queueClient.QueueServiceJoin(ctx, &pb.QueueServiceJoinRequest{
		GameId:   gameID,
		PlayerId: "player-2",
	})
	if err != nil {
		t.Fatalf("join2 failed: %v", err)
	}
	if joinResp2.Position != 2 {
		t.Errorf("position2 = %d, want 2", joinResp2.Position)
	}

	// Leave
	_, err = env.queueClient.QueueServiceLeave(ctx, &pb.QueueServiceLeaveRequest{
		GameId:   gameID,
		PlayerId: "player-1",
	})
	if err != nil {
		t.Fatalf("leave failed: %v", err)
	}

	// player-2 теперь первый
	hbResp2, err := env.queueClient.QueueServiceHeartbeat(ctx, &pb.QueueServiceHeartbeatRequest{
		GameId:   gameID,
		PlayerId: "player-2",
	})
	if err != nil {
		t.Fatalf("heartbeat2 failed: %v", err)
	}
	if hbResp2.Position != 1 {
		t.Errorf("position after leave = %d, want 1", hbResp2.Position)
	}
}

// E2E_Queue_02: Discovery возвращает QUEUE статус когда все инстансы заполнены.
func TestE2E_Queue_DiscoveryQueue(t *testing.T) {
	env := setupE2E(t)
	env.cleanupTables(t)

	ctx := withJWT(context.Background(), e2eJWTSecret, e2eIssuer)
	gameID := int64(43)

	// Политика с queue
	_, err := env.policyClient.Set(ctx, &pb.GamePolicyServiceSetRequest{
		GameId:                gameID,
		Mode:                  pb.OrchestrationMode_ORCHESTRATION_MODE_DISABLED,
		MaxInstancesPerGame:   1,
		MaxPlayersPerInstance: 2,
		ScaleBehavior:         pb.ScaleBehavior_SCALE_BEHAVIOR_QUEUE,
	})
	if err != nil {
		t.Fatalf("set policy failed: %v", err)
	}

	// Discovery без инстансов → UNAVAILABLE (disabled mode)
	discResp, err := env.discoveryClient.DiscoveryServiceDiscover(ctx, &pb.DiscoveryServiceDiscoverRequest{
		GameId: gameID,
	})
	if err != nil {
		t.Fatalf("discover failed: %v", err)
	}
	if discResp.Status != pb.DiscoveryStatus_DISCOVERY_STATUS_UNAVAILABLE {
		t.Errorf("status = %v, want UNAVAILABLE", discResp.Status)
	}
}

// E2E_Queue_03: Cleanup expired players.
func TestE2E_Queue_CleanupExpired(t *testing.T) {
	env := setupE2E(t)
	env.cleanupTables(t)

	ctx := withJWT(context.Background(), e2eJWTSecret, e2eIssuer)
	gameID := int64(44)

	// Политика с очень коротим heartbeat timeout
	_, err := env.policyClient.Set(ctx, &pb.GamePolicyServiceSetRequest{
		GameId:                gameID,
		Mode:                  pb.OrchestrationMode_ORCHESTRATION_MODE_KEEP_ALIVE,
		MaxInstancesPerGame:   1,
		MaxPlayersPerInstance: 10,
		ScaleBehavior:         pb.ScaleBehavior_SCALE_BEHAVIOR_QUEUE,
		QueueHeartbeatTimeout: 1,
		QueueMaxWaitSeconds:   300,
	})
	if err != nil {
		t.Fatalf("set policy failed: %v", err)
	}

	// Join
	_, err = env.queueClient.QueueServiceJoin(ctx, &pb.QueueServiceJoinRequest{
		GameId:   gameID,
		PlayerId: "expired-player",
	})
	if err != nil {
		t.Fatalf("join failed: %v", err)
	}

	// Ждём истечения heartbeat timeout
	time.Sleep(2 * time.Second)

	// Heartbeat должен вернуть EXPIRED
	hbResp, err := env.queueClient.QueueServiceHeartbeat(ctx, &pb.QueueServiceHeartbeatRequest{
		GameId:   gameID,
		PlayerId: "expired-player",
	})
	if err != nil {
		t.Fatalf("heartbeat failed: %v", err)
	}
	if hbResp.Status != pb.QueueStatus_QUEUE_STATUS_EXPIRED {
		t.Errorf("status = %v, want EXPIRED", hbResp.Status)
	}
}
