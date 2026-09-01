<template>
  <div class="app-layout">
    <transition name="toast-fade">
      <div v-if="toast.show" class="toast" :class="toast.type">
        {{ toast.message }}
      </div>
    </transition>
    <GlobalHeader v-if="$route.path !== '/login'" />
    <main class="page-content">
      <div
        v-if="userRole === 'Модератор' && $route.path.startsWith('/projects')"
        class="moderator-stub"
      >
        <h2>{{ t('roles.moderator') }}</h2>
        <p>
          {{ t('moderation.queueTitle') }}
        </p>
      </div>
      <router-view v-else />
    </main>
  </div>
</template>

<script setup>
import { computed, watch, onMounted } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { GlobalHeader } from '@/widgets/global-header';
import { toast } from '@/shared/lib';
import { useAuth } from '@/entities/user';
import { ROLE_MAP } from '@/shared/config';

const { t } = useI18n();
const router = useRouter();
const route = useRoute();
const { state: authState, loadUser } = useAuth();

const userRole = computed(() => ROLE_MAP[authState.user?.role] || 'Пользователь');

onMounted(() => {
  loadUser();
});

watch(userRole, (newRole) => {
  if (newRole === 'Разработчик') {
    if (route.path.startsWith('/moderator') || route.path.startsWith('/admin')) {
      router.push('/projects');
    }
  } else if (newRole === 'Модератор') {
    if (route.path.startsWith('/projects') || route.path.startsWith('/admin')) {
      router.push('/moderator/queue');
    }
  } else if (newRole === 'Администратор') {
    if (!route.path.startsWith('/admin') && route.path !== '/profile') {
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
