package domain

import "context"

// NodeRepo хранит персистентные данные о нодах (PostgreSQL).
// Отвечает за реестр: адреса, токены, региона, характеристики.
type NodeRepo interface {
	// Create добавляет новую ноду в реестр. Возвращает ErrAlreadyExists при повторной регистрации.
	Create(ctx context.Context, node *Node) error

	// Update обновляет данные ноды. Возвращает ErrNotFound если нода не существует.
	Update(ctx context.Context, node *Node) error

	// GetByID возвращает ноду по идентификатору. Возвращает ErrNotFound при отсутствии.
	GetByID(ctx context.Context, id int64) (*Node, error)

	// GetByAddress возвращает ноду по gRPC-адресу. Возвращает ErrNotFound при отсутствии.
	GetByAddress(ctx context.Context, address string) (*Node, error)

	// List возвращает все ноды. Опционально фильтрует по статусу (nil — без фильтра).
	List(ctx context.Context, status *NodeStatus) ([]*Node, error)

	// Delete удаляет ноду из реестра. Возвращает ErrNotFound при отсутствии.
	Delete(ctx context.Context, id int64) error

	// UpdateLastPing обновляет время последнего heartbeat ноды.
	UpdateLastPing(ctx context.Context, id int64) error
}
