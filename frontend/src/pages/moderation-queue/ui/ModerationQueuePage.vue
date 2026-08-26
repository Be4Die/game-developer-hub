<template>
  <div class="page-container">
    <div class="header-row">
      <div>
        <div class="page-subtitle">{{ t('header.moderation') }}</div>
        <h1>{{ t('moderation.queueTitle') }}</h1>
      </div>
      <div class="header-actions">
        <button class="btn-outline" @click="loadData">
          <RefreshCw class="icon-sm" :class="{ spinning: loading }" /> {{ t('common.refresh') }}
        </button>
      </div>
    </div>

    <!-- Фильтры и поиск -->
    <div class="filter-toolbar card">
      <div class="tabs-group">
        <button
          v-for="tab in filterTabs"
          :key="tab.value"
          class="tab-btn"
          :class="{ active: currentTab === tab.value }"
          @click="currentTab = tab.value"
        >
          {{ tab.label }}
          <span class="tab-badge" v-if="tab.count !== undefined">{{ tab.count }}</span>
        </button>
      </div>

      <div class="search-box">
        <Search class="icon-sm search-icon" />
        <input
          type="text"
          v-model="searchQuery"
          :placeholder="t('projects.searchPlaceholder')"
          class="search-input"
        />
      </div>
    </div>


    <!-- Список проектов -->
    <div v-if="loading && !filteredRequests.length" class="loading-state card">
      <RefreshCw class="icon-md spinning text-primary" />
      <p>Загрузка очереди проектов...</p>
    </div>

    <div v-else-if="!filteredRequests.length && !isChatTab" class="empty-state card">
      <Inbox class="icon-lg text-muted" />
      <p>Проектов не найдено</p>
      <span class="subtext">В выбранной категории сейчас нет проектов</span>
    </div>

    <!-- Список чатов -->
    <div v-if="isChatTab" class="requests-grid">
      <div v-if="!filteredChats.length && !loading" class="empty-state card">
        <Inbox class="icon-lg text-muted" />
        <p>Активных обсуждений нет</p>
      </div>

      <div
        v-for="chat in filteredChats"
        :key="chat.projectId"
        class="card request-card card-hover"
        @click="openProject(chat.projectId)"
      >
        <div class="request-card-header">
          <div class="id-and-status">
            <span class="badge badge-info">Новое сообщение</span>
          </div>
          <span class="date-text">
            Обновлено: {{ formatDateTime(chat.lastMessage?.createdAt) }}
          </span>
        </div>
        <div class="request-body">
          <h3 class="game-title">Проект #{{ chat.projectId }}</h3>
          <p class="game-desc">{{ chat.lastMessage?.content || 'Без текста' }}</p>
        </div>
        <div class="request-card-footer">
          <button class="btn-primary btn-sm" @click.stop="openProject(chat.projectId)">
            Открыть чат →
          </button>
        </div>
      </div>
    </div>

    <!-- Список проектов (заявок) -->
    <div v-else class="requests-grid">
      <div
        v-for="req in filteredRequests"
        :key="req.projectId"
        class="card request-card card-hover"
        @click="openProject(req.projectId)"
      >
        <div class="request-card-header">
          <div class="id-and-status">
            <span class="badge" :class="getStatusBadgeClass(req.status)">
              {{ getStatusText(req.status) }}
            </span>
          </div>
          <span class="date-text">Отправлено: {{ formatDateTime(req.submittedAt) }}</span>
        </div>

        <div class="request-body">
          <div class="game-meta-group">
            <h3 class="game-title">
              {{ req.snapshot.titleRu || req.snapshot.titleEn || `Проект #${req.projectId}` }}
            </h3>
            <div class="title-en" v-if="req.snapshot.titleEn && req.snapshot.titleRu">
              {{ req.snapshot.titleEn }}
            </div>
          </div>
          <p class="game-desc">{{ req.snapshot.aboutRu || req.snapshot.aboutEn || req.snapshot.about || 'Описание не заполнено' }}</p>
        </div>

        <div class="request-card-footer">
          <div class="tags-group">
            <span class="tag">Проект #{{ req.projectId }}</span>
            <span class="tag" v-if="req.snapshot.activeBuildVersion">
              Сборка v{{ req.snapshot.activeBuildVersion }}
            </span>
            <span class="tag" v-if="req.ownerId">
              Разработчик: {{ req.ownerId }}
            </span>
          </div>

          <div class="footer-actions">
            <button
              class="btn-primary btn-sm"
              @click.stop="openProject(req.projectId)"
            >
              Перейти к проверке →
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { RefreshCw, Search, Inbox } from 'lucide-vue-next';
import {
  moderationStore,
  getStatusText,
  getStatusBadgeClass,
  REQUEST_STATUS,
  formatDateTime,
} from '@/entities/moderation';

const { t } = useI18n();
const router = useRouter();
const loading = ref(false);
const currentTab = ref('pending');
const searchQuery = ref('');

const allRequests = computed(() => moderationStore.requests);
const activeChats = computed(() => moderationStore.activeChats);

const filterTabs = computed(() => {
  const pending = allRequests.value.filter((r) => isPending(r.status)).length;
  const inReview = allRequests.value.filter((r) => isInReview(r.status)).length;
  const approved = allRequests.value.filter((r) => isApproved(r.status)).length;
  const rejected = allRequests.value.filter((r) => isRejected(r.status)).length;

  return [
    { label: t('projects.draft'), value: 'pending', count: pending },
    { label: t('projects.moderation'), value: 'in_review', count: inReview },
    { label: t('projects.approved'), value: 'approved', count: approved },
    { label: t('projects.rejected'), value: 'rejected', count: rejected },
    { label: t('projects.allStatuses'), value: 'all', count: allRequests.value.length },
    { label: t('moderation.chatTitle'), value: 'chats', count: activeChats.value.length },
  ];
});


