package grpc

import (
	"errors"

	"github.com/Be4Die/game-developer-hub/moderation/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func domainError(err error, op string) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, domain.ErrNotFound):
		return status.Errorf(codes.NotFound, "%s: not found", op)
	case errors.Is(err, domain.ErrForbidden):
		return status.Errorf(codes.PermissionDenied, "%s: permission denied", op)
	case errors.Is(err, domain.ErrUnauthorized):
		return status.Errorf(codes.Unauthenticated, "%s: unauthenticated", op)
	case errors.Is(err, domain.ErrAlreadyClaimed):
		return status.Errorf(codes.AlreadyExists, "%s: request already claimed by another moderator", op)
	case errors.Is(err, domain.ErrInvalidStatus):
		return status.Errorf(codes.FailedPrecondition, "%s: invalid status transition", op)
	case errors.Is(err, domain.ErrEmptyReason):
		return status.Errorf(codes.InvalidArgument, "%s: rejection reason is required", op)
	case errors.Is(err, domain.ErrEmptyMessage):
		return status.Errorf(codes.InvalidArgument, "%s: message cannot be empty", op)
	case errors.Is(err, domain.ErrDeployFailed):
		return status.Errorf(codes.Internal, "%s: %v", op, err)
	default:
		return status.Errorf(codes.Internal, "%s: %v", op, err)
	}
}
