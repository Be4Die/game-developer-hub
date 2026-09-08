<template>
  <div class="tab-fade-in">
    <div class="form-grid">
      <!-- Уведомления статуса модерации (если есть) -->
      <div v-if="isUnderReview" class="card review-notice">
        ⏳ {{ t('projects.moderation') }} — Заявка проверяется модератором.
      </div>
      <div v-else-if="isRejected && rejectionReason" class="card rejection-notice">
        <strong>{{ t('moderation.rejectReasonLabel') }}</strong> {{ rejectionReason }}
      </div>

      <!-- БЛОК 1: МЕТАДАННЫЕ -->
      <div class="card form-section">
        <div class="section-head">
          <h3>{{ t('projectDraft.basicInfo') }}</h3>
        </div>
        <div class="input-row">
          <div class="input-group">
            <div class="input-header">
              <label class="input-label"
                >{{ t('projectDraft.gameTitleRu') }} <span class="req">*</span></label
              >
              <span class="char-count">{{ meta.title_ru.length }}/50</span>
            </div>
            <input
              v-model="meta.title_ru"
              type="text"
              class="input-control"
              maxlength="50"
              :disabled="!canEditInfo"
            />
          </div>
          <div class="input-group">
            <div class="input-header">
              <label class="input-label"
                >{{ t('projectDraft.gameTitleEn') }} <span class="req">*</span></label
              >
              <span class="char-count">{{ meta.title_en.length }}/50</span>
            </div>
            <input
              v-model="meta.title_en"
              type="text"
              class="input-control"
              maxlength="50"
              :disabled="!canEditInfo"
            />
          </div>
        </div>
        <div class="input-row">
          <div class="input-group">
            <div class="input-header">
              <label class="input-label">SEO (RU) <span class="req">*</span></label>
              <span class="char-count">{{ meta.seo_ru.length }}/180</span>
            </div>
            <textarea
              v-model="meta.seo_ru"
              class="input-control"
              rows="2"
              maxlength="180"
              :disabled="!canEditInfo"
            ></textarea>
          </div>
          <div class="input-group">
            <div class="input-header">
              <label class="input-label">SEO (EN) <span class="req">*</span></label>
              <span class="char-count">{{ meta.seo_en.length }}/180</span>
            </div>
            <textarea
              v-model="meta.seo_en"
              class="input-control"
              rows="2"
              maxlength="180"
              :disabled="!canEditInfo"
            ></textarea>
          </div>
        </div>
        <div class="input-row">
          <div class="input-group">
            <div class="input-header">
              <label class="input-label"
                >{{ t('projectDraft.gameDescriptionRu') }} <span class="req">*</span></label
              >
              <span class="char-count">{{ meta.about_ru.length }}/800</span>
            </div>
            <textarea
              v-model="meta.about_ru"
              class="input-control"
              rows="4"
              maxlength="800"
              :disabled="!canEditInfo"
            ></textarea>
          </div>
          <div class="input-group">
            <div class="input-header">
              <label class="input-label"
                >{{ t('projectDraft.gameDescriptionEn') }} <span class="req">*</span></label
              >
              <span class="char-count">{{ meta.about_en.length }}/800</span>
            </div>
            <textarea
              v-model="meta.about_en"
              class="input-control"
              rows="4"
              maxlength="800"
              :disabled="!canEditInfo"
            ></textarea>
          </div>
        </div>
      </div>

      <!-- СКРЫТЫЕ ИНПУТЫ ДЛЯ МЕДИА -->
      <input
        ref="fileIcon"
        type="file"
        accept="image/png"
        hidden
        @change="handleFileInput('icon', $event)"
      />
      <input
        ref="fileCoverMain"
        type="file"
        accept="image/png"
        hidden
        @change="handleFileInput('cover', $event)"
      />
      <input
        ref="fileVideo"
        type="file"
        accept="video/mp4"
        hidden
        @change="handleFileInput('video', $event)"
      />

      <!-- БЛОК 2: ПРОМО И МЕДИА-МАТЕРИАЛЫ -->
      <div class="card form-section">
        <div class="section-head">
          <h3>{{ t('projectDraft.seoAndMedia') }}</h3>
        </div>

        <div class="media-grid">
          <!-- СЛОТ 1: ИКОНКА ИГРЫ -->
          <div
            class="media-slot"
            :class="{
              'is-filled': media.icon && mediaUrls.icon,
              'is-empty': !media.icon || !mediaUrls.icon,
              'is-dragging': dragStates.icon,
              'is-loading': uploading.icon,
              'is-disabled': !canUploadMedia,
            }"
            @dragover.prevent="canUploadMedia && onDragOver('icon', $event)"
            @dragleave.prevent="canUploadMedia && onDragLeave('icon', $event)"
            @drop.prevent="canUploadMedia && onDrop('icon', $event)"
            @click="(!media.icon || !mediaUrls.icon) && canUploadMedia && triggerFileInput('icon')"
          >
            <div class="media-slot-header">
              <div class="media-slot-title-group">
                <ImageIcon class="icon-sm text-primary" />
                <span class="media-slot-title">{{ t('projectDraft.iconTitle') }}</span>
                <span class="req">*</span>
              </div>
              <span class="media-req-badge">{{ t('projectDraft.iconReq') }}</span>
            </div>

            <!-- Загруженное превью -->
            <div v-if="media.icon && mediaUrls.icon" class="media-preview-wrapper icon-size">
              <img
                :src="mediaUrls.icon"
                alt="Icon preview"
                class="media-preview-image"
                @error="handleMediaError('icon')"
              />
              <span v-if="pendingFiles.icon" class="staged-pill">Локальный файл</span>
              <button
                v-if="canUploadMedia"
                type="button"
                class="media-action-btn btn-delete"
                :title="t('projectDraft.removeFile')"
                @click.stop="removeMedia('icon')"
              >
                <Trash2 class="icon-xs" />
              </button>
            </div>

            <!-- Пустое состояние: область загрузки -->
            <div v-else class="media-upload-prompt">
              <div v-if="uploading.icon" class="upload-loader">
                <Loader2 class="icon-md spin text-primary" />
                <span>{{ t('projectDraft.uploadingFile') }}</span>
              </div>
              <div v-else class="upload-prompt-content">
                <div class="upload-icon-circle">
                  <Upload class="icon-md" />
                </div>
                <p class="upload-prompt-text">{{ t('projectDraft.dropOrClick') }}</p>
                <button
                  type="button"
                  class="btn-select-file"
                  :disabled="!canUploadMedia"
                  @click.stop="canUploadMedia && triggerFileInput('icon')"
                >
                  {{ t('projectDraft.selectFile') }}
                </button>
              </div>
            </div>
          </div>

          <!-- СЛОТ 2: ОБЛОЖКА ИГРЫ -->
          <div
            class="media-slot"
            :class="{
              'is-filled': media.cover && mediaUrls.cover,
              'is-empty': !media.cover || !mediaUrls.cover,
              'is-dragging': dragStates.cover,
              'is-loading': uploading.cover,
              'is-disabled': !canUploadMedia,
            }"
            @dragover.prevent="canUploadMedia && onDragOver('cover', $event)"
            @dragleave.prevent="canUploadMedia && onDragLeave('cover', $event)"
            @drop.prevent="canUploadMedia && onDrop('cover', $event)"
            @click="(!media.cover || !mediaUrls.cover) && canUploadMedia && triggerFileInput('cover')"
          >
            <div class="media-slot-header">
              <div class="media-slot-title-group">
                <ImageIcon class="icon-sm text-primary" />
                <span class="media-slot-title">{{ t('projectDraft.coverTitle') }}</span>
                <span class="req">*</span>
              </div>
              <span class="media-req-badge">{{ t('projectDraft.coverReq') }}</span>
            </div>

            <!-- Загруженное превью -->
            <div v-if="media.cover && mediaUrls.cover" class="media-preview-wrapper cover-size">
              <img
                :src="mediaUrls.cover"
                alt="Cover preview"
                class="media-preview-image"
                @error="handleMediaError('cover')"
              />
              <span v-if="pendingFiles.cover" class="staged-pill">Локальный файл</span>
              <button
                v-if="canUploadMedia"
                type="button"
                class="media-action-btn btn-delete"
                :title="t('projectDraft.removeFile')"
                @click.stop="removeMedia('cover')"
              >
                <Trash2 class="icon-xs" />
              </button>
            </div>

            <!-- Пустое состояние: область загрузки -->
            <div v-else class="media-upload-prompt">
              <div v-if="uploading.cover" class="upload-loader">
                <Loader2 class="icon-md spin text-primary" />
                <span>{{ t('projectDraft.uploadingFile') }}</span>
              </div>
              <div v-else class="upload-prompt-content">
                <div class="upload-icon-circle">
                  <Upload class="icon-md" />
                </div>
                <p class="upload-prompt-text">{{ t('projectDraft.dropOrClick') }}</p>
                <button
                  type="button"
                  class="btn-select-file"
                  :disabled="!canUploadMedia"
                  @click.stop="canUploadMedia && triggerFileInput('cover')"
                >
                  {{ t('projectDraft.selectFile') }}
                </button>
              </div>
            </div>
          </div>

          <!-- СЛОТ 3: ПРОМО-ВИДЕО -->
          <div
            class="media-slot"
            :class="{
              'is-filled': media.video && mediaUrls.video,
              'is-empty': !media.video || !mediaUrls.video,
              'is-dragging': dragStates.video,
              'is-loading': uploading.video,
              'is-disabled': !canUploadMedia,
            }"
            @dragover.prevent="canUploadMedia && onDragOver('video', $event)"
            @dragleave.prevent="canUploadMedia && onDragLeave('video', $event)"
            @drop.prevent="canUploadMedia && onDrop('video', $event)"
            @click="(!media.video || !mediaUrls.video) && canUploadMedia && triggerFileInput('video')"
          >
            <div class="media-slot-header">
              <div class="media-slot-title-group">
                <Film class="icon-sm text-primary" />
                <span class="media-slot-title">{{ t('projectDraft.videoTitle') }}</span>
              </div>
              <span class="media-req-badge">{{ t('projectDraft.videoReq') }}</span>
            </div>

            <!-- Загруженное видео -->
            <div v-if="media.video && mediaUrls.video" class="media-preview-wrapper video-size">
              <video
                :src="mediaUrls.video"
                autoplay
                loop
                muted
                playsinline
                class="media-preview-video"
                @error="handleMediaError('video')"
              ></video>
              <span v-if="pendingFiles.video" class="staged-pill">Локальный файл</span>
              <button
                v-if="canUploadMedia"
                type="button"
                class="media-action-btn btn-delete"
                :title="t('projectDraft.removeFile')"
                @click.stop="removeMedia('video')"
              >
                <Trash2 class="icon-xs" />
              </button>
            </div>

            <!-- Пустое состояние: область загрузки -->
            <div v-else class="media-upload-prompt">
              <div v-if="uploading.video" class="upload-loader">
                <Loader2 class="icon-md spin text-primary" />
                <span>{{ t('projectDraft.uploadingFile') }}</span>
              </div>
              <div v-else class="upload-prompt-content">
                <div class="upload-icon-circle">
                  <Upload class="icon-md" />
                </div>
                <p class="upload-prompt-text">{{ t('projectDraft.dropOrClick') }}</p>
                <button
                  type="button"
                  class="btn-select-file"
                  :disabled="!canUploadMedia"
                  @click.stop="canUploadMedia && triggerFileInput('video')"
                >
                  {{ t('projectDraft.selectFile') }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- БЛОК 3: БИЛД -->
      <div class="card form-section">
        <div class="section-head">
          <h3>{{ t('projectDraft.clientBuildSection') }}</h3>
        </div>

        <ClientBuildUploader
          v-if="canUploadBuild"
          :project-id="projectId"
          @build-uploaded="onBuildUploaded"
        />
        <div v-else class="build-no-perm-notice">
          <span class="text-muted">{{ t('access.permissions.PERM_UPLOAD_BUILD') }} — нет прав на загрузку новых билдов</span>
        </div>

        <!-- Список версий -->
        <div v-if="recentBuilds.length" class="build-versions">
          <h4 class="versions-title">{{ t('servers.buildsList') }}</h4>
          <div
            v-for="b in recentBuilds"
            :key="b.version"
            class="build-row"
            :class="{ active: activeBuildVersion === b.version }"
            @click="canUploadBuild && setActiveBuild(b.version)"
          >
            <div class="build-info">
              <div
                class="build-radio-indicator"
                :class="{ selected: activeBuildVersion === b.version }"
              >
                <div v-if="activeBuildVersion === b.version" class="radio-inner-dot"></div>
              </div>
              <div class="build-labels">
                <strong class="build-version-name">{{ b.version }}</strong>
                <span class="build-date">{{ b.created_at }}</span>
              </div>
            </div>

            <div class="build-actions">
              <span v-if="activeBuildVersion === b.version" class="active-badge">
                {{ t('common.active') }}
              </span>
              <button
                type="button"
                class="btn-icon-download"
                :title="t('common.download') || 'Скачать ZIP'"
                @click.stop="downloadBuild(b.version)"
              >
                <Download class="icon-xs" />
              </button>
            </div>
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
import { Image as ImageIcon, Film, Upload, Trash2, Loader2, Download } from 'lucide-vue-next';
import {
  getProject,
  updateProject,
  uploadMedia,
  getMediaUrl,
  submitForModeration as submitProjectForModeration,
} from '@/entities/project';
import { listClientBuilds } from '@/entities/build';
import { moderationApi, REQUEST_STATUS } from '@/entities/moderation';
import { ClientBuildUploader } from '@/features/upload-client-build';
import { showToast } from '@/shared/lib';

const { t } = useI18n();
const route = useRoute();
const projectId = computed(() => route.params.id);

const sharedProject = inject('project', null);
const draftActions = inject('draftActions', null);

const isOwner = computed(() => sharedProject?.value?.is_owner !== false);
const permissions = computed(() => sharedProject?.value?.current_user_permissions || []);
const canEditInfo = computed(() => isOwner.value || permissions.value.includes('PERM_EDIT_INFO'));
const canUploadMedia = computed(() => isOwner.value || permissions.value.includes('PERM_UPLOAD_MEDIA'));
const canUploadBuild = computed(() => isOwner.value || permissions.value.includes('PERM_UPLOAD_BUILD'));
const canSubmitModeration = computed(
  () => isOwner.value || permissions.value.includes('PERM_SUBMIT_MODERATION')
);

const meta = ref({
  title_ru: '',
  title_en: '',
  seo_ru: '',
  seo_en: '',
  about_ru: '',
  about_en: '',
});

const media = ref({ icon: false, cover: false, video: false });
const mediaUrls = ref({ icon: '', cover: '', video: '' });
const pendingFiles = reactive({ icon: null, cover: null, video: null });
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

const isRejected = computed(() => {
  const st = moderationStatus.value;
  return st === REQUEST_STATUS.REJECTED || st === 'REQUEST_STATUS_REJECTED' || st === 'rejected';
});

if (draftActions) {
  draftActions.value.save = () => saveMeta(false);
  draftActions.value.submit = () => submitForModeration();
  watch(
    submitting,
    (v) => {
      if (draftActions.value) draftActions.value.isSubmitting = v;
    },
    { immediate: true }
  );
  watch(
    isUnderReview,
    (v) => {
      if (draftActions.value) draftActions.value.isUnderReview = v;
    },
    { immediate: true }
  );
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
    stageMedia(type, file);
  }
}

