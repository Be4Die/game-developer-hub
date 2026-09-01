//go:build integration

package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
)

func TestInstanceRepo_CreateAndGet(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	node := makeTestNode(t, env, 1)
	if err := env.nodeRepo.Create(ctx, node); err != nil {
		t.Fatalf("create node failed: %v", err)
	}
	build := makeTestBuild(t, env, 42, 1)

	inst := makeTestInstance(t, 10, node.ID, build.ID, 42)
	if err := env.instanceRepo.Create(ctx, inst); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	got, err := env.instanceRepo.GetByID(ctx, inst.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.Name != "test-instance-10" {
		t.Errorf("name = %q, want %q", got.Name, "test-instance-10")
	}
	if got.Status != domain.InstanceStatusRunning {
		t.Errorf("status = %v, want running", got.Status)
	}
}

func TestInstanceRepo_Create_Duplicate(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	node := makeTestNode(t, env, 1)
	if err := env.nodeRepo.Create(ctx, node); err != nil {
		t.Fatalf("create node failed: %v", err)
	}
	build := makeTestBuild(t, env, 42, 1)

	inst := makeTestInstance(t, 20, node.ID, build.ID, 42)
	if err := env.instanceRepo.Create(ctx, inst); err != nil {
		t.Fatalf("first Create failed: %v", err)
	}

	inst2 := makeTestInstance(t, 20, node.ID, build.ID, 42) // same ID
	inst2.Name = "duplicate"
	err := env.instanceRepo.Create(ctx, inst2)
	if err == nil {
		t.Fatal("expected error on duplicate, got nil")
	}
}

func TestInstanceRepo_Update(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	node := makeTestNode(t, env, 1)
	if err := env.nodeRepo.Create(ctx, node); err != nil {
		t.Fatalf("create node failed: %v", err)
	}
	build := makeTestBuild(t, env, 42, 1)

	inst := makeTestInstance(t, 30, node.ID, build.ID, 42)
	if err := env.instanceRepo.Create(ctx, inst); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	inst.Status = domain.InstanceStatusStopped
	inst.UpdatedAt = time.Now()
	if err := env.instanceRepo.Update(ctx, inst); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	got, err := env.instanceRepo.GetByID(ctx, inst.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.Status != domain.InstanceStatusStopped {
		t.Errorf("status = %v, want stopped", got.Status)
	}
}

func TestInstanceRepo_Update_NotFound(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	inst := makeTestInstance(t, 99999, 1, 1, 42)
	err := env.instanceRepo.Update(ctx, inst)
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestInstanceRepo_ListByGame(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	node := makeTestNode(t, env, 1)
	if err := env.nodeRepo.Create(ctx, node); err != nil {
		t.Fatalf("create node failed: %v", err)
	}
	build42 := makeTestBuild(t, env, 42, 1)
	build99 := makeTestBuild(t, env, 99, 2)

	for i := range 2 {
		inst := makeTestInstance(t, int64(100+i), node.ID, build42.ID, 42)
		if err := env.instanceRepo.Create(ctx, inst); err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	inst99 := makeTestInstance(t, 200, node.ID, build99.ID, 99)
	if err := env.instanceRepo.Create(ctx, inst99); err != nil {
		t.Fatalf("Create for game 99 failed: %v", err)
	}

	instances, err := env.instanceRepo.ListByGame(ctx, 42, nil)
	if err != nil {
		t.Fatalf("ListByGame failed: %v", err)
	}
	if len(instances) != 2 {
		t.Fatalf("expected 2 instances for game 42, got %d", len(instances))
	}

	instances99, err := env.instanceRepo.ListByGame(ctx, 99, nil)
	if err != nil {
		t.Fatalf("ListByGame(99) failed: %v", err)
	}
	if len(instances99) != 1 {
		t.Fatalf("expected 1 instance for game 99, got %d", len(instances99))
	}
}

func TestInstanceRepo_ListByGame_FilterByStatus(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	node := makeTestNode(t, env, 1)
	if err := env.nodeRepo.Create(ctx, node); err != nil {
		t.Fatalf("create node failed: %v", err)
	}
	build := makeTestBuild(t, env, 42, 1)

	running := makeTestInstance(t, 300, node.ID, build.ID, 42)
	if err := env.instanceRepo.Create(ctx, running); err != nil {
		t.Fatalf("Create running failed: %v", err)
	}

	stopped := makeTestInstance(t, 301, node.ID, build.ID, 42)
	stopped.Status = domain.InstanceStatusStopped
	if err := env.instanceRepo.Create(ctx, stopped); err != nil {
		t.Fatalf("Create stopped failed: %v", err)
	}

	status := domain.InstanceStatusRunning
	instances, err := env.instanceRepo.ListByGame(ctx, 42, &status)
	if err != nil {
		t.Fatalf("ListByGame with status failed: %v", err)
	}
	if len(instances) != 1 {
		t.Fatalf("expected 1 running instance, got %d", len(instances))
	}
}

func TestInstanceRepo_ListByNode(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	node := makeTestNode(t, env, 1)
	if err := env.nodeRepo.Create(ctx, node); err != nil {
		t.Fatalf("create node failed: %v", err)
	}
	build := makeTestBuild(t, env, 42, 1)

	for i := range 2 {
		inst := makeTestInstance(t, int64(400+i), node.ID, build.ID, 42)
		if err := env.instanceRepo.Create(ctx, inst); err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	instances, err := env.instanceRepo.ListByNode(ctx, node.ID)
	if err != nil {
		t.Fatalf("ListByNode failed: %v", err)
	}
	if len(instances) != 2 {
		t.Fatalf("expected 2 instances on node, got %d", len(instances))
	}
}

func TestInstanceRepo_Delete(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	node := makeTestNode(t, env, 1)
	if err := env.nodeRepo.Create(ctx, node); err != nil {
		t.Fatalf("create node failed: %v", err)
	}
	build := makeTestBuild(t, env, 42, 1)

	inst := makeTestInstance(t, 500, node.ID, build.ID, 42)
	if err := env.instanceRepo.Create(ctx, inst); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := env.instanceRepo.Delete(ctx, inst.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err := env.instanceRepo.GetByID(ctx, inst.ID)
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestInstanceRepo_CountByGame(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	node := makeTestNode(t, env, 1)
	if err := env.nodeRepo.Create(ctx, node); err != nil {
		t.Fatalf("create node failed: %v", err)
	}
	build := makeTestBuild(t, env, 42, 1)

	for i := range 3 {
		inst := makeTestInstance(t, int64(600+i), node.ID, build.ID, 42)
		if err := env.instanceRepo.Create(ctx, inst); err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	count, err := env.instanceRepo.CountByGame(ctx, 42)
	if err != nil {
		t.Fatalf("CountByGame failed: %v", err)
	}
	if count != 3 {
		t.Errorf("count = %d, want 3", count)
	}
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func makeTestInstance(t *testing.T, id, nodeID, buildID, gameID int64) *domain.Instance {
	t.Helper()
	now := time.Now()
	return &domain.Instance{
		ID:            id,
		NodeID:        nodeID,
		ServerBuildID: buildID,
		GameID:        gameID,
		Name:          fmt.Sprintf("test-instance-%d", id),
		BuildVersion:  "v1.0.0",
		Protocol:      domain.ProtocolTCP,
		HostPort:      uint32(7000 + id),
		InternalPort:  8080,
		Status:        domain.InstanceStatusRunning,
		MaxPlayers:    10,
		ServerAddress: fmt.Sprintf("node-%d.example.com:44044", nodeID),
		StartedAt:     now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func makeTestBuild(t *testing.T, env *integrationTestEnv, gameID, buildID int64) *domain.ServerBuild {
	t.Helper()
	ctx := context.Background()
	build := &domain.ServerBuild{
		ID:           buildID,
		GameID:       gameID,
		Version:      fmt.Sprintf("v1.0.%d", buildID),
		ImageTag:     fmt.Sprintf("welwise/game-%d:v1.0.%d", gameID, buildID),
		Protocol:     domain.ProtocolTCP,
		InternalPort: 8080,
		MaxPlayers:   10,
		FileURL:      fmt.Sprintf("/builds/game-%d-v%d.tar", gameID, buildID),
		FileSize:     1000000,
		CreatedAt:    time.Now(),
	}
	if err := env.buildStorage.Create(ctx, build); err != nil {
		t.Fatalf("makeTestBuild failed: %v", err)
	}
	return build
}
