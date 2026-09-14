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

func (m *mockRequestRepo) GetLatestByProjectAndType(ctx context.Context, projectID int64, reqType domain.RequestType) (*domain.ModerationRequest, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var latest *domain.ModerationRequest
	for _, req := range m.requests {
		if req.ProjectID == projectID && req.Type == reqType {
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
		if filter.Type != nil && req.Type != *filter.Type {
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

func (m *mockRequestRepo) ResolveServerAccess(
	ctx context.Context,
	id int64,
	status domain.RequestStatus,
	maxInstances int32,
	maxTotalCPU uint32,
	maxTotalMemoryMB uint64,
	maxInstanceCPU uint32,
	maxInstanceMemoryMB uint64,
	moderatorComment string,
	rejectionReason string,
	moderatorID string,
) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	req, ok := m.requests[id]
	if !ok {
		return domain.ErrNotFound
	}

	req.Status = status
	req.MaxInstances = maxInstances
	req.MaxTotalCPUMillis = maxTotalCPU
	req.MaxTotalMemoryMB = maxTotalMemoryMB
	req.MaxInstanceCPUMillis = maxInstanceCPU
	req.MaxInstanceMemoryMB = maxInstanceMemoryMB
	req.ModeratorComment = moderatorComment
	req.RejectionReason = rejectionReason
	if req.ModeratorID == "" {
		req.ModeratorID = moderatorID
	}
	now := time.Now()
	req.ResolvedAt = &now
	req.UpdatedAt = now
	return nil
}

func (m *mockRequestRepo) GetModeratorStats(ctx context.Context, moderatorID string) (*domain.ModeratorStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := &domain.ModeratorStats{
		ModeratorID: moderatorID,
	}

	var totalDuration int64
	var durationCount int64

	for _, req := range m.requests {
		if req.ModeratorID != moderatorID {
			continue
		}
		stats.TotalAssigned++
		switch req.Status {
		case domain.RequestStatusInReview:
			stats.InReviewCount++
		case domain.RequestStatusApproved:
			stats.ApprovedCount++
		case domain.RequestStatusRejected:
			stats.RejectedCount++
		}
		if req.Status == domain.RequestStatusApproved || req.Status == domain.RequestStatusRejected {
			stats.TodayResolved++
			stats.WeekResolved++
			stats.MonthResolved++
			if req.StartedReviewAt != nil && req.ResolvedAt != nil {
				dur := req.ResolvedAt.Sub(*req.StartedReviewAt)
				if dur > 0 {
					totalDuration += int64(dur.Seconds())
					durationCount++
				}
			}
		}
	}

	stats.TotalResolved = stats.ApprovedCount + stats.RejectedCount
	if stats.TotalResolved > 0 {
		stats.ApprovalRate = float64(stats.ApprovedCount) / float64(stats.TotalResolved) * 100
		stats.RejectionRate = float64(stats.RejectedCount) / float64(stats.TotalResolved) * 100
	}
	if durationCount > 0 {
		stats.AvgReviewDurationSeconds = totalDuration / durationCount
	}

	return stats, nil
}

func (m *mockRequestRepo) ListModeratorsStats(ctx context.Context) ([]*domain.ModeratorStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	mods := make(map[string]bool)
	for _, req := range m.requests {
		if req.ModeratorID != "" {
			mods[req.ModeratorID] = true
		}
	}

	var result []*domain.ModeratorStats
	for modID := range mods {
		stats, _ := m.GetModeratorStats(ctx, modID)
		result = append(result, stats)
	}

	return result, nil
}

func (m *mockRequestRepo) ListModeratorActivity(ctx context.Context, moderatorID string, actionType string, limit, offset int) ([]*domain.ModeratorActivityItem, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var list []*domain.ModeratorActivityItem
	for _, req := range m.requests {
		if req.ModeratorID != moderatorID {
			continue
		}
		item := &domain.ModeratorActivityItem{
			ID:           req.ID,
			ProjectID:    req.ProjectID,
			ProjectTitle: req.Snapshot.TitleRu,
			CreatedAt:    req.UpdatedAt,
		}
		switch req.Status {
		case domain.RequestStatusApproved:
			item.ActionType = "approved"
			item.ActionTitle = "Одобрил публикацию"
		case domain.RequestStatusRejected:
			item.ActionType = "rejected"
			item.ActionTitle = "Отклонил заявку"
			item.Details = req.RejectionReason
		case domain.RequestStatusInReview:
			item.ActionType = "claimed"
			item.ActionTitle = "Взял на проверку"
		default:
			item.ActionType = "event"
			item.ActionTitle = "Действие"
		}
		if actionType != "" && item.ActionType != actionType {
			continue
		}
		list = append(list, item)
	}

	total := len(list)
	if offset > total {
		offset = total
	}
	end := offset + limit
	if limit <= 0 || end > total {
		end = total
	}

	return list[offset:end], total, nil
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

func (m *mockMessageRepo) ListActiveChats(ctx context.Context, limit, offset int) ([]*domain.ChatSummary, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var summaries []*domain.ChatSummary
	for pid, msgs := range m.messages {
		if len(msgs) == 0 {
			continue
		}
		last := msgs[len(msgs)-1]
		summaries = append(summaries, &domain.ChatSummary{
			ProjectID:   pid,
			LastMessage: last,
		})
	}

	total := len(summaries)
	if offset > total {
		offset = total
	}
	end := offset + limit
	if limit <= 0 || end > total {
		end = total
	}

	return summaries[offset:end], total, nil
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

type mockAttachmentRepo struct {
	mu          sync.RWMutex
	attachments map[string]*domain.Attachment
}

func newMockAttachmentRepo() *mockAttachmentRepo {
	return &mockAttachmentRepo{
		attachments: make(map[string]*domain.Attachment),
	}
}

func (m *mockAttachmentRepo) Create(ctx context.Context, att *domain.Attachment) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	att.CreatedAt = time.Now()
	m.attachments[att.ID] = att
	return nil
}

func (m *mockAttachmentRepo) Get(ctx context.Context, id string, projectID int64) (*domain.Attachment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	att, ok := m.attachments[id]
	if !ok || att.ProjectID != projectID {
		return nil, domain.ErrNotFound
	}
	return att, nil
}

func (m *mockAttachmentRepo) GetByIDs(ctx context.Context, ids []string) ([]*domain.Attachment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*domain.Attachment
	for _, id := range ids {
		if a, ok := m.attachments[id]; ok {
			result = append(result, a)
		}
	}
	return result, nil
}

func (m *mockAttachmentRepo) BindToMessage(ctx context.Context, attachmentIDs []string, messageID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, id := range attachmentIDs {
		if a, ok := m.attachments[id]; ok {
			a.MessageID = &messageID
		}
	}
	return nil
}

func (m *mockAttachmentRepo) ListByMessageIDs(ctx context.Context, messageIDs []int64) (map[int64][]*domain.Attachment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make(map[int64][]*domain.Attachment)
	idMap := make(map[int64]bool)
	for _, id := range messageIDs {
		idMap[id] = true
	}
	for _, a := range m.attachments {
		if a.MessageID != nil && idMap[*a.MessageID] {
			result[*a.MessageID] = append(result[*a.MessageID], a)
		}
	}
	return result, nil
}

func (m *mockAttachmentRepo) PurgeByProjectID(ctx context.Context, projectID int64) (int, []string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var paths []string
	count := 0
	for _, a := range m.attachments {
		if a.ProjectID == projectID && !a.IsPurged {
			a.IsPurged = true
			if a.StoragePath != "" {
				paths = append(paths, a.StoragePath)
			}
			count++
		}
	}
	return count, paths, nil
}

type mockSnapshotRepo struct {
	mu        sync.RWMutex
	snapshots map[int64]*domain.ModerationSnapshot
}

func newMockSnapshotRepo() *mockSnapshotRepo {
	return &mockSnapshotRepo{
		snapshots: make(map[int64]*domain.ModerationSnapshot),
	}
}

func (m *mockSnapshotRepo) Save(ctx context.Context, snapshot *domain.ModerationSnapshot) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.snapshots[snapshot.RequestID] = snapshot
	return nil
}

func (m *mockSnapshotRepo) GetByRequestID(ctx context.Context, requestID int64) (*domain.ModerationSnapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.snapshots[requestID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return s, nil
}

func (m *mockSnapshotRepo) GetByProjectID(ctx context.Context, projectID int64) (*domain.ModerationSnapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, s := range m.snapshots {
		if s.ProjectID == projectID {
			return s, nil
		}
	}
	return nil, domain.ErrNotFound
}
