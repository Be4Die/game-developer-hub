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
            'is-verdict': isRejectionVerdict(msg),
            'is-system': !isRejectionVerdict(msg) && isSystemMessage(msg),
            'is-own': !isRejectionVerdict(msg) && !isSystemMessage(msg) && isOwn(msg),
            'is-other': !isRejectionVerdict(msg) && !isSystemMessage(msg) && !isOwn(msg),
          }"
        >
          <!-- Вердикт модерации с конкретными нарушениями правил -->
          <div v-if="isRejectionVerdict(msg)" class="verdict-row-wrap">
            <ModerationVerdictCard :message="msg" />
          </div>

          <!-- Системное сообщение (строго по центру, отдельный стиль от обычных сообщений) -->
          <div v-else-if="isSystemMessage(msg)" class="system-event-wrapper">
            <div class="system-event-pill">
              <span class="system-badge">Система</span>
              <span class="system-text">{{ msg.content }}</span>
              <span class="system-time">{{ formatTime(msg.created_at || msg.createdAt) }}</span>
            </div>
          </div>

          <!-- Обычное сообщение пользователя / модератора -->
          <div v-else class="chat-bubble">
            <div class="bubble-header">
              <span class="author-name">{{ formatSenderRole(msg) }}</span>
              <span class="bubble-time">{{ formatTime(msg.created_at || msg.createdAt) }}</span>
            </div>
            <div v-if="msg.content" class="bubble-body">{{ msg.content }}</div>

            <!-- Вложения обычного сообщения -->
            <div
              v-if="msg.attachments && msg.attachments.length"
              class="bubble-attachments-stack"
            >
              <div
                v-for="att in msg.attachments"
                :key="att.id"
                class="bubble-attachment-item"
              >
                <!-- Видео -->
                <div v-if="isVideo(att)" class="bubble-video-wrapper">
                  <video
                    :src="getMediaUrl(att, msg)"
                    controls
                    preload="metadata"
                    class="embedded-bubble-video"
                  ></video>
                </div>

                <!-- Изображение -->
                <div
                  v-else-if="isImage(att)"
                  class="bubble-image-wrapper"
                  @click="openLightbox(att, msg)"
                >
                  <img
                    :src="getMediaUrl(att, msg)"
                    :alt="att.file_name || att.fileName"
                    class="embedded-bubble-image"
                    loading="lazy"
                  />
                  <div class="image-overlay">
                    <Maximize2 class="icon-xs" />
                    <span>Увеличить</span>
                  </div>
                </div>

                <!-- Скачивание файла -->
                <div class="attachment-download-bar">
                  <a
                    :href="getDownloadUrl(att, msg)"
                    :download="att.file_name || att.fileName || 'attachment'"
                    class="attachment-download-link"
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    <Paperclip class="icon-xs" />
                    <span class="att-filename">{{ att.file_name || att.fileName }}</span>
                    <span class="att-filesize">({{ formatSize(att.file_size || att.fileSize) }})</span>
                    <span class="att-dl-action">• Скачать</span>
                  </a>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Поле ввода сообщения -->
    <div
      v-if="!isChatReadOnly"
      class="chat-input-area"
      @dragover.prevent="isDragging = true"
      @dragleave.prevent="isDragging = false"
      @drop.prevent="handleDrop"
    >
      <!-- Превью выбранных вложений перед отправкой -->
      <div v-if="pendingAttachments.length" class="pending-attachments-row">
        <div
          v-for="(att, idx) in pendingAttachments"
          :key="att.id || idx"
          class="pending-attachment-chip"
        >
          <Film v-if="isVideo(att)" class="icon-xs text-primary" />
          <Image v-else class="icon-xs text-primary" />
          <span class="pending-filename">{{ att.file_name || att.name }}</span>
          <span class="pending-filesize">({{ formatSize(att.file_size || att.size) }})</span>
          <button
            type="button"
            class="btn-remove-pending"
            title="Удалить"
            @click="removePendingAttachment(idx)"
          >
            <X class="icon-xs" />
          </button>
        </div>
        <div v-if="uploadingCount > 0" class="uploading-indicator">
          <Loader2 class="icon-xs spin" />
          <span>Загрузка...</span>
        </div>
      </div>

      <div class="chat-input-row" :class="{ 'is-drag-over': isDragging }">
        <!-- Кнопка прикрепления файлов -->
        <label
          class="btn-attach"
          :class="{ disabled: uploadingCount > 0 || sending }"
          title="Прикрепить фото или видео"
        >
          <input
            ref="fileInputRef"
            type="file"
            accept="image/png,image/jpeg,image/webp,video/mp4,video/webm"
            multiple
            class="file-hidden-input"
            :disabled="uploadingCount > 0 || sending"
            @change="handleFileSelect"
          />
          <Paperclip class="icon-sm" />
        </label>

        <textarea
          v-model="inputContent"
          class="chat-textarea"
          :placeholder="t('moderation.chatPlaceholder') + ' (вставка скриншота по Ctrl+V)'"
          rows="2"
          :disabled="sending"
          @keydown.enter.exact.prevent="handleSend"
          @paste="handlePaste"
        ></textarea>

        <button
          class="btn-send"
          :disabled="
            (!inputContent.trim() && !pendingAttachments.length) ||
            sending ||
            uploadingCount > 0
          "
          :title="t('moderation.sendMessage')"
          @click="handleSend"
        >
          <Send class="icon-sm" />
        </button>
      </div>
    </div>

    <div v-else class="chat-readonly-banner">
      <Eye class="icon-xs text-muted" />
      <span>Режим аудита: чат доступен только для чтения</span>
    </div>

    <!-- Нижняя панель действий чата: Кнопка закрытия вопроса модератором -->
    <div v-if="!isChatReadOnly && showResolveButton" class="chat-footer-actions">
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

    <!-- Модальное окно просмотра картинок в оригинале -->
    <MediaLightboxModal
      v-if="lightboxMedia"
      :src="getMediaUrl(lightboxMedia.att, lightboxMedia.msg)"
      :file-name="lightboxMedia.att.file_name || lightboxMedia.att.fileName"
      :download-url="getDownloadUrl(lightboxMedia.att, lightboxMedia.msg)"
      @close="lightboxMedia = null"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import {
  MessageSquare,
  MessageSquareDashed,
  Info,
  Send,
  CheckCircle2,
  Eye,
  Paperclip,
  Image,
  Film,
  X,
  Maximize2,
  Loader2,
} from 'lucide-vue-next';
import { moderationApi } from '../api/moderationApi';
import {
  formatDateTime,
  determineDialogState,
  parseSenderRole,
  parseMessageType,
} from '../model/helpers';
import { useAuth } from '@/entities/user';
import { showToast } from '@/shared/lib';
import {
  validateChatFile,
  compressImageIfNeeded,
  formatBytes,
} from '@/shared/lib/mediaCompressor';
import ModerationVerdictCard from './ModerationVerdictCard.vue';
import MediaLightboxModal from './MediaLightboxModal.vue';

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
  readonly: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(['dialogStatusChanged']);

