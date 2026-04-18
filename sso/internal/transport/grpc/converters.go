// Package grpc предоставляет gRPC обработчики и конвертеры для SSO-сервиса.
package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/Be4Die/game-developer-hub/protos/sso/v1"
	"github.com/Be4Die/game-developer-hub/sso/internal/domain"
)

// ─── Domain → Proto ─────────────────────────────────────────────

func userToProto(u domain.User) *pb.User {
	return &pb.User{
		Id:            u.ID,
		Email:         u.Email,
		DisplayName:   u.DisplayName,
		Role:          userRoleToProto(u.Role),
		Status:        userStatusToProto(u.Status),
		EmailVerified: u.EmailVerified,
		CreatedAt:     timestamppb.New(u.CreatedAt),
		UpdatedAt:     timestamppb.New(u.UpdatedAt),
	}
}

func tokenInfoToProto(t domain.TokenPair) *pb.TokenInfo {
	return &pb.TokenInfo{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
		ExpiresAt:    timestamppb.New(t.ExpiresAt),
		TokenType:    t.TokenType,
	}
}

func sessionToProto(s domain.Session) *pb.Session {
	return &pb.Session{
		SessionId:  s.ID,
		UserId:     s.UserID,
		UserAgent:  s.UserAgent,
		IpAddress:  s.IPAddress,
		CreatedAt:  timestamppb.New(s.CreatedAt),
		LastUsedAt: timestamppb.New(s.LastUsedAt),
		ExpiresAt:  timestamppb.New(s.ExpiresAt),
	}
}

func userRoleToProto(r domain.UserRole) pb.UserRole {
	switch r {
	case domain.RoleDeveloper:
		return pb.UserRole_USER_ROLE_DEVELOPER
	case domain.RoleModerator:
		return pb.UserRole_USER_ROLE_MODERATOR
	case domain.RoleAdmin:
		return pb.UserRole_USER_ROLE_ADMIN
	default:
		return pb.UserRole_USER_ROLE_UNSPECIFIED
	}
}

func userStatusToProto(s domain.UserStatus) pb.UserStatus {
	switch s {
	case domain.StatusActive:
		return pb.UserStatus_USER_STATUS_ACTIVE
	case domain.StatusSuspended:
		return pb.UserStatus_USER_STATUS_SUSPENDED
	case domain.StatusDeleted:
		return pb.UserStatus_USER_STATUS_DELETED
	default:
		return pb.UserStatus_USER_STATUS_UNSPECIFIED
	}
}

// ─── Proto → Domain ─────────────────────────────────────────────

// ─── Errors ─────────────────────────────────────────────────────

func domainErrToStatus(err error) error {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return status.Errorf(codes.NotFound, "%v", err)
	case errors.Is(err, domain.ErrAlreadyExists):
		return status.Errorf(codes.AlreadyExists, "%v", err)
	case errors.Is(err, domain.ErrInvalidPassword):
		return status.Errorf(codes.Unauthenticated, "%v", err)
	case errors.Is(err, domain.ErrInvalidToken):
		return status.Errorf(codes.Unauthenticated, "%v", err)
	case errors.Is(err, domain.ErrTokenExpired):
		return status.Errorf(codes.Unauthenticated, "%v", err)
	case errors.Is(err, domain.ErrEmailNotVerified):
		return status.Errorf(codes.FailedPrecondition, "%v", err)
	case errors.Is(err, domain.ErrUserSuspended):
		return status.Errorf(codes.PermissionDenied, "%v", err)
	default:
		return status.Errorf(codes.Internal, "%v", err)
	}
}

// extractUserIDFromContext извлекает user_id из gRPC metadata.
// В production это должен делать auth interceptor, который парсит JWT
// и добавляет user_id в контекст.
func extractUserIDFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	values := md.Get("x-user-id")
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
