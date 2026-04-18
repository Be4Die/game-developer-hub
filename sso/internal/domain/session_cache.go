package domain

import "context"

// SessionCache — кэш активных сессий.
type SessionCache interface {
	Set(ctx context.Context, session *Session) error
	Get(ctx context.Context, sessionID string) (*Session, error)
	Invalidate(ctx context.Context, sessionID string) error
}
