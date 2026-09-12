<template>
  <div class="snapshot-view-root">
    <!-- ВЕРХНИЙ БАР НАВИГАЦИИ -->
    <header class="snapshot-nav-bar">
      <div class="nav-bar-left">
        <button class="btn-nav-back" @click="goBack">
          <ArrowLeft class="icon-sm" />
          <span>{{ t('journal.snapshotModal.backToJournal') || 'К журналу решений' }}</span>
        </button>
        <div class="nav-divider"></div>
        <div class="nav-title-stack">
          <div class="snapshot-badges">
            <span class="badge-camera">
              <Camera class="icon-xs" />
            </span>
            <span class="badge-snapshot-num">
              {{ t('journal.snapshotModal.title') || 'Снимок решения (Snapshot)' }} #{{ requestId }}
            </span>
            <span class="status-badge" :class="statusBadgeClass">
              {{ statusText }}
            </span>
            <span v-if="projectData.build_version" class="version-tag">
              v{{ projectData.build_version }}
            </span>
          </div>
          <h1 class="snapshot-heading">
            {{ projectData.title_ru || projectData.title_en || `Проект #${projectId}` }}
          </h1>
        </div>
      </div>

      <div class="nav-bar-right">
        <button
          v-if="projectId"
          class="btn-open-live"
          @click="goToLiveProject"
        >
          <ExternalLink class="icon-xs" />
          <span>{{ t('journal.snapshotModal.goToCurrent') || 'Перейти к актуальному проекту' }}</span>
        </button>
      </div>
    </header>

    <!-- ИНФОРМАЦИОННАЯ ПЛАШКА ЗАМОРОЗКИ ДАННЫХ -->
    <div class="freeze-banner">
      <Clock class="icon-xs text-primary flex-shrink-0" />
      <span class="freeze-banner-text">
        <strong>Замороженный аудит-снимок:</strong> данные проекта, медиафайлы и переписка зафиксированы на момент вынесения вердикта модератором ({{ formatDateTime(verdictData.resolved_at || snapshotMeta.created_at) }}).
      </span>
    </div>

    <!-- ИНДИКАТОР ЗАГРУЗКИ -->
    <div v-if="loading" class="state-screen">
      <div class="spinner-md"></div>
      <p>{{ t('common.loading') }}</p>
    </div>

    <!-- ОШИБКА -->
    <div v-else-if="error" class="state-screen">
      <AlertTriangle class="icon-lg text-danger" />
      <p>{{ error }}</p>
      <button class="btn-retry" @click="fetchSnapshot">Повторить попытку</button>
    </div>

    <!-- ОСНОВНОЙ РАБОЧИЙ ЛЕЙАУТ (ИДЕНТИЧЕН СТРАНИЦЕ МОДЕРАЦИИ) -->
    <div v-else class="moderator-workspace">
      <!-- ЛЕВАЯ КОЛОНКА: ИНСПЕКТОР СНИМКА -->
      <main class="review-inspector">
        <div class="inspector-scroll">
          <!-- Карточка 0: Шапка проекта (Идентификация) -->
          <div class="card project-identity-card">
            <div class="identity-header">
              <div class="identity-icon-box" @click="iconUrl && openLightbox(iconUrl, mediaData.icon?.file_name || 'icon.png')">
                <img
                  v-if="iconUrl"
                  :src="iconUrl"
                  alt="Icon"
                  class="identity-icon-img"
                />
                <div v-else class="identity-icon-mock">
                  <span>#{{ projectId }}</span>
                </div>
              </div>

              <div class="identity-text-box">
                <div class="identity-top-row">
                  <span class="project-id-tag">Проект #{{ projectId }}</span>
                  <span class="status-badge" :class="statusBadgeClass">
                    {{ statusText }}
                  </span>
                  <span v-if="projectData.build_version" class="version-tag">
                    v{{ projectData.build_version }}
                  </span>
                </div>
                <h1 class="project-main-title">
                  {{ projectData.title_ru || projectData.title_en || `Проект #${projectId}` }}
                </h1>
                <div v-if="projectData.title_en && projectData.title_ru" class="project-sub-title">
                  {{ projectData.title_en }}
                </div>
              </div>
            </div>

            <!-- Метаданные проверки -->
            <div class="identity-meta-grid">
              <div class="meta-item">
                <span class="meta-label">{{ t('moderation.developerColumn') }}</span>
                <span class="meta-value" :title="projectData.owner_id">
                  {{ projectData.developer_name || getUserDisplayName(projectData.owner_id) }}
                </span>
              </div>
              <div class="meta-item">
                <span class="meta-label">{{ t('moderation.submittedColumn') }}</span>
                <span class="meta-value">
                  {{ formatDateTime(verdictData.submitted_at) }}
                </span>
              </div>
              <div class="meta-item">
                <span class="meta-label">{{ t('moderation.moderator') }}</span>
                <span class="meta-value" :title="verdictData.moderator_id">
                  {{ verdictData.moderator_id ? getUserDisplayName(verdictData.moderator_id) : '—' }}
                </span>
              </div>
            </div>
          </div>

          <!-- Карточка 1: Основная информация (RU / EN) -->
          <div class="card section-card">
            <div class="section-head">
              <h3>{{ t('moderation.basicInfo') }}</h3>
            </div>

            <!-- Названия -->
            <div class="data-row">
              <div class="data-group">
                <label class="data-label">{{ t('projectDraft.gameTitleRu') }}</label>
                <div class="data-box">{{ projectData.title_ru || '—' }}</div>
              </div>
              <div class="data-group">
                <label class="data-label">{{ t('projectDraft.gameTitleEn') }}</label>
                <div class="data-box">{{ projectData.title_en || '—' }}</div>
              </div>
            </div>

            <!-- SEO описания -->
            <div class="data-row">
              <div class="data-group">
                <label class="data-label">SEO (RU)</label>
                <div class="data-box multiline">{{ projectData.seo_ru || '—' }}</div>
              </div>
              <div class="data-group">
                <label class="data-label">SEO (EN)</label>
                <div class="data-box multiline">{{ projectData.seo_en || '—' }}</div>
              </div>
            </div>

            <!-- Описание игры -->
            <div class="data-row">
              <div class="data-group">
                <label class="data-label">{{ t('projectDraft.gameDescriptionRu') }}</label>
                <div class="data-box multiline-lg">
                  {{ projectData.about_ru || t('moderation.noDescription') }}
                </div>
              </div>
              <div class="data-group">
                <label class="data-label">{{ t('projectDraft.gameDescriptionEn') }}</label>
                <div class="data-box multiline-lg">
                  {{ projectData.about_en || t('moderation.noDescription') }}
                </div>
              </div>
            </div>
          </div>

          <!-- Карточка 2: Медиа-материалы (со слепками и контрольными суммами) -->
          <div class="card section-card">
            <div class="section-head">
              <h3>{{ t('moderation.mediaMaterials') }}</h3>
            </div>

            <div class="media-inspection-grid">
              <!-- ИКОНКА ИГРЫ (512x512) -->
              <div class="media-box-slot">
                <div class="slot-title-row">
                  <span class="slot-label">{{ t('projectDraft.gameIcon') }}</span>
                  <span class="slot-spec">512×512 PNG</span>
                </div>
                <div class="media-view-panel">
                  <div
                    v-if="iconUrl"
                    class="img-preview-wrap icon-aspect clickable"
                    @click="openLightbox(iconUrl, mediaData.icon?.file_name || 'icon.png')"
                  >
                    <img :src="iconUrl" alt="Icon" class="preview-img" />
                  </div>
                  <div v-else class="media-empty-placeholder">
                    <ImageIcon class="icon-md text-muted" />
                    <span>{{ t('moderation.noMedia') }}</span>
                  </div>
                </div>
                <div v-if="mediaData.icon?.sha256" class="media-meta-bar" :title="`SHA256: ${mediaData.icon.sha256}`">
                  <ShieldCheck class="icon-xs text-success" />
                  <span class="hash-text">SHA256: {{ formatShortHash(mediaData.icon.sha256) }}</span>
                </div>
              </div>

              <!-- ОБЛОЖКА ИГРЫ (800x470) -->
              <div class="media-box-slot">
                <div class="slot-title-row">
                  <span class="slot-label">{{ t('projectDraft.coverMain') }}</span>
                  <span class="slot-spec">800×470 PNG</span>
                </div>
                <div class="media-view-panel">
                  <div
                    v-if="coverUrl"
                    class="img-preview-wrap cover-aspect clickable"
                    @click="openLightbox(coverUrl, mediaData.cover?.file_name || 'cover.png')"
                  >
                    <img :src="coverUrl" alt="Cover" class="preview-img" />
                  </div>
                  <div v-else class="media-empty-placeholder">
                    <ImageIcon class="icon-md text-muted" />
                    <span>{{ t('moderation.noMedia') }}</span>
                  </div>
                </div>
                <div v-if="mediaData.cover?.sha256" class="media-meta-bar" :title="`SHA256: ${mediaData.cover.sha256}`">
                  <ShieldCheck class="icon-xs text-success" />
                  <span class="hash-text">SHA256: {{ formatShortHash(mediaData.cover.sha256) }}</span>
                </div>
              </div>

              <!-- ПРОМО-ВИДЕО (≤ 12 МБ, MP4) -->
              <div class="media-box-slot">
                <div class="slot-title-row">
                  <span class="slot-label">{{ t('projectDraft.promoVideo') }}</span>
                  <span class="slot-spec">≤ 12 МБ, MP4</span>
                </div>
                <div class="media-view-panel">
                  <div v-if="videoUrl" class="img-preview-wrap video-aspect">
                    <video :src="videoUrl" controls class="preview-video"></video>
                  </div>
                  <div v-else class="media-empty-placeholder">
                    <VideoIcon class="icon-md text-muted" />
                    <span>{{ t('moderation.noMedia') }}</span>
                  </div>
                </div>
                <div v-if="mediaData.video?.sha256" class="media-meta-bar" :title="`SHA256: ${mediaData.video.sha256}`">
                  <ShieldCheck class="icon-xs text-success" />
                  <span class="hash-text">SHA256: {{ formatShortHash(mediaData.video.sha256) }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- Карточка 3: Сборка и тестирование -->
          <div class="card section-card">
            <div class="section-head">
              <h3>{{ t('moderation.buildTesting') }}</h3>
            </div>

            <div class="build-test-row">
              <div class="build-info-block">
                <span class="build-version-tag">
                  Версия сборки снимка: <strong>v{{ projectData.build_version || '1.0.0' }}</strong>
                </span>
                <p class="build-desc">
                  Зафиксированный билд на момент аудита. Доступен для запуска и воспроизведения поведения игры.
                </p>
              </div>

              <a
                v-if="projectData.dev_url"
                :href="projectData.dev_url"
                target="_blank"
                rel="noopener noreferrer"
                class="btn-play-dev-lg"
              >
                <Gamepad2 class="icon-sm" />
                <span>{{ t('journal.snapshotModal.openDevBuild') || 'Запустить Dev-билд снимка' }}</span>
                <ExternalLink class="icon-xs" />
              </a>
            </div>
          </div>

          <!-- Карточка 4: Вердикт модератора и нарушения -->
          <div class="card verdict-card">
            <div class="section-head">
              <h3>{{ t('moderation.verdictSection') }}</h3>
            </div>

            <!-- Вердикт: Отклонен -->
            <div v-if="isRejected" class="result-box box-rejected">
              <div class="result-header">
                <AlertTriangle class="icon-md text-danger flex-shrink-0" />
                <div class="result-text">
                  <div class="result-title-row">
                    <strong>{{ t('moderation.rejected') }}</strong>
                    <span class="verdict-time-tag">Проверено за: {{ verdictData.review_duration || '—' }}</span>
                  </div>
                  <p v-if="verdictData.rejection_reason" class="reason-summary">
                    <strong>Официальное обоснование:</strong> {{ verdictData.rejection_reason }}
                  </p>
                </div>
              </div>

              <!-- Список выявленных нарушений правил -->
              <div v-if="verdictData.violations && verdictData.violations.length" class="violations-block">
                <div class="violations-heading">
                  <span>Выявленные нарушения регламента ({{ verdictData.violations.length }}):</span>
                </div>

                <div
                  v-for="(v, idx) in verdictData.violations"
                  :key="idx"
                  class="violation-item-card"
                >
                  <div class="violation-top">
                    <span class="v-rule-pill">{{ v.rule_code }}</span>
                    <span class="v-title">{{ v.rule_title }}</span>
                  </div>
                  <p class="v-desc">{{ v.description }}</p>

                  <!-- Доказательства нарушения -->
                  <div v-if="v.attachments && v.attachments.length" class="v-proofs-grid">
                    <div
                      v-for="att in v.attachments"
                      :key="att.id"
                      class="v-proof-card"
                    >
                      <img
                        v-if="isImgAttachment(att)"
                        :src="getAttachmentPreview(att)"
                        :alt="att.file_name"
                        class="proof-thumb"
                        @click="openLightbox(getAttachmentPreview(att), att.file_name)"
                      />
                      <video
                        v-else-if="isVideoAttachment(att)"
                        :src="getAttachmentPreview(att)"
                        controls
                        class="proof-video"
                      ></video>
                      <span class="proof-name">{{ att.file_name }}</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Вердикт: Одобрен -->
            <div v-else-if="isApproved" class="result-box box-approved">
              <div class="result-header">
                <CheckCircle2 class="icon-md text-success flex-shrink-0" />
                <div class="result-text">
                  <div class="result-title-row">
                    <strong>{{ t('moderation.approved') }}</strong>
                    <span class="verdict-time-tag">Проверено за: {{ verdictData.review_duration || '—' }}</span>
                  </div>
                  <p>Проект одобрен и опубликован в основном каталоге платформы.</p>
                  <p v-if="verdictData.comment" class="comment-line">
                    <strong>Комментарий модератора:</strong> {{ verdictData.comment }}
                  </p>
                </div>
              </div>
            </div>

            <!-- Вердикт: Отозван / Отменен -->
            <div v-else class="result-box box-cancelled">
              <Info class="icon-md text-muted flex-shrink-0" />
              <div class="result-text">
                <strong>{{ statusText }}</strong>
                <p>Заявка была отозвана разработчиком или отменена администратором.</p>
              </div>
            </div>
          </div>
        </div>
      </main>

      <!-- ПРАВАЯ КОЛОНКА: СРЕЗ ПЕРЕПИСКИ ТИКЕТА -->
      <aside class="moderator-chat-aside">
        <div class="snapshot-chat-container">
          <!-- Шапка чата -->
          <div class="chat-header">
            <div class="chat-header-left">
              <MessageSquare class="icon-sm text-primary" />
              <h3 class="chat-title">{{ t('moderation.projectChatTitle') }}</h3>
              <span class="status-pill-sm pill-frozen">
                Архив тикета (Заморожен)
              </span>
            </div>
          </div>

          <!-- Список замороженных сообщений -->
          <div class="chat-messages-scroll">
            <div v-if="!chatMessages.length" class="chat-empty-state">
              <MessageSquareDashed class="icon-lg text-muted" />
              <p>В этом тикете нет сохранённых сообщений</p>
            </div>

            <div v-else class="messages-stack">
              <div
                v-for="msg in chatMessages"
                :key="msg.id"
                class="message-row"
                :class="{
                  'is-system': isSystemMsg(msg),
                  'is-moderator': !isSystemMsg(msg) && isModMsg(msg),
                  'is-developer': !isSystemMsg(msg) && !isModMsg(msg),
                }"
              >
                <!-- Системное событие (строго по центру в капсуле) -->
                <div v-if="isSystemMsg(msg)" class="system-event-wrapper">
                  <div class="system-event-pill">
                    <span class="system-badge">Система</span>
                    <span class="system-text">{{ msg.content }}</span>
                    <span class="system-time">{{ formatTime(msg.created_at) }}</span>
                  </div>
                </div>

                <!-- Сообщение пользователя / модератора -->
                <div v-else class="chat-bubble">
                  <div class="bubble-header">
                    <span class="author-name">{{ msg.sender_name || (isModMsg(msg) ? 'Модератор' : 'Разработчик') }}</span>
                    <span class="bubble-time">{{ formatTime(msg.created_at) }}</span>
                  </div>
                  <div v-if="msg.content" class="bubble-body">{{ msg.content }}</div>

                  <!-- Вложения сообщения -->
                  <div v-if="msg.attachments && msg.attachments.length" class="bubble-attachments-stack">
                    <div
                      v-for="att in msg.attachments"
                      :key="att.id"
                      class="bubble-attachment-item"
                    >
                      <!-- Видео -->
                      <div v-if="isVideoAttachment(att)" class="bubble-video-wrapper">
                        <video
                          :src="getAttachmentPreview(att)"
                          controls
                          class="embedded-bubble-video"
                        ></video>
                      </div>

                      <!-- Изображение -->
                      <div
                        v-else
                        class="bubble-image-wrapper"
                        @click="openLightbox(getAttachmentPreview(att), att.file_name)"
                      >
                        <img
                          :src="getAttachmentPreview(att)"
                          :alt="att.file_name"
                          class="embedded-bubble-image"
                          loading="lazy"
                        />
                      </div>

                      <!-- Скачивание -->
                      <div class="attachment-download-bar">
                        <a
                          :href="getAttachmentDownloadUrl(att)"
                          :download="att.file_name || 'attachment'"
                          class="attachment-download-link"
                          target="_blank"
                          rel="noopener noreferrer"
                        >
                          <Paperclip class="icon-xs" />
                          <span class="att-filename">{{ att.file_name }}</span>
                          <span v-if="att.file_size" class="att-filesize">({{ formatBytes(att.file_size) }})</span>
                          <span class="att-dl-action">• Скачать</span>
                        </a>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Нижняя плашка заблокированного чата -->
          <div class="chat-locked-bar">
            <Lock class="icon-xs text-muted" />
            <span>Переписка заморожена в рамках аудиторского снимка решения</span>
          </div>
        </div>
      </aside>
    </div>

    <!-- ПОЛНОЭКРАННЫЙ LIGHTBOX ДЛЯ ПРОСМОТРА ИЗОБРАЖЕНИЙ -->
    <MediaLightboxModal
      v-if="lightboxData"
      :src="lightboxData.src"
      :file-name="lightboxData.fileName"
      :download-url="lightboxData.downloadUrl"
      @close="closeLightbox"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import {
  ArrowLeft,
  Camera,
  ExternalLink,
  Clock,
  AlertTriangle,
  CheckCircle2,
  Info,
  Gamepad2,
  MessageSquare,
  MessageSquareDashed,
  Paperclip,
  Lock,
  Image as ImageIcon,
  Video as VideoIcon,
  ShieldCheck,
} from 'lucide-vue-next';
import {
  moderationApi,
  getStatusText,
  getStatusBadgeClass,
  formatDateTime,
} from '@/entities/moderation';
import { getUserDisplayName } from '@/entities/user';
import { getMediaUrl } from '@/entities/project';
import MediaLightboxModal from '@/entities/moderation/ui/MediaLightboxModal.vue';

