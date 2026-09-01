package grpc

import (
	"errors"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// domainError преобразует доменные ошибки в соответствующие статусы gRPC.
func domainError(err error, action string) error {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return status.Errorf(codes.NotFound, "%s: not found", action)
	case errors.Is(err, domain.ErrAlreadyExists):
		return status.Errorf(codes.AlreadyExists, "%s: already exists", action)
	case errors.Is(err, domain.ErrForbidden):
		return status.Errorf(codes.PermissionDenied, "%s: forbidden", action)
	case errors.Is(err, domain.ErrInvalidInput):
		return status.Errorf(codes.InvalidArgument, "%s: invalid input", action)
	case errors.Is(err, domain.ErrInvalidArchive):
		return status.Errorf(codes.InvalidArgument, "%s: invalid archive: %v", action, err)
	case errors.Is(err, domain.ErrNoIndexHTML):
		return status.Errorf(codes.InvalidArgument, "%s: index.html missing in archive", action)
	case errors.Is(err, domain.ErrDisallowedFileType):
		return status.Errorf(codes.InvalidArgument, "%s: disallowed file type: %v", action, err)
	case errors.Is(err, domain.ErrDraftNotReady):
		return status.Errorf(codes.FailedPrecondition, "%s: draft not ready", action)
	case errors.Is(err, domain.ErrAlreadyInModeration):
		return status.Errorf(codes.AlreadyExists, "%s: already in moderation", action)
	case errors.Is(err, domain.ErrNoActiveBuild):
		return status.Errorf(codes.FailedPrecondition, "%s: no active build", action)
	default:
		return status.Errorf(codes.Internal, "%s: %v", action, err)
	}
}
