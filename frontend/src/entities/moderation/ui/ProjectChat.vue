<template>
  <div class="project-chat-container">
    <div class="chat-header">
      <div class="chat-title-group">
        <MessageSquare class="icon-sm text-primary" />
        <h3 class="chat-title">{{ t('moderation.projectChatTitle') }}</h3>
      </div>
      <span class="messages-counter" v-if="messages.length">
        {{ messages.length }} сообщ.
      </span>
    </div>

    <!-- Область сообщений -->
    <div class="chat-messages-scroll" ref="messagesContainer">
      <div v-if="loading && !messages.length" class="chat-state-box">
        <span class="loader-spinner"></span>
        <p>Загрузка сообщений...</p>
      </div>

      <div v-else-if="!messages.length" class="chat-state-box empty">
        <MessageSquareDashed class="icon-lg text-muted" />
        <p>История сообщений пуста</p>
        <span class="subtext">Здесь фиксируются системные события и ведется диалог между разработчиком и модератором</span>
      </div>

      <div v-else class="messages-stack">
        <div
          v-for="msg in messages"
          :key="msg.id"
          class="message-row"
          :class="{
            'is-system': msg.is_system,
            'is-own': !msg.is_system && isOwn(msg),
            'is-other': !msg.is_system && !isOwn(msg),
          }"
        >
          <!-- Системное событие -->
          <div v-if="msg.is_system" class="system-event-card">
            <Info class="icon-xs system-icon" />
            <div class="system-content">
              <span class="system-text">{{ msg.content }}</span>
              <span class="system-time">{{ formatTime(msg.created_at) }}</span>
            </div>
          </div>

          <!-- Обычное сообщение -->
          <div v-else class="chat-bubble">
            <div class="bubble-header">
              <span class="author-name">{{ formatSenderRole(msg) }}</span>
              <span class="bubble-time">{{ formatTime(msg.created_at) }}</span>
            </div>
            <div class="bubble-body">{{ msg.content }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Поле ввода сообщения -->
    <div class="chat-input-row">
      <textarea
        v-model="inputContent"
        class="chat-textarea"
        placeholder="Напишите сообщение... (Enter для отправки)"
        rows="2"
        :disabled="sending"
        @keydown.enter.exact.prevent="handleSend"
      ></textarea>
      <button
        class="btn-send"
        :disabled="!inputContent.trim() || sending"
        @click="handleSend"
      >
        <Send class="icon-sm" />
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, nextTick, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { MessageSquare, MessageSquareDashed, Info, Send } from 'lucide-vue-next';
import { moderationApi } from '../api/moderationApi';
import { formatDateTime } from '../model/helpers';
import { useAuth } from '@/entities/user';

const { t } = useI18n();

const props = defineProps({
  projectId: {
    type: [Number, String],
    required: true,
  },
  autoPollInterval: {
    type: Number,
    default: 5000,
  },
});

const { state: authState } = useAuth();
const currentUserId = authState.user?.id;

const messages = ref([]);
const loading = ref(false);
const sending = ref(false);
const inputContent = ref('');
const messagesContainer = ref(null);

let pollTimer = null;

function isOwn(msg) {
  if (!currentUserId) return false;
  return String(msg.sender_id) === String(currentUserId);
}

function formatSenderRole(msg) {
  if (isOwn(msg)) return 'Вы';
  if (msg.sender_role === 'moderator' || msg.sender_role === 'USER_ROLE_MODERATOR') {
    return 'Модератор';
  }
  if (msg.sender_role === 'developer' || msg.sender_role === 'USER_ROLE_USER') {
    return 'Разработчик';
  }
  return msg.sender_id || 'Пользователь';
}

function formatTime(isoStr) {
  return formatDateTime(isoStr);
}

async function scrollToBottom() {
  await nextTick();
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight;
  }
}

async function fetchMessages(silent = false) {
  if (!props.projectId) return;
  if (!silent) loading.value = true;
  try {
    const data = await moderationApi.listMessages(props.projectId);
    const newMessages = data.messages || [];
    const hadChanges = newMessages.length !== messages.value.length;
    messages.value = newMessages;
    if (hadChanges) {
      scrollToBottom();
    }
  } catch (err) {
    console.error('Failed to fetch project messages:', err);
  } finally {
    if (!silent) loading.value = false;
  }
}

