package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/service"
	"github.com/go-chi/chi/v5"
)

// ─── Mock for DiscoveryHandler ───────────────────────────────────────────────

type discHandlerMockInstanceRepo struct {
	listByGameFn func(ctx context.Context, gameID int64, status *domain.InstanceStatus) ([]*domain.Instance, error)
}

func (m *discHandlerMockInstanceRepo) Create(ctx context.Context, instance *domain.Instance) error {
	return nil
}
func (m *discHandlerMockInstanceRepo) GetByID(ctx context.Context, id int64) (*domain.Instance, error) {
	return nil, nil
}
func (m *discHandlerMockInstanceRepo) ListByGame(ctx context.Context, gameID int64, status *domain.InstanceStatus) ([]*domain.Instance, error) {
	return m.listByGameFn(ctx, gameID, status)
}
func (m *discHandlerMockInstanceRepo) ListByNode(ctx context.Context, nodeID int64) ([]*domain.Instance, error) {
	return nil, nil
}
func (m *discHandlerMockInstanceRepo) Update(ctx context.Context, instance *domain.Instance) error {
	return nil
}
func (m *discHandlerMockInstanceRepo) Delete(ctx context.Context, id int64) error { return nil }
func (m *discHandlerMockInstanceRepo) CountByGame(ctx context.Context, gameID int64) (int, error) {
	return 0, nil
}

type discHandlerMockInstanceState struct{}

func (m *discHandlerMockInstanceState) SetStatus(ctx context.Context, instanceID int64, status domain.InstanceStatus) error {
	return nil
}
func (m *discHandlerMockInstanceState) GetStatus(ctx context.Context, instanceID int64) (domain.InstanceStatus, error) {
	return 0, nil
}
func (m *discHandlerMockInstanceState) SetPlayerCount(ctx context.Context, instanceID int64, count uint32) error {
	return nil
}
func (m *discHandlerMockInstanceState) GetPlayerCount(ctx context.Context, instanceID int64) (uint32, error) {
	return 0, nil
}
func (m *discHandlerMockInstanceState) SetUsage(ctx context.Context, instanceID int64, usage *domain.ResourceUsage) error {
	return nil
}
func (m *discHandlerMockInstanceState) GetUsage(ctx context.Context, instanceID int64) (*domain.ResourceUsage, error) {
	return nil, nil
}
func (m *discHandlerMockInstanceState) Delete(ctx context.Context, instanceID int64) error {
	return nil
}

type discHandlerMockNodeRepo struct {
	getByIDFn func(ctx context.Context, id int64) (*domain.Node, error)
}

func (m *discHandlerMockNodeRepo) Create(ctx context.Context, node *domain.Node) error { return nil }
func (m *discHandlerMockNodeRepo) Update(ctx context.Context, node *domain.Node) error { return nil }
func (m *discHandlerMockNodeRepo) GetByID(ctx context.Context, id int64) (*domain.Node, error) {
	return m.getByIDFn(ctx, id)
}
func (m *discHandlerMockNodeRepo) GetByAddress(ctx context.Context, addr string) (*domain.Node, error) {
	return nil, nil
}
func (m *discHandlerMockNodeRepo) List(ctx context.Context, status *domain.NodeStatus) ([]*domain.Node, error) {
	return nil, nil
}
func (m *discHandlerMockNodeRepo) Delete(ctx context.Context, id int64) error         { return nil }
func (m *discHandlerMockNodeRepo) UpdateLastPing(ctx context.Context, id int64) error { return nil }

// ─── Tests ───────────────────────────────────────────────────────────────────

func TestDiscoveryHandler_Discover_Success(t *testing.T) {
	instanceRepo := &discHandlerMockInstanceRepo{
		listByGameFn: func(ctx context.Context, gameID int64, status *domain.InstanceStatus) ([]*domain.Instance, error) {
			return []*domain.Instance{
				{ID: 1, NodeID: 10, HostPort: 7001, Protocol: domain.ProtocolTCP, MaxPlayers: 10},
			}, nil
		},
	}

	nodeRepo := &discHandlerMockNodeRepo{
		getByIDFn: func(ctx context.Context, id int64) (*domain.Node, error) {
			return &domain.Node{ID: id, Address: "node1:44044"}, nil
		},
	}

	svc := service.NewDiscoveryService(instanceRepo, &discHandlerMockInstanceState{}, nodeRepo)
	handler := NewDiscoveryHandler(svc)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/games/42/servers", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("gameId", "42")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.Discover(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	servers, ok := body["servers"].([]any)
	if !ok {
		t.Fatal("expected 'servers' array in response")
	}
	if len(servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(servers))
	}
}

func TestDiscoveryHandler_Discover_InvalidGameId(t *testing.T) {
	svc := service.NewDiscoveryService(&discHandlerMockInstanceRepo{}, &discHandlerMockInstanceState{}, &discHandlerMockNodeRepo{})
	handler := NewDiscoveryHandler(svc)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/games/abc/servers", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("gameId", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.Discover(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var body map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["code"] != "BAD_REQUEST" {
		t.Errorf("code = %q, want %q", body["code"], "BAD_REQUEST")
	}
}

func TestDiscoveryHandler_Discover_ServiceError(t *testing.T) {
	instanceRepo := &discHandlerMockInstanceRepo{
		listByGameFn: func(ctx context.Context, gameID int64, status *domain.InstanceStatus) ([]*domain.Instance, error) {
			return nil, context.Canceled
		},
	}

	svc := service.NewDiscoveryService(instanceRepo, &discHandlerMockInstanceState{}, &discHandlerMockNodeRepo{})
	handler := NewDiscoveryHandler(svc)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/games/42/servers", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("gameId", "42")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.Discover(rec, req)

	// Service error -> 500 INTERNAL_ERROR
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	var body map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["code"] != "INTERNAL_ERROR" {
		t.Errorf("code = %q, want %q", body["code"], "INTERNAL_ERROR")
	}
}

func TestDiscoveryHandler_Discover_EmptyList(t *testing.T) {
	instanceRepo := &discHandlerMockInstanceRepo{
		listByGameFn: func(ctx context.Context, gameID int64, status *domain.InstanceStatus) ([]*domain.Instance, error) {
			return []*domain.Instance{}, nil
		},
	}

	svc := service.NewDiscoveryService(instanceRepo, &discHandlerMockInstanceState{}, &discHandlerMockNodeRepo{})
	handler := NewDiscoveryHandler(svc)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/games/42/servers", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("gameId", "42")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.Discover(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	servers := body["servers"].([]any)
	if len(servers) != 0 {
		t.Errorf("expected 0 servers, got %d", len(servers))
	}
}
