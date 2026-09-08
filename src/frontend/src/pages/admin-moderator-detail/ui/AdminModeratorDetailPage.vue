<template>
  <div class="projects-page-container">
    <div class="main-content-wrap">
      <!-- Верхний заголовок и метаданные модератора -->
      <div class="moderator-header-card">
        <div class="header-main-row">
          <div class="identity-info">
            <div class="title-with-badges">
              <h1 class="moderator-name">
                {{ moderator.display_name || moderator.email || 'Модератор' }}
              </h1>
              <span class="role-badge">Модератор</span>
              <span class="status-pill" :class="statusBadgeClass(moderator.status)">
                {{ statusLabel(moderator.status) }}
              </span>
            </div>

            <div class="meta-row">
              <span v-if="moderator.email" class="meta-item">
                <span class="meta-label">Email:</span>
                <span class="meta-value">{{ moderator.email }}</span>
              </span>
              <span class="meta-dot">•</span>
              <span v-if="moderator.id" class="meta-item">
                <span class="meta-label">ID:</span>
                <span class="meta-value font-mono">{{ moderator.id }}</span>
              </span>
              <span class="meta-dot">•</span>
              <span v-if="moderator.created_at" class="meta-item">
                <span class="meta-label">Создан:</span>
                <span class="meta-value">{{ formatProjectDate(moderator.created_at) }}</span>
              </span>
            </div>
          </div>

          <div class="header-actions">
            <button
              class="btn-refresh"
              title="Обновить данные"
              :disabled="loadingStats || loadingActivity"
              @click="refreshAll"
            >
              <RefreshCw
                class="icon-xs"
                :class="{ 'spin-animation': loadingStats || loadingActivity }"
              />
              <span>Обновить</span>
            </button>
          </div>
        </div>
      </div>

      <!-- Секция: Ключевые показатели (без пестрых цветных плашек, строгий стиль) -->
      <div class="section-block">
        <div class="section-header">
          <h2 class="section-title">{{ t('moderation.keyIndicators') }}</h2>
        </div>

        <!-- Сетка показателей -->
        <div class="indicators-grid">
          <!-- 1. Рассмотрено всего -->
          <div class="indicator-card">
            <div class="indicator-label">{{ t('moderation.totalResolved') }}</div>
            <div class="indicator-value-row">
              <span class="indicator-value">{{ stats.total_resolved || 0 }}</span>
            </div>
            <div class="indicator-subtext">
              <span class="subtext-muted">
                {{ stats.approved_count || 0 }} одобрено · {{ stats.rejected_count || 0 }} отклонено
              </span>
            </div>
          </div>

          <!-- 2. Одобрено заявок -->
          <div class="indicator-card">
            <div class="indicator-label">{{ t('moderation.approvedStats') }}</div>
            <div class="indicator-value-row">
              <span class="indicator-value text-success">{{ stats.approved_count || 0 }}</span>
              <span v-if="stats.total_resolved > 0" class="indicator-rate-pill pill-success">
                {{ stats.approval_rate }}%
              </span>
            </div>
            <div class="indicator-subtext">
              <span class="subtext-muted">от рассмотренных</span>
            </div>
          </div>

          <!-- 3. Отклонено заявок -->
          <div class="indicator-card">
            <div class="indicator-label">{{ t('moderation.rejectedStats') }}</div>
            <div class="indicator-value-row">
              <span class="indicator-value text-danger">{{ stats.rejected_count || 0 }}</span>
              <span v-if="stats.total_resolved > 0" class="indicator-rate-pill pill-danger">
                {{ stats.rejection_rate }}%
              </span>
            </div>
            <div class="indicator-subtext">
              <span class="subtext-muted">от рассмотренных</span>
            </div>
          </div>

          <!-- 4. В процессе проверки -->
          <div class="indicator-card">
            <div class="indicator-label">{{ t('moderation.inReviewStats') }}</div>
            <div class="indicator-value-row">
              <span class="indicator-value text-info">{{ stats.in_review_count || 0 }}</span>
            </div>
            <div class="indicator-subtext">
              <span class="subtext-muted">текущие в работе</span>
            </div>
          </div>

          <!-- 5. Среднее время проверки -->
          <div class="indicator-card">
            <div class="indicator-label">{{ t('moderation.avgReviewDuration') }}</div>
            <div class="indicator-value-row">
              <span class="indicator-value">
                {{
                  stats.avg_review_duration_seconds
                    ? formatDurationSeconds(stats.avg_review_duration_seconds)
                    : '—'
                }}
              </span>
            </div>
            <div class="indicator-subtext">
              <span class="subtext-muted">время рассмотрения</span>
            </div>
          </div>

          <!-- 6. Отправлено сообщений -->
          <div class="indicator-card">
            <div class="indicator-label">{{ t('moderation.messagesSent') }}</div>
            <div class="indicator-value-row">
              <span class="indicator-value">{{ stats.messages_sent || 0 }}</span>
            </div>
            <div class="indicator-subtext">
              <span class="subtext-muted">в чатах проектов</span>
            </div>
          </div>
        </div>

        <!-- Разрез по периодам (лаконичный горизонтальный стрип) -->
        <div class="period-strip">
          <span class="period-title">Рассмотрено решений:</span>
          <div class="period-item">
            <span class="period-label">{{ t('moderation.periodToday') }}:</span>
            <span class="period-count">{{ stats.today_resolved || 0 }}</span>
          </div>
          <div class="period-divider"></div>
          <div class="period-item">
            <span class="period-label">{{ t('moderation.periodWeek') }}:</span>
            <span class="period-count">{{ stats.week_resolved || 0 }}</span>
          </div>
          <div class="period-divider"></div>
          <div class="period-item">
            <span class="period-label">{{ t('moderation.periodMonth') }}:</span>
            <span class="period-count">{{ stats.month_resolved || 0 }}</span>
          </div>
          <div class="period-divider"></div>
          <div class="period-item">
            <span class="period-label">{{ t('moderation.periodAllTime') }}:</span>
            <span class="period-count font-bold">{{ stats.total_resolved || 0 }}</span>
          </div>
        </div>
      </div>

      <!-- Секция: Журнал действий модератора -->
      <div class="section-block">
        <div class="section-header">
          <div class="section-title-wrap">
            <h2 class="section-title">{{ t('moderation.activityJournal') }}</h2>
            <span class="count-badge">{{ totalActivity }}</span>
          </div>
        </div>

        <!-- Панель фильтров журнала -->
        <div class="filters-toolbar">
          <!-- Поиск по названию игры, ID или содержанию -->
          <div class="filter-field field-search">
            <label class="field-label">Поиск</label>
            <div class="input-wrapper">
              <input
                v-model="searchQuery"
                type="text"
                placeholder="Поиск по названию игры, ID или тексту сообщения..."
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

          <!-- Фильтр по типу действия -->
          <div class="filter-field field-action-type">
            <label class="field-label">{{ t('moderation.actionColumn') }}</label>
            <div class="select-wrapper">
              <select v-model="actionTypeFilter" class="filter-select" @change="handleFilterChange">
                <option value="">{{ t('moderation.allActions') }}</option>
                <option value="chat_message">{{ t('moderation.actionChatMessage') }}</option>
                <option value="dialog_closed">{{ t('moderation.actionDialogClosed') }}</option>
                <option value="claimed">{{ t('moderation.actionClaimed') }}</option>
                <option value="approved">{{ t('moderation.actionApproved') }}</option>
                <option value="rejected">{{ t('moderation.actionRejected') }}</option>
              </select>
              <ChevronDown class="icon-xs select-arrow" />
            </div>
          </div>

          <!-- Сброс фильтров -->
          <button
            v-if="searchQuery || actionTypeFilter !== ''"
            class="btn-reset-filters"
            title="Сбросить фильтры"
            @click="resetJournalFilters"
          >
            <RotateCcw class="icon-xs" />
            <span>{{ t('common.reset') }}</span>
          </button>
        </div>

        <!-- Состояние загрузки журнала -->
        <div v-if="loadingActivity" class="state-container">
          <div class="spinner-md"></div>
          <p>{{ t('common.loading') }}</p>
        </div>

        <!-- Пустой журнал -->
        <div v-else-if="filteredActivity.length === 0" class="state-container empty-card">
          <Search class="icon-md text-muted" />
          <h3>{{ t('common.empty') }}</h3>
          <p>
            {{
              searchQuery || actionTypeFilter
                ? 'По заданным критериям действий не найдено.'
                : t('moderation.noActivity')
            }}
          </p>
          <button
            v-if="searchQuery || actionTypeFilter"
            class="btn-reset"
            @click="resetJournalFilters"
          >
            {{ t('common.reset') }}
          </button>
        </div>

        <!-- Таблица журнала действий -->
        <div v-else class="table-wrapper">
          <table class="yandex-games-table">
            <thead>
              <tr>
                <th class="col-date">{{ t('moderation.dateTimeColumn') }}</th>
                <th class="col-action-type">{{ t('moderation.actionColumn') }}</th>
                <th class="col-project">{{ t('moderation.projectColumn') }}</th>
                <th class="col-details">{{ t('moderation.detailsColumn') }}</th>
                <th class="col-actions"></th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in filteredActivity" :key="item.id" class="table-row">
                <!-- Дата и время -->
                <td class="col-date">
                  <span class="date-text">{{ formatDateTime(item.created_at) }}</span>
                </td>

                <!-- Тип действия (бейдж) -->
                <td class="col-action-type">
                  <span class="action-pill" :class="actionBadgeClass(item.action_type)">
                    {{ item.action_title || getActionLabel(item.action_type) }}
                  </span>
                </td>

                <!-- Проект -->
                <td class="col-project">
                  <div class="project-cell">
                    <img
                      v-if="item.project_icon"
                      :src="getMediaUrl(item.project_icon)"
                      class="project-thumb"
                      alt=""
                    />
                    <div v-else class="project-thumb-mock">
                      <Gamepad2 class="icon-xs" />
                    </div>
                    <div class="project-info">
                      <div class="project-title-row">
                        <span class="project-name">
                          {{ item.project_title || `Проект #${item.project_id}` }}
                        </span>
                        <span v-if="item.build_version" class="version-tag">
                          v{{ item.build_version }}
                        </span>
                      </div>
                      <span class="project-id-sub">#{{ item.project_id }}</span>
                    </div>
                  </div>
                </td>

                <!-- Детали / Содержание сообщения или вердикта -->
                <td class="col-details">
                  <div class="details-content-box">
                    <span v-if="item.details" class="details-text" :title="item.details">
                      {{ item.details }}
                    </span>
                    <span v-else class="text-muted">—</span>
                  </div>
                </td>

                <!-- Кнопка перехода к проекту -->
                <td class="col-actions">
                  <button
                    class="btn-action btn-to-project"
                    title="Открыть проект на проверке"
                    @click="goToProject(item.project_id)"
                  >
                    <span>{{ t('moderation.openProject') }}</span>
                    <ExternalLink class="icon-xs" />
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Пагинация журнала -->
        <div v-if="totalActivity > 0" class="pagination-container">
          <div class="pagination-center">
            <button
              class="page-nav-btn"
              :disabled="currentPage === 1"
              title="Первая страница"
              @click="goToPage(1)"
            >
              <ChevronsLeft class="icon-sm" />
            </button>
            <button
              class="page-nav-btn"
              :disabled="currentPage === 1"
              title="Предыдущая"
              @click="goToPage(currentPage - 1)"
            >
              <ChevronLeft class="icon-sm" />
            </button>

            <div class="page-numbers">
              <button
                v-for="page in visiblePages"
                :key="page"
                class="page-num-btn"
                :class="{ active: page === currentPage, 'page-ellipsis': page === '...' }"
                :disabled="page === '...'"
                @click="typeof page === 'number' && goToPage(page)"
              >
                {{ page }}
              </button>
            </div>

            <button
              class="page-nav-btn"
              :disabled="currentPage === totalPages"
              title="Следующая"
              @click="goToPage(currentPage + 1)"
            >
              <ChevronRight class="icon-sm" />
            </button>
            <button
              class="page-nav-btn"
              :disabled="currentPage === totalPages"
              title="Последняя страница"
              @click="goToPage(totalPages)"
            >
              <ChevronsRight class="icon-sm" />
            </button>

            <div class="page-size-wrap">
              <select v-model="pageSize" class="page-size-select" @change="handlePageSizeChange">
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
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import {
  RefreshCw,
  Search,
  X,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
  RotateCcw,
  ExternalLink,
  Gamepad2,
} from 'lucide-vue-next';
import { getUser, searchUsers } from '@/entities/user';
import { moderationApi, formatDurationSeconds } from '@/entities/moderation';
import { getMediaUrl } from '@/entities/project';
import { formatProjectDate, formatDateTime, showToast } from '@/shared/lib';

