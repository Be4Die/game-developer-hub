<template>
  <div class="tab-fade-in">
    <div class="form-grid">
      <!-- БЛОК 1: ОСНОВНАЯ ИНФОРМАЦИЯ -->
      <div class="card form-section">
        <div class="section-head">
          <h3>{{ t('projectDraft.basicInfo') }}</h3>
        </div>

        <div class="input-row">
          <div class="input-group">
            <div class="input-header">
              <label class="input-label">{{ t('projectDraft.gameTitleRu') }}</label>
            </div>
            <input
              type="text"
              class="input-control readonly"
              readonly
              :value="releaseData?.title_ru || '—'"
            />
          </div>
          <div class="input-group">
            <div class="input-header">
              <label class="input-label">{{ t('projectDraft.gameTitleEn') }}</label>
            </div>
            <input
              type="text"
              class="input-control readonly"
              readonly
              :value="releaseData?.title_en || '—'"
            />
          </div>
        </div>

        <div class="input-row">
          <div class="input-group">
            <div class="input-header">
              <label class="input-label">SEO (RU)</label>
            </div>
            <textarea
              class="input-control readonly"
              rows="2"
              readonly
              :value="releaseData?.seo_ru || '—'"
            ></textarea>
          </div>
          <div class="input-group">
            <div class="input-header">
              <label class="input-label">SEO (EN)</label>
            </div>
            <textarea
              class="input-control readonly"
              rows="2"
              readonly
              :value="releaseData?.seo_en || '—'"
            ></textarea>
          </div>
        </div>

        <div class="input-row">
          <div class="input-group">
            <div class="input-header">
              <label class="input-label">{{ t('projectDraft.gameDescriptionRu') }}</label>
            </div>
            <textarea
              class="input-control readonly"
              rows="4"
              readonly
              :value="releaseData?.about_ru || '—'"
            ></textarea>
          </div>
          <div class="input-group">
            <div class="input-header">
              <label class="input-label">{{ t('projectDraft.gameDescriptionEn') }}</label>
            </div>
            <textarea
              class="input-control readonly"
              rows="4"
              readonly
              :value="releaseData?.about_en || '—'"
            ></textarea>
          </div>
        </div>
      </div>

      <!-- БЛОК 2: ПРОМО И МЕДИА-МАТЕРИАЛЫ -->
      <div class="card form-section">
        <div class="section-head">
          <h3>{{ t('projectDraft.seoAndMedia') }}</h3>
        </div>

        <div class="media-grid">
          <!-- СЛОТ 1: ИКОНКА ИГРЫ -->
          <div
            class="media-slot"
            :class="iconUrl ? 'is-filled' : 'is-empty'"
          >
            <div class="media-slot-header">
              <div class="media-slot-title-group">
                <ImageIcon class="icon-sm text-primary" />
                <span class="media-slot-title">{{ t('projectDraft.iconTitle') }}</span>
              </div>
              <span class="media-req-badge">{{ t('projectDraft.iconReq') }}</span>
            </div>

            <div v-if="iconUrl" class="media-preview-wrapper icon-size">
              <img
                :src="iconUrl"
                alt="Icon preview"
                class="media-preview-image"
              />
            </div>
            <div v-else class="media-empty-info">
              <span>{{ t('projectDraft.mediaNotAttached') }}</span>
            </div>
          </div>

          <!-- СЛОТ 2: ОБЛОЖКА ИГРЫ -->
          <div
            class="media-slot"
            :class="coverUrl ? 'is-filled' : 'is-empty'"
          >
            <div class="media-slot-header">
              <div class="media-slot-title-group">
                <ImageIcon class="icon-sm text-primary" />
                <span class="media-slot-title">{{ t('projectDraft.coverTitle') }}</span>
              </div>
              <span class="media-req-badge">{{ t('projectDraft.coverReq') }}</span>
            </div>

            <div v-if="coverUrl" class="media-preview-wrapper cover-size">
              <img
                :src="coverUrl"
                alt="Cover preview"
                class="media-preview-image"
              />
            </div>
            <div v-else class="media-empty-info">
              <span>{{ t('projectDraft.mediaNotAttached') }}</span>
            </div>
          </div>

          <!-- СЛОТ 3: ПРОМО-ВИДЕО -->
          <div
            class="media-slot"
            :class="videoUrl ? 'is-filled' : 'is-empty'"
          >
            <div class="media-slot-header">
              <div class="media-slot-title-group">
                <Film class="icon-sm text-primary" />
                <span class="media-slot-title">{{ t('projectDraft.videoTitle') }}</span>
              </div>
              <span class="media-req-badge">{{ t('projectDraft.videoReq') }}</span>
            </div>

            <div v-if="videoUrl" class="media-preview-wrapper video-size">
              <video
                :src="videoUrl"
                controls
                playsinline
                class="media-preview-video"
              ></video>
            </div>
            <div v-else class="media-empty-info">
              <span>{{ t('projectDraft.mediaNotAttached') }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- БЛОК 3: СБОРКА И ОКРУЖЕНИЕ -->
      <div class="card form-section">
        <div class="section-head">
          <h3>{{ t('projectDraft.clientBuildSection') }}</h3>
        </div>

        <div class="published-build-box">
          <div class="build-ver-info">
            <span class="ver-label">{{ t('common.version') }}:</span>
            <strong class="ver-value">v{{ releaseVersion || '—' }}</strong>
          </div>
          <button
            v-if="releaseVersion"
            type="button"
            class="btn-download-action"
            @click="downloadBuild(releaseVersion)"
          >
            <Download class="icon-xs" />
            <span>{{ t('common.download') || 'Скачать' }} ZIP</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, inject } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import {
  Image as ImageIcon,
  Film,
  Download,
} from 'lucide-vue-next';
import { getProject, getPublished, getMediaUrl } from '@/entities/project';

