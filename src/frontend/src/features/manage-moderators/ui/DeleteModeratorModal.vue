<template>
  <transition name="modal-fade">
    <div
      v-if="target"
      class="modal-overlay"
      @click.self="$emit('cancel')"
    >
      <div class="modal-card">
        <button class="modal-close" @click="$emit('cancel')">&#x2715;</button>
        <h3>Подтверждение удаления</h3>
        <p>
          Вы уверены, что хотите удалить модератора
          <strong>{{ target.display_name }}</strong>?
        </p>
        <p class="warning-text">Это действие нельзя отменить.</p>
        <div class="modal-actions">
          <button class="btn btn-secondary" @click="$emit('cancel')">
            Отмена
          </button>
          <button
            class="btn btn-danger"
            @click="handleDelete"
            :disabled="deleting"
          >
            {{ deleting ? 'Удаление...' : 'Удалить' }}
          </button>
        </div>
      </div>
    </div>
  </transition>
</template>

<script setup>
import { ref } from 'vue';
import { deleteUser } from '@/entities/user';
import { showToast } from '@/shared/lib';

const props = defineProps({
  target: { type: Object, default: null },
});

const emit = defineEmits(['deleted', 'cancel']);
const deleting = ref(false);

async function handleDelete() {
  if (!props.target) return;
  deleting.value = true;
  try {
    await deleteUser(props.target.id);
    showToast(
      `Модератор "${props.target.display_name}" удалён`,
      'success'
    );
    emit('deleted', props.target.id);
  } catch (err) {
    showToast(
      err.response?.data?.message || 'Не удалось удалить модератора',
      'error'
    );
  } finally {
    deleting.value = false;
  }
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 200;
  display: flex;
  align-items: center;
  justify-content: center;
  backdrop-filter: blur(4px);
}

.modal-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  padding: 32px 28px;
  width: 420px;
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 12px;
  box-shadow: var(--shadow-lg);
}

.modal-close {
  position: absolute;
  top: 14px;
  right: 14px;
  background: none;
  border: none;
  font-size: 1rem;
  color: var(--text-muted);
  cursor: pointer;
  transition: color 0.2s;
}

.modal-close:hover {
  color: var(--text-main);
}

.modal-card h3 {
  margin: 0;
  font-size: 1.1rem;
  color: var(--text-main);
}

.modal-card p {
  margin: 0;
  color: var(--text-muted);
  font-size: 0.9rem;
  line-height: 1.5;
}

.modal-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  margin-top: 8px;
}

.warning-text {
  color: var(--danger) !important;
  font-weight: 600;
}

.modal-fade-enter-active,
.modal-fade-leave-active {
  transition: opacity 0.2s, transform 0.2s;
}

.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
  transform: scale(0.95);
}

@media (max-width: 768px) {
  .modal-card {
    width: 90vw;
  }
}
</style>
