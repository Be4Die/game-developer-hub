//go:build e2e

package e2e

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

// E2E_01: Health check -- проверка работоспособности оркестратора.
func TestE2E_Health(t *testing.T) {
	env := setupE2E(t)
	env.cleanupTables(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

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
	if _, ok := body["version"]; !ok {
		t.Error("version field missing")
	}
	if _, ok := body["uptime_seconds"]; !ok {
		t.Error("uptime_seconds field missing")
	}

	_ = ctx
}
