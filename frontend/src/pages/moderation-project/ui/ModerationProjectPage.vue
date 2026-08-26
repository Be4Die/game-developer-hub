<template>
  <div class="project-detail-page" v-if="projectData">
    <!-- Верхняя панель навигации -->
    <div class="top-nav-bar">
      <button class="back-btn" @click="$router.push('/moderator/queue')">
        <ArrowLeft class="icon-sm" /> Назад к списку проектов
      </button>
      <div class="header-badges">
        <span class="badge" v-if="activeRequest" :class="getStatusBadgeClass(requestStatus)">
          {{ getStatusText(requestStatus) }}
        </span>
        <span class="badge badge-info" v-else>Чат (Без заявки)</span>
      </div>
    </div>

    <!-- Основной лейаут: Карточка игры слева, Чат справа -->
    <div class="project-layout">
      <!-- Левая колонка: Проверка черновика игры -->
      <div class="inspection-column">
        <!-- Карточка игры -->
        <div class="card game-card">
          <div class="game-header">
            <div>
              <h2 class="game-title">
                {{ projectData.titleRu || projectData.titleEn || `Проект #${projectId}` }}
              </h2>
              <div class="game-title-en" v-if="projectData.titleEn && projectData.titleRu">
                {{ projectData.titleEn }}
              </div>
            </div>
            <div class="game-version-badge" v-if="projectData.activeBuildVersion">
              v{{ projectData.activeBuildVersion }}
            </div>
          </div>

          <!-- Метаданные -->
          <div class="meta-grid">
            <div class="meta-field">
              <span class="meta-label">ID Проекта</span>
              <span class="meta-value">#{{ projectId }}</span>
            </div>
            <div class="meta-field">
              <span class="meta-label">Разработчик</span>
              <span class="meta-value">{{ activeRequest?.ownerId || '—' }}</span>
            </div>
            <div class="meta-field">
              <span class="meta-label">Отправлено на проверку</span>
              <span class="meta-value">
                {{ activeRequest ? formatDateTime(activeRequest.submittedAt) : 'Нет активной заявки' }}
              </span>
            </div>
            <div class="meta-field">
              <span class="meta-label">Модератор</span>
              <span class="meta-value">{{ activeRequest?.moderatorId || 'Не назначен' }}</span>
            </div>
          </div>

          <!-- Описание RU / EN -->
          <div class="seo-grid">
            <div class="content-block">
              <label class="block-label">Описание игры (RU)</label>
              <p class="block-text">{{ projectData.aboutRu || projectData.about || 'Описание не заполнено.' }}</p>
            </div>
            <div class="content-block">
              <label class="block-label">Описание игры (EN)</label>
              <p class="block-text">{{ projectData.aboutEn || 'Описание не заполнено.' }}</p>
            </div>
          </div>

          <!-- SEO описания -->
          <div class="seo-grid">
            <div class="content-block" v-if="projectData.seoRu">
              <label class="block-label">SEO (RU)</label>
              <p class="block-text">{{ projectData.seoRu }}</p>
            </div>
            <div class="content-block" v-if="projectData.seoEn">
              <label class="block-label">SEO (EN)</label>
              <p class="block-text">{{ projectData.seoEn }}</p>
            </div>
          </div>

          <!-- Медиафайлы (Иконка, Обложка) -->
          <div class="media-section">
            <label class="block-label">Медиаресурсы</label>
            <div class="media-preview-grid">
              <div class="media-preview-card" v-if="projectData.iconPath">
                <span class="media-tag">Иконка (512x512)</span>
                <img :src="projectData.iconPath" alt="Icon" class="media-img icon-img" />
              </div>
              <div class="media-preview-card" v-if="projectData.coverPath">
                <span class="media-tag">Обложка (800x470)</span>
                <img :src="projectData.coverPath" alt="Cover" class="media-img cover-img" />
              </div>
              <div class="no-media-text" v-if="!projectData.iconPath && !projectData.coverPath">
                Медиафайлы еще не загружены
              </div>
            </div>
          </div>

          <!-- Кнопка тестирования игры в Dev окружении -->
          <div class="test-game-block">
            <button class="btn-play-dev" @click="openDevPreview">
              <Gamepad2 class="icon-sm" /> Запустить тестовую сборку (Dev)
            </button>
          </div>
        </div>

        <!-- Баннеры вердиктов -->
        <div v-if="isRejected" class="card result-banner rejected">
          <AlertTriangle class="icon-md text-danger" />
          <div>
            <strong>Проект отклонен.</strong>
            <p class="rejection-reason" v-if="activeRequest?.rejectionReason">
              Причина: {{ activeRequest.rejectionReason }}
            </p>
          </div>
        </div>

        <div v-else-if="isApproved" class="card result-banner approved">
          <CheckCircle2 class="icon-md text-success" />
          <div>
            <strong>Проект одобрен и опубликован!</strong>
            <p class="subtext">Вы можете продолжать общаться с разработчиком в чате проекта.</p>
          </div>
        </div>

        <!-- Панель действий модератора -->
        <div class="card actions-card" v-if="!isApproved && !isRejected && activeRequest">
          <div class="actions-header">
            <h4>Действия модератора</h4>
          </div>

          <div class="action-buttons-stack">
            <!-- Кнопка "Взять в работу" -->
            <button
              v-if="isPending"
              class="btn-action btn-claim"
              @click="handleClaim"
              :disabled="actionLoading"
            >
              <Eye class="icon-sm" /> Приступить к проверке
            </button>

            <!-- Кнопки решения (когда в проверке) -->
            <template v-else-if="isInReview">
              <button
                class="btn-action btn-success"
                @click="showApproveModal = true"
                :disabled="actionLoading"
              >
                <CheckCircle2 class="icon-sm" /> Одобрить и опубликовать в Prod
              </button>

              <button
                class="btn-action btn-danger"
                @click="showRejectModal = true"
                :disabled="actionLoading"
              >
                <XCircle class="icon-sm" /> Отклонить проект (вернуть на доработку)
              </button>
            </template>
          </div>
        </div>
      </div>

      <!-- Правая колонка: Чат проекта -->
      <div class="chat-column">
        <ProjectChat :projectId="projectId" />
      </div>
    </div>

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

  <div v-else-if="loading" class="page-state card">
    <RefreshCw class="icon-md spinning text-primary" />
    <p>Загрузка данных проекта...</p>
  </div>

  <div v-else class="page-state card">
    <AlertTriangle class="icon-lg text-danger" />
    <p>Проект не найден</p>
    <button class="btn-outline" @click="$router.push('/moderator/queue')">
      Вернуться в список
    </button>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import {
  ArrowLeft,
  Gamepad2,
  CheckCircle2,
  XCircle,
  AlertTriangle,
  Eye,
  RefreshCw,
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
import { getProject } from '@/entities/project';
import {
  ApproveRequestModal,
  RejectRequestModal,
} from '@/features/review-request';
import { showToast } from '@/shared/lib';

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
    st === REQUEST_STATUS.PENDING ||
    st === 'REQUEST_STATUS_PENDING' ||
    st === 1 ||
    st === 'pending'
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

async function loadProjectInfo() {
  loading.value = true;
  noRequestMode.value = false;
  try {
    const data = await moderationApi.getLatestByProject(projectId.value);
    if (data && data.request) {
      activeRequest.value = data.request;
      // Используем snapshot из заявки как источник данных для ревью
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
  const devUrl =
    projectData.value?.devUrl ||
    `/games/${projectId.value}/dev/index.html`;
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
.project-detail-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.top-nav-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.back-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  background: transparent;
  border: none;
  color: var(--text-secondary);
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  padding: 6px 0;
  transition: color 0.15s;
}

.back-btn:hover {
  color: var(--primary);
}

.header-badges {
  display: flex;
  align-items: center;
  gap: 10px;
}

.project-layout {
  display: grid;
  grid-template-columns: 1.15fr 0.85fr;
  gap: 24px;
}

.inspection-column {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.chat-column {
  display: flex;
  flex-direction: column;
}

.game-card {
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.game-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  border-bottom: 1px solid var(--border);
  padding-bottom: 16px;
}

.game-title {
  margin: 0;
  font-size: 1.4rem;
  font-weight: 700;
  color: var(--text-main);
}

.game-title-en {
  font-size: 0.9rem;
  color: var(--text-muted);
  margin-top: 4px;
}

.game-version-badge {
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  font-weight: 600;
  font-size: 0.82rem;
  padding: 4px 10px;
  border-radius: var(--radius-sm);
  color: var(--primary);
}

.meta-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  background: var(--bg-secondary);
  padding: 14px 16px;
  border-radius: var(--radius-md);
}

.meta-field {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.meta-label {
  font-size: 0.75rem;
  color: var(--text-tertiary);
  font-weight: 500;
}

.meta-value {
  font-size: 0.88rem;
  font-weight: 600;
  color: var(--text-main);
}

.block-label {
  display: block;
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--text-secondary);
  margin-bottom: 6px;
}

.block-text {
  margin: 0;
  font-size: 0.9rem;
  color: var(--text-main);
  line-height: 1.5;
  white-space: pre-wrap;
}

.seo-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.media-section {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.media-preview-grid {
  display: flex;
  gap: 16px;
  flex-wrap: wrap;
}

.media-preview-card {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.media-tag {
  font-size: 0.72rem;
  color: var(--text-muted);
}

.media-img {
  border-radius: var(--radius-md);
  border: 1px solid var(--border);
  object-fit: cover;
  background: var(--bg-secondary);
}

.icon-img {
  width: 72px;
  height: 72px;
}

.cover-img {
  width: 140px;
  height: 82px;
}

.no-media-text {
  font-size: 0.85rem;
  color: var(--text-tertiary);
  font-style: italic;
}

.test-game-block {
  padding-top: 10px;
  border-top: 1px solid var(--border);
}

.btn-play-dev {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  width: 100%;
  padding: 12px;
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  color: var(--primary);
  font-weight: 600;
  cursor: pointer;
  transition: 0.2s;
}

.btn-play-dev:hover {
  background: var(--bg-hover);
  border-color: var(--primary);
}

.result-banner {
  padding: 16px 20px;
  display: flex;
  align-items: center;
  gap: 14px;
}

.result-banner.approved {
  border-left: 4px solid var(--success);
  background: rgba(16, 185, 129, 0.08);
}

.result-banner.rejected {
  border-left: 4px solid var(--danger, #ef4444);
  background: rgba(239, 68, 68, 0.08);
}

.rejection-reason {
  margin: 4px 0 0;
  font-size: 0.88rem;
  color: var(--text-secondary);
}

.actions-card {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.actions-header h4 {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
}

.action-buttons-stack {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.btn-action {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 12px 18px;
  border-radius: var(--radius-md);
  font-weight: 600;
  font-size: 0.92rem;
  cursor: pointer;
  border: none;
  transition: opacity 0.15s;
}

.btn-action:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-claim {
  background: var(--primary);
  color: white;
}

.btn-success {
  background: var(--success, #10b981);
  color: white;
}

.btn-danger {
  background: var(--danger, #ef4444);
  color: white;
}

.page-state {
  padding: 64px 32px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  text-align: center;
}

.spinning {
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