function removeMedia(type) {
  if (pendingFiles[type]) {
    pendingFiles[type] = null;
  }
  if (mediaUrls.value[type] && mediaUrls.value[type].startsWith('blob:')) {
    URL.revokeObjectURL(mediaUrls.value[type]);
  }
  media.value[type] = false;
  mediaUrls.value[type] = '';
  if (fileIcon.value && type === 'icon') fileIcon.value.value = '';
  if (fileCoverMain.value && type === 'cover') fileCoverMain.value.value = '';
  if (fileVideo.value && type === 'video') fileVideo.value.value = '';
  showToast(t('projectDraft.removeFile') + ': ' + t(`projectDraft.${type}Title`), 'info');
}

function handleMediaError(type) {
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
      rejectionReason.value = data.request.rejection_reason || data.request.rejectionReason || '';
    } else {
      moderationRequest.value = null;
      moderationStatus.value = null;
      rejectionReason.value = '';
    }
  } catch (err) {
    // moderation service may not have a request yet
  }
}

async function loadProject(keepStaged = false) {
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
      about_ru:
        project.draft?.about_ru || project.about_ru || project.draft?.about || project.about || '',
      about_en: project.draft?.about_en || project.about_en || '',
    };

    const iconPath = project.draft?.icon_path || project.icon_path;
    const coverPath = project.draft?.cover_path || project.cover_path;
    const videoPath = project.draft?.video_path || project.video_path;

    if (!keepStaged || !pendingFiles.icon) {
      media.value.icon = !!iconPath;
      mediaUrls.value.icon = iconPath ? getMediaUrl(iconPath) : '';
    }
    if (!keepStaged || !pendingFiles.cover) {
      media.value.cover = !!coverPath;
      mediaUrls.value.cover = coverPath ? getMediaUrl(coverPath) : '';
    }
    if (!keepStaged || !pendingFiles.video) {
      media.value.video = !!videoPath;
      mediaUrls.value.video = videoPath ? getMediaUrl(videoPath) : '';
    }

    activeBuildVersion.value =
      project.draft?.active_build_version || project.active_build_version || '';

    const builds = await listClientBuilds(projectId.value);
    recentBuilds.value = builds;
    if (!activeBuildVersion.value && builds.length > 0) {
      activeBuildVersion.value = builds[0].version;
    }
    await loadModerationStatus();
  } catch (err) {
    showToast('Не удалось загрузить данные проекта', 'danger');
  }
}

