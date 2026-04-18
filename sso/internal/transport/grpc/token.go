package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/Be4Die/game-developer-hub/protos/sso/v1"
	"github.com/Be4Die/game-developer-hub/sso/internal/service"
)

// TokenHandler — gRPC обработчик для TokenService.
type TokenHandler struct {
	pb.UnimplementedTokenServiceServer
	svc *service.TokenService
}

// NewTokenHandler создаёт обработчик токенов.
func NewTokenHandler(svc *service.TokenService) *TokenHandler {
	return &TokenHandler{svc: svc}
}

// ValidateToken проверяет валидность access-токена и возвращает данные пользователя.
// Возвращает ErrInvalidToken или ErrTokenExpired при невалидном токене.
func (h *TokenHandler) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	if req.AccessToken == "" {
		return nil, status.Error(codes.InvalidArgument, "access_token is required")
	}

	claims, err := h.svc.ValidateToken(ctx, req.AccessToken)
	if err != nil {
		return nil, domainErrToStatus(err)
	}

	return &pb.ValidateTokenResponse{
		Valid: true,
		User: &pb.User{
			Id:    claims.UserID,
			Email: claims.Email,
			Role:  userRoleToProto(claims.Role),
		},
		SessionId: claims.SessionID,
	}, nil
}

// ListSessions возвращает список активных сессий пользователя.
func (h *TokenHandler) ListSessions(ctx context.Context, req *pb.ListSessionsRequest) (*pb.ListSessionsResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	sessions, err := h.svc.ListSessions(ctx, req.UserId)
	if err != nil {
		return nil, domainErrToStatus(err)
	}

	result := make([]*pb.Session, 0, len(sessions))
	for _, s := range sessions {
		result = append(result, sessionToProto(s))
	}

	return &pb.ListSessionsResponse{Sessions: result}, nil
}

// RevokeSession отзывает конкретную сессию пользователя.
// Возвращает ErrNotFound, если сессия не найдена.
func (h *TokenHandler) RevokeSession(ctx context.Context, req *pb.RevokeSessionRequest) (*pb.RevokeSessionResponse, error) {
	if req.UserId == "" || req.SessionId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id and session_id are required")
	}

	if err := h.svc.RevokeSession(ctx, req.UserId, req.SessionId); err != nil {
		return nil, domainErrToStatus(err)
	}

	return &pb.RevokeSessionResponse{Success: true}, nil
}

// RevokeAllSessions отзывает все сессии пользователя, опционально исключая указанную.
// Возвращает количество отозванных сессий.
func (h *TokenHandler) RevokeAllSessions(ctx context.Context, req *pb.RevokeAllSessionsRequest) (*pb.RevokeAllSessionsResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	excludeID := ""
	if req.ExcludeSessionId != nil {
		excludeID = *req.ExcludeSessionId
	}

	count, err := h.svc.RevokeAllSessions(ctx, req.UserId, excludeID)
	if err != nil {
		return nil, domainErrToStatus(err)
	}

	return &pb.RevokeAllSessionsResponse{RevokedCount: count}, nil
}
