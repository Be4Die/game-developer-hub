<template>
  <div class="sandbox-devtools">
    <!-- Заголовок панели -->
    <div class="devtools-header">
      <div class="devtools-title-group">
        <Terminal class="icon-sm text-primary" />
        <span class="devtools-title">SDK DevTools</span>
        <span class="badge-count">{{ logs.length }}</span>
      </div>
      <div class="header-actions">
        <button
          v-if="activeTab === 'logs'"
          class="btn-icon-xs"
          title="Очистить лог"
          @click="$emit('clear-logs')"
        >
          <Trash2 class="icon-xs" />
        </button>
      </div>
    </div>

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
        <span>Игрок и Сейвы</span>
      </button>
      <button
        class="tab-btn"
        :class="{ active: activeTab === 'scenarios' }"
        @click="activeTab = 'scenarios'"
      >
        <Sliders class="icon-xs" />
        <span>Реклама и QA</span>
      </button>
    </div>

    <!-- Контент активной вкладки -->
    <div class="devtools-content">
      <!-- ВКЛАДКА 1: ЛОГ СОБЫТИЙ -->
      <div v-if="activeTab === 'logs'" class="tab-pane logs-pane">
        <div class="logs-filter-bar">
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Фильтр событий..."
            class="filter-input-sm"
          />
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

      <!-- ВКЛАДКА 2: ИГРОК И СЕЙВЫ -->
      <div v-else-if="activeTab === 'storage'" class="tab-pane storage-pane">
        <!-- Настройки окружения игрока -->
        <div class="devtools-section">
          <h4 class="section-title">Окружение (AppEnvironment)</h4>
          <div class="form-group-sm">
            <label>PlayerId (Идентификатор игрока):</label>
            <input
              :value="appEnv.playerId"
              type="text"
              class="input-sm"
              placeholder="dev-player-1"
              @input="$emit('update-env', { ...appEnv, playerId: $event.target.value })"
            />
          </div>

          <div class="form-row-sm">
            <div class="form-group-sm">
              <label>Язык (Language):</label>
              <select
                :value="appEnv.language"
                class="select-sm"
                @change="$emit('update-env', { ...appEnv, language: $event.target.value })"
              >
                <option value="ru">Русский (ru)</option>
                <option value="en">English (en)</option>
              </select>
            </div>
            <div class="form-group-sm">
              <label>Тип устройства:</label>
              <select
                :value="appEnv.deviceType"
                class="select-sm"
                @change="$emit('update-env', { ...appEnv, deviceType: $event.target.value })"
              >
                <option value="desktop">Desktop</option>
                <option value="mobile">Mobile</option>
              </select>
            </div>
          </div>
        </div>

        <!-- Облачные сохранения -->
        <div class="devtools-section">
          <div class="section-head-between">
            <h4 class="section-title">Облачные сохранения (Cloud Saves)</h4>
            <button
              class="btn-danger-xs"
              title="Очистить все сохранения игры"
              @click="$emit('clear-storage')"
            >
              <RotateCcw class="icon-xs" />
              <span>Сбросить сейв</span>
            </button>
          </div>

          <div v-if="hasSavedData" class="storage-viewer">
            <pre class="json-preview">{{ JSON.stringify(playerStorage, null, 2) }}</pre>
          </div>
          <div v-else class="empty-storage-notice">
            <span>Сохранения отсутствуют (игра ещё не вызывала player.setData)</span>
          </div>
        </div>
      </div>

      <!-- ВКЛАДКА 3: РЕКЛАМА И СЦЕНАРИИ QA -->
      <div v-else-if="activeTab === 'scenarios'" class="tab-pane scenarios-pane">
        <div class="devtools-section">
          <h4 class="section-title">Эмуляция рекламы (Ad Simulator)</h4>

          <div class="form-group-sm">
            <label>Длительность автозакрытия:</label>
            <div class="radio-pill-group">
              <button
                v-for="sec in [0, 3, 5, 10]"
                :key="sec"
                class="pill-btn"
                :class="{ active: adConfig.autoCloseSeconds === sec }"
                @click="$emit('update-ad-config', { ...adConfig, autoCloseSeconds: sec })"
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
                @change="$emit('update-ad-config', { ...adConfig, simulateError: $event.target.checked })"
              />
              <span>Симулировать ошибку показа (AdBlock / Network Error)</span>
            </label>
            <label class="checkbox-label">
              <input
                :checked="adConfig.rewardedGranted"
                type="checkbox"
                @change="$emit('update-ad-config', { ...adConfig, rewardedGranted: $event.target.checked })"
              />
              <span>Начислять вознаграждение за Rewarded-видео</span>
            </label>
          </div>
        </div>

        <div class="devtools-section">
          <div class="section-head-between">
            <h4 class="section-title">Каталог товаров (In-App Purchases)</h4>
            <router-link
              v-if="projectId"
              :to="{ name: 'project-purchases', params: { id: projectId } }"
              class="link-action-xs"
              target="_blank"
            >
              <ExternalLink class="icon-xs" />
              <span>Управление</span>
            </router-link>
          </div>
          <p class="section-desc">
            {{ purchaseCatalog.length > 0
              ? `Загружено ${purchaseCatalog.length} товаров из проекта для метода purchases.getAvailableItems().`
              : 'В каталоге проекта пока нет товаров.'
            }}
          </p>
          <div v-if="purchaseCatalog.length > 0" class="catalog-list">
            <div v-for="item in purchaseCatalog" :key="item.itemId" class="catalog-item">
              <div class="catalog-item-info">
                <div class="item-name">{{ item.name }}</div>
                <div class="item-id-sub text-muted">ID: {{ item.itemId }}</div>
              </div>
              <div class="item-price">{{ item.priceCoins }} монет</div>
            </div>
          </div>
          <div v-else class="empty-catalog-notice">
            <span>Товары настраиваются во вкладке «Покупки». Добавьте товары в проект, чтобы тестировать SDK.</span>
            <router-link
              v-if="projectId"
              :to="{ name: 'project-purchases', params: { id: projectId } }"
              class="btn-manage-purchases"
              target="_blank"
            >
              <ExternalLink class="icon-xs" />
              <span>Перейти в Покупки</span>
            </router-link>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue';
