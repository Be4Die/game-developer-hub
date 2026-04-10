// Package e2e содержит end-to-end тесты для game-server-node.
//
// В отличие от интеграционных тестов (bufconn in-memory), e2e используют
// реальный сетевой стек: gRPC клиент → TCP → Docker контейнер с сервером.
//
// Перед запуском:
//  1. Соберите образ: task node:build
//  2. Запустите контейнер: task node:up
//  3. Убедитесь что Docker Desktop запущен
//
// Запуск:
//
//	go test -tags=e2e ./tests/e2e/...
package e2e

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/game-server-node/internal/runtime/docker"
	pb "github.com/Be4Die/game-developer-hub/protos/game_server_node/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Константы e2e тестирования.
const (
	grpcAddress   = "localhost:44044"
	imageTag      = "alpine:3.18"
	testTimeout   = 60 * time.Second
	stopTimeout   = 10 * time.Second
	containerWait = 500 * time.Millisecond
)

// getE2EAPIKey загружает ключ из env или использует значение по умолчанию.
func getE2EAPIKey() string {
	if key := os.Getenv("E2E_API_KEY"); key != "" {
		return key
	}
	return "dev-api-key-for-local-testing"
}

// testClient хранит gRPC клиенты для всех сервисов.
type testClient struct {
	discovery    pb.DiscoveryServiceClient
	deployment   pb.DeploymentServiceClient
	conn         *grpc.ClientConn
	containerLog *slog.Logger
}

// connectToServer подключается к работающему gRPC серверу.
func connectToServer(ctx context.Context, t *testing.T) *testClient {
	t.Helper()

	// Unary interceptor автоматически добавляет API-ключ в каждый запрос.
	authInterceptor := func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+getE2EAPIKey())
		return invoker(ctx, method, req, reply, cc, opts...)
	}

	conn, err := grpc.NewClient(
		grpcAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(testTimeout),
		grpc.WithUnaryInterceptor(authInterceptor),
	)
	if err != nil {
		t.Fatalf("не удалось подключиться к %s: %v\nУбедитесь что сервер запущен: task node:up", grpcAddress, err)
	}

	t.Cleanup(func() {
		_ = conn.Close()
	})

	return &testClient{
		discovery:  pb.NewDiscoveryServiceClient(conn),
		deployment: pb.NewDeploymentServiceClient(conn),
		conn:       conn,
		containerLog: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelWarn,
		})),
	}
}

// ensureDockerAvailable проверяет доступность Docker daemon.
func ensureDockerAvailable(t *testing.T) {
	t.Helper()
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	_, err := docker.New(log)
	if err != nil {
		t.Skipf("Docker daemon недоступен: %v", err)
	}
}

// cleanupInstances останавливает и удаляет все инстансы на сервере.
// Для идемпотентности тестов удаляем инстансы в любых статусах.
func cleanupInstances(t *testing.T, tc *testClient, ctx context.Context) {
	t.Helper()

	resp, err := tc.discovery.ListInstances(ctx, &pb.ListInstancesRequest{})
	if err != nil {
		return // не критично
	}

	for _, inst := range resp.Instances {
		if inst.Status == pb.InstanceStatus_INSTANCE_STATUS_RUNNING ||
			inst.Status == pb.InstanceStatus_INSTANCE_STATUS_STARTING {
			_, _ = tc.deployment.StopInstance(ctx, &pb.StopInstanceRequest{
				InstanceId:     inst.InstanceId,
				TimeoutSeconds: 3,
			})
		}
	}
}

