package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/service"
	"github.com/go-chi/chi/v5"
)

// ─── instance_handler tests ──────────────────────────────────────────────────

func TestEnrichedInstanceResponse(t *testing.T) {
	inst := &service.EnrichedInstance{
		Instance: &domain.Instance{
			ID:            1,
			GameID:        42,
			NodeID:        10,
			BuildVersion:  "v1.0",
			Name:          "test-server",
			Protocol:      domain.ProtocolTCP,
			HostPort:      7001,
			InternalPort:  8080,
			PlayerCount:   addrUint32(5),
			MaxPlayers:    10,
			ServerAddress: "node:44044",
		},
		Status:      domain.InstanceStatusRunning,
		PlayerCount: addrUint32(7), // KV overrides
	}

	resp := enrichedInstanceResponse(inst)

	if resp["id"].(int64) != 1 {
		t.Errorf("id = %v, want 1", resp["id"])
	}
	if resp["status"].(string) != "running" {
		t.Errorf("status = %q, want %q", resp["status"], "running")
	}
	if resp["protocol"].(string) != "tcp" {
		t.Errorf("protocol = %q, want %q", resp["protocol"], "tcp")
	}
	// KV player count должен приоритизироваться
	pc := resp["player_count"].(*uint32)
	if pc == nil || *pc != 7 {
		t.Errorf("player_count = %v, want 7", resp["player_count"])
	}
}

func TestEnrichedInstanceResponse_NilKVPlayerCount(t *testing.T) {
	inst := &service.EnrichedInstance{
		Instance: &domain.Instance{
			ID:          1,
			PlayerCount: addrUint32(3),
		},
		Status:      domain.InstanceStatusStarting,
		PlayerCount: nil, // KV недоступен
	}

	resp := enrichedInstanceResponse(inst)
	pc := resp["player_count"].(*uint32)
	if pc == nil || *pc != 3 {
		t.Errorf("expected fallback to instance player_count, got %v", resp["player_count"])
	}
}

func TestInstanceResponse(t *testing.T) {
	inst := &domain.Instance{
		ID:            100,
		GameID:        42,
		NodeID:        10,
		BuildVersion:  "v2.0",
		Name:          "my-server",
		Protocol:      domain.ProtocolWebSocket,
		HostPort:      9000,
		InternalPort:  8080,
		Status:        domain.InstanceStatusStopped,
		PlayerCount:   addrUint32(0),
		MaxPlayers:    20,
		ServerAddress: "node2:44044",
	}

	resp := instanceResponse(inst, nil)
	if resp["id"].(int64) != 100 {
		t.Errorf("id = %v, want 100", resp["id"])
	}
	if resp["status"].(string) != "stopped" {
		t.Errorf("status = %q, want %q", resp["status"], "stopped")
	}
	if resp["protocol"].(string) != "websocket" {
		t.Errorf("protocol = %q, want %q", resp["protocol"], "websocket")
	}
	pc := resp["player_count"].(*uint32)
	if pc == nil || *pc != 0 {
		t.Errorf("player_count = %v, want 0", resp["player_count"])
	}
}

func TestResourceUsageResponse(t *testing.T) {
	u := &domain.ResourceUsage{
		CPUUsagePercent:    55.5,
		MemoryUsedBytes:    1024,
		DiskUsedBytes:      2048,
		NetworkBytesPerSec: 500,
	}
	resp := resourceUsageResponse(u)
	if resp["cpu_usage_percent"].(float64) != 55.5 {
		t.Errorf("cpu = %v, want 55.5", resp["cpu_usage_percent"])
	}
	if resp["memory_used_bytes"].(uint64) != 1024 {
		t.Errorf("memory = %v, want 1024", resp["memory_used_bytes"])
	}
}

