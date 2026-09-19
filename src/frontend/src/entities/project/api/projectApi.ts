import { http } from '@/shared/api';
import type {
  Project,
  ProjectRelease,
  ProjectInvitation,
  ProjectMember,
  BlockedUser,
  PaginationParams,
} from '@/shared/types';

export const listProjects = (params: PaginationParams = {}): Promise<{ projects: Project[]; total: number }> => {
  const query = new URLSearchParams();
  if (params.limit !== undefined) query.append('limit', String(params.limit));
  if (params.offset !== undefined) query.append('offset', String(params.offset));
  const qStr = query.toString() ? `?${query.toString()}` : '';
  return http.get(`/projects${qStr}`).then((r) => ({
    projects: r.data.projects ?? [],
    total: r.data.total ?? (r.data.projects ? r.data.projects.length : 0),
  }));
};

export const listPublishedProjects = (params: PaginationParams = {}): Promise<{ projects: Project[]; total: number }> => {
  const query = new URLSearchParams();
  if (params.limit !== undefined) query.append('limit', String(params.limit));
  if (params.offset !== undefined) query.append('offset', String(params.offset));
  const qStr = query.toString() ? `?${query.toString()}` : '';
  return http.get(`/projects/published${qStr}`).then((r) => ({
    projects: r.data.projects ?? [],
    total: r.data.total ?? (r.data.projects ? r.data.projects.length : 0),
  }));
};

export const createProject = (payload: any): Promise<Project> =>
  http.post('/projects', payload).then((r) => r.data.project);

export const getProject = (id: string | number): Promise<Project> =>
  http.get(`/projects/${id}`).then((r) => r.data.project);

export const updateProject = (id: string | number, payload: any): Promise<Project> =>
  http.patch(`/projects/${id}`, payload).then((r) => r.data.project);

export const deleteProject = (id: string | number): Promise<any> =>
  http.delete(`/projects/${id}`).then((r) => r.data);

export const uploadMedia = (id: string | number, mediaType: 'icon' | 'cover' | 'video' | string, file: File): Promise<any> => {
  const form = new FormData();
  form.append('media_type', mediaType);
  form.append('file', file);
  return http
    .post(`/projects/${id}/media`, form, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    .then((r) => r.data);
};

export const submitForModeration = (id: string | number): Promise<any> =>
  http.post(`/projects/${id}/moderation/submit`).then((r) => r.data);

export const getPublished = (id: string | number): Promise<ProjectRelease> =>
  http.get(`/projects/${id}/published`).then((r) => r.data.release);

export const unpublish = (id: string | number): Promise<any> =>
  http.post(`/projects/${id}/unpublish`).then((r) => r.data);

export const PERMISSIONS = {
  EDIT_INFO: 'PERM_EDIT_INFO',
  UPLOAD_BUILD: 'PERM_UPLOAD_BUILD',
  UPLOAD_MEDIA: 'PERM_UPLOAD_MEDIA',
  VIEW_STATS: 'PERM_VIEW_STATS',
  MANAGE_SERVERS: 'PERM_MANAGE_SERVERS',
  SUBMIT_MODERATION: 'PERM_SUBMIT_MODERATION',
} as const;

export const ALL_PERMISSIONS = Object.values(PERMISSIONS);

export const sendInvitation = (
  projectId: string | number,
  payload: { invitee_id?: string; invitee_email?: string; permissions?: string[] }
): Promise<ProjectInvitation> =>
  http
    .post(`/projects/${projectId}/invitations`, payload)
    .then((r) => r.data.invitation);

export const listIncomingInvitations = (): Promise<ProjectInvitation[]> =>
  http.get('/invitations/incoming').then((r) => r.data.invitations || []);

export const listOutgoingInvitations = (projectId: string | number | null = null): Promise<ProjectInvitation[]> => {
  const query = projectId ? `?project_id=${projectId}` : '';
  return http.get(`/invitations/outgoing${query}`).then((r) => r.data.invitations || []);
};

export const respondInvitation = (invitationId: string | number, accept: boolean): Promise<any> =>
  http.post(`/invitations/${invitationId}/respond`, { accept }).then((r) => r.data);

export const cancelInvitation = (invitationId: string | number): Promise<any> =>
  http.post(`/invitations/${invitationId}/cancel`).then((r) => r.data);

export const listMembers = (projectId: string | number): Promise<ProjectMember[]> =>
  http.get(`/projects/${projectId}/members`).then((r) => r.data.members || []);

export const updateMemberPermissions = (
  projectId: string | number,
  userId: string,
  permissions: string[]
): Promise<ProjectMember> =>
  http
    .patch(`/projects/${projectId}/members/${userId}`, { permissions })
    .then((r) => r.data.member);

export const removeMember = (projectId: string | number, userId: string): Promise<any> =>
  http.delete(`/projects/${projectId}/members/${userId}`).then((r) => r.data);

export const leaveProject = (projectId: string | number): Promise<any> =>
  http.post(`/projects/${projectId}/leave`).then((r) => r.data);

export const listSharedProjects = (): Promise<Project[]> =>
  http.get('/projects/shared').then((r) => r.data.projects || r.data.shared_projects || []);

export const blockUser = (blockedUserId: string | number): Promise<BlockedUser> =>
  http.post('/access/blocks', { blocked_user_id: String(blockedUserId) }).then((r) => r.data.block);

export const unblockUser = (blockedUserId: string | number): Promise<any> =>
  http.delete(`/access/blocks/${blockedUserId}`).then((r) => r.data);

export const listBlockedUsers = (): Promise<BlockedUser[]> =>
  http.get('/access/blocks').then((r) => r.data.blocks || []);
