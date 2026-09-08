package integration

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
)

type mockProjectRepo struct {
	mu       sync.Mutex
	projects map[int64]*domain.Project
	nextID   int64
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
		}
	}
	return res, nil
}

func (m *mockProjectRepo) CountForUser(ctx context.Context, userID string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	count := 0
	for _, p := range m.projects {
		if p.OwnerID == userID {
			count++
		}
	}
	return count, nil
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
