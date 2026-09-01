//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/storage/postgres"
)

func TestGamePolicyRepo_SetAndGet(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	repo := postgres.NewGamePolicyRepo(env.pool)

	policy := &domain.GamePolicy{
		GameID:                1,
		OwnerID:               "user-test",
		Mode:                  domain.OrchestrationModeKeepAlive,
		TargetInstances:       3,
		AutoRestart:           true,
		ScaleToZeroTimeout:    10,
		DefaultBuildVersion:   "v1.0.0",
		MaxPlayersPerInstance: 50,
		MaxInstancesPerGame:   5,
		ScaleBehavior:         domain.ScaleBehaviorSpawn,
		NodePreference:        "auto",
	}

	if err := repo.Set(ctx, policy); err != nil {
		t.Fatalf("set policy: %v", err)
	}

	got, err := repo.Get(ctx, 1)
	if err != nil {
		t.Fatalf("get policy: %v", err)
	}

	if got.Mode != domain.OrchestrationModeKeepAlive {
		t.Errorf("mode: expected keep_alive, got %s", got.Mode)
	}
	if got.TargetInstances != 3 {
		t.Errorf("target_instances: expected 3, got %d", got.TargetInstances)
	}
	if !got.AutoRestart {
		t.Error("auto_restart: expected true")
	}
	if got.DefaultBuildVersion != "v1.0.0" {
		t.Errorf("default_build_version: expected v1.0.0, got %s", got.DefaultBuildVersion)
	}
	if got.OwnerID != "user-test" {
		t.Errorf("owner_id: expected user-test, got %s", got.OwnerID)
	}
}

func TestGamePolicyRepo_ListAll(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	repo := postgres.NewGamePolicyRepo(env.pool)

	for _, p := range []*domain.GamePolicy{
		{GameID: 1, OwnerID: "a", Mode: domain.OrchestrationModeKeepAlive, TargetInstances: 1},
		{GameID: 2, OwnerID: "b", Mode: domain.OrchestrationModeScaleToZero, TargetInstances: 0},
		{GameID: 3, OwnerID: "c", Mode: domain.OrchestrationModeDisabled, TargetInstances: 1},
	} {
		if err := repo.Set(ctx, p); err != nil {
			t.Fatalf("set policy: %v", err)
		}
	}

	policies, err := repo.ListAll(ctx)
	if err != nil {
		t.Fatalf("list all: %v", err)
	}

	if len(policies) != 3 {
		t.Fatalf("expected 3 policies, got %d", len(policies))
	}
}

func TestGamePolicyRepo_Get_NotFound(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	repo := postgres.NewGamePolicyRepo(env.pool)

	_, err := repo.Get(ctx, 999)
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGamePolicyRepo_Delete(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	repo := postgres.NewGamePolicyRepo(env.pool)

	policy := &domain.GamePolicy{GameID: 1, OwnerID: "d", Mode: domain.OrchestrationModeKeepAlive, TargetInstances: 1}
	if err := repo.Set(ctx, policy); err != nil {
		t.Fatalf("set policy: %v", err)
	}

	if err := repo.Delete(ctx, 1); err != nil {
		t.Fatalf("delete policy: %v", err)
	}

	_, err := repo.Get(ctx, 1)
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}
