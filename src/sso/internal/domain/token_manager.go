package domain

import (
	"context"
	"time"
)

// TokenManager — интерфейс для работы с JWT токенами.
type TokenManager interface {
	// GenerateAccessToken генерирует JWT access token.
	GenerateAccessToken(ctx context.Context, claims Claims) (string, time.Time, error)
	// GenerateRefreshToken генерирует случайный refresh token.
	GenerateRefreshToken(ctx context.Context) (string, error)
	// ParseAccessToken парсит и валидирует JWT access token.
	ParseAccessToken(ctx context.Context, token string) (*Claims, error)
}
