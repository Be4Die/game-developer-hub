<template>
  <div class="page-container">
    <div class="header-row">
      <div>
        <div class="page-subtitle">Модерация</div>
        <h1>Панель модератора</h1>
      </div>
      <div class="header-actions">
        <button class="btn-outline" @click="loadData">
          <RefreshCw class="icon-sm" :class="{ spinning: loading }" /> Обновить
        </button>
        <button class="btn-primary" @click="$router.push('/moderator/queue')">
          <ListOrdered class="icon-sm" /> Очередь проектов
        </button>
      </div>
    </div>

    <!-- Метрики -->
    <div class="metrics-grid">
      <div class="metric-card card">
        <div class="metric-icon-wrap bg-warning-light">
          <Clock class="icon-md text-warning" />
        </div>
        <div class="metric-content">
          <div class="metric-value">{{ pendingRequests.length }}</div>
          <div class="metric-label">Новые проекты</div>
        </div>
      </div>

      <div class="metric-card card">
        <div class="metric-icon-wrap bg-info-light">
          <Eye class="icon-md text-info" />
        </div>
        <div class="metric-content">
          <div class="metric-value">{{ inReviewRequests.length }}</div>
          <div class="metric-label">В проверке</div>
        </div>
      </div>

      <div class="metric-card card">
        <div class="metric-icon-wrap bg-success-light">
          <CheckCircle class="icon-md text-success" />
        </div>
        <div class="metric-content">
          <div class="metric-value">{{ approvedCount }}</div>
          <div class="metric-label">Одобрено</div>
        </div>
      </div>
    </div>

    <!-- Секции дашборда -->
    <div class="dashboard-grid">
      <!-- Активная очередь новых заявок -->
      <div class="dashboard-section">
        <div class="section-header">
          <div class="section-title-group">
            <Inbox class="icon-sm text-warning" />
            <h2>Новые проекты</h2>
          </div>
          <span class="badge badge-warning" v-if="pendingRequests.length > 0">
            {{ pendingRequests.length }}
          </span>
        </div>

        <div class="requests-list" v-if="pendingRequests.length > 0">
          <div
            v-for="req in pendingRequests"
            :key="req.projectId"
            class="request-card card card-hover"
            @click="openProject(req.projectId)"
          >
            <div class="request-header">
              <div class="title-with-badge">
                <span class="badge" :class="getStatusBadgeClass(req.status)">
                  {{ getStatusText(req.status) }}
                </span>
              </div>
              <span class="request-time">{{ formatDateTime(req.submittedAt) }}</span>
            </div>

            <h3 class="request-title">
              {{ req.snapshot.titleRu || req.snapshot.titleEn || `Проект #${req.projectId}` }}
            </h3>
            <p class="request-desc">{{ req.snapshot.about || 'Описание не указано' }}</p>

            <div class="request-meta-row">
              <span class="meta-tag">Проект: #{{ req.projectId }}</span>
              <span class="meta-tag" v-if="req.snapshot.activeBuildVersion">
                Версия: v{{ req.snapshot.activeBuildVersion }}
              </span>
              <button
                class="btn-claim-inline"
                @click.stop="openProject(req.projectId)"
              >
                К проверке →
              </button>
            </div>
          </div>
        </div>

        <div class="empty-state card" v-else>
          <CheckCircle2 class="icon-lg text-success" />
          <p>Очередь пуста</p>
          <span class="subtext">Все новые проекты проверены или взяты в работу</span>
        </div>
      </div>

      <!-- Заявки в проверке -->
      <div class="dashboard-section">
        <div class="section-header">
          <div class="section-title-group">
            <Eye class="icon-sm text-info" />
            <h2>Проекты в проверке</h2>
          </div>
          <span class="badge badge-info" v-if="inReviewRequests.length > 0">
            {{ inReviewRequests.length }}
          </span>
        </div>

        <div class="requests-list" v-if="inReviewRequests.length > 0">
          <div
            v-for="req in inReviewRequests"
            :key="req.projectId"
            class="request-card card card-hover in-review-card"
            @click="openProject(req.projectId)"
          >
            <div class="request-header">
              <div class="title-with-badge">
                <span class="badge badge-info">В проверке</span>
              </div>
              <span class="request-time">{{ formatDateTime(req.submittedAt) }}</span>
            </div>

            <h3 class="request-title">
              {{ req.snapshot.titleRu || req.snapshot.titleEn || `Проект #${req.projectId}` }}
            </h3>
            <p class="request-desc">{{ req.snapshot.about || 'Описание не указано' }}</p>

            <div class="request-meta-row">
              <span class="meta-tag">Проект: #{{ req.projectId }}</span>
              <span class="meta-tag">Модератор: {{ req.moderatorId || 'Вы' }}</span>
              <button class="btn-inspect-inline" @click.stop="openProject(req.projectId)">
                Продолжить →
              </button>
            </div>
          </div>
        </div>

        <div class="empty-state card" v-else>
          <Inbox class="icon-lg text-muted" />
          <p>Нет проектов в проверке</p>
          <span class="subtext">Возьмите проект из очереди слева для проведения проверки</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import {
  Inbox,
  Clock,
  Eye,
  CheckCircle,
  CheckCircle2,
  RefreshCw,
  ListOrdered,
} from 'lucide-vue-next';
import {
  moderationStore,
  getStatusText,
  getStatusBadgeClass,
  REQUEST_STATUS,
  formatDateTime,
} from '@/entities/moderation';

