<template>
    <div class="page-container">
        <!-- Шапка страницы -->
        <div class="header-row">
            <div>
                <div class="page-subtitle">Консоль разработчика</div>
                <div class="title-with-count">
                    <h1>Мои игры</h1>
                    <span v-if="!loading" class="count-badge">{{ totalProjects }}</span>
                </div>
            </div>
            <div class="header-actions">
                <button class="btn btn-primary" :disabled="creating" @click="createNewGame">
                    <Plus v-if="!creating" class="icon-sm" />
                    <span v-else class="spinner-sm"></span>
                    {{ creating ? 'Создание...' : 'Добавить игру' }}
                </button>
            </div>
        </div>

        <!-- Панель фильтров и поиска (по образу Яндекс Игр) -->
        <div class="toolbar-card">
            <div class="search-box">
                <Search class="icon-sm search-icon" />
                <input
                    type="text"
                    v-model="searchQuery"
                    placeholder="Поиск по названию или ID..."
                    class="search-input"
                />
                <button v-if="searchQuery" class="clear-btn" @click="searchQuery = ''">
                    <X class="icon-xs" />
                </button>
            </div>

            <div class="filters-group">
                <div class="filter-item">
                    <label class="filter-label">Статус:</label>
                    <select v-model="statusFilter" class="filter-select">
                        <option value="all">Все статусы</option>
                        <option value="draft">Черновик</option>
                        <option value="pending">На модерации</option>
                        <option value="published">Опубликована</option>
                        <option value="rejected">Отклонена</option>
                    </select>
                </div>

                <div class="filter-item">
                    <label class="filter-label">Сортировка:</label>
                    <select v-model="sortBy" class="filter-select">
                        <option value="newest">Сначала новые</option>
                        <option value="oldest">Сначала старые</option>
                        <option value="title">По названию (А–Я)</option>
                    </select>
                </div>

                <div class="view-toggle">
                    <button
                        class="toggle-btn"
                        :class="{ active: viewMode === 'table' }"
                        title="Табличный вид"
                        @click="viewMode = 'table'"
                    >
                        <List class="icon-sm" />
                    </button>
                    <button
                        class="toggle-btn"
                        :class="{ active: viewMode === 'grid' }"
                        title="Вид карточек"
                        @click="viewMode = 'grid'"
                    >
                        <LayoutGrid class="icon-sm" />
                    </button>
                </div>
            </div>
        </div>

        <!-- Загрузка -->
        <div v-if="loading" class="state-container">
            <div class="spinner-md"></div>
            <p>Загрузка списка игр...</p>
        </div>

        <!-- Пустой список без проектов -->
        <div v-else-if="games.length === 0 && !searchQuery && statusFilter === 'all'" class="state-container empty-card">
            <div class="empty-icon-wrap">
                <Gamepad2 class="icon-lg" />
            </div>
            <h3>У вас пока нет проектов</h3>
            <p>Нажмите «Добавить игру», чтобы мгновенно создать новый черновик и загрузить веб-сборку.</p>
            <button class="btn btn-primary" :disabled="creating" @click="createNewGame">
                <Plus v-if="!creating" class="icon-sm" />
                <span v-else class="spinner-sm"></span>
                {{ creating ? 'Создание...' : 'Создать первую игру' }}
            </button>
        </div>

        <!-- Пустой список по результатам поиска -->
        <div v-else-if="filteredGames.length === 0" class="state-container empty-card">
            <Search class="icon-md text-muted" />
            <h3>Ничего не найдено</h3>
            <p>Попробуйте изменить поисковый запрос или сбросить фильтры.</p>
            <button class="btn btn-secondary btn-sm" @click="resetFilters">
                Сбросить фильтры
            </button>
        </div>

        <!-- Основной табличный вид (как в Яндекс Играх) -->
        <div v-else-if="viewMode === 'table'" class="table-container">
            <table class="games-table">
                <thead>
                    <tr>
                        <th class="col-game">Игра</th>
                        <th class="col-status">Статус</th>
                        <th class="col-version">Версия</th>
                        <th class="col-date">Дата обновления</th>
                        <th class="col-links">Окружение</th>
                        <th class="col-actions">Действия</th>
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
                                        <span v-if="game.title_en && game.title_ru" class="meta-sep">•</span>
                                        <span v-if="game.title_en && game.title_ru" class="en-tag">{{ game.title_en }}</span>
                                    </div>
                                </div>
                            </div>
                        </td>

                        <!-- Колонка: Статус -->
                        <td class="col-status">
                            <span class="status-pill" :class="statusClass(game.status)">
                                <span class="status-dot"></span>
                                {{ statusLabel(game.status) }}
                            </span>
                        </td>

                        <!-- Колонка: Версия сборки -->
                        <td class="col-version">
                            <span v-if="game.active_build_version" class="version-badge">
                                {{ game.active_build_version }}
                            </span>
                            <span v-else class="text-dim">—</span>
                        </td>

                        <!-- Колонка: Дата обновления -->
                        <td class="col-date">
                            <div class="date-cell">
                                <Clock class="icon-xs text-muted" />
                                <span>{{ formatDate(game.updated_at || game.created_at) }}</span>
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
                                <span v-else class="env-placeholder" title="Сборка не загружена">Dev —</span>

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

        <!-- Вид сетки (альтернативный режим) -->
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
                    <h3 class="project-title">{{ game.title_ru || game.title_en || 'Без названия' }}</h3>
                    <p class="project-id-text">ID: {{ game.id }}</p>
                    <div class="project-meta-row">
                        <span v-if="game.active_build_version" class="version-badge">
                            {{ game.active_build_version }}
                        </span>
                        <span class="date-text">{{ formatDate(game.updated_at || game.created_at) }}</span>
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

        <!-- Пагинация (как в Яндекс Играх) -->
        <div v-if="totalPages > 1 || filteredGames.length > pageSize" class="pagination-bar">
            <div class="pagination-info">
                Показано {{ pageStart + 1 }}–{{ Math.min(pageStart + pageSize, filteredGames.length) }} из {{ filteredGames.length }}
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
                    :class="{ active: currentPage === page, ellipsis: page === '...' }"
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
import { ref, computed, onMounted } from "vue";
import { useRouter } from "vue-router";
import {
    Plus,
    Gamepad2,
    Search,
    X,
    List,
    LayoutGrid,
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
} from "lucide-vue-next";
import { draftProject, showToast } from "../store";
import { listProjects, createProject, deleteProject } from "../api/projects";

