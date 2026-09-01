package jwt_test

import (
	"context"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/sso/internal/domain"
	jwt "github.com/Be4Die/game-developer-hub/sso/internal/infrastructure/jwt"
)

func TestNewTokenManager(t *testing.T) {
	t.Parallel()

	t.Run("valid secret", func(t *testing.T) {
		t.Parallel()

		tm, err := jwt.NewTokenManager("secret", 15*time.Minute, "test-issuer")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if tm == nil {
			t.Fatal("expected TokenManager instance, got nil")
		}
	})

	t.Run("empty secret", func(t *testing.T) {
		t.Parallel()

		_, err := jwt.NewTokenManager("", 15*time.Minute, "test-issuer")

		if err == nil {
			t.Fatal("expected error for empty secret, got nil")
		}
	})
}

func TestTokenManager_GenerateAccessToken(t *testing.T) {
	t.Parallel()

	tm, err := jwt.NewTokenManager("test-secret", 15*time.Minute, "test-issuer")
	if err != nil {
		t.Fatalf("failed to create TokenManager: %v", err)
	}

	claims := domain.Claims{
		UserID:    "user-1",
		SessionID: "session-1",
		Email:     "test@example.com",
		Role:      domain.RoleDeveloper,
	}

	token, expiresAt, err := tm.GenerateAccessToken(context.Background(), claims)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token == "" {
		t.Fatal("expected non-empty token")
	}

	if expiresAt.Before(time.Now()) {
		t.Error("expected expiration in the future")
	}
}

func TestTokenManager_GenerateRefreshToken(t *testing.T) {
	t.Parallel()

	tm, err := jwt.NewTokenManager("test-secret", 15*time.Minute, "test-issuer")
	if err != nil {
		t.Fatalf("failed to create TokenManager: %v", err)
	}

	token, err := tm.GenerateRefreshToken(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token == "" {
		t.Fatal("expected non-empty refresh token")
	}

	// Refresh token should be 64 hex chars (32 bytes)
	if len(token) != 64 {
		t.Errorf("expected refresh token length 64, got %d", len(token))
	}
}

func TestTokenManager_ParseAccessToken(t *testing.T) {
	t.Parallel()

	tm, err := jwt.NewTokenManager("test-secret", 15*time.Minute, "test-issuer")
	if err != nil {
		t.Fatalf("failed to create TokenManager: %v", err)
	}

	t.Run("valid token", func(t *testing.T) {
		t.Parallel()

		claims := domain.Claims{
			UserID:    "user-1",
			SessionID: "session-1",
			Email:     "test@example.com",
			Role:      domain.RoleDeveloper,
		}

		token, _, err := tm.GenerateAccessToken(context.Background(), claims)
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		parsed, err := tm.ParseAccessToken(context.Background(), token)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if parsed.UserID != claims.UserID {
			t.Errorf("expected UserID %s, got %s", claims.UserID, parsed.UserID)
		}

		if parsed.SessionID != claims.SessionID {
			t.Errorf("expected SessionID %s, got %s", claims.SessionID, parsed.SessionID)
		}

		if parsed.Email != claims.Email {
			t.Errorf("expected Email %s, got %s", claims.Email, parsed.Email)
		}

		if parsed.Role != claims.Role {
			t.Errorf("expected Role %v, got %v", claims.Role, parsed.Role)
		}
	})

	t.Run("invalid token format", func(t *testing.T) {
		t.Parallel()

		_, err := tm.ParseAccessToken(context.Background(), "invalid-token")

		if err == nil {
			t.Fatal("expected error for invalid token, got nil")
		}
	})

	t.Run("tampered token", func(t *testing.T) {
		t.Parallel()

		claims := domain.Claims{
			UserID:    "user-1",
			SessionID: "session-1",
			Email:     "test@example.com",
			Role:      domain.RoleDeveloper,
		}

		token, _, err := tm.GenerateAccessToken(context.Background(), claims)
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		// Tamper with the token
		tampered := token[:len(token)-5] + "XXXXX"

		_, err = tm.ParseAccessToken(context.Background(), tampered)

		if err == nil {
			t.Fatal("expected error for tampered token, got nil")
		}
	})

	t.Run("wrong issuer", func(t *testing.T) {
		t.Parallel()

		tm2, err := jwt.NewTokenManager("other-secret", 15*time.Minute, "other-issuer")
		if err != nil {
			t.Fatalf("failed to create TokenManager: %v", err)
		}

		claims := domain.Claims{
			UserID:    "user-1",
			SessionID: "session-1",
			Email:     "test@example.com",
			Role:      domain.RoleDeveloper,
		}

		token, _, err := tm2.GenerateAccessToken(context.Background(), claims)
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		// Try to parse with tm (different issuer)
		_, err = tm.ParseAccessToken(context.Background(), token)

		if err == nil {
			t.Fatal("expected error for wrong issuer, got nil")
		}
	})
}

func TestTokenManager_RoundTrip(t *testing.T) {
	t.Parallel()

	tm, err := jwt.NewTokenManager("round-trip-secret", 15*time.Minute, "test-issuer")
	if err != nil {
		t.Fatalf("failed to create TokenManager: %v", err)
	}

	originalClaims := domain.Claims{
		UserID:    "user-42",
		SessionID: "session-99",
		Email:     "roundtrip@example.com",
		Role:      domain.RoleAdmin,
	}

	token, expiresAt, err := tm.GenerateAccessToken(context.Background(), originalClaims)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	parsedClaims, err := tm.ParseAccessToken(context.Background(), token)
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}

	if parsedClaims.UserID != originalClaims.UserID {
		t.Errorf("UserID mismatch: expected %s, got %s", originalClaims.UserID, parsedClaims.UserID)
	}

	if parsedClaims.SessionID != originalClaims.SessionID {
		t.Errorf("SessionID mismatch: expected %s, got %s", originalClaims.SessionID, parsedClaims.SessionID)
	}

	if parsedClaims.Email != originalClaims.Email {
		t.Errorf("Email mismatch: expected %s, got %s", originalClaims.Email, parsedClaims.Email)
	}

	if parsedClaims.Role != originalClaims.Role {
		t.Errorf("Role mismatch: expected %v, got %v", originalClaims.Role, parsedClaims.Role)
	}

	if expiresAt.Before(time.Now()) {
		t.Error("expiration should be in the future")
	}
}
