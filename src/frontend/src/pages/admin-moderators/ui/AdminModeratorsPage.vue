<template>
  <div class="projects-page-container">
    <div class="main-content-wrap">
      <!-- Панель фильтров и поиска в стиле консоли -->
      <div class="filters-toolbar">
        <!-- Поиск по названию / email / ID -->
        <div class="filter-field field-name">
          <label class="field-label">{{ t('common.name') }}</label>
          <div class="input-wrapper">
            <input
              v-model="searchQuery"
              type="text"
              placeholder="Поиск по имени, логину или ID..."
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

        <!-- Статус модератора -->
        <div class="filter-field field-status">
          <label class="field-label">{{ t('common.status') }}</label>
          <div class="select-wrapper">
            <select v-model="statusFilter" class="filter-select">
              <option value="all">—</option>
              <option value="active">Активные</option>
              <option value="deleted">Удалённые</option>
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
              <option value="name">{{ t('common.name') }} (A–Z)</option>
            </select>
            <ChevronDown class="icon-xs select-arrow" />
          </div>
        </div>

        <!-- Сброс фильтров -->
        <button
          v-if="searchQuery || statusFilter !== 'all' || sortBy !== 'newest'"
          class="btn-reset-filters"
          title="Сбросить фильтры"
          @click="resetFilters"
        >
          <RotateCcw class="icon-xs" />
          <span>{{ t('common.reset') }}</span>
        </button>

        <!-- Кнопка добавления модератора -->
        <button class="btn-action-primary" @click="showCreateModal = true">
          <Plus class="icon-sm" />
          <span>Добавить модератора</span>
        </button>
      </div>

      <!-- Состояние загрузки -->
      <div v-if="loading" class="state-container">
        <div class="spinner-md"></div>
        <p>{{ t('common.loading') }}</p>
      </div>

      <!-- Пустой список без модераторов -->
      <div
        v-else-if="moderators.length === 0 && !searchQuery && statusFilter === 'all'"
        class="state-container empty-card"
      >
        <div class="empty-icon-wrap">
          <ShieldCheck class="icon-lg text-warning" />
        </div>
        <h3>Модераторы ещё не созданы</h3>
        <p>Добавьте первых сотрудников модерации для проверки проектов и ведения диалогов.</p>
        <button class="btn-action-primary" @click="showCreateModal = true">
          <Plus class="icon-sm" />
          <span>Добавить модератора</span>
        </button>
      </div>

      <!-- Пустой список по результатам поиска -->
      <div v-else-if="filteredModerators.length === 0" class="state-container empty-card">
        <Search class="icon-md text-muted" />
        <h3>{{ t('common.empty') }}</h3>
        <p>По заданным критериям модераторов не найдено.</p>
        <button class="btn-reset" @click="resetFilters">
          {{ t('common.reset') }}
        </button>
      </div>

      <!-- Таблица модераторов -->
      <div v-else class="table-wrapper">
        <table class="yandex-games-table">
          <thead>
            <tr>
              <th class="col-name">Название</th>
              <th class="col-email">Почта</th>
              <th class="col-resolved">{{ t('moderation.resolvedColumn') }}</th>
              <th class="col-in-review">{{ t('moderation.inReviewColumn') }}</th>
              <th class="col-avg-time">{{ t('moderation.avgDurationColumn') }}</th>
              <th class="col-date">Создан</th>
              <th class="col-status">{{ t('common.status') }}</th>
              <th class="col-actions"></th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="mod in paginatedModerators"
              :key="mod.id"
              class="table-row table-row-clickable"
              @click="goToModerator(mod.id)"
            >
              <!-- Колонка 1: Название / Имя -->
              <td class="col-name">
                <span class="user-name-text">{{ mod.display_name || 'Модератор' }}</span>
              </td>

              <!-- Колонка 2: Почта -->
              <td class="col-email">
                <span class="email-text">{{ mod.email }}</span>
              </td>

              <!-- Колонка 3: Рассмотрено заявок -->
              <td class="col-resolved">
                <span class="stat-number">{{ statsMap[mod.id]?.total_resolved || 0 }}</span>
                <span v-if="statsMap[mod.id]?.total_resolved > 0" class="stat-subtext">
                  ({{ statsMap[mod.id]?.approved_count || 0 }} од.)
                </span>
              </td>

              <!-- Колонка 4: В работе -->
              <td class="col-in-review">
                <span v-if="statsMap[mod.id]?.in_review_count > 0" class="in-review-badge">
                  {{ statsMap[mod.id]?.in_review_count }}
                </span>
                <span v-else class="text-muted">—</span>
              </td>

              <!-- Колонка 5: Среднее время проверки -->
              <td class="col-avg-time">
                <span class="stat-time">
                  {{
                    statsMap[mod.id]?.avg_review_duration_seconds
                      ? formatDurationSeconds(statsMap[mod.id]?.avg_review_duration_seconds)
                      : '—'
                  }}
                </span>
              </td>

              <!-- Колонка 6: Дата создания -->
              <td class="col-date">
                <span class="date-text">{{ formatProjectDate(mod.created_at) }}</span>
              </td>

              <!-- Колонка 7: Статус -->
              <td class="col-status">
                <span class="status-pill" :class="statusBadgeClass(mod.status)">
                  {{ statusLabel(mod.status) }}
                </span>
              </td>

              <!-- Колонка 8: Действия -->
              <td class="col-actions" @click.stop>
                <div class="row-actions">
                  <!-- Кнопка восстановления (если удален) -->
                  <button
                    v-if="isUserDeleted(mod.status)"
                    class="btn-action btn-action-success"
                    title="Восстановить учётную запись"
                    :disabled="actionPendingId === mod.id"
                    @click="handleRestore(mod)"
                  >
                    <RotateCcw class="icon-xs" />
                    <span>Восстановить</span>
                  </button>

                  <!-- Кнопка удаления (soft-delete) -->
                  <button
                    v-else
                    class="btn-icon-danger"
                    title="Удалить модератора"
                    :disabled="actionPendingId === mod.id"
                    @click="deleteTarget = mod"
                  >
                    <Trash2 class="icon-xs" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Пагинация в стиле Яндекс.Консоли -->
    <div v-if="filteredModerators.length > 0" class="pagination-container">
      <div class="pagination-center">
        <button
          class="page-nav-btn"
          :disabled="currentPage === 1"
          title="Первая страница"
          @click="currentPage = 1"
        >
          <ChevronsLeft class="icon-sm" />
        </button>
        <button
          class="page-nav-btn"
          :disabled="currentPage === 1"
          title="Предыдущая"
          @click="currentPage--"
        >
          <ChevronLeft class="icon-sm" />
        </button>

        <div class="page-numbers">
          <button
            v-for="page in visiblePages"
            :key="page"
            class="page-num-btn"
            :class="{
              active: currentPage === page,
              ellipsis: page === '...',
            }"
            :disabled="page === '...'"
            @click="typeof page === 'number' && (currentPage = page)"
          >
            {{ page }}
          </button>
        </div>

        <button
          class="page-nav-btn"
          :disabled="currentPage === totalPages"
          title="Следующая"
          @click="currentPage++"
        >
          <ChevronRight class="icon-sm" />
        </button>
        <button
          class="page-nav-btn"
          :disabled="currentPage === totalPages"
          title="Последняя страница"
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

    <!-- Модалка создания модератора -->
    <CreateModeratorModal
      v-if="showCreateModal"
      @created="handleModeratorCreated"
      @cancel="showCreateModal = false"
    />

    <!-- Модалка подтверждения удаления -->
    <DeleteModeratorModal
      v-if="deleteTarget"
      :target="deleteTarget"
      @deleted="handleModeratorDeleted"
      @cancel="deleteTarget = null"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import {
  ShieldCheck,
  Plus,
  Search,
  X,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
  RotateCcw,
  Trash2,
} from 'lucide-vue-next';
import { searchUsers, setUserStatus } from '@/entities/user';
import { CreateModeratorModal, DeleteModeratorModal } from '@/features/manage-moderators';
import { moderationApi, formatDurationSeconds } from '@/entities/moderation';
import { formatProjectDate, showToast } from '@/shared/lib';