onMounted(() => loadProject(false));

async function submitForModeration() {
  if (!canSubmitModeration.value) {
    showToast('Недостаточно прав для отправки на модерацию', 'danger');
    return;
  }
  const pId = parseInt(projectId.value, 10);
  if (!pId) {
    showToast('Не удалось определить ID игры', 'danger');
    return;
  }
  if (!meta.value.title_ru.trim() && !meta.value.title_en.trim()) {
    showToast('Укажите название игры', 'danger');
    return;
  }
  if (!meta.value.about_ru.trim() && !meta.value.about_en.trim()) {
    showToast('Заполните описание игры', 'danger');
    return;
  }
  submitting.value = true;
  try {
    await saveMeta(true);
    await submitProjectForModeration(pId);
    await loadModerationStatus();
    showToast('Заявка на модерацию успешно отправлена!', 'success');
  } catch (e) {
    showToast(e.response?.data?.message || e.message || 'Ошибка отправки на модерацию', 'danger');
  } finally {
    submitting.value = false;
  }
}

async function saveMeta(silent = false) {
  try {
    // 1. Сначала загружаем все локально прикрепленные медиафайлы (если есть права)
    if (canUploadMedia.value) {
      for (const type of ['icon', 'cover', 'video']) {
        if (pendingFiles[type]) {
          uploading[type] = true;
          try {
            await uploadMedia(projectId.value, type, pendingFiles[type]);
            pendingFiles[type] = null;
          } catch (uploadErr) {
            showToast(
              `Ошибка загрузки медиафайла (${type}): ${uploadErr.message || uploadErr}`,
              'danger'
            );
            throw uploadErr;
          } finally {
            uploading[type] = false;
          }
        }
      }
    }

    // 2. Обновляем текстовые метаданные черновика (если есть права)
    if (canEditInfo.value || canUploadBuild.value) {
      const payload = {};
      if (canEditInfo.value) {
        Object.assign(payload, meta.value);
      }
      if (canUploadBuild.value) {
        payload.active_build_version = activeBuildVersion.value;
      }
      await updateProject(projectId.value, payload);
      if (sharedProject && sharedProject.value && canEditInfo.value) {
        sharedProject.value = {
          ...sharedProject.value,
          title_ru: meta.value.title_ru,
          title_en: meta.value.title_en,
        };
      }
      await loadProject(false);
    }
    if (!silent) showToast('Черновик успешно сохранён!', 'success');
  } catch (err) {
    if (!silent) showToast('Ошибка сохранения черновика', 'danger');
    throw err;
  }
}

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
  await stageMedia(type, file);
  event.target.value = '';
};

