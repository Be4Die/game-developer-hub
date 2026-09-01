package integration

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	modclient "github.com/Be4Die/game-developer-hub/moderation/internal/infrastructure/client"
	modservice "github.com/Be4Die/game-developer-hub/moderation/internal/service"
	modgrpc "github.com/Be4Die/game-developer-hub/moderation/internal/transport/grpc"
	modpb "github.com/Be4Die/game-developer-hub/protos/moderation/v1"
	pmpb "github.com/Be4Die/game-developer-hub/protos/project_manager/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
)

// fakeProjectServer реализует pmpb.ProjectServiceServer для эмуляции поведения реального project-manager.
type fakeProjectServer struct {
	pmpb.UnimplementedProjectServiceServer
	mu             sync.Mutex
	publishedGames map[int64]*pmpb.ProjectRelease
	rejectedGames  map[int64]string
}

func newFakeProjectServer() *fakeProjectServer {
	return &fakeProjectServer{
		publishedGames: make(map[int64]*pmpb.ProjectRelease),
		rejectedGames:  make(map[int64]string),
	}
}

func (s *fakeProjectServer) PublishRelease(ctx context.Context, req *pmpb.ProjectPublishReleaseRequest) (*pmpb.ProjectPublishReleaseResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rel := &pmpb.ProjectRelease{
		ProjectId:   req.GetProjectId(),
		Version:     req.GetVersion(),
		PublishedAt: time.Now().Format(time.RFC3339),
		ProdUrl:     "/games/" + req.GetVersion() + "/prod/index.html",
	}
	s.publishedGames[req.GetProjectId()] = rel

	return &pmpb.ProjectPublishReleaseResponse{
		Success: true,
		Release: rel,
	}, nil
}

func (s *fakeProjectServer) RejectDraft(ctx context.Context, req *pmpb.ProjectRejectDraftRequest) (*pmpb.ProjectRejectDraftResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.rejectedGames[req.GetProjectId()] = req.GetReason()
	return &pmpb.ProjectRejectDraftResponse{Success: true}, nil
}

