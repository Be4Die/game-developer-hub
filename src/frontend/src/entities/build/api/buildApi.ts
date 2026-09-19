import { http } from '@/shared/api';
import type { ClientBuild, ServerBuild } from '@/shared/types';

// ─── Client Web Builds (Projects API) ───────────────────────────

export const listClientBuilds = (projectId: string | number): Promise<ClientBuild[]> =>
  http.get(`/projects/${projectId}/builds`).then((r) => r.data.builds ?? []);

export const deleteClientBuild = (projectId: string | number, version: string): Promise<any> =>
  http.delete(`/projects/${projectId}/builds/${version}`).then((r) => r.data);

export const uploadClientBuild = (
  projectId: string | number,
  version: string,
  file: File,
  onProgress?: (percent: number) => void
): Promise<any> => {
  const form = new FormData();
  form.append('version', version);
  form.append('file', file);
  return http
    .post(`/projects/${projectId}/builds`, form, {
      headers: { 'Content-Type': 'multipart/form-data' },
      onUploadProgress: (e) => {
        if (onProgress && e.total) {
          onProgress(Math.round((e.loaded * 100) / e.total));
        }
      },
    })
    .then((r) => r.data);
};

// ─── Server Game Builds (Orchestrator API) ───────────────────────

export function listServerBuilds(gameId: string | number): Promise<ServerBuild[]> {
  return http.get(`/games/${gameId}/builds`).then((r) => r.data.builds ?? []);
}

export function getServerBuild(gameId: string | number, buildVersion: string): Promise<ServerBuild> {
  return http
    .get(`/games/${gameId}/builds/${encodeURIComponent(buildVersion)}`)
    .then((r) => r.data);
}

export function uploadServerBuild(
  gameId: string | number,
  formData: FormData,
  onProgress?: (e: any) => void
): Promise<any> {
  return http
    .post(`/games/${gameId}/builds`, formData, {
      onUploadProgress: onProgress,
    })
    .then((r) => r.data);
}

export function deleteServerBuild(gameId: string | number, buildVersion: string): Promise<any> {
  return http.delete(`/games/${gameId}/builds/${encodeURIComponent(buildVersion)}`);
}
