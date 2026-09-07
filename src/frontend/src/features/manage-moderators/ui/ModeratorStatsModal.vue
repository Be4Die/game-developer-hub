<template>
  <div class="modal-overlay" @click.self="$emit('close')">
    <div class="modal-card modal-large">
      <!-- Шапка модального окна -->
      <div class="modal-header">
        <div class="modal-title-wrap">
          <div class="user-avatar-badge">
            <ShieldCheck class="icon-md text-primary" />
          </div>
          <div class="user-info-header">
            <div class="user-title-row">
              <h3>{{ moderator.display_name || moderator.email || 'Модератор' }}</h3>
              <span class="status-pill" :class="statusBadgeClass(moderator.status)">
                {{ statusLabel(moderator.status) }}
              </span>
            </div>
            <p class="modal-subtitle">
              <span>{{ moderator.email }}</span>
              <span v-if="moderator.id" class="meta-dot">•</span>
              <span v-if="moderator.id">ID: {{ moderator.id }}</span>
              <span v-if="moderator.created_at" class="meta-dot">•</span>
              <span v-if="moderator.created_at">
                Создан: {{ formatProjectDate(moderator.created_at) }}
              </span>
            </p>
          </div>
        </div>
        <button class="modal-close" title="Закрыть" @click="$emit('close')">
          <X class="icon-sm" />
        </button>
      </div>

      <!-- Содержимое модального окна -->
      <div class="modal-body">
        <!-- Состояние загрузки общей статистики -->
        <div v-if="loadingStats" class="loading-state">
          <div class="spinner-md"></div>
          <p>Загрузка статистики модератора...</p>
        </div>

        <div v-else class="stats-content">
          <!-- Сетка ключевых числовых показателей -->
          <div class="metrics-grid">
            <!-- 1. Всего рассмотрено -->
            <div class="metric-card">
              <div class="metric-icon-wrap bg-primary-soft">
                <CheckSquare class="icon-md text-primary" />
              </div>
              <div class="metric-content">
                <div class="metric-value">{{ stats.total_resolved || 0 }}</div>
                <div class="metric-label">{{ t('moderation.totalResolved') }}</div>
              </div>
            </div>

            <!-- 2. Одобрено -->
            <div class="metric-card">
              <div class="metric-icon-wrap bg-success-soft">
                <CheckCircle2 class="icon-md text-success" />
              </div>
              <div class="metric-content">
                <div class="metric-value-row">
                  <span class="metric-value text-success">{{ stats.approved_count || 0 }}</span>
                  <span v-if="stats.total_resolved > 0" class="metric-sub-badge bg-success-soft">
                    {{ stats.approval_rate }}%
                  </span>
                </div>
                <div class="metric-label">{{ t('moderation.approvedStats') }}</div>
              </div>
            </div>

            <!-- 3. Отклонено -->
            <div class="metric-card">
              <div class="metric-icon-wrap bg-danger-soft">
                <XCircle class="icon-md text-danger" />
              </div>
              <div class="metric-content">
                <div class="metric-value-row">
                  <span class="metric-value text-danger">{{ stats.rejected_count || 0 }}</span>
                  <span v-if="stats.total_resolved > 0" class="metric-sub-badge bg-danger-soft">
                    {{ stats.rejection_rate }}%
                  </span>
                </div>
                <div class="metric-label">{{ t('moderation.rejectedStats') }}</div>
              </div>
            </div>

            <!-- 4. В процессе проверки -->
            <div class="metric-card">
              <div class="metric-icon-wrap bg-info-soft">
                <Clock class="icon-md text-info" />
              </div>
              <div class="metric-content">
                <div class="metric-value">{{ stats.in_review_count || 0 }}</div>
                <div class="metric-label">{{ t('moderation.inReviewStats') }}</div>
              </div>
            </div>

            <!-- 5. Среднее время проверки -->
            <div class="metric-card">
              <div class="metric-icon-wrap bg-warning-soft">
                <Hourglass class="icon-md text-warning" />
              </div>
              <div class="metric-content">
                <div class="metric-value">
                  {{ formatDurationSeconds(stats.avg_review_duration_seconds) }}
                </div>
                <div class="metric-label">{{ t('moderation.avgReviewDuration') }}</div>
              </div>
            </div>

            <!-- 6. Сообщений в чатах -->
            <div class="metric-card">
              <div class="metric-icon-wrap bg-purple-soft">
                <MessageSquare class="icon-md text-purple" />
              </div>
              <div class="metric-content">
                <div class="metric-value">{{ stats.messages_sent || 0 }}</div>
                <div class="metric-label">{{ t('moderation.messagesSent') }}</div>
              </div>
            </div>
          </div>

          <!-- Срез по периодам активности -->
          <div class="periods-panel">
            <span class="periods-title">Рассмотрено решений по периодам:</span>
            <div class="periods-chips">
              <div class="period-chip">
                <span class="period-label">{{ t('moderation.periodToday') }}:</span>
                <span class="period-val">{{ stats.today_resolved || 0 }}</span>
              </div>
              <div class="period-chip">
                <span class="period-label">{{ t('moderation.periodWeek') }}:</span>
                <span class="period-val">{{ stats.week_resolved || 0 }}</span>
              </div>
              <div class="period-chip">
                <span class="period-label">{{ t('moderation.periodMonth') }}:</span>
                <span class="period-val">{{ stats.month_resolved || 0 }}</span>
              </div>
              <div class="period-chip">
                <span class="period-label">{{ t('moderation.periodAllTime') }}:</span>
                <span class="period-val">{{ stats.total_resolved || 0 }}</span>
              </div>
            </div>
          </div>

          <!-- Разделитель -->
          <div class="section-divider"></div>

          <!-- Журнал действий модератора -->
          <div class="journal-section">
            <div class="journal-header">
              <div class="journal-title-wrap">
                <FileText class="icon-sm text-primary" />
                <h4>{{ t('moderation.activityJournal') }}</h4>
                <span class="journal-badge">{{ activityTotal }}</span>
              </div>

              <!-- Панель управления журналом -->
              <div class="journal-controls">
                <!-- Поиск по журналу -->
                <div class="journal-search-wrap">
                  <Search class="icon-xs search-icon" />
                  <input
                    v-model="journalSearch"
                    type="text"
                    placeholder="Поиск по проекту или ID..."
                    class="journal-search-input"
                  />
                  <button
                    v-if="journalSearch"
                    class="journal-clear-btn"
                    @click="journalSearch = ''"
                  >
                    <X class="icon-xs" />
                  </button>
                </div>

                <!-- Фильтр по статусу -->
                <div class="journal-select-wrap">
                  <select
                    v-model="statusFilter"
                    class="journal-select"
                    @change="onStatusFilterChange"
                  >
                    <option value="">{{ t('moderation.allVerdicts') }}</option>
                    <option value="3">{{ t('moderation.approved') }}</option>
                    <option value="4">{{ t('moderation.rejected') }}</option>
                    <option value="2">{{ t('moderation.inReview') }}</option>
                  </select>
                  <ChevronDown class="icon-xs select-caret" />
                </div>

                <!-- Кнопка обновления журнала -->
                <button
                  class="btn-icon-refresh"
                  title="Обновить журнал"
                  :disabled="loadingActivity"
                  @click="loadActivity"
                >
                  <RefreshCw class="icon-xs" :class="{ spin: loadingActivity }" />
                </button>
              </div>
            </div>

            <!-- Таблица журнала -->
            <div v-if="loadingActivity" class="journal-loading">
              <div class="spinner-sm"></div>
              <span>Загрузка журнала действий...</span>
            </div>

            <div v-else-if="filteredActivity.length === 0" class="journal-empty">
              <FileQuestion class="icon-md text-muted" />
              <p>{{ t('moderation.noActivity') }}</p>
            </div>

            <div v-else class="journal-table-wrap">
              <table class="journal-table">
                <thead>
                  <tr>
                    <th class="col-date">Дата и время</th>
                    <th class="col-project">Проект</th>
                    <th class="col-status">Вердикт</th>
                    <th class="col-duration">{{ t('moderation.duration') }}</th>
                    <th class="col-reason">Причина / Примечание</th>
                    <th class="col-link"></th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="item in filteredActivity" :key="item.id" class="journal-row">
                    <!-- 1. Дата и время -->
                    <td class="col-date">
                      <span class="cell-date">
                        {{
                          formatDateTime(
                            item.resolved_at || item.started_review_at || item.submitted_at
                          )
                        }}
                      </span>
                    </td>

                    <!-- 2. Проект -->
                    <td class="col-project">
                      <div class="project-cell-wrap">
                        <img
                          v-if="item.snapshot && item.snapshot.icon_path"
                          :src="item.snapshot.icon_path"
                          alt="icon"
                          class="project-mini-icon"
                          @error="$event.target.style.display = 'none'"
                        />
                        <div v-else class="project-mini-placeholder">
                          <Gamepad2 class="icon-xs" />
                        </div>
                        <div class="project-cell-info">
                          <span class="project-cell-title">
                            {{
                              item.snapshot?.title_ru ||
                              item.snapshot?.title_en ||
                              `Проект #${item.project_id}`
                            }}
                          </span>
                          <span class="project-cell-meta">
                            #{{ item.project_id }}
                            <template v-if="item.snapshot?.active_build_version">
                              • v{{ item.snapshot.active_build_version }}
                            </template>
                          </span>
                        </div>
                      </div>
                    </td>

                    <!-- 3. Статус / Вердикт -->
                    <td class="col-status">
                      <span class="badge" :class="getStatusBadge(item.status)">
                        {{ getStatusLabel(item.status) }}
                      </span>
                    </td>

                    <!-- 4. Время проверки -->
                    <td class="col-duration">
                      <span class="duration-cell">
                        {{ getRequestDuration(item) }}
                      </span>
                    </td>

                    <!-- 5. Причина отказа / замечания -->
                    <td class="col-reason">
                      <div
                        v-if="item.rejection_reason"
                        class="reason-pill"
                        :title="item.rejection_reason"
                      >
                        {{ item.rejection_reason }}
                      </div>
                      <span v-else class="text-muted">—</span>
                    </td>

                    <!-- 6. Ссылка на проект -->
                    <td class="col-link">
                      <button
                        class="btn-open-project"
                        title="Открыть проект модерации"
                        @click="openProject(item.project_id)"
                      >
                        <span>Проект</span>
                        <ExternalLink class="icon-xs" />
                      </button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <!-- Пагинация журнала -->
            <div v-if="activityTotal > pageSize" class="journal-pagination">
              <button
                class="btn-page"
                :disabled="currentPage === 1"
                @click="changePage(currentPage - 1)"
              >
                <ChevronLeft class="icon-xs" />
                <span>Назад</span>
              </button>

              <span class="page-indicator"> Страница {{ currentPage }} из {{ totalPages }} </span>

              <button
                class="btn-page"
                :disabled="currentPage === totalPages"
                @click="changePage(currentPage + 1)"
              >
                <span>Вперед</span>
                <ChevronRight class="icon-xs" />
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Подвал модалки -->
      <div class="modal-footer">
        <button class="btn-modal-secondary" @click="$emit('close')">Закрыть</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import {
  ShieldCheck,
  X,
  CheckSquare,
  CheckCircle2,
  XCircle,
  Clock,
  Hourglass,
  MessageSquare,
  FileText,
  Search,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  RefreshCw,
  Gamepad2,
  ExternalLink,
  FileQuestion,
} from 'lucide-vue-next';
import {
  moderationApi,
  formatDurationSeconds,
  getRequestDuration,
  formatDateTime,
} from '@/entities/moderation';
import { formatProjectDate, showToast } from '@/shared/lib';

