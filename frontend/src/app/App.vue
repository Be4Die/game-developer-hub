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
          $route.path.startsWith('/projects')
        "
        class="moderator-stub"
      >
        <h2>Режим модератора</h2>
        <p>
          Прямой просмотр проектов недоступен. Перейдите во вкладку "Очередь модерации".
        </p>
      </div>
      <router-view v-else />
    </main>
  </div>
</template>

<script setup>
import { computed, watch, onMounted } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { GlobalHeader } from '@/widgets/global-header';
import { ChatWidget } from '@/widgets/chat-widget';
import { toast } from '@/shared/lib';
import { useAuth } from '@/entities/user';
import {
  startChatNotifications,
  stopChatNotifications,
} from '@/entities/chat';
import { ROLE_MAP } from '@/shared/config';

const router = useRouter();
const route = useRoute();
const { state: authState, loadUser } = useAuth();

const userRole = computed(
  () => ROLE_MAP[authState.user?.role] || 'Пользователь'
);

const isChatRoute = computed(() => {
  return (
    route.path.startsWith('/moderator') ||
    route.path.startsWith('/chat/') ||
    route.path.startsWith('/projects') ||
    route.path === '/login'
  );
});

onMounted(() => {
  loadUser();
});

watch(
  () => authState.accessToken,
  (token) => {
    if (token) startChatNotifications();
    else stopChatNotifications();
  },
  { immediate: true }
);

watch(userRole, (newRole) => {
  if (newRole === 'Разработчик') {
    if (route.path.startsWith('/moderator')) {
      router.push('/projects');
    }
  } else if (newRole === 'Модератор') {
    if (route.path.startsWith('/projects')) {
      router.push('/moderator/queue');
    }
  } else if (newRole === 'Администратор') {
    if (!route.path.startsWith('/admin/dashboard')) {
      router.push('/admin/dashboard');
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
