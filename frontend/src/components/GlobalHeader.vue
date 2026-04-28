<template>
  <header class="top-header">
    <div class="header-left">
      <div class="logo"><Layers class="logo-icon" /><span>WELWISE</span></div>
      <nav class="main-nav">
        <template v-if="userRole === 'Разработчик' || userRole === 'Владелец'">
          <router-link to="/projects" class="nav-item" active-class="active"><FolderGit2 class="icon-sm" /> Черновики</router-link>
        </template>
        <template v-if="userRole === 'Разработчик' || userRole === 'Владелец'">
          <router-link to="/nodes" class="nav-item" active-class="active"><Server class="icon-sm" /> Ноды</router-link>
        </template>
        <template v-if="userRole === 'Модератор'">
          <router-link to="/moderator/tickets" class="nav-item" active-class="active"><Inbox class="icon-sm" /> Очередь тикетов</router-link>
          <router-link to="/moderator/history" class="nav-item" active-class="active"><History class="icon-sm" /> История тикетов</router-link>
        </template>
        <template v-if="userRole === 'Администратор'">
          <router-link to="/moderator/roles" class="nav-item" active-class="active"><Users class="icon-sm" /> Управление ролями</router-link>
        </template>
      </nav>
    </div>

    <div class="header-right relative">
      <button class="profile-btn" @click="menuOpen = !menuOpen">
        <User class="icon-sm" /> {{ displayName }}
      </button>

      <div v-if="menuOpen" class="dropdown-overlay" @click="menuOpen = false"></div>
      <div v-if="menuOpen" class="dropdown">
        <div class="dropdown-header">
          <strong>{{ displayName }}</strong>
          <div class="text-muted">{{ userEmail }}</div>
          <div class="text-muted">Роль: {{ userRole }}</div>
        </div>
        <div class="dropdown-body">
          <router-link to="/settings" class="dropdown-item"><Settings class="icon-sm" /> Настройки</router-link>
          <button class="dropdown-item text-danger" @click="handleLogout"><LogOut class="icon-sm" /> Выйти</button>
        </div>
      </div>
    </div>
  </header>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuth } from '../store/auth'
import { Layers, FolderGit2, Server, User, Inbox, Settings, LogOut, Users, History } from 'lucide-vue-next'

const router = useRouter()
const { state: authState, logout } = useAuth()
const menuOpen = ref(false)

const ROLE_MAP = {
  0: 'Пользователь',
  1: 'Разработчик',
  2: 'Модератор',
  3: 'Администратор',
  'USER_ROLE_UNSPECIFIED': 'Пользователь',
  'USER_ROLE_DEVELOPER': 'Разработчик',
  'USER_ROLE_MODERATOR': 'Модератор',
  'USER_ROLE_ADMIN': 'Администратор',
}

const displayName = computed(() => authState.user?.display_name || 'Пользователь')
const userEmail = computed(() => authState.user?.email || '')
const userRole = computed(() => ROLE_MAP[authState.user?.role] || 'Пользователь')

async function handleLogout() {
  await logout()
  menuOpen.value = false
  router.push('/login')
}
</script>

<style scoped>
.top-header {
  height: 60px;
  background: var(--bg-card);
  border-bottom: 1px solid var(--border);
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 24px;
  position: sticky;
  top: 0;
  z-index: 50;
}
.header-left, .header-right {
  display: flex;
  align-items: center;
  gap: 32px;
}
.logo {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 800;
  font-size: 1.1rem;
  color: var(--text-main);
}
.logo-icon { color: var(--primary); }
.main-nav { display: flex; gap: 24px; }
.nav-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-weight: 500;
  color: var(--text-main);
  padding: 8px 0;
  border-bottom: 2px solid transparent;
}
.nav-item.active {
  border-color: var(--primary);
  color: var(--primary);
}
.text-muted { color: var(--text-muted); }

.profile-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  background: transparent;
  border: 1px solid var(--border);
  padding: 6px 16px;
  border-radius: 20px;
  font-weight: 500;
  cursor: pointer;
  color: var(--text-main);
}
.relative { position: relative; }
.dropdown-overlay { position: fixed; inset: 0; z-index: 90; }
.dropdown {
  position: absolute;
  top: calc(100% + 10px);
  right: 0;
  width: 220px;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  box-shadow: 0 10px 25px rgba(0,0,0,0.1);
  z-index: 100;
  display: flex;
  flex-direction: column;
}
.dropdown-header {
  padding: 12px;
  background: var(--bg-app);
  font-size: 0.85rem;
  border-bottom: 1px solid var(--border);
  border-radius: 8px 8px 0 0;
}
.dropdown-body { padding: 8px; display: flex; flex-direction: column; }
.dropdown-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border: none;
  background: none;
  text-align: left;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
  color: var(--text-main);
  text-decoration: none;
}
.dropdown-item:hover { background: var(--bg-app); }
.divider { height: 1px; background: var(--border); margin: 8px 0; }
.role-select {
  width: 100%;
  padding: 8px;
  border-radius: 6px;
  border: 1px solid var(--border);
  background: var(--bg-card);
  color: var(--text-main);
  font-weight: 600;
}
.text-danger { color: var(--danger); }
.text-danger:hover { background: var(--danger-light); }
</style>