const router = useRouter();
const games = ref([]);
const totalProjects = ref(0);
const loading = ref(false);
const creating = ref(false);

// Состояние фильтрации и вида
const searchQuery = ref("");
const statusFilter = ref("all");
const sortBy = ref("newest");
const viewMode = ref("table");

// Состояние пагинации
const currentPage = ref(1);
const pageSize = ref(10);

function resetDraftState() {
    draftProject.meta.titleRu = "";
    draftProject.meta.titleEn = "";
    draftProject.meta.seoRu = "";
    draftProject.meta.seoEn = "";
    draftProject.meta.about = "";
    draftProject.media.icon = null;
    draftProject.media.coverMain = null;
    draftProject.media.video = null;
    draftProject.builds = [];
    draftProject.activeBuildVersion = null;
}

function statusClass(status) {
    if (status === 3) return "status-published";
    if (status === 2) return "status-pending";
    if (status === 4) return "status-rejected";
    return "status-draft";
}

function statusLabel(status) {
    const map = { 1: "Черновик", 2: "На модерации", 3: "Опубликована", 4: "Отклонена" };
    return map[status] || "Черновик";
}

function formatDate(dateStr) {
    if (!dateStr) return "—";
    try {
        const d = new Date(dateStr);
        if (isNaN(d.getTime())) return dateStr;
        return d.toLocaleDateString("ru-RU", {
            day: "numeric",
            month: "short",
            year: "numeric",
            hour: "2-digit",
            minute: "2-digit",
        });
    } catch {
        return dateStr;
    }
}

