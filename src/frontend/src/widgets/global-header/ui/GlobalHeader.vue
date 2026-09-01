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
            to="/moderator/queue"
            class="nav-item"
            active-class="active"
          >
            <CheckSquare class="icon-sm" /> {{ t('header.moderation') }}
          </router-link>
          <router-link
            to="/moderator/chats"
            class="nav-item"
            active-class="active"
          >
            <MessageSquare class="icon-sm" /> {{ t('header.moderatorChats') }}
          </router-link>
          <router-link
            to="/moderator/archive"
            class="nav-item"
            active-class="active"
          >
            <Archive class="icon-sm" /> {{ t('header.moderationArchive') }}
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
          <router-link
            to="/moderator/queue"
            class="nav-item"
            active-class="active"
          >
            <CheckSquare class="icon-sm" /> {{ t('header.moderation') }}
          </router-link>
          <router-link
            to="/moderator/chats"
            class="nav-item"
            active-class="active"
          >
            <MessageSquare class="icon-sm" /> {{ t('header.moderatorChats') }}
          </router-link>
          <router-link
            to="/moderator/archive"
            class="nav-item"
            active-class="active"
          >
            <Archive class="icon-sm" /> {{ t('header.moderationArchive') }}
          </router-link>
        </template>
        <!-- Переход в профиль в общей панели навигации -->
        <router-link
          v-if="isAuthed"
          to="/profile"
          class="nav-item"
          active-class="active"
        >
          <User class="icon-sm" /> {{ t('header.profile') }}
        </router-link>
      </nav>
    </div>

    <!-- Правая панель: отображение пользователя и кнопка выхода -->
    <div class="header-right">
      <div v-if="isAuthed" class="header-user-section">
        <span class="user-name" :title="userEmail">{{ displayName }}</span>
        <button
          class="btn-logout"
          @click="handleLogout"
          :title="t('header.logout')"
        >
          <LogOut class="icon-sm" />
          <span>{{ t('header.logout') }}</span>
        </button>
      </div>
    </div>
  </header>
</template>

<script setup>
import { computed } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { useAuth } from '@/entities/user';
import { LogoIcon } from '@/shared/ui';
import {
  FolderGit2,
  Server,
  User,
  CheckSquare,
  MessageSquare,
  Archive,
  LogOut,
  Users,
} from 'lucide-vue-next';

const { t } = useI18n();
const router = useRouter();
const { state: authState, logout } = useAuth();

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
    router.push('/login');
  }
}
</script>

<style scoped>
.top-header {
  height: 60px;
  background: var(--bg-card, #161b22);
  border-bottom: 1px solid var(--border, #30363d);
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
  gap: 24px;
}

.logo-link {
  display: flex;
  align-items: center;
  text-decoration: none;
}

.main-nav {
  display: flex;
  gap: 6px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-weight: 500;
  font-size: 13px;
  color: var(--text-muted, #b0b8c4);
  padding: 6px 12px;
  border-radius: var(--radius-sm, 6px);
  border: 1px solid transparent;
  transition: all 0.15s ease;
}

.nav-item:hover {
  color: var(--text-main, #f0f6fc);
  background: var(--bg-secondary, #21262d);
}

.nav-item.active {
  color: var(--primary, #58a6ff);
  background: var(--primary-light, rgba(88, 166, 255, 0.1));
  border-color: transparent;
}

.header-user-section {
  display: flex;
  align-items: center;
  gap: 16px;
}

.user-name {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-main, #f0f6fc);
  max-width: 160px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.btn-logout {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: transparent;
  border: 1px solid var(--border, #30363d);
  padding: 6px 12px;
  border-radius: var(--radius-sm, 6px);
  font-weight: 500;
  font-size: 13px;
  cursor: pointer;
  color: var(--text-muted, #b0b8c4);
  transition: all 0.15s ease;
}

.btn-logout:hover {
  color: var(--danger, #f85149);
  background: var(--danger-light, rgba(248, 81, 73, 0.1));
  border-color: rgba(248, 81, 73, 0.3);
}

@media (max-width: 768px) {
  .main-nav {
    gap: 2px;
  }
  .user-name {
    display: none;
  }
  .nav-item span {
    display: none;
  }
}
</style>
