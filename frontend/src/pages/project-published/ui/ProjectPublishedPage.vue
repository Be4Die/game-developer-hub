<template>
  <div class="tab-fade-in">
    <div class="form-grid">
      <!-- ЗАГОЛОВОК + КНОПКИ -->
      <div class="form-toolbar">
        <div class="title-block">
          <h1 style="margin: 0 0 8px 0; font-size: 1.5rem;">
            {{ t('projectWorkspace.publishedTab') }}
          </h1>
          <span class="status-badge bg-green">{{ t('projects.published') }}</span>
        </div>
        <div class="actions">
          <button class="btn-prod-link" @click="openProdGame">
            {{ t('projectDraft.openTest') }} (Prod)
          </button>
          <button class="btn-outline" @click="loadProject">{{ t('common.refresh') }}</button>
          <button class="btn btn-danger" @click="unpublishGame">
            {{ t('common.delete') }}
          </button>
        </div>
      </div>

      <!-- БЛОК 1: МЕТАДАННЫЕ -->
      <div class="card form-section">
        <div class="section-head">
          <h3>{{ t('projectDraft.basicInfo') }}</h3>
          <p class="version-info">{{ t('common.version') }}: {{ activeBuildDisplay }}</p>
        </div>

        <div class="input-row">
          <div class="input-group">
            <label>{{ t('projectDraft.gameTitle') }} (RU)</label>
            <div class="readonly-field">{{ project?.title_ru || '—' }}</div>
          </div>
          <div class="input-group">
            <label>{{ t('projectDraft.gameTitle') }} (EN)</label>
            <div class="readonly-field">{{ project?.title_en || '—' }}</div>
          </div>
        </div>

        <div class="input-row">
          <div class="input-group">
            <label>SEO (RU)</label>
            <div class="readonly-field">{{ project?.seo_ru || '—' }}</div>
          </div>
          <div class="input-group">
            <label>SEO (EN)</label>
            <div class="readonly-field">{{ project?.seo_en || '—' }}</div>
          </div>
        </div>

        <div class="input-row">
          <div class="input-group">
            <label>{{ t('projectDraft.gameDescriptionRu') }}</label>
            <div class="readonly-field multiline">
              {{ project?.about_ru || project?.about || '—' }}
            </div>
          </div>
          <div class="input-group">
            <label>{{ t('projectDraft.gameDescriptionEn') }}</label>
            <div class="readonly-field multiline">
              {{ project?.about_en || '—' }}
            </div>
          </div>
        </div>
      </div>

      <!-- БЛОК 2: ПРОМО -->
      <div class="card form-section">
        <div class="section-head"><h3>{{ t('projectDraft.seoAndMedia') }}</h3></div>

        <div class="media-list">
          <div
            class="media-item"
            :class="{
              uploaded: !!project?.icon_path,
              empty: !project?.icon_path,
            }"
          >
            <template v-if="project?.icon_path">
              <CheckCircle class="icon-md text-green" />
              <span class="m-title">{{ t('common.saved') }}</span>
              <span class="m-req">512 x 512, png</span>
              <span class="upload-label">Icon</span>
            </template>
            <template v-else>
              <ImageIcon class="icon-md" />
              <span class="m-title">{{ t('common.empty') }}</span>
              <span class="m-req">512 x 512, png</span>
              <span class="upload-label">Icon</span>
            </template>
          </div>

          <div
            class="media-item"
            :class="{
              uploaded: !!project?.cover_path,
              empty: !project?.cover_path,
            }"
          >
            <template v-if="project?.cover_path">
              <CheckCircle class="icon-md text-green" />
              <span class="m-title">{{ t('common.saved') }}</span>
              <span class="m-req">800 x 470, png</span>
              <span class="upload-label">Cover</span>
            </template>
            <template v-else>
              <ImageIcon class="icon-md" />
              <span class="m-title">{{ t('common.empty') }}</span>
              <span class="m-req">800 x 470, png</span>
              <span class="upload-label">Cover</span>
            </template>
          </div>

          <div
            class="media-item"
            :class="{
              uploaded: !!project?.video_path,
              empty: !project?.video_path,
            }"
          >
            <template v-if="project?.video_path">
              <CheckCircle class="icon-md text-green" />
              <span class="m-title">{{ t('common.saved') }}</span>
              <span class="m-req">≤ 12 MB</span>
              <span class="upload-label">Video</span>
            </template>
            <template v-else>
              <Film class="icon-md" />
              <span class="m-title">{{ t('common.empty') }}</span>
              <span class="m-req">≤ 12 MB</span>
              <span class="upload-label">Video</span>
            </template>
          </div>
        </div>
      </div>

      <!-- БЛОК 3: БИЛД -->
      <div class="card form-section">
        <div class="section-head"><h3>{{ t('projectDraft.clientBuildSection') }}</h3></div>

        <div v-if="activeBuildDisplay !== '—'" class="build-info">
          <div class="build-success-box">
            <CheckCircle class="icon-md text-green" />
            <div>
              <span style="display: block; font-weight: 600;">
                {{ t('common.version') }} {{ activeBuildDisplay }} {{ t('common.active') }}
              </span>
              <span
                style="
                  display: block;
                  font-size: 0.85rem;
                  color: var(--success);
                "
              >
                {{ t('moderation.verdictApproved') }}
              </span>
            </div>
          </div>
        </div>
        <div v-else class="build-info">
          <div class="build-empty-box">
            <AlertCircle class="icon-md" style="color: var(--warning);" />
            <div>
              <span style="display: block; font-weight: 600;">
                {{ t('servers.noBuilds') }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import {
  CheckCircle,
  Image as ImageIcon,
  Film,
  AlertCircle,
} from 'lucide-vue-next';
import { getProject, unpublish } from '@/entities/project';
import { listClientBuilds } from '@/entities/build';
import { showToast } from '@/shared/lib';

const { t } = useI18n();
const route = useRoute();
const projectId = computed(() => route.params.id);
const project = ref(null);
const builds = ref([]);

const activeBuildDisplay = computed(() => {
  return (
    project.value?.release?.build_version ||
    project.value?.active_build_version ||
    '—'
  );
});

async function loadProject() {
  try {
    project.value = await getProject(projectId.value);
    builds.value = await listClientBuilds(projectId.value);
  } catch (err) {
    showToast(t('common.error'), 'danger');
  }
}

function openProdGame() {
  const url =
    project.value?.release?.prod_url ||
    project.value?.prod_url ||
    `/games/${projectId.value}/prod/index.html`;
  window.open(url, '_blank');
}

async function unpublishGame() {
  try {
    await unpublish(projectId.value);
    showToast(t('common.success'), 'info');
    await loadProject();
  } catch (err) {
    showToast(t('common.error'), 'danger');
  }
}

onMounted(loadProject);
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
  max-width: 800px;
  padding-bottom: 60px;
}
.form-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
}
.actions {
  display: flex;
  gap: 12px;
}
.btn-prod-link {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border: 1px solid var(--success);
  border-radius: var(--radius-md);
  background: transparent;
  color: var(--success);
  font-weight: 600;
  font-size: 0.85rem;
  cursor: pointer;
  transition: 0.2s;
}
.btn-prod-link:hover {
  background: var(--success-light);
}

