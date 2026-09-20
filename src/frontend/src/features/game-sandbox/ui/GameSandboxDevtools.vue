<template>
  <div class="sandbox-devtools">
    <!-- Вкладки панели -->
    <div class="devtools-tabs">
      <button
        class="tab-btn"
        :class="{ active: activeTab === 'logs' }"
        @click="activeTab = 'logs'"
      >
        <Activity class="icon-xs" />
        <span>События</span>
      </button>
      <button
        class="tab-btn"
        :class="{ active: activeTab === 'storage' }"
        @click="activeTab = 'storage'"
      >
        <Database class="icon-xs" />
        <span>Данные</span>
      </button>
      <button
        class="tab-btn"
        :class="{ active: activeTab === 'scenarios' }"
        @click="activeTab = 'scenarios'"
      >
        <Sliders class="icon-xs" />
        <span>Реклама</span>
      </button>
    </div>

    <!-- Контент активной вкладки -->
    <div class="devtools-content">
      <!-- ВКЛАДКА 1: ЛОГ СОБЫТИЙ -->
      <div v-if="activeTab === 'logs'" class="tab-pane logs-pane">
        <div class="logs-filter-bar">
          <div class="logs-search-row">
            <input
              v-model="searchQuery"
              type="text"
              placeholder="Фильтр событий..."
              class="filter-input-sm"
            />
            <button
              class="btn-clear-logs"
              title="Очистить консоль"
              @click="emit('clear-logs')"
            >
              <Eraser class="icon-xs" />
            </button>
          </div>
          <div class="quick-filters">
            <button
              class="tag-filter-btn"
              :class="{ active: currentCategory === 'all' }"
              @click="currentCategory = 'all'"
            >
              Все
            </button>
            <button
              class="tag-filter-btn"
              :class="{ active: currentCategory === 'adv' }"
              @click="currentCategory = 'adv'"
            >
              Реклама
            </button>
            <button
              class="tag-filter-btn"
              :class="{ active: currentCategory === 'data' }"
              @click="currentCategory = 'data'"
            >
              Сейвы
            </button>
            <button
              class="tag-filter-btn"
              :class="{ active: currentCategory === 'purchase' }"
              @click="currentCategory = 'purchase'"
            >
              Покупки
            </button>
          </div>
        </div>

        <div v-if="filteredLogs.length === 0" class="empty-logs">
          <Info class="icon-md text-muted" />
          <p>Ожидание вызовов WelwiseGames SDK...</p>
        </div>

        <div v-else class="logs-list">
          <div
            v-for="log in filteredLogs"
            :key="log.id"
            class="log-item"
            :class="`status-${log.status}`"
            @click="toggleExpandLog(log.id)"
          >
            <div class="log-item-summary">
              <span class="log-time">{{ log.timestamp }}</span>
              <span class="log-dir-badge" :class="log.direction">
                {{ log.direction === 'in' ? 'IN' : log.direction === 'out' ? 'OUT' : 'SYS' }}
              </span>
              <span class="log-action" :title="log.actionName || log.channel">
                {{ log.actionName || log.channel }}
              </span>
              <span class="expand-arrow">
                {{ expandedLogs.has(log.id) ? '▲' : '▼' }}
              </span>
            </div>

            <!-- Детали по клику -->
            <div v-if="expandedLogs.has(log.id)" class="log-details">
              <pre class="json-preview">{{ JSON.stringify(log.payload, null, 2) }}</pre>
            </div>
          </div>
        </div>
      </div>

      <!-- ВКЛАДКА 2: ДАННЫЕ -->
      <div v-else-if="activeTab === 'storage'" class="tab-pane storage-pane">
        <!-- Настройки окружения игрока -->
        <div class="devtools-section">
          <h4 class="section-title">Окружение</h4>
          <div class="form-group-sm">
            <label>Идентификатор:</label>
            <input
              :value="appEnv.playerId"
              type="text"
              class="input-sm"
              placeholder="dev-player-1"
              @input="onPlayerIdInput"
            />
          </div>

          <div class="form-row-sm">
            <div class="form-group-sm">
              <label>Язык:</label>
              <select
                :value="appEnv.language"
                class="select-sm"
                @change="onLanguageChange"
              >
                <option value="ru">Русский (ru)</option>
                <option value="en">English (en)</option>
              </select>
            </div>
            <div class="form-group-sm">
              <label>Устройство:</label>
              <select
                :value="appEnv.deviceType"
                class="select-sm"
                @change="onDeviceTypeChange"
              >
                <option value="desktop">Desktop</option>
                <option value="mobile">Mobile</option>
              </select>
            </div>
          </div>
        </div>

        <!-- Кнопка сброса под блоком окружения отдельной строкой -->
        <div class="storage-action-row">
          <button
            class="btn-reset-storage"
            title="Очистить все сохранения игры"
            @click="emit('clear-storage')"
          >
            <RotateCcw class="icon-xs" />
            <span>Сбросить данные</span>
          </button>
        </div>
      </div>

      <!-- ВКЛАДКА 3: РЕКЛАМА -->
      <div v-else-if="activeTab === 'scenarios'" class="tab-pane scenarios-pane">
        <div class="devtools-section">
          <h4 class="section-title">Эмуляция рекламы</h4>

          <div class="form-group-sm">
            <label>Длительность автозакрытия:</label>
            <div class="radio-pill-group">
              <button
                v-for="sec in [0, 3, 5, 10]"
                :key="sec"
                class="pill-btn"
                :class="{ active: adConfig.autoCloseSeconds === sec }"
                @click="onAutoCloseClick(sec)"
              >
                {{ sec === 0 ? 'Мгновенно' : `${sec} сек` }}
              </button>
            </div>
          </div>

          <div class="checkbox-group-sm">
            <label class="checkbox-label">
              <input
                :checked="adConfig.simulateError"
                type="checkbox"
                class="styled-checkbox"
                @change="onSimulateErrorChange"
              />
              <span>Симулировать ошибку показа</span>
            </label>
            <label class="checkbox-label">
              <input
                :checked="adConfig.rewardedGranted"
                type="checkbox"
                class="styled-checkbox"
                @change="onRewardedGrantedChange"
              />
              <span>Начислять вознаграждение за Rewarded-видео</span>
            </label>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import {
  Activity,
  Database,
  Sliders,
  Eraser,
  RotateCcw,
  Info,
} from 'lucide-vue-next';
import type { SandboxLogEntry } from '../model/useGameSdkSandbox';

