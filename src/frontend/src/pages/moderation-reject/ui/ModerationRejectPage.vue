<template>
  <div class="reject-page-container">
    <div class="reject-content-wrap">
      <!-- Навигационная панель / Назад -->
      <div class="top-nav-bar">
        <button class="btn-back" @click="goBack">
          <ArrowLeft class="icon-sm" />
          <span>Назад к проверке проекта #{{ projectId }}</span>
        </button>
      </div>

      <!-- Состояние загрузки проекта -->
      <div v-if="loadingProject" class="loading-container">
        <div class="spinner-md"></div>
        <p>Загрузка данных проекта...</p>
      </div>

      <div v-else class="reject-form-layout">
        <!-- Шапка проекта / Сводка -->
        <div class="project-header-card">
          <div class="project-identity">
            <div class="game-avatar">
              <img
                v-if="projectIconUrl"
                :src="projectIconUrl"
                alt="Game Icon"
                class="avatar-img"
              />
              <div v-else class="avatar-mock">
                <Gamepad2 class="icon-md text-muted" />
              </div>
            </div>

            <div class="identity-info">
              <div class="title-row">
                <h1 class="game-title">{{ projectTitle }}</h1>
                <span v-if="projectVersion" class="version-tag">v{{ projectVersion }}</span>
                <span class="badge-status-in-review">На проверке</span>
              </div>
              <div class="meta-row">
                <span class="meta-item">
                  <User class="icon-xs text-muted" />
                  <strong>Разработчик:</strong> {{ devDisplayName }}
                </span>
                <span class="meta-item">
                  <strong>ID проекта:</strong> #{{ projectId }}
                </span>
              </div>
            </div>
          </div>

          <div class="header-warning-banner">
            <AlertTriangle class="icon-md text-danger banner-icon" />
            <div>
              <h4>Оформление решения об отклонении заявки</h4>
              <p>
                Опишите конкретные нарушения регламента платформы и прикрепите фото/видео доказательства.
                Структурированный отчёт поступит в чат проекта, а черновик вернётся разработчику на доработку.
              </p>
            </div>
          </div>
        </div>

        <!-- Секция: Конструктор пунктов нарушений -->
        <div class="section-card">
          <div class="section-header">
            <div class="section-title-wrap">
              <ShieldAlert class="icon-sm text-danger" />
              <h2>Пункты нарушений ({{ violations.length }})</h2>
            </div>
            <p class="section-subtitle">
              Разработчику будет проще исправить замечания, если каждое нарушение выделено в отдельный пункт с доказательством.
            </p>
          </div>

          <div class="violations-list">
            <div
              v-for="(item, idx) in violations"
              :key="idx"
              class="violation-box"
            >
              <div class="violation-box-header">
                <div class="violation-badge">
                  <span class="number-tag">Пункт #{{ idx + 1 }}</span>
                  <span v-if="item.ruleCode" class="rule-preview-tag">{{ item.ruleCode }}</span>
                </div>
                <button
                  v-if="violations.length > 1"
                  type="button"
                  class="btn-remove-box"
                  title="Удалить данный пункт нарушения"
                  @click="removeViolation(idx)"
                >
                  <Trash2 class="icon-xs" />
                  <span>Удалить пункт</span>
                </button>
              </div>

              <!-- Выбор правила из каталога регламента с поиском -->
              <div class="rule-selector-field">
                <label class="form-label">
                  Выберите пункт из регламента платформы (или введите вручную):
                </label>
                <RuleSearchSelect
                  v-model="item.ruleCode"
                  placeholder="Начните вводить: SEC-04, SRV-02, квоты, вызовы, баг..."
                  @select="onRuleSelected(item, $event)"
                  @clear="onRuleCleared(item)"
                />
              </div>

              <div class="form-grid-two">
                <div class="form-group">
                  <label class="form-label">
                    Код / № правила <span class="req">*</span>
                  </label>
                  <input
                    v-model="item.ruleCode"
                    type="text"
                    class="form-input code-font"
                    placeholder="Например: SEC-04 или 2.3.1"
                  />
                </div>

                <div class="form-group">
                  <label class="form-label">
                    Название правила платформы <span class="req">*</span>
                  </label>
                  <input
                    v-model="item.ruleTitle"
                    type="text"
                    class="form-input"
                    placeholder="Например: Несанкционированные сетевые запросы"
                  />
                </div>
              </div>

              <div class="form-group">
                <label class="form-label">
                  Подробное описание проблемы и где обнаружено <span class="req">*</span>
                </label>
                <textarea
                  v-model="item.description"
                  class="form-textarea"
                  rows="3"
                  placeholder="Опишите, в какой сцене, на каком уровне или при каких действиях воспроизводится нарушение, шаги для воспроизведения..."
                ></textarea>
              </div>

              <!-- Доказательства (скриншот или видео) -->
              <div class="evidence-block">
                <label class="form-label">Доказательства нарушения (скриншот или видео бага):</label>
                
                <div class="evidence-items-wrap">
                  <div
                    v-for="(att, aIdx) in item.attachments"
                    :key="att.id || aIdx"
                    class="evidence-chip"
                  >
                    <Film v-if="isVideo(att)" class="icon-xs text-primary" />
                    <Image v-else class="icon-xs text-primary" />
                    <span class="chip-name" :title="att.file_name || att.name">
                      {{ att.file_name || att.name }}
                    </span>
                    <span class="chip-size">({{ formatSize(att.file_size || att.size) }})</span>
                    <button
                      type="button"
                      class="btn-delete-chip"
                      title="Удалить файл"
                      @click="removeAttachment(item, aIdx)"
                    >
                      <X class="icon-xs" />
                    </button>
                  </div>

                  <!-- Кнопка прикрепления файла -->
                  <label class="btn-upload-evidence" :class="{ 'is-loading': item.uploading }">
                    <input
                      type="file"
                      accept="image/png,image/jpeg,image/webp,video/mp4,video/webm"
                      class="file-hidden-input"
                      :disabled="item.uploading"
                      @change="handleFileUpload($event, item)"
                    />
                    <Loader2 v-if="item.uploading" class="icon-xs spin" />
                    <Paperclip v-else class="icon-xs" />
                    <span>{{ item.uploading ? 'Загрузка файла...' : '+ Прикрепить скриншот / видео' }}</span>
                  </label>
                </div>
              </div>
            </div>
          </div>

          <!-- Кнопка добавления нового пункта -->
          <div class="add-more-row">
            <button type="button" class="btn-add-violation" @click="addViolation">
              <Plus class="icon-sm" />
              <span>Добавить еще одно нарушение</span>
            </button>
          </div>
        </div>

        <!-- Секция: Общее заключение / комментарий -->
        <div class="section-card">
          <div class="section-header">
            <div class="section-title-wrap">
              <MessageSquare class="icon-sm text-primary" />
              <h2>Общее заключение модератора</h2>
            </div>
            <p class="section-subtitle">
              Финальные рекомендации и напутствие разработчику перед повторной отправкой билда на проверку.
            </p>
          </div>

          <div class="form-group">
            <textarea
              v-model="generalComment"
              class="form-textarea"
              rows="3"
              placeholder="Например: Пожалуйста, устраните сетевые вызовы и приведите возрастные ограничения в порядок, после чего загрузите обновленный билд."
            ></textarea>
          </div>
        </div>

        <!-- Нижняя панель действий -->
        <div class="bottom-actions-panel">
          <div v-if="!canSubmit" class="validation-tip">
            <AlertCircle class="icon-xs text-muted" />
            <span>Заполните код, название и описание хотя бы для одного нарушения.</span>
          </div>

          <div class="actions-group">
            <button
              type="button"
              class="btn-cancel"
              :disabled="submitting"
              @click="goBack"
            >
              Отмена
            </button>

            <button
              type="button"
              class="btn-submit-reject"
              :disabled="!canSubmit || submitting"
              @click="handleSubmit"
            >
              <Loader2 v-if="submitting" class="icon-xs spin" />
              <XCircle v-else class="icon-xs" />
              <span>{{ submitting ? 'Отклонение проекта...' : 'Отклонить проект и отправить вердикт' }}</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import {
  ArrowLeft,
  Gamepad2,
  User,
  AlertTriangle,
  ShieldAlert,
  Trash2,
  Plus,
  Paperclip,
  Film,
  Image,
  X,
  Loader2,
  MessageSquare,
  AlertCircle,
  XCircle,
} from 'lucide-vue-next';
import { moderationApi, RuleSearchSelect } from '@/entities/moderation';
import { getProject, getMediaUrl } from '@/entities/project';
import { getUserDisplayName } from '@/entities/user';
import { showToast } from '@/shared/lib';

