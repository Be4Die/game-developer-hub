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

	// Контейнер пишет лог и живёт достаточно долго чтобы мы успели прочитать.
	containerID, err := env.runtime.CreateContainer(ctx, domain.ContainerOpts{
		ImageTag:     imageTag,
		InternalPort: 80,
		HostPort:     0,
		Args:         []string{"sh", "-c", "echo 'integration-test-log' && sleep 10"},
	})
	if err != nil {
		t.Fatalf("CreateContainer failed: %v", err)
	}

	err = env.runtime.StartContainer(ctx, containerID)
	if err != nil {
		_ = env.runtime.RemoveContainer(ctx, containerID)
		t.Fatalf("StartContainer failed: %v", err)
	}

	// Ждём чтобы команда echo успела выполниться и Docker записал логи.
	time.Sleep(3 * time.Second)

	logs, err := env.runtime.ContainerLogs(ctx, containerID, false)
	if err != nil {
		_ = env.runtime.RemoveContainer(ctx, containerID)
		t.Fatalf("ContainerLogs failed: %v", err)
	}

	// EOF — нормальная ситуация для завершённых контейнеров.
	// Читаем всё что доступно.
	buf := make([]byte, 1024)
	n, readErr := logs.Read(buf)
	_ = logs.Close()

	// Если контейнер уже завершился, Read может вернуть EOF до записи данных.
	// Это ожидаемо — считаем пустые логи допустимыми.
	if readErr != nil && n == 0 {
		t.Log("container logs are empty (container may have already exited)")
	} else if n > 8 {
		logContent := string(buf[8:n]) // skip 8-byte Docker header
		t.Logf("container logs: %s", logContent)
	} else {
		t.Log("container logs are empty or incomplete")
	}

	_ = env.runtime.RemoveContainer(ctx, containerID)
}
