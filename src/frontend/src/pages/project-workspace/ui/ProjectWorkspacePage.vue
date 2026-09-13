<template>
  <div class="game-workspace">
    <!-- ЛЕВОЕ МЕНЮ ИГРЫ -->
    <aside class="game-sidebar">
      <div class="game-header">
        <div class="game-identity-row">
          <div class="game-icon-box">
            <img v-if="projectIconUrl" :src="projectIconUrl" alt="Icon" class="game-icon-img" />
            <div v-else class="game-icon-mock">
              <span>Draft</span>
            </div>
          </div>
          <div class="game-title-wrap">
            <h2 class="game-title-short" :title="projectTitle">
              {{ projectTitle }}
            </h2>
            <div v-if="project" class="role-access-badge-wrap">
              <span v-if="isOwner" class="badge-role-owner">
                {{ t('access.statuses.owner') }}
              </span>
              <span
                v-else
                class="badge-role-collab"
                :title="collaboratorPermissionsText"
              >
                <Users class="icon-xs" />
                <span>{{ t('access.statuses.collaborator') }}</span>
              </span>
            </div>
          </div>
        </div>

        <div class="game-links-row">
          <button v-if="isPublished" class="btn-prod-link" @click="openProdGame">
            <ExternalLink class="icon-xs" />
            <span>Игра (Prod)</span>
          </button>
          <button class="btn-dev-link" @click="openDevGame">
            <ExternalLink class="icon-xs" />
            <span>{{ t('projectDraft.openGameDev') }}</span>
          </button>
        </div>
      </div>

      <nav class="game-nav">
        <router-link :to="`/projects/${id}/draft`" class="nav-btn" active-class="active">
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
          v-if="canManageServers && isOnline"
          :to="`/projects/${id}/servers`"
          class="nav-btn"
          active-class="active"
        >
          <Server class="icon-sm" /> {{ t('projectWorkspace.serversTab') }}
        </router-link>
        <router-link
          v-if="canViewStats"
          :to="`/projects/${id}/stats`"
          class="nav-btn"
          active-class="active"
        >
          <BarChart2 class="icon-sm" /> {{ t('projectWorkspace.statsTab') }}
        </router-link>
        <router-link
          v-if="isOwner"
          :to="`/projects/${id}/access`"
          class="nav-btn"
          active-class="active"
        >
          <Users class="icon-sm" /> {{ t('projectWorkspace.accessTab') }}
        </router-link>
      </nav>

      <!-- Футер сайдбара: Сохранить и Отправить на модерацию -->
      <div class="sidebar-footer">
        <button
          class="btn-sidebar-save"
          :disabled="
            draftActions.isSaving ||
            draftActions.isSubmitting ||
            (!canEditInfo && !canUploadMedia && !canUploadBuild)
          "
          @click="handleSidebarSave"
        >
          <Loader2 v-if="draftActions.isSaving" class="icon-xs spin" />
          <Save v-else class="icon-xs" />
          <span>{{ draftActions.isSaving ? t('common.saving') : t('common.save') }}</span>
        </button>

        <button
          class="btn-sidebar-submit"
          :disabled="
            draftActions.isSubmitting ||
            draftActions.isUnderReview ||
            !canSubmitModeration
          "
          @click="handleSidebarSubmit"
        >
          <Loader2 v-if="draftActions.isSubmitting" class="icon-xs spin" />
          <Send v-else class="icon-xs" />
          <span>
            {{
              draftActions.isSubmitting
                ? t('projectDraft.sending')
                : draftActions.isUnderReview
                  ? t('projects.moderation')
                  : isPublished
                    ? t('projectDraft.sendUpdateToModeration')
                    : t('projectDraft.sendToModeration')
            }}
          </span>
        </button>
      </div>
    </aside>

    <!-- ЦЕНТР (Подгружает табы) -->
    <main class="content-area">
      <router-view />
    </main>

    <!-- ПРАВЫЙ САЙДБАР: Чат проекта -->
    <aside class="chat-sidebar">
      <ProjectChat :project-id="id" class="workspace-chat" />
    </aside>
  </div>
</template>

