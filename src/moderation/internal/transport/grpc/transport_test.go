package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/moderation/internal/domain"
	pb "github.com/Be4Die/game-developer-hub/protos/moderation/v1"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestConverters_RequestAndMessageToProto(t *testing.T) {
	assert.Nil(t, requestToProto(nil))
	assert.Nil(t, messageToProto(nil))

	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	reqID := int64(99)

	req := &domain.ModerationRequest{
		ID:              1,
		ProjectID:       10,
		OwnerID:         "dev-1",
		ModeratorID:     "mod-1",
		Status:          domain.RequestStatusApproved,
		RejectionReason: "",
		SubmittedAt:     now,
		StartedReviewAt: &now,
		ResolvedAt:      &now,
		Snapshot: domain.ProjectSnapshot{
			ProjectID:          10,
			TitleRu:            "Тест",
			TitleEn:            "Test",
			ActiveBuildVersion: "1.0.0",
		},
	}

	protoReq := requestToProto(req)
	require.NotNil(t, protoReq)
	assert.Equal(t, int64(1), protoReq.Id)
	assert.Equal(t, pb.RequestStatus_REQUEST_STATUS_APPROVED, protoReq.Status)
	assert.Equal(t, "Тест", protoReq.Snapshot.TitleRu)

	msg := &domain.ChatMessage{
		ID:          5,
		ProjectID:   10,
		RequestID:   &reqID,
		SenderID:    "mod-1",
		SenderRole:  domain.SenderRoleModerator,
		MessageType: domain.MessageTypeText,
		Content:     "Hello world",
		CreatedAt:   now,
	}

	protoMsg := messageToProto(msg)
	require.NotNil(t, protoMsg)
	assert.Equal(t, int64(5), protoMsg.Id)
	assert.Equal(t, pb.SenderRole_SENDER_ROLE_MODERATOR, protoMsg.SenderRole)
	assert.Equal(t, "Hello world", protoMsg.Content)
	assert.Equal(t, reqID, protoMsg.RequestId)

	snap := snapshotFromProto(protoReq.Snapshot)
	assert.Equal(t, int64(10), snap.ProjectID)
	assert.Equal(t, "Тест", snap.TitleRu)
}

func TestDomainError_Mapping(t *testing.T) {
	tests := []struct {
		err          error
		action       string
		expectedCode codes.Code
	}{
		{domain.ErrNotFound, "get", codes.NotFound},
		{domain.ErrForbidden, "update", codes.PermissionDenied},
		{domain.ErrUnauthorized, "auth", codes.Unauthenticated},
		{domain.ErrAlreadyClaimed, "claim", codes.AlreadyExists},
		{domain.ErrInvalidStatus, "transition", codes.FailedPrecondition},
		{domain.ErrEmptyReason, "reject", codes.InvalidArgument},
		{domain.ErrEmptyMessage, "message", codes.InvalidArgument},
		{domain.ErrDeployFailed, "deploy", codes.Internal},
		{errors.New("generic"), "action", codes.Internal},
	}

	for _, tt := range tests {
		t.Run(tt.action+"_"+tt.expectedCode.String(), func(t *testing.T) {
			st := domainError(tt.err, tt.action)
			assert.Equal(t, tt.expectedCode, status.Code(st))
		})
	}
}

func TestJWTAuth_Unary(t *testing.T) {
	secret := "test-secret-moderation"
	issuer := "sso-service"

	_, err := NewJWTAuth("", issuer)
	require.Error(t, err)

	auth, err := NewJWTAuth(secret, issuer)
	require.NoError(t, err)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  "mod-101",
		"role": float64(2),
		"iss":  issuer,
		"exp":  time.Now().Add(time.Hour).Unix(),
	})
	tokenStr, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	md := metadata.Pairs("authorization", "Bearer "+tokenStr)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	interceptor := auth.Unary()
	called := false
	_, err = interceptor(ctx, "req", &grpc.UnaryServerInfo{}, func(ctx context.Context, req any) (any, error) {
		called = true
		uid, ok := UserIDFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, "mod-101", uid)

		role := UserRoleFromContext(ctx)
		assert.Equal(t, "moderator", role)
		return "ok", nil
	})
	require.NoError(t, err)
	assert.True(t, called)
}
