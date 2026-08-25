<template>
  <div class="projects-page-container">
    <div class="main-content-wrap">
      <!-- Панель фильтров и поиска в стиле Яндекс.Консоли -->
      <ProjectFilters
        v-model:searchQuery="searchQuery"
        v-model:statusFilter="statusFilter"
        v-model:sortBy="sortBy"
        :creating="creating"
        @reset="resetFilters"
        @create="createNewGame"
      />

      <!-- Загрузка -->
      <div v-if="loading" class="state-container">
        <div class="spinner-md"></div>
        <p>{{ t('common.loading') }}</p>
      </div>

      <!-- Пустой список без проектов -->
      <div
        v-else-if="
          games.length === 0 && !searchQuery && statusFilter === 'all'
        "
        class="state-container empty-card"
      >
        <div class="empty-icon-wrap">
          <Gamepad2 class="icon-lg" />
        </div>
        <h3>{{ t('projects.noProjects') }}</h3>
        <p>
          {{ t('projects.noProjectsDesc') }}
        </p>
        <button
          class="btn-add-game-primary"
          :disabled="creating"
          @click="createNewGame"
        >
          <span v-if="creating" class="spinner-sm"></span>
          <Plus v-else class="icon-sm" />
          {{ creating ? t('common.saving') : t('projects.createBtn') }}
        </button>
      </div>

      <!-- Пустой список по результатам поиска -->
      <div
        v-else-if="filteredGames.length === 0"
        class="state-container empty-card"
      >
        <Search class="icon-md text-muted" />
        <h3>{{ t('common.empty') }}</h3>
        <p>{{ t('stats.noData') }}</p>
        <button class="btn-reset" @click="resetFilters">
          {{ t('common.reset') }}
        </button>
      </div>

      <!-- Основной табличный вид -->
      <div v-else class="table-wrapper">
        <table class="yandex-games-table">
          <thead>
            <tr>
              <th class="col-game">{{ t('projects.projectNameLabel') }}</th>
              <th class="col-date">{{ t('common.updated') }}</th>
              <th class="col-status">{{ t('common.status') }}</th>
              <th class="col-actions"></th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="game in paginatedGames"
              :key="game.id"
              class="table-row"
              @click="openProject(game.id)"
            >
              <!-- 1 колонка: Игра (Иконка / Mock Draft + 2 строки: статус и название) -->
              <td class="col-game">
                <div class="game-cell">
                  <div class="game-icon-box">
                    <img
                      v-if="game.icon_path"
                      :src="getMediaUrl(game.icon_path)"
                      alt="Icon"
                      class="game-icon-img"
                    />
                    <div v-else class="game-icon-mock">
                      <span>Draft</span>
                    </div>
                  </div>
                  <div class="game-text">
                    <div class="game-type-label">
                      {{ getGameTypeLabel(game) }}
                    </div>
                    <div class="game-title">
                      {{ game.title_ru || game.title_en || '—' }}
                    </div>
                  </div>
                </div>
              </td>

              <!-- 2 колонка: Дата обновления -->
              <td class="col-date">
                <span class="date-text">
                  {{ formatProjectDate(game.updated_at || game.created_at) }}
                </span>
              </td>

              <!-- 3 колонка: Статус -->
              <td class="col-status">
                <span
                  class="status-pill"
                  :class="statusClass(game.status)"
                >
                  {{ statusLabel(game.status) }}
                </span>
              </td>

              <!-- Быстрые действия / Ссылки -->
              <td class="col-actions" @click.stop>
                <div class="row-actions">
                  <a
                    v-if="game.dev_url"
                    :href="game.dev_url"
                    target="_blank"
                    class="action-link"
                    title="Dev-версия"
                  >
                    <Play class="icon-xs" />
                  </a>
                  <a
                    v-if="game.prod_url && game.status === 3"
                    :href="game.prod_url"
                    target="_blank"
                    class="action-link"
                    title="Prod-версия"
                  >
                    <ExternalLink class="icon-xs" />
                  </a>
                  <button
                    class="btn-icon text-danger-hover"
                    title="Удалить проект"
                    @click="confirmDeleteProject(game)"
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

    <!-- Пагинация прикреплена к низу страницы по центру в стиле Яндекс.Консоли -->
    <div
      v-if="filteredGames.length > 0"
      class="pagination-container"
    >
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
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import {
  Plus,
  Gamepad2,
  Search,
  Play,
  Trash2,
  ExternalLink,
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
  ChevronDown,
} from 'lucide-vue-next';
import {
  listProjects,
  createProject,
  deleteProject,
  resetDraftState,
  statusClass,
  statusLabel,
  getMediaUrl,
} from '@/entities/project';
import { ProjectFilters } from '@/features/manage-projects';
import { formatProjectDate, showToast } from '@/shared/lib';

