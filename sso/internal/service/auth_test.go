package service

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/sso/internal/domain"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

// --- stubs for AuthService tests ---

type stubPasswordHasher struct {
	hashToReturn []byte
	hashErr      error
	compareErr   error
}

func (s *stubPasswordHasher) Hash(_ context.Context, _ string) ([]byte, error) {
	return s.hashToReturn, s.hashErr
}

func (s *stubPasswordHasher) Compare(_ context.Context, _ []byte, _ string) error {
	return s.compareErr
}

type stubUserRepo struct {
	createFunc     func(context.Context, domain.User) error
	getByIDFunc    func(context.Context, string) (*domain.User, error)
	getByEmailFunc func(context.Context, string) (*domain.User, error)
	updateFunc     func(context.Context, domain.User) error
	searchFunc     func(context.Context, string, int, int) ([]domain.User, int64, error)
}

func (s *stubUserRepo) Create(ctx context.Context, user domain.User) error {
	return s.createFunc(ctx, user)
}

func (s *stubUserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return s.getByIDFunc(ctx, id)
}

func (s *stubUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return s.getByEmailFunc(ctx, email)
}

func (s *stubUserRepo) Update(ctx context.Context, user domain.User) error {
	return s.updateFunc(ctx, user)
}

func (s *stubUserRepo) Search(ctx context.Context, query string, limit, offset int) ([]domain.User, int64, error) {
	return s.searchFunc(ctx, query, limit, offset)
}

type stubSessionRepo struct {
	createFunc            func(context.Context, domain.Session) error
	getByIDFunc           func(context.Context, string) (*domain.Session, error)
	getByUserIDFunc       func(context.Context, string) ([]domain.Session, error)
	getByRefreshTokenHash func(context.Context, string) (*domain.Session, error)
	updateFunc            func(context.Context, domain.Session) error
	revokeFunc            func(context.Context, string) error
	revokeAllForUserFunc  func(context.Context, string, string) (int64, error)
}

func (s *stubSessionRepo) Create(ctx context.Context, session domain.Session) error {
	return s.createFunc(ctx, session)
}

func (s *stubSessionRepo) GetByID(ctx context.Context, id string) (*domain.Session, error) {
	return s.getByIDFunc(ctx, id)
}

func (s *stubSessionRepo) GetByUserID(ctx context.Context, userID string) ([]domain.Session, error) {
	return s.getByUserIDFunc(ctx, userID)
}

func (s *stubSessionRepo) GetByRefreshTokenHash(ctx context.Context, hash string) (*domain.Session, error) {
	return s.getByRefreshTokenHash(ctx, hash)
}

func (s *stubSessionRepo) Update(ctx context.Context, session domain.Session) error {
	return s.updateFunc(ctx, session)
}

func (s *stubSessionRepo) Revoke(ctx context.Context, id string) error {
	return s.revokeFunc(ctx, id)
}

func (s *stubSessionRepo) RevokeAllForUser(ctx context.Context, userID, excludeSessionID string) (int64, error) {
	return s.revokeAllForUserFunc(ctx, userID, excludeSessionID)
}

type stubTokenManager struct {
	generateAccessFunc  func(context.Context, domain.Claims) (string, time.Time, error)
	generateRefreshFunc func(context.Context) (string, error)
	parseAccessFunc     func(context.Context, string) (*domain.Claims, error)
}

func (s *stubTokenManager) GenerateAccessToken(ctx context.Context, claims domain.Claims) (string, time.Time, error) {
	return s.generateAccessFunc(ctx, claims)
}

func (s *stubTokenManager) GenerateRefreshToken(ctx context.Context) (string, error) {
	return s.generateRefreshFunc(ctx)
}

func (s *stubTokenManager) ParseAccessToken(ctx context.Context, token string) (*domain.Claims, error) {
	return s.parseAccessFunc(ctx, token)
}

type stubEmailSender struct {
	sendVerificationFunc  func(context.Context, string, string) error
	sendPasswordResetFunc func(context.Context, string, string) error
}

