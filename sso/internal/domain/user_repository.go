package domain

import "context"

// UserRepository — интерфейс хранилища пользователей.
type UserRepository interface {
	Create(ctx context.Context, user User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user User) error
	Search(ctx context.Context, query string, limit, offset int) ([]User, int64, error)
	Delete(ctx context.Context, id string) error
}
