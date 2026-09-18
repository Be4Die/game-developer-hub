<template>
  <div class="verdict-card">
    <div class="verdict-header">
      <div class="verdict-header-left">
        <AlertOctagon class="icon-sm text-danger" />
        <span class="verdict-title">Заявка отклонена модератором</span>
      </div>
      <span class="verdict-time">{{ formattedTime }}</span>
    </div>

    <!-- Общая вводная часть / причина -->
    <div v-if="leadText" class="verdict-lead">
      {{ leadText }}
    </div>

    <!-- Список нарушений правил -->
    <div v-if="parsedViolations.length" class="violations-list">
      <div
        v-for="(item, idx) in parsedViolations"
        :key="idx"
        class="violation-item"
      >
        <div class="violation-head">
          <router-link
            v-if="item.rule_code || item.ruleCode"
            :to="`/docs/rules#${(item.rule_code || item.ruleCode).toUpperCase()}`"
            class="rule-badge-link"
            :title="`Открыть регламент: Пункт ${item.rule_code || item.ruleCode}`"
          >
            <BookOpen class="icon-xs" />
            <span>Пункт {{ item.rule_code || item.ruleCode }}</span>
            <ExternalLink class="icon-xs link-arrow" />
          </router-link>
          <span v-if="item.rule_title || item.ruleTitle" class="rule-title">
            «{{ item.rule_title || item.ruleTitle }}»
          </span>
        </div>

        <div class="violation-desc">
          {{ item.description }}
        </div>

        <!-- Прикрепленные доказательства к конкретному пункту -->
        <div
          v-if="item.attachments && item.attachments.length"
          class="violation-media-stack"
        >
          <div
            v-for="att in item.attachments"
            :key="att.id"
            class="media-container"
          >
            <!-- Видео -->
            <div v-if="isVideo(att)" class="video-wrapper">
              <video
                :src="getMediaUrl(att)"
                controls
                preload="metadata"
                class="embedded-video"
              ></video>
            </div>

            <!-- Изображение -->
            <div
              v-else-if="isImage(att)"
              class="image-wrapper"
              @click="openLightbox(att)"
            >
              <img
                :src="getMediaUrl(att)"
                :alt="att.file_name || att.fileName"
                class="embedded-image"
                loading="lazy"
              />
              <div class="image-overlay">
                <Maximize2 class="icon-xs" />
                <span>Увеличить</span>
              </div>
            </div>

            <!-- Строка со ссылкой на скачивание -->
            <div class="media-download-row">
              <a
                :href="getDownloadUrl(att)"
                :download="att.file_name || att.fileName || 'attachment'"
                class="media-download-link"
                target="_blank"
                rel="noopener noreferrer"
              >
                <Paperclip class="icon-xs" />
                <span class="file-name">{{ att.file_name || att.fileName }}</span>
                <span class="file-size">({{ formatSize(att.file_size || att.fileSize) }})</span>
                <span class="download-action">• Скачать оригинал</span>
              </a>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Резервный вывод текста, если нет структурированных нарушений -->
    <div v-else-if="fallbackContent" class="verdict-fallback">
      {{ fallbackContent }}
    </div>

    <!-- Общие вложения сообщения (если прикреплены к вердикту целиком) -->
    <div v-if="rootAttachments.length" class="root-attachments-stack">
      <div
        v-for="att in rootAttachments"
        :key="att.id"
        class="media-container"
      >
        <div v-if="isVideo(att)" class="video-wrapper">
          <video
            :src="getMediaUrl(att)"
            controls
            preload="metadata"
            class="embedded-video"
          ></video>
        </div>
        <div
          v-else-if="isImage(att)"
          class="image-wrapper"
          @click="openLightbox(att)"
        >
          <img
            :src="getMediaUrl(att)"
            :alt="att.file_name || att.fileName"
            class="embedded-image"
            loading="lazy"
          />
          <div class="image-overlay">
            <Maximize2 class="icon-xs" />
            <span>Увеличить</span>
          </div>
        </div>
        <div class="media-download-row">
          <a
            :href="getDownloadUrl(att)"
            :download="att.file_name || att.fileName || 'attachment'"
            class="media-download-link"
            target="_blank"
            rel="noopener noreferrer"
          >
            <Paperclip class="icon-xs" />
            <span class="file-name">{{ att.file_name || att.fileName }}</span>
            <span class="file-size">({{ formatSize(att.file_size || att.fileSize) }})</span>
            <span class="download-action">• Скачать оригинал</span>
          </a>
        </div>
      </div>
    </div>

    <!-- Общий комментарий модератора -->
    <div v-if="generalComment" class="verdict-footer">
      <MessageSquare class="icon-xs text-muted" />
      <span class="footer-label">Комментарий модератора:</span>
      <span class="footer-text">{{ generalComment }}</span>
    </div>

    <!-- Модальное окно просмотра картинок -->
    <MediaLightboxModal
      v-if="lightboxMedia"
      :src="getMediaUrl(lightboxMedia)"
      :file-name="lightboxMedia.file_name || lightboxMedia.fileName"
      :download-url="getDownloadUrl(lightboxMedia)"
      @close="lightboxMedia = null"
    />
  </div>
