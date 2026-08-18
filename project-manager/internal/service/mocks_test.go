package service

import (
	"context"
	"fmt"
	"io"
	"sync"
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

type mockModerationRepo struct {
	mu      sync.Mutex
	tickets map[int64]*domain.ModerationTicket
	nextID  int64
}

func newMockModerationRepo() *mockModerationRepo {
	return &mockModerationRepo{
		tickets: make(map[int64]*domain.ModerationTicket),
		nextID:  1,
	}
}

func (m *mockModerationRepo) CreateTicket(ctx context.Context, t *domain.ModerationTicket) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t.ID = m.nextID
	t.SubmittedAt = time.Now()
	m.tickets[t.ID] = t
	m.nextID++
	return t.ID, nil
}

func (m *mockModerationRepo) GetTicket(ctx context.Context, id int64) (*domain.ModerationTicket, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tickets[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return t, nil
}

func (m *mockModerationRepo) GetLatestTicketByProject(ctx context.Context, projectID int64) (*domain.ModerationTicket, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var latest *domain.ModerationTicket
	for _, t := range m.tickets {
		if t.ProjectID == projectID {
			if latest == nil || t.SubmittedAt.After(latest.SubmittedAt) {
				latest = t
			}
		}
	}
	if latest == nil {
		return nil, domain.ErrNotFound
	}
	return latest, nil
}

func (m *mockModerationRepo) ListTickets(ctx context.Context, status *domain.ModerationStatus, limit, offset int) ([]*domain.ModerationTicket, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var res []*domain.ModerationTicket
	for _, t := range m.tickets {
		if status == nil || t.Status == *status {
			res = append(res, t)
		}
	}
	return res, nil
}

func (m *mockModerationRepo) ResolveTicket(ctx context.Context, id int64, status domain.ModerationStatus, rejectionReason, moderatorID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tickets[id]
	if !ok {
		return domain.ErrNotFound
	}
	t.Status = status
	t.RejectionReason = rejectionReason
	t.ModeratorID = moderatorID
	now := time.Now()
	t.ResolvedAt = &now
	return nil
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
