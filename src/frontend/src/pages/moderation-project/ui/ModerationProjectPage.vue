<template>
  <div v-if="projectData" class="moderator-workspace">
    <!-- ЛЕВАЯ КОЛОНКА: ИНСПЕКТОР ЧЕРНОВИКА И ДЕЙСТВИЯ -->
    <main class="review-inspector">
      <div class="inspector-scroll">
        <!-- Шапка проекта (Карточка идентификации) -->
        <div class="card project-identity-card">
          <div class="identity-header">
            <div class="identity-icon-box">
              <img
                v-if="projectIconUrl"
                :src="projectIconUrl"
                alt="Icon"
                class="identity-icon-img"
              />
              <div v-else class="identity-icon-mock">
                <span>Draft</span>
              </div>
            </div>

            <div class="identity-text-box">
              <div class="identity-top-row">
                <span class="project-id-tag">Проект #{{ projectId }}</span>
                <span
                  v-if="activeRequest"
                  class="status-badge"
                  :class="getStatusBadgeClass(requestStatus)"
                >
                  {{ getStatusText(requestStatus) }}
                </span>
                <span v-if="projectData.activeBuildVersion" class="version-tag">
                  v{{ projectData.activeBuildVersion }}
                </span>
              </div>
              <h1 class="project-main-title">
                {{ projectData.titleRu || projectData.titleEn || `Проект #${projectId}` }}
              </h1>
              <div v-if="projectData.titleEn && projectData.titleRu" class="project-sub-title">
                {{ projectData.titleEn }}
              </div>
            </div>
          </div>

          <!-- Метаданные проверки -->
          <div class="identity-meta-grid">
            <div class="meta-item">
              <span class="meta-label">{{ t('moderation.developerColumn') }}</span>
              <span class="meta-value">{{ activeRequest?.ownerId || '—' }}</span>
            </div>
            <div class="meta-item">
              <span class="meta-label">{{ t('moderation.submittedColumn') }}</span>
              <span class="meta-value">
                {{ activeRequest ? formatDateTime(activeRequest.submittedAt) : '—' }}
              </span>
            </div>
            <div class="meta-item">
              <span class="meta-label">{{ t('moderation.moderator') }}</span>
              <span class="meta-value">
                {{ activeRequest?.moderatorId || t('moderation.notAssigned') }}
              </span>
            </div>
          </div>
        </div>

        <!-- КАРТОЧКА 1: ОСНОВНАЯ ИНФОРМАЦИЯ -->
        <div class="card section-card">
          <div class="section-head">
            <h3>{{ t('moderation.basicInfo') }}</h3>
          </div>

          <!-- Названия -->
          <div class="data-row">
            <div class="data-group">
              <label class="data-label">{{ t('projectDraft.gameTitleRu') }}</label>
              <div class="data-box">{{ projectData.titleRu || '—' }}</div>
            </div>
            <div class="data-group">
              <label class="data-label">{{ t('projectDraft.gameTitleEn') }}</label>
              <div class="data-box">{{ projectData.titleEn || '—' }}</div>
            </div>
          </div>

          <!-- SEO описания -->
          <div class="data-row">
            <div class="data-group">
              <label class="data-label">SEO (RU)</label>
              <div class="data-box multiline">{{ projectData.seoRu || '—' }}</div>
            </div>
            <div class="data-group">
              <label class="data-label">SEO (EN)</label>
              <div class="data-box multiline">{{ projectData.seoEn || '—' }}</div>
            </div>
          </div>

          <!-- Описание игры -->
          <div class="data-row">
            <div class="data-group">
              <label class="data-label">{{ t('projectDraft.gameDescriptionRu') }}</label>
              <div class="data-box multiline-lg">
                {{ projectData.aboutRu || projectData.about || t('moderation.noDescription') }}
              </div>
            </div>
            <div class="data-group">
              <label class="data-label">{{ t('projectDraft.gameDescriptionEn') }}</label>
              <div class="data-box multiline-lg">
                {{ projectData.aboutEn || t('moderation.noDescription') }}
              </div>
            </div>
          </div>
        </div>

        <!-- КАРТОЧКА 2: МЕДИА-МАТЕРИАЛЫ -->
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
                <div v-if="projectIconUrl" class="img-preview-wrap icon-aspect">
                  <img :src="projectIconUrl" alt="Icon" class="preview-img" />
                </div>
                <div v-else class="media-empty-placeholder">
                  <Image class="icon-md text-muted" />
                  <span>{{ t('moderation.noMedia') }}</span>
                </div>
              </div>
            </div>

            <!-- ОБЛОЖКА ИГРЫ (800x470) -->
            <div class="media-box-slot">
              <div class="slot-title-row">
                <span class="slot-label">{{ t('projectDraft.coverMain') }}</span>
                <span class="slot-spec">800×470 PNG</span>
              </div>
              <div class="media-view-panel">
                <div v-if="projectCoverUrl" class="img-preview-wrap cover-aspect">
                  <img :src="projectCoverUrl" alt="Cover" class="preview-img" />
                </div>
                <div v-else class="media-empty-placeholder">
                  <Image class="icon-md text-muted" />
                  <span>{{ t('moderation.noMedia') }}</span>
                </div>
              </div>
            </div>

            <!-- ПРОМО-ВИДЕО (≤ 12 МБ, MP4) -->
            <div class="media-box-slot">
              <div class="slot-title-row">
                <span class="slot-label">{{ t('projectDraft.promoVideo') }}</span>
                <span class="slot-spec">≤ 12 МБ, MP4</span>
              </div>
              <div class="media-view-panel">
                <div v-if="projectVideoUrl" class="img-preview-wrap video-aspect">
                  <video :src="projectVideoUrl" controls class="preview-video"></video>
                </div>
                <div v-else class="media-empty-placeholder">
                  <Video class="icon-md text-muted" />
                  <span>{{ t('moderation.noMedia') }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- КАРТОЧКА 3: СБОРКА И ТЕСТИРОВАНИЕ -->
        <div class="card section-card">
          <div class="section-head">
            <h3>{{ t('moderation.buildTesting') }}</h3>
          </div>

          <div class="build-test-row">
            <div class="build-info-block">
              <span class="build-version-tag">
                Версия сборки: <strong>v{{ projectData.activeBuildVersion || '1.0.0' }}</strong>
              </span>
              <p class="build-desc">
                Проверьте работоспособность игры, управление, отсутствие критических ошибок и
                соответствие контента правилам платформы.
              </p>
            </div>

            <button class="btn-play-dev-lg" @click="openDevPreview">
              <Gamepad2 class="icon-sm" />
              <span>{{ t('moderation.runDevBuild') }}</span>
              <ExternalLink class="icon-xs" />
            </button>
          </div>
        </div>

        <!-- КАРТОЧКА 4: ВЕРДИКТ И ПАНЕЛЬ ДЕЙСТВИЙ -->
        <div class="card verdict-card">
          <div class="section-head">
            <h3>{{ t('moderation.verdictSection') }}</h3>
          </div>

          <!-- Баннер вердикта: Отклонен -->
          <div v-if="isRejected" class="result-box box-rejected">
            <AlertTriangle class="icon-md text-danger" />
            <div class="result-text">
              <strong>{{ t('moderation.rejected') }}</strong>
              <p v-if="activeRequest?.rejectionReason">
                {{ t('moderation.rejectionReason') }}: {{ activeRequest.rejectionReason }}
              </p>
            </div>
          </div>

          <!-- Баннер вердикта: Одобрен -->
          <div v-else-if="isApproved" class="result-box box-approved">
            <CheckCircle2 class="icon-md text-success" />
            <div class="result-text">
              <strong>{{ t('moderation.approved') }}</strong>
              <p>Проект одобрен и опубликован в основном каталоге платформы.</p>
            </div>
          </div>

          <!-- Панель действий: когда заявка активна -->
          <div v-else-if="activeRequest" class="verdict-actions-row">
            <!-- Кнопка "Взять в работу" (Pending) -->
            <button
              v-if="isPending"
              class="btn-verdict btn-claim-ticket"
              :disabled="actionLoading"
              @click="handleClaim"
            >
              <Eye class="icon-sm" />
              <span>{{ t('moderation.claimBtn') }}</span>
            </button>

            <!-- Кнопки решения (In Review) -->
            <template v-else-if="isInReview">
              <button
                class="btn-verdict btn-approve-ticket"
                :disabled="actionLoading"
                @click="showApproveModal = true"
              >
                <CheckCircle2 class="icon-sm" />
                <span>{{ t('moderation.approve') }}</span>
              </button>

              <button
                class="btn-verdict btn-reject-ticket"
                :disabled="actionLoading"
                @click="showRejectModal = true"
              >
                <XCircle class="icon-sm" />
                <span>{{ t('moderation.reject') }}</span>
              </button>
            </template>
          </div>
        </div>
      </div>
    </main>

    <!-- ПРАВАЯ КОЛОНКА: ЧАТ ПРОЕКТА -->
    <aside class="moderator-chat-aside">
      <ProjectChat :project-id="projectId" />
    </aside>

    <!-- Модальные окна одобрения и отклонения -->
    <ApproveRequestModal
      v-if="showApproveModal"
      :loading="actionLoading"
      @confirm="handleApprove"
      @cancel="showApproveModal = false"
    />

    <RejectRequestModal
      v-if="showRejectModal"
      :loading="actionLoading"
      @confirm="handleReject"
      @cancel="showRejectModal = false"
    />
  </div>

  <!-- Состояния загрузки и ошибки -->
  <div v-else-if="loading" class="state-loading-screen">
    <div class="spinner-md"></div>
    <p>{{ t('common.loading') }}</p>
  </div>

  <div v-else class="state-loading-screen">
    <AlertTriangle class="icon-lg text-danger" />
    <p>Проект не найден</p>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import {
  Gamepad2,
  CheckCircle2,
  XCircle,
  AlertTriangle,
  Eye,
  Image,
  Video,
  ExternalLink,
} from 'lucide-vue-next';
import {
  moderationApi,
  moderationStore,
  getStatusText,
  getStatusBadgeClass,
  REQUEST_STATUS,
  formatDateTime,
  ProjectChat,
} from '@/entities/moderation';
import { getProject, getMediaUrl } from '@/entities/project';
import { ApproveRequestModal, RejectRequestModal } from '@/features/review-request';
import { showToast } from '@/shared/lib';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const projectId = computed(() => route.params.projectId);

const activeRequest = ref(null);
const projectData = ref(null);
const loading = ref(true);
const actionLoading = ref(false);
const showApproveModal = ref(false);
const showRejectModal = ref(false);
const noRequestMode = ref(false);

const requestStatus = computed(() => activeRequest.value?.status);

const isPending = computed(() => {
  const st = requestStatus.value;
  return (
    st === REQUEST_STATUS.PENDING || st === 'REQUEST_STATUS_PENDING' || st === 1 || st === 'pending'
  );
});

const isInReview = computed(() => {
  const st = requestStatus.value;
  return (
    st === REQUEST_STATUS.IN_REVIEW ||
    st === 'REQUEST_STATUS_IN_REVIEW' ||
    st === 2 ||
    st === 'in_review'
  );
});

const isApproved = computed(() => {
  const st = requestStatus.value;
  return (
    st === REQUEST_STATUS.APPROVED ||
    st === 'REQUEST_STATUS_APPROVED' ||
    st === 3 ||
    st === 'approved'
  );
});

const isRejected = computed(() => {
  const st = requestStatus.value;
  return (
    st === REQUEST_STATUS.REJECTED ||
    st === 'REQUEST_STATUS_REJECTED' ||
    st === 4 ||
    st === 'rejected'
  );
});

const projectIconUrl = computed(() => {
  const path = projectData.value?.iconPath || projectData.value?.icon_path;
  if (!path) return '';
  return getMediaUrl(path);
});

const projectCoverUrl = computed(() => {
  const path = projectData.value?.coverPath || projectData.value?.cover_path;
  if (!path) return '';
  return getMediaUrl(path);
});

const projectVideoUrl = computed(() => {
  const path = projectData.value?.videoPath || projectData.value?.video_path;
  if (!path) return '';
  return getMediaUrl(path);
});

async function loadProjectInfo() {
  loading.value = true;
  noRequestMode.value = false;
  try {
    const data = await moderationApi.getLatestByProject(projectId.value);
    if (data && data.request) {
      activeRequest.value = data.request;
      projectData.value = data.request.snapshot || {};
    } else {
      noRequestMode.value = true;
      activeRequest.value = null;
      try {
        const p = await getProject(projectId.value);
        projectData.value = {
          titleRu: p.title_ru,
          titleEn: p.title_en,
          seoRu: p.seo_ru,
          seoEn: p.seo_en,
          aboutRu: p.about_ru || p.about,
          aboutEn: p.about_en,
          about: p.about_ru || p.about,
          iconPath: p.icon_path,
          coverPath: p.cover_path,
          videoPath: p.video_path,
          devUrl: p.dev_url,
          activeBuildVersion: p.active_build_version || '1.0.0',
        };
      } catch (err) {
        showToast('Проект не найден', 'warning');
        router.push('/moderator/queue');
      }
    }
  } catch (err) {
    console.error('Failed to load project request:', err);
    showToast('Не удалось загрузить данные проекта', 'danger');
  } finally {
    loading.value = false;
  }
}

function openDevPreview() {
  const devUrl = projectData.value?.devUrl || `/games/${projectId.value}/dev/index.html`;
  window.open(devUrl, '_blank');
}

async function handleClaim() {
  if (!activeRequest.value) return;
  actionLoading.value = true;
  try {
    const updated = await moderationStore.claimRequest(activeRequest.value.id);
    activeRequest.value = updated;
    showToast('Проект взят в работу', 'success');
  } catch (err) {
    showToast('Ошибка при взятии проекта в работу', 'danger');
  } finally {
    actionLoading.value = false;
  }
}

async function handleApprove(comment) {
  actionLoading.value = true;
  try {
    await moderationStore.approveRequest(projectId.value, comment);
    showApproveModal.value = false;
    showToast('Проект одобрен и опубликован', 'success');
    await loadProjectInfo();
  } catch (err) {
    showToast('Ошибка при одобрении проекта', 'danger');
  } finally {
    actionLoading.value = false;
  }
}

async function handleReject(reason) {
  actionLoading.value = true;
  try {
    await moderationStore.rejectRequest(projectId.value, reason);
    showRejectModal.value = false;
    showToast('Проект отклонен, отправлено уведомление', 'warning');
    await loadProjectInfo();
  } catch (err) {
    showToast('Ошибка при отклонении проекта', 'danger');
  } finally {
    actionLoading.value = false;
  }
}

onMounted(() => {
  loadProjectInfo();
});
</script>

<style scoped>
.moderator-workspace {
  display: flex;
  width: 100%;
  height: calc(100vh - 60px);
  background: var(--bg-app, #0d1117);
  overflow: hidden;
  box-sizing: border-box;
}

/* ЛЕВАЯ КОЛОНКА: ИНСПЕКТОР */
.review-inspector {
  flex: 1;
  min-width: 0;
  height: 100%;
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

/* ПРАВАЯ КОЛОНКА: ЧАТ */
.moderator-chat-aside {
  width: 440px;
  flex-shrink: 0;
  height: 100%;
  border-left: 1px solid var(--border, #30363d);
  background: var(--bg-card, #161b22);
}

/* КАРТОЧКИ */
.card {
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-md, 8px);
  padding: 20px;
  box-sizing: border-box;
}

/* Шапка проекта */
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
}

.identity-icon-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.identity-icon-mock {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
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

/* Секционные карточки */
.section-head {
  margin-bottom: 16px;
}

.section-head h3 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
  letter-spacing: 0.1px;
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

/* Медиа превью */
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
  gap: 6px;
  color: var(--text-tertiary, #8b949e);
  font-size: 12px;
}

/* Сборка и тестирование */
.build-test-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.build-info-block {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.build-version-tag {
  font-size: 13px;
  color: var(--text-main, #f0f6fc);
}

.build-desc {
  margin: 0;
  font-size: 12px;
  color: var(--text-tertiary, #8b949e);
  max-width: 500px;
  line-height: 1.4;
}

.btn-play-dev-lg {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  height: 38px;
  padding: 0 16px;
  background: var(--bg-secondary, #21262d);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  color: var(--primary, #58a6ff);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.15s;
  flex-shrink: 0;
}

.btn-play-dev-lg:hover {
  background: rgba(88, 166, 255, 0.1);
  border-color: var(--primary, #58a6ff);
}

/* Панель решений */
.verdict-actions-row {
  display: flex;
  gap: 12px;
}

.btn-verdict {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  height: 38px;
  padding: 0 18px;
  border-radius: var(--radius-sm, 6px);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s;
  border: none;
}

.btn-claim-ticket {
  background: var(--primary, #58a6ff);
  color: #ffffff;
}

.btn-claim-ticket:hover:not(:disabled) {
  background: var(--primary-hover, #79c0ff);
}

.btn-approve-ticket {
  background: #238636;
  color: #ffffff;
}

.btn-approve-ticket:hover:not(:disabled) {
  background: #2ea043;
}

.btn-reject-ticket {
  background: #da3633;
  color: #ffffff;
}

.btn-reject-ticket:hover:not(:disabled) {
  background: #f85149;
}

.btn-verdict:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.result-box {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 12px 16px;
  border-radius: var(--radius-sm, 6px);
  font-size: 13px;
}

.result-text strong {
  display: block;
  font-size: 14px;
  margin-bottom: 2px;
}

.result-text p {
  margin: 0;
  color: var(--text-secondary, #b0b8c4);
}

.box-rejected {
  background: rgba(218, 54, 51, 0.1);
  border: 1px solid rgba(218, 54, 51, 0.3);
  color: #f85149;
}

.box-approved {
  background: rgba(35, 134, 54, 0.1);
  border: 1px solid rgba(35, 134, 54, 0.3);
  color: #2ea043;
}

.state-loading-screen {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: calc(100vh - 60px);
  gap: 12px;
  color: var(--text-muted, #b0b8c4);
}

@media (max-width: 1000px) {
  .moderator-workspace {
    flex-direction: column;
    height: auto;
    overflow: visible;
  }
  .moderator-chat-aside {
    width: 100%;
    height: 500px;
  }
}
</style>
