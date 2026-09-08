<template>
  <div class="filters-toolbar">
    <!-- Поле фильтрации по названию (растягивается на всё доступное пространство) -->
    <div class="filter-field field-name">
      <label class="field-label">{{ t('common.name') }}</label>
      <div class="input-wrapper">
        <input
          type="text"
          :value="searchQuery"
          :placeholder="t('common.name')"
          class="filter-input"
          @input="$emit('update:searchQuery', $event.target.value)"
        />
        <button
          v-if="searchQuery"
          class="clear-input-btn"
          title="Очистить"
          @click="$emit('update:searchQuery', '')"
        >
          <X class="icon-xs" />
        </button>
      </div>
    </div>

    <!-- Селектор статуса -->
    <div class="filter-field field-status">
      <label class="field-label">{{ t('common.status') }}</label>
      <div class="select-wrapper">
        <select
          :value="statusFilter"
          class="filter-select"
          @change="$emit('update:statusFilter', $event.target.value)"
        >
          <option value="all">—</option>
          <option value="draft">{{ t('projects.draft') }}</option>
          <option value="pending">{{ t('projects.moderation') }}</option>
          <option value="published">{{ t('projects.published') }}</option>
          <option value="rejected">{{ t('projects.rejected') }}</option>
        </select>
        <ChevronDown class="icon-xs select-arrow" />
      </div>
    </div>

    <!-- Селектор принадлежности / доступа -->
    <div class="filter-field field-role">
      <label class="field-label">{{ t('projects.accessColumn') }}</label>
      <div class="select-wrapper">
        <select
          :value="roleFilter"
          class="filter-select"
          @change="$emit('update:roleFilter', $event.target.value)"
        >
          <option value="all">{{ t('common.all') }}</option>
          <option value="owned">{{ t('access.statuses.owner') }}</option>
          <option value="shared">{{ t('access.statuses.member') }}</option>
        </select>
        <ChevronDown class="icon-xs select-arrow" />
      </div>
    </div>

    <!-- Селектор сортировки -->
    <div class="filter-field field-sort">
      <label class="field-label">{{ t('common.actions') }}</label>
      <div class="select-wrapper">
        <select
          :value="sortBy"
          class="filter-select"
          @change="$emit('update:sortBy', $event.target.value)"
        >
          <option value="newest">{{ t('stats.today') }} / {{ t('common.created') }}</option>
          <option value="oldest">{{ t('common.created') }} ↑</option>
          <option value="title">{{ t('common.name') }} (A–Z)</option>
        </select>
        <ChevronDown class="icon-xs select-arrow" />
      </div>
    </div>

    <!-- Кнопка сброса фильтров -->
    <button
      v-if="searchQuery || statusFilter !== 'all' || roleFilter !== 'all' || sortBy !== 'newest'"
      class="btn-reset-filters"
      title="Сбросить фильтры"
      @click="$emit('reset')"
    >
      <RotateCcw class="icon-xs" />
      <span>{{ t('common.reset') }}</span>
    </button>

    <!-- Кнопка создания игры -->
    <button class="btn-add-game" :disabled="creating" @click="$emit('create')">
      <span v-if="creating" class="spinner-btn"></span>
      <span>{{ creating ? t('common.saving') : t('projects.createBtn') }}</span>
    </button>
  </div>
</template>

<script setup>
import { useI18n } from 'vue-i18n';
import { X, ChevronDown, RotateCcw } from 'lucide-vue-next';

const { t } = useI18n();

defineProps({
  searchQuery: { type: String, default: '' },
  statusFilter: { type: String, default: 'all' },
  roleFilter: { type: String, default: 'all' },
  sortBy: { type: String, default: 'newest' },
  creating: { type: Boolean, default: false },
});

defineEmits([
  'update:searchQuery',
  'update:statusFilter',
  'update:roleFilter',
  'update:sortBy',
  'reset',
  'create',
]);
</script>

<style scoped>
.filters-toolbar {
  display: flex;
  align-items: flex-end;
  gap: 16px;
  margin-bottom: 24px;
  width: 100%;
}

.filter-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.field-name {
  flex: 1;
  min-width: 220px;
}

.field-status {
  width: 170px;
  flex-shrink: 0;
}

.field-role {
  width: 170px;
  flex-shrink: 0;
}

.field-sort {
  width: 190px;
  flex-shrink: 0;
}

.field-label {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-tertiary, #8b949e);
  letter-spacing: 0.1px;
}

.input-wrapper,
.select-wrapper {
  position: relative;
  display: flex;
  align-items: center;
  width: 100%;
}

.filter-input {
  width: 100%;
  height: 36px;
  padding: 0 32px 0 12px;
  background: var(--bg-secondary, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  color: var(--text-main, #f0f6fc);
  font-size: 13px;
  outline: none;
  transition: border-color 0.15s;
}

.filter-input:focus {
  border-color: var(--primary, #58a6ff);
}

.filter-input::placeholder {
  color: var(--text-tertiary, #6e7681);
}

.clear-input-btn {
  position: absolute;
  right: 8px;
  background: transparent;
  border: none;
  color: var(--text-tertiary, #8b949e);
  cursor: pointer;
  padding: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.clear-input-btn:hover {
  color: var(--text-main, #f0f6fc);
}

.filter-select {
  width: 100%;
  height: 36px;
  padding: 0 30px 0 12px;
  background: var(--bg-secondary, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  color: var(--text-main, #f0f6fc);
  font-size: 13px;
  outline: none;
  cursor: pointer;
  appearance: none;
  -webkit-appearance: none;
  transition: border-color 0.15s;
}

.filter-select:focus {
  border-color: var(--primary, #58a6ff);
}

.select-arrow {
  position: absolute;
  right: 10px;
  pointer-events: none;
  color: var(--text-tertiary, #8b949e);
}

.btn-reset-filters {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 36px;
  padding: 0 12px;
  background: transparent;
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  color: var(--text-muted, #b0b8c4);
  font-size: 13px;
  cursor: pointer;
  flex-shrink: 0;
  transition: all 0.15s;
}

.btn-reset-filters:hover {
  background: var(--bg-tertiary, #21262d);
  color: var(--text-main, #f0f6fc);
  border-color: var(--border-secondary, #484f58);
}

.btn-add-game {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  height: 36px;
  padding: 0 18px;
  background: var(--primary, #58a6ff);
  color: #ffffff;
  border: none;
  border-radius: var(--radius-sm, 6px);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  flex-shrink: 0;
  transition:
    background-color 0.15s,
    opacity 0.15s;
  white-space: nowrap;
}

.btn-add-game:hover {
  background: var(--primary-hover, #79c0ff);
}

.btn-add-game:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.spinner-btn {
  width: 14px;
  height: 14px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: #ffffff;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  display: inline-block;
  flex-shrink: 0;
  vertical-align: middle;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 800px) {
  .filters-toolbar {
    flex-wrap: wrap;
  }
  .field-name {
    width: 100%;
    min-width: 100%;
  }
}
</style>