</template>

<script setup>
import { ref, computed } from 'vue';
import { AlertOctagon, Maximize2, Paperclip, MessageSquare, BookOpen, ExternalLink } from 'lucide-vue-next';
import { formatDateTime } from '../model/helpers';
import { formatBytes } from '@/shared/lib/mediaCompressor';
import MediaLightboxModal from './MediaLightboxModal.vue';

const props = defineProps({
  message: {
    type: Object,
    required: true,
  },
});

const lightboxMedia = ref(null);

const formattedTime = computed(() => {
  return formatDateTime(props.message.created_at || props.message.createdAt);
});

// Извлечение полезной нагрузки (JSON)
const payload = computed(() => {
  let p = props.message.payload;
  if (!p && props.message.payload_json) {
    try {
      p = JSON.parse(props.message.payload_json);
    } catch {
      p = {};
    }
  }
  return p || {};
});

const parsedViolations = computed(() => {
  if (Array.isArray(payload.value.violations) && payload.value.violations.length > 0) {
    return payload.value.violations;
  }
  return [];
});

const leadText = computed(() => {
  if (parsedViolations.value.length > 0) {
    return 'Проект отправлен на доработку. При проверке обнаружены следующие замечания:';
  }
  return '';
});

const generalComment = computed(() => {
  return (
    payload.value.general_comment ||
    payload.value.generalComment ||
    payload.value.comment ||
    ''
  );
});

const fallbackContent = computed(() => {
  const content = props.message.content || '';
  return content.replace(/^Заявка отклонена:\s*/i, '').trim();
});

// Вложения, прикрепленные к сообщению на верхнем уровне, но не вошедшие в пункты нарушений
const rootAttachments = computed(() => {
  const allAtts = props.message.attachments || [];
  if (!allAtts.length) return [];
  const violationAttIds = new Set();
  parsedViolations.value.forEach((v) => {
    (v.attachments || []).forEach((a) => violationAttIds.add(a.id));
    (v.attachment_ids || v.attachmentIds || []).forEach((id) => violationAttIds.add(id));
  });
  return allAtts.filter((a) => !violationAttIds.has(a.id));
});

function isVideo(att) {
  const mime = (att.mime_type || att.mimeType || '').toLowerCase();
  const name = (att.file_name || att.fileName || '').toLowerCase();
  return mime.startsWith('video/') || name.endsWith('.mp4') || name.endsWith('.webm');
}

function isImage(att) {
  const mime = (att.mime_type || att.mimeType || '').toLowerCase();
  const name = (att.file_name || att.fileName || '').toLowerCase();
  return (
    mime.startsWith('image/') ||
    name.endsWith('.png') ||
    name.endsWith('.jpg') ||
    name.endsWith('.jpeg') ||
    name.endsWith('.webp')
  );
}

function getMediaUrl(att) {
  if (att.thumbnail_data || att.thumbnailData) {
    return att.thumbnail_data || att.thumbnailData;
  }
  const token = localStorage.getItem('gdh_access_token');
  const base = att.url || `/api/v1/projects/${props.message.project_id || props.message.projectId}/chat/attachments/${att.id}`;
  if (token && !base.includes('token=')) {
    const sep = base.includes('?') ? '&' : '?';
    return `${base}${sep}token=${encodeURIComponent(token)}`;
  }
  return base;
}

function getDownloadUrl(att) {
  const token = localStorage.getItem('gdh_access_token');
  const base = att.download_url || `/api/v1/projects/${props.message.project_id || props.message.projectId}/chat/attachments/${att.id}/download`;
  if (token && !base.includes('token=')) {
    const sep = base.includes('?') ? '&' : '?';
    return `${base}${sep}token=${encodeURIComponent(token)}`;
  }
  return base;
}

function formatSize(bytes) {
  return formatBytes(bytes);
}

function openLightbox(att) {
  lightboxMedia.value = att;
}
</script>

