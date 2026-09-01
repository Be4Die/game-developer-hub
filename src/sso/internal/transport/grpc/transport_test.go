package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	pb "github.com/Be4Die/game-developer-hub/protos/sso/v1"
	"github.com/Be4Die/game-developer-hub/sso/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type mockTokenManager struct {
	claims *domain.Claims
	err    error
}

func (m *mockTokenManager) GenerateAccessToken(ctx context.Context, claims domain.Claims) (string, time.Time, error) {
	return "access-token", time.Now().Add(time.Hour), nil
}

func (m *mockTokenManager) GenerateRefreshToken(ctx context.Context) (string, error) {
	return "refresh-token", nil
}

func (m *mockTokenManager) ParseAccessToken(ctx context.Context, token string) (*domain.Claims, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.claims, nil
}

func TestConverters_SSO(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	u := domain.User{
		ID:            "user-100",
		Email:         "dev@example.com",
		DisplayName:   "Developer",
		Role:          domain.RoleDeveloper,
		Status:        domain.StatusActive,
		EmailVerified: true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	protoUser := userToProto(u)
	assert.Equal(t, "user-100", protoUser.Id)
	assert.Equal(t, pb.UserRole_USER_ROLE_DEVELOPER, protoUser.Role)
	assert.Equal(t, pb.UserStatus_USER_STATUS_ACTIVE, protoUser.Status)
	assert.True(t, protoUser.EmailVerified)

	sess := domain.Session{
		ID:         "sess-1",
		UserID:     "user-100",
		UserAgent:  "Mozilla",
		IPAddress:  "127.0.0.1",
		CreatedAt:  now,
		LastUsedAt: now,
		ExpiresAt:  now.Add(time.Hour),
	}

	protoSess := sessionToProto(sess)
	assert.Equal(t, "sess-1", protoSess.SessionId)
	assert.Equal(t, "user-100", protoSess.UserId)

	tokens := domain.TokenPair{
		AccessToken:  "at",
		RefreshToken: "rt",
		ExpiresAt:    now.Add(time.Hour),
		TokenType:    "Bearer",
	}

	protoTokens := tokenInfoToProto(tokens)
	assert.Equal(t, "at", protoTokens.AccessToken)
	assert.Equal(t, "Bearer", protoTokens.TokenType)
}

func TestDomainErrToStatus(t *testing.T) {
	tests := []struct {
		err          error
		expectedCode codes.Code
	}{
		{domain.ErrNotFound, codes.NotFound},
		{domain.ErrAlreadyExists, codes.AlreadyExists},
		{domain.ErrInvalidPassword, codes.Unauthenticated},
		{domain.ErrInvalidToken, codes.Unauthenticated},
		{domain.ErrTokenExpired, codes.Unauthenticated},
		{domain.ErrEmailNotVerified, codes.FailedPrecondition},
		{domain.ErrUserSuspended, codes.PermissionDenied},
		{domain.ErrModeratorManagedByAdmin, codes.PermissionDenied},
		{domain.ErrProfileImmutable, codes.PermissionDenied},
		{domain.ErrPermissionDenied, codes.PermissionDenied},
		{domain.ErrCannotModifyAdminStatus, codes.PermissionDenied},
		{domain.ErrCannotModifyModeratorStatus, codes.PermissionDenied},
		{domain.ErrCannotDeleteAdmin, codes.PermissionDenied},
		{errors.New("unknown"), codes.Internal},
	}

	for _, tt := range tests {
		st := domainErrToStatus(tt.err)
		assert.Equal(t, tt.expectedCode, status.Code(st))
	}
}

func TestJWTAuthInterceptor_Unary(t *testing.T) {
	tm := &mockTokenManager{
		claims: &domain.Claims{
			UserID: "user-42",
			Role:   domain.RoleDeveloper,
		},
	}

	interceptor := NewJWTAuth(tm)

	t.Run("public method bypasses auth", func(t *testing.T) {
		info := &grpc.UnaryServerInfo{FullMethod: "/sso.v1.AuthService/Login"}
		called := false
		_, err := interceptor.Unary()(context.Background(), "req", info, func(ctx context.Context, req any) (any, error) {
			called = true
			return "ok", nil
		})
		require.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("protected method with valid token", func(t *testing.T) {
		info := &grpc.UnaryServerInfo{FullMethod: "/sso.v1.UserService/GetProfile"}
		md := metadata.Pairs("authorization", "Bearer valid-token")
		ctx := metadata.NewIncomingContext(context.Background(), md)

		called := false
		_, err := interceptor.Unary()(ctx, "req", info, func(ctx context.Context, req any) (any, error) {
			called = true
			uid := extractUserIDFromContext(ctx)
			assert.Equal(t, "user-42", uid)
			return "ok", nil
		})
		require.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("missing authorization", func(t *testing.T) {
		info := &grpc.UnaryServerInfo{FullMethod: "/sso.v1.UserService/GetProfile"}
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs())

		_, err := interceptor.Unary()(ctx, "req", info, func(ctx context.Context, req any) (any, error) {
			return nil, nil
		})
		require.Error(t, err)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})
}
