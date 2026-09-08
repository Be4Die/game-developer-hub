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
              placeholder="Поиск по имени, email или ID..."
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

        <!-- Статус разработчика -->
        <div class="filter-field field-status">
          <label class="field-label">{{ t('common.status') }}</label>
          <div class="select-wrapper">
            <select v-model="statusFilter" class="filter-select">
              <option value="all">—</option>
              <option value="active">Активные</option>
              <option value="suspended">Заблокированные</option>
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
      </div>

      <!-- Состояние загрузки -->
      <div v-if="loading" class="state-container">
        <div class="spinner-md"></div>
        <p>{{ t('common.loading') }}</p>
      </div>

      <!-- Пустой список без разработчиков -->
      <div
        v-else-if="developers.length === 0 && !searchQuery && statusFilter === 'all'"
        class="state-container empty-card"
      >
        <div class="empty-icon-wrap">
          <Users class="icon-lg text-muted" />
        </div>
        <h3>Разработчиков пока нет</h3>
        <p>Зарегистрированные разработчики игр будут отображаться в этом списке.</p>
        <button class="btn-reset" @click="loadUsers">
          <RefreshCw class="icon-xs" />
          <span>{{ t('common.refresh') }}</span>
        </button>
      </div>

      <!-- Пустой список по результатам поиска -->
      <div v-else-if="filteredDevelopers.length === 0" class="state-container empty-card">
        <Search class="icon-md text-muted" />
        <h3>{{ t('common.empty') }}</h3>
        <p>По заданным критериям разработчиков не найдено.</p>
        <button class="btn-reset" @click="resetFilters">
          {{ t('common.reset') }}
        </button>
      </div>

      <!-- Таблица разработчиков -->
      <div v-else class="table-wrapper">
        <table class="yandex-games-table">
          <thead>
            <tr>
              <th class="col-name">Название</th>
              <th class="col-email">Почта</th>
              <th class="col-date">Регистрация</th>
              <th class="col-status">{{ t('common.status') }}</th>
              <th class="col-actions"></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="dev in paginatedDevelopers" :key="dev.id" class="table-row">
              <!-- Колонка 1: Название / Имя -->
              <td class="col-name">
                <span class="user-name-text">{{ dev.display_name || 'Без имени' }}</span>
              </td>

              <!-- Колонка 2: Почта -->
              <td class="col-email">
                <span class="email-text">{{ dev.email }}</span>
              </td>

              <!-- Колонка 3: Дата регистрации -->
              <td class="col-date">
                <span class="date-text">{{ formatProjectDate(dev.created_at) }}</span>
              </td>

              <!-- Колонка 4: Статус -->
              <td class="col-status">
                <span class="status-pill" :class="statusBadgeClass(dev.status)">
                  {{ statusLabel(dev.status) }}
                </span>
              </td>

              <!-- Колонка 4: Действия -->
              <td class="col-actions" @click.stop>
                <div class="row-actions">
                  <!-- Кнопка блокировки (если активен) -->
                  <button
                    v-if="isUserActive(dev.status)"
                    class="btn-action btn-action-danger"
                    title="Заблокировать доступ"
                    :disabled="actionPendingId === dev.id"
                    @click="openStatusModal(dev, 'suspend')"
                  >
                    <Ban class="icon-xs" />
                    <span>Заблокировать</span>
                  </button>

                  <!-- Кнопка разблокировки (если заблокирован) -->
                  <button
                    v-else-if="isUserSuspended(dev.status)"
                    class="btn-action btn-action-success"
                    title="Разблокировать доступ"
                    :disabled="actionPendingId === dev.id"
                    @click="handleSetStatus(dev, 'USER_STATUS_ACTIVE')"
                  >
                    <CheckCircle class="icon-xs" />
                    <span>Разблокировать</span>
                  </button>

                  <!-- Кнопка восстановления (если удален) -->
                  <button
                    v-if="isUserDeleted(dev.status)"
                    class="btn-action btn-action-success"
                    title="Восстановить аккаунт"
                    :disabled="actionPendingId === dev.id"
                    @click="handleSetStatus(dev, 'USER_STATUS_ACTIVE')"
                  >
                    <RotateCcw class="icon-xs" />
                    <span>Восстановить</span>
                  </button>

                  <!-- Кнопка удаления (soft-delete) -->
                  <button
                    v-if="!isUserDeleted(dev.status)"
                    class="btn-icon-danger"
                    title="Удалить аккаунт"
                    :disabled="actionPendingId === dev.id"
                    @click="openDeleteModal(dev)"
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
    <div v-if="filteredDevelopers.length > 0" class="pagination-container">
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

    <!-- Модальное окно подтверждения блокировки -->
    <transition name="modal-fade">
      <div v-if="statusTargetModal" class="modal-overlay" @click.self="statusTargetModal = null">
        <div class="modal-card">
          <div class="modal-header">
            <div class="modal-title-wrap">
              <Ban class="icon-md text-danger" />
              <div>
                <h3>Блокировка разработчика</h3>
                <p class="modal-subtitle">Подтверждение ограничения доступа</p>
              </div>
            </div>
            <button class="modal-close" @click="statusTargetModal = null">
              <X class="icon-sm" />
            </button>
          </div>
          <p class="modal-text">
            Вы уверены, что хотите заблокировать разработчика
            <strong>{{ statusTargetModal.display_name || statusTargetModal.email }}</strong
            >?
          </p>
          <p class="warning-text">
            Пользователь будет немедленно отключён от системы и не сможет авторизоваться.
          </p>
          <div class="modal-actions">
            <button
              class="btn-modal-secondary"
              :disabled="actionPendingId === statusTargetModal.id"
              @click="statusTargetModal = null"
            >
              Отмена
            </button>
            <button
              class="btn-modal-danger"
              :disabled="actionPendingId === statusTargetModal.id"
              @click="handleSetStatus(statusTargetModal, 'USER_STATUS_SUSPENDED')"
            >
              <Loader2 v-if="actionPendingId === statusTargetModal.id" class="icon-xs spin" />
              <Ban v-else class="icon-xs" />
              <span>Заблокировать</span>
            </button>
          </div>
        </div>
      </div>
    </transition>

    <!-- Модальное окно подтверждения удаления -->
    <transition name="modal-fade">
      <div v-if="deleteTargetModal" class="modal-overlay" @click.self="deleteTargetModal = null">
        <div class="modal-card">
          <div class="modal-header">
            <div class="modal-title-wrap">
              <Trash2 class="icon-md text-danger" />
              <div>
                <h3>Удаление пользователя</h3>
                <p class="modal-subtitle">Soft-delete учётной записи</p>
              </div>
            </div>
            <button class="modal-close" @click="deleteTargetModal = null">
              <X class="icon-sm" />
            </button>
          </div>
          <p class="modal-text">
            Вы уверены, что хотите удалить учётную запись
            <strong>{{ deleteTargetModal.display_name || deleteTargetModal.email }}</strong
            >?
          </p>
          <p class="hint-text">
            Аккаунт будет помечен как удалённый. Вы сможете восстановить его позже при
            необходимости.
          </p>
          <div class="modal-actions">
            <button
              class="btn-modal-secondary"
              :disabled="actionPendingId === deleteTargetModal.id"
              @click="deleteTargetModal = null"
            >
              Отмена
            </button>
            <button
              class="btn-modal-danger"
              :disabled="actionPendingId === deleteTargetModal.id"
              @click="handleSoftDelete(deleteTargetModal)"
            >
              <Loader2 v-if="actionPendingId === deleteTargetModal.id" class="icon-xs spin" />
              <Trash2 v-else class="icon-xs" />
              <span>Удалить</span>
            </button>
          </div>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import {
  Users,
  Search,
  X,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
  RotateCcw,
  RefreshCw,
  Ban,
  CheckCircle,
  Trash2,
  Loader2,
} from 'lucide-vue-next';
import { searchUsers, setUserStatus, deleteUser } from '@/entities/user';
import { formatProjectDate, showToast } from '@/shared/lib';

