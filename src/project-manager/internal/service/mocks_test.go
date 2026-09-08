// Package service contains unit tests and mocks for project-manager service layer.
package service

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
)

type mockProjectRepo struct {
	mu         sync.Mutex
	projects   map[int64]*domain.Project
	nextID     int64
	memberRepo domain.MemberRepo
}

func newMockProjectRepo() *mockProjectRepo {
	return &mockProjectRepo{
		projects: make(map[int64]*domain.Project),
		nextID:   1,
	}
}

func (m *mockProjectRepo) Create(ctx context.Context, p *domain.Project) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p.ID = m.nextID
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	m.projects[p.ID] = p
	m.nextID++
	return p.ID, nil
}

func (m *mockProjectRepo) Get(ctx context.Context, id int64) (*domain.Project, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.projects[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return p, nil
}

func (m *mockProjectRepo) ListByOwner(ctx context.Context, ownerID string, limit, offset int) ([]*domain.Project, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var res []*domain.Project
	for _, p := range m.projects {
		if p.OwnerID == ownerID {
			res = append(res, p)
		}
	}
	return res, nil
}

func (m *mockProjectRepo) CountByOwner(ctx context.Context, ownerID string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	count := 0
	for _, p := range m.projects {
		if p.OwnerID == ownerID {
			count++
		}
	}
	return count, nil
}

func (m *mockProjectRepo) ListForUser(ctx context.Context, userID string, limit, offset int) ([]*domain.Project, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var res []*domain.Project
	for _, p := range m.projects {
		if p.OwnerID == userID {
			res = append(res, p)
		} else if m.memberRepo != nil {
			if isMem, _ := m.memberRepo.IsMember(ctx, p.ID, userID); isMem {
				res = append(res, p)
			}
		}
	}
	return res, nil
}

func (m *mockProjectRepo) CountForUser(ctx context.Context, userID string) (int, error) {
	list, err := m.ListForUser(ctx, userID, 0, 0)
	if err != nil {
		return 0, err
	}
	return len(list), nil
}

func (m *mockProjectRepo) UpdateStatus(ctx context.Context, id int64, status domain.ProjectStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.projects[id]
	if !ok {
		return domain.ErrNotFound
	}
	p.Status = status
	return nil
}

func (m *mockProjectRepo) Delete(ctx context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.projects, id)
	return nil
}

type mockDraftRepo struct {
	mu     sync.Mutex
	drafts map[int64]*domain.Draft
}

func newMockDraftRepo() *mockDraftRepo {
	return &mockDraftRepo{drafts: make(map[int64]*domain.Draft)}
}

func (m *mockDraftRepo) Create(ctx context.Context, d *domain.Draft) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.drafts[d.ProjectID] = d
	return nil
}

func (m *mockDraftRepo) Get(ctx context.Context, projectID int64) (*domain.Draft, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	d, ok := m.drafts[projectID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return d, nil
}

func (m *mockDraftRepo) Update(ctx context.Context, d *domain.Draft) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.drafts[d.ProjectID] = d
	return nil
}

func (m *mockDraftRepo) UpdateActiveBuild(ctx context.Context, projectID int64, version, devURL string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	d, ok := m.drafts[projectID]
	if !ok {
		return domain.ErrNotFound
	}
	d.ActiveBuildVersion = version
	d.DevURL = devURL
	return nil
}

func (m *mockDraftRepo) UpdateMedia(ctx context.Context, projectID int64, mediaType, path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	d, ok := m.drafts[projectID]
	if !ok {
		return domain.ErrNotFound
	}
	switch mediaType {
	case "icon":
		d.IconPath = path
	case "cover":
		d.CoverPath = path
	case "video":
		d.VideoPath = path
	}
	return nil
}

type mockBuildRepo struct {
	mu     sync.Mutex
	builds map[int64][]*domain.Build
	nextID int64
}

func newMockBuildRepo() *mockBuildRepo {
	return &mockBuildRepo{
		builds: make(map[int64][]*domain.Build),
		nextID: 1,
	}
}

func (m *mockBuildRepo) Create(ctx context.Context, b *domain.Build) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	b.ID = m.nextID
	m.nextID++
	m.builds[b.ProjectID] = append([]*domain.Build{b}, m.builds[b.ProjectID]...)
	return nil
}

