// Package valkey provides distributed locking implementations using Valkey/Redis.
package valkey

import (
	"context"
	"fmt"
	"time"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
	"github.com/redis/go-redis/v9"
)

// Locker реализует domain.Locker с использованием Valkey/Redis (SetNX).
type Locker struct {
	client *redis.Client
}

// NewLocker создаёт новый экземпляр распределенного локера.
func NewLocker(addr, password string, db int) (*Locker, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping valkey %s: %w", addr, err)
	}

	return &Locker{client: client}, nil
}

// Close закрывает соединение с Valkey.
func (l *Locker) Close() error {
	if l.client != nil {
		return l.client.Close()
	}
	return nil
}

// Acquire пытается захватить распределенную блокировку.
func (l *Locker) Acquire(ctx context.Context, key string, ttl time.Duration) (func(), error) {
	lockKey := "lock:" + key
	ok, err := l.client.SetNX(ctx, lockKey, "1", ttl).Result()
	if err != nil {
		return nil, fmt.Errorf("valkey acquire lock: %w", err)
	}
	if !ok {
		return nil, domain.ErrLockBusy
	}

	unlock := func() {
		_ = l.client.Del(context.WithoutCancel(ctx), lockKey).Err()
	}

	return unlock, nil
}

// NoOpLocker заглушка локера без Valkey (для локальных тестов).
type NoOpLocker struct{}

// NewNoOpLocker создает фиктивный локер.
func NewNoOpLocker() *NoOpLocker {
	return &NoOpLocker{}
}

// Acquire всегда успешно возвращает пустой unlock.
func (l *NoOpLocker) Acquire(_ context.Context, _ string, _ time.Duration) (func(), error) {
	return func() {}, nil
}
