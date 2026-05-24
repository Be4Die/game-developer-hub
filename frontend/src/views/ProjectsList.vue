<template>
    <div class="page-container">
        <div class="header-row">
            <div>
                <div class="page-subtitle">Проекты</div>
                <h1>Мои игры</h1>
            </div>
            <button class="btn btn-primary" @click="createNewGame">
                <Plus class="icon-sm" />
                Добавить игру
            </button>
        </div>

        <div v-if="loading" class="empty-state">
            <p>Загрузка проектов...</p>
        </div>

        <div v-else-if="games.length === 0" class="empty-state">
            <div class="empty-icon">
                <Gamepad2 class="icon-md" />
            </div>
            <h3>Пока нет проектов</h3>
            <p>Создайте свою первую игру, чтобы начать работу</p>
            <button class="btn btn-primary" @click="createNewGame">
                <Plus class="icon-sm" />
                Создать проект
            </button>
        </div>

        <div v-else class="projects-grid">
            <div
                class="card project-card card-hover"
                v-for="game in games"
                :key="game.id"
                @click="$router.push(`/projects/${game.id}`)"
            >
                <div class="project-status-row">
                    <span class="status-indicator" :class="statusIndicatorClass(game.status)"></span>
                    <span class="badge" :class="statusClass(game.status)">
                        {{ statusLabel(game.status) }}
                    </span>
                </div>
                <h3 class="project-title">{{ game.title_ru || game.title_en || 'Без названия' }}</h3>
                <p class="project-meta">ID: {{ game.id }}</p>
                <div class="project-arrow">
                    <ArrowRight class="icon-sm" />
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import { useRouter } from "vue-router";
import { Plus, Gamepad2, ArrowRight } from "lucide-vue-next";
import { draftProject } from "../store";
import { listProjects, createProject } from "../api/projects";
import { showToast } from "../store";

const router = useRouter();
const games = ref([]);
const loading = ref(false);

function resetDraftProject() {
    draftProject.meta.titleRu = ''
    draftProject.meta.titleEn = ''
    draftProject.meta.seoRu = ''
    draftProject.meta.seoEn = ''
    draftProject.meta.about = ''
    draftProject.media.icon = null
    draftProject.media.coverMain = null
    draftProject.media.video = null
    draftProject.builds = []
    draftProject.activeBuildVersion = null
}

function statusClass(status) {
    if (status === 3) return "badge-success"; // published
    if (status === 1) return "badge-neutral"; // draft
    return "badge-neutral";
}

function statusIndicatorClass(status) {
    if (status === 3) return "indicator-success";
    if (status === 1) return "indicator-neutral";
    return "indicator-neutral";
}

function statusLabel(status) {
    const map = { 1: "Черновик", 2: "На модерации", 3: "Опубликована", 4: "Отклонена" };
    return map[status] || "Черновик";
}

async function loadProjects() {
    loading.value = true;
    try {
        games.value = await listProjects();
    } catch (err) {
        showToast("Не удалось загрузить проекты", "danger");
    } finally {
        loading.value = false;
    }
}

const createNewGame = async () => {
    try {
        const project = await createProject({ title_ru: "Новый проект", title_en: "New Project" });
        resetDraftProject();
        games.value.push(project);
        router.push(`/projects/${project.id}/draft`);
    } catch (err) {
        showToast("Не удалось создать проект", "danger");
    }
};

onMounted(loadProjects);
</script>

<style scoped>
.page-subtitle {
    font-size: 12px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    color: var(--text-tertiary);
    margin-bottom: 4px;
}

.header-row h1 {
    font-size: 1.5rem;
    font-weight: 700;
    letter-spacing: -0.5px;
    margin: 0;
}

.projects-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 20px;
}

.project-card {
    cursor: pointer;
    padding: 24px;
    display: flex;
    flex-direction: column;
    gap: 10px;
    position: relative;
}

.project-status-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
}

.status-indicator {
    width: 8px;
    height: 8px;
    border-radius: 50%;
}

.indicator-success {
    background: var(--success);
    box-shadow: 0 0 0 3px var(--success-light);
}

.indicator-neutral {
    background: var(--text-tertiary);
    box-shadow: 0 0 0 3px var(--bg-tertiary);
}

.project-title {
    font-size: 1.15rem;
    font-weight: 700;
    color: var(--text-main);
    margin: 0;
    letter-spacing: -0.3px;
}

.project-meta {
    font-size: 0.8rem;
    color: var(--text-tertiary);
    margin: 0;
    font-family: monospace;
}

.project-arrow {
    position: absolute;
    right: 20px;
    bottom: 20px;
    color: var(--text-tertiary);
    opacity: 0;
    transform: translateX(-4px);
    transition: all 0.2s ease;
}

.project-card:hover .project-arrow {
    opacity: 1;
    transform: translateX(0);
    color: var(--primary);
}

.empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 80px 24px;
    text-align: center;
    color: var(--text-muted);
    gap: 16px;
}

.empty-icon {
    width: 64px;
    height: 64px;
    border-radius: var(--radius-lg);
    background: var(--bg-secondary);
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--primary);
}

.empty-state h3 {
    color: var(--text-main);
    margin: 0;
    font-size: 1.1rem;
}

.empty-state p {
    margin: 0;
    font-size: 0.9rem;
}

@media (max-width: 968px) {
    .projects-grid {
        grid-template-columns: repeat(2, 1fr);
    }
}

@media (max-width: 640px) {
    .projects-grid {
        grid-template-columns: 1fr;
    }
}
</style>
