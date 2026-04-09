// Package memory предоставляет in-memory хранилище для инстансов.
package memory

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"sync"

	"github.com/Be4Die/game-developer-hub/game-server-node/internal/domain"
)

// Storage хранит инстансы в памяти.
// Безопасна для конкурентного использования.
type Storage struct {
	data  map[int64]domain.Instance
	mutex sync.RWMutex
}

// NewStorage создаёт хранилище с начальной ёмкостью 10 элементов.
func NewStorage() *Storage {
	return &Storage{
		data: make(map[int64]domain.Instance, 10),
	}
}

// GetInstanceByID возвращает инстанс по ID. Возвращает ErrNotFound при отсутствии.
func (s *Storage) GetInstanceByID(ctx context.Context, id int64) (*domain.Instance, error) { //nolint:revive // ctx required by interface
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	instance, ok := s.data[id]
	if !ok {
		return nil, fmt.Errorf("instance with id %d: %w", id, domain.ErrNotFound)
	}
	return &instance, nil
}

// GetInstancesByGameID возвращает все инстансы указанной игры.
func (s *Storage) GetInstancesByGameID(ctx context.Context, gameID int64) ([]domain.Instance, error) { //nolint:revive // ctx required by interface
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	instances := make([]domain.Instance, 0)
	for _, v := range s.data {
		if v.GameID == gameID {
			instances = append(instances, v)
		}
	}
	return instances, nil
}

// GetInstanceByContainerID возвращает инстанс по containerID. Возвращает ErrNotFound при отсутствии.
func (s *Storage) GetInstanceByContainerID(ctx context.Context, containerID string) (*domain.Instance, error) { //nolint:revive // ctx required by interface
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, v := range s.data {
		if v.ContainerID == containerID {
			return &v, nil
		}
	}
	return nil, fmt.Errorf("instance with container_id %s: %w", containerID, domain.ErrNotFound)
}

// GetAllInstances возвращает копию всех зарегистрированных инстансов.
func (s *Storage) GetAllInstances(ctx context.Context) ([]domain.Instance, error) { //nolint:revive // ctx required by interface
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return slices.Collect(maps.Values(s.data)), nil
}

// RecordInstance сохраняет или обновляет данные инстанса.
func (s *Storage) RecordInstance(ctx context.Context, instance domain.Instance) error { //nolint:revive // ctx required by interface
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.data[instance.ID] = instance
	return nil
}

// DeleteInstance удаляет инстанс по ID. Возвращает ErrNotFound при отсутствии.
func (s *Storage) DeleteInstance(ctx context.Context, id int64) error { //nolint:revive // ctx required by interface
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.data[id]; !ok {
		return fmt.Errorf("instance with id %d: %w", id, domain.ErrNotFound)
	}
	delete(s.data, id)
	return nil
}
