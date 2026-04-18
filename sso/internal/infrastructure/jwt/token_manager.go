// Package jwt предоставляет реализацию TokenManager на основе JWT.
package jwt

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/Be4Die/game-developer-hub/sso/internal/domain"
	"github.com/golang-jwt/jwt/v5"
)

// TokenManager — реализация domain.TokenManager через JWT.
type TokenManager struct {
	secret         []byte
	accessTokenTTL time.Duration
	issuer         string
}

// NewTokenManager создаёт менеджер JWT-токенов.
func NewTokenManager(secret string, accessTokenTTL time.Duration, issuer string) (*TokenManager, error) {
	if secret == "" {
		return nil, fmt.Errorf("jwt.NewTokenManager: secret is required")
	}
	return &TokenManager{
		secret:         []byte(secret),
		accessTokenTTL: accessTokenTTL,
		issuer:         issuer,
	}, nil
}

// GenerateAccessToken генерирует JWT access token.
func (tm *TokenManager) GenerateAccessToken(_ context.Context, claims domain.Claims) (string, time.Time, error) {
	const op = "jwt.TokenManager.GenerateAccessToken"

	now := time.Now()
	expiresAt := now.Add(tm.accessTokenTTL)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   claims.UserID,
		"sid":   claims.SessionID,
		"email": claims.Email,
		"role":  int(claims.Role),
		"iss":   tm.issuer,
		"iat":   now.Unix(),
		"exp":   expiresAt.Unix(),
	})

	accessToken, err := token.SignedString(tm.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("%s: %w", op, err)
	}

	return accessToken, expiresAt, nil
}

// GenerateRefreshToken генерирует случайный refresh token.
func (tm *TokenManager) GenerateRefreshToken(_ context.Context) (string, error) {
	const op = "jwt.TokenManager.GenerateRefreshToken"

	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return hex.EncodeToString(bytes), nil
}

// ParseAccessToken парсит и валидирует JWT access token.
func (tm *TokenManager) ParseAccessToken(_ context.Context, tokenStr string) (*domain.Claims, error) {
	const op = "jwt.TokenManager.ParseAccessToken"

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return tm.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("%s: %w", op, domain.ErrInvalidToken)
	}

	// Проверка issuer
	if iss, _ := claims.GetIssuer(); iss != tm.issuer {
		return nil, fmt.Errorf("%s: invalid issuer: %w", op, domain.ErrInvalidToken)
	}

	userID, _ := claims.GetSubject()
	sessionID, _ := claims["sid"].(string)
	email, _ := claims["email"].(string)
	roleFloat, _ := claims["role"].(float64)
	role := domain.UserRole(uint8(roleFloat))

	return &domain.Claims{
		UserID:    userID,
		SessionID: sessionID,
		Email:     email,
		Role:      role,
	}, nil
}
