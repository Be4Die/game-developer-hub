// Package service содержит бизнес-логику SSO-сервиса.
package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Be4Die/game-developer-hub/sso/internal/domain"
)

// AdminInitializer обеспечивает создание учётной записи администратора при старте.
type AdminInitializer struct {
	log            *slog.Logger
	userRepo       domain.UserRepository
	passwordHasher domain.PasswordHasher
	email          string
	password       string
	displayName    string
}

// NewAdminInitializer создаёт инициализатор администратора.
func NewAdminInitializer(
	log *slog.Logger,
	userRepo domain.UserRepository,
	passwordHasher domain.PasswordHasher,
	email, password, displayName string,
) *AdminInitializer {
	return &AdminInitializer{
		log:            log,
		userRepo:       userRepo,
		passwordHasher: passwordHasher,
		email:          email,
		password:       password,
		displayName:    displayName,
	}
}

// EnsureAdmin проверяет наличие администратора и создаёт его при отсутствии.
func (i *AdminInitializer) EnsureAdmin(ctx context.Context) error {
	const op = "AdminInitializer.EnsureAdmin"

	// Проверяем, существует ли уже администратор.
	existing, err := i.userRepo.GetByEmail(ctx, i.email)
	if err == nil && existing != nil {
		// Администратор уже существует.
		i.log.Info("admin user already exists",
			slog.String("user_id", existing.ID),
			slog.String("email", i.email),
		)
		return nil
	}

	// Администратор не найден — создаём.
	i.log.Info("creating admin user", slog.String("email", i.email))

	passwordHash, err := i.passwordHasher.Hash(ctx, i.password)
	if err != nil {
		return fmt.Errorf("%s: hash password: %w", op, err)
	}

	now := time.Now()
	admin := domain.User{
		Email:         i.email,
		PasswordHash:  passwordHash,
		DisplayName:   i.displayName,
		Role:          domain.RoleAdmin,
		Status:        domain.StatusActive,
		EmailVerified: true, // Внутренние пользователи не требуют верификации.
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := i.userRepo.Create(ctx, admin); err != nil {
		return fmt.Errorf("%s: create admin: %w", op, err)
	}

	created, err := i.userRepo.GetByEmail(ctx, i.email)
	if err != nil {
		i.log.Warn("admin created but failed to fetch",
			slog.String("email", i.email),
			slog.String("error", err.Error()),
		)
		return nil
	}

	i.log.Info("admin user created successfully",
		slog.String("user_id", created.ID),
		slog.String("email", i.email),
	)

	return nil
}