const { t } = useI18n();
const route = useRoute();
const projectId = computed(() => route.params.id);

const sharedProject = inject('project', null);
const directRelease = ref(null);
const loading = ref(false);

const releaseData = computed(() => {
  if (directRelease.value) return directRelease.value;
  if (sharedProject?.value?.release) return sharedProject.value.release;
  return sharedProject?.value || null;
});

const releaseVersion = computed(() => {
  return (
    releaseData.value?.version ||
    releaseData.value?.active_build_version ||
    ''
  );
});

const iconUrl = computed(() => {
  const path = releaseData.value?.icon_path;
  return path ? getMediaUrl(path) : null;
});

const coverUrl = computed(() => {
  const path = releaseData.value?.cover_path;
  return path ? getMediaUrl(path) : null;
});

const videoUrl = computed(() => {
  const path = releaseData.value?.video_path;
  return path ? getMediaUrl(path) : null;
});

async function loadData() {
  loading.value = true;
  try {
    const rel = await getPublished(projectId.value);
    directRelease.value = rel;
    if (sharedProject) {
      const p = await getProject(projectId.value);
      sharedProject.value = p;
    }
  } catch (err) {
    // fallback to project data
  } finally {
    loading.value = false;
  }
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

onMounted(loadData);
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

/* Карточки формы */
.card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-md, 8px);
  padding: 24px;
}

.section-head {
  margin-bottom: 20px;
  border-bottom: 1px solid var(--border);
  padding-bottom: 12px;
}

.section-head h3 {
  margin: 0;
  font-size: 1.1rem;
  font-weight: 700;
  color: var(--text-main);
}

.input-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  margin-bottom: 16px;
}

.input-group {
  display: flex;
  flex-direction: column;
}

.input-header {
  display: flex;
  align-items: center;
  margin-bottom: 8px;
}

.input-label {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-main);
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

.input-control.readonly {
  background: var(--bg-secondary);
  cursor: default;
  color: var(--text-main);
}

.input-control.readonly:focus {
  outline: none;
  border-color: var(--border);
}

/* Медиа сетка */
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
  box-sizing: border-box;
}

.media-slot.is-filled {
  border: 1px solid var(--border);
  background: var(--bg-secondary);
}

.media-slot.is-empty {
  border: 1px dashed var(--border);
  background: var(--bg-card);
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

.media-empty-info {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
  color: var(--text-muted);
  font-size: 0.85rem;
}

/* Компактный блок опубликованной сборки */
.published-build-box {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 14px 18px;
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  border-radius: var(--radius-md, 8px);
}

.build-ver-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.ver-label {
  font-size: 0.9rem;
  color: var(--text-muted);
}

.ver-value {
  font-size: 0.95rem;
  font-weight: 700;
  color: var(--text-main);
}

.btn-download-action {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm, 6px);
  background: var(--bg-card);
  color: var(--text-main);
  font-weight: 600;
  font-size: 0.85rem;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-download-action:hover {
  border-color: var(--primary);
  color: var(--primary);
  background: var(--bg-secondary);
}
</style>

