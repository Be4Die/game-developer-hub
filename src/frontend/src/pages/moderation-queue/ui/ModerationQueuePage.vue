<template>
  <div class="moderation-page-container">
    <div class="main-content-wrap">
      <!-- Панель фильтрации и поиска (компактный стиль консоли) -->
      <div class="filters-toolbar">
        <!-- Поиск по названию, ID, разработчику -->
        <div class="filter-field field-search">
          <label class="field-label">{{ t('common.search') }}</label>
          <div class="input-wrapper">
            <input
              v-model="searchQuery"
              type="text"
              :placeholder="t('moderation.searchPlaceholder')"
              class="filter-input"
            />
            <button
              v-if="searchQuery"
              class="clear-input-btn"
              title="Очистить"
              @click="searchQuery = ''"
            >
              <X class="icon-xs" />
            </button>
          </div>
        </div>

        <!-- Фильтр по статусу -->
        <div class="filter-field field-status">
          <label class="field-label">{{ t('common.status') }}</label>
          <div class="select-wrapper">
            <select v-model="statusFilter" class="filter-select">
              <option value="all">{{ t('moderation.allActive') }}</option>
              <option value="pending">{{ t('moderation.pending') }}</option>
              <option value="in_review">{{ t('moderation.inReview') }}</option>
              <option value="my">{{ t('moderation.assignedToMe') }}</option>
            </select>
            <ChevronDown class="icon-xs select-arrow" />
          </div>
        </div>

        <!-- Сортировка -->
        <div class="filter-field field-sort">
          <label class="field-label">{{ t('common.actions') }}</label>
          <div class="select-wrapper">
            <select v-model="sortBy" class="filter-select">
              <option value="newest">{{ t('stats.today') }} / {{ t('common.created') }} ↓</option>
              <option value="oldest">{{ t('common.created') }} ↑</option>
              <option value="title">{{ t('common.name') }} (A–Z)</option>
            </select>
            <ChevronDown class="icon-xs select-arrow" />
          </div>
        </div>

        <!-- Кнопка сброса фильтров -->
        <button
          v-if="searchQuery || statusFilter !== 'all' || sortBy !== 'newest'"
          class="btn-reset-filters"
          title="Сбросить фильтры"
          @click="resetFilters"
        >
          <RotateCcw class="icon-xs" />
          <span>{{ t('common.reset') }}</span>
        </button>

        <!-- Кнопка обновления -->
        <button class="btn-refresh" :disabled="loading" title="Обновить" @click="loadQueue">
          <RefreshCw class="icon-xs" :class="{ spin: loading }" />
          <span>{{ t('common.refresh') }}</span>
        </button>
      </div>

      <!-- Состояние загрузки -->
      <div v-if="loading" class="state-container">
        <div class="spinner-md"></div>
        <p>{{ t('common.loading') }}</p>
      </div>

      <!-- Пустой список заявок -->
      <div
        v-else-if="requests.length === 0 && !searchQuery && statusFilter === 'all'"
        class="state-container empty-card"
      >
        <div class="empty-icon-wrap">
          <CheckCircle2 class="icon-lg text-success" />
        </div>
        <h3>{{ t('moderation.noActiveRequests') }}</h3>
        <p>{{ t('moderation.emptyQueue') }}</p>
        <button class="btn-primary-sm" @click="loadQueue">
          <RefreshCw class="icon-xs" />
          <span>{{ t('common.refresh') }}</span>
        </button>
      </div>

      <!-- Пустой список по результатам поиска/фильтров -->
      <div v-else-if="filteredRequests.length === 0" class="state-container empty-card">
        <Search class="icon-md text-muted" />
        <h3>{{ t('common.empty') }}</h3>
        <p>{{ t('stats.noData') }}</p>
        <button class="btn-reset-filters" @click="resetFilters">
          {{ t('common.reset') }}
        </button>
      </div>

      <!-- Таблица активных заявок на модерацию -->
      <div v-else class="table-wrapper">
        <table class="moderation-table">
          <thead>
            <tr>
              <th class="col-game">{{ t('moderation.projectColumn') }}</th>
              <th class="col-version">{{ t('common.version') }}</th>
              <th class="col-dev">{{ t('moderation.developerColumn') }}</th>
              <th class="col-date">{{ t('moderation.submittedColumn') }}</th>
              <th class="col-status">{{ t('common.status') }}</th>
              <th class="col-mod">{{ t('moderation.moderator') }}</th>
              <th class="col-actions"></th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="req in paginatedRequests"
              :key="req.id"
              class="table-row"
              @click="openProject(req.projectId)"
            >
              <!-- 1 колонка: Игра / Проект -->
              <td class="col-game">
                <div class="game-cell">
                  <div class="game-icon-box">
                    <img
                      v-if="req.snapshot.iconPath"
                      :src="getMediaUrl(req.snapshot.iconPath)"
                      alt="Icon"
                      class="game-icon-img"
                    />
                    <div v-else class="game-icon-mock">
                      <span>Draft</span>
                    </div>
                  </div>
                  <div class="game-text">
                    <div class="game-type-label">Проект #{{ req.projectId }}</div>
                    <div class="game-title">
                      {{
                        req.snapshot.titleRu || req.snapshot.titleEn || `Проект #${req.projectId}`
                      }}
                    </div>
                  </div>
                </div>
              </td>

              <!-- 2 колонка: Версия сборки -->
              <td class="col-version">
                <span v-if="req.snapshot.activeBuildVersion" class="version-badge">
                  v{{ req.snapshot.activeBuildVersion }}
                </span>
                <span v-else class="text-muted text-sm">—</span>
              </td>

              <!-- 3 колонка: Разработчик -->
              <td class="col-dev">
                <div class="dev-cell" :title="req.ownerId">
                  <User class="icon-xs text-muted" />
                  <span class="dev-name">{{ req.ownerId || '—' }}</span>
                </div>
              </td>

              <!-- 4 колонка: Дата подачи -->
              <td class="col-date">
                <span class="date-text">
                  {{ formatDateTime(req.submittedAt) }}
                </span>
              </td>

              <!-- 5 колонка: Статус -->
              <td class="col-status">
                <span class="status-pill" :class="reqStatusClass(req.status)">
                  {{ reqStatusLabel(req.status) }}
                </span>
              </td>

              <!-- 6 колонка: Модератор -->
              <td class="col-mod">
                <span v-if="req.moderatorId" class="mod-name">
                  {{ req.moderatorId }}
                </span>
                <span v-else class="unassigned-text">
                  {{ t('moderation.notAssigned') }}
                </span>
              </td>

              <!-- 7 колонка: Действия -->
              <td class="col-actions" @click.stop>
                <div class="row-actions">
                  <template v-if="isAdmin">
                    <button
                      class="btn-inspect-sm"
                      title="Просмотр проекта"
                      @click="openProject(req.projectId)"
                    >
                      <Eye class="icon-xs" />
                      <span>Просмотр</span>
                    </button>
                  </template>
                  <template v-else>
                    <button
                      v-if="isPending(req.status)"
                      class="btn-claim-sm"
                      :disabled="claimingId === req.id"
                      :title="t('moderation.claimBtn')"
                      @click="claimAndOpen(req)"
                    >
                      <Loader2 v-if="claimingId === req.id" class="icon-xs spin" />
                      <CheckSquare v-else class="icon-xs" />
                      <span>{{ t('moderation.claimBtn') }}</span>
                    </button>

                    <button
                      v-else
                      class="btn-inspect-sm"
                      :title="t('moderation.continueBtn')"
                      @click="openProject(req.projectId)"
                    >
                      <ArrowRight class="icon-xs" />
                      <span>{{ t('moderation.continueBtn') }}</span>
                    </button>
                  </template>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Пагинация -->
      <div v-if="totalPages > 1" class="pagination-bar">
        <span class="page-info">
          {{ (currentPage - 1) * pageSize + 1 }}–{{
            Math.min(currentPage * pageSize, filteredRequests.length)
          }}
          из {{ filteredRequests.length }}
        </span>

        <div class="page-nav">
          <button class="page-nav-btn" :disabled="currentPage === 1" @click="currentPage = 1">
            <ChevronsLeft class="icon-sm" />
          </button>
          <button class="page-nav-btn" :disabled="currentPage === 1" @click="currentPage--">
            <ChevronLeft class="icon-sm" />
          </button>
          <span class="page-current">{{ currentPage }} / {{ totalPages }}</span>
          <button
            class="page-nav-btn"
            :disabled="currentPage === totalPages"
            @click="currentPage++"
          >
            <ChevronRight class="icon-sm" />
          </button>
          <button
            class="page-nav-btn"
            :disabled="currentPage === totalPages"
            @click="currentPage = totalPages"
          >
            <ChevronsRight class="icon-sm" />
          </button>

          <div class="page-size-wrap">
            <select v-model="pageSize" class="page-size-select">
              <option :value="10">10</option>
              <option :value="25">25</option>
              <option :value="50">50</option>
            </select>
            <ChevronDown class="icon-xs select-caret" />
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
import {
  Search,
  X,
  ChevronDown,
  RotateCcw,
  RefreshCw,
  CheckCircle2,
  User,
  CheckSquare,
  ArrowRight,
  Loader2,
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
  Eye,
} from 'lucide-vue-next';
import {
  moderationApi,
  normalizeRequest,
  formatDateTime,
  REQUEST_STATUS,
} from '@/entities/moderation';
import { getMediaUrl } from '@/entities/project';
import { useAuth } from '@/entities/user';
import { showToast } from '@/shared/lib';

