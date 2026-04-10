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
	"log/slog"
	"net"
	"os"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/game-server-node/internal/config"
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
	_ = r
}

// testConfig создаёт конфигурацию для тестов.
func testConfig(port int) *config.Config {
	return &config.Config{
		Env:         config.EnvLocal,
		StoragePath: "/tmp/test.db",
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

	storage := memory.NewStorage()
	runtime, err := docker.New(log)
	if err != nil {
		t.Fatalf("failed to create docker runtime: %v", err)
	}

	cfg := testConfig(0)

	discoverySvc := service.NewDiscoveryService(storage, runtime, cfg)
	deploymentSvc := service.NewDeploymentService(log, storage, runtime)

	discoveryHdl := grpctransport.NewDiscoveryHandler(discoverySvc)
	deploymentHdl := grpctransport.NewDeploymentHandler(deploymentSvc)

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
