<template>
  <div class="profile-page">
    <div class="profile-container">
      <!-- Для модератора/админа (только учетная запись и интерфейс) -->
      <div v-if="isModeratorOrAdmin" class="profile-main-grid">
        <!-- Колонка 1: Учетная запись и безопасность -->
        <div class="profile-col">
          <section class="profile-card card-account">
            <div class="card-header-identity">
              <div class="header-identity-left">
                <div class="header-title-wrap">
                  <UserCheck class="icon-md text-primary" />
                  <h2 class="card-title-lg">{{ t('profile.accountAndSecurity') }}</h2>
                </div>
                <span class="role-badge" :class="roleBadgeClass">
                  <Shield class="icon-xs" />
                  {{ localizedRoleName }}
                </span>
              </div>

              <div class="header-identity-right">
                <div class="identity-meta-pill">
                  <Mail class="icon-xs text-muted" />
                  <span>{{ userEmail }}</span>
                </div>
                <div v-if="registeredDate" class="identity-meta-pill">
                  <Calendar class="icon-xs text-muted" />
                  <span>{{ t('profile.registeredAt') }}: {{ registeredDate }}</span>
                </div>
              </div>
            </div>

            <div class="account-body">
              <div class="account-sub-section">
                <label class="form-label" for="displayNameInputMod">
                  {{ t('profile.editProfile') }}
                </label>
                <form class="inline-name-form" @submit.prevent="handleUpdateDisplayName">
                  <input
                    id="displayNameInputMod"
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
              </div>

              <div class="card-divider"></div>

              <div class="account-sub-section">
                <label class="form-label">{{ t('profile.security') }}</label>
                <form class="password-form" @submit.prevent="handleChangePassword">
                  <div class="form-group">
                    <input
                      v-model="passwordForm.currentPassword"
                      type="password"
                      class="form-input"
                      :placeholder="t('auth.currentPassword')"
                      :disabled="passwordSaving"
                      required
                    />
                  </div>

                  <div class="form-row-2">
                    <div class="form-group">
                      <input
                        v-model="passwordForm.newPassword"
                        type="password"
                        class="form-input"
                        :placeholder="t('auth.newPassword')"
                        :disabled="passwordSaving"
                        minlength="6"
                        required
                      />
                    </div>

                    <div class="form-group">
                      <input
                        v-model="passwordForm.confirmPassword"
                        type="password"
                        class="form-input"
                        :placeholder="t('auth.confirmPassword')"
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
                      <Loader2 v-if="passwordSaving" class="icon-sm spin" />
                      <Key v-else class="icon-sm" />
                      <span>{{
                        passwordSaving ? t('common.saving') : t('profile.changePasswordBtn')
                      }}</span>
                    </button>
                  </div>
                </form>
              </div>
            </div>
          </section>
        </div>

        <!-- Колонка 2: Интерфейс и оформление -->
        <div class="profile-col">
          <section class="profile-card card-appearance">
            <h2 class="card-title">
              <Palette class="icon-md text-primary" />
              {{ t('profile.appearance') }}
            </h2>

            <div class="appearance-row-split">
              <div class="pref-block-half">
                <span class="sub-block-label">{{ t('profile.interfaceLanguage') }}</span>
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
              </div>

              <div class="pref-block-half">
                <span class="sub-block-label">{{ t('profile.theme') }}</span>
                <ThemeCardSelector />
              </div>
            </div>
          </section>
        </div>
      </div>

      <!-- Основная 2-колоночная сетка для разработчиков на всю ширину -->
      <div v-else class="profile-main-grid">
        <!-- Левая колонка: Учетная запись + Персонализация -->
        <div class="profile-col">
          <!-- Карточка 1: Учетная запись и безопасность -->
          <section class="profile-card card-account">
            <!-- Шапка учетной записи: Мета-информация без дублирования имени и без псевдо-аватара -->
            <div class="card-header-identity">
              <div class="header-identity-left">
                <div class="header-title-wrap">
                  <UserCheck class="icon-md text-primary" />
                  <h2 class="card-title-lg">{{ t('profile.accountAndSecurity') }}</h2>
                </div>
                <span class="role-badge" :class="roleBadgeClass">
                  <Code2 class="icon-xs" />
                  {{ localizedRoleName }}
                </span>
              </div>

              <div class="header-identity-right">
                <div class="identity-meta-pill">
                  <Mail class="icon-xs text-muted" />
                  <span>{{ userEmail }}</span>
                </div>
                <div v-if="registeredDate" class="identity-meta-pill">
                  <Calendar class="icon-xs text-muted" />
                  <span>{{ t('profile.registeredAt') }}: {{ registeredDate }}</span>
                </div>
              </div>
            </div>

            <div class="account-body">
              <!-- Отображаемое имя (единственное место отображения и редактирования) -->
              <div class="account-sub-section">
                <label class="form-label" for="displayNameInput">
                  {{ t('profile.editProfile') }}
                </label>
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
              </div>

              <div class="card-divider"></div>

              <!-- Смена пароля -->
              <div class="account-sub-section">
                <label class="form-label">{{ t('profile.security') }}</label>
                <form class="password-form" @submit.prevent="handleChangePassword">
                  <div class="form-group">
                    <input
                      id="currentPasswordInput"
                      v-model="passwordForm.currentPassword"
                      type="password"
                      class="form-input"
                      :placeholder="t('auth.currentPassword')"
                      :disabled="passwordSaving"
                      required
                    />
                  </div>

                  <div class="form-row-2">
                    <div class="form-group">
                      <input
                        id="newPasswordInput"
                        v-model="passwordForm.newPassword"
                        type="password"
                        class="form-input"
                        :placeholder="t('auth.newPassword')"
                        :disabled="passwordSaving"
                        minlength="6"
                        required
                      />
                    </div>

                    <div class="form-group">
                      <input
                        id="confirmPasswordInput"
                        v-model="passwordForm.confirmPassword"
                        type="password"
                        class="form-input"
                        :placeholder="t('auth.confirmPassword')"
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
                      <Loader2 v-if="passwordSaving" class="icon-sm spin" />
                      <Key v-else class="icon-sm" />
                      <span>{{
                        passwordSaving ? t('common.saving') : t('profile.changePasswordBtn')
                      }}</span>
                    </button>
                  </div>
                </form>
              </div>
            </div>
          </section>

          <!-- Карточка 2: Интерфейс и персонализация -->
          <section class="profile-card card-appearance">
            <h2 class="card-title">
              <Palette class="icon-md text-primary" />
              {{ t('profile.appearance') }}
            </h2>

            <div class="appearance-row-split">
              <!-- Язык интерфейса -->
              <div class="pref-block-half">
                <span class="sub-block-label">{{ t('profile.interfaceLanguage') }}</span>
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
              </div>

              <!-- Тема оформления -->
              <div class="pref-block-half">
                <span class="sub-block-label">{{ t('profile.theme') }}</span>
                <ThemeCardSelector />
              </div>
            </div>
          </section>
        </div>

        <!-- Правая колонка: Коллаборация (Приглашения + Черный список) -->
        <div class="profile-col">
          <!-- Карточка 3: Входящие приглашения -->
          <section class="profile-card card-invitations flex-card">
            <div class="card-header-flex">
              <div class="header-title-wrap">
                <Mail class="icon-md text-primary" />
                <h2 class="card-title-lg">{{ t('profile.invitations') }}</h2>
              </div>
              <span
                class="count-badge"
                :class="{ 'count-active': incomingList.length > 0 }"
              >
                {{ incomingList.length }}
              </span>
            </div>

            <div v-if="invitesLoading" class="mini-loader-wrap">
              <div class="spinner-sm"></div>
              <span>{{ t('common.loading') }}</span>
            </div>

            <div v-else-if="incomingList.length === 0" class="empty-state-modern">
              <div class="empty-icon-circle">
                <Mail class="icon-md" />
              </div>
              <span class="empty-title">{{ t('profile.emptyInvitations') }}</span>
              <span class="empty-desc">{{ t('profile.emptyInvitationsHint') }}</span>
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
                      :disabled="actionInProgress === inv.inviter_id"
                      :title="t('access.actions.block')"
                      @click="
                        promptBlockFromInvite(
                          inv.inviter_id,
                          inv.inviter_name || inv.inviter_email
                        )
                      "
                    >
                      <Ban class="icon-sm" />
                    </button>
                  </div>
                </div>

                <div class="inv-permissions-row">
                  <span
                    v-if="hasAllPermissions(inv.permissions)"
                    class="badge-full-access-sm"
                    :title="inv.permissions.map(permissionLabel).join('\n')"
                  >
                    <ShieldCheck class="icon-xs" />
                    {{ t('access.statuses.fullAccess') }}
                  </span>
                  <template v-else>
                    <span v-for="perm in inv.permissions" :key="perm" class="badge-role-collab">
                      {{ permissionLabel(perm) }}
                    </span>
                  </template>
                  <span class="inv-date text-muted">
                    <Clock class="icon-xs" style="margin-right: 2px; vertical-align: middle;" />
                    {{ formatDate(inv.created_at) }}
                  </span>
                </div>
              </div>
            </div>
          </section>

          <!-- Карточка 4: Черный список -->
          <section class="profile-card card-blacklist flex-card">
            <div class="card-header-flex">
              <div class="header-title-wrap">
                <Ban class="icon-md text-danger" />
                <h2 class="card-title-lg">{{ t('profile.blacklist') }}</h2>
              </div>
              <span
                class="count-badge"
                :class="{ 'count-active': blockedUsersList.length > 0 }"
              >
                {{ blockedUsersList.length }}
              </span>
            </div>

            <div v-if="blacklistLoading" class="mini-loader-wrap">
              <div class="spinner-sm"></div>
              <span>{{ t('common.loading') }}</span>
            </div>

            <div v-else-if="blockedUsersList.length === 0" class="empty-state-modern">
              <div class="empty-icon-circle">
                <UserX class="icon-md" />
              </div>
              <span class="empty-title">{{ t('profile.emptyBlacklist') }}</span>
              <span class="empty-desc">{{ t('profile.emptyBlacklistHint') }}</span>
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
  </div>
