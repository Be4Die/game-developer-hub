package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHealthHandler_Check(t *testing.T) {
	handler := NewHealthHandler("v1.2.3")
	// Подождем чтобы uptime был > 0.
	time.Sleep(10 * time.Millisecond)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	handler.Check(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %q, want %q", rec.Header().Get("Content-Type"), "application/json")
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	if body["status"] != "ok" {
		t.Errorf("body[status] = %q, want %q", body["status"], "ok")
	}
	if body["version"] != "v1.2.3" {
		t.Errorf("body[version] = %q, want %q", body["version"], "v1.2.3")
	}
	uptime, ok := body["uptime_seconds"].(float64)
	if !ok {
		t.Fatal("uptime_seconds is not a number")
	}
	if uptime <= 0 {
		t.Errorf("uptime_seconds = %f, want > 0", uptime)
	}
}