const { t } = useI18n();
const router = useRouter();

const loading = ref(false);
const allUsers = ref([]);
const statsMap = ref({});
const searchQuery = ref('');
const statusFilter = ref('all');
const sortBy = ref('newest');
const currentPage = ref(1);
const pageSize = ref(10);

const showCreateModal = ref(false);
const deleteTarget = ref(null);
const actionPendingId = ref(null);

onMounted(() => {
  loadUsers();
});

async function loadUsers() {
  loading.value = true;
  try {
    const [res] = await Promise.all([
      searchUsers({ query: '', limit: 100 }),
      loadModeratorsStats(),
    ]);
    allUsers.value = res.users || [];
  } catch (err) {
    allUsers.value = [];
    showToast('Не удалось загрузить список модераторов', 'danger');
  } finally {
    loading.value = false;
  }
}

async function loadModeratorsStats() {
  try {
    const res = await moderationApi.listModeratorsStats();
    const map = {};
    for (const s of res.stats || []) {
      map[s.moderator_id] = s;
    }
    statsMap.value = map;
  } catch (err) {
    console.warn('Failed to load moderators stats', err);
  }
}

// Фильтруем только модераторов
const moderators = computed(() => {
  return allUsers.value.filter((u) => {
    const r = u.role;
    return r === 'USER_ROLE_MODERATOR' || r === 'moderator' || r === 2;
  });
});

