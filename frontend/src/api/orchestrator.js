import axios from 'axios'

const http = axios.create({
  baseURL: '/api/orchestrator/v1',
  headers: { 'Content-Type': 'application/json' },
})

// ─── Билды ────────────────────────────────────────

export function listBuilds(gameId) {
  return http.get(`/games/${gameId}/builds`).then(r => r.data.builds ?? [])
}

export function getBuild(gameId, buildVersion) {
  return http.get(`/games/${gameId}/builds/${encodeURIComponent(buildVersion)}`).then(r => r.data)
}

export function uploadBuild(gameId, formData, onProgress) {
  return http.post(`/games/${gameId}/builds`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    onUploadProgress: onProgress,
  }).then(r => r.data)
}

export function deleteBuild(gameId, buildVersion) {
  return http.delete(`/games/${gameId}/builds/${encodeURIComponent(buildVersion)}`)
}

// ─── Инстансы ─────────────────────────────────────

export function listInstances(gameId, status) {
  const params = {}
  if (status && status !== 'all') params.status = status
  return http.get(`/games/${gameId}/instances`, { params }).then(r => r.data.instances ?? [])
}

export function getInstance(gameId, instanceId) {
  return http.get(`/games/${gameId}/instances/${instanceId}`).then(r => r.data)
}

export function startInstance(gameId, payload) {
  return http.post(`/games/${gameId}/instances`, payload).then(r => r.data)
}

export function stopInstance(gameId, instanceId, timeout = 30) {
  return http.delete(`/games/${gameId}/instances/${instanceId}`, { params: { timeout } }).then(r => r.data.instance ?? r.data)
}

export function getInstanceUsage(gameId, instanceId) {
  return http.get(`/games/${gameId}/instances/${instanceId}/usage`).then(r => r.data.usage ?? r.data)
}

// ─── Логи (SSE) ───────────────────────────────────

export function createLogStream(gameId, instanceId, { follow = true, tail = 100, source, since } = {}) {
  const params = new URLSearchParams()
  params.set('follow', String(follow))
  params.set('tail', String(tail))
  if (source && source !== 'all') params.set('source', source)
  if (since) params.set('since', since)

  const url = `/api/orchestrator/v1/games/${gameId}/instances/${instanceId}/logs?${params.toString()}`
  return new EventSource(url)
}

// ─── Ноды ─────────────────────────────────────────

export function listNodes(status) {
  const params = {}
  if (status && status !== 'all') params.status = status
  return http.get('/nodes', { params }).then(r => r.data.nodes ?? [])
}

export function getNode(nodeId) {
  return http.get(`/nodes/${nodeId}`).then(r => r.data)
}

export function registerNode(payload) {
  return http.post('/nodes', payload).then(r => r.data)
}

export function deleteNode(nodeId) {
  return http.delete(`/nodes/${nodeId}`)
}

export function getNodeUsage(nodeId) {
  return http.get(`/nodes/${nodeId}/usage`).then(r => r.data)
}