const { state: authState } = useAuth();
const currentUserId = computed(() => authState.user?.id || authState.user?.sub || '');
const currentUserRole = computed(() => authState.user?.role || '');

const isAdmin = computed(() => {
  const r = currentUserRole.value;
  return r === 'USER_ROLE_ADMIN' || r === 'admin' || r === 3;
});

const isChatReadOnly = computed(() => props.readonly || isAdmin.value);

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
const fileInputRef = ref(null);

// Вложения
const pendingAttachments = ref([]);
const uploadingCount = ref(0);
const isDragging = ref(false);
const lightboxMedia = ref(null);

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

const hasDisputeOrQuestions = computed(() => {
  return messages.value.some((m) => {
    const msgType = parseMessageType(m.message_type ?? m.messageType);
    if (msgType === 5) return true;
    const content = (m.content || '').toLowerCase();
    return (
      content.includes('замечани') ||
      content.includes('вопрос') ||
      content.includes('нарушен') ||
      content.includes('отклон') ||
      content.includes('баг') ||
      content.includes('почему') ||
      content.includes('исправ')
    );
  });
});

const showResolveButton = computed(() => {
  return (
    isModeratorOrAdmin.value &&
    hasDisputeOrQuestions.value &&
    (dialogState.value === 'unanswered' || dialogState.value === 'in_dialog')
  );
});

function isRejectionVerdict(msg) {
  const msgType = parseMessageType(msg.message_type ?? msg.messageType);
  if (msgType === 5) return true;
  let p = msg.payload;
  if (!p && msg.payload_json) {
    try {
      p = JSON.parse(msg.payload_json);
    } catch {
      p = null;
    }
  }
  return p && (p.type === 'moderation_verdict' || Array.isArray(p.violations));
}

function isSystemMessage(msg) {
  if (isRejectionVerdict(msg)) return false;
  const msgType = parseMessageType(msg.message_type ?? msg.messageType);
  const senderRole = parseSenderRole(msg.sender_role ?? msg.senderRole);
  return msg.is_system || msgType > 1 || senderRole === 3 || msg.sender_id === 'system';
}

