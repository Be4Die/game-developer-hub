package domain

import "errors"

var (
	// ErrNotFound возвращается при отсутствии запрашиваемого ресурса.
	ErrNotFound = errors.New("not found")
	// ErrAlreadyExists возвращается при попытке создать дубликат ресурса.
	ErrAlreadyExists = errors.New("already exists")
	// ErrBuildInUse возвращается при попытке удалить билд, используемый активными инстансами.
	ErrBuildInUse = errors.New("build is in use by active instances")
	// ErrNodeUnauthorized возвращается при попытке выполнить операцию с неавторизованной нодой.
	ErrNodeUnauthorized = errors.New("node is not authorized")
	// ErrInvalidToken возвращается при неверном токене авторизации ноды.
	ErrInvalidToken = errors.New("invalid node token")
	// ErrNoAvailableNode возвращается когда нет нод с достаточными ресурсами.
	ErrNoAvailableNode = errors.New("no available node with sufficient resources")
	// ErrForbidden возвращается при попытке выполнить операцию над чужим ресурсом.
	ErrForbidden = errors.New("forbidden")
)