const { t } = useI18n();
const router = useRouter();
const { state: authState } = useAuth();

const isAdmin = computed(() => {
  const r = authState.user?.role;
  return r === 'USER_ROLE_ADMIN' || r === 'admin' || r === 3;
});

const currentUserId = computed(() => authState.user?.id || authState.user?.email || '');

const requests = ref([]);
const loading = ref(false);
const claimingId = ref(null);

const searchQuery = ref('');
const statusFilter = ref('all');
const sortBy = ref('newest');
const currentPage = ref(1);
const pageSize = ref(10);

async function loadQueue() {
  loading.value = true;
  try {
    const res = await moderationApi.listRequests({ limit: 100, offset: 0 });
    const raw = res.requests || [];
    // Отбираем только активные заявки (Pending=1 или In Review=2)
    const active = raw
      .map(normalizeRequest)
      .filter(
        (r) =>
          r.status === REQUEST_STATUS.PENDING ||
          r.status === REQUEST_STATUS.IN_REVIEW ||
          r.status === 1 ||
          r.status === 2 ||
          r.status === 'REQUEST_STATUS_PENDING' ||
          r.status === 'REQUEST_STATUS_IN_REVIEW' ||
          r.status === 'pending' ||
          r.status === 'in_review'
      );
    requests.value = active;
  } catch (err) {
    showToast(t('common.error'), 'danger');
  } finally {
    loading.value = false;
  }
}

