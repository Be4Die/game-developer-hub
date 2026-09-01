<template>
  <div class="theme-grid">
    <div class="theme-card" :class="{ active: !isDark }" @click="selectTheme('light')">
      <div class="theme-preview theme-preview-light">
        <div class="preview-header"></div>
        <div class="preview-body">
          <div class="preview-card"></div>
          <div class="preview-line"></div>
          <div class="preview-line short"></div>
        </div>
      </div>
      <div class="theme-info">
        <Sun class="icon-md theme-icon" />
        <span>{{ t('profile.themeLight') }}</span>
      </div>
    </div>
    <div class="theme-card" :class="{ active: isDark }" @click="selectTheme('dark')">
      <div class="theme-preview theme-preview-dark">
        <div class="preview-header"></div>
        <div class="preview-body">
          <div class="preview-card"></div>
          <div class="preview-line"></div>
          <div class="preview-line short"></div>
        </div>
      </div>
      <div class="theme-info">
        <Moon class="icon-md theme-icon" />
        <span>{{ t('profile.themeDark') }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { useI18n } from 'vue-i18n';
import { Sun, Moon } from 'lucide-vue-next';
import { useTheme } from '@/shared/lib';

const { t } = useI18n();
const { isDark, setTheme } = useTheme();

function selectTheme(value) {
  setTheme(value);
}
</script>

<style scoped>
.theme-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
  height: 100%;
  flex: 1;
}

.theme-card {
  border: 1px solid var(--border);
  border-radius: var(--radius-sm, 6px);
  cursor: pointer;
  transition: all 0.2s;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--bg-secondary);
}

.theme-card:hover {
  border-color: var(--border-secondary);
}

.theme-card.active {
  border-color: var(--primary);
  box-shadow: 0 0 0 1px var(--primary);
  background: var(--bg-card);
}

.theme-preview {
  flex: 1;
  min-height: 140px;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.theme-preview-light {
  background: #f3f4f6;
}

.theme-preview-light .preview-header {
  height: 18px;
  background: #ffffff;
  border-radius: 4px;
  border: 1px solid #e5e7eb;
}

.theme-preview-light .preview-body {
  flex: 1;
  background: #ffffff;
  border-radius: 4px;
  border: 1px solid #e5e7eb;
  padding: 8px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.theme-preview-light .preview-card {
  height: 20px;
  background: #f3f4f6;
  border-radius: 3px;
}

.theme-preview-light .preview-line {
  height: 6px;
  background: #e5e7eb;
  border-radius: 3px;
}

.theme-preview-light .preview-line.short {
  width: 60%;
}

.theme-preview-dark {
  background: #181b20;
}

.theme-preview-dark .preview-header {
  height: 18px;
  background: #21262d;
  border-radius: 4px;
  border: 1px solid #30363d;
}

.theme-preview-dark .preview-body {
  flex: 1;
  background: #21262d;
  border-radius: 4px;
  border: 1px solid #30363d;
  padding: 8px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.theme-preview-dark .preview-card {
  height: 20px;
  background: #161b22;
  border-radius: 3px;
}

.theme-preview-dark .preview-line {
  height: 6px;
  background: #30363d;
  border-radius: 3px;
}

.theme-preview-dark .preview-line.short {
  width: 60%;
}

.theme-info {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 14px;
  background: var(--bg-secondary);
  font-weight: 500;
  font-size: 0.92rem;
  color: var(--text-main);
  border-top: 1px solid var(--border);
}

.theme-icon {
  color: var(--primary);
}

@media (max-width: 640px) {
  .theme-grid {
    grid-template-columns: 1fr;
  }
}
</style>