func (s *stubEmailSender) SendVerificationEmail(ctx context.Context, email, code string) error {
	return s.sendVerificationFunc(ctx, email, code)
}

func (s *stubEmailSender) SendPasswordResetEmail(ctx context.Context, email, token string) error {
	return s.sendPasswordResetFunc(ctx, email, token)
}

type stubEmailVerifyStore struct {
	storeFunc      func(context.Context, string, string) error
	verifyFunc     func(context.Context, string, string) (bool, error)
	getEmailByCode func(context.Context, string) (string, error)
}

func (s *stubEmailVerifyStore) Store(ctx context.Context, email, code string) error {
	return s.storeFunc(ctx, email, code)
}

func (s *stubEmailVerifyStore) Verify(ctx context.Context, email, code string) (bool, error) {
	return s.verifyFunc(ctx, email, code)
}

func (s *stubEmailVerifyStore) GetEmailByCode(ctx context.Context, code string) (string, error) {
	return s.getEmailByCode(ctx, code)
}

type stubPasswordResetStore struct {
	storeFunc   func(context.Context, string, string) error
	consumeFunc func(context.Context, string) (string, error)
}

func (s *stubPasswordResetStore) Store(ctx context.Context, email, token string) error {
	return s.storeFunc(ctx, email, token)
}

func (s *stubPasswordResetStore) Consume(ctx context.Context, token string) (string, error) {
	return s.consumeFunc(ctx, token)
}

type stubSessionCache struct {
	setFunc        func(context.Context, *domain.Session) error
	getFunc        func(context.Context, string) (*domain.Session, error)
	invalidateFunc func(context.Context, string) error
}

func (s *stubSessionCache) Set(ctx context.Context, session *domain.Session) error {
	return s.setFunc(ctx, session)
}

func (s *stubSessionCache) Get(ctx context.Context, sessionID string) (*domain.Session, error) {
	return s.getFunc(ctx, sessionID)
}

func (s *stubSessionCache) Invalidate(ctx context.Context, sessionID string) error {
	return s.invalidateFunc(ctx, sessionID)
}

// --- AuthService tests ---

