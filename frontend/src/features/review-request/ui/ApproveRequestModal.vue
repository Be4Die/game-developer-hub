<template>
  <div class="modal-overlay" @click.self="$emit('cancel')">
    <div class="modal-card">
      <div class="modal-header">
        <CheckCircle2 class="icon-md text-success" />
        <h3>Одобрить проект</h3>
      </div>
      <p class="modal-hint">
        Проект будет опубликован в Production. Вы можете оставить комментарий для разработчика (необязательно).
      </p>
      <div class="input-group">
        <label class="input-label">Комментарий модератора</label>
        <textarea
          v-model="comment"
          class="input-control"
          placeholder="Например: Проект соответствует всем правилам площадки."
          rows="3"
        ></textarea>
      </div>
      <div class="modal-actions">
        <button class="btn-cancel" @click="$emit('cancel')">Отмена</button>
        <button
          class="btn-confirm-approve"
          @click="onConfirm"
          :disabled="loading"
        >
          {{ loading ? 'Одобрение...' : '✓ Подтвердить одобрение' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { CheckCircle2 } from 'lucide-vue-next';

defineProps({
  loading: { type: Boolean, default: false },
});

const emit = defineEmits(['confirm', 'cancel']);

const comment = ref('Проект проверен и одобрен к публикации.');

function onConfirm() {
  emit('confirm', comment.value.trim());
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
  border-color: var(--success);
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

.btn-confirm-approve {
  padding: 9px 18px;
  border: none;
  border-radius: var(--radius-md);
  background: var(--success);
  color: white;
  cursor: pointer;
  font-weight: 600;
}

.btn-confirm-approve:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
