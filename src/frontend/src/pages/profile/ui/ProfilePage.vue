<template>
  <div class="profile-page">
    <div class="profile-container" :class="{ 'is-compact': isModeratorOrAdmin }">
      <!-- Карточка информации об аккаунте (без псевдо-аватара) -->
      <section class="user-info-card">
        <div class="user-main-info">
          <span class="user-title">{{ userDisplayName }}</span>
          <span class="role-badge" :class="roleBadgeClass">
            <Shield v-if="isModeratorOrAdmin" class="icon-xs" />
            <Code2 v-else class="icon-xs" />
            {{ localizedRoleName }}
          </span>
        </div>
        <div class="user-meta-details">
          <div class="detail-item">
            <Mail class="icon-xs text-muted" />
            <span>{{ userEmail }}</span>
          </div>
          <div v-if="registeredDate" class="detail-item">
            <Calendar class="icon-xs text-muted" />
            <span>{{ t('profile.registeredAt') }}: {{ registeredDate }}</span>
          </div>
        </div>
      </section>

      <!-- 2-колоночная адаптивная сетка (для модератора и админа — вертикальный стек) -->
      <div class="profile-grid" :class="{ 'grid-vertical': isModeratorOrAdmin }">
        <!-- Изменение отображаемого имени (только для разработчиков) -->
        <section v-if="!isModeratorOrAdmin" class="profile-card card-name">
          <h2 class="card-title">
            <UserCheck class="icon-sm text-primary" />
            {{ t('profile.editProfile') }}
          </h2>

          <form class="inline-name-form" @submit.prevent="handleUpdateDisplayName">
            <input
              id="displayNameInput"
              v-model="displayNameForm"
              type="text"
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
              <Loader2 v-if="nameSaving" class="icon-sm spin" />
              <Check v-else class="icon-sm" />
              <span>{{ nameSaving ? t('common.saving') : t('profile.updateProfileBtn') }}</span>
            </button>
          </form>
        </section>

        <!-- Язык интерфейса -->
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
              <Check v-if="currentLocale === 'ru'" class="icon-sm lang-check" />
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
              <Check v-if="currentLocale === 'en'" class="icon-sm lang-check" />
            </button>
          </div>
        </section>

        <!-- Смена пароля (только для разработчиков) -->
        <section v-if="!isModeratorOrAdmin" class="profile-card card-password">
          <h2 class="card-title">
            <Lock class="icon-sm text-primary" />
            {{ t('profile.security') }}
          </h2>

          <form class="password-form" @submit.prevent="handleChangePassword">
            <div class="form-group">
              <label class="form-label" for="currentPasswordInput">
                {{ t('auth.currentPassword') }}
              </label>
              <input
                id="currentPasswordInput"
                v-model="passwordForm.currentPassword"
                type="password"
                class="form-input"
                :placeholder="t('auth.passwordPlaceholder')"
                :disabled="passwordSaving"
                required
              />
            </div>

            <div class="form-group">
              <label class="form-label" for="newPasswordInput">
                {{ t('auth.newPassword') }}
              </label>
              <input
                id="newPasswordInput"
                v-model="passwordForm.newPassword"
                type="password"
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
                v-model="passwordForm.confirmPassword"
                type="password"
                class="form-input"
                :placeholder="t('auth.passwordPlaceholder')"
                :disabled="passwordSaving"
                minlength="6"
                required
              />
            </div>

            <div class="form-actions-full">
              <button
                type="submit"
                class="btn-primary btn-password-submit"
                :disabled="passwordSaving || !isPasswordFormFilled"
              >
                <Loader2 v-if="passwordSaving" class="icon-sm spin" />
                <Key v-else class="icon-sm" />
                <span>{{
                  passwordSaving ? t('common.saving') : t('profile.changePasswordBtn')
                }}</span>
              </button>
            </div>
          </form>
        </section>

        <!-- Тема оформления -->
        <section class="profile-card card-theme">
          <h2 class="card-title">
            <Palette class="icon-sm text-primary" />
            {{ t('profile.theme') }}
          </h2>

          <ThemeCardSelector class="theme-selector-wrap" />
        </section>

        <!-- Входящие приглашения в проекты (только для разработчиков) -->
        <section v-if="!isModeratorOrAdmin" class="profile-card card-invitations">
          <div class="card-header-flex">
            <h2 class="card-title no-margin">
              <Mail class="icon-sm text-primary" />
              {{ t('profile.invitations') }}
            </h2>
            <span v-if="incomingList.length > 0" class="count-badge">
              {{ incomingList.length }}
            </span>
          </div>
          <p class="card-subtitle-text">{{ t('profile.invitationsDesc') }}</p>

          <div v-if="invitesLoading" class="mini-loader-wrap">
            <div class="spinner-sm"></div>
            <span>{{ t('common.loading') }}</span>
          </div>

          <div v-else-if="incomingList.length === 0" class="empty-compact-box">
            <Mail class="icon-md text-muted" />
            <span>{{ t('profile.emptyInvitations') }}</span>
          </div>

          <div v-else class="invitations-stack">
            <div v-for="inv in incomingList" :key="inv.id" class="invitation-item-card">
              <div class="inv-project-header">
                <div class="inv-project-icon">
                  <img
                    v-if="inv.project_icon"
                    :src="getMediaUrl(inv.project_icon)"
                    alt="Icon"
                    class="inv-icon-img"
                  />
                  <span v-else class="inv-mock-icon">Draft</span>
                </div>
                <div class="inv-project-meta">
                  <span class="inv-project-title">{{ inv.project_title || '—' }}</span>
                  <span class="inv-inviter-info text-muted">
                    {{ inv.inviter_name || inv.inviter_email }}
                    <span v-if="inv.inviter_name && inv.inviter_email">({{ inv.inviter_email }})</span>
                  </span>
                </div>
                <div class="inv-actions">
                  <button
                    class="btn-icon text-success-hover"
                    :disabled="actionInProgress === inv.id"
                    :title="t('access.actions.accept')"
                    @click="handleRespondInvite(inv.id, true)"
                  >
                    <Check class="icon-sm" />
                  </button>
                  <button
                    class="btn-icon text-warning-hover"
                    :disabled="actionInProgress === inv.id"
                    :title="t('access.actions.decline')"
                    @click="handleRespondInvite(inv.id, false)"
                  >
                    <X class="icon-sm" />
                  </button>
                  <button
                    class="btn-icon text-danger-hover"
                    :disabled="actionInProgress === inv.id"
                    :title="t('access.actions.block')"
                    @click="promptBlockFromInvite(inv.inviter_id, inv.inviter_name || inv.inviter_email)"
                  >
                    <Ban class="icon-sm" />
                  </button>
                </div>
              </div>

              <div class="inv-permissions-row">
                <span v-for="perm in inv.permissions" :key="perm" class="badge-role-collab">
                  {{ permissionLabel(perm) }}
                </span>
                <span class="inv-date text-muted">
                  <Clock class="icon-xs" style="margin-right: 2px; vertical-align: middle;" />
                  {{ formatDate(inv.created_at) }}
                </span>
              </div>
            </div>
          </div>
        </section>

        <!-- Черный список (только для разработчиков) -->
        <section v-if="!isModeratorOrAdmin" class="profile-card card-blacklist">
          <div class="card-header-flex">
            <h2 class="card-title no-margin">
              <Ban class="icon-sm text-primary" />
              {{ t('profile.blacklist') }}
            </h2>
            <span v-if="blockedUsersList.length > 0" class="count-badge danger-badge">
              {{ blockedUsersList.length }}
            </span>
          </div>
          <p class="card-subtitle-text">{{ t('profile.blacklistDesc') }}</p>

          <div v-if="blacklistLoading" class="mini-loader-wrap">
            <div class="spinner-sm"></div>
            <span>{{ t('common.loading') }}</span>
          </div>

          <div v-else-if="blockedUsersList.length === 0" class="empty-compact-box">
            <UserCheck class="icon-md text-muted" />
            <span>{{ t('profile.emptyBlacklist') }}</span>
          </div>

          <div v-else class="blacklist-stack">
            <div v-for="b in blockedUsersList" :key="b.id" class="blocked-item-row">
              <div class="blocked-user-meta">
                <span class="blocked-user-name">{{ b.blocked_user_name || b.blocked_user_email }}</span>
                <span v-if="b.blocked_user_name && b.blocked_user_email" class="blocked-user-email text-muted">
                  {{ b.blocked_user_email }}
                </span>
              </div>
              <button
                class="btn-unblock"
                :disabled="actionInProgress === b.id"
                @click="handleUnblock(b.blocked_user_id)"
              >
                <Check class="icon-xs" />
                <span>{{ t('access.actions.unblock') }}</span>
              </button>
            </div>
          </div>
        </section>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, reactive, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import {
  Mail,
  Calendar,
  Lock,
  Key,
  Check,
  X,
  Ban,
  Clock,
  Loader2,
  UserCheck,
  Shield,
  Code2,
  Languages,
  Palette,
} from 'lucide-vue-next';
import { useAuth, updateProfile, changePassword } from '@/entities/user';
import {
  listIncomingInvitations,
  respondInvitation,
  listBlockedUsers,
  blockUser,
  unblockUser,
  getMediaUrl,
  permissionLabel,
} from '@/entities/project';
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

