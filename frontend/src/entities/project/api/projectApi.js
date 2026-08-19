import { http } from '@/shared/api';

export const listProjects = (params = {}) => {
  const query = new URLSearchParams();
  if (params.limit !== undefined) query.append('limit', params.limit);
  if (params.offset !== undefined) query.append('offset', params.offset);
  const qStr = query.toString() ? `?${query.toString()}` : '';
  return http.get(`/projects${qStr}`).then((r) => ({
    projects: r.data.projects ?? [],
    total: r.data.total ?? (r.data.projects ? r.data.projects.length : 0),
  }));
};

export const createProject = (payload) =>
  http.post('/projects', payload).then((r) => r.data.project);

export const getProject = (id) =>
  http.get(`/projects/${id}`).then((r) => r.data.project);

export const updateProject = (id, payload) =>
  http.patch(`/projects/${id}`, payload).then((r) => r.data.project);

export const deleteProject = (id) =>
  http.delete(`/projects/${id}`).then((r) => r.data);

export const uploadMedia = (id, mediaType, file) => {
  const form = new FormData();
  form.append('media_type', mediaType); // icon, cover, video
  form.append('file', file);
  return http
    .post(`/projects/${id}/media`, form, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    .then((r) => r.data);
};

export const submitForModeration = (id) =>
  http.post(`/projects/${id}/moderation/submit`).then((r) => r.data);

export const getPublished = (id) =>
  http.get(`/projects/${id}/published`).then((r) => r.data.release);

export const unpublish = (id) =>
  http.post(`/projects/${id}/unpublish`).then((r) => r.data);