</template>

<script setup>
import { ref, computed, watch, reactive, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import {
  Mail,
  Calendar,
  Key,
  Check,
  X,
  Ban,
  Clock,
  Loader2,
  UserCheck,
  Shield,
  Code2,
  Palette,
  ShieldCheck,
  UserX,
} from 'lucide-vue-next';
import { useAuth, updateProfile, changePassword } from '@/entities/user';
import {
  listIncomingInvitations,
  respondInvitation,
  listBlockedUsers,
  blockUser,
  unblockUser,
  getMediaUrl,
  ALL_PERMISSIONS,
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

function hasAllPermissions(perms) {
  if (!perms || !Array.isArray(perms) || perms.length === 0) return false;
  return ALL_PERMISSIONS.every((p) => perms.includes(p));
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
  max-width: 100%;
  min-height: calc(100vh - 60px);
  background: var(--bg-app);
  padding: 24px 32px;
  box-sizing: border-box;
}

.profile-container {
  width: 100%;
  max-width: 100%;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* 2-колоночная сетка на всю ширину */
.profile-main-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  width: 100%;
  align-items: stretch;
}

.profile-col {
  display: flex;
  flex-direction: column;
  gap: 20px;
  min-width: 0;
}

.profile-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-md, 8px);
  padding: 20px 22px;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  transition: border-color 0.15s ease;
}