func (m *mockBuildRepo) Get(ctx context.Context, projectID int64, version string) (*domain.Build, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, b := range m.builds[projectID] {
		if b.Version == version {
			return b, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *mockBuildRepo) ListByProject(ctx context.Context, projectID int64, limit int) ([]*domain.Build, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	list := m.builds[projectID]
	if len(list) > limit {
		list = list[:limit]
	}
	return list, nil
}

func (m *mockBuildRepo) MarkUnpacked(ctx context.Context, projectID int64, version, unpackedPath string) error {
	return nil
}

func (m *mockBuildRepo) Delete(ctx context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for pid, list := range m.builds {
		var filtered []*domain.Build
		for _, b := range list {
			if b.ID != id {
				filtered = append(filtered, b)
			}
		}
		m.builds[pid] = filtered
	}
	return nil
}

type mockModerationClient struct {
	mu       sync.RWMutex
	requests map[int64]*domain.ModerationRequestInfo
	nextID   int64
}

func newMockModerationClient() *mockModerationClient {
	return &mockModerationClient{
		requests: make(map[int64]*domain.ModerationRequestInfo),
	}
}

func (m *mockModerationClient) SubmitDraft(ctx context.Context, snapshot *domain.ProjectSnapshot) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	id := atomic.AddInt64(&m.nextID, 1)
	m.requests[snapshot.ProjectID] = &domain.ModerationRequestInfo{
		RequestID: id,
		ProjectID: snapshot.ProjectID,
		Status:    1,
	}
	return id, nil
}

func (m *mockModerationClient) GetLatestRequest(ctx context.Context, projectID int64) (*domain.ModerationRequestInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	req, ok := m.requests[projectID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return req, nil
}

type mockReleaseRepo struct {
	mu       sync.Mutex
	releases map[int64]*domain.Release
	nextID   int64
}

func newMockReleaseRepo() *mockReleaseRepo {
	return &mockReleaseRepo{
		releases: make(map[int64]*domain.Release),
		nextID:   1,
	}
}

func (m *mockReleaseRepo) Create(ctx context.Context, r *domain.Release) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r.ID = m.nextID
	r.PublishedAt = time.Now()
	r.IsActive = true
	m.releases[r.ProjectID] = r
	m.nextID++
	return r.ID, nil
}

func (m *mockReleaseRepo) GetActive(ctx context.Context, projectID int64) (*domain.Release, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.releases[projectID]
	if !ok || !r.IsActive {
		return nil, domain.ErrNotFound
	}
	return r, nil
}

func (m *mockReleaseRepo) ListByProject(ctx context.Context, projectID int64) ([]*domain.Release, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var res []*domain.Release
	if r, ok := m.releases[projectID]; ok {
		res = append(res, r)
	}
	return res, nil
}

func (m *mockReleaseRepo) Deactivate(ctx context.Context, projectID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if r, ok := m.releases[projectID]; ok {
		r.IsActive = false
		now := time.Now()
		r.UnpublishAt = &now
	}
	return nil
}

type mockDeploymentRepo struct {
	mu      sync.Mutex
	records []*domain.DeploymentRecord
}

func newMockDeploymentRepo() *mockDeploymentRepo {
	return &mockDeploymentRepo{}
}

func (m *mockDeploymentRepo) Create(ctx context.Context, d *domain.DeploymentRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	d.DeployedAt = time.Now()
	m.records = append(m.records, d)
	return nil
}

func (m *mockDeploymentRepo) ListByProject(ctx context.Context, projectID int64, limit int) ([]*domain.DeploymentRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.records, nil
}

type mockBuildStorage struct{}

func (s *mockBuildStorage) SaveArchiveStream(ctx context.Context, projectID int64, version string, src io.Reader) (string, int64, error) {
	return fmt.Sprintf("/data/archives/%d/%s.zip", projectID, version), 1024, nil
}

func (s *mockBuildStorage) SaveArchive(ctx context.Context, projectID int64, version string, data []byte) (string, int64, error) {
	return fmt.Sprintf("/data/archives/%d/%s.zip", projectID, version), int64(len(data)), nil
}

func (s *mockBuildStorage) GetArchivePath(projectID int64, version string) string {
	return fmt.Sprintf("/data/archives/%d/%s.zip", projectID, version)
}

func (s *mockBuildStorage) DeleteBuild(projectID int64, version string) error { return nil }
func (s *mockBuildStorage) DeleteProject(projectID int64) error               { return nil }

type mockMediaStorage struct{}

func (s *mockMediaStorage) SaveMediaStream(ctx context.Context, projectID int64, mediaType string, src io.Reader) (string, error) {
	return fmt.Sprintf("/data/media/%d/%s.png", projectID, mediaType), nil
}

func (s *mockMediaStorage) SaveMedia(ctx context.Context, projectID int64, mediaType string, data []byte) (string, error) {
	return fmt.Sprintf("/data/media/%d/%s.png", projectID, mediaType), nil
}

func (s *mockMediaStorage) DeleteMedia(projectID int64, mediaType string) error { return nil }
func (s *mockMediaStorage) SnapshotMediaForRelease(ctx context.Context, projectID int64, version string, srcPath string, mediaType string) (string, error) {
	return fmt.Sprintf("/data/media/%d/releases/%s/%s.png", projectID, version, mediaType), nil
}

type mockDeployer struct{}

func (d *mockDeployer) DeployDev(ctx context.Context, projectID int64, version string, archivePath string) (*domain.DeploymentResult, error) {
	return &domain.DeploymentResult{
		URL:          fmt.Sprintf("/games/%d/dev/index.html", projectID),
		UnpackedPath: fmt.Sprintf("/data/games/%d/versions/%s", projectID, version),
		Success:      true,
	}, nil
}

func (d *mockDeployer) DeployProd(ctx context.Context, projectID int64, version string, archivePath string) (*domain.DeploymentResult, error) {
	return &domain.DeploymentResult{
		URL:          fmt.Sprintf("/games/%d/prod/index.html", projectID),
		UnpackedPath: fmt.Sprintf("/data/games/%d/versions/%s", projectID, version),
		Success:      true,
	}, nil
}

func (d *mockDeployer) UndeployProd(ctx context.Context, projectID int64) error { return nil }
func (d *mockDeployer) DeleteVersion(ctx context.Context, projectID int64, version string) error {
	return nil
}
func (d *mockDeployer) DeleteProject(ctx context.Context, projectID int64) error { return nil }

type mockMemberRepo struct {
	mu      sync.Mutex
	members map[string]*domain.Member
}

func newMockMemberRepo() *mockMemberRepo {
	return &mockMemberRepo{members: make(map[string]*domain.Member)}
}

func (m *mockMemberRepo) Add(ctx context.Context, mem *domain.Member) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := fmt.Sprintf("%d:%s", mem.ProjectID, mem.UserID)
	m.members[key] = mem
	return nil
}

