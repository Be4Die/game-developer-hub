<template>
  <div ref="containerRef" class="rule-search-select">
    <!-- Кнопка-триггер / поле быстрого выбора -->
    <div class="select-trigger-bar" @click="toggleDropdown">
      <div class="trigger-left">
        <Search class="icon-xs text-muted" />
        <span v-if="selectedRule" class="selected-text">
          <span class="selected-code">{{ selectedRule.code }}</span>
          <span class="selected-title">{{ selectedRule.title }}</span>
        </span>
        <span v-else class="placeholder-text">
          {{ placeholder || 'Поиск по регламенту (напр. SEC-04, сервер, квоты, баг...)' }}
        </span>
      </div>

      <div class="trigger-right">
        <button
          v-if="selectedRule"
          type="button"
          class="btn-clear"
          title="Сбросить выбор"
          @click.stop="clearSelection"
        >
          <X class="icon-xs" />
        </button>
        <ChevronDown class="icon-xs text-muted transition-icon" :class="{ 'is-open': isOpen }" />
      </div>
    </div>

    <!-- Выпадающее меню со строкой поиска и списком правил -->
    <div v-if="isOpen" class="dropdown-panel">
      <div class="dropdown-search-wrap">
        <Search class="icon-xs text-muted" />
        <input
          ref="searchInputRef"
          v-model="searchQuery"
          type="text"
          class="dropdown-search-input"
          placeholder="Введите код (SEC-01), название или ключевое слово..."
          @keydown.esc="closeDropdown"
        />
        <button
          v-if="searchQuery"
          type="button"
          class="btn-clear-search"
          @click="searchQuery = ''"
        >
          <X class="icon-xs" />
        </button>
      </div>

      <!-- Фильтры по категориям (быстрые табы) -->
      <div class="category-tabs">
        <button
          type="button"
          class="cat-tab-chip"
          :class="{ active: selectedCategory === 'all' }"
          @click="selectedCategory = 'all'"
        >
          Все ({{ PLATFORM_RULES.length }})
        </button>
        <button
          v-for="cat in categories"
          :key="cat.id"
          type="button"
          class="cat-tab-chip"
          :class="{ active: selectedCategory === cat.id }"
          @click="selectedCategory = cat.id"
        >
          {{ cat.title }} ({{ cat.rules.length }})
        </button>
      </div>

      <!-- Список результатов -->
      <div class="rules-scroll-list">
        <div v-if="filteredRules.length === 0" class="no-results">
          <AlertCircle class="icon-sm text-muted" />
          <span>Правила по запросу «{{ searchQuery }}» не найдены</span>
        </div>

        <div
          v-for="rule in filteredRules"
          :key="rule.code"
          class="rule-option-item"
          :class="{
            'is-selected': selectedRule?.code === rule.code,
            'is-critical': rule.severity === 'critical',
          }"
          @click="selectRule(rule)"
        >
          <div class="option-header">
            <span class="rule-code-badge" :class="`cat-${rule.category}`">
              {{ rule.code }}
            </span>
            <span class="rule-title-text">{{ rule.title }}</span>
            <span v-if="rule.severity === 'critical'" class="badge-critical" title="Критическое нарушение">
              Критично
            </span>
          </div>

          <p class="rule-summary-text">{{ rule.summary }}</p>

          <div class="option-footer">
            <span class="cat-label">{{ rule.categoryTitle }}</span>
            <span class="updated-label">ред. {{ rule.lastUpdated }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, nextTick } from 'vue';
import { Search, X, ChevronDown, AlertCircle } from 'lucide-vue-next';
import {
  PLATFORM_RULES,
  getRulesByCategory,
  getRuleByCode,
} from '@/entities/documentation';

const props = defineProps({
  modelValue: {
    type: String,
    default: '',
  },
  placeholder: {
    type: String,
    default: '',
  },
});

const emit = defineEmits(['update:modelValue', 'select', 'clear']);

const containerRef = ref(null);
const searchInputRef = ref(null);
const isOpen = ref(false);
const searchQuery = ref('');
const selectedCategory = ref('all');

const categories = computed(() => getRulesByCategory());

const selectedRule = computed(() => {
  return getRuleByCode(props.modelValue);
});

const filteredRules = computed(() => {
  let list = PLATFORM_RULES;

  if (selectedCategory.value !== 'all') {
    list = list.filter((r) => r.category === selectedCategory.value);
  }

  if (searchQuery.value.trim()) {
    const q = searchQuery.value.trim().toLowerCase();
    list = list.filter(
      (r) =>
        r.code.toLowerCase().includes(q) ||
        r.title.toLowerCase().includes(q) ||
        r.summary.toLowerCase().includes(q) ||
        r.description.toLowerCase().includes(q) ||
        r.categoryTitle.toLowerCase().includes(q)
    );
  }

  return list;
});

function toggleDropdown() {
  if (isOpen.value) {
    closeDropdown();
  } else {
    openDropdown();
  }
}

function openDropdown() {
  isOpen.value = true;
  nextTick(() => {
    searchInputRef.value?.focus();
  });
}

function closeDropdown() {
  isOpen.value = false;
  searchQuery.value = '';
}

function selectRule(rule) {
  emit('update:modelValue', rule.code);
  emit('select', rule);
  closeDropdown();
}

