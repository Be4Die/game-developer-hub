package domain

import "time"

// Session хранит информацию о пользовательской сессии.
type Session struct {
	ID               string
	UserID           string
	UserAgent        string
	IPAddress        string
	RefreshTokenHash string
	CreatedAt        time.Time
	LastUsedAt       time.Time
	ExpiresAt        time.Time
	Revoked          bool
	RevokedAt        time.Time
}

// IsActive возвращает true, если сессия активна (не отозвана и не истекла).
func (s *Session) IsActive() bool {
	return !s.Revoked && time.Now().Before(s.ExpiresAt)
}
