<template>
  <div class="auth-page">
    <div class="auth-card">
      <div class="auth-header">
        <h1>Welwise Games</h1>
        <p>Войдите в аккаунт или создайте новый</p>
      </div>

      <div v-if="authState.error" class="auth-error">{{ authState.error }}</div>
      <div v-if="successMessage" class="auth-success">{{ successMessage }}</div>

      <!-- Login Form -->
      <form v-if="mode === 'login'" @submit.prevent="handleLogin" class="auth-form">
        <h2>Вход</h2>
        <div class="form-group">
          <label for="email">Email</label>
          <input id="email" v-model="form.email" type="email" placeholder="you@example.com" required />
        </div>
        <div class="form-group">
          <label for="password">Пароль</label>
          <input id="password" v-model="form.password" type="password" placeholder="Введите пароль" required />
        </div>
        <button type="submit" class="btn-primary btn-full" :disabled="authState.loading">
          {{ authState.loading ? 'Вход...' : 'Войти' }}
        </button>
        <p class="auth-switch">
          Нет аккаунта? <a href="#" @click.prevent="mode = 'register'">Зарегистрироваться</a>
        </p>
      </form>

      <!-- Registration Form -->
      <form v-if="mode === 'register'" @submit.prevent="handleRegister" class="auth-form">
        <h2>Регистрация</h2>
        <div class="form-group">
          <label for="reg-name">Имя</label>
          <input id="reg-name" v-model="form.display_name" type="text" placeholder="Как вас зовут?" required />
        </div>
        <div class="form-group">
          <label for="reg-email">Email</label>
          <input id="reg-email" v-model="form.email" type="email" placeholder="you@example.com" required />
        </div>
        <div class="form-group">
          <label for="reg-password">Пароль</label>
          <input id="reg-password" v-model="form.password" type="password" placeholder="Минимум 6 символов" required minlength="6" />
        </div>
        <button type="submit" class="btn-primary btn-full" :disabled="authState.loading">
          {{ authState.loading ? 'Регистрация...' : 'Зарегистрироваться' }}
        </button>
        <p class="auth-switch">
          Уже есть аккаунт? <a href="#" @click.prevent="mode = 'login'">Войти</a>
        </p>
      </form>

      <!-- Email Verification Form -->
      <form v-if="mode === 'verify'" @submit.prevent="handleVerify" class="auth-form">
        <h2>Подтверждение email</h2>
        <p class="verify-info">
          На email <strong>{{ form.email }}</strong> отправлен код подтверждения.<br>
          Введите 6-значный код из письма.
        </p>
        <div class="form-group">
          <label for="verify-code">Код подтверждения</label>
          <input id="verify-code" v-model="form.verification_code" type="text" placeholder="000000" required maxlength="6" pattern="\d{6}" />
        </div>
        <button type="submit" class="btn-primary btn-full" :disabled="authState.loading">
          {{ authState.loading ? 'Проверка...' : 'Подтвердить' }}
        </button>
        <p class="auth-switch">
          Не получили код? <a href="#" @click.prevent="resendCode">Отправить повторно</a>
        </p>
        <p class="auth-switch">
          <a href="#" @click.prevent="mode = 'login'">← Назад к входу</a>
        </p>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAuth } from '../store/auth'
import { verifyEmail, resendVerificationEmail } from '../api/sso'

const router = useRouter()
const { state: authState, login, register } = useAuth()

const mode = ref('login')
const successMessage = ref('')
const form = reactive({
  email: '',
  password: '',
  display_name: '',
  verification_code: '',
})

async function handleLogin() {
  try {
    await login({ email: form.email, password: form.password })
    router.push('/projects')
  } catch (err) {
    // Ошибка уже установлена в authState.error
  }
}

async function handleRegister() {
  try {
    await register({ email: form.email, password: form.password, display_name: form.display_name })
    successMessage.value = 'Регистрация успешна! Подтвердите email.'
    mode.value = 'verify'
  } catch (err) {
    // Ошибка уже установлена в authState.error
  }
}

async function handleVerify() {
  try {
    await verifyEmail(form.verification_code)
    successMessage.value = 'Email подтвержден! Теперь войдите в систему.'
    mode.value = 'login'
    form.password = ''
  } catch (err) {
    authState.error = err.response?.data?.message || 'Неверный код подтверждения'
  }
}

async function resendCode() {
  try {
    await resendVerificationEmail(form.email)
    successMessage.value = 'Код отправлен повторно!'
  } catch (err) {
    authState.error = 'Не удалось отправить код'
  }
}
</script>

<style scoped>
.auth-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-app);
  padding: 24px;
}

.auth-card {
  width: 100%;
  max-width: 420px;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  padding: 40px 32px;
}

.auth-header {
  text-align: center;
  margin-bottom: 32px;
}

.auth-header h1 {
  font-size: 1.5rem;
  color: var(--primary);
  margin: 0 0 8px;
}

.auth-header p {
  color: var(--text-muted);
  margin: 0;
  font-size: 0.9rem;
}

.auth-form h2 {
  margin: 0 0 24px;
  font-size: 1.25rem;
  text-align: center;
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 6px;
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-muted);
}

.form-group input {
  width: 100%;
  padding: 10px 14px;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  background: var(--bg-input);
  color: var(--text-main);
  font-size: 0.95rem;
  transition: border-color 0.2s;
}

.form-group input:focus {
  outline: none;
  border-color: var(--primary);
}

.btn-full {
  width: 100%;
  justify-content: center;
  margin-top: 8px;
}

.btn-full:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.auth-switch {
  text-align: center;
  margin-top: 20px;
  font-size: 0.9rem;
  color: var(--text-muted);
}

.auth-switch a {
  color: var(--primary);
  font-weight: 600;
}

.auth-error {
  background: var(--danger-light);
  color: var(--danger);
  padding: 10px 14px;
  border-radius: var(--radius-sm);
  margin-bottom: 16px;
  font-size: 0.9rem;
  text-align: center;
}

.auth-success {
  background: var(--success-light);
  color: var(--success);
  padding: 10px 14px;
  border-radius: var(--radius-sm);
  margin-bottom: 16px;
  font-size: 0.9rem;
  text-align: center;
}

.verify-info {
  text-align: center;
  color: var(--text-muted);
  margin-bottom: 20px;
  font-size: 0.9rem;
  line-height: 1.5;
}
</style>
