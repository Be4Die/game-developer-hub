// Package domain содержит бизнес-модели и интерфейсы SSO-сервиса.
package domain

// RegisterRequest — данные для регистрации пользователя.
type RegisterRequest struct {
	Email       string
	Password    string
	DisplayName string
}

// RegisterResponse — результат регистрации.
type RegisterResponse struct {
	User   User
	Tokens TokenPair
}

// LoginRequest — данные для входа.
type LoginRequest struct {
	Email     string
	Password  string
	UserAgent string
	IPAddress string
}

// LoginResponse — результат входа.
type LoginResponse struct {
	User   User
	Tokens TokenPair
}

// RefreshTokenRequest — запрос на обновление токена.
type RefreshTokenRequest struct {
	RefreshToken string
}

// RefreshTokenResponse — обновлённая пара токенов.
type RefreshTokenResponse struct {
	Tokens TokenPair
}

// LogoutRequest — запрос на выход.
type LogoutRequest struct {
	RefreshToken string
}

// VerifyEmailRequest — запрос на подтверждение email.
type VerifyEmailRequest struct {
	VerificationCode string
}

// ResendVerificationRequest — повторная отправка письма верификации.
type ResendVerificationRequest struct {
	Email string
}

// PasswordResetRequest — запрос на сброс пароля.
type PasswordResetRequest struct {
	Email string
}

// ResetPasswordRequest — установка нового пароля.
type ResetPasswordRequest struct {
	ResetToken  string
	NewPassword string
}

// ChangePasswordRequest — смена пароля (требует текущий).
type ChangePasswordRequest struct {
	UserID          string
	CurrentPassword string
	NewPassword     string
}

// UpdateProfileRequest — обновление профиля.
type UpdateProfileRequest struct {
	UserID      string
	DisplayName *string
}

// SearchUsersRequest — поиск пользователей.
type SearchUsersRequest struct {
	Query  string
	Limit  int
	Offset int
}

// SearchUsersResponse — результат поиска.
type SearchUsersResponse struct {
	Users      []User
	TotalCount int64
}

// CreateModeratorRequest — запрос на создание модератора администратором.
type CreateModeratorRequest struct {
	Login       string
	Password    string
	DisplayName string
}

// CreateModeratorResponse — результат создания модератора.
type CreateModeratorResponse struct {
	User User
}

// DeleteUserRequest — запрос на удаление пользователя.
type DeleteUserRequest struct {
	UserID string
}