const route = useRoute();
const router = useRouter();
const { t } = useI18n();

const requestId = computed(() => Number(route.params.requestId));

const loading = ref(true);
const error = ref(null);
const snapshotMeta = ref({});
const projectData = ref({});
const mediaData = ref({});
const verdictData = ref({});
const chatMessages = ref([]);
const lightboxData = ref(null);

const projectId = computed(() => {
  return projectData.value.id || snapshotMeta.value.projectId || snapshotMeta.value.project_id || '';
});

const status = computed(() => {
  return verdictData.value.status ?? snapshotMeta.value.status ?? 0;
});

const statusText = computed(() => {
  return getStatusText(status.value);
});

const statusBadgeClass = computed(() => {
  return getStatusBadgeClass(status.value);
});

const isApproved = computed(() => {
  const s = status.value;
  return s === 3 || s === 'REQUEST_STATUS_APPROVED' || s === 'approved';
});

const isRejected = computed(() => {
  const s = status.value;
  return s === 4 || s === 'REQUEST_STATUS_REJECTED' || s === 'rejected';
});

// Вычисление URL медиа с поддержкой превью из снимка и fallback на original_url
const iconUrl = computed(() => {
  const m = mediaData.value.icon;
  if (!m) return '';
  if (m.thumbnail_data) return m.thumbnail_data;
  return getMediaUrl(m.original_url);
});