<script setup>
import { ref, computed, watch, provide } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import {
  BarChart2,
  PenTool,
  CheckCircle,
  Server,
  ExternalLink,
  Save,
  Send,
  Loader2,
  Users,
} from 'lucide-vue-next';
import { getProject, getMediaUrl, permissionLabel } from '@/entities/project';
import { ProjectChat } from '@/entities/moderation';

const { t } = useI18n();
const props = defineProps({
  id: { type: [String, Number], default: null },
});
const route = useRoute();
const router = useRouter();

// ─── Project data (shared with child tabs) ───────────────────
const project = ref(null);
provide('project', project);

const isOwner = computed(() => project.value?.is_owner !== false);
const permissions = computed(() => project.value?.current_user_permissions || []);
const canEditInfo = computed(() => isOwner.value || permissions.value.includes('PERM_EDIT_INFO'));
const canUploadMedia = computed(() => isOwner.value || permissions.value.includes('PERM_UPLOAD_MEDIA'));
const canUploadBuild = computed(() => isOwner.value || permissions.value.includes('PERM_UPLOAD_BUILD'));
const canViewStats = computed(() => isOwner.value || permissions.value.includes('PERM_VIEW_STATS'));
const canManageServers = computed(() => isOwner.value || permissions.value.includes('PERM_MANAGE_SERVERS'));
const canSubmitModeration = computed(
  () => isOwner.value || permissions.value.includes('PERM_SUBMIT_MODERATION')
);

const isOnline = computed(() => {
  return project.value?.draft?.is_online ?? project.value?.is_online ?? false;
});

const collaboratorPermissionsText = computed(() => {
  if (isOwner.value) return '';
  return permissions.value.map((p) => permissionLabel(p)).join(', ');
});

const draftActions = ref({
  save: null,
  submit: null,
  isSaving: false,
  isSubmitting: false,
  isUnderReview: false,
  isApproved: false,
});
provide('draftActions', draftActions);

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

const projectIconUrl = computed(() => {
  const path = project.value?.icon_path || project.value?.draft?.icon_path;
  if (!path) return null;
  return getMediaUrl(path);
});

const isPublished = computed(() => {
  return (
    project.value?.status === 3 ||
    project.value?.status === 'PROJECT_STATUS_PUBLISHED' ||
    !!project.value?.release
  );
});

// Редирект с "published" на "draft", если проект загружен и не опубликован
watch(
  () => [route.name, project.value?.status, project.value?.release],
  ([name]) => {
    if (name === 'published' && project.value && !isPublished.value) {
      router.replace(`/projects/${props.id}/draft`);
    }
  },
  { immediate: true }
);

// Редирект с табов stats/servers, если у участника нет соответствующих прав
watch(
  () => [route.name, route.path, project.value],
  () => {
    if (!project.value) return;
    if (route.name === 'stats' && !canViewStats.value) {
      router.replace(`/projects/${props.id}/draft`);
    }
    if ((route.name === 'servers' || route.path.includes('/servers')) && (!canManageServers.value || !isOnline.value)) {
      router.replace(`/projects/${props.id}/draft`);
    }
  }
);

function openDevGame() {
  const url =
    project.value?.draft?.dev_url || project.value?.dev_url || `/games/${props.id}/dev/index.html`;
  window.open(url, '_blank');
}

function openProdGame() {
  const url =
    project.value?.release?.prod_url ||
    project.value?.prod_url ||
    `/games/${props.id}/prod/index.html`;
  window.open(url, '_blank');
}

async function handleSidebarSave() {
  if (draftActions.value.save) {
    draftActions.value.isSaving = true;
    try {
      await draftActions.value.save();
    } finally {
      draftActions.value.isSaving = false;
    }
  }
}

async function handleSidebarSubmit() {
  if (!canSubmitModeration.value) return;
  if (draftActions.value.submit) {
    await draftActions.value.submit();
  }
}
</script>

<style scoped>
.role-access-badge-wrap {
  margin-top: 4px;
}

.badge-role-owner {
  display: inline-block;
  padding: 2px 6px;
  font-size: 11px;
  font-weight: 600;
  border-radius: var(--radius-sm);
  background: var(--primary-light);
  color: var(--primary);
}

