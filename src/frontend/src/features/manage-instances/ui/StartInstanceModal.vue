<template>
  <div class="modal-overlay" @click.self="$emit('cancel')">
    <div class="modal card">
      <h3>{{ t('servers.startInstance') }}</h3>
      <div class="form-grid">
        <div class="form-group">
          <label>{{ t('common.version') }} *</label>
          <select v-model="startForm.build_version" class="form-input">
            <option value="" disabled>Select build</option>
            <option v-for="b in availableBuilds" :key="b.build_version" :value="b.build_version">
              {{ b.build_version }}
            </option>
          </select>
        </div>
        <div class="form-group">
          <label>{{ t('common.name') }}</label>
          <input
            v-model="startForm.name"
            type="text"
            class="form-input"
            placeholder="EU-1"
            maxlength="128"
          />
        </div>
        <div class="form-group">
          <label>Max Players</label>
          <input
            v-model.number="startForm.max_players"
            type="number"
            class="form-input"
            min="1"
            placeholder="Default"
          />
        </div>
        <div class="form-group form-group-wide">
          <label>Environment Variables</label>
          <KeyValueEditor v-model="startForm.env_vars" />
        </div>
        <div class="form-group form-group-wide">
          <label>Arguments</label>
          <div class="args-list">
            <div v-for="(arg, i) in startForm.args" :key="i" class="arg-row">
              <input
                v-model="startForm.args[i]"
                type="text"
                class="form-input"
                placeholder="--flag value"
              />
              <button class="arg-remove" @click="startForm.args.splice(i, 1)">&times;</button>
            </div>
            <button class="arg-add" @click="startForm.args.push('')">+ Add argument</button>
          </div>
        </div>
      </div>
      <div v-if="startError" class="start-error">{{ startError }}</div>
      <div class="modal-actions">
        <button
          class="btn-primary"
          :disabled="!startForm.build_version || starting"
          @click="submitStart"
        >
          {{ starting ? t('common.loading') : t('servers.startInstance') }}
        </button>
        <button class="btn-outline" @click="$emit('cancel')">{{ t('common.cancel') }}</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue';
import { useI18n } from 'vue-i18n';
import { KeyValueEditor } from '@/shared/ui';
import { startInstance } from '@/entities/instance';
import { showToast } from '@/shared/lib';

const { t } = useI18n();

const props = defineProps({
  gameId: { type: [String, Number], required: true },
  availableBuilds: { type: Array, default: () => [] },
});

const emit = defineEmits(['started', 'cancel']);

const starting = ref(false);
const startError = ref(null);

const startForm = reactive({
  build_version: '',
  name: '',
  max_players: null,
  env_vars: {},
  args: [],
});

async function submitStart() {
  starting.value = true;
  startError.value = null;
  const payload = {
    build_version: startForm.build_version,
  };
  if (startForm.name) payload.name = startForm.name;
  if (startForm.max_players) payload.max_players = startForm.max_players;
  if (Object.keys(startForm.env_vars).length) {
    payload.env_vars = startForm.env_vars;
  }
  if (startForm.args.filter(Boolean).length) {
    payload.args = startForm.args.filter(Boolean);
  }

  try {
    await startInstance(props.gameId, payload);
    showToast('Инстанс запускается...', 'info');
    Object.assign(startForm, {
      build_version: '',
      name: '',
      max_players: null,
      env_vars: {},
      args: [],
    });
    emit('started');
  } catch (e) {
    if (e.response?.status === 409) {
      startError.value = 'Недостаточно ресурсов на доступных нодах';
    } else {
      startError.value = e.response?.data?.message ?? 'Ошибка запуска инстанса';
    }
  } finally {
    starting.value = false;
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
  max-width: 540px;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  padding: 24px;
  max-height: 90vh;
  overflow-y: auto;
}

.modal h3 {
  margin: 0 0 20px 0;
  font-size: 1.1rem;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
  margin-bottom: 20px;
}

.form-group-wide {
  grid-column: 1 / -1;
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

.args-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.arg-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.arg-remove {
  background: none;
  border: none;
  color: var(--danger);
  font-size: 1.2rem;
  cursor: pointer;
  padding: 0 4px;
}

.arg-add {
  align-self: flex-start;
  background: none;
  border: 1px dashed var(--border);
  color: var(--primary);
  padding: 4px 12px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-size: 0.85rem;
}

.start-error {
  margin-bottom: 16px;
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
}
</style>
