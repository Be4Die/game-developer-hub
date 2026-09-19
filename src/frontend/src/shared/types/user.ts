export type UserRole =
  | 'USER_ROLE_UNSPECIFIED'
  | 'USER_ROLE_DEVELOPER'
  | 'USER_ROLE_MODERATOR'
  | 'USER_ROLE_ADMIN'
  | 'USER_ROLE_USER'
  | 'admin'
  | 'moderator'
  | 'developer'
  | 'user'
  | number;

export type UserStatus =
  | 'USER_STATUS_UNSPECIFIED'
  | 'USER_STATUS_ACTIVE'
  | 'USER_STATUS_SUSPENDED'
  | 'USER_STATUS_DELETED'
  | 'USER_STATUS_BANNED'
  | 'active'
  | 'suspended'
  | 'banned'
  | 'deleted'
  | number;

export interface User {
  id: string;
  email: string;
  display_name?: string;
  username?: string;
  role: UserRole;
  status: UserStatus;
  email_verified?: boolean;
  created_at?: string;
  updated_at?: string;
}

export interface AuthTokens {
  access_token: string;
  refresh_token: string;
  expires_at?: string;
  token_type?: string;
}

export interface LoginResponse {
  tokens: AuthTokens;
  user: User;
}

export interface RegisterResponse {
  tokens?: AuthTokens;
  user: User;
}

export interface RefreshResponse {
  tokens: AuthTokens;
  user?: User;
}

export interface AuthState {
  user: User | null;
  accessToken: string | null;
  refreshToken: string | null;
  loading: boolean;
  error: string | null;
}
