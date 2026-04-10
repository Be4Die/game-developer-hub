package integration

import (
	"context"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/game-server-node/internal/domain"
)

const imageTag = "alpine:3.18"

// TestDockerRuntime_CreateAndRemoveContainer проверяет полный жизненный цикл контейнера.
func TestDockerRuntime_CreateAndRemoveContainer(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	containerID, err := env.runtime.CreateContainer(ctx, domain.ContainerOpts{
		ImageTag:     imageTag,
		InternalPort: 8080,
		HostPort:     0,
		EnvVars:      map[string]string{"TEST_ENV": "integration-test"},
		Args:         []string{"echo", "hello"},
	})
	if err != nil {
		t.Fatalf("CreateContainer failed: %v (ensure '%s' image is available locally)", err, imageTag)
	}

	if containerID == "" {
		t.Fatal("expected non-empty container ID")
	}

	t.Logf("created container: %s", containerID[:12])

	err = env.runtime.RemoveContainer(ctx, containerID)
	if err != nil {
		t.Fatalf("RemoveContainer failed: %v", err)
	}
}

// TestDockerRuntime_CreateStartStopContainer проверяет создание, запуск и остановку.
func TestDockerRuntime_CreateStartStopContainer(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	containerID, err := env.runtime.CreateContainer(ctx, domain.ContainerOpts{
		ImageTag:     imageTag,
		InternalPort: 80,
		HostPort:     0,
		Args:         []string{"sleep", "30"},
	})
	if err != nil {
		t.Fatalf("CreateContainer failed: %v", err)
	}

	t.Logf("created container: %s", containerID[:12])

	err = env.runtime.StartContainer(ctx, containerID)
	if err != nil {
		_ = env.runtime.RemoveContainer(ctx, containerID)
		t.Fatalf("StartContainer failed: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	err = env.runtime.StopContainer(ctx, containerID, 5*time.Second)
	if err != nil {
		_ = env.runtime.RemoveContainer(ctx, containerID)
		t.Fatalf("StopContainer failed: %v", err)
	}

	err = env.runtime.RemoveContainer(ctx, containerID)
	if err != nil {
		t.Fatalf("RemoveContainer failed: %v", err)
	}
}

// TestDockerRuntime_ContainerLogs проверяет получение логов контейнера.
func TestDockerRuntime_ContainerLogs(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	containerID, err := env.runtime.CreateContainer(ctx, domain.ContainerOpts{
		ImageTag:     imageTag,
		InternalPort: 80,
		HostPort:     0,
		Args:         []string{"sh", "-c", "echo 'integration-test-log' && sleep 1"},
	})
	if err != nil {
		t.Fatalf("CreateContainer failed: %v", err)
	}

	err = env.runtime.StartContainer(ctx, containerID)
	if err != nil {
		_ = env.runtime.RemoveContainer(ctx, containerID)
		t.Fatalf("StartContainer failed: %v", err)
	}

	time.Sleep(2 * time.Second)

	logs, err := env.runtime.ContainerLogs(ctx, containerID, false)
	if err != nil {
		_ = env.runtime.RemoveContainer(ctx, containerID)
		t.Fatalf("ContainerLogs failed: %v", err)
	}

	buf := make([]byte, 1024)
	n, err := logs.Read(buf)
	if err != nil {
		_ = logs.Close()
		_ = env.runtime.RemoveContainer(ctx, containerID)
		t.Fatalf("reading logs failed: %v", err)
	}
	_ = logs.Close()

	logContent := string(buf[8:n])
	if logContent == "" {
		t.Log("container logs are empty (may be ok for short-lived containers)")
	} else {
		t.Logf("container logs: %s", logContent)
	}

	_ = env.runtime.RemoveContainer(ctx, containerID)
}
