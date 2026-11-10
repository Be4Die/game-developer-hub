// Package service содержит бизнес-логику SSO-сервиса.
package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/Be4Die/game-developer-hub/sso/internal/domain"
	cryptoprov "github.com/Be4Die/game-developer-hub/sso/internal/infrastructure/crypto"
)

// AuthService реализует логику аутентификации.
type AuthService struct {
	log                *slog.Logger
	userRepo           domain.UserRepository
	sessionRepo        domain.SessionRepository
	tokenManager       domain.TokenManager
	passwordHasher     domain.PasswordHasher
	emailVerifyStore   domain.EmailVerificationStore
	passwordResetStore domain.PasswordResetStore
	emailSender        domain.EmailSender
	refreshTokenTTL    time.Duration
	// loginLimiter и verifyLimiter опциональны: nil отключает ограничение попыток.
	loginLimiter  domain.AttemptLimiter
	verifyLimiter domain.AttemptLimiter
}

// WithAttemptLimiters подключает ограничители неудачных попыток входа и подтверждения email.
func (s *AuthService) WithAttemptLimiters(login, verify domain.AttemptLimiter) *AuthService {
	s.loginLimiter = login
	s.verifyLimiter = verify
	return s
}

// blocked проверяет лимит попыток. Ошибка хранилища не должна открывать перебор, поэтому трактуется как блокировка.
func (s *AuthService) blocked(ctx context.Context, l domain.AttemptLimiter, key string) bool {
	if l == nil {
		return false
	}
	isBlocked, err := l.Blocked(ctx, key)
	if err != nil {
		s.log.Error("attempt limiter unavailable", slog.String("error", err.Error()))
		return true
	}
	return isBlocked
}

func (s *AuthService) registerFailure(ctx context.Context, l domain.AttemptLimiter, key string) {
	if l == nil {
		return
	}
	if err := l.Fail(ctx, key); err != nil {
		s.log.Error("failed to register failed attempt", slog.String("error", err.Error()))
	}
}

// NewAuthService создаёт сервис аутентификации.
func NewAuthService(
	log *slog.Logger,
	userRepo domain.UserRepository,
	sessionRepo domain.SessionRepository,
	tokenManager domain.TokenManager,
	passwordHasher domain.PasswordHasher,
	emailVerifyStore domain.EmailVerificationStore,
	passwordResetStore domain.PasswordResetStore,
	emailSender domain.EmailSender,
	refreshTokenTTL time.Duration,
) *AuthService {
	return &AuthService{
		log:                log,
		userRepo:           userRepo,
		sessionRepo:        sessionRepo,
		tokenManager:       tokenManager,
		passwordHasher:     passwordHasher,
		emailVerifyStore:   emailVerifyStore,
		passwordResetStore: passwordResetStore,
		emailSender:        emailSender,
		refreshTokenTTL:    refreshTokenTTL,
	}
}

// Register регистрирует нового пользователя.
func (s *AuthService) Register(ctx context.Context, req domain.RegisterRequest) (domain.RegisterResponse, error) {
	const op = "AuthService.Register"

	// Хешируем пароль.
	passwordHash, err := s.passwordHasher.Hash(ctx, req.Password)
	if err != nil {
		return domain.RegisterResponse{}, fmt.Errorf("%s: hash password: %w", op, err)
	}

	now := time.Now()
	user := domain.User{
		Email:         req.Email,
		PasswordHash:  passwordHash,
		DisplayName:   req.DisplayName,
		Role:          domain.RoleDeveloper,
		Status:        domain.StatusSuspended,
		EmailVerified: false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// Создаём пользователя.
	if err := s.userRepo.Create(ctx, user); err != nil {
		return domain.RegisterResponse{}, fmt.Errorf("%s: create user: %w", op, err)
	}

	// Получаем пользователя с ID из БД.
	createdUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return domain.RegisterResponse{}, fmt.Errorf("%s: get created user: %w", op, err)
	}

	// Генерируем 6-значный код верификации.
	code, err := generateVerificationCode()
	if err != nil {
		return domain.RegisterResponse{}, fmt.Errorf("%s: generate code: %w", op, err)
	}

	// Сохраняем код в Valkey.
	if err := s.emailVerifyStore.Store(ctx, req.Email, code); err != nil {
		return domain.RegisterResponse{}, fmt.Errorf("%s: store verification code: %w", op, err)
	}

	// Отправляем письмо с кодом.
	if err := s.emailSender.SendVerificationEmail(ctx, req.Email, code); err != nil {
		s.log.Warn("failed to send verification email", slog.String("email", req.Email), slog.String("error", err.Error()))
	}

	s.log.Info("user registered, awaiting email verification",
		slog.String("user_id", createdUser.ID),
		slog.String("email", req.Email),
	)

	// Токены не выдаём до верификации email.
	return domain.RegisterResponse{User: *createdUser}, nil
}

