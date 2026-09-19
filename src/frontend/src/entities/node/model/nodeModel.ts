import type { Node, NodeService, NodeStatus, NodeRole, ServiceType, ServiceStatus } from '@/shared/types';

export const nodeStatusMap: Record<string, NodeStatus> = {
  NODE_STATUS_UNSPECIFIED: 'offline',
  NODE_STATUS_UNAUTHORIZED: 'unauthorized',
  NODE_STATUS_ONLINE: 'online',
  NODE_STATUS_OFFLINE: 'offline',
  NODE_STATUS_MAINTENANCE: 'maintenance',
};

export const nodeRoleMap: Record<string, NodeRole> = {
  NODE_ROLE_UNSPECIFIED: 'mixed',
  NODE_ROLE_MIXED: 'mixed',
  NODE_ROLE_COMPUTE: 'compute',
  NODE_ROLE_STORAGE: 'storage',
};

export const nodeRoleToProto: Record<string, string> = {
  mixed: 'NODE_ROLE_MIXED',
  compute: 'NODE_ROLE_COMPUTE',
  storage: 'NODE_ROLE_STORAGE',
};

export const serviceTypeMap: Record<string, ServiceType> = {
  '0': 'postgres',
  '1': 'postgres',
  '2': 'redis',
  '3': 'mysql',
  '5': 'volume',
  '6': 'adminer',
  '7': 'pgadmin',
  SERVICE_TYPE_UNSPECIFIED: 'postgres',
  SERVICE_TYPE_POSTGRES: 'postgres',
  SERVICE_TYPE_REDIS: 'redis',
  SERVICE_TYPE_MYSQL: 'mysql',
  SERVICE_TYPE_VOLUME: 'volume',
  SERVICE_TYPE_ADMINER: 'adminer',
  SERVICE_TYPE_PGADMIN: 'pgadmin',
};

export const serviceTypeToProto: Record<string, string> = {
  postgres: 'SERVICE_TYPE_POSTGRES',
  redis: 'SERVICE_TYPE_REDIS',
  mysql: 'SERVICE_TYPE_MYSQL',
  volume: 'SERVICE_TYPE_VOLUME',
  adminer: 'SERVICE_TYPE_ADMINER',
  pgadmin: 'SERVICE_TYPE_PGADMIN',
};

export const serviceStatusMap: Record<string, ServiceStatus> = {
  '0': 'unknown',
  '1': 'starting',
  '2': 'running',
  '3': 'stopped',
  '4': 'failed',
  SERVICE_STATUS_UNSPECIFIED: 'unknown',
  SERVICE_STATUS_STARTING: 'starting',
  SERVICE_STATUS_RUNNING: 'running',
  SERVICE_STATUS_STOPPED: 'stopped',
  SERVICE_STATUS_FAILED: 'failed',
};

export function normalizeNodeStatus(status?: string | number | null): NodeStatus {
  if (!status) return 'offline';
  const str = String(status);
  if (nodeStatusMap[str]) return nodeStatusMap[str];
  return str as NodeStatus;
}

export function normalizeNodeRole(role?: string | number | null): NodeRole {
  if (!role) return 'mixed';
  const str = String(role);
  if (nodeRoleMap[str]) return nodeRoleMap[str];
  return str.toLowerCase() as NodeRole;
}

export function normalizeService(raw?: any): NodeService {
  if (!raw) return raw;
  const data = raw.service ? { ...raw.service } : { ...raw };
  const rawType = data.type ?? data.service_type;
  const typeKey = rawType !== undefined && rawType !== null ? String(rawType) : '';
  const statusKey = data.status !== undefined && data.status !== null ? String(data.status) : '';

  return {
    ...data,
    service_type: serviceTypeMap[typeKey] || (rawType ? String(rawType).toLowerCase() : 'postgres'),
    status: serviceStatusMap[statusKey] || (data.status ? String(data.status).toLowerCase() : 'unknown'),
    allowed_game_ids: data.allowed_game_ids ?? [],
    credentials: data.credentials ?? {},
    volume_size_bytes: Number(data.volume_size_bytes ?? 0),
    host_port: Number(data.host_port ?? data.port ?? 0),
    volume_path: data.volume_path || '',
    connection_uri: data.connection_uri || '',
    auto_backup_enabled: !!data.auto_backup_enabled,
  };
}

export const ingressModeMap: Record<string, 'platform_proxy' | 'direct'> = {
  '0': 'platform_proxy',
  '1': 'platform_proxy',
  '2': 'direct',
  INGRESS_MODE_UNSPECIFIED: 'platform_proxy',
  INGRESS_MODE_PLATFORM_PROXY: 'platform_proxy',
  INGRESS_MODE_DIRECT: 'direct',
};

export const ingressModeToProto: Record<string, string> = {
  platform_proxy: 'INGRESS_MODE_PLATFORM_PROXY',
  direct: 'INGRESS_MODE_DIRECT',
};

export function normalizeIngressMode(mode?: string | number | null): 'platform_proxy' | 'direct' {
  if (!mode) return 'platform_proxy';
  const str = String(mode);
  if (ingressModeMap[str]) return ingressModeMap[str];
  return 'platform_proxy';
}

export const platformAccessStatusMap: Record<string, string> = {
  PLATFORM_ACCESS_STATUS_UNSPECIFIED: 'pending',
  PLATFORM_ACCESS_STATUS_PENDING: 'pending',
  PLATFORM_ACCESS_STATUS_APPROVED: 'approved',
  PLATFORM_ACCESS_STATUS_REJECTED: 'rejected',
  '0': 'pending',
  '1': 'pending',
  '2': 'approved',
  '3': 'rejected',
  pending: 'pending',
  approved: 'approved',
  rejected: 'rejected',
};

export function normalizePlatformAccessStatus(status?: string | number | null): string {
  if (!status) return 'pending';
  const str = String(status);
  return platformAccessStatusMap[str] || str.toLowerCase();
}

export function normalizeNode(raw?: any): Node {
  if (!raw) return raw;
  const data = raw.node ? { ...raw.node } : { ...raw };
  return {
    ...data,
    status: normalizeNodeStatus(data.status),
    role: normalizeNodeRole(data.role),
    ingress_mode: normalizeIngressMode(data.ingress_mode),
    custom_domain: data.custom_domain || '',
    backups_enabled: !!data.backups_enabled,
    is_platform: Boolean(data.is_platform ?? raw.is_platform ?? false),
  };
}
