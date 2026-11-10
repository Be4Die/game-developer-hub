package domain

import "context"

// EmailVerificationStore — хранилище кодов верификации email.
type EmailVerificationStore interface {
	Store(ctx context.Context, email, code string) error
	Verify(ctx context.Context, email, code string) (bool, error)
}

// AttemptLimiter ограничивает число неудачных попыток (вход, ввод кода подтверждения) по ключу.
type AttemptLimiter interface {
	// Blocked сообщает, исчерпан ли лимит попыток для ключа.
	Blocked(ctx context.Context, key string) (bool, error)
	// Fail регистрирует неудачную попытку.
	Fail(ctx context.Context, key string) error
	// Reset сбрасывает счётчик после успешной попытки.
	Reset(ctx context.Context, key string) error
}

// PasswordResetStore — хранилище токенов сброса пароля.
type PasswordResetStore interface {
	Store(ctx context.Context, email, token string) error
	Consume(ctx context.Context, token string) (string, error)
}