<style scoped>
.verdict-card {
  background: rgba(239, 68, 68, 0.04);
  border: 1px solid rgba(239, 68, 68, 0.35);
  border-radius: 8px;
  padding: 14px 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  width: 100%;
  box-sizing: border-box;
}

.verdict-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 8px;
  border-bottom: 1px solid rgba(239, 68, 68, 0.2);
}

.verdict-header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.text-danger {
  color: #ef4444;
}

.verdict-title {
  font-size: 13px;
  font-weight: 700;
  color: #ef4444;
  letter-spacing: 0.3px;
  text-transform: uppercase;
}

.verdict-time {
  font-size: 11px;
  color: var(--text-tertiary, #8b949e);
}

.verdict-lead {
  font-size: 13px;
  color: var(--text-main, #f0f6fc);
  font-weight: 500;
  line-height: 1.4;
}

.violations-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.violation-item {
  background: var(--bg-card, #161b22);
  border: 1px solid rgba(239, 68, 68, 0.25);
  border-left: 3px solid #ef4444;
  border-radius: 6px;
  padding: 10px 14px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.violation-head {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.rule-badge-link {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  background: rgba(239, 68, 68, 0.15);
  border: 1px solid rgba(239, 68, 68, 0.4);
  color: #fca5a5;
  font-size: 11px;
  font-weight: 700;
  padding: 2px 8px;
  border-radius: 4px;
  text-decoration: none;
  transition: all 0.15s ease;
  cursor: pointer;
}

.rule-badge-link:hover {
  background: rgba(239, 68, 68, 0.25);
  border-color: #ef4444;
  color: #ffffff;
  transform: translateY(-1px);
}

.link-arrow {
  opacity: 0.7;
}

.rule-badge-link:hover .link-arrow {
  opacity: 1;
}

.rule-badge {
  background: rgba(239, 68, 68, 0.15);
  border: 1px solid rgba(239, 68, 68, 0.4);
  color: #fca5a5;
  font-size: 11px;
  font-weight: 700;
  padding: 2px 8px;
  border-radius: 4px;
}

.rule-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
}

.violation-desc {
  font-size: 13px;
  color: var(--text-secondary, #c9d1d9);
  line-height: 1.45;
  white-space: pre-wrap;
  word-break: break-word;
}

.violation-media-stack,
.root-attachments-stack {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-top: 4px;
}

.media-container {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.image-wrapper {
  position: relative;
  display: inline-block;
  max-width: 320px;
  max-height: 220px;
  border-radius: 6px;
  overflow: hidden;
  cursor: pointer;
  border: 1px solid var(--border, #30363d);
  background: #000;
}

.embedded-image {
  width: 100%;
  max-height: 220px;
  object-fit: contain;
  display: block;
  transition: transform 0.15s ease;
}

.image-wrapper:hover .embedded-image {
  transform: scale(1.02);
}

.image-overlay {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  color: #fff;
  font-size: 12px;
  font-weight: 500;
  opacity: 0;
  transition: opacity 0.15s ease;
}

.image-wrapper:hover .image-overlay {
  opacity: 1;
}

.video-wrapper {
  max-width: 420px;
  border-radius: 6px;
  overflow: hidden;
  border: 1px solid var(--border, #30363d);
  background: #000;
}

.embedded-video {
  width: 100%;
  max-height: 240px;
  display: block;
  outline: none;
}

.media-download-row {
  display: flex;
  align-items: center;
}

.media-download-link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--primary, #58a6ff);
  text-decoration: none;
  background: rgba(88, 166, 255, 0.08);
  padding: 4px 10px;
  border-radius: 4px;
  border: 1px solid rgba(88, 166, 255, 0.2);
  transition: all 0.15s ease;
}

.media-download-link:hover {
  background: rgba(88, 166, 255, 0.18);
  border-color: rgba(88, 166, 255, 0.4);
}

.file-name {
  font-weight: 500;
  max-width: 220px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-size {
  color: var(--text-tertiary, #8b949e);
}

.download-action {
  color: var(--primary, #58a6ff);
  font-weight: 600;
}

.verdict-fallback {
  font-size: 13px;
  color: var(--text-main, #f0f6fc);
  white-space: pre-wrap;
  line-height: 1.45;
}

.verdict-footer {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  margin-top: 4px;
  padding-top: 8px;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  font-size: 12px;
}

.footer-label {
  color: var(--text-tertiary, #8b949e);
  font-weight: 600;
  flex-shrink: 0;
}

.footer-text {
  color: var(--text-main, #f0f6fc);
}
</style>
