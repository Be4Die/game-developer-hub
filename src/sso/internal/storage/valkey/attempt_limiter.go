package valkey

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// AttemptLimiter — счётчик неудачных попыток в Valkey с фиксированным окном.
// После maxAttempts неудач ключ считается заблокированным до истечения окна.
type AttemptLimiter struct {
	client      *redis.Client
	prefix      string
	maxAttempts int64
	window      time.Duration
}

// NewAttemptLimiter создаёт ограничитель попыток.
func NewAttemptLimiter(client *redis.Client, prefix string, maxAttempts int, window time.Duration) *AttemptLimiter {
	return &AttemptLimiter{client: client, prefix: prefix, maxAttempts: int64(maxAttempts), window: window}
}

func (l *AttemptLimiter) key(key string) string {
	return "sso:attempts:" + l.prefix + ":" + strings.ToLower(key)
}

// Blocked сообщает, исчерпан ли лимит попыток для ключа.
func (l *AttemptLimiter) Blocked(ctx context.Context, key string) (bool, error) {
	n, err := l.client.Get(ctx, l.key(key)).Int64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, fmt.Errorf("valkey.AttemptLimiter.Blocked: %w", err)
	}
	return n >= l.maxAttempts, nil
}

// Fail регистрирует неудачную попытку. Окно отсчитывается от первой неудачи.
func (l *AttemptLimiter) Fail(ctx context.Context, key string) error {
	k := l.key(key)
	n, err := l.client.Incr(ctx, k).Result()
	if err != nil {
		return fmt.Errorf("valkey.AttemptLimiter.Fail: %w", err)
	}
	if n == 1 {
		if err := l.client.Expire(ctx, k, l.window).Err(); err != nil {
			return fmt.Errorf("valkey.AttemptLimiter.Fail: expire: %w", err)
		}
	}
	return nil
}

// Reset сбрасывает счётчик.
func (l *AttemptLimiter) Reset(ctx context.Context, key string) error {
	if err := l.client.Del(ctx, l.key(key)).Err(); err != nil {
		return fmt.Errorf("valkey.AttemptLimiter.Reset: %w", err)
	}
	return nil
}