onMounted(loadQueue);

function resetFilters() {
  searchQuery.value = '';
  statusFilter.value = 'all';
  sortBy.value = 'newest';
  currentPage.value = 1;
}

const filteredRequests = computed(() => {
  let list = [...requests.value];

  // Фильтр по статусу
  if (statusFilter.value === 'pending') {
    list = list.filter((r) => isPending(r.status));
  } else if (statusFilter.value === 'in_review') {
    list = list.filter((r) => isInReview(r.status));
  } else if (statusFilter.value === 'my') {
    list = list.filter(
      (r) =>
        r.moderatorId &&
        currentUserId.value &&
        (r.moderatorId === currentUserId.value || r.moderatorId === authState.user?.email)
    );
  }

  // Поиск
  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase().trim();
    list = list.filter((r) => {
      const titleRu = (r.snapshot.titleRu || '').toLowerCase();
      const titleEn = (r.snapshot.titleEn || '').toLowerCase();
      const pId = String(r.projectId);
      const owner = (r.ownerId || '').toLowerCase();
      return titleRu.includes(q) || titleEn.includes(q) || pId.includes(q) || owner.includes(q);
    });
  }

  // Сортировка
  if (sortBy.value === 'newest') {
    list.sort((a, b) => new Date(b.submittedAt) - new Date(a.submittedAt));
  } else if (sortBy.value === 'oldest') {
    list.sort((a, b) => new Date(a.submittedAt) - new Date(b.submittedAt));
  } else if (sortBy.value === 'title') {
    list.sort((a, b) => {
      const nameA = a.snapshot.titleRu || a.snapshot.titleEn || '';
      const nameB = b.snapshot.titleRu || b.snapshot.titleEn || '';
      return nameA.localeCompare(nameB);
    });
  }

  return list;
});