const props = defineProps({
  id: {
    type: String,
    default: '',
  },
});

const route = useRoute();
const router = useRouter();
const { t } = useI18n();

const moderatorId = computed(() => props.id || route.params.id);

const moderator = ref({});
const stats = ref({});
const activityItems = ref([]);
const totalActivity = ref(0);

const loadingStats = ref(false);
const loadingActivity = ref(false);

const actionTypeFilter = ref('');
const searchQuery = ref('');
const currentPage = ref(1);
const pageSize = ref(10);

onMounted(() => {
  if (moderatorId.value) {
    refreshAll();
  }
});

async function refreshAll() {
  await Promise.all([loadModeratorInfo(), loadStats(), loadActivity()]);
}

async function loadModeratorInfo() {
  try {
    const res = await getUser(moderatorId.value);
    if (res && res.user) {
      moderator.value = res.user;
    } else if (res && res.id) {
      moderator.value = res;
    } else {
      await fallbackSearchModerator();
    }
  } catch {
    await fallbackSearchModerator();
  }
}

async function fallbackSearchModerator() {
  try {
    const sRes = await searchUsers({ query: moderatorId.value, limit: 10 });
    const found = (sRes.users || []).find((u) => u.id === moderatorId.value);
    if (found) {
      moderator.value = found;
    } else {
      moderator.value = { id: moderatorId.value };
    }
  } catch {
    moderator.value = { id: moderatorId.value };
  }
}

