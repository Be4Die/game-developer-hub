<template>
  <transition name="modal-fade">
    <div v-if="target" class="modal-overlay" @click.self="$emit('cancel')">
      <div class="modal-card">
        <div class="modal-header">
          <div class="modal-title-wrap">
            <Trash2 class="icon-md text-danger" />
            <div>
              <h3>Удаление модератора</h3>
              <p class="modal-subtitle">Подтверждение удаления учётной записи</p>
            </div>
          </div>
          <button class="modal-close" @click="$emit('cancel')">
            <X class="icon-sm" />
          </button>
        </div>

        <p class="modal-text">
          Вы уверены, что хотите удалить модератора
          <strong>{{ target.display_name || target.email }}</strong
          >?
        </p>
        <p class="hint-text">
          Аккаунт будет помечен как удалённый. Администратор сможет восстановить его в любой момент.
        </p>

        <div class="modal-actions">
          <button class="btn-modal-secondary" :disabled="deleting" @click="$emit('cancel')">
            Отмена
          </button>
          <button class="btn-modal-danger" :disabled="deleting" @click="handleDelete">
            <Loader2 v-if="deleting" class="icon-xs spin" />
            <Trash2 v-else class="icon-xs" />
            <span>{{ deleting ? 'Удаление...' : 'Удалить' }}</span>
          </button>
        </div>
      </div>
    </div>
  </transition>
</template>

<script setup>
import { ref } from 'vue';
import { Trash2, X, Loader2 } from 'lucide-vue-next';
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
    showToast(`Модератор "${props.target.display_name || props.target.email}" удалён`, 'success');
    emit('deleted', props.target.id);
  } catch (err) {
    showToast(err.response?.data?.message || 'Не удалось удалить модератора', 'danger');
  } finally {
    deleting.value = false;
  }
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.7);
  z-index: 300;
  display: flex;
  align-items: center;
  justify-content: center;
  backdrop-filter: blur(4px);
  padding: 16px;
}

.modal-card {
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-md, 8px);
  padding: 24px;
  width: 100%;
  max-width: 440px;
  box-shadow: 0 16px 32px rgba(0, 0, 0, 0.4);
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.modal-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
}

.modal-title-wrap {
  display: flex;
  align-items: center;
  gap: 10px;
}

.modal-title-wrap h3 {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
}

.modal-subtitle {
  margin: 2px 0 0 0;
  font-size: 12px;
  color: var(--text-tertiary, #8b949e);
}

.modal-close {
  background: transparent;
  border: none;
  color: var(--text-tertiary, #8b949e);
  cursor: pointer;
  padding: 4px;
}

.modal-close:hover {
  color: var(--text-main, #f0f6fc);
}

.modal-text {
  margin: 0;
  font-size: 13px;
  color: var(--text-main, #f0f6fc);
  line-height: 1.5;
}

.hint-text {
  margin: 0;
  font-size: 12px;
  color: var(--text-tertiary, #8b949e);
  line-height: 1.4;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding-top: 12px;
  border-top: 1px solid var(--border, #30363d);
}

.btn-modal-danger {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  background: #da3633;
  color: #ffffff;
  border: 1px solid rgba(248, 81, 73, 0.4);
  border-radius: var(--radius-sm, 6px);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.15s;
}

.btn-modal-danger:hover:not(:disabled) {
  background: #f85149;
}

.btn-modal-secondary {
  display: inline-flex;
  align-items: center;
  padding: 6px 14px;
  background: var(--bg-secondary, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  color: var(--text-main, #f0f6fc);
  font-size: 13px;
  cursor: pointer;
}

.btn-modal-secondary:hover:not(:disabled) {
  background: var(--bg-tertiary, #21262d);
}

.modal-fade-enter-active,
.modal-fade-leave-active {
  transition:
    opacity 0.15s ease,
    transform 0.15s ease;
}
.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
  transform: scale(0.96);
}

.spin {
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
