<template>
  <div class="moderation-page-container">
    <div class="main-content-wrap">
      <!-- Панель фильтрации и поиска (компактный стиль консоли) -->
      <div class="filters-toolbar">
        <!-- Поиск по названию, ID, разработчику, причине -->
        <div class="filter-field field-search">
          <label class="field-label">{{ t('common.search') }}</label>
          <div class="input-wrapper">
            <input
              type="text"
              v-model="searchQuery"
              :placeholder="t('moderation.searchArchivePlaceholder')"
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

        <!-- Фильтр по вердикту / результату -->
        <div class="filter-field field-status">
          <label class="field-label">{{ t('moderation.verdict') }}</label>
          <div class="select-wrapper">
            <select v-model="statusFilter" class="filter-select">
              <option value="all">{{ t('moderation.allResolved') }}</option>
              <option value="approved">{{ t('moderation.approved') }}</option>
              <option value="rejected">{{ t('moderation.rejected') }}</option>
              <option value="cancelled">{{ t('moderation.cancelled') }}</option>
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
          @click="resetFilters"
          title="Сбросить фильтры"
        >
          <RotateCcw class="icon-xs" />
          <span>{{ t('common.reset') }}</span>
        </button>

        <!-- Кнопка обновления -->
        <button
          class="btn-refresh"
          :disabled="loading"
          @click="loadArchive"
          title="Обновить"
        >
          <RefreshCw class="icon-xs" :class="{ spin: loading }" />
          <span>{{ t('common.refresh') }}</span>
        </button>
      </div>

      <!-- Состояние загрузки -->
      <div v-if="loading" class="state-container">
        <div class="spinner-md"></div>
        <p>{{ t('common.loading') }}</p>
      </div>

      <!-- Пустой архив -->
      <div
        v-else-if="requests.length === 0 && !searchQuery && statusFilter === 'all'"
        class="state-container empty-card"
      >
        <div class="empty-icon-wrap">
          <Archive class="icon-lg text-muted" />
        </div>
        <h3>{{ t('moderation.noArchive') }}</h3>
        <p>Здесь сохраняется журнал всех проверенных и вынесенных модераторами решений.</p>
        <button class="btn-primary-sm" @click="loadArchive">
          <RefreshCw class="icon-xs" />
          <span>{{ t('common.refresh') }}</span>
        </button>
      </div>

      <!-- Пустой список по результатам поиска/фильтров -->
      <div
        v-else-if="filteredRequests.length === 0"
        class="state-container empty-card"
      >
        <Search class="icon-md text-muted" />
        <h3>{{ t('common.empty') }}</h3>
        <p>{{ t('stats.noData') }}</p>
        <button class="btn-reset-filters" @click="resetFilters">
          {{ t('common.reset') }}
        </button>
      </div>

      <!-- Таблица архива решений -->
      <div v-else class="table-wrapper">
        <table class="moderation-table">
          <thead>
            <tr>
              <th class="col-game">{{ t('moderation.projectColumn') }}</th>
              <th class="col-version">{{ t('common.version') }}</th>
              <th class="col-dev">{{ t('moderation.developerColumn') }}</th>
              <th class="col-verdict">{{ t('moderation.verdict') }}</th>
              <th class="col-mod">{{ t('moderation.moderator') }}</th>
              <th class="col-reason">{{ t('moderation.rejectionReason') }} / Комментарий</th>
              <th class="col-date">{{ t('common.date') }}</th>
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
                    <div class="game-type-label">
                      Заявка #{{ req.id }} • Проект #{{ req.projectId }}
                    </div>
                    <div class="game-title">
                      {{ req.snapshot.titleRu || req.snapshot.titleEn || `Проект #${req.projectId}` }}
                    </div>
                  </div>
                </div>
              </td>

              <!-- 2 колонка: Версия сборки -->
              <td class="col-version">
                <span class="version-badge" v-if="req.snapshot.activeBuildVersion">
                  v{{ req.snapshot.activeBuildVersion }}
                </span>
                <span class="text-muted text-sm" v-else>—</span>
              </td>

              <!-- 3 колонка: Разработчик -->
              <td class="col-dev">
                <div class="dev-cell" :title="req.ownerId">
                  <User class="icon-xs text-muted" />
                  <span class="dev-name">{{ req.ownerId || '—' }}</span>
                </div>
              </td>

              <!-- 4 колонка: Вердикт / Статус -->
              <td class="col-verdict">
                <span
                  class="status-pill"
                  :class="verdictClass(req.status)"
                >
                  {{ verdictLabel(req.status) }}
                </span>
              </td>

              <!-- 5 колонка: Модератор -->
              <td class="col-mod">
                <span class="mod-name" v-if="req.moderatorId">
                  {{ req.moderatorId }}
                </span>
                <span class="unassigned-text" v-else>
                  {{ t('moderation.notAssigned') }}
                </span>
              </td>

              <!-- 6 колонка: Причина / Замечания -->
              <td class="col-reason">
                <div class="reason-cell" :title="req.rejectionReason || 'Без замечаний'">
                  <span v-if="req.rejectionReason" class="reason-text">
                    {{ req.rejectionReason }}
                  </span>
                  <span v-else class="text-muted text-sm">—</span>
                </div>
              </td>

              <!-- 7 колонка: Дата решения -->
              <td class="col-date">
                <span class="date-text">
                  {{ formatDateTime(req.reviewedAt || req.submittedAt) }}
                </span>
              </td>

              <!-- 8 колонка: Действия -->
              <td class="col-actions" @click.stop>
                <div class="row-actions">
                  <button
                    class="btn-inspect-sm"
                    @click="openProject(req.projectId)"
                    :title="t('moderation.viewDetails')"
                  >
                    <Eye class="icon-xs" />
                    <span>{{ t('moderation.viewDetails') }}</span>
                  </button>
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
          <button
            class="page-nav-btn"
            :disabled="currentPage === 1"
            @click="currentPage = 1"
          >
            <ChevronsLeft class="icon-sm" />
          </button>
          <button
            class="page-nav-btn"
            :disabled="currentPage === 1"
            @click="currentPage--"
          >
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
  Archive,
  User,
  Eye,
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
} from 'lucide-vue-next';
import {
  moderationApi,
  normalizeRequest,
  formatDateTime,
  REQUEST_STATUS,
} from '@/entities/moderation';
import { getMediaUrl } from '@/entities/project';
import { showToast } from '@/shared/lib';

