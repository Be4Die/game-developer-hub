<template>
  <div class="modal-overlay" @click.self="$emit('cancel')">
    <div class="modal-card">
      <div class="modal-header">
        <div class="header-left">
          <AlertTriangle class="icon-md text-danger" />
          <h3>Отклонить проект и отправить замечания</h3>
        </div>
        <button class="btn-close-modal" @click="$emit('cancel')">
          <X class="icon-sm" />
        </button>
      </div>

      <p class="modal-hint">
        Укажите конкретные пункты нарушений правил и прикрепите доказательства (скриншоты или видео).
        Разработчик получит структурированный вердикт в чате проекта.
      </p>

      <!-- Список пунктов нарушений -->
      <div class="violations-builder-container">
        <div
          v-for="(item, idx) in violations"
          :key="idx"
          class="violation-card"
        >
          <div class="card-top-row">
            <span class="card-number-badge">Пункт #{{ idx + 1 }}</span>
            <button
              v-if="violations.length > 1"
              class="btn-remove-violation"
              title="Удалить этот пункт"
              @click="removeViolation(idx)"
            >
              <Trash2 class="icon-xs" />
              <span>Удалить</span>
            </button>
          </div>

          <div class="rule-inputs-row">
            <div class="input-group field-rule-code">
              <label class="input-label">№ правила</label>
              <input
                v-model="item.ruleCode"
                type="text"
                class="input-control code-input"
                placeholder="Напр. 2.3.1"
              />
            </div>
            <div class="input-group field-rule-title">
              <label class="input-label">Название правила платформы</label>
              <input
                v-model="item.ruleTitle"
                type="text"
                class="input-control"
                placeholder="Напр. Сексуализированный контент"
              />
            </div>
          </div>

          <div class="input-group">
            <label class="input-label">
              Описание проблемы и где обнаружено <span class="req">*</span>
            </label>
            <textarea
              v-model="item.description"
              class="input-control desc-textarea"
              placeholder="Опишите, в какой сцене, на каком уровне или при каких действиях возникает нарушение..."
              rows="2"
            ></textarea>
          </div>

          <!-- Прикрепление доказательств (фото / видео) -->
          <div class="violation-attachments-section">
            <label class="input-label sub-label">Доказательства (скриншот или видео бага):</label>
            
            <div class="attachments-list-chips">
              <div
                v-for="(att, aIdx) in item.attachments"
                :key="att.id || aIdx"
                class="attachment-chip"
              >
                <Film v-if="isVideo(att)" class="icon-xs text-primary" />
                <Image v-else class="icon-xs text-primary" />
                <span class="att-name">{{ att.file_name || att.name }}</span>
                <span class="att-size">({{ formatSize(att.file_size || att.size) }})</span>
                <button
                  type="button"
                  class="btn-remove-att"
                  title="Удалить вложение"
                  @click="removeAttachment(item, aIdx)"
                >
                  <X class="icon-xs" />
                </button>
              </div>

              <!-- Кнопка загрузки вложения -->
              <label class="btn-attach-file" :class="{ 'is-loading': item.uploading }">
                <input
                  type="file"
                  accept="image/png,image/jpeg,image/webp,video/mp4,video/webm"
                  class="file-hidden-input"
                  :disabled="item.uploading"
                  @change="handleFileUpload($event, item)"
                />
                <Loader2 v-if="item.uploading" class="icon-xs spin" />
                <Paperclip v-else class="icon-xs" />
                <span>{{ item.uploading ? 'Загрузка...' : '+ Прикрепить фото/видео' }}</span>
              </label>
            </div>
          </div>
        </div>

        <!-- Кнопка добавления нового пункта -->
        <button type="button" class="btn-add-violation" @click="addViolation">
          <Plus class="icon-xs" />
          <span>Добавить еще пункт нарушения</span>
        </button>
      </div>

      <!-- Общий комментарий / вердикт -->
      <div class="input-group general-comment-group">
        <label class="input-label">Общий вердикт / комментарий разработчику (необязательно)</label>
        <textarea
          v-model="generalComment"
          class="input-control"
          placeholder="Например: Пожалуйста, устраните замечания и загрузите обновленный билд на повторную проверку."
          rows="2"
        ></textarea>
      </div>

      <div class="modal-actions">
        <button class="btn-cancel" :disabled="loading" @click="$emit('cancel')">Отмена</button>
        <button
          class="btn-confirm-reject"
          :disabled="!canSubmit || loading"
          @click="onConfirm"
        >
          {{ loading ? 'Отклонение...' : '✕ Отклонить и отправить вердикт' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue';
import {
  AlertTriangle,
  Plus,
  Trash2,
  Paperclip,
  Image,
  Film,
  X,
  Loader2,
} from 'lucide-vue-next';
import { moderationApi } from '@/entities/moderation/api/moderationApi';
import {
  validateChatFile,
  compressImageIfNeeded,
  formatBytes,
} from '@/shared/lib/mediaCompressor';
import { showToast } from '@/shared/lib';

const props = defineProps({
  projectId: {
    type: [Number, String],
    required: true,
  },
  loading: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(['confirm', 'cancel']);

const generalComment = ref('');

const violations = ref([
  {
    ruleCode: '',
    ruleTitle: '',
    description: '',
    attachments: [],
    uploading: false,
  },
]);

function addViolation() {
  violations.value.push({
    ruleCode: '',
    ruleTitle: '',
    description: '',
    attachments: [],
    uploading: false,
  });
}

function removeViolation(index) {
  if (violations.value.length > 1) {
    violations.value.splice(index, 1);
  }
}

function removeAttachment(item, aIdx) {
  item.attachments.splice(aIdx, 1);
}

function isVideo(att) {
  const mime = (att.mime_type || att.type || '').toLowerCase();
  const name = (att.file_name || att.name || '').toLowerCase();
  return mime.startsWith('video/') || name.endsWith('.mp4') || name.endsWith('.webm');
}

function formatSize(bytes) {
  return formatBytes(bytes);
}

async function handleFileUpload(event, item) {
  const file = event.target.files?.[0];
  event.target.value = ''; // сброс инпута
  if (!file) return;

  // 1. Валидация
  const validation = validateChatFile(file);
  if (!validation.valid) {
    showToast(validation.error, 'danger');
    return;
  }

  item.uploading = true;
  try {
    // 2. Легковесная компрессия для тяжелых фото
    let processedFile = file;
    if (validation.isImage) {
      processedFile = await compressImageIfNeeded(file, 1920, 0.85);
    }

    // 3. Загрузка через Gateway
    const uploaded = await moderationApi.uploadAttachment(props.projectId, processedFile);
    item.attachments.push(uploaded);
    showToast('Файл успешно прикреплен', 'success');
  } catch (err) {
    console.error('Failed to upload attachment:', err);
    const msg =
      err.response?.data?.message ||
      (typeof err.response?.data === 'string' ? err.response.data.trim() : null);
    showToast(msg || 'Ошибка загрузки файла', 'danger');
  } finally {
    item.uploading = false;
  }
}

const canSubmit = computed(() => {
  const hasDesc = violations.value.some((v) => v.description.trim().length > 0);
  const hasGeneral = generalComment.value.trim().length > 0;
  return (hasDesc || hasGeneral) && !violations.value.some((v) => v.uploading);
});

function onConfirm() {
  const validViolations = violations.value
    .filter((v) => v.description.trim() || v.ruleCode.trim() || v.attachments.length)
    .map((v) => ({
      rule_code: v.ruleCode.trim(),
      rule_title: v.ruleTitle.trim(),
      description: v.description.trim(),
      attachment_ids: v.attachments.map((a) => a.id),
      attachments: v.attachments,
    }));

  emit('confirm', {
    reason: generalComment.value.trim(),
    violations: validViolations,
  });
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.65);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  backdrop-filter: blur(4px);
  padding: 20px;
  box-sizing: border-box;
}

.modal-card {
  background: var(--bg-card, #161b22);
  border-radius: var(--radius-lg, 12px);
  border: 1px solid var(--border, #30363d);
  padding: 24px;
  width: 100%;
  max-width: 680px;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  gap: 16px;
  box-shadow: 0 15px 35px rgba(0, 0, 0, 0.4);
  box-sizing: border-box;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.modal-header h3 {
  margin: 0;
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--text-main, #f0f6fc);
}

.btn-close-modal {
  background: transparent;
  border: none;
  color: var(--text-tertiary, #8b949e);
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
}

.btn-close-modal:hover {
  color: var(--text-main, #f0f6fc);
}

.text-danger {
  color: var(--danger, #ef4444);
}

.modal-hint {
  margin: 0;
  font-size: 0.85rem;
  color: var(--text-muted, #8b949e);
  line-height: 1.4;
}

.violations-builder-container {
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 12px;
  max-height: 50vh;
  padding-right: 4px;
}

.violation-card {
  background: var(--bg-secondary, #0d1117);
  border: 1px solid var(--border, #30363d);
  border-left: 3px solid #ef4444;
  border-radius: 8px;
  padding: 12px 14px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.card-top-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.card-number-badge {
  font-size: 11px;
  font-weight: 700;
  background: rgba(239, 68, 68, 0.15);
  color: #fca5a5;
  border: 1px solid rgba(239, 68, 68, 0.3);
  padding: 2px 8px;
  border-radius: 4px;
  text-transform: uppercase;
  letter-spacing: 0.3px;
}

.btn-remove-violation {
  display: flex;
  align-items: center;
  gap: 4px;
  background: transparent;
  border: none;
  color: var(--text-tertiary, #8b949e);
  font-size: 11px;
  cursor: pointer;
  padding: 2px 6px;
  border-radius: 4px;
  transition: color 0.15s;
}

.btn-remove-violation:hover {
  color: #ef4444;
}

.rule-inputs-row {
  display: grid;
  grid-template-columns: 140px 1fr;
  gap: 10px;
}

.input-label {
  display: block;
  font-size: 0.78rem;
  font-weight: 600;
  color: var(--text-secondary, #c9d1d9);
  margin-bottom: 4px;
}

.sub-label {
  font-weight: 500;
  font-size: 0.75rem;
  color: var(--text-tertiary, #8b949e);
}

.req {
  color: var(--danger, #ef4444);
}

.input-control {
  width: 100%;
  padding: 8px 10px;
  border: 1px solid var(--border, #30363d);
  border-radius: 6px;
  font-family: inherit;
  font-size: 13px;
  box-sizing: border-box;
  background: var(--bg-app, #0d1117);
  color: var(--text-main, #f0f6fc);
  outline: none;
  transition: border-color 0.15s;
}

.input-control:focus {
  border-color: #ef4444;
}

.desc-textarea {
  resize: vertical;
  min-height: 54px;
}

.violation-attachments-section {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.attachments-list-chips {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

.attachment-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: 6px;
  padding: 4px 8px;
  font-size: 12px;
  color: var(--text-main, #f0f6fc);
}

.att-name {
  max-width: 140px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.att-size {
  color: var(--text-tertiary, #8b949e);
  font-size: 11px;
}

.btn-remove-att {
  background: transparent;
  border: none;
  color: var(--text-tertiary, #8b949e);
  cursor: pointer;
  padding: 2px;
  display: flex;
  align-items: center;
}

.btn-remove-att:hover {
  color: #ef4444;
}

.file-hidden-input {
  display: none;
}

.btn-attach-file {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: rgba(88, 166, 255, 0.08);
  border: 1px dashed rgba(88, 166, 255, 0.4);
  color: var(--primary, #58a6ff);
  padding: 5px 10px;
  border-radius: 6px;
  font-size: 11px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-attach-file:hover:not(.is-loading) {
  background: rgba(88, 166, 255, 0.16);
  border-color: var(--primary, #58a6ff);
}

.btn-attach-file.is-loading {
  opacity: 0.6;
  cursor: wait;
}

.btn-add-violation {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 8px;
  background: rgba(88, 166, 255, 0.06);
  border: 1px dashed var(--border, #30363d);
  color: var(--text-secondary, #c9d1d9);
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-add-violation:hover {
  background: rgba(88, 166, 255, 0.12);
  border-color: var(--primary, #58a6ff);
  color: var(--primary, #58a6ff);
}

.general-comment-group {
  margin-top: 4px;
}

.modal-actions {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px solid var(--border, #30363d);
}

.btn-cancel {
  padding: 8px 16px;
  border: 1px solid var(--border, #30363d);
  border-radius: 6px;
  background: transparent;
  cursor: pointer;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
  font-size: 13px;
}

.btn-confirm-reject {
  padding: 8px 18px;
  border: none;
  border-radius: 6px;
  background: var(--danger, #ef4444);
  color: white;
  cursor: pointer;
  font-weight: 600;
  font-size: 13px;
  transition: opacity 0.15s;
}

.btn-confirm-reject:disabled {
  opacity: 0.5;
  cursor: not-allowed;
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
