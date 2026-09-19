import type { UserRole, UserStatus } from '@/shared/types';
export * from '@/shared/types/user';

export function roleClass(role?: UserRole | null): string {
  switch (role) {
    case 'USER_ROLE_ADMIN':
    case 'admin':
    case 3:
      return 'badge-danger';
    case 'USER_ROLE_MODERATOR':
    case 'moderator':
    case 2:
      return 'badge-warning';
    default:
      return 'badge-success';
  }
}

export function roleLabel(role?: UserRole | null): string {
  switch (role) {
    case 'USER_ROLE_ADMIN':
    case 'admin':
    case 3:
      return 'Администратор';
    case 'USER_ROLE_MODERATOR':
    case 'moderator':
    case 2:
      return 'Модератор';
    default:
      return 'Разработчик';
  }
}

export function statusBadgeClass(status?: UserStatus | null): string {
  switch (status) {
    case 'USER_STATUS_ACTIVE':
    case 'active':
    case 1:
      return 'badge-success';
    default:
      return 'badge-danger';
  }
}

export function statusLabel(status?: UserStatus | null): string {
  switch (status) {
    case 'USER_STATUS_ACTIVE':
    case 'active':
    case 1:
      return 'Активен';
    case 'USER_STATUS_BANNED':
    case 'USER_STATUS_SUSPENDED':
    case 'banned':
    case 'suspended':
    case 2:
      return 'Заблокирован';
    default:
      return 'Неизвестно';
  }
}