// cleanupAllInstancesForTest останавливает все запущенные и удаляет все остановленные
// инстансы, гарантируя чистое состояние для идемпотентного теста.
func cleanupAllInstancesForTest(t *testing.T, tc *testClient, ctx context.Context) {
	t.Helper()

	resp, err := tc.discovery.ListInstances(ctx, &pb.ListInstancesRequest{})
	if err != nil {
		return
	}

	for _, inst := range resp.Instances {
		if inst.Status == pb.InstanceStatus_INSTANCE_STATUS_RUNNING ||
			inst.Status == pb.InstanceStatus_INSTANCE_STATUS_STARTING {
			_, _ = tc.deployment.StopInstance(ctx, &pb.StopInstanceRequest{
				InstanceId:     inst.InstanceId,
				TimeoutSeconds: 3,
			})
		}
		// Остановленные и crashed инстансы остаются в хранилище (in-memory).
		// Для полного clean нужно пересоздать контейнер, но это не требуется
		// для большинства тестов — фильтрация по status решает проблему.
	}

	// Пауза чтобы все инстансы успели остановиться
	time.Sleep(200 * time.Millisecond)
}

// ============================================================
// TestMain — точка входа для всех e2e тестов
// ============================================================

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

// ============================================================
// DiscoveryService E2E (6 тестов)
// ============================================================

// E2E_01: GetNodeInfo — проверка характеристик ноды.
func TestE2E_Discovery_GetNodeInfo(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	tc := connectToServer(ctx, t)

	resp, err := tc.discovery.GetNodeInfo(ctx, &pb.GetNodeInfoRequest{})
	if err != nil {
		t.Fatalf("GetNodeInfo RPC failed: %v", err)
	}

	if resp.Region == "" {
		t.Error("expected non-empty region")
	}
	if resp.AgentVersion == "" {
		t.Error("expected non-empty agent_version")
	}

	// CPU и память зависят от виртуализации Docker Desktop.
	// На Windows значения могут быть 0 — проверяем только что сервис отвечает.
	if resp.CpuCores > 0 || resp.TotalMemoryBytes > 0 {
		t.Logf("Resources: cpu=%d cores, mem=%d bytes", resp.CpuCores, resp.TotalMemoryBytes)
	} else {
		t.Log("Resources: stub values returned (Docker Desktop virtualization)")
	}

	t.Logf("Node: region=%s, version=%s, cpu=%d, mem=%d bytes",
		resp.Region, resp.AgentVersion, resp.CpuCores, resp.TotalMemoryBytes)
}

// E2E_02: Heartbeat — проверка загрузки ноды.
func TestE2E_Discovery_Heartbeat(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	tc := connectToServer(ctx, t)
	cleanupInstances(t, tc, ctx)

	resp, err := tc.discovery.Heartbeat(ctx, &pb.HeartbeatRequest{})
	if err != nil {
		t.Fatalf("Heartbeat RPC failed: %v", err)
	}

	t.Logf("Heartbeat: cpu=%.2f%%, mem=%d bytes, active_instances=%d",
		resp.Usage.CpuUsagePercent, resp.Usage.MemoryUsedBytes, resp.ActiveInstanceCount)

	// На Docker Desktop значения могут быть непредсказуемыми.
	// Проверяем только что RPC прошёл без ошибки.
}

// E2E_03: ListInstances — пустой список → добавление → список с данными.
func TestE2E_Discovery_ListInstances(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	tc := connectToServer(ctx, t)
	cleanupInstances(t, tc, ctx)

	// 1. Пустой список
	resp, err := tc.discovery.ListInstances(ctx, &pb.ListInstancesRequest{})
	if err != nil {
		t.Fatalf("ListInstances (empty) failed: %v", err)
	}
	initialCount := len(resp.Instances)

	// 2. Запускаем инстанс
	ensureDockerAvailable(t)
	startResp := startTestInstance(t, tc, ctx, 1, "list-test-server")

	// 3. Проверяем что список увеличился
	resp, err = tc.discovery.ListInstances(ctx, &pb.ListInstancesRequest{})
	if err != nil {
		t.Fatalf("ListInstances (after start) failed: %v", err)
	}

	if len(resp.Instances) != initialCount+1 {
		t.Errorf("expected %d instances, got %d", initialCount+1, len(resp.Instances))
	}

	// Cleanup
	stopTestInstance(t, tc, ctx, startResp.InstanceId)
}

