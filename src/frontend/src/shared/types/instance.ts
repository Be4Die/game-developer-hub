export type InstanceStatus =
  | 'stopped'
  | 'starting'
  | 'running'
  | 'stopping'
  | 'crashed'
  | 'INSTANCE_STATUS_UNSPECIFIED'
  | 'INSTANCE_STATUS_STARTING'
  | 'INSTANCE_STATUS_RUNNING'
  | 'INSTANCE_STATUS_STOPPING'
  | 'INSTANCE_STATUS_STOPPED'
  | 'INSTANCE_STATUS_CRASHED'
  | string;

export interface InstanceUsage {
  cpu_percent?: number;
  memory_bytes?: number;
  memory_limit?: number;
  memory_percent?: number;
  network_rx_bytes?: number;
  network_tx_bytes?: number;
  timestamp?: string;
  [key: string]: any;
}

export interface Instance {
  id: string;
  game_id?: string | number;
  build_id?: string | number;
  node_id?: string;
  status: InstanceStatus;
  host?: string;
  port?: number;
  players_count?: number;
  max_players?: number;
  created_at?: string;
  started_at?: string;
  stopped_at?: string;
  region?: string;
  cpu_limit?: number;
  memory_limit?: number;
  environment?: Record<string, string>;
  labels?: Record<string, string>;
  usage?: InstanceUsage;
  [key: string]: any;
}

export interface LogEntry {
  timestamp?: string;
  message: string;
  source?: 'stdout' | 'stderr' | 'system' | string;
  level?: string;
  [key: string]: any;
}