const { t } = useI18n();
const router = useRouter();

const requests = ref([]);
const loading = ref(false);

const searchQuery = ref('');
const statusFilter = ref('all');
const sortBy = ref('newest');
const currentPage = ref(1);
const pageSize = ref(10);

async function loadArchive() {
  loading.value = true;
  try {
    const res = await moderationApi.listRequests({ limit: 100, offset: 0 });
    const raw = res.requests || [];
    // Отбираем только завершённые заявки (Approved=3, Rejected=4, Cancelled=5)
    const resolved = raw
      .map(normalizeRequest)
      .filter(
        (r) =>
          r.status === REQUEST_STATUS.APPROVED ||
          r.status === REQUEST_STATUS.REJECTED ||
          r.status === REQUEST_STATUS.CANCELLED ||
          r.status === 3 ||
          r.status === 4 ||
          r.status === 5 ||
          r.status === 'REQUEST_STATUS_APPROVED' ||
          r.status === 'REQUEST_STATUS_REJECTED' ||
          r.status === 'REQUEST_STATUS_CANCELLED' ||
          r.status === 'approved' ||
          r.status === 'rejected' ||
          r.status === 'cancelled'
      );
    requests.value = resolved;
  } catch (err) {
    showToast(t('common.error'), 'danger');
  } finally {
    loading.value = false;
  }
}

onMounted(loadArchive);

