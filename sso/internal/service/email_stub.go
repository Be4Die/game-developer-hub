package service

import (
	"context"
	"log/slog"
)

// StubEmailSender — заглушка EmailSender для разработки.
type StubEmailSender struct {
	log *slog.Logger
}

// NewStubEmailSender создаёт заглушку отправщика почты.
func NewStubEmailSender(log *slog.Logger) *StubEmailSender {
	return &StubEmailSender{log: log}
}

// SendVerificationEmail логирует код верификации вместо отправки письма.
func (s *StubEmailSender) SendVerificationEmail(_ context.Context, email, code string) error {
	s.log.Info("[EMAIL STUB] verification code", slog.String("email", email), slog.String("code", code))
	return nil
}

// SendPasswordResetEmail логирует токен сброса пароля вместо отправки письма.
func (s *StubEmailSender) SendPasswordResetEmail(_ context.Context, email, token string) error {
	s.log.Info("[EMAIL STUB] password reset token", slog.String("email", email), slog.String("token", token))
	return nil
}