const isChatTab = computed(() => currentTab.value === 'chats');

const filteredRequests = computed(() => {
  if (isChatTab.value) return [];
  let list = allRequests.value;

  if (currentTab.value === 'pending') {
    list = list.filter((r) => isPending(r.status));
  } else if (currentTab.value === 'in_review') {
    list = list.filter((r) => isInReview(r.status));
  } else if (currentTab.value === 'approved') {
    list = list.filter((r) => isApproved(r.status));
  } else if (currentTab.value === 'rejected') {
    list = list.filter((r) => isRejected(r.status));
  }

  const q = searchQuery.value.trim().toLowerCase();
  if (q) {
    list = list.filter((r) => {
      const matchId = String(r.projectId).includes(q);
      const matchTitleRu = (r.snapshot.titleRu || '').toLowerCase().includes(q);
      const matchTitleEn = (r.snapshot.titleEn || '').toLowerCase().includes(q);
      return matchId || matchTitleRu || matchTitleEn;
    });
  }

  return list;
});

const filteredChats = computed(() => {
  if (!isChatTab.value) return [];
  let list = activeChats.value;
  const q = searchQuery.value.trim().toLowerCase();
  if (q) {
    list = list.filter((c) => {
      return String(c.projectId).includes(q);
    });
  }
  return list;
});

function isPending(status) {
  return (
    status === REQUEST_STATUS.PENDING ||
    status === 'REQUEST_STATUS_PENDING' ||
    status === 1 ||
    status === 'pending'
  );
}

function isInReview(status) {
  return (
    status === REQUEST_STATUS.IN_REVIEW ||
    status === 'REQUEST_STATUS_IN_REVIEW' ||
    status === 2 ||
    status === 'in_review'
  );
}

function isApproved(status) {
  return (
    status === REQUEST_STATUS.APPROVED ||
    status === 'REQUEST_STATUS_APPROVED' ||
    status === 3 ||
    status === 'approved'
  );
}

function isRejected(status) {
  return (
    status === REQUEST_STATUS.REJECTED ||
    status === 'REQUEST_STATUS_REJECTED' ||
    status === 4 ||
    status === 'rejected'
  );
}

async function loadData() {
  loading.value = true;
  try {
    await Promise.all([
      moderationStore.loadRequests({ limit: 100 }),
      moderationStore.loadActiveChats({ limit: 100 })
    ]);
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
  margin-bottom: 20px;
}

.header-row h1 {
  font-size: 1.5rem;
  font-weight: 700;
  letter-spacing: -0.5px;
  margin: 0;
}

.filter-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  margin-bottom: 20px;
  gap: 16px;
  flex-wrap: wrap;
}

.tabs-group {
  display: flex;
  gap: 6px;
}

.tab-btn {
  padding: 6px 14px;
  border-radius: var(--radius-md);
  border: 1px solid transparent;
  background: transparent;
  color: var(--text-secondary);
  font-size: 0.88rem;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
  transition: 0.15s;
}

.tab-btn:hover {
  background: var(--bg-hover);
  color: var(--text-main);
}

.tab-btn.active {
  background: var(--bg-secondary);
  border-color: var(--border);
  color: var(--primary);
  font-weight: 600;
}

.tab-badge {
  font-size: 0.72rem;
  padding: 1px 6px;
  border-radius: 10px;
  background: var(--bg-hover);
  color: var(--text-muted);
}

.tab-btn.active .tab-badge {
  background: var(--primary);
  color: white;
}

.search-box {
  position: relative;
  min-width: 280px;
}

.search-icon {
  position: absolute;
  left: 10px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--text-tertiary);
}

.search-input {
  width: 100%;
  padding: 8px 12px 8px 34px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  background: var(--bg-app);
  color: var(--text-main);
  font-size: 0.88rem;
  box-sizing: border-box;
}

.search-input:focus {
  outline: none;
  border-color: var(--primary);
}

.requests-grid {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.request-card {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  cursor: pointer;
  transition: transform 0.15s, border-color 0.15s;
}

.request-card:hover {
  transform: translateY(-2px);
  border-color: var(--primary);
}

.request-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.id-and-status {
  display: flex;
  align-items: center;
  gap: 10px;
}

.date-text {
  font-size: 0.78rem;
  color: var(--text-tertiary);
}

.game-title {
  margin: 0;
  font-size: 1.1rem;
  font-weight: 700;
  color: var(--text-main);
}

.title-en {
  font-size: 0.85rem;
  color: var(--text-muted);
  margin-top: 2px;
}

.game-desc {
  margin: 6px 0 0 0;
  font-size: 0.88rem;
  color: var(--text-secondary);
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.request-card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 4px;
  padding-top: 10px;
  border-top: 1px solid var(--border);
}

.tags-group {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.tag {
  font-size: 0.75rem;
  padding: 3px 8px;
  border-radius: var(--radius-sm);
  background: var(--bg-secondary);
  color: var(--text-tertiary);
}

.btn-sm {
  padding: 6px 14px;
  font-size: 0.82rem;
  font-weight: 600;
  border-radius: var(--radius-md);
}

.loading-state,
.empty-state {
  padding: 48px;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: var(--text-tertiary);
}

.empty-state p,
.loading-state p {
  margin: 0;
  font-weight: 600;
  color: var(--text-secondary);
}

.empty-state .subtext {
  font-size: 0.82rem;
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
