// Package domain содержит доменные сущности, ошибки и интерфейсы подсистемы управления проектами.
package domain

import "errors"

// Ошибки доступа и валидации.
var (
	// ErrNotFound возвращается, когда запрашиваемый ресурс не найден в хранилище.
	ErrNotFound = errors.New("not found")

	// ErrAlreadyExists возвращается при попытке создать сущность с уже существующим уникальным ключом.
	ErrAlreadyExists = errors.New("already exists")

	// ErrForbidden возвращается при отсутствии прав доступа пользователя к запрашиваемому ресурсу.
	ErrForbidden = errors.New("forbidden")

	// ErrInvalidInput возвращается при передаче некорректных входных параметров.
	ErrInvalidInput = errors.New("invalid input")

	// ErrInvalidArchive возвращается при попытке загрузить поврежденный или небезопасный архив.
	ErrInvalidArchive = errors.New("invalid archive format or corrupt data")

	// ErrNoIndexHTML возвращается, когда в корне загруженного архива отсутствует файл index.html.
	ErrNoIndexHTML = errors.New("index.html is required in the root of the archive")

	// ErrDraftNotReady возвращается при попытке отправить на модерацию черновик без обязательных полей или билда.
	ErrDraftNotReady = errors.New("draft is missing required fields or active build")

	// ErrAlreadyInModeration возвращается при попытке повторно отправить черновик, уже находящийся на проверке.
	ErrAlreadyInModeration = errors.New("project is already pending moderation")

	// ErrNotApproved возвращается при попытке публикации игры, не прошедшей модерацию.
	ErrNotApproved = errors.New("project is not approved for publication")

	// ErrNoActiveBuild возвращается при попытке развертывания проекта без выбранной активной сборки.
	ErrNoActiveBuild = errors.New("no active build selected for project")

	// ErrDisallowedFileType возвращается при наличии недопустимых расширений файлов в архиве.
	ErrDisallowedFileType = errors.New("archive contains disallowed file type")

	// ErrLockBusy возвращается при попытке выполнения параллельной операции над заблокированным проектом.
	ErrLockBusy = errors.New("project is locked by another operation, please retry later")

	// ErrUserBlocked возвращается, когда получатель приглашения заблокировал отправителя.
	ErrUserBlocked = errors.New("user has blocked invitations from you")

	// ErrAlreadyMember возвращается, если пользователь уже является участником проекта.
	ErrAlreadyMember = errors.New("user is already a member of this project")

	// ErrAlreadyInvited возвращается, если открытое приглашение пользователю уже отправлено.
	ErrAlreadyInvited = errors.New("invitation already pending for this user")

	// ErrCannotInviteSelf возвращается при попытке отправить приглашение самому себе.
	ErrCannotInviteSelf = errors.New("cannot invite self to project")

	// ErrInvitationNotFound возвращается, если указанное приглашение не найдено.
	ErrInvitationNotFound = errors.New("invitation not found")

	// ErrInvitationClosed возвращается при попытке ответить на уже закрытое приглашение.
	ErrInvitationClosed = errors.New("invitation is already closed")
)