const filteredModerators = computed(() => {
  let list = [...moderators.value];

  // Поиск
  if (searchQuery.value.trim()) {
    const q = searchQuery.value.trim().toLowerCase();
    list = list.filter(
      (u) =>
        (u.display_name && u.display_name.toLowerCase().includes(q)) ||
        (u.email && u.email.toLowerCase().includes(q)) ||
        (u.id && u.id.toLowerCase().includes(q))
    );
  }

  // Фильтр по статусу
  if (statusFilter.value !== 'all') {
    list = list.filter((u) => {
      if (statusFilter.value === 'active') return isUserActive(u.status);
      if (statusFilter.value === 'deleted') return isUserDeleted(u.status);
      return true;
    });
  }

  // Сортировка
  if (sortBy.value === 'name') {
    list.sort((a, b) =>
      (a.display_name || a.email || '').localeCompare(b.display_name || b.email || '')
    );
  } else if (sortBy.value === 'newest') {
    list.sort((a, b) => new Date(b.created_at || 0) - new Date(a.created_at || 0));
  } else if (sortBy.value === 'oldest') {
    list.sort((a, b) => new Date(a.created_at || 0) - new Date(b.created_at || 0));
  }

  return list;
});

const totalPages = computed(() => Math.ceil(filteredModerators.value.length / pageSize.value) || 1);
const pageStart = computed(() => (currentPage.value - 1) * pageSize.value);

const paginatedModerators = computed(() => {
  return filteredModerators.value.slice(pageStart.value, pageStart.value + pageSize.value);
});

const visiblePages = computed(() => {
  const total = totalPages.value;
  const current = currentPage.value;
  if (total <= 7) {
    return Array.from({ length: total }, (_, i) => i + 1);
  }
  if (current <= 4) {
    return [1, 2, 3, 4, 5, '...', total];
  }
  if (current >= total - 3) {
    return [1, '...', total - 4, total - 3, total - 2, total - 1, total];
  }
  return [1, '...', current - 1, current, current + 1, '...', total];
});

function resetFilters() {
  searchQuery.value = '';
  statusFilter.value = 'all';
  sortBy.value = 'newest';
  currentPage.value = 1;
}

function isUserActive(st) {
  return st === 'USER_STATUS_ACTIVE' || st === 'active' || st === 1 || !st;
}

function isUserDeleted(st) {
  return st === 'USER_STATUS_DELETED' || st === 'deleted' || st === 3;
}

function statusBadgeClass(st) {
  if (isUserDeleted(st)) return 'status-deleted';
  return 'status-active';
}

function statusLabel(st) {
  if (isUserDeleted(st)) return 'Удалён';
  return 'Активен';
}

async function handleRestore(mod) {
  actionPendingId.value = mod.id;
  try {
    await setUserStatus(mod.id, 'USER_STATUS_ACTIVE');
    showToast(`Модератор "${mod.display_name || mod.email}" восстановлен`, 'success');
    await loadUsers();
  } catch (err) {
    showToast(err.response?.data?.message || 'Не удалось восстановить модератора', 'danger');
  } finally {
    actionPendingId.value = null;
  }
}

function goToModerator(modId) {
  if (modId) {
    router.push(`/admin/moderators/${modId}`);
  }
}

function handleModeratorCreated() {
  showCreateModal.value = false;
  loadUsers();
}

function handleModeratorDeleted() {
  deleteTarget.value = null;
  loadUsers();
}
</script>

<style scoped>
.projects-page-container {
  width: 100%;
  min-height: calc(100vh - 60px);
  background: var(--bg-app);
  padding: 24px 32px;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
}

.main-content-wrap {
  width: 100%;
  flex: 1;
  display: flex;
  flex-direction: column;
}

/* Панель фильтров */
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

