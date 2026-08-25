<template>
  <div class="profile-page">
    <div class="profile-container">
      <div class="profile-header">
        <div class="header-titles">
          <h1>{{ t('profile.title') }}</h1>
          <p class="subtitle">{{ t('profile.subtitle') }}</p>
        </div>
      </div>

      <!-- Карточка информации об аккаунте -->
      <section class="profile-section user-card-section">
        <div class="user-avatar-wrap">
          <div class="user-avatar">
            {{ userInitials }}
          </div>
        </div>
        <div class="user-meta-info">
          <div class="meta-row">
            <span class="user-title">{{ userDisplayName }}</span>
            <span class="role-badge" :class="roleBadgeClass">
              <Shield class="icon-xs" v-if="isModeratorOrAdmin" />
              <Code2 class="icon-xs" v-else />
              {{ localizedRoleName }}
            </span>
          </div>
          <div class="meta-details">
            <div class="detail-item">
              <Mail class="icon-xs text-muted" />
              <span>{{ userEmail }}</span>
            </div>
            <div class="detail-item" v-if="registeredDate">
              <Calendar class="icon-xs text-muted" />
              <span>{{ t('profile.registeredAt') }}: {{ registeredDate }}</span>
            </div>
          </div>
        </div>
      </section>

      <!-- СЕКЦИЯ: Изменение отображаемого имени (Только для разработчиков) -->
      <section class="profile-section" v-if="!isModeratorOrAdmin">
        <div class="section-header-block">
          <h2 class="section-title">
            <UserCheck class="icon-sm text-primary" />
            {{ t('profile.editProfile') }}
          </h2>
          <p class="section-desc">{{ t('profile.editProfileDesc') }}</p>
        </div>

        <form @submit.prevent="handleUpdateDisplayName" class="form-grid">
          <div class="form-group">
            <label class="form-label" for="displayNameInput">
              {{ t('profile.studioOrName') }}
            </label>
            <div class="input-wrap">
              <input
                id="displayNameInput"
                type="text"
                v-model="displayNameForm"
                class="form-input"
                :placeholder="t('auth.displayNamePlaceholder')"
                :disabled="nameSaving"
                required
              />
            </div>
          </div>

          <div class="form-actions">
            <button
              type="submit"
              class="btn-primary"
              :disabled="nameSaving || !isNameChanged"
            >
              <Loader2 class="icon-sm spin" v-if="nameSaving" />
              <Check class="icon-sm" v-else />
              {{ nameSaving ? t('common.saving') : t('profile.updateProfileBtn') }}
            </button>
          </div>
        </form>
      </section>

      <!-- СЕКЦИЯ: Смена пароля (Только для разработчиков) -->
      <section class="profile-section" v-if="!isModeratorOrAdmin">
        <div class="section-header-block">
          <h2 class="section-title">
            <Lock class="icon-sm text-primary" />
            {{ t('profile.security') }}
          </h2>
          <p class="section-desc">{{ t('profile.securityDesc') }}</p>
        </div>

        <form @submit.prevent="handleChangePassword" class="form-grid">
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

          <div class="form-actions">
            <button
              type="submit"
              class="btn-primary"
              :disabled="passwordSaving || !isPasswordFormFilled"
            >
              <Loader2 class="icon-sm spin" v-if="passwordSaving" />
              <Key class="icon-sm" v-else />
              {{ passwordSaving ? t('common.saving') : t('profile.changePasswordBtn') }}
            </button>
          </div>
        </form>
      </section>

      <!-- Уведомление для модераторов/администраторов -->
      <div class="info-alert" v-if="isModeratorOrAdmin">
        <Info class="icon-sm alert-icon" />
        <div class="alert-text">
          {{ t('profile.readonlyNotice') }}
        </div>
      </div>

      <!-- СЕКЦИЯ: Язык интерфейса (Доступно всем ролям) -->
      <section class="profile-section">
        <div class="section-header-block">
          <h2 class="section-title">
            <Languages class="icon-sm text-primary" />
            {{ t('profile.interfaceLanguage') }}
          </h2>
          <p class="section-desc">{{ t('profile.interfaceLanguageDesc') }}</p>
        </div>

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

      <!-- СЕКЦИЯ: Тема оформления (Доступно всем ролям) -->
      <section class="profile-section">
        <div class="section-header-block">
          <h2 class="section-title">
            <Palette class="icon-sm text-primary" />
            {{ t('profile.theme') }}
          </h2>
          <p class="section-desc">{{ t('profile.themeDesc') }}</p>
        </div>

        <ThemeCardSelector />
      </section>
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

