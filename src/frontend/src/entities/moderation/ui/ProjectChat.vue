<template>
  <div class="project-chat-container">
    <!-- Шапка чата: Заголовок, статус диалога и кнопка закрытия вопроса -->
    <div class="chat-header">
      <div class="chat-header-left">
        <MessageSquare class="icon-sm text-primary" />
        <h3 class="chat-title">{{ t('moderation.projectChatTitle') }}</h3>
        <!-- Статус диалога (отображается только если есть сообщения) -->
        <span
          v-if="dialogState !== 'none' && dialogStatusLabel"
          class="status-pill-sm"
          :class="dialogStatusClass"
        >
          {{ dialogStatusLabel }}
        </span>
      </div>
    </div>

    <!-- Область сообщений -->
    <div ref="messagesContainer" class="chat-messages-scroll">
      <div v-if="loading && !messages.length" class="chat-state-box">
        <span class="loader-spinner"></span>
        <p>{{ t('common.loading') }}</p>
      </div>

      <div v-else-if="!messages.length" class="chat-state-box empty">
        <MessageSquareDashed class="icon-lg text-muted" />
        <p>{{ t('moderation.noChats') }}</p>
        <span class="subtext"
          >Здесь фиксируются системные события и ведется диалог между разработчиком и
          модератором</span
        >
      </div>

      <div v-else class="messages-stack">
        <div
          v-for="msg in messages"
          :key="msg.id"
          class="message-row"
          :class="{
            'is-system': isSystemMessage(msg),
            'is-own': !isSystemMessage(msg) && isOwn(msg),
            'is-other': !isSystemMessage(msg) && !isOwn(msg),
          }"
        >
          <!-- Системное событие -->
          <div v-if="isSystemMessage(msg)" class="system-event-card">
            <Info class="icon-xs system-icon" />
            <div class="system-content">
              <span class="system-text">{{ msg.content }}</span>
              <span class="system-time">{{ formatTime(msg.created_at || msg.createdAt) }}</span>
            </div>
          </div>

          <!-- Обычное сообщение -->
          <div v-else class="chat-bubble">
            <div class="bubble-header">
              <span class="author-name">{{ formatSenderRole(msg) }}</span>
              <span class="bubble-time">{{ formatTime(msg.created_at || msg.createdAt) }}</span>
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
        :placeholder="t('moderation.chatPlaceholder')"
        rows="2"
        :disabled="sending"
        @keydown.enter.exact.prevent="handleSend"
      ></textarea>
      <button
        class="btn-send"
        :disabled="!inputContent.trim() || sending"
        :title="t('moderation.sendMessage')"
        @click="handleSend"
      >
        <Send class="icon-sm" />
      </button>
    </div>

    <!-- Нижняя панель действий чата: Кнопка закрытия вопроса модератором -->
    <div v-if="showResolveButton" class="chat-footer-actions">
      <button
        class="btn-resolve-dialog"
        :disabled="closingDialog"
        :title="t('moderation.closeDialogBtn')"
        @click="handleCloseDialog"
      >
        <CheckCircle2 class="icon-xs text-success" />
        <span>{{ t('moderation.closeDialogBtn') }}</span>
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { MessageSquare, MessageSquareDashed, Info, Send, CheckCircle2 } from 'lucide-vue-next';
import { moderationApi } from '../api/moderationApi';
import {
  formatDateTime,
  determineDialogState,
  parseSenderRole,
  parseMessageType,
} from '../model/helpers';
import { useAuth } from '@/entities/user';
import { showToast } from '@/shared/lib';

const { t } = useI18n();

const props = defineProps({
  projectId: {
    type: [Number, String],
    required: true,
  },
  autoPollInterval: {
    type: Number,
    default: 4000,
  },
});

const emit = defineEmits(['dialogStatusChanged']);

const { state: authState } = useAuth();
const currentUserId = computed(() => authState.user?.id || authState.user?.sub || '');
const currentUserRole = computed(() => authState.user?.role || '');

const isModeratorOrAdmin = computed(() => {
  const r = currentUserRole.value;
  return (
    r === 'USER_ROLE_MODERATOR' ||
    r === 'USER_ROLE_ADMIN' ||
    r === 'moderator' ||
    r === 'admin' ||
    r === 2 ||
    r === 3
  );
});

const messages = ref([]);
const loading = ref(false);
const sending = ref(false);
const closingDialog = ref(false);
const inputContent = ref('');
const messagesContainer = ref(null);

let pollTimer = null;

const dialogState = computed(() => {
  return determineDialogState(messages.value);
});

const dialogStatusLabel = computed(() => {
  if (dialogState.value === 'unanswered') {
    return isModeratorOrAdmin.value ? t('moderation.unanswered') : t('moderation.waitingResponse');
  }
  if (dialogState.value === 'in_dialog') {
    return t('moderation.inDialog');
  }
  if (dialogState.value === 'resolved') {
    return t('moderation.resolvedDialog');
  }
  return '';
});

const dialogStatusClass = computed(() => {
  if (dialogState.value === 'unanswered') return 'status-pill-unanswered';
  if (dialogState.value === 'in_dialog') return 'status-pill-in-dialog';
  if (dialogState.value === 'resolved') return 'status-pill-resolved';
  return 'status-pill-neutral';
});

const showResolveButton = computed(() => {
  return (
    isModeratorOrAdmin.value &&
    (dialogState.value === 'unanswered' || dialogState.value === 'in_dialog')
  );
});

