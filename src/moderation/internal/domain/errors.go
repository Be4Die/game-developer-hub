package domain

import "errors"

var (
	// ErrNotFound возвращается, когда запрашиваемый ресурс не найден.
	ErrNotFound = errors.New("not found")

	// ErrForbidden возвращается при отсутствии прав на выполнение операции.
	ErrForbidden = errors.New("forbidden")

	// ErrUnauthorized возвращается при отсутствии аутентификации.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrAlreadyClaimed возвращается при попытке взять в работу заявку, которая уже проверяется другим модератором.
	ErrAlreadyClaimed = errors.New("request already claimed by another moderator")

	// ErrInvalidStatus возвращается при попытке выполнить недопустимый переход статуса заявки.
	ErrInvalidStatus = errors.New("invalid status transition")

	// ErrEmptyReason возвращается при попытке отклонить заявку без указания причины.
	ErrEmptyReason = errors.New("rejection reason is required")

	// ErrEmptyMessage возвращается при попытке отправить пустое сообщение.
	ErrEmptyMessage = errors.New("message content cannot be empty")

	// ErrDeployFailed возвращается при сбое развёртывания в сервисе управления проектами.
	ErrDeployFailed = errors.New("project deployment failed")
)
