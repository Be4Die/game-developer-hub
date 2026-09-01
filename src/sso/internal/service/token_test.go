package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/sso/internal/domain"
)

func TestTokenService_ValidateToken(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	log := newTestLogger()

	t.Run("success with cache hit", func(t *testing.T) {
		t.Parallel()

		expectedClaims := &domain.Claims{UserID: "user-1", SessionID: "session-1", Email: "test@example.com", Role: domain.RoleDeveloper}
		tokenManager := &stubTokenManager{parseAccessFunc: func(context.Context, string) (*domain.Claims, error) { return expectedClaims, nil }}
		sessionCache := &stubSessionCache{getFunc: func(context.Context, string) (*domain.Session, error) {
			return &domain.Session{ID: "session-1", UserID: "user-1", ExpiresAt: time.Now().Add(1 * time.Hour), Revoked: false}, nil
		}}
		svc := NewTokenService(log, &stubSessionRepo{}, sessionCache, tokenManager)

		claims, err := svc.ValidateToken(ctx, "valid-token")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if claims.UserID != expectedClaims.UserID {
			t.Errorf("expected user ID %s, got %s", expectedClaims.UserID, claims.UserID)
		}
	})

	t.Run("cache miss, fallback to db", func(t *testing.T) {
		t.Parallel()

		expectedClaims := &domain.Claims{UserID: "user-1", SessionID: "session-1", Email: "test@example.com", Role: domain.RoleDeveloper}
		tokenManager := &stubTokenManager{parseAccessFunc: func(context.Context, string) (*domain.Claims, error) { return expectedClaims, nil }}
		sessionCache := &stubSessionCache{
			getFunc: func(context.Context, string) (*domain.Session, error) { return nil, errors.New("cache miss") },
			setFunc: func(context.Context, *domain.Session) error { return nil },
		}
		sessionRepo := &stubSessionRepo{getByIDFunc: func(context.Context, string) (*domain.Session, error) {
			return &domain.Session{ID: "session-1", UserID: "user-1", ExpiresAt: time.Now().Add(1 * time.Hour), Revoked: false}, nil
		}}
		svc := NewTokenService(log, sessionRepo, sessionCache, tokenManager)

		claims, err := svc.ValidateToken(ctx, "valid-token")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if claims.SessionID != "session-1" {
			t.Errorf("expected session ID session-1, got %s", claims.SessionID)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		t.Parallel()

		tokenManager := &stubTokenManager{parseAccessFunc: func(context.Context, string) (*domain.Claims, error) { return nil, errors.New("invalid token") }}
		svc := NewTokenService(log, &stubSessionRepo{}, &stubSessionCache{}, tokenManager)

		_, err := svc.ValidateToken(ctx, "invalid-token")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, domain.ErrInvalidToken) {
			t.Errorf("expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("session not found", func(t *testing.T) {
		t.Parallel()

		expectedClaims := &domain.Claims{UserID: "user-1", SessionID: "session-1"}
		tokenManager := &stubTokenManager{parseAccessFunc: func(context.Context, string) (*domain.Claims, error) { return expectedClaims, nil }}
		sessionCache := &stubSessionCache{getFunc: func(context.Context, string) (*domain.Session, error) { return nil, errors.New("cache miss") }}
		sessionRepo := &stubSessionRepo{getByIDFunc: func(context.Context, string) (*domain.Session, error) { return nil, errors.New("not found") }}
		svc := NewTokenService(log, sessionRepo, sessionCache, tokenManager)

		_, err := svc.ValidateToken(ctx, "valid-token")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("expired session", func(t *testing.T) {
		t.Parallel()

		expectedClaims := &domain.Claims{UserID: "user-1", SessionID: "session-1"}
		tokenManager := &stubTokenManager{parseAccessFunc: func(context.Context, string) (*domain.Claims, error) { return expectedClaims, nil }}
		sessionCache := &stubSessionCache{
			getFunc: func(context.Context, string) (*domain.Session, error) { return nil, errors.New("cache miss") },
			setFunc: func(context.Context, *domain.Session) error { return nil },
		}
		sessionRepo := &stubSessionRepo{getByIDFunc: func(context.Context, string) (*domain.Session, error) {
			return &domain.Session{ID: "session-1", UserID: "user-1", ExpiresAt: time.Now().Add(-1 * time.Hour), Revoked: true}, nil
		}}
		svc := NewTokenService(log, sessionRepo, sessionCache, tokenManager)

		_, err := svc.ValidateToken(ctx, "valid-token")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, domain.ErrTokenExpired) {
			t.Errorf("expected ErrTokenExpired, got %v", err)
		}
	})
}

func TestTokenService_ListSessions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	log := newTestLogger()

	t.Run("success with filtering", func(t *testing.T) {
		t.Parallel()

		now := time.Now()
		sessionRepo := &stubSessionRepo{getByUserIDFunc: func(context.Context, string) ([]domain.Session, error) {
			return []domain.Session{
				{ID: "session-1", UserID: "user-1", Revoked: false, ExpiresAt: now.Add(1 * time.Hour)},
				{ID: "session-2", UserID: "user-1", Revoked: true, ExpiresAt: now.Add(1 * time.Hour)},
				{ID: "session-3", UserID: "user-1", Revoked: false, ExpiresAt: now.Add(1 * time.Hour)},
			}, nil
		}}
		svc := NewTokenService(log, sessionRepo, &stubSessionCache{}, &stubTokenManager{})

		sessions, err := svc.ListSessions(ctx, "user-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(sessions) != 2 {
			t.Errorf("expected 2 active sessions, got %d", len(sessions))
		}
	})

	t.Run("empty sessions", func(t *testing.T) {
		t.Parallel()

		sessionRepo := &stubSessionRepo{getByUserIDFunc: func(context.Context, string) ([]domain.Session, error) { return []domain.Session{}, nil }}
		svc := NewTokenService(log, sessionRepo, &stubSessionCache{}, &stubTokenManager{})

		sessions, err := svc.ListSessions(ctx, "user-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(sessions) != 0 {
			t.Errorf("expected empty slice, got %v", sessions)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		sessionRepo := &stubSessionRepo{getByUserIDFunc: func(context.Context, string) ([]domain.Session, error) { return nil, errors.New("db error") }}
		svc := NewTokenService(log, sessionRepo, &stubSessionCache{}, &stubTokenManager{})

		_, err := svc.ListSessions(ctx, "user-1")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestTokenService_RevokeSession(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	log := newTestLogger()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		sessionRepo := &stubSessionRepo{
			getByIDFunc: func(context.Context, string) (*domain.Session, error) {
				return &domain.Session{ID: "session-1", UserID: "user-1"}, nil
			},
			revokeFunc: func(context.Context, string) error { return nil },
		}
		cacheInvalidated := false
		sessionCache := &stubSessionCache{invalidateFunc: func(context.Context, string) error { cacheInvalidated = true; return nil }}
		svc := NewTokenService(log, sessionRepo, sessionCache, &stubTokenManager{})

		err := svc.RevokeSession(ctx, "user-1", "session-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !cacheInvalidated {
			t.Error("expected cache to be invalidated")
		}
	})

	t.Run("session not found", func(t *testing.T) {
		t.Parallel()

		sessionRepo := &stubSessionRepo{getByIDFunc: func(context.Context, string) (*domain.Session, error) { return nil, errors.New("not found") }}
		svc := NewTokenService(log, sessionRepo, &stubSessionCache{}, &stubTokenManager{})

		err := svc.RevokeSession(ctx, "user-1", "nonexistent")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("session belongs to different user", func(t *testing.T) {
		t.Parallel()

		sessionRepo := &stubSessionRepo{getByIDFunc: func(context.Context, string) (*domain.Session, error) {
			return &domain.Session{ID: "session-1", UserID: "other-user"}, nil
		}}
		svc := NewTokenService(log, sessionRepo, &stubSessionCache{}, &stubTokenManager{})

		err := svc.RevokeSession(ctx, "user-1", "session-1")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestTokenService_RevokeAllSessions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	log := newTestLogger()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		sessionRepo := &stubSessionRepo{revokeAllForUserFunc: func(context.Context, string, string) (int64, error) { return 3, nil }}
		svc := NewTokenService(log, sessionRepo, &stubSessionCache{}, &stubTokenManager{})

		count, err := svc.RevokeAllSessions(ctx, "user-1", "session-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if count != 3 {
			t.Errorf("expected count 3, got %d", count)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		sessionRepo := &stubSessionRepo{revokeAllForUserFunc: func(context.Context, string, string) (int64, error) { return 0, errors.New("db error") }}
		svc := NewTokenService(log, sessionRepo, &stubSessionCache{}, &stubTokenManager{})

		_, err := svc.RevokeAllSessions(ctx, "user-1", "")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