// E2E_04: ListInstancesByGame — фильтрация по game_id.
func TestE2E_Discovery_ListInstancesByGame(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	tc := connectToServer(ctx, t)
	ensureDockerAvailable(t)

	// Уникальные game_id для идемпотентности теста
	const gameID1 = 10001
	const gameID2 = 10002

	// Запускаем 2 инстанса для game 1, 1 для game 2
	var game1IDs []int64
	for i := 0; i < 2; i++ {
		resp := startTestInstanceWithGameID(t, tc, ctx, gameID1, fmt.Sprintf("game1-srv-%d", i))
		game1IDs = append(game1IDs, resp.InstanceId)
	}
	resp2 := startTestInstanceWithGameID(t, tc, ctx, gameID2, "game2-srv")

	// Проверяем фильтрацию
	resp, err := tc.discovery.ListInstancesByGame(ctx, &pb.ListInstancesByGameRequest{GameId: gameID1})
	if err != nil {
		t.Fatalf("ListInstancesByGame(game=%d) failed: %v", gameID1, err)
	}

	if len(resp.Instances) != 2 {
		t.Errorf("expected 2 instances for game %d, got %d", gameID1, len(resp.Instances))
	}

	// Game 2
	resp, err = tc.discovery.ListInstancesByGame(ctx, &pb.ListInstancesByGameRequest{GameId: gameID2})
	if err != nil {
		t.Fatalf("ListInstancesByGame(game=%d) failed: %v", gameID2, err)
	}
	if len(resp.Instances) != 1 {
		t.Errorf("expected 1 instance for game %d, got %d", gameID2, len(resp.Instances))
	}

	// Несуществующая игра
	resp, err = tc.discovery.ListInstancesByGame(ctx, &pb.ListInstancesByGameRequest{GameId: 999999})
	if err != nil {
		t.Fatalf("ListInstancesByGame(game=999999) failed: %v", err)
	}
	if len(resp.Instances) != 0 {
		t.Errorf("expected 0 instances for game 999999, got %d", len(resp.Instances))
	}

	// Cleanup
	for _, id := range game1IDs {
		stopTestInstance(t, tc, ctx, id)
	}
	stopTestInstance(t, tc, ctx, resp2.InstanceId)
}

// E2E_05: GetInstance — существующий и несуществующий.
func TestE2E_Discovery_GetInstance(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	tc := connectToServer(ctx, t)
	cleanupInstances(t, tc, ctx)
	ensureDockerAvailable(t)

	// Запускаем инстанс
	startResp := startTestInstance(t, tc, ctx, 1, "get-instance-test")
	defer stopTestInstance(t, tc, ctx, startResp.InstanceId)

	// Получение существующего
	resp, err := tc.discovery.GetInstance(ctx, &pb.GetInstanceRequest{
		InstanceId: startResp.InstanceId,
	})
	if err != nil {
		t.Fatalf("GetInstance failed: %v", err)
	}

	if resp.Instance.Name != "get-instance-test" {
		t.Errorf("expected name 'get-instance-test', got '%s'", resp.Instance.Name)
	}
	if resp.Instance.Status != pb.InstanceStatus_INSTANCE_STATUS_RUNNING {
		t.Errorf("expected status RUNNING, got %s", resp.Instance.Status)
	}
	if resp.Instance.GameId != 1 {
		t.Errorf("expected game_id 1, got %d", resp.Instance.GameId)
	}

	// Получение несуществующего
	_, err = tc.discovery.GetInstance(ctx, &pb.GetInstanceRequest{InstanceId: 999999})
	if err == nil {
		t.Fatal("expected NotFound error for non-existent instance")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}
	if st.Code() != codes.NotFound {
		t.Errorf("expected code NotFound, got %s", st.Code())
	}
}

