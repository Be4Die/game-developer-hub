import { http } from '@/shared/api';
import type { Instance, InstanceUsage, LogEntry } from '@/shared/types';
import { normalizeInstance } from '../model/instanceModel';

export function listInstances(gameId: string | number, status?: string): Promise<Instance[]> {
  const params: Record<string, string> = {};
  if (status && status !== 'all') params.status = status;
  return http
    .get(`/games/${gameId}/instances`, { params })
    .then((r) => (r.data.instances ?? []).map(normalizeInstance));
}

export function getInstance(gameId: string | number, instanceId: string | number): Promise<Instance> {
  return http
    .get(`/games/${gameId}/instances/${instanceId}`)
    .then((r) => normalizeInstance(r.data));
}

export function startInstance(gameId: string | number, payload: any): Promise<Instance> {
  return http.post(`/games/${gameId}/instances`, payload).then((r) => normalizeInstance(r.data));
}

export function stopInstance(gameId: string | number, instanceId: string | number, timeout = 30): Promise<Instance> {
  return http
    .post(`/games/${gameId}/instances/${instanceId}:stop`, { timeout })
    .then((r) => normalizeInstance(r.data.instance ?? r.data));
}

export function deleteInstance(gameId: string | number, instanceId: string | number): Promise<any> {
  return http.delete(`/games/${gameId}/instances/${instanceId}`).then((r) => r.data);
}

export function restartInstance(gameId: string | number, instanceId: string | number): Promise<Instance> {
  return http
    .post(`/games/${gameId}/instances/${instanceId}:restart`)
    .then((r) => normalizeInstance(r.data.instance ?? r.data));
}

export function resumeInstance(gameId: string | number, instanceId: string | number): Promise<Instance> {
  return http
    .post(`/games/${gameId}/instances/${instanceId}:resume`)
    .then((r) => normalizeInstance(r.data.instance ?? r.data));
}

export function getInstanceUsage(gameId: string | number, instanceId: string | number): Promise<InstanceUsage> {
  return http
    .get(`/games/${gameId}/instances/${instanceId}/usage`)
    .then((r) => r.data.usage ?? r.data);
}

export function createLogStream(
  gameId: string | number,
  instanceId: string | number,
  { follow = true, tail = 100, source, since }: { follow?: boolean; tail?: number; source?: string; since?: string } = {}
): EventSource {
  const params = new URLSearchParams();
  params.set('follow', String(follow));
  params.set('tail', String(tail));
  if (source && source !== 'all') params.set('source', source);
  if (since) params.set('since', since);

  const token = localStorage.getItem('gdh_access_token');
  if (token) params.set('token', token);

  const url = `/api/v1/games/${gameId}/instances/${instanceId}/logs?${params.toString()}`;
  return new EventSource(url);
}

export async function fetchLogs(
  gameId: string | number,
  instanceId: string | number,
  { tail = 100, source }: { tail?: number; source?: string } = {}
): Promise<LogEntry[]> {
  const params = new URLSearchParams();
  params.set('follow', 'false');
  params.set('tail', String(tail));
  if (source && source !== 'all') params.set('source', source);

  const token = localStorage.getItem('gdh_access_token');
  if (token) params.set('token', token);

  const url = `/api/v1/games/${gameId}/instances/${instanceId}/logs?${params.toString()}`;
  const resp = await fetch(url);
  if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
  const text = await resp.text();

  const entries: LogEntry[] = [];
  const lines = text.split('\n');
  let currentData: string | null = null;
  for (const line of lines) {
    if (line.startsWith('event: ')) {
      // ignore event type
    } else if (line.startsWith('data: ')) {
      currentData = line.slice(6);
    } else if (line === '' && currentData !== null) {
      try {
        entries.push(JSON.parse(currentData));
      } catch {
        // ignore invalid data
      }
      currentData = null;
    }
  }
  return entries;
}
