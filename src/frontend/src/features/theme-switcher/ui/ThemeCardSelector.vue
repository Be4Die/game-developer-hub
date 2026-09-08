<template>
  <div class="theme-grid">
    <button
      type="button"
      class="theme-card"
      :class="{ active: !isDark }"
      @click="selectTheme('light')"
    >
      <div class="theme-icon-box light-box">
        <Sun class="icon-md" />
      </div>
      <div class="theme-text">
        <span class="theme-name">{{ t('profile.themeLight') }}</span>
        <span class="theme-sub">Light</span>
      </div>
      <Check v-if="!isDark" class="icon-sm theme-check" />
    </button>

    <button
      type="button"
      class="theme-card"
      :class="{ active: isDark }"
      @click="selectTheme('dark')"
    >
      <div class="theme-icon-box dark-box">
        <Moon class="icon-md" />
      </div>
      <div class="theme-text">
        <span class="theme-name">{{ t('profile.themeDark') }}</span>
        <span class="theme-sub">Dark</span>
      </div>
      <Check v-if="isDark" class="icon-sm theme-check" />
    </button>
  </div>
</template>

<script setup>
import { useI18n } from 'vue-i18n';
import { Sun, Moon, Check } from 'lucide-vue-next';
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
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  width: 100%;
}

.theme-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 14px;
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm, 6px);
  cursor: pointer;
  transition: all 0.15s ease;
  text-align: left;
  position: relative;
  outline: none;
}

.theme-card:hover {
  border-color: var(--border-secondary);
}

.theme-card.active {
  border-color: var(--primary);
  background: var(--bg-card);
}

.theme-icon-box {
  width: 34px;
  height: 34px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.light-box {
  background: rgba(245, 176, 39, 0.12);
  color: #f5b027;
  border: 1px solid rgba(245, 176, 39, 0.25);
}

.dark-box {
  background: rgba(88, 166, 255, 0.12);
  color: #58a6ff;
  border: 1px solid rgba(88, 166, 255, 0.25);
}

.theme-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex: 1;
  min-width: 0;
}

.theme-name {
  font-size: 0.92rem;
  font-weight: 600;
  color: var(--text-main);
  line-height: 1.2;
}

.theme-sub {
  font-size: 0.75rem;
  color: var(--text-muted);
  line-height: 1.2;
}

.theme-check {
  color: var(--primary);
  flex-shrink: 0;
  margin-left: auto;
}

@media (max-width: 480px) {
  .theme-grid {
    grid-template-columns: 1fr;
  }
}
</style>