async function stageMedia(type, file) {
  if (!file) return;
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

    pendingFiles[type] = file;
    if (mediaUrls.value[type] && mediaUrls.value[type].startsWith('blob:')) {
      URL.revokeObjectURL(mediaUrls.value[type]);
    }
    const localUrl = URL.createObjectURL(file);
    mediaUrls.value[type] = localUrl;
    media.value[type] = true;

    showToast(`Файл "${file.name}" выбран. Нажмите "Сохранить", чтобы загрузить.`, 'info');
  } catch (err) {
    showToast(err.message || 'Ошибка выбора файла', 'danger');
  }
}

async function onBuildUploaded(version) {
  activeBuildVersion.value = version;
  await loadProject(true);
}

function setActiveBuild(version) {
  activeBuildVersion.value = version;
  showToast(`Активная версия изменена на ${version}`, 'success');
}

function downloadBuild(version) {
  if (!version) return;
  const link = document.createElement('a');
  link.href = `/api/v1/projects/${projectId.value}/builds/${version}/download`;
  link.download = `project_${projectId.value}_v${version}.zip`;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
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

.published-info-notice {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  background: rgba(16, 185, 129, 0.08);
  border: 1px solid rgba(16, 185, 129, 0.3);
  border-left: 4px solid #10b981;
  padding: 12px 16px;
  color: var(--text-main);
  border-radius: var(--radius-md, 8px);
}

.notice-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  border-radius: 50%;
  background: #10b981;
  color: #fff;
  font-size: 0.8rem;
  font-weight: 700;
  flex-shrink: 0;
}

