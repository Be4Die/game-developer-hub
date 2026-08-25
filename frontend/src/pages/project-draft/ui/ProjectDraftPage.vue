<template>
  <div class="tab-fade-in">
    <div class="form-grid">
      <!-- Уведомления статуса модерации (если есть) -->
      <div
        v-if="isRejected && rejectionReason"
        class="card rejection-notice"
      >
        <strong>{{ t('moderation.rejectReasonLabel') }}</strong> {{ rejectionReason }}
      </div>
      <div
        v-else-if="isApproved"
        class="card approval-notice"
      >
        ✓ {{ t('moderation.verdictApproved') }}
      </div>
      <div
        v-else-if="isUnderReview"
        class="card review-notice"
      >
        ⏳ {{ t('projects.moderation') }}
      </div>

      <!-- БЛОК 1: МЕТАДАННЫЕ -->
      <div class="card form-section">
        <div class="section-head"><h3>{{ t('projectDraft.basicInfo') }}</h3></div>
        <div class="input-row">
          <div class="input-group">
            <label>{{ t('projectDraft.gameTitle') }} (RU) <span class="req">*</span></label>
            <input
              type="text"
              class="input-control"
              v-model="meta.title_ru"
            />
          </div>
          <div class="input-group">
            <label>{{ t('projectDraft.gameTitle') }} (EN) <span class="req">*</span></label>
            <input
              type="text"
              class="input-control"
              v-model="meta.title_en"
            />
          </div>
        </div>
        <div class="input-row">
          <div class="input-group">
            <label
              >SEO (RU) <span class="req">*</span>
              <span class="char-count">{{ meta.seo_ru.length }}/180</span></label
            >
            <textarea
              class="input-control"
              rows="2"
              maxlength="180"
              v-model="meta.seo_ru"
            ></textarea>
          </div>
          <div class="input-group">
            <label
              >SEO (EN) <span class="req">*</span>
              <span class="char-count">{{ meta.seo_en.length }}/180</span></label
            >
            <textarea
              class="input-control"
              rows="2"
              maxlength="180"
              v-model="meta.seo_en"
            ></textarea>
          </div>
        </div>
        <div class="input-group" style="margin-top: 16px;">
          <label
            >{{ t('projectDraft.gameDescription') }} <span class="req">*</span>
            <span class="char-count">{{ meta.about.length }}/800</span></label
          >
          <textarea
            class="input-control"
            rows="4"
            maxlength="800"
            v-model="meta.about"
          ></textarea>
        </div>
      </div>

      <!-- СКРЫТЫЕ ИНПУТЫ ДЛЯ МЕДИА -->
      <input
        type="file"
        ref="fileIcon"
        accept="image/png"
        hidden
        @change="handleFileInput('icon', $event)"
      />
      <input
        type="file"
        ref="fileCoverMain"
        accept="image/png"
        hidden
        @change="handleFileInput('cover', $event)"
      />
      <input
        type="file"
        ref="fileVideo"
        accept="video/mp4"
        hidden
        @change="handleFileInput('video', $event)"
      />

      <!-- БЛОК 2: ПРОМО И МЕДИА-МАТЕРИАЛЫ -->
      <div class="card form-section">
        <div class="section-head"><h3>{{ t('projectDraft.seoAndMedia') }}</h3></div>
        
        <div class="media-grid">
          <!-- СЛОТ 1: ИКОНКА ИГРЫ -->
          <div class="media-card">
            <div class="media-card-header">
              <div class="media-card-title-group">
                <ImageIcon class="icon-sm text-primary" />
                <span class="media-card-title">{{ t('projectDraft.iconTitle') }}</span>
              </div>
              <span class="media-req-badge">{{ t('projectDraft.iconReq') }}</span>
            </div>

            <!-- Загруженное превью -->
            <div
              v-if="media.icon && mediaUrls.icon"
              class="media-preview-container icon-ratio"
              @dragover.prevent="onDragOver('icon', $event)"
              @dragleave.prevent="onDragLeave('icon', $event)"
              @drop.prevent="onDrop('icon', $event)"
              :class="{ 'is-dragging': dragStates.icon }"
            >
              <img
                :src="mediaUrls.icon"
                alt="Icon preview"
                class="media-preview-image icon-fit"
                @error="handleMediaError('icon')"
              />
              
              <div class="media-overlay-actions">
                <button
                  type="button"
                  class="media-action-btn btn-delete"
                  @click="removeMedia('icon')"
                  :title="t('projectDraft.removeFile')"
                >
                  <Trash2 class="icon-xs" />
                </button>
              </div>
            </div>

            <!-- Интерактивный Dropzone -->
            <div
              v-else
              class="media-dropzone"
              :class="{ 'is-dragging': dragStates.icon, 'is-loading': uploading.icon }"
              @dragover.prevent="onDragOver('icon', $event)"
              @dragleave.prevent="onDragLeave('icon', $event)"
              @drop.prevent="onDrop('icon', $event)"
              @click="triggerFileInput('icon')"
            >
              <div v-if="uploading.icon" class="dropzone-loader">
                <Loader2 class="icon-md spin text-primary" />
                <span>{{ t('projectDraft.uploadingFile') }}</span>
              </div>
              <div v-else class="dropzone-content">
                <div class="dropzone-icon-circle">
                  <Upload class="icon-md" />
                </div>
                <p class="dropzone-text">{{ t('projectDraft.dropOrClick') }}</p>
                <button type="button" class="btn-select-file" @click.stop="triggerFileInput('icon')">
                  {{ t('projectDraft.selectFile') }}
                </button>
              </div>
            </div>
          </div>

          <!-- СЛОТ 2: ОБЛОЖКА ИГРЫ -->
          <div class="media-card">
            <div class="media-card-header">
              <div class="media-card-title-group">
                <ImageIcon class="icon-sm text-primary" />
                <span class="media-card-title">{{ t('projectDraft.coverTitle') }}</span>
              </div>
              <span class="media-req-badge">{{ t('projectDraft.coverReq') }}</span>
            </div>

            <!-- Загруженное превью -->
            <div
              v-if="media.cover && mediaUrls.cover"
              class="media-preview-container cover-ratio"
              @dragover.prevent="onDragOver('cover', $event)"
              @dragleave.prevent="onDragLeave('cover', $event)"
              @drop.prevent="onDrop('cover', $event)"
              :class="{ 'is-dragging': dragStates.cover }"
            >
              <img
                :src="mediaUrls.cover"
                alt="Cover preview"
                class="media-preview-image cover-fit"
                @error="handleMediaError('cover')"
              />
              
              <div class="media-overlay-actions">
                <button
                  type="button"
                  class="media-action-btn btn-delete"
                  @click="removeMedia('cover')"
                  :title="t('projectDraft.removeFile')"
                >
                  <Trash2 class="icon-xs" />
                </button>
              </div>
            </div>

            <!-- Интерактивный Dropzone -->
            <div
              v-else
              class="media-dropzone"
              :class="{ 'is-dragging': dragStates.cover, 'is-loading': uploading.cover }"
              @dragover.prevent="onDragOver('cover', $event)"
              @dragleave.prevent="onDragLeave('cover', $event)"
              @drop.prevent="onDrop('cover', $event)"
              @click="triggerFileInput('cover')"
            >
              <div v-if="uploading.cover" class="dropzone-loader">
                <Loader2 class="icon-md spin text-primary" />
                <span>{{ t('projectDraft.uploadingFile') }}</span>
              </div>
              <div v-else class="dropzone-content">
                <div class="dropzone-icon-circle">
                  <Upload class="icon-md" />
                </div>
                <p class="dropzone-text">{{ t('projectDraft.dropOrClick') }}</p>
                <button type="button" class="btn-select-file" @click.stop="triggerFileInput('cover')">
                  {{ t('projectDraft.selectFile') }}
                </button>
              </div>
            </div>
          </div>

          <!-- СЛОТ 3: ПРОМО-ВИДЕО -->
          <div class="media-card">
            <div class="media-card-header">
              <div class="media-card-title-group">
                <Film class="icon-sm text-primary" />
                <span class="media-card-title">{{ t('projectDraft.videoTitle') }}</span>
              </div>
              <span class="media-req-badge">{{ t('projectDraft.videoReq') }}</span>
            </div>

            <!-- Загруженное видео -->
            <div
              v-if="media.video && mediaUrls.video"
              class="media-preview-container video-ratio"
              @dragover.prevent="onDragOver('video', $event)"
              @dragleave.prevent="onDragLeave('video', $event)"
              @drop.prevent="onDrop('video', $event)"
              :class="{ 'is-dragging': dragStates.video }"
            >
              <video
                :src="mediaUrls.video"
                autoplay
                loop
                muted
                playsinline
                class="media-preview-video"
                @error="handleMediaError('video')"
              ></video>
              
              <div class="media-overlay-actions">
                <button
                  type="button"
                  class="media-action-btn btn-delete"
                  @click="removeMedia('video')"
                  :title="t('projectDraft.removeFile')"
                >
                  <Trash2 class="icon-xs" />
                </button>
              </div>
            </div>

            <!-- Интерактивный Dropzone -->
            <div
              v-else
              class="media-dropzone"
              :class="{ 'is-dragging': dragStates.video, 'is-loading': uploading.video }"
              @dragover.prevent="onDragOver('video', $event)"
              @dragleave.prevent="onDragLeave('video', $event)"
              @drop.prevent="onDrop('video', $event)"
              @click="triggerFileInput('video')"
            >
              <div v-if="uploading.video" class="dropzone-loader">
                <Loader2 class="icon-md spin text-primary" />
                <span>{{ t('projectDraft.uploadingFile') }}</span>
              </div>
              <div v-else class="dropzone-content">
                <div class="dropzone-icon-circle">
                  <Upload class="icon-md" />
                </div>
                <p class="dropzone-text">{{ t('projectDraft.dropOrClick') }}</p>
                <button type="button" class="btn-select-file" @click.stop="triggerFileInput('video')">
                  {{ t('projectDraft.selectFile') }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- БЛОК 3: БИЛД -->
      <div class="card form-section">
        <div class="section-head"><h3>{{ t('projectDraft.clientBuildSection') }}</h3></div>

        <ClientBuildUploader
          :project-id="projectId"
          @build-uploaded="onBuildUploaded"
        />

        <!-- Список версий -->
        <div v-if="recentBuilds.length" class="build-versions">
          <h4 class="versions-title">{{ t('servers.buildsList') }}</h4>
          <div
            v-for="b in recentBuilds"
            :key="b.version"
            class="build-row"
            :class="{ active: activeBuildVersion === b.version }"
          >
            <div class="build-info">
              <strong>{{ b.version }}</strong>
              <span class="build-date">{{ b.created_at }}</span>
            </div>
            <button
              v-if="activeBuildVersion !== b.version"
              class="btn-text"
              @click="setActiveBuild(b.version)"
            >
              {{ t('common.apply') }}
            </button>
            <span v-else class="active-label">{{ t('common.active') }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch, inject, reactive } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import {
  CheckCircle,
  Image as ImageIcon,
  Film,
  Upload,
  Trash2,
  Loader2,
} from 'lucide-vue-next';
import {
  getProject,
  updateProject,
  uploadMedia,
  getMediaUrl,
  submitForModeration as submitProjectForModeration,
} from '@/entities/project';
import { listClientBuilds } from '@/entities/build';
import {
  moderationApi,
  getStatusText,
  getStatusBadgeClass,
  REQUEST_STATUS,
} from '@/entities/moderation';
import { ClientBuildUploader } from '@/features/upload-client-build';
import { showToast } from '@/shared/lib';

const { t } = useI18n();
const route = useRoute();
const projectId = computed(() => route.params.id);

const sharedProject = inject('project', null);
const draftActions = inject('draftActions', null);

const meta = ref({
  title_ru: '',
  title_en: '',
  seo_ru: '',
  seo_en: '',
  about: '',
});

const media = ref({ icon: false, cover: false, video: false });
const mediaUrls = ref({ icon: '', cover: '', video: '' });
const uploading = reactive({ icon: false, cover: false, video: false });
const dragStates = reactive({ icon: false, cover: false, video: false });

const fileIcon = ref(null);
const fileCoverMain = ref(null);
const fileVideo = ref(null);

const activeBuildVersion = ref('');
const submitting = ref(false);
const moderationRequest = ref(null);
const moderationStatus = ref(null);
const rejectionReason = ref('');
const recentBuilds = ref([]);
const projectData = ref(null);

let autoSaveTimeout = null;
let skipAutoSave = false;

const isUnderReview = computed(() => {
  const st = moderationStatus.value;
  return (
    st === REQUEST_STATUS.PENDING ||
    st === REQUEST_STATUS.IN_REVIEW ||
    st === 'REQUEST_STATUS_PENDING' ||
    st === 'REQUEST_STATUS_IN_REVIEW' ||
    st === 'pending' ||
    st === 'in_review'
  );
});

const isApproved = computed(() => {
  const st = moderationStatus.value;
  return (
    st === REQUEST_STATUS.APPROVED ||
    st === 'REQUEST_STATUS_APPROVED' ||
    st === 'approved'
  );
});

const isRejected = computed(() => {
  const st = moderationStatus.value;
  return (
    st === REQUEST_STATUS.REJECTED ||
    st === 'REQUEST_STATUS_REJECTED' ||
    st === 'rejected'
  );
});

if (draftActions) {
  draftActions.value.save = () => saveMeta(false);
  draftActions.value.submit = () => submitForModeration();
  watch(submitting, (v) => { if (draftActions.value) draftActions.value.isSubmitting = v; }, { immediate: true });
  watch(isUnderReview, (v) => { if (draftActions.value) draftActions.value.isUnderReview = v; }, { immediate: true });
  watch(isApproved, (v) => { if (draftActions.value) draftActions.value.isApproved = v; }, { immediate: true });
}

function triggerFileInput(type) {
  if (type === 'icon' && fileIcon.value) fileIcon.value.click();
  if (type === 'cover' && fileCoverMain.value) fileCoverMain.value.click();
  if (type === 'video' && fileVideo.value) fileVideo.value.click();
}

function onDragOver(type, e) {
  e.preventDefault();
  dragStates[type] = true;
}

function onDragLeave(type, e) {
  e.preventDefault();
  dragStates[type] = false;
}

function onDrop(type, e) {
  e.preventDefault();
  dragStates[type] = false;
  const file = e.dataTransfer?.files?.[0];
  if (file) {
    processUpload(type, file);
  }
}

function removeMedia(type) {
  media.value[type] = false;
  mediaUrls.value[type] = '';
  if (fileIcon.value && type === 'icon') fileIcon.value.value = '';
  if (fileCoverMain.value && type === 'cover') fileCoverMain.value.value = '';
  if (fileVideo.value && type === 'video') fileVideo.value.value = '';
  showToast(t('projectDraft.removeFile') + ': ' + t(`projectDraft.${type}Title`), 'info');
}

function handleMediaError(type) {
  // If the server URL fails, fallback gracefully
  console.warn(`Media failed to load for ${type}: ${mediaUrls.value[type]}`);
}

async function loadModerationStatus() {
  const pId = parseInt(projectId.value, 10);
  if (!pId) return;
  try {
    const data = await moderationApi.getLatestByProject(pId);
    if (data && data.request) {
      moderationRequest.value = data.request;
      moderationStatus.value = data.request.status;
      rejectionReason.value =
        data.request.rejection_reason || data.request.rejectionReason || '';
    } else {
      moderationRequest.value = null;
      moderationStatus.value = null;
      rejectionReason.value = '';
    }
  } catch (err) {
    // moderation service may not have a request yet
  }
}

async function loadProject() {
  skipAutoSave = true;
  try {
    const project = await getProject(projectId.value);
    projectData.value = project;
    if (sharedProject) {
      sharedProject.value = project;
    }
    meta.value = {
      title_ru: project.draft?.title_ru || project.title_ru || '',
      title_en: project.draft?.title_en || project.title_en || '',
      seo_ru: project.draft?.seo_ru || project.seo_ru || '',
      seo_en: project.draft?.seo_en || project.seo_en || '',
      about: project.draft?.about || project.about || '',
    };

    const iconPath = project.draft?.icon_path || project.icon_path;
    const coverPath = project.draft?.cover_path || project.cover_path;
    const videoPath = project.draft?.video_path || project.video_path;

    media.value.icon = !!iconPath;
    media.value.cover = !!coverPath;
    media.value.video = !!videoPath;

    if (iconPath && (!mediaUrls.value.icon || !mediaUrls.value.icon.startsWith('blob:'))) {
      mediaUrls.value.icon = getMediaUrl(iconPath);
    }
    if (coverPath && (!mediaUrls.value.cover || !mediaUrls.value.cover.startsWith('blob:'))) {
      mediaUrls.value.cover = getMediaUrl(coverPath);
    }
    if (videoPath && (!mediaUrls.value.video || !mediaUrls.value.video.startsWith('blob:'))) {
      mediaUrls.value.video = getMediaUrl(videoPath);
    }

    activeBuildVersion.value =
      project.draft?.active_build_version ||
      project.active_build_version ||
      '';

    const builds = await listClientBuilds(projectId.value);
    recentBuilds.value = builds;
    if (!activeBuildVersion.value && builds.length > 0) {
      activeBuildVersion.value = builds[0].version;
    }
    await loadModerationStatus();
  } catch (err) {
    showToast('Не удалось загрузить данные проекта', 'danger');
  }
  setTimeout(() => {
    skipAutoSave = false;
  }, 1600);
}

onMounted(loadProject);

async function submitForModeration() {
  const pId = parseInt(projectId.value, 10);
  if (!pId) {
    showToast('Не удалось определить ID игры', 'danger');
    return;
  }
  if (!meta.value.title_ru.trim()) {
    showToast('Укажите название игры на русском', 'danger');
    return;
  }
  submitting.value = true;
  try {
    await saveMeta(true);
    await submitProjectForModeration(pId);
    await loadModerationStatus();
    showToast('Заявка на модерацию успешно отправлена!', 'success');
  } catch (e) {
    showToast(
      e.response?.data?.message || e.message || 'Ошибка отправки на модерацию',
      'danger'
    );
  } finally {
    submitting.value = false;
  }
}

async function saveMeta(silent = false) {
  try {
    const payload = {
      ...meta.value,
      active_build_version: activeBuildVersion.value,
    };
    await updateProject(projectId.value, payload);
    if (sharedProject && sharedProject.value) {
      sharedProject.value = {
        ...sharedProject.value,
        title_ru: meta.value.title_ru,
        title_en: meta.value.title_en,
      };
    }
    if (!silent) showToast('Сохранено!', 'success');
  } catch (err) {
    if (!silent) showToast('Ошибка сохранения', 'danger');
  }
}

watch(
  () => ({ ...meta.value, active_build_version: activeBuildVersion.value }),
  () => {
    if (skipAutoSave) return;
    if (autoSaveTimeout) clearTimeout(autoSaveTimeout);
    autoSaveTimeout = setTimeout(() => saveMeta(true), 1500);
  },
  { deep: true }
);

function validateImageDimensions(file, expectedWidth, expectedHeight) {
  return new Promise((resolve, reject) => {
    const img = new Image();
    img.onload = () => {
      URL.revokeObjectURL(img.src);
      if (img.width === expectedWidth && img.height === expectedHeight) {
        resolve(true);
      } else {
        reject(
          new Error(
            `Разрешение должно быть ${expectedWidth}x${expectedHeight}px (загружено ${img.width}x${img.height})`
          )
        );
      }
    };
    img.onerror = () => {
      URL.revokeObjectURL(img.src);
      reject(new Error('Не удалось загрузить изображение'));
    };
    img.src = URL.createObjectURL(file);
  });
}

const handleFileInput = async (type, event) => {
  const file = event.target.files[0];
  if (!file) return;
  await processUpload(type, file);
  event.target.value = '';
};

async function processUpload(type, file) {
  if (!file) return;
  uploading[type] = true;
  try {
    if (type === 'icon') {
      await validateImageDimensions(file, 512, 512);
    } else if (type === 'cover') {
      await validateImageDimensions(file, 800, 470);
    } else if (type === 'video') {
      if (file.size > 12 * 1024 * 1024) {
        throw new Error('Размер видео не должен превышать 12 МБ');
      }
    }

    // Instant local preview
    const localUrl = URL.createObjectURL(file);
    mediaUrls.value[type] = localUrl;
    media.value[type] = true;

    await uploadMedia(projectId.value, type, file);
    await loadProject();
    showToast('Медиафайл успешно сохранен', 'success');
  } catch (err) {
    if (!mediaUrls.value[type]) {
      media.value[type] = false;
    }
    showToast(err.message || 'Ошибка загрузки', 'danger');
  } finally {
    uploading[type] = false;
  }
}

async function onBuildUploaded(version) {
  activeBuildVersion.value = version;
  await loadProject();
}

function setActiveBuild(version) {
  activeBuildVersion.value = version;
  showToast(`Активная версия изменена на ${version}`, 'success');
}
</script>

<style scoped>
.tab-fade-in {
  animation: fadeIn 0.3s ease;
}
@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.form-grid {
  display: flex;
  flex-direction: column;
  gap: 24px;
  max-width: 900px;
  padding-bottom: 60px;
}

.rejection-notice {
  padding: 16px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  color: #b91c1c;
  font-size: 0.9rem;
  line-height: 1.5;
  border-radius: var(--radius-md, 8px);
}

.approval-notice {
  padding: 16px;
  background: var(--success-light);
  border: 1px solid var(--success);
  color: var(--success);
  font-size: 0.9rem;
  border-radius: var(--radius-md, 8px);
}

.review-notice {
  background: var(--bg-secondary);
  border-left: 4px solid var(--info, #3b82f6);
  padding: 12px 16px;
  color: var(--text-main);
  font-size: 0.9rem;
  border-radius: var(--radius-md, 8px);
}

.section-head {
  margin-bottom: 20px;
  border-bottom: 1px solid var(--border);
  padding-bottom: 12px;
}
.section-head h3 {
  margin: 0;
  font-size: 1.1rem;
}

.input-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  margin-bottom: 16px;
}
.input-group label {
  font-size: 0.85rem;
  font-weight: 600;
  margin-bottom: 8px;
  display: flex;
  justify-content: space-between;
}
.req {
  color: var(--danger);
}
.char-count {
  font-weight: 400;
  color: var(--text-muted);
}
.input-control {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  background: var(--bg-input);
  font-family: inherit;
  box-sizing: border-box;
  resize: vertical;
  color: var(--text-main);
}
.input-control:focus {
  outline: none;
  border-color: var(--primary);
  background: var(--bg-card);
}

/* Сетка медиа-материалов */
.media-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 20px;
}