async function loadStats() {
  loadingStats.value = true;
  try {
    const res = await moderationApi.getModeratorStats(moderatorId.value);
    stats.value = res.stats || {};
  } catch (err) {
    console.warn('Failed to load moderator stats', err);
    stats.value = {};
  } finally {
    loadingStats.value = false;
  }
}

async function loadActivity() {
  loadingActivity.value = true;
  try {
    const offset = (currentPage.value - 1) * pageSize.value;
    const res = await moderationApi.listModeratorActivity(moderatorId.value, {
      action_type: actionTypeFilter.value,
      limit: pageSize.value,
      offset,
    });
    activityItems.value = res.items || [];
    totalActivity.value = res.total || 0;
  } catch (err) {
    console.warn('Failed to load moderator activity', err);
    activityItems.value = [];
    totalActivity.value = 0;
    showToast('Не удалось загрузить журнал действий', 'danger');
  } finally {
    loadingActivity.value = false;
  }
}

function handleFilterChange() {
  currentPage.value = 1;
  loadActivity();
}

function handlePageSizeChange() {
  currentPage.value = 1;
  loadActivity();
}

function goToPage(page) {
  if (page < 1 || page > totalPages.value) return;
  currentPage.value = page;
  loadActivity();
}

function resetJournalFilters() {
  searchQuery.value = '';
  actionTypeFilter.value = '';
  currentPage.value = 1;
  loadActivity();
}

