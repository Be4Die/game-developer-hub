package domain

import "context"

// PasswordHasher — интерфейс для хеширования паролей.
type PasswordHasher interface {
	Hash(ctx context.Context, password string) ([]byte, error)
	Compare(ctx context.Context, hash []byte, password string) error
}
