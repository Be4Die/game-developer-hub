package domain

import "context"

// SessionRepository — интерфейс хранилища сессий.
type SessionRepository interface {
	Create(ctx context.Context, session Session) error
	GetByID(ctx context.Context, id string) (*Session, error)
	GetByUserID(ctx context.Context, userID string) ([]Session, error)
	GetByRefreshTokenHash(ctx context.Context, hash string) (*Session, error)
	Update(ctx context.Context, session Session) error
	Revoke(ctx context.Context, id string) error
	RevokeAllForUser(ctx context.Context, userID string, excludeSessionID string) (int64, error)
}
