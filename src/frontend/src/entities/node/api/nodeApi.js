import { http } from '@/shared/api';
import {
  normalizeNode,
  normalizeService,
  nodeRoleToProto,
  serviceTypeToProto,
  ingressModeToProto,
} from '../model/nodeModel';

export function listNodes(status, gameId = null) {
  const params = {};
  if (status && status !== 'all') {
    const statusMap = {
      unauthorized: 'NODE_STATUS_UNAUTHORIZED',
      online: 'NODE_STATUS_ONLINE',
      offline: 'NODE_STATUS_OFFLINE',
      maintenance: 'NODE_STATUS_MAINTENANCE',
    };
    params.status = statusMap[status] || status;
  }
  if (gameId) {
    params.game_id = gameId;
  }
  return http.get('/nodes', { params }).then((r) => (r.data.nodes ?? []).map(normalizeNode));
}


export function getNode(nodeId) {
  return http.get(`/nodes/${nodeId}`).then((r) => normalizeNode(r.data.node || r.data));
}

export function registerNode(payload) {
  let requestBody;
  if (payload.node_id !== undefined) {
    // Authorize mode (auto-discovery)
    requestBody = {
      authorize: {
        node_id: payload.node_id,
        token: payload.token,
      },
    };
  } else {
    // Manual mode
    const manual = {
      address: payload.address,
      token: payload.token,
      region: payload.region,
    };
    if (payload.ingress_mode) {
      manual.ingress_mode = ingressModeToProto[payload.ingress_mode] || payload.ingress_mode;
    }
    if (payload.custom_domain) {
      manual.custom_domain = payload.custom_domain;
    }
    requestBody = {
      manual,
    };
  }
  return http.post('/nodes', requestBody).then((r) => normalizeNode(r.data));
}

export function updateNodeIngress(nodeId, ingressMode, customDomain = '', skipDnsCheck = false) {
  const protoMode = ingressModeToProto[ingressMode] || ingressMode;
  return http
    .patch(`/nodes/${nodeId}/ingress`, {
      ingress_mode: protoMode,
      custom_domain: customDomain,
      skip_dns_check: skipDnsCheck,
    })
    .then((r) => normalizeNode(r.data.node || r.data));
}

export function verifyNodeDomain(nodeId, customDomain) {
  return http
    .post(`/nodes/${nodeId}/verify-domain`, {
      custom_domain: customDomain,
    })
    .then((r) => r.data);
}

export function deleteNode(nodeId) {
  return http.delete(`/nodes/${nodeId}`);
}

export function getNodeUsage(nodeId) {
  return http.get(`/nodes/${nodeId}/usage`).then((r) => {
    const data = r.data;
    return {
      cpu_usage_percent: data.usage?.cpu_usage_percent ?? 0,
      memory_used_bytes: data.usage?.memory_used_bytes ?? 0,
      disk_used_bytes: data.usage?.disk_used_bytes ?? 0,
      network_bytes_per_sec: data.usage?.network_bytes_per_sec ?? 0,
      active_instance_count: data.active_instance_count ?? 0,
    };
  });
}

export function listNodeInstances(nodeId) {
  return http.get(`/nodes/${nodeId}/instances`).then((r) => r.data.instances ?? []);
}

export const STORAGE_TRANSITION_STOP = 'STORAGE_TRANSITION_ACTION_STOP';
export const STORAGE_TRANSITION_DELETE = 'STORAGE_TRANSITION_ACTION_DELETE';
export const COMPUTE_TRANSITION_TERMINATE = 'COMPUTE_TRANSITION_ACTION_TERMINATE';

export function updateNodeRole(nodeId, role, options = {}) {
  const protoRole = nodeRoleToProto[role] || role;
  const payload = { role: protoRole };
  if (options.storage_action !== undefined) {
    payload.storage_action = options.storage_action;
  }
  if (options.compute_action !== undefined) {
    payload.compute_action = options.compute_action;
  }
  return http.patch(`/nodes/${nodeId}/role`, payload).then((r) => normalizeNode(r.data.node));
}

export function listNodeServices(nodeId, gameId = null) {
  const params = {};
  if (gameId) params.game_id = gameId;
  return http
    .get(`/nodes/${nodeId}/services`, { params })
    .then((r) => (r.data.services ?? []).map(normalizeService));
}

export function createNodeService(nodeId, payload) {
  const protoType = serviceTypeToProto[payload.service_type] || payload.service_type;
  const requestBody = {
    service_type: protoType,
    name: payload.name,
    password: payload.password || '',
    db_name: payload.db_name || '',
    port: payload.port ? Number(payload.port) : 0,
    allowed_game_ids: (payload.allowed_game_ids ?? []).map(Number),
  };
  return http
    .post(`/nodes/${nodeId}/services`, requestBody)
    .then((r) => normalizeService(r.data.service));
}

export function deleteNodeService(nodeId, serviceId, deleteVolume = false) {
  return http.delete(`/nodes/${nodeId}/services/${serviceId}`, {
    params: { delete_volume: deleteVolume },
  });
}

