<template>
  <Teleport to="body">
    <div class="modal-overlay" @click.self="handleCancel">
      <div class="modal card role-modal">
        <div class="modal-header">
          <h3>{{ targetRole === 'storage' ? 'Переход в режим Storage' : 'Переход в режим Compute' }}</h3>
          <button class="close-btn" :disabled="processing" @click="handleCancel">&times;</button>
        </div>

        <!-- Переход в Storage -->
        <div v-if="targetRole === 'storage'" class="modal-body">
          <p class="modal-desc">
            На ноде запущены игровые серверы ({{ instances.length }} шт.). В режиме <strong>Storage</strong> игровые комнаты не поддерживаются.
          </p>
          <p class="danger-hint">
            Все активные серверы на этой ноде будут остановлены и удалены.
          </p>
        </div>

        <!-- Переход в Compute -->
        <div v-else-if="targetRole === 'compute'" class="modal-body">
          <p class="modal-desc">
            На ноде есть сервисы хранения: <strong>{{ servicesList }}</strong>.
            Перед переключением будет автоматически создан бэкап.
          </p>

          <div class="options-list">
            <label class="option-item" :class="{ selected: selectedAction === 'stop' }">
              <input type="radio" v-model="selectedAction" value="stop" :disabled="processing" />
              <div class="option-content">
                <span class="option-title">Остановить базы данных</span>
                <span class="option-subtitle">Данные сохранятся на диске, сервисы можно включить позже</span>
              </div>
            </label>

            <label class="option-item" :class="{ selected: selectedAction === 'delete' }">
              <input type="radio" v-model="selectedAction" value="delete" :disabled="processing" />
              <div class="option-content">
                <span class="option-title text-danger">Удалить базы данных</span>
                <span class="option-subtitle">Контейнеры и файлы данных будут удалены с диска</span>
              </div>
            </label>
          </div>
        </div>

        <div class="modal-actions">
          <button class="btn-outline" :disabled="processing" @click="handleCancel">
            Отмена
          </button>

          <button
            v-if="targetRole === 'storage'"
            class="btn-danger"
            :disabled="processing"
            @click="confirmStorage"
          >
            {{ processing ? 'Завершение...' : 'Завершить серверы и перейти' }}
          </button>

          <button
            v-else-if="targetRole === 'compute'"
            :class="selectedAction === 'delete' ? 'btn-danger' : 'btn-primary'"
            :disabled="processing"
            @click="confirmCompute"
          >
            {{ processing ? 'Выполняется...' : (selectedAction === 'delete' ? 'Удалить БД и перейти' : 'Остановить БД и перейти') }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, computed } from 'vue';
import {
  STORAGE_TRANSITION_STOP,
  STORAGE_TRANSITION_DELETE,
  COMPUTE_TRANSITION_TERMINATE,
} from '@/entities/node';

const props = defineProps({
  node: {
    type: Object,
    required: true,
  },
  targetRole: {
    type: String,
    required: true,
  },
  instances: {
    type: Array,
    default: () => [],
  },
  services: {
    type: Array,
    default: () => [],
  },
  processing: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(['close', 'confirm']);

const selectedAction = ref('stop');

const servicesList = computed(() => {
  if (!props.services || !props.services.length) return '';
  return props.services.map((s) => s.name).join(', ');
});

function handleCancel() {
  if (props.processing) return;
  emit('close');
}

function confirmStorage() {
  emit('confirm', {
    compute_action: COMPUTE_TRANSITION_TERMINATE,
  });
}

function confirmCompute() {
  emit('confirm', {
    storage_action: selectedAction.value === 'delete' ? STORAGE_TRANSITION_DELETE : STORAGE_TRANSITION_STOP,
  });
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.75);
  backdrop-filter: blur(2px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 16px;
}

.role-modal {
  width: 100%;
  max-width: 450px;
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-md, 8px);
  padding: 24px;
  box-shadow: 0 16px 32px rgba(0, 0, 0, 0.4);
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
}

.modal-header h3 {
  margin: 0;
  font-size: 1.12rem;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.35rem;
  line-height: 1;
  color: var(--text-muted, #8b949e);
  cursor: pointer;
  padding: 0 4px;
}

.close-btn:hover:not(:disabled) {
  color: var(--text-main, #f0f6fc);
}

.modal-body {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.modal-desc {
  margin: 0;
  font-size: 0.88rem;
  color: var(--text-muted, #8b949e);
  line-height: 1.45;
}

.modal-desc strong {
  color: var(--text-main, #f0f6fc);
}

.danger-hint {
  margin: 0;
  font-size: 0.84rem;
  color: var(--danger, #f85149);
  line-height: 1.4;
}

.options-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-top: 4px;
}

.option-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 10px 12px;
  background: var(--bg-app, #0d1117);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  cursor: pointer;
  transition: all 0.15s ease;
}

.option-item:hover {
  border-color: var(--border-secondary, #484f58);
}

.option-item.selected {
  border-color: var(--primary, #58a6ff);
  background: rgba(88, 166, 255, 0.06);
}

.option-item input[type="radio"] {
  margin-top: 3px;
  accent-color: var(--primary, #58a6ff);
  cursor: pointer;
}

.option-content {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.option-title {
  font-size: 0.88rem;
  font-weight: 500;
  color: var(--text-main, #f0f6fc);
}

.option-subtitle {
  font-size: 0.78rem;
  color: var(--text-muted, #8b949e);
  line-height: 1.35;
}

.text-danger {
  color: var(--danger, #f85149);
}

.modal-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 20px;
}
</style>
