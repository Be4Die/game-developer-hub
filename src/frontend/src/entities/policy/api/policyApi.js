import { http } from '@/shared/api';

export function getPolicy(gameId) {
  return http.get(`/games/${gameId}/policy`).then((r) => r.data.policy ?? r.data);
}

export function setPolicy(gameId, payload) {
  return http.post(`/games/${gameId}/policy`, payload).then((r) => r.data.policy ?? r.data);
}

export function joinQueue(gameId, playerId, mode = '') {
  return http
    .post(`/games/${gameId}/queue/join`, { player_id: playerId, mode })
    .then((r) => r.data);
}

export function heartbeatQueue(gameId, playerId) {
  return http.post(`/games/${gameId}/queue/heartbeat`, { player_id: playerId }).then((r) => r.data);
}

export function leaveQueue(gameId, playerId) {
  return http
    .delete(`/games/${gameId}/queue/leave`, { params: { player_id: playerId } })
    .then((r) => r.data);
}

export function statusQueue(gameId, playerId) {
  return http
    .get(`/games/${gameId}/queue/status`, { params: { player_id: playerId } })
    .then((r) => r.data);
}

export function getQueueCount(gameId) {
  return http.get(`/games/${gameId}/queue/count`).then((r) => r.data.count ?? 0);
}

export function discoverServers(gameId, playerId = '') {
  const params = {};
  if (playerId) params.player_id = playerId;
  return http.get(`/games/${gameId}/discover`, { params }).then((r) => r.data);
}