const route = useRoute();
const router = useRouter();

const projectId = computed(() => Number(route.params.projectId));

const loadingProject = ref(true);
const submitting = ref(false);
const activeRequest = ref(null);
const projectData = ref(null);

const violations = ref([
  {
    ruleCode: '',
    ruleTitle: '',
    description: '',
    attachments: [],
    uploading: false,
  },
]);

const generalComment = ref('');

const projectTitle = computed(() => {
  return (
    projectData.value?.titleRu ||
    projectData.value?.title_ru ||
    projectData.value?.titleEn ||
    projectData.value?.title_en ||
    `Проект #${projectId.value}`
  );
});

const projectVersion = computed(() => {
  return (
    projectData.value?.activeBuildVersion ||
    projectData.value?.active_build_version ||
    '1.0.0'
  );
});

const projectIconUrl = computed(() => {
  const p = projectData.value?.iconPath || projectData.value?.icon_path;
  if (!p) return '';
  return getMediaUrl(p);
});

const devDisplayName = computed(() => {
  const uid = activeRequest.value?.ownerId || projectData.value?.owner_id;
  return getUserDisplayName(uid);
});

const canSubmit = computed(() => {
  if (violations.value.length === 0) return false;
  return violations.value.every(
    (v) => v.ruleCode.trim() && v.ruleTitle.trim() && v.description.trim()
  );
});

