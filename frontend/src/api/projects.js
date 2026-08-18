import axios from "axios";

const http = axios.create({ baseURL: "/api/v1" });

http.interceptors.request.use((config) => {
  const token = localStorage.getItem("gdh_access_token");
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

export const listProjects = (params = {}) => {
  const query = new URLSearchParams();
  if (params.limit !== undefined) query.append("limit", params.limit);
  if (params.offset !== undefined) query.append("offset", params.offset);
  const qStr = query.toString() ? `?${query.toString()}` : "";
  return http.get(`/projects${qStr}`).then((r) => ({
    projects: r.data.projects ?? [],
    total: r.data.total ?? (r.data.projects ? r.data.projects.length : 0),
  }));
};

export const createProject = (payload) => http.post("/projects", payload).then((r) => r.data.project);

export const getProject = (id) => http.get(`/projects/${id}`).then((r) => r.data.project);

export const updateProject = (id, payload) =>
  http.patch(`/projects/${id}`, payload).then((r) => r.data.project);

export const deleteProject = (id) => http.delete(`/projects/${id}`).then((r) => r.data);

export const listBuilds = (id) =>
  http.get(`/projects/${id}/builds`).then((r) => r.data.builds ?? []);

export const deleteBuild = (id, version) =>
  http.delete(`/projects/${id}/builds/${version}`).then((r) => r.data);

// Multipart upload для билдов (через Gateway custom handler)
export const uploadBuild = (id, version, file, onProgress) => {
  const form = new FormData();
  form.append("version", version);
  form.append("file", file);
  return http.post(`/projects/${id}/builds`, form, {
    headers: { "Content-Type": "multipart/form-data" },
    onUploadProgress: (e) => onProgress?.(Math.round((e.loaded * 100) / e.total)),
  }).then((r) => r.data);
};

// Multipart upload для медиа (через Gateway custom handler)
export const uploadMedia = (id, mediaType, file) => {
  const form = new FormData();
  form.append("media_type", mediaType); // icon, cover, video
  form.append("file", file);
  return http.post(`/projects/${id}/media`, form, {
    headers: { "Content-Type": "multipart/form-data" },
  }).then((r) => r.data);
};

// Отправка проекта на модерацию
export const submitForModeration = (id) =>
  http.post(`/projects/${id}/moderation`).then((r) => r.data.ticket);

// Получение опубликованного релиза
export const getPublished = (id) =>
  http.get(`/projects/${id}/published`).then((r) => r.data.release);

// Снятие проекта с публикации
export const unpublish = (id) =>
  http.post(`/projects/${id}/unpublish`).then((r) => r.data);

