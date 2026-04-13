// Package valkey реализует хранилище горячего состояния в Valkey.
// Использует go-redis — Valkey wire-protocol-совместим с Redis OSS 7.x.
package valkey

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
	"github.com/redis/go-redis/v9"
)

// Префиксы ключей в хранилище.
const (
	keyInstanceStatus = "inst:st:"
	keyInstanceCount  = "inst:pc:"
	keyInstanceUsage  = "inst:us:"
	keyNodeUsage      = "node:us:"
	keyNodeInstCount  = "node:ic:"
)

// NodeStateStore реализует domain.NodeStateStore поверх Valkey.
// Безопасен для конкурентного использования.
type NodeStateStore struct {
	client *redis.Client
	ttl    time.Duration
}

// NewNodeStateStore создаёт хранилище горячего состояния нод.
func NewNodeStateStore(client *redis.Client, ttl time.Duration) *NodeStateStore {
	return &NodeStateStore{client: client, ttl: ttl}
}

// UpdateHeartbeat обновляет время последнего ping ноды и её метрики потребления.
// Сбрасывает TTL ключа.
func (s *NodeStateStore) UpdateHeartbeat(ctx context.Context, nodeID int64, usage *domain.ResourceUsage) error {
	key := keyNodeUsage + strconv.FormatInt(nodeID, 10)
	data, err := json.Marshal(usage)
	if err != nil {
		return fmt.Errorf("valkey.NodeStateStore.UpdateHeartbeat: marshal: %w", err)
	}

	err = s.client.Set(ctx, key, data, s.ttl).Err()
	if err != nil {
		return fmt.Errorf("valkey.NodeStateStore.UpdateHeartbeat: %w", err)
	}

	return nil
}

// GetUsage возвращает текущую загруженность ноды.
// Возвращает ErrNotFound если ключ истёк.
func (s *NodeStateStore) GetUsage(ctx context.Context, nodeID int64) (*domain.ResourceUsage, error) {
	key := keyNodeUsage + strconv.FormatInt(nodeID, 10)

	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("valkey.NodeStateStore.GetUsage: %w", err)
	}

	var usage domain.ResourceUsage
	if err := json.Unmarshal(data, &usage); err != nil {
		return nil, fmt.Errorf("valkey.NodeStateStore.GetUsage: unmarshal: %w", err)
	}

	return &usage, nil
}

// GetActiveInstanceCount возвращает количество активных инстансов на ноде.
// Возвращает ErrNotFound если ключ отсутствует.
func (s *NodeStateStore) GetActiveInstanceCount(ctx context.Context, nodeID int64) (uint32, error) {
	key := keyNodeInstCount + strconv.FormatInt(nodeID, 10)

	count, err := s.client.Get(ctx, key).Uint64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, domain.ErrNotFound
		}
		return 0, fmt.Errorf("valkey.NodeStateStore.GetActiveInstanceCount: %w", err)
	}

	if count > uint64(^uint32(0)) {
		return 0, fmt.Errorf("valkey.NodeStateStore.GetActiveInstanceCount: count %d overflows uint32", count)
	}

	return uint32(count), nil
}

// SetActiveInstanceCount обновляет счётчик активных инстансов на ноде.
// Сбрасывает TTL ключа.
func (s *NodeStateStore) SetActiveInstanceCount(ctx context.Context, nodeID int64, count uint32) error {
	key := keyNodeInstCount + strconv.FormatInt(nodeID, 10)

	err := s.client.Set(ctx, key, count, s.ttl).Err()
	if err != nil {
		return fmt.Errorf("valkey.NodeStateStore.SetActiveInstanceCount: %w", err)
	}

	return nil
}

// Delete удаляет все ключи состояния ноды.
func (s *NodeStateStore) Delete(ctx context.Context, nodeID int64) error {
	keys := []string{
		keyNodeUsage + strconv.FormatInt(nodeID, 10),
		keyNodeInstCount + strconv.FormatInt(nodeID, 10),
	}

	err := s.client.Del(ctx, keys...).Err()
	if err != nil {
		return fmt.Errorf("valkey.NodeStateStore.Delete: %w", err)
	}

	return nil
}