// ─── Incoming Invitations & Blacklist ─────────────────────────
const incomingList = ref([]);
const invitesLoading = ref(false);
const blockedUsersList = ref([]);
const blacklistLoading = ref(false);
const actionInProgress = ref(null);

async function loadIncoming() {
  if (isModeratorOrAdmin.value) return;
  invitesLoading.value = true;
  try {
    incomingList.value = await listIncomingInvitations();
  } catch {
    incomingList.value = [];
  } finally {
    invitesLoading.value = false;
  }
}

async function loadBlacklist() {
  if (isModeratorOrAdmin.value) return;
  blacklistLoading.value = true;
  try {
    blockedUsersList.value = await listBlockedUsers();
  } catch {
    blockedUsersList.value = [];
  } finally {
    blacklistLoading.value = false;
  }
}

async function handleRespondInvite(invitationId, accept) {
  actionInProgress.value = invitationId;
  try {
    await respondInvitation(invitationId, accept);
    showToast(
      accept ? t('access.messages.inviteAccepted') : t('access.messages.inviteDeclined'),
      'success'
    );
    await loadIncoming();
  } catch (err) {
    showToast(err.response?.data?.message || t('common.error'), 'danger');
  } finally {
    actionInProgress.value = null;
  }
}

async function handleUnblock(userId) {
  actionInProgress.value = userId;
  try {
    await unblockUser(userId);
    showToast(t('access.messages.userUnblocked'), 'success');
    await loadBlacklist();
  } catch (err) {
    showToast(err.response?.data?.message || t('common.error'), 'danger');
  } finally {
    actionInProgress.value = null;
  }
}

