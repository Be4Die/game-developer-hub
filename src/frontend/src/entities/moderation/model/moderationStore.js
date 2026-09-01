import { reactive } from 'vue';
import { moderationApi } from '../api/moderationApi';
import { normalizeRequest } from './helpers';
import { showToast } from '@/shared/lib';

export const moderationStore = reactive({
  requests: [],
  total: 0,
  currentRequest: null,
  activeChats: [],
  totalChats: 0,
  messages: [],
  loading: false,
  messagesLoading: false,

  async loadRequests(params = {}) {
    this.loading = true;
    try {
      const data = await moderationApi.listRequests(params);
      this.requests = (data.requests || []).map(normalizeRequest);
      this.total = data.total;
    } catch (err) {
      console.error('Failed to load moderation requests:', err);
      showToast('Ошибка загрузки заявок на модерацию', 'danger');
    } finally {
      this.loading = false;
    }
  },

  async loadRequestById(requestId) {
    this.loading = true;
    try {
      const data = await moderationApi.getRequest(requestId);
      this.currentRequest = normalizeRequest(data.request);
      return this.currentRequest;
    } catch (err) {
      console.error('Failed to load request:', err);
      showToast('Ошибка загрузки заявки', 'danger');
      return null;
    } finally {
      this.loading = false;
    }
  },

  async claimRequest(requestId) {
    try {
      const data = await moderationApi.claimRequest(requestId);
      const updated = normalizeRequest(data.request);
      if (this.currentRequest && this.currentRequest.id === updated.id) {
        this.currentRequest = updated;
      }
      const idx = this.requests.findIndex((r) => r.id === updated.id);
      if (idx >= 0) {
        this.requests[idx] = updated;
      }
      showToast('Заявка взята в работу', 'success');
      return updated;
    } catch (err) {
      console.error('Failed to claim request:', err);
      showToast(err.response?.data?.message || 'Не удалось взять заявку в работу', 'danger');
      throw err;
    }
  },

  async approveRequest(projectId, comment) {
    try {
      const res = await moderationApi.approve(projectId, comment);
      showToast('Заявка одобрена, игра опубликована!', 'success');
      return res;
    } catch (err) {
      console.error('Failed to approve request:', err);
      showToast(err.response?.data?.message || 'Ошибка одобрения заявки', 'danger');
      throw err;
    }
  },

  async rejectRequest(projectId, reason) {
    try {
      const res = await moderationApi.reject(projectId, reason);
      showToast('Заявка отклонена', 'info');
      return res;
    } catch (err) {
      console.error('Failed to reject request:', err);
      showToast(err.response?.data?.message || 'Ошибка отклонения заявки', 'danger');
      throw err;
    }
  },

  async loadMessages(projectId) {
    this.messagesLoading = true;
    try {
      const data = await moderationApi.listMessages(projectId);
      this.messages = data.messages || [];
      return this.messages;
    } catch (err) {
      console.error('Failed to load chat messages:', err);
      return [];
    } finally {
      this.messagesLoading = false;
    }
  },

  async sendMessage(projectId, content) {
    try {
      const data = await moderationApi.sendMessage(projectId, content);
      if (data.message) {
        this.messages.push(data.message);
      }
      return data.message;
    } catch (err) {
      console.error('Failed to send message:', err);
      showToast('Не удалось отправить сообщение', 'danger');
      throw err;
    }
  },
  async loadActiveChats(params = {}) {
    this.loading = true;
    try {
      const data = await moderationApi.listActiveChats(params);
      this.activeChats = (data.chats || []).map(c => ({
        projectId: c.project_id,
        lastMessage: c.last_message ? {
          id: c.last_message.id,
          content: c.last_message.content,
          createdAt: c.last_message.created_at,
          senderRole: c.last_message.sender_role,
        } : null
      }));
      this.totalChats = data.total || 0;
    } catch (err) {
      console.error('Failed to load active chats:', err);
    } finally {
      this.loading = false;
    }
  },
});
