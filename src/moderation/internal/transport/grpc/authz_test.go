package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	pb "github.com/Be4Die/game-developer-hub/protos/moderation/v1"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func signedToken(t *testing.T, secret, issuer, sub string, role float64) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": sub, "role": role, "iss": issuer, "exp": time.Now().Add(time.Hour).Unix(),
	})
	s, err := token.SignedString([]byte(secret))
	require.NoError(t, err)
	return s
}

func TestJWTAuth_DoesNotTrustClientHeaders(t *testing.T) {
	const secret, issuer = "test-secret-moderation", "sso"

	auth, err := NewJWTAuth(secret, issuer)
	require.NoError(t, err)
	auth = auth.WithInternalToken("service-secret")

	noop := func(ctx context.Context, _ any) (any, error) { return "ok", nil }

	t.Run("headers without token are rejected", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-user-id", "anyone", "x-user-role", "3"))
		_, err := auth.Unary()(ctx, "req", &grpc.UnaryServerInfo{}, noop)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("token signed with another secret is rejected", func(t *testing.T) {
		forged := signedToken(t, "attacker-secret", issuer, "attacker", 3)
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+forged))
		_, err := auth.Unary()(ctx, "req", &grpc.UnaryServerInfo{}, noop)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("token without expiry is rejected", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "dev-1", "role": float64(1), "iss": issuer})
		eternal, err := token.SignedString([]byte(secret))
		require.NoError(t, err)
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+eternal))
		_, err = auth.Unary()(ctx, "req", &grpc.UnaryServerInfo{}, noop)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("spoofed role and internal marker are overwritten", func(t *testing.T) {
		devToken := signedToken(t, secret, issuer, "dev-1", 1)
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
			"authorization", "Bearer "+devToken,
			"x-user-id", "mod-1",
			"x-user-role", "admin",
			"x-internal-call", "1",
			"x-internal-token", "wrong",
		))
		_, err := auth.Unary()(ctx, "req", &grpc.UnaryServerInfo{}, func(ctx context.Context, _ any) (any, error) {
			uid, _ := UserIDFromContext(ctx)
			assert.Equal(t, "dev-1", uid)
			assert.Equal(t, "developer", UserRoleFromContext(ctx))
			assert.False(t, IsStaff(ctx))
			assert.False(t, IsInternalCall(ctx))
			return "ok", nil
		})
		require.NoError(t, err)
	})

	t.Run("valid internal token marks the call", func(t *testing.T) {
		devToken := signedToken(t, secret, issuer, "dev-1", 1)
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
			"authorization", "Bearer "+devToken, "x-internal-token", "service-secret",
		))
		_, err := auth.Unary()(ctx, "req", &grpc.UnaryServerInfo{}, func(ctx context.Context, _ any) (any, error) {
			assert.True(t, IsInternalCall(ctx))
			return "ok", nil
		})
		require.NoError(t, err)
	})
}

type denyAccess struct{}

func (denyAccess) CheckAccess(context.Context, int64) error { return errors.New("forbidden") }

// Проверки прав выполняются до обращения к сервису, поэтому сервис в этих тестах не нужен.
func TestModerationHandler_Authorization(t *testing.T) {
	devCtx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-user-id", "dev-1", "x-user-role", "developer"))

	h := NewModerationHandler(nil).WithAccessChecker(denyAccess{}).WithInternalOnly()

	staffOnly := map[string]func() error{
		"ListRequests": func() error { _, err := h.ListRequests(devCtx, &pb.ListModerationRequestsRequest{}); return err },
		"GetSnapshot":  func() error { _, err := h.GetSnapshot(devCtx, &pb.GetSnapshotRequest{RequestId: 1}); return err },
		"ClaimRequest": func() error {
			_, err := h.ClaimRequest(devCtx, &pb.ClaimModerationRequestRequest{RequestId: 1})
			return err
		},
		"Approve": func() error { _, err := h.Approve(devCtx, &pb.ApproveModerationRequest{ProjectId: 1}); return err },
		"Reject":  func() error { _, err := h.Reject(devCtx, &pb.RejectModerationRequest{ProjectId: 1}); return err },
		"ReviewServerAccess": func() error {
			_, err := h.ReviewServerAccess(devCtx, &pb.ReviewServerAccessRequest{RequestId: 1, Approved: true})
			return err
		},
		"ListActiveChats": func() error { _, err := h.ListActiveChats(devCtx, &pb.ListActiveChatsRequest{}); return err },
		"PurgeProjectMedia": func() error {
			_, err := h.PurgeProjectMedia(devCtx, &pb.PurgeProjectMediaRequest{ProjectId: 1})
			return err
		},
	}
	for name, call := range staffOnly {
		t.Run("developer cannot call "+name, func(t *testing.T) {
			assert.Equal(t, codes.PermissionDenied, status.Code(call()))
		})
	}

	foreignProject := map[string]func() error{
		"ListMessages": func() error { _, err := h.ListMessages(devCtx, &pb.ListChatMessagesRequest{ProjectId: 7}); return err },
		"SendMessage": func() error {
			_, err := h.SendMessage(devCtx, &pb.SendChatMessageRequest{ProjectId: 7, Content: "hi"})
			return err
		},
		"GetLatestRequestByProject": func() error {
			_, err := h.GetLatestRequestByProject(devCtx, &pb.GetLatestRequestByProjectRequest{ProjectId: 7})
			return err
		},
		"GetServerAccess": func() error { _, err := h.GetServerAccess(devCtx, &pb.GetServerAccessRequest{ProjectId: 7}); return err },
		"SubmitServerAccess": func() error {
			_, err := h.SubmitServerAccess(devCtx, &pb.SubmitServerAccessRequest{ProjectId: 7})
			return err
		},
		"GetAttachment": func() error {
			_, err := h.GetAttachment(devCtx, &pb.GetAttachmentRequest{ProjectId: 7, AttachmentId: "a"})
			return err
		},
	}
	for name, call := range foreignProject {
		t.Run("developer without project access cannot call "+name, func(t *testing.T) {
			assert.Equal(t, codes.PermissionDenied, status.Code(call()))
		})
	}

	serviceOnly := map[string]func() error{
		"SubmitDraft": func() error { _, err := h.SubmitDraft(devCtx, &pb.SubmitDraftRequest{ProjectId: 7}); return err },
		"RegisterAttachment": func() error {
			_, err := h.RegisterAttachment(devCtx, &pb.RegisterAttachmentRequest{ProjectId: 7})
			return err
		},
	}
	for name, call := range serviceOnly {
		t.Run("user cannot call service method "+name, func(t *testing.T) {
			assert.Equal(t, codes.PermissionDenied, status.Code(call()))
		})
	}
}
