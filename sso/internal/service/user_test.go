package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Be4Die/game-developer-hub/sso/internal/domain"
)

func TestUserService_GetProfile(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	log := newTestLogger()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		expectedUser := &domain.User{ID: "user-1", Email: "test@example.com", DisplayName: "Test User", Role: domain.RoleDeveloper, Status: domain.StatusActive}
		userRepo := &stubUserRepo{getByIDFunc: func(context.Context, string) (*domain.User, error) { return expectedUser, nil }}
		svc := NewUserService(log, userRepo, &stubPasswordHasher{})

		user, err := svc.GetProfile(ctx, "user-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if user.ID != expectedUser.ID {
			t.Errorf("expected user ID %s, got %s", expectedUser.ID, user.ID)
		}
		if user.Email != expectedUser.Email {
			t.Errorf("expected email %s, got %s", expectedUser.Email, user.Email)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		t.Parallel()

		userRepo := &stubUserRepo{getByIDFunc: func(context.Context, string) (*domain.User, error) { return nil, errors.New("not found") }}
		svc := NewUserService(log, userRepo, &stubPasswordHasher{})

		_, err := svc.GetProfile(ctx, "nonexistent")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestUserService_UpdateProfile(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	log := newTestLogger()

	t.Run("update display name", func(t *testing.T) {
		t.Parallel()

		displayName := "New Name"
		existingUser := &domain.User{ID: "user-1", Email: "test@example.com", DisplayName: "Old Name"}
		updatedUser := &domain.User{ID: "user-1", Email: "test@example.com", DisplayName: "New Name"}

		callCount := 0
		userRepo := &stubUserRepo{
			getByIDFunc: func(context.Context, string) (*domain.User, error) {
				callCount++
				if callCount == 1 {
					return existingUser, nil
				}
				return updatedUser, nil
			},
			updateFunc: func(context.Context, domain.User) error { return nil },
		}
		svc := NewUserService(log, userRepo, &stubPasswordHasher{})

		user, err := svc.UpdateProfile(ctx, domain.UpdateProfileRequest{UserID: "user-1", DisplayName: &displayName})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if user.DisplayName != "New Name" {
			t.Errorf("expected display name 'New Name', got %s", user.DisplayName)
		}
	})

	t.Run("update with nil display name", func(t *testing.T) {
		t.Parallel()

		existingUser := &domain.User{ID: "user-1", Email: "test@example.com", DisplayName: "Old Name"}
		userRepo := &stubUserRepo{
			getByIDFunc: func(context.Context, string) (*domain.User, error) { return existingUser, nil },
			updateFunc:  func(context.Context, domain.User) error { return nil },
		}
		svc := NewUserService(log, userRepo, &stubPasswordHasher{})

		user, err := svc.UpdateProfile(ctx, domain.UpdateProfileRequest{UserID: "user-1", DisplayName: nil})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if user.DisplayName != "Old Name" {
			t.Errorf("expected display name 'Old Name', got %s", user.DisplayName)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		t.Parallel()

		displayName := "New Name"
		userRepo := &stubUserRepo{getByIDFunc: func(context.Context, string) (*domain.User, error) { return nil, errors.New("not found") }}
		svc := NewUserService(log, userRepo, &stubPasswordHasher{})

		_, err := svc.UpdateProfile(ctx, domain.UpdateProfileRequest{UserID: "nonexistent", DisplayName: &displayName})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("update error", func(t *testing.T) {
		t.Parallel()

		displayName := "New Name"
		existingUser := &domain.User{ID: "user-1", DisplayName: "Old Name"}
		userRepo := &stubUserRepo{
			getByIDFunc: func(context.Context, string) (*domain.User, error) { return existingUser, nil },
			updateFunc:  func(context.Context, domain.User) error { return errors.New("db error") },
		}
		svc := NewUserService(log, userRepo, &stubPasswordHasher{})

		_, err := svc.UpdateProfile(ctx, domain.UpdateProfileRequest{UserID: "user-1", DisplayName: &displayName})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestUserService_ChangePassword(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	log := newTestLogger()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		userRepo := &stubUserRepo{
			getByIDFunc: func(_ context.Context, id string) (*domain.User, error) {
				return &domain.User{ID: id, PasswordHash: []byte("old-hash")}, nil
			},
			updateFunc: func(context.Context, domain.User) error { return nil },
		}
		svc := NewUserService(log, userRepo, &stubPasswordHasher{compareErr: nil, hashToReturn: []byte("new-hash")})

		err := svc.ChangePassword(ctx, domain.ChangePasswordRequest{UserID: "user-1", CurrentPassword: "old-password", NewPassword: "new-password"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("wrong current password", func(t *testing.T) {
		t.Parallel()

		userRepo := &stubUserRepo{getByIDFunc: func(_ context.Context, id string) (*domain.User, error) {
			return &domain.User{ID: id, PasswordHash: []byte("old-hash")}, nil
		}}
		svc := NewUserService(log, userRepo, &stubPasswordHasher{compareErr: domain.ErrInvalidPassword})

		err := svc.ChangePassword(ctx, domain.ChangePasswordRequest{UserID: "user-1", CurrentPassword: "wrong-password", NewPassword: "new-password"})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("hash new password error", func(t *testing.T) {
		t.Parallel()

		userRepo := &stubUserRepo{getByIDFunc: func(_ context.Context, id string) (*domain.User, error) {
			return &domain.User{ID: id, PasswordHash: []byte("old-hash")}, nil
		}}
		svc := NewUserService(log, userRepo, &stubPasswordHasher{hashErr: errors.New("hash error")})

		err := svc.ChangePassword(ctx, domain.ChangePasswordRequest{UserID: "user-1", CurrentPassword: "old-password", NewPassword: "new-password"})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("user not found", func(t *testing.T) {
		t.Parallel()

		userRepo := &stubUserRepo{getByIDFunc: func(context.Context, string) (*domain.User, error) { return nil, errors.New("not found") }}
		svc := NewUserService(log, userRepo, &stubPasswordHasher{})

		err := svc.ChangePassword(ctx, domain.ChangePasswordRequest{UserID: "nonexistent", CurrentPassword: "old-password", NewPassword: "new-password"})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestUserService_GetUserByID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	log := newTestLogger()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		expectedUser := &domain.User{ID: "user-1", Email: "test@example.com"}
		userRepo := &stubUserRepo{getByIDFunc: func(context.Context, string) (*domain.User, error) { return expectedUser, nil }}
		svc := NewUserService(log, userRepo, &stubPasswordHasher{})

		user, err := svc.GetUserByID(ctx, "user-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if user.ID != expectedUser.ID {
			t.Errorf("expected user ID %s, got %s", expectedUser.ID, user.ID)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		t.Parallel()

		userRepo := &stubUserRepo{getByIDFunc: func(context.Context, string) (*domain.User, error) { return nil, errors.New("not found") }}
		svc := NewUserService(log, userRepo, &stubPasswordHasher{})

		_, err := svc.GetUserByID(ctx, "nonexistent")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestUserService_SearchUsers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	log := newTestLogger()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		expectedUsers := []domain.User{
			{ID: "user-1", Email: "test1@example.com"},
			{ID: "user-2", Email: "test2@example.com"},
		}
		userRepo := &stubUserRepo{searchFunc: func(context.Context, string, int, int) ([]domain.User, int64, error) { return expectedUsers, 2, nil }}
		svc := NewUserService(log, userRepo, &stubPasswordHasher{})

		resp, err := svc.SearchUsers(ctx, domain.SearchUsersRequest{Query: "test", Limit: 10, Offset: 0})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.TotalCount != 2 {
			t.Errorf("expected total count 2, got %d", resp.TotalCount)
		}
		if len(resp.Users) != 2 {
			t.Errorf("expected 2 users, got %d", len(resp.Users))
		}
	})

	t.Run("search error", func(t *testing.T) {
		t.Parallel()

		userRepo := &stubUserRepo{searchFunc: func(context.Context, string, int, int) ([]domain.User, int64, error) {
			return nil, 0, errors.New("db error")
		}}
		svc := NewUserService(log, userRepo, &stubPasswordHasher{})

		_, err := svc.SearchUsers(ctx, domain.SearchUsersRequest{Query: "test"})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
