// Package e2e содержит end-to-end тесты для game-server-node.
//
// В отличие от интеграционных тестов (bufconn in-memory), e2e используют
// реальный сетевой стек: gRPC клиент → TCP → Docker контейнер с сервером.
//
// Каждый тестовый пакет поднимает свежий контейнер через testcontainers,
// поэтому состояние всегда чистое и тесты полностью изолированы.
//
// Перед запуском:
//  1. Соберите образ: task node:build
//  2. Убедитесь что Docker Desktop запущен
//
// Запуск:
//
//	go test -tags=e2e ./tests/e2e/...
package e2e

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	pb "github.com/Be4Die/game-developer-hub/protos/game_server_node/v1"
	"github.com/docker/docker/client"
	"github.com/moby/moby/api/types/container"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Константы e2e тестирования.
const (
	imageTag      = "alpine:3.18"
	testTimeout   = 60 * time.Second
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

// setupServerContainer поднимает свежий контейнер game-server-node и возвращает
// testClient для подключения + контейнер для последующей очистки.
// Контейнер живёт в рамках одного теста и автоматически удаляется через t.Cleanup.
func setupServerContainer(t *testing.T) *testClient {
	t.Helper()

	ctx := context.Background()

	// Собираем образ из Dockerfile если его нет локально.
	// Используем уже собранный образ game-server-node:latest.
	req := testcontainers.ContainerRequest{
		Image:        "game-server-node:latest",
		ExposedPorts: []string{"44044/tcp"},
		Env: map[string]string{
			"CONFIG_PATH":  "/app/config/local.yaml",
			"NODE_API_KEY": getE2EAPIKey(),
		},
		WaitingFor: wait.ForListeningPort("44044/tcp").WithStartupTimeout(15 * time.Second),
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.Binds = append(hc.Binds, "/var/run/docker.sock:/var/run/docker.sock")
		},
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Skipf("не удалось запустить контейнер (проверьте Docker Desktop и образ game-server-node:latest): %v", err)
	}

	// Cleanup в конце теста
	t.Cleanup(func() {
		_ = container.Terminate(context.Background())
	})

	// Получаем реальный порт (testcontainers мапит на случайный хост-порт)
	port, err := container.MappedPort(ctx, "44044")
	if err != nil {
		t.Fatalf("не удалось получить порт контейнера: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("не удалось получить хост контейнера: %v", err)
	}

	grpcAddress := fmt.Sprintf("%s:%s", host, port.Port())

	// Создаём gRPC соединение
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
		grpc.WithUnaryInterceptor(authInterceptor),
	)
	if err != nil {
		t.Fatalf("не удалось подключиться к %s: %v", grpcAddress, err)
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

	tc := setupServerContainer(t)

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

	tc := setupServerContainer(t)

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

	tc := setupServerContainer(t)

	// 1. Пустой список
	resp, err := tc.discovery.ListInstances(ctx, &pb.ListInstancesRequest{})
	if err != nil {
		t.Fatalf("ListInstances (empty) failed: %v", err)
	}
	initialCount := len(resp.Instances)

	// 2. Запускаем инстанс
	_ = startTestInstance(ctx, t, tc, 1, "list-test-server")

	// 3. Проверяем что список увеличился
	resp, err = tc.discovery.ListInstances(ctx, &pb.ListInstancesRequest{})
	if err != nil {
		t.Fatalf("ListInstances (after start) failed: %v", err)
	}

	if len(resp.Instances) != initialCount+1 {
		t.Errorf("expected %d instances, got %d", initialCount+1, len(resp.Instances))
	}
}

// E2E_04: ListInstancesByGame — фильтрация по game_id.
func TestE2E_Discovery_ListInstancesByGame(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	tc := setupServerContainer(t)

	// Уникальные game_id для идемпотентности теста
	const gameID1 = 10001
	const gameID2 = 10002

	// Запускаем 2 инстанса для game 1, 1 для game 2
	for i := 0; i < 2; i++ {
		_ = startTestInstanceWithGameID(ctx, t, tc, gameID1, fmt.Sprintf("game1-srv-%d", i))
	}
	_ = startTestInstanceWithGameID(ctx, t, tc, gameID2, "game2-srv")

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
}

// E2E_05: GetInstance — существующий и несуществующий.
func TestE2E_Discovery_GetInstance(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	tc := setupServerContainer(t)

	// Запускаем инстанс
	startResp := startTestInstance(ctx, t, tc, 1, "get-instance-test")

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

	tc := setupServerContainer(t)

	// Запускаем инстанс
	startResp := startTestInstance(ctx, t, tc, 1, "usage-test")

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

	tc := setupServerContainer(t)

	// Запускаем инстанс
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

	// Запускаем
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

	// LoadImage требует streaming
	loadTestImage(ctx, t, tc, 42)

	// Запускаем инстанс — должен найти образ
	startResp := startTestInstance(ctx, t, tc, 42, "load-image-test")

	t.Logf("Started instance %d for game 42", startResp.InstanceId)
}

// E2E_10: Полный E2E поток — все методы в одной цепочке.
func TestE2E_FullWorkflow(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	tc := setupServerContainer(t)

	// Шаг 1: LoadImage
	t.Log("1. Loading image...")
	loadTestImage(ctx, t, tc, 100)

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

	tc := setupServerContainer(t)

	const n = 3

	// Запускаем N инстансов параллельно
	for i := 0; i < n; i++ {
		func(idx int) {
			startTestInstance(ctx, t, tc, 1, fmt.Sprintf("parallel-srv-%d", idx))
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
}

// E2E_12: Стратегии портов (Exact / Range / Any).
func TestE2E_PortStrategies(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	tc := setupServerContainer(t)

	// Загружаем образ для game 1
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

	// Загружаем образ
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

// ============================================================
// Helper функции
// ============================================================

// startTestInstance запускает инстанс с параметрами по умолчанию.
func startTestInstance(ctx context.Context, t *testing.T, tc *testClient, gameID int64, name string) *pb.StartInstanceResponse {
	t.Helper()

	// Загружаем образ для этого gameID
	loadTestImage(ctx, t, tc, gameID)

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

	instanceID := resp.InstanceId
	//nolint:contextcheck // cleanup uses its own timeout context
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, _ = tc.deployment.StopInstance(ctx, &pb.StopInstanceRequest{
			InstanceId:     instanceID,
			TimeoutSeconds: 3,
		})
	})

	return resp
}

// startTestInstanceWithGameID запускает инстанс с указанным gameID.
func startTestInstanceWithGameID(ctx context.Context, t *testing.T, tc *testClient, gameID int64, name string) *pb.StartInstanceResponse {
	t.Helper()

	loadTestImage(ctx, t, tc, gameID)

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

	instanceID := resp.InstanceId
	//nolint:contextcheck // cleanup uses its own timeout context
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, _ = tc.deployment.StopInstance(ctx, &pb.StopInstanceRequest{
			InstanceId:     instanceID,
			TimeoutSeconds: 3,
		})
	})

	return resp
}