// Login выполняет вход пользователя.
func (s *AuthService) Login(ctx context.Context, req domain.LoginRequest) (domain.LoginResponse, error) {
	const op = "AuthService.Login"

	if s.blocked(ctx, s.loginLimiter, req.Email) {
		return domain.LoginResponse{}, fmt.Errorf("%s: %w", op, domain.ErrTooManyAttempts)
	}

	// Неизвестный email и неверный пароль дают одну и ту же ошибку, чтобы не раскрывать наличие учётной записи.
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		s.registerFailure(ctx, s.loginLimiter, req.Email)
		return domain.LoginResponse{}, fmt.Errorf("%s: %w", op, domain.ErrInvalidPassword)
	}

	// Пароль проверяется до статуса: о блокировке узнаёт только владелец учётной записи.
	if err := s.passwordHasher.Compare(ctx, user.PasswordHash, req.Password); err != nil {
		s.registerFailure(ctx, s.loginLimiter, req.Email)
		return domain.LoginResponse{}, fmt.Errorf("%s: %w", op, domain.ErrInvalidPassword)
	}

	if user.Status != domain.StatusActive {
		// Незавершённая регистрация: аккаунт создаётся приостановленным до подтверждения email.
		if !user.EmailVerified && !user.Role.IsInternal() && user.Status == domain.StatusSuspended {
			return domain.LoginResponse{}, fmt.Errorf("%s: %w", op, domain.ErrEmailNotVerified)
		}
		return domain.LoginResponse{}, fmt.Errorf("%s: %w", op, domain.ErrUserSuspended)
	}

	// Проверяем email_verified (кроме внутренних пользователей).
	if !user.EmailVerified && !user.Role.IsInternal() {
		return domain.LoginResponse{}, fmt.Errorf("%s: %w", op, domain.ErrEmailNotVerified)
	}

	if s.loginLimiter != nil {
		_ = s.loginLimiter.Reset(ctx, req.Email)
	}

	// Создаём сессию и токены.
	tokens, session, err := s.createSession(ctx, user.ID, req.UserAgent, req.IPAddress)
	if err != nil {
		return domain.LoginResponse{}, fmt.Errorf("%s: create session: %w", op, err)
	}

	s.log.Info("user logged in",
		slog.String("user_id", user.ID),
		slog.String("session_id", session.ID),
	)

	return domain.LoginResponse{
		User:   *user,
		Tokens: tokens,
	}, nil
}

// RefreshToken обновляет пару токенов (rotation).
func (s *AuthService) RefreshToken(ctx context.Context, req domain.RefreshTokenRequest) (domain.RefreshTokenResponse, error) {
	const op = "AuthService.RefreshToken"

	// Хеш текущего refresh token.
	rtHash := cryptoprov.HashToken(req.RefreshToken)

	// Находим сессию по hash refresh token.
	session, err := s.sessionRepo.GetByRefreshTokenHash(ctx, rtHash)
	if errors.Is(err, domain.ErrNotFound) {
		// Отозванный или чужой токен — это ошибка аутентификации, а не «ресурс не найден».
		return domain.RefreshTokenResponse{}, fmt.Errorf("%s: %w", op, domain.ErrInvalidToken)
	}
	if err != nil {
		return domain.RefreshTokenResponse{}, fmt.Errorf("%s: find session: %w", op, err)
	}

	// Проверяем что сессия активна.
	if !session.IsActive() {
		// Отзываем истёкшую сессию.
		_ = s.sessionRepo.Revoke(ctx, session.ID)
		return domain.RefreshTokenResponse{}, fmt.Errorf("%s: %w", op, domain.ErrTokenExpired)
	}

	// Обновляем last_used_at.
	session.LastUsedAt = time.Now()

	// Генерируем новую пару токенов.
	newTokens, newSession, err := s.rotateSession(ctx, session)
	if err != nil {
		return domain.RefreshTokenResponse{}, fmt.Errorf("%s: rotate session: %w", op, err)
	}

	// Старая сессия инвалидируется (revoked = true).
	if err := s.sessionRepo.Revoke(ctx, session.ID); err != nil {
		s.log.Warn("failed to revoke old session during rotation",
			slog.String("session_id", session.ID),
			slog.String("error", err.Error()),
		)
	}

	s.log.Debug("token rotated",
		slog.String("user_id", session.UserID),
		slog.String("old_session_id", session.ID),
		slog.String("new_session_id", newSession.ID),
	)

	return domain.RefreshTokenResponse{Tokens: newTokens}, nil
}