interface Props {
  projectId?: string | number | null;
  logs?: SandboxLogEntry[];
  appEnv: Record<string, any>;
  playerStorage?: Record<string, any>;
  adConfig: Record<string, any>;
  purchaseCatalog?: any[];
}

const props = withDefaults(defineProps<Props>(), {
  projectId: null,
  logs: () => [],
  playerStorage: () => ({}),
  purchaseCatalog: () => [],
});

const emit = defineEmits<{
  (e: 'clear-logs'): void;
  (e: 'clear-storage'): void;
  (e: 'update-env', value: Record<string, any>): void;
  (e: 'update-ad-config', value: Record<string, any>): void;
}>();

function onPlayerIdInput(e: Event) {
  const target = e.target as HTMLInputElement;
  emit('update-env', { ...props.appEnv, playerId: target.value });
}

function onLanguageChange(e: Event) {
  const target = e.target as HTMLSelectElement;
  emit('update-env', { ...props.appEnv, language: target.value });
}

function onDeviceTypeChange(e: Event) {
  const target = e.target as HTMLSelectElement;
  emit('update-env', { ...props.appEnv, deviceType: target.value });
}

function onAutoCloseClick(sec: number) {
  emit('update-ad-config', { ...props.adConfig, autoCloseSeconds: sec });
}

function onSimulateErrorChange(e: Event) {
  const target = e.target as HTMLInputElement;
  emit('update-ad-config', { ...props.adConfig, simulateError: target.checked });
}

function onRewardedGrantedChange(e: Event) {
  const target = e.target as HTMLInputElement;
  emit('update-ad-config', { ...props.adConfig, rewardedGranted: target.checked });
}

const activeTab = ref<'logs' | 'storage' | 'scenarios'>('logs');
const searchQuery = ref('');
const currentCategory = ref<'all' | 'adv' | 'data' | 'purchase'>('all');
const expandedLogs = ref(new Set<string>());