func TestCrossService_ModerationWithProjectManager(t *testing.T) {
	// 1. Запуск fake Project Manager gRPC сервера
	pmLis := bufconn.Listen(bufSize)
	fakePM := newFakeProjectServer()

	pmServer := grpc.NewServer()
	pmpb.RegisterProjectServiceServer(pmServer, fakePM)

	go func() {
		_ = pmServer.Serve(pmLis)
	}()
	defer pmServer.Stop()

	// 2. Создание gRPC клиента к Project Manager
	pmConn, err := grpc.NewClient(
		"passthrough://bufnet-pm",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return pmLis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("dial PM: %v", err)
	}
	defer func() { _ = pmConn.Close() }()

	pmClientAdapter := &grpcPMAdapter{client: pmpb.NewProjectServiceClient(pmConn)}

	// 3. Запуск Moderation Service gRPC сервера с реальным PM клиентом
	modLis := bufconn.Listen(bufSize)
	reqRepo := newInMemoryRequestRepo()
	msgRepo := newInMemoryMessageRepo()

	modSvc := modservice.NewModerationService(reqRepo, msgRepo, pmClientAdapter)
	modHandler := modgrpc.NewModerationHandler(modSvc)

	modServer := grpc.NewServer()
	modpb.RegisterModerationServiceServer(modServer, modHandler)

	go func() {
		_ = modServer.Serve(modLis)
	}()
	defer modServer.Stop()

	// 4. Клиент к Moderation Service
	modConn, err := grpc.NewClient(
		"passthrough://bufnet-mod",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return modLis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("dial Moderation: %v", err)
	}
	defer func() { _ = modConn.Close() }()

	modClient := modpb.NewModerationServiceClient(modConn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	devCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs("x-user-id", "developer-101", "x-user-role", "developer"))
	modCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs("x-user-id", "moderator-202", "x-user-role", "moderator"))

	projectID := int64(1001)

	// ─── Сценарий 1: Отправка черновика на модерацию ──────────────
	submitResp, err := modClient.SubmitDraft(devCtx, &modpb.SubmitDraftRequest{
		ProjectId: projectID,
		OwnerId:   "developer-101",
		Snapshot: &modpb.ProjectSnapshot{
			ProjectId:          projectID,
			TitleRu:            "Супер РПГ",
			TitleEn:            "Super RPG",
			AboutRu:            "Браузерная ролевая игра нового поколения.",
			AboutEn:            "Next-gen browser RPG.",
			ActiveBuildVersion: "2.1.0",
			DevUrl:             "/games/1001/dev/index.html",
		},
	})
	if err != nil {
		t.Fatalf("SubmitDraft failed: %v", err)
	}
	requestID := submitResp.GetRequest().GetId()

	// ─── Сценарий 2: Модератор берет заявку в работу ─────────────
	claimResp, err := modClient.ClaimRequest(modCtx, &modpb.ClaimModerationRequestRequest{
		RequestId: requestID,
	})
	if err != nil {
		t.Fatalf("ClaimRequest failed: %v", err)
	}
	if claimResp.GetRequest().GetStatus() != modpb.RequestStatus_REQUEST_STATUS_IN_REVIEW {
		t.Errorf("expected status IN_REVIEW, got: %v", claimResp.GetRequest().GetStatus())
	}

	// ─── Сценарий 3: Переписка в чате проекта ─────────────────────
	_, err = modClient.SendMessage(devCtx, &modpb.SendChatMessageRequest{
		ProjectId: projectID,
		Content:   "Добрый день! Проверьте, пожалуйста, сохранение прогресса.",
	})
	if err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}

	_, err = modClient.SendMessage(modCtx, &modpb.SendChatMessageRequest{
		ProjectId: projectID,
		Content:   "Здравствуйте! Сохранения проверил, всё работает стабильно.",
	})
	if err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}

	// ─── Сценарий 4: Одобрение и вызов Project Manager по gRPC ───
	approveResp, err := modClient.Approve(modCtx, &modpb.ApproveModerationRequest{
		ProjectId: projectID,
		Comment:   "Игра полностью соответствует требованиям площадки.",
	})
	if err != nil {
		t.Fatalf("Approve failed: %v", err)
	}
	if !approveResp.GetSuccess() || approveResp.GetProdUrl() == "" {
		t.Errorf("expected approve success with prod url: %+v", approveResp)
	}

	// Проверяем, что в Project Manager пришел gRPC вызов и зафиксирован релиз
	fakePM.mu.Lock()
	rel, ok := fakePM.publishedGames[projectID]
	fakePM.mu.Unlock()

	if !ok || rel.GetVersion() != "2.1.0" {
		t.Fatalf("expected project 1001 to be published in Project Manager with version 2.1.0, got: %+v", rel)
	}
}

type grpcPMAdapter struct {
	client pmpb.ProjectServiceClient
}

func (a *grpcPMAdapter) PublishRelease(ctx context.Context, projectID int64, version, approvedBy, comment string) (string, error) {
	resp, err := a.client.PublishRelease(ctx, &pmpb.ProjectPublishReleaseRequest{
		ProjectId:   projectID,
		Version:     version,
		PublishedBy: approvedBy,
		Comment:     comment,
	})
	if err != nil {
		return "", err
	}
	if resp.GetRelease() != nil {
		return resp.GetRelease().GetProdUrl(), nil
	}
	return "", nil
}

func (a *grpcPMAdapter) RejectDraft(ctx context.Context, projectID int64, reason string) error {
	_, err := a.client.RejectDraft(ctx, &pmpb.ProjectRejectDraftRequest{
		ProjectId: projectID,
		Reason:    reason,
	})
	return err
}

var _ = modclient.NewProjectStubClient
