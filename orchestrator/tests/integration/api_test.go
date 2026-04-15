//go:build integration

package integration

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/config"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/service"
	orchhttp "github.com/Be4Die/game-developer-hub/orchestrator/internal/transport/http"
)

// apiTestEnv хранит компоненты для API-тестов.
type apiTestEnv struct {
	*integrationTestEnv
	server  *httptest.Server
	baseURL string
}

// setupAPITest запускает контейнеры и создаёт HTTP сервер через httptest.
func setupAPITest(t *testing.T) *apiTestEnv {
	t.Helper()

	intEnv := setupIntegration(t)

	limits := config.LimitsConfig{
		MaxBuildsPerGame:    10,
		MaxInstancesPerGame: 50,
		MaxLogTailLines:     5000,
		MaxBuildSizeBytes:   2147483648,
	}

	buildPipeline := service.NewBuildPipeline(
		intEnv.buildStorage, nil, nil, intEnv.nodeRepo, intEnv.nodeState, limits,
	)

	instanceService := service.NewInstanceService(
		intEnv.instanceRepo, intEnv.instanceState, intEnv.buildStorage,
		intEnv.nodeRepo, intEnv.nodeState, nil, limits,
	)

	discoveryService := service.NewDiscoveryService(
		intEnv.instanceRepo, intEnv.instanceState, intEnv.nodeRepo,
	)

	nodeService := service.NewNodeService(
		intEnv.nodeRepo, intEnv.nodeState, intEnv.instanceRepo, intEnv.instanceState, nil,
	)

	buildHandler := orchhttp.NewBuildHandler(buildPipeline)
	instanceHandler := orchhttp.NewInstanceHandler(instanceService, limits.MaxLogTailLines)
	discoveryHandler := orchhttp.NewDiscoveryHandler(discoveryService)
	nodeHandler := orchhttp.NewNodeHandler(nodeService)
	healthHandler := orchhttp.NewHealthHandler("test-1.0.0")

	router := orchhttp.NewRouter(
		buildHandler, instanceHandler, discoveryHandler, nodeHandler, healthHandler, intEnv.log,
	)

	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)

	return &apiTestEnv{
		integrationTestEnv: intEnv,
		server:             srv,
		baseURL:            srv.URL,
	}
}

// ─── Health API ──────────────────────────────────────────────────────────────