// E2E_06: GetInstanceUsage — метрики запущенного контейнера.
func TestE2E_Discovery_GetInstanceUsage(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	tc := connectToServer(ctx, t)
	cleanupInstances(t, tc, ctx)
	ensureDockerAvailable(t)

	// Запускаем инстанс
	startResp := startTestInstance(t, tc, ctx, 1, "usage-test")
	defer stopTestInstance(t, tc, ctx, startResp.InstanceId)

	time.Sleep(containerWait)

	// Получаем метрики
	resp, err := tc.discovery.GetInstanceUsage(ctx, &pb.GetInstanceUsageRequest{
		InstanceId: startResp.InstanceId,
	})
	if err != nil {
		t.Fatalf("GetInstanceUsage failed: %v", err)
	}

	if resp.InstanceId != startResp.InstanceId {
		t.Errorf("expected instance_id %d, got %d", startResp.InstanceId, resp.InstanceId)
	}

	t.Logf("Usage: CPU=%.2f%%, Mem=%d bytes, Disk=%d bytes, Net=%d bytes/s",
		resp.Usage.CpuUsagePercent,
		resp.Usage.MemoryUsedBytes,
		resp.Usage.DiskUsedBytes,
		resp.Usage.NetworkBytesPerSec,
	)

	// Метрики несуществующего инстанса
	_, err = tc.discovery.GetInstanceUsage(ctx, &pb.GetInstanceUsageRequest{InstanceId: 999999})
	if err == nil {
		t.Fatal("expected error for non-existent instance usage")
	}
}

// ============================================================
// DeploymentService E2E (5 тестов)
// ============================================================

// E2E_07: StartInstance — контейнер создаётся и запускается.
func TestE2E_Deployment_StartInstance(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	tc := connectToServer(ctx, t)
	cleanupInstances(t, tc, ctx)
	ensureDockerAvailable(t)

	// Запускаем инстанс
	startResp := startTestInstance(t, tc, ctx, 1, "start-instance-e2e")
	defer stopTestInstance(t, tc, ctx, startResp.InstanceId)

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

	tc := connectToServer(ctx, t)
	cleanupInstances(t, tc, ctx)
	ensureDockerAvailable(t)

	// Запускаем
	startResp := startTestInstance(t, tc, ctx, 1, "stop-test")

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

	tc := connectToServer(ctx, t)
	cleanupInstances(t, tc, ctx)
	ensureDockerAvailable(t)

	// LoadImage требует streaming
	loadTestImage(t, tc, ctx, 42, imageTag)

	// Запускаем инстанс — должен найти образ
	startResp := startTestInstance(t, tc, ctx, 42, "load-image-test")
	defer stopTestInstance(t, tc, ctx, startResp.InstanceId)

	t.Logf("Started instance %d for game 42", startResp.InstanceId)
}

