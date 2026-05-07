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

        <div v-if="games.length === 0" class="empty-state">
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
                        {{ game.status }}
                    </span>
                </div>
                <h3 class="project-title">{{ game.title }}</h3>
                <p class="project-meta">ID: {{ game.id }}</p>
                <div class="project-arrow">
                    <ArrowRight class="icon-sm" />
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref } from "vue";
import { useRouter } from "vue-router";
import { Plus, Gamepad2, ArrowRight } from "lucide-vue-next";

const router = useRouter();

const games = ref([
    { id: 1, title: "RIVALS", status: "Опубликована" },
    { id: 2, title: "Новый проект", status: "Черновик" },
]);

function statusClass(status) {
    if (status === "Опубликована") return "badge-success";
    if (status === "Черновик") return "badge-neutral";
    return "badge-neutral";
}

function statusIndicatorClass(status) {
    if (status === "Опубликована") return "indicator-success";
    if (status === "Черновик") return "indicator-neutral";
    return "indicator-neutral";
}

const createNewGame = () => {
    const newId = Date.now();
    router.push(`/projects/${newId}/draft`);
};
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
