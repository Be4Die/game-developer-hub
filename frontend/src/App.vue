<template>
    <div class="app-layout">
        <transition name="toast-fade">
            <div v-if="toast.show" class="toast" :class="toast.type">
                {{ toast.message }}
            </div>
        </transition>
        <GlobalHeader />
        <main class="page-content">
            <div
                v-if="
                    userRole === 'Модератор' &&
                    $route.path.includes('/projects') &&
                    !$route.path.includes('/nodes')
                "
                class="moderator-stub"
            >
                <h2>Режим модератора</h2>
                <p>
                    Перейдите во вкладку "Очередь тикетов". Просмотр проектов
                    недоступен.
                </p>
            </div>
            <router-view v-else />
        </main>
    </div>
</template>

<script setup>
import { computed, watch, onMounted } from "vue";
import { useRouter, useRoute } from "vue-router";
import GlobalHeader from "./components/GlobalHeader.vue";
import { toast } from "./store";
import { useAuth } from "./store/auth";

const router = useRouter();
const route = useRoute();
const { state: authState, loadUser } = useAuth();

const ROLE_MAP = {
    USER_ROLE_UNSPECIFIED: "Разработчик",
    USER_ROLE_DEVELOPER: "Разработчик",
    USER_ROLE_MODERATOR: "Модератор",
    USER_ROLE_ADMIN: "Администратор",
};

const userRole = computed(
    () => ROLE_MAP[authState.user?.role] || "Пользователь",
);

onMounted(() => {
    loadUser();
});

watch(userRole, (newRole) => {
    if (newRole === "Разработчик") {
        if (route.path.startsWith("/moderator")) {
            router.push("/projects");
        }
    } else if (newRole === "Модератор") {
        if (
            route.path.includes("/projects") ||
            route.path === "/moderator/roles"
        ) {
            router.push("/moderator/tickets");
        }
    } else if (newRole === "Администратор") {
        if (!route.path.startsWith("/moderator/roles")) {
            router.push("/moderator/roles");
        }
    }
});
</script>

<style>
/* СВЕТЛАЯ ТЕМА (По умолчанию) */
:root {
    --bg-app: #f3f4f6;
    --bg-card: #ffffff;
    --bg-secondary: #f9fafb;
    --bg-hover: #f3f4f6;
    --bg-input: #f9fafb;
    --text-main: #111827;
    --text-muted: #6b7280;
    --border: #e5e7eb;
    --primary: #2563eb;
    --primary-hover: #1d4ed8;
    --success: #10b981;
    --success-light: #d1fae5;
    --warning: #d97706;
    --warning-light: #fef3c7;
    --danger: #ef4444;
    --danger-light: #fef2f2;
    --radius-lg: 12px;
    --radius-md: 8px;
    --radius-sm: 6px;
}

/* ТЁМНАЯ ТЕМА */
[data-theme="dark"] {
    --bg-app: #111827;
    --bg-card: #1f2937;
    --bg-secondary: #1f2937;
    --bg-hover: #374151;
    --bg-input: #374151;
    --text-main: #f9fafb;
    --text-muted: #9ca3af;
    --border: #374151;
    --primary: #3b82f6;
    --primary-hover: #2563eb;
    --success: #10b981;
    --success-light: #064e3b;
    --warning: #f59e0b;
    --warning-light: #78350f;
    --danger: #ef4444;
    --danger-light: #7f1d1d;
    --radius-lg: 12px;
    --radius-md: 8px;
    --radius-sm: 6px;
}

body {
    margin: 0;
    background: var(--bg-app);
    color: var(--text-main);
    font-family: "Inter", sans-serif;
    transition:
        background 0.3s ease,
        color 0.3s ease;
}

a {
    text-decoration: none;
}

.app-layout {
    display: flex;
    flex-direction: column;
    min-height: 100vh;
}

.page-content {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
}

.card {
    background: var(--bg-card);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    padding: 24px;
}

.btn-primary {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    background: var(--primary);
    color: #ffffff;
    border: none;
    padding: 8px 16px;
    border-radius: var(--radius-md);
    font-weight: 600;
    cursor: pointer;
    transition: 0.2s;
}

.btn-primary:hover {
    background: var(--primary-hover);
}

.btn-outline {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    background: var(--bg-card);
    border: 1px solid var(--border);
    padding: 8px 16px;
    border-radius: var(--radius-md);
    font-weight: 600;
    cursor: pointer;
    color: var(--text-main);
    transition: 0.2s;
}

.btn-outline:hover {
    background: var(--bg-hover);
}

.icon-sm {
    width: 18px;
    height: 18px;
}
.icon-md {
    width: 24px;
    height: 24px;
}

.toast {
    position: fixed;
    top: 20px;
    left: 50%;
    transform: translateX(-50%);
    padding: 12px 24px;
    border-radius: var(--radius-md);
    font-weight: 600;
    color: white;
    z-index: 9999;
}

.toast.success {
    background: var(--success);
}
.toast.error {
    background: var(--danger);
}
.toast.info {
    background: var(--primary);
}

.toast-fade-enter-active,
.toast-fade-leave-active {
    transition: all 0.3s;
}
.toast-fade-enter-from,
.toast-fade-leave-to {
    opacity: 0;
    transform: translate(-50%, -20px);
}

.moderator-stub {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 60vh;
    text-align: center;
    color: var(--text-muted);
}

.moderator-stub h2 {
    color: #8b5cf6;
    font-size: 2rem;
    margin-bottom: 8px;
}
</style>
