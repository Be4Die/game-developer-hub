import { http } from '@/shared/api';

// ─── Auth ──────────────────────────────────────────────

export function register({ email, password, display_name }) {
  return http.post('/auth/register', { email, password, display_name }).then((r) => r.data);
}

export function login({ email, password }) {
  return http.post('/auth/login', { email, password }).then((r) => r.data);
}

export function refreshToken(refresh_token) {
  return http.post('/auth/refresh', { refresh_token }).then((r) => r.data);
}

export function logout(refresh_token) {
  return http.post('/auth/logout', { refresh_token }).then((r) => r.data);
}

export function verifyEmail(verification_code) {
  return http.post('/auth/verify-email', { verification_code }).then((r) => r.data);
}

export function resendVerificationEmail(email) {
  return http.post('/auth/resend-verification', { email }).then((r) => r.data);
}

export function requestPasswordReset(email) {
  return http.post('/auth/password-reset', { email }).then((r) => r.data);
}

export function resetPassword({ reset_token, new_password }) {
  return http.post('/auth/reset-password', { reset_token, new_password }).then((r) => r.data);
}

// ─── Users ─────────────────────────────────────────────

export function getUser(userId) {
  return http.get(`/users/${userId}`).then((r) => r.data);
}

export function searchUsers({ query = '', limit = 100, offset = 0 } = {}) {
  return http.get('/users', { params: { query, limit, offset } }).then((r) => r.data);
}

export function getCurrentUser() {
  return http.get('/user/profile').then((r) => r.data);
}

export function updateProfile({ display_name } = {}) {
  return http.patch('/user/profile', { display_name }).then((r) => r.data);
}

export function updateUser(userId, { display_name, avatar_url } = {}) {
  return http.patch('/user/profile', { display_name, avatar_url }).then((r) => r.data);
}

export function changePassword({ current_password, new_password, old_password }) {
  const current = current_password || old_password;
  return http
    .post('/user/profile:change-password', { current_password: current, new_password })
    .then((r) => r.data);
}

// ─── Tokens ────────────────────────────────────────────

export function revokeToken(session_id) {
  return http.delete(`/tokens/${session_id}`).then((r) => r.data);
}

export function listSessions() {
  return http.get('/tokens/sessions').then((r) => r.data);
}

// ─── Admin: Moderator Management ───────────────────────

export function createModerator({ login, password, display_name }) {
  return http.post('/users/moderators', { login, password, display_name }).then((r) => r.data);
}

export function deleteUser(userId) {
  return http.delete(`/users/${userId}`).then((r) => r.data);
}

export function getModerators() {
  return http.get('/users', { params: { limit: 100, offset: 0 } }).then((r) => {
    const users = r.data.users || [];
    const mods = users.filter(
      (u) => u.role === 'USER_ROLE_MODERATOR' || u.role === 'moderator' || u.role === 2
    );
    mods.sort((a, b) => {
      const aScore = (a.email || '').includes('moderator') ? 0 : 1;
      const bScore = (b.email || '').includes('moderator') ? 0 : 1;
      return aScore - bScore;
    });
    return mods;
  });
}
