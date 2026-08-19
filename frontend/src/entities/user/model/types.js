export function roleClass(role) {
  switch (role) {
    case 'USER_ROLE_ADMIN':
    case 'admin':
      return 'badge-danger';
    case 'USER_ROLE_MODERATOR':
    case 'moderator':
      return 'badge-warning';
    default:
      return 'badge-success';
  }
}

export function roleLabel(role) {
  switch (role) {
    case 'USER_ROLE_ADMIN':
    case 'admin':
      return 'Администратор';
    case 'USER_ROLE_MODERATOR':
    case 'moderator':
      return 'Модератор';
    default:
      return 'Разработчик';
  }
}

export function statusBadgeClass(status) {
  switch (status) {
    case 'USER_STATUS_ACTIVE':
    case 'active':
      return 'badge-success';
    default:
      return 'badge-danger';
  }
}

export function statusLabel(status) {
  switch (status) {
    case 'USER_STATUS_ACTIVE':
    case 'active':
      return 'Активен';
    case 'USER_STATUS_BANNED':
    case 'USER_STATUS_SUSPENDED':
    case 'banned':
    case 'suspended':
      return 'Заблокирован';
    default:
      return 'Неизвестно';
  }
}
