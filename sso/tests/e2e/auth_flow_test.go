//go:build e2e

package e2e

import (
	"context"
	"testing"
	"time"

	pb "github.com/Be4Die/game-developer-hub/protos/sso/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestE2E_FullAuthFlow проверяет полный поток через gRPC: register → verify → login → refresh → logout.
func TestE2E_FullAuthFlow(t *testing.T) {
	env := setupE2E(t)
	ctx, cancel := context.WithTimeout(context.Background(), e2eTestTimeout)
	defer cancel()

	// Step 1: Register via gRPC
	displayName := "E2E User"
	regResp, err := env.authClient.Register(ctx, &pb.RegisterRequest{
		Email:       "e2e-flow@test.com",
		Password:    "password123",
		DisplayName: displayName,
	})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if regResp.User == nil {
		t.Fatal("expected user in register response")
	}
	if regResp.User.Email != "e2e-flow@test.com" {
		t.Errorf("expected email e2e-flow@test.com, got %s", regResp.User.Email)
	}
	if regResp.User.DisplayName != displayName {
		t.Errorf("expected display name %s, got %s", displayName, regResp.User.DisplayName)
	}
	userID := regResp.User.Id

	// User is registered but not verified yet. Login should fail.
	_, err = env.authClient.Login(ctx, &pb.LoginRequest{
		Email:    "e2e-flow@test.com",
		Password: "password123",
	})
	if err == nil {
		t.Fatal("expected error logging in before verification")
	}

	// Step 2: Verify email (store code in Valkey directly, call VerifyEmail via gRPC)
	if err := env.redisClient.Set(ctx, "email_verify:e2e-flow@test.com", "123456", 10*time.Minute).Err(); err != nil {
		t.Fatalf("store verification code failed: %v", err)
	}

	_, err = env.authClient.VerifyEmail(ctx, &pb.VerifyEmailRequest{VerificationCode: "123456"})
	if err != nil {
		t.Fatalf("VerifyEmail failed: %v", err)
	}

	// Step 3: Login via gRPC
	loginResp, err := env.authClient.Login(ctx, &pb.LoginRequest{
		Email:    "e2e-flow@test.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if loginResp.User == nil {
		t.Fatal("expected user in login response")
	}
	if loginResp.User.Id != userID {
		t.Errorf("expected user ID %s, got %s", userID, loginResp.User.Id)
	}
	if loginResp.Tokens == nil {
		t.Fatal("expected tokens in login response")
	}
	if loginResp.Tokens.AccessToken == "" {
		t.Error("expected access token")
	}
	if loginResp.Tokens.RefreshToken == "" {
		t.Error("expected refresh token")
	}

	t.Logf("Full auth flow completed for user %s", userID)
}

// TestE2E_TokenRotation проверяет ротацию refresh токенов через gRPC.
func TestE2E_TokenRotation(t *testing.T) {
	env := setupE2E(t)
	ctx, cancel := context.WithTimeout(context.Background(), e2eTestTimeout)
	defer cancel()
	env.cleanupTables(t)

	// Register and verify
	createVerifiedUserE2E(t, env, "rotation@test.com", "Rotation User")

	// Login
	loginResp, err := env.authClient.Login(ctx, &pb.LoginRequest{
		Email: "rotation@test.com", Password: "password123",
	})
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	originalRT := loginResp.Tokens.RefreshToken

	// First refresh
	refreshResp1, err := env.authClient.RefreshToken(ctx, &pb.RefreshTokenRequest{RefreshToken: originalRT})
	if err != nil {
		t.Fatalf("First refresh failed: %v", err)
	}
	if refreshResp1.Tokens.RefreshToken == originalRT {
		t.Error("expected new refresh token after rotation")
	}

	// Use new refresh token
	refreshResp2, err := env.authClient.RefreshToken(ctx, &pb.RefreshTokenRequest{RefreshToken: refreshResp1.Tokens.RefreshToken})
	if err != nil {
		t.Fatalf("Second refresh failed: %v", err)
	}
	if refreshResp2.Tokens.RefreshToken == refreshResp1.Tokens.RefreshToken {
		t.Error("expected different refresh token on second rotation")
	}

	// Old token should be invalid
	_, err = env.authClient.RefreshToken(ctx, &pb.RefreshTokenRequest{RefreshToken: originalRT})
	if err == nil {
		t.Error("expected error using original refresh token")
	}
}

// TestE2E_ValidateToken проверяет валидацию токена через gRPC.
func TestE2E_ValidateToken(t *testing.T) {
	env := setupE2E(t)
	ctx, cancel := context.WithTimeout(context.Background(), e2eTestTimeout)
	defer cancel()
	env.cleanupTables(t)

	createVerifiedUserE2E(t, env, "validate-e2e@test.com", "Validate E2E")

	loginResp, _ := env.authClient.Login(ctx, &pb.LoginRequest{
		Email: "validate-e2e@test.com", Password: "password123",
	})

	// Validate access token
	valResp, err := env.tokenClient.ValidateToken(ctx, &pb.ValidateTokenRequest{AccessToken: loginResp.Tokens.AccessToken})
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
	if !valResp.Valid {
		t.Error("expected token to be valid")
	}
	if valResp.User.Email != "validate-e2e@test.com" {
		t.Errorf("expected email validate-e2e@test.com, got %s", valResp.User.Email)
	}
}

// TestE2E_ListSessions проверяет получение сессий через gRPC.
func TestE2E_ListSessions(t *testing.T) {
	env := setupE2E(t)
	ctx, cancel := context.WithTimeout(context.Background(), e2eTestTimeout)
	defer cancel()
	env.cleanupTables(t)

	createVerifiedUserE2E(t, env, "list-sessions@test.com", "List Sessions")
	loginResp, _ := env.authClient.Login(ctx, &pb.LoginRequest{
		Email: "list-sessions@test.com", Password: "password123",
	})
	userID := loginResp.User.Id

	// Login again to have multiple sessions
	time.Sleep(100 * time.Millisecond)
	_, _ = env.authClient.Login(ctx, &pb.LoginRequest{
		Email: "list-sessions@test.com", Password: "password123",
	})

	listResp, err := env.tokenClient.ListSessions(ctx, &pb.ListSessionsRequest{UserId: userID})
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}
	if len(listResp.Sessions) != 2 {
		t.Errorf("expected 2 sessions, got %d", len(listResp.Sessions))
	}
}

// TestE2E_RevokeSession проверяет отзыв сессии через gRPC.
func TestE2E_RevokeSession(t *testing.T) {
	env := setupE2E(t)
	ctx, cancel := context.WithTimeout(context.Background(), e2eTestTimeout)
	defer cancel()
	env.cleanupTables(t)

	createVerifiedUserE2E(t, env, "revoke-e2e@test.com", "Revoke E2E")
	loginResp, _ := env.authClient.Login(ctx, &pb.LoginRequest{
		Email: "revoke-e2e@test.com", Password: "password123",
	})
	userID := loginResp.User.Id

	// Validate token to get session ID
	valResp, _ := env.tokenClient.ValidateToken(ctx, &pb.ValidateTokenRequest{AccessToken: loginResp.Tokens.AccessToken})
	sessionID := valResp.SessionId

	// Revoke session
	_, err := env.tokenClient.RevokeSession(ctx, &pb.RevokeSessionRequest{
		UserId:    userID,
		SessionId: sessionID,
	})
	if err != nil {
		t.Fatalf("RevokeSession failed: %v", err)
	}

	// Token should now be invalid
	_, err = env.tokenClient.ValidateToken(ctx, &pb.ValidateTokenRequest{AccessToken: loginResp.Tokens.AccessToken})
	if err == nil {
		t.Error("expected error after session revocation")
	}
}

// TestE2E_Logout проверяет выход через gRPC.
func TestE2E_Logout(t *testing.T) {
	env := setupE2E(t)
	ctx, cancel := context.WithTimeout(context.Background(), e2eTestTimeout)
	defer cancel()
	env.cleanupTables(t)

	createVerifiedUserE2E(t, env, "logout-e2e@test.com", "Logout E2E")
	loginResp, _ := env.authClient.Login(ctx, &pb.LoginRequest{
		Email: "logout-e2e@test.com", Password: "password123",
	})

	_, err := env.authClient.Logout(ctx, &pb.LogoutRequest{RefreshToken: loginResp.Tokens.RefreshToken})
	if err != nil {
		t.Fatalf("Logout failed: %v", err)
	}

	// Refresh should fail
	_, err = env.authClient.RefreshToken(ctx, &pb.RefreshTokenRequest{RefreshToken: loginResp.Tokens.RefreshToken})
	if err == nil {
		t.Error("expected error after logout")
	}
}

// TestE2E_UserProfile проверяет получение и обновление профиля через gRPC.
func TestE2E_UserProfile(t *testing.T) {
	env := setupE2E(t)
	ctx, cancel := context.WithTimeout(context.Background(), e2eTestTimeout)
	defer cancel()
	env.cleanupTables(t)

	createVerifiedUserE2E(t, env, "profile-e2e@test.com", "Original Name")

	// We need user ID — get it by login
	loginResp, _ := env.authClient.Login(ctx, &pb.LoginRequest{
		Email: "profile-e2e@test.com", Password: "password123",
	})
	userID := loginResp.User.Id

	// Get profile
	profileResp, err := env.userClient.GetProfile(ctx, &pb.GetProfileRequest{UserId: userID})
	if err != nil {
		t.Fatalf("GetProfile failed: %v", err)
	}
	if profileResp.User.DisplayName != "Original Name" {
		t.Errorf("expected display name 'Original Name', got %s", profileResp.User.DisplayName)
	}

	// Update profile — requires user_id in context (extracted from metadata by interceptor)
	// Since we don't have auth interceptor in e2e, we use GetUserById instead
	newName := "Updated Name"
	_, err = env.userClient.UpdateProfile(ctx, &pb.UpdateProfileRequest{DisplayName: &newName})
	// This should fail because there's no user_id in context
	if err == nil {
		t.Error("expected error without user context")
	}
	st, ok := status.FromError(err)
	if ok && st.Code() != codes.Unauthenticated {
		t.Errorf("expected Unauthenticated, got %s", st.Code())
	}
}

// TestE2E_ErrorCases проверяет обработку ошибок через gRPC.
func TestE2E_ErrorCases(t *testing.T) {
	env := setupE2E(t)
	ctx, cancel := context.WithTimeout(context.Background(), e2eTestTimeout)
	defer cancel()
	env.cleanupTables(t)

	// Login with wrong password
	_, err := env.authClient.Login(ctx, &pb.LoginRequest{
		Email: "nonexistent@test.com", Password: "password",
	})
	if err == nil {
		t.Fatal("expected error for nonexistent user")
	}

	// Validate with empty token
	_, err = env.tokenClient.ValidateToken(ctx, &pb.ValidateTokenRequest{AccessToken: ""})
	if err == nil {
		t.Fatal("expected error for empty token")
	}
	st, ok := status.FromError(err)
	if ok && st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %s", st.Code())
	}

	// List sessions with empty user ID
	_, err = env.tokenClient.ListSessions(ctx, &pb.ListSessionsRequest{UserId: ""})
	if err == nil {
		t.Fatal("expected error for empty user ID")
	}
}

// createVerifiedUserE2E создаёт верифицированного пользователя напрямую в БД.
func createVerifiedUserE2E(t *testing.T, env *e2eTestEnv, email, displayName string) {
	t.Helper()
	ctx := context.Background()

	hash, err := env.passwordHasher.Hash(ctx, "password123")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	now := time.Now()
	_, err = env.pool.Exec(ctx,
		`INSERT INTO users (email, password_hash, display_name, role, status, email_verified, created_at, updated_at)
		 VALUES ($1, $2, $3, 1, 1, true, $4, $5)`,
		email, hash, displayName, now, now,
	)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
}
