<template>
  <div class="catalog-page-container">
    <div class="main-content-wrap">
      <!-- Панель фильтров, поиска и переключения вида -->
      <div class="filters-toolbar">
        <!-- Поиск -->
        <div class="filter-field field-search">
          <label class="field-label">{{ t('common.search') }}</label>
          <div class="input-wrapper">
            <input
              v-model="searchQuery"
              type="text"
              :placeholder="t('catalog.searchPlaceholder')"
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

        <!-- Сброс фильтров -->
        <button
          v-if="searchQuery || sortBy !== 'newest'"
          class="btn-reset-filters"
          title="Сбросить фильтры"
          @click="resetFilters"
        >
          <RotateCcw class="icon-xs" />
          <span>{{ t('common.reset') }}</span>
        </button>

        <!-- Кнопка обновления -->
        <button class="btn-refresh" :disabled="loading" title="Обновить" @click="loadGames">
          <RefreshCw class="icon-xs" :class="{ spin: loading }" />
          <span>{{ t('common.refresh') }}</span>
        </button>
      </div>

      <!-- Состояние загрузки -->
      <div v-if="loading" class="state-container loading-card">
        <div class="spinner-md"></div>
        <p>{{ t('common.loading') }}</p>
      </div>

      <!-- Пустой каталог -->
      <div
        v-else-if="games.length === 0 && !searchQuery"
        class="state-container empty-card"
      >
        <div class="empty-icon-wrap">
          <Gamepad2 class="icon-lg text-muted" />
        </div>
        <h3>{{ t('catalog.emptyCatalog') }}</h3>
        <p>{{ t('catalog.emptyCatalogDesc') }}</p>
        <button class="btn-primary-sm" @click="loadGames">
          <RefreshCw class="icon-xs" />
          <span>{{ t('common.refresh') }}</span>
        </button>
      </div>

      <!-- Пустой результат поиска -->
      <div v-else-if="filteredGames.length === 0" class="state-container empty-card">
        <Search class="icon-md text-muted" />
        <h3>{{ t('common.empty') }}</h3>
        <p>{{ t('catalog.emptySearch') }}</p>
        <button class="btn-reset-filters" @click="resetFilters">
          {{ t('common.reset') }}
        </button>
      </div>

      <!-- ВИД: ТАБЛИЦА (TABLE) -->
      <div v-else class="table-wrapper">
        <table class="catalog-table">
          <thead>
            <tr>
              <th class="col-game">{{ t('moderation.projectColumn') }}</th>
              <th class="col-version">{{ t('catalog.version') }}</th>
              <th class="col-dev">{{ t('catalog.developer') }}</th>
              <th class="col-date">{{ t('catalog.publishedAt') }}</th>
              <th class="col-actions"></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="game in filteredGames" :key="game.id" class="table-row">
              <!-- Проект / Игра -->
              <td class="col-game">
                <div class="game-cell">
                  <div class="game-icon-box">
                    <img
                      v-if="getGameIcon(game)"
                      :src="getMediaUrl(getGameIcon(game))"
                      alt="Icon"
                      class="game-icon-img"
                    />
                    <div v-else class="game-icon-mock">
                      <Gamepad2 class="icon-xs text-muted" />
                    </div>
                  </div>
                  <div class="game-text">
                    <div class="game-title" :title="getGameTitle(game)">{{ getGameTitle(game) }}</div>
                  </div>
                </div>
              </td>

              <!-- Версия -->
              <td class="col-version">
                <span class="version-badge">v{{ getGameVersion(game) }}</span>
              </td>

              <!-- Студия / Разработчик -->
              <td class="col-dev">
                <div class="dev-cell" :title="game.owner_id">
                  <User class="icon-xs text-muted" />
                  <span>{{ game.owner_id ? getUserDisplayName(game.owner_id) : '—' }}</span>
                </div>
              </td>

              <!-- Дата публикации -->
              <td class="col-date">
                <span class="date-text">{{ formatDateTime(getPublishedAt(game)) }}</span>
              </td>

              <!-- Действия -->
              <td class="col-actions">
                <div class="row-actions">
                  <a
                    v-if="getProdUrl(game)"
                    :href="getProdUrl(game)"
                    target="_blank"
                    rel="noopener"
                    class="btn-icon-action"
                    :title="t('catalog.playProd')"
                  >
                    <Play class="icon-xs" />
                  </a>
                  <button
                    class="btn-icon-action"
                    :title="t('catalog.inspect')"
                    @click="openProject(game.id)"
                  >
                    <Eye class="icon-xs" />
                  </button>
                  <button
                    class="btn-action-revoke-sm"
                    :title="t('catalog.revokeBtn')"
                    @click="openRevokeModal(game)"
                  >
                    <ShieldAlert class="icon-xs" />
                    <span>{{ t('catalog.revokeBtn') }}</span>
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- МОДАЛЬНОЕ ОКНО ЭКСТРЕННОГО ОТЗЫВА (KILLSWITCH) -->
      <div v-if="showRevokeModal" class="modal-overlay" @click.self="closeRevokeModal">
        <div class="modal-card revoke-modal-card">
          <div class="modal-header">
            <div class="modal-header-title">
              <div class="warning-icon-badge">
                <ShieldAlert class="icon-md text-danger" />
              </div>
              <div>
                <h3 class="modal-title">{{ t('catalog.revokeModal.title') }}</h3>
                <span class="modal-subtitle">
                  {{ targetGame ? getGameTitle(targetGame) : '' }} (ID #{{ targetGame?.id }})
                </span>
              </div>
            </div>
            <button class="btn-close-modal" @click="closeRevokeModal">
              <X class="icon-sm" />
            </button>
          </div>

          <div class="modal-body">
            <!-- Предупреждающий баннер -->
            <div class="alert-danger-banner">
              <AlertTriangle class="icon-sm flex-shrink-0" />
              <span>{{ t('catalog.revokeModal.warning') }}</span>
            </div>

            <!-- Поле ввода причины -->
            <div class="form-field">
              <label class="form-label">
                {{ t('catalog.revokeModal.reasonLabel') }}
                <span class="required-star">*</span>
              </label>
              <textarea
                v-model="revokeReason"
                rows="4"
                :placeholder="t('catalog.revokeModal.reasonPlaceholder')"
                class="form-textarea"
                :class="{ 'has-error': reasonError }"
              ></textarea>
              <span v-if="reasonError" class="field-error-text">{{ reasonError }}</span>
            </div>

            <!-- Чекбокс бана разработчика (только для Администратора) -->
            <div v-if="isAdmin" class="form-checkbox-row">
              <label class="checkbox-label">
                <input v-model="banDeveloper" type="checkbox" class="checkbox-input" />
                <span class="checkbox-text">{{ t('catalog.revokeModal.banDevCheckbox') }}</span>
              </label>
            </div>
          </div>

          <div class="modal-footer">
            <button class="btn-cancel" :disabled="revoking" @click="closeRevokeModal">
              {{ t('catalog.revokeModal.cancel') }}
            </button>
            <button
              class="btn-danger-confirm"
              :disabled="revoking"
              @click="confirmRevoke"
            >
              <RefreshCw v-if="revoking" class="icon-xs spin" />
              <ShieldAlert v-else class="icon-xs" />
              <span>
                {{
                  revoking
                    ? t('catalog.revokeModal.revoking')
                    : t('catalog.revokeModal.confirmRevoke')
                }}
              </span>
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
  RotateCcw,
  RefreshCw,
  Play,
  Eye,
  ShieldAlert,
  AlertTriangle,
  ChevronDown,
  User,
  Gamepad2,
} from 'lucide-vue-next';
import {
  listPublishedProjects,
  unpublish,
  getMediaUrl,
} from '@/entities/project';
import { moderationApi } from '@/entities/moderation';
import { useAuth, setUserStatus, getUserDisplayName } from '@/entities/user';
import { showToast } from '@/shared/lib';