func TestAPI_Health(t *testing.T) {
	env := setupAPITest(t)
	env.Cleanup(t)

	resp, err := http.Get(env.baseURL + "/health")
	if err != nil {
		t.Fatalf("GET /health failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("status = %q, want %q", body["status"], "ok")
	}
	if body["version"] != "test-1.0.0" {
		t.Errorf("version = %q, want %q", body["version"], "test-1.0.0")
	}
}

// ─── Node API ────────────────────────────────────────────────────────────────

func TestAPI_Nodes_RegisterAndList(t *testing.T) {
	env := setupAPITest(t)
	env.Cleanup(t)
	ctx := context.Background()

	now := time.Now()
	tokenHash := sha256.Sum256([]byte("test-token"))
	node := &domain.Node{
		ID:         1,
		Address:    "api-node-list.example.com:44044",
		TokenHash:  tokenHash[:],
		Status:     domain.NodeStatusUnauthorized,
		LastPingAt: now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := env.nodeRepo.Create(ctx, node); err != nil {
		t.Fatalf("Create node failed: %v", err)
	}
	// Populate KV to avoid nil Usage in enrichedNodeResponse
	_ = env.nodeState.UpdateHeartbeat(ctx, node.ID, &domain.ResourceUsage{})
	// Populate KV to avoid nil Usage in enrichedNodeResponse
	_ = env.nodeState.UpdateHeartbeat(ctx, node.ID, &domain.ResourceUsage{})

	resp, err := http.Get(env.baseURL + "/nodes")
	if err != nil {
		t.Fatalf("GET /nodes failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, body = %s", resp.StatusCode, body)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	nodes := body["nodes"].([]any)
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
}

func TestAPI_Nodes_Get(t *testing.T) {
	env := setupAPITest(t)
	env.Cleanup(t)
	ctx := context.Background()

	now := time.Now()
	tokenHash := sha256.Sum256([]byte("test-token"))
	node := &domain.Node{
		ID:         10,
		Address:    "api-node-get.example.com:44044",
		TokenHash:  tokenHash[:],
		Status:     domain.NodeStatusOnline,
		LastPingAt: now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := env.nodeRepo.Create(ctx, node); err != nil {
		t.Fatalf("Create node failed: %v", err)
	}
	// Populate KV to avoid nil Usage in enrichedNodeResponse
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
}

func TestAPI_Nodes_Get_NotFound(t *testing.T) {
	env := setupAPITest(t)
	env.Cleanup(t)

	resp, err := http.Get(env.baseURL + "/nodes/99999")
	if err != nil {
		t.Fatalf("GET /nodes/99999 failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestAPI_Nodes_Delete(t *testing.T) {
	env := setupAPITest(t)
	env.Cleanup(t)
	ctx := context.Background()

	now := time.Now()
	tokenHash := sha256.Sum256([]byte("test-token"))
	node := &domain.Node{
		ID:         20,
		Address:    "api-node-del.example.com:44044",
		TokenHash:  tokenHash[:],
		Status:     domain.NodeStatusOnline,
		LastPingAt: now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := env.nodeRepo.Create(ctx, node); err != nil {
		t.Fatalf("Create node failed: %v", err)
	}
	// Populate KV to avoid nil Usage in enrichedNodeResponse
	_ = env.nodeState.UpdateHeartbeat(ctx, node.ID, &domain.ResourceUsage{})

	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/nodes/%d", env.baseURL, node.ID), nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("DELETE /nodes/%d failed: %v", node.ID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, body = %s, want %d", resp.StatusCode, body, http.StatusNoContent)
	}

	_, err = env.nodeRepo.GetByID(ctx, node.ID)
	if err != domain.ErrNotFound {
		t.Errorf("expected node to be deleted, got %v", err)
	}
}

// ─── Builds API ──────────────────────────────────────────────────────────────

func TestAPI_Builds_List_Empty(t *testing.T) {
	env := setupAPITest(t)
	env.Cleanup(t)

	resp, err := http.Get(env.baseURL + "/games/42/builds")
	if err != nil {
		t.Fatalf("GET /games/42/builds failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	builds := body["builds"].([]any)
	if len(builds) != 0 {
		t.Errorf("expected 0 builds, got %d", len(builds))
	}
}

func TestAPI_Builds_Get_NotFound(t *testing.T) {
	env := setupAPITest(t)
	env.Cleanup(t)

	resp, err := http.Get(env.baseURL + "/games/42/builds/nonexistent")
	if err != nil {
		t.Fatalf("GET /games/42/builds/nonexistent failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestAPI_Builds_Upload_MissingFields(t *testing.T) {
	env := setupAPITest(t)
	env.Cleanup(t)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("image", "test.zip")
	part.Write([]byte("dummy data"))
	writer.Close()

	req, _ := http.NewRequest(http.MethodPost, env.baseURL+"/games/42/builds", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("POST /games/42/builds failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestAPI_Builds_Upload_InvalidGameId(t *testing.T) {
	env := setupAPITest(t)
	env.Cleanup(t)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("image", "test.zip")
	part.Write([]byte("dummy data"))
	writer.WriteField("build_version", "v1.0.0")
	writer.WriteField("protocol", "tcp")
	writer.WriteField("internal_port", "8080")
	writer.WriteField("max_players", "10")
	writer.Close()

	req, _ := http.NewRequest(http.MethodPost, env.baseURL+"/games/abc/builds", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("POST /games/abc/builds failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

// ─── Instances API ───────────────────────────────────────────────────────────

func TestAPI_Instances_List_Empty(t *testing.T) {
	env := setupAPITest(t)
	env.Cleanup(t)

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
}

func TestAPI_Instances_Get_NotFound(t *testing.T) {
	env := setupAPITest(t)
	env.Cleanup(t)

	resp, err := http.Get(env.baseURL + "/games/42/instances/99999")
	if err != nil {
		t.Fatalf("GET /games/42/instances/99999 failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestAPI_Instances_Start_MissingBody(t *testing.T) {
	env := setupAPITest(t)
	env.Cleanup(t)

	resp, err := http.Post(env.baseURL+"/games/42/instances", "application/json", nil)
	if err != nil {
		t.Fatalf("POST /games/42/instances failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestAPI_Instances_Stop_InvalidGameId(t *testing.T) {
	env := setupAPITest(t)
	env.Cleanup(t)

	req, _ := http.NewRequest(http.MethodDelete, env.baseURL+"/games/abc/instances/1", nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("DELETE /games/abc/instances/1 failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

// ─── Discovery API ───────────────────────────────────────────────────────────

func TestAPI_Discovery_Empty(t *testing.T) {
	env := setupAPITest(t)
	env.Cleanup(t)

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
}

func TestAPI_Discovery_InvalidGameId(t *testing.T) {
	env := setupAPITest(t)
	env.Cleanup(t)

	resp, err := http.Get(env.baseURL + "/games/abc/servers")
	if err != nil {
		t.Fatalf("GET /games/abc/servers failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

// ─── Full User Journey ──────────────────────────────────────────────────────

func TestAPI_FullLifecycle(t *testing.T) {
	env := setupAPITest(t)
	env.Cleanup(t)
	ctx := context.Background()

	// 1. Create node.
	now := time.Now()
	tokenHash := sha256.Sum256([]byte("lifecycle-token"))
	node := &domain.Node{
		ID:         100,
		Address:    "lifecycle-node.example.com:44044",
		TokenHash:  tokenHash[:],
		Status:     domain.NodeStatusOnline,
		LastPingAt: now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := env.nodeRepo.Create(ctx, node); err != nil {
		t.Fatalf("Create node failed: %v", err)
	}
	// Populate KV to avoid nil Usage in enrichedNodeResponse
	_ = env.nodeState.UpdateHeartbeat(ctx, node.ID, &domain.ResourceUsage{})

	// 2. Create build.
	build := &domain.ServerBuild{
		ID:           100,
		GameID:       42,
		Version:      "v1.0.0",
		ImageTag:     "welwise/game-42:v1.0.0",
		Protocol:     domain.ProtocolTCP,
		InternalPort: 8080,
		MaxPlayers:   10,
		FileURL:      "/builds/lifecycle.tar",
		FileSize:     1000000,
		CreatedAt:    now,
	}
	if err := env.buildStorage.Create(ctx, build); err != nil {
		t.Fatalf("Create build failed: %v", err)
	}

	// 3. Verify build exists.
	resp, err := http.Get(env.baseURL + "/games/42/builds")
	if err != nil {
		t.Fatalf("GET /games/42/builds failed: %v", err)
	}
	defer resp.Body.Close()

	var buildBody map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&buildBody); err != nil {
		t.Fatalf("decode builds failed: %v", err)
	}
	builds := buildBody["builds"].([]any)
	if len(builds) != 1 {
		t.Fatalf("expected 1 build, got %d", len(builds))
	}

	// 4. Verify node exists.
	resp, err = http.Get(env.baseURL + "/nodes")
	if err != nil {
		t.Fatalf("GET /nodes failed: %v", err)
	}
	defer resp.Body.Close()

	var nodeBody map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&nodeBody); err != nil {
		t.Fatalf("decode nodes failed: %v", err)
	}
	nodes := nodeBody["nodes"].([]any)
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}

	// 5. Verify discovery is empty.
	resp, err = http.Get(env.baseURL + "/games/42/servers")
	if err != nil {
		t.Fatalf("GET /games/42/servers failed: %v", err)
	}
	defer resp.Body.Close()

	var discBody map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&discBody); err != nil {
		t.Fatalf("decode discovery failed: %v", err)
	}
	servers := discBody["servers"].([]any)
	if len(servers) != 0 {
		t.Fatalf("expected 0 servers, got %d", len(servers))
	}

	// 6. Delete node.
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/nodes/%d", env.baseURL, node.ID), nil)
	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("DELETE /nodes/%d failed: %v", node.ID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("delete node status = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}
}

// ─── Response Format Tests ──────────────────────────────────────────────────

func TestAPI_ResponseFormats(t *testing.T) {
	env := setupAPITest(t)
	env.Cleanup(t)
	ctx := context.Background()

	now := time.Now()
	tokenHash := sha256.Sum256([]byte("format-token"))
	node := &domain.Node{
		ID:           200,
		Address:      "format-node.example.com:44044",
		TokenHash:    tokenHash[:],
		Region:       "us-east",
		Status:       domain.NodeStatusOnline,
		CPUCores:     4,
		TotalMemory:  8000000000,
		TotalDisk:    250000000000,
		AgentVersion: "1.0.0",
		LastPingAt:   now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := env.nodeRepo.Create(ctx, node); err != nil {
		t.Fatalf("Create node failed: %v", err)
	}
	// Populate KV to avoid nil Usage in enrichedNodeResponse
	_ = env.nodeState.UpdateHeartbeat(ctx, node.ID, &domain.ResourceUsage{})

	resp, err := http.Get(fmt.Sprintf("%s/nodes/%d", env.baseURL, node.ID))
	if err != nil {
		t.Fatalf("GET /nodes/%d failed: %v", node.ID, err)
	}
	defer resp.Body.Close()

	if resp.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %q, want %q", resp.Header.Get("Content-Type"), "application/json")
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	requiredFields := []string{
		"id", "address", "region", "status", "cpu_cores",
		"total_memory_bytes", "total_disk_bytes", "agent_version",
		"last_ping_at", "created_at", "updated_at",
	}
	for _, field := range requiredFields {
		if _, ok := body[field]; !ok {
			t.Errorf("missing field %q in response", field)
		}
	}
}

// ─── Filter Tests ───────────────────────────────────────────────────────────

func TestAPI_Nodes_FilterByStatus(t *testing.T) {
	env := setupAPITest(t)
	env.Cleanup(t)
	ctx := context.Background()

	now := time.Now()
	tokenHash := sha256.Sum256([]byte("filter-token"))

	online := &domain.Node{
		ID:         300,
		Address:    "online-filter.example.com:44044",
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
		ID:         301,
		Address:    "unauth-filter.example.com:44044",
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

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, body = %s", resp.StatusCode, body)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	nodes := body["nodes"].([]any)
	if len(nodes) != 1 {
		t.Fatalf("expected 1 online node, got %d", len(nodes))
	}
}

func TestAPI_Instances_FilterByStatus(t *testing.T) {
	env := setupAPITest(t)
	env.Cleanup(t)
	ctx := context.Background()

	node := makeTestNode(t, env.integrationTestEnv, 400)
	if err := env.nodeRepo.Create(ctx, node); err != nil {
		t.Fatalf("create node failed: %v", err)
	}
	// Populate KV to avoid nil Usage in enrichedNodeResponse
	_ = env.nodeState.UpdateHeartbeat(ctx, node.ID, &domain.ResourceUsage{})
	build := makeTestBuildFull(t, 400, 42, "v1.0.0")
	if err := env.buildStorage.Create(ctx, build); err != nil {
		t.Fatalf("Create build failed: %v", err)
	}

	running := makeTestInstance(t, 4000, node.ID, build.ID, 42)
	if err := env.instanceRepo.Create(ctx, running); err != nil {
		t.Fatalf("Create running failed: %v", err)
	}

	stopped := makeTestInstance(t, 4001, node.ID, build.ID, 42)
	stopped.Status = domain.InstanceStatusStopped
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
}
