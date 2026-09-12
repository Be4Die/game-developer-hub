import { http } from '@/shared/api';
import { normalizeRequest } from '../model/helpers';

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
    const rawList = res.data.requests ?? [];
    return {
      requests: rawList.map(normalizeRequest),
      total: res.data.total ?? rawList.length,
    };
  },

  /**
   * Получить конкретную заявку по ID
   */
  async getRequest(requestId) {
    const res = await http.get(`/moderation/requests/${requestId}`);
    return { request: normalizeRequest(res.data.request) };
  },

  /**
   * Получить неизменяемый аудит-снимок по ID заявки
   */
  async getSnapshot(requestId) {
    try {
      const res = await http.get(`/moderation/requests/${requestId}/snapshot`);
      let parsed = {};
      const rawJson = res.data.snapshot_json || res.data.snapshotJson;
      if (rawJson) {
        try {
          parsed = typeof rawJson === 'string'
            ? JSON.parse(rawJson)
            : rawJson;
        } catch (e) {
          console.warn('Failed to parse snapshot_json:', e);
        }
      }
      return {
        requestId: res.data.request_id || res.data.requestId,
        projectId: res.data.project_id || res.data.projectId,
        status: res.data.status,
        createdAt: res.data.created_at || res.data.createdAt,
        payload: parsed,
      };
    } catch (err) {
      console.warn('Snapshot endpoint fallback to getRequest:', err);
      const fallback = await this.getRequest(requestId);
      const r = fallback.request;
      const snap = r.snapshot || {};
      return {
        requestId: r.id,
        projectId: r.projectId,
        status: r.status,
        createdAt: r.resolvedAt || r.submittedAt,
        payload: {
          project: {
            id: r.projectId,
            owner_id: r.ownerId,
            developer_name: r.ownerId,
            title_ru: snap.titleRu || '',
            title_en: snap.titleEn || '',
            seo_ru: snap.seoRu || '',
            seo_en: snap.seoEn || '',
            about_ru: snap.aboutRu || snap.about || '',
            about_en: snap.aboutEn || '',
            build_version: snap.activeBuildVersion || '',
            dev_url: snap.devUrl || '',
          },
          media: {
            icon: {
              file_name: 'icon.png',
              mime_type: 'image/png',
              original_url: snap.iconPath || '',
            },
            cover: {
              file_name: 'cover.png',
              mime_type: 'image/png',
              original_url: snap.coverPath || '',
            },
            video: {
              file_name: 'video.mp4',
              mime_type: 'video/mp4',
              original_url: snap.videoPath || '',
            },
          },
          verdict: {
            status: r.status,
            moderator_id: r.moderatorId,
            moderator_name: r.moderatorId,
            submitted_at: r.submittedAt,
            resolved_at: r.resolvedAt,
            rejection_reason: r.rejectionReason,
          },
          chat_transcript: [],
        },
      };
    }
  },

  /**
   * Получить последнюю заявку по ID проекта
   */
  async getLatestByProject(projectId) {
    try {
      const res = await http.get(`/moderation/projects/${projectId}/latest`);
      return { request: normalizeRequest(res.data.request) };
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
    return { request: normalizeRequest(res.data.request) };
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
   * Загрузить медиа-вложение (фото или видео) для чата проекта
   */
  async uploadAttachment(projectId, file) {
    const formData = new FormData();
    formData.append('file', file);
    const res = await http.post(`/projects/${projectId}/chat/attachments`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return res.data;
  },

  /**
   * Отклонить заявку на модерацию со списком нарушений
   */
  async reject(projectId, reason, violations = []) {
    const res = await http.post(`/moderation/projects/${projectId}/reject`, {
      project_id: projectId,
      reason,
      violations,
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
   * Отправить сообщение в чат проекта с возможными вложениями
   */
  async sendMessage(projectId, content, attachmentIds = [], payloadJson = '') {
    const res = await http.post(`/moderation/projects/${projectId}/messages`, {
      project_id: projectId,
      content,
      attachment_ids: attachmentIds,
      payload_json: payloadJson,
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
    if (params.action_type !== undefined && params.action_type !== '') {
      query.append('action_type', params.action_type);
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
      items: res.data.items ?? [],
      total: res.data.total ?? 0,
    };
  },
};
