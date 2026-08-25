<template>
  <div class="profile-page">
    <div class="profile-container">
      <!-- Карточка информации об аккаунте (без псевдо-аватара) -->
      <section class="user-info-card">
        <div class="user-main-info">
          <span class="user-title">{{ userDisplayName }}</span>
          <span class="role-badge" :class="roleBadgeClass">
            <Shield class="icon-xs" v-if="isModeratorOrAdmin" />
            <Code2 class="icon-xs" v-else />
            {{ localizedRoleName }}
          </span>
        </div>
        <div class="user-meta-details">
          <div class="detail-item">
            <Mail class="icon-xs text-muted" />
            <span>{{ userEmail }}</span>
          </div>
          <div class="detail-item" v-if="registeredDate">
            <Calendar class="icon-xs text-muted" />
            <span>{{ t('profile.registeredAt') }}: {{ registeredDate }}</span>
          </div>
        </div>
      </section>

      <!-- 2-колоночная адаптивная сетка: строка 1 (Имя + Язык), строка 2 (Пароль + Тема) -->
      <div class="profile-grid">
        <!-- РЯД 1 / КОЛОНКА 1: Изменение отображаемого имени -->
        <section class="profile-card card-name" v-if="!isModeratorOrAdmin">
          <h2 class="card-title">
            <UserCheck class="icon-sm text-primary" />
            {{ t('profile.editProfile') }}
          </h2>

          <form @submit.prevent="handleUpdateDisplayName" class="inline-name-form">
            <input
              id="displayNameInput"
              type="text"
              v-model="displayNameForm"
              class="form-input field-name-input"
              :placeholder="t('auth.displayNamePlaceholder')"
              :disabled="nameSaving"
              required
            />
            <button
              type="submit"
              class="btn-primary btn-save-name"
              :disabled="nameSaving || !isNameChanged"
            >
              <Loader2 class="icon-sm spin" v-if="nameSaving" />
              <Check class="icon-sm" v-else />
              <span>{{ nameSaving ? t('common.saving') : t('profile.updateProfileBtn') }}</span>
            </button>
          </form>
        </section>

        <!-- Для модераторов/администраторов вместо имени (РЯД 1 / КОЛОНКА 1) -->
        <div class="info-alert card-name" v-else>
          <Info class="icon-sm alert-icon" />
          <div class="alert-text">
            {{ t('profile.readonlyNotice') }}
          </div>
        </div>

        <!-- РЯД 1 / КОЛОНКА 2: Язык интерфейса -->
        <section class="profile-card card-lang">
          <h2 class="card-title">
            <Languages class="icon-sm text-primary" />
            {{ t('profile.interfaceLanguage') }}
          </h2>

          <div class="languages-grid">
            <button
              type="button"
              class="lang-card"
              :class="{ active: currentLocale === 'ru' }"
              @click="selectLanguage('ru')"
            >
              <span class="lang-flag">🇷🇺</span>
              <div class="lang-text">
                <span class="lang-name">Русский</span>
                <span class="lang-sub">Russian</span>
              </div>
              <Check class="icon-sm lang-check" v-if="currentLocale === 'ru'" />
            </button>

            <button
              type="button"
              class="lang-card"
              :class="{ active: currentLocale === 'en' }"
              @click="selectLanguage('en')"
            >
              <span class="lang-flag">🇬🇧</span>
              <div class="lang-text">
                <span class="lang-name">English</span>
                <span class="lang-sub">Английский</span>
              </div>
              <Check class="icon-sm lang-check" v-if="currentLocale === 'en'" />
            </button>
          </div>
        </section>

        <!-- РЯД 2 / КОЛОНКА 1: Смена пароля -->
        <section class="profile-card card-password" v-if="!isModeratorOrAdmin">
          <h2 class="card-title">
            <Lock class="icon-sm text-primary" />
            {{ t('profile.security') }}
          </h2>

          <form @submit.prevent="handleChangePassword" class="password-form">
            <div class="form-group">
              <label class="form-label" for="currentPasswordInput">
                {{ t('auth.currentPassword') }}
              </label>
              <input
                id="currentPasswordInput"
                type="password"
                v-model="passwordForm.currentPassword"
                class="form-input"
                :placeholder="t('auth.passwordPlaceholder')"
                :disabled="passwordSaving"
                required
              />
            </div>

            <div class="form-row-2">
              <div class="form-group">
                <label class="form-label" for="newPasswordInput">
                  {{ t('auth.newPassword') }}
                </label>
                <input
                  id="newPasswordInput"
                  type="password"
                  v-model="passwordForm.newPassword"
                  class="form-input"
                  :placeholder="t('auth.passwordPlaceholder')"
                  :disabled="passwordSaving"
                  minlength="6"
                  required
                />
              </div>

              <div class="form-group">
                <label class="form-label" for="confirmPasswordInput">
                  {{ t('auth.confirmPassword') }}
                </label>
                <input
                  id="confirmPasswordInput"
                  type="password"
                  v-model="passwordForm.confirmPassword"
                  class="form-input"
                  :placeholder="t('auth.passwordPlaceholder')"
                  :disabled="passwordSaving"
                  minlength="6"
                  required
                />
              </div>
            </div>

            <div class="form-actions-full">
              <button
                type="submit"
                class="btn-primary btn-password-submit"
                :disabled="passwordSaving || !isPasswordFormFilled"
              >
                <Loader2 class="icon-sm spin" v-if="passwordSaving" />
                <Key class="icon-sm" v-else />
                <span>{{ passwordSaving ? t('common.saving') : t('profile.changePasswordBtn') }}</span>
              </button>
            </div>
          </form>
        </section>

        <!-- Пустой заполнитель для модераторов/админов если нужно -->
        <div v-else></div>

        <!-- РЯД 2 / КОЛОНКА 2: Тема оформления -->
        <section class="profile-card card-theme">
          <h2 class="card-title">
            <Palette class="icon-sm text-primary" />
            {{ t('profile.theme') }}
          </h2>

          <ThemeCardSelector class="theme-selector-wrap" />
        </section>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, reactive } from 'vue';
