import { http } from '@/shared/api';

export const moderationApi = {
  async submitForReview(gameId) {
    const res = await http.post(`/projects/${gameId}/moderation`);
    return { moderation: res.data.ticket };
  },

  async listPending(limit = 50, offset = 0) {
    const res = await http.get(
      `/moderation/tickets?status=1&limit=${limit}&offset=${offset}`
    );
    return {
      moderations: res.data.tickets ?? [],
    };
  },

  async getStatus(gameId) {
    try {
      const res = await http.get(`/moderation/tickets/by-project/${gameId}`);
      return { moderation: res.data.ticket };
    } catch {
      return null;
    }
  },

  async approve(gameId, comment = 'Одобрено') {
    const res = await http.post(`/moderation/tickets/${gameId}/approve`, {
      comment,
    });
    return { release: res.data.release };
  },

  async reject(gameId, reason) {
    const res = await http.post(`/moderation/tickets/${gameId}/reject`, {
      reason,
    });
    return { ticket: res.data.ticket };
  },
};
