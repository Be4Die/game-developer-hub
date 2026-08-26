<template>
  <div class="moderation-page-container">
    <div class="main-content-wrap">
      <!-- Панель фильтрации и поиска (компактный стиль консоли) -->
      <div class="filters-toolbar">
        <!-- Поиск по названию, ID, тексту сообщения -->
        <div class="filter-field field-search">
          <label class="field-label">{{ t('common.search') }}</label>
          <div class="input-wrapper">
            <input
              type="text"
              v-model="searchQuery"
              :placeholder="t('moderation.searchChatsPlaceholder')"
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

        <!-- Фильтр по статусу диалога -->
        <div class="filter-field field-status">
          <label class="field-label">{{ t('moderation.replyStatus') }}</label>
          <div class="select-wrapper">
            <select v-model="statusFilter" class="filter-select">
              <option value="all">{{ t('moderation.allChats') }}</option>
              <option value="unanswered">{{ t('moderation.unanswered') }}</option>
              <option value="in_dialog">{{ t('moderation.inDialog') }}</option>
              <option value="resolved">{{ t('moderation.resolvedDialog') }}</option>
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
          @click="loadChats"
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

      <!-- Пустой список чатов -->
      <div
        v-else-if="chats.length === 0 && !searchQuery && statusFilter === 'all'"
        class="state-container empty-card"
      >
        <div class="empty-icon-wrap">
          <MessageSquare class="icon-lg text-muted" />
        </div>
        <h3>{{ t('moderation.noChats') }}</h3>
        <p>Когда разработчики или модераторы отправят сообщения, они появятся в этом списке.</p>
        <button class="btn-primary-sm" @click="loadChats">
          <RefreshCw class="icon-xs" />
          <span>{{ t('common.refresh') }}</span>
        </button>
      </div>

      <!-- Пустой список по результатам поиска/фильтров -->
      <div
        v-else-if="filteredChats.length === 0"
        class="state-container empty-card"
      >
        <Search class="icon-md text-muted" />
        <h3>{{ t('common.empty') }}</h3>
        <p>{{ t('stats.noData') }}</p>
        <button class="btn-reset-filters" @click="resetFilters">
          {{ t('common.reset') }}
        </button>
      </div>

      <!-- Таблица чатов проектов -->
      <div v-else class="table-wrapper">
        <table class="moderation-table">
          <thead>
            <tr>
              <th class="col-game">{{ t('moderation.projectColumn') }}</th>
              <th class="col-last-msg">{{ t('moderation.lastMessage') }}</th>
              <th class="col-reply-status">{{ t('moderation.replyStatus') }}</th>
              <th class="col-date">{{ t('common.date') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="item in paginatedChats"
              :key="item.projectId"
              class="table-row clickable-row"
              @click="openChat(item.projectId)"
            >
              <!-- 1 колонка: Игра / Проект -->
              <td class="col-game">
                <div class="game-cell">
                  <div class="game-icon-box">
                    <img
                      v-if="item.iconPath"
                      :src="getMediaUrl(item.iconPath)"
                      alt="Icon"
                      class="game-icon-img"
                    />
                    <div v-else class="game-icon-mock">
                      <span>Draft</span>
                    </div>
                  </div>
                  <div class="game-text">
                    <div class="game-type-label">
                      Проект #{{ item.projectId }}
                    </div>
                    <div class="game-title">
                      {{ item.titleRu || item.titleEn || `Проект #${item.projectId}` }}
                    </div>
                  </div>
                </div>
              </td>

              <!-- 2 колонка: Последнее сообщение -->
              <td class="col-last-msg">
                <div class="msg-preview-cell">
                  <div class="msg-meta-row">
                    <span class="sender-role-pill" :class="senderRoleClass(item)">
                      {{ senderRoleLabel(item) }}
                    </span>
                    <span class="sender-id" v-if="item.lastMessage?.sender_id && item.lastMessage?.sender_id !== 'system'">
                      {{ item.lastMessage?.sender_id }}
                    </span>
                  </div>
                  <div class="msg-content-text" :title="item.lastMessage?.content">
                    {{ item.lastMessage?.content || '—' }}
                  </div>
                </div>
              </td>

              <!-- 3 колонка: Статус ответа / диалога -->
              <td class="col-reply-status">
                <span
                  class="status-pill"
                  :class="dialogStatusClass(item.dialogState)"
                >
                  {{ dialogStatusLabel(item.dialogState) }}
                </span>
              </td>

              <!-- 4 колонка: Дата последнего сообщения -->
              <td class="col-date">
                <span class="date-text">
                  {{ formatDateTime(item.lastMessage?.created_at || item.lastMessage?.createdAt) }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Пагинация -->
      <div v-if="totalPages > 1" class="pagination-bar">
        <span class="page-info">
          {{ (currentPage - 1) * pageSize + 1 }}–{{
            Math.min(currentPage * pageSize, filteredChats.length)
          }}
          из {{ filteredChats.length }}
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
  MessageSquare,
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
} from 'lucide-vue-next';
import {
  moderationApi,
  normalizeRequest,
  formatDateTime,
} from '@/entities/moderation';
import { getMediaUrl } from '@/entities/project';
import { showToast } from '@/shared/lib';

const { t } = useI18n();
const router = useRouter();

const chats = ref([]);
const loading = ref(false);

const searchQuery = ref('');
const statusFilter = ref('all');
const sortBy = ref('newest');
const currentPage = ref(1);
const pageSize = ref(10);

function determineDialogState(lastMsg) {
  if (!lastMsg) return 'resolved';
  const role = Number(lastMsg.sender_role || lastMsg.senderRole || 1);
  const msgType = Number(lastMsg.message_type || lastMsg.messageType || 1);
  const content = (lastMsg.content || '').toLowerCase();

  // Если закрыт диалог или вынесен вердикт (одобрен/отклонен)
  if (
    msgType === 3 || // status_changed
    msgType === 4 || // approved
    msgType === 5 || // rejected
    content.includes('закрыл диалог') ||
    content.includes('вопрос решён') ||
    content.includes('одобрен') ||
    content.includes('отклонен')
  ) {
    return 'resolved';
  }

  // Если последнее сообщение от разработчика
  if (role === 1) {
    return 'unanswered';
  }

  // Если последнее сообщение от модератора
  if (role === 2) {
    return 'in_dialog';
  }

  return 'resolved';
}

async function loadChats() {
  loading.value = true;
  try {
    const [chatsRes, reqsRes] = await Promise.all([
      moderationApi.listActiveChats({ limit: 100 }),
      moderationApi.listRequests({ limit: 100 }),
    ]);

    const reqsMap = new Map();
    (reqsRes.requests || []).forEach((r) => {
      const norm = normalizeRequest(r);
      if (norm && !reqsMap.has(norm.projectId)) {
        reqsMap.set(norm.projectId, norm);
      }
    });

    const rawChats = chatsRes.chats || [];
    chats.value = rawChats.map((c) => {
      const pId = Number(c.project_id || c.projectId);
      const req = reqsMap.get(pId);
      const lastMsg = c.last_message || c.lastMessage || {};
      const dialogState = determineDialogState(lastMsg);

      return {
        projectId: pId,
        lastMessage: lastMsg,
        dialogState,
        titleRu: req?.snapshot?.titleRu || '',
        titleEn: req?.snapshot?.titleEn || '',
        iconPath: req?.snapshot?.iconPath || '',
        ownerId: req?.ownerId || '',
      };
    });
  } catch (err) {
    showToast(t('common.error'), 'danger');
  } finally {
    loading.value = false;
  }
}

onMounted(loadChats);

function resetFilters() {
  searchQuery.value = '';
  statusFilter.value = 'all';
  sortBy.value = 'newest';
  currentPage.value = 1;
}

const filteredChats = computed(() => {
  let list = [...chats.value];

  // Фильтр по статусу диалога
  if (statusFilter.value !== 'all') {
    list = list.filter((c) => c.dialogState === statusFilter.value);
  }

  // Поиск
  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase().trim();
    list = list.filter((c) => {
      const titleRu = (c.titleRu || '').toLowerCase();
      const titleEn = (c.titleEn || '').toLowerCase();
      const pId = String(c.projectId);
      const content = (c.lastMessage?.content || '').toLowerCase();
      const sender = (c.lastMessage?.sender_id || '').toLowerCase();
      return (
        titleRu.includes(q) ||
        titleEn.includes(q) ||
        pId.includes(q) ||
        content.includes(q) ||
        sender.includes(q)
      );
    });
  }

  // Сортировка
  if (sortBy.value === 'newest') {
    list.sort((a, b) => {
      const dateA = new Date(a.lastMessage?.created_at || a.lastMessage?.createdAt || 0);
      const dateB = new Date(b.lastMessage?.created_at || b.lastMessage?.createdAt || 0);
      return dateB - dateA;
    });
  } else if (sortBy.value === 'oldest') {
    list.sort((a, b) => {
      const dateA = new Date(a.lastMessage?.created_at || a.lastMessage?.createdAt || 0);
      const dateB = new Date(b.lastMessage?.created_at || b.lastMessage?.createdAt || 0);
      return dateA - dateB;
    });
  } else if (sortBy.value === 'title') {
    list.sort((a, b) => {
      const nameA = a.titleRu || a.titleEn || '';
      const nameB = b.titleRu || b.titleEn || '';
      return nameA.localeCompare(nameB);
    });
  }

  return list;
});

const totalPages = computed(() =>
  Math.max(1, Math.ceil(filteredChats.value.length / pageSize.value))
);

const paginatedChats = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value;
  return filteredChats.value.slice(start, start + pageSize.value);
});

