import { http } from '@/shared/api';

export const moderationApi = {
  /**
   * Получить список заявок на модерацию с фильтрацией
   */
  async listRequests(params = {}) {
    const query = new URLSearchParams();
    if (params.status !== undefined && params.status !== '') {
      query.append('status', params.status);
    }
    if (params.moderator_id) {
      query.append('moderator_id', params.moderator_id);
    }
    if (params.limit !== undefined) {
      query.append('limit', params.limit);
    }
    if (params.offset !== undefined) {
      query.append('offset', params.offset);
    }
    const qStr = query.toString() ? `?${query.toString()}` : '';
    const res = await http.get(`/moderation/requests${qStr}`);
    return {
      requests: res.data.requests ?? [],
      total: res.data.total ?? (res.data.requests ? res.data.requests.length : 0),
    };
  },

  /**
   * Получить конкретную заявку по ID
   */
  async getRequest(requestId) {
    const res = await http.get(`/moderation/requests/${requestId}`);
    return { request: res.data.request };
  },

  /**
   * Получить последнюю заявку по ID проекта
   */
  async getLatestByProject(projectId) {
    try {
      const res = await http.get(`/moderation/projects/${projectId}/latest`);
      return { request: res.data.request };
    } catch (err) {
      if (err.response && err.response.status === 404) {
        return { request: null };
      }
      throw err;
    }
  },

  /**
   * Взять заявку в работу модератором
   */
  async claimRequest(requestId) {
    const res = await http.post(`/moderation/requests/${requestId}/claim`);
    return { request: res.data.request };
  },

  /**
   * Одобрить заявку на модерацию
   */
  async approve(projectId, comment = 'Одобрено') {
    const res = await http.post(`/moderation/projects/${projectId}/approve`, {
      comment,
    });
    return {
      success: res.data.success,
      prod_url: res.data.prod_url,
    };
  },

  /**
   * Отклонить заявку на модерацию
   */
  async reject(projectId, reason) {
    const res = await http.post(`/moderation/projects/${projectId}/reject`, {
      reason,
    });
    return {
      success: res.data.success,
    };
  },

  /**
   * Получить сообщения чата проекта
   */
  async listMessages(projectId, params = {}) {
    const query = new URLSearchParams();
    if (params.limit !== undefined) query.append('limit', params.limit);
    if (params.offset !== undefined) query.append('offset', params.offset);
    const qStr = query.toString() ? `?${query.toString()}` : '';
    const res = await http.get(`/moderation/projects/${projectId}/messages${qStr}`);
    return {
      messages: res.data.messages ?? [],
      total: res.data.total ?? (res.data.messages ? res.data.messages.length : 0),
    };
  },

  /**
   * Отправить сообщение в чат проекта
   */
  async sendMessage(projectId, content) {
    const res = await http.post(`/moderation/projects/${projectId}/messages`, {
      content,
    });
    return {
      message: res.data.message,
    };
  },

  /**
   * Получить список активных чатов
   */
  async listActiveChats(params = {}) {
    const query = new URLSearchParams();
    if (params.limit !== undefined) query.append('limit', params.limit);
    if (params.offset !== undefined) query.append('offset', params.offset);
    const qStr = query.toString() ? `?${query.toString()}` : '';
    const res = await http.get(`/moderation/chats${qStr}`);
    return {
      chats: res.data.chats ?? [],
      total: res.data.total ?? (res.data.chats ? res.data.chats.length : 0),
    };
  },

  /**
   * Закрыть диалог модерации по проекту (Вопрос решён)
   */
  async closeDialog(projectId, comment = '') {
    const res = await http.post(`/moderation/projects/${projectId}/close-dialog`, {
      comment,
    });
    return {
      success: res.data.success,
      message: res.data.message,
    };
  },

  /**
   * Получить статистику конкретного модератора
   */
  async getModeratorStats(moderatorId) {
    const res = await http.get(`/moderation/moderators/${moderatorId}/stats`);
    return { stats: res.data.stats };
  },

  /**
   * Получить сводную статистику по всем модераторам
   */
  async listModeratorsStats() {
    const res = await http.get('/moderation/moderators/stats');
    return { stats: res.data.stats ?? [] };
  },

  /**
   * Получить журнал действий конкретного модератора
   */
  async listModeratorActivity(moderatorId, params = {}) {
    const query = new URLSearchParams();
    if (params.status !== undefined && params.status !== '') {
      query.append('status', params.status);
    }
    if (params.limit !== undefined) {
      query.append('limit', params.limit);
    }
    if (params.offset !== undefined) {
      query.append('offset', params.offset);
    }
    const qStr = query.toString() ? `?${query.toString()}` : '';
    const res = await http.get(`/moderation/moderators/${moderatorId}/activity${qStr}`);
    return {
      requests: res.data.requests ?? [],
      total: res.data.total ?? (res.data.requests ? res.data.requests.length : 0),
    };
  },
};
