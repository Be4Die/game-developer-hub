<template>
  <div class="game-workspace">
    <!-- ЛЕВОЕ МЕНЮ ИГРЫ -->
    <aside class="game-sidebar">
      <div class="game-header">
        <button class="back-btn" @click="$router.push('/projects')">
          <ArrowLeft class="icon-sm" /> {{ t('common.back') }}
        </button>
        <h2 class="game-title-short">{{ projectTitle }}</h2>
      </div>
      <nav class="game-nav">
        <router-link
          :to="`/projects/${id}/stats`"
          class="nav-btn"
          active-class="active"
        >
          <BarChart2 class="icon-sm" /> {{ t('projectWorkspace.statsTab') }}
        </router-link>
        <router-link
          :to="`/projects/${id}/draft`"
          class="nav-btn"
          active-class="active"
        >
          <PenTool class="icon-sm" /> {{ t('projectWorkspace.draftTab') }}
        </router-link>
        <router-link
          v-if="isPublished"
          :to="`/projects/${id}/published`"
          class="nav-btn"
          active-class="active"
        >
          <CheckCircle class="icon-sm" /> {{ t('projectWorkspace.publishedTab') }}
        </router-link>
        <router-link
          :to="`/projects/${id}/servers`"
          class="nav-btn"
          active-class="active"
        >
          <Server class="icon-sm" /> {{ t('projectWorkspace.serversTab') }}
        </router-link>
      </nav>
    </aside>

    <!-- ЦЕНТР (Подгружает табы) -->
    <main class="content-area">
      <router-view />
    </main>

    <!-- ПРАВЫЙ САЙДБАР: Чат -->
    <aside class="chat-sidebar">
      <div class="chat-sidebar-header">
        <MessageSquare class="icon-sm" />
        <h3>{{ t('moderation.chatTitle') }}</h3>
      </div>
      <ProjectChat :projectId="id" class="workspace-chat" />
    </aside>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, watch, provide } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import {
  ArrowLeft,
  BarChart2,
  PenTool,
  CheckCircle,
  Server,
  MessageSquare,
} from 'lucide-vue-next';
import { getProject } from '@/entities/project';
import { ProjectChat } from '@/entities/moderation';
import { useAuth } from '@/entities/user';
import { showToast, formatTime } from '@/shared/lib';

const { t } = useI18n();
const props = defineProps(['id']);
const route = useRoute();
const router = useRouter();

// ─── Project data (shared with child tabs) ───────────────────
const project = ref(null);
provide('project', project);

async function loadProject() {
  try {
    project.value = await getProject(props.id);
  } catch (err) {
    // silently fail, fallback to id
  }
}

watch(() => props.id, loadProject, { immediate: true });

const projectTitle = computed(() => {
  return (
    project.value?.title_ru ||
    project.value?.title_en ||
    `${t('projects.projectNameLabel')} #${props.id}`
  );
});


const isPublished = computed(() => project.value?.status === 3);

// Редирект с "published" на "draft", если проект загружен и не опубликован
watch(
  () => [route.name, project.value?.status],
  ([name]) => {
    if (name === 'published' && project.value && !isPublished.value) {
      router.replace(`/projects/${props.id}/draft`);
    }
  },
  { immediate: true }
);

const { state: authState } = useAuth();
const currentUserId = computed(() => authState.user?.id);

</script>

<style scoped>
.game-workspace {
  display: flex;
  min-height: calc(100vh - 60px);
  background: var(--bg-app);
}
.scrollable {
  overflow-y: auto;
}

/* Левый сайдбар */
.game-sidebar {
  width: 260px;
  background: var(--bg-card);
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
}
.game-header {
  padding: 20px;
  border-bottom: 1px solid var(--border);
}
.back-btn {
  background: none;
  border: none;
  color: var(--text-muted);
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  padding: 0;
  font-size: 0.85rem;
  margin-bottom: 12px;
}
.game-title-short {
  margin: 0;
  font-size: 1.2rem;
  font-weight: 700;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.game-nav {
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.nav-btn {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 10px 12px;
  border: none;
  background: transparent;
  border-radius: var(--radius-md);
  font-size: 0.9rem;
  font-weight: 500;
  color: var(--text-muted);
  cursor: pointer;
  text-decoration: none;
}
.nav-btn:hover {
  background: var(--bg-app);
  color: var(--text-main);
}
.nav-btn.active {
  background: var(--primary-light);
  color: var(--primary);
}

.content-area {
  flex: 1;
  padding: 32px 40px;
}

/* Правый сайдбар с чатом */
.chat-sidebar {
  width: 340px;
  background: var(--bg-card);
  border-left: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
}

.chat-sidebar-header {
  padding: 16px 20px;
  border-bottom: 1px solid var(--border);
  display: flex;
  align-items: center;
  gap: 10px;
}

.chat-sidebar-header h3 {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-main);
}

.workspace-chat {
  flex: 1;
  display: flex;
  flex-direction: column;
  height: 100%;
}
</style>
