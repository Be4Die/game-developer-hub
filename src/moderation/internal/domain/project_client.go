package domain

import "context"

// ProjectClient определяет контракт взаимодействия с сервисом управления проектами (project-manager).
type ProjectClient interface {
	// PublishRelease разворачивает сборку игры в продуктивное окружение и фиксирует релиз.
	PublishRelease(ctx context.Context, projectID int64, version, approvedBy, comment string) (string, error)
	// RejectDraft переводит черновик проекта в статус для доработки разработчиком.
	RejectDraft(ctx context.Context, projectID int64, reason string) error
}
