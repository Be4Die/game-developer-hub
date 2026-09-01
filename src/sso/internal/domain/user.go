package domain

import "time"

// User представляет учётную запись пользователя платформы.
type User struct {
	ID            string
	Email         string
	PasswordHash  []byte
	DisplayName   string
	Role          UserRole
	Status        UserStatus
	EmailVerified bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
