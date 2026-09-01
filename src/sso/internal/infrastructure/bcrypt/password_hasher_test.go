package bcryptpkg_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Be4Die/game-developer-hub/sso/internal/domain"
	bcryptpkg "github.com/Be4Die/game-developer-hub/sso/internal/infrastructure/bcrypt"
	"golang.org/x/crypto/bcrypt"
)

func TestNewPasswordHasher(t *testing.T) {
	t.Parallel()

	t.Run("valid cost", func(t *testing.T) {
		t.Parallel()

		h, err := bcryptpkg.NewPasswordHasher(bcryptpkg.DefaultCost)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if h == nil {
			t.Fatal("expected PasswordHasher instance, got nil")
		}
	})

	t.Run("cost too low", func(t *testing.T) {
		t.Parallel()

		_, err := bcryptpkg.NewPasswordHasher(bcrypt.MinCost - 1)

		if err == nil {
			t.Fatal("expected error for low cost, got nil")
		}
	})

	t.Run("cost too high", func(t *testing.T) {
		t.Parallel()

		_, err := bcryptpkg.NewPasswordHasher(bcrypt.MaxCost + 1)

		if err == nil {
			t.Fatal("expected error for high cost, got nil")
		}
	})

	t.Run("minimum cost", func(t *testing.T) {
		t.Parallel()

		h, err := bcryptpkg.NewPasswordHasher(bcrypt.MinCost)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if h == nil {
			t.Fatal("expected PasswordHasher instance, got nil")
		}
	})

	t.Run("maximum cost", func(t *testing.T) {
		t.Parallel()

		h, err := bcryptpkg.NewPasswordHasher(bcrypt.MaxCost)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if h == nil {
			t.Fatal("expected PasswordHasher instance, got nil")
		}
	})
}

func TestPasswordHasher_Hash(t *testing.T) {
	t.Parallel()

	h, err := bcryptpkg.NewPasswordHasher(bcryptpkg.DefaultCost)
	if err != nil {
		t.Fatalf("failed to create PasswordHasher: %v", err)
	}

	t.Run("hash password", func(t *testing.T) {
		t.Parallel()

		hash, err := h.Hash(context.Background(), "password123")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(hash) == 0 {
			t.Fatal("expected non-empty hash")
		}

		// Bcrypt hashes start with $2a$ or $2b$
		if string(hash[:4]) != "$2a$" && string(hash[:4]) != "$2b$" {
			t.Errorf("expected bcrypt format, got prefix: %s", string(hash[:4]))
		}
	})

	t.Run("different hashes for same password", func(t *testing.T) {
		t.Parallel()

		hash1, err := h.Hash(context.Background(), "password123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		hash2, err := h.Hash(context.Background(), "password123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Bcrypt uses random salt, so hashes should be different
		if string(hash1) == string(hash2) {
			t.Error("expected different hashes for same password (different salts)")
		}
	})

	t.Run("empty password", func(t *testing.T) {
		t.Parallel()

		hash, err := h.Hash(context.Background(), "")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(hash) == 0 {
			t.Fatal("expected non-empty hash for empty password")
		}
	})
}

func TestPasswordHasher_Compare(t *testing.T) {
	t.Parallel()

	h, err := bcryptpkg.NewPasswordHasher(bcryptpkg.DefaultCost)
	if err != nil {
		t.Fatalf("failed to create PasswordHasher: %v", err)
	}

	t.Run("correct password", func(t *testing.T) {
		t.Parallel()

		password := "password123"
		hash, err := h.Hash(context.Background(), password)
		if err != nil {
			t.Fatalf("failed to hash password: %v", err)
		}

		err = h.Compare(context.Background(), hash, password)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		t.Parallel()

		hash, err := h.Hash(context.Background(), "password123")
		if err != nil {
			t.Fatalf("failed to hash password: %v", err)
		}

		err = h.Compare(context.Background(), hash, "wrongpassword")

		if err == nil {
			t.Fatal("expected error for wrong password, got nil")
		}

		if !errors.Is(err, domain.ErrInvalidPassword) {
			t.Errorf("expected ErrInvalidPassword, got %v", err)
		}
	})

	t.Run("empty hash", func(t *testing.T) {
		t.Parallel()

		err := h.Compare(context.Background(), []byte(""), "password")

		if err == nil {
			t.Fatal("expected error for empty hash, got nil")
		}
	})

	t.Run("invalid hash format", func(t *testing.T) {
		t.Parallel()

		err := h.Compare(context.Background(), []byte("not-a-valid-hash"), "password")

		if err == nil {
			t.Fatal("expected error for invalid hash format, got nil")
		}
	})
}

func TestPasswordHasher_RoundTrip(t *testing.T) {
	t.Parallel()

	h, err := bcryptpkg.NewPasswordHasher(bcryptpkg.DefaultCost)
	if err != nil {
		t.Fatalf("failed to create PasswordHasher: %v", err)
	}

	passwords := []string{
		"simple",
		"with spaces",
		"unicode: привет мир",
		"special: !@#$%^&*()",
	}

	for _, password := range passwords {
		t.Run(password[:minLen(len(password), 20)], func(t *testing.T) {
			t.Parallel()

			hash, err := h.Hash(context.Background(), password)
			if err != nil {
				t.Fatalf("failed to hash password: %v", err)
			}

			err = h.Compare(context.Background(), hash, password)
			if err != nil {
				t.Fatalf("comparison failed: %v", err)
			}
		})
	}
}

func minLen(a, b int) int {
	if a < b {
		return a
	}
	return b
}