const { t } = useI18n();
const router = useRouter();
const games = ref([]);
const totalProjects = ref(0);
const loading = ref(false);
const creating = ref(false);

const searchQuery = ref('');
const statusFilter = ref('all');
const sortBy = ref('newest');

const currentPage = ref(1);
const pageSize = ref(10);

function getGameTypeLabel(game) {
  if (game.status === 3) return t('projects.published');
  if (game.status === 2) return t('projects.moderation');
  if (game.status === 4) return t('projects.rejected');
  return t('projects.draft');
}

async function loadProjects() {
  loading.value = true;
  try {
    const res = await listProjects({ limit: 100, offset: 0 });
    games.value = res.projects || [];
    totalProjects.value = res.total || games.value.length;
  } catch (err) {
    showToast(t('common.error'), 'danger');
  } finally {
    loading.value = false;
  }
}

const createNewGame = async () => {
  creating.value = true;
  try {
    const project = await createProject({
      title_ru: 'Новый проект',
      title_en: 'New Project',
    });
    resetDraftState();
    showToast(t('projects.createModalTitle') + ` #${project.id}`, 'success');
    router.push(`/projects/${project.id}/draft`);
  } catch (err) {
    showToast(t('common.error'), 'danger');
  } finally {
    creating.value = false;
  }
};

const openProject = (id) => {
  router.push(`/projects/${id}/draft`);
};

const confirmDeleteProject = async (game) => {
  const title = game.title_ru || game.title_en || `#${game.id}`;
  if (
    !confirm(
      `Вы действительно хотите удалить проект «${title}» и все его сборки?`
    )
  ) {
    return;
  }
  try {
    await deleteProject(game.id);
    games.value = games.value.filter((g) => g.id !== game.id);
    totalProjects.value = Math.max(0, totalProjects.value - 1);
    showToast('Проект успешно удалён', 'success');
  } catch (err) {
    showToast('Ошибка при удалении проекта', 'danger');
  }
};

const resetFilters = () => {
  searchQuery.value = '';
  statusFilter.value = 'all';
  sortBy.value = 'newest';
  currentPage.value = 1;
};

const filteredGames = computed(() => {
  let list = [...games.value];

  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase().trim();
    list = list.filter((g) => {
      const titleRu = (g.title_ru || '').toLowerCase();
      const titleEn = (g.title_en || '').toLowerCase();
      const idStr = String(g.id);
      return titleRu.includes(q) || titleEn.includes(q) || idStr.includes(q);
    });
  }

  if (statusFilter.value !== 'all') {
    const statusMap = { draft: 1, pending: 2, published: 3, rejected: 4 };
    const targetStatus = statusMap[statusFilter.value];
    list = list.filter((g) => g.status === targetStatus);
  }

  if (sortBy.value === 'newest') {
    list.sort(
      (a, b) => new Date(b.created_at || 0) - new Date(a.created_at || 0)
    );
  } else if (sortBy.value === 'oldest') {
    list.sort(
      (a, b) => new Date(a.created_at || 0) - new Date(b.created_at || 0)
    );
  } else if (sortBy.value === 'title') {
    list.sort((a, b) =>
      (a.title_ru || a.title_en || '').localeCompare(
        b.title_ru || b.title_en || ''
      )
    );
  }

  return list;
});