function resetFilters() {
  searchQuery.value = '';
  statusFilter.value = 'all';
  sortBy.value = 'newest';
  currentPage.value = 1;
}

const filteredRequests = computed(() => {
  let list = [...requests.value];

  // Фильтр по статусу
  if (statusFilter.value === 'approved') {
    list = list.filter((r) => isApproved(r.status));
  } else if (statusFilter.value === 'rejected') {
    list = list.filter((r) => isRejected(r.status));
  } else if (statusFilter.value === 'cancelled') {
    list = list.filter((r) => isCancelled(r.status));
  }

  // Поиск
  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase().trim();
    list = list.filter((r) => {
      const titleRu = (r.snapshot.titleRu || '').toLowerCase();
      const titleEn = (r.snapshot.titleEn || '').toLowerCase();
      const pId = String(r.projectId);
      const reqId = String(r.id);
      const owner = (r.ownerId || '').toLowerCase();
      const mod = (r.moderatorId || '').toLowerCase();
      const reason = (r.rejectionReason || '').toLowerCase();
      return (
        titleRu.includes(q) ||
        titleEn.includes(q) ||
        pId.includes(q) ||
        reqId.includes(q) ||
        owner.includes(q) ||
        mod.includes(q) ||
        reason.includes(q)
      );
    });
  }

  // Сортировка
  if (sortBy.value === 'newest') {
    list.sort(
      (a, b) =>
        new Date(b.reviewedAt || b.submittedAt) -
        new Date(a.reviewedAt || a.submittedAt)
    );
  } else if (sortBy.value === 'oldest') {
    list.sort(
      (a, b) =>
        new Date(a.reviewedAt || a.submittedAt) -
        new Date(b.reviewedAt || b.submittedAt)
    );
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

function isApproved(status) {
  return (
    status === REQUEST_STATUS.APPROVED ||
    status === 3 ||
    status === 'REQUEST_STATUS_APPROVED' ||
    status === 'approved'
  );
}

function isRejected(status) {
  return (
    status === REQUEST_STATUS.REJECTED ||
    status === 4 ||
    status === 'REQUEST_STATUS_REJECTED' ||
    status === 'rejected'
  );
}

function isCancelled(status) {
  return (
    status === REQUEST_STATUS.CANCELLED ||
    status === 5 ||
    status === 'REQUEST_STATUS_CANCELLED' ||
    status === 'cancelled'
  );
}

function verdictLabel(status) {
  if (isApproved(status)) return t('moderation.approved');
  if (isRejected(status)) return t('moderation.rejected');
  if (isCancelled(status)) return t('moderation.cancelled');
  return t('common.unknown');
}

function verdictClass(status) {
  if (isApproved(status)) return 'status-approved';
  if (isRejected(status)) return 'status-rejected';
  if (isCancelled(status)) return 'status-neutral';
  return 'status-neutral';
}

function openProject(projectId) {
  router.push(`/moderator/projects/${projectId}`);
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

/* Панель фильтрации */
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
  width: 28%;
}

.col-version {
  width: 8%;
}

.col-dev {
  width: 14%;
}

.col-verdict {
  width: 10%;
}

.col-mod {
  width: 10%;
}

.col-reason {
  width: 18%;
}

.col-date {
  width: 12%;
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
  max-width: 140px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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

.status-approved {
  background: rgba(46, 204, 113, 0.12);
  border: 1px solid rgba(46, 204, 113, 0.35);
  color: #2ecc71;
}

.status-rejected {
  background: rgba(248, 81, 73, 0.12);
  border: 1px solid rgba(248, 81, 73, 0.35);
  color: #f85149;
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

.reason-cell {
  max-width: 220px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.reason-text {
  font-size: 13px;
  color: var(--text-muted, #b0b8c4);
}

.date-text {
  font-size: 13px;
  color: var(--text-muted, #b0b8c4);
}

.row-actions {
  display: inline-flex;
  align-items: center;
  gap: 8px;
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