function getMediaUrl(path) {
    if (!path) return "";
    if (path.startsWith("http://") || path.startsWith("https://") || path.startsWith("/api/")) {
        return path;
    }
    return `/api/v1/media/${path}`;
}

async function loadProjects() {
    loading.value = true;
    try {
        const res = await listProjects({ limit: 100, offset: 0 });
        games.value = res.projects || [];
        totalProjects.value = res.total || games.value.length;
    } catch (err) {
        showToast("Не удалось загрузить проекты", "danger");
    } finally {
        loading.value = false;
    }
}

// Быстрое создание проекта в стиле Яндекс Игр (клик -> черновик с ID -> редирект)
const createNewGame = async () => {
    creating.value = true;
    try {
        const project = await createProject({
            title_ru: "Новый проект",
            title_en: "New Project",
        });
        resetDraftState();
        showToast(`Создан черновик #${project.id}`, "success");
        router.push(`/projects/${project.id}/draft`);
    } catch (err) {
        showToast("Не удалось создать проект", "danger");
    } finally {
        creating.value = false;
    }
};

const openProject = (id) => {
    router.push(`/projects/${id}/draft`);
};

const confirmDeleteProject = async (game) => {
    const title = game.title_ru || game.title_en || `#${game.id}`;
    if (!confirm(`Вы действительно хотите удалить проект «${title}» и все его сборки?`)) {
        return;
    }
    try {
        await deleteProject(game.id);
        games.value = games.value.filter((g) => g.id !== game.id);
        totalProjects.value = Math.max(0, totalProjects.value - 1);
        showToast("Проект успешно удалён", "success");
    } catch (err) {
        showToast("Ошибка при удалении проекта", "danger");
    }
};

const resetFilters = () => {
    searchQuery.value = "";
    statusFilter.value = "all";
    sortBy.value = "newest";
    currentPage.value = 1;
};

// Фильтрация и сортировка
const filteredGames = computed(() => {
    let list = [...games.value];

    // Поиск по названию или ID
    if (searchQuery.value.trim()) {
        const q = searchQuery.value.toLowerCase().trim();
        list = list.filter((g) => {
            const titleRu = (g.title_ru || "").toLowerCase();
            const titleEn = (g.title_en || "").toLowerCase();
            const idStr = String(g.id);
            return titleRu.includes(q) || titleEn.includes(q) || idStr.includes(q);
        });
    }

    // Фильтр по статусу
    if (statusFilter.value !== "all") {
        const statusMap = { draft: 1, pending: 2, published: 3, rejected: 4 };
        const targetStatus = statusMap[statusFilter.value];
        list = list.filter((g) => g.status === targetStatus);
    }

    // Сортировка
    if (sortBy.value === "newest") {
        list.sort((a, b) => new Date(b.created_at || 0) - new Date(a.created_at || 0));
    } else if (sortBy.value === "oldest") {
        list.sort((a, b) => new Date(a.created_at || 0) - new Date(b.created_at || 0));
    } else if (sortBy.value === "title") {
        list.sort((a, b) => (a.title_ru || a.title_en || "").localeCompare(b.title_ru || b.title_en || ""));
    }

    return list;
});

// Пагинация
const totalPages = computed(() => Math.ceil(filteredGames.value.length / pageSize.value) || 1);
const pageStart = computed(() => (currentPage.value - 1) * pageSize.value);

const paginatedGames = computed(() => {
    return filteredGames.value.slice(pageStart.value, pageStart.value + pageSize.value);
});

const visiblePages = computed(() => {
    const total = totalPages.value;
    const current = currentPage.value;
    if (total <= 7) {
        return Array.from({ length: total }, (_, i) => i + 1);
    }
    if (current <= 4) {
        return [1, 2, 3, 4, 5, "...", total];
    }
    if (current >= total - 3) {
        return [1, "...", total - 4, total - 3, total - 2, total - 1, total];
    }
    return [1, "...", current - 1, current, current + 1, "...", total];
});

onMounted(loadProjects);
</script>

<style scoped>
.page-container {
    max-width: 1360px;
    margin: 0 auto;
    padding: 32px 24px 64px;
}

/* Шапка страницы */
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