func TestEscapeSSE(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"hello", "hello"},
		{"line1\nline2", "line1\\nline2"},
		{"say \"hi\"", "say \\\"hi\\\""},
		{"back\\slash", "back\\\\slash"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := escapeSSE(tt.input); got != tt.want {
				t.Errorf("escapeSSE(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestPortAllocReq_ToDomain(t *testing.T) {
	tests := []struct {
		name  string
		req   portAllocReq
		check func(domain.PortAllocation) bool
	}{
		{
			name:  "any",
			req:   portAllocReq{Strategy: "any"},
			check: func(pa domain.PortAllocation) bool { return pa.Any },
		},
		{
			name:  "exact",
			req:   portAllocReq{Strategy: "exact", Port: addrUint32(8080)},
			check: func(pa domain.PortAllocation) bool { return pa.Exact == 8080 },
		},
		{
			name: "range",
			req:  portAllocReq{Strategy: "range", MinPort: addrUint32(7000), MaxPort: addrUint32(8000)},
			check: func(pa domain.PortAllocation) bool {
				return pa.Range != nil && pa.Range.Min == 7000 && pa.Range.Max == 8000
			},
		},
		{
			name:  "unknown strategy",
			req:   portAllocReq{Strategy: "foo"},
			check: func(pa domain.PortAllocation) bool { return !pa.Any && pa.Exact == 0 && pa.Range == nil },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.req.toDomain()
			if !tt.check(result) {
				t.Errorf("toDomain() check failed for %+v", tt.req)
			}
		})
	}
}

func TestResourceLimitsReq_ToDomain(t *testing.T) {
	req := &resourceLimitsReq{
		CPUMillis:   addrUint32(500),
		MemoryBytes: addrUint64(1073741824),
	}
	result := req.toDomain()
	if result.CPUMillis == nil || *result.CPUMillis != 500 {
		t.Errorf("cpu_millis = %v, want 500", result.CPUMillis)
	}
	if result.MemoryBytes == nil || *result.MemoryBytes != 1073741824 {
		t.Errorf("memory_bytes = %v, want 1073741824", result.MemoryBytes)
	}

	// Nil receiver
	var nilReq *resourceLimitsReq
	if nilReq.toDomain() != nil {
		t.Error("nil receiver should return nil")
	}
}

func TestInstanceHandler_Stop_InvalidGameId(t *testing.T) {
	handler := NewInstanceHandler(nil, 5000)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/games/abc/instances/1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("gameId", "abc")
	rctx.URLParams.Add("instanceId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.Stop(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestInstanceHandler_Stop_InvalidInstanceId(t *testing.T) {
	handler := NewInstanceHandler(nil, 5000)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/games/42/instances/xyz", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("gameId", "42")
	rctx.URLParams.Add("instanceId", "xyz")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.Stop(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestInstanceHandler_Get_InvalidGameId(t *testing.T) {
	handler := NewInstanceHandler(nil, 5000)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/games/abc/instances/1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("gameId", "abc")
	rctx.URLParams.Add("instanceId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.Get(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestInstanceHandler_GetUsage_InvalidGameId(t *testing.T) {
	handler := NewInstanceHandler(nil, 5000)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/games/abc/instances/1/usage", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("gameId", "abc")
	rctx.URLParams.Add("instanceId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.GetUsage(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

// ─── node_handler tests ──────────────────────────────────────────────────────

func TestEnrichedNodeResponse(t *testing.T) {
	now := time.Now()
	usage := &domain.ResourceUsage{CPUUsagePercent: 30.0}
	activeCount := uint32(5)

	n := &service.EnrichedNode{
		Node: &domain.Node{
			ID:           1,
			Address:      "node1:44044",
			Region:       "eu-west",
			Status:       domain.NodeStatusOnline,
			CPUCores:     8,
			TotalMemory:  16000000000,
			TotalDisk:    500000000000,
			AgentVersion: "1.0.0",
			LastPingAt:   now,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		Usage:               usage,
		ActiveInstanceCount: &activeCount,
	}

	resp := enrichedNodeResponse(n)
	if resp["id"].(int64) != 1 {
		t.Errorf("id = %v, want 1", resp["id"])
	}
	if resp["status"].(string) != "online" {
		t.Errorf("status = %q, want %q", resp["status"], "online")
	}
	if resp["cpu_cores"].(uint32) != 8 {
		t.Errorf("cpu_cores = %v, want 8", resp["cpu_cores"])
	}
	if resp["active_instance_count"].(*uint32) == nil || *resp["active_instance_count"].(*uint32) != 5 {
		t.Errorf("active_instance_count = %v, want 5", resp["active_instance_count"])
	}
}

func TestNodeResponse(t *testing.T) {
	now := time.Now()
	n := &domain.Node{
		ID:           42,
		Address:      "node2:44044",
		Region:       "us-east",
		Status:       domain.NodeStatusOffline,
		CPUCores:     4,
		TotalMemory:  8000000000,
		TotalDisk:    250000000000,
		AgentVersion: "0.9.0",
		LastPingAt:   now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	resp := nodeResponse(n, nil)
	if resp["id"].(int64) != 42 {
		t.Errorf("id = %v, want 42", resp["id"])
	}
	if resp["status"].(string) != "offline" {
		t.Errorf("status = %q, want %q", resp["status"], "offline")
	}
	if resp["address"].(string) != "node2:44044" {
		t.Errorf("address = %q, want %q", resp["address"], "node2:44044")
	}
}

func TestNodeHandler_Register_InvalidJSON(t *testing.T) {
	handler := NewNodeHandler(nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/nodes", strings.NewReader("{invalid"))

	handler.Register(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestNodeHandler_Get_InvalidNodeId(t *testing.T) {
	handler := NewNodeHandler(nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/nodes/abc", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("nodeId", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.Get(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestNodeHandler_Delete_InvalidNodeId(t *testing.T) {
	handler := NewNodeHandler(nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/nodes/xyz", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("nodeId", "xyz")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.Delete(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestNodeHandler_GetUsage_InvalidNodeId(t *testing.T) {
	handler := NewNodeHandler(nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/nodes/abc/usage", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("nodeId", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.GetUsage(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestNodeHandler_List_StatusFilter(t *testing.T) {
	// Используем моки из service пакета
	svc := newTestNodeService()
	handler := NewNodeHandler(svc)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/nodes?status=online", nil)

	handler.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}
	nodes := body["nodes"].([]any)
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
}

// ─── build_handler tests ────────────────────────────────────────────────────

func TestBuildResponse(t *testing.T) {
	now := time.Now()
	b := &domain.ServerBuild{
		ID:           1,
		GameID:       42,
		Version:      "v1.0.0",
		ImageTag:     "welwise/game-42:v1.0.0",
		Protocol:     domain.ProtocolUDP,
		InternalPort: 7777,
		MaxPlayers:   32,
		FileSize:     1000000,
		CreatedAt:    now,
	}

	resp := buildResponse(b)
	if resp["id"].(int64) != 1 {
		t.Errorf("id = %v, want 1", resp["id"])
	}
	if resp["protocol"].(string) != "udp" {
		t.Errorf("protocol = %q, want %q", resp["protocol"], "udp")
	}
	if resp["build_version"].(string) != "v1.0.0" {
		t.Errorf("build_version = %q, want %q", resp["build_version"], "v1.0.0")
	}
}

func TestBuildHandler_Upload_InvalidGameId(t *testing.T) {
	handler := NewBuildHandler(nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/games/abc/builds", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("gameId", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.Upload(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestBuildHandler_Upload_MissingMultipart(t *testing.T) {
	handler := NewBuildHandler(nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/games/42/builds", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("gameId", "42")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.Upload(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestBuildHandler_List_InvalidGameId(t *testing.T) {
	handler := NewBuildHandler(nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/games/abc/builds", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("gameId", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.List(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestBuildHandler_Get_MissingBuildVersion(t *testing.T) {
	handler := NewBuildHandler(nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/games/42/builds/", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("gameId", "42")
	rctx.URLParams.Add("buildVersion", "")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.Get(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestBuildHandler_Delete_MissingBuildVersion(t *testing.T) {
	handler := NewBuildHandler(nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/games/42/builds/", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("gameId", "42")
	rctx.URLParams.Add("buildVersion", "")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.Delete(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

// ─── Helper: build a test NodeService with mock ──────────────────────────────

func newTestNodeService() *service.NodeService {
	nodeRepo := &testNodeRepo{
		listFn: func(ctx context.Context, status *domain.NodeStatus) ([]*domain.Node, error) {
			if status != nil && *status == domain.NodeStatusOnline {
				return []*domain.Node{{
					ID:      1,
					Address: "node1:44044",
					Status:  domain.NodeStatusOnline,
				}}, nil
			}
			return nil, nil
		},
	}
	return service.NewNodeService(
		nodeRepo,
		&testNodeStateStore{},
		&testInstanceRepo{},
		&testInstanceState{},
		nil, // nodeClient
	)
}

type testNodeRepo struct {
	createFn         func(ctx context.Context, node *domain.Node) error
	updateFn         func(ctx context.Context, node *domain.Node) error
	getByIDFn        func(ctx context.Context, id int64) (*domain.Node, error)
	getByAddressFn   func(ctx context.Context, addr string) (*domain.Node, error)
	listFn           func(ctx context.Context, status *domain.NodeStatus) ([]*domain.Node, error)
	deleteFn         func(ctx context.Context, id int64) error
	updateLastPingFn func(ctx context.Context, id int64) error
}

func (m *testNodeRepo) Create(ctx context.Context, node *domain.Node) error {
	if m.createFn != nil {
		return m.createFn(ctx, node)
	}
	return nil
}
func (m *testNodeRepo) Update(ctx context.Context, node *domain.Node) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, node)
	}
	return nil
}
func (m *testNodeRepo) GetByID(ctx context.Context, id int64) (*domain.Node, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}
func (m *testNodeRepo) GetByAddress(ctx context.Context, addr string) (*domain.Node, error) {
	if m.getByAddressFn != nil {
		return m.getByAddressFn(ctx, addr)
	}
	return nil, nil
}
func (m *testNodeRepo) List(ctx context.Context, status *domain.NodeStatus) ([]*domain.Node, error) {
	if m.listFn != nil {
		return m.listFn(ctx, status)
	}
	return nil, nil
}
func (m *testNodeRepo) Delete(ctx context.Context, id int64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}
func (m *testNodeRepo) UpdateLastPing(ctx context.Context, id int64) error {
	if m.updateLastPingFn != nil {
		return m.updateLastPingFn(ctx, id)
	}
	return nil
}

type testNodeStateStore struct{}

func (m *testNodeStateStore) UpdateHeartbeat(ctx context.Context, nodeID int64, usage *domain.ResourceUsage) error {
	return nil
}
func (m *testNodeStateStore) GetUsage(ctx context.Context, nodeID int64) (*domain.ResourceUsage, error) {
	return &domain.ResourceUsage{}, nil
}
func (m *testNodeStateStore) GetActiveInstanceCount(ctx context.Context, nodeID int64) (uint32, error) {
	return 0, nil
}
func (m *testNodeStateStore) SetActiveInstanceCount(ctx context.Context, nodeID int64, count uint32) error {
	return nil
}
func (m *testNodeStateStore) Delete(ctx context.Context, nodeID int64) error {
	return nil
}

type testInstanceRepo struct{}

func (m *testInstanceRepo) Create(ctx context.Context, instance *domain.Instance) error { return nil }
func (m *testInstanceRepo) GetByID(ctx context.Context, id int64) (*domain.Instance, error) {
	return nil, nil
}
func (m *testInstanceRepo) ListByGame(ctx context.Context, gameID int64, status *domain.InstanceStatus) ([]*domain.Instance, error) {
	return nil, nil
}
func (m *testInstanceRepo) ListByNode(ctx context.Context, nodeID int64) ([]*domain.Instance, error) {
	return nil, nil
}
func (m *testInstanceRepo) Update(ctx context.Context, instance *domain.Instance) error { return nil }
func (m *testInstanceRepo) Delete(ctx context.Context, id int64) error                  { return nil }
func (m *testInstanceRepo) CountByGame(ctx context.Context, gameID int64) (int, error)  { return 0, nil }
func (m *testInstanceRepo) GetByVersion(ctx context.Context, gameID int64, version string) (*domain.ServerBuild, error) {
	return nil, nil
}
func (m *testInstanceRepo) ListByGameWithLimit(ctx context.Context, gameID int64, limit int) ([]*domain.ServerBuild, error) {
	return nil, nil
}
func (m *testInstanceRepo) CountActiveInstancesByBuild(ctx context.Context, buildID int64) (int, error) {
	return 0, nil
}

type testInstanceState struct{}

func (m *testInstanceState) SetStatus(ctx context.Context, instanceID int64, status domain.InstanceStatus) error {
	return nil
}
func (m *testInstanceState) GetStatus(ctx context.Context, instanceID int64) (domain.InstanceStatus, error) {
	return 0, nil
}
func (m *testInstanceState) SetPlayerCount(ctx context.Context, instanceID int64, count uint32) error {
	return nil
}
func (m *testInstanceState) GetPlayerCount(ctx context.Context, instanceID int64) (uint32, error) {
	return 0, nil
}
func (m *testInstanceState) SetUsage(ctx context.Context, instanceID int64, usage *domain.ResourceUsage) error {
	return nil
}
func (m *testInstanceState) GetUsage(ctx context.Context, instanceID int64) (*domain.ResourceUsage, error) {
	return nil, nil
}
func (m *testInstanceState) Delete(ctx context.Context, instanceID int64) error { return nil }

// ─── Helpers ─────────────────────────────────────────────────────────────────

func addrUint32(v uint32) *uint32 { return &v }
func addrUint64(v uint64) *uint64 { return &v }
