<template>
  <div class="auth-card">
    <div v-if="authState.error" class="alert alert-danger">
      {{ authState.error }}
    </div>
    <div v-if="successMessage" class="alert alert-success">
      {{ successMessage }}
    </div>

    <!-- Login Form -->
    <form v-if="mode === 'login'" class="auth-form" @submit.prevent="handleLogin">
      <h2>{{ t('auth.loginTitle') }}</h2>
      <div class="form-group">
        <label for="email">{{ t('auth.email') }}</label>
        <input
          id="email"
          v-model="form.email"
          type="email"
          :placeholder="t('auth.emailPlaceholder')"
          required
          @input="validateEmail"
        />
        <span v-if="emailError" class="field-error">{{ emailError }}</span>
      </div>
      <div class="form-group">
        <label for="password">{{ t('auth.password') }}</label>
        <input
          id="password"
          v-model="form.password"
          type="password"
          :placeholder="t('auth.passwordPlaceholder')"
          required
        />
      </div>
      <button type="submit" class="btn btn-primary btn-full" :disabled="authState.loading">
        {{ authState.loading ? t('common.loading') : t('auth.signIn') }}
      </button>
      <p class="auth-switch">
        {{ t('auth.noAccount') }}
        <a href="#" @click.prevent="mode = 'register'">{{ t('auth.signUp') }}</a>
      </p>
    </form>

    <!-- Registration Form -->
    <form v-if="mode === 'register'" class="auth-form" @submit.prevent="handleRegister">
      <h2>{{ t('auth.registerTitle') }}</h2>
      <div class="form-group">
        <label for="reg-name">{{ t('auth.displayName') }}</label>
        <input
          id="reg-name"
          v-model="form.display_name"
          type="text"
          :placeholder="t('auth.displayNamePlaceholder')"
          required
        />
      </div>
      <div class="form-group">
        <label for="reg-email">{{ t('auth.email') }}</label>
        <input
          id="reg-email"
          v-model="form.email"
          type="email"
          :placeholder="t('auth.emailPlaceholder')"
          required
          @input="validateEmail"
        />
        <span v-if="emailError" class="field-error">{{ emailError }}</span>
      </div>
      <div class="form-group">
        <label for="reg-password">{{ t('auth.password') }}</label>
        <input
          id="reg-password"
          v-model="form.password"
          type="password"
          :placeholder="t('auth.passwordPlaceholder')"
          required
          minlength="6"
          @input="checkPasswordStrength"
        />
        <div v-if="form.password" class="password-strength">
          <div class="strength-bar">
            <div
              class="strength-fill"
              :class="passwordStrength.class"
              :style="{ width: passwordStrength.percent + '%' }"
            ></div>
          </div>
          <span class="strength-label" :class="passwordStrength.class">{{
            passwordStrength.text
          }}</span>
        </div>
      </div>
      <button type="submit" class="btn btn-primary btn-full" :disabled="authState.loading">
        {{ authState.loading ? t('common.loading') : t('auth.signUp') }}
      </button>
      <p class="auth-switch">
        {{ t('auth.haveAccount') }}
        <a href="#" @click.prevent="mode = 'login'">{{ t('auth.signIn') }}</a>
      </p>
    </form>

    <!-- Email Verification Form -->
    <form v-if="mode === 'verify'" class="auth-form" @submit.prevent="handleVerify">
      <h2>Подтверждение email</h2>
      <p class="verify-info">
        На email <strong>{{ form.email }}</strong> отправлен код подтверждения.<br />
        Введите 6-значный код из письма.
      </p>
      <div class="form-group">
        <label for="verify-code">Код подтверждения</label>
        <input
          id="verify-code"
          v-model="form.verification_code"
          type="text"
          placeholder="000000"
          required
          maxlength="6"
          pattern="\d{6}"
        />
      </div>
      <button type="submit" class="btn btn-primary btn-full" :disabled="authState.loading">
        {{ authState.loading ? t('common.loading') : t('common.confirm') }}
      </button>
      <p class="auth-switch">
        Не получили код?
        <a href="#" @click.prevent="resendCode">Отправить повторно</a>
      </p>
      <p class="auth-switch">
        <a href="#" @click.prevent="mode = 'login'">← {{ t('auth.backToLogin') }}</a>
      </p>
    </form>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { useAuth, verifyEmail, resendVerificationEmail } from '@/entities/user';

