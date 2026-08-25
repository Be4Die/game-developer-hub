<template>
  <div class="page-container">
    <!-- Шапка страницы -->
    <div class="header-row">
      <div>
        <div class="page-subtitle">{{ t('projects.subtitle') }}</div>
        <div class="title-with-count">
          <h1>{{ t('projects.title') }}</h1>
          <span v-if="!loading" class="count-badge">{{ totalProjects }}</span>
        </div>
      </div>
      <div class="header-actions">
        <button
          class="btn btn-primary"
          :disabled="creating"
          @click="createNewGame"
        >
          <Plus v-if="!creating" class="icon-sm" />
          <span v-else class="spinner-sm"></span>
          {{ creating ? t('common.saving') : t('projects.createBtn') }}
        </button>
      </div>
    </div>

    <!-- Панель фильтров и поиска -->
    <ProjectFilters
      v-model:searchQuery="searchQuery"
      v-model:statusFilter="statusFilter"
      v-model:sortBy="sortBy"
      v-model:viewMode="viewMode"
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
        class="btn btn-primary"
        :disabled="creating"
        @click="createNewGame"
      >
        <Plus v-if="!creating" class="icon-sm" />
        <span v-else class="spinner-sm"></span>
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
      <button class="btn btn-secondary btn-sm" @click="resetFilters">
        {{ t('common.reset') }}
      </button>
    </div>

    <!-- Основной табличный вид -->
    <div v-else-if="viewMode === 'table'" class="table-container">
      <table class="games-table">
        <thead>
          <tr>
            <th class="col-game">{{ t('projects.projectNameLabel') }}</th>
            <th class="col-status">{{ t('common.status') }}</th>
            <th class="col-version">{{ t('common.version') }}</th>
            <th class="col-date">{{ t('common.updated') }}</th>
            <th class="col-links">Env</th>
            <th class="col-actions">{{ t('common.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="game in paginatedGames"
            :key="game.id"
            class="table-row"
            @click="openProject(game.id)"
          >
            <!-- Колонка: Иконка + Название + ID -->
            <td class="col-game">
              <div class="game-cell">

                <div class="game-avatar">
                  <img
                    v-if="game.icon_path"
                    :src="getMediaUrl(game.icon_path)"
                    alt="Icon"
                    class="avatar-img"
                  />
                  <Gamepad2 v-else class="icon-sm avatar-fallback" />
                </div>
                <div class="game-info">
                  <div class="game-title">
                    {{ game.title_ru || game.title_en || 'Без названия' }}
                  </div>
                  <div class="game-meta">
                    <span class="id-tag">ID: {{ game.id }}</span>
                    <span
                      v-if="game.title_en && game.title_ru"
                      class="meta-sep"
                      >•</span
                    >
                    <span
                      v-if="game.title_en && game.title_ru"
                      class="en-tag"
                      >{{ game.title_en }}</span
                    >
                  </div>
                </div>
              </div>
            </td>

            <!-- Колонка: Статус -->
            <td class="col-status">
              <span
                class="status-pill"
                :class="statusClass(game.status)"
              >
                <span class="status-dot"></span>
                {{ statusLabel(game.status) }}
              </span>
            </td>

            <!-- Колонка: Версия сборки -->
            <td class="col-version">
              <span
                v-if="game.active_build_version"
                class="version-badge"
              >
                {{ game.active_build_version }}
              </span>
              <span v-else class="text-dim">—</span>
            </td>

            <!-- Колонка: Дата обновления -->
            <td class="col-date">
              <div class="date-cell">
                <Clock class="icon-xs text-muted" />
                <span>{{
                  formatDate(game.updated_at || game.created_at)
                }}</span>
              </div>
            </td>

            <!-- Колонка: Быстрые ссылки Dev / Prod -->
            <td class="col-links" @click.stop>
              <div class="env-buttons">
                <a
                  v-if="game.dev_url"
                  :href="game.dev_url"
                  target="_blank"
                  class="env-link env-dev"
                  title="Запустить Dev-сборку в новой вкладке"
                >
                  <Play class="icon-xs" />
                  Dev
                  <ExternalLink class="icon-xxs" />
                </a>
                <span
                  v-else
                  class="env-placeholder"
                  title="Сборка не загружена"
                  >Dev —</span
                >

                <a
                  v-if="game.prod_url && game.status === 3"
                  :href="game.prod_url"
                  target="_blank"
                  class="env-link env-prod"
                  title="Открыть опубликованную игру"
                >
                  <CheckCircle2 class="icon-xs" />
                  Prod
                  <ExternalLink class="icon-xxs" />
                </a>
              </div>
            </td>

            <!-- Колонка: Действия -->
            <td class="col-actions" @click.stop>
              <div class="action-buttons">
                <button
                  class="btn-icon"
                  title="Редактировать проект"
                  @click="openProject(game.id)"
                >
                  <Edit3 class="icon-xs" />
                </button>
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

    <!-- Вид сетки -->
    <div v-else class="projects-grid">
      <div
        v-for="game in paginatedGames"
        :key="game.id"
        class="card project-card card-hover"
        @click="openProject(game.id)"
      >
        <div class="project-card-header">
          <div class="game-avatar-card">
            <img
              v-if="game.icon_path"
              :src="getMediaUrl(game.icon_path)"
              alt="Icon"
              class="avatar-img"
            />
            <Gamepad2 v-else class="icon-md avatar-fallback" />
          </div>
          <span class="status-pill" :class="statusClass(game.status)">
            <span class="status-dot"></span>
            {{ statusLabel(game.status) }}
          </span>
        </div>

        <div class="project-card-body">
          <h3 class="project-title">
            {{ game.title_ru || game.title_en || 'Без названия' }}
          </h3>
          <p class="project-id-text">ID: {{ game.id }}</p>
          <div class="project-meta-row">
            <span
              v-if="game.active_build_version"
              class="version-badge"
            >
              {{ game.active_build_version }}
            </span>
            <span class="date-text">{{
              formatDate(game.updated_at || game.created_at)
            }}</span>
          </div>
        </div>

        <div class="project-card-footer" @click.stop>
          <div class="env-buttons">
            <a
              v-if="game.dev_url"
              :href="game.dev_url"
              target="_blank"
              class="env-link env-dev"
            >
              <Play class="icon-xs" />
              Dev
            </a>
            <a
              v-if="game.prod_url && game.status === 3"
              :href="game.prod_url"
              target="_blank"
              class="env-link env-prod"
            >
              <CheckCircle2 class="icon-xs" />
              Prod
            </a>
          </div>
          <div class="action-buttons">
            <button
              class="btn-icon"
              title="Редактировать"
              @click="openProject(game.id)"
            >
              <Edit3 class="icon-xs" />
            </button>
            <button
              class="btn-icon text-danger-hover"
              title="Удалить"
              @click="confirmDeleteProject(game)"
            >
              <Trash2 class="icon-xs" />
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Пагинация -->
    <div
      v-if="totalPages > 1 || filteredGames.length > pageSize"
      class="pagination-bar"
    >
      <div class="pagination-info">
        Показано {{ pageStart + 1 }}–{{
          Math.min(pageStart + pageSize, filteredGames.length)
        }}
        из {{ filteredGames.length }}
      </div>

      <div class="pagination-controls">
        <button
          class="page-btn"
          :disabled="currentPage === 1"
          title="Первая страница"
          @click="currentPage = 1"
        >
          <ChevronsLeft class="icon-sm" />
        </button>
        <button
          class="page-btn"
          :disabled="currentPage === 1"
          title="Предыдущая"
          @click="currentPage--"
        >
          <ChevronLeft class="icon-sm" />
        </button>

        <button
          v-for="page in visiblePages"
          :key="page"
          class="page-btn"
          :class="{
            active: currentPage === page,
            ellipsis: page === '...',
          }"
          :disabled="page === '...'"
          @click="typeof page === 'number' && (currentPage = page)"
        >
          {{ page }}
        </button>

        <button
          class="page-btn"
          :disabled="currentPage === totalPages"
          title="Следующая"
          @click="currentPage++"
        >
          <ChevronRight class="icon-sm" />
        </button>
        <button
          class="page-btn"
          :disabled="currentPage === totalPages"
          title="Последняя страница"
          @click="currentPage = totalPages"
        >
          <ChevronsRight class="icon-sm" />
        </button>
      </div>

      <div class="page-size-selector">
        <select v-model="pageSize" class="size-select">
          <option :value="10">10 на стр.</option>
          <option :value="25">25 на стр.</option>
          <option :value="50">50 на стр.</option>
        </select>
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
  Clock,
  Play,
  CheckCircle2,
  Edit3,
  Trash2,
  ExternalLink,
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
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
import { formatDate, showToast } from '@/shared/lib';

const { t } = useI18n();
const router = useRouter();
const games = ref([]);
const totalProjects = ref(0);
const loading = ref(false);
const creating = ref(false);

const searchQuery = ref('');
const statusFilter = ref('all');
const sortBy = ref('newest');
const viewMode = ref('table');

const currentPage = ref(1);
const pageSize = ref(10);

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
.page-container {
  max-width: 1360px;
  margin: 0 auto;
  padding: 32px 24px 64px;
}

.header-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
}

.page-subtitle {
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.8px;
  color: var(--text-tertiary);
  margin-bottom: 4px;
}

.title-with-count {
  display: flex;
  align-items: center;
  gap: 12px;
}

.title-with-count h1 {
  font-size: 1.75rem;
  font-weight: 700;
  letter-spacing: -0.5px;
  margin: 0;
  color: var(--text-main);
}

.count-badge {
  background: var(--bg-tertiary);
  color: var(--text-secondary);
  font-size: 0.85rem;
  font-weight: 600;
  padding: 2px 10px;
  border-radius: 12px;
}

.spinner-sm {
  width: 14px;
  height: 14px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  display: inline-block;
}

.spinner-md {
  width: 32px;
  height: 32px;
  border: 3px solid var(--border-color);
  border-top-color: var(--primary);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.state-container {
  padding: 60px 20px;
  text-align: center;
  color: var(--text-muted);
}

.empty-card {
  background: var(--bg-card);
  border: 1px dashed var(--border);
  border-radius: var(--radius-lg);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.empty-icon-wrap {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  background: var(--bg-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--primary);
  margin-bottom: 8px;
}

.table-container {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  overflow-x: auto;
  box-shadow: var(--shadow-sm);
}

.games-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
}

.games-table th {
  background: var(--bg-tertiary);
  color: var(--text-tertiary);
  font-size: 0.75rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.6px;
  padding: 12px 18px;
  border-bottom: 1px solid var(--border-color);
}

.games-table td {
  padding: 14px 18px;
  border-bottom: 1px solid var(--border-color);
  vertical-align: middle;
}

.table-row {
  cursor: pointer;
  transition: background-color 0.15s ease;
}

.table-row:hover {
  background: rgba(255, 255, 255, 0.03);
}

.game-cell {
  display: flex;
  align-items: center;
  gap: 14px;
}

.game-avatar {
  width: 44px;
  height: 44px;
  border-radius: var(--radius-sm);
  background: var(--bg-card);
  border: 1px solid var(--border);
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  flex-shrink: 0;
}

.avatar-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.avatar-fallback {
  color: var(--text-muted);
}

.game-title {
  font-weight: 600;
  font-size: 0.95rem;
  color: var(--text-main);
  margin-bottom: 3px;
}

.game-meta {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.75rem;
  color: var(--text-tertiary);
}

.status-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 0.8rem;
  font-weight: 600;
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: currentColor;
}

.status-draft {
  background: var(--bg-tertiary);
  color: var(--text-muted);
}

.status-pending {
  background: var(--warning-light);
  color: var(--warning);
}

.status-published {
  background: var(--success-light);
  color: var(--success);
}

.status-rejected {
  background: var(--danger-light);
  color: var(--danger);
}

.version-badge {
  font-family: monospace;
  font-size: 0.8rem;
  background: var(--bg-card);
  padding: 2px 6px;
  border-radius: 4px;
  border: 1px solid var(--border);
}

.date-cell {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.85rem;
  color: var(--text-muted);
}

.env-buttons {
  display: flex;
  align-items: center;
  gap: 8px;
}

.env-link {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  border-radius: var(--radius-sm);
  font-size: 0.75rem;
  font-weight: 600;
  text-decoration: none;
  transition: opacity 0.2s;
}

.env-link:hover {
  opacity: 0.8;
}

.env-dev {
  background: var(--primary-light);
  color: var(--primary);
}

.env-prod {
  background: var(--success-light);
  color: var(--success);
}

.env-placeholder {
  font-size: 0.75rem;
  color: var(--text-tertiary);
}

.action-buttons {
  display: flex;
  gap: 8px;
}

.btn-icon {
  background: none;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  padding: 6px;
  cursor: pointer;
  color: var(--text-muted);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s;
}

.btn-icon:hover {
  color: var(--text-main);
  background: var(--bg-card);
}

.text-danger-hover:hover {
  color: var(--danger);
  border-color: var(--danger);
}

/* Grid mode */
.projects-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 20px;
}

