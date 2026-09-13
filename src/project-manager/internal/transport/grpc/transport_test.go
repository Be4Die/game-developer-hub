package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
	pb "github.com/Be4Die/game-developer-hub/protos/project_manager/v1"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestConverters_ProjectToProto(t *testing.T) {
	assert.Nil(t, projectToProto(nil))

	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	p := &domain.Project{
		ID:        123,
		OwnerID:   "user-1",
		Status:    domain.ProjectStatusPublished,
		IsOnline:  true,
		CreatedAt: now,
		UpdatedAt: now,
		Draft: &domain.Draft{
			ProjectID:          123,
			TitleRu:            "Название",
			TitleEn:            "Title",
			SeoRu:              "seo-ru",
			SeoEn:              "seo-en",
			AboutRu:            "О игре",
			AboutEn:            "About game",
			IconPath:           "/icon.png",
			CoverPath:          "/cover.png",
			VideoPath:          "/video.mp4",
			ActiveBuildVersion: "v1.0.0",
			DevURL:             "https://dev.local/123",
			IsOnline:           true,
			UpdatedAt:          now,
		},
		Release: &domain.Release{
			ID:          1,
			ProjectID:   123,
			Version:     "v1.0.0",
			TitleRu:     "Релиз",
			TitleEn:     "Release",
			SeoRu:       "seo-rel-ru",
			SeoEn:       "seo-rel-en",
			AboutRu:     "Описание релиза",
			AboutEn:     "About release",
			IconPath:    "/rel_icon.png",
			CoverPath:   "/rel_cover.png",
			VideoPath:   "/rel_video.mp4",
			ProdURL:     "https://prod.local/123",
			IsOnline:    true,
			PublishedAt: now,
		},
	}

	proto := projectToProto(p)
	require.NotNil(t, proto)
	assert.Equal(t, int64(123), proto.Id)
	assert.Equal(t, "user-1", proto.OwnerId)
	assert.Equal(t, pb.ProjectStatus_PROJECT_STATUS_PUBLISHED, proto.Status)
	assert.True(t, proto.IsOnline)
	assert.Equal(t, "Название", proto.TitleRu)
	assert.Equal(t, "https://prod.local/123", proto.ProdUrl)
	assert.Equal(t, "https://dev.local/123", proto.DevUrl)
	assert.NotNil(t, proto.Draft)
	assert.True(t, proto.Draft.IsOnline)
	assert.NotNil(t, proto.Release)
	assert.True(t, proto.Release.IsOnline)
}

func TestConverters_BuildAndReleaseToProto(t *testing.T) {
	assert.Nil(t, buildToProto(nil))
	assert.Nil(t, releaseToProto(nil))
	assert.Nil(t, draftToProto(nil))

	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	b := &domain.Build{
		ID:         5,
		ProjectID:  10,
		Version:    "v1.0.1",
		FilePath:   "/path/build.zip",
		FileSize:   2048,
		IsUnpacked: true,
		CreatedAt:  now,
	}

	pbBuild := buildToProto(b)
	require.NotNil(t, pbBuild)
	assert.Equal(t, int64(5), pbBuild.Id)
	assert.Equal(t, "v1.0.1", pbBuild.Version)
	assert.True(t, pbBuild.IsUnpacked)
}

func TestDomainError_Mapping(t *testing.T) {
	tests := []struct {
		err          error
		action       string
		expectedCode codes.Code
	}{
		{domain.ErrNotFound, "get", codes.NotFound},
		{domain.ErrAlreadyExists, "create", codes.AlreadyExists},
		{domain.ErrForbidden, "update", codes.PermissionDenied},
		{domain.ErrInvalidInput, "validate", codes.InvalidArgument},
		{domain.ErrInvalidArchive, "unpack", codes.InvalidArgument},
		{domain.ErrNoIndexHTML, "unpack", codes.InvalidArgument},
		{domain.ErrDisallowedFileType, "upload", codes.InvalidArgument},
		{domain.ErrDraftNotReady, "submit", codes.FailedPrecondition},
		{domain.ErrAlreadyInModeration, "submit", codes.AlreadyExists},
		{domain.ErrNoActiveBuild, "deploy", codes.FailedPrecondition},
		{errors.New("unknown error"), "custom", codes.Internal},
	}

	for _, tt := range tests {
		t.Run(tt.action+"_"+tt.expectedCode.String(), func(t *testing.T) {
			st := domainError(tt.err, tt.action)
			assert.Equal(t, tt.expectedCode, status.Code(st))
		})
	}
}

func TestJWTAuth_UnaryAndStream(t *testing.T) {
	secret := "test-secret-key-123456"
	issuer := "sso-service"

	_, err := NewJWTAuth("", issuer)
	require.Error(t, err)

	auth, err := NewJWTAuth(secret, issuer)
	require.NoError(t, err)

	// Generate valid token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  "user-999",
		"role": float64(2),
		"iss":  issuer,
		"exp":  time.Now().Add(time.Hour).Unix(),
	})
	tokenStr, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	t.Run("valid Bearer token in unary", func(t *testing.T) {
		md := metadata.Pairs("authorization", "Bearer "+tokenStr)
		ctx := metadata.NewIncomingContext(context.Background(), md)

		interceptor := auth.Unary()
		handlerCalled := false
		resp, err := interceptor(ctx, "req", &grpc.UnaryServerInfo{}, func(ctx context.Context, req any) (any, error) {
			handlerCalled = true
			userID, ok := UserIDFromContext(ctx)
			assert.True(t, ok)
			assert.Equal(t, "user-999", userID)

			role, ok := UserRoleFromContext(ctx)
			assert.True(t, ok)
			assert.Equal(t, 2, role)
			return "ok", nil
		})

		require.NoError(t, err)
		assert.True(t, handlerCalled)
		assert.Equal(t, "ok", resp)
	})

	t.Run("valid x-user-id header in unary", func(t *testing.T) {
		md := metadata.Pairs("x-user-id", "gateway-user-1", "x-user-role", "1")
		ctx := metadata.NewIncomingContext(context.Background(), md)

		interceptor := auth.Unary()
		handlerCalled := false
		_, err := interceptor(ctx, "req", &grpc.UnaryServerInfo{}, func(ctx context.Context, req any) (any, error) {
			handlerCalled = true
			userID, ok := UserIDFromContext(ctx)
			assert.True(t, ok)
			assert.Equal(t, "gateway-user-1", userID)

			role, ok := UserRoleFromContext(ctx)
			assert.True(t, ok)
			assert.Equal(t, 1, role)
			return "ok", nil
		})

		require.NoError(t, err)
		assert.True(t, handlerCalled)
	})

	t.Run("missing metadata", func(t *testing.T) {
		interceptor := auth.Unary()
		_, err := interceptor(context.Background(), "req", &grpc.UnaryServerInfo{}, func(ctx context.Context, req any) (any, error) {
			return nil, nil
		})
		require.Error(t, err)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("invalid token", func(t *testing.T) {
		md := metadata.Pairs("authorization", "Bearer invalid.jwt.token")
		ctx := metadata.NewIncomingContext(context.Background(), md)

		interceptor := auth.Unary()
		_, err := interceptor(ctx, "req", &grpc.UnaryServerInfo{}, func(ctx context.Context, req any) (any, error) {
			return nil, nil
		})
		require.Error(t, err)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})
}