.status-badge {
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 0.8rem;
  font-weight: 600;
  display: inline-block;
}
.bg-green {
  background: var(--success-light);
  color: var(--success);
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
.version-info {
  font-size: 0.85rem;
  color: var(--text-muted);
  margin-top: 4px;
  margin-bottom: 0;
}

.input-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  margin-bottom: 16px;
}
.input-group label {
  display: block;
  font-size: 0.85rem;
  font-weight: 600;
  margin-bottom: 8px;
}
.readonly-field {
  padding: 10px 12px;
  background: var(--bg-secondary);
  border-radius: var(--radius-md);
  color: var(--text-main);
  font-size: 0.9rem;
  border: 1px solid var(--border);
}
.readonly-field.multiline {
  line-height: 1.5;
}

.media-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  align-items: center;
}
.media-item {
  border-radius: var(--radius-md);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 6px;
  text-align: center;
  padding: 14px;
  width: 100%;
  max-width: 510px;
  height: 110px;
}
.media-item.uploaded {
  border: 1px solid var(--success);
  background: var(--success-light);
  color: var(--success);
}
.media-item.empty {
  border: 1px dashed var(--border);
  background: var(--bg-secondary);
  color: var(--text-muted);
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
}
.media-item.uploaded .m-title {
  color: var(--success);
}
.m-req {
  font-size: 0.65rem;
}
.upload-label {
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--text-main);
  margin-top: 4px;
}

.build-info {
  margin-top: 24px;
}
.build-success-box {
  padding: 24px;
  border: 1px solid var(--success);
  border-radius: var(--radius-md);
  background: var(--success-light);
  display: flex;
  align-items: center;
  gap: 16px;
}
.build-empty-box {
  padding: 24px;
  border: 1px solid var(--warning);
  border-radius: var(--radius-md);
  background: var(--warning-light);
  display: flex;
  align-items: center;
  gap: 16px;
}
</style>
