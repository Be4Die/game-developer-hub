// Package bcryptpkg реализует хеширование паролей на основе алгоритма bcrypt.
package bcryptpkg

import (
	"context"
	"fmt"

	"github.com/Be4Die/game-developer-hub/sso/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

// DefaultCost — стандартная сложность bcrypt для production.
const DefaultCost = 12

// PasswordHasher — реализация domain.PasswordHasher через bcrypt.
type PasswordHasher struct {
	cost int
}

// NewPasswordHasher создаёт хешер паролей.
func NewPasswordHasher(cost int) (*PasswordHasher, error) {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		return nil, fmt.Errorf("bcrypt.NewPasswordHasher: cost must be between %d and %d", bcrypt.MinCost, bcrypt.MaxCost)
	}
	return &PasswordHasher{cost: cost}, nil
}

// Hash хеширует пароль через bcrypt.
func (h *PasswordHasher) Hash(_ context.Context, password string) ([]byte, error) {
	const op = "bcrypt.PasswordHasher.Hash"

	hash, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return hash, nil
}

// Compare сравнивает пароль с хешом.
func (h *PasswordHasher) Compare(_ context.Context, hash []byte, password string) error {
	const op = "bcrypt.PasswordHasher.Compare"

	if err := bcrypt.CompareHashAndPassword(hash, []byte(password)); err != nil {
		return fmt.Errorf("%s: %w", op, domain.ErrInvalidPassword)
	}

	return nil
}