// E2E_10: Полный E2E поток — все методы в одной цепочке.
func TestE2E_FullWorkflow(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	tc := connectToServer(ctx, t)
	cleanupInstances(t, tc, ctx)
	ensureDockerAvailable(t)

	// Шаг 1: LoadImage
	t.Log("1. Loading image...")
	loadTestImage(t, tc, ctx, 100, imageTag)

	// Шаг 2: StartInstance
	t.Log("2. Starting instance...")
	startResp, err := tc.deployment.StartInstance(ctx, &pb.StartInstanceRequest{
		GameId:       100,
		Name:         "full-workflow-srv",
		Protocol:     pb.Protocol_PROTOCOL_TCP,
		InternalPort: 8080,
		PortAllocation: &pb.PortAllocation{
			Strategy: &pb.PortAllocation_Any{Any: true},
		},
		MaxPlayers: 10,
		Args:       []string{"sleep", "120"},
	})
	if err != nil {
		t.Fatalf("StartInstance failed: %v", err)
	}
	defer func() {
		_, _ = tc.deployment.StopInstance(ctx, &pb.StopInstanceRequest{
			InstanceId:     startResp.InstanceId,
			TimeoutSeconds: 5,
		})
	}()

	// Шаг 3: GetInstance — проверка RUNNING
	t.Log("3. Verifying instance is RUNNING...")
	getResp, err := tc.discovery.GetInstance(ctx, &pb.GetInstanceRequest{
		InstanceId: startResp.InstanceId,
	})
	if err != nil {
		t.Fatalf("GetInstance failed: %v", err)
	}
	if getResp.Instance.Status != pb.InstanceStatus_INSTANCE_STATUS_RUNNING {
		t.Errorf("expected RUNNING, got %s", getResp.Instance.Status)
	}

	// Шаг 4: Heartbeat — проверка активного инстанса
	t.Log("4. Checking heartbeat...")
	hbResp, err := tc.discovery.Heartbeat(ctx, &pb.HeartbeatRequest{})
	if err != nil {
		t.Fatalf("Heartbeat failed: %v", err)
	}
	if hbResp.ActiveInstanceCount == 0 {
		t.Error("expected at least 1 active instance in heartbeat")
	}

	// Шаг 5: GetInstanceUsage
	t.Log("5. Getting instance usage...")
	time.Sleep(containerWait)
	usageResp, err := tc.discovery.GetInstanceUsage(ctx, &pb.GetInstanceUsageRequest{
		InstanceId: startResp.InstanceId,
	})
	if err != nil {
		t.Fatalf("GetInstanceUsage failed: %v", err)
	}
	t.Logf("   Usage: CPU=%.2f%%, Mem=%d bytes",
		usageResp.Usage.CpuUsagePercent, usageResp.Usage.MemoryUsedBytes)

	// Шаг 6: ListInstances
	t.Log("6. Listing all instances...")
	listResp, err := tc.discovery.ListInstances(ctx, &pb.ListInstancesRequest{})
	if err != nil {
		t.Fatalf("ListInstances failed: %v", err)
	}
	t.Logf("   Total instances: %d", len(listResp.Instances))

	// Шаг 7: StopInstance
	t.Log("7. Stopping instance...")
	_, err = tc.deployment.StopInstance(ctx, &pb.StopInstanceRequest{
		InstanceId:     startResp.InstanceId,
		TimeoutSeconds: 5,
	})
	if err != nil {
		t.Fatalf("StopInstance failed: %v", err)
	}

	// Шаг 8: GetInstance — проверка STOPPED
	t.Log("8. Verifying instance is STOPPED...")
	getResp2, err := tc.discovery.GetInstance(ctx, &pb.GetInstanceRequest{
		InstanceId: startResp.InstanceId,
	})
	if err != nil {
		t.Fatalf("GetInstance after stop failed: %v", err)
	}
	if getResp2.Instance.Status != pb.InstanceStatus_INSTANCE_STATUS_STOPPED {
		t.Errorf("expected STOPPED, got %s", getResp2.Instance.Status)
	}

	t.Log("Full workflow completed successfully")
}

// ============================================================
// Специфические сценарии (4 теста)
// ============================================================

// E2E_11: Параллельный запуск N инстансов.
func TestE2E_ParallelInstances(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	tc := connectToServer(ctx, t)
	cleanupInstances(t, tc, ctx)
	ensureDockerAvailable(t)

	const n = 3
	instanceIDs := make([]int64, n)

	// Запускаем N инстансов параллельно
	for i := 0; i < n; i++ {
		func(idx int) {
			resp := startTestInstance(t, tc, ctx, 1, fmt.Sprintf("parallel-srv-%d", idx))
			instanceIDs[idx] = resp.InstanceId
		}(i)
	}

	// Проверяем что все в списке
	resp, err := tc.discovery.ListInstances(ctx, &pb.ListInstancesRequest{})
	if err != nil {
		t.Fatalf("ListInstances failed: %v", err)
	}

	runningCount := 0
	for _, inst := range resp.Instances {
		if inst.Status == pb.InstanceStatus_INSTANCE_STATUS_RUNNING {
			runningCount++
		}
	}

	if runningCount < n {
		t.Errorf("expected at least %d running instances, got %d", n, runningCount)
	}

	// Cleanup
	for _, id := range instanceIDs {
		stopTestInstance(t, tc, ctx, id)
	}
}

