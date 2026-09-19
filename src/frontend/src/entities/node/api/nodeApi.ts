import { http } from '@/shared/api';
import type { Node, NodeService, ServiceBackup } from '@/shared/types';
import {
  normalizeNode,
  normalizeService,
  nodeRoleToProto,
  serviceTypeToProto,
  ingressModeToProto,
} from '../model/nodeModel';

export function listNodes(status?: string | null, gameId: string | number | null = null): Promise<Node[]> {
  const params: Record<string, any> = {};
  if (status && status !== 'all') {
    const statusMap: Record<string, string> = {
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

export function getNode(nodeId: string | number): Promise<Node> {
  return http.get(`/nodes/${nodeId}`).then((r) => normalizeNode(r.data.node || r.data));
}

export function registerNode(payload: any): Promise<Node> {
  let requestBody: any;
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
    const manual: any = {
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

export function updateNodeIngress(
  nodeId: string | number,
  ingressMode: string,
  customDomain = '',
  skipDnsCheck = false
): Promise<Node> {
  const protoMode = ingressModeToProto[ingressMode] || ingressMode;
  return http
    .patch(`/nodes/${nodeId}/ingress`, {
      ingress_mode: protoMode,
      custom_domain: customDomain,
      skip_dns_check: skipDnsCheck,
    })
    .then((r) => normalizeNode(r.data.node || r.data));
}

export function verifyNodeDomain(nodeId: string | number, customDomain: string): Promise<any> {
  return http
    .post(`/nodes/${nodeId}/verify-domain`, {
      custom_domain: customDomain,
    })
    .then((r) => r.data);
}

export function deleteNode(nodeId: string | number): Promise<any> {
  return http.delete(`/nodes/${nodeId}`);
}

export function getNodeUsage(nodeId: string | number): Promise<{
  cpu_usage_percent: number;
  memory_used_bytes: number;
  disk_used_bytes: number;
  network_bytes_per_sec: number;
  active_instance_count: number;
}> {
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

export function listNodeInstances(nodeId: string | number): Promise<any[]> {
  return http.get(`/nodes/${nodeId}/instances`).then((r) => r.data.instances ?? []);
}

export const STORAGE_TRANSITION_STOP = 'STORAGE_TRANSITION_ACTION_STOP';
export const STORAGE_TRANSITION_DELETE = 'STORAGE_TRANSITION_ACTION_DELETE';
export const COMPUTE_TRANSITION_TERMINATE = 'COMPUTE_TRANSITION_ACTION_TERMINATE';

export function updateNodeRole(nodeId: string | number, role: string, options: any = {}): Promise<Node> {
  const protoRole = nodeRoleToProto[role] || role;
  const payload: any = { role: protoRole };
  if (options.storage_action !== undefined) {
    payload.storage_action = options.storage_action;
  }
  if (options.compute_action !== undefined) {
    payload.compute_action = options.compute_action;
  }
  return http.patch(`/nodes/${nodeId}/role`, payload).then((r) => normalizeNode(r.data.node));
}

export function listNodeServices(nodeId: string | number, gameId: string | number | null = null): Promise<NodeService[]> {
  const params: Record<string, any> = {};
  if (gameId) params.game_id = gameId;
  return http
    .get(`/nodes/${nodeId}/services`, { params })
    .then((r) => (r.data.services ?? []).map(normalizeService));
}

export function createNodeService(nodeId: string | number, payload: any): Promise<NodeService> {
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

export function deleteNodeService(nodeId: string | number, serviceId: string, deleteVolume = false): Promise<any> {
  return http.delete(`/nodes/${nodeId}/services/${serviceId}`, {
    params: { delete_volume: deleteVolume },
  });
}

export function startManagedService(nodeId: string | number, serviceId: string): Promise<NodeService> {
  return http
    .post(`/nodes/${nodeId}/services/${serviceId}/start`, {})
    .then((r) => normalizeService(r.data.service || r.data));
}

export function stopManagedService(nodeId: string | number, serviceId: string): Promise<NodeService> {
  return http
    .post(`/nodes/${nodeId}/services/${serviceId}/stop`, {})
    .then((r) => normalizeService(r.data.service || r.data));
}

// ─── Managed Service Backups ───────────────────────────────────────────────

export function listServiceBackups(nodeId: string | number, serviceName: string): Promise<ServiceBackup[]> {
  return http
    .get(`/nodes/${nodeId}/services/${serviceName}/backups`)
    .then((r) => r.data.backups ?? []);
}

export function createServiceBackup(nodeId: string | number, serviceName: string): Promise<ServiceBackup> {
  return http
    .post(`/nodes/${nodeId}/services/${serviceName}/backups`, {})
    .then((r) => r.data.backup);
}

export function restoreServiceBackup(nodeId: string | number, serviceName: string, backupId: string): Promise<any> {
  return http
    .post(`/nodes/${nodeId}/services/${serviceName}/backups/${backupId}/restore`, {})
    .then((r) => r.data);
}

export function deleteServiceBackup(nodeId: string | number, serviceName: string, backupId: string): Promise<any> {
  return http.delete(`/nodes/${nodeId}/services/${serviceName}/backups/${backupId}`);
}

export function getServiceBackupTicket(nodeId: string | number, serviceName: string, backupId: string): Promise<string> {
  return http
    .post(`/nodes/${nodeId}/services/${serviceName}/backups/${backupId}/ticket`, {})
    .then((r) => r.data.ticket);
}

export async function downloadServiceBackup(
  nodeId: string | number,
  serviceName: string,
  backupId: string,
  fileName?: string
): Promise<void> {
  const ticket = await getServiceBackupTicket(nodeId, serviceName, backupId);
  const downloadUrl = `/api/v1/nodes/${nodeId}/services/${serviceName}/backups/${backupId}/download?ticket=${encodeURIComponent(ticket)}`;

  const link = document.createElement('a');
  link.href = downloadUrl;
  link.setAttribute('download', fileName || `${backupId}.archive`);
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
}

export function uploadServiceBackup(
  nodeId: string | number,
  serviceName: string,
  file: File,
  restoreImmediately = false,
  onProgress: ((percent: number) => void) | null = null
): Promise<ServiceBackup> {
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

export function toggleNodeBackups(nodeId: string | number, enabled: boolean): Promise<Node> {
  return http
    .put(`/nodes/${nodeId}/backups/toggle`, { enabled })
    .then((r) => normalizeNode(r.data.node || r.data));
}

export function toggleServiceAutoBackup(nodeId: string | number, serviceName: string, enabled: boolean): Promise<NodeService> {
  return http
    .put(`/nodes/${nodeId}/services/${serviceName}/auto-backup/toggle`, { enabled })
    .then((r) => normalizeService(r.data.service || r.data));
}

// ─── Platform Nodes & Access Requests ─────────────────────────────────────

export function updateNodePlatform(nodeId: string | number, isPlatform: boolean): Promise<Node> {
  return http
    .patch(`/nodes/${nodeId}/platform`, { is_platform: isPlatform })
    .then((r) => normalizeNode(r.data.node || r.data));
}

export function getPlatformAccess(gameId: string | number): Promise<any> {
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

export function createPlatformAccess(gameId: string | number, reason: string): Promise<any> {
  return http
    .post(`/projects/${gameId}/platform-access`, { reason })
    .then((r) => r.data.request);
}

export function listPlatformAccessRequests(status: string | null = null): Promise<any[]> {
  const params: Record<string, string> = {};
  if (status && status !== 'all') {
    const statusToProto: Record<string, string> = {
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

export function reviewPlatformAccess(
  requestId: string | number,
  payload: { approved?: boolean; rejection_reason?: string; max_instances?: number | string; status?: string; reviewer_comment?: string }
): Promise<any> {
  const finalStatus = payload.status || (payload.approved ? 'PLATFORM_ACCESS_STATUS_APPROVED' : 'PLATFORM_ACCESS_STATUS_REJECTED');
  const finalComment = payload.reviewer_comment ?? payload.rejection_reason ?? '';
  return http
    .post(`/platform-access/requests/${requestId}/review`, {
      status: finalStatus,
      reviewer_comment: finalComment,
      max_instances: Number(payload.max_instances ?? 5),
    })
    .then((r) => r.data.request);
}
