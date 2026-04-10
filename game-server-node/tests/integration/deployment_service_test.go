package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/game-server-node/internal/domain"
	"github.com/Be4Die/game-developer-hub/game-server-node/internal/service"
)

// TestDeploymentService_FullLifecycle проверяет полный цикл: LoadImage → StartInstance → StopInstance.
func TestDeploymentService_FullLifecycle(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	if err := env.deploymentSvc.LoadImage(ctx, 1, imageTag, nil); err != nil {
		t.Fatalf("LoadImage failed: %v", err)
	}

	instanceID, hostPort, err := env.deploymentSvc.StartInstance(ctx, service.StartInstanceOpts{
		GameID:       1,
		Name:         "test-game-server",
		Protocol:     domain.ProtocolTCP,
		InternalPort: 8080,
		PortStrategy: domain.PortStrategy{Any: true},
		MaxPlayers:   10,
		EnvVars:      map[string]string{"GAME_MODE": "test"},
		Args:         []string{"sleep", "60"},
	})
	if err != nil {
		t.Fatalf("StartInstance failed: %v", err)
	}

	if instanceID <= 0 {
		t.Errorf("expected positive instance ID, got %d", instanceID)
	}
	t.Logf("started instance %d on port %d", instanceID, hostPort)

	instance, err := env.storage.GetInstanceByID(ctx, instanceID)
	if err != nil {
		t.Fatalf("GetInstanceByID failed: %v", err)
	}

	if instance.Status != domain.InstanceStatusRunning {
		t.Errorf("expected status Running (%d), got %d", domain.InstanceStatusRunning, instance.Status)
	}
	if instance.GameID != 1 {
		t.Errorf("expected game ID 1, got %d", instance.GameID)
	}
	if instance.Name != "test-game-server" {
		t.Errorf("expected name 'test-game-server', got %s", instance.Name)
	}

	err = env.deploymentSvc.StopInstance(ctx, instanceID, 5*time.Second)
	if err != nil {
		t.Fatalf("StopInstance failed: %v", err)
	}

	instance, err = env.storage.GetInstanceByID(ctx, instanceID)
	if err != nil {
		t.Fatalf("GetInstanceByID after stop failed: %v", err)
	}

	if instance.Status != domain.InstanceStatusStopped {
		t.Errorf("expected status Stopped (%d), got %d", domain.InstanceStatusStopped, instance.Status)
	}
}

// TestDeploymentService_MultipleInstances проверяет запуск нескольких инстансов.
func TestDeploymentService_MultipleInstances(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	if err := env.deploymentSvc.LoadImage(ctx, 1, imageTag, nil); err != nil {
		t.Fatalf("LoadImage failed: %v", err)
	}

	var instanceIDs []int64
	for i := 0; i < 3; i++ {
		id, port, err := env.deploymentSvc.StartInstance(ctx, service.StartInstanceOpts{
			GameID:       1,
			Name:         fmt.Sprintf("game-server-%d", i),
			Protocol:     domain.ProtocolTCP,
			InternalPort: 8080,
			PortStrategy: domain.PortStrategy{Any: true},
			MaxPlayers:   10,
			Args:         []string{"sleep", "60"},
		})
		if err != nil {
			for _, prevID := range instanceIDs {
				_ = env.deploymentSvc.StopInstance(ctx, prevID, 3*time.Second)
			}
			t.Fatalf("StartInstance %d failed: %v", i, err)
		}
		instanceIDs = append(instanceIDs, id)
		t.Logf("instance %d started on port %d", id, port)
	}

	all, err := env.storage.GetAllInstances(ctx)
	if err != nil {
		t.Fatalf("GetAllInstances failed: %v", err)
	}

	if len(all) != 3 {
		t.Errorf("expected 3 instances, got %d", len(all))
	}

	for _, id := range instanceIDs {
		err := env.deploymentSvc.StopInstance(ctx, id, 3*time.Second)
		if err != nil {
			t.Logf("warning: StopInstance %d failed: %v", id, err)
		}
	}
}

// TestDeploymentService_ExactPortStrategy проверяет стратегию Exact.
func TestDeploymentService_ExactPortStrategy(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	if err := env.deploymentSvc.LoadImage(ctx, 1, imageTag, nil); err != nil {
		t.Fatalf("LoadImage failed: %v", err)
	}

	instanceID, hostPort, err := env.deploymentSvc.StartInstance(ctx, service.StartInstanceOpts{
		GameID:       1,
		Name:         "exact-port-test",
		Protocol:     domain.ProtocolTCP,
		InternalPort: 8080,
		PortStrategy: domain.PortStrategy{Exact: 27015},
		MaxPlayers:   10,
		Args:         []string{"sleep", "30"},
	})
	if err != nil {
		t.Fatalf("StartInstance failed: %v", err)
	}
	defer func() {
		_ = env.deploymentSvc.StopInstance(ctx, instanceID, 3*time.Second)
	}()

	if hostPort != 27015 {
		t.Errorf("expected host port 27015, got %d", hostPort)
	}
}

// TestDeploymentService_RangePortStrategy проверяет стратегию Range.
func TestDeploymentService_RangePortStrategy(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	if err := env.deploymentSvc.LoadImage(ctx, 1, imageTag, nil); err != nil {
		t.Fatalf("LoadImage failed: %v", err)
	}

	instanceID, hostPort, err := env.deploymentSvc.StartInstance(ctx, service.StartInstanceOpts{
		GameID:       1,
		Name:         "range-port-test",
		Protocol:     domain.ProtocolTCP,
		InternalPort: 8080,
		PortStrategy: domain.PortStrategy{Range: &domain.PortRange{Min: 27000, Max: 27100}},
		MaxPlayers:   10,
		Args:         []string{"sleep", "30"},
	})
	if err != nil {
		t.Fatalf("StartInstance failed: %v", err)
	}
	defer func() {
		_ = env.deploymentSvc.StopInstance(ctx, instanceID, 3*time.Second)
	}()

	if hostPort < 27000 || hostPort > 27100 {
		t.Errorf("expected host port in range [27000, 27100], got %d", hostPort)
	}
}

// TestDeploymentService_StopNonExistentInstance проверяет остановку несуществующего инстанса.
func TestDeploymentService_StopNonExistentInstance(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	err := env.deploymentSvc.StopInstance(ctx, 99999, 5*time.Second)
	if err == nil {
		t.Fatal("expected error when stopping non-existent instance")
	}

	t.Logf("got expected error: %v", err)
}
