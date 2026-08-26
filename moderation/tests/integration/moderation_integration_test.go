package integration

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/moderation/internal/domain"
	"github.com/Be4Die/game-developer-hub/moderation/internal/service"
	grpctransport "github.com/Be4Die/game-developer-hub/moderation/internal/transport/grpc"
	pb "github.com/Be4Die/game-developer-hub/protos/moderation/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

type inMemoryRequestRepo struct {
	mu       sync.RWMutex
	requests map[int64]*domain.ModerationRequest
	nextID   int64
}

func newInMemoryRequestRepo() *inMemoryRequestRepo {
	return &inMemoryRequestRepo{
		requests: make(map[int64]*domain.ModerationRequest),
	}
}

func (m *inMemoryRequestRepo) Create(ctx context.Context, req *domain.ModerationRequest) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := atomic.AddInt64(&m.nextID, 1)
	req.ID = id
	req.SubmittedAt = time.Now()
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()

	copied := *req
	m.requests[id] = &copied
	return id, nil
}

func (m *inMemoryRequestRepo) Get(ctx context.Context, id int64) (*domain.ModerationRequest, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	req, ok := m.requests[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	copied := *req
	return &copied, nil
}

func (m *inMemoryRequestRepo) GetLatestByProject(ctx context.Context, projectID int64) (*domain.ModerationRequest, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var latest *domain.ModerationRequest
	for _, req := range m.requests {
		if req.ProjectID == projectID {
			if latest == nil || req.SubmittedAt.After(latest.SubmittedAt) {
				latest = req
			}
		}
	}
	if latest == nil {
		return nil, domain.ErrNotFound
	}
	copied := *latest
	return &copied, nil
}

func (m *inMemoryRequestRepo) List(ctx context.Context, filter domain.RequestFilter) ([]*domain.ModerationRequest, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var list []*domain.ModerationRequest
	for _, req := range m.requests {
		if filter.Status != nil && req.Status != *filter.Status {
			continue
		}
		if filter.ModeratorID != nil && req.ModeratorID != *filter.ModeratorID {
			continue
		}
		copied := *req
		list = append(list, &copied)
	}

	total := len(list)
	return list, total, nil
}

func (m *inMemoryRequestRepo) Claim(ctx context.Context, id int64, moderatorID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	req, ok := m.requests[id]
	if !ok {
		return domain.ErrNotFound
	}
	if req.ModeratorID != "" && req.ModeratorID != moderatorID {
		return domain.ErrAlreadyClaimed
	}

	req.ModeratorID = moderatorID
	req.Status = domain.RequestStatusInReview
	now := time.Now()
	req.StartedReviewAt = &now
	req.UpdatedAt = now
	return nil
}

func (m *inMemoryRequestRepo) Resolve(ctx context.Context, id int64, status domain.RequestStatus, reason, moderatorID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	req, ok := m.requests[id]
	if !ok {
		return domain.ErrNotFound
	}

	req.Status = status
	req.RejectionReason = reason
	if req.ModeratorID == "" {
		req.ModeratorID = moderatorID
	}
	now := time.Now()
	req.ResolvedAt = &now
	req.UpdatedAt = now
	return nil
}

type inMemoryMessageRepo struct {
	mu       sync.RWMutex
	messages map[int64][]*domain.ChatMessage
	nextID   int64
}

func newInMemoryMessageRepo() *inMemoryMessageRepo {
	return &inMemoryMessageRepo{
		messages: make(map[int64][]*domain.ChatMessage),
	}
}

func (m *inMemoryMessageRepo) Create(ctx context.Context, msg *domain.ChatMessage) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := atomic.AddInt64(&m.nextID, 1)
	msg.ID = id
	msg.CreatedAt = time.Now()

	copied := *msg
	m.messages[msg.ProjectID] = append(m.messages[msg.ProjectID], &copied)
	return id, nil
}

func (m *inMemoryMessageRepo) ListByProject(ctx context.Context, projectID int64, limit, offset int) ([]*domain.ChatMessage, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	all := m.messages[projectID]
	total := len(all)
	return all, total, nil
}

func (m *inMemoryMessageRepo) ListActiveChats(ctx context.Context, limit, offset int) ([]*domain.ChatSummary, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return nil, 0, nil
}

type inMemoryProjectClient struct {
	mu       sync.RWMutex
	releases map[int64]string
}

func newInMemoryProjectClient() *inMemoryProjectClient {
	return &inMemoryProjectClient{
		releases: make(map[int64]string),
	}
}

func (c *inMemoryProjectClient) PublishRelease(ctx context.Context, projectID int64, version, approvedBy, comment string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	prodURL := fmt.Sprintf("/games/%d/prod/index.html", projectID)
	c.releases[projectID] = prodURL
	return prodURL, nil
}

func (c *inMemoryProjectClient) RejectDraft(ctx context.Context, projectID int64, reason string) error {
	return nil
}

