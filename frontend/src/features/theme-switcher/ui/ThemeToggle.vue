<template>
  <button class="theme-toggle" @click="toggleTheme" title="Переключить тему">
    <Sun v-if="isDark" class="icon-sm" />
    <Moon v-else class="icon-sm" />
  </button>
</template>

<script setup>
import { Sun, Moon } from 'lucide-vue-next';
import { useTheme } from '@/shared/lib';

const { isDark, toggleTheme } = useTheme();
</script>

<style scoped>
.theme-toggle {
  width: 36px;
  height: 36px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border);
  background: var(--bg-secondary);
  color: var(--text-muted);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
  font-size: 16px;
}

.theme-toggle:hover {
  background: var(--bg-tertiary);
  color: var(--text-main);
  border-color: var(--border-secondary);
}
</style>