// Logout выполняет выход пользователя.
func (s *AuthService) Logout(ctx context.Context, req domain.LogoutRequest) error {
	const op = "AuthService.Logout"

	rtHash := cryptoprov.HashToken(req.RefreshToken)

	session, err := s.sessionRepo.GetByRefreshTokenHash(ctx, rtHash)
	if err != nil {
		return fmt.Errorf("%s: find session: %w", op, err)
	}

	if err := s.sessionRepo.Revoke(ctx, session.ID); err != nil {
		return fmt.Errorf("%s: revoke session: %w", op, err)
	}

	s.log.Info("user logged out", slog.String("session_id", session.ID))
	return nil
}

// VerifyEmail подтверждает email пользователя по паре «адрес + код».
// Число неверных вводов ограничено по адресу, поэтому шестизначный код нельзя подобрать перебором.
func (s *AuthService) VerifyEmail(ctx context.Context, req domain.VerifyEmailRequest) error {
	const op = "AuthService.VerifyEmail"

	if req.Email == "" || req.VerificationCode == "" {
		return fmt.Errorf("%s: %w", op, domain.ErrInvalidToken)
	}
	if s.blocked(ctx, s.verifyLimiter, req.Email) {
		return fmt.Errorf("%s: %w", op, domain.ErrTooManyAttempts)
	}

	valid, err := s.emailVerifyStore.Verify(ctx, req.Email, req.VerificationCode)
	if err != nil {
		return fmt.Errorf("%s: verify code: %w", op, err)
	}
	if !valid {
		s.registerFailure(ctx, s.verifyLimiter, req.Email)
		return fmt.Errorf("%s: %w", op, domain.ErrInvalidToken)
	}

	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return fmt.Errorf("%s: get user: %w", op, err)
	}

	// Активируем аккаунт. Токены не выдаются: пользователь входит обычным способом.
	user.EmailVerified = true
	user.Status = domain.StatusActive
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, *user); err != nil {
		return fmt.Errorf("%s: update user: %w", op, err)
	}

	if s.verifyLimiter != nil {
		_ = s.verifyLimiter.Reset(ctx, req.Email)
	}

	s.log.Info("email verified, user activated", slog.String("user_id", user.ID))
	return nil
}

// ResendVerificationEmail отправляет повторное письмо верификации.
func (s *AuthService) ResendVerificationEmail(ctx context.Context, req domain.ResendVerificationRequest) error {
	const op = "AuthService.ResendVerificationEmail"

	// Проверяем что пользователь существует и не верифицирован.
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		// Не раскрываем وجود пользователя — возвращаем успех.
		s.log.Debug("resend verification: user not found", slog.String("email", req.Email))
		return nil //nolint:nilerr
	}

	if user.EmailVerified {
		return nil //nolint:nilerr
	}

	// Генерируем новый код.
	code, err := generateVerificationCode()
	if err != nil {
		return fmt.Errorf("%s: generate code: %w", op, err)
	}

	if err := s.emailVerifyStore.Store(ctx, req.Email, code); err != nil {
		return fmt.Errorf("%s: store code: %w", op, err)
	}

	if err := s.emailSender.SendVerificationEmail(ctx, req.Email, code); err != nil {
		s.log.Warn("failed to resend verification email",
			slog.String("email", req.Email),
			slog.String("error", err.Error()),
		)
	}

	return nil
}

