package domain

import "context"

// InstanceStorage хранит и предоставляет данные об игровых инстансах.
type InstanceStorage interface {
	// GetInstanceByID возвращает инстанс по уникальному ID. Возвращает ErrNotFound при отсутствии.
	GetInstanceByID(ctx context.Context, id int64) (*Instance, error)
	// GetInstancesByGameID возвращает все инстансы для указанной игры.
	GetInstancesByGameID(ctx context.Context, gameID int64) ([]Instance, error)
	// GetInstanceByContainerID возвращает инстанс по ID контейнера. Возвращает ErrNotFound при отсутствии.
	GetInstanceByContainerID(ctx context.Context, containerID string) (*Instance, error)
	// GetAllInstances возвращает все зарегистрированные инстансы.
	GetAllInstances(ctx context.Context) ([]Instance, error)

	// RecordInstance сохраняет или обновляет данные инстанса.
	RecordInstance(ctx context.Context, instance Instance) error

	// DeleteInstance удаляет инстанс из хранилища.
	DeleteInstance(ctx context.Context, id int64) error
}