.profile-card.flex-card {
  flex: 1;
}

/* Шапка карточки идентичности и учетной записи */
.card-header-identity {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--border);
  margin-bottom: 16px;
}

.header-identity-left {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.header-title-wrap {
  display: inline-flex;
  align-items: center;
  gap: 10px;
}

.card-title-lg {
  font-size: 1.05rem;
  font-weight: 600;
  color: var(--text-main);
  margin: 0;
}

.card-title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 1.02rem;
  font-weight: 600;
  color: var(--text-main);
  margin: 0 0 16px 0;
}

.role-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 3px 12px;
  border-radius: 20px;
  font-size: 0.8rem;
  font-weight: 600;
}

.role-developer {
  background: var(--primary-light, rgba(59, 130, 246, 0.15));
  color: var(--primary, #3b82f6);
}

.role-moderator {
  background: var(--warning-light, rgba(245, 158, 11, 0.15));
  color: var(--warning, #d97706);
}

.role-admin {
  background: var(--danger-light, rgba(239, 68, 68, 0.15));
  color: var(--danger, #dc2626);
}

.header-identity-right {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.identity-meta-pill {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  padding: 5px 12px;
  border-radius: 6px;
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  font-size: 0.84rem;
  color: var(--text-muted);
}

.card-header-flex {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.count-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 22px;
  height: 22px;
  padding: 0 8px;
  border-radius: 12px;
  font-size: 0.78rem;
  font-weight: 600;
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  color: var(--text-muted);
}

.count-badge.count-active {
  background: var(--primary-light, rgba(59, 130, 246, 0.15));
  border-color: var(--primary, #3b82f6);
  color: var(--primary, #3b82f6);
}

/* Тело учетной записи */
.account-body {
  display: flex;
  flex-direction: column;
}

.account-sub-section {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.card-divider {
  width: 100%;
  height: 1px;
  background: var(--border);
  margin: 16px 0;
  opacity: 0.7;
}

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
  gap: 12px;
}

.form-row-2 {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-label {
  font-size: 0.84rem;
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
  margin-top: 4px;
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
  padding: 0 18px;
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

/* Блок персонализации и темы */
.appearance-row-split {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  width: 100%;
}

.pref-block-half {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.sub-block-label {
  font-size: 0.84rem;
  font-weight: 500;
  color: var(--text-muted);
}

.languages-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
  width: 100%;
}

.lang-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 14px;
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm, 6px);
  cursor: pointer;
  transition: all 0.15s ease;
  text-align: left;
  outline: none;
  position: relative;
}

.lang-card:hover {
  border-color: var(--border-secondary, rgba(255, 255, 255, 0.2));
}

.lang-card.active {
  border-color: var(--primary);
  background: var(--bg-card);
}

.lang-flag {
  font-size: 1.35rem;
  line-height: 1;
  flex-shrink: 0;
}

.lang-text {
  display: flex;
  flex-direction: column;
  min-width: 0;
  flex: 1;
}

.lang-name {
  font-size: 0.92rem;
  font-weight: 600;
  color: var(--text-main);
  line-height: 1.2;
}

.lang-sub {
  font-size: 0.78rem;
  color: var(--text-muted);
  line-height: 1.2;
}

.lang-check {
  color: var(--primary);
  margin-left: auto;
  flex-shrink: 0;
}

/* Пустые состояния */
.empty-state-modern {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 28px 20px;
  text-align: center;
  background: var(--bg-secondary);
  border: 1px dashed var(--border);
  border-radius: var(--radius-sm, 6px);
  flex: 1;
}

.empty-icon-circle {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  background: var(--bg-card);
  border: 1px solid var(--border);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-muted);
  margin-bottom: 10px;
}

.empty-title {
  font-size: 0.94rem;
  font-weight: 600;
  color: var(--text-main);
  margin-bottom: 4px;
}

.empty-desc {
  font-size: 0.82rem;
  color: var(--text-muted);
  max-width: 320px;
  line-height: 1.4;
}

/* Списки приглашений и черного списка */
.invitations-stack,
.blacklist-stack {
  display: flex;
  flex-direction: column;
  gap: 10px;
  overflow-y: auto;
  padding-right: 2px;
}

.invitations-stack {
  max-height: 240px;
}

.blacklist-stack {
  max-height: 240px;
}

.invitation-item-card {
  padding: 12px 14px;
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm, 6px);
  display: flex;
  flex-direction: column;
  gap: 8px;
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
  overflow: hidden;
  background: var(--bg-card);
  border: 1px solid var(--border);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.inv-icon-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.inv-mock-icon {
  font-size: 0.7rem;
  font-weight: 600;
  color: var(--text-muted);
}

.inv-project-meta {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-width: 0;
}

.inv-project-title {
  font-size: 0.92rem;
  font-weight: 600;
  color: var(--text-main);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.inv-inviter-info {
  font-size: 0.8rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.inv-actions {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-left: auto;
}

.btn-icon {
  background: none;
  border: none;
  cursor: pointer;
  padding: 5px;
  border-radius: 4px;
  color: var(--text-muted);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s ease;
}

.btn-icon:hover:not(:disabled) {
  background: var(--bg-card);
}

.text-success-hover:hover:not(:disabled) {
  color: var(--success, #10b981);
}

.text-warning-hover:hover:not(:disabled) {
  color: var(--warning, #f59e0b);
}

.text-danger-hover:hover:not(:disabled) {
  color: var(--danger, #ef4444);
}

.inv-permissions-row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
}

.badge-full-access-sm {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 3px 8px;
  border-radius: 4px;
  font-size: 0.78rem;
  font-weight: 600;
  background: rgba(99, 102, 241, 0.15);
  color: #818cf8;
  border: 1px solid rgba(99, 102, 241, 0.3);
}

.badge-role-collab {
  display: inline-flex;
  align-items: center;
  padding: 2px 7px;
  border-radius: 4px;
  font-size: 0.76rem;
  font-weight: 500;
  background: var(--bg-card);
  color: var(--text-muted);
  border: 1px solid var(--border);
}

.inv-date {
  margin-left: auto;
  font-size: 0.78rem;
  white-space: nowrap;
}

.blocked-item-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm, 6px);
  gap: 12px;
}

.blocked-user-meta {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.blocked-user-name {
  font-size: 0.9rem;
  font-weight: 500;
  color: var(--text-main);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.blocked-user-email {
  font-size: 0.8rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.btn-unblock {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 4px;
  padding: 5px 10px;
  font-size: 0.8rem;
  color: var(--text-main);
  cursor: pointer;
  transition: all 0.15s ease;
  white-space: nowrap;
}

.btn-unblock:hover:not(:disabled) {
  border-color: var(--primary);
  color: var(--primary);
}

.mini-loader-wrap {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 24px 0;
  color: var(--text-muted);
  font-size: 0.85rem;
}

.spinner-sm {
  width: 18px;
  height: 18px;
  border: 2px solid var(--border);
  border-top-color: var(--primary);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

.icon-xs {
  width: 14px;
  height: 14px;
}

.icon-sm {
  width: 16px;
  height: 16px;
}

.icon-md {
  width: 20px;
  height: 20px;
}

.text-primary {
  color: var(--primary);
}

.text-danger {
  color: var(--danger, #ef4444);
}

.text-muted {
  color: var(--text-muted);
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
  .profile-main-grid {
    grid-template-columns: 1fr;
  }
  .appearance-row-split {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 600px) {
  .profile-page {
    padding: 16px;
  }
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
