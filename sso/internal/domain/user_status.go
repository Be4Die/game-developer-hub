package domain

// UserStatus определяет состояние учётной записи пользователя.
type UserStatus uint8

const (
	// StatusActive — учётная запись активна.
	StatusActive UserStatus = iota + 1
	// StatusSuspended — учётная запись заблокирована.
	StatusSuspended
	// StatusDeleted — учётная запись удалена.
	StatusDeleted
)

func (s UserStatus) String() string {
	switch s {
	case StatusActive:
		return "active"
	case StatusSuspended:
		return "suspended"
	case StatusDeleted:
		return "deleted"
	default:
		return "unknown"
	}
}

// ParseUserStatus преобразует строковое представление статуса в UserStatus.
func ParseUserStatus(s string) UserStatus {
	switch s {
	case "active":
		return StatusActive
	case "suspended":
		return StatusSuspended
	case "deleted":
		return StatusDeleted
	default:
		return 0
	}
}