async function promptBlockFromInvite(userId, userName) {
  if (!confirm(`${t('access.modals.blockConfirmTitle')} (${userName})`)) return;
  actionInProgress.value = userId;
  try {
    await blockUser(userId);
    showToast(t('access.messages.userBlocked'), 'success');
    await Promise.all([loadIncoming(), loadBlacklist()]);
  } catch (err) {
    showToast(err.response?.data?.message || t('common.error'), 'danger');
  } finally {
    actionInProgress.value = null;
  }
}

onMounted(() => {
  if (!isModeratorOrAdmin.value) {
    loadIncoming();
    loadBlacklist();
  }
});
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

.profile-container.is-compact {
  max-width: 680px;
  margin: 0 auto;
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

.profile-grid.grid-vertical {
  display: flex;
  flex-direction: column;
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

/* Invitations & Blacklist Cards */
.card-header-flex {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 4px;
}
.no-margin {
  margin: 0 !important;
}
.card-subtitle-text {
  margin: 0 0 16px 0;
  font-size: 13px;
  color: var(--text-muted);
}
.count-badge {
  padding: 2px 8px;
  font-size: 11px;
  font-weight: 700;
  border-radius: 12px;
  background: var(--primary, #58a6ff);
  color: #fff;
}
.count-badge.danger-badge {
  background: var(--danger, #f85149);
}

.mini-loader-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 20px;
  color: var(--text-muted);
  font-size: 13px;
  justify-content: center;
}
.empty-compact-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 28px 16px;
  color: var(--text-muted);
  font-size: 13px;
  background: var(--bg-secondary, #0d1117);
  border: 1px dashed var(--border, #30363d);
  border-radius: 6px;
  text-align: center;
}

.invitations-stack,
.blacklist-stack {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.invitation-item-card {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 12px 14px;
  background: var(--bg-secondary, #0d1117);
  border: 1px solid var(--border, #30363d);
  border-radius: 6px;
}
.inv-project-header {
  display: flex;
  align-items: center;
  gap: 12px;
}
.inv-project-icon {
  width: 36px;
  height: 36px;
  border-radius: 6px;
  background: var(--bg-tertiary, #21262d);
  border: 1px solid var(--border, #30363d);
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  flex-shrink: 0;
}
.inv-icon-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.inv-mock-icon {
  font-size: 10px;
  font-weight: 600;
  color: var(--text-muted);
}
.inv-project-meta {
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex: 1;
  min-width: 0;
}
.inv-project-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.inv-inviter-info {
  font-size: 12px;
}
.inv-actions {
  display: flex;
  align-items: center;
  gap: 4px;
}
.inv-permissions-row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
  padding-top: 6px;
  border-top: 1px solid var(--border, #21262d);
}
.inv-date {
  font-size: 11px;
  margin-left: auto;
}

.badge-role-collab {
  display: inline-flex;
  align-items: center;
  font-size: 11px;
  font-weight: 500;
  padding: 1px 7px;
  border-radius: 10px;
  background: var(--bg-tertiary, #21262d);
  color: var(--primary, #58a6ff);
  border: 1px solid var(--border, #30363d);
}

.blocked-item-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  background: var(--bg-secondary, #0d1117);
  border: 1px solid var(--border, #30363d);
  border-radius: 6px;
}
.blocked-user-meta {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.blocked-user-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
}
.blocked-user-email {
  font-size: 12px;
}
.btn-unblock {
  padding: 5px 12px;
  background: var(--bg-tertiary, #21262d);
  border: 1px solid var(--border, #30363d);
  border-radius: 6px;
  color: var(--text-main, #f0f6fc);
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 5px;
  transition: all 0.15s ease;
}
.btn-unblock:hover {
  background: var(--primary, #58a6ff);
  border-color: var(--primary, #58a6ff);
  color: #fff;
}

.btn-icon {
  background: none;
  border: none;
  padding: 6px;
  cursor: pointer;
  color: var(--text-tertiary, #8b949e);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  transition: all 0.15s;
}
.btn-icon:hover {
  color: var(--text-main, #f0f6fc);
  background: var(--bg-tertiary, #21262d);
}
.text-danger-hover:hover { color: var(--danger, #f85149) !important; }
.text-warning-hover:hover { color: var(--warning, #d29922) !important; }
.text-success-hover:hover { color: #2ecc71 !important; }

.spinner-sm {
  width: 14px;
  height: 14px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: #ffffff;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  display: inline-block;
  flex-shrink: 0;
  vertical-align: middle;
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
