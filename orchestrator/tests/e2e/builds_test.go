//go:build e2e

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
)

// E2E_09: List builds — empty.
func TestE2E_Builds_List_Empty(t *testing.T) {
	env := setupE2E(t)
	env.cleanupDB(t)

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

	t.Log("Builds list: empty as expected")
}

// E2E_10: List builds — with data.
func TestE2E_Builds_List(t *testing.T) {
	env := setupE2E(t)
	env.cleanupDB(t)
	ctx := context.Background()

	// Создаём 3 билда.
	for i := range 3 {
		build := &domain.ServerBuild{
			ID:           int64(i + 1),
			GameID:       42,
			Version:      fmt.Sprintf("v1.0.%d", i),
			ImageTag:     fmt.Sprintf("welwise/game-42:v1.0.%d", i),
			Protocol:     domain.ProtocolTCP,
			InternalPort: 8080,
			MaxPlayers:   10,
			FileURL:      fmt.Sprintf("/builds/v1.0.%d.tar", i),
			FileSize:     int64(1000000 + i*1000),
			CreatedAt:    time.Now(),
		}
		if err := env.buildStorage.Create(ctx, build); err != nil {
			t.Fatalf("Create build %d failed: %v", i, err)
		}
	}

	resp, err := http.Get(env.baseURL + "/games/42/builds")
	if err != nil {
		t.Fatalf("GET /games/42/builds failed: %v", err)
	}
	defer resp.Body.Close()

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	builds := body["builds"].([]any)
	if len(builds) != 3 {
		t.Fatalf("expected 3 builds, got %d", len(builds))
	}

	t.Logf("Listed %d builds for game 42", len(builds))
}

// E2E_11: Get build by version.
func TestE2E_Builds_Get(t *testing.T) {
	env := setupE2E(t)
	env.cleanupDB(t)
	ctx := context.Background()

	build := &domain.ServerBuild{
		ID:           10,
		GameID:       42,
		Version:      "v2.0.0",
		ImageTag:     "welwise/game-42:v2.0.0",
		Protocol:     domain.ProtocolUDP,
		InternalPort: 7777,
		MaxPlayers:   32,
		FileURL:      "/builds/v2.0.0.tar",
		FileSize:     2000000,
		CreatedAt:    time.Now(),
	}
	if err := env.buildStorage.Create(ctx, build); err != nil {
		t.Fatalf("Create build failed: %v", err)
	}

	resp, err := http.Get(env.baseURL + "/games/42/builds/v2.0.0")
	if err != nil {
		t.Fatalf("GET /games/42/builds/v2.0.0 failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if body["build_version"] != "v2.0.0" {
		t.Errorf("build_version = %q, want %q", body["build_version"], "v2.0.0")
	}
	if body["protocol"] != "udp" {
		t.Errorf("protocol = %q, want %q", body["protocol"], "udp")
	}
	if int64(body["max_players"].(float64)) != 32 {
		t.Errorf("max_players = %v, want 32", body["max_players"])
	}

	t.Logf("Got build: version=%s, protocol=%s, max_players=%v",
		body["build_version"], body["protocol"], body["max_players"])
}

// E2E_12: Get build — not found.
func TestE2E_Builds_Get_NotFound(t *testing.T) {
	env := setupE2E(t)
	env.cleanupDB(t)

	resp, err := http.Get(env.baseURL + "/games/42/builds/nonexistent")
	if err != nil {
		t.Fatalf("GET /games/42/builds/nonexistent failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

// E2E_13: Delete build.
func TestE2E_Builds_Delete(t *testing.T) {
	env := setupE2E(t)
	env.cleanupDB(t)
	ctx := context.Background()

	build := &domain.ServerBuild{
		ID:           100,
		GameID:       42,
		Version:      "v1.0.0",
		ImageTag:     "welwise/game-42:v1.0.0",
		Protocol:     domain.ProtocolTCP,
		InternalPort: 8080,
		MaxPlayers:   10,
		FileURL:      "/builds/delete-me.tar",
		FileSize:     1000000,
		CreatedAt:    time.Now(),
	}
	if err := env.buildStorage.Create(ctx, build); err != nil {
		t.Fatalf("Create build failed: %v", err)
	}

	req, _ := http.NewRequest(http.MethodDelete, env.baseURL+"/games/42/builds/v1.0.0", nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("DELETE /games/42/builds/v1.0.0 failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}

	// Verify deleted.
	_, err = env.buildStorage.GetByID(ctx, build.ID)
	if err != domain.ErrNotFound {
		t.Errorf("expected build to be deleted, got %v", err)
	}

	t.Log("Build deleted successfully")
}

// E2E_14: Delete build — not found.
func TestE2E_Builds_Delete_NotFound(t *testing.T) {
	env := setupE2E(t)
	env.cleanupDB(t)

	req, _ := http.NewRequest(http.MethodDelete, env.baseURL+"/games/42/builds/nonexistent", nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("DELETE failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}