.notice-content {
  flex: 1;
}

.notice-sub {
  margin: 4px 0 0;
  font-size: 0.85rem;
  color: var(--text-muted);
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

.input-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.input-label {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-main);
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.req {
  color: var(--danger);
  font-weight: 700;
}

.char-count {
  font-weight: 400;
  font-size: 0.8rem;
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

/* Сетка медиа-материалов (единый компонент карточки без лишней вложенности) */
.media-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 20px;
}

.media-slot {
  border-radius: var(--radius-md, 8px);
  padding: 18px 20px;
  display: flex;
  flex-direction: column;
  gap: 14px;
  transition: all 0.2s ease;
  box-sizing: border-box;
}

/* Пустое состояние: пунктирная граница (линии) */
.media-slot.is-empty {
  border: 2px dashed var(--border);
  background: var(--bg-card);
  cursor: pointer;
}

.media-slot.is-empty:hover {
  border-color: var(--primary);
  background: var(--bg-hover);
}

.media-slot.is-dragging {
  border-color: var(--primary) !important;
  background: var(--primary-light, rgba(88, 166, 255, 0.08)) !important;
  box-shadow: 0 0 0 3px rgba(88, 166, 255, 0.2);
}

/* Заполненное состояние: сплошная граница */
.media-slot.is-filled {
  border: 1px solid var(--border);
  background: var(--bg-secondary);
}

.media-slot-header {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.media-slot-title-group {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.media-slot-title {
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

/* Область выбора/загрузки файла */
.media-upload-prompt {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px 10px;
}

.upload-prompt-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  text-align: center;
}

.upload-icon-circle {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  background: var(--bg-secondary);
  color: var(--primary);
  display: flex;
  align-items: center;
  justify-content: center;
}

.upload-prompt-text {
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

.upload-loader {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  font-size: 0.85rem;
  color: var(--text-muted);
  padding: 16px;
}

/* Превью загруженного медиа (центрировано без лишних рамок) */
.media-preview-wrapper {
  position: relative;
  border-radius: var(--radius-md, 8px);
  overflow: hidden;
  margin: 0 auto;
  border: 1px solid var(--border);
  background: var(--bg-card);
}

.media-preview-wrapper.icon-size {
  width: 160px;
  height: 160px;
}

.media-preview-wrapper.cover-size {
  width: 100%;
  max-width: 540px;
  aspect-ratio: 800 / 470;
}

.media-preview-wrapper.video-size {
  width: 100%;
  max-width: 540px;
  aspect-ratio: 16 / 9;
}

.media-preview-image,
.media-preview-video {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.media-action-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: none;
  cursor: pointer;
}

.btn-delete {
  position: absolute;
  top: 10px;
  right: 10px;
  background: rgba(220, 38, 38, 0.85);
  color: #fff;
  border: 1px solid rgba(239, 68, 68, 0.5);
  padding: 6px 8px;
  border-radius: var(--radius-sm, 6px);
  transition: all 0.15s ease;
  backdrop-filter: blur(8px);
  opacity: 0;
  z-index: 10;
}

.media-preview-wrapper:hover .btn-delete {
  opacity: 1;
}

.btn-delete:hover {
  background: #dc2626;
  transform: scale(1.05);
}

.staged-pill {
  position: absolute;
  top: 10px;
  left: 10px;
  background: rgba(88, 166, 255, 0.9);
  color: #ffffff;
  font-size: 0.72rem;
  font-weight: 600;
  padding: 3px 8px;
  border-radius: var(--radius-sm, 6px);
  backdrop-filter: blur(8px);
  pointer-events: none;
  z-index: 10;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
}

/* Список билдов */
.build-versions {
  margin-top: 20px;
}

.versions-title {
  margin: 0 0 10px;
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--text-main);
}

.build-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md, 8px);
  background: var(--bg-secondary);
  margin-bottom: 8px;
  cursor: pointer;
  transition: all 0.15s ease;
  user-select: none;
}

.build-row:hover {
  border-color: var(--primary);
  background: var(--bg-hover, rgba(255, 255, 255, 0.03));
}

.build-row.active {
  border-color: rgba(16, 185, 129, 0.4);
  background: rgba(16, 185, 129, 0.08);
}

.build-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.build-radio-indicator {
  width: 18px;
  height: 18px;
  border-radius: 50%;
  border: 2px solid var(--border-secondary, #484f58);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s ease;
  flex-shrink: 0;
}

.build-radio-indicator.selected {
  border-color: #10b981;
}

.radio-inner-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #10b981;
}

.build-labels {
  display: flex;
  align-items: center;
  gap: 10px;
}

.build-version-name {
  font-size: 0.92rem;
  font-weight: 700;
  color: var(--text-main);
}

.build-date {
  font-size: 0.8rem;
  color: var(--text-muted);
}

.build-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.active-badge {
  font-size: 0.76rem;
  font-weight: 600;
  color: #10b981;
  background: rgba(16, 185, 129, 0.12);
  padding: 3px 8px;
  border-radius: 4px;
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.btn-icon-download {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm, 6px);
  background: var(--bg-card);
  color: var(--text-muted);
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-icon-download:hover {
  color: var(--primary);
  border-color: var(--primary);
  background: var(--bg-secondary);
}

.media-slot.is-disabled {
  opacity: 0.6;
  cursor: not-allowed;
  pointer-events: none;
}

.build-no-perm-notice {
  padding: 16px;
  background: var(--bg-tertiary);
  border-radius: var(--radius-sm, 6px);
  font-size: 13px;
  text-align: center;
}
</style>
