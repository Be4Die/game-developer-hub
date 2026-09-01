package domain

import "time"

// ProjectStatus определяет статус жизненного цикла игрового проекта.
type ProjectStatus int16

const (
	// ProjectStatusDraft обозначает статус редактирования черновика.
	ProjectStatusDraft ProjectStatus = 1
	// ProjectStatusPending обозначает статус нахождения на модерации.
	ProjectStatusPending ProjectStatus = 2
	// ProjectStatusPublished обозначает статус опубликованной игры.
	ProjectStatusPublished ProjectStatus = 3
	// ProjectStatusRejected обозначает статус отклоненной модератором заявки.
	ProjectStatusRejected ProjectStatus = 4
)

// Project представляет агрегат игрового проекта на платформе.
// Содержит базовую идентификационную информацию и ссылки на черновик и релиз.
type Project struct {
	ID        int64
	OwnerID   string
	Status    ProjectStatus
	CreatedAt time.Time
	UpdatedAt time.Time
	Draft     *Draft
	Release   *Release
}
