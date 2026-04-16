<template>
  <div class="app-layout">
    <transition name="toast-fade">
      <div v-if="toast.show" class="toast" :class="toast.type">{{ toast.message }}</div>
    </transition>
    <GlobalHeader />
    <main class="page-content">
      <div v-if="user.role === 'Модератор' && $route.path.includes('/projects')" class="moderator-stub">
        <h2>Режим модератора</h2>
        <p>Перейдите во вкладку "Очередь тикетов". Просмотр проектов недоступен.</p>
      </div>
      <router-view v-else />
    </main>
  </div>
</template>

<script setup>
import { watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import GlobalHeader from './components/GlobalHeader.vue'
import { toast, user } from './store'

const router = useRouter()
const route = useRoute()

watch(() => user.role, (newRole) => {
  if (newRole === 'Разработчик') {
    if (route.path.startsWith('/moderator')) {
      router.push('/projects')
    }
  } else if (newRole === 'Модератор') {
    if (route.path.includes('/projects') || route.path === '/moderator/roles') {
      router.push('/moderator/tickets')
    }
  } else if (newRole === 'Администратор') {
    if (!route.path.startsWith('/moderator/roles')) {
      router.push('/moderator/roles')
    }
  }
})
</script>

<style>
/* СВЕТЛАЯ ТЕМА (По умолчанию) */
:root {
  --bg-app: #F3F4F6;
  --bg-card: #FFFFFF;
  --bg-secondary: #F9FAFB;
  --bg-hover: #F3F4F6;
  --bg-input: #F9FAFB;
  --text-main: #111827;
  --text-muted: #6B7280;
  --border: #E5E7EB;
  --primary: #2563EB;
  --primary-hover: #1D4ED8;
  --success: #10B981;
  --success-light: #D1FAE5;
  --warning: #D97706;
  --warning-light: #FEF3C7;
  --danger: #EF4444;
  --danger-light: #FEF2F2;
  --radius-lg: 12px;
  --radius-md: 8px;
  --radius-sm: 6px;
}

/* ТЁМНАЯ ТЕМА */
[data-theme="dark"] {
  --bg-app: #111827;
  --bg-card: #1F2937;
  --bg-secondary: #1F2937;
  --bg-hover: #374151;
  --bg-input: #374151;
  --text-main: #F9FAFB;
  --text-muted: #9CA3AF;
  --border: #374151;
  --primary: #3B82F6;
  --primary-hover: #2563EB;
  --success: #10B981;
  --success-light: #064E3B;
  --warning: #F59E0B;
  --warning-light: #78350F;
  --danger: #EF4444;
  --danger-light: #7F1D1D;
  --radius-lg: 12px;
  --radius-md: 8px;
  --radius-sm: 6px;
}

body {
  margin: 0;
  background: var(--bg-app);
  color: var(--text-main);
  font-family: 'Inter', sans-serif;
  transition: background 0.3s ease, color 0.3s ease;
}

a { text-decoration: none; }

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
  color: #FFFFFF;
  border: none;
  padding: 8px 16px;
  border-radius: var(--radius-md);
  font-weight: 600;
  cursor: pointer;
  transition: 0.2s;
}

.btn-primary:hover { background: var(--primary-hover); }

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

.btn-outline:hover { background: var(--bg-hover); }

.icon-sm { width: 18px; height: 18px; }
.icon-md { width: 24px; height: 24px; }

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

.toast.success { background: var(--success); }
.toast.error { background: var(--danger); }
.toast.info { background: var(--primary); }

.toast-fade-enter-active, .toast-fade-leave-active { transition: all 0.3s; }
.toast-fade-enter-from, .toast-fade-leave-to { opacity: 0; transform: translate(-50%, -20px); }

.moderator-stub {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 60vh;
  text-align: center;
  color: var(--text-muted);
}

.moderator-stub h2 { color: #8B5CF6; font-size: 2rem; margin-bottom: 8px; }
</style>