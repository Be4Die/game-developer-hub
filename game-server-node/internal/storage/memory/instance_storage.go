package memory

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"sync"

	"github.com/Be4Die/game-developer-hub/game-server-node/internal/domain"
)

// MemoryInstanceStorage хранит инстансы в памяти.
// Безопасна для конкурентного использования.
type MemoryInstanceStorage struct {
	data  map[int64]domain.Instance
	mutex sync.RWMutex
}

// NewMemoryInstanceStorage создаёт хранилище с начальной ёмкостью 10 элементов.
func NewMemoryInstanceStorage() *MemoryInstanceStorage {
	return &MemoryInstanceStorage{
		data: make(map[int64]domain.Instance, 10),
	}
}

// GetInstanceByID возвращает инстанс по ID. Возвращает ErrNotFound при отсутствии.
func (m *MemoryInstanceStorage) GetInstanceByID(ctx context.Context, id int64) (*domain.Instance, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	instance, ok := m.data[id]
	if !ok {
		return nil, fmt.Errorf("instance with id %d: %w", id, domain.ErrNotFound)
	}
	return &instance, nil
}

// GetInstancesByGameID возвращает все инстансы указанной игры.
func (m *MemoryInstanceStorage) GetInstancesByGameID(ctx context.Context, gameID int64) ([]domain.Instance, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	instances := make([]domain.Instance, 0)
	for _, v := range m.data {
		if v.GameID == gameID {
			instances = append(instances, v)
		}
	}
	return instances, nil
}

// GetInstanceByContainerID возвращает инстанс по containerID. Возвращает ErrNotFound при отсутствии.
func (m *MemoryInstanceStorage) GetInstanceByContainerID(ctx context.Context, containerID string) (*domain.Instance, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for _, v := range m.data {
		if v.ContainerID == containerID {
			return &v, nil
		}
	}
	return nil, fmt.Errorf("instance with container_id %s: %w", containerID, domain.ErrNotFound)
}

// GetAllInstances возвращает копию всех зарегистрированных инстансов.
func (m *MemoryInstanceStorage) GetAllInstances(ctx context.Context) ([]domain.Instance, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return slices.Collect(maps.Values(m.data)), nil
}

// RecordInstance сохраняет или обновляет данные инстанса.
func (m *MemoryInstanceStorage) RecordInstance(ctx context.Context, instance domain.Instance) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.data[instance.ID] = instance
	return nil
}

// DeleteInstance удаляет инстанс по ID. Возвращает ErrNotFound при отсутствии.
func (m *MemoryInstanceStorage) DeleteInstance(ctx context.Context, id int64) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, ok := m.data[id]; !ok {
		return fmt.Errorf("instance with id %d: %w", id, domain.ErrNotFound)
	}
	delete(m.data, id)
	return nil
}
