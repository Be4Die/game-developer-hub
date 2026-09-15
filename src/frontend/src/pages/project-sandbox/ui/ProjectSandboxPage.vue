<template>
  <div class="tab-content tab-fade-in sandbox-page-layout">
    <!-- Шапка страницы песочницы -->
    <div class="sandbox-page-header">
      <div class="header-main">
        <div class="title-with-badge">
          <h1 class="page-title">{{ t('projectSandbox.title') || 'Песочница тестирования игры' }}</h1>
          <span class="version-tag">
            Версия: <strong>v{{ activeVersion || '1.0.0' }}</strong>
          </span>
        </div>
        <p class="page-subtitle">
          {{ t('projectSandbox.subtitle') || 'Изолированная среда выполнения игры с поддержкой WelwiseGames JS-SDK. Все события и вызовы API эмулируются локально в консоли.' }}
        </p>
      </div>

      <!-- Селектор версий сборок для тестирования -->
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

<script setup>
import { ref, computed, inject, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { GameSandboxPlayer } from '@/features/game-sandbox';
import { listClientBuilds } from '@/entities/build';

const { t } = useI18n();
const route = useRoute();
const project = inject('project', ref(null));

const projectId = computed(() => {
  return route.params.id || project.value?.id || '';
});

const activeVersion = computed(() => {
  return (
    project.value?.draft?.active_build_version ||
    project.value?.active_build_version ||
    '1.0.0'
  );
});

const selectedVersion = ref('');
const buildsList = ref([]);

// Загружаем список доступных сборок игры
async function loadBuilds() {
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

const currentPlayUrl = computed(() => {
  const id = projectId.value;
  if (!id) return '';

  // Если выбрана конкретная версия из списка
  if (selectedVersion.value && selectedVersion.value !== activeVersion.value) {
    return `/games/${id}/versions/${encodeURIComponent(selectedVersion.value)}/index.html`;
  }

  // По умолчанию: dev-симлинк черновика проекта
  return (
    project.value?.draft?.dev_url ||
    project.value?.dev_url ||
    `/games/${id}/dev/index.html`
  );
});
</script>

<style scoped>
.sandbox-page-layout {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 72px);
  padding: 16px 20px 20px;
  gap: 16px;
  box-sizing: border-box;
}

.sandbox-page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 20px;
  flex-wrap: wrap;
}

.header-main {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.title-with-badge {
  display: flex;
  align-items: center;
  gap: 12px;
}

.page-title {
  margin: 0;
  font-size: 20px;
  font-weight: 700;
  color: var(--text-main, #f0f2f5);
}

.version-tag {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  background: var(--bg-hover, #2b2f38);
  border: 1px solid var(--border-color, #383c46);
  border-radius: 6px;
  font-size: 12px;
  color: var(--text-muted, #9ba1ad);
}

.version-tag strong {
  color: var(--primary, #3b82f6);
}

.page-subtitle {
  margin: 0;
  font-size: 13px;
  color: var(--text-muted, #9ba1ad);
  max-width: 780px;
  line-height: 1.4;
}

.version-selector-group {
  display: flex;
  align-items: center;
  gap: 10px;
  background: var(--bg-card, #1c1e24);
  border: 1px solid var(--border-color, #2d3139);
  padding: 6px 12px;
  border-radius: 8px;
}

.selector-label {
  font-size: 12px;
  color: var(--text-muted, #9ba1ad);
}

.version-select {
  padding: 4px 8px;
  background: var(--bg-surface, #131418);
  border: 1px solid var(--border-color, #2d3139);
  border-radius: 6px;
  color: var(--text-main, #fff);
  font-size: 12px;
  outline: none;
  cursor: pointer;
}

.sandbox-player-wrapper {
  flex: 1;
  min-height: 0;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.25);
}
</style>
