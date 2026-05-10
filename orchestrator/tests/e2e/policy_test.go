//go:build e2e

package e2e

import (
	"testing"

	pb "github.com/Be4Die/game-developer-hub/protos/orchestrator/v1"
)

func TestE2E_Policy_CRUD(t *testing.T) {
	env := setupE2E(t)
	defer env.cleanupTables(t)
	ctx := withJWT(t.Context(), e2eJWTSecret, e2eIssuer)

	gameID := int64(42)

	// Set policy
	setResp, err := env.policyClient.Set(ctx, &pb.GamePolicyServiceSetRequest{
		GameId:                gameID,
		Mode:                  pb.OrchestrationMode_ORCHESTRATION_MODE_KEEP_ALIVE,
		TargetInstances:       3,
		AutoRestart:           true,
		ScaleToZeroTimeout:    10,
		DefaultBuildVersion:   "latest",
		MaxPlayersPerInstance: 100,
		MaxInstancesPerGame:   5,
		ScaleBehavior:         pb.ScaleBehavior_SCALE_BEHAVIOR_SPAWN,
		NodePreference:        "auto",
	})
	if err != nil {
		t.Fatalf("set policy: %v", err)
	}
	if setResp.Policy == nil {
		t.Fatal("expected policy in set response")
	}
	if setResp.Policy.TargetInstances != 3 {
		t.Errorf("expected target_instances=3, got %d", setResp.Policy.TargetInstances)
	}

	// Get policy
	getResp, err := env.policyClient.Get(ctx, &pb.GamePolicyServiceGetRequest{GameId: gameID})
	if err != nil {
		t.Fatalf("get policy: %v", err)
	}
	if getResp.Policy == nil {
		t.Fatal("expected policy in get response")
	}
	if getResp.Policy.Mode != pb.OrchestrationMode_ORCHESTRATION_MODE_KEEP_ALIVE {
		t.Errorf("expected keep_alive mode, got %v", getResp.Policy.Mode)
	}
	if getResp.Policy.NodePreference != "auto" {
		t.Errorf("expected node_preference=auto, got %s", getResp.Policy.NodePreference)
	}
}

func TestE2E_Policy_DefaultDisabled(t *testing.T) {
	env := setupE2E(t)
	defer env.cleanupTables(t)
	ctx := withJWT(t.Context(), e2eJWTSecret, e2eIssuer)

	gameID := int64(99)

	// Get policy for game without saved policy — should return default (disabled).
	getResp, err := env.policyClient.Get(ctx, &pb.GamePolicyServiceGetRequest{GameId: gameID})
	if err != nil {
		t.Fatalf("get policy: %v", err)
	}
	if getResp.Policy == nil {
		t.Fatal("expected default policy in get response")
	}
	if getResp.Policy.Mode != pb.OrchestrationMode_ORCHESTRATION_MODE_DISABLED {
		t.Errorf("expected default disabled mode, got %v", getResp.Policy.Mode)
	}
}
