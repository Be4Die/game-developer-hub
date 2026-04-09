// Package integration содержит интеграционные тесты для game-server-node.
//
// Эти тесты проверяют реальное взаимодействие компонентов:
// - Docker Runtime с реальным Docker daemon
// - gRPC сервер и клиент через сетевой стек
// - Полные потоки: загрузка образа → запуск инстанса → получение метрик → остановка
//
// Требования:
// - Доступный Docker daemon
// - Запуск: go test -tags=integration ./tests/integration/...
package integration

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/game-server-node/internal/config"
	"github.com/Be4Die/game-developer-hub/game-server-node/internal/domain"
	"github.com/Be4Die/game-developer-hub/game-server-node/internal/runtime/docker"
	"github.com/Be4Die/game-developer-hub/game-server-node/internal/service"
	"github.com/Be4Die/game-developer-hub/game-server-node/internal/storage/memory"
	grpctransport "github.com/Be4Die/game-developer-hub/game-server-node/internal/transport/grpc"
	pb "github.com/Be4Die/game-developer-hub/protos/game_server_node/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

// skipIfNoDocker пропускает тест если Docker daemon недоступен.
func skipIfNoDocker(t *testing.T) {
	t.Helper()

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	r, err := docker.New(log)
	if err != nil {
		t.Skipf("Docker daemon not available: %v", err)
	}
	_ = r // runtime создан успешно
}

// testConfig создаёт конфигурацию для тестов.
func testConfig(port int) *config.Config {
	return &config.Config{
		Env:         config.EnvLocal,
		StoragePath: "/tmp/test.db",
		TokenTTL:    time.Hour,
		GRPC: config.GRPCConfig{
			Port:    port,
			Timeout: 30 * time.Second,
		},
		Node: config.NodeConfig{
			Region:  "test-region",
			Version: "test-0.0.1",
			EthName: "",
		},
	}
}

// integrationTestEnv хранит зависимости для интеграционных тестов.
type integrationTestEnv struct {
	storage          *memory.Storage
	runtime          *docker.Runtime
	discoverySvc     *service.DiscoveryService
	deploymentSvc    *service.DeploymentService
	discoveryHdl     *grpctransport.DiscoveryHandler
	deploymentHdl    *grpctransport.DeploymentHandler
	grpcServer       *grpc.Server
	bufDialer        func(context.Context, string) (net.Conn, error)
	listener         *bufconn.Listener
	clientConn       *grpc.ClientConn
	discoveryClient  pb.DiscoveryServiceClient
	deploymentClient pb.DeploymentServiceClient
	log              *slog.Logger
}

// setupIntegration поднимает полный стек для интеграционного теста.
func setupIntegration(t *testing.T) *integrationTestEnv {
	t.Helper()
	skipIfNoDocker(t)

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// Создаём компоненты
	storage := memory.NewStorage()
	runtime, err := docker.New(log)
	if err != nil {
		t.Fatalf("failed to create docker runtime: %v", err)
	}

	cfg := testConfig(0) // порт не важен для bufconn

	discoverySvc := service.NewDiscoveryService(storage, runtime, cfg)
	deploymentSvc := service.NewDeploymentService(log, storage, runtime)

	discoveryHdl := grpctransport.NewDiscoveryHandler(discoverySvc)
	deploymentHdl := grpctransport.NewDeploymentHandler(deploymentSvc)

	// bufconn listener для in-memory gRPC
	const bufSize = 1024 * 1024
	listener := bufconn.Listen(bufSize)

	grpcServer := grpc.NewServer()
	pb.RegisterDiscoveryServiceServer(grpcServer, discoveryHdl)
	pb.RegisterDeploymentServiceServer(grpcServer, deploymentHdl)

	go func() {
		_ = grpcServer.Serve(listener)
	}()

	bufDialer := func(ctx context.Context, address string) (net.Conn, error) {
		return listener.Dial()
	}

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("failed to connect to gRPC server: %v", err)
	}

	t.Cleanup(func() {
		_ = conn.Close()
		grpcServer.GracefulStop()
		_ = listener.Close()
	})

	return &integrationTestEnv{
		storage:          storage,
		runtime:          runtime,
		discoverySvc:     discoverySvc,
		deploymentSvc:    deploymentSvc,
		discoveryHdl:     discoveryHdl,
		deploymentHdl:    deploymentHdl,
		grpcServer:       grpcServer,
		bufDialer:        bufDialer,
		listener:         listener,
		clientConn:       conn,
		discoveryClient:  pb.NewDiscoveryServiceClient(conn),
		deploymentClient: pb.NewDeploymentServiceClient(conn),
		log:              log,
	}
}