const props = defineProps({
  moderator: {
    type: Object,
    required: true,
  },
});

defineEmits(['close']);

const { t } = useI18n();
const router = useRouter();

const loadingStats = ref(true);
const loadingActivity = ref(false);

const stats = ref({
  total_resolved: 0,
  approved_count: 0,
  rejected_count: 0,
  in_review_count: 0,
  approval_rate: 0,
  rejection_rate: 0,
  avg_review_duration_seconds: 0,
  messages_sent: 0,
  today_resolved: 0,
  week_resolved: 0,
  month_resolved: 0,
});

const activityList = ref([]);
const activityTotal = ref(0);
const statusFilter = ref('');
const journalSearch = ref('');
const currentPage = ref(1);
const pageSize = ref(10);

onMounted(() => {
  loadAll();
});

async function loadAll() {
  await Promise.all([loadStats(), loadActivity()]);
}

async function loadStats() {
  loadingStats.value = true;
  try {
    const res = await moderationApi.getModeratorStats(props.moderator.id);
    if (res.stats) {
      stats.value = res.stats;
    }
  } catch (err) {
    showToast('Не удалось загрузить статистику модератора', 'danger');
  } finally {
    loadingStats.value = false;
  }
}

async function loadActivity() {
  loadingActivity.value = true;
  try {
    const offset = (currentPage.value - 1) * pageSize.value;
    const res = await moderationApi.listModeratorActivity(props.moderator.id, {
      status: statusFilter.value ? Number(statusFilter.value) : undefined,
      limit: pageSize.value,
      offset,
    });
    activityList.value = res.requests || [];
    activityTotal.value = res.total || 0;
  } catch (err) {
    showToast('Не удалось загрузить журнал действий', 'danger');
  } finally {
    loadingActivity.value = false;
  }
}

