//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
)

func TestInstanceStateStore_SetAndGetStatus(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	instanceID := int64(1)

	if err := env.instanceState.SetStatus(ctx, instanceID, domain.InstanceStatusRunning); err != nil {
		t.Fatalf("SetStatus failed: %v", err)
	}

	status, err := env.instanceState.GetStatus(ctx, instanceID)
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}
	if status != domain.InstanceStatusRunning {
		t.Errorf("status = %v, want %v", status, domain.InstanceStatusRunning)
	}
}

func TestInstanceStateStore_GetStatus_NotFound(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	_, err := env.instanceState.GetStatus(ctx, 999)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestInstanceStateStore_SetAndGetPlayerCount(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	instanceID := int64(1)

	if err := env.instanceState.SetPlayerCount(ctx, instanceID, 7); err != nil {
		t.Fatalf("SetPlayerCount failed: %v", err)
	}

	count, err := env.instanceState.GetPlayerCount(ctx, instanceID)
	if err != nil {
		t.Fatalf("GetPlayerCount failed: %v", err)
	}
	if count != 7 {
		t.Errorf("count = %d, want 7", count)
	}
}

func TestInstanceStateStore_GetPlayerCount_NotFound(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	_, err := env.instanceState.GetPlayerCount(ctx, 999)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestInstanceStateStore_SetAndGetUsage(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	instanceID := int64(1)
	usage := &domain.ResourceUsage{
		CPUUsagePercent:    30.0,
		MemoryUsedBytes:    2000000000,
		DiskUsedBytes:      50000000000,
		NetworkBytesPerSec: 250000,
	}

	if err := env.instanceState.SetUsage(ctx, instanceID, usage); err != nil {
		t.Fatalf("SetUsage failed: %v", err)
	}

	got, err := env.instanceState.GetUsage(ctx, instanceID)
	if err != nil {
		t.Fatalf("GetUsage failed: %v", err)
	}
	if got.CPUUsagePercent != 30.0 {
		t.Errorf("cpu = %f, want 30.0", got.CPUUsagePercent)
	}
	if got.MemoryUsedBytes != 2000000000 {
		t.Errorf("memory = %d, want 2000000000", got.MemoryUsedBytes)
	}
}

func TestInstanceStateStore_GetUsage_NotFound(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	_, err := env.instanceState.GetUsage(ctx, 999)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestInstanceStateStore_Delete(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	instanceID := int64(1)

	// Set all types of data.
	if err := env.instanceState.SetStatus(ctx, instanceID, domain.InstanceStatusRunning); err != nil {
		t.Fatalf("SetStatus failed: %v", err)
	}
	if err := env.instanceState.SetPlayerCount(ctx, instanceID, 5); err != nil {
		t.Fatalf("SetPlayerCount failed: %v", err)
	}
	if err := env.instanceState.SetUsage(ctx, instanceID, &domain.ResourceUsage{CPUUsagePercent: 20.0}); err != nil {
		t.Fatalf("SetUsage failed: %v", err)
	}

	// Delete.
	if err := env.instanceState.Delete(ctx, instanceID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// All keys should be gone.
	_, err := env.instanceState.GetStatus(ctx, instanceID)
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound for status after delete, got %v", err)
	}
	_, err = env.instanceState.GetPlayerCount(ctx, instanceID)
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound for player_count after delete, got %v", err)
	}
	_, err = env.instanceState.GetUsage(ctx, instanceID)
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound for usage after delete, got %v", err)
	}
}

func TestInstanceStateStore_StatusRoundTrip_AllValues(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	statuses := []domain.InstanceStatus{
		domain.InstanceStatusStarting,
		domain.InstanceStatusRunning,
		domain.InstanceStatusStopping,
		domain.InstanceStatusStopped,
		domain.InstanceStatusCrashed,
	}

	for i, st := range statuses {
		instanceID := int64(100 + i)
		if err := env.instanceState.SetStatus(ctx, instanceID, st); err != nil {
			t.Fatalf("SetStatus(%v) failed: %v", st, err)
		}
		got, err := env.instanceState.GetStatus(ctx, instanceID)
		if err != nil {
			t.Fatalf("GetStatus(%v) failed: %v", st, err)
		}
		if got != st {
			t.Errorf("status(%d) = %v, want %v", instanceID, got, st)
		}
	}
}
