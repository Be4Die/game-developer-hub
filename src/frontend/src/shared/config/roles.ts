export const ROLE_MAP: Record<string, string> = {
  USER_ROLE_UNSPECIFIED: 'Разработчик',
  USER_ROLE_DEVELOPER: 'Разработчик',
  USER_ROLE_MODERATOR: 'Модератор',
  USER_ROLE_ADMIN: 'Администратор',
  developer: 'Разработчик',
  moderator: 'Модератор',
  admin: 'Администратор',
};

export const USER_ROLES = {
  UNSPECIFIED: 'USER_ROLE_UNSPECIFIED',
  DEVELOPER: 'USER_ROLE_DEVELOPER',
  MODERATOR: 'USER_ROLE_MODERATOR',
  ADMIN: 'USER_ROLE_ADMIN',
} as const;

export type UserRoleKey = keyof typeof USER_ROLES;
