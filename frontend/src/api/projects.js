import axios from "axios";

const http = axios.create({ baseURL: "/api/v1" });

http.interceptors.request.use((config) => {
  const token = localStorage.getItem("gdh_access_token");
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

export const listProjects = () => http.get("/projects").then((r) => r.data.projects ?? []);

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
