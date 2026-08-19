package integration

import (
	"archive/zip"
	"bytes"
	"context"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/service"
	"github.com/Be4Die/game-developer-hub/project-manager/internal/storage/deployment"
	"github.com/Be4Die/game-developer-hub/project-manager/internal/storage/filesystem"
	grpctransport "github.com/Be4Die/game-developer-hub/project-manager/internal/transport/grpc"
	pb "github.com/Be4Die/game-developer-hub/protos/project_manager/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

func setupIntegrationServer(t *testing.T) (pb.ProjectServiceClient, func()) {
	t.Helper()

	lis := bufconn.Listen(bufSize)
	tmpDir := t.TempDir()

	pRepo := newMockProjectRepo()
	dRepo := newMockDraftRepo()
	bRepo := newMockBuildRepo()
	mClient := newMockModerationClient()
	rRepo := newMockReleaseRepo()
	depRepo := newMockDeploymentRepo()

	bStorage := filesystem.NewBuildStorage(filepath.Join(tmpDir, "projects"))
	mStorage := filesystem.NewMediaStorage(filepath.Join(tmpDir, "projects"))
	deployer := deployment.NewLocalDeployer(filepath.Join(tmpDir, "games"), "/games")

	projSvc := service.NewProjectService(
		pRepo, dRepo, bRepo, rRepo, depRepo, mClient,
		bStorage, mStorage, deployer, nil, 5,
	)

	projHandler := grpctransport.NewProjectHandler(projSvc)

	s := grpc.NewServer()
	pb.RegisterProjectServiceServer(s, projHandler)

	go func() {
		if err := s.Serve(lis); err != nil {
			// ignore on close
		}
	}()

	conn, err := grpc.NewClient(
		"passthrough://bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("failed to dial bufnet: %v", err)
	}

	cleanup := func() {
		_ = conn.Close()
		s.Stop()
		_ = lis.Close()
	}

	return pb.NewProjectServiceClient(conn), cleanup
}

func createTestGameZip(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	f, err := zw.Create("index.html")
	if err != nil {
		t.Fatalf("create index.html: %v", err)
	}
	_, _ = f.Write([]byte("<!DOCTYPE html><html><body><h1>Space Shooter Web Game</h1></body></html>"))

	asset, err := zw.Create("game.js")
	if err != nil {
		t.Fatalf("create game.js: %v", err)
	}
	_, _ = asset.Write([]byte("console.log('game loaded');"))

	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return buf.Bytes()
}

func TestIntegration_FullProjectLifecycle(t *testing.T) {
	projClient, cleanup := setupIntegrationServer(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs("x-user-id", "dev-user-42", "x-user-role", "developer"))

	// 1. Создание проекта
	createResp, err := projClient.Create(userCtx, &pb.ProjectCreateRequest{
		TitleRu: "Космический Шутер",
		TitleEn: "Space Shooter",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	projectID := createResp.GetProject().GetId()
	if projectID == 0 {
		t.Fatalf("expected non-zero project id")
	}

	// 2. Обновление метаданных черновика
	_, err = projClient.Update(userCtx, &pb.ProjectUpdateRequest{
		Id:                 projectID,
		TitleRu:            "Космический Шутер Ремастер",
		TitleEn:            "Space Shooter Remastered",
		About:              "Захватывающая игра про космос с новейшей графикой.",
		SeoRu:              "космос, игра, шутер",
		SeoEn:              "space, game, shooter",
		ActiveBuildVersion: "1.0.0",
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// 3. Потоковая загрузка клиентской веб-сборки (64 КБ чанки)
	zipData := createTestGameZip(t)
	stream, err := projClient.UploadBuildStream(userCtx)
	if err != nil {
		t.Fatalf("open upload build stream: %v", err)
	}

	// Первое сообщение: метаданные
	err = stream.Send(&pb.ProjectUploadBuildStreamRequest{
		Payload: &pb.ProjectUploadBuildStreamRequest_Metadata{
			Metadata: &pb.BuildUploadStreamMetadata{
				ProjectId: projectID,
				Version:   "1.0.0",
			},
		},
	})
	if err != nil {
		t.Fatalf("send metadata: %v", err)
	}

	// Отправка чанков
	chunkSize := 32
	for i := 0; i < len(zipData); i += chunkSize {
		end := i + chunkSize
		if end > len(zipData) {
			end = len(zipData)
		}
		err = stream.Send(&pb.ProjectUploadBuildStreamRequest{
			Payload: &pb.ProjectUploadBuildStreamRequest_Chunk{
				Chunk: zipData[i:end],
			},
		})
		if err != nil {
			t.Fatalf("send chunk: %v", err)
		}
	}

	uploadResp, err := stream.CloseAndRecv()
	if err != nil {
		t.Fatalf("upload stream close and recv: %v", err)
	}
	if !uploadResp.GetSuccess() {
		t.Fatalf("expected success upload")
	}
	if uploadResp.GetDevUrl() == "" {
		t.Errorf("expected dev_url to be returned")
	}

	// 4. Отправка на модерацию (через Stub Moderation Client)
	submitResp, err := projClient.SubmitForModeration(userCtx, &pb.SubmitForModerationRequest{
		ProjectId: projectID,
	})
	if err != nil {
		t.Fatalf("SubmitForModeration failed: %v", err)
	}
	if !submitResp.GetSuccess() || submitResp.GetRequestId() == 0 {
		t.Errorf("expected success and non-zero request_id, got: %+v", submitResp)
	}

	// 5. Публикация релиза (вызов от имени модератора/сервиса модерации)
	publishResp, err := projClient.PublishRelease(ctx, &pb.ProjectPublishReleaseRequest{
		ProjectId:   projectID,
		Version:     "1.0.0",
		PublishedBy: "mod-999",
		Comment:     "Проверено, игра отличная",
	})
	if err != nil {
		t.Fatalf("PublishRelease failed: %v", err)
	}
	if !publishResp.GetSuccess() || publishResp.GetRelease().GetProdUrl() == "" {
		t.Errorf("expected release with prod_url, got: %+v", publishResp)
	}

	// 6. Получение опубликованного релиза
	pubResp, err := projClient.GetPublished(userCtx, &pb.ProjectGetPublishedRequest{
		Id: projectID,
	})
	if err != nil {
		t.Fatalf("GetPublished failed: %v", err)
	}
	if pubResp.GetRelease().GetVersion() != "1.0.0" {
		t.Errorf("expected release version 1.0.0, got: %s", pubResp.GetRelease().GetVersion())
	}

	// 7. Снятие с публикации
	unpubResp, err := projClient.Unpublish(userCtx, &pb.ProjectUnpublishRequest{
		Id: projectID,
	})
	if err != nil {
		t.Fatalf("Unpublish failed: %v", err)
	}
	if !unpubResp.GetSuccess() {
		t.Errorf("expected unpublish success")
	}
}
