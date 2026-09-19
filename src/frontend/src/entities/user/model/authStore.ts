import { reactive, readonly } from 'vue';
import type { User, AuthState, LoginResponse, RegisterResponse } from '@/shared/types';
import {
  login as ssoLogin,
  register as ssoRegister,
  logout as ssoLogout,
  refreshToken as ssoRefreshToken,
  getCurrentUser,
} from '../api/userApi';

const STORAGE_KEYS = {
  accessToken: 'gdh_access_token',
  refreshToken: 'gdh_refresh_token',
  user: 'gdh_user',
} as const;

const state = reactive<AuthState>({
  user: null,
  accessToken: localStorage.getItem(STORAGE_KEYS.accessToken) || null,
  refreshToken: localStorage.getItem(STORAGE_KEYS.refreshToken) || null,
  loading: false,
  error: null,
});

function setTokens(accessToken: string, refreshToken: string, user?: User | null): void {
  state.accessToken = accessToken;
  state.refreshToken = refreshToken;
  state.user = user ?? null;
  localStorage.setItem(STORAGE_KEYS.accessToken, accessToken);
  localStorage.setItem(STORAGE_KEYS.refreshToken, refreshToken);
  if (user) {
    localStorage.setItem(STORAGE_KEYS.user, JSON.stringify(user));
  }
}

function clearTokens(): void {
  state.accessToken = null;
  state.refreshToken = null;
  state.user = null;
  localStorage.removeItem(STORAGE_KEYS.accessToken);
  localStorage.removeItem(STORAGE_KEYS.refreshToken);
  localStorage.removeItem(STORAGE_KEYS.user);
}

export async function login(payload: { email: string; password: string }): Promise<LoginResponse> {
  state.loading = true;
  state.error = null;
  try {
    const res = await ssoLogin(payload);
    setTokens(res.tokens.access_token, res.tokens.refresh_token, res.user);
    return res;
  } catch (err: any) {
    state.error = err.response?.data?.message || 'Ошибка входа';
    throw err;
  } finally {
    state.loading = false;
  }
}

export async function register(payload: { email: string; password: string; display_name?: string }): Promise<RegisterResponse> {
  state.loading = true;
  state.error = null;
  try {
    const res = await ssoRegister(payload);
    state.user = res.user;
    return res;
  } catch (err: any) {
    state.error = err.response?.data?.message || 'Ошибка регистрации';
    throw err;
  } finally {
    state.loading = false;
  }
}

export async function logout(): Promise<void> {
  try {
    if (state.refreshToken) {
      await ssoLogout(state.refreshToken);
    }
  } finally {
    clearTokens();
    state.user = null;
  }
}

export async function refreshSession(): Promise<boolean> {
  if (!state.refreshToken) return false;
  try {
    const res = await ssoRefreshToken(state.refreshToken);
    setTokens(res.tokens.access_token, res.tokens.refresh_token, res.user);
    return true;
  } catch {
    clearTokens();
    state.user = null;
    return false;
  }
}

export async function loadUser(): Promise<void> {
  if (!state.accessToken) {
    const savedUser = localStorage.getItem(STORAGE_KEYS.user);
    if (savedUser) {
      try {
        state.user = JSON.parse(savedUser);
      } catch {
        // ignore
      }
    }
    return;
  }
  try {
    const res = await getCurrentUser();
    state.user = res.user;
    localStorage.setItem(STORAGE_KEYS.user, JSON.stringify(res.user));
  } catch {
    const refreshed = await refreshSession();
    if (refreshed && state.accessToken) {
      try {
        const res = await getCurrentUser();
        state.user = res.user;
        localStorage.setItem(STORAGE_KEYS.user, JSON.stringify(res.user));
      } catch {
        clearTokens();
      }
    }
  }
}

export function updateCurrentUser(updatedUser?: Partial<User> | null): void {
  if (!updatedUser || !state.user) return;
  state.user = {
    ...state.user,
    ...updatedUser,
  };
  localStorage.setItem(STORAGE_KEYS.user, JSON.stringify(state.user));
}

export function isAuthenticated(): boolean {
  return !!state.accessToken;
}

export function useAuth() {
  return {
    state: readonly(state),
    login,
    register,
    logout,
    refreshSession,
    loadUser,
    updateCurrentUser,
    isAuthenticated,
  };
}
