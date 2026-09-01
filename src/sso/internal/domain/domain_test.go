package domain_test

import (
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/sso/internal/domain"
)

func TestSession_IsActive(t *testing.T) {
	t.Parallel()

	t.Run("active session", func(t *testing.T) {
		t.Parallel()

		session := &domain.Session{
			Revoked:   false,
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}

		if !session.IsActive() {
			t.Error("expected session to be active")
		}
	})

	t.Run("revoked session", func(t *testing.T) {
		t.Parallel()

		session := &domain.Session{
			Revoked:   true,
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}

		if session.IsActive() {
			t.Error("expected revoked session to be inactive")
		}
	})

	t.Run("expired session", func(t *testing.T) {
		t.Parallel()

		session := &domain.Session{
			Revoked:   false,
			ExpiresAt: time.Now().Add(-1 * time.Hour),
		}

		if session.IsActive() {
			t.Error("expected expired session to be inactive")
		}
	})

	t.Run("revoked and expired session", func(t *testing.T) {
		t.Parallel()

		session := &domain.Session{
			Revoked:   true,
			ExpiresAt: time.Now().Add(-1 * time.Hour),
		}

		if session.IsActive() {
			t.Error("expected revoked and expired session to be inactive")
		}
	})
}

func TestUserRole_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		role     domain.UserRole
		expected string
	}{
		{"developer", domain.RoleDeveloper, "developer"},
		{"moderator", domain.RoleModerator, "moderator"},
		{"admin", domain.RoleAdmin, "admin"},
		{"unknown", domain.UserRole(0), "unknown"},
		{"unknown high", domain.UserRole(100), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if result := tt.role.String(); result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestParseUserRole(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected domain.UserRole
	}{
		{"developer", "developer", domain.RoleDeveloper},
		{"moderator", "moderator", domain.RoleModerator},
		{"admin", "admin", domain.RoleAdmin},
		{"unknown", "unknown", domain.UserRole(0)},
		{"empty", "", domain.UserRole(0)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if result := domain.ParseUserRole(tt.input); result != tt.expected {
				t.Errorf("ParseUserRole(%s) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestUserStatus_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		status   domain.UserStatus
		expected string
	}{
		{"active", domain.StatusActive, "active"},
		{"suspended", domain.StatusSuspended, "suspended"},
		{"deleted", domain.StatusDeleted, "deleted"},
		{"unknown", domain.UserStatus(0), "unknown"},
		{"unknown high", domain.UserStatus(100), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if result := tt.status.String(); result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestParseUserStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected domain.UserStatus
	}{
		{"active", "active", domain.StatusActive},
		{"suspended", "suspended", domain.StatusSuspended},
		{"deleted", "deleted", domain.StatusDeleted},
		{"unknown", "unknown", domain.UserStatus(0)},
		{"empty", "", domain.UserStatus(0)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if result := domain.ParseUserStatus(tt.input); result != tt.expected {
				t.Errorf("ParseUserStatus(%s) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}
