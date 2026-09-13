package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// ErrTargetNotFound возвращается, когда маршрут инстанса не найден в Valkey.
var ErrTargetNotFound = errors.New("instance target route not found")

// RouteResolver определяет контракт поиска целевого адреса инстанса игрового сервера.
type RouteResolver interface {
	ResolveInstanceTarget(ctx context.Context, instanceID int64) (string, error)
	Close() error
}

// ValkeyStore реализует RouteResolver с хранением маршрутов в Valkey/Redis.
type ValkeyStore struct {
	client *redis.Client
}

// NewValkeyStore создаёт клиент Valkey.
func NewValkeyStore(addr, password string, db int) *ValkeyStore {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &ValkeyStore{client: client}
}

// ResolveInstanceTarget ищет целевой адрес host:port инстанса по ключу instance:{id}:target.
func (s *ValkeyStore) ResolveInstanceTarget(ctx context.Context, instanceID int64) (string, error) {
	key := fmt.Sprintf("instance:%d:target", instanceID)
	target, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrTargetNotFound
		}
		return "", fmt.Errorf("valkey get %s: %w", key, err)
	}
	if target == "" {
		return "", ErrTargetNotFound
	}
	return target, nil
}

// Close закрывает соединение с Valkey.
func (s *ValkeyStore) Close() error {
	return s.client.Close()
}

// Ping проверяет доступность сервера Valkey.
func (s *ValkeyStore) Ping(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}
