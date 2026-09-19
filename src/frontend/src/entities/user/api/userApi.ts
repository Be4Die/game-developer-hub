import { http } from '@/shared/api';
import type { User, LoginResponse, RegisterResponse, RefreshResponse } from '@/shared/types';

// ─── Auth ──────────────────────────────────────────────

export function register(payload: { email: string; password: string; display_name?: string }): Promise<RegisterResponse> {
  return http.post('/auth/register', payload).then((r) => r.data);
}

export function login(payload: { email: string; password: string }): Promise<LoginResponse> {
  return http.post('/auth/login', payload).then((r) => r.data);
}

export function refreshToken(refresh_token: string): Promise<RefreshResponse> {
  return http.post('/auth/refresh', { refresh_token }).then((r) => r.data);
}

export function logout(refresh_token: string): Promise<any> {
  return http.post('/auth/logout', { refresh_token }).then((r) => r.data);
}

export function verifyEmail(verification_code: string): Promise<any> {
  return http.post('/auth/verify-email', { verification_code }).then((r) => r.data);
}

export function resendVerificationEmail(email: string): Promise<any> {
  return http.post('/auth/resend-verification', { email }).then((r) => r.data);
}

export function requestPasswordReset(email: string): Promise<any> {
  return http.post('/auth/password-reset', { email }).then((r) => r.data);
}

export function resetPassword(payload: { reset_token: string; new_password: string }): Promise<any> {
  return http.post('/auth/reset-password', payload).then((r) => r.data);
}

// ─── Users ─────────────────────────────────────────────

export function getUser(userId: string): Promise<{ user: User }> {
  return http.get(`/users/${userId}`).then((r) => r.data);
}

export function searchUsers(params: { query?: string; limit?: number; offset?: number } = {}): Promise<{ users: User[]; total_count?: number }> {
  return http.get('/users', { params }).then((r) => r.data);
}

export function getCurrentUser(): Promise<{ user: User }> {
  return http.get('/user/profile').then((r) => r.data);
}

export function updateProfile(payload: { display_name?: string } = {}): Promise<{ user: User }> {
  return http.patch('/user/profile', payload).then((r) => r.data);
}

export function updateUser(userId: string, payload: { display_name?: string; avatar_url?: string } = {}): Promise<{ user: User }> {
  return http.patch('/user/profile', payload).then((r) => r.data);
}

export function changePassword(payload: { current_password?: string; new_password?: string; old_password?: string }): Promise<any> {
  const current = payload.current_password || payload.old_password;
  return http
    .post('/user/profile:change-password', { current_password: current, new_password: payload.new_password })
    .then((r) => r.data);
}

// ─── Tokens ────────────────────────────────────────────

export function revokeToken(session_id: string): Promise<any> {
  return http.delete(`/tokens/${session_id}`).then((r) => r.data);
}

export function listSessions(): Promise<{ sessions: any[] }> {
  return http.get('/tokens/sessions').then((r) => r.data);
}

// ─── Admin: Moderator Management ───────────────────────

export function createModerator(payload: { login: string; password: string; display_name: string }): Promise<{ user: User }> {
  return http.post('/users/moderators', payload).then((r) => r.data);
}

export function deleteUser(userId: string): Promise<any> {
  return http.delete(`/users/${userId}`).then((r) => r.data);
}

export function setUserStatus(userId: string, newStatus: string | number): Promise<any> {
  return http.patch(`/users/${userId}:set-status`, { new_status: newStatus }).then((r) => r.data);
}

export function getModerators(): Promise<User[]> {
  return http.get('/users', { params: { limit: 100, offset: 0 } }).then((r) => {
    const users: User[] = r.data.users || [];
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
