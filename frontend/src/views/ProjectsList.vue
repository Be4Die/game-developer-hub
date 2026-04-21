<template>
    <div class="page-container">
        <div class="header-row">
            <h1>Мои проекты</h1>
            <button class="btn-sm" @click="createNewGame">
                + Добавить игру
            </button>
        </div>

        <div class="projects-grid">
            <div
                class="card project-card"
                v-for="game in games"
                :key="game.id"
                @click="$router.push(`/projects/${game.id}`)"
            >
                <h3>{{ game.title }}</h3>
                <span class="badge">{{ game.status }}</span>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref } from "vue";
import { useRouter } from "vue-router";
const router = useRouter();

const games = ref([
    { id: 1, title: "RIVALS", status: "Опубликована" },
    { id: 2, title: "Новый проект", status: "Черновик" },
]);

const createNewGame = () => {
    const newId = Date.now();
    router.push(`/projects/${newId}/draft`);
};
</script>

<style scoped>
.page-container {
    padding: 32px 40px;
    max-width: 1200px;
    margin: 0 auto;
    width: 100%;
}
.header-row {
    display: flex;
    justify-content: space-between;
    margin-bottom: 24px;
}
.header-row .btn-sm {
    background: var(--primary);
    color: white;
    border: none;
    padding: 6px 14px;
    border-radius: 6px;
    cursor: pointer;
    font-weight: 600;
    font-size: 0.85rem;
    line-height: 1.4;
}
.header-row .btn-sm:hover {
    background: var(--primary-hover);
}
.projects-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 20px;
}
.project-card {
    background: var(--bg-card);
    border-color: var(--border);
    cursor: pointer;
}
.project-card h3 {
    color: var(--text-main);
    margin: 0 0 12px;
}
.project-card:hover {
    transform: translateY(-4px);
    box-shadow: 0 10px 15px rgba(0, 0, 0, 0.1);
}
.badge {
    background: var(--bg-app);
    color: var(--text-muted);
    padding: 4px 8px;
    border-radius: 12px;
    font-size: 0.8rem;
    font-weight: 600;
    display: inline-block;
}
</style>
