import { http } from '@/shared/api';

// ─── Client Web Builds (Projects API) ───────────────────────────

export const listClientBuilds = (projectId) =>
  http.get(`/projects/${projectId}/builds`).then((r) => r.data.builds ?? []);

export const deleteClientBuild = (projectId, version) =>
  http.delete(`/projects/${projectId}/builds/${version}`).then((r) => r.data);

export const uploadClientBuild = (projectId, version, file, onProgress) => {
  const form = new FormData();
  form.append('version', version);
  form.append('file', file);
  return http
    .post(`/projects/${projectId}/builds`, form, {
      headers: { 'Content-Type': 'multipart/form-data' },
      onUploadProgress: (e) => onProgress?.(Math.round((e.loaded * 100) / e.total)),
    })
    .then((r) => r.data);
};

// ─── Server Game Builds (Orchestrator API) ───────────────────────

export function listServerBuilds(gameId) {
  return http.get(`/games/${gameId}/builds`).then((r) => r.data.builds ?? []);
}

export function getServerBuild(gameId, buildVersion) {
  return http
    .get(`/games/${gameId}/builds/${encodeURIComponent(buildVersion)}`)
    .then((r) => r.data);
}

export function uploadServerBuild(gameId, formData, onProgress) {
  return http
    .post(`/games/${gameId}/builds`, formData, {
      onUploadProgress: onProgress,
    })
    .then((r) => r.data);
}

export function deleteServerBuild(gameId, buildVersion) {
  return http.delete(`/games/${gameId}/builds/${encodeURIComponent(buildVersion)}`);
}
