// Package postgres предоставляет PostgreSQL-реализации репозиториев SSO.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Be4Die/game-developer-hub/sso/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SessionRepository — PostgreSQL-реализация domain.SessionRepository.
type SessionRepository struct {
	pool *pgxpool.Pool
}

// NewSessionRepository создаёт репозиторий сессий.
func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{pool: pool}
}

func (r *SessionRepository) Create(ctx context.Context, session domain.Session) error { //nolint:revive
	const op = "postgres.SessionRepository.Create"

	_, err := r.pool.Exec(ctx, `
		INSERT INTO sessions (id, user_id, user_agent, ip_address, refresh_token_hash, created_at, last_used_at, expires_at, revoked, revoked_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, session.ID, session.UserID, session.UserAgent, session.IPAddress, session.RefreshTokenHash, session.CreatedAt, session.LastUsedAt, session.ExpiresAt, session.Revoked, session.RevokedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("%s: %w", op, domain.ErrAlreadyExists)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *SessionRepository) GetByID(ctx context.Context, id string) (*domain.Session, error) { //nolint:revive
	const op = "postgres.SessionRepository.GetByID"

	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, user_agent, ip_address, refresh_token_hash, created_at, last_used_at, expires_at, revoked, revoked_at
		FROM sessions WHERE id = $1
	`, id)

	session, err := scanSession(row)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return session, nil
}

func (r *SessionRepository) GetByUserID(ctx context.Context, userID string) ([]domain.Session, error) { //nolint:revive
	const op = "postgres.SessionRepository.GetByUserID"

	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, user_agent, ip_address, refresh_token_hash, created_at, last_used_at, expires_at, revoked, revoked_at
		FROM sessions WHERE user_id = $1 ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var sessions []domain.Session
	for rows.Next() {
		s, err := scanSession(rows)
		if err != nil {
			return nil, fmt.Errorf("%s: scan: %w", op, err)
		}
		sessions = append(sessions, *s)
	}

	return sessions, nil
}

func (r *SessionRepository) Update(ctx context.Context, session domain.Session) error { //nolint:revive
	const op = "postgres.SessionRepository.Update"

	tag, err := r.pool.Exec(ctx, `
		UPDATE sessions SET
			user_agent = $2,
			ip_address = $3,
			refresh_token_hash = $4,
			last_used_at = $5,
			expires_at = $6,
			revoked = $7,
			revoked_at = $8
		WHERE id = $1
	`, session.ID, session.UserAgent, session.IPAddress, session.RefreshTokenHash, session.LastUsedAt, session.ExpiresAt, session.Revoked, session.RevokedAt)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", op, domain.ErrNotFound)
	}

	return nil
}

func (r *SessionRepository) GetByRefreshTokenHash(ctx context.Context, hash string) (*domain.Session, error) { //nolint:revive
	const op = "postgres.SessionRepository.GetByRefreshTokenHash"

	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, user_agent, ip_address, refresh_token_hash, created_at, last_used_at, expires_at, revoked, revoked_at
		FROM sessions WHERE refresh_token_hash = $1 AND revoked = false
	`, hash)

	session, err := scanSession(row)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return session, nil
}

func (r *SessionRepository) Revoke(ctx context.Context, id string) error { //nolint:revive
	const op = "postgres.SessionRepository.Revoke"

	tag, err := r.pool.Exec(ctx, `
		UPDATE sessions SET revoked = true, revoked_at = $2 WHERE id = $1
	`, id, time.Now())
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", op, domain.ErrNotFound)
	}

	return nil
}

func (r *SessionRepository) RevokeAllForUser(ctx context.Context, userID string, excludeSessionID string) (int64, error) { //nolint:revive
	const op = "postgres.SessionRepository.RevokeAllForUser"

	tag, err := r.pool.Exec(ctx, `
		UPDATE sessions SET revoked = true, revoked_at = $3
		WHERE user_id = $1 AND id != $2
	`, userID, excludeSessionID, time.Now())
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return tag.RowsAffected(), nil
}

type sessionScanner interface {
	Scan(dest ...any) error
}

func scanSession(s sessionScanner) (*domain.Session, error) {
	session := &domain.Session{}
	err := s.Scan(&session.ID, &session.UserID, &session.UserAgent, &session.IPAddress, &session.RefreshTokenHash, &session.CreatedAt, &session.LastUsedAt, &session.ExpiresAt, &session.Revoked, &session.RevokedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("postgres.scanSession: %w", err)
	}
	return session, nil
}