import {
  Terminal,
  Activity,
  Database,
  Sliders,
  Trash2,
  RotateCcw,
  Info,
  ExternalLink,
} from 'lucide-vue-next';

const props = defineProps({
  projectId: {
    type: [String, Number],
    default: null,
  },
  logs: {
    type: Array,
    default: () => [],
  },
  appEnv: {
    type: Object,
    required: true,
  },
  playerStorage: {
    type: Object,
    default: () => ({}),
  },
  adConfig: {
    type: Object,
    required: true,
  },
  purchaseCatalog: {
    type: Array,
    default: () => [],
  },
});

defineEmits(['clear-logs', 'clear-storage', 'update-env', 'update-ad-config']);

const activeTab = ref('logs');
const searchQuery = ref('');
const currentCategory = ref('all');
const expandedLogs = ref(new Set());

function toggleExpandLog(id) {
  if (expandedLogs.value.has(id)) {
    expandedLogs.value.delete(id);
  } else {
    expandedLogs.value.add(id);
  }
}

const hasSavedData = computed(() => {
  return props.playerStorage && Object.keys(props.playerStorage).length > 0;
});

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

.devtools-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  border-bottom: 1px solid var(--border-color, #2d3139);
  background: var(--bg-surface, #181a1f);
}

.devtools-title-group {
  display: flex;
  align-items: center;
  gap: 8px;
}

.devtools-title {
  font-weight: 600;
  color: var(--text-main, #f0f2f5);
}

.badge-count {
  padding: 1px 6px;
  background: var(--bg-hover, #2b2f38);
  border-radius: 10px;
  font-size: 11px;
  color: var(--text-muted, #9ba1ad);
}

.btn-icon-xs {
  background: transparent;
  border: none;
  color: var(--text-muted, #9ba1ad);
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  transition: all 0.15s ease;
}

.btn-icon-xs:hover {
  color: var(--text-main, #fff);
  background: var(--bg-hover, #2b2f38);
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

.btn-danger-xs {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 3px 8px;
  background: rgba(239, 68, 68, 0.15);
  color: #f87171;
  border: 1px solid rgba(239, 68, 68, 0.3);
  border-radius: 4px;
  font-size: 11px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-danger-xs:hover {
  background: rgba(239, 68, 68, 0.25);
}

.empty-storage-notice {
  font-size: 11px;
  color: var(--text-muted, #7e8494);
  font-style: italic;
  padding: 6px 0;
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
  gap: 8px;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--text-main, #f0f2f5);
  cursor: pointer;
}

.catalog-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.catalog-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 8px;
  background: var(--bg-input, #131418);
  border-radius: 4px;
}

.item-name {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-main, #fff);
}

.item-price {
  font-size: 11px;
  color: #fbbf24;
}

.link-action-xs {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  color: var(--primary, #3b82f6);
  text-decoration: none;
  transition: color 0.15s ease;
}

.link-action-xs:hover {
  text-decoration: underline;
  color: #60a5fa;
}

.empty-catalog-notice {
  font-size: 11px;
  color: var(--text-muted, #7e8494);
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 6px 0;
}

.btn-manage-purchases {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  align-self: flex-start;
  padding: 4px 10px;
  background: rgba(59, 130, 246, 0.12);
  border: 1px solid rgba(59, 130, 246, 0.3);
  border-radius: 4px;
  color: var(--primary, #3b82f6);
  font-size: 11px;
  text-decoration: none;
  font-weight: 500;
  transition: all 0.15s ease;
}

.btn-manage-purchases:hover {
  background: rgba(59, 130, 246, 0.2);
  border-color: var(--primary, #3b82f6);
}

.catalog-item-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.item-id-sub {
  font-size: 10px;
  font-family: monospace;
}
</style>
