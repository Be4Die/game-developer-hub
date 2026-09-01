package domain

import "errors"

var (
	// ErrNotFound возвращается при отсутствии запрашиваемого ресурса.
	ErrNotFound = errors.New("not found")
	// ErrAlreadyExists возвращается при попытке создать дублирующийся ресурс.
	ErrAlreadyExists = errors.New("already exists")
	// ErrInvalidPassword возвращается при неверном пароле.
	ErrInvalidPassword = errors.New("invalid password")
	// ErrInvalidToken возвращается при невалидном токене.
	ErrInvalidToken = errors.New("invalid token")
	// ErrTokenExpired возвращается при истёкшем сроке действия токена.
	ErrTokenExpired = errors.New("token expired")
	// ErrEmailNotVerified возвращается при попытке действий без подтверждения email.
	ErrEmailNotVerified = errors.New("email not verified")
	// ErrUserSuspended возвращается при попытке входа заблокированного пользователя.
	ErrUserSuspended = errors.New("user suspended")
	// ErrCannotDeleteAdmin возвращается при попытке удалить администратора.
	ErrCannotDeleteAdmin = errors.New("cannot delete admin user")
	// ErrCannotModifyAdminStatus возвращается при попытке изменить статус администратора.
	ErrCannotModifyAdminStatus = errors.New("cannot modify admin status")
	// ErrCannotModifyModeratorStatus возвращается при попытке изменить статус модератора.
	ErrCannotModifyModeratorStatus = errors.New("cannot modify moderator status")
	// ErrPermissionDenied возвращается при отсутствии прав на выполнение операции.
	ErrPermissionDenied = errors.New("permission denied")
	// ErrNotInternalUser возвращается при попытке создать внутреннего пользователя с внешним email.
	ErrNotInternalUser = errors.New("only internal users can be created with welwise.com domain")
	// ErrModeratorManagedByAdmin возвращается при попытке модератора или администратора изменить свои данные или пароль.
	ErrModeratorManagedByAdmin = errors.New("moderator profile is managed by administrator")
	// ErrProfileImmutable возвращается при попытке изменить защищённый профиль (модератора или администратора).
	ErrProfileImmutable = errors.New("profile cannot be modified for this role")
)
