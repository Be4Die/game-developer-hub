//go:build integration

package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
)

func TestBuildStorage_CreateAndGetByID(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	build := makeTestBuildFull(t, 1, 42, "v1.0.0")
	if err := env.buildStorage.Create(ctx, build); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	got, err := env.buildStorage.GetByID(ctx, build.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.Version != "v1.0.0" {
		t.Errorf("version = %q, want %q", got.Version, "v1.0.0")
	}
	if got.FileSize != 1000000 {
		t.Errorf("file_size = %d, want 1000000", got.FileSize)
	}
}

func TestBuildStorage_GetByVersion(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	build := makeTestBuildFull(t, 10, 42, "v2.0.0")
	if err := env.buildStorage.Create(ctx, build); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	got, err := env.buildStorage.GetByVersion(ctx, 42, "v2.0.0")
	if err != nil {
		t.Fatalf("GetByVersion failed: %v", err)
	}
	if got.ImageTag != "welwise/game-42:v2.0.0" {
		t.Errorf("image_tag = %q, want %q", got.ImageTag, "welwise/game-42:v2.0.0")
	}
}

func TestBuildStorage_GetByVersion_NotFound(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	_, err := env.buildStorage.GetByVersion(ctx, 999, "nonexistent")
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestBuildStorage_Create_Duplicate(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	build := makeTestBuildFull(t, 20, 42, "v1.0.0")
	if err := env.buildStorage.Create(ctx, build); err != nil {
		t.Fatalf("first Create failed: %v", err)
	}

	build2 := makeTestBuildFull(t, 21, 42, "v1.0.0") // same game_id + version
	err := env.buildStorage.Create(ctx, build2)
	if err != domain.ErrAlreadyExists {
		t.Errorf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestBuildStorage_ListByGame(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	for i := range 3 {
		build := makeTestBuildFull(t, int64(100+i), 42, fmt.Sprintf("v1.0.%d", i))
		if err := env.buildStorage.Create(ctx, build); err != nil {
			t.Fatalf("Create build %d failed: %v", i, err)
		}
	}

	build99 := makeTestBuildFull(t, 200, 99, "v1.0.0")
	if err := env.buildStorage.Create(ctx, build99); err != nil {
		t.Fatalf("Create build for game 99 failed: %v", err)
	}

	builds, err := env.buildStorage.ListByGame(ctx, 42, 0)
	if err != nil {
		t.Fatalf("ListByGame failed: %v", err)
	}
	if len(builds) != 3 {
		t.Fatalf("expected 3 builds for game 42, got %d", len(builds))
	}

	buildsLimited, err := env.buildStorage.ListByGame(ctx, 42, 2)
	if err != nil {
		t.Fatalf("ListByGame with limit failed: %v", err)
	}
	if len(buildsLimited) != 2 {
		t.Fatalf("expected 2 builds with limit, got %d", len(buildsLimited))
	}
}

func TestBuildStorage_CountByGame(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	for i := range 5 {
		build := makeTestBuildFull(t, int64(300+i), 42, fmt.Sprintf("v1.0.%d", i))
		if err := env.buildStorage.Create(ctx, build); err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	count, err := env.buildStorage.CountByGame(ctx, 42)
	if err != nil {
		t.Fatalf("CountByGame failed: %v", err)
	}
	if count != 5 {
		t.Errorf("count = %d, want 5", count)
	}
}

func TestBuildStorage_Delete(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	build := makeTestBuildFull(t, 400, 42, "v1.0.0")
	if err := env.buildStorage.Create(ctx, build); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := env.buildStorage.Delete(ctx, build.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err := env.buildStorage.GetByID(ctx, build.ID)
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestBuildStorage_Delete_NotFound(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	err := env.buildStorage.Delete(ctx, 99999)
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestBuildStorage_CountActiveInstancesByBuild(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	node := makeTestNode(t, env, 1)
	if err := env.nodeRepo.Create(ctx, node); err != nil {
		t.Fatalf("create node failed: %v", err)
	}
	build := makeTestBuildFull(t, 500, 42, "v1.0.0")
	if err := env.buildStorage.Create(ctx, build); err != nil {
		t.Fatalf("Create build failed: %v", err)
	}

	// 2 running instances.
	for i := range 2 {
		inst := makeTestInstance(t, int64(1000+i), node.ID, build.ID, 42)
		if err := env.instanceRepo.Create(ctx, inst); err != nil {
			t.Fatalf("Create running failed: %v", err)
		}
	}

	// 1 stopped instance (not counted as active).
	stopped := makeTestInstance(t, 1002, node.ID, build.ID, 42)
	stopped.Status = domain.InstanceStatusStopped
	if err := env.instanceRepo.Create(ctx, stopped); err != nil {
		t.Fatalf("Create stopped failed: %v", err)
	}

	count, err := env.buildStorage.CountActiveInstancesByBuild(ctx, build.ID)
	if err != nil {
		t.Fatalf("CountActiveInstancesByBuild failed: %v", err)
	}
	if count != 2 {
		t.Errorf("active count = %d, want 2", count)
	}
}

// ─── Helper ──────────────────────────────────────────────────────────────────

func makeTestBuildFull(t *testing.T, id, gameID int64, version string) *domain.ServerBuild {
	t.Helper()
	return &domain.ServerBuild{
		ID:           id,
		GameID:       gameID,
		Version:      version,
		ImageTag:     fmt.Sprintf("welwise/game-%d:%s", gameID, version),
		Protocol:     domain.ProtocolTCP,
		InternalPort: 8080,
		MaxPlayers:   10,
		FileURL:      fmt.Sprintf("/builds/game-%d-%s.tar", gameID, version),
		FileSize:     1000000,
		CreatedAt:    time.Now(),
	}
}
