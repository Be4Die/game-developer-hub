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
        @change="handleFile('icon', $event)"
      />
      <input
        type="file"
        ref="fileCoverMain"
        accept="image/png"
        hidden
        @change="handleFile('cover', $event)"
      />
      <input
        type="file"
        ref="fileVideo"
        accept="video/mp4"
        hidden
        @change="handleFile('video', $event)"
      />

      <!-- БЛОК 2: ПРОМО -->
      <div class="card form-section">
        <div class="section-head"><h3>{{ t('projectDraft.seoAndMedia') }}</h3></div>
        <div class="media-list">
          <!-- Иконка -->
          <div
            class="media-item"
            :class="{ uploaded: media.icon }"
            @click="$refs.fileIcon.click()"
          >
            <CheckCircle v-if="media.icon" class="icon-md text-green" />
            <ImageIcon v-else class="icon-md" />
            <span class="m-title">{{
              media.icon ? t('common.saved') : 'Icon'
            }}</span>
            <span class="m-req">512 x 512, png</span>
          </div>
          <!-- Обложка -->
          <div
            class="media-item"
            :class="{ uploaded: media.cover }"
            @click="$refs.fileCoverMain.click()"
          >
            <CheckCircle v-if="media.cover" class="icon-md text-green" />
            <ImageIcon v-else class="icon-md" />
            <span class="m-title">{{
              media.cover ? t('common.saved') : 'Cover'
            }}</span>
            <span class="m-req">800 x 470, png</span>
          </div>
          <!-- Видео -->
          <div
            class="media-item"
            :class="{ uploaded: media.video }"
            @click="$refs.fileVideo.click()"
          >
            <CheckCircle v-if="media.video" class="icon-md text-green" />
            <Film v-else class="icon-md" />
            <span class="m-title">{{
              media.video ? t('common.saved') : 'Video'
            }}</span>
            <span class="m-req">≤ 12 MB</span>
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
import { ref, computed, onMounted, watch, inject } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { CheckCircle, Image as ImageIcon, Film } from 'lucide-vue-next';
import {
  getProject,
  updateProject,
  uploadMedia,
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
    media.value.icon = !!(project.draft?.icon_path || project.icon_path);
    media.value.cover = !!(project.draft?.cover_path || project.cover_path);
    media.value.video = !!(project.draft?.video_path || project.video_path);
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

const handleFile = async (type, event) => {
  const file = event.target.files[0];
  if (!file) return;
  try {
    if (type === 'icon') {
      await validateImageDimensions(file, 512, 512);
    } else if (type === 'cover') {
      await validateImageDimensions(file, 800, 470);
    }
    await uploadMedia(projectId.value, type, file);
    media.value[type] = true;
    await loadProject();
    showToast('Медиафайл успешно сохранен', 'success');
  } catch (err) {
    showToast(err.message || 'Ошибка загрузки', 'danger');
  } finally {
    event.target.value = '';
  }
};

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

.media-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  align-items: center;
}
.media-item {
  border: 1px dashed var(--border);
  border-radius: var(--radius-md);
  background: var(--bg-secondary);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 6px;
  color: var(--text-muted);
  text-align: center;
  cursor: pointer;
  padding: 14px;
  transition: 0.2s;
  width: 100%;
  max-width: 510px;
  height: 110px;
}
.media-item:hover {
  border-color: var(--primary);
  background: var(--bg-hover);
  color: var(--primary);
}
.media-item.uploaded {
  border: 1px solid var(--success);
  background: var(--success-light);
  color: var(--success);
}
.media-item .icon-md {
  width: 16px;
  height: 16px;
}
.text-green {
  color: var(--success);
}
.m-title {
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--text-main);
}
.media-item.uploaded .m-title {
  color: var(--success);
}
.m-req {
  font-size: 0.65rem;
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