.badge-role-collab {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 6px;
  font-size: 11px;
  font-weight: 600;
  border-radius: var(--radius-sm);
  background: var(--bg-tertiary);
  color: var(--text-muted);
  border: 1px solid var(--border);
  cursor: help;
}

.game-workspace {
  display: flex;
  height: calc(100vh - 60px);
  max-height: calc(100vh - 60px);
  overflow: hidden;
  background: var(--bg-app);
}

/* Левый сайдбар */
.game-sidebar {
  width: 280px;
  height: 100%;
  background: var(--bg-card);
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  box-sizing: border-box;
}

.game-header {
  padding: 16px 16px;
  border-bottom: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.game-identity-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.game-icon-box {
  width: 42px;
  height: 42px;
  border-radius: var(--radius-sm, 6px);
  overflow: hidden;
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.game-icon-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.game-icon-mock {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-tertiary);
  color: var(--text-tertiary);
  font-size: 0.72rem;
  font-weight: 700;
  text-transform: uppercase;
}

.game-title-wrap {
  flex: 1;
  min-width: 0;
}

.game-title-short {
  margin: 0;
  font-size: 1.05rem;
  font-weight: 700;
  color: var(--text-main);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.game-links-row {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
}

.btn-dev-link,
.btn-prod-link {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: var(--radius-sm, 6px);
  font-weight: 500;
  font-size: 0.85rem;
  cursor: pointer;
  transition: all 0.15s ease;
  width: 100%;
}

.btn-dev-link {
  border: 1px solid var(--border);
  background: var(--bg-secondary);
  color: var(--text-main);
}

.btn-dev-link:hover {
  border-color: var(--primary);
  color: var(--primary);
  background: var(--bg-card);
}

.btn-prod-link {
  border: 1px solid rgba(16, 185, 129, 0.4);
  background: rgba(16, 185, 129, 0.1);
  color: #10b981;
}

.btn-prod-link:hover {
  border-color: #10b981;
  background: rgba(16, 185, 129, 0.18);
}

.game-nav {
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex: 1;
  overflow-y: auto;
}

.nav-btn {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 10px 12px;
  border: none;
  background: transparent;
  border-radius: var(--radius-sm, 6px);
  font-size: 0.9rem;
  font-weight: 500;
  color: var(--text-muted);
  cursor: pointer;
  text-decoration: none;
  transition: all 0.15s ease;
}

.nav-btn:hover {
  background: var(--bg-hover);
  color: var(--text-main);
}

.nav-btn.active {
  background: var(--primary-light);
  color: var(--primary);
  font-weight: 600;
}

/* Футер сайдбара */
.sidebar-footer {
  margin-top: auto;
  padding: 14px 16px;
  border-top: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  gap: 8px;
  background: var(--bg-card);
  flex-shrink: 0;
}

.btn-sidebar-save {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  width: 100%;
  height: 38px;
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm, 6px);
  color: var(--text-main);
  font-size: 0.88rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-sidebar-save:hover:not(:disabled) {
  border-color: var(--border-secondary);
  background: var(--bg-hover);
}

.btn-sidebar-submit {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  width: 100%;
  height: 38px;
  background: var(--primary);
  border: none;
  border-radius: var(--radius-sm, 6px);
  color: #fff;
  font-size: 0.88rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-sidebar-submit:hover:not(:disabled) {
  opacity: 0.92;
}

.btn-sidebar-submit:disabled,
.btn-sidebar-save:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Центральная область */
.content-area {
  flex: 1;
  height: 100%;
  overflow-y: auto;
  padding: 28px 40px;
  box-sizing: border-box;
}

/* Правый сайдбар с чатом */
.chat-sidebar {
  width: 480px;
  min-width: 400px;
  max-width: 560px;
  height: 100%;
  background: var(--bg-card);
  border-left: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  overflow: hidden;
  box-sizing: border-box;
  transition: width 0.2s ease;
}

@media (min-width: 1600px) {
  .chat-sidebar {
    width: 520px;
  }
}

@media (max-width: 1280px) {
  .chat-sidebar {
    width: 420px;
  }
}

.workspace-chat {
  flex: 1;
  height: 100%;
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>