const totalPages = computed(
  () => Math.ceil(filteredGames.value.length / pageSize.value) || 1
);
const pageStart = computed(() => (currentPage.value - 1) * pageSize.value);

const paginatedGames = computed(() => {
  return filteredGames.value.slice(
    pageStart.value,
    pageStart.value + pageSize.value
  );
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

onMounted(loadProjects);
</script>

<style scoped>
.projects-page-container {
  width: 100%;
  max-width: 100%;
  min-height: calc(100vh - 60px);
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 24px 32px;
  box-sizing: border-box;
}

.main-content-wrap {
  width: 100%;
}

.state-container {
  padding: 60px 20px;
  text-align: center;
  color: var(--text-tertiary, #8b949e);
}

.empty-card {
  background: var(--bg-card, #161b22);
  border: 1px dashed var(--border, #30363d);
  border-radius: var(--radius-md, 8px);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 48px 24px;
}

.empty-icon-wrap {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: var(--bg-tertiary, #21262d);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--primary, #58a6ff);
  margin-bottom: 6px;
}

.btn-add-game-primary {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 18px;
  background: var(--primary, #58a6ff);
  color: #ffffff;
  border: none;
  border-radius: var(--radius-sm, 6px);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.15s;
}

.btn-add-game-primary:hover {
  background: var(--primary-hover, #79c0ff);
}

.btn-reset {
  padding: 6px 14px;
  background: var(--bg-secondary, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  color: var(--text-main, #f0f6fc);
  font-size: 13px;
  cursor: pointer;
}

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

.yandex-games-table th.col-game {
  width: 48%;
  padding-left: 8px;
}

.yandex-games-table th.col-date {
  width: 26%;
}

.yandex-games-table th.col-status {
  width: 16%;
}

.yandex-games-table th.col-actions {
  width: 10%;
  text-align: right;
  padding-right: 12px;
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

.table-row td.col-game {
  padding-left: 8px;
}

.table-row td.col-actions {
  padding-right: 12px;
  text-align: right;
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
  letter-spacing: 0.2px;
}

.game-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.game-type-label {
  font-size: 13px;
  font-weight: 400;
  color: var(--text-tertiary, #8b949e);
  line-height: 1.3;
}

.game-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
  line-height: 1.3;
}

.date-text {
  font-size: 13px;
  font-weight: 400;
  color: var(--text-muted, #b0b8c4);
}

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

.status-draft {
  background: var(--bg-tertiary, #21262d);
  border-color: var(--border, #30363d);
  color: var(--text-muted, #b0b8c4);
}

.status-pending {
  background: rgba(245, 176, 39, 0.12);
  border-color: rgba(245, 176, 39, 0.35);
  color: #f5b027;
}

.status-published {
  background: rgba(46, 204, 113, 0.12);
  border-color: rgba(46, 204, 113, 0.35);
  color: #2ecc71;
}

.status-rejected {
  background: rgba(248, 81, 73, 0.12);
  border-color: rgba(248, 81, 73, 0.35);
  color: #f85149;
}

.row-actions {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.action-link {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 4px;
  background: var(--bg-tertiary, #21262d);
  border: 1px solid var(--border, #30363d);
  color: var(--text-tertiary, #8b949e);
  transition: all 0.15s;
}

.action-link:hover {
  background: var(--bg-secondary, #161b22);
  color: var(--text-main, #f0f6fc);
  border-color: var(--border-secondary, #484f58);
}

.btn-icon {
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

.btn-icon:hover {
  color: var(--text-main, #f0f6fc);
  background: var(--bg-tertiary, #21262d);
}

.text-danger-hover:hover {
  color: var(--danger, #f85149) !important;
}

/* Пагинация прикреплена к низу */
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

.spinner-sm {
  width: 14px;
  height: 14px;
  border: 2px solid rgba(0, 0, 0, 0.25);
  border-top-color: #000000;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  display: inline-block;
}

.spinner-md {
  width: 32px;
  height: 32px;
  border: 3px solid var(--border, #30363d);
  border-top-color: #f5b027;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
