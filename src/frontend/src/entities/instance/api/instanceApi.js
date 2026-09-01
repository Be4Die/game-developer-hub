import { http } from '@/shared/api';
import { normalizeInstance } from '../model/instanceModel';

export function listInstances(gameId, status) {
  const params = {};
  if (status && status !== 'all') params.status = status;
  return http
    .get(`/games/${gameId}/instances`, { params })
    .then((r) => (r.data.instances ?? []).map(normalizeInstance));
}

export function getInstance(gameId, instanceId) {
  return http
    .get(`/games/${gameId}/instances/${instanceId}`)
    .then((r) => normalizeInstance(r.data));
}

export function startInstance(gameId, payload) {
  return http.post(`/games/${gameId}/instances`, payload).then((r) => normalizeInstance(r.data));
}

export function stopInstance(gameId, instanceId, timeout = 30) {
  return http
    .post(`/games/${gameId}/instances/${instanceId}:stop`, { timeout })
    .then((r) => normalizeInstance(r.data.instance ?? r.data));
}

export function deleteInstance(gameId, instanceId) {
  return http.delete(`/games/${gameId}/instances/${instanceId}`).then((r) => r.data);
}

export function restartInstance(gameId, instanceId) {
  return http
    .post(`/games/${gameId}/instances/${instanceId}:restart`)
    .then((r) => normalizeInstance(r.data.instance ?? r.data));
}

export function resumeInstance(gameId, instanceId) {
  return http
    .post(`/games/${gameId}/instances/${instanceId}:resume`)
    .then((r) => normalizeInstance(r.data.instance ?? r.data));
}

export function getInstanceUsage(gameId, instanceId) {
  return http
    .get(`/games/${gameId}/instances/${instanceId}/usage`)
    .then((r) => r.data.usage ?? r.data);
}

export function createLogStream(
  gameId,
  instanceId,
  { follow = true, tail = 100, source, since } = {}
) {
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

export async function fetchLogs(gameId, instanceId, { tail = 100, source } = {}) {
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

  const entries = [];
  const lines = text.split('\n');
  let currentData = null;
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
