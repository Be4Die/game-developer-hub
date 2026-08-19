// Package client предоставляет реализации клиентов к внешним сервисам для project-manager.
package client

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
)

// StubModerationClient реализует in-memory заглушку для работы и тестирования без запущенного сервиса модерации.
// Потокобезопасен для конкурентного использования.
type StubModerationClient struct {
	mu       sync.RWMutex
	requests map[int64]*domain.ModerationRequestInfo
	nextID   int64
}

// NewStubModerationClient создаёт новый экземпляр StubModerationClient.
func NewStubModerationClient() *StubModerationClient {
	return &StubModerationClient{
		requests: make(map[int64]*domain.ModerationRequestInfo),
	}
}

// SubmitDraft сохраняет заявку в памяти и возвращает сгенерированный идентификатор.
func (s *StubModerationClient) SubmitDraft(ctx context.Context, snapshot *domain.ProjectSnapshot) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := atomic.AddInt64(&s.nextID, 1)
	s.requests[snapshot.ProjectID] = &domain.ModerationRequestInfo{
		RequestID: id,
		ProjectID: snapshot.ProjectID,
		Status:    1, // Pending
	}
	return id, nil
}

// GetLatestRequest возвращает последнюю заявку по проекту из памяти.
func (s *StubModerationClient) GetLatestRequest(ctx context.Context, projectID int64) (*domain.ModerationRequestInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	req, ok := s.requests[projectID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return req, nil
}
