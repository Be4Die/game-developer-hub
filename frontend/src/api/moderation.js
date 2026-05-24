import { getAuthHeaders } from './sso'

const API_URL = '/api/v1/moderation'

export const moderationApi = {
  async submitForReview(gameId, gameName, gameDescription) {
    const res = await fetch(`${API_URL}/games/${gameId}/submit`, {
      method: 'POST',
      headers: { ...getAuthHeaders(), 'Content-Type': 'application/json' },
      body: JSON.stringify({
        game_name: gameName,
        game_description: gameDescription,
      }),
    })
    if (!res.ok) {
      const err = await res.json().catch(() => ({}))
      throw new Error(err.message || `Ошибка сервера (${res.status})`)
    }
    return res.json()
  },

  async listPending(limit = 50, offset = 0) {
    const res = await fetch(`${API_URL}/pending?limit=${limit}&offset=${offset}`, {
      headers: getAuthHeaders(),
    })
    if (!res.ok) throw new Error('Failed to fetch pending games')
    return res.json()
  },

  async getStatus(gameId) {
    const res = await fetch(`${API_URL}/games/${gameId}/status`, {
      headers: getAuthHeaders(),
    })
    if (res.status === 404) return null
    if (!res.ok) throw new Error('Failed to fetch moderation status')
    return res.json()
  },

  async approve(gameId) {
    const res = await fetch(`${API_URL}/games/${gameId}/approve`, {
      method: 'POST',
      headers: getAuthHeaders(),
    })
    if (!res.ok) throw new Error('Failed to approve game')
    return res.json()
  },

  async reject(gameId, reason) {
    const res = await fetch(`${API_URL}/games/${gameId}/reject`, {
      method: 'POST',
      headers: { ...getAuthHeaders(), 'Content-Type': 'application/json' },
      body: JSON.stringify({ reason }),
    })
    if (!res.ok) throw new Error('Failed to reject game')
    return res.json()
  },
}

export function moderationToTicket(m) {
  const statusMap = {
    MODERATION_STATUS_PENDING: 'pending',
    MODERATION_STATUS_APPROVED: 'approved',
    MODERATION_STATUS_REJECTED: 'rejected',
  }
  const submittedAt = m.submitted_at || m.submittedAt
  let created = ''
  if (submittedAt) {
    const ts = submittedAt.seconds ? submittedAt.seconds * 1000 : new Date(submittedAt).getTime()
    created = new Date(ts).toLocaleDateString('ru-RU')
  }
  return {
    id: Number(m.game_id ?? m.gameId),
    title: m.game_name || m.gameName || `Игра #${m.game_id || m.gameId}`,
    description: m.game_description || m.gameDescription || '',
    status: statusMap[m.status] || 'pending',
    priority: 'Средний',
    created,
    developerId: m.developer_id || m.developerId,
    rejectionReason: m.rejection_reason || m.rejectionReason || '',
    moderationStatus: m.status,
  }
}

export function ticketStatusText(status) {
  const map = {
    pending: 'На модерации',
    approved: 'Одобрено',
    rejected: 'Отклонено',
    new: 'Новый',
    in_progress: 'В работе',
    resolved: 'Решён',
  }
  return map[status] || status
}