import { useI18n } from 'vue-i18n';
import {
  Mail,
  Calendar,
  Lock,
  Key,
  Check,
  Loader2,
  UserCheck,
  Shield,
  Code2,
  Languages,
  Palette,
  Info,
} from 'lucide-vue-next';
import { useAuth, updateProfile, changePassword } from '@/entities/user';
import { ThemeCardSelector } from '@/features/theme-switcher';
import { showToast, setLocale, formatDate } from '@/shared/lib';

const { t, locale } = useI18n();
const { state: authState, updateCurrentUser } = useAuth();

const currentLocale = computed(() => locale.value);

const user = computed(() => authState.user || {});
const userDisplayName = computed(
  () => user.value.display_name || user.value.email?.split('@')[0] || t('roles.user')
);
const userEmail = computed(() => user.value.email || '');

const isModeratorOrAdmin = computed(() => {
  const role = user.value.role;
  return (
    role === 'USER_ROLE_MODERATOR' ||
    role === 'USER_ROLE_ADMIN' ||
    role === 'moderator' ||
    role === 'admin' ||
    role === 2 ||
    role === 3
  );
});

const localizedRoleName = computed(() => {
  const role = user.value.role;
  if (role === 'USER_ROLE_ADMIN' || role === 'admin' || role === 3) {
    return t('roles.admin');
  }
  if (role === 'USER_ROLE_MODERATOR' || role === 'moderator' || role === 2) {
    return t('roles.moderator');
  }
  return t('roles.developer');
});

const roleBadgeClass = computed(() => {
  const role = user.value.role;
  if (role === 'USER_ROLE_ADMIN' || role === 'admin' || role === 3) return 'role-admin';
  if (role === 'USER_ROLE_MODERATOR' || role === 'moderator' || role === 2) return 'role-moderator';
  return 'role-developer';
});

const registeredDate = computed(() => {
  if (!user.value.created_at) return '';
  return formatDate(user.value.created_at);
});

// Display name form
const displayNameForm = ref(user.value.display_name || '');
const nameSaving = ref(false);

watch(
  () => user.value.display_name,
  (newVal) => {
    if (newVal !== undefined) {
      displayNameForm.value = newVal;
    }
  },
  { immediate: true }
);

const isNameChanged = computed(() => {
  const current = (user.value.display_name || '').trim();
  const form = displayNameForm.value.trim();
  return form.length > 0 && form !== current;
});

async function handleUpdateDisplayName() {
  if (!isNameChanged.value) return;
  nameSaving.value = true;
  try {
    const res = await updateProfile({ display_name: displayNameForm.value.trim() });
    const updatedUser = res.user || { display_name: displayNameForm.value.trim() };
    updateCurrentUser(updatedUser);
    showToast(t('profile.profileUpdated'), 'success');
  } catch (err) {
    showToast(err.response?.data?.message || t('common.error'), 'danger');
  } finally {
    nameSaving.value = false;
  }
}

// Password form
const passwordForm = reactive({
  currentPassword: '',
  newPassword: '',
  confirmPassword: '',
});
const passwordSaving = ref(false);

const isPasswordFormFilled = computed(() => {
  return (
    passwordForm.currentPassword.length > 0 &&
    passwordForm.newPassword.length >= 6 &&
    passwordForm.confirmPassword.length >= 6
  );
});

async function handleChangePassword() {
  if (passwordForm.newPassword !== passwordForm.confirmPassword) {
    showToast(t('auth.errorPasswordMismatch'), 'danger');
    return;
  }

  passwordSaving.value = true;
  try {
    await changePassword({
      current_password: passwordForm.currentPassword,
      new_password: passwordForm.newPassword,
    });
    passwordForm.currentPassword = '';
    passwordForm.newPassword = '';
    passwordForm.confirmPassword = '';
    showToast(t('profile.passwordChanged'), 'success');
  } catch (err) {
    showToast(err.response?.data?.message || t('common.error'), 'danger');
  } finally {
    passwordSaving.value = false;
  }
}