/* Панель фильтров и поиска */
.toolbar-card {
    background: var(--bg-secondary);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-md);
    padding: 14px 18px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    margin-bottom: 20px;
    flex-wrap: wrap;
}

.search-box {
    position: relative;
    display: flex;
    align-items: center;
    flex: 1;
    min-width: 260px;
}

.search-icon {
    position: absolute;
    left: 12px;
    color: var(--text-tertiary);
}

.search-input {
    width: 100%;
    padding: 8px 32px 8px 36px;
    background: var(--bg-main);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-sm);
    color: var(--text-main);
    font-size: 0.9rem;
    outline: none;
    transition: border-color 0.2s;
}

.search-input:focus {
    border-color: var(--primary);
}

.clear-btn {
    position: absolute;
    right: 10px;
    background: transparent;
    border: none;
    color: var(--text-tertiary);
    cursor: pointer;
    padding: 2px;
    display: flex;
    align-items: center;
}

.filters-group {
    display: flex;
    align-items: center;
    gap: 16px;
    flex-wrap: wrap;
}

.filter-item {
    display: flex;
    align-items: center;
    gap: 8px;
}

.filter-label {
    font-size: 0.85rem;
    color: var(--text-tertiary);
    font-weight: 500;
}

.filter-select,
.size-select {
    padding: 7px 12px;
    background: var(--bg-main);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-sm);
    color: var(--text-main);
    font-size: 0.85rem;
    outline: none;
    cursor: pointer;
}

.view-toggle {
    display: flex;
    background: var(--bg-main);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-sm);
    overflow: hidden;
}

.toggle-btn {
    background: transparent;
    border: none;
    padding: 6px 10px;
    color: var(--text-tertiary);
    cursor: pointer;
    display: flex;
    align-items: center;
    transition: all 0.2s;
}

.toggle-btn.active {
    background: var(--bg-tertiary);
    color: var(--text-main);
}

/* Табличный вид */
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

.table-row:last-child td {
    border-bottom: none;
}

/* Колонки таблицы */
.game-cell {
    display: flex;
    align-items: center;
    gap: 14px;
}

.game-avatar {
    width: 44px;
    height: 44px;
    border-radius: 8px;
    background: var(--bg-tertiary);
    border: 1px solid var(--border-color);
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
    color: var(--text-tertiary);
}

.game-title {
    font-size: 0.95rem;
    font-weight: 600;
    color: var(--text-main);
    margin-bottom: 2px;
}

.game-meta {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 0.8rem;
    color: var(--text-tertiary);
    font-family: monospace;
}

.id-tag {
    color: var(--text-muted);
}

.meta-sep {
    color: var(--border-color);
}

.en-tag {
    color: var(--text-tertiary);
    font-family: sans-serif;
}

/* Бейджи статусов */
.status-pill {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 4px 10px;
    border-radius: 20px;
    font-size: 0.75rem;
    font-weight: 600;
}

.status-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
}

.status-draft {
    background: rgba(148, 163, 184, 0.15);
    color: #94A3B8;
}
.status-draft .status-dot {
    background: #94A3B8;
}

.status-pending {
    background: rgba(234, 179, 8, 0.15);
    color: #EAB308;
}
.status-pending .status-dot {
    background: #EAB308;
}

.status-published {
    background: rgba(34, 197, 94, 0.15);
    color: #22C55E;
}
.status-published .status-dot {
    background: #22C55E;
}

.status-rejected {
    background: rgba(239, 68, 68, 0.15);
    color: #EF4444;
}
.status-rejected .status-dot {
    background: #EF4444;
}

.version-badge {
    background: var(--bg-main);
    border: 1px solid var(--border-color);
    padding: 2px 8px;
    border-radius: 4px;
    font-size: 0.75rem;
    font-family: monospace;
    color: var(--primary);
    font-weight: 600;
}

.date-cell {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 0.85rem;
    color: var(--text-secondary);
}

.text-dim {
    color: var(--text-tertiary);
}

/* Ссылки на окружение */
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
    border-radius: 4px;
    font-size: 0.75rem;
    font-weight: 600;
    text-decoration: none;
    transition: opacity 0.2s;
}