const { t } = useI18n();
const router = useRouter();
const { state: authState } = useAuth();

const games = ref([]);
const loading = ref(true);

const searchQuery = ref('');
const sortBy = ref('newest');

// Модальное окно отзыва
const showRevokeModal = ref(false);
const targetGame = ref(null);
const revokeReason = ref('');
const reasonError = ref('');
const banDeveloper = ref(false);
const revoking = ref(false);

const isAdmin = computed(() => {
  const r = authState.user?.role;
  return r === 'USER_ROLE_ADMIN' || r === 'admin' || r === 3;
});

async function loadGames() {
  loading.value = true;
  try {
    const res = await listPublishedProjects({ limit: 100, offset: 0 });
    games.value = res.projects || [];
  } catch (err) {
    showToast(t('common.error'), 'danger');
  } finally {
    loading.value = false;
  }
}

onMounted(loadGames);

function resetFilters() {
  searchQuery.value = '';
  sortBy.value = 'newest';
}

function getGameTitle(p) {
  return (
    p.release?.title_ru ||
    p.release?.title_en ||
    p.draft?.title_ru ||
    p.draft?.title_en ||
    p.title_ru ||
    p.title_en ||
    `Проект #${p.id}`
  );
}

function getGameVersion(p) {
  return (
    p.release?.version ||
    p.draft?.active_build_version ||
    p.active_build_version ||
    '1.0.0'
  );
}

