<template>
  <Teleport to="body">
    <Transition name="dialog">
      <div v-if="modelValue" class="dialog-overlay" @click.self="onCancel">
        <div class="dialog-box">
          <div class="dialog-header">
            <AlertTriangle v-if="type === 'warning'" class="dialog-icon warning" />
            <AlertCircle v-else-if="type === 'danger'" class="dialog-icon danger" />
            <HelpCircle v-else class="dialog-icon info" />
            <h3>{{ title }}</h3>
          </div>
          <div class="dialog-body">
            <p>{{ message }}</p>
          </div>
          <div class="dialog-footer">
            <button class="btn-outline" @click="onCancel">{{ cancelText }}</button>
            <button :class="confirmClass" @click="onConfirm">{{ confirmText }}</button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { computed } from 'vue'
import { AlertTriangle, AlertCircle, HelpCircle } from 'lucide-vue-next'

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  title: { type: String, default: 'Подтвердите действие' },
  message: { type: String, default: 'Вы уверены?' },
  confirmText: { type: String, default: 'Подтвердить' },
  cancelText: { type: String, default: 'Отмена' },
  type: { type: String, default: 'warning' }, // warning | danger | info
})

const emit = defineEmits(['update:modelValue', 'confirm', 'cancel'])

const confirmClass = computed(() => {
  switch (props.type) {
    case 'danger': return 'btn-danger'
    case 'info': return 'btn-primary'
    default: return 'btn-warning'
  }
})

function onConfirm() {
  emit('confirm')
  emit('update:modelValue', false)
}

function onCancel() {
  emit('cancel')
  emit('update:modelValue', false)
}
</script>

<style scoped>
.dialog-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 16px;
}

.dialog-box {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  width: 100%;
  max-width: 420px;
  box-shadow: 0 20px 40px rgba(0,0,0,0.2);
  overflow: hidden;
}

.dialog-header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 16px 20px 0;
}
.dialog-header h3 {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
}

.dialog-icon {
  width: 22px;
  height: 22px;
  flex-shrink: 0;
}
.dialog-icon.warning { color: var(--warning); }
.dialog-icon.danger  { color: var(--danger); }
.dialog-icon.info    { color: var(--primary); }

.dialog-body {
  padding: 12px 20px;
}
.dialog-body p {
  margin: 0;
  font-size: 0.9rem;
  color: var(--text-muted);
  line-height: 1.5;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 12px 20px 16px;
  border-top: 1px solid var(--border);
}

.btn-warning {
  padding: 8px 16px;
  font-size: 0.85rem;
  font-weight: 500;
  color: #fff;
  background: var(--warning);
  border: none;
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: 0.15s;
}
.btn-warning:hover { background: var(--warning-dark, #e6a000); }

.btn-danger {
  padding: 8px 16px;
  font-size: 0.85rem;
  font-weight: 500;
  color: #fff;
  background: var(--danger);
  border: none;
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: 0.15s;
}
.btn-danger:hover { background: var(--danger-dark, #c53030); }

.btn-outline {
  padding: 8px 16px;
  font-size: 0.85rem;
  font-weight: 500;
  color: var(--text-main);
  background: transparent;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: 0.15s;
}
.btn-outline:hover {
  border-color: var(--primary);
  color: var(--primary);
  background: var(--primary-light);
}

.dialog-enter-active,
.dialog-leave-active {
  transition: opacity 0.2s ease;
}
.dialog-enter-active .dialog-box,
.dialog-leave-active .dialog-box {
  transition: transform 0.2s ease, opacity 0.2s ease;
}
.dialog-enter-from,
.dialog-leave-to {
  opacity: 0;
}
.dialog-enter-from .dialog-box,
.dialog-leave-to .dialog-box {
  opacity: 0;
  transform: scale(0.95) translateY(10px);
}
</style>