export function startManagedService(nodeId, serviceId) {
  return http
    .post(`/nodes/${nodeId}/services/${serviceId}/start`, {})
    .then((r) => normalizeService(r.data.service || r.data));
}

export function stopManagedService(nodeId, serviceId) {
  return http
    .post(`/nodes/${nodeId}/services/${serviceId}/stop`, {})
    .then((r) => normalizeService(r.data.service || r.data));
}

// ─── Managed Service Backups ───────────────────────────────────────────────

export function listServiceBackups(nodeId, serviceName) {
  return http
    .get(`/nodes/${nodeId}/services/${serviceName}/backups`)
    .then((r) => r.data.backups ?? []);
}

export function createServiceBackup(nodeId, serviceName) {
  return http
    .post(`/nodes/${nodeId}/services/${serviceName}/backups`, {})
    .then((r) => r.data.backup);
}

export function restoreServiceBackup(nodeId, serviceName, backupId) {
  return http
    .post(`/nodes/${nodeId}/services/${serviceName}/backups/${backupId}/restore`, {})
    .then((r) => r.data);
}

export function deleteServiceBackup(nodeId, serviceName, backupId) {
  return http.delete(`/nodes/${nodeId}/services/${serviceName}/backups/${backupId}`);
}

export function getServiceBackupTicket(nodeId, serviceName, backupId) {
  return http
    .post(`/nodes/${nodeId}/services/${serviceName}/backups/${backupId}/ticket`, {})
    .then((r) => r.data.ticket);
}

export async function downloadServiceBackup(nodeId, serviceName, backupId, fileName) {
  const ticket = await getServiceBackupTicket(nodeId, serviceName, backupId);
  const downloadUrl = `/api/v1/nodes/${nodeId}/services/${serviceName}/backups/${backupId}/download?ticket=${encodeURIComponent(ticket)}`;

  const link = document.createElement('a');
  link.href = downloadUrl;
  link.setAttribute('download', fileName || `${backupId}.archive`);
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
}

export function uploadServiceBackup(nodeId, serviceName, file, restoreImmediately = false, onProgress = null) {
  const formData = new FormData();
  formData.append('file', file);
  formData.append('restore_immediately', restoreImmediately ? 'true' : 'false');

  return http
    .post(`/nodes/${nodeId}/services/${serviceName}/backups/upload`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
      onUploadProgress: (progressEvent) => {
        if (onProgress && progressEvent.total) {
          const percent = Math.round((progressEvent.loaded * 100) / progressEvent.total);
          onProgress(percent);
        }
      },
    })
    .then((r) => r.data.backup);
}



export function toggleNodeBackups(nodeId, enabled) {
  return http
    .put(`/nodes/${nodeId}/backups/toggle`, { enabled })
    .then((r) => normalizeNode(r.data.node || r.data));
}

export function toggleServiceAutoBackup(nodeId, serviceName, enabled) {
  return http
    .put(`/nodes/${nodeId}/services/${serviceName}/auto-backup/toggle`, { enabled })
    .then((r) => normalizeService(r.data.service || r.data));
}

// ─── Platform Nodes & Access Requests ─────────────────────────────────────

export function updateNodePlatform(nodeId, isPlatform) {
  return http
    .patch(`/nodes/${nodeId}/platform`, { is_platform: isPlatform })
    .then((r) => normalizeNode(r.data.node || r.data));
}

export function getPlatformAccess(gameId) {
  return http
    .get(`/projects/${gameId}/platform-access`)
    .then((r) => r.data.request || null)
    .catch((err) => {
      if (err.response && (err.response.status === 404 || err.response.status === 400)) {
        return null;
      }
      throw err;
    });
}

export function createPlatformAccess(gameId, reason) {
  return http
    .post(`/projects/${gameId}/platform-access`, { reason })
    .then((r) => r.data.request);
}

export function listPlatformAccessRequests(status = null) {
  const params = {};
  if (status && status !== 'all') {
    const statusToProto = {
      pending: 'PLATFORM_ACCESS_STATUS_PENDING',
      approved: 'PLATFORM_ACCESS_STATUS_APPROVED',
      rejected: 'PLATFORM_ACCESS_STATUS_REJECTED',
    };
    params.status = statusToProto[status] || status;
  }
  return http
    .get('/platform-access/requests', { params })
    .then((r) => r.data.requests ?? []);
}

export function reviewPlatformAccess(requestId, { approved, rejection_reason = '', max_instances = 5, status, reviewer_comment }) {
  const finalStatus = status || (approved ? 'PLATFORM_ACCESS_STATUS_APPROVED' : 'PLATFORM_ACCESS_STATUS_REJECTED');
  const finalComment = reviewer_comment ?? rejection_reason ?? '';
  return http
    .post(`/platform-access/requests/${requestId}/review`, {
      status: finalStatus,
      reviewer_comment: finalComment,
      max_instances: Number(max_instances),
    })
    .then((r) => r.data.request);
}

