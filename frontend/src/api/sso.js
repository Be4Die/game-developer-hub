import axios from "axios";

const http = axios.create({
  baseURL: "/api/v1",
  headers: { "Content-Type": "application/json" },
});

// Add JWT token to requests
http.interceptors.request.use((config) => {
  const token = localStorage.getItem("gdh_access_token");
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// ─── Auth ──────────────────────────────────────────────

export function register({ email, password, display_name }) {
  return http
    .post("/auth/register", { email, password, display_name })
    .then((r) => r.data);
}

export function login({ email, password }) {
  return http.post("/auth/login", { email, password }).then((r) => r.data);
}

export function refreshToken(refresh_token) {
  return http.post("/auth/refresh", { refresh_token }).then((r) => r.data);
}

export function logout(refresh_token) {
  return http.post("/auth/logout", { refresh_token }).then((r) => r.data);
}

export function verifyEmail(verification_code) {
  return http
    .post("/auth/verify-email", { verification_code })
    .then((r) => r.data);
}

export function resendVerificationEmail(email) {
  return http.post("/auth/resend-verification", { email }).then((r) => r.data);
}

export function requestPasswordReset(email) {
  return http.post("/auth/password-reset", { email }).then((r) => r.data);
}

export function resetPassword({ reset_token, new_password }) {
  return http
    .post("/auth/reset-password", { reset_token, new_password })
    .then((r) => r.data);
}

// ─── Users ─────────────────────────────────────────────

export function getUser(userId) {
  return http.get(`/users/${userId}`).then((r) => r.data);
}

export function getCurrentUser() {
  return http.get("/user/profile").then((r) => r.data);
}

export function updateUser(userId, { display_name, avatar_url } = {}) {
  return http
    .patch(`/users/${userId}`, { display_name, avatar_url })
    .then((r) => r.data);
}

export function changePassword({ old_password, new_password }) {
  return http
    .post("/users/change-password", { old_password, new_password })
    .then((r) => r.data);
}

// ─── Tokens ────────────────────────────────────────────

export function revokeToken(session_id) {
  return http.delete(`/tokens/${session_id}`).then((r) => r.data);
}

export function listSessions() {
  return http.get("/tokens/sessions").then((r) => r.data);
}
