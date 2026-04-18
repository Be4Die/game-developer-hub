package domain

import "time"

// TokenPair — пара access и refresh токенов.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	TokenType    string // "Bearer"
}

// Claims — данные, храные в JWT.
type Claims struct {
	UserID    string
	SessionID string
	Email     string
	Role      UserRole
}
