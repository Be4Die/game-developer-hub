<template>
  <div class="locale-switcher">
    <button
      type="button"
      class="locale-btn"
      :class="{ active: currentLocale === 'ru' }"
      title="Русский язык"
      @click="handleSelect('ru')"
    >
      <span class="flag">🇷🇺</span>
      <span class="code">RU</span>
    </button>
    <div class="divider"></div>
    <button
      type="button"
      class="locale-btn"
      :class="{ active: currentLocale === 'en' }"
      title="English language"
      @click="handleSelect('en')"
    >
      <span class="flag">🇬🇧</span>
      <span class="code">EN</span>
    </button>
  </div>
</template>

<script setup>
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { setLocale } from '@/shared/lib';

const { locale } = useI18n();
const currentLocale = computed(() => locale.value);

function handleSelect(code) {
  setLocale(code);
}
</script>

<style scoped>
.locale-switcher {
  display: inline-flex;
  align-items: center;
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  padding: 2px;
  gap: 2px;
  user-select: none;
}

.locale-btn {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  background: transparent;
  border: none;
  border-radius: calc(var(--radius-sm) - 2px);
  padding: 4px 8px;
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--text-muted);
  cursor: pointer;
  transition: all 0.2s;
  line-height: 1;
}

.locale-btn:hover {
  color: var(--text-main);
  background: var(--bg-hover);
}

.locale-btn.active {
  background: var(--bg-card);
  color: var(--primary);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
}

.flag {
  font-size: 0.95rem;
}

.divider {
  width: 1px;
  height: 14px;
  background: var(--border);
}
</style>