.project-card {
  cursor: pointer;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.project-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.game-avatar-card {
  width: 48px;
  height: 48px;
  border-radius: var(--radius-sm);
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.project-title {
  margin: 0 0 4px 0;
  font-size: 1.05rem;
}

.project-id-text {
  font-size: 0.8rem;
  color: var(--text-tertiary);
  margin: 0 0 8px 0;
}

.project-meta-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.date-text {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.project-card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 12px;
  border-top: 1px solid var(--border);
}

/* Pagination */
.pagination-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 24px;
  gap: 16px;
  flex-wrap: wrap;
}

.pagination-info {
  font-size: 0.85rem;
  color: var(--text-muted);
}

.pagination-controls {
  display: flex;
  gap: 4px;
}

.page-btn {
  padding: 6px 12px;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  background: var(--bg-card);
  color: var(--text-main);
  cursor: pointer;
  font-size: 0.85rem;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s;
}

.page-btn:hover:not(:disabled) {
  background: var(--bg-secondary);
  border-color: var(--primary);
}

.page-btn.active {
  background: var(--primary);
  color: white;
  border-color: var(--primary);
}

.page-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.page-btn.ellipsis {
  border: none;
  background: none;
}

.size-select {
  padding: 6px 10px;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  background: var(--bg-card);
  color: var(--text-main);
  font-size: 0.85rem;
  outline: none;
}
</style>
