import { http } from '@/shared/api';
import type { OrchestrationPolicy, QueueStatus } from '@/shared/types';

export function getPolicy(gameId: string | number): Promise<OrchestrationPolicy> {
  return http.get(`/games/${gameId}/policy`).then((r) => r.data.policy ?? r.data);
}

export function setPolicy(gameId: string | number, payload: any): Promise<OrchestrationPolicy> {
  return http.post(`/games/${gameId}/policy`, payload).then((r) => r.data.policy ?? r.data);
}

export function joinQueue(gameId: string | number, playerId: string, mode = ''): Promise<any> {
  return http
    .post(`/games/${gameId}/queue/join`, { player_id: playerId, mode })
    .then((r) => r.data);
}

export function heartbeatQueue(gameId: string | number, playerId: string): Promise<any> {
  return http.post(`/games/${gameId}/queue/heartbeat`, { player_id: playerId }).then((r) => r.data);
}

export function leaveQueue(gameId: string | number, playerId: string): Promise<any> {
  return http
    .delete(`/games/${gameId}/queue/leave`, { params: { player_id: playerId } })
    .then((r) => r.data);
}

export function statusQueue(gameId: string | number, playerId: string): Promise<QueueStatus> {
  return http
    .get(`/games/${gameId}/queue/status`, { params: { player_id: playerId } })
    .then((r) => r.data);
}

export function getQueueCount(gameId: string | number): Promise<number> {
  return http.get(`/games/${gameId}/queue/count`).then((r) => r.data.count ?? 0);
}

export function discoverServers(gameId: string | number, playerId = ''): Promise<any> {
  const params: Record<string, string> = {};
  if (playerId) params.player_id = playerId;
  return http.get(`/games/${gameId}/discover`, { params }).then((r) => r.data);
}
