// Package e2e содержит end-to-end тесты для game-server-node.
//
// В отличие от интеграционных тестов (bufconn in-memory), e2e используют
// реальный сетевой стек: gRPC клиент → TCP → Docker контейнер с сервером.
//
// Каждый тест поднимает свежий контейнер через testcontainers,
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
	"log/slog"
	"os"
	"testing"
	"time"

	pb "github.com/Be4Die/game-developer-hub/protos/game_server_node/v1"
	"github.com/moby/moby/api/types/container"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
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
// testClient для подключения. Контейнер автоматически уничтожается через t.Cleanup.
func setupServerContainer(t *testing.T) *testClient {
	t.Helper()

	ctx := context.Background()

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

	cont, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Skipf("не удалось запустить контейнер (проверьте Docker Desktop и образ game-server-node:latest): %v", err)
	}

	t.Cleanup(func() {
		_ = cont.Terminate(context.Background())
	})

	port, err := cont.MappedPort(ctx, "44044")
	if err != nil {
		t.Fatalf("не удалось получить порт контейнера: %v", err)
	}

	host, err := cont.Host(ctx)
	if err != nil {
		t.Fatalf("не удалось получить хост контейнера: %v", err)
	}

	grpcAddress := host + ":" + port.Port()

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

// TestMain — точка входа для всех e2e тестов.
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