const emit = defineEmits(['success']);

const { t } = useI18n();
const router = useRouter();
const { state: authState, login, register } = useAuth();

const mode = ref('login');
const successMessage = ref('');
const emailError = ref('');
const passwordStrength = ref({ percent: 0, text: '', class: '' });
const form = reactive({
  email: '',
  password: '',
  display_name: '',
  verification_code: '',
});

const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

function validateEmail() {
  if (!form.email) {
    emailError.value = '';
    return;
  }
  if (!EMAIL_REGEX.test(form.email)) {
    emailError.value = t('auth.errorInvalidCredentials');
  } else {
    emailError.value = '';
  }
}

function checkPasswordStrength() {
  const pwd = form.password;
  if (!pwd) {
    passwordStrength.value = { percent: 0, text: '', class: '' };
    return;
  }
  let score = 0;
  if (pwd.length >= 6) score++;
  if (pwd.length >= 10) score++;
  if (/[a-z]/.test(pwd) && /[A-Z]/.test(pwd)) score++;
  if (/\d/.test(pwd)) score++;
  if (/[^a-zA-Z0-9]/.test(pwd)) score++;

  if (score <= 2) {
    passwordStrength.value = { percent: 35, text: '•••', class: 'weak' };
  } else if (score === 3) {
    passwordStrength.value = { percent: 65, text: '••••', class: 'medium' };
  } else {
    passwordStrength.value = { percent: 100, text: '•••••', class: 'strong' };
  }
}

async function handleLogin() {
  try {
    const res = await login({ email: form.email, password: form.password });
    emit('success');

    const role = res.user?.role || authState.user?.role;
    if (role === 'USER_ROLE_ADMIN' || role === 3) {
      router.push('/admin/dashboard');
    } else if (role === 'USER_ROLE_MODERATOR' || role === 2) {
      router.push('/moderator/queue');
    } else {
      router.push('/projects');
    }
  } catch (err) {
    // handled in state.error
  }
}

async function handleRegister() {
  try {
    await register({
      email: form.email,
      password: form.password,
      display_name: form.display_name,
    });
    successMessage.value = t('auth.registerSuccess');
    mode.value = 'verify';
  } catch (err) {
    // handled in state.error
  }
}

async function handleVerify() {
  try {
    await verifyEmail(form.verification_code);
    successMessage.value = t('auth.loginSuccess');
    mode.value = 'login';
    form.password = '';
  } catch (err) {
    authState.error = err.response?.data?.message || t('common.error');
  }
}

async function resendCode() {
  try {
    await resendVerificationEmail(form.email);
    successMessage.value = t('auth.resetSuccess');
  } catch (err) {
    authState.error = t('common.error');
  }
}
</script>

<style scoped>
.auth-card {
  width: 100%;
  max-width: 420px;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  padding: 40px 32px;
  box-shadow: var(--shadow-lg);
}

.auth-form h2 {
  margin: 0 0 24px;
  font-size: 1.25rem;
  font-weight: 700;
  text-align: center;
  letter-spacing: -0.3px;
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
  text-decoration: none;
}

.verify-info {
  text-align: center;
  color: var(--text-muted);
  margin-bottom: 20px;
  font-size: 0.9rem;
  line-height: 1.5;
}

.field-error {
  display: block;
  color: var(--danger);
  font-size: 0.8rem;
  margin-top: 4px;
}

.password-strength {
  margin-top: 8px;
}

.strength-bar {
  height: 4px;
  background: var(--bg-app);
  border-radius: 2px;
  overflow: hidden;
}

.strength-fill {
  height: 100%;
  border-radius: 2px;
  transition:
    width 0.3s,
    background 0.3s;
}

.strength-fill.weak {
  background: var(--danger);
}

.strength-fill.medium {
  background: var(--warning);
}

.strength-fill.strong {
  background: var(--success);
}

.strength-label {
  font-size: 0.75rem;
  margin-top: 4px;
  display: block;
}

.strength-label.weak {
  color: var(--danger);
}

.strength-label.medium {
  color: var(--warning);
}

.strength-label.strong {
  color: var(--success);
}
</style>