function isOwn(msg) {
  if (!currentUserId.value) return false;
  return String(msg.sender_id) === String(currentUserId.value);
}

function formatSenderRole(msg) {
  const r = parseSenderRole(msg.sender_role ?? msg.senderRole);
  if (r === 3 || msg.sender_id === 'system') return 'Система';
  if (isOwn(msg)) return 'Вы';
  if (r === 2) return t('moderation.moderatorRole');
  if (r === 1) return t('moderation.developerRole');
  return 'Пользователь';
}

function formatTime(isoStr) {
  return formatDateTime(isoStr);
}

function formatSize(bytes) {
  return formatBytes(bytes);
}

function isVideo(att) {
  const mime = (att.mime_type || att.mimeType || att.type || '').toLowerCase();
  const name = (att.file_name || att.fileName || att.name || '').toLowerCase();
  return mime.startsWith('video/') || name.endsWith('.mp4') || name.endsWith('.webm');
}

function isImage(att) {
  const mime = (att.mime_type || att.mimeType || att.type || '').toLowerCase();
  const name = (att.file_name || att.fileName || att.name || '').toLowerCase();
  return (
    mime.startsWith('image/') ||
    name.endsWith('.png') ||
    name.endsWith('.jpg') ||
    name.endsWith('.jpeg') ||
    name.endsWith('.webp')
  );
}

function getMediaUrl(att, msg) {
  const token = localStorage.getItem('gdh_access_token');
  const projId = msg?.project_id || msg?.projectId || props.projectId;
  const base = att.url || `/api/v1/projects/${projId}/chat/attachments/${att.id}`;
  if (token && !base.includes('token=')) {
    const sep = base.includes('?') ? '&' : '?';
    return `${base}${sep}token=${encodeURIComponent(token)}`;
  }
  return base;
}

function getDownloadUrl(att, msg) {
  const token = localStorage.getItem('gdh_access_token');
  const projId = msg?.project_id || msg?.projectId || props.projectId;
  const base = att.download_url || `/api/v1/projects/${projId}/chat/attachments/${att.id}/download`;
  if (token && !base.includes('token=')) {
    const sep = base.includes('?') ? '&' : '?';
    return `${base}${sep}token=${encodeURIComponent(token)}`;
  }
  return base;
}

function openLightbox(att, msg) {
  lightboxMedia.value = { att, msg };
}

function removePendingAttachment(index) {
  pendingAttachments.value.splice(index, 1);
}

async function processAndUploadFile(file) {
  const validation = validateChatFile(file);
  if (!validation.valid) {
    showToast(validation.error, 'danger');
    return;
  }

  uploadingCount.value++;
  try {
    let processedFile = file;
    if (validation.isImage) {
      processedFile = await compressImageIfNeeded(file, 1920, 0.85);
    }
    const uploaded = await moderationApi.uploadAttachment(props.projectId, processedFile);
    pendingAttachments.value.push(uploaded);
    showToast('Файл прикреплен', 'success');
  } catch (err) {
    console.error('Failed to upload attachment:', err);
    const msg =
      err.response?.data?.message ||
      (typeof err.response?.data === 'string' ? err.response.data.trim() : null);
    showToast(msg || 'Ошибка загрузки файла', 'danger');
  } finally {
    uploadingCount.value--;
  }
}

function handleFileSelect(e) {
  const files = Array.from(e.target.files || []);
  e.target.value = '';
  files.forEach(processAndUploadFile);
}

function handleDrop(e) {
  isDragging.value = false;
  const files = Array.from(e.dataTransfer?.files || []);
  files.forEach(processAndUploadFile);
}

