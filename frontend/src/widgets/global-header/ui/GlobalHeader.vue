<template>
  <header class="top-header">
    <div class="header-left">
      <router-link to="/" class="logo-link">
        <LogoIcon :size="28" :textSize="16" :showSub="false" :noHover="true" />
      </router-link>
      <nav class="main-nav">
        <template v-if="isDeveloper">
          <router-link
            to="/projects"
            class="nav-item"
            active-class="active"
          >
            <FolderGit2 class="icon-sm" /> {{ t('header.projects') }}
          </router-link>
          <router-link
            to="/nodes"
            class="nav-item"
            active-class="active"
          >
            <Server class="icon-sm" /> {{ t('header.gameServers') }}
          </router-link>
        </template>
        <template v-if="isModerator">
          <router-link
            to="/moderator"
            class="nav-item"
            active-class="active"
          >
            <Inbox class="icon-sm" /> {{ t('header.moderatorPanel') }}
          </router-link>
        </template>
        <template v-if="isAdmin">
          <router-link
            to="/admin/dashboard"
            class="nav-item"
            active-class="active"
          >
            <Users class="icon-sm" /> {{ t('header.adminPanel') }}
          </router-link>
        </template>
      </nav>
    </div>

    <div class="header-right">
      <div v-if="isAuthed" class="profile-wrap relative">
        <button class="profile-btn" @click="menuOpen = !menuOpen">
          <User class="icon-sm" />
          <span class="profile-name">{{ displayName }}</span>
          <ChevronDown class="icon-sm" :class="{ rotate: menuOpen }" />
        </button>

        <div
          v-if="menuOpen"
          class="dropdown-overlay"
          @click="menuOpen = false"
        ></div>
        <transition name="dropdown">
          <div v-if="menuOpen" class="dropdown">
            <div class="dropdown-header">
              <div class="dropdown-name">{{ displayName }}</div>
              <div class="dropdown-email">{{ userEmail }}</div>
            </div>
            <div class="dropdown-body">
              <router-link
                to="/profile"
                class="dropdown-item"
                @click="menuOpen = false"
              >
                <User class="icon-sm" /> {{ t('header.profile') }}
              </router-link>
              <button class="dropdown-item text-danger" @click="handleLogout">
                <LogOut class="icon-sm" /> {{ t('header.logout') }}
              </button>
            </div>
          </div>
        </transition>
      </div>
    </div>
  </header>
</template>

<script setup>
import { ref, computed } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { useAuth } from '@/entities/user';
import { LogoIcon } from '@/shared/ui';
import {
  FolderGit2,
  Server,
  User,
  Inbox,
  LogOut,
  Users,
  ChevronDown,
} from 'lucide-vue-next';


const { t } = useI18n();
const router = useRouter();
const { state: authState, logout } = useAuth();
const menuOpen = ref(false);

const isAuthed = computed(() => !!authState.user);
const displayName = computed(
  () => authState.user?.display_name || authState.user?.email?.split('@')[0] || t('roles.user')
);
const userEmail = computed(() => authState.user?.email || '');

const userRole = computed(() => authState.user?.role);

const isAdmin = computed(() => {
  const r = userRole.value;
  return r === 'USER_ROLE_ADMIN' || r === 'admin' || r === 3;
});

const isModerator = computed(() => {
  const r = userRole.value;
  return r === 'USER_ROLE_MODERATOR' || r === 'moderator' || r === 2;
});

const isDeveloper = computed(() => {
  return !isAdmin.value && !isModerator.value;
});

async function handleLogout() {
  try {
    await logout();
  } catch {
    /* ignore api error on logout */
  } finally {
    menuOpen.value = false;
    router.push('/login');
  }
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
  backdrop-filter: blur(12px);
}

.header-left,
.header-right {
  display: flex;
  align-items: center;
  gap: 32px;
}

.logo-link {
  display: flex;
  align-items: center;
  text-decoration: none;
}

.main-nav {
  display: flex;
  gap: 8px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-weight: 500;
  font-size: 14px;
  color: var(--text-muted);
  padding: 6px 12px;
  border-radius: var(--radius-sm);
  border: 1px solid transparent;
  transition: all 0.2s;
}

.nav-item:hover {
  color: var(--text-main);
  background: var(--bg-secondary);
}

.nav-item.active {
  color: var(--primary);
  background: var(--primary-light);
  border-color: var(--primary-light);
}

.profile-wrap {
  position: relative;
}

.profile-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  background: transparent;
  border: 1px solid var(--border);
  padding: 6px 12px 6px 10px;
  border-radius: 20px;
  font-weight: 500;
  font-size: 14px;
  cursor: pointer;
  color: var(--text-main);
  transition: all 0.2s;
}

.profile-btn:hover {
  background: var(--bg-secondary);
  border-color: var(--border-secondary);
}

.profile-name {
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.profile-btn .rotate {
  transform: rotate(180deg);
  transition: transform 0.2s;
}

.relative {
  position: relative;
}

.dropdown-overlay {
  position: fixed;
  inset: 0;
  z-index: 90;
}

.dropdown {
  position: absolute;
  top: calc(100% + 10px);
  right: 0;
  width: 240px;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  z-index: 100;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  animation: dropdownIn 0.2s ease-out;
}

@keyframes dropdownIn {
  from {
    opacity: 0;
    transform: translateY(-6px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.dropdown-header {
  padding: 12px 16px;
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border);
}

.dropdown-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-main);
}

.dropdown-email {
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 2px;
}

.dropdown-body {
  padding: 8px;
  display: flex;
  flex-direction: column;
}

.dropdown-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border: none;
  background: none;
  text-align: left;
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-weight: 500;
  font-size: 14px;
  color: var(--text-main);
  text-decoration: none;
  transition: background 0.15s;
}

.dropdown-item:hover {
  background: var(--bg-secondary);
}

.text-danger {
  color: var(--danger);
}

.text-danger:hover {
  background: var(--danger-light);
}

@media (max-width: 768px) {
  .main-nav {
    display: none;
  }
  .profile-name {
    display: none;
  }
}
</style>