function senderRoleLabel(item) {
  const msgType = Number(item.lastMessage?.message_type || item.lastMessage?.messageType || 1);
  if (msgType > 1) return t('moderation.systemRole');

  const r = Number(item.lastMessage?.sender_role || item.lastMessage?.senderRole || 1);
  if (r === 1) return t('moderation.developerRole');
  if (r === 2) return t('moderation.moderatorRole');
  if (r === 3) return t('moderation.systemRole');
  return '—';
}

function senderRoleClass(item) {
  const msgType = Number(item.lastMessage?.message_type || item.lastMessage?.messageType || 1);
  if (msgType > 1) return 'role-sys';

  const r = Number(item.lastMessage?.sender_role || item.lastMessage?.senderRole || 1);
  if (r === 1) return 'role-dev';
  if (r === 2) return 'role-mod';
  return 'role-sys';
}

function dialogStatusLabel(state) {
  if (state === 'unanswered') return t('moderation.unanswered');
  if (state === 'in_dialog') return t('moderation.inDialog');
  if (state === 'resolved') return t('moderation.resolvedDialog');
  return t('common.unknown');
}

function dialogStatusClass(state) {
  if (state === 'unanswered') return 'status-unanswered';
  if (state === 'in_dialog') return 'status-in-dialog';
  if (state === 'resolved') return 'status-resolved';
  return 'status-neutral';
}

