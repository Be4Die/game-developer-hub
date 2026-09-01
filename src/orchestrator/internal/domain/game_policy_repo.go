package domain

import "context"

// GamePolicyRepo хранит политики оркестрации серверов по проектам.
type GamePolicyRepo interface {
	// Get возвращает политику игры. Возвращает ErrNotFound при отсутствии.
	Get(ctx context.Context, gameID int64) (*GamePolicy, error)

	// Set создаёт или обновляет политику игры.
	Set(ctx context.Context, policy *GamePolicy) error

	// Delete удаляет политику игры. Возвращает ErrNotFound при отсутствии.
	Delete(ctx context.Context, gameID int64) error

	// ListAll возвращает все сохранённые политики.
	ListAll(ctx context.Context) ([]*GamePolicy, error)
}
