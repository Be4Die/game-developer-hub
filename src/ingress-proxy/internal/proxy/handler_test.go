package proxy

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/ingress-proxy/internal/store"
)

type mockResolver struct {
	targets map[int64]string
}

func (m *mockResolver) ResolveInstanceTarget(_ context.Context, instanceID int64) (string, error) {
	if target, ok := m.targets[instanceID]; ok {
		return target, nil
	}
	return "", store.ErrTargetNotFound
}

func (m *mockResolver) Close() error {
	return nil
}

func TestHandler_Healthz(t *testing.T) {
	h := NewHandler(&mockResolver{}, slog.Default(), time.Second)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHandler_InvalidPath(t *testing.T) {
	h := NewHandler(&mockResolver{}, slog.Default(), time.Second)

	req := httptest.NewRequest(http.MethodGet, "/game-proxy/only-one-param", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestHandler_InstanceNotFound(t *testing.T) {
	resolver := &mockResolver{targets: map[int64]string{}}
	h := NewHandler(resolver, slog.Default(), time.Second)

	req := httptest.NewRequest(http.MethodGet, "/game-proxy/42/999/api/status", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestHandler_HTTPReverseProxy(t *testing.T) {
	var receivedPath string
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"hello":"game"}`))
	}))
	defer backend.Close()

	// backend.URL is e.g. "http://127.0.0.1:54321"
	targetAddr := backend.Listener.Addr().String()

	resolver := &mockResolver{
		targets: map[int64]string{
			100: targetAddr,
		},
	}

	h := NewHandler(resolver, slog.Default(), time.Second)

	req := httptest.NewRequest(http.MethodGet, "/game-proxy/42/100/api/v1/players", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if receivedPath != "/api/v1/players" {
		t.Fatalf("expected backend to receive /api/v1/players, got %s", receivedPath)
	}

	body, _ := io.ReadAll(rr.Body)
	if string(body) != `{"hello":"game"}` {
		t.Fatalf("unexpected body: %s", string(body))
	}
}
