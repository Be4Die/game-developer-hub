<template>
  <div class="modal-overlay" @click.self="$emit('cancel')">
    <div class="modal-card">
      <div class="modal-header">
        <div class="modal-title-wrap">
          <ShieldCheck class="icon-md text-warning" />
          <div>
            <h3>Добавить модератора</h3>
            <p class="modal-subtitle">Создание учётной записи сотрудника модерации</p>
          </div>
        </div>
        <button class="modal-close" @click="$emit('cancel')">
          <X class="icon-sm" />
        </button>
      </div>

      <form class="modal-form" @submit.prevent="handleCreate">
        <div class="form-group">
          <label class="form-label" for="mod-login">Логин сотрудника *</label>
          <div class="email-input-group">
            <input
              id="mod-login"
              v-model="form.login"
              type="text"
              class="field-input"
              placeholder="username"
              required
              :disabled="loading"
            />
            <span class="email-domain">@welwise.com</span>
          </div>
          <span class="form-hint">Email формируется автоматически как login@welwise.com</span>
        </div>

        <div class="form-group">
          <label class="form-label" for="mod-name">Отображаемое имя *</label>
          <input
            id="mod-name"
            v-model="form.display_name"
            type="text"
            class="form-input"
            placeholder="Иван Модераторов"
            required
            :disabled="loading"
          />
        </div>

        <div class="form-group">
          <label class="form-label" for="mod-pass">Пароль для входа *</label>
          <input
            id="mod-pass"
            v-model="form.password"
            type="password"
            class="form-input"
            placeholder="Минимум 6 символов"
            required
            minlength="6"
            :disabled="loading"
          />
        </div>

        <div class="modal-actions">
          <button
            type="button"
            class="btn-modal-secondary"
            :disabled="loading"
            @click="$emit('cancel')"
          >
            Отмена
          </button>
          <button type="submit" class="btn-modal-primary" :disabled="loading">
            <Loader2 v-if="loading" class="icon-xs spin" />
            <ShieldCheck v-else class="icon-xs" />
            <span>{{ loading ? 'Создание...' : 'Создать модератора' }}</span>
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue';
import { ShieldCheck, X, Loader2 } from 'lucide-vue-next';
import { createModerator } from '@/entities/user';
import { showToast } from '@/shared/lib';

const emit = defineEmits(['created', 'cancel']);

const loading = ref(false);
const form = reactive({
  login: '',
  display_name: '',
  password: '',
});

async function handleCreate() {
  loading.value = true;
  try {
    const res = await createModerator({
      login: form.login.trim(),
      password: form.password,
      display_name: form.display_name.trim(),
    });
    showToast(`Модератор "${res.user?.display_name || form.display_name}" успешно создан`, 'success');
    emit('created', res.user);
  } catch (err) {
    showToast(err.response?.data?.message || 'Не удалось создать модератора', 'danger');
  } finally {
    loading.value = false;
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
  max-width: 460px;
  box-shadow: 0 16px 32px rgba(0, 0, 0, 0.4);
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.modal-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.modal-title-wrap {
  display: flex;
  align-items: center;
  gap: 12px;
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
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s;
}

.modal-close:hover {
  color: var(--text-main, #f0f6fc);
}

.modal-form {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-main, #f0f6fc);
}

.form-input {
  width: 100%;
  height: 38px;
  padding: 0 12px;
  background: var(--bg-secondary, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  color: var(--text-main, #f0f6fc);
  font-size: 14px;
  outline: none;
  box-sizing: border-box;
  transition: border-color 0.15s;
}

.form-input:focus {
  border-color: var(--primary, #58a6ff);
}

.email-input-group {
  display: flex;
  align-items: center;
  height: 38px;
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  background: var(--bg-secondary, #161b22);
  overflow: hidden;
  box-sizing: border-box;
  transition: border-color 0.15s;
}

.email-input-group:focus-within {
  border-color: var(--primary, #58a6ff);
}

.field-input {
  border: none;
  flex: 1;
  min-width: 0;
  height: 100%;
  padding: 0 12px;
  background: transparent;
  color: var(--text-main, #f0f6fc);
  font-size: 14px;
  outline: none;
}

.email-domain {
  padding: 0 12px;
  height: 100%;
  display: flex;
  align-items: center;
  color: var(--text-tertiary, #8b949e);
  font-size: 13px;
  background: var(--bg-tertiary, #21262d);
  border-left: 1px solid var(--border, #30363d);
  white-space: nowrap;
  user-select: none;
}

.form-hint {
  font-size: 11px;
  color: var(--text-tertiary, #8b949e);
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding-top: 14px;
  border-top: 1px solid var(--border, #30363d);
}

.btn-modal-primary {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 34px;
  padding: 0 16px;
  font-size: 13px;
  font-weight: 500;
  border-radius: var(--radius-sm, 6px);
  cursor: pointer;
  background: var(--primary, #58a6ff);
  color: #ffffff;
  border: none;
  transition: background-color 0.15s;
}

.btn-modal-primary:hover:not(:disabled) {
  background: var(--primary-hover, #79c0ff);
}

.btn-modal-secondary {
  display: inline-flex;
  align-items: center;
  height: 34px;
  padding: 0 16px;
  font-size: 13px;
  border-radius: var(--radius-sm, 6px);
  cursor: pointer;
  background: var(--bg-secondary, #161b22);
  color: var(--text-main, #f0f6fc);
  border: 1px solid var(--border, #30363d);
  transition: background-color 0.15s;
}

.btn-modal-secondary:hover:not(:disabled) {
  background: var(--bg-tertiary, #21262d);
}

.btn-modal-primary:disabled,
.btn-modal-secondary:disabled {
  opacity: 0.4;
  cursor: not-allowed;
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
