import { http } from '@/shared/api';
import { normalizeNode } from '../model/nodeModel';

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
  return http.get(`/nodes/${nodeId}`).then((r) => normalizeNode(r.data));
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
