<template>
  <div class="modal-overlay" @click.self="$emit('cancel')">
    <div class="modal-card">
      <div class="modal-header">
        <AlertTriangle class="icon-md text-danger" />
        <h3>Отклонить проект</h3>
      </div>
      <p class="modal-hint">
        Укажите обязательную причину отказа. Разработчик получит уведомление и увидит причину в карточке черновика для устранения замечаний.
      </p>
      <div class="input-group">
        <label class="input-label">Причина отказа <span class="req">*</span></label>
        <textarea
          v-model="reason"
          class="input-control"
          placeholder="Например: Некорректное описание, не работает управление на пробел, отсутствуют иконки..."
          rows="4"
          autofocus
        ></textarea>
      </div>
      <div class="modal-actions">
        <button class="btn-cancel" @click="$emit('cancel')">Отмена</button>
        <button
          class="btn-confirm-reject"
          @click="onConfirm"
          :disabled="!reason.trim() || loading"
        >
          {{ loading ? 'Отклонение...' : '✕ Отклонить проект' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { AlertTriangle } from 'lucide-vue-next';

defineProps({
  loading: { type: Boolean, default: false },
});

const emit = defineEmits(['confirm', 'cancel']);

const reason = ref('');

function onConfirm() {
  if (!reason.value.trim()) return;
  emit('confirm', reason.value.trim());
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  backdrop-filter: blur(2px);
}

.modal-card {
  background: var(--bg-card);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border);
  padding: 24px;
  width: 100%;
  max-width: 480px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  margin: 20px;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
}

.modal-header {
  display: flex;
  align-items: center;
  gap: 10px;
}

.modal-header h3 {
  margin: 0;
  font-size: 1.2rem;
  font-weight: 700;
}

.text-danger {
  color: var(--danger, #ef4444);
}

.modal-hint {
  margin: 0;
  font-size: 0.88rem;
  color: var(--text-muted);
  line-height: 1.4;
}

.input-label {
  display: block;
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--text-secondary);
  margin-bottom: 6px;
}

.req {
  color: var(--danger, #ef4444);
}

.input-control {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  font-family: inherit;
  box-sizing: border-box;
  background: var(--bg-app);
  color: var(--text-main);
  resize: vertical;
}

.input-control:focus {
  outline: none;
  border-color: var(--danger, #ef4444);
}

.modal-actions {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
  margin-top: 8px;
}

.btn-cancel {
  padding: 9px 18px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  background: transparent;
  cursor: pointer;
  font-weight: 600;
  color: var(--text-main);
}

.btn-confirm-reject {
  padding: 9px 18px;
  border: none;
  border-radius: var(--radius-md);
  background: var(--danger, #ef4444);
  color: white;
  cursor: pointer;
  font-weight: 600;
}

.btn-confirm-reject:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
