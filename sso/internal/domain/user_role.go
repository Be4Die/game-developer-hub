package domain

// UserRole определяет роль пользователя в системе.
type UserRole uint8

const (
	// RoleDeveloper — разработчик, стандартная роль пользователя.
	RoleDeveloper UserRole = iota + 1
	// RoleModerator — модератор, управляет контентом.
	RoleModerator
	// RoleAdmin — администратор, полный доступ.
	RoleAdmin
)

func (r UserRole) String() string {
	switch r {
	case RoleDeveloper:
		return "developer"
	case RoleModerator:
		return "moderator"
	case RoleAdmin:
		return "admin"
	default:
		return "unknown"
	}
}

// ParseUserRole преобразует строковое представление роли в UserRole.
func ParseUserRole(s string) UserRole {
	switch s {
	case "developer":
		return RoleDeveloper
	case "moderator":
		return RoleModerator
	case "admin":
		return RoleAdmin
	default:
		return 0
	}
}