async function handleSend() {
  const text = inputContent.value.trim();
  if (!text || sending.value) return;

  sending.value = true;
  try {
    await moderationApi.sendMessage(props.projectId, text);
    inputContent.value = '';
    await fetchMessages(true);
    scrollToBottom();
  } catch (err) {
    console.error('Failed to send message:', err);
  } finally {
    sending.value = false;
  }
}

watch(
  () => props.projectId,
  (newId) => {
    if (newId) {
      fetchMessages();
    }
  }
);

onMounted(() => {
  fetchMessages();
  if (props.autoPollInterval > 0) {
    pollTimer = setInterval(() => {
      fetchMessages(true);
    }, props.autoPollInterval);
  }
});

onUnmounted(() => {
  if (pollTimer) {
    clearInterval(pollTimer);
    pollTimer = null;
  }
});
</script>

<style scoped>
.project-chat-container {
  display: flex;
  flex-direction: column;
  background: var(--bg-card);
  border: none;
  border-radius: 0;
  overflow: hidden;
  height: 100%;
  width: 100%;
  box-sizing: border-box;
}

.chat-header {
  padding: 14px 18px;
  background: var(--bg-card);
  border-bottom: 1px solid var(--border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
}

.chat-title-group {
  display: flex;
  align-items: center;
  gap: 8px;
}

.chat-title {
  margin: 0;
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--text-main);
}

.messages-counter {
  font-size: 0.75rem;
  color: var(--text-tertiary);
  font-weight: 500;
}

.chat-messages-scroll {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  background: var(--bg-app);
}

.chat-state-box {
  margin: auto;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 24px;
  text-align: center;
  color: var(--text-tertiary);
}

.chat-state-box.empty .subtext {
  font-size: 0.8rem;
  max-width: 280px;
}

.messages-stack {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.message-row {
  display: flex;
  width: 100%;
}

.message-row.is-system {
  justify-content: center;
}

.message-row.is-own {
  justify-content: flex-end;
}

.message-row.is-other {
  justify-content: flex-start;
}

/* Системное событие */
.system-event-card {
  display: flex;
  align-items: center;
  gap: 8px;
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  padding: 6px 12px;
  max-width: 85%;
  font-size: 0.8rem;
  color: var(--text-secondary);
}

.system-icon {
  color: var(--primary);
  flex-shrink: 0;
}

.system-content {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.system-text {
  font-weight: 500;
}

.system-time {
  font-size: 0.7rem;
  color: var(--text-tertiary);
}

/* Пузыри сообщений */
.chat-bubble {
  max-width: 75%;
  padding: 10px 14px;
  border-radius: var(--radius-lg);
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.message-row.is-own .chat-bubble {
  background: var(--primary);
  color: white;
  border-bottom-right-radius: 2px;
}

.message-row.is-other .chat-bubble {
  background: var(--bg-card);
  border: 1px solid var(--border);
  color: var(--text-main);
  border-bottom-left-radius: 2px;
}

.bubble-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  font-size: 0.75rem;
}

.message-row.is-own .author-name {
  font-weight: 600;
  color: rgba(255, 255, 255, 0.9);
}

.message-row.is-own .bubble-time {
  color: rgba(255, 255, 255, 0.7);
}

.message-row.is-other .author-name {
  font-weight: 600;
  color: var(--primary);
}

.message-row.is-other .bubble-time {
  color: var(--text-tertiary);
}

.bubble-body {
  font-size: 0.88rem;
  line-height: 1.4;
  white-space: pre-wrap;
  word-break: break-word;
}

/* Поле ввода */
.chat-input-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px;
  background: var(--bg-card);
  border-top: 1px solid var(--border);
}

.chat-textarea {
  flex: 1;
  border: 1px solid var(--border);
  background: var(--bg-app);
  color: var(--text-main);
  border-radius: var(--radius-md);
  padding: 8px 12px;
  font-family: inherit;
  font-size: 0.88rem;
  resize: none;
  box-sizing: border-box;
}

.chat-textarea:focus {
  outline: none;
  border-color: var(--primary);
}

.btn-send {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  background: var(--primary);
  color: white;
  border: none;
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: opacity 0.2s;
  flex-shrink: 0;
}

.btn-send:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}
</style>
