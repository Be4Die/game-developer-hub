//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
)

func TestNodeStateStore_UpdateHeartbeatAndGetUsage(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	nodeID := int64(1)
	usage := &domain.ResourceUsage{
		CPUUsagePercent:    45.5,
		MemoryUsedBytes:    4000000000,
		DiskUsedBytes:      100000000000,
		NetworkBytesPerSec: 500000,
	}

	// UpdateHeartbeat.
	if err := env.nodeState.UpdateHeartbeat(ctx, nodeID, usage); err != nil {
		t.Fatalf("UpdateHeartbeat failed: %v", err)
	}

	// GetUsage.
	got, err := env.nodeState.GetUsage(ctx, nodeID)
	if err != nil {
		t.Fatalf("GetUsage failed: %v", err)
	}
	if got.CPUUsagePercent != 45.5 {
		t.Errorf("cpu = %f, want 45.5", got.CPUUsagePercent)
	}
	if got.MemoryUsedBytes != 4000000000 {
		t.Errorf("memory = %d, want 4000000000", got.MemoryUsedBytes)
	}
}

func TestNodeStateStore_GetUsage_NotFound(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	_, err := env.nodeState.GetUsage(ctx, 999)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestNodeStateStore_SetAndGetActiveInstanceCount(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	nodeID := int64(1)

	// Set.
	if err := env.nodeState.SetActiveInstanceCount(ctx, nodeID, 5); err != nil {
		t.Fatalf("SetActiveInstanceCount failed: %v", err)
	}

	// Get.
	count, err := env.nodeState.GetActiveInstanceCount(ctx, nodeID)
	if err != nil {
		t.Fatalf("GetActiveInstanceCount failed: %v", err)
	}
	if count != 5 {
		t.Errorf("count = %d, want 5", count)
	}
}

func TestNodeStateStore_GetActiveInstanceCount_NotFound(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	_, err := env.nodeState.GetActiveInstanceCount(ctx, 999)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestNodeStateStore_Delete(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	nodeID := int64(1)

	// Set some data.
	usage := &domain.ResourceUsage{CPUUsagePercent: 10.0}
	if err := env.nodeState.UpdateHeartbeat(ctx, nodeID, usage); err != nil {
		t.Fatalf("UpdateHeartbeat failed: %v", err)
	}
	if err := env.nodeState.SetActiveInstanceCount(ctx, nodeID, 3); err != nil {
		t.Fatalf("SetActiveInstanceCount failed: %v", err)
	}

	// Delete.
	if err := env.nodeState.Delete(ctx, nodeID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify keys are gone.
	_, err := env.nodeState.GetUsage(ctx, nodeID)
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}

	_, err = env.nodeState.GetActiveInstanceCount(ctx, nodeID)
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestNodeStateStore_UpdateHeartbeat_TTLExpiry(t *testing.T) {
	env := setupIntegration(t)
	env.Cleanup(t)
	ctx := context.Background()

	nodeID := int64(42)
	usage := &domain.ResourceUsage{CPUUsagePercent: 50.0}

	if err := env.nodeState.UpdateHeartbeat(ctx, nodeID, usage); err != nil {
		t.Fatalf("UpdateHeartbeat failed: %v", err)
	}

	// Should be present immediately.
	got, err := env.nodeState.GetUsage(ctx, nodeID)
	if err != nil {
		t.Fatalf("GetUsage before expiry failed: %v", err)
	}
	if got.CPUUsagePercent != 50.0 {
		t.Errorf("cpu = %f, want 50.0", got.CPUUsagePercent)
	}

	// Note: We can't easily test TTL expiry in a fast test.
	// The key would expire after ttl (45s by default) but we don't want to wait.
	// This test just verifies set/get roundtrip works.
}
