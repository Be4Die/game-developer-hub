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

export const getProject = (id) => http.get(`/projects/${id}`).then((r) => r.data.project);

export const updateProject = (id, payload) =>
  http.patch(`/projects/${id}`, payload).then((r) => r.data.project);

export const deleteProject = (id) => http.delete(`/projects/${id}`).then((r) => r.data);

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

export const unpublish = (id) => http.post(`/projects/${id}/unpublish`).then((r) => r.data);

export const PERMISSIONS = {
  EDIT_INFO: 'PERM_EDIT_INFO',
  UPLOAD_BUILD: 'PERM_UPLOAD_BUILD',
  UPLOAD_MEDIA: 'PERM_UPLOAD_MEDIA',
  VIEW_STATS: 'PERM_VIEW_STATS',
  MANAGE_SERVERS: 'PERM_MANAGE_SERVERS',
  SUBMIT_MODERATION: 'PERM_SUBMIT_MODERATION',
};

export const ALL_PERMISSIONS = Object.values(PERMISSIONS);

export const sendInvitation = (projectId, { invitee_id, invitee_email, permissions }) =>
  http
    .post(`/projects/${projectId}/invitations`, {
      invitee_id,
      invitee_email,
      permissions,
    })
    .then((r) => r.data.invitation);

export const listIncomingInvitations = () =>
  http.get('/invitations/incoming').then((r) => r.data.invitations || []);

export const listOutgoingInvitations = (projectId = null) => {
  const query = projectId ? `?project_id=${projectId}` : '';
  return http.get(`/invitations/outgoing${query}`).then((r) => r.data.invitations || []);
};

export const respondInvitation = (invitationId, accept) =>
  http.post(`/invitations/${invitationId}/respond`, { accept }).then((r) => r.data);

export const cancelInvitation = (invitationId) =>
  http.post(`/invitations/${invitationId}/cancel`).then((r) => r.data);

export const listMembers = (projectId) =>
  http.get(`/projects/${projectId}/members`).then((r) => r.data.members || []);

export const updateMemberPermissions = (projectId, userId, permissions) =>
  http
    .patch(`/projects/${projectId}/members/${userId}`, { permissions })
    .then((r) => r.data.member);

export const removeMember = (projectId, userId) =>
  http.delete(`/projects/${projectId}/members/${userId}`).then((r) => r.data);

export const leaveProject = (projectId) =>
  http.post(`/projects/${projectId}/leave`).then((r) => r.data);

export const listSharedProjects = () =>
  http.get('/projects/shared').then((r) => r.data.projects || r.data.shared_projects || []);

export const blockUser = (blockedUserId) =>
  http.post('/access/blocks', { blocked_user_id: blockedUserId }).then((r) => r.data.block);

export const unblockUser = (blockedUserId) =>
  http.delete(`/access/blocks/${blockedUserId}`).then((r) => r.data);

export const listBlockedUsers = () =>
  http.get('/access/blocks').then((r) => r.data.blocks || []);