.field-name {
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
  font-size: 13px;
  font-weight: 400;
  color: var(--text-tertiary, #8b949e);
  line-height: 1;
}

.input-wrapper,
.select-wrapper {
  position: relative;
  display: flex;
  align-items: center;
  height: 38px;
  border-radius: var(--radius-sm, 6px);
  border: 1px solid var(--border, #30363d);
  background: var(--bg-secondary, #161b22);
  box-sizing: border-box;
  transition:
    border-color 0.15s,
    box-shadow 0.15s;
}

.input-wrapper:focus-within,
.select-wrapper:focus-within {
  border-color: var(--primary, #58a6ff);
}

.filter-input {
  width: 100%;
  height: 100%;
  padding: 0 32px 0 12px;
  background: transparent;
  border: none;
  color: var(--text-main, #f0f6fc);
  font-size: 14px;
  outline: none;
}

.filter-select {
  width: 100%;
  height: 100%;
  padding: 0 32px 0 12px;
  background: transparent;
  border: none;
  color: var(--text-main, #f0f6fc);
  font-size: 14px;
  outline: none;
  cursor: pointer;
  appearance: none;
  -webkit-appearance: none;
}

.clear-input-btn {
  position: absolute;
  right: 8px;
  background: none;
  border: none;
  color: var(--text-tertiary, #8b949e);
  cursor: pointer;
  padding: 2px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
}

.clear-input-btn:hover {
  color: var(--text-main, #f0f6fc);
}

.select-arrow {
  position: absolute;
  right: 12px;
  pointer-events: none;
  color: var(--text-tertiary, #8b949e);
  width: 14px;
  height: 14px;
}

.btn-reset-filters {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 38px;
  padding: 0 16px;
  border-radius: var(--radius-sm, 6px);
  background: var(--bg-secondary, #161b22);
  border: 1px solid var(--border, #30363d);
  color: var(--text-main, #f0f6fc);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition:
    background-color 0.15s,
    border-color 0.15s;
  white-space: nowrap;
}

.btn-reset-filters:hover {
  background: var(--bg-tertiary, #21262d);
  border-color: var(--border-secondary, #484f58);
}

.btn-action-primary {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  height: 38px;
  padding: 0 20px;
  border-radius: var(--radius-sm, 6px);
  background: var(--primary, #58a6ff);
  color: #ffffff;
  border: none;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.15s;
  white-space: nowrap;
}

.btn-action-primary:hover:not(:disabled) {
  background: var(--primary-hover, #79c0ff);
}

.btn-reset {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  background: var(--bg-secondary, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  color: var(--text-main, #f0f6fc);
  font-size: 13px;
  cursor: pointer;
}

.btn-reset:hover {
  background: var(--bg-tertiary, #21262d);
}

/* Таблица */
.table-wrapper {
  width: 100%;
  overflow-x: auto;
}

.yandex-games-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
}

.yandex-games-table th {
  color: var(--text-tertiary, #8b949e);
  font-size: 13px;
  font-weight: 500;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border, #30363d);
  background: transparent;
  white-space: nowrap;
}

.yandex-games-table th.col-name {
  width: 20%;
  padding-left: 12px;
}

.yandex-games-table th.col-email {
  width: 18%;
}

.yandex-games-table th.col-resolved {
  width: 14%;
}

.yandex-games-table th.col-in-review {
  width: 10%;
}

.yandex-games-table th.col-avg-time {
  width: 11%;
}

.yandex-games-table th.col-date {
  width: 11%;
}

.yandex-games-table th.col-status {
  width: 8%;
}

.yandex-games-table th.col-actions {
  width: 8%;
  text-align: right;
  padding-right: 12px;
}

.table-row {
  border-bottom: 1px solid var(--border, #21262d);
  transition: background-color 0.15s ease;
}

.table-row-clickable {
  cursor: pointer;
}

.stat-number {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
}

.stat-subtext {
  font-size: 11px;
  color: var(--text-tertiary, #8b949e);
  margin-left: 4px;
}

.in-review-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 20px;
  padding: 0 6px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
  background: rgba(56, 189, 248, 0.12);
  border: 1px solid rgba(56, 189, 248, 0.35);
  color: #38bdf8;
}

.stat-time {
  font-size: 13px;
  color: var(--text-muted, #b0b8c4);
  font-weight: 500;
}

.btn-action-stats:hover:not(:disabled) {
  color: var(--primary, #58a6ff);
  border-color: rgba(88, 166, 255, 0.5);
  background: rgba(88, 166, 255, 0.1);
}

.table-row:hover {
  background: var(--bg-secondary, #161b22);
}

.table-row td {
  padding: 12px 16px;
  vertical-align: middle;
}

.table-row td.col-name {
  padding-left: 12px;
}

.table-row td.col-actions {
  padding-right: 12px;
  text-align: right;
}

.user-name-text {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
}

.email-text {
  font-size: 13px;
  font-weight: 400;
  color: var(--text-muted, #b0b8c4);
}

.date-text {
  font-size: 13px;
  font-weight: 400;
  color: var(--text-muted, #b0b8c4);
}

/* Статус пилюли */
.status-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 26px;
  padding: 0 12px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  background: var(--bg-tertiary, #21262d);
  border: 1px solid var(--border, #30363d);
  color: var(--text-muted, #b0b8c4);
  white-space: nowrap;
}

.status-active {
  background: rgba(46, 204, 113, 0.12);
  border-color: rgba(46, 204, 113, 0.35);
  color: #2ecc71;
}

.status-deleted {
  background: var(--bg-tertiary, #21262d);
  border-color: var(--border, #30363d);
  color: var(--text-tertiary, #8b949e);
}

/* Действия */
.row-actions {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  justify-content: flex-end;
}

.btn-action {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 28px;
  padding: 0 10px;
  border-radius: 4px;
  background: var(--bg-tertiary, #21262d);
  border: 1px solid var(--border, #30363d);
  color: var(--text-main, #f0f6fc);
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
  white-space: nowrap;
}

.btn-action:hover:not(:disabled) {
  background: var(--bg-secondary, #161b22);
  border-color: var(--border-secondary, #484f58);
}

.btn-action-success:hover:not(:disabled) {
  color: #2ecc71;
  border-color: rgba(46, 204, 113, 0.5);
  background: rgba(46, 204, 113, 0.1);
}

.btn-icon-danger {
  background: none;
  border: none;
  padding: 6px;
  cursor: pointer;
  color: var(--text-tertiary, #8b949e);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  transition: all 0.15s;
}

.btn-icon-danger:hover:not(:disabled) {
  color: var(--danger, #f85149);
  background: var(--bg-tertiary, #21262d);
}

/* Пагинация */
.pagination-container {
  display: flex;
  justify-content: center;
  align-items: center;
  margin-top: auto;
  padding-top: 32px;
  padding-bottom: 8px;
  width: 100%;
}

.pagination-center {
  display: flex;
  align-items: center;
  gap: 6px;
}

.page-numbers {
  display: flex;
  align-items: center;
  gap: 4px;
}

.page-nav-btn,
.page-num-btn {
  height: 32px;
  min-width: 32px;
  padding: 0 8px;
  background: var(--bg-secondary, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: 4px;
  color: var(--text-muted, #8b949e);
  font-size: 13px;
  font-weight: 500;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.15s;
}

.page-nav-btn:hover:not(:disabled),
.page-num-btn:hover:not(:disabled) {
  background: var(--bg-tertiary, #21262d);
  border-color: var(--border-secondary, #484f58);
  color: var(--text-main, #f0f6fc);
}

.page-num-btn.active {
  background: var(--primary, #58a6ff);
  border-color: var(--primary, #58a6ff);
  color: #ffffff;
  font-weight: 600;
}

.page-nav-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.page-num-btn.ellipsis {
  border: none;
  background: none;
  cursor: default;
  color: var(--text-tertiary, #6e7681);
}

.page-size-wrap {
  position: relative;
  display: inline-flex;
  align-items: center;
  margin-left: 8px;
}

.page-size-select {
  height: 32px;
  min-width: 62px;
  padding: 0 26px 0 12px;
  background: var(--bg-secondary, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: 4px;
  color: var(--text-muted, #8b949e);
  font-size: 13px;
  outline: none;
  cursor: pointer;
  appearance: none;
  -webkit-appearance: none;
  transition: border-color 0.15s;
}

.page-size-select:focus {
  border-color: var(--primary, #58a6ff);
}

.select-caret {
  position: absolute;
  right: 8px;
  width: 12px;
  height: 12px;
  pointer-events: none;
  color: var(--text-tertiary, #8b949e);
}

/* Empty & Loading states */
.state-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px 24px;
  gap: 12px;
  text-align: center;
}

.empty-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-md, 8px);
}

.empty-icon-wrap {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: var(--bg-tertiary, #21262d);
  display: flex;
  align-items: center;
  justify-content: center;
}

.spinner-md {
  width: 32px;
  height: 32px;
  border: 3px solid var(--border, #30363d);
  border-top-color: var(--primary, #58a6ff);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin: 0 auto;
  display: block;
  flex-shrink: 0;
}

.spin {
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