function onStatusFilterChange() {
  currentPage.value = 1;
  loadActivity();
}

function changePage(page) {
  if (page < 1 || page > totalPages.value) return;
  currentPage.value = page;
  loadActivity();
}

const totalPages = computed(() => {
  return Math.ceil(activityTotal.value / pageSize.value) || 1;
});

const filteredActivity = computed(() => {
  if (!journalSearch.value.trim()) {
    return activityList.value;
  }
  const q = journalSearch.value.trim().toLowerCase();
  return activityList.value.filter((item) => {
    const titleRu = item.snapshot?.title_ru?.toLowerCase() || '';
    const titleEn = item.snapshot?.title_en?.toLowerCase() || '';
    const pId = String(item.project_id || '');
    return titleRu.includes(q) || titleEn.includes(q) || pId.includes(q);
  });
});

function openProject(projectId) {
  router.push(`/moderator/projects/${projectId}`);
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

function getStatusBadge(status) {
  if (status === 3 || status === 'REQUEST_STATUS_APPROVED' || status === 'approved') {
    return 'badge-success';
  }
  if (status === 4 || status === 'REQUEST_STATUS_REJECTED' || status === 'rejected') {
    return 'badge-danger';
  }
  if (status === 2 || status === 'REQUEST_STATUS_IN_REVIEW' || status === 'in_review') {
    return 'badge-info';
  }
  return 'badge-neutral';
}

function getStatusLabel(status) {
  if (status === 3 || status === 'REQUEST_STATUS_APPROVED' || status === 'approved') {
    return t('moderation.approved');
  }
  if (status === 4 || status === 'REQUEST_STATUS_REJECTED' || status === 'rejected') {
    return t('moderation.rejected');
  }
  if (status === 2 || status === 'REQUEST_STATUS_IN_REVIEW' || status === 'in_review') {
    return t('moderation.inReview');
  }
  return '—';
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.75);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 20px;
  box-sizing: border-box;
}