func (m *mockMemberRepo) Get(ctx context.Context, projectID int64, userID string) (*domain.Member, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := fmt.Sprintf("%d:%s", projectID, userID)
	mem, ok := m.members[key]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return mem, nil
}

func (m *mockMemberRepo) ListByProject(ctx context.Context, projectID int64) ([]*domain.Member, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var res []*domain.Member
	for _, mem := range m.members {
		if mem.ProjectID == projectID {
			res = append(res, mem)
		}
	}
	return res, nil
}

func (m *mockMemberRepo) ListByUser(ctx context.Context, userID string) ([]*domain.Member, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var res []*domain.Member
	for _, mem := range m.members {
		if mem.UserID == userID {
			res = append(res, mem)
		}
	}
	return res, nil
}

func (m *mockMemberRepo) UpdatePermissions(ctx context.Context, projectID int64, userID string, perms []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := fmt.Sprintf("%d:%s", projectID, userID)
	mem, ok := m.members[key]
	if !ok {
		return domain.ErrNotFound
	}
	mem.Permissions = perms
	return nil
}

func (m *mockMemberRepo) Delete(ctx context.Context, projectID int64, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := fmt.Sprintf("%d:%s", projectID, userID)
	delete(m.members, key)
	return nil
}

func (m *mockMemberRepo) IsMember(ctx context.Context, projectID int64, userID string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := fmt.Sprintf("%d:%s", projectID, userID)
	_, ok := m.members[key]
	return ok, nil
}