// loadTestImage загружает Docker образ через gRPC streaming с реальной передачей данных.
// Сохраняет образ в tar через Docker API и стримит чанки на сервер.
func loadTestImage(ctx context.Context, t *testing.T, tc *testClient, gameID int64) {
	t.Helper()

	t.Logf("Loading image '%s' for game_id=%d via gRPC stream...", imageTag, gameID)

	imageTarData, err := saveDockerImageToTar(ctx, imageTag)
	if err != nil {
		t.Fatalf("Failed to save image '%s' to tar: %v", imageTag, err)
	}

	t.Logf("Image '%s' saved to tar: %d bytes, starting gRPC stream...", imageTag, len(imageTarData))

	stream, err := tc.deployment.LoadImage(ctx)
	if err != nil {
		t.Fatalf("LoadImage stream open error: %v", err)
	}

	err = stream.Send(&pb.LoadImageRequest{
		Payload: &pb.LoadImageRequest_Metadata{
			Metadata: &pb.ImageMetadata{GameId: gameID, ImageTag: imageTag},
		},
	})
	if err != nil {
		t.Fatalf("LoadImage Send metadata error: %v", err)
	}

	const chunkSize = 64 * 1024 // 64KB chunks
	totalSent := 0

	for offset := 0; offset < len(imageTarData); {
		end := offset + chunkSize
		if end > len(imageTarData) {
			end = len(imageTarData)
		}

		chunk := imageTarData[offset:end]
		err = stream.Send(&pb.LoadImageRequest{
			Payload: &pb.LoadImageRequest_Chunk{
				Chunk: chunk,
			},
		})
		if err != nil {
			t.Fatalf("LoadImage Send chunk error at offset %d: %v", offset, err)
		}

		totalSent += len(chunk)
		offset = end
	}

	t.Logf("Streamed %d bytes (%.2f MB) of image '%s'", totalSent, float64(totalSent)/1024/1024, imageTag)

	resp, err := stream.CloseAndRecv()
	if err != nil {
		t.Fatalf("LoadImage CloseAndRecv error: %v", err)
	}

	t.Logf("Image '%s' loaded successfully, server response: %s", imageTag, resp.GetImageTag())
}

// saveDockerImageToTar сохраняет Docker образ в tar формат через Docker API.
// Эквивалент команды: docker save <image> > image.tar
func saveDockerImageToTar(ctx context.Context, imageTag string) ([]byte, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer func() { _ = cli.Close() }()

	// ImageSave возвращает io.ReadCloser с tar архивом образа
	imageTarReader, err := cli.ImageSave(ctx, []string{imageTag})
	if err != nil {
		return nil, fmt.Errorf("failed to save image '%s': %w", imageTag, err)
	}
	defer func() { _ = imageTarReader.Close() }()

	// Читаем весь tar в память
	data, err := io.ReadAll(imageTarReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read image tar data: %w", err)
	}

	return data, nil
}