const totalPages = computed(() =>
  Math.max(1, Math.ceil(filteredRequests.value.length / pageSize.value))
);

const paginatedRequests = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value;
  return filteredRequests.value.slice(start, start + pageSize.value);
});

function isPending(status) {
  return (
    status === REQUEST_STATUS.PENDING ||
    status === 1 ||
    status === 'REQUEST_STATUS_PENDING' ||
    status === 'pending'
  );
}

function isInReview(status) {
  return (
    status === REQUEST_STATUS.IN_REVIEW ||
    status === 2 ||
    status === 'REQUEST_STATUS_IN_REVIEW' ||
    status === 'in_review'
  );
}

function reqStatusLabel(status) {
  if (isPending(status)) return t('moderation.pending');
  if (isInReview(status)) return t('moderation.inReview');
  return t('common.unknown');
}

function reqStatusClass(status) {
  if (isPending(status)) return 'status-pending';
  if (isInReview(status)) return 'status-in-review';
  return 'status-neutral';
}

function openProject(projectId) {
  router.push(`/moderator/projects/${projectId}`);
}

async function claimAndOpen(req) {
  claimingId.value = req.id;
  try {
    await moderationApi.claimRequest(req.id);
    showToast(t('moderation.claimBtn') + ' — успешно', 'success');
    openProject(req.projectId);
  } catch (err) {
    showToast(err.response?.data?.message || t('common.error'), 'danger');
  } finally {
    claimingId.value = null;
  }
}
</script>