function getGameIcon(p) {
  return p.release?.icon_path || p.draft?.icon_path || p.icon_path || '';
}

function getGameCover(p) {
  return p.release?.cover_path || p.draft?.cover_path || p.cover_path || '';
}

function getProdUrl(p) {
  return p.release?.prod_url || p.prod_url || '';
}

function getPublishedAt(p) {
  return p.release?.published_at || p.updated_at || p.created_at;
}

function formatDateTime(val) {
  if (!val) return '—';
  try {
    const d = new Date(val);
    if (isNaN(d.getTime())) return String(val);
    return d.toLocaleString('ru-RU', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  } catch {
    return String(val);
  }
}

const filteredGames = computed(() => {
  let list = [...games.value];

  // Поиск
  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase().trim();
    list = list.filter((g) => {
      const title = getGameTitle(g).toLowerCase();
      const owner = (g.owner_id || '').toLowerCase();
      const pId = String(g.id);
      const ver = getGameVersion(g).toLowerCase();
      return (
        title.includes(q) ||
        owner.includes(q) ||
        pId.includes(q) ||
        ver.includes(q)
      );
    });
  }

  // Сортировка
  if (sortBy.value === 'newest') {
    list.sort(
      (a, b) => new Date(getPublishedAt(b) || 0) - new Date(getPublishedAt(a) || 0)
    );
  } else if (sortBy.value === 'oldest') {
    list.sort(
      (a, b) => new Date(getPublishedAt(a) || 0) - new Date(getPublishedAt(b) || 0)
    );
  } else if (sortBy.value === 'title') {
    list.sort((a, b) => getGameTitle(a).localeCompare(getGameTitle(b)));
  }

  return list;
});

function openProject(projectId) {
  router.push(`/moderator/projects/${projectId}`);
}

function openRevokeModal(game) {
  targetGame.value = game;
  revokeReason.value = '';
  reasonError.value = '';
  banDeveloper.value = false;
  showRevokeModal.value = true;
}

function closeRevokeModal() {
  if (revoking.value) return;
  showRevokeModal.value = false;
  targetGame.value = null;
}

