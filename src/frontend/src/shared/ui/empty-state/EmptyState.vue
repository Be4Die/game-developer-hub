<template>
  <div
    class="empty-state"
    :class="{
      'empty-state--bordered': bordered,
      'empty-state--compact': compact,
    }"
  >
    <slot name="icon">
      <div v-if="icon" class="empty-state__icon-wrap">
        <component :is="icon" class="empty-state__icon" :class="iconClass" />
      </div>
    </slot>

    <slot name="title">
      <h3 v-if="title" class="empty-state__title">{{ title }}</h3>
    </slot>

    <slot name="action">
      <div v-if="actionText" class="empty-state__action">
        <button
          type="button"
          class="empty-state__btn"
          :class="buttonClass"
          :disabled="actionDisabled || actionLoading"
          @click="$emit('action')"
        >
          <span v-if="actionLoading" class="empty-state__spinner"></span>
          <component
            :is="actionIcon"
            v-else-if="actionIcon"
            class="empty-state__btn-icon"
          />
          <span>{{ actionText }}</span>
        </button>
      </div>
    </slot>
  </div>
</template>

<script setup lang="ts">
import { computed, type Component } from 'vue';

interface Props {
  icon?: Component;
  iconClass?: string;
  title?: string;
  actionText?: string;
  actionIcon?: Component;
  actionVariant?: 'primary' | 'secondary' | 'outline' | 'ghost' | 'reset';
  actionLoading?: boolean;
  actionDisabled?: boolean;
  bordered?: boolean;
  compact?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  icon: undefined,
  iconClass: 'text-muted',
  title: '',
  actionText: '',
  actionIcon: undefined,
  actionVariant: 'primary',
  actionLoading: false,
  actionDisabled: false,
  bordered: true,
  compact: false,
});

defineEmits<{
  (e: 'action'): void;
}>();

const buttonClass = computed(() => {
  switch (props.actionVariant) {
    case 'secondary':
      return 'empty-state__btn--secondary';
    case 'outline':
      return 'empty-state__btn--outline';
    case 'ghost':
      return 'empty-state__btn--ghost';
    case 'reset':
      return 'empty-state__btn--reset';
    case 'primary':
    default:
      return 'empty-state__btn--primary';
  }
});
</script>

<style scoped>
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: 48px 24px;
  color: var(--text-muted, #b0b8c4);
  box-sizing: border-box;
  width: 100%;
}

.empty-state--bordered {
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-md, 8px);
}

.empty-state--compact {
  padding: 28px 16px;
}

.empty-state__icon-wrap {
  width: 52px;
  height: 52px;
  border-radius: 50%;
  background: var(--bg-secondary, #21262d);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-muted, #b0b8c4);
  margin-bottom: 12px;
  flex-shrink: 0;
}

.empty-state--compact .empty-state__icon-wrap {
  width: 42px;
  height: 42px;
  margin-bottom: 8px;
}

.empty-state__icon {
  width: 24px;
  height: 24px;
}

.empty-state--compact .empty-state__icon {
  width: 20px;
  height: 20px;
}

.empty-state__title {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
  line-height: 1.35;
  text-align: center;
  max-width: 520px;
}

.empty-state--compact .empty-state__title {
  font-size: 14px;
}

.empty-state__action {
  margin-top: 14px;
  display: flex;
  justify-content: center;
}

.empty-state--compact .empty-state__action {
  margin-top: 10px;
}

.empty-state__btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  height: 34px;
  padding: 0 16px;
  border-radius: var(--radius-sm, 6px);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  border: 1px solid transparent;
  transition: all 0.15s ease;
  white-space: nowrap;
}

.empty-state__btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.empty-state__btn--primary {
  background: var(--primary, #58a6ff);
  color: #ffffff;
  border-color: var(--primary, #58a6ff);
}

.empty-state__btn--primary:hover:not(:disabled) {
  background: var(--primary-hover, #4a94ec);
  border-color: var(--primary-hover, #4a94ec);
}

.empty-state__btn--secondary {
  background: var(--bg-secondary, #21262d);
  color: var(--text-main, #f0f6fc);
  border-color: var(--border, #30363d);
}

.empty-state__btn--secondary:hover:not(:disabled) {
  background: var(--bg-tertiary, #30363d);
}

.empty-state__btn--outline {
  background: transparent;
  color: var(--text-main, #f0f6fc);
  border-color: var(--border, #30363d);
}

.empty-state__btn--outline:hover:not(:disabled) {
  background: var(--bg-secondary, #21262d);
}

.empty-state__btn--ghost {
  background: transparent;
  color: var(--text-muted, #b0b8c4);
  border-color: transparent;
}

.empty-state__btn--ghost:hover:not(:disabled) {
  background: var(--bg-secondary, #21262d);
  color: var(--text-main, #f0f6fc);
}

.empty-state__btn--reset {
  background: transparent;
  color: var(--text-muted, #b0b8c4);
  border-color: var(--border, #30363d);
}

.empty-state__btn--reset:hover:not(:disabled) {
  background: var(--bg-secondary, #21262d);
  color: var(--primary, #58a6ff);
  border-color: var(--primary, #58a6ff);
}

.empty-state__btn-icon {
  width: 14px;
  height: 14px;
}

.empty-state__spinner {
  width: 14px;
  height: 14px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: currentColor;
  border-radius: 50%;
  animation: empty-state-spin 0.8s linear infinite;
}

@keyframes empty-state-spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
