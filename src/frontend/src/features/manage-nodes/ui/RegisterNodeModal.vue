<template>
  <div class="modal-overlay" @click.self="$emit('cancel')">
    <div class="modal card">
      <h3>Подключить ноду</h3>
      <div class="tabs">
        <button
          class="tab-btn"
          :class="{ active: registerTab === 'available' }"
          @click="registerTab = 'available'"
        >
          Доступные ноды
        </button>
        <button
          class="tab-btn"
          :class="{ active: registerTab === 'manual' }"
          @click="registerTab = 'manual'"
        >
          Ручной ввод
        </button>
      </div>

      <!-- Доступные ноды (auto-discovery) -->
      <div v-if="registerTab === 'available'" class="form-content">
        <div class="form-group">
          <label>Доступная нода *</label>
          <select v-model="availableForm.node_id" class="form-input">
            <option value="" disabled>Выберите ноду</option>
            <option v-for="n in availableNodes" :key="n.id" :value="n.id">
              {{ n.address }} (ID: {{ n.id }})
            </option>
          </select>
          <p v-if="!availableNodes.length" class="hint">
            Нет доступных нод для подключения. Убедитесь, что нода запущена в режиме auto-discovery.
          </p>
        </div>
        <div class="form-group">
          <label>Ключ авторизации (API-ключ ноды) *</label>
          <input
            v-model="availableForm.token"
            type="text"
            class="form-input"
            placeholder="dev-api-key-for-local-testing"
          />
        </div>
        <p class="hint">
          Ноды в этом списке самостоятельно анонсировали себя оркестратору и ожидают авторизации.
          Введите API-ключ ноды (NODE_API_KEY) для подключения.
        </p>
      </div>

      <!-- Ручной ввод -->
      <div v-if="registerTab === 'manual'" class="form-content">
        <div class="form-group">
          <label>Адрес (host:port) *</label>
          <input
            v-model="manualForm.address"
            type="text"
            class="form-input"
            placeholder="192.168.1.100:44044"
          />
        </div>
        <div class="form-group">
          <label>Ключ авторизации (API-ключ ноды) *</label>
          <input
            v-model="manualForm.token"
            type="text"
            class="form-input"
            placeholder="dev-api-key-for-local-testing"
          />
        </div>
        <div class="form-group">
          <label>Регион (опционально)</label>
          <input v-model="manualForm.region" type="text" class="form-input" placeholder="EU" />
        </div>
        <p class="hint">Введите адрес ноды и её API-ключ (NODE_API_KEY) для подключения.</p>
      </div>

      <div v-if="registerError" class="form-error">
        {{ registerError }}
      </div>

      <div class="modal-actions">
        <button class="btn-primary" :disabled="registering || !canSubmit" @click="submitRegister">
          {{ registering ? 'Подключение...' : 'Подключить' }}
        </button>
        <button class="btn-outline" @click="$emit('cancel')">Отмена</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue';
import { registerNode } from '@/entities/node';
import { showToast } from '@/shared/lib';

defineProps({
  availableNodes: { type: Array, default: () => [] },
});

const emit = defineEmits(['registered', 'cancel']);

const registerTab = ref('available');
const registering = ref(false);
const registerError = ref(null);

const availableForm = ref({
  node_id: '',
  token: '',
});

const manualForm = ref({
  address: '',
  token: '',
  region: '',
});

const canSubmit = computed(() => {
  if (registerTab.value === 'available') {
    return availableForm.value.node_id && availableForm.value.token;
  }
  return manualForm.value.address && manualForm.value.token;
});

async function submitRegister() {
  registering.value = true;
  registerError.value = null;

  try {
    let payload;
    if (registerTab.value === 'available') {
      payload = {
        node_id: availableForm.value.node_id,
        token: availableForm.value.token,
      };
    } else {
      payload = {
        address: manualForm.value.address,
        token: manualForm.value.token,
        region: manualForm.value.region || undefined,
      };
    }
    await registerNode(payload);
    showToast('Нода успешно подключена', 'success');
    emit('registered');
  } catch (e) {
    if (e.response?.status === 401) {
      registerError.value = 'Неверный ключ авторизации ноды';
    } else if (e.response?.status === 409) {
      registerError.value = 'Нода с таким адресом уже зарегистрирована';
    } else {
      registerError.value = e.response?.data?.message ?? 'Ошибка подключения ноды';
    }
  } finally {
    registering.value = false;
  }
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
  padding: 16px;
}

.modal {
  width: 100%;
  max-width: 480px;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  padding: 24px;
}

.modal h3 {
  margin: 0 0 16px 0;
  font-size: 1.1rem;
}

.tabs {
  display: flex;
  gap: 4px;
  margin-bottom: 20px;
  border-bottom: 1px solid var(--border);
}

.tab-btn {
  padding: 8px 16px;
  border: none;
  background: transparent;
  color: var(--text-muted);
  font-size: 0.88rem;
  font-weight: 500;
  cursor: pointer;
  border-bottom: 2px solid transparent;
  transition: all 0.15s;
}

.tab-btn:hover {
  color: var(--text-main);
}

.tab-btn.active {
  color: var(--primary);
  border-bottom-color: var(--primary);
  font-weight: 600;
}

.form-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-group label {
  display: block;
  font-size: 0.82rem;
  font-weight: 600;
  color: var(--text-muted);
  margin-bottom: 6px;
}

.form-input {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  background: var(--bg-app);
  color: var(--text-main);
  font-size: 0.9rem;
  outline: none;
  box-sizing: border-box;
}

.hint {
  font-size: 0.8rem;
  color: var(--text-muted);
  margin: 0;
  line-height: 1.4;
}

.form-error {
  margin-top: 16px;
  padding: 10px 14px;
  background: var(--danger-light);
  color: var(--danger);
  border-radius: var(--radius-md);
  font-size: 0.85rem;
}

.modal-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  margin-top: 24px;
}
</style>