function toggleExpandLog(id: string) {
  if (expandedLogs.value.has(id)) {
    expandedLogs.value.delete(id);
  } else {
    expandedLogs.value.add(id);
  }
}


const filteredLogs = computed(() => {
  let list = props.logs;

  if (currentCategory.value === 'adv') {
    list = list.filter((l) => (l.actionName && l.actionName.includes('ADV_')) || l.channel === 'adv-manager');
  } else if (currentCategory.value === 'data') {
    list = list.filter((l) => l.actionName && (l.actionName.includes('PLAYER_DATA') || l.actionName.includes('STORAGE')));
  } else if (currentCategory.value === 'purchase') {
    list = list.filter((l) => (l.actionName && l.actionName.includes('PURCHASE')) || l.channel === 'purchase-manager');
  }

  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase();
    list = list.filter(
      (l) =>
        (l.actionName && l.actionName.toLowerCase().includes(q)) ||
        (l.channel && l.channel.toLowerCase().includes(q)) ||
        JSON.stringify(l.payload).toLowerCase().includes(q)
    );
  }

  return list;
});
</script>

<style scoped>
.sandbox-devtools {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--bg-card, #1c1e24);
  border-left: 1px solid var(--border-color, #2d3139);
  font-family: inherit;
  font-size: 13px;
  overflow: hidden;
}

.devtools-tabs {
  display: flex;
  border-bottom: 1px solid var(--border-color, #2d3139);
  background: var(--bg-surface, #181a1f);
}

.tab-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 8px 6px;
  background: transparent;
  border: none;
  border-bottom: 2px solid transparent;
  color: var(--text-muted, #9ba1ad);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.tab-btn:hover {
  color: var(--text-main, #fff);
  background: var(--bg-hover, #22252c);
}

.tab-btn.active {
  color: var(--primary, #3b82f6);
  border-bottom-color: var(--primary, #3b82f6);
  font-weight: 600;
  background: var(--bg-card, #1c1e24);
}

.devtools-content {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}

.tab-pane {
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  flex: 1;
}

.logs-filter-bar {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.logs-search-row {
  display: flex;
  align-items: center;
  gap: 6px;
}

.btn-clear-logs {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  background: var(--bg-input, #131418);
  border: 1px solid var(--border-color, #2d3139);
  border-radius: 6px;
  color: var(--text-muted, #9ba1ad);
  cursor: pointer;
  flex-shrink: 0;
  transition: all 0.15s ease;
}

.btn-clear-logs:hover {
  background: var(--bg-hover, #2b2f38);
  border-color: var(--primary, #3b82f6);
  color: var(--primary, #3b82f6);
}

.filter-input-sm {
  width: 100%;
  padding: 6px 10px;
  border-radius: 6px;
  background: var(--bg-input, #131418);
  border: 1px solid var(--border-color, #2d3139);
  color: var(--text-main, #fff);
  font-size: 12px;
  outline: none;
}

.filter-input-sm:focus {
  border-color: var(--primary, #3b82f6);
}

.quick-filters {
  display: flex;
  gap: 4px;
}

.tag-filter-btn {
  background: var(--bg-input, #131418);
  border: 1px solid var(--border-color, #2d3139);
  color: var(--text-muted, #9ba1ad);
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  cursor: pointer;
}

.tag-filter-btn.active {
  background: var(--primary-light, rgba(59, 130, 246, 0.15));
  color: var(--primary, #3b82f6);
  border-color: var(--primary, #3b82f6);
}

.logs-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex: 1;
  overflow-y: auto;
}

.empty-logs {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 10px;
  color: var(--text-muted, #9ba1ad);
  text-align: center;
  gap: 8px;
}

.log-item {
  background: var(--bg-surface, #181a1f);
  border: 1px solid var(--border-color, #2d3139);
  border-radius: 6px;
  padding: 6px 8px;
  cursor: pointer;
  transition: background 0.15s ease;
}

.log-item:hover {
  background: var(--bg-hover, #22252c);
}

.log-item.status-error {
  border-left: 3px solid #ef4444;
}

.log-item.status-success {
  border-left: 3px solid #10b981;
}

.log-item.status-warn {
  border-left: 3px solid #f59e0b;
}

.log-item-summary {
  display: flex;
  align-items: center;
  gap: 6px;
}

.log-time {
  font-size: 10px;
  color: var(--text-muted, #7e8494);
  font-family: monospace;
}

.log-dir-badge {
  font-size: 9px;
  font-weight: 700;
  padding: 1px 4px;
  border-radius: 3px;
}

.log-dir-badge.in {
  background: rgba(59, 130, 246, 0.2);
  color: #60a5fa;
}

.log-dir-badge.out {
  background: rgba(16, 185, 129, 0.2);
  color: #34d399;
}

.log-dir-badge.system {
  background: rgba(245, 158, 11, 0.2);
  color: #fbbf24;
}

.log-action {
  font-family: monospace;
  font-size: 11px;
  font-weight: 600;
  color: var(--text-main, #f0f2f5);
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.expand-arrow {
  font-size: 9px;
  color: var(--text-muted, #7e8494);
}

.log-details {
  margin-top: 6px;
  padding-top: 6px;
  border-top: 1px solid var(--border-color, #2d3139);
}

.json-preview {
  margin: 0;
  padding: 6px;
  background: var(--bg-input, #101114);
  border-radius: 4px;
  font-size: 11px;
  color: #a5b4fc;
  max-height: 200px;
  overflow: auto;
}

.devtools-section {
  background: var(--bg-surface, #181a1f);
  border: 1px solid var(--border-color, #2d3139);
  border-radius: 8px;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.section-head-between {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.section-title {
  margin: 0;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-main, #f0f2f5);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.section-desc {
  margin: 0;
  font-size: 11px;
  color: var(--text-muted, #9ba1ad);
}

.form-group-sm {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.form-group-sm label {
  font-size: 11px;
  color: var(--text-muted, #9ba1ad);
}

.form-row-sm {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
}

.input-sm,
.select-sm {
  width: 100%;
  padding: 6px 8px;
  background: var(--bg-input, #131418);
  border: 1px solid var(--border-color, #2d3139);
  border-radius: 6px;
  color: var(--text-main, #fff);
  font-size: 12px;
  outline: none;
}

.storage-action-row {
  display: flex;
  align-items: center;
  padding-top: 4px;
}

.btn-reset-storage {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: rgba(239, 68, 68, 0.12);
  color: #f87171;
  border: 1px solid rgba(239, 68, 68, 0.25);
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-reset-storage:hover {
  background: rgba(239, 68, 68, 0.22);
  border-color: rgba(239, 68, 68, 0.45);
  color: #fca5a5;
}

.radio-pill-group {
  display: flex;
  gap: 4px;
}

.pill-btn {
  flex: 1;
  padding: 4px 6px;
  background: var(--bg-input, #131418);
  border: 1px solid var(--border-color, #2d3139);
  color: var(--text-muted, #9ba1ad);
  border-radius: 4px;
  font-size: 11px;
  cursor: pointer;
}

.pill-btn.active {
  background: var(--primary-light, rgba(59, 130, 246, 0.15));
  border-color: var(--primary, #3b82f6);
  color: var(--primary, #3b82f6);
  font-weight: 600;
}

.checkbox-group-sm {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.checkbox-label {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--text-main, #f0f2f5);
  cursor: pointer;
  user-select: none;
}

.styled-checkbox {
  appearance: none;
  -webkit-appearance: none;
  width: 16px;
  height: 16px;
  border: 1px solid var(--border-color, #383c46);
  border-radius: 4px;
  background: var(--bg-input, #131418);
  cursor: pointer;
  outline: none;
  transition: all 0.15s ease;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  position: relative;
  margin: 0;
}

.styled-checkbox:hover {
  border-color: var(--primary, #3b82f6);
}

.styled-checkbox:focus-visible {
  border-color: var(--primary, #3b82f6);
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.25);
}

.styled-checkbox:checked {
  background: var(--primary, #3b82f6);
  border-color: var(--primary, #3b82f6);
}

.styled-checkbox:checked::after {
  content: '';
  width: 4px;
  height: 8px;
  border: solid #ffffff;
  border-width: 0 2px 2px 0;
  transform: rotate(45deg) translate(-1px, -1px);
}
</style>
