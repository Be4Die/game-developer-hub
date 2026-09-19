import type { Instance, InstanceStatus } from '@/shared/types';

export const instanceStatusMap: Record<string, string> = {
  INSTANCE_STATUS_UNSPECIFIED: 'stopped',
  INSTANCE_STATUS_STARTING: 'starting',
  INSTANCE_STATUS_RUNNING: 'running',
  INSTANCE_STATUS_STOPPING: 'stopping',
  INSTANCE_STATUS_STOPPED: 'stopped',
  INSTANCE_STATUS_CRASHED: 'crashed',
};

export function normalizeInstanceStatus(status?: string | null): InstanceStatus {
  if (!status) return 'stopped';
  if (instanceStatusMap[status]) return instanceStatusMap[status];
  return status;
}

export function normalizeInstance(raw?: any): Instance {
  if (!raw) return raw;
  return {
    ...raw,
    status: normalizeInstanceStatus(raw.status),
  };
}