const { t } = useI18n();

const loading = ref(false);
const allUsers = ref([]);
const searchQuery = ref('');
const statusFilter = ref('all');
const sortBy = ref('newest');
const currentPage = ref(1);
const pageSize = ref(10);

const actionPendingId = ref(null);
const statusTargetModal = ref(null);
const deleteTargetModal = ref(null);

onMounted(() => {
  loadUsers();
});

async function loadUsers() {
  loading.value = true;
  try {
    const res = await searchUsers({ query: '', limit: 100 });
    allUsers.value = res.users || [];
  } catch (err) {
    allUsers.value = [];
    showToast('Не удалось загрузить список пользователей', 'danger');
  } finally {
    loading.value = false;
  }
}

// Фильтруем только разработчиков (исключаем модераторов и админа)
const developers = computed(() => {
  return allUsers.value.filter((u) => {
    const r = u.role;
    return r === 'USER_ROLE_DEVELOPER' || r === 'developer' || r === 1 || (!r && !isModOrAdmin(u));
  });
});

function isModOrAdmin(u) {
  const r = u.role;
  return (
    r === 'USER_ROLE_ADMIN' ||
    r === 'admin' ||
    r === 3 ||
    r === 'USER_ROLE_MODERATOR' ||
    r === 'moderator' ||
    r === 2
  );
}

