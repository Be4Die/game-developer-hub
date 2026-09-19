import axios, {
  type InternalAxiosRequestConfig,
  type AxiosResponse,
  type AxiosError,
} from 'axios';

export const http = axios.create({
  baseURL: '/api/v1',
});

// Request interceptor: attach Bearer token and set appropriate Content-Type
http.interceptors.request.use((config: InternalAxiosRequestConfig) => {
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

let isRefreshing = false;
let failedQueue: Array<{
  resolve: (token: string | null) => void;
  reject: (error: any) => void;
}> = [];

const processQueue = (error: any, token: string | null = null) => {
  failedQueue.forEach((prom) => {
    if (error) {
      prom.reject(error);
    } else {
      prom.resolve(token);
    }
  });
  failedQueue = [];
};

interface CustomAxiosRequestConfig extends InternalAxiosRequestConfig {
  _retry?: boolean;
}

http.interceptors.response.use(
  (response: AxiosResponse) => response,
  (error: AxiosError) => {
    const originalRequest = error.config as CustomAxiosRequestConfig | undefined;
    if (
      error.response?.status === 401 &&
      originalRequest &&
      !originalRequest._retry &&
      originalRequest.url !== '/auth/refresh'
    ) {
      if (isRefreshing) {
        return new Promise<string | null>(function (resolve, reject) {
          failedQueue.push({ resolve, reject });
        })
          .then((token) => {
            if (token) {
              originalRequest.headers.Authorization = `Bearer ${token}`;
            }
            return http(originalRequest);
          })
          .catch((err) => Promise.reject(err));
      }

      originalRequest._retry = true;
      isRefreshing = true;

      const refreshToken = localStorage.getItem('gdh_refresh_token');
      if (!refreshToken) {
        return Promise.reject(error);
      }

      return new Promise((resolve, reject) => {
        axios
          .post('/api/v1/auth/refresh', { refresh_token: refreshToken })
          .then(({ data }) => {
            const newAccessToken = data.tokens.access_token;
            localStorage.setItem('gdh_access_token', newAccessToken);
            localStorage.setItem('gdh_refresh_token', data.tokens.refresh_token);

            // update original request header
            originalRequest.headers.Authorization = `Bearer ${newAccessToken}`;

            processQueue(null, newAccessToken);
            resolve(http(originalRequest));
          })
          .catch((err) => {
            processQueue(err, null);
            localStorage.removeItem('gdh_access_token');
            localStorage.removeItem('gdh_refresh_token');
            localStorage.removeItem('gdh_user');
            reject(err);
            window.location.href = '/login'; // Redirect to login on refresh failure
          })
          .finally(() => {
            isRefreshing = false;
          });
      });
    }
    return Promise.reject(error);
  }
);

export function getAuthHeaders(): Record<string, string> {
  const token = localStorage.getItem('gdh_access_token');
  const user = JSON.parse(localStorage.getItem('gdh_user') || 'null');
  const headers: Record<string, string> = token ? { Authorization: `Bearer ${token}` } : {};
  // Now that gateway ignores x-user-id from headers, the backend relies on JWT.
  // We still send these for compatibility just in case, but they are safely ignored.
  if (user?.id) {
    headers['x-user-id'] = user.id;
    headers['x-user-name'] = user.display_name || user.username || user.email || 'User';
    headers['x-user-role'] = user.role || 'USER_ROLE_USER';
  }
  return headers;
}
