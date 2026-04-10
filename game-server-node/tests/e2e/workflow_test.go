package e2e

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	pb "github.com/Be4Die/game-developer-hub/protos/game_server_node/v1"
	"github.com/docker/docker/client"
)

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

// ============================================================
// Helper функции
// ============================================================

// startTestInstance запускает инстанс с параметрами по умолчанию.
func startTestInstance(ctx context.Context, t *testing.T, tc *testClient, gameID int64, name string) *pb.StartInstanceResponse {
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
func saveDockerImageToTar(ctx context.Context, imageTag string) ([]byte, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer func() { _ = cli.Close() }()

	imageTarReader, err := cli.ImageSave(ctx, []string{imageTag})
	if err != nil {
		return nil, fmt.Errorf("failed to save image '%s': %w", imageTag, err)
	}
	defer func() { _ = imageTarReader.Close() }()

	data, err := io.ReadAll(imageTarReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read image tar data: %w", err)
	}

	return data, nil
}
