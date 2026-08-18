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

func setupIntegrationServer(t *testing.T) (pb.ProjectServiceClient, pb.ModerationServiceClient, func()) {
	t.Helper()

	lis := bufconn.Listen(bufSize)
	tmpDir := t.TempDir()

	pRepo := newMockProjectRepo()
	dRepo := newMockDraftRepo()
	bRepo := newMockBuildRepo()
	mRepo := newMockModerationRepo()
	rRepo := newMockReleaseRepo()
	depRepo := newMockDeploymentRepo()

	bStorage := filesystem.NewBuildStorage(filepath.Join(tmpDir, "projects"))
	mStorage := filesystem.NewMediaStorage(filepath.Join(tmpDir, "projects"))
	deployer := deployment.NewLocalDeployer(filepath.Join(tmpDir, "games"), "/games")

	projSvc := service.NewProjectService(
		pRepo, dRepo, bRepo, mRepo, rRepo, depRepo,
		bStorage, mStorage, deployer, nil, 5,
	)

	modSvc := service.NewModerationService(
		pRepo, dRepo, bRepo, mRepo, rRepo, depRepo,
		bStorage, deployer,
	)

	projHandler := grpctransport.NewProjectHandler(projSvc)
	modHandler := grpctransport.NewModerationHandler(modSvc, projSvc)

	s := grpc.NewServer()
	pb.RegisterProjectServiceServer(s, projHandler)
	pb.RegisterModerationServiceServer(s, modHandler)

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

	return pb.NewProjectServiceClient(conn), pb.NewModerationServiceClient(conn), cleanup
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

	_ = zw.Close()
	return buf.Bytes()
}

func withAuth(userID string) context.Context {
	md := metadata.Pairs("x-user-id", userID)
	return metadata.NewOutgoingContext(context.Background(), md)
}

func TestIntegration_FullPublishingLifecycle(t *testing.T) {
	t.Parallel()

	projClient, modClient, cleanup := setupIntegrationServer(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userCtx := withAuth("dev-user-007")
	modCtx := withAuth("mod-user-999")

	// 1. Создание проекта
	createResp, err := projClient.Create(userCtx, &pb.ProjectCreateRequest{
		TitleRu: "Космический Шутер",
		TitleEn: "Space Shooter",
	})
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}
	projectID := createResp.GetProject().GetId()
	if projectID == 0 {
		t.Fatalf("expected non-zero project id")
	}

	// 2. Обновление метаданных черновика
	updateResp, err := projClient.Update(userCtx, &pb.ProjectUpdateRequest{
		Id:      projectID,
		TitleRu: "Космический Шутер HD",
		TitleEn: "Space Shooter HD",
		About:   "Увлекательная аркадная игра в браузере",
		SeoRu:   "космос, аркада, игра",
		SeoEn:   "space, arcade, game",
	})
	if err != nil {
		t.Fatalf("Update draft failed: %v", err)
	}
	if updateResp.GetProject().GetDraft().GetTitleRu() != "Космический Шутер HD" {
		t.Errorf("expected TitleRu 'Космический Шутер HD', got: %s", updateResp.GetProject().GetDraft().GetTitleRu())
	}

	// 3. Потоковая загрузка клиентской веб-сборки (64 КБ чанками)
	zipData := createTestGameZip(t)
	stream, err := projClient.UploadBuildStream(userCtx)
	if err != nil {
		t.Fatalf("UploadBuildStream client init failed: %v", err)
	}

	// 3.1 Метаданные
	if err := stream.Send(&pb.ProjectUploadBuildStreamRequest{
		Payload: &pb.ProjectUploadBuildStreamRequest_Metadata{
			Metadata: &pb.BuildUploadStreamMetadata{
				ProjectId: projectID,
				Version:   "1.0.0",
			},
		},
	}); err != nil {
		t.Fatalf("Send metadata failed: %v", err)
	}

	// 3.2 Чанки
	chunkSize := 64 * 1024
	for i := 0; i < len(zipData); i += chunkSize {
		end := i + chunkSize
		if end > len(zipData) {
			end = len(zipData)
		}
		if err := stream.Send(&pb.ProjectUploadBuildStreamRequest{
			Payload: &pb.ProjectUploadBuildStreamRequest_Chunk{
				Chunk: zipData[i:end],
			},
		}); err != nil {
			t.Fatalf("Send chunk failed: %v", err)
		}
	}

	uploadResp, err := stream.CloseAndRecv()
	if err != nil {
		t.Fatalf("UploadBuildStream failed: %v", err)
	}
	if !uploadResp.GetSuccess() {
		t.Fatalf("expected success upload")
	}
	if uploadResp.GetDevUrl() == "" {
		t.Errorf("expected dev_url to be returned")
	}

	// 4. Отправка на модерацию
	submitResp, err := projClient.SubmitForModeration(userCtx, &pb.SubmitForModerationRequest{
		ProjectId: projectID,
	})
	if err != nil {
		t.Fatalf("SubmitForModeration failed: %v", err)
	}
	if submitResp.GetTicket().GetStatus() != pb.ModerationStatus_MODERATION_STATUS_PENDING {
		t.Errorf("expected ticket status PENDING, got: %v", submitResp.GetTicket().GetStatus())
	}

	// 5. Проверка модератором списка тикетов
	listTicketsResp, err := modClient.ListTickets(modCtx, &pb.ListModerationTicketsRequest{
		Status: pb.ModerationStatus_MODERATION_STATUS_PENDING,
	})
	if err != nil {
		t.Fatalf("ListTickets failed: %v", err)
	}
	if len(listTicketsResp.GetTickets()) == 0 {
		t.Fatalf("expected at least 1 ticket in moderation queue")
	}

	// 6. Модератор утверждает проект (Approve)
	approveResp, err := modClient.Approve(modCtx, &pb.ApproveModerationRequest{
		ProjectId: projectID,
		Comment:   "Игра протестирована, всё отлично!",
	})
	if err != nil {
		t.Fatalf("Approve failed: %v", err)
	}
	if !approveResp.GetSuccess() {
		t.Fatalf("expected approve success")
	}
	if approveResp.GetRelease().GetProdUrl() == "" {
		t.Errorf("expected release prod_url to be set")
	}

	// 7. Получение опубликованного релиза
	pubResp, err := projClient.GetPublished(userCtx, &pb.ProjectGetPublishedRequest{
		Id: projectID,
	})
	if err != nil {
		t.Fatalf("GetPublished failed: %v", err)
	}
	if pubResp.GetRelease().GetVersion() != "1.0.0" {
		t.Errorf("expected release version 1.0.0, got: %s", pubResp.GetRelease().GetVersion())
	}

	// 8. Снятие с публикации
	unpubResp, err := projClient.Unpublish(userCtx, &pb.ProjectUnpublishRequest{
		Id: projectID,
	})
	if err != nil {
		t.Fatalf("Unpublish failed: %v", err)
	}
	if !unpubResp.GetSuccess() {
		t.Errorf("expected unpublish success")
	}

	_ = ctx
}
