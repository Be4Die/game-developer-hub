export type NodeStatus =
  | 'online'
  | 'offline'
  | 'degraded'
  | 'NODE_STATUS_UNSPECIFIED'
  | 'NODE_STATUS_ONLINE'
  | 'NODE_STATUS_OFFLINE'
  | 'NODE_STATUS_DEGRADED'
  | string;

export type NodeRole =
  | 'mixed'
  | 'game_only'
  | 'service_only'
  | 'NODE_ROLE_UNSPECIFIED'
  | 'NODE_ROLE_MIXED'
  | 'NODE_ROLE_GAME_ONLY'
  | 'NODE_ROLE_SERVICE_ONLY'
  | string;

export type ServiceType =
  | 'postgres'
  | 'redis'
  | 'mysql'
  | 'volume'
  | 'adminer'
  | 'pgadmin'
  | 'SERVICE_TYPE_POSTGRES'
  | 'SERVICE_TYPE_REDIS'
  | 'SERVICE_TYPE_MYSQL'
  | 'SERVICE_TYPE_VOLUME'
  | 'SERVICE_TYPE_ADMINER'
  | 'SERVICE_TYPE_PGADMIN'
  | string;

export type ServiceStatus =
  | 'unknown'
  | 'starting'
  | 'running'
  | 'stopped'
  | 'failed'
  | string;

export interface NodeService {
  id: string;
  node_id?: string;
  name?: string;
  service_type: ServiceType;
  status: ServiceStatus;
  host_port?: number;
  container_id?: string;
  credentials?: Record<string, string>;
  allowed_game_ids?: (string | number)[];
  volume_path?: string;
  volume_size_bytes?: number;
  connection_uri?: string;
  auto_backup_enabled?: boolean;
  backup_schedule?: string;
  backup_retention_days?: number;
  created_at?: string;
  updated_at?: string;
  [key: string]: any;
}

export interface NodeStats {
  cpu_percent?: number;
  memory_bytes?: number;
  memory_limit?: number;
  memory_percent?: number;
  disk_used_bytes?: number;
  disk_total_bytes?: number;
  instances_count?: number;
  services_count?: number;
  timestamp?: string;
  [key: string]: any;
}

export interface Node {
  id: string;
  name?: string;
  owner_id?: string;
  ip_address?: string;
  port?: number;
  status: NodeStatus;
  role: NodeRole;
  region?: string;
  labels?: Record<string, string>;
  services?: NodeService[];
  stats?: NodeStats;
  ingress_mode?: 'platform_proxy' | 'direct' | string;
  custom_domain?: string;
  backups_enabled?: boolean;
  is_platform?: boolean;
  platform_access_status?: 'pending' | 'approved' | 'rejected' | string;
  created_at?: string;
  updated_at?: string;
  last_heartbeat_at?: string;
  [key: string]: any;
}

export type NodeInfo = Node;

export interface ServiceBackup {
  id: string;
  service_id: string;
  node_id: string;
  name?: string;
  size_bytes?: number;
  status?: string;
  created_at?: string;
  [key: string]: any;
}