const coverUrl = computed(() => {
  const m = mediaData.value.cover;
  if (!m) return '';
  if (m.thumbnail_data) return m.thumbnail_data;
  return getMediaUrl(m.original_url);
});

const videoUrl = computed(() => {
  const m = mediaData.value.video;
  if (!m) return '';
  return getMediaUrl(m.original_url);
});

function goBack() {
  router.push('/moderator/journal');
}

function goToLiveProject() {
  if (projectId.value) {
    router.push(`/moderator/projects/${projectId.value}`);
  }
}

function openLightbox(src, fileName = '', downloadUrl = '') {
  if (!src) return;
  lightboxData.value = {
    src,
    fileName,
    downloadUrl: downloadUrl || src,
  };
}

function closeLightbox() {
  lightboxData.value = null;
}

function isSystemMsg(msg) {
  const role = msg.sender_role ?? msg.senderRole;
  const type = msg.message_type ?? msg.messageType;
  return role === 3 || role === 'SENDER_ROLE_SYSTEM' || type === 2 || type === 3 || type === 4 || type === 5;
}

function isModMsg(msg) {
  const role = msg.sender_role ?? msg.senderRole;
  return role === 2 || role === 'SENDER_ROLE_MODERATOR';
}

function isImgAttachment(att) {
  const mime = (att.mime_type || att.mimeType || '').toLowerCase();
  const name = (att.file_name || att.fileName || '').toLowerCase();
  return mime.startsWith('image/') || name.endsWith('.png') || name.endsWith('.jpg') || name.endsWith('.jpeg') || name.endsWith('.webp');
}

