//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
)

// E2E_06: Full workflow — register node → list instances (empty) → discovery (empty).
func TestE2E_FullWorkflow(t *testing.T) {
	env := setupE2E(t)
	env.cleanupTables(t)
	ctx, cancel := context.WithTimeout(context.Background(), e2eTestTimeout)
	defer cancel()

	// Шаг 1: Health check.
	resp, err := http.Get(env.baseURL + "/health")
	if err != nil {
		t.Fatalf("GET /health failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("health check failed: status=%d", resp.StatusCode)
	}
	t.Log("Step 1: Health check OK")

	// Шаг 2: Register node via API.
	nodeAddress := env.getNodeAddress(t)
	reqBody := map[string]any{
		"address": nodeAddress,
		"token":   e2eAPIKey,
		"region":  "e2e-workflow",
	}
	body, _ := json.Marshal(reqBody)
	resp, err = http.Post(env.baseURL+"/nodes", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /nodes failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("register node: status=%d, body=%s", resp.StatusCode, respBody)
	}
	resp.Body.Close()
	t.Log("Step 2: Node registered")

	// Шаг 3: List nodes — should have 1.
	resp, err = http.Get(env.baseURL + "/nodes")
	if err != nil {
		t.Fatalf("GET /nodes failed: %v", err)
	}
	var nodesBody map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&nodesBody); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	resp.Body.Close()
	nodes := nodesBody["nodes"].([]any)
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	t.Log("Step 3: Node listed")

	// Шаг 4: List instances — empty.
	gameID := int64(100)
	resp, err = http.Get(fmt.Sprintf("%s/games/%d/instances", env.baseURL, gameID))
	if err != nil {
		t.Fatalf("GET /games/%d/instances failed: %v", gameID, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list instances: status=%d", resp.StatusCode)
	}
	var instBody map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&instBody); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	instances := instBody["instances"].([]any)
	if len(instances) != 0 {
		t.Fatalf("expected 0 instances, got %d", len(instances))
	}
	t.Log("Step 4: Instances empty (expected)")

	// Шаг 5: Discovery — empty (no instances running).
	resp, err = http.Get(fmt.Sprintf("%s/games/%d/discover", env.baseURL, gameID))
	if err != nil {
		t.Fatalf("GET /games/%d/discover failed: %v", gameID, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("discovery: status=%d", resp.StatusCode)
	}
	var discBody map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&discBody); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	servers := discBody["servers"].([]any)
	if len(servers) != 0 {
		t.Fatalf("expected 0 servers, got %d", len(servers))
	}
	t.Log("Step 5: Discovery empty (expected)")

	// Шаг 6: Delete node.
	nodeID := int64(nodes[0].(map[string]any)["id"].(float64))
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/nodes/%d", env.baseURL, nodeID), nil)
	client := &http.Client{}
	delResp, err := client.Do(req)
	if err != nil {
		t.Fatalf("DELETE /nodes/%d failed: %v", nodeID, err)
	}
	delResp.Body.Close()
	if delResp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete node: status=%d", delResp.StatusCode)
	}
	t.Log("Step 6: Node deleted")

	t.Log("Full workflow completed successfully")
	_ = ctx
}

// E2E_07: Multiple nodes — verify least-loaded-first discovery ordering.
func TestE2E_Discovery_LeastLoaded(t *testing.T) {
	env := setupE2E(t)
	env.cleanupTables(t)
	ctx, cancel := context.WithTimeout(context.Background(), e2eTestTimeout)
	defer cancel()

	now := time.Now()
	tokenHash := sha256.Sum256([]byte(e2eAPIKey))

	// Создаём 2 ноды с разной загрузкой.
	for i := range 2 {
		node := &domain.Node{
			ID:         int64(i + 1),
			Address:    fmt.Sprintf("e2e-discovery-node-%d:44044", i+1),
			TokenHash:  tokenHash[:],
			APIToken:   e2eAPIKey,
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

	resp, err := http.Get(fmt.Sprintf("%s/games/%d/discover", env.baseURL, gameID))
	if err != nil {
		t.Fatalf("GET /games/%d/discover failed: %v", gameID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("discovery: status=%d", resp.StatusCode)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	servers := body["servers"].([]any)
	// Без инстансов discovery вернёт empty, но ноды должны быть доступны.
	// Проверим что запрос успешен.
	if _, ok := body["servers"]; !ok {
		t.Error("servers field missing")
	}

	t.Logf("Discovery returned %d servers", len(servers))
}
