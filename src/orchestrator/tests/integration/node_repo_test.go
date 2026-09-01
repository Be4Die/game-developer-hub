//go:build integration

package integration

import (
	"context"
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
)

func TestNodeRepo_CreateAndGet(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	node := makeTestNode(t, env, 1)
	if err := env.nodeRepo.Create(ctx, node); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	got, err := env.nodeRepo.GetByAddress(ctx, node.Address)
	if err != nil {
		t.Fatalf("GetByAddress failed: %v", err)
	}
	if got.Region != "eu-west" {
		t.Errorf("region = %q, want %q", got.Region, "eu-west")
	}
	if got.CPUCores != 8 {
		t.Errorf("cpu_cores = %d, want 8", got.CPUCores)
	}
}

func TestNodeRepo_Create_Duplicate(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	node := makeTestNode(t, env, 10)
	if err := env.nodeRepo.Create(ctx, node); err != nil {
		t.Fatalf("first Create failed: %v", err)
	}

	node2 := makeTestNode(t, env, 11)
	node2.Address = node.Address // Same address -> duplicate
	err := env.nodeRepo.Create(ctx, node2)
	if err != domain.ErrAlreadyExists {
		t.Errorf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestNodeRepo_Update(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	node := makeTestNode(t, env, 20)
	if err := env.nodeRepo.Create(ctx, node); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	node.Status = domain.NodeStatusOffline
	node.UpdatedAt = time.Now()
	if err := env.nodeRepo.Update(ctx, node); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	got, err := env.nodeRepo.GetByAddress(ctx, node.Address)
	if err != nil {
		t.Fatalf("GetByAddress failed: %v", err)
	}
	if got.Status != domain.NodeStatusOffline {
		t.Errorf("status = %v, want offline", got.Status)
	}
}

func TestNodeRepo_Update_NotFound(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	node := makeTestNode(t, env, 99999)
	err := env.nodeRepo.Update(ctx, node)
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestNodeRepo_GetByAddress(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	node := makeTestNode(t, env, 30)
	if err := env.nodeRepo.Create(ctx, node); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	got, err := env.nodeRepo.GetByAddress(ctx, node.Address)
	if err != nil {
		t.Fatalf("GetByAddress failed: %v", err)
	}
	if got.ID != node.ID {
		t.Errorf("id = %d, want %d", got.ID, node.ID)
	}
}

func TestNodeRepo_GetByAddress_NotFound(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	_, err := env.nodeRepo.GetByAddress(ctx, "nonexistent:44044")
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestNodeRepo_List_All(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	for i := range 3 {
		node := makeTestNode(t, env, int64(40+i))
		if err := env.nodeRepo.Create(ctx, node); err != nil {
			t.Fatalf("Create node %d failed: %v", i, err)
		}
	}

	nodes, err := env.nodeRepo.List(ctx, nil)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(nodes))
	}
}

func TestNodeRepo_List_FilterByStatus(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	online := makeTestNode(t, env, 50)
	online.Address = "online-filter-test:44044"
	if err := env.nodeRepo.Create(ctx, online); err != nil {
		t.Fatalf("Create online failed: %v", err)
	}

	offline := makeTestNode(t, env, 51)
	offline.Address = "offline-filter-test:44044"
	offline.Status = domain.NodeStatusOffline
	if err := env.nodeRepo.Create(ctx, offline); err != nil {
		t.Fatalf("Create offline failed: %v", err)
	}

	status := domain.NodeStatusOnline
	nodes, err := env.nodeRepo.List(ctx, &status)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 online node, got %d", len(nodes))
	}
}

func TestNodeRepo_Delete(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	node := makeTestNode(t, env, 60)
	if err := env.nodeRepo.Create(ctx, node); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := env.nodeRepo.Delete(ctx, node.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err := env.nodeRepo.GetByID(ctx, node.ID)
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestNodeRepo_Delete_NotFound(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	err := env.nodeRepo.Delete(ctx, 99999)
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestNodeRepo_UpdateLastPing(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	node := makeTestNode(t, env, 70)
	if err := env.nodeRepo.Create(ctx, node); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	beforePing := time.Now()
	if err := env.nodeRepo.UpdateLastPing(ctx, node.ID); err != nil {
		t.Fatalf("UpdateLastPing failed: %v", err)
	}

	got, err := env.nodeRepo.GetByID(ctx, node.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if !got.LastPingAt.After(beforePing) {
		t.Errorf("last_ping_at = %v, expected > %v", got.LastPingAt, beforePing)
	}
}

func TestNodeRepo_UpdateLastPing_NotFound(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	err := env.nodeRepo.UpdateLastPing(ctx, 99999)
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// ─── Helper ──────────────────────────────────────────────────────────────────

func makeTestNode(t *testing.T, _ *integrationTestEnv, id int64) *domain.Node {
	t.Helper()
	now := time.Now()
	tokenHash := sha256.Sum256([]byte(fmt.Sprintf("token-%d", id)))

	return &domain.Node{
		ID:           id,
		Address:      fmt.Sprintf("test-node-%d.example.com:44044", id),
		TokenHash:    tokenHash[:],
		Region:       "eu-west",
		Status:       domain.NodeStatusOnline,
		CPUCores:     8,
		TotalMemory:  16000000000,
		TotalDisk:    500000000000,
		AgentVersion: "1.0.0",
		LastPingAt:   now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