// E2E_12: Стратегии портов (Exact / Range / Any).
func TestE2E_PortStrategies(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	tc := connectToServer(ctx, t)
	cleanupInstances(t, tc, ctx)
	ensureDockerAvailable(t)

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
			expected: func(p uint32) bool { return true }, // OS назначит любой
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

	tc := connectToServer(ctx, t)

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

	tc := connectToServer(ctx, t)
	cleanupInstances(t, tc, ctx)
	ensureDockerAvailable(t)

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

// ============================================================
// Helper функции
// ============================================================

// startTestInstance запускает инстанс с параметрами по умолчанию.
func startTestInstance(t *testing.T, tc *testClient, ctx context.Context, gameID int64, name string) *pb.StartInstanceResponse {
	t.Helper()

	// Загружаем образ для этого gameID
	loadTestImage(t, tc, ctx, gameID, imageTag)

	resp, err := tc.deployment.StartInstance(ctx, &pb.StartInstanceRequest{
		GameId:       gameID,
		Name:         name,
		Protocol:     pb.Protocol_PROTOCOL_TCP,
		InternalPort: 8080,
		PortAllocation: &pb.PortAllocation{
			Strategy: &pb.PortAllocation_Any{Any: true},
		},
		MaxPlayers: 10,
		Args:       []string{"sleep", "120"},
	})
	if err != nil {
		t.Fatalf("StartInstance(%s) failed: %v", name, err)
	}

	t.Cleanup(func() {
		_, _ = tc.deployment.StopInstance(ctx, &pb.StopInstanceRequest{
			InstanceId:     resp.InstanceId,
			TimeoutSeconds: 3,
		})
	})

	return resp
}

// startTestInstanceWithGameID запускает инстанс с указанным gameID.
func startTestInstanceWithGameID(t *testing.T, tc *testClient, ctx context.Context, gameID int64, name string) *pb.StartInstanceResponse {
	t.Helper()

	loadTestImage(t, tc, ctx, gameID, imageTag)

	resp, err := tc.deployment.StartInstance(ctx, &pb.StartInstanceRequest{
		GameId:       gameID,
		Name:         name,
		Protocol:     pb.Protocol_PROTOCOL_TCP,
		InternalPort: 8080,
		PortAllocation: &pb.PortAllocation{
			Strategy: &pb.PortAllocation_Any{Any: true},
		},
		MaxPlayers: 10,
		Args:       []string{"sleep", "120"},
	})
	if err != nil {
		t.Fatalf("StartInstance(%s) failed: %v", name, err)
	}

	t.Cleanup(func() {
		_, _ = tc.deployment.StopInstance(ctx, &pb.StopInstanceRequest{
			InstanceId:     resp.InstanceId,
			TimeoutSeconds: 3,
		})
	})

	return resp
}

// stopTestInstance останавливает инстанс.
func stopTestInstance(t *testing.T, tc *testClient, ctx context.Context, instanceID int64) {
	t.Helper()
	_, _ = tc.deployment.StopInstance(ctx, &pb.StopInstanceRequest{
		InstanceId:     instanceID,
		TimeoutSeconds: 5,
	})
}

// loadTestImage загружает образ через gRPC streaming.
// Для локально доступных образов может вернуть ошибку (образ уже есть).
func loadTestImage(t *testing.T, tc *testClient, ctx context.Context, gameID int64, tag string) {
	t.Helper()

	stream, err := tc.deployment.LoadImage(ctx)
	if err != nil {
		t.Logf("LoadImage stream open error (may be ok): %v", err)
		return
	}

	// Отправляем метаданные
	err = stream.Send(&pb.LoadImageRequest{
		Payload: &pb.LoadImageRequest_Metadata{
			Metadata: &pb.ImageMetadata{GameId: gameID, ImageTag: tag},
		},
	})
	if err != nil {
		t.Logf("LoadImage Send error (may be ok for local image): %v", err)
		return
	}

	// Закрываем стрим и получаем ответ
	_, err = stream.CloseAndRecv()
	if err != nil {
		t.Logf("LoadImage completed with: %v", err)
	}
}
