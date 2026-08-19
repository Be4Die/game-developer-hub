export const nodeStatusMap = {
  NODE_STATUS_UNSPECIFIED: 'offline',
  NODE_STATUS_UNAUTHORIZED: 'unauthorized',
  NODE_STATUS_ONLINE: 'online',
  NODE_STATUS_OFFLINE: 'offline',
  NODE_STATUS_MAINTENANCE: 'maintenance',
};

export function normalizeNodeStatus(status) {
  if (!status) return 'offline';
  if (nodeStatusMap[status]) return nodeStatusMap[status];
  return status;
}

export function normalizeNode(raw) {
  if (!raw) return raw;
  return {
    ...raw,
    status: normalizeNodeStatus(raw.status),
  };
}
