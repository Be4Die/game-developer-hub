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
		svc := NewUserService(log, userRepo, &stubPasswordHasher{}, nil)

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
		svc := NewUserService(log, userRepo, &stubPasswordHasher{}, nil)

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

	t.Run("update display name for developer", func(t *testing.T) {
		t.Parallel()

		displayName := "New Name"
		existingUser := &domain.User{ID: "user-1", Role: domain.RoleDeveloper, Email: "test@example.com", DisplayName: "Old Name"}
		updatedUser := &domain.User{ID: "user-1", Role: domain.RoleDeveloper, Email: "test@example.com", DisplayName: "New Name"}

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
		svc := NewUserService(log, userRepo, &stubPasswordHasher{}, nil)

		user, err := svc.UpdateProfile(ctx, domain.UpdateProfileRequest{UserID: "user-1", DisplayName: &displayName})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if user.DisplayName != "New Name" {
			t.Errorf("expected display name 'New Name', got %s", user.DisplayName)
		}
	})

	t.Run("cannot update display name for moderator", func(t *testing.T) {
		t.Parallel()

		displayName := "New Mod Name"
		moderator := &domain.User{ID: "mod-1", Role: domain.RoleModerator, Email: "mod@welwise.com", DisplayName: "Mod"}
		userRepo := &stubUserRepo{
			getByIDFunc: func(context.Context, string) (*domain.User, error) { return moderator, nil },
		}
		svc := NewUserService(log, userRepo, &stubPasswordHasher{}, nil)

		_, err := svc.UpdateProfile(ctx, domain.UpdateProfileRequest{UserID: "mod-1", DisplayName: &displayName})
		if !errors.Is(err, domain.ErrProfileImmutable) {
			t.Fatalf("expected ErrProfileImmutable, got %v", err)
		}
	})

	t.Run("cannot update display name for admin", func(t *testing.T) {
		t.Parallel()

		displayName := "New Admin Name"
		admin := &domain.User{ID: "admin-1", Role: domain.RoleAdmin, Email: "admin@welwise.com", DisplayName: "Admin"}
		userRepo := &stubUserRepo{
			getByIDFunc: func(context.Context, string) (*domain.User, error) { return admin, nil },
		}
		svc := NewUserService(log, userRepo, &stubPasswordHasher{}, nil)

		_, err := svc.UpdateProfile(ctx, domain.UpdateProfileRequest{UserID: "admin-1", DisplayName: &displayName})
		if !errors.Is(err, domain.ErrProfileImmutable) {
			t.Fatalf("expected ErrProfileImmutable, got %v", err)
		}
	})

	t.Run("update with nil display name", func(t *testing.T) {
		t.Parallel()

		existingUser := &domain.User{ID: "user-1", Role: domain.RoleDeveloper, Email: "test@example.com", DisplayName: "Old Name"}
		userRepo := &stubUserRepo{
			getByIDFunc: func(context.Context, string) (*domain.User, error) { return existingUser, nil },
			updateFunc:  func(context.Context, domain.User) error { return nil },
		}
		svc := NewUserService(log, userRepo, &stubPasswordHasher{}, nil)

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
		svc := NewUserService(log, userRepo, &stubPasswordHasher{}, nil)

		_, err := svc.UpdateProfile(ctx, domain.UpdateProfileRequest{UserID: "nonexistent", DisplayName: &displayName})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("update error", func(t *testing.T) {
		t.Parallel()

		displayName := "New Name"
		existingUser := &domain.User{ID: "user-1", Role: domain.RoleDeveloper, DisplayName: "Old Name"}
		userRepo := &stubUserRepo{
			getByIDFunc: func(context.Context, string) (*domain.User, error) { return existingUser, nil },
			updateFunc:  func(context.Context, domain.User) error { return errors.New("db error") },
		}
		svc := NewUserService(log, userRepo, &stubPasswordHasher{}, nil)

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

	t.Run("success for developer", func(t *testing.T) {
		t.Parallel()

		existingUser := &domain.User{ID: "user-1", Role: domain.RoleDeveloper, PasswordHash: []byte("old-hash")}
		userRepo := &stubUserRepo{
			getByIDFunc: func(context.Context, string) (*domain.User, error) { return existingUser, nil },
			updateFunc:  func(context.Context, domain.User) error { return nil },
		}
		hasher := &stubPasswordHasher{hashToReturn: []byte("new-hash")}
		svc := NewUserService(log, userRepo, hasher, nil)

		err := svc.ChangePassword(ctx, domain.ChangePasswordRequest{
			UserID:          "user-1",
			CurrentPassword: "old-password",
			NewPassword:     "new-password",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("cannot change password for moderator", func(t *testing.T) {
		t.Parallel()

		moderator := &domain.User{ID: "mod-1", Role: domain.RoleModerator, PasswordHash: []byte("hash")}
		userRepo := &stubUserRepo{
			getByIDFunc: func(context.Context, string) (*domain.User, error) { return moderator, nil },
		}
		svc := NewUserService(log, userRepo, &stubPasswordHasher{}, nil)

		err := svc.ChangePassword(ctx, domain.ChangePasswordRequest{
			UserID:          "mod-1",
			CurrentPassword: "current",
			NewPassword:     "new",
		})
		if !errors.Is(err, domain.ErrProfileImmutable) {
			t.Fatalf("expected ErrProfileImmutable, got %v", err)
		}
	})

	t.Run("cannot change password for admin", func(t *testing.T) {
		t.Parallel()

		admin := &domain.User{ID: "admin-1", Role: domain.RoleAdmin, PasswordHash: []byte("hash")}
		userRepo := &stubUserRepo{
			getByIDFunc: func(context.Context, string) (*domain.User, error) { return admin, nil },
		}
		svc := NewUserService(log, userRepo, &stubPasswordHasher{}, nil)

		err := svc.ChangePassword(ctx, domain.ChangePasswordRequest{
			UserID:          "admin-1",
			CurrentPassword: "current",
			NewPassword:     "new",
		})
		if !errors.Is(err, domain.ErrProfileImmutable) {
			t.Fatalf("expected ErrProfileImmutable, got %v", err)
		}
	})

	t.Run("invalid current password", func(t *testing.T) {
		t.Parallel()

		existingUser := &domain.User{ID: "user-1", Role: domain.RoleDeveloper, PasswordHash: []byte("old-hash")}
		userRepo := &stubUserRepo{
			getByIDFunc: func(context.Context, string) (*domain.User, error) { return existingUser, nil },
		}
		hasher := &stubPasswordHasher{compareErr: domain.ErrInvalidPassword}
		svc := NewUserService(log, userRepo, hasher, nil)

		err := svc.ChangePassword(ctx, domain.ChangePasswordRequest{
			UserID:          "user-1",
			CurrentPassword: "wrong-password",
			NewPassword:     "new-password",
		})
		if !errors.Is(err, domain.ErrInvalidPassword) {
			t.Fatalf("expected ErrInvalidPassword, got %v", err)
		}
	})

	t.Run("hash error", func(t *testing.T) {
		t.Parallel()

		existingUser := &domain.User{ID: "user-1", Role: domain.RoleDeveloper, PasswordHash: []byte("old-hash")}
		userRepo := &stubUserRepo{
			getByIDFunc: func(context.Context, string) (*domain.User, error) { return existingUser, nil },
		}
		hasher := &stubPasswordHasher{hashErr: errors.New("hash error")}
		svc := NewUserService(log, userRepo, hasher, nil)

		err := svc.ChangePassword(ctx, domain.ChangePasswordRequest{
			UserID:          "user-1",
			CurrentPassword: "old-password",
			NewPassword:     "new-password",
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("user not found", func(t *testing.T) {
		t.Parallel()

		userRepo := &stubUserRepo{
			getByIDFunc: func(context.Context, string) (*domain.User, error) { return nil, errors.New("not found") },
		}
		svc := NewUserService(log, userRepo, &stubPasswordHasher{}, nil)

		err := svc.ChangePassword(ctx, domain.ChangePasswordRequest{
			UserID:          "nonexistent",
			CurrentPassword: "old-password",
			NewPassword:     "new-password",
		})
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
		userRepo := &stubUserRepo{
			getByIDFunc: func(context.Context, string) (*domain.User, error) { return expectedUser, nil },
		}
		svc := NewUserService(log, userRepo, &stubPasswordHasher{}, nil)

		user, err := svc.GetUserByID(ctx, "user-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if user.ID != "user-1" {
			t.Errorf("expected user ID 'user-1', got %s", user.ID)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		t.Parallel()

		userRepo := &stubUserRepo{
			getByIDFunc: func(context.Context, string) (*domain.User, error) { return nil, errors.New("not found") },
		}
		svc := NewUserService(log, userRepo, &stubPasswordHasher{}, nil)

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

		users := []domain.User{
			{ID: "user-1", DisplayName: "Alice"},
			{ID: "user-2", DisplayName: "Bob"},
		}
		userRepo := &stubUserRepo{
			searchFunc: func(context.Context, string, int, int) ([]domain.User, int64, error) {
				return users, 2, nil
			},
		}
		svc := NewUserService(log, userRepo, &stubPasswordHasher{}, nil)

		res, err := svc.SearchUsers(ctx, domain.SearchUsersRequest{Query: "test", Limit: 10, Offset: 0})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(res.Users) != 2 {
			t.Errorf("expected 2 users, got %d", len(res.Users))
		}
		if res.TotalCount != 2 {
			t.Errorf("expected total count 2, got %d", res.TotalCount)
		}
	})

	t.Run("search error", func(t *testing.T) {
		t.Parallel()

		userRepo := &stubUserRepo{
			searchFunc: func(context.Context, string, int, int) ([]domain.User, int64, error) {
				return nil, 0, errors.New("db error")
			},
		}
		svc := NewUserService(log, userRepo, &stubPasswordHasher{}, nil)

		_, err := svc.SearchUsers(ctx, domain.SearchUsersRequest{Query: "test", Limit: 10, Offset: 0})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestUserService_CreateModerator(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	log := newTestLogger()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		var createdUser domain.User
		userRepo := &stubUserRepo{
			createFunc: func(_ context.Context, u domain.User) error {
				createdUser = u
				createdUser.ID = "mod-1"
				return nil
			},
			getByEmailFunc: func(_ context.Context, _ string) (*domain.User, error) {
				return &createdUser, nil
			},
		}
		hasher := &stubPasswordHasher{hashToReturn: []byte("hashed-password")}
		svc := NewUserService(log, userRepo, hasher, nil)

		res, err := svc.CreateModerator(ctx, domain.CreateModeratorRequest{
			Login:       "moduser",
			Password:    "password123",
			DisplayName: "Moderator User",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.User.Email != "moduser@welwise.com" {
			t.Errorf("expected email moduser@welwise.com, got %s", res.User.Email)
		}
		if res.User.Role != domain.RoleModerator {
			t.Errorf("expected role Moderator, got %v", res.User.Role)
		}
		if res.User.Status != domain.StatusActive {
			t.Errorf("expected status Active, got %v", res.User.Status)
		}
	})

	t.Run("create error", func(t *testing.T) {
		t.Parallel()

		userRepo := &stubUserRepo{
			createFunc: func(context.Context, domain.User) error { return errors.New("already exists") },
		}
		hasher := &stubPasswordHasher{hashToReturn: []byte("hash")}
		svc := NewUserService(log, userRepo, hasher, nil)

		_, err := svc.CreateModerator(ctx, domain.CreateModeratorRequest{
			Login:       "moduser",
			Password:    "password123",
			DisplayName: "Mod",
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestUserService_SetUserStatus(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	log := newTestLogger()

	admin := &domain.User{ID: "admin-1", Role: domain.RoleAdmin}
	moderator := &domain.User{ID: "mod-1", Role: domain.RoleModerator}
	developer := &domain.User{ID: "dev-1", Role: domain.RoleDeveloper, Status: domain.StatusActive}

	t.Run("admin suspends developer and revokes sessions", func(t *testing.T) {
		t.Parallel()

		var updated domain.User
		var revokedUserID string
		userRepo := &stubUserRepo{
			getByIDFunc: func(_ context.Context, id string) (*domain.User, error) {
				if id == "admin-1" {
					return admin, nil
				}
				if id == "dev-1" {
					return developer, nil
				}
				return nil, domain.ErrNotFound
			},
			updateFunc: func(_ context.Context, u domain.User) error {
				updated = u
				return nil
			},
		}
		sessionRepo := &stubSessionRepo{
			revokeAllForUserFunc: func(_ context.Context, userID, _ string) (int64, error) {
				revokedUserID = userID
				return 1, nil
			},
		}
		svc := NewUserService(log, userRepo, &stubPasswordHasher{}, sessionRepo)

		res, err := svc.SetUserStatus(ctx, domain.SetUserStatusRequest{
			CallerID: "admin-1",
			UserID:   "dev-1",
			Status:   domain.StatusSuspended,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Status != domain.StatusSuspended || updated.Status != domain.StatusSuspended {
			t.Errorf("expected status suspended, got %v", res.Status)
		}
		if revokedUserID != "dev-1" {
			t.Errorf("expected sessions revoked for dev-1, got %s", revokedUserID)
		}
	})

	t.Run("admin activates suspended developer", func(t *testing.T) {
		t.Parallel()

		suspendedDev := &domain.User{ID: "dev-1", Role: domain.RoleDeveloper, Status: domain.StatusSuspended}
		userRepo := &stubUserRepo{
			getByIDFunc: func(_ context.Context, id string) (*domain.User, error) {
				if id == "admin-1" {
					return admin, nil
				}
				if id == "dev-1" {
					return suspendedDev, nil
				}
				return nil, domain.ErrNotFound
			},
			updateFunc: func(_ context.Context, _ domain.User) error { return nil },
		}
		svc := NewUserService(log, userRepo, &stubPasswordHasher{}, nil)

		res, err := svc.SetUserStatus(ctx, domain.SetUserStatusRequest{
			CallerID: "admin-1",
			UserID:   "dev-1",
			Status:   domain.StatusActive,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Status != domain.StatusActive {
			t.Errorf("expected status active, got %v", res.Status)
		}
	})

	t.Run("moderator suspends developer", func(t *testing.T) {
		t.Parallel()

		userRepo := &stubUserRepo{
			getByIDFunc: func(_ context.Context, id string) (*domain.User, error) {
				if id == "mod-1" {
					return moderator, nil
				}
				if id == "dev-1" {
					return developer, nil
				}
				return nil, domain.ErrNotFound
			},
			updateFunc: func(_ context.Context, _ domain.User) error { return nil },
		}
		svc := NewUserService(log, userRepo, &stubPasswordHasher{}, nil)

		res, err := svc.SetUserStatus(ctx, domain.SetUserStatusRequest{
			CallerID: "mod-1",
			UserID:   "dev-1",
			Status:   domain.StatusSuspended,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Status != domain.StatusSuspended {
			t.Errorf("expected status suspended, got %v", res.Status)
		}
	})

	t.Run("moderator cannot suspend other moderator", func(t *testing.T) {
		t.Parallel()

		targetMod := &domain.User{ID: "mod-2", Role: domain.RoleModerator, Status: domain.StatusActive}
		userRepo := &stubUserRepo{
			getByIDFunc: func(_ context.Context, id string) (*domain.User, error) {
				if id == "mod-1" {
					return moderator, nil
				}
				if id == "mod-2" {
					return targetMod, nil
				}
				return nil, domain.ErrNotFound
			},
		}
		svc := NewUserService(log, userRepo, &stubPasswordHasher{}, nil)

		_, err := svc.SetUserStatus(ctx, domain.SetUserStatusRequest{
			CallerID: "mod-1",
			UserID:   "mod-2",
			Status:   domain.StatusSuspended,
		})
		if !errors.Is(err, domain.ErrCannotModifyModeratorStatus) {
			t.Fatalf("expected ErrCannotModifyModeratorStatus, got %v", err)
		}
	})

	t.Run("admin cannot suspend moderator (moderator cannot be suspended)", func(t *testing.T) {
		t.Parallel()

		targetMod := &domain.User{ID: "mod-2", Role: domain.RoleModerator, Status: domain.StatusActive}
		userRepo := &stubUserRepo{
			getByIDFunc: func(_ context.Context, id string) (*domain.User, error) {
				if id == "admin-1" {
					return admin, nil
				}
				if id == "mod-2" {
					return targetMod, nil
				}
				return nil, domain.ErrNotFound
			},
		}
		svc := NewUserService(log, userRepo, &stubPasswordHasher{}, nil)

		_, err := svc.SetUserStatus(ctx, domain.SetUserStatusRequest{
			CallerID: "admin-1",
			UserID:   "mod-2",
			Status:   domain.StatusSuspended,
		})
		if !errors.Is(err, domain.ErrCannotModifyModeratorStatus) {
			t.Fatalf("expected ErrCannotModifyModeratorStatus, got %v", err)
		}
	})

	t.Run("cannot modify admin status", func(t *testing.T) {
		t.Parallel()

		userRepo := &stubUserRepo{
			getByIDFunc: func(_ context.Context, id string) (*domain.User, error) {
				if id == "admin-1" {
					return admin, nil
				}
				return nil, domain.ErrNotFound
			},
		}
		svc := NewUserService(log, userRepo, &stubPasswordHasher{}, nil)

		_, err := svc.SetUserStatus(ctx, domain.SetUserStatusRequest{
			CallerID: "admin-1",
			UserID:   "admin-1",
			Status:   domain.StatusSuspended,
		})
		if !errors.Is(err, domain.ErrCannotModifyAdminStatus) {
			t.Fatalf("expected ErrCannotModifyAdminStatus, got %v", err)
		}
	})

	t.Run("moderator cannot set status to deleted", func(t *testing.T) {
		t.Parallel()

		userRepo := &stubUserRepo{
			getByIDFunc: func(_ context.Context, id string) (*domain.User, error) {
				if id == "mod-1" {
					return moderator, nil
				}
				if id == "dev-1" {
					return developer, nil
				}
				return nil, domain.ErrNotFound
			},
		}
		svc := NewUserService(log, userRepo, &stubPasswordHasher{}, nil)

		_, err := svc.SetUserStatus(ctx, domain.SetUserStatusRequest{
			CallerID: "mod-1",
			UserID:   "dev-1",
			Status:   domain.StatusDeleted,
		})
		if !errors.Is(err, domain.ErrPermissionDenied) {
			t.Fatalf("expected ErrPermissionDenied, got %v", err)
		}
	})

	t.Run("developer cannot change status", func(t *testing.T) {
		t.Parallel()

		userRepo := &stubUserRepo{
			getByIDFunc: func(_ context.Context, id string) (*domain.User, error) {
				if id == "dev-1" {
					return developer, nil
				}
				return nil, domain.ErrNotFound
			},
		}
		svc := NewUserService(log, userRepo, &stubPasswordHasher{}, nil)

		_, err := svc.SetUserStatus(ctx, domain.SetUserStatusRequest{
			CallerID: "dev-1",
			UserID:   "dev-1",
			Status:   domain.StatusSuspended,
		})
		if !errors.Is(err, domain.ErrPermissionDenied) {
			t.Fatalf("expected ErrPermissionDenied, got %v", err)
		}
	})
}

func TestUserService_DeleteUser(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	log := newTestLogger()

	admin := &domain.User{ID: "admin-1", Role: domain.RoleAdmin}
	moderator := &domain.User{ID: "mod-1", Role: domain.RoleModerator, Email: "mod@welwise.com"}
	developer := &domain.User{ID: "dev-1", Role: domain.RoleDeveloper, Email: "dev@example.com"}

	t.Run("admin soft-deletes developer", func(t *testing.T) {
		t.Parallel()

		var updated domain.User
		var revokedUserID string
		userRepo := &stubUserRepo{
			getByIDFunc: func(_ context.Context, id string) (*domain.User, error) {
				if id == "admin-1" {
					return admin, nil
				}
				if id == "dev-1" {
					return developer, nil
				}
				return nil, domain.ErrNotFound
			},
			updateFunc: func(_ context.Context, u domain.User) error {
				updated = u
				return nil
			},
		}
		sessionRepo := &stubSessionRepo{
			revokeAllForUserFunc: func(_ context.Context, userID, _ string) (int64, error) {
				revokedUserID = userID
				return 1, nil
			},
		}
		svc := NewUserService(log, userRepo, &stubPasswordHasher{}, sessionRepo)

		err := svc.DeleteUser(ctx, domain.DeleteUserRequest{CallerID: "admin-1", UserID: "dev-1"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Status != domain.StatusDeleted {
			t.Errorf("expected status deleted, got %v", updated.Status)
		}
		if revokedUserID != "dev-1" {
			t.Errorf("expected sessions revoked for dev-1, got %s", revokedUserID)
		}
	})

	t.Run("admin soft-deletes moderator", func(t *testing.T) {
		t.Parallel()

		var updated domain.User
		userRepo := &stubUserRepo{
			getByIDFunc: func(_ context.Context, id string) (*domain.User, error) {
				if id == "admin-1" {
					return admin, nil
				}
				if id == "mod-1" {
					return moderator, nil
				}
				return nil, domain.ErrNotFound
			},
			updateFunc: func(_ context.Context, u domain.User) error {
				updated = u
				return nil
			},
		}
		svc := NewUserService(log, userRepo, &stubPasswordHasher{}, nil)

		err := svc.DeleteUser(ctx, domain.DeleteUserRequest{CallerID: "admin-1", UserID: "mod-1"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Status != domain.StatusDeleted {
			t.Errorf("expected status deleted, got %v", updated.Status)
		}
	})

	t.Run("moderator cannot delete user", func(t *testing.T) {
		t.Parallel()

		userRepo := &stubUserRepo{
			getByIDFunc: func(_ context.Context, id string) (*domain.User, error) {
				if id == "mod-1" {
					return moderator, nil
				}
				if id == "dev-1" {
					return developer, nil
				}
				return nil, domain.ErrNotFound
			},
		}
		svc := NewUserService(log, userRepo, &stubPasswordHasher{}, nil)

		err := svc.DeleteUser(ctx, domain.DeleteUserRequest{CallerID: "mod-1", UserID: "dev-1"})
		if !errors.Is(err, domain.ErrPermissionDenied) {
			t.Fatalf("expected ErrPermissionDenied, got %v", err)
		}
	})

	t.Run("cannot delete admin", func(t *testing.T) {
		t.Parallel()

		otherAdmin := &domain.User{ID: "admin-2", Role: domain.RoleAdmin, Email: "admin2@welwise.com"}
		userRepo := &stubUserRepo{
			getByIDFunc: func(_ context.Context, id string) (*domain.User, error) {
				if id == "admin-1" {
					return admin, nil
				}
				if id == "admin-2" {
					return otherAdmin, nil
				}
				return nil, domain.ErrNotFound
			},
		}
		svc := NewUserService(log, userRepo, &stubPasswordHasher{}, nil)

		err := svc.DeleteUser(ctx, domain.DeleteUserRequest{CallerID: "admin-1", UserID: "admin-2"})
		if !errors.Is(err, domain.ErrCannotDeleteAdmin) {
			t.Fatalf("expected ErrCannotDeleteAdmin, got %v", err)
		}
	})
}