function handlePaste(e) {
  const items = e.clipboardData?.items;
  if (!items) return;
  for (let i = 0; i < items.length; i++) {
    const item = items[i];
    if (item.type.indexOf('image') !== -1) {
      const file = item.getAsFile();
      if (file) {
        processAndUploadFile(file);
      }
    }
  }
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
  const attachmentIds = pendingAttachments.value.map((a) => a.id);
  if ((!text && !attachmentIds.length) || sending.value || uploadingCount.value > 0) return;

  sending.value = true;
  try {
    await moderationApi.sendMessage(props.projectId, text, attachmentIds);
    inputContent.value = '';
    pendingAttachments.value = [];
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
      pendingAttachments.value = [];
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
  gap: 12px;
}

.message-row {
  display: flex;
  width: 100%;
}

.message-row.is-verdict {
  justify-content: center;
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

.verdict-row-wrap {
  width: 100%;
  max-width: 90%;
}

/* Системные события (по центру) */
.message-row.is-system {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 100%;
  margin: 6px 0;
}

.system-event-wrapper {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 100%;
}

.system-event-pill {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  background: var(--bg-secondary, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: 9999px;
  padding: 5px 14px;
  max-width: 90%;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.15);
}

.system-badge {
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  color: #79c0ff;
  background: rgba(56, 139, 253, 0.15);
  border: 1px solid rgba(56, 139, 253, 0.3);
  border-radius: 4px;
  padding: 1px 6px;
  line-height: 1.3;
  flex-shrink: 0;
}

.system-text {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-muted, #8b949e);
  line-height: 1.4;
  text-align: center;
}

.system-time {
  font-size: 11px;
  color: var(--text-tertiary, #6e7681);
  white-space: nowrap;
  flex-shrink: 0;
}

/* Пузыри сообщений */
.chat-bubble {
  max-width: 88%;
  padding: 10px 14px;
  border-radius: var(--radius-md, 8px);
  display: flex;
  flex-direction: column;
  gap: 6px;
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

/* Вложения внутри пузыря */
.bubble-attachments-stack {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-top: 4px;
}

.bubble-attachment-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.bubble-image-wrapper {
  position: relative;
  max-width: 100%;
  border-radius: 6px;
  overflow: hidden;
  cursor: pointer;
  border: 1px solid rgba(255, 255, 255, 0.15);
  background: #000;
}

.embedded-bubble-image {
  width: 100%;
  max-height: 280px;
  object-fit: contain;
  display: block;
}

.image-overlay {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  color: #fff;
  font-size: 11px;
  font-weight: 500;
  opacity: 0;
  transition: opacity 0.15s ease;
}

.bubble-image-wrapper:hover .image-overlay {
  opacity: 1;
}

.bubble-video-wrapper {
  max-width: 100%;
  border-radius: 6px;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.15);
  background: #000;
}

.embedded-bubble-video {
  width: 100%;
  max-height: 280px;
  display: block;
}

.attachment-download-bar {
  display: flex;
  align-items: center;
}

.attachment-download-link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 11.5px;
  color: inherit;
  text-decoration: none;
  background: rgba(0, 0, 0, 0.2);
  padding: 4px 10px;
  border-radius: 4px;
  transition: opacity 0.15s;
  max-width: 100%;
}

.attachment-download-link:hover {
  opacity: 0.85;
}

.att-filename {
  max-width: 240px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.att-filesize {
  opacity: 0.8;
  white-space: nowrap;
}

.att-dl-action {
  font-weight: 600;
  white-space: nowrap;
}

/* Область ввода и превью вложений */
.chat-input-area {
  display: flex;
  flex-direction: column;
  background: var(--bg-card, #161b22);
  border-top: 1px solid var(--border, #30363d);
}

.pending-attachments-row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  padding: 8px 12px 0;
}

.pending-attachment-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: var(--bg-secondary, #0d1117);
  border: 1px solid var(--border, #30363d);
  border-radius: 6px;
  padding: 3px 8px;
  font-size: 11px;
  color: var(--text-main, #f0f6fc);
}

.pending-filename {
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.pending-filesize {
  color: var(--text-tertiary, #8b949e);
}

.btn-remove-pending {
  background: transparent;
  border: none;
  color: var(--text-tertiary, #8b949e);
  cursor: pointer;
  padding: 2px;
  display: flex;
  align-items: center;
}

.btn-remove-pending:hover {
  color: #ef4444;
}

.uploading-indicator {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  color: var(--primary, #58a6ff);
}

.chat-input-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  transition: background-color 0.15s;
}

.chat-input-row.is-drag-over {
  background: rgba(88, 166, 255, 0.08);
  outline: 2px dashed var(--primary, #58a6ff);
  outline-offset: -2px;
}

.file-hidden-input {
  display: none;
}

.btn-attach {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  color: var(--text-secondary, #c9d1d9);
  border-radius: var(--radius-sm, 6px);
  cursor: pointer;
  transition: all 0.15s ease;
  flex-shrink: 0;
}

.btn-attach:hover:not(.disabled) {
  background: var(--bg-secondary, #0d1117);
  color: var(--primary, #58a6ff);
}

.btn-attach.disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.chat-readonly-banner {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: var(--bg-secondary, #0d1117);
  border-top: 1px solid var(--border, #30363d);
  color: var(--text-muted, #8b949e);
  font-size: 0.8rem;
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
  flex-shrink: 0;
  vertical-align: middle;
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