.env-link:hover {
    opacity: 0.85;
}

.env-dev {
    background: rgba(59, 130, 246, 0.15);
    color: #60A5FA;
    border: 1px solid rgba(59, 130, 246, 0.3);
}

.env-prod {
    background: rgba(34, 197, 94, 0.15);
    color: #4ADE80;
    border: 1px solid rgba(34, 197, 94, 0.3);
}

.env-placeholder {
    font-size: 0.75rem;
    color: var(--text-tertiary);
}

.action-buttons {
    display: flex;
    align-items: center;
    gap: 6px;
}

.btn-icon {
    background: transparent;
    border: 1px solid transparent;
    color: var(--text-tertiary);
    border-radius: 4px;
    padding: 6px;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.2s;
}

.btn-icon:hover {
    background: var(--bg-tertiary);
    color: var(--text-main);
    border-color: var(--border-color);
}

.text-danger-hover:hover {
    color: #EF4444;
    background: rgba(239, 68, 68, 0.1);
    border-color: rgba(239, 68, 68, 0.2);
}

/* Сетка карточек */
.projects-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
    gap: 20px;
}

.project-card {
    background: var(--bg-secondary);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-md);
    padding: 18px;
    display: flex;
    flex-direction: column;
    gap: 14px;
    cursor: pointer;
    transition: transform 0.2s, border-color 0.2s, box-shadow 0.2s;
}

.project-card:hover {
    border-color: var(--primary);
    transform: translateY(-2px);
    box-shadow: var(--shadow-md);
}

.project-card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
}

.game-avatar-card {
    width: 48px;
    height: 48px;
    border-radius: 10px;
    background: var(--bg-tertiary);
    border: 1px solid var(--border-color);
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
}

.project-card-body {
    display: flex;
    flex-direction: column;
    gap: 4px;
}

.project-id-text {
    font-size: 0.8rem;
    color: var(--text-tertiary);
    font-family: monospace;
    margin: 0;
}

.project-meta-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-top: 8px;
}

.date-text {
    font-size: 0.75rem;
    color: var(--text-tertiary);
}

.project-card-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding-top: 12px;
    border-top: 1px solid var(--border-color);
}

/* Состояния */
.state-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 60px 24px;
    text-align: center;
    color: var(--text-secondary);
    gap: 14px;
}

.empty-card {
    background: var(--bg-secondary);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-md);
}

.empty-icon-wrap {
    width: 64px;
    height: 64px;
    border-radius: 16px;
    background: rgba(59, 130, 246, 0.1);
    color: var(--primary);
    display: flex;
    align-items: center;
    justify-content: center;
}

/* Пагинация */
.pagination-bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-top: 24px;
    padding: 12px 16px;
    background: var(--bg-secondary);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-md);
    flex-wrap: wrap;
    gap: 12px;
}

.pagination-info {
    font-size: 0.85rem;
    color: var(--text-tertiary);
}

.pagination-controls {
    display: flex;
    align-items: center;
    gap: 4px;
}

.page-btn {
    min-width: 32px;
    height: 32px;
    padding: 0 6px;
    background: var(--bg-main);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-sm);
    color: var(--text-secondary);
    font-size: 0.85rem;
    font-weight: 500;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    transition: all 0.15s ease;
}

.page-btn:hover:not(:disabled) {
    background: var(--bg-tertiary);
    color: var(--text-main);
    border-color: var(--text-tertiary);
}

.page-btn.active {
    background: var(--primary);
    color: #fff;
    border-color: var(--primary);
    font-weight: 700;
}

.page-btn:disabled {
    opacity: 0.4;
    cursor: not-allowed;
}

.page-btn.ellipsis {
    border: none;
    background: transparent;
    cursor: default;
}

.icon-xxs {
    width: 10px;
    height: 10px;
}

@media (max-width: 768px) {
    .toolbar-card {
        flex-direction: column;
        align-items: stretch;
    }
    .filters-group {
        justify-content: space-between;
    }
    .pagination-bar {
        flex-direction: column;
        align-items: center;
    }
}
</style>
