package domain

import "context"

// BuildStorage управляет сохранением и извлечением данных о серверных билдах.
type BuildStorage interface {
	// Create регистрирует новый билд в хранилище. Возвращает ErrAlreadyExists при дубликате версии.
	Create(ctx context.Context, build *ServerBuild) error

	// GetByID возвращает билд по идентификатору. Возвращает ErrNotFound при отсутствии.
	GetByID(ctx context.Context, id int64) (*ServerBuild, error)

	// GetByVersion возвращает билд по game_id и версии. Возвращает ErrNotFound при отсутствии.
	GetByVersion(ctx context.Context, gameID int64, version string) (*ServerBuild, error)

	// ListByGame возвращает последние версии билдов для игры, отсортированные по дате создания (новые первыми).
	// Limit ограничивает количество результатов (0 — без ограничения).
	ListByGame(ctx context.Context, gameID int64, limit int) ([]*ServerBuild, error)

	// CountByGame возвращает количество билдов для игры.
	CountByGame(ctx context.Context, gameID int64) (int, error)

	// Delete удаляет билд из хранилища. Возвращает ErrNotFound при отсутствии.
	Delete(ctx context.Context, id int64) error

	// CountActiveInstancesByBuild возвращает количество активных инстансов, использующих билд.
	// Активные — это инстансы в статусах starting, running, stopping.
	CountActiveInstancesByBuild(ctx context.Context, buildID int64) (int, error)
}