async function confirmRevoke() {
  if (!revokeReason.value.trim()) {
    reasonError.value = t('catalog.revokeModal.reasonRequired');
    return;
  }
  reasonError.value = '';
  revoking.value = true;

  const game = targetGame.value;
  try {
    // 1. Физическое снятие игры с публикации (UndeployProd + Deactivate release + Status draft)
    await unpublish(game.id);

    // 2. Отправка системного сообщения о снятии в чат проекта
    try {
      await moderationApi.sendMessage(
        game.id,
        `[ЭКСТРЕННЫЙ ОТЗЫВ ИЗ КАТАЛОГА] Игра снята с публикации. Причина: ${revokeReason.value.trim()}`
      );
    } catch (e) {
      console.warn('Failed to send takedown notice to chat:', e);
    }

    // 3. Если выбран бан разработчика (для админа)
    if (banDeveloper.value && game.owner_id) {
      try {
        await setUserStatus(game.owner_id, 'USER_STATUS_SUSPENDED');
        showToast(`Аккаунт разработчика ${game.owner_id} заблокирован`, 'warning');
      } catch (e) {
        console.warn('Failed to suspend developer:', e);
      }
    }

    showToast(
      t('catalog.revokeModal.successRevoke', { title: getGameTitle(game) }),
      'success'
    );
    showRevokeModal.value = false;
    targetGame.value = null;
    await loadGames();
  } catch (err) {
    showToast(
      err.response?.data?.message || 'Не удалось отозвать игру из каталога',
      'danger'
    );
  } finally {
    revoking.value = false;
  }
}
</script>