function onRuleSelected(item, rule) {
  item.ruleCode = rule.code;
  item.ruleTitle = rule.title;
  if (!item.description.trim()) {
    item.description = rule.summary;
  }
}

function onRuleCleared(item) {
  item.ruleCode = '';
  item.ruleTitle = '';
}

function addViolation() {
  violations.value.push({
    ruleCode: '',
    ruleTitle: '',
    description: '',
    attachments: [],
    uploading: false,
  });
}

function removeViolation(idx) {
  if (violations.value.length > 1) {
    violations.value.splice(idx, 1);
  }
}

function removeAttachment(item, idx) {
  item.attachments.splice(idx, 1);
}

function isVideo(att) {
  const mt = att.mime_type || att.type || '';
  const n = att.file_name || att.name || '';
  return mt.startsWith('video/') || n.endsWith('.mp4') || n.endsWith('.webm');
}

function formatSize(bytes) {
  if (!bytes) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}

async function handleFileUpload(e, item) {
  const file = e.target.files?.[0];
  if (!file) return;

  const isVid = file.type.startsWith('video/');
  const maxBytes = isVid ? 50 * 1024 * 1024 : 15 * 1024 * 1024;
  if (file.size > maxBytes) {
    showToast(
      `Файл слишком велик (${formatSize(file.size)}). Максимум: ${isVid ? '50 МБ' : '15 МБ'}`,
      'danger'
    );
    e.target.value = '';
    return;
  }

  item.uploading = true;
  try {
    const res = await moderationApi.uploadAttachment(projectId.value, file);
    if (res && res.attachment) {
      item.attachments.push({
        id: res.attachment.id,
        project_id: projectId.value,
        file_name: res.attachment.file_name,
        file_size: res.attachment.file_size,
        mime_type: res.attachment.mime_type,
        url: res.attachment.url,
        download_url: res.attachment.download_url,
      });
      showToast(`Файл "${file.name}" прикреплен`, 'success');
    }
  } catch (err) {
    console.error('Failed to upload evidence attachment:', err);
    showToast('Не удалось загрузить файл', 'danger');
  } finally {
    item.uploading = false;
    e.target.value = '';
  }
}

async function loadProject() {
  loadingProject.value = true;
  try {
    const data = await moderationApi.getLatestByProject(projectId.value);
    if (data && data.request) {
      activeRequest.value = data.request;
      projectData.value = data.request.snapshot || {};
    } else {
      const p = await getProject(projectId.value);
      projectData.value = p || {};
    }
  } catch (err) {
    console.error('Failed to load project details:', err);
    showToast('Ошибка загрузки проекта', 'danger');
  } finally {
    loadingProject.value = false;
  }
}

function goBack() {
  router.push(`/moderator/projects/${projectId.value}`);
}

