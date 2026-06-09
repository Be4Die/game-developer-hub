import { getAuthHeaders } from './sso'

const API_URL = '/api/v1/chat'

export const chatApi = {
  async getConversations() {
    const res = await fetch(`${API_URL}/conversations`, {
      headers: getAuthHeaders(),
    })
    if (!res.ok) throw new Error('Failed to fetch conversations')
    return res.json()
  },

  async getMessages(conversationId, limit = 50, offset = 0) {
    const res = await fetch(
      `${API_URL}/conversations/${conversationId}/messages?limit=${limit}&offset=${offset}`,
      { headers: getAuthHeaders() }
    )
    if (!res.ok) throw new Error('Failed to fetch messages')
    return res.json()
  },

  async sendMessage(conversationId, content) {
    const res = await fetch(`${API_URL}/messages`, {
      method: 'POST',
      headers: { ...getAuthHeaders(), 'Content-Type': 'application/json' },
      body: JSON.stringify({ conversationId, content }),
    })
    if (!res.ok) throw new Error('Failed to send message')
    return res.json()
  },

  async markAsRead(conversationId) {
    const res = await fetch(`${API_URL}/conversations/${conversationId}/read`, {
      method: 'POST',
      headers: getAuthHeaders(),
    })
    if (!res.ok) throw new Error('Failed to mark as read')
  },

  async getUnreadCount() {
    const res = await fetch(`${API_URL}/unread`, {
      headers: getAuthHeaders(),
    })
    if (!res.ok) throw new Error('Failed to fetch unread count')
    const data = await res.json()
    return data.totalUnread || 0
  },

  async createConversation(participantId, participantName) {
    const res = await fetch(`${API_URL}/conversations`, {
      method: 'POST',
      headers: { ...getAuthHeaders(), 'Content-Type': 'application/json' },
      body: JSON.stringify({ participantId, participantName }),
    })
    if (!res.ok) throw new Error('Failed to create conversation')
    return res.json()
  },
}
