//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/sso/internal/domain"
)

// TestAuthService_FullRegistrationFlow проверяет полный поток: register → verify email → login.
func TestAuthService_FullRegistrationFlow(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	// Step 1: Register (creates user with email_verified=false, status=suspended)
	// We bypass Register() since it sends emails; instead create directly.
	now := time.Now()
	hash, err := env.passwordHasher.Hash(ctx, "password123")
	if err != nil {
		t.Fatalf("hash failed: %v", err)
	}

	user := domain.User{
		Email:         "flow@test.com",
		PasswordHash:  hash,
		DisplayName:   "Flow User",
		Role:          domain.RoleDeveloper,
		Status:        domain.StatusSuspended,
		EmailVerified: false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := env.userRepo.Create(ctx, user); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	created, _ := env.userRepo.GetByEmail(ctx, "flow@test.com")

	// Step 2: Try to login before verification — should fail
	_, err = env.authService.Login(ctx, domain.LoginRequest{
		Email: "flow@test.com", Password: "password123",
	})
	if err == nil {
		t.Fatal("expected error logging in before verification")
	}

	// Step 3: Verify email
	if err := env.emailStore.Store(ctx, "flow@test.com", "123456"); err != nil {
		t.Fatalf("store verification code failed: %v", err)
	}

	err = env.authService.VerifyEmail(ctx, domain.VerifyEmailRequest{VerificationCode: "123456"})
	if err != nil {
		t.Fatalf("VerifyEmail failed: %v", err)
	}

	// Step 4: Login after verification
	loginResp, err := env.authService.Login(ctx, domain.LoginRequest{
		Email: "flow@test.com", Password: "password123",
	})
	if err != nil {
		t.Fatalf("Login after verification failed: %v", err)
	}
	if loginResp.Tokens.AccessToken == "" {
		t.Error("expected access token after login")
	}
	if loginResp.Tokens.RefreshToken == "" {
		t.Error("expected refresh token after login")
	}
	if loginResp.User.ID != created.ID {
		t.Errorf("expected user ID %s, got %s", created.ID, loginResp.User.ID)
	}

	t.Logf("Full registration flow completed for user %s", loginResp.User.ID)
}

// TestAuthService_TokenRefresh проверяет обновление токенов.
func TestAuthService_TokenRefresh(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	// Create verified user and login
	createVerifiedUser(t, env, "refresh@test.com", "password123", "Refresh User")
	loginResp, err := env.authService.Login(ctx, domain.LoginRequest{
		Email: "refresh@test.com", Password: "password123",
	})
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	originalRefreshToken := loginResp.Tokens.RefreshToken

	// Refresh token
	refreshResp, err := env.authService.RefreshToken(ctx, domain.RefreshTokenRequest{
		RefreshToken: originalRefreshToken,
	})
	if err != nil {
		t.Fatalf("RefreshToken failed: %v", err)
	}
	if refreshResp.Tokens.AccessToken == "" {
		t.Error("expected new access token")
	}
	if refreshResp.Tokens.RefreshToken == "" {
		t.Error("expected new refresh token")
	}
	if refreshResp.Tokens.RefreshToken == originalRefreshToken {
		t.Error("expected different refresh token after rotation")
	}

	// Old refresh token should be invalid
	_, err = env.authService.RefreshToken(ctx, domain.RefreshTokenRequest{
		RefreshToken: originalRefreshToken,
	})
	if err == nil {
		t.Error("expected error using old refresh token")
	}
}

// TestAuthService_Logout проверяет выход пользователя.
func TestAuthService_Logout(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	createVerifiedUser(t, env, "logout@test.com", "password123", "Logout User")
	loginResp, err := env.authService.Login(ctx, domain.LoginRequest{
		Email: "logout@test.com", Password: "password123",
	})
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	err = env.authService.Logout(ctx, domain.LogoutRequest{
		RefreshToken: loginResp.Tokens.RefreshToken,
	})
	if err != nil {
		t.Fatalf("Logout failed: %v", err)
	}

	// Using the refresh token after logout should fail
	_, err = env.authService.RefreshToken(ctx, domain.RefreshTokenRequest{
		RefreshToken: loginResp.Tokens.RefreshToken,
	})
	if err == nil {
		t.Error("expected error using refresh token after logout")
	}
}

// TestAuthService_PasswordReset проверяет сброс пароля.
func TestAuthService_PasswordReset(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	createVerifiedUser(t, env, "reset@test.com", "old-password", "Reset User")

	// Request password reset
	err := env.authService.RequestPasswordReset(ctx, domain.PasswordResetRequest{
		Email: "reset@test.com",
	})
	if err != nil {
		t.Fatalf("RequestPasswordReset failed: %v", err)
	}

	// The reset token would be in Valkey, but we can't access it from email.
	// Directly store for testing.
	if err := env.resetStore.Store(ctx, "reset@test.com", "test-reset-token"); err != nil {
		t.Fatalf("store reset token failed: %v", err)
	}

	// Reset password
	err = env.authService.ResetPassword(ctx, domain.ResetPasswordRequest{
		ResetToken:  "test-reset-token",
		NewPassword: "new-password",
	})
	if err != nil {
		t.Fatalf("ResetPassword failed: %v", err)
	}

	// Login with new password
	_, err = env.authService.Login(ctx, domain.LoginRequest{
		Email: "reset@test.com", Password: "new-password",
	})
	if err != nil {
		t.Fatalf("Login with new password failed: %v", err)
	}

	// Old password should not work
	_, err = env.authService.Login(ctx, domain.LoginRequest{
		Email: "reset@test.com", Password: "old-password",
	})
	if err == nil {
		t.Error("expected error logging in with old password")
	}
}

// TestUserService_GetProfile проверяет получение профиля.
func TestUserService_GetProfile(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	createVerifiedUser(t, env, "profile@test.com", "password123", "Profile User")
	user, _ := env.userRepo.GetByEmail(ctx, "profile@test.com")

	profile, err := env.userService.GetProfile(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetProfile failed: %v", err)
	}
	if profile.Email != "profile@test.com" {
		t.Errorf("expected email profile@test.com, got %s", profile.Email)
	}
	if profile.DisplayName != "Profile User" {
		t.Errorf("expected display name 'Profile User', got %s", profile.DisplayName)
	}
}

// TestUserService_UpdateProfile проверяет обновление профиля.
func TestUserService_UpdateProfile(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	createVerifiedUser(t, env, "update@test.com", "password123", "Old Name")
	user, _ := env.userRepo.GetByEmail(ctx, "update@test.com")

	newName := "New Name"
	updated, err := env.userService.UpdateProfile(ctx, domain.UpdateProfileRequest{
		UserID:      user.ID,
		DisplayName: &newName,
	})
	if err != nil {
		t.Fatalf("UpdateProfile failed: %v", err)
	}
	if updated.DisplayName != "New Name" {
		t.Errorf("expected display name 'New Name', got %s", updated.DisplayName)
	}
}

// TestUserService_ChangePassword проверяет смену пароля.
func TestUserService_ChangePassword(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	createVerifiedUser(t, env, "changepw@test.com", "old-pass", "ChangePW User")
	user, _ := env.userRepo.GetByEmail(ctx, "changepw@test.com")

	err := env.userService.ChangePassword(ctx, domain.ChangePasswordRequest{
		UserID:          user.ID,
		CurrentPassword: "old-pass",
		NewPassword:     "new-pass",
	})
	if err != nil {
		t.Fatalf("ChangePassword failed: %v", err)
	}

	// Verify login works with new password
	_, err = env.authService.Login(ctx, domain.LoginRequest{
		Email: "changepw@test.com", Password: "new-pass",
	})
	if err != nil {
		t.Fatalf("Login with new password failed: %v", err)
	}
}

// TestTokenService_ValidateToken проверяет валидацию токена.
func TestTokenService_ValidateToken(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	createVerifiedUser(t, env, "validate@test.com", "password123", "Validate User")
	loginResp, _ := env.authService.Login(ctx, domain.LoginRequest{
		Email: "validate@test.com", Password: "password123",
	})

	claims, err := env.tokenService.ValidateToken(ctx, loginResp.Tokens.AccessToken)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
	if claims.Email != "validate@test.com" {
		t.Errorf("expected email validate@test.com, got %s", claims.Email)
	}
}

// TestTokenService_ListSessions проверяет получение списка сессий.
func TestTokenService_ListSessions(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	createVerifiedUser(t, env, "sessions@test.com", "password123", "Sessions User")
	user, _ := env.userRepo.GetByEmail(ctx, "sessions@test.com")

	// Login twice
	_, _ = env.authService.Login(ctx, domain.LoginRequest{Email: "sessions@test.com", Password: "password123"})
	time.Sleep(100 * time.Millisecond)
	_, _ = env.authService.Login(ctx, domain.LoginRequest{Email: "sessions@test.com", Password: "password123"})

	sessions, err := env.tokenService.ListSessions(ctx, user.ID)
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}
	if len(sessions) != 2 {
		t.Errorf("expected 2 sessions, got %d", len(sessions))
	}
}

