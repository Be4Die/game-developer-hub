<template>
  <div class="moderation-page-container">
    <div class="main-content-wrap">
      <!-- Панель фильтрации и поиска (консольный стиль) -->
      <div class="filters-toolbar">
        <!-- Поиск по названию, ID, разработчику, причине -->
        <div class="filter-field field-search">
          <label class="field-label">{{ t('common.search') }}</label>
          <div class="input-wrapper">
            <input
              v-model="searchQuery"
              type="text"
              :placeholder="t('journal.searchPlaceholder')"
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
          title="Сбросить фильтры"
          @click="resetFilters"
        >
          <RotateCcw class="icon-xs" />
          <span>{{ t('common.reset') }}</span>
        </button>

        <!-- Кнопка обновления -->
        <button class="btn-refresh" :disabled="loading" title="Обновить" @click="loadArchive">
          <RefreshCw class="icon-xs" :class="{ spin: loading }" />
          <span>{{ t('common.refresh') }}</span>
        </button>
      </div>

      <!-- Состояние загрузки -->
      <div v-if="loading" class="state-container">
        <div class="spinner-md"></div>
        <p>{{ t('common.loading') }}</p>
      </div>

      <!-- Пустой журнал -->
      <div
        v-else-if="requests.length === 0 && !searchQuery && statusFilter === 'all'"
        class="state-container empty-card"
      >
        <div class="empty-icon-wrap">
          <BookOpen class="icon-lg text-muted" />
        </div>
        <h3>{{ t('journal.emptyJournal') }}</h3>
        <p>{{ t('journal.emptyJournalDesc') }}</p>
        <button class="btn-primary-sm" @click="loadArchive">
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

      <!-- Таблица журнала решений -->
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
              @click="openSnapshot(req)"
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
                      <span>#{{ req.projectId }}</span>
                    </div>
                  </div>
                  <div class="game-text">
                    <div class="game-type-label">
                      Заявка #{{ req.id }} • Проект #{{ req.projectId }}
                    </div>
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

              <!-- 4 колонка: Вердикт / Статус -->
              <td class="col-verdict">
                <span class="status-pill" :class="verdictClass(req.status)">
                  {{ verdictLabel(req.status) }}
                </span>
              </td>

              <!-- 5 колонка: Модератор -->
              <td class="col-mod">
                <span v-if="req.moderatorId" class="mod-name">
                  {{ req.moderatorId }}
                </span>
                <span v-else class="unassigned-text">
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
                    class="btn-snapshot-action"
                    title="Просмотреть снимок решения"
                    @click="openSnapshot(req)"
                  >
                    <Camera class="icon-xs" />
                    <span>Снимок</span>
                  </button>
                  <button
                    class="btn-inspect-sm"
                    :title="t('moderation.viewDetails')"
                    @click="openProject(req.projectId)"
                  >
                    <ExternalLink class="icon-xs" />
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

      <!-- МОДАЛЬНОЕ ОКНО СНАПШОТА РЕШЕНИЯ (SNAPSHOT AUDIT MODAL) -->
      <div v-if="showSnapshotModal" class="modal-overlay" @click.self="closeSnapshotModal">
        <div class="modal-card snapshot-modal-card">
          <!-- Шапка модального окна -->
          <div class="modal-header">
            <div class="modal-header-title">
              <div class="snapshot-icon-badge">
                <Camera class="icon-md text-primary" />
              </div>
              <div>
                <div class="snapshot-title-row">
                  <h3 class="modal-title">{{ t('journal.snapshotModal.title') }}</h3>
                  <span
                    v-if="selectedRequest"
                    class="status-pill-sm"
                    :class="verdictClass(selectedRequest.status)"
                  >
                    {{ verdictLabel(selectedRequest.status) }}
                  </span>
                </div>
                <span class="modal-subtitle">
                  Заявка #{{ selectedRequest?.id }} • Проект #{{ selectedRequest?.projectId }}
                </span>
              </div>
            </div>
            <button class="btn-close-modal" @click="closeSnapshotModal">
              <X class="icon-sm" />
            </button>
          </div>

          <!-- Тело модального окна снапшота -->
          <div v-if="selectedRequest" class="modal-body-scrollable">
            <!-- Плашка замороженного состояния -->
            <div class="frozen-banner">
              <History class="icon-sm text-primary flex-shrink-0" />
              <span>{{ t('journal.snapshotModal.frozenNotice') }}</span>
            </div>

            <!-- Сетка содержимого снапшота -->
            <div class="snapshot-grid">
              <!-- Левая колонка: Проект и сборка снапшота -->
              <div class="snapshot-col snapshot-project-info">
                <div class="snapshot-card">
                  <!-- Обложка снапшота -->
                  <div class="snapshot-cover-wrap">
                    <img
                      v-if="selectedRequest.snapshot.coverPath"
                      :src="getMediaUrl(selectedRequest.snapshot.coverPath)"
                      alt="Cover"
                      class="snapshot-cover-img"
                    />
                    <div v-else class="snapshot-cover-placeholder">
                      <Gamepad2 class="icon-lg text-muted" />
                    </div>
                  </div>

                  <div class="snapshot-card-content">
                    <div class="snapshot-game-identity">
                      <div class="snapshot-icon-box">
                        <img
                          v-if="selectedRequest.snapshot.iconPath"
                          :src="getMediaUrl(selectedRequest.snapshot.iconPath)"
                          alt="Icon"
                          class="snapshot-icon-img"
                        />
                        <div v-else class="snapshot-icon-placeholder">
                          <span>#{{ selectedRequest.projectId }}</span>
                        </div>
                      </div>
                      <div class="identity-text">
                        <h4 class="snapshot-game-title">
                          {{
                            selectedRequest.snapshot.titleRu ||
                            selectedRequest.snapshot.titleEn ||
                            `Проект #${selectedRequest.projectId}`
                          }}
                        </h4>
                        <span class="snapshot-version-tag">
                          v{{ selectedRequest.snapshot.activeBuildVersion || '1.0.0' }}
                        </span>
                      </div>
                    </div>

                    <!-- Описание снапшота -->
                    <div class="snapshot-section-group">
                      <label class="snapshot-label">Описание (RU / EN):</label>
                      <p class="snapshot-description-text">
                        {{
                          selectedRequest.snapshot.aboutRu ||
                          selectedRequest.snapshot.aboutEn ||
                          'Описание отсутствует в данном снапшоте'
                        }}
                      </p>
                    </div>

                    <!-- Ссылка на Dev-билд снапшота -->
                    <div v-if="selectedRequest.snapshot.devUrl" class="snapshot-section-group">
                      <a
                        :href="selectedRequest.snapshot.devUrl"
                        target="_blank"
                        rel="noopener"
                        class="btn-dev-build"
                      >
                        <Play class="icon-xs" />
                        <span>{{ t('journal.snapshotModal.openDevBuild') }}</span>
                      </a>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Правая колонка: Вердикт и срез переписки -->
              <div class="snapshot-col snapshot-audit-info">
                <!-- Блок вердикта -->
                <div class="snapshot-card verdict-card">
                  <h4 class="section-title">
                    <CheckSquare class="icon-xs text-primary" />
                    {{ t('journal.snapshotModal.verdictInfo') }}
                  </h4>

                  <div class="verdict-meta-grid">
                    <div class="verdict-meta-item">
                      <span class="meta-label">{{ t('journal.snapshotModal.moderator') }}:</span>
                      <span class="meta-value font-medium">
                        {{ selectedRequest.moderatorId || t('moderation.notAssigned') }}
                      </span>
                    </div>

                    <div class="verdict-meta-item">
                      <span class="meta-label">{{ t('journal.snapshotModal.date') }}:</span>
                      <span class="meta-value">
                        {{
                          formatDateTime(selectedRequest.reviewedAt || selectedRequest.submittedAt)
                        }}
                      </span>
                    </div>
                  </div>

                  <!-- Официальная причина / Замечания -->
                  <div class="verdict-reason-box">
                    <label class="reason-label">
                      {{ t('journal.snapshotModal.reasonOrComment') }}:
                    </label>
                    <p class="reason-content" :class="{ 'text-danger': isRejected(selectedRequest.status) }">
                      {{ selectedRequest.rejectionReason || t('journal.snapshotModal.noReason') }}
                    </p>
                  </div>
                </div>

                <!-- Блок среза переписки (Ticket Chat Transcript) -->
                <div class="snapshot-card chat-transcript-card">
                  <h4 class="section-title">
                    <MessageSquare class="icon-xs text-primary" />
                    {{ t('journal.snapshotModal.chatTranscript') }}
                  </h4>

                  <div v-if="loadingChat" class="chat-loading-wrap">
                    <div class="spinner-sm"></div>
                    <span>Загрузка истории тикета...</span>
                  </div>

                  <div v-else-if="chatMessages.length === 0" class="chat-empty-wrap">
                    <MessageSquare class="icon-md text-muted" />
                    <span>{{ t('journal.snapshotModal.noChatMessages') }}</span>
                  </div>

                  <div v-else class="chat-messages-container">
                    <div
                      v-for="msg in chatMessages"
                      :key="msg.id"
                      class="chat-bubble-item"
                      :class="messageRoleClass(msg.senderRole)"
                    >
                      <div class="bubble-header">
                        <span class="sender-tag">{{ formatSenderRole(msg.senderRole) }}</span>
                        <span class="bubble-time">{{ formatDateTime(msg.createdAt) }}</span>
                      </div>
                      <div class="bubble-content">
                        {{ msg.content }}
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Подвал модального окна -->
          <div class="modal-footer">
            <button
              v-if="selectedRequest"
              class="btn-live-project"
              @click="openProject(selectedRequest.projectId)"
            >
              <ExternalLink class="icon-xs" />
              <span>{{ t('journal.snapshotModal.openLiveProject') }}</span>
            </button>
            <button class="btn-primary-sm" @click="closeSnapshotModal">
              {{ t('journal.snapshotModal.close') }}
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
import {
  Search,
  X,
  ChevronDown,
  RotateCcw,
  RefreshCw,
  BookOpen,
  User,
  ExternalLink,
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
  Camera,
  History,
  Gamepad2,
  Play,
  CheckSquare,
  MessageSquare,
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

// Модальное окно снапшота
const showSnapshotModal = ref(false);
const selectedRequest = ref(null);
const chatMessages = ref([]);
const loadingChat = ref(false);

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
      (a, b) => new Date(b.reviewedAt || b.submittedAt) - new Date(a.reviewedAt || a.submittedAt)
    );
  } else if (sortBy.value === 'oldest') {
    list.sort(
      (a, b) => new Date(a.reviewedAt || a.submittedAt) - new Date(b.reviewedAt || b.submittedAt)
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
  if (isApproved(status)) return t('journal.snapshotModal.approved');
  if (isRejected(status)) return t('journal.snapshotModal.rejected');
  if (isCancelled(status)) return t('journal.snapshotModal.cancelled');
  return t('common.unknown');
}

function verdictClass(status) {
  if (isApproved(status)) return 'status-approved';
  if (isRejected(status)) return 'status-rejected';
  if (isCancelled(status)) return 'status-neutral';
  return 'status-neutral';
}

function openProject(projectId) {
  closeSnapshotModal();
  router.push(`/moderator/projects/${projectId}`);
}

async function openSnapshot(req) {
  selectedRequest.value = req;
  showSnapshotModal.value = true;
  chatMessages.value = [];
  loadingChat.value = true;
  try {
    const res = await moderationApi.listMessages(req.projectId, { limit: 100 });
    chatMessages.value = res.messages || [];
  } catch (e) {
    console.warn('Failed to load chat history for snapshot:', e);
  } finally {
    loadingChat.value = false;
  }
}

function closeSnapshotModal() {
  showSnapshotModal.value = false;
  selectedRequest.value = null;
  chatMessages.value = [];
}

function messageRoleClass(role) {
  const r = String(role).toUpperCase();
  if (r.includes('MODERATOR') || role === 2) return 'bubble-moderator';
  if (r.includes('DEVELOPER') || role === 1) return 'bubble-developer';
  return 'bubble-system';
}

function formatSenderRole(role) {
  const r = String(role).toUpperCase();
  if (r.includes('MODERATOR') || role === 2) return t('moderation.moderatorRole');
  if (r.includes('DEVELOPER') || role === 1) return t('moderation.developerRole');
  return t('moderation.systemRole');
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
  box-sizing: border-box;
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
  width: 10%;
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
  font-size: 11px;
  font-weight: 600;
  color: var(--text-tertiary, #8b949e);
}

.game-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.game-type-label {
  font-size: 12px;
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

.status-pill-sm {
  display: inline-flex;
  align-items: center;
  padding: 1px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
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
  gap: 6px;
}

.btn-snapshot-action {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  height: 28px;
  padding: 0 10px;
  background: rgba(88, 166, 255, 0.1);
  border: 1px solid rgba(88, 166, 255, 0.3);
  color: #58a6ff;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
}

.btn-snapshot-action:hover {
  background: rgba(88, 166, 255, 0.2);
}

.btn-inspect-sm {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  background: var(--bg-secondary, #21262d);
  border: 1px solid var(--border, #30363d);
  color: var(--text-main, #f0f6fc);
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s;
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
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
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
  border-radius: 6px;
  color: var(--text-main, #f0f6fc);
  cursor: pointer;
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
  border-radius: 6px;
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

.spinner-md {
  width: 32px;
  height: 32px;
  border: 3px solid var(--border, #30363d);
  border-top-color: var(--primary, #58a6ff);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

.spinner-sm {
  width: 20px;
  height: 20px;
  border: 2px solid var(--border, #30363d);
  border-top-color: var(--primary, #58a6ff);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

/* МОДАЛЬНОЕ ОКНО СНАПШОТА */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.8);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 20px;
}

.snapshot-modal-card {
  width: 100%;
  max-width: 900px;
  max-height: 90vh;
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: 10px;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 50px rgba(0, 0, 0, 0.6);
  overflow: hidden;
}

.modal-header {
  padding: 16px 20px;
  border-bottom: 1px solid var(--border, #21262d);
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
}

.modal-header-title {
  display: flex;
  align-items: center;
  gap: 12px;
}

.snapshot-icon-badge {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  background: rgba(88, 166, 255, 0.12);
  border: 1px solid rgba(88, 166, 255, 0.3);
  display: flex;
  align-items: center;
  justify-content: center;
}

.snapshot-title-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.modal-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
  margin: 0;
}

.modal-subtitle {
  font-size: 12px;
  color: var(--text-tertiary, #8b949e);
}

.btn-close-modal {
  background: transparent;
  border: none;
  color: var(--text-tertiary, #8b949e);
  cursor: pointer;
  padding: 4px;
}

.modal-body-scrollable {
  padding: 20px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.frozen-banner {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  background: rgba(88, 166, 255, 0.08);
  border: 1px solid rgba(88, 166, 255, 0.25);
  border-radius: 6px;
  color: var(--primary, #58a6ff);
  font-size: 13px;
}

.snapshot-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.snapshot-card {
  background: var(--bg-secondary, #0d1117);
  border: 1px solid var(--border, #30363d);
  border-radius: 8px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.snapshot-cover-wrap {
  width: 100%;
  height: 120px;
  background: var(--bg-tertiary, #21262d);
  position: relative;
}

.snapshot-cover-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.snapshot-cover-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.snapshot-card-content {
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.snapshot-game-identity {
  display: flex;
  align-items: center;
  gap: 12px;
}

.snapshot-icon-box {
  width: 44px;
  height: 44px;
  border-radius: 8px;
  background: var(--bg-tertiary, #21262d);
  border: 1px solid var(--border, #30363d);
  overflow: hidden;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.snapshot-icon-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.snapshot-icon-placeholder {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-tertiary, #8b949e);
}

.identity-text {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.snapshot-game-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
  margin: 0;
}

.snapshot-version-tag {
  display: inline-block;
  font-size: 11px;
  font-weight: 600;
  color: var(--primary, #58a6ff);
}

.snapshot-section-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.snapshot-label {
  font-size: 11px;
  font-weight: 500;
  color: var(--text-tertiary, #8b949e);
}

.snapshot-description-text {
  font-size: 13px;
  line-height: 1.4;
  color: var(--text-muted, #b0b8c4);
  margin: 0;
  max-height: 100px;
  overflow-y: auto;
}

.btn-dev-build {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  height: 32px;
  padding: 0 14px;
  background: var(--bg-tertiary, #21262d);
  border: 1px solid var(--border, #30363d);
  color: var(--text-main, #f0f6fc);
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  text-decoration: none;
  transition: all 0.15s;
}

.btn-dev-build:hover {
  border-color: var(--primary, #58a6ff);
  color: var(--primary, #58a6ff);
}

.snapshot-col {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
  margin: 0 0 12px 0;
}

.verdict-card {
  padding: 16px;
}

.verdict-meta-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
  font-size: 13px;
  margin-bottom: 12px;
}

.verdict-meta-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.meta-label {
  font-size: 11px;
  color: var(--text-tertiary, #8b949e);
}

.meta-value {
  color: var(--text-main, #f0f6fc);
}

.verdict-reason-box {
  background: var(--bg-tertiary, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: 6px;
  padding: 10px 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.reason-label {
  font-size: 11px;
  font-weight: 500;
  color: var(--text-tertiary, #8b949e);
}

.reason-content {
  font-size: 13px;
  line-height: 1.4;
  color: var(--text-muted, #b0b8c4);
  margin: 0;
}

.chat-transcript-card {
  padding: 16px;
  flex: 1;
  min-height: 220px;
  display: flex;
  flex-direction: column;
}

.chat-loading-wrap,
.chat-empty-wrap {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 30px;
  gap: 8px;
  color: var(--text-tertiary, #8b949e);
  font-size: 13px;
  flex: 1;
}

.chat-messages-container {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 260px;
  overflow-y: auto;
  padding-right: 4px;
}

.chat-bubble-item {
  padding: 8px 12px;
  border-radius: 6px;
  font-size: 13px;
  line-height: 1.4;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.bubble-developer {
  background: rgba(88, 166, 255, 0.08);
  border: 1px solid rgba(88, 166, 255, 0.2);
}

.bubble-moderator {
  background: rgba(46, 204, 113, 0.08);
  border: 1px solid rgba(46, 204, 113, 0.2);
}

.bubble-system {
  background: var(--bg-tertiary, #21262d);
  border: 1px solid var(--border, #30363d);
}

.bubble-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 11px;
}

.sender-tag {
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
}

.bubble-time {
  color: var(--text-tertiary, #8b949e);
}

.bubble-content {
  color: var(--text-muted, #b0b8c4);
}

.modal-footer {
  padding: 14px 20px;
  border-top: 1px solid var(--border, #21262d);
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
}

.btn-live-project {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 34px;
  padding: 0 14px;
  background: var(--bg-secondary, #21262d);
  border: 1px solid var(--border, #30363d);
  color: var(--text-main, #f0f6fc);
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
}

.btn-live-project:hover {
  border-color: var(--primary, #58a6ff);
  color: var(--primary, #58a6ff);
}

@media (max-width: 800px) {
  .snapshot-grid {
    grid-template-columns: 1fr;
  }
}
</style>
