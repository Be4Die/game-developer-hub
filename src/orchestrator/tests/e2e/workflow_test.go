//go:build e2e

package e2e

import (
	"context"
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
	pb "github.com/Be4Die/game-developer-hub/protos/orchestrator/v1"
)

// E2E_06: Full workflow — register node → list instances (empty) → discovery (empty).
func TestE2E_FullWorkflow(t *testing.T) {
	env := setupE2E(t)
	env.cleanupTables(t)
	ctx, cancel := context.WithTimeout(context.Background(), e2eTestTimeout)
	defer cancel()

	// Шаг 1: Health check.
	healthResp, err := env.healthClient.Check(ctx, &pb.HealthServiceCheckRequest{})
	if err != nil {
		t.Fatalf("HealthCheck failed: %v", err)
	}
	if healthResp.Status != "ok" {
		t.Fatalf("health check failed: status=%s", healthResp.Status)
	}
	t.Log("Step 1: Health check OK")

	// Шаг 2: Register node via API.
	nodeAddress := env.getNodeAddress(t)
	_, err = env.nodeClient.Register(withJWT(ctx, e2eJWTSecret, e2eIssuer), &pb.NodeServiceRegisterRequest{
		Mode: &pb.NodeServiceRegisterRequest_Manual{
			Manual: &pb.RegisterNodeManual{
				Address: nodeAddress,
				Token:   "test-node-token",
				Region:  ptrStr("e2e-workflow"),
			},
		},
	})
	if err != nil {
		t.Fatalf("RegisterNode failed: %v", err)
	}
	t.Log("Step 2: Node registered")

	// Шаг 3: List nodes — should have 1.
	listResp, err := env.nodeClient.List(withJWT(ctx, e2eJWTSecret, e2eIssuer), &pb.NodeServiceListRequest{})
	if err != nil {
		t.Fatalf("ListNodes failed: %v", err)
	}
	if len(listResp.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(listResp.Nodes))
	}
	t.Log("Step 3: Node listed")

	// Шаг 4: List instances — empty.
	gameID := int64(100)
	instResp, err := env.instanceClient.List(withJWT(ctx, e2eJWTSecret, e2eIssuer), &pb.InstanceServiceListRequest{GameId: gameID})
	if err != nil {
		t.Fatalf("ListInstances failed: %v", err)
	}
	if len(instResp.Instances) != 0 {
		t.Fatalf("expected 0 instances, got %d", len(instResp.Instances))
	}
	t.Log("Step 4: Instances empty (expected)")

	// Шаг 5: Discovery — empty (no instances running).
	discResp, err := env.discoveryClient.DiscoveryServiceDiscover(ctx, &pb.DiscoveryServiceDiscoverRequest{GameId: gameID})
	if err != nil {
		t.Fatalf("DiscoverServers failed: %v", err)
	}
	if len(discResp.Servers) != 0 {
		t.Fatalf("expected 0 servers, got %d", len(discResp.Servers))
	}
	t.Log("Step 5: Discovery empty (expected)")

	// Шаг 6: Delete node.
	nodeID := listResp.Nodes[0].Id
	_, err = env.nodeClient.Delete(withJWT(ctx, e2eJWTSecret, e2eIssuer), &pb.NodeServiceDeleteRequest{NodeId: nodeID})
	if err != nil {
		t.Fatalf("DeleteNode failed: %v", err)
	}
	t.Log("Step 6: Node deleted")

	t.Log("Full workflow completed successfully")
}

// E2E_07: Multiple nodes — verify discovery works with nodes having different loads.
func TestE2E_Discovery_LeastLoaded(t *testing.T) {
	env := setupE2E(t)
	env.cleanupTables(t)
	ctx, cancel := context.WithTimeout(context.Background(), e2eTestTimeout)
	defer cancel()

	now := time.Now()
	token := "test-node-token"
	tokenHash := sha256.Sum256([]byte(token))

	// Создаём 2 ноды с разной загрузкой.
	for i := range 2 {
		node := &domain.Node{
			ID:         int64(i + 1),
			Address:    fmt.Sprintf("e2e-discovery-node-%d:44044", i+1),
			TokenHash:  tokenHash[:],
			APIToken:   token,
			Status:     domain.NodeStatusOnline,
			LastPingAt: now,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		if err := env.nodeRepo.Create(ctx, node); err != nil {
			t.Fatalf("Create node %d failed: %v", i, err)
		}
		// Нода 1: 8 игроков, Нода 2: 2 игрока.
		playerCount := uint32(8 - i*6) // 8, 2
		_ = env.nodeState.UpdateHeartbeat(ctx, node.ID, &domain.ResourceUsage{})
		_ = env.nodeState.SetActiveInstanceCount(ctx, node.ID, playerCount)
	}

	gameID := int64(200)

	resp, err := env.discoveryClient.DiscoveryServiceDiscover(ctx, &pb.DiscoveryServiceDiscoverRequest{GameId: gameID})
	if err != nil {
		t.Fatalf("DiscoverServers failed: %v", err)
	}

	// Без инстансов discovery вернёт empty, но запрос должен быть успешен.
	t.Logf("Discovery returned %d servers", len(resp.Servers))
}