<style scoped>
.moderation-page-container {
  width: 100%;
  min-height: calc(100vh - 60px);
  background: var(--bg-app, #0d1117);
  padding: 24px 32px 48px;
  box-sizing: border-box;
}

.main-content-wrap {
  width: 100%;
  max-width: 100%;
  margin: 0;
  display: flex;
  flex-direction: column;
}

/* Панель фильтрации в стиле Яндекс.Игр / Developer Hub */
.filters-toolbar {
  display: flex;
  align-items: flex-end;
  gap: 16px;
  margin-bottom: 24px;
  width: 100%;
}

.filter-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.field-search {
  flex: 1;
  min-width: 220px;
}

.field-status {
  width: 180px;
  flex-shrink: 0;
}

.field-sort {
  width: 210px;
  flex-shrink: 0;
}

.field-label {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-tertiary, #8b949e);
  letter-spacing: 0.1px;
}

.input-wrapper,
.select-wrapper {
  position: relative;
  display: flex;
  align-items: center;
  width: 100%;
}

.filter-input {
  width: 100%;
  height: 36px;
  padding: 0 32px 0 12px;
  background: var(--bg-secondary, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  color: var(--text-main, #f0f6fc);
  font-size: 13px;
  outline: none;
  transition: border-color 0.15s;
  box-sizing: border-box;
}

.filter-input:focus {
  border-color: var(--primary, #58a6ff);
}

.filter-input::placeholder {
  color: var(--text-tertiary, #6e7681);
}

.clear-input-btn {
  position: absolute;
  right: 8px;
  background: transparent;
  border: none;
  color: var(--text-tertiary, #8b949e);
  cursor: pointer;
  padding: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.clear-input-btn:hover {
  color: var(--text-main, #f0f6fc);
}

.filter-select {
  width: 100%;
  height: 36px;
  padding: 0 30px 0 12px;
  background: var(--bg-secondary, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  color: var(--text-main, #f0f6fc);
  font-size: 13px;
  outline: none;
  cursor: pointer;
  appearance: none;
  -webkit-appearance: none;
  transition: border-color 0.15s;
  box-sizing: border-box;
}

.filter-select:focus {
  border-color: var(--primary, #58a6ff);
}

.select-arrow {
  position: absolute;
  right: 10px;
  pointer-events: none;
  color: var(--text-tertiary, #8b949e);
}

.btn-reset-filters {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 36px;
  padding: 0 12px;
  background: transparent;
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  color: var(--text-muted, #b0b8c4);
  font-size: 13px;
  cursor: pointer;
  flex-shrink: 0;
  transition: all 0.15s;
}

.btn-reset-filters:hover {
  background: var(--bg-tertiary, #21262d);
  color: var(--text-main, #f0f6fc);
  border-color: var(--border-secondary, #484f58);
}

.btn-refresh {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 36px;
  padding: 0 14px;
  background: var(--bg-secondary, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  color: var(--text-main, #f0f6fc);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  flex-shrink: 0;
  transition: all 0.15s;
}

.btn-refresh:hover:not(:disabled) {
  background: var(--bg-tertiary, #21262d);
  border-color: var(--border-secondary, #484f58);
}

.btn-refresh:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Табличный вид */
.table-wrapper {
  width: 100%;
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-md, 8px);
  overflow-x: auto;
}

.moderation-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
}

.moderation-table thead {
  background: var(--bg-card, #161b22);
  border-bottom: 1px solid var(--border, #30363d);
}

.moderation-table th {
  padding: 12px 16px;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-tertiary, #8b949e);
  letter-spacing: 0.2px;
}

.col-game {
  width: 32%;
}

.col-version {
  width: 10%;
}

.col-dev {
  width: 16%;
}

.col-date {
  width: 14%;
}

.col-status {
  width: 12%;
}

.col-mod {
  width: 10%;
}

.col-actions {
  width: 6%;
  text-align: right;
  padding-right: 16px;
}

.table-row {
  cursor: pointer;
  border-bottom: 1px solid var(--border, #21262d);
  transition: background-color 0.15s ease;
}

.table-row:hover {
  background: var(--bg-secondary, #161b22);
}

.table-row td {
  padding: 12px 16px;
  vertical-align: middle;
}

.table-row td.col-actions {
  text-align: right;
  padding-right: 16px;
}

.game-cell {
  display: flex;
  align-items: center;
  gap: 14px;
}

.game-icon-box {
  width: 40px;
  height: 40px;
  border-radius: var(--radius-sm, 8px);
  background: var(--bg-tertiary, #21262d);
  border: 1px solid var(--border, #30363d);
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  flex-shrink: 0;
}

.game-icon-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.game-icon-mock {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-tertiary, #21262d);
  color: var(--text-tertiary, #8b949e);
  font-size: 11px;
  font-weight: 600;
}

.game-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.game-type-label {
  font-size: 12px;
  font-weight: 400;
  color: var(--text-tertiary, #8b949e);
}

.game-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
}

.version-badge {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
  background: var(--bg-tertiary, #21262d);
  border: 1px solid var(--border, #30363d);
  color: var(--primary, #58a6ff);
}

.dev-cell {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--text-muted, #b0b8c4);
  max-width: 160px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.date-text {
  font-size: 13px;
  color: var(--text-muted, #b0b8c4);
}

.status-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 24px;
  padding: 0 10px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  white-space: nowrap;
}

.status-pending {
  background: rgba(245, 176, 39, 0.12);
  border: 1px solid rgba(245, 176, 39, 0.35);
  color: #f5b027;
}

.status-in-review {
  background: rgba(88, 166, 255, 0.12);
  border: 1px solid rgba(88, 166, 255, 0.35);
  color: #58a6ff;
}

.status-neutral {
  background: var(--bg-tertiary, #21262d);
  border: 1px solid var(--border, #30363d);
  color: var(--text-muted, #b0b8c4);
}

.mod-name {
  font-size: 13px;
  color: var(--text-main, #f0f6fc);
  font-weight: 500;
}

.unassigned-text {
  font-size: 13px;
  color: var(--text-tertiary, #6e7681);
}

.row-actions {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.btn-claim-sm {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 30px;
  padding: 0 12px;
  background: var(--primary, #58a6ff);
  color: #ffffff;
  border: none;
  border-radius: var(--radius-sm, 6px);
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.15s;
  white-space: nowrap;
}

.btn-claim-sm:hover:not(:disabled) {
  background: var(--primary-hover, #79c0ff);
}

.btn-claim-sm:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-inspect-sm {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 30px;
  padding: 0 12px;
  background: var(--bg-secondary, #21262d);
  border: 1px solid var(--border, #30363d);
  color: var(--text-main, #f0f6fc);
  border-radius: var(--radius-sm, 6px);
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
  white-space: nowrap;
}

.btn-inspect-sm:hover {
  border-color: var(--primary, #58a6ff);
  color: var(--primary, #58a6ff);
}

/* Пустые состояния */
.state-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  gap: 12px;
  color: var(--text-muted, #b0b8c4);
}

.empty-card {
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-md, 8px);
}

.empty-icon-wrap {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: var(--bg-secondary, #21262d);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 8px;
}

.btn-primary-sm {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 34px;
  padding: 0 16px;
  background: var(--primary, #58a6ff);
  color: #ffffff;
  border: none;
  border-radius: var(--radius-sm, 6px);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  margin-top: 12px;
}

/* Пагинация */
.pagination-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 0;
  font-size: 13px;
  color: var(--text-muted, #b0b8c4);
}

.page-nav {
  display: flex;
  align-items: center;
  gap: 6px;
}

.page-nav-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  background: var(--bg-secondary, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  color: var(--text-main, #f0f6fc);
  cursor: pointer;
  transition: all 0.15s;
}

.page-nav-btn:hover:not(:disabled) {
  border-color: var(--primary, #58a6ff);
  color: var(--primary, #58a6ff);
}

.page-nav-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.page-current {
  padding: 0 8px;
  font-weight: 500;
}

.page-size-wrap {
  position: relative;
  display: flex;
  align-items: center;
  margin-left: 8px;
}

.page-size-select {
  height: 32px;
  padding: 0 24px 0 8px;
  background: var(--bg-secondary, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  color: var(--text-main, #f0f6fc);
  font-size: 12px;
  outline: none;
  cursor: pointer;
  appearance: none;
  -webkit-appearance: none;
}

.select-caret {
  position: absolute;
  right: 6px;
  pointer-events: none;
  color: var(--text-tertiary, #8b949e);
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.audit-mode-banner {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 16px;
  background: rgba(59, 130, 246, 0.1);
  border: 1px solid rgba(59, 130, 246, 0.25);
  border-radius: var(--radius-md, 8px);
  color: #60a5fa;
  font-size: 0.84rem;
  font-weight: 500;
}

@media (max-width: 800px) {
  .moderation-page-container {
    padding: 16px;
  }
  .filters-toolbar {
    flex-wrap: wrap;
  }
  .field-search {
    width: 100%;
    min-width: 100%;
  }
}
</style>
