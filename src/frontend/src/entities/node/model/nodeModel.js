export const nodeStatusMap = {
  NODE_STATUS_UNSPECIFIED: 'offline',
  NODE_STATUS_UNAUTHORIZED: 'unauthorized',
  NODE_STATUS_ONLINE: 'online',
  NODE_STATUS_OFFLINE: 'offline',
  NODE_STATUS_MAINTENANCE: 'maintenance',
};

export const nodeRoleMap = {
  NODE_ROLE_UNSPECIFIED: 'mixed',
  NODE_ROLE_MIXED: 'mixed',
  NODE_ROLE_COMPUTE: 'compute',
  NODE_ROLE_STORAGE: 'storage',
};

export const nodeRoleToProto = {
  mixed: 'NODE_ROLE_MIXED',
  compute: 'NODE_ROLE_COMPUTE',
  storage: 'NODE_ROLE_STORAGE',
};

export const serviceTypeMap = {
  SERVICE_TYPE_UNSPECIFIED: 'postgres',
  SERVICE_TYPE_POSTGRES: 'postgres',
  SERVICE_TYPE_REDIS: 'redis',
  SERVICE_TYPE_MYSQL: 'mysql',
  SERVICE_TYPE_MINIO: 'minio',
  SERVICE_TYPE_VOLUME: 'volume',
  SERVICE_TYPE_ADMINER: 'adminer',
  SERVICE_TYPE_PGADMIN: 'pgadmin',
};

export const serviceTypeToProto = {
  postgres: 'SERVICE_TYPE_POSTGRES',
  redis: 'SERVICE_TYPE_REDIS',
  mysql: 'SERVICE_TYPE_MYSQL',
  minio: 'SERVICE_TYPE_MINIO',
  volume: 'SERVICE_TYPE_VOLUME',
  adminer: 'SERVICE_TYPE_ADMINER',
  pgadmin: 'SERVICE_TYPE_PGADMIN',
};

export const serviceStatusMap = {
  SERVICE_STATUS_UNSPECIFIED: 'unknown',
  SERVICE_STATUS_STARTING: 'starting',
  SERVICE_STATUS_RUNNING: 'running',
  SERVICE_STATUS_STOPPED: 'stopped',
  SERVICE_STATUS_FAILED: 'failed',
};

export function normalizeNodeStatus(status) {
  if (!status) return 'offline';
  if (nodeStatusMap[status]) return nodeStatusMap[status];
  return status;
}

export function normalizeNodeRole(role) {
  if (!role) return 'mixed';
  if (nodeRoleMap[role]) return nodeRoleMap[role];
  return String(role).toLowerCase();
}

export function normalizeService(raw) {
  if (!raw) return raw;
  return {
    ...raw,
    service_type: serviceTypeMap[raw.service_type] || (raw.service_type ? String(raw.service_type).toLowerCase() : 'postgres'),
    status: serviceStatusMap[raw.status] || (raw.status ? String(raw.status).toLowerCase() : 'unknown'),
    allowed_game_ids: raw.allowed_game_ids ?? [],
    credentials: raw.credentials ?? {},
    volume_size_bytes: Number(raw.volume_size_bytes ?? 0),
    host_port: Number(raw.host_port ?? 0),
    volume_path: raw.volume_path || '',
    connection_uri: raw.connection_uri || '',
  };
}

export function normalizeNode(raw) {
  if (!raw) return raw;
  return {
    ...raw,
    status: normalizeNodeStatus(raw.status),
    role: normalizeNodeRole(raw.role),
  };
}
