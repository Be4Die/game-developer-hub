//go:build e2e

package e2e

import (
	"context"
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
	pb "github.com/Be4Die/game-developer-hub/protos/orchestrator/v1"
)

// E2E_02: Register node via manual flow — GetNodeInfo gRPC к реальной ноде.
func TestE2E_Nodes_Register_Manual(t *testing.T) {
	env := setupE2E(t)
	env.cleanupTables(t)
	ctx, cancel := context.WithTimeout(context.Background(), e2eTestTimeout)
	defer cancel()

	nodeAddress := env.getNodeAddress(t)

	resp, err := env.nodeClient.Register(withJWT(ctx, e2eJWTSecret, e2eIssuer), &pb.NodeServiceRegisterRequest{
		Mode: &pb.NodeServiceRegisterRequest_Manual{
			Manual: &pb.RegisterNodeManual{
				Address: nodeAddress,
				Token:   "test-node-token",
				Region:  ptrStr("e2e-test"),
			},
		},
	})
	if err != nil {
		t.Fatalf("RegisterNode failed: %v", err)
	}

	if resp.Node.Status != pb.NodeStatus_NODE_STATUS_ONLINE {
		t.Errorf("status = %q, want %q", resp.Node.Status, pb.NodeStatus_NODE_STATUS_ONLINE)
	}
	if resp.Node.GetRegion() != "e2e-test" {
		t.Errorf("region = %q, want %q", resp.Node.GetRegion(), "e2e-test")
	}
	if resp.Node.GetCpuCores() == 0 {
		t.Error("cpu_cores = 0 — GetNodeInfo gRPC likely failed")
	}

	t.Logf("Node registered: id=%d, address=%s, status=%s, cpu_cores=%d",
		resp.Node.Id, resp.Node.Address, resp.Node.Status, resp.Node.CpuCores)
}

// E2E_03: Register node via authorize flow.
func TestE2E_Nodes_Register_Authorize(t *testing.T) {
	env := setupE2E(t)
	env.cleanupTables(t)
	ctx, cancel := context.WithTimeout(context.Background(), e2eTestTimeout)
	defer cancel()

	// Создаём ноду в unauthorized состоянии.
	now := time.Now()
	token := "test-node-token"
	tokenHash := sha256.Sum256([]byte(token))

	node := &domain.Node{
		ID:         1,
		Address:    "e2e-test-node:44044",
		TokenHash:  tokenHash[:],
		APIToken:   token,
		Status:     domain.NodeStatusUnauthorized,
		LastPingAt: now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := env.nodeRepo.Create(ctx, node); err != nil {
		t.Fatalf("Create node failed: %v", err)
	}

	// Авторизуем через API.
	resp, err := env.nodeClient.Register(withJWT(ctx, e2eJWTSecret, e2eIssuer), &pb.NodeServiceRegisterRequest{
		Mode: &pb.NodeServiceRegisterRequest_Authorize{
			Authorize: &pb.RegisterNodeAuthorize{
				NodeId: node.ID,
				Token:  token,
			},
		},
	})
	if err != nil {
		t.Fatalf("RegisterNode failed: %v", err)
	}

	if resp.Node.Status != pb.NodeStatus_NODE_STATUS_ONLINE {
		t.Errorf("status = %q, want %q", resp.Node.Status, pb.NodeStatus_NODE_STATUS_ONLINE)
	}

	t.Logf("Node authorized: id=%d, status=%s", resp.Node.Id, resp.Node.Status)
}

// E2E_04: List nodes.
func TestE2E_Nodes_List(t *testing.T) {
	env := setupE2E(t)
	env.cleanupTables(t)
	ctx, cancel := context.WithTimeout(context.Background(), e2eTestTimeout)
	defer cancel()

	// Создаём ноды напрямую.
	now := time.Now()
	token := "test-node-token"
	tokenHash := sha256.Sum256([]byte(token))
	for i := range 2 {
		node := &domain.Node{
			ID:         int64(i + 1),
			Address:    fmt.Sprintf("e2e-node-%d:44044", i+1),
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
		_ = env.nodeState.UpdateHeartbeat(ctx, node.ID, &domain.ResourceUsage{})
	}

	resp, err := env.nodeClient.List(withJWT(ctx, e2eJWTSecret, e2eIssuer), &pb.NodeServiceListRequest{})
	if err != nil {
		t.Fatalf("ListNodes failed: %v", err)
	}

	if len(resp.Nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(resp.Nodes))
	}

	t.Logf("Listed %d nodes", len(resp.Nodes))
}

// E2E_05: Delete node.
func TestE2E_Nodes_Delete(t *testing.T) {
	env := setupE2E(t)
	env.cleanupTables(t)
	ctx, cancel := context.WithTimeout(context.Background(), e2eTestTimeout)
	defer cancel()

	now := time.Now()
	token := "test-node-token"
	tokenHash := sha256.Sum256([]byte(token))

	node := &domain.Node{
		ID:         99,
		Address:    "e2e-delete-node:44044",
		TokenHash:  tokenHash[:],
		APIToken:   token,
		Status:     domain.NodeStatusOnline,
		LastPingAt: now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := env.nodeRepo.Create(ctx, node); err != nil {
		t.Fatalf("Create node failed: %v", err)
	}

	_, err := env.nodeClient.Delete(withJWT(ctx, e2eJWTSecret, e2eIssuer), &pb.NodeServiceDeleteRequest{NodeId: node.ID})
	if err != nil {
		t.Fatalf("DeleteNode failed: %v", err)
	}

	_, err = env.nodeRepo.GetByID(ctx, node.ID)
	if err != domain.ErrNotFound {
		t.Errorf("expected node to be deleted, got %v", err)
	}

	t.Logf("Node %d deleted successfully", node.ID)
}

// getNodeAddress возвращает адрес запущенного game-server-node контейнера.
func (env *e2eTestEnv) getNodeAddress(t *testing.T) string {
	t.Helper()
	ctx := context.Background()
	port, err := env.nodeContainer.MappedPort(ctx, "44044")
	if err != nil {
		t.Fatalf("failed to get node port: %v", err)
	}
	host, err := env.nodeContainer.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get node host: %v", err)
	}
	return host + ":" + port.Port()
}

var _ = codes.OK
var _ = status.Error