<style scoped>
.catalog-page-container {
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

.field-search {
  flex: 1;
  min-width: 240px;
}

.field-sort {
  width: 220px;
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

.view-toggle-group {
  display: flex;
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  overflow: hidden;
  height: 36px;
  background: var(--bg-secondary, #161b22);
}

.btn-view-toggle {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 100%;
  background: transparent;
  border: none;
  color: var(--text-tertiary, #8b949e);
  cursor: pointer;
  transition: all 0.15s;
}

.btn-view-toggle.active {
  background: var(--bg-tertiary, #21262d);
  color: var(--primary, #58a6ff);
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

.spin {
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.version-badge {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
  background: rgba(13, 17, 23, 0.85);
  border: 1px solid var(--border, #30363d);
  color: var(--primary, #58a6ff);
}

.game-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
  margin: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* ТАБЛИЦА */
.table-wrapper {
  width: 100%;
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-md, 8px);
  overflow-x: auto;
}

.catalog-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
}

.catalog-table thead {
  background: var(--bg-card, #161b22);
  border-bottom: 1px solid var(--border, #30363d);
}

.catalog-table th {
  padding: 12px 16px;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-tertiary, #8b949e);
}

.table-row {
  border-bottom: 1px solid var(--border, #21262d);
  transition: background-color 0.15s;
}

.table-row:hover {
  background: var(--bg-secondary, #161b22);
}

.table-row td {
  padding: 12px 16px;
  vertical-align: middle;
}

.col-game {
  width: 40%;
}

.col-version {
  width: 12%;
}

.col-dev {
  width: 18%;
}

.col-date {
  width: 15%;
}

.col-actions {
  width: 15%;
  text-align: right;
}

.game-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.game-icon-box {
  width: 36px;
  height: 36px;
  border-radius: 6px;
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
  font-size: 11px;
  color: var(--text-tertiary, #8b949e);
}

.dev-cell {
  display: inline-flex;
  align-items: center;
  gap: 6px;
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

.btn-icon-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 6px;
  background: var(--bg-secondary, #21262d);
  border: 1px solid var(--border, #30363d);
  color: var(--text-main, #f0f6fc);
  cursor: pointer;
  text-decoration: none;
  box-sizing: border-box;
  flex-shrink: 0;
  transition: all 0.15s;
}

.btn-icon-action:hover {
  border-color: var(--primary, #58a6ff);
  color: var(--primary, #58a6ff);
}

.btn-action-revoke-sm {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 32px;
  padding: 0 12px;
  background: rgba(248, 81, 73, 0.1);
  border: 1px solid rgba(248, 81, 73, 0.3);
  color: #f85149;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  box-sizing: border-box;
  flex-shrink: 0;
  transition: all 0.15s;
}

.btn-action-revoke-sm:hover {
  background: rgba(248, 81, 73, 0.2);
  border-color: #f85149;
}

/* СОСТОЯНИЯ */
.state-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  gap: 12px;
  color: var(--text-muted, #b0b8c4);
}

.empty-card,
.loading-card {
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-md, 8px);
}

.loading-card {
  min-height: 280px;
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
  margin-top: 12px;
}

.spinner-md {
  width: 32px;
  height: 32px;
  border: 3px solid var(--border, #30363d);
  border-top-color: var(--primary, #58a6ff);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

/* МОДАЛЬНОЕ ОКНО */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.75);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 999;
  padding: 20px;
}

.revoke-modal-card {
  width: 100%;
  max-width: 520px;
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-md, 10px);
  display: flex;
  flex-direction: column;
  box-shadow: 0 16px 40px rgba(0, 0, 0, 0.5);
  overflow: hidden;
}

.modal-header {
  padding: 16px 20px;
  border-bottom: 1px solid var(--border, #21262d);
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.modal-header-title {
  display: flex;
  align-items: center;
  gap: 12px;
}

.warning-icon-badge {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  background: rgba(248, 81, 73, 0.12);
  border: 1px solid rgba(248, 81, 73, 0.3);
  display: flex;
  align-items: center;
  justify-content: center;
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

.modal-body {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.alert-danger-banner {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 12px 14px;
  background: rgba(248, 81, 73, 0.1);
  border: 1px solid rgba(248, 81, 73, 0.3);
  border-radius: 6px;
  color: #f85149;
  font-size: 13px;
  line-height: 1.4;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-main, #f0f6fc);
}

.required-star {
  color: #f85149;
}

.form-textarea {
  width: 100%;
  padding: 10px 12px;
  background: var(--bg-secondary, #0d1117);
  border: 1px solid var(--border, #30363d);
  border-radius: 6px;
  color: var(--text-main, #f0f6fc);
  font-size: 13px;
  outline: none;
  box-sizing: border-box;
  resize: vertical;
  font-family: inherit;
}

.form-textarea:focus {
  border-color: #f85149;
}

.form-textarea.has-error {
  border-color: #f85149;
}

.field-error-text {
  font-size: 12px;
  color: #f85149;
}

.form-checkbox-row {
  display: flex;
  align-items: center;
  padding: 10px 12px;
  background: var(--bg-secondary, #0d1117);
  border: 1px solid var(--border, #30363d);
  border-radius: 6px;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  user-select: none;
}

.checkbox-input {
  width: 16px;
  height: 16px;
  accent-color: #f85149;
  cursor: pointer;
}

.checkbox-text {
  font-size: 13px;
  color: var(--text-main, #f0f6fc);
  font-weight: 500;
}

.modal-footer {
  padding: 14px 20px;
  border-top: 1px solid var(--border, #21262d);
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
}

.btn-cancel {
  height: 34px;
  padding: 0 16px;
  background: transparent;
  border: 1px solid var(--border, #30363d);
  border-radius: 6px;
  color: var(--text-muted, #b0b8c4);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.15s;
}

.btn-cancel:hover {
  background: var(--bg-tertiary, #21262d);
  color: var(--text-main, #f0f6fc);
}

.btn-danger-confirm {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 34px;
  padding: 0 16px;
  background: #da3633;
  color: #ffffff;
  border: 1px solid rgba(248, 81, 73, 0.4);
  border-radius: 6px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s;
}

.btn-danger-confirm:hover:not(:disabled) {
  background: #b62324;
}

.btn-danger-confirm:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

@media (max-width: 768px) {
  .catalog-page-container {
    padding: 16px;
  }
  .filters-toolbar {
    flex-wrap: wrap;
  }
  .field-search {
    width: 100%;
    min-width: 100%;
  }
  .catalog-grid {
    grid-template-columns: 1fr;
  }
}
</style>
