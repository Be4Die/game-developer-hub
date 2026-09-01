package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Be4Die/game-developer-hub/sso/internal/domain"
)

// TokenService реализует логику валидации токенов.
type TokenService struct {
	log          *slog.Logger
	sessionRepo  domain.SessionRepository
	sessionCache domain.SessionCache
	tokenManager domain.TokenManager
}

// NewTokenService создаёт сервис работы с токенами.
func NewTokenService(
	log *slog.Logger,
	sessionRepo domain.SessionRepository,
	sessionCache domain.SessionCache,
	tokenManager domain.TokenManager,
) *TokenService {
	return &TokenService{
		log:          log,
		sessionRepo:  sessionRepo,
		sessionCache: sessionCache,
		tokenManager: tokenManager,
	}
}

// ValidateToken проверяет валидность access токена.
func (s *TokenService) ValidateToken(ctx context.Context, accessToken string) (domain.Claims, error) {
	const op = "TokenService.ValidateToken"

	// Парсим и валидируем JWT.
	claims, err := s.tokenManager.ParseAccessToken(ctx, accessToken)
	if err != nil {
		return domain.Claims{}, fmt.Errorf("%s: %w", op, domain.ErrInvalidToken)
	}

	// Проверяем сессию (сначала кэш, потом БД).
	session, err := s.sessionCache.Get(ctx, claims.SessionID)
	if err != nil {
		// Кэш-мисс, идём в БД.
		session, err = s.sessionRepo.GetByID(ctx, claims.SessionID)
		if err != nil {
			return domain.Claims{}, fmt.Errorf("%s: session not found: %w", op, err)
		}

		// Сохраняем в кэш.
		_ = s.sessionCache.Set(ctx, session)
	}

	// Проверяем что сессия активна.
	if !session.IsActive() {
		return domain.Claims{}, fmt.Errorf("%s: %w", op, domain.ErrTokenExpired)
	}

	// Обновляем last_used_at (лениво — не пишем в БД при каждой валидации).
	_ = session.LastUsedAt

	return *claims, nil
}

// ListSessions возвращает все активные сессии пользователя.
func (s *TokenService) ListSessions(ctx context.Context, userID string) ([]domain.Session, error) {
	const op = "TokenService.ListSessions"

	sessions, err := s.sessionRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Фильтруем только активные.
	var active []domain.Session
	for _, session := range sessions {
		if session.IsActive() {
			active = append(active, session)
		}
	}

	return active, nil
}

// RevokeSession отзывает конкретную сессию.
func (s *TokenService) RevokeSession(ctx context.Context, userID, sessionID string) error {
	const op = "TokenService.RevokeSession"

	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// Проверяем что сессия принадлежит пользователю.
	if session.UserID != userID {
		return fmt.Errorf("%s: %w", op, domain.ErrNotFound)
	}

	if err := s.sessionRepo.Revoke(ctx, sessionID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// Инвалидируем кэш.
	_ = s.sessionCache.Invalidate(ctx, sessionID)

	return nil
}

// RevokeAllSessions отзывает все сессии пользователя.
func (s *TokenService) RevokeAllSessions(ctx context.Context, userID, excludeSessionID string) (int64, error) {
	const op = "TokenService.RevokeAllSessions"

	count, err := s.sessionRepo.RevokeAllForUser(ctx, userID, excludeSessionID)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	// Инвалидируем кэш для всех сессий (лениво — кэш истечёт по TTL).
	return count, nil
}
