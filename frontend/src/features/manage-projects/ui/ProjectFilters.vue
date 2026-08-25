<template>
  <div class="toolbar-card">
    <div class="search-box">
      <Search class="icon-sm search-icon" />
      <input
        type="text"
        :value="searchQuery"
        @input="$emit('update:searchQuery', $event.target.value)"
        :placeholder="t('projects.searchPlaceholder')"
        class="search-input"
      />
      <button
        v-if="searchQuery"
        class="clear-btn"
        @click="$emit('update:searchQuery', '')"
      >
        <X class="icon-xs" />
      </button>
    </div>

    <div class="filters-group">
      <div class="filter-item">
        <label class="filter-label">{{ t('common.status') }}:</label>
        <select
          :value="statusFilter"
          @change="$emit('update:statusFilter', $event.target.value)"
          class="filter-select"
        >
          <option value="all">{{ t('projects.allStatuses') }}</option>
          <option value="draft">{{ t('projects.draft') }}</option>
          <option value="pending">{{ t('projects.moderation') }}</option>
          <option value="published">{{ t('projects.approved') }}</option>
          <option value="rejected">{{ t('projects.rejected') }}</option>
        </select>
      </div>

      <div class="filter-item">
        <label class="filter-label">{{ t('common.actions') }}:</label>
        <select
          :value="sortBy"
          @change="$emit('update:sortBy', $event.target.value)"
          class="filter-select"
        >
          <option value="newest">{{ t('stats.today') }} / {{ t('common.created') }}</option>
          <option value="oldest">{{ t('common.created') }} ↑</option>
          <option value="title">{{ t('common.name') }} (A–Z)</option>
        </select>
      </div>

      <div class="view-toggle">
        <button
          class="toggle-btn"
          :class="{ active: viewMode === 'table' }"
          :title="t('servers.tabs.overview')"
          @click="$emit('update:viewMode', 'table')"
        >
          <List class="icon-sm" />
        </button>
        <button
          class="toggle-btn"
          :class="{ active: viewMode === 'grid' }"
          :title="t('projects.title')"
          @click="$emit('update:viewMode', 'grid')"
        >
          <LayoutGrid class="icon-sm" />
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { useI18n } from 'vue-i18n';
import { Search, X, List, LayoutGrid } from 'lucide-vue-next';

const { t } = useI18n();

defineProps({
  searchQuery: { type: String, default: '' },
  statusFilter: { type: String, default: 'all' },
  sortBy: { type: String, default: 'newest' },
  viewMode: { type: String, default: 'table' },
});

defineEmits([
  'update:searchQuery',
  'update:statusFilter',
  'update:sortBy',
  'update:viewMode',
]);
</script>


<style scoped>
.toolbar-card {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 14px 18px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.search-box {
  position: relative;
  display: flex;
  align-items: center;
  flex: 1;
  min-width: 260px;
}

.search-icon {
  position: absolute;
  left: 12px;
  color: var(--text-tertiary);
}

.search-input {
  width: 100%;
  padding: 8px 32px 8px 36px;
  background: var(--bg-main);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  color: var(--text-main);
  font-size: 0.9rem;
  outline: none;
  transition: border-color 0.2s;
}

.search-input:focus {
  border-color: var(--primary);
}

.clear-btn {
  position: absolute;
  right: 10px;
  background: transparent;
  border: none;
  color: var(--text-tertiary);
  cursor: pointer;
  padding: 2px;
  display: flex;
  align-items: center;
}

.filters-group {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
}

.filter-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.filter-label {
  font-size: 0.85rem;
  color: var(--text-tertiary);
  font-weight: 500;
}

.filter-select {
  padding: 7px 12px;
  background: var(--bg-main);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  color: var(--text-main);
  font-size: 0.85rem;
  outline: none;
  cursor: pointer;
}

.view-toggle {
  display: flex;
  background: var(--bg-main);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  overflow: hidden;
}

.toggle-btn {
  background: transparent;
  border: none;
  padding: 6px 10px;
  color: var(--text-tertiary);
  cursor: pointer;
  display: flex;
  align-items: center;
  transition: all 0.2s;
}

.toggle-btn.active {
  background: var(--bg-tertiary);
  color: var(--text-main);
}
</style>