.media-card {
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  border-radius: var(--radius-md, 8px);
  padding: 18px 20px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 14px;
}

.media-card-header {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.media-card-title-group {
  display: flex;
  align-items: center;
  gap: 8px;
}

.media-card-title {
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--text-main);
}

.media-req-badge {
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--text-muted);
  background: var(--bg-card);
  padding: 3px 8px;
  border-radius: 4px;
  border: 1px solid var(--border);
}

/* Dropzone (Empty State) */
.media-dropzone {
  width: 100%;
  border: 2px dashed var(--border);
  border-radius: var(--radius-md, 8px);
  background: var(--bg-card);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px 20px;
  cursor: pointer;
  transition: all 0.2s ease;
  position: relative;
  min-height: 140px;
  box-sizing: border-box;
}

.media-dropzone:hover {
  border-color: var(--primary);
  background: var(--bg-hover);
}

.media-dropzone.is-dragging {
  border-color: var(--primary);
  background: var(--primary-light, rgba(88, 166, 255, 0.1));
  box-shadow: 0 0 0 3px rgba(88, 166, 255, 0.2);
}

.dropzone-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  text-align: center;
}

.dropzone-icon-circle {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  background: var(--bg-secondary);
  color: var(--primary);
  display: flex;
  align-items: center;
  justify-content: center;
}