// TestMain позволяет запускать интеграционные тесты с флагом -tags=integration.
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

// ============================================================
// Интеграционные тесты DeploymentService с реальным Docker
// ============================================================

// TestDockerRuntime_CreateAndRemoveContainer проверяет полный жизненный цикл контейнера.
func TestDockerRuntime_CreateAndRemoveContainer(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	const imageTag = "alpine:3.18"

	// Создаём контейнер (образ должен быть локально доступен)
	containerID, err := env.runtime.CreateContainer(ctx, domain.ContainerOpts{
		ImageTag:     imageTag,
		InternalPort: 8080,
		HostPort:     0, // OS назначит сама
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

	// Удаляем контейнер (cleanup)
	err = env.runtime.RemoveContainer(ctx, containerID)
	if err != nil {
		t.Fatalf("RemoveContainer failed: %v", err)
	}
}

// TestDockerRuntime_CreateStartStopContainer проверяет создание, запуск и остановку.
func TestDockerRuntime_CreateStartStopContainer(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	const imageTag = "alpine:3.18"

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

	// Запускаем
	err = env.runtime.StartContainer(ctx, containerID)
	if err != nil {
		_ = env.runtime.RemoveContainer(ctx, containerID)
		t.Fatalf("StartContainer failed: %v", err)
	}

	// Даём контейнеру время запуститься
	time.Sleep(500 * time.Millisecond)

	// Останавливаем
	err = env.runtime.StopContainer(ctx, containerID, 5*time.Second)
	if err != nil {
		_ = env.runtime.RemoveContainer(ctx, containerID)
		t.Fatalf("StopContainer failed: %v", err)
	}

	// Удаляем
	err = env.runtime.RemoveContainer(ctx, containerID)
	if err != nil {
		t.Fatalf("RemoveContainer failed: %v", err)
	}
}

// TestDockerRuntime_ContainerLogs проверяет получение логов контейнера.
func TestDockerRuntime_ContainerLogs(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	const imageTag = "alpine:3.18"

	containerID, err := env.runtime.CreateContainer(ctx, domain.ContainerOpts{
		ImageTag:     imageTag,
		InternalPort: 80,
		HostPort:     0,
		Args:         []string{"sh", "-c", "echo 'integration-test-log' && sleep 1"},
	})
	if err != nil {
		t.Fatalf("CreateContainer failed: %v", err)
	}

	// Запускаем и ждём завершения
	err = env.runtime.StartContainer(ctx, containerID)
	if err != nil {
		_ = env.runtime.RemoveContainer(ctx, containerID)
		t.Fatalf("StartContainer failed: %v", err)
	}

	time.Sleep(2 * time.Second)

	// Читаем логи
	logs, err := env.runtime.ContainerLogs(ctx, containerID, false)
	if err != nil {
		_ = env.runtime.RemoveContainer(ctx, containerID)
		t.Fatalf("ContainerLogs failed: %v", err)
	}

	// Docker логирует в формате: [8字节header][payload]
	buf := make([]byte, 1024)
	n, err := logs.Read(buf)
	if err != nil {
		_ = logs.Close()
		_ = env.runtime.RemoveContainer(ctx, containerID)
		t.Fatalf("reading logs failed: %v", err)
	}
	_ = logs.Close()

	logContent := string(buf[8:n]) // skip 8-byte Docker header
	if logContent == "" {
		t.Log("container logs are empty (may be ok for short-lived containers)")
	} else {
		t.Logf("container logs: %s", logContent)
	}

	// Cleanup
	_ = env.runtime.RemoveContainer(ctx, containerID)
}

// ============================================================
// Интеграционные тесты DeploymentService (сервисный слой)
// ============================================================

// TestDeploymentService_FullLifecycle проверяет полный цикл: LoadImage → StartInstance → StopInstance.
func TestDeploymentService_FullLifecycle(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	const imageTag = "alpine:3.18"

	// 1. Загружаем образ (LoadImage требует io.Reader — передаём nil для локального образа)
	// Для реального теста нужен tar-архив от docker save.
	// Здесь проверяем что сервис корректно работает с уже локальным образом.
	env.deploymentSvc.LoadImage(ctx, 1, imageTag, nil)

	// 2. Запускаем инстанс
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

	// 3. Проверяем что инстанс в хранилище
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

	// 4. Останавливаем инстанс
	err = env.deploymentSvc.StopInstance(ctx, instanceID, 5*time.Second)
	if err != nil {
		t.Fatalf("StopInstance failed: %v", err)
	}

	// 5. Проверяем статус
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

	const imageTag = "alpine:3.18"
	env.deploymentSvc.LoadImage(ctx, 1, imageTag, nil)

	// Запускаем 3 инстанса
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
			// Cleanup ранее запущенных
			for _, prevID := range instanceIDs {
				_ = env.deploymentSvc.StopInstance(ctx, prevID, 3*time.Second)
			}
			t.Fatalf("StartInstance %d failed: %v", i, err)
		}
		instanceIDs = append(instanceIDs, id)
		t.Logf("instance %d started on port %d", id, port)
	}

	// Проверяем что все в хранилище
	all, err := env.storage.GetAllInstances(ctx)
	if err != nil {
		t.Fatalf("GetAllInstances failed: %v", err)
	}

	if len(all) != 3 {
		t.Errorf("expected 3 instances, got %d", len(all))
	}

	// Останавливаем все
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

	const imageTag = "alpine:3.18"
	env.deploymentSvc.LoadImage(ctx, 1, imageTag, nil)

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

	const imageTag = "alpine:3.18"
	env.deploymentSvc.LoadImage(ctx, 1, imageTag, nil)

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

// ============================================================
// Интеграционные тесты полного gRPC стека
// ============================================================

// TestGRPC_FullLifecycle_StartAndStop проверяет полный цикл через gRPC клиент-сервер.
func TestGRPC_FullLifecycle_StartAndStop(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	const imageTag = "alpine:3.18"

	// Предварительно загружаем образ через сервис (т.к. LoadImage — streaming RPC)
	env.deploymentSvc.LoadImage(ctx, 1, imageTag, nil)

	// StartInstance через gRPC
	startReq := &pb.StartInstanceRequest{
		GameId:       1,
		Name:         "grpc-test-server",
		Protocol:     pb.Protocol_PROTOCOL_TCP,
		InternalPort: 8080,
		PortAllocation: &pb.PortAllocation{
			Strategy: &pb.PortAllocation_Any{Any: true},
		},
		MaxPlayers: 10,
		Args:       []string{"sleep", "60"},
	}

	startResp, err := env.deploymentClient.StartInstance(ctx, startReq)
	if err != nil {
		t.Fatalf("StartInstance RPC failed: %v", err)
	}

	if startResp.InstanceId <= 0 {
		t.Errorf("expected positive instance ID, got %d", startResp.InstanceId)
	}
	t.Logf("gRPC StartInstance: instance_id=%d, host_port=%d", startResp.InstanceId, startResp.HostPort)

	// Проверяем через DiscoveryService
	getReq := &pb.GetInstanceRequest{InstanceId: startResp.InstanceId}
	getResp, err := env.discoveryClient.GetInstance(ctx, getReq)
	if err != nil {
		t.Fatalf("GetInstance RPC failed: %v", err)
	}

	if getResp.Instance.Name != "grpc-test-server" {
		t.Errorf("expected name 'grpc-test-server', got '%s'", getResp.Instance.Name)
	}
	if getResp.Instance.Status != pb.InstanceStatus_INSTANCE_STATUS_RUNNING {
		t.Errorf("expected status RUNNING, got %s", getResp.Instance.Status)
	}

	// ListInstances
	listResp, err := env.discoveryClient.ListInstances(ctx, &pb.ListInstancesRequest{})
	if err != nil {
		t.Fatalf("ListInstances RPC failed: %v", err)
	}

	if len(listResp.Instances) < 1 {
		t.Errorf("expected at least 1 instance, got %d", len(listResp.Instances))
	}

	// StopInstance
	stopResp, err := env.deploymentClient.StopInstance(ctx, &pb.StopInstanceRequest{
		InstanceId:     startResp.InstanceId,
		TimeoutSeconds: 5,
	})
	if err != nil {
		t.Fatalf("StopInstance RPC failed: %v", err)
	}
	_ = stopResp

	// Проверяем статус после остановки
	getResp2, err := env.discoveryClient.GetInstance(ctx, getReq)
	if err != nil {
		t.Fatalf("GetInstance after stop failed: %v", err)
	}

	if getResp2.Instance.Status != pb.InstanceStatus_INSTANCE_STATUS_STOPPED {
		t.Errorf("expected status STOPPED, got %s", getResp2.Instance.Status)
	}
}

// TestGRPC_Heartbeat проверяет heartbeat через gRPC.
func TestGRPC_Heartbeat(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	resp, err := env.discoveryClient.Heartbeat(ctx, &pb.HeartbeatRequest{})
	if err != nil {
		t.Fatalf("Heartbeat RPC failed: %v", err)
	}

	// На Windows sysinfo возвращает stub-данные
	t.Logf("Heartbeat: active_instances=%d, cpu=%.2f%%, memory=%d bytes",
		resp.ActiveInstanceCount, resp.Usage.CpuUsagePercent, resp.Usage.MemoryUsedBytes)
}

// TestGRPC_GetNodeInfo проверяет GetNodeInfo через gRPC.
func TestGRPC_GetNodeInfo(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	resp, err := env.discoveryClient.GetNodeInfo(ctx, &pb.GetNodeInfoRequest{})
	if err != nil {
		t.Fatalf("GetNodeInfo RPC failed: %v", err)
	}

	if resp.Region != "test-region" {
		t.Errorf("expected region 'test-region', got '%s'", resp.Region)
	}
	if resp.AgentVersion != "test-0.0.1" {
		t.Errorf("expected version 'test-0.0.1', got '%s'", resp.AgentVersion)
	}

	t.Logf("GetNodeInfo: region=%s, version=%s, cpu_cores=%d, total_memory=%d bytes",
		resp.Region, resp.AgentVersion, resp.CpuCores, resp.TotalMemoryBytes)
}

// TestGRPC_ListInstancesByGame проверяет фильтрацию по game_id через gRPC.
func TestGRPC_ListInstancesByGame(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	const imageTag = "alpine:3.18"
	env.deploymentSvc.LoadImage(ctx, 1, imageTag, nil)
	env.deploymentSvc.LoadImage(ctx, 2, imageTag, nil)

	// Запускаем 2 инстанса для game 1 и 1 для game 2
	var game1IDs []int64
	for i := 0; i < 2; i++ {
		resp, err := env.deploymentClient.StartInstance(ctx, &pb.StartInstanceRequest{
			GameId:       1,
			Name:         fmt.Sprintf("game1-server-%d", i),
			Protocol:     pb.Protocol_PROTOCOL_TCP,
			InternalPort: 8080,
			PortAllocation: &pb.PortAllocation{
				Strategy: &pb.PortAllocation_Any{Any: true},
			},
			MaxPlayers: 10,
			Args:       []string{"sleep", "60"},
		})
		if err != nil {
			t.Fatalf("StartInstance for game 1 failed: %v", err)
		}
		game1IDs = append(game1IDs, resp.InstanceId)
	}

	resp2, err := env.deploymentClient.StartInstance(ctx, &pb.StartInstanceRequest{
		GameId:       2,
		Name:         "game2-server",
		Protocol:     pb.Protocol_PROTOCOL_TCP,
		InternalPort: 8080,
		PortAllocation: &pb.PortAllocation{
			Strategy: &pb.PortAllocation_Any{Any: true},
		},
		MaxPlayers: 10,
		Args:       []string{"sleep", "60"},
	})
	if err != nil {
		t.Fatalf("StartInstance for game 2 failed: %v", err)
	}
	defer func() {
		_, _ = env.deploymentClient.StopInstance(ctx, &pb.StopInstanceRequest{InstanceId: resp2.InstanceId, TimeoutSeconds: 3})
	}()

	// Запрашиваем инстансы для game 1
	listByGame, err := env.discoveryClient.ListInstancesByGame(ctx, &pb.ListInstancesByGameRequest{GameId: 1})
	if err != nil {
		t.Fatalf("ListInstancesByGame RPC failed: %v", err)
	}

	if len(listByGame.Instances) != 2 {
		t.Errorf("expected 2 instances for game 1, got %d", len(listByGame.Instances))
	}

	// Cleanup
	for _, id := range game1IDs {
		_, _ = env.deploymentClient.StopInstance(ctx, &pb.StopInstanceRequest{InstanceId: id, TimeoutSeconds: 3})
	}
}

// TestGRPC_StopNonExistentInstance проверяет корректную обработку ошибки через gRPC.
func TestGRPC_StopNonExistentInstance(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	_, err := env.deploymentClient.StopInstance(ctx, &pb.StopInstanceRequest{
		InstanceId:     99999,
		TimeoutSeconds: 5,
	})
	if err == nil {
		t.Fatal("expected error when stopping non-existent instance via gRPC")
	}

	t.Logf("got expected gRPC error: %v", err)
}

// TestGRPC_GetInstanceUsage проверяет получение метрик через gRPC.
func TestGRPC_GetInstanceUsage(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	const imageTag = "alpine:3.18"
	env.deploymentSvc.LoadImage(ctx, 1, imageTag, nil)

	// Запускаем инстанс
	startResp, err := env.deploymentClient.StartInstance(ctx, &pb.StartInstanceRequest{
		GameId:       1,
		Name:         "usage-test-server",
		Protocol:     pb.Protocol_PROTOCOL_TCP,
		InternalPort: 8080,
		PortAllocation: &pb.PortAllocation{
			Strategy: &pb.PortAllocation_Any{Any: true},
		},
		MaxPlayers: 10,
		Args:       []string{"sleep", "60"},
	})
	if err != nil {
		t.Fatalf("StartInstance failed: %v", err)
	}
	defer func() {
		_, _ = env.deploymentClient.StopInstance(ctx, &pb.StopInstanceRequest{InstanceId: startResp.InstanceId, TimeoutSeconds: 3})
	}()

	time.Sleep(500 * time.Millisecond)

	// Получаем метрики
	usageResp, err := env.discoveryClient.GetInstanceUsage(ctx, &pb.GetInstanceUsageRequest{
		InstanceId: startResp.InstanceId,
	})
	if err != nil {
		t.Fatalf("GetInstanceUsage RPC failed: %v", err)
	}

	if usageResp.InstanceId != startResp.InstanceId {
		t.Errorf("expected instance ID %d, got %d", startResp.InstanceId, usageResp.InstanceId)
	}

	t.Logf("Instance usage: CPU=%.2f%%, Memory=%d bytes, Disk=%d, Network=%d",
		usageResp.Usage.CpuUsagePercent,
		usageResp.Usage.MemoryUsedBytes,
		usageResp.Usage.DiskUsedBytes,
		usageResp.Usage.NetworkBytesPerSec,
	)
}
