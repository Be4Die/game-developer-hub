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

// E2E_02: Register node via authorization flow.
func TestE2E_Nodes_Register_Authorize(t *testing.T) {
	env := setupE2E(t)
	env.cleanupDB(t)
	ctx := context.Background()

	// Создаём ноду в unauthorized состоянии.
	now := time.Now()
	token := "test-node-token-123"
	tokenHash := sha256.Sum256([]byte(token))

	node := &domain.Node{
		ID:         1,
		Address:    "node1.example.com:44044",
		TokenHash:  tokenHash[:],
		Status:     domain.NodeStatusUnauthorized,
		LastPingAt: now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := env.nodeRepo.Create(ctx, node); err != nil {
		t.Fatalf("Create node failed: %v", err)
	}

	// Авторизуем ноду через API.
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
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, body = %s, want %d", resp.StatusCode, respBody, http.StatusCreated)
	}

	var respBody map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if respBody["status"] != "online" {
		t.Errorf("status = %q, want %q", respBody["status"], "online")
	}
	if respBody["address"] != "node1.example.com:44044" {
		t.Errorf("address = %q, want %q", respBody["address"], "node1.example.com:44044")
	}

	t.Logf("Node registered: id=%v, status=%s", respBody["id"], respBody["status"])
}

// E2E_03: Register node — invalid token.
func TestE2E_Nodes_Register_InvalidToken(t *testing.T) {
	env := setupE2E(t)
	env.cleanupDB(t)
	ctx := context.Background()

	now := time.Now()
	tokenHash := sha256.Sum256([]byte("correct-token"))

	node := &domain.Node{
		ID:         1,
		Address:    "node1.example.com:44044",
		TokenHash:  tokenHash[:],
		Status:     domain.NodeStatusUnauthorized,
		LastPingAt: now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := env.nodeRepo.Create(ctx, node); err != nil {
		t.Fatalf("Create node failed: %v", err)
	}

	// Пытаемся авторизоваться с неправильным токеном.
	reqBody := map[string]any{
		"node_id": node.ID,
		"token":   "wrong-token",
	}
	body, _ := json.Marshal(reqBody)

	resp, err := http.Post(env.baseURL+"/nodes", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /nodes failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

// E2E_04: List nodes.
func TestE2E_Nodes_List(t *testing.T) {
	env := setupE2E(t)
	env.cleanupDB(t)
	ctx := context.Background()

	// Создаём 2 ноды.
	now := time.Now()
	tokenHash := sha256.Sum256([]byte("token"))

	for i := range 2 {
		node := &domain.Node{
			ID:         int64(i + 1),
			Address:    fmt.Sprintf("node%d.example.com:44044", i+1),
			TokenHash:  tokenHash[:],
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

// E2E_05: List nodes — filter by status.
func TestE2E_Nodes_List_FilterByStatus(t *testing.T) {
	env := setupE2E(t)
	env.cleanupDB(t)
	ctx := context.Background()

	now := time.Now()
	tokenHash := sha256.Sum256([]byte("token"))

	online := &domain.Node{
		ID:         1,
		Address:    "online-node.example.com:44044",
		TokenHash:  tokenHash[:],
		Status:     domain.NodeStatusOnline,
		LastPingAt: now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := env.nodeRepo.Create(ctx, online); err != nil {
		t.Fatalf("Create online failed: %v", err)
	}
	_ = env.nodeState.UpdateHeartbeat(ctx, online.ID, &domain.ResourceUsage{})

	unauth := &domain.Node{
		ID:         2,
		Address:    "unauth-node.example.com:44044",
		TokenHash:  tokenHash[:],
		Status:     domain.NodeStatusUnauthorized,
		LastPingAt: now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := env.nodeRepo.Create(ctx, unauth); err != nil {
		t.Fatalf("Create unauth failed: %v", err)
	}

	resp, err := http.Get(env.baseURL + "/nodes?status=online")
	if err != nil {
		t.Fatalf("GET /nodes?status=online failed: %v", err)
	}
	defer resp.Body.Close()

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	nodes := body["nodes"].([]any)
	if len(nodes) != 1 {
		t.Fatalf("expected 1 online node, got %d", len(nodes))
	}

	t.Logf("Filtered list: %d online nodes", len(nodes))
}

// E2E_06: Get node by ID.
func TestE2E_Nodes_Get(t *testing.T) {
	env := setupE2E(t)
	env.cleanupDB(t)
	ctx := context.Background()

	now := time.Now()
	tokenHash := sha256.Sum256([]byte("token"))

	node := &domain.Node{
		ID:         10,
		Address:    "get-node.example.com:44044",
		TokenHash:  tokenHash[:],
		Region:     "us-east",
		Status:     domain.NodeStatusOnline,
		CPUCores:   8,
		LastPingAt: now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := env.nodeRepo.Create(ctx, node); err != nil {
		t.Fatalf("Create node failed: %v", err)
	}
	_ = env.nodeState.UpdateHeartbeat(ctx, node.ID, &domain.ResourceUsage{})

	resp, err := http.Get(fmt.Sprintf("%s/nodes/%d", env.baseURL, node.ID))
	if err != nil {
		t.Fatalf("GET /nodes/%d failed: %v", node.ID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if int64(body["id"].(float64)) != node.ID {
		t.Errorf("id = %v, want %d", body["id"], node.ID)
	}
	if body["address"] != "get-node.example.com:44044" {
		t.Errorf("address = %q, want %q", body["address"], "get-node.example.com:44044")
	}
	if body["region"] != "us-east" {
		t.Errorf("region = %q, want %q", body["region"], "us-east")
	}
	if body["status"] != "online" {
		t.Errorf("status = %q, want %q", body["status"], "online")
	}

	t.Logf("Got node: id=%v, address=%s, region=%s", body["id"], body["address"], body["region"])
}

// E2E_07: Get node — not found.
func TestE2E_Nodes_Get_NotFound(t *testing.T) {
	env := setupE2E(t)
	env.cleanupDB(t)

	resp, err := http.Get(env.baseURL + "/nodes/99999")
	if err != nil {
		t.Fatalf("GET /nodes/99999 failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

// E2E_08: Delete node.
func TestE2E_Nodes_Delete(t *testing.T) {
	env := setupE2E(t)
	env.cleanupDB(t)
	ctx := context.Background()

	now := time.Now()
	tokenHash := sha256.Sum256([]byte("token"))

	node := &domain.Node{
		ID:         20,
		Address:    "delete-node.example.com:44044",
		TokenHash:  tokenHash[:],
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

	// Verify deleted.
	_, err = env.nodeRepo.GetByID(ctx, node.ID)
	if err != domain.ErrNotFound {
		t.Errorf("expected node to be deleted, got %v", err)
	}

	t.Logf("Node %d deleted successfully", node.ID)
}
