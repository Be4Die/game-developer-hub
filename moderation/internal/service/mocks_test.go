package service

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Be4Die/game-developer-hub/moderation/internal/domain"
)

type mockRequestRepo struct {
	mu       sync.RWMutex
	requests map[int64]*domain.ModerationRequest
	nextID   int64
}

func newMockRequestRepo() *mockRequestRepo {
	return &mockRequestRepo{
		requests: make(map[int64]*domain.ModerationRequest),
	}
}

func (m *mockRequestRepo) Create(ctx context.Context, req *domain.ModerationRequest) (int64, error) {
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

func (m *mockRequestRepo) Get(ctx context.Context, id int64) (*domain.ModerationRequest, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	req, ok := m.requests[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	copied := *req
	return &copied, nil
}

func (m *mockRequestRepo) GetLatestByProject(ctx context.Context, projectID int64) (*domain.ModerationRequest, error) {
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

func (m *mockRequestRepo) List(ctx context.Context, filter domain.RequestFilter) ([]*domain.ModerationRequest, int, error) {
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
	offset := filter.Offset
	if offset > total {
		offset = total
	}
	end := offset + filter.Limit
	if filter.Limit <= 0 || end > total {
		end = total
	}

	return list[offset:end], total, nil
}

func (m *mockRequestRepo) Claim(ctx context.Context, id int64, moderatorID string) error {
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
	if req.StartedReviewAt == nil {
		req.StartedReviewAt = &now
	}
	req.UpdatedAt = now
	return nil
}

func (m *mockRequestRepo) Resolve(ctx context.Context, id int64, status domain.RequestStatus, reason, moderatorID string) error {
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

type mockMessageRepo struct {
	mu       sync.RWMutex
	messages map[int64][]*domain.ChatMessage
	nextID   int64
}

func newMockMessageRepo() *mockMessageRepo {
	return &mockMessageRepo{
		messages: make(map[int64][]*domain.ChatMessage),
	}
}

func (m *mockMessageRepo) Create(ctx context.Context, msg *domain.ChatMessage) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := atomic.AddInt64(&m.nextID, 1)
	msg.ID = id
	msg.CreatedAt = time.Now()

	copied := *msg
	m.messages[msg.ProjectID] = append(m.messages[msg.ProjectID], &copied)
	return id, nil
}

func (m *mockMessageRepo) ListByProject(ctx context.Context, projectID int64, limit, offset int) ([]*domain.ChatMessage, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	all := m.messages[projectID]
	total := len(all)

	if offset > total {
		offset = total
	}
	end := offset + limit
	if limit <= 0 || end > total {
		end = total
	}

	return all[offset:end], total, nil
}

type mockProjectClient struct {
	mu           sync.RWMutex
	deployFails  bool
	rejectedMods map[int64]string
	releases     map[int64]string
}

func newMockProjectClient() *mockProjectClient {
	return &mockProjectClient{
		rejectedMods: make(map[int64]string),
		releases:     make(map[int64]string),
	}
}

func (m *mockProjectClient) PublishRelease(ctx context.Context, projectID int64, version, approvedBy, comment string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.deployFails {
		return "", domain.ErrDeployFailed
	}

	url := fmt.Sprintf("/games/%d/prod/index.html", projectID)
	m.releases[projectID] = url
	return url, nil
}

func (m *mockProjectClient) RejectDraft(ctx context.Context, projectID int64, reason string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.rejectedMods[projectID] = reason
	return nil
}