type mockInvitationRepo struct {
	mu          sync.Mutex
	invitations map[int64]*domain.Invitation
	nextID      int64
}

func newMockInvitationRepo() *mockInvitationRepo {
	return &mockInvitationRepo{
		invitations: make(map[int64]*domain.Invitation),
		nextID:      1,
	}
}

func (r *mockInvitationRepo) Create(ctx context.Context, inv *domain.Invitation) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	inv.ID = r.nextID
	r.nextID++
	inv.CreatedAt = time.Now()
	inv.UpdatedAt = time.Now()
	r.invitations[inv.ID] = inv
	return inv.ID, nil
}

func (r *mockInvitationRepo) Get(ctx context.Context, id int64) (*domain.Invitation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	inv, ok := r.invitations[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return inv, nil
}

func (r *mockInvitationRepo) GetPending(ctx context.Context, projectID int64, inviteeID string) (*domain.Invitation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, inv := range r.invitations {
		if inv.ProjectID == projectID && inv.InviteeID == inviteeID && inv.Status == domain.InvitationStatusPending {
			return inv, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *mockInvitationRepo) ListIncoming(ctx context.Context, inviteeID string) ([]*domain.Invitation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var res []*domain.Invitation
	for _, inv := range r.invitations {
		if inv.InviteeID == inviteeID && inv.Status == domain.InvitationStatusPending {
			res = append(res, inv)
		}
	}
	return res, nil
}

func (r *mockInvitationRepo) ListOutgoing(ctx context.Context, inviterID string, projectID int64) ([]*domain.Invitation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var res []*domain.Invitation
	for _, inv := range r.invitations {
		if inv.InviterID == inviterID && (projectID == 0 || inv.ProjectID == projectID) && inv.Status == domain.InvitationStatusPending {
			res = append(res, inv)
		}
	}
	return res, nil
}

func (r *mockInvitationRepo) UpdateStatus(ctx context.Context, id int64, status domain.InvitationStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	inv, ok := r.invitations[id]
	if !ok {
		return domain.ErrNotFound
	}
	inv.Status = status
	inv.UpdatedAt = time.Now()
	return nil
}

func (r *mockInvitationRepo) CancelAllPendingBetween(ctx context.Context, inviterID, inviteeID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, inv := range r.invitations {
		if inv.InviterID == inviterID && inv.InviteeID == inviteeID && inv.Status == domain.InvitationStatusPending {
			inv.Status = domain.InvitationStatusCanceled
			inv.UpdatedAt = time.Now()
		}
	}
	return nil
}

func (r *mockInvitationRepo) IsSystemUser(ctx context.Context, userID, email string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if strings.Contains(email, "moderator") || strings.Contains(email, "admin") ||
		strings.HasPrefix(userID, "mod-") || strings.HasPrefix(userID, "admin-") {
		return true, nil
	}
	return false, nil
}

type mockBlockRepo struct {
	mu     sync.Mutex
	blocks map[string]*domain.UserBlock
}

func newMockBlockRepo() *mockBlockRepo {
	return &mockBlockRepo{blocks: make(map[string]*domain.UserBlock)}
}

func (r *mockBlockRepo) Block(ctx context.Context, b *domain.UserBlock) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := fmt.Sprintf("%s:%s", b.UserID, b.BlockedUserID)
	b.CreatedAt = time.Now()
	r.blocks[key] = b
	return nil
}

func (r *mockBlockRepo) Unblock(ctx context.Context, userID, blockedUserID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := fmt.Sprintf("%s:%s", userID, blockedUserID)
	delete(r.blocks, key)
	return nil
}

func (r *mockBlockRepo) IsBlocked(ctx context.Context, blockerID, targetID string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := fmt.Sprintf("%s:%s", blockerID, targetID)
	_, ok := r.blocks[key]
	return ok, nil
}

func (r *mockBlockRepo) ListBlocked(ctx context.Context, userID string) ([]*domain.UserBlock, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var res []*domain.UserBlock
	for _, b := range r.blocks {
		if b.UserID == userID {
			res = append(res, b)
		}
	}
	return res, nil
}

