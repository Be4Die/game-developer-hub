<template>
  <div class="theme-grid">
    <div
      class="theme-card"
      :class="{ active: !isDark }"
      @click="selectTheme('light')"
    >
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
        <span>Светлая</span>
      </div>
    </div>
    <div
      class="theme-card"
      :class="{ active: isDark }"
      @click="selectTheme('dark')"
    >
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
        <span>Тёмная</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { Sun, Moon } from 'lucide-vue-next';
import { useTheme } from '@/shared/lib';

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
}

.theme-card {
  border: 2px solid var(--border);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all 0.2s;
  overflow: hidden;
}

.theme-card:hover {
  border-color: var(--primary);
}

.theme-card.active {
  border-color: var(--primary);
  box-shadow: 0 0 0 1px var(--primary);
}

.theme-preview {
  height: 120px;
  padding: 8px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.theme-preview-light {
  background: #F3F4F6;
}

.theme-preview-light .preview-header {
  height: 16px;
  background: #FFFFFF;
  border-radius: 4px;
  border: 1px solid #E5E7EB;
}

.theme-preview-light .preview-body {
  flex: 1;
  background: #FFFFFF;
  border-radius: 4px;
  border: 1px solid #E5E7EB;
  padding: 6px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.theme-preview-light .preview-card {
  height: 16px;
  background: #F3F4F6;
  border-radius: 3px;
}

.theme-preview-light .preview-line {
  height: 6px;
  background: #E5E7EB;
  border-radius: 3px;
}

.theme-preview-light .preview-line.short {
  width: 60%;
}

.theme-preview-dark {
  background: #1F2937;
}

.theme-preview-dark .preview-header {
  height: 16px;
  background: #111827;
  border-radius: 4px;
  border: 1px solid #374151;
}

.theme-preview-dark .preview-body {
  flex: 1;
  background: #111827;
  border-radius: 4px;
  border: 1px solid #374151;
  padding: 6px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.theme-preview-dark .preview-card {
  height: 16px;
  background: #1F2937;
  border-radius: 3px;
}

.theme-preview-dark .preview-line {
  height: 6px;
  background: #374151;
  border-radius: 3px;
}

.theme-preview-dark .preview-line.short {
  width: 60%;
}

.theme-info {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  background: var(--bg-secondary);
  font-weight: 500;
  color: var(--text-main);
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
