package client

import (
	"context"
	"fmt"
	"sync"

	"github.com/Be4Die/game-developer-hub/moderation/internal/domain"
)

// ProjectStubClient реализует in-memory заглушку клиента project-manager для локального запуска и тестов.
type ProjectStubClient struct {
	mu           sync.RWMutex
	releases     map[int64]string
	deployFails  bool
	rejectedMods map[int64]string
}

// NewProjectStubClient создаёт новый экземпляр ProjectStubClient.
func NewProjectStubClient() *ProjectStubClient {
	return &ProjectStubClient{
		releases:     make(map[int64]string),
		rejectedMods: make(map[int64]string),
	}
}

// SetDeployFails управляет эмуляцией сбоя развёртывания.
func (s *ProjectStubClient) SetDeployFails(fail bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deployFails = fail
}

// PublishRelease эмулирует развёртывание и публикацию игры.
func (s *ProjectStubClient) PublishRelease(_ context.Context, projectID int64, _, _, _ string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.deployFails {
		return "", domain.ErrDeployFailed
	}

	prodURL := fmt.Sprintf("/games/%d/prod/index.html", projectID)
	s.releases[projectID] = prodURL
	return prodURL, nil
}

// RejectDraft эмулирует отклонение черновика.
func (s *ProjectStubClient) RejectDraft(_ context.Context, projectID int64, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rejectedMods[projectID] = reason
	return nil
}
