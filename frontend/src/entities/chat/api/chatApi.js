// Mock Chat API (in-memory & localStorage persistent)

const STORAGE_KEY = 'gdh_mock_chat_v1';

function getStoredData() {
  const defaultData = {
    conversations: [
      {
        id: 'conv-1',
        participant_id: '2',
        participantId: '2',
        participant_name: 'Анна (Модератор)',
        participantName: 'Анна (Модератор)',
        last_message: 'Пожалуйста, проверьте описание игры перед публикацией.',
        lastMessage: 'Пожалуйста, проверьте описание игры перед публикацией.',
        last_message_at: Math.floor((Date.now() - 1800000) / 1000),
        lastMessageAt: Math.floor((Date.now() - 1800000) / 1000),
        unread_count: 1,
        unreadCount: 1,
      },
      {
        id: 'conv-2',
        participant_id: '3',
        participantId: '3',
        participant_name: 'Максим (Разработчик)',
        participantName: 'Максим (Разработчик)',
        last_message: 'Спасибо, всё исправил!',
        lastMessage: 'Спасибо, всё исправил!',
        last_message_at: Math.floor((Date.now() - 3600000) / 1000),
        lastMessageAt: Math.floor((Date.now() - 3600000) / 1000),
        unread_count: 0,
        unreadCount: 0,
      },
    ],
    messages: [
      {
        id: 'm1',
        conversation_id: 'conv-1',
        conversationId: 'conv-1',
        sender_id: '2',
        senderId: '2',
        sender_name: 'Анна (Модератор)',
        senderName: 'Анна (Модератор)',
        content: 'Здравствуйте! Ваша заявка принята в работу.',
        created_at: Math.floor((Date.now() - 3600000) / 1000),
        createdAt: Math.floor((Date.now() - 3600000) / 1000),
        is_read: true,
        isRead: true,
      },
      {
        id: 'm2',
        conversation_id: 'conv-1',
        conversationId: 'conv-1',
        sender_id: '2',
        senderId: '2',
        sender_name: 'Анна (Модератор)',
        senderName: 'Анна (Модератор)',
        content: 'Пожалуйста, проверьте описание игры перед публикацией.',
        created_at: Math.floor((Date.now() - 1800000) / 1000),
        createdAt: Math.floor((Date.now() - 1800000) / 1000),
        is_read: false,
        isRead: false,
      },
      {
        id: 'm3',
        conversation_id: 'conv-2',
        conversationId: 'conv-2',
        sender_id: '3',
        senderId: '3',
        sender_name: 'Максим',
        senderName: 'Максим',
        content: 'Привет! Обновил билд до v1.1.0.',
        created_at: Math.floor((Date.now() - 7200000) / 1000),
        createdAt: Math.floor((Date.now() - 7200000) / 1000),
        is_read: true,
        isRead: true,
      },
      {
        id: 'm4',
        conversation_id: 'conv-2',
        conversationId: 'conv-2',
        sender_id: '3',
        senderId: '3',
        sender_name: 'Максим',
        senderName: 'Максим',
        content: 'Спасибо, всё исправил!',
        created_at: Math.floor((Date.now() - 3600000) / 1000),
        createdAt: Math.floor((Date.now() - 3600000) / 1000),
        is_read: true,
        isRead: true,
      },
    ],
  };

  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw) {
      return JSON.parse(raw);
    }
  } catch {
    // fallback
  }
  return defaultData;
}

function saveStoredData(data) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(data));
  } catch {
    // ignore
  }
}

function getCurrentUser() {
  try {
    const raw = localStorage.getItem('gdh_user');
    if (raw) return JSON.parse(raw);
  } catch {}
  return { id: '1', name: 'Михаил' };
}

export const chatApi = {
  async getConversations() {
    const data = getStoredData();
    return {
      conversations: data.conversations || [],
    };
  },

  async getMessages(conversationId, limit = 50, offset = 0) {
    const data = getStoredData();
    const msgs = (data.messages || [])
      .filter(
        (m) =>
          m.conversation_id === conversationId ||
          m.conversationId === conversationId
      )
      .sort(
        (a, b) =>
          (b.created_at || b.createdAt || 0) -
          (a.created_at || a.createdAt || 0)
      );
    const paged = msgs.slice(offset, offset + limit);
    return {
      messages: paged,
    };
  },

  async sendMessage(conversationId, content) {
    const data = getStoredData();
    const currentUser = getCurrentUser();
    const now = Math.floor(Date.now() / 1000);
    const newMsg = {
      id: `msg-${Date.now()}-${Math.random().toString(36).substr(2, 4)}`,
      conversation_id: conversationId,
      conversationId: conversationId,
      sender_id: currentUser.id || '1',
      senderId: currentUser.id || '1',
      sender_name: currentUser.name || 'Пользователь',
      senderName: currentUser.name || 'Пользователь',
      content: content.trim(),
      created_at: now,
      createdAt: now,
      is_read: true,
      isRead: true,
    };

    if (!data.messages) data.messages = [];
    data.messages.push(newMsg);

    const conv = (data.conversations || []).find((c) => c.id === conversationId);
    if (conv) {
      conv.last_message = newMsg.content;
      conv.lastMessage = newMsg.content;
      conv.last_message_at = now;
      conv.lastMessageAt = now;
    }

    saveStoredData(data);
    return { message: newMsg };
  },

  async markAsRead(conversationId) {
    const data = getStoredData();
    const conv = (data.conversations || []).find((c) => c.id === conversationId);
    if (conv) {
      conv.unread_count = 0;
      conv.unreadCount = 0;
    }
    if (data.messages) {
      data.messages.forEach((m) => {
        if (
          m.conversation_id === conversationId ||
          m.conversationId === conversationId
        ) {
          m.is_read = true;
          m.isRead = true;
        }
      });
    }
    saveStoredData(data);
  },

  async getUnreadCount() {
    const data = getStoredData();
    let total = 0;
    for (const c of data.conversations || []) {
      total += c.unread_count || c.unreadCount || 0;
    }
    return total;
  },

  async createConversation(participantId, participantName) {
    const data = getStoredData();
    let conv = (data.conversations || []).find(
      (c) =>
        c.participant_id === String(participantId) ||
        c.participantId === String(participantId)
    );
    if (!conv) {
      const now = Math.floor(Date.now() / 1000);
      conv = {
        id: `conv-${Date.now()}-${Math.random().toString(36).substr(2, 4)}`,
        participant_id: String(participantId),
        participantId: String(participantId),
        participant_name: participantName || 'Пользователь',
        participantName: participantName || 'Пользователь',
        last_message: '',
        lastMessage: '',
        last_message_at: now,
        lastMessageAt: now,
        unread_count: 0,
        unreadCount: 0,
      };
      if (!data.conversations) data.conversations = [];
      data.conversations.unshift(conv);
      saveStoredData(data);
    }
    return { conversation: conv };
  },
};
