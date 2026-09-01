package domain

import "context"

// EmailVerificationStore — хранилище кодов верификации email.
type EmailVerificationStore interface {
	Store(ctx context.Context, email, code string) error
	Verify(ctx context.Context, email, code string) (bool, error)
	GetEmailByCode(ctx context.Context, code string) (string, error)
}

// PasswordResetStore — хранилище токенов сброса пароля.
type PasswordResetStore interface {
	Store(ctx context.Context, email, token string) error
	Consume(ctx context.Context, token string) (string, error)
}