const filteredDevelopers = computed(() => {
  let list = [...developers.value];

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
      if (statusFilter.value === 'suspended') return isUserSuspended(u.status);
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

const totalPages = computed(() => Math.ceil(filteredDevelopers.value.length / pageSize.value) || 1);
const pageStart = computed(() => (currentPage.value - 1) * pageSize.value);

const paginatedDevelopers = computed(() => {
  return filteredDevelopers.value.slice(pageStart.value, pageStart.value + pageSize.value);
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

function isUserSuspended(st) {
  return st === 'USER_STATUS_SUSPENDED' || st === 'suspended' || st === 2;
}

function isUserDeleted(st) {
  return st === 'USER_STATUS_DELETED' || st === 'deleted' || st === 3;
}

function statusBadgeClass(st) {
  if (isUserSuspended(st)) return 'status-suspended';
  if (isUserDeleted(st)) return 'status-deleted';
  return 'status-active';
}

function statusLabel(st) {
  if (isUserSuspended(st)) return 'Заблокирован';
  if (isUserDeleted(st)) return 'Удалён';
  return 'Активен';
}

function openStatusModal(user, action) {
  if (action === 'suspend') {
    statusTargetModal.value = user;
  }
}

function openDeleteModal(user) {
  deleteTargetModal.value = user;
}

async function handleSetStatus(user, newStatus) {
  actionPendingId.value = user.id;
  try {
    await setUserStatus(user.id, newStatus);
    const actionWord = isUserActive(newStatus) ? 'активирован' : 'заблокирован';
    showToast(`Пользователь "${user.display_name || user.email}" ${actionWord}`, 'success');
    statusTargetModal.value = null;
    await loadUsers();
  } catch (err) {
    showToast(err.response?.data?.message || 'Ошибка изменения статуса', 'danger');
  } finally {
    actionPendingId.value = null;
  }
}

async function handleSoftDelete(user) {
  actionPendingId.value = user.id;
  try {
    await deleteUser(user.id);
    showToast(`Пользователь "${user.display_name || user.email}" удалён`, 'success');
    deleteTargetModal.value = null;
    await loadUsers();
  } catch (err) {
    showToast(err.response?.data?.message || 'Не удалось удалить пользователя', 'danger');
  } finally {
    actionPendingId.value = null;
  }
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
  width: 32%;
  padding-left: 12px;
}

.yandex-games-table th.col-email {
  width: 30%;
}

.yandex-games-table th.col-date {
  width: 15%;
}

.yandex-games-table th.col-status {
  width: 11%;
}

.yandex-games-table th.col-actions {
  width: 12%;
  text-align: right;
  padding-right: 12px;
}

.table-row {
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

.status-suspended {
  background: rgba(248, 81, 73, 0.12);
  border-color: rgba(248, 81, 73, 0.35);
  color: #f85149;
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

.btn-action-danger:hover:not(:disabled) {
  color: #f85149;
  border-color: rgba(248, 81, 73, 0.5);
  background: rgba(248, 81, 73, 0.1);
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

/* Modals */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.7);
  z-index: 300;
  display: flex;
  align-items: center;
  justify-content: center;
  backdrop-filter: blur(4px);
  padding: 16px;
}

.modal-card {
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-md, 8px);
  padding: 24px;
  width: 100%;
  max-width: 440px;
  box-shadow: 0 16px 32px rgba(0, 0, 0, 0.4);
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.modal-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
}

.modal-title-wrap {
  display: flex;
  align-items: center;
  gap: 10px;
}

.modal-title-wrap h3 {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
}

.modal-subtitle {
  margin: 2px 0 0 0;
  font-size: 12px;
  color: var(--text-tertiary, #8b949e);
}

.modal-close {
  background: transparent;
  border: none;
  color: var(--text-tertiary, #8b949e);
  cursor: pointer;
  padding: 4px;
}

.modal-close:hover {
  color: var(--text-main, #f0f6fc);
}

.modal-text {
  margin: 0;
  font-size: 13px;
  color: var(--text-main, #f0f6fc);
  line-height: 1.5;
}

.warning-text {
  margin: 0;
  font-size: 12px;
  color: #f85149;
  line-height: 1.4;
}

.hint-text {
  margin: 0;
  font-size: 12px;
  color: var(--text-tertiary, #8b949e);
  line-height: 1.4;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding-top: 12px;
  border-top: 1px solid var(--border, #30363d);
}

.btn-modal-danger {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  background: #da3633;
  color: #ffffff;
  border: 1px solid rgba(248, 81, 73, 0.4);
  border-radius: var(--radius-sm, 6px);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.15s;
}

.btn-modal-danger:hover:not(:disabled) {
  background: #f85149;
}

.btn-modal-secondary {
  display: inline-flex;
  align-items: center;
  padding: 6px 14px;
  background: var(--bg-secondary, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  color: var(--text-main, #f0f6fc);
  font-size: 13px;
  cursor: pointer;
}

.btn-modal-secondary:hover:not(:disabled) {
  background: var(--bg-tertiary, #21262d);
}

.modal-fade-enter-active,
.modal-fade-leave-active {
  transition:
    opacity 0.15s ease,
    transform 0.15s ease;
}
.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
  transform: scale(0.96);
}
</style>
