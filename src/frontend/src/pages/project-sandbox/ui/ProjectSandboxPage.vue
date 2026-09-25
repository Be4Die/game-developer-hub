<template>
  <div class="tab-content tab-fade-in sandbox-page-layout">
    <!-- Встроенный плеер песочницы -->
    <div class="sandbox-player-wrapper">
      <GameSandboxPlayer
        :game-url="currentPlayUrl"
        :project-id="projectId"
        :show-devtools="true"
      >
        <template #toolbar-left>
          <!-- Селектор версий сборок для тестирования (если их больше одной) -->
          <div v-if="buildsList.length > 1" class="version-selector-group">
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
        </template>
      </GameSandboxPlayer>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, inject, onMounted, type Ref } from 'vue';
import { useRoute } from 'vue-router';
import { GameSandboxPlayer } from '@/features/game-sandbox';
import { listClientBuilds } from '@/entities/build';
import type { Project, ClientBuild } from '@/shared/types';

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

.version-selector-group {
  display: flex;
  align-items: center;
  gap: 8px;
  background: var(--bg-surface, #131418);
  border: 1px solid var(--border-color, #2d3139);
  padding: 2px 8px;
  border-radius: 6px;
  height: 28px;
  box-sizing: border-box;
}

.selector-label {
  font-size: 11px;
  color: var(--text-muted, #9ba1ad);
  white-space: nowrap;
  user-select: none;
}

.select-wrapper {
  display: flex;
  align-items: center;
}

.version-select {
  padding: 2px 6px;
  background: var(--bg-card, #1a1c22);
  border: 1px solid var(--border-color, #2d3139);
  border-radius: 4px;
  color: var(--text-main, #fff);
  font-size: 11px;
  outline: none;
  cursor: pointer;
  height: 22px;
  transition: border-color 0.15s ease;
}

.version-select:hover {
  border-color: var(--border-color-hover, #424754);
}

.version-select:focus {
  border-color: var(--primary, #3b82f6);
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
