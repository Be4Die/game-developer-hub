<template>
  <div class="servers-layout">
    <div class="servers-subnav">
      <div class="subnav-left">
        <router-link
          :to="`/projects/${gameId}/servers`"
          class="subnav-btn"
          exact-active-class="active"
        >
          <LayoutDashboard class="icon-sm" /> {{ t('servers.tabs.overview') }}
        </router-link>
        <router-link
          :to="`/projects/${gameId}/servers/builds`"
          class="subnav-btn"
          active-class="active"
        >
          <Package class="icon-sm" /> {{ t('servers.tabs.builds') }}
        </router-link>
        <router-link
          :to="`/projects/${gameId}/servers/instances`"
          class="subnav-btn"
          active-class="active"
        >
          <Play class="icon-sm" /> {{ t('servers.tabs.instances') }}
          <span v-if="serverLayoutContext.maxInstances > 0" class="subnav-counter">
            {{ serverLayoutContext.activeCount }} / {{ serverLayoutContext.maxInstances }}
          </span>
        </router-link>
      </div>

      <div class="subnav-actions">
        <!-- Вкладка сборок: кнопка Загрузить сборку -->
        <button
          v-if="isBuildsTab && serverLayoutContext.triggerUpload"
          class="btn-primary btn-sm"
          @click="serverLayoutContext.triggerUpload"
        >
          <Upload class="icon-sm" />
          <span>{{ t('servers.uploadBuild') }}</span>
        </button>

        <!-- Вкладка инстансов: фильтр статуса + Запустить экземпляр -->
        <template v-else-if="isInstancesTab">
          <select
            v-model="serverLayoutContext.statusFilter"
            class="filter-select-sm"
          >
            <option value="all">{{ t('projects.allStatuses') }}</option>
            <option value="starting">Starting</option>
            <option value="running">Running</option>
            <option value="stopping">Stopping</option>
            <option value="stopped">Stopped</option>
            <option value="crashed">Crashed</option>
          </select>
          <button
            v-if="serverLayoutContext.triggerStart"
            class="btn-primary btn-sm"
            @click="serverLayoutContext.triggerStart"
          >
            <Play class="icon-sm" />
            <span>{{ t('servers.startInstance') }}</span>
          </button>
        </template>
      </div>
    </div>
    <div class="servers-content">
      <router-view />
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, computed, provide } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { LayoutDashboard, Package, Play, Upload } from 'lucide-vue-next';

const { t } = useI18n();
const route = useRoute();

defineProps<{
  gameId: string | number;
}>();

const isBuildsTab = computed(() => route.name === 'server-builds' || route.path.endsWith('/builds'));
const isInstancesTab = computed(() => route.name === 'server-instances' || route.path.endsWith('/instances'));

const serverLayoutContext = reactive({
  triggerUpload: null as (() => void) | null,
  triggerStart: null as (() => void) | null,
  statusFilter: 'all',
  activeCount: 0,
  maxInstances: 0,
});

provide('serverLayoutContext', serverLayoutContext);
</script>

<style scoped>
.servers-layout {
  display: flex;
  flex-direction: column;
}
.servers-subnav {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  padding: 8px 12px;
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  margin: 12px 32px 0;
  background: var(--bg-card);
  flex-wrap: wrap;
}
.subnav-left {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-wrap: wrap;
}
.subnav-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}
.subnav-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border: none;
  background: transparent;
  border-radius: var(--radius-md);
  font-size: 0.88rem;
  font-weight: 500;
  color: var(--text-muted);
  cursor: pointer;
  text-decoration: none;
  transition: 0.15s;
}
.subnav-btn:hover {
  background: var(--bg-hover);
  color: var(--text-main);
}
.subnav-btn.active {
  background: #eff6ff;
  color: var(--primary);
  font-weight: 600;
}
[data-theme='dark'] .subnav-btn.active {
  background: #1e3a5f;
}

.subnav-counter {
  display: inline-flex;
  align-items: center;
  padding: 2px 7px;
  border-radius: 10px;
  font-size: 0.75rem;
  font-weight: 600;
  background: var(--bg-app);
  color: var(--text-muted);
  border: 1px solid var(--border);
  margin-left: 4px;
}
.subnav-btn.active .subnav-counter {
  background: rgba(88, 166, 255, 0.15);
  color: var(--primary);
  border-color: rgba(88, 166, 255, 0.3);
}

.filter-select-sm {
  padding: 6px 10px;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  background: var(--bg-app);
  color: var(--text-main);
  font-size: 0.82rem;
  outline: none;
  cursor: pointer;
  font-family: inherit;
  transition: border-color 0.15s;
}
.filter-select-sm:focus {
  border-color: var(--primary);
}

.servers-content {
  flex: 1;
  padding: 16px 32px 24px;
}
</style>
