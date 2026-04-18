// Package valkey предоставляет Valkey/Redis-реализации хранилищ SSO.
package valkey

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Be4Die/game-developer-hub/sso/internal/domain"
	"github.com/redis/go-redis/v9"
)

// EmailVerificationStore — хранилище кодов верификации email в Valkey.
type EmailVerificationStore struct {
	client *redis.Client
	ttl    time.Duration
}

// NewEmailVerificationStore создаёт хранилище кодов верификации.
func NewEmailVerificationStore(client *redis.Client, ttl time.Duration) *EmailVerificationStore {
	return &EmailVerificationStore{client: client, ttl: ttl}
}

// Store сохраняет код верификации для email.
func (s *EmailVerificationStore) Store(ctx context.Context, email, code string) error {
	const op = "valkey.EmailVerificationStore.Store"

	key := verificationKey(email)
	if err := s.client.Set(ctx, key, code, s.ttl).Err(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// Обратный маппинг: code -> email для верификации по коду.
	codeKey := codeToEmailKey(code)
	if err := s.client.Set(ctx, codeKey, email, s.ttl).Err(); err != nil {
		return fmt.Errorf("%s: reverse map: %w", op, err)
	}

	return nil
}

// Verify проверяет код верификации и удаляет его при успехе.
func (s *EmailVerificationStore) Verify(ctx context.Context, email, code string) (bool, error) {
	const op = "valkey.EmailVerificationStore.Verify"

	key := verificationKey(email)
	storedCode, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}

	if storedCode != code {
		return false, nil
	}

	// Удаляем код после успешной верификации
	if err := s.client.Del(ctx, key).Err(); err != nil {
		return false, fmt.Errorf("%s: del: %w", op, err)
	}
	// Удаляем обратный маппинг.
	_ = s.client.Del(ctx, codeToEmailKey(code)).Err()

	return true, nil
}

// GetEmailByCode находит email по коду верификации.
func (s *EmailVerificationStore) GetEmailByCode(ctx context.Context, code string) (string, error) {
	const op = "valkey.EmailVerificationStore.GetEmailByCode"

	key := codeToEmailKey(code)
	email, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", nil
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return email, nil
}

// PasswordResetStore — хранилище токенов сброса пароля в Valkey.
type PasswordResetStore struct {
	client *redis.Client
	ttl    time.Duration
}

// NewPasswordResetStore создаёт хранилище токенов сброса пароля.
func NewPasswordResetStore(client *redis.Client, ttl time.Duration) *PasswordResetStore {
	return &PasswordResetStore{client: client, ttl: ttl}
}

// Store сохраняет токен сброса пароля.
func (s *PasswordResetStore) Store(ctx context.Context, email, token string) error {
	const op = "valkey.PasswordResetStore.Store"

	key := resetKey(token)
	data, _ := json.Marshal(map[string]string{"email": email})
	if err := s.client.Set(ctx, key, data, s.ttl).Err(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// Consume извлекает email по токену и удаляет токен.
func (s *PasswordResetStore) Consume(ctx context.Context, token string) (string, error) {
	const op = "valkey.PasswordResetStore.Consume"

	key := resetKey(token)
	data, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", domain.ErrInvalidToken
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	var payload map[string]string
	if err := json.Unmarshal([]byte(data), &payload); err != nil {
		return "", fmt.Errorf("%s: unmarshal: %w", op, err)
	}

	// Удаляем токен после использования
	if err := s.client.Del(ctx, key).Err(); err != nil {
		return "", fmt.Errorf("%s: del: %w", op, err)
	}

	return payload["email"], nil
}

// SessionCache — кэш активных сессий в Valkey для быстрого доступа.
type SessionCache struct {
	client *redis.Client
	ttl    time.Duration
}

// NewSessionCache создаёт кэш сессий.
func NewSessionCache(client *redis.Client, ttl time.Duration) *SessionCache {
	return &SessionCache{client: client, ttl: ttl}
}

// Set сохраняет сессию в кэш.
func (c *SessionCache) Set(ctx context.Context, session *domain.Session) error {
	const op = "valkey.SessionCache.Set"

	key := sessionKey(session.ID)
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("%s: marshal: %w", op, err)
	}

	if err := c.client.Set(ctx, key, data, c.ttl).Err(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// Get получает сессию из кэша.
func (c *SessionCache) Get(ctx context.Context, sessionID string) (*domain.Session, error) {
	const op = "valkey.SessionCache.Get"

	key := sessionKey(sessionID)
	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var session domain.Session
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, fmt.Errorf("%s: unmarshal: %w", op, err)
	}

	return &session, nil
}

// Invalidate удаляет сессию из кэша.
func (c *SessionCache) Invalidate(ctx context.Context, sessionID string) error {
	const op = "valkey.SessionCache.Invalidate"

	key := sessionKey(sessionID)
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func verificationKey(email string) string {
	return "sso:verify:" + email
}

func codeToEmailKey(code string) string {
	return "sso:verify:code:" + code
}

func resetKey(token string) string {
	return "sso:reset:" + token
}

func sessionKey(sessionID string) string {
	return "sso:session:" + sessionID
}
