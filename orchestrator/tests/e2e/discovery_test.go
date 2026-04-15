//go:build e2e

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
)

// E2E_24: Discovery — empty (no running instances).
func TestE2E_Discovery_Empty(t *testing.T) {
	env := setupE2E(t)
	env.cleanupDB(t)

	resp, err := http.Get(env.baseURL + "/games/42/servers")
	if err != nil {
		t.Fatalf("GET /games/42/servers failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	servers := body["servers"].([]any)
	if len(servers) != 0 {
		t.Errorf("expected 0 servers, got %d", len(servers))
	}

	t.Log("Discovery: no servers available for game 42")
}

// E2E_25: Discovery — with running instances.
func TestE2E_Discovery_ServersAvailable(t *testing.T) {
	env := setupE2E(t)
	env.cleanupDB(t)
	ctx := context.Background()

	node := createE2ENode(t, env, 1)
	build := createE2EBuild(t, env, 42, 1)

	now := time.Now()

	// Создаём 2 running инстанса с разным количеством игроков.
	inst1 := &domain.Instance{
		ID:            1,
		NodeID:        node.ID,
		ServerBuildID: build.ID,
		GameID:        42,
		Name:          "server-1",
		BuildVersion:  "v1.0.1",
		Protocol:      domain.ProtocolTCP,
		HostPort:      7001,
		InternalPort:  8080,
		Status:        domain.InstanceStatusRunning,
		MaxPlayers:    10,
		ServerAddress: node.Address,
		StartedAt:     now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := env.instanceRepo.Create(ctx, inst1); err != nil {
		t.Fatalf("Create instance 1 failed: %v", err)
	}
	_ = env.instanceState.SetPlayerCount(ctx, inst1.ID, 8) // почти полный

	inst2 := &domain.Instance{
		ID:            2,
		NodeID:        node.ID,
		ServerBuildID: build.ID,
		GameID:        42,
		Name:          "server-2",
		BuildVersion:  "v1.0.1",
		Protocol:      domain.ProtocolTCP,
		HostPort:      7002,
		InternalPort:  8080,
		Status:        domain.InstanceStatusRunning,
		MaxPlayers:    10,
		ServerAddress: node.Address,
		StartedAt:     now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := env.instanceRepo.Create(ctx, inst2); err != nil {
		t.Fatalf("Create instance 2 failed: %v", err)
	}
	_ = env.instanceState.SetPlayerCount(ctx, inst2.ID, 2) // почти пустой

	resp, err := http.Get(env.baseURL + "/games/42/servers")
	if err != nil {
		t.Fatalf("GET /games/42/servers failed: %v", err)
	}
	defer resp.Body.Close()

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	servers := body["servers"].([]any)
	if len(servers) != 2 {
		t.Fatalf("expected 2 servers, got %d", len(servers))
	}

	// Least-loaded first: server-2 (2 players) before server-1 (8 players).
	first := servers[0].(map[string]any)
	if int64(first["instance_id"].(float64)) != 2 {
		t.Errorf("expected first server to be instance 2 (least loaded), got %v", first["instance_id"])
	}

	t.Logf("Discovery: %d servers returned, first=%v (player_count=%v)",
		len(servers), first["instance_id"], first["player_count"])
}

// E2E_26: Discovery — invalid game ID.
func TestE2E_Discovery_InvalidGameId(t *testing.T) {
	env := setupE2E(t)
	env.cleanupDB(t)

	resp, err := http.Get(env.baseURL + "/games/abc/servers")
	if err != nil {
		t.Fatalf("GET /games/abc/servers failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

// E2E_27: Instance usage endpoint.
func TestE2E_Instances_GetUsage(t *testing.T) {
	env := setupE2E(t)
	env.cleanupDB(t)
	ctx := context.Background()

	node := createE2ENode(t, env, 1)
	build := createE2EBuild(t, env, 42, 1)

	now := time.Now()
	inst := &domain.Instance{
		ID:            300,
		NodeID:        node.ID,
		ServerBuildID: build.ID,
		GameID:        42,
		Name:          "usage-test",
		BuildVersion:  "v1.0.1",
		Protocol:      domain.ProtocolTCP,
		HostPort:      7300,
		InternalPort:  8080,
		Status:        domain.InstanceStatusRunning,
		MaxPlayers:    10,
		ServerAddress: node.Address,
		StartedAt:     now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := env.instanceRepo.Create(ctx, inst); err != nil {
		t.Fatalf("Create instance failed: %v", err)
	}
	_ = env.instanceState.SetUsage(ctx, inst.ID, &domain.ResourceUsage{
		CPUUsagePercent:    25.5,
		MemoryUsedBytes:    512000000,
		DiskUsedBytes:      1000000000,
		NetworkBytesPerSec: 100000,
	})

	resp, err := http.Get(fmt.Sprintf("%s/games/42/instances/%d/usage", env.baseURL, inst.ID))
	if err != nil {
		t.Fatalf("GET /games/42/instances/%d/usage failed: %v", inst.ID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	usage := body["usage"].(map[string]any)
	if usage["cpu_usage_percent"].(float64) != 25.5 {
		t.Errorf("cpu = %v, want 25.5", usage["cpu_usage_percent"])
	}

	t.Logf("Instance usage: CPU=%.1f%%, Mem=%.0f bytes",
		usage["cpu_usage_percent"], usage["memory_used_bytes"])
}
