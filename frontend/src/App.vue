<template>
    <div class="app-layout">
        <ChatWidget v-if="!isChatRoute" />
        <transition name="toast-fade">
            <div v-if="toast.show" class="toast" :class="toast.type">
                {{ toast.message }}
            </div>
        </transition>
        <GlobalHeader v-if="$route.path !== '/login'" />
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
import ChatWidget from "./components/chat/ChatWidget.vue";
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

const isChatRoute = computed(() => {
    return route.path.startsWith("/moderator") || route.path.startsWith("/chat/") || route.path.startsWith("/projects") || route.path === "/login";
});

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
        if (!route.path.startsWith("/admin/dashboard")) {
            router.push("/admin/dashboard");
        }
    }
});
</script>

<style>
.app-layout {
    display: flex;
    flex-direction: column;
    min-height: 100vh;
}

.page-content {
    flex: 1;
    display: flex;
    flex-direction: column;
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
    color: var(--primary);
    font-size: 2rem;
    margin-bottom: 8px;
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
</style>