const userInitials = computed(() => {
  const name = userDisplayName.value || 'U';
  return name.slice(0, 2).toUpperCase();
});

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
  min-height: 100vh;
  background: var(--bg-app);
  padding: 32px 24px 64px;
}

.profile-container {
  max-width: 760px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.profile-header {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 8px;
}

.header-titles h1 {
  font-size: 1.75rem;
  font-weight: 700;
  color: var(--text-main);
  margin: 0;
  letter-spacing: -0.02em;
}

.subtitle {
  font-size: 0.95rem;
  color: var(--text-muted);
  margin: 6px 0 0;
}

.profile-section {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  padding: 24px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
}

.user-card-section {
  display: flex;
  align-items: center;
  gap: 20px;
  background: linear-gradient(to right, var(--bg-card), var(--bg-secondary));
}

.user-avatar-wrap {
  flex-shrink: 0;
}

.user-avatar {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  background: var(--primary);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.4rem;
  font-weight: 700;
  letter-spacing: 1px;
  box-shadow: 0 4px 12px rgba(var(--primary-rgb, 59, 130, 246), 0.3);
}

.user-meta-info {
  display: flex;
  flex-direction: column;
  gap: 8px;
  flex: 1;
}

.meta-row {
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
  padding: 3px 10px;
  border-radius: 20px;
  font-size: 0.75rem;
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

.meta-details {
  display: flex;
  align-items: center;
  gap: 20px;
  flex-wrap: wrap;
  font-size: 0.85rem;
  color: var(--text-muted);
}

.detail-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.section-header-block {
  margin-bottom: 20px;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 1.15rem;
  font-weight: 600;
  color: var(--text-main);
  margin: 0 0 6px;
}

.section-desc {
  font-size: 0.875rem;
  color: var(--text-muted);
  margin: 0;
}

.form-grid {
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
  color: var(--text-main);
}

.form-input {
  width: 100%;
  padding: 10px 14px;
  border-radius: var(--radius-md);
  border: 1px solid var(--border);
  background: var(--bg-secondary);
  color: var(--text-main);
  font-size: 0.95rem;
  transition: all 0.2s;
  box-sizing: border-box;
}

.form-input:focus {
  outline: none;
  border-color: var(--primary);
  background: var(--bg-card);
  box-shadow: 0 0 0 3px rgba(var(--primary-rgb, 59, 130, 246), 0.15);
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 8px;
}

.btn-primary {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: var(--primary);
  color: #fff;
  border: none;
  border-radius: var(--radius-md);
  padding: 10px 20px;
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary:hover:not(:disabled) {
  opacity: 0.92;
  transform: translateY(-1px);
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
  border-radius: var(--radius-md);
  padding: 14px 16px;
  color: var(--text-muted);
  font-size: 0.875rem;
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
  gap: 16px;
}

.lang-card {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 16px 20px;
  background: var(--bg-secondary);
  border: 2px solid var(--border);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all 0.2s;
  text-align: left;
}

.lang-card:hover {
  border-color: var(--border-secondary);
  background: var(--bg-hover);
}

.lang-card.active {
  border-color: var(--primary);
  background: var(--bg-card);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.lang-flag {
  font-size: 1.75rem;
}

.lang-text {
  display: flex;
  flex-direction: column;
  flex: 1;
}

.lang-name {
  font-size: 1rem;
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

@media (max-width: 640px) {
  .profile-page {
    padding: 20px 16px;
  }
  .form-row-2 {
    grid-template-columns: 1fr;
  }
  .languages-grid {
    grid-template-columns: 1fr;
  }
  .user-card-section {
    flex-direction: column;
    text-align: center;
  }
  .meta-row,
  .meta-details {
    justify-content: center;
  }
}
</style>