function goToProject(projectId) {
  if (!projectId) return;
  router.push(`/moderator/projects/${projectId}`);
}

const filteredActivity = computed(() => {
  let list = activityItems.value;
  if (searchQuery.value.trim()) {
    const q = searchQuery.value.trim().toLowerCase();
    list = list.filter(
      (item) =>
        (item.project_title && item.project_title.toLowerCase().includes(q)) ||
        String(item.project_id).includes(q) ||
        (item.details && item.details.toLowerCase().includes(q)) ||
        (item.action_title && item.action_title.toLowerCase().includes(q))
    );
  }
  return list;
});

const totalPages = computed(() => {
  return Math.max(1, Math.ceil(totalActivity.value / pageSize.value));
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

function actionBadgeClass(actionType) {
  switch (actionType) {
    case 'chat_message':
      return 'action-chat';
    case 'dialog_closed':
      return 'action-closed';
    case 'claimed':
      return 'action-claimed';
    case 'approved':
      return 'action-approved';
    case 'rejected':
      return 'action-rejected';
    default:
      return 'action-default';
  }
}

function getActionLabel(actionType) {
  switch (actionType) {
    case 'chat_message':
      return t('moderation.actionChatMessage');
    case 'dialog_closed':
      return t('moderation.actionDialogClosed');
    case 'claimed':
      return t('moderation.actionClaimed');
    case 'approved':
      return t('moderation.actionApproved');
    case 'rejected':
      return t('moderation.actionRejected');
    default:
      return actionType || 'Действие';
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
  gap: 28px;
}

/* Карточка заголовка модератора */
.moderator-header-card {
  width: 100%;
  background: var(--bg-secondary, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-md, 8px);
  padding: 20px 24px;
  box-sizing: border-box;
}

.header-main-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 20px;
}

.identity-info {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.title-with-badges {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.moderator-name {
  margin: 0;
  font-size: 22px;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
  letter-spacing: -0.2px;
}

.role-badge {
  display: inline-flex;
  align-items: center;
  height: 22px;
  padding: 0 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
  background: rgba(88, 166, 255, 0.12);
  border: 1px solid rgba(88, 166, 255, 0.35);
  color: var(--primary, #58a6ff);
}

.status-pill {
  display: inline-flex;
  align-items: center;
  height: 22px;
  padding: 0 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
}

.status-active {
  background: rgba(63, 185, 80, 0.12);
  border: 1px solid rgba(63, 185, 80, 0.35);
  color: #3fb950;
}

.status-deleted {
  background: rgba(248, 81, 73, 0.12);
  border: 1px solid rgba(248, 81, 73, 0.35);
  color: #f85149;
}

.meta-row {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--text-secondary, #8b949e);
  flex-wrap: wrap;
}

.meta-item {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.meta-label {
  color: var(--text-tertiary, #6e7681);
}

.meta-value {
  color: var(--text-main, #f0f6fc);
}

.font-mono {
  font-family: var(--font-mono, monospace);
  font-size: 12px;
}

.meta-dot {
  color: var(--border, #30363d);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.btn-refresh {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  height: 34px;
  padding: 0 14px;
  background: var(--bg-tertiary, #21262d);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  color: var(--text-main, #f0f6fc);
  font-size: 13px;
  cursor: pointer;
  transition:
    background-color 0.15s,
    border-color 0.15s;
}

.btn-refresh:hover:not(:disabled) {
  background: var(--border, #30363d);
  border-color: var(--text-tertiary, #8b949e);
}

.spin-animation {
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

/* Секции */
.section-block {
  display: flex;
  flex-direction: column;
  gap: 16px;
  width: 100%;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.section-title-wrap {
  display: flex;
  align-items: center;
  gap: 10px;
}

.section-title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
  letter-spacing: -0.1px;
}

.count-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 2px 8px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
  background: var(--bg-secondary, #161b22);
  border: 1px solid var(--border, #30363d);
  color: var(--text-secondary, #8b949e);
}

/* Сетка показателей (лаконичный, не пестрый стиль) */
.indicators-grid {
  display: grid;
  grid-template-columns: repeat(6, minmax(0, 1fr));
  gap: 14px;
  width: 100%;
}

@media (max-width: 1280px) {
  .indicators-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

@media (max-width: 768px) {
  .indicators-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

.indicator-card {
  background: var(--bg-secondary, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  padding: 14px 16px;
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-height: 86px;
  box-sizing: border-box;
}

.indicator-label {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary, #8b949e);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.indicator-value-row {
  display: flex;
  align-items: baseline;
  gap: 8px;
}

.indicator-value {
  font-size: 24px;
  font-weight: 700;
  color: var(--text-main, #f0f6fc);
  line-height: 1.1;
  font-variant-numeric: tabular-nums;
}

.indicator-rate-pill {
  display: inline-flex;
  align-items: center;
  padding: 1px 6px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
  line-height: 1.2;
}

.pill-success {
  background: rgba(63, 185, 80, 0.12);
  color: #3fb950;
  border: 1px solid rgba(63, 185, 80, 0.3);
}

.pill-danger {
  background: rgba(248, 81, 73, 0.12);
  color: #f85149;
  border: 1px solid rgba(248, 81, 73, 0.3);
}

.indicator-subtext {
  font-size: 11px;
  line-height: 1.3;
}

.subtext-muted {
  color: var(--text-tertiary, #6e7681);
}

.text-success {
  color: #3fb950 !important;
}

.text-danger {
  color: #f85149 !important;
}

.text-info {
  color: #38bdf8 !important;
}

/* Горизонтальный стрип периодов */
.period-strip {
  display: flex;
  align-items: center;
  gap: 16px;
  background: var(--bg-secondary, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  padding: 10px 16px;
  width: 100%;
  box-sizing: border-box;
  font-size: 13px;
  flex-wrap: wrap;
}

.period-title {
  color: var(--text-tertiary, #6e7681);
  font-weight: 500;
  margin-right: 4px;
}

.period-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.period-label {
  color: var(--text-secondary, #8b949e);
}

.period-count {
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
  font-variant-numeric: tabular-nums;
}

.period-divider {
  width: 1px;
  height: 14px;
  background: var(--border, #30363d);
}

.font-bold {
  font-weight: 700;
}

/* Панель фильтров */
.filters-toolbar {
  display: flex;
  align-items: flex-end;
  gap: 16px;
  width: 100%;
}

.filter-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.field-search {
  flex: 1;
  min-width: 260px;
}

.field-action-type {
  width: 220px;
  flex-shrink: 0;
}

.field-label {
  font-size: 13px;
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
  transition: border-color 0.15s;
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

.select-arrow {
  position: absolute;
  right: 12px;
  pointer-events: none;
  color: var(--text-tertiary, #8b949e);
}

.clear-input-btn {
  position: absolute;
  right: 8px;
  background: transparent;
  border: none;
  color: var(--text-tertiary, #8b949e);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 4px;
}

.clear-input-btn:hover {
  color: var(--text-main, #f0f6fc);
}

.btn-reset-filters {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 38px;
  padding: 0 14px;
  background: transparent;
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  color: var(--text-secondary, #8b949e);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.15s;
}

.btn-reset-filters:hover {
  background: var(--bg-secondary, #161b22);
  color: var(--text-main, #f0f6fc);
}

/* Таблица журнала */
.table-wrapper {
  width: 100%;
  overflow-x: auto;
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  background: var(--bg-secondary, #161b22);
}

.yandex-games-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
}

.yandex-games-table th {
  color: var(--text-tertiary, #8b949e);
  font-size: 12px;
  font-weight: 500;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border, #30363d);
  background: rgba(255, 255, 255, 0.02);
  white-space: nowrap;
}

.col-date {
  width: 170px;
}

.col-action-type {
  width: 160px;
}

.col-project {
  width: 260px;
}

.col-details {
  min-width: 280px;
}

.col-actions {
  width: 130px;
  text-align: right;
  padding-right: 16px;
}

.table-row {
  border-bottom: 1px solid var(--border, #21262d);
  transition: background-color 0.15s ease;
}

.table-row:last-child {
  border-bottom: none;
}

.table-row:hover {
  background: rgba(255, 255, 255, 0.03);
}

.table-row td {
  padding: 12px 16px;
  vertical-align: middle;
}

.date-text {
  font-size: 13px;
  color: var(--text-secondary, #8b949e);
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}

/* Бейджи действий */
.action-pill {
  display: inline-flex;
  align-items: center;
  height: 22px;
  padding: 0 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
  white-space: nowrap;
}

.action-chat {
  background: rgba(139, 148, 158, 0.15);
  border: 1px solid rgba(139, 148, 158, 0.35);
  color: var(--text-main, #f0f6fc);
}

.action-closed {
  background: rgba(56, 189, 248, 0.12);
  border: 1px solid rgba(56, 189, 248, 0.35);
  color: #38bdf8;
}

.action-claimed {
  background: rgba(210, 153, 34, 0.12);
  border: 1px solid rgba(210, 153, 34, 0.35);
  color: #e3b341;
}

.action-approved {
  background: rgba(63, 185, 80, 0.12);
  border: 1px solid rgba(63, 185, 80, 0.35);
  color: #3fb950;
}

.action-rejected {
  background: rgba(248, 81, 73, 0.12);
  border: 1px solid rgba(248, 81, 73, 0.35);
  color: #f85149;
}

.action-default {
  background: var(--bg-tertiary, #21262d);
  border: 1px solid var(--border, #30363d);
  color: var(--text-secondary, #8b949e);
}

/* Ячейка проекта */
.project-cell {
  display: flex;
  align-items: center;
  gap: 10px;
}

.project-thumb {
  width: 32px;
  height: 32px;
  border-radius: 6px;
  object-fit: cover;
  background: var(--bg-tertiary, #21262d);
  flex-shrink: 0;
  border: 1px solid var(--border, #30363d);
}

.project-thumb-mock {
  width: 32px;
  height: 32px;
  border-radius: 6px;
  background: var(--bg-tertiary, #21262d);
  border: 1px solid var(--border, #30363d);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-tertiary, #8b949e);
  flex-shrink: 0;
}

.project-info {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.project-title-row {
  display: flex;
  align-items: center;
  gap: 6px;
}

.project-name {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-main, #f0f6fc);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 170px;
}

.version-tag {
  font-size: 10px;
  padding: 1px 4px;
  border-radius: 3px;
  background: var(--bg-tertiary, #21262d);
  border: 1px solid var(--border, #30363d);
  color: var(--text-secondary, #8b949e);
  font-family: var(--font-mono, monospace);
  flex-shrink: 0;
}

.project-id-sub {
  font-size: 11px;
  color: var(--text-tertiary, #6e7681);
  font-family: var(--font-mono, monospace);
}

/* Блок деталей */
.details-content-box {
  max-width: 480px;
}

.details-text {
  font-size: 13px;
  color: var(--text-main, #f0f6fc);
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Кнопка "К проекту" */
.btn-to-project {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: transparent;
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  color: var(--primary, #58a6ff);
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
  white-space: nowrap;
}

.btn-to-project:hover {
  background: rgba(88, 166, 255, 0.1);
  border-color: rgba(88, 166, 255, 0.4);
}

/* Состояния пустоты / загрузки */
.state-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px 24px;
  gap: 12px;
  color: var(--text-secondary, #8b949e);
  background: var(--bg-secondary, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
}

.state-container h3 {
  margin: 0;
  font-size: 16px;
  color: var(--text-main, #f0f6fc);
}

.state-container p {
  margin: 0;
  font-size: 13px;
}

.btn-reset {
  margin-top: 8px;
  padding: 6px 14px;
  background: var(--bg-tertiary, #21262d);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  color: var(--text-main, #f0f6fc);
  font-size: 13px;
  cursor: pointer;
}

.btn-reset:hover {
  background: var(--border, #30363d);
}

/* Пагинация */
.pagination-container {
  display: flex;
  justify-content: center;
  margin-top: 16px;
  width: 100%;
}

.pagination-center {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.page-nav-btn,
.page-num-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 32px;
  height: 32px;
  padding: 0 6px;
  border-radius: var(--radius-sm, 6px);
  background: var(--bg-secondary, #161b22);
  border: 1px solid var(--border, #30363d);
  color: var(--text-main, #f0f6fc);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.15s;
}

.page-nav-btn:hover:not(:disabled),
.page-num-btn:hover:not(:disabled) {
  background: var(--bg-tertiary, #21262d);
  border-color: var(--text-tertiary, #8b949e);
}

.page-nav-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.page-num-btn.active {
  background: var(--primary, #58a6ff);
  border-color: var(--primary, #58a6ff);
  color: #ffffff;
  font-weight: 600;
}

.page-numbers {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.page-ellipsis {
  border: none;
  background: transparent;
  cursor: default;
}

.page-size-wrap {
  position: relative;
  display: inline-flex;
  align-items: center;
  margin-left: 8px;
}

.page-size-select {
  height: 32px;
  padding: 0 24px 0 10px;
  border-radius: var(--radius-sm, 6px);
  background: var(--bg-secondary, #161b22);
  border: 1px solid var(--border, #30363d);
  color: var(--text-main, #f0f6fc);
  font-size: 13px;
  cursor: pointer;
  appearance: none;
  -webkit-appearance: none;
  outline: none;
}

.select-caret {
  position: absolute;
  right: 8px;
  pointer-events: none;
  color: var(--text-tertiary, #8b949e);
}

/* Иконки */
.icon-xs {
  width: 14px;
  height: 14px;
}

.icon-sm {
  width: 16px;
  height: 16px;
}

.icon-md {
  width: 24px;
  height: 24px;
}

.spinner-md {
  width: 28px;
  height: 28px;
  border: 3px solid rgba(88, 166, 255, 0.2);
  border-top-color: var(--primary, #58a6ff);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}
</style>