function openChat(projectId) {
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
  width: 190px;
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

.col-last-msg {
  width: 42%;
}

.col-reply-status {
  width: 14%;
}

.col-date {
  width: 12%;
  text-align: right;
  padding-right: 20px;
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

.table-row td.col-date {
  text-align: right;
  padding-right: 20px;
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

.msg-preview-cell {
  display: flex;
  flex-direction: column;
  gap: 4px;
  max-width: 520px;
}

.msg-meta-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.sender-role-pill {
  font-size: 11px;
  font-weight: 600;
  padding: 1px 6px;
  border-radius: 4px;
  text-transform: uppercase;
  letter-spacing: 0.3px;
}

.role-dev {
  background: rgba(245, 176, 39, 0.15);
  color: #f5b027;
  border: 1px solid rgba(245, 176, 39, 0.3);
}

.role-mod {
  background: rgba(88, 166, 255, 0.15);
  color: #58a6ff;
  border: 1px solid rgba(88, 166, 255, 0.3);
}

.role-sys {
  background: var(--bg-tertiary, #21262d);
  color: var(--text-tertiary, #8b949e);
  border: 1px solid var(--border, #30363d);
}

.sender-id {
  font-size: 12px;
  color: var(--text-muted, #b0b8c4);
  max-width: 140px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.msg-content-text {
  font-size: 13px;
  color: var(--text-main, #f0f6fc);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  line-height: 1.3;
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

.status-unanswered {
  background: rgba(245, 176, 39, 0.12);
  border: 1px solid rgba(245, 176, 39, 0.35);
  color: #f5b027;
}

.status-in-dialog {
  background: rgba(88, 166, 255, 0.12);
  border: 1px solid rgba(88, 166, 255, 0.35);
  color: #58a6ff;
}

.status-resolved {
  background: rgba(46, 204, 113, 0.12);
  border: 1px solid rgba(46, 204, 113, 0.35);
  color: #2ecc71;
}

.status-neutral {
  background: var(--bg-tertiary, #21262d);
  border: 1px solid var(--border, #30363d);
  color: var(--text-muted, #b0b8c4);
}

.date-text {
  font-size: 13px;
  color: var(--text-muted, #b0b8c4);
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
