package domain

import "context"

type ModerationRepository interface {
	Create(ctx context.Context, m *GameModeration) error
	GetByGameID(ctx context.Context, gameID int64) (*GameModeration, error)
	GetPending(ctx context.Context, limit, offset int) ([]GameModeration, int, error)
	Update(ctx context.Context, m *GameModeration) error
}
