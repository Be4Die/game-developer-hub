import { http } from '@/shared/api';
import {
  normalizeNode,
  normalizeService,
  nodeRoleToProto,
  serviceTypeToProto,
} from '../model/nodeModel';

export function listNodes(status) {
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
    requestBody = {
      manual: {
        address: payload.address,
        token: payload.token,
        region: payload.region,
      },
    };
  }
  return http.post('/nodes', requestBody).then((r) => normalizeNode(r.data));
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

export function updateNodeRole(nodeId, role) {
  const protoRole = nodeRoleToProto[role] || role;
  return http.patch(`/nodes/${nodeId}/role`, { role: protoRole }).then((r) => normalizeNode(r.data.node));
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
