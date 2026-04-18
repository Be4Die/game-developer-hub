package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/Be4Die/game-developer-hub/protos/sso/v1"
	"github.com/Be4Die/game-developer-hub/sso/internal/domain"
	"github.com/Be4Die/game-developer-hub/sso/internal/service"
)

// AuthHandler — gRPC обработчик для AuthService.
type AuthHandler struct {
	pb.UnimplementedAuthServiceServer
	svc *service.AuthService
}

// NewAuthHandler создаёт обработчик аутентификации.
func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// Register обрабатывает запрос на регистрацию нового пользователя.
// Возвращает ErrAlreadyExists, если email уже занят.
func (h *AuthHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	resp, err := h.svc.Register(ctx, domain.RegisterRequest{
		Email:       req.Email,
		Password:    req.Password,
		DisplayName: req.DisplayName,
	})
	if err != nil {
		return nil, domainErrToStatus(err)
	}

	return &pb.RegisterResponse{
		User:   userToProto(resp.User),
		Tokens: tokenInfoToProto(resp.Tokens),
	}, nil
}

// Login обрабатывает запрос на аутентификацию по email и паролю.
// Возвращает ErrInvalidPassword при неверных учётных данных.
func (h *AuthHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	loginReq := domain.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}
	if req.UserAgent != nil {
		loginReq.UserAgent = *req.UserAgent
	}
	if req.IpAddress != nil {
		loginReq.IPAddress = *req.IpAddress
	}

	resp, err := h.svc.Login(ctx, loginReq)
	if err != nil {
		return nil, domainErrToStatus(err)
	}

	return &pb.LoginResponse{
		User:   userToProto(resp.User),
		Tokens: tokenInfoToProto(resp.Tokens),
	}, nil
}

// RefreshToken обновляет пару access/refresh токенов по истёкшему refresh-токену.
// Возвращает ErrInvalidToken или ErrTokenExpired при невалидном токене.
func (h *AuthHandler) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	resp, err := h.svc.RefreshToken(ctx, domain.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		return nil, domainErrToStatus(err)
	}

	return &pb.RefreshTokenResponse{
		Tokens: tokenInfoToProto(resp.Tokens),
	}, nil
}

// Logout выполняет выход пользователя и инвалидирует refresh-токен.
func (h *AuthHandler) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	if err := h.svc.Logout(ctx, domain.LogoutRequest{
		RefreshToken: req.RefreshToken,
	}); err != nil {
		return nil, domainErrToStatus(err)
	}

	return &pb.LogoutResponse{}, nil
}

// VerifyEmail подтверждает email-адрес пользователя по коду верификации.
// Возвращает ErrInvalidToken при неверном или истёкшем коде.
func (h *AuthHandler) VerifyEmail(ctx context.Context, req *pb.VerifyEmailRequest) (*pb.VerifyEmailResponse, error) {
	if err := h.svc.VerifyEmail(ctx, domain.VerifyEmailRequest{
		VerificationCode: req.VerificationCode,
	}); err != nil {
		return nil, domainErrToStatus(err)
	}

	return &pb.VerifyEmailResponse{Success: true}, nil
}

// ResendVerificationEmail повторно отправляет письмо с кодом верификации.
// Возвращает ошибку, если пользователь не найден или email уже подтверждён.
func (h *AuthHandler) ResendVerificationEmail(ctx context.Context, req *pb.ResendVerificationEmailRequest) (*pb.ResendVerificationEmailResponse, error) {
	err := h.svc.ResendVerificationEmail(ctx, domain.ResendVerificationRequest{
		Email: req.Email,
	})
	if err != nil {
		return nil, domainErrToStatus(err)
	}

	return &pb.ResendVerificationEmailResponse{Sent: true}, nil
}

// RequestPasswordReset инициирует процесс сброса пароля.
// Всегда возвращает Ok:true для защиты от enumeration attack.
func (h *AuthHandler) RequestPasswordReset(ctx context.Context, req *pb.RequestPasswordResetRequest) (*pb.RequestPasswordResetResponse, error) {
	_ = h.svc.RequestPasswordReset(ctx, domain.PasswordResetRequest{
		Email: req.Email,
	})

	// Всегда возвращаем true — защита от enumeration attack.
	return &pb.RequestPasswordResetResponse{Ok: true}, nil
}

// ResetPassword устанавливает новый пароль по токену сброса.
// Возвращает ErrInvalidToken или ErrTokenExpired при невалидном токене.
func (h *AuthHandler) ResetPassword(ctx context.Context, req *pb.ResetPasswordRequest) (*pb.ResetPasswordResponse, error) {
	if err := h.svc.ResetPassword(ctx, domain.ResetPasswordRequest{
		ResetToken:  req.ResetToken,
		NewPassword: req.NewPassword,
	}); err != nil {
		return nil, domainErrToStatus(err)
	}

	return &pb.ResetPasswordResponse{Success: true}, nil
}

// validateUserRole проверяет что роль допустима для изменения.
func validateUserRole(role pb.UserRole) error {
	switch role {
	case pb.UserRole_USER_ROLE_DEVELOPER,
		pb.UserRole_USER_ROLE_MODERATOR,
		pb.UserRole_USER_ROLE_ADMIN:
		return nil
	default:
		return status.Errorf(codes.InvalidArgument, "invalid user role: %v", role)
	}
}

// validateUserStatus проверяет что статус допустим для изменения.
func validateUserStatus(s pb.UserStatus) error {
	switch s {
	case pb.UserStatus_USER_STATUS_ACTIVE,
		pb.UserStatus_USER_STATUS_SUSPENDED,
		pb.UserStatus_USER_STATUS_DELETED:
		return nil
	default:
		return status.Errorf(codes.InvalidArgument, "invalid user status: %v", s)
	}
}
