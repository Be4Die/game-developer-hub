import axios from 'axios';

export const http = axios.create({
  baseURL: '/api/v1',
});

// Request interceptor: attach Bearer token and set appropriate Content-Type
http.interceptors.request.use((config) => {
  const token = localStorage.getItem('gdh_access_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  // If data is provided and not FormData, default to application/json if not already set
  if (config.data && !(config.data instanceof FormData) && !config.headers['Content-Type']) {
    config.headers['Content-Type'] = 'application/json';
  }
  return config;
});

export function getAuthHeaders() {
  const token = localStorage.getItem('gdh_access_token');
  const user = JSON.parse(localStorage.getItem('gdh_user') || 'null');
  const headers = token ? { Authorization: `Bearer ${token}` } : {};
  if (user?.id) {
    headers['x-user-id'] = user.id;
    headers['x-user-name'] = user.display_name || user.username || user.email || 'User';
    headers['x-user-role'] = user.role || 'USER_ROLE_USER';
  }
  return headers;
}