// TestTokenService_RevokeSession проверяет отзыв сессии.
func TestTokenService_RevokeSession(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	createVerifiedUser(t, env, "revoke@test.com", "password123", "Revoke User")
	user, _ := env.userRepo.GetByEmail(ctx, "revoke@test.com")
	loginResp, _ := env.authService.Login(ctx, domain.LoginRequest{Email: "revoke@test.com", Password: "password123"})

	claims, _ := env.tokenService.ValidateToken(ctx, loginResp.Tokens.AccessToken)
	sessionID := claims.SessionID

	err := env.tokenService.RevokeSession(ctx, user.ID, sessionID)
	if err != nil {
		t.Fatalf("RevokeSession failed: %v", err)
	}

	// Token should be invalid after revocation
	_, err = env.tokenService.ValidateToken(ctx, loginResp.Tokens.AccessToken)
	if err == nil {
		t.Error("expected error validating token after session revocation")
	}
}

// createVerifiedUser — helper для создания верифицированного пользователя.
func createVerifiedUser(t *testing.T, env *IntegrationTestEnv, email, password, displayName string) {
	t.Helper()
	ctx := context.Background()

	hash, err := env.passwordHasher.Hash(ctx, password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	now := time.Now()
	user := domain.User{
		Email:         email,
		PasswordHash:  hash,
		DisplayName:   displayName,
		Role:          domain.RoleDeveloper,
		Status:        domain.StatusActive,
		EmailVerified: true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := env.userRepo.Create(ctx, user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
}
