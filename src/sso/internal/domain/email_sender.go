package domain

import "context"

// EmailSender — интерфейс отправки почты.
type EmailSender interface {
	SendVerificationEmail(ctx context.Context, email, code string) error
	SendPasswordResetEmail(ctx context.Context, email, token string) error
}