function selectLanguage(lang) {
  setLocale(lang);
}
</script>

<style scoped>
.profile-page {
  width: 100%;
  min-height: calc(100vh - 60px);
  background: var(--bg-app);
  padding: 24px 32px 48px;
  box-sizing: border-box;
}

.profile-container {
  width: 100%;
  max-width: 100%;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.user-info-card {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-md, 8px);
  padding: 20px 28px;
  flex-wrap: wrap;
  box-sizing: border-box;
}

.user-main-info {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.user-title {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--text-main);
}

.role-badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 0.8rem;
  font-weight: 600;
}

.role-developer {
  background: var(--primary-light);
  color: var(--primary);
}

.role-moderator {
  background: var(--warning-light, rgba(245, 158, 11, 0.15));
  color: var(--warning, #d97706);
}

.role-admin {
  background: var(--danger-light, rgba(239, 68, 68, 0.15));
  color: var(--danger, #dc2626);
}

.user-meta-details {
  display: flex;
  align-items: center;
  gap: 24px;
  flex-wrap: wrap;
  font-size: 0.9rem;
  color: var(--text-muted);
}

.detail-item {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

/* 2-колоночная сетка: равная высота элементов по строкам */
.profile-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 24px;
  width: 100%;
}

.profile-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-md, 8px);
  padding: 24px 28px;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
}

.card-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 1.05rem;
  font-weight: 600;
  color: var(--text-main);
  margin: 0 0 20px 0;
}

/* Строка 1: Карточка имени и Карточка языка */
.card-name {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.card-name .inline-name-form {
  flex: 1;
  display: flex;
  gap: 12px;
  align-items: center;
}

.card-lang {
  display: flex;
  flex-direction: column;
}

.card-lang .languages-grid {
  flex: 1;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
  align-items: stretch;
}

/* Строка 2: Карточка пароля и Карточка темы */
.card-password {
  display: flex;
  flex-direction: column;
}

.card-password .password-form {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 16px;
}

.card-theme {
  display: flex;
  flex-direction: column;
}

.card-theme .theme-selector-wrap {
  flex: 1;
  display: flex;
  flex-direction: column;
}

/* Однострочная форма изменения имени */
.inline-name-form {
  display: flex;
  gap: 12px;
  align-items: center;
}

.field-name-input {
  flex: 1;
  min-width: 0;
}

.btn-save-name {
  flex-shrink: 0;
  white-space: nowrap;
}

.password-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-row-2 {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-label {
  font-size: 0.85rem;
  font-weight: 500;
  color: var(--text-muted);
}

.form-input {
  width: 100%;
  height: 40px;
  padding: 0 14px;
  border-radius: var(--radius-sm, 6px);
  border: 1px solid var(--border);
  background: var(--bg-secondary);
  color: var(--text-main);
  font-size: 0.92rem;
  transition: all 0.15s ease;
  box-sizing: border-box;
  outline: none;
}

.form-input:focus {
  border-color: var(--primary);
  background: var(--bg-card);
}

.form-actions-full {
  display: flex;
  width: 100%;
  margin-top: 6px;
}

.btn-password-submit {
  width: 100%;
}

.btn-primary {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  background: var(--primary);
  color: #fff;
  border: none;
  border-radius: var(--radius-sm, 6px);
  padding: 0 20px;
  height: 40px;
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-primary:hover:not(:disabled) {
  opacity: 0.92;
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.info-alert {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  border-radius: var(--radius-md, 8px);
  padding: 16px 20px;
  color: var(--text-muted);
  font-size: 0.88rem;
  line-height: 1.4;
}

.alert-icon {
  color: var(--primary);
  flex-shrink: 0;
  margin-top: 2px;
}

/* Языковая сетка */
.languages-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
}

.lang-card {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 16px 20px;
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm, 6px);
  cursor: pointer;
  transition: all 0.15s ease;
  text-align: left;
}

.lang-card:hover {
  border-color: var(--border-secondary);
}

.lang-card.active {
  border-color: var(--primary);
  background: var(--bg-card);
}

.lang-flag {
  font-size: 1.6rem;
}

.lang-text {
  display: flex;
  flex-direction: column;
  flex: 1;
}

.lang-name {
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--text-main);
}

.lang-sub {
  font-size: 0.8rem;
  color: var(--text-muted);
}

.lang-check {
  color: var(--primary);
  flex-shrink: 0;
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 960px) {
  .profile-page {
    padding: 16px;
  }
  .profile-grid {
    grid-template-columns: 1fr;
  }
  .user-info-card {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
}

@media (max-width: 540px) {
  .inline-name-form {
    flex-direction: column;
    align-items: stretch;
  }
  .form-row-2 {
    grid-template-columns: 1fr;
  }
  .languages-grid {
    grid-template-columns: 1fr;
  }
}
</style>