func TestAuthService_Register(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	log := newTestLogger()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		userRepo := &stubUserRepo{
			createFunc: func(context.Context, domain.User) error { return nil },
			getByEmailFunc: func(_ context.Context, email string) (*domain.User, error) {
				return &domain.User{ID: "user-1", Email: email, EmailVerified: false, Status: domain.StatusSuspended, Role: domain.RoleDeveloper}, nil
			},
		}

		emailStore := &stubEmailVerifyStore{
			storeFunc: func(context.Context, string, string) error { return nil },
		}

		emailSender := &stubEmailSender{
			sendVerificationFunc: func(context.Context, string, string) error { return nil },
		}

		svc := NewAuthService(log, userRepo, &stubSessionRepo{}, &stubTokenManager{}, &stubPasswordHasher{hashToReturn: []byte("hashed")}, emailStore, &stubPasswordResetStore{}, emailSender, 24*time.Hour)

		resp, err := svc.Register(ctx, domain.RegisterRequest{Email: "test@example.com", Password: "password123", DisplayName: "Test User"})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.User.Email != "test@example.com" {
			t.Errorf("expected email test@example.com, got %s", resp.User.Email)
		}
		if resp.User.EmailVerified {
			t.Error("expected email to be unverified after registration")
		}
	})

	t.Run("hash password error", func(t *testing.T) {
		t.Parallel()

		svc := NewAuthService(log, &stubUserRepo{}, &stubSessionRepo{}, &stubTokenManager{}, &stubPasswordHasher{hashErr: errors.New("hash error")}, &stubEmailVerifyStore{}, &stubPasswordResetStore{}, &stubEmailSender{}, 24*time.Hour)

		_, err := svc.Register(ctx, domain.RegisterRequest{Email: "test@example.com", Password: "password123"})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("create user error", func(t *testing.T) {
		t.Parallel()

		userRepo := &stubUserRepo{createFunc: func(context.Context, domain.User) error { return errors.New("db error") }}
		svc := NewAuthService(log, userRepo, &stubSessionRepo{}, &stubTokenManager{}, &stubPasswordHasher{hashToReturn: []byte("hashed")}, &stubEmailVerifyStore{}, &stubPasswordResetStore{}, &stubEmailSender{}, 24*time.Hour)

		_, err := svc.Register(ctx, domain.RegisterRequest{Email: "test@example.com", Password: "password123"})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("get created user error", func(t *testing.T) {
		t.Parallel()

		userRepo := &stubUserRepo{
			createFunc:     func(context.Context, domain.User) error { return nil },
			getByEmailFunc: func(context.Context, string) (*domain.User, error) { return nil, errors.New("not found") },
		}
		svc := NewAuthService(log, userRepo, &stubSessionRepo{}, &stubTokenManager{}, &stubPasswordHasher{hashToReturn: []byte("hashed")}, &stubEmailVerifyStore{}, &stubPasswordResetStore{}, &stubEmailSender{}, 24*time.Hour)

		_, err := svc.Register(ctx, domain.RegisterRequest{Email: "test@example.com", Password: "password123"})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("store verification code error", func(t *testing.T) {
		t.Parallel()

		userRepo := &stubUserRepo{
			createFunc: func(context.Context, domain.User) error { return nil },
			getByEmailFunc: func(_ context.Context, email string) (*domain.User, error) {
				return &domain.User{ID: "user-1", Email: email}, nil
			},
		}
		emailStore := &stubEmailVerifyStore{storeFunc: func(context.Context, string, string) error { return errors.New("store error") }}
		svc := NewAuthService(log, userRepo, &stubSessionRepo{}, &stubTokenManager{}, &stubPasswordHasher{hashToReturn: []byte("hashed")}, emailStore, &stubPasswordResetStore{}, &stubEmailSender{}, 24*time.Hour)

		_, err := svc.Register(ctx, domain.RegisterRequest{Email: "test@example.com", Password: "password123"})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("email sender failure should not fail registration", func(t *testing.T) {
		t.Parallel()

		userRepo := &stubUserRepo{
			createFunc: func(context.Context, domain.User) error { return nil },
			getByEmailFunc: func(_ context.Context, email string) (*domain.User, error) {
				return &domain.User{ID: "user-1", Email: email}, nil
			},
		}
		emailStore := &stubEmailVerifyStore{storeFunc: func(context.Context, string, string) error { return nil }}
		emailSender := &stubEmailSender{sendVerificationFunc: func(context.Context, string, string) error { return errors.New("smtp error") }}
		svc := NewAuthService(log, userRepo, &stubSessionRepo{}, &stubTokenManager{}, &stubPasswordHasher{hashToReturn: []byte("hashed")}, emailStore, &stubPasswordResetStore{}, emailSender, 24*time.Hour)

		_, err := svc.Register(ctx, domain.RegisterRequest{Email: "test@example.com", Password: "password123"})
		if err != nil {
			t.Fatalf("expected no error when email fails, got: %v", err)
		}
	})
}

func TestAuthService_Login(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	log := newTestLogger()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		userRepo := &stubUserRepo{
			getByEmailFunc: func(_ context.Context, email string) (*domain.User, error) {
				return &domain.User{ID: "user-1", Email: email, EmailVerified: true, Status: domain.StatusActive, PasswordHash: []byte("hashed"), Role: domain.RoleDeveloper}, nil
			},
			getByIDFunc: func(_ context.Context, id string) (*domain.User, error) {
				return &domain.User{ID: id, Email: "test@example.com", Role: domain.RoleDeveloper}, nil
			},
		}
		sessionRepo := &stubSessionRepo{
			createFunc: func(context.Context, domain.Session) error { return nil },
			getByRefreshTokenHash: func(context.Context, string) (*domain.Session, error) {
				return &domain.Session{ID: "session-1", UserID: "user-1", ExpiresAt: time.Now().Add(24 * time.Hour)}, nil
			},
		}
		tokenManager := &stubTokenManager{
			generateRefreshFunc: func(context.Context) (string, error) { return "refresh-token", nil },
			generateAccessFunc: func(context.Context, domain.Claims) (string, time.Time, error) {
				return "access-token", time.Now().Add(15 * time.Minute), nil
			},
		}
		svc := NewAuthService(log, userRepo, sessionRepo, tokenManager, &stubPasswordHasher{}, &stubEmailVerifyStore{}, &stubPasswordResetStore{}, &stubEmailSender{}, 24*time.Hour)

		resp, err := svc.Login(ctx, domain.LoginRequest{Email: "test@example.com", Password: "password123", UserAgent: "test-agent", IPAddress: "127.0.0.1"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Tokens.AccessToken != "access-token" {
			t.Errorf("expected access-token, got %s", resp.Tokens.AccessToken)
		}
		if resp.Tokens.RefreshToken != "refresh-token" {
			t.Errorf("expected refresh-token, got %s", resp.Tokens.RefreshToken)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		t.Parallel()

		userRepo := &stubUserRepo{getByEmailFunc: func(context.Context, string) (*domain.User, error) { return nil, errors.New("not found") }}
		svc := NewAuthService(log, userRepo, &stubSessionRepo{}, &stubTokenManager{}, &stubPasswordHasher{}, &stubEmailVerifyStore{}, &stubPasswordResetStore{}, &stubEmailSender{}, 24*time.Hour)

		_, err := svc.Login(ctx, domain.LoginRequest{Email: "test@example.com", Password: "password123"})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("suspended user", func(t *testing.T) {
		t.Parallel()

		userRepo := &stubUserRepo{getByEmailFunc: func(context.Context, string) (*domain.User, error) {
			return &domain.User{EmailVerified: true, Status: domain.StatusSuspended}, nil
		}}
		svc := NewAuthService(log, userRepo, &stubSessionRepo{}, &stubTokenManager{}, &stubPasswordHasher{}, &stubEmailVerifyStore{}, &stubPasswordResetStore{}, &stubEmailSender{}, 24*time.Hour)

		_, err := svc.Login(ctx, domain.LoginRequest{Email: "test@example.com", Password: "password123"})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, domain.ErrUserSuspended) {
			t.Errorf("expected ErrUserSuspended, got %v", err)
		}
	})

	t.Run("email not verified", func(t *testing.T) {
		t.Parallel()

		userRepo := &stubUserRepo{getByEmailFunc: func(context.Context, string) (*domain.User, error) {
			return &domain.User{EmailVerified: false, Status: domain.StatusActive}, nil
		}}
		svc := NewAuthService(log, userRepo, &stubSessionRepo{}, &stubTokenManager{}, &stubPasswordHasher{}, &stubEmailVerifyStore{}, &stubPasswordResetStore{}, &stubEmailSender{}, 24*time.Hour)

		_, err := svc.Login(ctx, domain.LoginRequest{Email: "test@example.com", Password: "password123"})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, domain.ErrEmailNotVerified) {
			t.Errorf("expected ErrEmailNotVerified, got %v", err)
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		t.Parallel()

		userRepo := &stubUserRepo{getByEmailFunc: func(context.Context, string) (*domain.User, error) {
			return &domain.User{EmailVerified: true, Status: domain.StatusActive, PasswordHash: []byte("hashed")}, nil
		}}
		svc := NewAuthService(log, userRepo, &stubSessionRepo{}, &stubTokenManager{}, &stubPasswordHasher{compareErr: domain.ErrInvalidPassword}, &stubEmailVerifyStore{}, &stubPasswordResetStore{}, &stubEmailSender{}, 24*time.Hour)

		_, err := svc.Login(ctx, domain.LoginRequest{Email: "test@example.com", Password: "wrongpassword"})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestAuthService_RefreshToken(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	log := newTestLogger()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		sessionRepo := &stubSessionRepo{
			getByRefreshTokenHash: func(context.Context, string) (*domain.Session, error) {
				return &domain.Session{ID: "session-2", UserID: "user-1", ExpiresAt: time.Now().Add(24 * time.Hour)}, nil
			},
			createFunc: func(context.Context, domain.Session) error { return nil },
			revokeFunc: func(context.Context, string) error { return nil },
		}
		userRepo := &stubUserRepo{
			getByIDFunc: func(_ context.Context, id string) (*domain.User, error) {
				return &domain.User{ID: id, Email: "test@example.com", Role: domain.RoleDeveloper}, nil
			},
		}
		tokenManager := &stubTokenManager{
			generateRefreshFunc: func(context.Context) (string, error) { return "new-refresh-token", nil },
			generateAccessFunc: func(context.Context, domain.Claims) (string, time.Time, error) {
				return "new-access-token", time.Now().Add(15 * time.Minute), nil
			},
		}
		svc := NewAuthService(log, userRepo, sessionRepo, tokenManager, &stubPasswordHasher{}, &stubEmailVerifyStore{}, &stubPasswordResetStore{}, &stubEmailSender{}, 24*time.Hour)

		resp, err := svc.RefreshToken(ctx, domain.RefreshTokenRequest{RefreshToken: "old-refresh-token"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Tokens.AccessToken != "new-access-token" {
			t.Errorf("expected new-access-token, got %s", resp.Tokens.AccessToken)
		}
	})

	t.Run("session not found", func(t *testing.T) {
		t.Parallel()

		sessionRepo := &stubSessionRepo{getByRefreshTokenHash: func(context.Context, string) (*domain.Session, error) { return nil, errors.New("not found") }}
		svc := NewAuthService(log, &stubUserRepo{}, sessionRepo, &stubTokenManager{}, &stubPasswordHasher{}, &stubEmailVerifyStore{}, &stubPasswordResetStore{}, &stubEmailSender{}, 24*time.Hour)

		_, err := svc.RefreshToken(ctx, domain.RefreshTokenRequest{RefreshToken: "invalid-token"})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("expired session", func(t *testing.T) {
		t.Parallel()

		sessionRepo := &stubSessionRepo{
			getByRefreshTokenHash: func(context.Context, string) (*domain.Session, error) {
				return &domain.Session{ID: "session-1", UserID: "user-1", ExpiresAt: time.Now().Add(-1 * time.Hour), Revoked: false}, nil
			},
			revokeFunc: func(context.Context, string) error { return nil },
		}
		svc := NewAuthService(log, &stubUserRepo{}, sessionRepo, &stubTokenManager{}, &stubPasswordHasher{}, &stubEmailVerifyStore{}, &stubPasswordResetStore{}, &stubEmailSender{}, 24*time.Hour)

		_, err := svc.RefreshToken(ctx, domain.RefreshTokenRequest{RefreshToken: "expired-token"})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, domain.ErrTokenExpired) {
			t.Errorf("expected ErrTokenExpired, got %v", err)
		}
	})
}

func TestAuthService_Logout(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	log := newTestLogger()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		sessionRepo := &stubSessionRepo{
			getByRefreshTokenHash: func(context.Context, string) (*domain.Session, error) {
				return &domain.Session{ID: "session-1", UserID: "user-1"}, nil
			},
			revokeFunc: func(context.Context, string) error { return nil },
		}
		svc := NewAuthService(log, &stubUserRepo{}, sessionRepo, &stubTokenManager{}, &stubPasswordHasher{}, &stubEmailVerifyStore{}, &stubPasswordResetStore{}, &stubEmailSender{}, 24*time.Hour)

		err := svc.Logout(ctx, domain.LogoutRequest{RefreshToken: "refresh-token"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("session not found", func(t *testing.T) {
		t.Parallel()

		sessionRepo := &stubSessionRepo{getByRefreshTokenHash: func(context.Context, string) (*domain.Session, error) { return nil, errors.New("not found") }}
		svc := NewAuthService(log, &stubUserRepo{}, sessionRepo, &stubTokenManager{}, &stubPasswordHasher{}, &stubEmailVerifyStore{}, &stubPasswordResetStore{}, &stubEmailSender{}, 24*time.Hour)

		err := svc.Logout(ctx, domain.LogoutRequest{RefreshToken: "invalid-token"})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestAuthService_VerifyEmail(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	log := newTestLogger()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		emailStore := &stubEmailVerifyStore{
			getEmailByCode: func(context.Context, string) (string, error) { return "test@example.com", nil },
			verifyFunc:     func(context.Context, string, string) (bool, error) { return true, nil },
		}
		userRepo := &stubUserRepo{
			getByEmailFunc: func(_ context.Context, email string) (*domain.User, error) {
				return &domain.User{ID: "user-1", Email: email, EmailVerified: false, Status: domain.StatusSuspended}, nil
			},
			updateFunc: func(context.Context, domain.User) error { return nil },
			getByIDFunc: func(_ context.Context, id string) (*domain.User, error) {
				return &domain.User{ID: id, Email: "test@example.com", Role: domain.RoleDeveloper}, nil
			},
		}
		sessionRepo := &stubSessionRepo{
			createFunc: func(context.Context, domain.Session) error { return nil },
			getByRefreshTokenHash: func(context.Context, string) (*domain.Session, error) {
				return &domain.Session{ID: "session-1", UserID: "user-1", ExpiresAt: time.Now().Add(24 * time.Hour)}, nil
			},
		}
		tokenManager := &stubTokenManager{
			generateRefreshFunc: func(context.Context) (string, error) { return "refresh-token", nil },
			generateAccessFunc: func(context.Context, domain.Claims) (string, time.Time, error) {
				return "access-token", time.Now().Add(15 * time.Minute), nil
			},
		}
		svc := NewAuthService(log, userRepo, sessionRepo, tokenManager, &stubPasswordHasher{}, emailStore, &stubPasswordResetStore{}, &stubEmailSender{}, 24*time.Hour)

		err := svc.VerifyEmail(ctx, domain.VerifyEmailRequest{VerificationCode: "123456"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("invalid verification code", func(t *testing.T) {
		t.Parallel()

		emailStore := &stubEmailVerifyStore{getEmailByCode: func(context.Context, string) (string, error) { return "", nil }}
		svc := NewAuthService(log, &stubUserRepo{}, &stubSessionRepo{}, &stubTokenManager{}, &stubPasswordHasher{}, emailStore, &stubPasswordResetStore{}, &stubEmailSender{}, 24*time.Hour)

		err := svc.VerifyEmail(ctx, domain.VerifyEmailRequest{VerificationCode: "invalid"})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, domain.ErrInvalidToken) {
			t.Errorf("expected ErrInvalidToken, got %v", err)
		}
	})
}

func TestAuthService_ResendVerificationEmail(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	log := newTestLogger()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		userRepo := &stubUserRepo{getByEmailFunc: func(_ context.Context, email string) (*domain.User, error) {
			return &domain.User{ID: "user-1", Email: email, EmailVerified: false}, nil
		}}
		emailStore := &stubEmailVerifyStore{storeFunc: func(context.Context, string, string) error { return nil }}
		emailSender := &stubEmailSender{sendVerificationFunc: func(context.Context, string, string) error { return nil }}
		svc := NewAuthService(log, userRepo, &stubSessionRepo{}, &stubTokenManager{}, &stubPasswordHasher{}, emailStore, &stubPasswordResetStore{}, emailSender, 24*time.Hour)

		err := svc.ResendVerificationEmail(ctx, domain.ResendVerificationRequest{Email: "test@example.com"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("user not found should return success", func(t *testing.T) {
		t.Parallel()

		userRepo := &stubUserRepo{getByEmailFunc: func(context.Context, string) (*domain.User, error) { return nil, errors.New("not found") }}
		svc := NewAuthService(log, userRepo, &stubSessionRepo{}, &stubTokenManager{}, &stubPasswordHasher{}, &stubEmailVerifyStore{}, &stubPasswordResetStore{}, &stubEmailSender{}, 24*time.Hour)

		err := svc.ResendVerificationEmail(ctx, domain.ResendVerificationRequest{Email: "nonexistent@example.com"})
		if err != nil {
			t.Fatalf("expected no error for nonexistent user, got: %v", err)
		}
	})

	t.Run("already verified should return success", func(t *testing.T) {
		t.Parallel()

		userRepo := &stubUserRepo{getByEmailFunc: func(_ context.Context, email string) (*domain.User, error) {
			return &domain.User{ID: "user-1", Email: email, EmailVerified: true}, nil
		}}
		svc := NewAuthService(log, userRepo, &stubSessionRepo{}, &stubTokenManager{}, &stubPasswordHasher{}, &stubEmailVerifyStore{}, &stubPasswordResetStore{}, &stubEmailSender{}, 24*time.Hour)

		err := svc.ResendVerificationEmail(ctx, domain.ResendVerificationRequest{Email: "verified@example.com"})
		if err != nil {
			t.Fatalf("expected no error for verified user, got: %v", err)
		}
	})
}

func TestAuthService_RequestPasswordReset(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	log := newTestLogger()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		userRepo := &stubUserRepo{getByEmailFunc: func(_ context.Context, email string) (*domain.User, error) {
			return &domain.User{ID: "user-1", Email: email}, nil
		}}
		resetStore := &stubPasswordResetStore{storeFunc: func(context.Context, string, string) error { return nil }}
		emailSender := &stubEmailSender{sendPasswordResetFunc: func(context.Context, string, string) error { return nil }}
		svc := NewAuthService(log, userRepo, &stubSessionRepo{}, &stubTokenManager{}, &stubPasswordHasher{}, &stubEmailVerifyStore{}, resetStore, emailSender, 24*time.Hour)

		err := svc.RequestPasswordReset(ctx, domain.PasswordResetRequest{Email: "test@example.com"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("nonexistent user should return success", func(t *testing.T) {
		t.Parallel()

		userRepo := &stubUserRepo{getByEmailFunc: func(context.Context, string) (*domain.User, error) { return nil, errors.New("not found") }}
		svc := NewAuthService(log, userRepo, &stubSessionRepo{}, &stubTokenManager{}, &stubPasswordHasher{}, &stubEmailVerifyStore{}, &stubPasswordResetStore{}, &stubEmailSender{}, 24*time.Hour)

		err := svc.RequestPasswordReset(ctx, domain.PasswordResetRequest{Email: "nonexistent@example.com"})
		if err != nil {
			t.Fatalf("expected no error for nonexistent user, got: %v", err)
		}
	})
}

func TestAuthService_ResetPassword(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	log := newTestLogger()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		resetStore := &stubPasswordResetStore{consumeFunc: func(context.Context, string) (string, error) { return "test@example.com", nil }}
		userRepo := &stubUserRepo{
			getByEmailFunc: func(_ context.Context, email string) (*domain.User, error) {
				return &domain.User{ID: "user-1", Email: email, PasswordHash: []byte("old-hash")}, nil
			},
			updateFunc: func(context.Context, domain.User) error { return nil },
		}
		sessionRepo := &stubSessionRepo{revokeAllForUserFunc: func(context.Context, string, string) (int64, error) { return 1, nil }}
		svc := NewAuthService(log, userRepo, sessionRepo, &stubTokenManager{}, &stubPasswordHasher{hashToReturn: []byte("new-hash")}, &stubEmailVerifyStore{}, resetStore, &stubEmailSender{}, 24*time.Hour)

		err := svc.ResetPassword(ctx, domain.ResetPasswordRequest{ResetToken: "reset-token", NewPassword: "new-password"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("invalid reset token", func(t *testing.T) {
		t.Parallel()

		resetStore := &stubPasswordResetStore{consumeFunc: func(context.Context, string) (string, error) { return "", errors.New("token not found") }}
		svc := NewAuthService(log, &stubUserRepo{}, &stubSessionRepo{}, &stubTokenManager{}, &stubPasswordHasher{}, &stubEmailVerifyStore{}, resetStore, &stubEmailSender{}, 24*time.Hour)

		err := svc.ResetPassword(ctx, domain.ResetPasswordRequest{ResetToken: "invalid-token", NewPassword: "new-password"})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