// RequestPasswordReset запрашивает сброс пароля.
func (s *AuthService) RequestPasswordReset(ctx context.Context, req domain.PasswordResetRequest) error {
	const op = "AuthService.RequestPasswordReset"

	// Проверяем существование пользователя.
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		// Не раскрываем وجود — всегда возвращаем успех.
		return nil //nolint:nilerr
	}

	// Генерируем токен сброса.
	resetToken := generateResetToken()

	// Сохраняем в Valkey.
	if err := s.passwordResetStore.Store(ctx, req.Email, resetToken); err != nil {
		return fmt.Errorf("%s: store reset token: %w", op, err)
	}

	// Отправляем письмо.
	if err := s.emailSender.SendPasswordResetEmail(ctx, req.Email, resetToken); err != nil {
		s.log.Warn("failed to send password reset email",
			slog.String("email", req.Email),
			slog.String("error", err.Error()),
		)
	}

	_ = user
	return nil
}

// ResetPassword сбрасывает пароль по токену.
func (s *AuthService) ResetPassword(ctx context.Context, req domain.ResetPasswordRequest) error {
	const op = "AuthService.ResetPassword"

	// Извлекаем email из токена.
	email, err := s.passwordResetStore.Consume(ctx, req.ResetToken)
	if err != nil {
		return fmt.Errorf("%s: consume reset token: %w", op, err)
	}

	// Находим пользователя.
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("%s: get user: %w", op, err)
	}

	// Хешируем новый пароль.
	passwordHash, err := s.passwordHasher.Hash(ctx, req.NewPassword)
	if err != nil {
		return fmt.Errorf("%s: hash password: %w", op, err)
	}

	user.PasswordHash = passwordHash
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, *user); err != nil {
		return fmt.Errorf("%s: update user: %w", op, err)
	}

	// Отзываем все сессии пользователя (безопасность).
	_, _ = s.sessionRepo.RevokeAllForUser(ctx, user.ID, "")

	s.log.Info("password reset", slog.String("user_id", user.ID))
	return nil
}

// createSession создаёт новую сессию и генерирует токены.
func (s *AuthService) createSession(ctx context.Context, userID, userAgent, ipAddress string) (domain.TokenPair, *domain.Session, error) {
	const op = "AuthService.createSession"

	// Генерируем refresh token.
	refreshToken, err := s.tokenManager.GenerateRefreshToken(ctx)
	if err != nil {
		return domain.TokenPair{}, nil, fmt.Errorf("%s: generate refresh token: %w", op, err)
	}

	rtHash := cryptoprov.HashToken(refreshToken)

	now := time.Now()
	session := &domain.Session{
		UserID:           userID,
		UserAgent:        userAgent,
		IPAddress:        ipAddress,
		RefreshTokenHash: rtHash,
		CreatedAt:        now,
		LastUsedAt:       now,
		ExpiresAt:        now.Add(s.refreshTokenTTL),
	}

	// Сохраняем сессию.
	if err := s.sessionRepo.Create(ctx, *session); err != nil {
		return domain.TokenPair{}, nil, fmt.Errorf("%s: create session: %w", op, err)
	}

	// Получаем сессию с ID из БД.
	savedSession, err := s.sessionRepo.GetByRefreshTokenHash(ctx, rtHash)
	if err != nil {
		return domain.TokenPair{}, nil, fmt.Errorf("%s: get saved session: %w", op, err)
	}

	// Генерируем access token.
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return domain.TokenPair{}, nil, fmt.Errorf("%s: get user for claims: %w", op, err)
	}

	accessToken, expiresAt, err := s.tokenManager.GenerateAccessToken(ctx, domain.Claims{
		UserID:    savedSession.UserID,
		SessionID: savedSession.ID,
		Email:     user.Email,
		Role:      user.Role,
	})
	if err != nil {
		return domain.TokenPair{}, nil, fmt.Errorf("%s: generate access token: %w", op, err)
	}

	tokens := domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		TokenType:    "Bearer",
	}

	return tokens, savedSession, nil
}

// rotateSession инвалидирует текущую сессию и создаёт новую.
func (s *AuthService) rotateSession(ctx context.Context, oldSession *domain.Session) (domain.TokenPair, *domain.Session, error) {
	return s.createSession(ctx, oldSession.UserID, oldSession.UserAgent, oldSession.IPAddress)
}

func generateVerificationCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", fmt.Errorf("generateVerificationCode: %w", err)
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func generateResetToken() string {
	bytes := make([]byte, 32)
	_, _ = rand.Read(bytes)
	return fmt.Sprintf("%x", bytes)
}
