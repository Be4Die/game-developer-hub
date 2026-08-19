<template>
  <div class="modal-overlay" @click.self="$emit('cancel')">
    <div class="modal-card">
      <h3>Отклонить игру</h3>
      <p class="modal-hint">
        Укажите причину отказа — разработчик увидит её в карточке заявки.
      </p>
      <textarea
        v-model="reason"
        placeholder="Например: не загружена иконка, описание слишком короткое..."
        rows="4"
        autofocus
      ></textarea>
      <div class="modal-actions">
        <button class="btn-cancel" @click="$emit('cancel')">Отмена</button>
        <button
          class="btn-confirm-reject"
          @click="onConfirm"
          :disabled="!reason.trim() || loading"
        >
          {{ loading ? 'Отклонение...' : 'Отклонить' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';

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
  background: rgba(0, 0, 0, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-card {
  background: var(--bg-card);
  border-radius: var(--radius-lg);
  padding: 28px;
  width: 100%;
  max-width: 480px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  margin: 20px;
}

.modal-card h3 {
  margin: 0;
}

.modal-hint {
  margin: 0;
  font-size: 0.88rem;
  color: var(--text-muted);
}

.modal-card textarea {
  width: 100%;
  padding: 12px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  font-family: inherit;
  box-sizing: border-box;
  background: var(--bg-app);
  color: var(--text-main);
  resize: vertical;
}

.modal-actions {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
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
