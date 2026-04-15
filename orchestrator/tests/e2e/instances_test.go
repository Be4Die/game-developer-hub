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

// E2E_15: List instances — empty.
func TestE2E_Instances_List_Empty(t *testing.T) {
	env := setupE2E(t)
	env.cleanupDB(t)

	resp, err := http.Get(env.baseURL + "/games/42/instances")
	if err != nil {
		t.Fatalf("GET /games/42/instances failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	instances := body["instances"].([]any)
	if len(instances) != 0 {
		t.Errorf("expected 0 instances, got %d", len(instances))
	}

	t.Log("Instances list: empty as expected")
}

// E2E_16: List instances — with data.
func TestE2E_Instances_List(t *testing.T) {
	env := setupE2E(t)
	env.cleanupDB(t)
	ctx := context.Background()
	_ = ctx

	node := createE2ENode(t, env, 1)
	build := createE2EBuild(t, env, 42, 1)

	now := time.Now()
	for i := range 3 {
		inst := &domain.Instance{
			ID:            int64(i + 1),
			NodeID:        node.ID,
			ServerBuildID: build.ID,
			GameID:        42,
			Name:          fmt.Sprintf("instance-%d", i+1),
			BuildVersion:  "v1.0.1",
			Protocol:      domain.ProtocolTCP,
			HostPort:      uint32(7001 + i),
			InternalPort:  8080,
			Status:        domain.InstanceStatusRunning,
			MaxPlayers:    10,
			ServerAddress: node.Address,
			StartedAt:     now,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		if err := env.instanceRepo.Create(ctx, inst); err != nil {
			t.Fatalf("Create instance %d failed: %v", i, err)
		}
		_ = env.instanceState.SetStatus(ctx, inst.ID, domain.InstanceStatusRunning)
		_ = env.instanceState.SetPlayerCount(ctx, inst.ID, uint32(i*2))
	}

	resp, err := http.Get(env.baseURL + "/games/42/instances")
	if err != nil {
		t.Fatalf("GET /games/42/instances failed: %v", err)
	}
	defer resp.Body.Close()

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	instances := body["instances"].([]any)
	if len(instances) != 3 {
		t.Fatalf("expected 3 instances, got %d", len(instances))
	}

	t.Logf("Listed %d instances for game 42", len(instances))
}

// E2E_17: List instances — filter by status.
func TestE2E_Instances_List_FilterByStatus(t *testing.T) {
	env := setupE2E(t)
	env.cleanupDB(t)
	ctx := context.Background()
	_ = ctx

	node := createE2ENode(t, env, 1)
	build := createE2EBuild(t, env, 42, 1)

	now := time.Now()

	running := &domain.Instance{
		ID:            10,
		NodeID:        node.ID,
		ServerBuildID: build.ID,
		GameID:        42,
		Name:          "running-inst",
		BuildVersion:  "v1.0.1",
		Protocol:      domain.ProtocolTCP,
		HostPort:      7010,
		InternalPort:  8080,
		Status:        domain.InstanceStatusRunning,
		MaxPlayers:    10,
		ServerAddress: node.Address,
		StartedAt:     now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := env.instanceRepo.Create(ctx, running); err != nil {
		t.Fatalf("Create running failed: %v", err)
	}

	stopped := &domain.Instance{
		ID:            11,
		NodeID:        node.ID,
		ServerBuildID: build.ID,
		GameID:        42,
		Name:          "stopped-inst",
		BuildVersion:  "v1.0.1",
		Protocol:      domain.ProtocolTCP,
		HostPort:      7011,
		InternalPort:  8080,
		Status:        domain.InstanceStatusStopped,
		MaxPlayers:    10,
		ServerAddress: node.Address,
		StartedAt:     now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := env.instanceRepo.Create(ctx, stopped); err != nil {
		t.Fatalf("Create stopped failed: %v", err)
	}

	resp, err := http.Get(env.baseURL + "/games/42/instances?status=running")
	if err != nil {
		t.Fatalf("GET /games/42/instances?status=running failed: %v", err)
	}
	defer resp.Body.Close()

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	instances := body["instances"].([]any)
	if len(instances) != 1 {
		t.Fatalf("expected 1 running instance, got %d", len(instances))
	}

	t.Logf("Filtered: %d running instances", len(instances))
}

// E2E_18: Get instance by ID.
func TestE2E_Instances_Get(t *testing.T) {
	env := setupE2E(t)
	env.cleanupDB(t)
	ctx := context.Background()
	_ = ctx

	node := createE2ENode(t, env, 1)
	build := createE2EBuild(t, env, 42, 1)

	now := time.Now()
	inst := &domain.Instance{
		ID:            100,
		NodeID:        node.ID,
		ServerBuildID: build.ID,
		GameID:        42,
		Name:          "get-instance-test",
		BuildVersion:  "v1.0.1",
		Protocol:      domain.ProtocolTCP,
		HostPort:      7100,
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
	_ = env.instanceState.SetStatus(ctx, inst.ID, domain.InstanceStatusRunning)
	_ = env.instanceState.SetPlayerCount(ctx, inst.ID, 5)

	resp, err := http.Get(fmt.Sprintf("%s/games/42/instances/%d", env.baseURL, inst.ID))
	if err != nil {
		t.Fatalf("GET /games/42/instances/%d failed: %v", inst.ID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if int64(body["id"].(float64)) != inst.ID {
		t.Errorf("id = %v, want %d", body["id"], inst.ID)
	}
	if body["name"] != "get-instance-test" {
		t.Errorf("name = %q, want %q", body["name"], "get-instance-test")
	}
	if body["status"] != "running" {
		t.Errorf("status = %q, want %q", body["status"], "running")
	}

	t.Logf("Got instance: id=%v, name=%s, status=%s", body["id"], body["name"], body["status"])
}

// E2E_19: Get instance — not found.
func TestE2E_Instances_Get_NotFound(t *testing.T) {
	env := setupE2E(t)
	env.cleanupDB(t)

	resp, err := http.Get(env.baseURL + "/games/42/instances/99999")
	if err != nil {
		t.Fatalf("GET /games/42/instances/99999 failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

// E2E_20: Start instance — via HTTP API with mock node client.
func TestE2E_Instances_Start(t *testing.T) {
	env := setupE2E(t)
	env.cleanupDB(t)
	ctx := context.Background()
	_ = ctx

	// Создаём ноду и билд.
	node := createE2ENode(t, env, 1)
	build := createE2EBuild(t, env, 42, 1)
	_ = node // node используется сервисом для выбора ноды
	_ = build

	// Запускаем инстанс через API.
	reqBody := map[string]any{
		"build_version": "v1.0.1",
		"name":          "e2e-start-test",
	}
	body, _ := json.Marshal(reqBody)

	resp, err := http.Post(env.baseURL+"/games/42/instances", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /games/42/instances failed: %v", err)
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

	if respBody["name"] != "e2e-start-test" {
		t.Errorf("name = %q, want %q", respBody["name"], "e2e-start-test")
	}
	if respBody["status"] != "running" {
		t.Errorf("status = %q, want %q", respBody["status"], "running")
	}

	t.Logf("Started instance: id=%v, name=%s, status=%s", respBody["id"], respBody["name"], respBody["status"])
}

// E2E_21: Start instance — missing build_version.
func TestE2E_Instances_Start_MissingBuildVersion(t *testing.T) {
	env := setupE2E(t)
	env.cleanupDB(t)

	reqBody := map[string]any{
		"name": "no-version-test",
	}
	body, _ := json.Marshal(reqBody)

	resp, err := http.Post(env.baseURL+"/games/42/instances", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /games/42/instances failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

// E2E_22: Start instance — build not found.
func TestE2E_Instances_Start_BuildNotFound(t *testing.T) {
	env := setupE2E(t)
	env.cleanupDB(t)
	ctx := context.Background()
	_ = ctx

	// Создаём ноду но не билд.
	_ = createE2ENode(t, env, 1)

	reqBody := map[string]any{
		"build_version": "nonexistent",
		"name":          "no-build-test",
	}
	body, _ := json.Marshal(reqBody)

	resp, err := http.Post(env.baseURL+"/games/42/instances", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /games/42/instances failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

// E2E_23: Stop instance.
func TestE2E_Instances_Stop(t *testing.T) {
	env := setupE2E(t)
	env.cleanupDB(t)
	ctx := context.Background()
	_ = ctx

	node := createE2ENode(t, env, 1)
	build := createE2EBuild(t, env, 42, 1)

	now := time.Now()
	inst := &domain.Instance{
		ID:            200,
		NodeID:        node.ID,
		ServerBuildID: build.ID,
		GameID:        42,
		Name:          "stop-instance-test",
		BuildVersion:  "v1.0.1",
		Protocol:      domain.ProtocolTCP,
		HostPort:      7200,
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
	_ = env.instanceState.SetStatus(ctx, inst.ID, domain.InstanceStatusRunning)

	req, _ := http.NewRequest(http.MethodDelete,
		fmt.Sprintf("%s/games/42/instances/%d", env.baseURL, inst.ID), nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("DELETE /games/42/instances/%d failed: %v", inst.ID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, body = %s, want %d", resp.StatusCode, respBody, http.StatusOK)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	instBody := body["instance"].(map[string]any)
	if instBody["status"] != "stopped" {
		t.Errorf("status = %q, want %q", instBody["status"], "stopped")
	}

	t.Log("Instance stopped successfully")
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func createE2ENode(t *testing.T, env *e2eTestEnv, id int64) *domain.Node {
	t.Helper()
	ctx := context.Background()
	_ = ctx
	now := time.Now()
	tokenHash := sha256.Sum256([]byte(fmt.Sprintf("token-%d", id)))

	node := &domain.Node{
		ID:         id,
		Address:    fmt.Sprintf("e2e-node-%d.example.com:44044", id),
		TokenHash:  tokenHash[:],
		Status:     domain.NodeStatusOnline,
		LastPingAt: now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := env.nodeRepo.Create(ctx, node); err != nil {
		t.Fatalf("createE2ENode(%d) failed: %v", id, err)
	}
	_ = env.nodeState.UpdateHeartbeat(ctx, node.ID, &domain.ResourceUsage{})
	return node
}

func createE2EBuild(t *testing.T, env *e2eTestEnv, gameID, buildID int64) *domain.ServerBuild {
	t.Helper()
	ctx := context.Background()
	_ = ctx

	build := &domain.ServerBuild{
		ID:           buildID,
		GameID:       gameID,
		Version:      "v1.0.1",
		ImageTag:     fmt.Sprintf("welwise/game-%d:v1.0.1", gameID),
		Protocol:     domain.ProtocolTCP,
		InternalPort: 8080,
		MaxPlayers:   10,
		FileURL:      fmt.Sprintf("/builds/game-%d-v1.tar", gameID),
		FileSize:     1000000,
		CreatedAt:    time.Now(),
	}
	if err := env.buildStorage.Create(ctx, build); err != nil {
		t.Fatalf("createE2EBuild(%d) failed: %v", buildID, err)
	}
	return build
}
