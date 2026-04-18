//go:build e2e

package e2e

import (
	"context"
	"testing"
	"time"

	pb "github.com/Be4Die/game-developer-hub/protos/orchestrator/v1"
)

// E2E_01: Health check -- проверка работоспособности оркестратора.
func TestE2E_Health(t *testing.T) {
	env := setupE2E(t)
	env.cleanupTables(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := env.healthClient.Check(withAPIKey(ctx, e2eAPIKey), &pb.HealthCheckRequest{})
	if err != nil {
		t.Fatalf("HealthCheck failed: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("status = %q, want %q", resp.Status, "ok")
	}
	if resp.Version == "" {
		t.Error("version field missing")
	}
	if resp.UptimeSeconds <= 0 {
		t.Errorf("uptime_seconds = %f, want > 0", resp.UptimeSeconds)
	}
}
