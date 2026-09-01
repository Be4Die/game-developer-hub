<template>
  <div class="card">
    <h2>Создать модератора</h2>
    <form class="form-grid" @submit.prevent="handleCreate">
      <div class="form-group">
        <label for="login">Логин</label>
        <div class="email-input-group">
          <input id="login" v-model="form.login" type="text" placeholder="username" required />
          <span class="email-domain">@welwise.com</span>
        </div>
        <span class="form-hint">Email будет сформирован автоматически</span>
      </div>
      <div class="form-group">
        <label for="display_name">Имя</label>
        <input
          id="display_name"
          v-model="form.display_name"
          type="text"
          placeholder="Имя модератора"
          required
        />
      </div>
      <div class="form-group">
        <label for="password">Пароль</label>
        <input
          id="password"
          v-model="form.password"
          type="password"
          placeholder="Минимум 6 символов"
          required
          minlength="6"
        />
      </div>
      <div class="form-group form-actions">
        <button type="submit" class="btn btn-primary" :disabled="loading">
          {{ loading ? 'Создание...' : 'Создать' }}
        </button>
      </div>
    </form>
    <div v-if="createdEmail" class="alert alert-success" style="margin-top: 16px">
      <strong>Email для входа:</strong> {{ createdEmail }}
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue';
import { createModerator } from '@/entities/user';
import { showToast } from '@/shared/lib';

const emit = defineEmits(['created']);

const loading = ref(false);
const createdEmail = ref('');
const form = reactive({
  login: '',
  display_name: '',
  password: '',
});

async function handleCreate() {
  loading.value = true;
  try {
    const res = await createModerator({
      login: form.login,
      password: form.password,
      display_name: form.display_name,
    });
    createdEmail.value = res.user.email;
    showToast(`Модератор "${res.user.display_name}" успешно создан`, 'success');
    form.login = '';
    form.display_name = '';
    form.password = '';
    emit('created', res.user);
  } catch (err) {
    showToast(err.response?.data?.message || 'Не удалось создать модератора', 'error');
  } finally {
    loading.value = false;
  }
}
</script>

<style scoped>
.card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  padding: 24px;
  margin-bottom: 24px;
}

.card h2 {
  margin: 0 0 20px 0;
  font-size: 1.1rem;
}

.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.form-actions {
  justify-content: flex-end;
  padding-top: 8px;
  display: flex;
}

.email-input-group {
  display: flex;
  align-items: center;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  background: var(--bg-input);
  overflow: hidden;
  transition:
    border-color 0.2s,
    box-shadow 0.2s;
}

.email-input-group:focus-within {
  border-color: var(--primary);
  box-shadow: 0 0 0 3px var(--primary-light);
}

.email-input-group input {
  border: none;
  flex: 1;
  min-width: 0;
  padding: 10px 14px;
  background: transparent;
  color: var(--text-main);
  font-size: 0.95rem;
  font-family: inherit;
  outline: none;
}

.email-domain {
  padding: 10px 12px;
  color: var(--text-muted);
  font-size: 0.95rem;
  background: var(--bg-app);
  border-left: 1px solid var(--border);
  white-space: nowrap;
  user-select: none;
}

.form-hint {
  margin-top: 4px;
  font-size: 0.75rem;
  color: var(--text-muted);
}

@media (max-width: 768px) {
  .form-grid {
    grid-template-columns: 1fr;
  }
}
</style>
