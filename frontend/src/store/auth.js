import { reactive, readonly } from "vue";
import {
  login as ssoLogin,
  register as ssoRegister,
  logout as ssoLogout,
  refreshToken as ssoRefreshToken,
  getCurrentUser,
} from "../api/sso";

const STORAGE_KEYS = {
  accessToken: "gdh_access_token",
  refreshToken: "gdh_refresh_token",
  user: "gdh_user",
};

const state = reactive({
  user: null,
  accessToken: localStorage.getItem(STORAGE_KEYS.accessToken) || null,
  refreshToken: localStorage.getItem(STORAGE_KEYS.refreshToken) || null,
  loading: false,
  error: null,
});

function setTokens(accessToken, refreshToken, user) {
  state.accessToken = accessToken;
  state.refreshToken = refreshToken;
  state.user = user;
  localStorage.setItem(STORAGE_KEYS.accessToken, accessToken);
  localStorage.setItem(STORAGE_KEYS.refreshToken, refreshToken);
  if (user) {
    localStorage.setItem(STORAGE_KEYS.user, JSON.stringify(user));
  }
}

function clearTokens() {
  state.accessToken = null;
  state.refreshToken = null;
  state.user = null;
  localStorage.removeItem(STORAGE_KEYS.accessToken);
  localStorage.removeItem(STORAGE_KEYS.refreshToken);
  localStorage.removeItem(STORAGE_KEYS.user);
}

export async function login({ email, password }) {
  state.loading = true;
  state.error = null;
  try {
    const res = await ssoLogin({ email, password });
    setTokens(res.tokens.access_token, res.tokens.refresh_token, res.user);
    return res;
  } catch (err) {
    state.error = err.response?.data?.message || "Ошибка входа";
    throw err;
  } finally {
    state.loading = false;
  }
}

export async function register({ email, password, display_name }) {
  state.loading = true;
  state.error = null;
  try {
    const res = await ssoRegister({ email, password, display_name });
    // После регистрации токены НЕ выдаются — нужна верификация email.
    // Сохраняем только информацию о пользователе.
    state.user = res.user;
    return res;
  } catch (err) {
    state.error = err.response?.data?.message || "Ошибка регистрации";
    throw err;
  } finally {
    state.loading = false;
  }
}

export async function logout() {
  try {
    if (state.refreshToken) {
      await ssoLogout(state.refreshToken);
    }
  } finally {
    clearTokens();
    state.user = null;
  }
}

export async function refreshSession() {
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

export async function loadUser() {
  if (!state.accessToken) {
    // Try to restore from localStorage
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

export function isAuthenticated() {
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
    isAuthenticated,
  };
}
