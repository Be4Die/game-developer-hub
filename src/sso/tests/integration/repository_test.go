//go:build integration

package integration

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/sso/internal/domain"
)

// TestUserRepository_CRUD проверяет полный цикл операций с пользователями.
func TestUserRepository_CRUD(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	// Create
	now := time.Now()
	user := domain.User{
		Email:         "crud@test.com",
		PasswordHash:  []byte("hash"),
		DisplayName:   "CRUD User",
		Role:          domain.RoleDeveloper,
		Status:        domain.StatusActive,
		EmailVerified: true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := env.userRepo.Create(ctx, user); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// GetByEmail
	fetched, err := env.userRepo.GetByEmail(ctx, "crud@test.com")
	if err != nil {
		t.Fatalf("GetByEmail failed: %v", err)
	}
	if fetched.Email != user.Email {
		t.Errorf("expected email %s, got %s", user.Email, fetched.Email)
	}
	if fetched.DisplayName != "CRUD User" {
		t.Errorf("expected display name 'CRUD User', got %s", fetched.DisplayName)
	}
	userID := fetched.ID

	// GetByID
	byID, err := env.userRepo.GetByID(ctx, userID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if byID.ID != userID {
		t.Errorf("expected ID %s, got %s", userID, byID.ID)
	}

	// Update
	byID.DisplayName = "Updated Name"
	byID.UpdatedAt = time.Now()
	if err := env.userRepo.Update(ctx, *byID); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	updated, err := env.userRepo.GetByID(ctx, userID)
	if err != nil {
		t.Fatalf("GetByID after update failed: %v", err)
	}
	if updated.DisplayName != "Updated Name" {
		t.Errorf("expected updated display name 'Updated Name', got %s", updated.DisplayName)
	}
}

// TestUserRepository_DuplicateEmail проверяет что дублирующийся email возвращает ErrAlreadyExists.
func TestUserRepository_DuplicateEmail(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	now := time.Now()
	user := domain.User{
		Email: "dup@test.com", PasswordHash: []byte("hash"), DisplayName: "Dup",
		Role: domain.RoleDeveloper, Status: domain.StatusActive, EmailVerified: true,
		CreatedAt: now, UpdatedAt: now,
	}

	if err := env.userRepo.Create(ctx, user); err != nil {
		t.Fatalf("first Create failed: %v", err)
	}

	err := env.userRepo.Create(ctx, user)
	if err == nil {
		t.Fatal("expected error for duplicate email")
	}
	if !errors.Is(err, domain.ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists, got: %v", err)
	}
}

// TestUserRepository_NotFound проверяет ErrNotFound для несуществующих пользователей.
func TestUserRepository_NotFound(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	_, err := env.userRepo.GetByEmail(ctx, "nonexistent@test.com")
	if err == nil {
		t.Fatal("expected error for nonexistent email")
	}

	_, err = env.userRepo.GetByID(ctx, "00000000-0000-0000-0000-000000000000")
	if err == nil {
		t.Fatal("expected error for nonexistent ID")
	}
}

// TestUserRepository_Search проверяет поиск пользователей.
func TestUserRepository_Search(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	now := time.Now()
	users := []domain.User{
		{Email: "alice@test.com", PasswordHash: []byte("h"), DisplayName: "Alice", Role: domain.RoleDeveloper, Status: domain.StatusActive, EmailVerified: true, CreatedAt: now, UpdatedAt: now},
		{Email: "bob@test.com", PasswordHash: []byte("h"), DisplayName: "Bob", Role: domain.RoleDeveloper, Status: domain.StatusActive, EmailVerified: true, CreatedAt: now, UpdatedAt: now},
		{Email: "charlie@test.com", PasswordHash: []byte("h"), DisplayName: "Charlie", Role: domain.RoleDeveloper, Status: domain.StatusActive, EmailVerified: true, CreatedAt: now, UpdatedAt: now},
	}
	for _, u := range users {
		if err := env.userRepo.Create(ctx, u); err != nil {
			t.Fatalf("Create %s failed: %v", u.Email, err)
		}
	}

	// Search by name
	results, total, err := env.userRepo.Search(ctx, "bob", 10, 0)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if results[0].DisplayName != "Bob" {
		t.Errorf("expected Bob, got %s", results[0].DisplayName)
	}

	// Search with no matches
	results, total, err = env.userRepo.Search(ctx, "zzz", 10, 0)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if total != 0 {
		t.Errorf("expected total 0, got %d", total)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

// TestSessionRepository_CRUD проверяет полный цикл операций с сессиями.
func TestSessionRepository_CRUD(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	// Create user first
	now := time.Now()
	user := domain.User{
		Email: "session-user@test.com", PasswordHash: []byte("hash"), DisplayName: "Session User",
		Role: domain.RoleDeveloper, Status: domain.StatusActive, EmailVerified: true,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := env.userRepo.Create(ctx, user); err != nil {
		t.Fatalf("Create user failed: %v", err)
	}
	createdUser, err := env.userRepo.GetByEmail(ctx, user.Email)
	if err != nil {
		t.Fatalf("GetByEmail failed: %v", err)
	}

	// Create session
	session := domain.Session{
		UserID:           createdUser.ID,
		UserAgent:        "test-agent",
		IPAddress:        "127.0.0.1",
		RefreshTokenHash: "test-hash",
		CreatedAt:        now,
		LastUsedAt:       now,
		ExpiresAt:        now.Add(24 * time.Hour),
	}
	if err := env.sessionRepo.Create(ctx, session); err != nil {
		t.Fatalf("Create session failed: %v", err)
	}

	// GetByRefreshTokenHash
	found, err := env.sessionRepo.GetByRefreshTokenHash(ctx, "test-hash")
	if err != nil {
		t.Fatalf("GetByRefreshTokenHash failed: %v", err)
	}
	if found.UserID != createdUser.ID {
		t.Errorf("expected user ID %s, got %s", createdUser.ID, found.UserID)
	}
	sessionID := found.ID

	// GetByID
	byID, err := env.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if byID.ID != sessionID {
		t.Errorf("expected session ID %s, got %s", sessionID, byID.ID)
	}

	// GetByUserID
	sessions, err := env.sessionRepo.GetByUserID(ctx, createdUser.ID)
	if err != nil {
		t.Fatalf("GetByUserID failed: %v", err)
	}
	if len(sessions) != 1 {
		t.Errorf("expected 1 session, got %d", len(sessions))
	}

	// Revoke
	if err := env.sessionRepo.Revoke(ctx, sessionID); err != nil {
		t.Fatalf("Revoke failed: %v", err)
	}

	revoked, err := env.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		t.Fatalf("GetByID after revoke failed: %v", err)
	}
	if !revoked.Revoked {
		t.Error("expected session to be revoked")
	}
}

// TestSessionRepository_RevokeAllForUser проверяет отзыв всех сессий пользователя.
func TestSessionRepository_RevokeAllForUser(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	now := time.Now()
	user := domain.User{
		Email: "revoke-all@test.com", PasswordHash: []byte("hash"), DisplayName: "Revoke All",
		Role: domain.RoleDeveloper, Status: domain.StatusActive, EmailVerified: true,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := env.userRepo.Create(ctx, user); err != nil {
		t.Fatalf("Create user failed: %v", err)
	}
	createdUser, err := env.userRepo.GetByEmail(ctx, user.Email)
	if err != nil {
		t.Fatalf("GetByEmail failed: %v", err)
	}

	// Create 3 sessions
	for i := 0; i < 3; i++ {
		session := domain.Session{
			UserID: createdUser.ID, RefreshTokenHash: "hash", CreatedAt: now, LastUsedAt: now,
			ExpiresAt: now.Add(24 * time.Hour),
		}
		if err := env.sessionRepo.Create(ctx, session); err != nil {
			t.Fatalf("Create session %d failed: %v", i, err)
		}
	}

	// Revoke all except first
	allSessions, err := env.sessionRepo.GetByUserID(ctx, createdUser.ID)
	if err != nil {
		t.Fatalf("GetByUserID failed: %v", err)
	}
	excludeID := allSessions[0].ID

	count, err := env.sessionRepo.RevokeAllForUser(ctx, createdUser.ID, excludeID)
	if err != nil {
		t.Fatalf("RevokeAllForUser failed: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 revoked, got %d", count)
	}

	remaining, err := env.sessionRepo.GetByUserID(ctx, createdUser.ID)
	if err != nil {
		t.Fatalf("GetByUserID after revoke failed: %v", err)
	}
	activeCount := 0
	for _, s := range remaining {
		if !s.Revoked {
			activeCount++
		}
	}
	if activeCount != 1 {
		t.Errorf("expected 1 active session, got %d", activeCount)
	}
}

// TestValkeyStores проверяет работу хранилищ Valkey.
func TestValkeyStores(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	// EmailVerificationStore
	if err := env.emailStore.Store(ctx, "test@example.com", "123456"); err != nil {
		t.Fatalf("emailStore.Store failed: %v", err)
	}

	valid, err := env.emailStore.Verify(ctx, "test@example.com", "123456")
	if err != nil {
		t.Fatalf("emailStore.Verify failed: %v", err)
	}
	if !valid {
		t.Error("expected verification to be valid")
	}

	// PasswordResetStore
	if err := env.resetStore.Store(ctx, "reset@example.com", "reset-token"); err != nil {
		t.Fatalf("resetStore.Store failed: %v", err)
	}

	email, err := env.resetStore.Consume(ctx, "reset-token")
	if err != nil {
		t.Fatalf("resetStore.Consume failed: %v", err)
	}
	if email != "reset@example.com" {
		t.Errorf("expected email reset@example.com, got %s", email)
	}

	// Consume again should fail (one-time use)
	_, err = env.resetStore.Consume(ctx, "reset-token")
	if err == nil {
		t.Error("expected error on second consume")
	}
}

// TestValkeySessionCache проверяет кэш сессий в Valkey.
func TestValkeySessionCache(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	session := &domain.Session{
		ID:        "cache-session-1",
		UserID:    "user-1",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		Revoked:   false,
	}

	if err := env.sessionCache.Set(ctx, session); err != nil {
		t.Fatalf("sessionCache.Set failed: %v", err)
	}

	cached, err := env.sessionCache.Get(ctx, "cache-session-1")
	if err != nil {
		t.Fatalf("sessionCache.Get failed: %v", err)
	}
	if cached.UserID != "user-1" {
		t.Errorf("expected user ID user-1, got %s", cached.UserID)
	}

	if err := env.sessionCache.Invalidate(ctx, "cache-session-1"); err != nil {
		t.Fatalf("sessionCache.Invalidate failed: %v", err)
	}

	_, err = env.sessionCache.Get(ctx, "cache-session-1")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound after invalidation, got: %v", err)
	}
}