.dropzone-text {
  margin: 0;
  font-size: 0.85rem;
  color: var(--text-muted);
}

.btn-select-file {
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm, 6px);
  color: var(--text-main);
  font-size: 0.82rem;
  font-weight: 600;
  padding: 6px 14px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-select-file:hover {
  border-color: var(--primary);
  color: var(--primary);
  background: var(--bg-card);
}

.dropzone-loader {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  font-size: 0.85rem;
  color: var(--text-muted);
}

/* Preview Container (Uploaded State) */
.media-preview-container {
  position: relative;
  border-radius: var(--radius-md, 8px);
  overflow: hidden;
  border: 1px solid var(--border);
  background: var(--bg-card);
  display: flex;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
  margin: 0 auto;
}

.media-preview-container.icon-ratio {
  width: 160px;
  height: 160px;
}

.media-preview-container.cover-ratio {
  width: 100%;
  max-width: 540px;
  aspect-ratio: 800 / 470;
}

.media-preview-container.video-ratio {
  width: 100%;
  max-width: 540px;
  aspect-ratio: 16 / 9;
}

.media-preview-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.media-preview-video {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.media-overlay-actions {
  position: absolute;
  top: 10px;
  right: 10px;
  display: flex;
  align-items: center;
  gap: 8px;
  opacity: 0;
  transition: opacity 0.2s ease;
  z-index: 10;
}

.media-preview-container:hover .media-overlay-actions {
  opacity: 1;
}

.btn-delete {
  background: rgba(220, 38, 38, 0.85);
  color: #fff;
  border: 1px solid rgba(239, 68, 68, 0.5);
  padding: 6px 8px;
  border-radius: var(--radius-sm, 6px);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s ease;
  backdrop-filter: blur(8px);
}

.btn-delete:hover {
  background: #dc2626;
  transform: scale(1.05);
}

.btn-text {
  background: none;
  border: none;
  color: var(--primary);
  font-weight: 600;
  cursor: pointer;
  text-decoration: underline;
  padding: 0;
}

.build-versions {
  margin-top: 16px;
}
.versions-title {
  margin: 16px 0 8px;
  font-size: 0.9rem;
  color: var(--text-main);
}
.build-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  background: var(--bg-secondary);
  margin-bottom: 8px;
  transition: 0.2s;
}
.build-row.active {
  border-color: var(--success);
  background: var(--success-light);
}
.build-info {
  display: flex;
  align-items: center;
  gap: 8px;
}
</style>
