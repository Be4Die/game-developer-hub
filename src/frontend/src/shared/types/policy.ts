export interface OrchestrationPolicy {
  game_id?: string | number;
  allocation_strategy?: 'least_loaded' | 'most_loaded' | 'round_robin' | string;
  min_ready_instances?: number;
  max_instances?: number;
  scale_up_threshold_cpu?: number;
  scale_up_threshold_memory?: number;
  scale_down_threshold_idle_seconds?: number;
  enabled?: boolean;
  [key: string]: any;
}

export interface QueueStatus {
  position?: number;
  estimated_wait_seconds?: number;
  state?: 'waiting' | 'assigned' | 'expired' | string;
  server?: {
    host?: string;
    port?: number;
    instance_id?: string;
    [key: string]: any;
  };
  [key: string]: any;
}
