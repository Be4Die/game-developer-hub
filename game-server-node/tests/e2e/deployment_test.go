package e2e

import (
	"context"
	"fmt"
	"strings"
	"testing"

	pb "github.com/Be4Die/game-developer-hub/protos/game_server_node/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// E2E_07: StartInstance — контейнер создаётся и запускается.
func TestE2E_Deployment_StartInstance(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	tc := setupServerContainer(t)

	startResp := startTestInstance(ctx, t, tc, 1, "start-instance-e2e")

	if startResp.InstanceId <= 0 {
		t.Errorf("expected positive instance_id, got %d", startResp.InstanceId)
	}
	if startResp.HostPort == 0 {
		t.Log("host_port is 0 (OS assigned)")
	}

	// Проверяем через Discovery
	resp, err := tc.discovery.GetInstance(ctx, &pb.GetInstanceRequest{
		InstanceId: startResp.InstanceId,
	})
	if err != nil {
		t.Fatalf("GetInstance after start failed: %v", err)
	}

	if resp.Instance.Status != pb.InstanceStatus_INSTANCE_STATUS_RUNNING {
		t.Errorf("expected RUNNING, got %s", resp.Instance.Status)
	}
}

// E2E_08: StopInstance — graceful остановка.
func TestE2E_Deployment_StopInstance(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	tc := setupServerContainer(t)

	startResp := startTestInstance(ctx, t, tc, 1, "stop-test")

	// Останавливаем
	_, err := tc.deployment.StopInstance(ctx, &pb.StopInstanceRequest{
		InstanceId:     startResp.InstanceId,
		TimeoutSeconds: 5,
	})
	if err != nil {
		t.Fatalf("StopInstance failed: %v", err)
	}

	// Проверяем статус
	resp, err := tc.discovery.GetInstance(ctx, &pb.GetInstanceRequest{
		InstanceId: startResp.InstanceId,
	})
	if err != nil {
		t.Fatalf("GetInstance after stop failed: %v", err)
	}

	if resp.Instance.Status != pb.InstanceStatus_INSTANCE_STATUS_STOPPED {
		t.Errorf("expected STOPPED, got %s", resp.Instance.Status)
	}
}

// E2E_09: LoadImage + StartInstance — загрузка образа и запуск.
func TestE2E_Deployment_LoadImageAndStart(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	tc := setupServerContainer(t)

	loadTestImage(ctx, t, tc, 42)

	startResp := startTestInstance(ctx, t, tc, 42, "load-image-test")

	t.Logf("Started instance %d for game 42", startResp.InstanceId)
}

// E2E_12: Стратегии портов (Exact / Range / Any).
func TestE2E_PortStrategies(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	tc := setupServerContainer(t)

	loadTestImage(ctx, t, tc, 1)

	tests := []struct {
		name      string
		portAlloc *pb.PortAllocation
		expected  func(uint32) bool
	}{
		{
			name: "Exact",
			portAlloc: &pb.PortAllocation{
				Strategy: &pb.PortAllocation_Exact{Exact: 27015},
			},
			expected: func(p uint32) bool { return p == 27015 },
		},
		{
			name: "Range",
			portAlloc: &pb.PortAllocation{
				Strategy: &pb.PortAllocation_Range{
					Range: &pb.PortRange{MinPort: 27000, MaxPort: 27100},
				},
			},
			expected: func(p uint32) bool { return p >= 27000 && p <= 27100 },
		},
		{
			name: "Any",
			portAlloc: &pb.PortAllocation{
				Strategy: &pb.PortAllocation_Any{Any: true},
			},
			expected: func(p uint32) bool { return true },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := tc.deployment.StartInstance(ctx, &pb.StartInstanceRequest{
				GameId:         1,
				Name:           fmt.Sprintf("port-%s", strings.ToLower(tt.name)),
				Protocol:       pb.Protocol_PROTOCOL_TCP,
				InternalPort:   8080,
				PortAllocation: tt.portAlloc,
				MaxPlayers:     10,
				Args:           []string{"sleep", "30"},
			})
			if err != nil {
				t.Fatalf("StartInstance(%s) failed: %v", tt.name, err)
			}
			defer func() {
				_, _ = tc.deployment.StopInstance(ctx, &pb.StopInstanceRequest{
					InstanceId:     resp.InstanceId,
					TimeoutSeconds: 3,
				})
			}()

			if !tt.expected(resp.HostPort) {
				t.Errorf("%s: unexpected host port %d", tt.name, resp.HostPort)
			}
			t.Logf("%s: host_port=%d", tt.name, resp.HostPort)
		})
	}
}

// E2E_13: Остановка несуществующего инстанса.
func TestE2E_StopNonExistentInstance(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	tc := setupServerContainer(t)

	_, err := tc.deployment.StopInstance(ctx, &pb.StopInstanceRequest{
		InstanceId:     999999,
		TimeoutSeconds: 5,
	})
	if err == nil {
		t.Fatal("expected error when stopping non-existent instance")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}
	if st.Code() != codes.NotFound {
		t.Errorf("expected code NotFound, got %s", st.Code())
	}

	t.Logf("Got expected NotFound error: %v", err)
}

// E2E_14: Ресурсные лимиты (cpu_millis, memory_bytes).
func TestE2E_ResourceLimits(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	tc := setupServerContainer(t)

	loadTestImage(ctx, t, tc, 1)

	cpuMillis := uint32(500)                // 0.5 ядра
	memoryBytes := uint64(64 * 1024 * 1024) // 64 MB

	resp, err := tc.deployment.StartInstance(ctx, &pb.StartInstanceRequest{
		GameId:       1,
		Name:         "resource-limits-test",
		Protocol:     pb.Protocol_PROTOCOL_TCP,
		InternalPort: 8080,
		PortAllocation: &pb.PortAllocation{
			Strategy: &pb.PortAllocation_Any{Any: true},
		},
		MaxPlayers: 10,
		Args:       []string{"sleep", "30"},
		ResourceLimits: &pb.ResourceLimits{
			CpuMillis:   &cpuMillis,
			MemoryBytes: &memoryBytes,
		},
	})
	if err != nil {
		t.Fatalf("StartInstance with resource limits failed: %v", err)
	}
	defer func() {
		_, _ = tc.deployment.StopInstance(ctx, &pb.StopInstanceRequest{
			InstanceId:     resp.InstanceId,
			TimeoutSeconds: 3,
		})
	}()

	// Проверяем что инстанс запущен
	instResp, err := tc.discovery.GetInstance(ctx, &pb.GetInstanceRequest{
		InstanceId: resp.InstanceId,
	})
	if err != nil {
		t.Fatalf("GetInstance failed: %v", err)
	}

	if instResp.Instance.Status != pb.InstanceStatus_INSTANCE_STATUS_RUNNING {
		t.Errorf("expected RUNNING, got %s", instResp.Instance.Status)
	}

	t.Logf("Instance with limits: id=%d, port=%d, status=%s",
		resp.InstanceId, resp.HostPort, instResp.Instance.Status)
}
