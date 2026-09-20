<template>
  <div class="tab-content tab-fade-in sandbox-page-layout">
    <!-- Селектор версий сборок для тестирования (если их больше одной) -->
    <div v-if="buildsList.length > 1" class="sandbox-page-header">
      <div class="version-selector-group">
        <label class="selector-label">Сборка для запуска:</label>
        <div class="select-wrapper">
          <select v-model="selectedVersion" class="version-select">
            <option
              v-for="b in buildsList"
              :key="b.version"
              :value="b.version"
            >
              v{{ b.version }} {{ b.version === activeVersion ? '(Активный черновик)' : '' }}
            </option>
          </select>
        </div>
      </div>
    </div>

    <!-- Встроенный плеер песочницы -->
    <div class="sandbox-player-wrapper">
      <GameSandboxPlayer
        :game-url="currentPlayUrl"
        :project-id="projectId"
        :show-devtools="true"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, inject, onMounted, type Ref } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { GameSandboxPlayer } from '@/features/game-sandbox';
import { listClientBuilds } from '@/entities/build';
import type { Project, ClientBuild } from '@/shared/types';

const { t } = useI18n();
const route = useRoute();
const project = inject<Ref<Project | null> | null>('project', null);

const projectId = computed<string | number>(() => {
  return (route.params.id as string) || project?.value?.id || '';
});

const activeVersion = computed<string>(() => {
  return (
    project?.value?.draft?.active_build_version ||
    project?.value?.active_build_version ||
    '1.0.0'
  );
});

const selectedVersion = ref<string>('');
const buildsList = ref<ClientBuild[]>([]);

// Загружаем список доступных сборок игры
async function loadBuilds(): Promise<void> {
  if (!projectId.value) return;
  try {
    const list = await listClientBuilds(projectId.value);
    buildsList.value = list || [];
    if (!selectedVersion.value && buildsList.value.length > 0) {
      selectedVersion.value = activeVersion.value || buildsList.value[0].version;
    }
  } catch {
    // Тихо игнорируем ошибку получения списка сборок
  }
}

onMounted(() => {
  loadBuilds();
});

const currentPlayUrl = computed<string>(() => {
  const id = projectId.value;
  if (!id) return '';

  // Если выбрана конкретная версия из списка
  if (selectedVersion.value && selectedVersion.value !== activeVersion.value) {
    return `/games/${id}/versions/${encodeURIComponent(selectedVersion.value)}/index.html`;
  }

  // По умолчанию: dev-симлинк черновика проекта
  return (
    project?.value?.draft?.dev_url ||
    project?.value?.dev_url ||
    `/games/${id}/dev/index.html`
  );
});
</script>

<style scoped>
.sandbox-page-layout {
  display: flex;
  flex-direction: column;
  height: 100%;
  width: 100%;
  padding: 0;
  margin: 0;
  gap: 0;
  box-sizing: border-box;
  overflow: hidden;
}

.sandbox-page-header {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  padding: 8px 16px;
  background: var(--bg-card, #1c1e24);
  border-bottom: 1px solid var(--border-color, #2d3139);
  flex-shrink: 0;
}

.version-selector-group {
  display: flex;
  align-items: center;
  gap: 10px;
  background: var(--bg-surface, #131418);
  border: 1px solid var(--border-color, #2d3139);
  padding: 4px 10px;
  border-radius: 6px;
}

.selector-label {
  font-size: 12px;
  color: var(--text-muted, #9ba1ad);
}

.version-select {
  padding: 3px 8px;
  background: var(--bg-surface, #131418);
  border: 1px solid var(--border-color, #2d3139);
  border-radius: 4px;
  color: var(--text-main, #fff);
  font-size: 12px;
  outline: none;
  cursor: pointer;
}

.sandbox-player-wrapper {
  flex: 1;
  min-height: 0;
  height: 100%;
  width: 100%;
  border-radius: 0;
  overflow: hidden;
  box-shadow: none;
  display: flex;
  flex-direction: column;
}
</style>