func setupIntegrationServer(t *testing.T) (pb.ModerationServiceClient, func()) {
	t.Helper()

	lis := bufconn.Listen(bufSize)

	reqRepo := newInMemoryRequestRepo()
	msgRepo := newInMemoryMessageRepo()
	pmClient := newInMemoryProjectClient()

	svc := service.NewModerationService(reqRepo, msgRepo, pmClient)
	handler := grpctransport.NewModerationHandler(svc)

	s := grpc.NewServer()
	pb.RegisterModerationServiceServer(s, handler)

	go func() {
		if err := s.Serve(lis); err != nil {
			// ignore on stop
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

	return pb.NewModerationServiceClient(conn), cleanup
}

func TestIntegration_ModerationFullWorkflow(t *testing.T) {
	client, cleanup := setupIntegrationServer(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	devCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs("x-user-id", "dev-user-777", "x-user-role", "developer"))
	modCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs("x-user-id", "mod-user-888", "x-user-role", "moderator"))

	projectID := int64(42)

	// 1. Разработчик отправляет черновик на модерацию
	submitResp, err := client.SubmitDraft(devCtx, &pb.SubmitDraftRequest{
		ProjectId: projectID,
		OwnerId:   "dev-user-777",
		Snapshot: &pb.ProjectSnapshot{
			ProjectId:          projectID,
			TitleRu:            "Киберпанк 2099",
			TitleEn:            "Cyberpunk 2099",
			AboutRu:            "Крутая игра в браузере",
			AboutEn:            "Cool browser game",
			ActiveBuildVersion: "1.0.0",
			DevUrl:             "/games/42/dev/index.html",
		},
	})
	if err != nil {
		t.Fatalf("SubmitDraft failed: %v", err)
	}
	if !submitResp.GetSuccess() || submitResp.GetRequest().GetId() == 0 {
		t.Fatalf("expected submit success with ID, got: %+v", submitResp)
	}
	requestID := submitResp.GetRequest().GetId()

	// 2. Модератор просматривает список запросов
	listResp, err := client.ListRequests(modCtx, &pb.ListModerationRequestsRequest{
		Status: pb.RequestStatus_REQUEST_STATUS_PENDING,
	})
	if err != nil {
		t.Fatalf("ListRequests failed: %v", err)
	}
	if len(listResp.GetRequests()) == 0 {
		t.Fatalf("expected at least 1 pending request")
	}

	// 3. Модератор берет заявку в работу (Claim)
	claimResp, err := client.ClaimRequest(modCtx, &pb.ClaimModerationRequestRequest{
		RequestId: requestID,
	})
	if err != nil {
		t.Fatalf("ClaimRequest failed: %v", err)
	}
	if claimResp.GetRequest().GetStatus() != pb.RequestStatus_REQUEST_STATUS_IN_REVIEW {
		t.Errorf("expected status IN_REVIEW, got: %v", claimResp.GetRequest().GetStatus())
	}
	if claimResp.GetRequest().GetModeratorId() != "mod-user-888" {
		t.Errorf("expected moderator_id mod-user-888, got: %s", claimResp.GetRequest().GetModeratorId())
	}

	// 4. Общение в чате проекта
	// Разработчик отправляет сообщение
	_, err = client.SendMessage(devCtx, &pb.SendChatMessageRequest{
		ProjectId: projectID,
		Content:   "Привет! Подскажите, когда проверите?",
	})
	if err != nil {
		t.Fatalf("SendMessage developer failed: %v", err)
	}

	// Модератор отвечает
	_, err = client.SendMessage(modCtx, &pb.SendChatMessageRequest{
		ProjectId: projectID,
		Content:   "Здравствуйте! Сейчас запускаю сборку на тестовом стенде.",
	})
	if err != nil {
		t.Fatalf("SendMessage moderator failed: %v", err)
	}

	// Проверка истории сообщений
	listMsgsResp, err := client.ListMessages(devCtx, &pb.ListChatMessagesRequest{
		ProjectId: projectID,
	})
	if err != nil {
		t.Fatalf("ListMessages failed: %v", err)
	}
	// Сообщения: 1 (submitted) + 1 (status_changed) + 1 (dev) + 1 (mod) = 4
	if len(listMsgsResp.GetMessages()) != 4 {
		t.Errorf("expected 4 chat messages, got: %d", len(listMsgsResp.GetMessages()))
	}

	// 5. Модератор одобряет проект
	approveResp, err := client.Approve(modCtx, &pb.ApproveModerationRequest{
		ProjectId: projectID,
		Comment:   "Игра протестирована, багов нет, публикация подтверждена.",
	})
	if err != nil {
		t.Fatalf("Approve failed: %v", err)
	}
	if !approveResp.GetSuccess() {
		t.Fatalf("expected approve success")
	}
	if approveResp.GetProdUrl() != "/games/42/prod/index.html" {
		t.Errorf("expected prod_url /games/42/prod/index.html, got: %s", approveResp.GetProdUrl())
	}
	if approveResp.GetRequest().GetStatus() != pb.RequestStatus_REQUEST_STATUS_APPROVED {
		t.Errorf("expected status APPROVED, got: %v", approveResp.GetRequest().GetStatus())
	}

	// 6. Проверка статуса последней заявки
	latestResp, err := client.GetLatestRequestByProject(devCtx, &pb.GetLatestRequestByProjectRequest{
		ProjectId: projectID,
	})
	if err != nil {
		t.Fatalf("GetLatestRequestByProject failed: %v", err)
	}
	if latestResp.GetRequest().GetStatus() != pb.RequestStatus_REQUEST_STATUS_APPROVED {
		t.Errorf("expected latest status APPROVED, got: %v", latestResp.GetRequest().GetStatus())
	}
}
