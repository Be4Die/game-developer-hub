package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/Be4Die/game-developer-hub/protos/sso/v1"
	"github.com/Be4Die/game-developer-hub/sso/internal/domain"
	"github.com/Be4Die/game-developer-hub/sso/internal/service"
)

// UserHandler — gRPC обработчик для UserService.
type UserHandler struct {
	pb.UnimplementedUserServiceServer
	svc *service.UserService
}

// NewUserHandler создаёт обработчик пользователей.
func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// GetProfile возвращает профиль текущего пользователя (по JWT).
// UserID извлекается из контекста, установленного JWT interceptor.
func (h *UserHandler) GetProfile(ctx context.Context, _ *pb.UserServiceGetProfileRequest) (*pb.UserServiceGetProfileResponse, error) {
	// UserID извлекается из JWT контекста interceptor'ом.
	userID := extractUserIDFromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "missing user context")
	}

	user, err := h.svc.GetProfile(ctx, userID)
	if err != nil {
		return nil, domainErrToStatus(err)
	}

	return &pb.UserServiceGetProfileResponse{User: userToProto(user)}, nil
}

// UpdateProfile обновляет данные профиля текущего пользователя.
// Требует валидный user_id в контексте (из JWT).
func (h *UserHandler) UpdateProfile(ctx context.Context, req *pb.UserServiceUpdateProfileRequest) (*pb.UserServiceUpdateProfileResponse, error) {
	// user_id должен извлекаться из JWT контекста interceptor'ом.
	// Для простоты ожидаем что caller передаёт user_id через metadata.
	userID := extractUserIDFromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "missing user context")
	}

	updateReq := domain.UpdateProfileRequest{
		UserID: userID,
	}
	if req.DisplayName != nil {
		updateReq.DisplayName = req.DisplayName
	}

	user, err := h.svc.UpdateProfile(ctx, updateReq)
	if err != nil {
		return nil, domainErrToStatus(err)
	}

	return &pb.UserServiceUpdateProfileResponse{User: userToProto(user)}, nil
}

// ChangePassword изменяет пароль текущего пользователя.
// Требует валидный user_id в контексте и текущий пароль.
// Возвращает ErrInvalidPassword при неверном текущем пароле.
func (h *UserHandler) ChangePassword(ctx context.Context, req *pb.UserServiceChangePasswordRequest) (*pb.UserServiceChangePasswordResponse, error) {
	userID := extractUserIDFromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "missing user context")
	}
	if req.CurrentPassword == "" || req.NewPassword == "" {
		return nil, status.Error(codes.InvalidArgument, "current_password and new_password are required")
	}

	if err := h.svc.ChangePassword(ctx, domain.ChangePasswordRequest{
		UserID:          userID,
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	}); err != nil {
		return nil, domainErrToStatus(err)
	}

	return &pb.UserServiceChangePasswordResponse{Success: true}, nil
}

// Get возвращает данные пользователя по идентификатору.
// Возвращает ErrNotFound, если пользователь не найден.
func (h *UserHandler) Get(ctx context.Context, req *pb.UserServiceGetRequest) (*pb.UserServiceGetResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	user, err := h.svc.GetUserByID(ctx, req.UserId)
	if err != nil {
		return nil, domainErrToStatus(err)
	}

	return &pb.UserServiceGetResponse{User: userToProto(user)}, nil
}

// SearchUsers выполняет поиск пользователей по запросу с пагинацией.
// Возвращает до 100 пользователей за один запрос.
func (h *UserHandler) SearchUsers(ctx context.Context, req *pb.UserServiceSearchRequest) (*pb.UserServiceSearchResponse, error) {
	limit := 20
	if req.Limit != nil {
		limit = int(*req.Limit)
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	offset := 0
	if req.Offset != nil {
		offset = int(*req.Offset)
	}
	if offset < 0 {
		offset = 0
	}

	resp, err := h.svc.SearchUsers(ctx, domain.SearchUsersRequest{
		Query:  req.Query,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, domainErrToStatus(err)
	}

	users := make([]*pb.User, 0, len(resp.Users))
	for _, u := range resp.Users {
		users = append(users, userToProto(u))
	}

	return &pb.UserServiceSearchResponse{
		Users:      users,
		TotalCount: resp.TotalCount,
	}, nil
}

// ChangeUserRole изменяет роль указанного пользователя.
// Требует роль admin или moderator. Возвращает codes.Unimplemented, пока не реализовано в сервисе.
func (h *UserHandler) ChangeUserRole(_ context.Context, req *pb.UserServiceChangeRoleRequest) (*pb.UserServiceChangeRoleResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if err := validateUserRole(req.NewRole); err != nil {
		return nil, err
	}

	// TODO: реализовать в UserService
	return nil, status.Error(codes.Unimplemented, "ChangeUserRole not yet implemented in service layer")
}

// SetUserStatus изменяет статус учётной записи пользователя.
// Требует роль admin. Возвращает codes.Unimplemented, пока не реализовано в сервисе.
func (h *UserHandler) SetUserStatus(_ context.Context, req *pb.UserServiceSetStatusRequest) (*pb.UserServiceSetStatusResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if err := validateUserStatus(req.NewStatus); err != nil {
		return nil, err
	}

	// TODO: реализовать в UserService
	return nil, status.Error(codes.Unimplemented, "SetUserStatus not yet implemented in service layer")
}