function isSystemMessage(msg) {
  const msgType = parseMessageType(msg.message_type ?? msg.messageType);
  const senderRole = parseSenderRole(msg.sender_role ?? msg.senderRole);
  return msg.is_system || msgType > 1 || senderRole === 3 || msg.sender_id === 'system';
}

function isOwn(msg) {
  if (!currentUserId.value) return false;
  return String(msg.sender_id) === String(currentUserId.value);
}

function formatSenderRole(msg) {
  if (isOwn(msg)) return 'Вы';
  const r = parseSenderRole(msg.sender_role ?? msg.senderRole);
  if (r === 2) return t('moderation.moderatorRole');
  if (r === 1) return t('moderation.developerRole');
  return 'Пользователь';
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
      emit('dialogStatusChanged', dialogState.value);
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
    showToast(t('common.error'), 'danger');
  } finally {
    sending.value = false;
  }
}

async function handleCloseDialog() {
  closingDialog.value = true;
  try {
    await moderationApi.closeDialog(props.projectId);
    showToast(t('moderation.closeDialogSuccess'), 'success');
    await fetchMessages(true);
    scrollToBottom();
  } catch (err) {
    showToast(err.response?.data?.message || t('common.error'), 'danger');
  } finally {
    closingDialog.value = false;
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
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-md, 8px);
  overflow: hidden;
  height: 100%;
  width: 100%;
  box-sizing: border-box;
}

.chat-header {
  padding: 12px 16px;
  background: var(--bg-card, #161b22);
  border-bottom: 1px solid var(--border, #30363d);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-shrink: 0;
}

.chat-header-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.chat-title {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
}

.status-pill-sm {
  font-size: 11px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 4px;
  white-space: nowrap;
}

.status-pill-unanswered {
  background: rgba(245, 176, 39, 0.12);
  border: 1px solid rgba(245, 176, 39, 0.35);
  color: #f5b027;
}

.status-pill-in-dialog {
  background: rgba(88, 166, 255, 0.12);
  border: 1px solid rgba(88, 166, 255, 0.35);
  color: #58a6ff;
}

.status-pill-resolved {
  background: rgba(46, 204, 113, 0.12);
  border: 1px solid rgba(46, 204, 113, 0.35);
  color: #2ecc71;
}

.status-pill-neutral {
  background: var(--bg-tertiary, #21262d);
  border: 1px solid var(--border, #30363d);
  color: var(--text-muted, #b0b8c4);
}

.chat-footer-actions {
  padding: 0 12px 12px;
  background: var(--bg-card, #161b22);
}

.btn-resolve-dialog {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  height: 32px;
  padding: 0 12px;
  background: rgba(46, 204, 113, 0.08);
  border: 1px solid rgba(46, 204, 113, 0.3);
  color: #2ecc71;
  border-radius: var(--radius-sm, 6px);
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-resolve-dialog:hover:not(:disabled) {
  background: rgba(46, 204, 113, 0.18);
  border-color: #2ecc71;
  color: #2ecc71;
}

.btn-resolve-dialog:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.chat-messages-scroll {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  background: var(--bg-app, #0d1117);
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
  color: var(--text-tertiary, #8b949e);
}

.chat-state-box.empty .subtext {
  font-size: 12px;
  max-width: 280px;
  line-height: 1.4;
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
  background: var(--bg-secondary, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  padding: 6px 12px;
  max-width: 85%;
  font-size: 12px;
  color: var(--text-muted, #b0b8c4);
}

.system-icon {
  color: var(--primary, #58a6ff);
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
  font-size: 11px;
  color: var(--text-tertiary, #6e7681);
}

/* Пузыри сообщений */
.chat-bubble {
  max-width: 75%;
  padding: 10px 14px;
  border-radius: var(--radius-md, 8px);
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.message-row.is-own .chat-bubble {
  background: var(--primary, #58a6ff);
  color: #ffffff;
  border-bottom-right-radius: 2px;
}

.message-row.is-other .chat-bubble {
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border, #30363d);
  color: var(--text-main, #f0f6fc);
  border-bottom-left-radius: 2px;
}

.bubble-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  font-size: 11px;
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
  color: var(--primary, #58a6ff);
}

.message-row.is-other .bubble-time {
  color: var(--text-tertiary, #8b949e);
}

.bubble-body {
  font-size: 13px;
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
  background: var(--bg-card, #161b22);
  border-top: 1px solid var(--border, #30363d);
}

.chat-textarea {
  flex: 1;
  border: 1px solid var(--border, #30363d);
  background: var(--bg-secondary, #0d1117);
  color: var(--text-main, #f0f6fc);
  border-radius: var(--radius-sm, 6px);
  padding: 8px 12px;
  font-family: inherit;
  font-size: 13px;
  resize: none;
  box-sizing: border-box;
  outline: none;
  transition: border-color 0.15s;
}

.chat-textarea:focus {
  border-color: var(--primary, #58a6ff);
}

.btn-send {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  background: var(--primary, #58a6ff);
  color: #ffffff;
  border: none;
  border-radius: var(--radius-sm, 6px);
  cursor: pointer;
  transition:
    background-color 0.15s,
    opacity 0.15s;
  flex-shrink: 0;
}

.btn-send:hover:not(:disabled) {
  background: var(--primary-hover, #79c0ff);
}

.btn-send:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.loader-spinner {
  width: 18px;
  height: 18px;
  border: 2px solid var(--border, #30363d);
  border-top-color: var(--primary, #58a6ff);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  display: inline-block;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