async function handleSubmit() {
  if (!canSubmit.value || submitting.value) return;

  submitting.value = true;
  try {
    const violationItems = violations.value.map((v) => ({
      rule_code: v.ruleCode.trim(),
      rule_title: v.ruleTitle.trim(),
      description: v.description.trim(),
      attachment_ids: v.attachments.map((a) => a.id).filter(Boolean),
      attachments: v.attachments,
    }));

    const primaryReason =
      violations.value[0]?.ruleTitle ||
      violations.value[0]?.ruleCode ||
      'Замечания по модерации';

    await moderationApi.reject(projectId.value, primaryReason, violationItems);

    showToast('Проект отклонен, вердикт с замечаниями отправлен разработчику', 'success');
    router.push(`/moderator/projects/${projectId.value}`);
  } catch (err) {
    console.error('Failed to reject project:', err);
    const msg = err.response?.data?.message || err.message || 'Ошибка отправки решения';
    showToast(`Не удалось отклонить проект: ${msg}`, 'danger');
  } finally {
    submitting.value = false;
  }
}

onMounted(() => {
  loadProject();
});
</script>

<style scoped>
.reject-page-container {
  min-height: calc(100vh - 60px);
  background: var(--bg-app, #0d1117);
  padding: 24px 32px 64px;
  box-sizing: border-box;
}

.reject-content-wrap {
  max-width: 1040px;
  margin: 0 auto;
}

.top-nav-bar {
  margin-bottom: 20px;
}

.btn-back {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: transparent;
  border: 1px solid var(--border, #30363d);
  color: var(--text-muted, #b0b8c4);
  padding: 8px 16px;
  border-radius: var(--radius-sm, 6px);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
}

.btn-back:hover {
  background: var(--bg-secondary, #21262d);
  color: var(--text-main, #f0f6fc);
  border-color: var(--primary, #58a6ff);
}

.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 0;
  gap: 16px;
  color: var(--text-muted, #b0b8c4);
}

.reject-form-layout {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

/* Шапка проекта */
.project-header-card {
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-md, 8px);
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.project-identity {
  display: flex;
  align-items: center;
  gap: 16px;
}

.game-avatar {
  width: 56px;
  height: 56px;
  border-radius: 10px;
  background: var(--bg-secondary, #21262d);
  border: 1px solid var(--border, #30363d);
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  flex-shrink: 0;
}

.avatar-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.identity-info {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.title-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.game-title {
  margin: 0;
  font-size: 20px;
  font-weight: 700;
  color: var(--text-main, #f0f6fc);
}

.version-tag {
  font-size: 11px;
  font-weight: 600;
  background: rgba(88, 166, 255, 0.15);
  color: var(--primary, #58a6ff);
  padding: 2px 8px;
  border-radius: 12px;
}

.badge-status-in-review {
  font-size: 11px;
  font-weight: 600;
  background: rgba(210, 153, 34, 0.15);
  color: #d29922;
  padding: 2px 8px;
  border-radius: 12px;
}

.meta-row {
  display: flex;
  align-items: center;
  gap: 16px;
  font-size: 13px;
  color: var(--text-muted, #b0b8c4);
}

.meta-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.header-warning-banner {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  background: rgba(248, 81, 73, 0.08);
  border: 1px solid rgba(248, 81, 73, 0.25);
  padding: 14px 18px;
  border-radius: 6px;
}

.banner-icon {
  margin-top: 2px;
  flex-shrink: 0;
}

.header-warning-banner h4 {
  margin: 0 0 4px;
  font-size: 14px;
  font-weight: 600;
  color: #f85149;
}

.header-warning-banner p {
  margin: 0;
  font-size: 13px;
  color: var(--text-muted, #b0b8c4);
  line-height: 1.4;
}

/* Секционные карточки */
.section-card {
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-md, 8px);
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.section-header {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.section-title-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
}

.section-title-wrap h2 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
}

.section-subtitle {
  margin: 0;
  font-size: 13px;
  color: var(--text-muted, #b0b8c4);
}

/* Список нарушений */
.violations-list {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.violation-box {
  background: var(--bg-secondary, #1c2128);
  border: 1px solid var(--border, #30363d);
  border-left: 3px solid #f85149;
  border-radius: 6px;
  padding: 18px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.violation-box-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.violation-badge {
  display: flex;
  align-items: center;
  gap: 8px;
}

.number-tag {
  font-size: 12px;
  font-weight: 700;
  background: rgba(248, 81, 73, 0.2);
  color: #f85149;
  padding: 2px 8px;
  border-radius: 4px;
}

.rule-preview-tag {
  font-size: 12px;
  font-family: monospace;
  background: var(--bg-tertiary, #2d333b);
  color: var(--text-main, #f0f6fc);
  padding: 2px 6px;
  border-radius: 4px;
}

.btn-remove-box {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: transparent;
  border: none;
  color: #f85149;
  font-size: 12px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
  transition: background 0.15s;
}

.btn-remove-box:hover {
  background: rgba(248, 81, 73, 0.15);
}

.rule-selector-field {
  margin-bottom: 12px;
}

.form-grid-two {
  display: grid;
  grid-template-columns: 1fr 2fr;
  gap: 14px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-label {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-muted, #b0b8c4);
}

.req {
  color: #f85149;
}

.form-input {
  height: 36px;
  background: var(--bg-tertiary, #0d1117);
  border: 1px solid var(--border, #30363d);
  border-radius: 6px;
  padding: 0 12px;
  color: var(--text-main, #f0f6fc);
  font-size: 13px;
  outline: none;
  transition: border-color 0.15s;
}

.form-input:focus,
.form-textarea:focus {
  border-color: var(--primary, #58a6ff);
}

.code-font {
  font-family: monospace;
}

.form-textarea {
  background: var(--bg-tertiary, #0d1117);
  border: 1px solid var(--border, #30363d);
  border-radius: 6px;
  padding: 10px 12px;
  color: var(--text-main, #f0f6fc);
  font-size: 13px;
  font-family: inherit;
  resize: vertical;
  outline: none;
  line-height: 1.4;
  transition: border-color 0.15s;
}

/* Доказательства */
.evidence-block {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.evidence-items-wrap {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

.evidence-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: var(--bg-tertiary, #0d1117);
  border: 1px solid var(--border, #30363d);
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 12px;
}

.chip-name {
  max-width: 180px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--text-main, #f0f6fc);
}

.chip-size {
  color: var(--text-tertiary, #8b949e);
  font-size: 11px;
}

.btn-delete-chip {
  background: transparent;
  border: none;
  color: var(--text-muted, #b0b8c4);
  cursor: pointer;
  display: flex;
  align-items: center;
  padding: 0;
}

.btn-delete-chip:hover {
  color: #f85149;
}

.btn-upload-evidence {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 30px;
  padding: 0 12px;
  background: var(--bg-secondary, #21262d);
  border: 1px dashed var(--border, #30363d);
  border-radius: 6px;
  color: var(--primary, #58a6ff);
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
}

.btn-upload-evidence:hover {
  border-color: var(--primary, #58a6ff);
  background: rgba(88, 166, 255, 0.08);
}

.btn-upload-evidence.is-loading {
  opacity: 0.6;
  cursor: wait;
}

.file-hidden-input {
  display: none;
}

.add-more-row {
  display: flex;
  justify-content: flex-start;
}

.btn-add-violation {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  height: 38px;
  padding: 0 16px;
  background: transparent;
  border: 1px dashed var(--border, #30363d);
  border-radius: 6px;
  color: var(--text-main, #f0f6fc);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
}

.btn-add-violation:hover {
  border-color: var(--primary, #58a6ff);
  color: var(--primary, #58a6ff);
  background: rgba(88, 166, 255, 0.05);
}

/* Нижняя панель */
.bottom-actions-panel {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-md, 8px);
  padding: 18px 24px;
}

.validation-tip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--text-muted, #b0b8c4);
}

.actions-group {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-left: auto;
}

.btn-cancel {
  height: 38px;
  padding: 0 18px;
  background: transparent;
  border: 1px solid var(--border, #30363d);
  color: var(--text-muted, #b0b8c4);
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
}

.btn-cancel:hover {
  background: var(--bg-secondary, #21262d);
  color: var(--text-main, #f0f6fc);
}

.btn-submit-reject {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  height: 38px;
  padding: 0 20px;
  background: #da3633;
  border: 1px solid #f85149;
  color: #ffffff;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.15s;
}

.btn-submit-reject:hover:not(:disabled) {
  opacity: 0.9;
}

.btn-submit-reject:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>
