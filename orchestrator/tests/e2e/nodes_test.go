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

// E2E_02: Register node via manual flow — GetNodeInfo gRPC к реальной ноде.
func TestE2E_Nodes_Register_Manual(t *testing.T) {
	env := setupE2E(t)
	env.cleanupTables(t)
	ctx, cancel := context.WithTimeout(context.Background(), e2eTestTimeout)
	defer cancel()

	// Регистрируем ноду через manual flow (адрес + токен).
	// Токен = e2eAPIKey, так как это API key реальной ноды.
	reqBody := map[string]any{
		"address": env.getNodeAddress(t),
		"token":   e2eAPIKey,
		"region":  "e2e-test",
	}
	body, _ := json.Marshal(reqBody)

	resp, err := http.Post(env.baseURL+"/nodes", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /nodes failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, body = %s, want %d", resp.StatusCode, respBody, http.StatusCreated)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if result["status"] != "online" {
		t.Errorf("status = %q, want %q", result["status"], "online")
	}
	if result["region"] != "e2e-test" {
		t.Errorf("region = %q, want %q", result["region"], "e2e-test")
	}
	if _, ok := result["cpu_cores"]; !ok {
		t.Error("cpu_cores field missing — GetNodeInfo gRPC likely failed")
	}

	t.Logf("Node registered: id=%v, address=%s, status=%s, cpu_cores=%v",
		result["id"], result["address"], result["status"], result["cpu_cores"])

	_ = ctx
}

// E2E_03: Register node via authorize flow.
func TestE2E_Nodes_Register_Authorize(t *testing.T) {
	env := setupE2E(t)
	env.cleanupTables(t)
	ctx, cancel := context.WithTimeout(context.Background(), e2eTestTimeout)
	defer cancel()

	// Создаём ноду в unauthorized состоянии.
	now := time.Now()
	token := e2eAPIKey
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
	reqBody := map[string]any{
		"node_id": node.ID,
		"token":   token,
	}
	body, _ := json.Marshal(reqBody)

	resp, err := http.Post(env.baseURL+"/nodes", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /nodes failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusCreated)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if result["status"] != "online" {
		t.Errorf("status = %q, want %q", result["status"], "online")
	}

	t.Logf("Node authorized: id=%v, status=%s", result["id"], result["status"])
}

// E2E_04: List nodes.
func TestE2E_Nodes_List(t *testing.T) {
	env := setupE2E(t)
	env.cleanupTables(t)
	ctx, cancel := context.WithTimeout(context.Background(), e2eTestTimeout)
	defer cancel()

	// Создаём ноду напрямую.
	now := time.Now()
	tokenHash := sha256.Sum256([]byte(e2eAPIKey))
	for i := range 2 {
		node := &domain.Node{
			ID:         int64(i + 1),
			Address:    fmt.Sprintf("e2e-node-%d:44044", i+1),
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
		_ = env.nodeState.UpdateHeartbeat(ctx, node.ID, &domain.ResourceUsage{})
	}

	resp, err := http.Get(env.baseURL + "/nodes")
	if err != nil {
		t.Fatalf("GET /nodes failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	nodes := body["nodes"].([]any)
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}

	t.Logf("Listed %d nodes", len(nodes))
}

// E2E_05: Delete node.
func TestE2E_Nodes_Delete(t *testing.T) {
	env := setupE2E(t)
	env.cleanupTables(t)
	ctx, cancel := context.WithTimeout(context.Background(), e2eTestTimeout)
	defer cancel()

	now := time.Now()
	tokenHash := sha256.Sum256([]byte(e2eAPIKey))

	node := &domain.Node{
		ID:         99,
		Address:    "e2e-delete-node:44044",
		TokenHash:  tokenHash[:],
		APIToken:   e2eAPIKey,
		Status:     domain.NodeStatusOnline,
		LastPingAt: now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := env.nodeRepo.Create(ctx, node); err != nil {
		t.Fatalf("Create node failed: %v", err)
	}

	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/nodes/%d", env.baseURL, node.ID), nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("DELETE /nodes/%d failed: %v", node.ID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNoContent)
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
