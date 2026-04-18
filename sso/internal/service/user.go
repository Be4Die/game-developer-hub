package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Be4Die/game-developer-hub/sso/internal/domain"
)

// UserService реализует логику управления профилем.
type UserService struct {
	log            *slog.Logger
	userRepo       domain.UserRepository
	passwordHasher domain.PasswordHasher
}

// NewUserService создаёт сервис управления пользователями.
func NewUserService(log *slog.Logger, userRepo domain.UserRepository, passwordHasher domain.PasswordHasher) *UserService {
	return &UserService{
		log:            log,
		userRepo:       userRepo,
		passwordHasher: passwordHasher,
	}
}

// GetProfile возвращает профиль пользователя.
func (s *UserService) GetProfile(ctx context.Context, userID string) (domain.User, error) {
	const op = "UserService.GetProfile"

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return domain.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return *user, nil
}

// UpdateProfile обновляет профиль пользователя.
func (s *UserService) UpdateProfile(ctx context.Context, req domain.UpdateProfileRequest) (domain.User, error) {
	const op = "UserService.UpdateProfile"

	user, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return domain.User{}, fmt.Errorf("%s: %w", op, err)
	}

	if req.DisplayName != nil {
		user.DisplayName = *req.DisplayName
	}

	if err := s.userRepo.Update(ctx, *user); err != nil {
		return domain.User{}, fmt.Errorf("%s: %w", op, err)
	}

	updated, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return domain.User{}, fmt.Errorf("%s: get updated: %w", op, err)
	}

	return *updated, nil
}

// ChangePassword меняет пароль пользователя.
func (s *UserService) ChangePassword(ctx context.Context, req domain.ChangePasswordRequest) error {
	const op = "UserService.ChangePassword"

	user, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// Проверяем текущий пароль.
	if err := s.passwordHasher.Compare(ctx, user.PasswordHash, req.CurrentPassword); err != nil {
		return fmt.Errorf("%s: verify current password: %w", op, err)
	}

	// Хешируем новый пароль.
	newHash, err := s.passwordHasher.Hash(ctx, req.NewPassword)
	if err != nil {
		return fmt.Errorf("%s: hash new password: %w", op, err)
	}

	user.PasswordHash = newHash

	if err := s.userRepo.Update(ctx, *user); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// GetUserByID возвращает пользователя по ID.
func (s *UserService) GetUserByID(ctx context.Context, userID string) (domain.User, error) {
	const op = "UserService.GetUserByID"

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return domain.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return *user, nil
}

// SearchUsers ищет пользователей.
func (s *UserService) SearchUsers(ctx context.Context, req domain.SearchUsersRequest) (domain.SearchUsersResponse, error) {
	const op = "UserService.SearchUsers"

	users, total, err := s.userRepo.Search(ctx, req.Query, req.Limit, req.Offset)
	if err != nil {
		return domain.SearchUsersResponse{}, fmt.Errorf("%s: %w", op, err)
	}

	return domain.SearchUsersResponse{
		Users:      users,
		TotalCount: total,
	}, nil
}