function isVideoAttachment(att) {
  const mime = (att.mime_type || att.mimeType || '').toLowerCase();
  const name = (att.file_name || att.fileName || '').toLowerCase();
  return mime.startsWith('video/') || name.endsWith('.mp4') || name.endsWith('.webm');
}

function getAttachmentPreview(att) {
  if (att.thumbnail_data) return att.thumbnail_data;
  if (att.url) return att.url;
  return `/api/v1/projects/${projectId.value}/chat/attachments/${att.id}`;
}

function getAttachmentDownloadUrl(att) {
  if (att.download_url) return att.download_url;
  return `/api/v1/projects/${projectId.value}/chat/attachments/${att.id}/download`;
}

function formatShortHash(hash) {
  if (!hash || hash.length < 12) return hash || '—';
  return `${hash.slice(0, 8)}...${hash.slice(-6)}`;
}

function formatBytes(bytes) {
  if (!bytes || bytes <= 0) return '0 Б';
  const k = 1024;
  const sizes = ['Б', 'КБ', 'МБ', 'ГБ'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}

function formatTime(val) {
  if (!val) return '';
  try {
    const d = new Date(val);
    if (isNaN(d.getTime())) return '';
    return d.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' });
  } catch {
    return '';
  }
}

async function fetchSnapshot() {
  loading.value = true;
  error.value = null;
  try {
    const res = await moderationApi.getSnapshot(requestId.value);
    snapshotMeta.value = res;
    const payload = res.payload || {};
    projectData.value = payload.project || {};
    mediaData.value = payload.media || {};
    verdictData.value = payload.verdict || {};
    chatMessages.value = payload.chat_transcript || [];
  } catch (err) {
    console.error('Failed to load snapshot:', err);
    error.value = err.response?.data?.message || err.message || 'Ошибка загрузки аудит-снимка';
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  fetchSnapshot();
});
</script>

<style scoped>
.snapshot-view-root {
  display: flex;
  flex-direction: column;
  width: 100%;
  min-height: calc(100vh - 60px);
  background: var(--bg-app, #0d1117);
  box-sizing: border-box;
}

/* ВЕРХНИЙ БАР НАВИГАЦИИ */
.snapshot-nav-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 32px;
  background: var(--bg-card, #161b22);
  border-bottom: 1px solid var(--border, #30363d);
  gap: 16px;
  flex-wrap: wrap;
}

.nav-bar-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.btn-nav-back {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: var(--bg-secondary, #0d1117);
  border: 1px solid var(--border, #30363d);
  color: var(--text-main, #f0f6fc);
  padding: 6px 12px;
  border-radius: var(--radius-sm, 6px);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-nav-back:hover {
  background: var(--bg-tertiary, #21262d);
  border-color: var(--border-hover, #8b949e);
}

.nav-divider {
  width: 1px;
  height: 28px;
  background: var(--border, #30363d);
}

.nav-title-stack {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.snapshot-badges {
  display: flex;
  align-items: center;
  gap: 8px;
}

.badge-camera {
  color: var(--primary, #58a6ff);
  display: flex;
  align-items: center;
}

.badge-snapshot-num {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-muted, #8b949e);
}

.snapshot-heading {
  margin: 0;
  font-size: 18px;
  font-weight: 700;
  color: var(--text-main, #f0f6fc);
}

.btn-open-live {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: var(--primary, #58a6ff);
  color: #ffffff;
  border: none;
  padding: 8px 16px;
  border-radius: var(--radius-sm, 6px);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.15s ease;
}

.btn-open-live:hover {
  opacity: 0.9;
}

/* ЗАМОРОЖЕННЫЙ БАННЕР */
.freeze-banner {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 32px;
  background: rgba(56, 139, 253, 0.08);
  border-bottom: 1px solid rgba(56, 139, 253, 0.2);
  color: var(--text-main, #f0f6fc);
  font-size: 13px;
}

/* РАБОЧАЯ ОБЛАСТЬ (ИДЕНТИЧНА СТРАНИЦЕ МОДЕРАЦИИ) */
.moderator-workspace {
  display: flex;
  width: 100%;
  flex: 1;
  overflow: hidden;
  box-sizing: border-box;
}

.review-inspector {
  flex: 1;
  min-width: 0;
  height: calc(100vh - 120px);
  overflow-y: auto;
  padding: 24px 32px 48px;
  box-sizing: border-box;
}

.inspector-scroll {
  max-width: 900px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.card {
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-md, 8px);
  padding: 20px;
  box-sizing: border-box;
}

/* КАРТОЧКА ИДЕНТИФИКАЦИИ */
.project-identity-card {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.identity-header {
  display: flex;
  align-items: center;
  gap: 16px;
}

.identity-icon-box {
  width: 60px;
  height: 60px;
  border-radius: var(--radius-md, 8px);
  background: var(--bg-tertiary, #21262d);
  border: 1px solid var(--border, #30363d);
  overflow: hidden;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.identity-icon-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.identity-icon-mock {
  color: var(--text-tertiary, #8b949e);
  font-size: 12px;
  font-weight: 600;
}

.identity-text-box {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.identity-top-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.project-id-tag {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-tertiary, #8b949e);
}

.version-tag {
  font-size: 11px;
  font-weight: 600;
  padding: 1px 6px;
  border-radius: 4px;
  background: var(--bg-secondary, #21262d);
  color: var(--text-muted, #b0b8c4);
  border: 1px solid var(--border, #30363d);
}

.project-main-title {
  margin: 0;
  font-size: 18px;
  font-weight: 700;
  color: var(--text-main, #f0f6fc);
}

.project-sub-title {
  font-size: 13px;
  color: var(--text-tertiary, #8b949e);
}

.identity-meta-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  padding-top: 14px;
  border-top: 1px solid var(--border, #21262d);
}

.meta-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.meta-label {
  font-size: 11px;
  font-weight: 500;
  color: var(--text-tertiary, #8b949e);
}

.meta-value {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-main, #f0f6fc);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* СЕКЦИОННЫЕ КАРТОЧКИ */
.section-head {
  margin-bottom: 16px;
}

.section-head h3 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
}

.data-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  margin-bottom: 14px;
}

.data-row:last-child {
  margin-bottom: 0;
}

.data-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.data-label {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-tertiary, #8b949e);
}

.data-box {
  padding: 8px 12px;
  background: var(--bg-secondary, #0d1117);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  color: var(--text-main, #f0f6fc);
  font-size: 13px;
  min-height: 34px;
  display: flex;
  align-items: center;
  box-sizing: border-box;
}

.data-box.multiline {
  min-height: 52px;
  align-items: flex-start;
  line-height: 1.4;
  white-space: pre-wrap;
}

.data-box.multiline-lg {
  min-height: 90px;
  align-items: flex-start;
  line-height: 1.4;
  white-space: pre-wrap;
}

/* МЕДИА СЛОТЫ */
.media-inspection-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 16px;
}

.media-box-slot {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.slot-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.slot-label {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-main, #f0f6fc);
}

.slot-spec {
  font-size: 11px;
  color: var(--text-tertiary, #8b949e);
}

.media-view-panel {
  background: var(--bg-secondary, #0d1117);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  padding: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 140px;
}

.img-preview-wrap {
  border-radius: var(--radius-sm, 6px);
  overflow: hidden;
  border: 1px solid var(--border, #30363d);
  background: var(--bg-tertiary, #21262d);
  display: flex;
  align-items: center;
  justify-content: center;
}

.img-preview-wrap.clickable {
  cursor: pointer;
  transition: transform 0.15s ease;
}

.img-preview-wrap.clickable:hover {
  transform: scale(1.02);
}

.icon-aspect {
  width: 100px;
  height: 100px;
}

.cover-aspect {
  width: 100%;
  max-width: 200px;
  aspect-ratio: 800 / 470;
}

.video-aspect {
  width: 100%;
  max-width: 240px;
  aspect-ratio: 16 / 9;
}

.preview-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.preview-video {
  width: 100%;
  height: 100%;
  object-fit: contain;
  background: #000000;
  border-radius: 4px;
}

.media-empty-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: var(--text-tertiary, #8b949e);
  font-size: 12px;
}

.media-meta-bar {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  color: var(--text-tertiary, #8b949e);
  padding: 2px 4px;
}

.hash-text {
  font-family: monospace;
}

/* СБОРКА И ТЕСТИРОВАНИЕ */
.build-test-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
}

.build-info-block {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.build-version-tag {
  font-size: 13px;
  color: var(--text-main, #f0f6fc);
}

.build-desc {
  margin: 0;
  font-size: 13px;
  color: var(--text-tertiary, #8b949e);
  line-height: 1.4;
}

.btn-play-dev-lg {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 18px;
  background: var(--bg-tertiary, #21262d);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  color: var(--text-main, #f0f6fc);
  font-size: 13px;
  font-weight: 600;
  text-decoration: none;
  flex-shrink: 0;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-play-dev-lg:hover {
  background: var(--border, #30363d);
  border-color: var(--border-hover, #8b949e);
}

/* КАРТОЧКА ВЕРДИКТА */
.verdict-card {
  border-left: 4px solid var(--border, #30363d);
}

.result-box {
  padding: 16px;
  border-radius: var(--radius-sm, 6px);
}

.box-rejected {
  background: rgba(239, 68, 68, 0.08);
  border: 1px solid rgba(239, 68, 68, 0.25);
}

.box-approved {
  background: rgba(34, 197, 94, 0.08);
  border: 1px solid rgba(34, 197, 94, 0.25);
}

.box-cancelled {
  background: rgba(139, 148, 158, 0.08);
  border: 1px solid rgba(139, 148, 158, 0.25);
}

.result-header {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.result-text {
  flex: 1;
}

.result-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 6px;
}

.result-text strong {
  font-size: 15px;
  color: var(--text-main, #f0f6fc);
}

.result-text p {
  margin: 4px 0 0;
  font-size: 13px;
  color: var(--text-muted, #8b949e);
  line-height: 1.4;
}

.verdict-time-tag {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-tertiary, #8b949e);
}

.reason-summary {
  color: var(--text-main, #f0f6fc) !important;
}

.violations-block {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid rgba(239, 68, 68, 0.2);
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.violations-heading {
  font-size: 13px;
  font-weight: 600;
  color: #ef4444;
}

.violation-item-card {
  background: rgba(239, 68, 68, 0.05);
  border: 1px solid rgba(239, 68, 68, 0.2);
  border-radius: var(--radius-sm, 6px);
  padding: 12px;
}

.violation-top {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}

.v-rule-pill {
  font-size: 11px;
  font-weight: 700;
  background: rgba(239, 68, 68, 0.2);
  color: #ef4444;
  padding: 2px 6px;
  border-radius: 4px;
}

.v-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
}

.v-desc {
  margin: 0;
  font-size: 12px;
  color: var(--text-muted, #8b949e);
  line-height: 1.4;
}

.v-proofs-grid {
  display: flex;
  gap: 10px;
  margin-top: 8px;
  flex-wrap: wrap;
}

.v-proof-card {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.proof-thumb {
  width: 100px;
  height: 70px;
  object-fit: cover;
  border-radius: 4px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  cursor: pointer;
}

.proof-video {
  width: 140px;
  height: 80px;
  border-radius: 4px;
}

.proof-name {
  font-size: 10px;
  color: var(--text-tertiary, #8b949e);
  max-width: 100px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* ПРАВАЯ КОЛОНКА: ЧАТ */
.moderator-chat-aside {
  width: 480px;
  min-width: 400px;
  max-width: 560px;
  flex-shrink: 0;
  height: calc(100vh - 120px);
  border-left: 1px solid var(--border, #30363d);
  background: var(--bg-card, #161b22);
}

.snapshot-chat-container {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.chat-header {
  padding: 16px 20px;
  border-bottom: 1px solid var(--border, #30363d);
}

.chat-header-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.chat-title {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
}

.pill-frozen {
  font-size: 11px;
  font-weight: 600;
  background: rgba(88, 166, 255, 0.15);
  color: var(--primary, #58a6ff);
  padding: 2px 8px;
  border-radius: 10px;
}

.chat-messages-scroll {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
}

.chat-empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  gap: 10px;
  color: var(--text-tertiary, #8b949e);
  font-size: 13px;
}

.messages-stack {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.message-row {
  display: flex;
  flex-direction: column;
}

.is-developer {
  align-items: flex-start;
}

.is-moderator {
  align-items: flex-end;
}

.system-event-wrapper {
  display: flex;
  justify-content: center;
  width: 100%;
  margin: 8px 0;
}

.system-event-pill {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: var(--bg-secondary, #0d1117);
  border: 1px solid var(--border, #30363d);
  padding: 5px 12px;
  border-radius: 16px;
  font-size: 12px;
}

.system-badge {
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  background: var(--border, #30363d);
  color: var(--text-tertiary, #8b949e);
  padding: 1px 5px;
  border-radius: 4px;
}

.system-text {
  color: var(--text-main, #f0f6fc);
}

.system-time {
  font-size: 10px;
  color: var(--text-tertiary, #8b949e);
}

.chat-bubble {
  max-width: 85%;
  padding: 10px 14px;
  border-radius: var(--radius-md, 8px);
  background: var(--bg-secondary, #0d1117);
  border: 1px solid var(--border, #30363d);
}

.is-moderator .chat-bubble {
  background: rgba(56, 139, 253, 0.15);
  border-color: rgba(56, 139, 253, 0.3);
}

.bubble-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 4px;
}

.author-name {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
}

.bubble-time {
  font-size: 10px;
  color: var(--text-tertiary, #8b949e);
}

.bubble-body {
  font-size: 13px;
  color: var(--text-main, #f0f6fc);
  line-height: 1.4;
  white-space: pre-wrap;
}

.bubble-attachments-stack {
  margin-top: 8px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.bubble-image-wrapper {
  max-width: 280px;
  border-radius: 6px;
  overflow: hidden;
  border: 1px solid var(--border, #30363d);
  cursor: pointer;
}

.embedded-bubble-image {
  width: 100%;
  display: block;
}

.bubble-video-wrapper {
  max-width: 300px;
}

.embedded-bubble-video {
  width: 100%;
  border-radius: 6px;
}

.attachment-download-bar {
  margin-top: 4px;
}

.attachment-download-link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  color: var(--text-muted, #8b949e);
  text-decoration: none;
}

.attachment-download-link:hover {
  color: var(--primary, #58a6ff);
}

.chat-locked-bar {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 12px 16px;
  background: var(--bg-secondary, #0d1117);
  border-top: 1px solid var(--border, #30363d);
  font-size: 12px;
  color: var(--text-tertiary, #8b949e);
}

.state-screen {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 400px;
  gap: 12px;
  color: var(--text-muted, #8b949e);
}

.btn-retry {
  padding: 6px 14px;
  background: var(--primary, #58a6ff);
  color: #fff;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}
</style>