function clearSelection() {
  emit('update:modelValue', '');
  emit('clear');
}

function handleClickOutside(e) {
  if (containerRef.value && !containerRef.value.contains(e.target)) {
    closeDropdown();
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside);
});

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside);
});
</script>

<style scoped>
.rule-search-select {
  position: relative;
  width: 100%;
}

.select-trigger-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  background: var(--bg-secondary, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  cursor: pointer;
  min-height: 40px;
  transition: all 0.15s ease;
  user-select: none;
}

.select-trigger-bar:hover {
  border-color: var(--primary, #58a6ff);
  background: var(--bg-hover, #21262d);
}

.trigger-left {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  overflow: hidden;
}

.selected-text {
  display: flex;
  align-items: center;
  gap: 8px;
  overflow: hidden;
}

.selected-code {
  background: rgba(88, 166, 255, 0.15);
  color: var(--primary, #58a6ff);
  font-family: monospace;
  font-weight: 700;
  font-size: 11px;
  padding: 2px 6px;
  border-radius: 4px;
  border: 1px solid rgba(88, 166, 255, 0.3);
}

.selected-title {
  color: var(--text-main, #f0f6fc);
  font-size: 13px;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.placeholder-text {
  color: var(--text-muted, #8b949e);
  font-size: 13px;
}

.trigger-right {
  display: flex;
  align-items: center;
  gap: 6px;
}

.btn-clear {
  background: transparent;
  border: none;
  color: var(--text-muted, #8b949e);
  padding: 2px;
  cursor: pointer;
  border-radius: 4px;
  display: flex;
  align-items: center;
}

.btn-clear:hover {
  color: var(--danger, #f85149);
  background: rgba(248, 81, 73, 0.1);
}

.transition-icon {
  transition: transform 0.2s ease;
}

.transition-icon.is-open {
  transform: rotate(180deg);
}

.dropdown-panel {
  position: absolute;
  top: calc(100% + 6px);
  left: 0;
  right: 0;
  z-index: 100;
  background: var(--bg-secondary, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-md, 8px);
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.5);
  overflow: hidden;
  animation: fadeIn 0.15s ease-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(-4px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.dropdown-search-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  border-bottom: 1px solid var(--border, #30363d);
  background: var(--bg-app, #0d1117);
}

.dropdown-search-input {
  flex: 1;
  background: transparent;
  border: none;
  color: var(--text-main, #f0f6fc);
  font-size: 13px;
  outline: none;
}

.btn-clear-search {
  background: transparent;
  border: none;
  color: var(--text-muted, #8b949e);
  cursor: pointer;
}

.category-tabs {
  display: flex;
  gap: 6px;
  padding: 8px 12px;
  border-bottom: 1px solid var(--border, #30363d);
  background: rgba(0, 0, 0, 0.2);
  overflow-x: auto;
}

.cat-tab-chip {
  background: transparent;
  border: 1px solid var(--border, #30363d);
  color: var(--text-muted, #8b949e);
  font-size: 11px;
  padding: 3px 8px;
  border-radius: 12px;
  white-space: nowrap;
  cursor: pointer;
  transition: all 0.12s;
}

.cat-tab-chip:hover {
  background: var(--bg-hover, #21262d);
  color: var(--text-main, #f0f6fc);
}

.cat-tab-chip.active {
  background: rgba(88, 166, 255, 0.15);
  color: var(--primary, #58a6ff);
  border-color: var(--primary, #58a6ff);
  font-weight: 600;
}

.rules-scroll-list {
  max-height: 280px;
  overflow-y: auto;
  padding: 6px;
}

.no-results {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 24px;
  color: var(--text-muted, #8b949e);
  font-size: 13px;
}

.rule-option-item {
  padding: 8px 10px;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.12s;
  margin-bottom: 4px;
  border-left: 3px solid transparent;
}

.rule-option-item:hover {
  background: var(--bg-hover, #21262d);
}

.rule-option-item.is-selected {
  background: rgba(88, 166, 255, 0.1);
  border-left-color: var(--primary, #58a6ff);
}

.rule-option-item.is-critical {
  border-left-color: rgba(248, 81, 73, 0.4);
}

.option-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.rule-code-badge {
  font-family: monospace;
  font-size: 11px;
  font-weight: 700;
  padding: 1px 5px;
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.08);
  color: var(--text-main, #f0f6fc);
}

.rule-code-badge.cat-security {
  color: #58a6ff;
  background: rgba(88, 166, 255, 0.15);
}

.rule-code-badge.cat-servers {
  color: #3fb950;
  background: rgba(63, 185, 80, 0.15);
}

.rule-code-badge.cat-builds {
  color: #d29922;
  background: rgba(210, 153, 34, 0.15);
}

.rule-code-badge.cat-content {
  color: #bc8cff;
  background: rgba(188, 140, 255, 0.15);
}

.rule-title-text {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
  flex: 1;
}

.badge-critical {
  font-size: 10px;
  padding: 1px 5px;
  border-radius: 4px;
  background: rgba(248, 81, 73, 0.15);
  color: var(--danger, #f85149);
  font-weight: 600;
}

.rule-summary-text {
  font-size: 12px;
  color: var(--text-muted, #8b949e);
  margin: 0 0 6px;
  line-height: 1.4;
}

.option-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 11px;
  color: #6e7681;
}
</style>
