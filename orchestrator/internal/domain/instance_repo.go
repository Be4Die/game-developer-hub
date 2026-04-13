package domain

import "context"

// InstanceRepo хранит персистентные метаданные инстансов (PostgreSQL).
// Отвечает за историю: кто, когда, что запустил, на какой ноде, из какого билда.
type InstanceRepo interface {
	// Create регистрирует новый инстанс.
	Create(ctx context.Context, instance *Instance) error

	// GetByID возвращает инстанс по идентификатору. Возвращает ErrNotFound при отсутствии.
	GetByID(ctx context.Context, id int64) (*Instance, error)

	// ListByGame возвращает все инстансы указанной игры.
	// Опционально фильтрует по статусу (nil — без фильтра).
	ListByGame(ctx context.Context, gameID int64, status *InstanceStatus) ([]*Instance, error)

	// ListByNode возвращает все инстансы, запущенные на указанной ноде.
	ListByNode(ctx context.Context, nodeID int64) ([]*Instance, error)

	// Update обновляет персистентные поля инстанса. Возвращает ErrNotFound при отсутствии.
	Update(ctx context.Context, instance *Instance) error

	// Delete удаляет запись инстанса из БД. Возвращает ErrNotFound при отсутствии.
	Delete(ctx context.Context, id int64) error

	// CountByGame возвращает количество записей инстансов для игры.
	CountByGame(ctx context.Context, gameID int64) (int, error)
}
