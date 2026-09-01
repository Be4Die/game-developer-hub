package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Be4Die/game-developer-hub/sso/internal/domain"
)

// UserService реализует логику управления профилем и пользователями.
type UserService struct {
	log            *slog.Logger
	userRepo       domain.UserRepository
	sessionRepo    domain.SessionRepository
	passwordHasher domain.PasswordHasher
}

// NewUserService создаёт сервис управления пользователями.
func NewUserService(
	log *slog.Logger,
	userRepo domain.UserRepository,
	passwordHasher domain.PasswordHasher,
	sessionRepo domain.SessionRepository,
) *UserService {
	return &UserService{
		log:            log,
		userRepo:       userRepo,
		sessionRepo:    sessionRepo,
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

	if user.Role == domain.RoleModerator || user.Role == domain.RoleAdmin {
		return domain.User{}, fmt.Errorf("%s: %w", op, domain.ErrProfileImmutable)
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

	if user.Role == domain.RoleModerator || user.Role == domain.RoleAdmin {
		return fmt.Errorf("%s: %w", op, domain.ErrProfileImmutable)
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

// CreateModerator создаёт учётную запись модератора.
// Внутренние пользователи не требуют верификации email.
// Email формируется как {login}@welwise.com.
func (s *UserService) CreateModerator(ctx context.Context, req domain.CreateModeratorRequest) (domain.CreateModeratorResponse, error) {
	const op = "UserService.CreateModerator"

	// Формируем email из логина.
	email := req.Login + domain.WelwiseDomain

	// Хешируем пароль.
	passwordHash, err := s.passwordHasher.Hash(ctx, req.Password)
	if err != nil {
		return domain.CreateModeratorResponse{}, fmt.Errorf("%s: hash password: %w", op, err)
	}

	now := time.Now()
	moderator := domain.User{
		Email:         email,
		PasswordHash:  passwordHash,
		DisplayName:   req.DisplayName,
		Role:          domain.RoleModerator,
		Status:        domain.StatusActive,
		EmailVerified: true, // Внутренние пользователи не требуют верификации.
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.userRepo.Create(ctx, moderator); err != nil {
		return domain.CreateModeratorResponse{}, fmt.Errorf("%s: create moderator: %w", op, err)
	}

	created, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return domain.CreateModeratorResponse{}, fmt.Errorf("%s: get created moderator: %w", op, err)
	}

	s.log.Info("moderator created",
		slog.String("user_id", created.ID),
		slog.String("email", email),
		slog.String("display_name", req.DisplayName),
	)

	return domain.CreateModeratorResponse{User: *created}, nil
}

// SetUserStatus изменяет статус учётной записи пользователя.
// Модератор может блокировать/разблокировать только разработчиков.
// Администратор может блокировать/разблокировать/удалять разработчиков и восстанавливать их.
// Модератора нельзя заблокировать/разблокировать. Администратора нельзя заблокировать/изменить статус.
func (s *UserService) SetUserStatus(ctx context.Context, req domain.SetUserStatusRequest) (domain.User, error) {
	const op = "UserService.SetUserStatus"

	caller, err := s.userRepo.GetByID(ctx, req.CallerID)
	if err != nil {
		return domain.User{}, fmt.Errorf("%s: get caller: %w", op, err)
	}

	if caller.Role != domain.RoleAdmin && caller.Role != domain.RoleModerator {
		return domain.User{}, fmt.Errorf("%s: %w", op, domain.ErrPermissionDenied)
	}

	target, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return domain.User{}, fmt.Errorf("%s: get target: %w", op, err)
	}

	// Администратора нельзя заблокировать/разблокировать или перевести в удалён
	if target.Role == domain.RoleAdmin {
		return domain.User{}, fmt.Errorf("%s: %w", op, domain.ErrCannotModifyAdminStatus)
	}

	// Модератора нельзя заблокировать или разблокировать
	if target.Role == domain.RoleModerator {
		if req.Status == domain.StatusSuspended {
			return domain.User{}, fmt.Errorf("%s: %w", op, domain.ErrCannotModifyModeratorStatus)
		}
		// Модератор не может менять статус другого модератора вообще
		if caller.Role == domain.RoleModerator {
			return domain.User{}, fmt.Errorf("%s: %w", op, domain.ErrPermissionDenied)
		}
	}

	// Модератор может менять статус только разработчиков на Active или Suspended
	if caller.Role == domain.RoleModerator {
		if target.Role != domain.RoleDeveloper {
			return domain.User{}, fmt.Errorf("%s: %w", op, domain.ErrPermissionDenied)
		}
		if req.Status == domain.StatusDeleted {
			return domain.User{}, fmt.Errorf("%s: %w", op, domain.ErrPermissionDenied)
		}
	}

	target.Status = req.Status
	target.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, *target); err != nil {
		return domain.User{}, fmt.Errorf("%s: update user status: %w", op, err)
	}

	// При блокировке или soft-delete отзываем сессии
	if (req.Status == domain.StatusSuspended || req.Status == domain.StatusDeleted) && s.sessionRepo != nil {
		if _, err := s.sessionRepo.RevokeAllForUser(ctx, target.ID, ""); err != nil {
			s.log.Warn("failed to revoke sessions on status change", slog.String("user_id", target.ID), slog.String("error", err.Error()))
		}
	}

	s.log.Info("user status changed",
		slog.String("target_id", target.ID),
		slog.String("new_status", target.Status.String()),
		slog.String("caller_id", caller.ID),
	)

	return *target, nil
}

// DeleteUser выполняет soft delete пользователя (StatusDeleted).
// Доступно только администраторам. Администраторы не могут быть удалены.
func (s *UserService) DeleteUser(ctx context.Context, req domain.DeleteUserRequest) error {
	const op = "UserService.DeleteUser"

	caller, err := s.userRepo.GetByID(ctx, req.CallerID)
	if err != nil {
		return fmt.Errorf("%s: get caller: %w", op, err)
	}

	if caller.Role != domain.RoleAdmin {
		return fmt.Errorf("%s: %w", op, domain.ErrPermissionDenied)
	}

	// Проверяем что пользователь существует.
	user, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return fmt.Errorf("%s: get user: %w", op, err)
	}

	// Запрещаем удаление администраторов.
	if user.Role == domain.RoleAdmin {
		return fmt.Errorf("%s: %w", op, domain.ErrCannotDeleteAdmin)
	}

	user.Status = domain.StatusDeleted
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, *user); err != nil {
		return fmt.Errorf("%s: update user status to deleted: %w", op, err)
	}

	if s.sessionRepo != nil {
		if _, err := s.sessionRepo.RevokeAllForUser(ctx, req.UserID, ""); err != nil {
			s.log.Warn("failed to revoke sessions for deleted user", slog.String("user_id", req.UserID), slog.String("error", err.Error()))
		}
	}

	s.log.Info("user soft deleted",
		slog.String("user_id", req.UserID),
		slog.String("email", user.Email),
		slog.String("role", user.Role.String()),
	)

	return nil
}

