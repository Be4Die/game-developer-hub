//go:build e2e

package e2e

import (
	"encoding/json"
	"net/http"
	"testing"
)

// E2E_01: Health check — проверка работоспособности.
func TestE2E_Health(t *testing.T) {
	env := setupE2E(t)
	env.cleanupDB(t)

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
	if body["version"] != "e2e-test-1.0.0" {
		t.Errorf("version = %q, want %q", body["version"], "e2e-test-1.0.0")
	}

	uptime, ok := body["uptime_seconds"].(float64)
	if !ok {
		t.Fatal("uptime_seconds is not a number")
	}
	if uptime <= 0 {
		t.Errorf("uptime_seconds = %f, want > 0", uptime)
	}

	t.Logf("Health: status=ok, version=%s, uptime=%.2fs", body["version"], uptime)
}