const router = useRouter();
const loading = ref(false);

const pendingRequests = computed(() =>
  moderationStore.requests.filter(
    (r) =>
      r.status === REQUEST_STATUS.PENDING ||
      r.status === 'REQUEST_STATUS_PENDING' ||
      r.status === 1 ||
      r.status === 'pending'
  )
);

const inReviewRequests = computed(() =>
  moderationStore.requests.filter(
    (r) =>
      r.status === REQUEST_STATUS.IN_REVIEW ||
      r.status === 'REQUEST_STATUS_IN_REVIEW' ||
      r.status === 2 ||
      r.status === 'in_review'
  )
);

const approvedCount = computed(
  () =>
    moderationStore.requests.filter(
      (r) =>
        r.status === REQUEST_STATUS.APPROVED ||
        r.status === 'REQUEST_STATUS_APPROVED' ||
        r.status === 3 ||
        r.status === 'approved'
    ).length
);

async function loadData() {
  loading.value = true;
  try {
    await moderationStore.loadRequests({ limit: 100 });
  } finally {
    loading.value = false;
  }
}

function openProject(projectId) {
  router.push(`/moderator/projects/${projectId}`);
}

onMounted(() => {
  loadData();
});
</script>

<style scoped>
.page-subtitle {
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: var(--text-tertiary);
  margin-bottom: 4px;
}

.header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.header-row h1 {
  font-size: 1.5rem;
  font-weight: 700;
  letter-spacing: -0.5px;
  margin: 0;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.metrics-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}

.metric-card {
  padding: 18px 20px;
  display: flex;
  align-items: center;
  gap: 16px;
}

.metric-icon-wrap {
  width: 44px;
  height: 44px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.bg-warning-light {
  background: rgba(245, 158, 11, 0.12);
}

.bg-info-light {
  background: rgba(59, 130, 246, 0.12);
}

.bg-success-light {
  background: rgba(16, 185, 129, 0.12);
}

.text-warning {
  color: var(--warning, #f59e0b);
}

.text-info {
  color: var(--info, #3b82f6);
}

.text-success {
  color: var(--success, #10b981);
}

.metric-value {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--text-main);
  line-height: 1.2;
}

.metric-label {
  font-size: 0.82rem;
  color: var(--text-muted);
  font-weight: 500;
}

.dashboard-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 24px;
}

.dashboard-section {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.section-title-group {
  display: flex;
  align-items: center;
  gap: 8px;
}

.section-header h2 {
  font-size: 1.05rem;
  font-weight: 600;
  margin: 0;
}

.requests-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.request-card {
  cursor: pointer;
  padding: 18px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  transition: transform 0.15s, border-color 0.15s;
}

.request-card:hover {
  transform: translateY(-2px);
  border-color: var(--primary);
}

.request-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.title-with-badge {
  display: flex;
  align-items: center;
  gap: 8px;
}

.request-id {
  font-weight: 700;
  font-family: monospace;
  font-size: 0.85rem;
  color: var(--text-tertiary);
}

.request-time {
  font-size: 0.75rem;
  color: var(--text-tertiary);
}

.request-title {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-main);
}

.request-desc {
  margin: 0;
  font-size: 0.85rem;
  color: var(--text-muted);
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.request-meta-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 4px;
}

.meta-tag {
  font-size: 0.75rem;
  color: var(--text-tertiary);
  background: var(--bg-secondary);
  padding: 2px 6px;
  border-radius: var(--radius-sm);
}

.btn-claim-inline {
  margin-left: auto;
  padding: 4px 10px;
  font-size: 0.78rem;
  font-weight: 600;
  border-radius: var(--radius-sm);
  background: var(--primary);
  color: white;
  border: none;
  cursor: pointer;
}

.btn-claim-inline:hover {
  opacity: 0.9;
}

.btn-inspect-inline {
  margin-left: auto;
  padding: 4px 10px;
  font-size: 0.78rem;
  font-weight: 600;
  border-radius: var(--radius-sm);
  background: var(--bg-secondary);
  color: var(--primary);
  border: 1px solid var(--border);
  cursor: pointer;
}

.btn-inspect-inline:hover {
  border-color: var(--primary);
}

.empty-state {
  padding: 40px 24px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  text-align: center;
  color: var(--text-tertiary);
}

.empty-state p {
  margin: 0;
  font-weight: 600;
  color: var(--text-secondary);
}

.empty-state .subtext {
  font-size: 0.8rem;
  max-width: 280px;
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
