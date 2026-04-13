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

// InstanceStateStore реализует domain.InstanceStateStore поверх Valkey.
// Безопасен для конкурентного использования.
type InstanceStateStore struct {
	client *redis.Client
	ttl    time.Duration
}

// NewInstanceStateStore создаёт хранилище горячего состояния инстансов.
func NewInstanceStateStore(client *redis.Client, ttl time.Duration) *InstanceStateStore {
	return &InstanceStateStore{client: client, ttl: ttl}
}

// SetStatus обновляет статус инстанса и сбрасывает TTL.
func (s *InstanceStateStore) SetStatus(ctx context.Context, instanceID int64, status domain.InstanceStatus) error {
	key := keyInstanceStatus + strconv.FormatInt(instanceID, 10)

	err := s.client.Set(ctx, key, uint8(status), s.ttl).Err()
	if err != nil {
		return fmt.Errorf("valkey.InstanceStateStore.SetStatus: %w", err)
	}

	return nil
}

// GetStatus возвращает текущий статус инстанса.
// Возвращает ErrNotFound если ключ истёк.
func (s *InstanceStateStore) GetStatus(ctx context.Context, instanceID int64) (domain.InstanceStatus, error) {
	key := keyInstanceStatus + strconv.FormatInt(instanceID, 10)

	val, err := s.client.Get(ctx, key).Uint64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, domain.ErrNotFound
		}
		return 0, fmt.Errorf("valkey.InstanceStateStore.GetStatus: %w", err)
	}

	if val > 255 {
		return 0, fmt.Errorf("valkey.InstanceStateStore.GetStatus: status %d out of range", val)
	}

	return domain.InstanceStatus(val), nil
}

// SetPlayerCount обновляет количество игроков в инстансе и сбрасывает TTL.
func (s *InstanceStateStore) SetPlayerCount(ctx context.Context, instanceID int64, count uint32) error {
	key := keyInstanceCount + strconv.FormatInt(instanceID, 10)

	err := s.client.Set(ctx, key, count, s.ttl).Err()
	if err != nil {
		return fmt.Errorf("valkey.InstanceStateStore.SetPlayerCount: %w", err)
	}

	return nil
}

// GetPlayerCount возвращает текущее количество игроков в инстансе.
// Возвращает ErrNotFound если ключ отсутствует.
func (s *InstanceStateStore) GetPlayerCount(ctx context.Context, instanceID int64) (uint32, error) {
	key := keyInstanceCount + strconv.FormatInt(instanceID, 10)

	count, err := s.client.Get(ctx, key).Uint64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, domain.ErrNotFound
		}
		return 0, fmt.Errorf("valkey.InstanceStateStore.GetPlayerCount: %w", err)
	}

	if count > uint64(^uint32(0)) {
		return 0, fmt.Errorf("valkey.InstanceStateStore.GetPlayerCount: count %d overflows uint32", count)
	}

	return uint32(count), nil
}

// SetUsage обновляет метрики потребления ресурсов инстанса и сбрасывает TTL.
func (s *InstanceStateStore) SetUsage(ctx context.Context, instanceID int64, usage *domain.ResourceUsage) error {
	key := keyInstanceUsage + strconv.FormatInt(instanceID, 10)

	data, err := json.Marshal(usage)
	if err != nil {
		return fmt.Errorf("valkey.InstanceStateStore.SetUsage: marshal: %w", err)
	}

	err = s.client.Set(ctx, key, data, s.ttl).Err()
	if err != nil {
		return fmt.Errorf("valkey.InstanceStateStore.SetUsage: %w", err)
	}

	return nil
}

// GetUsage возвращает текущие метрики потребления ресурсов инстанса.
// Возвращает ErrNotFound если ключ истёк.
func (s *InstanceStateStore) GetUsage(ctx context.Context, instanceID int64) (*domain.ResourceUsage, error) {
	key := keyInstanceUsage + strconv.FormatInt(instanceID, 10)

	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("valkey.InstanceStateStore.GetUsage: %w", err)
	}

	var usage domain.ResourceUsage
	if err := json.Unmarshal(data, &usage); err != nil {
		return nil, fmt.Errorf("valkey.InstanceStateStore.GetUsage: unmarshal: %w", err)
	}

	return &usage, nil
}

// Delete удаляет все ключи состояния инстанса.
func (s *InstanceStateStore) Delete(ctx context.Context, instanceID int64) error {
	keys := []string{
		keyInstanceStatus + strconv.FormatInt(instanceID, 10),
		keyInstanceCount + strconv.FormatInt(instanceID, 10),
		keyInstanceUsage + strconv.FormatInt(instanceID, 10),
	}

	err := s.client.Del(ctx, keys...).Err()
	if err != nil {
		return fmt.Errorf("valkey.InstanceStateStore.Delete: %w", err)
	}

	return nil
}