.modal-card.modal-large {
  background: var(--bg-secondary, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-md, 8px);
  width: 100%;
  max-width: 960px;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 12px 36px rgba(0, 0, 0, 0.5);
  overflow: hidden;
}

/* Шапка */
.modal-header {
  padding: 20px 24px;
  border-bottom: 1px solid var(--border, #30363d);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  background: var(--bg-app, #0d1117);
}

.modal-title-wrap {
  display: flex;
  align-items: center;
  gap: 14px;
}

.user-avatar-badge {
  width: 44px;
  height: 44px;
  border-radius: 8px;
  background: rgba(88, 166, 255, 0.1);
  border: 1px solid rgba(88, 166, 255, 0.3);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.user-info-header {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.user-title-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.user-title-row h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
}

.modal-subtitle {
  margin: 0;
  font-size: 13px;
  color: var(--text-tertiary, #8b949e);
  display: flex;
  align-items: center;
  gap: 6px;
}

.meta-dot {
  color: var(--border, #484f58);
}

.modal-close {
  background: none;
  border: none;
  color: var(--text-tertiary, #8b949e);
  cursor: pointer;
  padding: 6px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s;
}

.modal-close:hover {
  color: var(--text-main, #f0f6fc);
  background: var(--bg-tertiary, #21262d);
}

/* Тело */
.modal-body {
  padding: 20px 24px;
  overflow-y: auto;
  flex: 1;
  box-sizing: border-box;
}

.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 0;
  gap: 12px;
  color: var(--text-tertiary, #8b949e);
}

/* Сетка метрик */
.metrics-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 14px;
  margin-bottom: 16px;
}

@media (max-width: 768px) {
  .metrics-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

.metric-card {
  background: var(--bg-app, #0d1117);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  padding: 14px 16px;
  display: flex;
  align-items: center;
  gap: 14px;
}

.metric-icon-wrap {
  width: 40px;
  height: 40px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.bg-primary-soft {
  background: rgba(88, 166, 255, 0.12);
}
.bg-success-soft {
  background: rgba(46, 204, 113, 0.12);
}
.bg-danger-soft {
  background: rgba(248, 81, 73, 0.12);
}
.bg-info-soft {
  background: rgba(56, 189, 248, 0.12);
}
.bg-warning-soft {
  background: rgba(234, 179, 8, 0.12);
}
.bg-purple-soft {
  background: rgba(168, 85, 247, 0.12);
}

.text-primary {
  color: var(--primary, #58a6ff);
}
.text-success {
  color: #2ecc71;
}
.text-danger {
  color: var(--danger, #f85149);
}
.text-info {
  color: #38bdf8;
}
.text-warning {
  color: #eab308;
}
.text-purple {
  color: #a855f7;
}

.metric-content {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.metric-value-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.metric-value {
  font-size: 20px;
  font-weight: 700;
  color: var(--text-main, #f0f6fc);
  line-height: 1.2;
}

.metric-sub-badge {
  font-size: 11px;
  font-weight: 600;
  padding: 2px 6px;
  border-radius: 4px;
}

.metric-label {
  font-size: 12px;
  color: var(--text-tertiary, #8b949e);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Периоды */
.periods-panel {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px;
  background: var(--bg-tertiary, #21262d);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  padding: 10px 16px;
}

.periods-title {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-muted, #b0b8c4);
}

.periods-chips {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.period-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: var(--bg-secondary, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: 4px;
  padding: 4px 10px;
  font-size: 12px;
}

.period-label {
  color: var(--text-tertiary, #8b949e);
}

.period-val {
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
}

.section-divider {
  height: 1px;
  background: var(--border, #30363d);
  margin: 20px 0;
}

/* Журнал */
.journal-section {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.journal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px;
}

.journal-title-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
}

.journal-title-wrap h4 {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
}

.journal-badge {
  font-size: 12px;
  font-weight: 600;
  padding: 2px 8px;
  background: var(--bg-tertiary, #21262d);
  border: 1px solid var(--border, #30363d);
  border-radius: 12px;
  color: var(--text-muted, #b0b8c4);
}

.journal-controls {
  display: flex;
  align-items: center;
  gap: 10px;
}

.journal-search-wrap {
  position: relative;
  display: flex;
  align-items: center;
  width: 220px;
  height: 32px;
  border-radius: 4px;
  border: 1px solid var(--border, #30363d);
  background: var(--bg-app, #0d1117);
}

.search-icon {
  position: absolute;
  left: 8px;
  color: var(--text-tertiary, #8b949e);
}

.journal-search-input {
  width: 100%;
  height: 100%;
  padding: 0 26px 0 28px;
  background: transparent;
  border: none;
  color: var(--text-main, #f0f6fc);
  font-size: 12px;
  outline: none;
}

.journal-clear-btn {
  position: absolute;
  right: 6px;
  background: none;
  border: none;
  color: var(--text-tertiary, #8b949e);
  cursor: pointer;
  padding: 2px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.journal-select-wrap {
  position: relative;
  display: flex;
  align-items: center;
  height: 32px;
  border-radius: 4px;
  border: 1px solid var(--border, #30363d);
  background: var(--bg-app, #0d1117);
}

.journal-select {
  height: 100%;
  padding: 0 24px 0 10px;
  background: transparent;
  border: none;
  color: var(--text-main, #f0f6fc);
  font-size: 12px;
  outline: none;
  cursor: pointer;
  appearance: none;
  -webkit-appearance: none;
}

.select-caret {
  position: absolute;
  right: 8px;
  pointer-events: none;
  color: var(--text-tertiary, #8b949e);
}

.btn-icon-refresh {
  width: 32px;
  height: 32px;
  border-radius: 4px;
  border: 1px solid var(--border, #30363d);
  background: var(--bg-app, #0d1117);
  color: var(--text-muted, #b0b8c4);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s;
}

.btn-icon-refresh:hover:not(:disabled) {
  color: var(--text-main, #f0f6fc);
  background: var(--bg-tertiary, #21262d);
}

/* Таблица журнала */
.journal-loading,
.journal-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 0;
  gap: 10px;
  color: var(--text-tertiary, #8b949e);
  font-size: 13px;
  background: var(--bg-app, #0d1117);
  border: 1px solid var(--border, #30363d);
  border-radius: 6px;
}

.journal-table-wrap {
  width: 100%;
  overflow-x: auto;
  border: 1px solid var(--border, #30363d);
  border-radius: 6px;
  background: var(--bg-app, #0d1117);
}

.journal-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
  font-size: 12px;
}

.journal-table th {
  padding: 10px 12px;
  color: var(--text-tertiary, #8b949e);
  font-weight: 500;
  border-bottom: 1px solid var(--border, #30363d);
  background: var(--bg-secondary, #161b22);
  white-space: nowrap;
}

.journal-table td {
  padding: 10px 12px;
  border-bottom: 1px solid var(--border, #21262d);
  vertical-align: middle;
}

.journal-row:hover {
  background: var(--bg-secondary, #161b22);
}

.col-date {
  width: 15%;
  white-space: nowrap;
}

.cell-date {
  color: var(--text-muted, #b0b8c4);
  font-size: 12px;
}

.col-project {
  width: 30%;
}

.project-cell-wrap {
  display: flex;
  align-items: center;
  gap: 10px;
}

.project-mini-icon {
  width: 28px;
  height: 28px;
  border-radius: 4px;
  object-fit: cover;
  background: var(--bg-tertiary, #21262d);
  flex-shrink: 0;
}

.project-mini-placeholder {
  width: 28px;
  height: 28px;
  border-radius: 4px;
  background: var(--bg-tertiary, #21262d);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-tertiary, #8b949e);
  flex-shrink: 0;
}

.project-cell-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.project-cell-title {
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 200px;
}

.project-cell-meta {
  font-size: 11px;
  color: var(--text-tertiary, #8b949e);
}

.col-status {
  width: 13%;
  white-space: nowrap;
}

.col-duration {
  width: 14%;
  white-space: nowrap;
}

.duration-cell {
  font-weight: 500;
  color: var(--text-muted, #b0b8c4);
}

.col-reason {
  width: 20%;
}

.reason-pill {
  max-width: 180px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  background: rgba(248, 81, 73, 0.1);
  border: 1px solid rgba(248, 81, 73, 0.3);
  color: #ff7b72;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 11px;
  display: inline-block;
}

.col-link {
  width: 8%;
  text-align: right;
  white-space: nowrap;
}

.btn-open-project {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  background: var(--bg-tertiary, #21262d);
  border: 1px solid var(--border, #30363d);
  color: var(--text-main, #f0f6fc);
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 11px;
  cursor: pointer;
  transition: all 0.15s;
}

.btn-open-project:hover {
  background: var(--bg-secondary, #161b22);
  border-color: var(--primary, #58a6ff);
  color: var(--primary, #58a6ff);
}

/* Пагинация */
.journal-pagination {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 12px;
  padding-top: 6px;
}

.btn-page {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  background: var(--bg-app, #0d1117);
  border: 1px solid var(--border, #30363d);
  color: var(--text-main, #f0f6fc);
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 12px;
  cursor: pointer;
}

.btn-page:hover:not(:disabled) {
  background: var(--bg-tertiary, #21262d);
}

.btn-page:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.page-indicator {
  font-size: 12px;
  color: var(--text-tertiary, #8b949e);
}

/* Подвал */
.modal-footer {
  padding: 14px 24px;
  border-top: 1px solid var(--border, #30363d);
  display: flex;
  align-items: center;
  justify-content: flex-end;
  background: var(--bg-app, #0d1117);
}

.btn-modal-secondary {
  background: var(--bg-tertiary, #21262d);
  border: 1px solid var(--border, #30363d);
  color: var(--text-main, #f0f6fc);
  padding: 8px 18px;
  border-radius: var(--radius-sm, 6px);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
}

.btn-modal-secondary:hover {
  background: var(--bg-secondary, #161b22);
  border-color: var(--border-secondary, #484f58);
}

/* Статусы */
.status-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 22px;
  padding: 0 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
  white-space: nowrap;
}

.status-active {
  background: rgba(46, 204, 113, 0.12);
  border: 1px solid rgba(46, 204, 113, 0.35);
  color: #2ecc71;
}

.status-deleted {
  background: var(--bg-tertiary, #21262d);
  border: 1px solid var(--border, #30363d);
  color: var(--text-tertiary, #8b949e);
}

.badge {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
  white-space: nowrap;
}

.badge-success {
  background: rgba(46, 204, 113, 0.12);
  border: 1px solid rgba(46, 204, 113, 0.35);
  color: #2ecc71;
}

.badge-danger {
  background: rgba(248, 81, 73, 0.12);
  border: 1px solid rgba(248, 81, 73, 0.35);
  color: #f85149;
}

.badge-info {
  background: rgba(56, 189, 248, 0.12);
  border: 1px solid rgba(56, 189, 248, 0.35);
  color: #38bdf8;
}

.badge-neutral {
  background: var(--bg-tertiary, #21262d);
  border: 1px solid var(--border, #30363d);
  color: var(--text-tertiary, #8b949e);
}

.spin {
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

.spinner-md {
  width: 28px;
  height: 28px;
  border: 2px solid var(--border, #30363d);
  border-top-color: var(--primary, #58a6ff);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

.spinner-sm {
  width: 18px;
  height: 18px;
  border: 2px solid var(--border, #30363d);
  border-top-color: var(--primary, #58a6ff);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}
</style>
