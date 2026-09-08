<template>
  <div class="tab-content tab-fade-in">
    <div class="page-header-row">
      <h1 class="page-title">{{ t('projectWorkspace.accessTab') }}</h1>
      <button class="btn-add-game-primary" @click="openInviteModal">
        <UserPlus class="icon-sm" />
        <span>{{ t('access.actions.invite') }}</span>
      </button>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="state-container">
      <div class="spinner-md"></div>
      <p>{{ t('common.loading') }}</p>
    </div>

    <!-- Members & Invites Table -->
    <div v-else class="access-table-card">
      <div
        v-if="members.length === 0 && pendingInvites.length === 0"
        class="empty-state"
      >
        <div class="empty-icon-wrap">
          <Users class="icon-lg" />
        </div>
        <h3>{{ t('access.messages.noMembersOrInvites') }}</h3>
        <p class="text-muted">{{ t('access.messages.noMembers') }}</p>
        <button class="btn-add-game-primary" @click="openInviteModal">
          <UserPlus class="icon-sm" />
          <span>{{ t('access.actions.invite') }}</span>
        </button>
      </div>

      <div v-else class="table-wrapper">
        <table class="yandex-games-table">
          <thead>
            <tr>
              <th class="col-user">{{ t('access.table.user') }}</th>
              <th class="col-perms">{{ t('access.table.permissions') }}</th>
              <th class="col-status">{{ t('access.table.status') }}</th>
              <th class="col-date">{{ t('access.table.date') }}</th>
              <th class="col-actions"></th>
            </tr>
          </thead>
          <tbody>
            <!-- Active Members -->
            <tr v-for="m in members" :key="'member-' + m.id" class="table-row">
              <td class="col-user">
                <div class="user-text">
                  <div class="user-name">{{ m.user_name || m.user_email }}</div>
                  <div v-if="m.user_name" class="user-email text-muted">{{ m.user_email }}</div>
                </div>
              </td>
              <td class="col-perms">
                <div
                  v-if="hasAllPermissions(m.permissions)"
                  class="badge-full-access"
                  :title="m.permissions.map(permissionLabel).join('\n')"
                >
                  <ShieldCheck class="icon-xs" />
                  <span>{{ t('access.statuses.fullAccess') }}</span>
                </div>
                <div v-else class="perms-badges-wrap">
                  <span v-for="p in m.permissions" :key="p" class="badge-role-collab">
                    {{ permissionLabel(p) }}
                  </span>
                </div>
              </td>
              <td class="col-status">
                <span class="status-pill status-active">{{ t('access.statuses.member') }}</span>
              </td>
              <td class="col-date">
                <span class="date-text">{{ formatDate(m.created_at) }}</span>
              </td>
              <td class="col-actions">
                <div class="row-actions">
                  <button
                    class="btn-icon"
                    :title="t('access.actions.editPermissions')"
                    @click="openEditPermissionsModal(m)"
                  >
                    <Sliders class="icon-xs" />
                  </button>
                  <button
                    class="btn-icon text-danger-hover"
                    :title="t('access.actions.removeMember')"
                    @click="promptRemoveMember(m)"
                  >
                    <UserX class="icon-xs" />
                  </button>
                </div>
              </td>
            </tr>

            <!-- Pending Invitations -->
            <tr
              v-for="inv in pendingInvites"
              :key="'invite-' + inv.id"
              class="table-row pending-row"
            >
              <td class="col-user">
                <div class="user-text">
                  <div class="user-name">{{ inv.invitee_email || inv.invitee_id }}</div>
                </div>
              </td>
              <td class="col-perms">
                <div
                  v-if="hasAllPermissions(inv.permissions)"
                  class="badge-full-access"
                  :title="inv.permissions.map(permissionLabel).join('\n')"
                >
                  <ShieldCheck class="icon-xs" />
                  <span>{{ t('access.statuses.fullAccess') }}</span>
                </div>
                <div v-else class="perms-badges-wrap">
                  <span v-for="perm in inv.permissions" :key="perm" class="badge-role-collab">
                    {{ permissionLabel(perm) }}
                  </span>
                </div>
              </td>
              <td class="col-status">
                <span class="status-pill status-pending">{{ t('access.statuses.pending') }}</span>
              </td>
              <td class="col-date">
                <span class="date-text">
                  <Clock class="icon-xs" style="margin-right: 4px; vertical-align: middle;" />
                  {{ formatDate(inv.created_at) }}
                </span>
              </td>
              <td class="col-actions">
                <div class="row-actions">
                  <button
                    class="btn-icon text-danger-hover"
                    :disabled="actionInProgress === inv.id"
                    :title="t('access.actions.cancel')"
                    @click="handleCancelInvitation(inv.id)"
                  >
                    <X class="icon-xs" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- ─── МОДАЛКА: Приглашение разработчика ─────────────────────── -->
    <div v-if="inviteModal.open" class="modal-overlay" @click.self="closeInviteModal">
      <div class="modal-card">
        <div class="modal-header">
          <h3>{{ t('access.modals.inviteTitle') }}</h3>
          <button class="btn-icon" @click="closeInviteModal">
            <X class="icon-sm" />
          </button>
        </div>

        <div class="modal-body">
          <div class="form-group">
            <label class="form-label">{{ t('access.modals.searchUser') }} <span class="req">*</span></label>
            <div class="search-user-wrapper">
              <div class="search-input-box">
                <Search class="icon-xs search-icon" />
                <input
                  v-model="inviteModal.query"
                  type="text"
                  class="form-input has-search-icon"
                  :placeholder="t('access.modals.searchUserPlaceholder')"
                  @input="onSearchUserInput"
                  @keydown.down.prevent="onSearchKeyDown"
                  @keydown.up.prevent="onSearchKeyUp"
                  @keydown.enter.prevent="onSearchKeyEnter"
                  @keydown.esc="closeSearchDropdown"
                />
                <div v-if="inviteModal.searching" class="spinner-sm search-loader"></div>
              </div>

              <div v-if="inviteModal.searchResults.length > 0" class="user-dropdown">
                <div
                  v-for="(u, idx) in inviteModal.searchResults"
                  :key="u.id"
                  class="user-dropdown-item"
                  :class="{ active: searchHighlightedIndex === idx }"
                  @mouseenter="searchHighlightedIndex = idx"
                  @click="selectUserToInvite(u)"
                >
                  <strong class="user-item-name">{{ u.display_name || u.email }}</strong>
                  <span class="user-item-email">{{ u.email }}</span>
                </div>
              </div>

              <div v-if="inviteModal.selectedUser" class="selected-user-pill">
                <div class="pill-info">
                  <strong>{{ inviteModal.selectedUser.display_name || inviteModal.selectedUser.email }}</strong>
                  <span class="pill-email">{{ inviteModal.selectedUser.email }}</span>
                </div>
                <button class="btn-icon" title="Очистить" @click="clearSelectedUser">
                  <X class="icon-xs" />
                </button>
              </div>
            </div>
          </div>

          <div class="form-group">
            <div class="permissions-header-row">
              <label class="form-label">{{ t('access.modals.permissionsSelect') }} <span class="req">*</span></label>
              <div class="quick-perms-actions">
                <button type="button" class="btn-text" @click="selectAllPerms">Все</button>
                <span class="text-separator">/</span>
                <button type="button" class="btn-text" @click="clearAllPerms">Сбросить</button>
              </div>
            </div>

            <div class="permissions-list">
              <div
                v-for="p in ALL_PERMISSIONS"
                :key="p"
                class="perm-item-row"
                :class="{ selected: inviteModal.permissions.includes(p) }"
                @click="togglePermission(inviteModal.permissions, p)"
              >
                <div class="custom-checkbox-box" :class="{ checked: inviteModal.permissions.includes(p) }">
                  <Check v-if="inviteModal.permissions.includes(p)" class="custom-check-icon" />
                </div>
                <div class="perm-label-wrap">
                  <span class="perm-title">{{ permissionLabel(p) }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="modal-footer">
          <button class="btn-reset" @click="closeInviteModal">
            {{ t('common.cancel') }}
          </button>
          <button
            class="btn-add-game-primary"
            :disabled="inviteModal.submitting"
            @click="submitInvitation"
          >
            <span v-if="inviteModal.submitting" class="spinner-sm"></span>
            <span v-else>{{ t('access.modals.sendInvite') }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- ─── МОДАЛКА: Редактирование прав ─────────────────────────── -->
    <div v-if="editPermsModal.open" class="modal-overlay" @click.self="closeEditPermsModal">
      <div class="modal-card">
        <div class="modal-header">
          <h3>{{ t('access.modals.editPermissionsTitle') }}</h3>
          <button class="btn-icon" @click="closeEditPermsModal">
            <X class="icon-sm" />
          </button>
        </div>

        <div class="modal-body">
          <div class="member-header-summary">
            <strong class="user-summary-name">{{
              editPermsModal.member?.user_name || editPermsModal.member?.user_email
            }}</strong>
            <span class="user-summary-email text-muted">{{ editPermsModal.member?.user_email }}</span>
          </div>

          <div class="form-group">
            <div class="permissions-header-row">
              <label class="form-label">{{ t('access.modals.permissionsSelect') }}</label>
              <div class="quick-perms-actions">
                <button type="button" class="btn-text" @click="selectAllEditPerms">Все</button>
                <span class="text-separator">/</span>
                <button type="button" class="btn-text" @click="clearAllEditPerms">Сбросить</button>
              </div>
            </div>

            <div class="permissions-list">
              <div
                v-for="p in ALL_PERMISSIONS"
                :key="p"
                class="perm-item-row"
                :class="{ selected: editPermsModal.permissions.includes(p) }"
                @click="togglePermission(editPermsModal.permissions, p)"
              >
                <div class="custom-checkbox-box" :class="{ checked: editPermsModal.permissions.includes(p) }">
                  <Check v-if="editPermsModal.permissions.includes(p)" class="custom-check-icon" />
                </div>
                <div class="perm-label-wrap">
                  <span class="perm-title">{{ permissionLabel(p) }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="modal-footer">
          <button class="btn-reset" @click="closeEditPermsModal">
            {{ t('common.cancel') }}
          </button>
          <button
            class="btn-add-game-primary"
            :disabled="editPermsModal.submitting"
            @click="submitEditPermissions"
          >
            <span v-if="editPermsModal.submitting" class="spinner-sm"></span>
            <span v-else>{{ t('common.save') }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- ─── МОДАЛКА: Подтверждение действий ──────────────────────── -->
    <div v-if="confirmModal.open" class="modal-overlay" @click.self="closeConfirmModal">
      <div class="modal-card">
        <div class="modal-header">
          <h3>{{ confirmModal.title }}</h3>
          <button class="btn-icon" @click="closeConfirmModal">
            <X class="icon-sm" />
          </button>
        </div>
        <div class="modal-body">
          <p class="confirm-desc">{{ confirmModal.desc }}</p>
        </div>
        <div class="modal-footer">
          <button class="btn-reset" @click="closeConfirmModal">
            {{ t('common.cancel') }}
          </button>
          <button
            class="btn-add-game-primary"
            :class="{ 'btn-danger': confirmModal.isDanger }"
            :disabled="confirmModal.submitting"
            @click="confirmModal.onConfirm"
          >
            <span v-if="confirmModal.submitting" class="spinner-sm"></span>
            <span v-else>{{ confirmModal.confirmText }}</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import {
  Users,
  UserPlus,
  UserX,
  Check,
  X,
  Search,
  Sliders,
  Clock,
  ShieldCheck,
} from 'lucide-vue-next';
import {
  listMembers,
  listOutgoingInvitations,
  sendInvitation,
  cancelInvitation,
  updateMemberPermissions,
  removeMember,
  ALL_PERMISSIONS,
  permissionLabel,
} from '@/entities/project';
import { searchUsers } from '@/entities/user';
import { showToast } from '@/shared/lib';

const { t } = useI18n();
const route = useRoute();

const projectId = computed(() => route.params.id);

const loading = ref(true);
const actionInProgress = ref(null);
const members = ref([]);
const pendingInvites = ref([]);

// ─── Modal States ─────────────────────────────────────────────
const searchHighlightedIndex = ref(-1);

const inviteModal = reactive({
  open: false,
  query: '',
  selectedUser: null,
  searchResults: [],
  searching: false,
  permissions: [...ALL_PERMISSIONS],
  submitting: false,
});

const editPermsModal = reactive({
  open: false,
  member: null,
  permissions: [],
  submitting: false,
});

const confirmModal = reactive({
  open: false,
  title: '',
  desc: '',
  confirmText: '',
  isDanger: false,
  submitting: false,
  onConfirm: null,
});

// ─── Data Loading ─────────────────────────────────────────────
async function loadData() {
  if (!projectId.value) return;
  loading.value = true;
  try {
    await Promise.all([loadMembersList(), loadPendingList()]);
  } catch (err) {
    console.error('Failed to load project collaborators', err);
  } finally {
    loading.value = false;
  }
}

async function loadMembersList() {
  try {
    members.value = await listMembers(projectId.value);
  } catch {
    members.value = [];
  }
}

async function loadPendingList() {
  try {
    const list = await listOutgoingInvitations(projectId.value);
    pendingInvites.value = list.filter(
      (inv) => Number(inv.project_id) === Number(projectId.value)
    );
  } catch {
    pendingInvites.value = [];
  }
}

onMounted(() => {
  loadData();
});

// ─── Permissions Helpers ──────────────────────────────────────
function hasAllPermissions(perms) {
  if (!perms || !Array.isArray(perms) || perms.length === 0) return false;
  return ALL_PERMISSIONS.every((p) => perms.includes(p));
}

function togglePermission(list, perm) {
  const idx = list.indexOf(perm);
  if (idx > -1) {
    list.splice(idx, 1);
  } else {
    list.push(perm);
  }
}

function selectAllPerms() {
  inviteModal.permissions = [...ALL_PERMISSIONS];
}

function clearAllPerms() {
  inviteModal.permissions = [];
}

function selectAllEditPerms() {
  editPermsModal.permissions = [...ALL_PERMISSIONS];
}

function clearAllEditPerms() {
  editPermsModal.permissions = [];
}

// ─── Invite Modal Handlers ────────────────────────────────────
function openInviteModal() {
  inviteModal.query = '';
  inviteModal.selectedUser = null;
  inviteModal.searchResults = [];
  searchHighlightedIndex.value = -1;
  inviteModal.permissions = [...ALL_PERMISSIONS];
  inviteModal.open = true;
}

function closeInviteModal() {
  inviteModal.open = false;
}

let searchTimer = null;
function onSearchUserInput() {
  clearTimeout(searchTimer);
  const q = inviteModal.query.trim();
  if (!q) {
    inviteModal.searchResults = [];
    searchHighlightedIndex.value = -1;
    return;
  }
  inviteModal.searching = true;
  searchTimer = setTimeout(async () => {
    try {
      const res = await searchUsers({ query: q, limit: 10 });
      const users = res.users || [];
      inviteModal.searchResults = users.filter((u) => {
        const isSystem =
          u.role === 'USER_ROLE_MODERATOR' ||
          u.role === 'moderator' ||
          u.role === 2 ||
          u.role === 'USER_ROLE_ADMIN' ||
          u.role === 'admin' ||
          u.role === 3;
        return !isSystem;
      });
      searchHighlightedIndex.value = -1;
    } catch {
      inviteModal.searchResults = [];
      searchHighlightedIndex.value = -1;
    } finally {
      inviteModal.searching = false;
    }
  }, 300);
}

function onSearchKeyDown() {
  if (inviteModal.searchResults.length === 0) return;
  searchHighlightedIndex.value =
    (searchHighlightedIndex.value + 1) % inviteModal.searchResults.length;
}

function onSearchKeyUp() {
  if (inviteModal.searchResults.length === 0) return;
  searchHighlightedIndex.value =
    (searchHighlightedIndex.value - 1 + inviteModal.searchResults.length) %
    inviteModal.searchResults.length;
}

function onSearchKeyEnter() {
  if (
    searchHighlightedIndex.value >= 0 &&
    searchHighlightedIndex.value < inviteModal.searchResults.length
  ) {
    selectUserToInvite(inviteModal.searchResults[searchHighlightedIndex.value]);
  }
}

function closeSearchDropdown() {
  inviteModal.searchResults = [];
  searchHighlightedIndex.value = -1;
}

function selectUserToInvite(user) {
  inviteModal.selectedUser = user;
  inviteModal.searchResults = [];
  inviteModal.query = '';
  searchHighlightedIndex.value = -1;
}

function clearSelectedUser() {
  inviteModal.selectedUser = null;
  searchHighlightedIndex.value = -1;
}

async function submitInvitation() {
  const targetEmail = inviteModal.selectedUser?.email || inviteModal.query.trim();
  const targetId = inviteModal.selectedUser?.id || '';
  if (!targetEmail && !targetId) {
    showToast(t('access.messages.selectUserWarning'), 'warning');
    return;
  }
  if (inviteModal.selectedUser) {
    const u = inviteModal.selectedUser;
    const isSystem =
      u.role === 'USER_ROLE_MODERATOR' ||
      u.role === 'moderator' ||
      u.role === 2 ||
      u.role === 'USER_ROLE_ADMIN' ||
      u.role === 'admin' ||
      u.role === 3;
    if (isSystem) {
      showToast(t('access.messages.cannotInviteSystemUser'), 'error');
      return;
    }
  }
  if (inviteModal.permissions.length === 0) {
    showToast(t('access.messages.selectPermsWarning'), 'warning');
    return;
  }

  inviteModal.submitting = true;
  try {
    await sendInvitation(projectId.value, {
      invitee_id: targetId,
      invitee_email: targetEmail,
      permissions: inviteModal.permissions,
    });
    showToast(t('access.messages.inviteSent'), 'success');
    closeInviteModal();
    await loadPendingList();
  } catch (err) {
    const msg = err.response?.data?.message || err.message || '';
    if (msg.includes('system user') || msg.includes('cannot invite system user')) {
      showToast(t('access.messages.cannotInviteSystemUser'), 'error');
    } else if (msg.includes('blocked') || msg.includes('user is blocked')) {
      showToast(t('access.messages.userIsBlocked'), 'error');
    } else if (msg.includes('already member')) {
      showToast(t('access.messages.alreadyMember'), 'error');
    } else if (msg.includes('already invited')) {
      showToast(t('access.messages.alreadyInvited'), 'error');
    } else if (msg.includes('cannot invite self')) {
      showToast(t('access.messages.cannotInviteSelf'), 'error');
    } else {
      showToast(msg || t('common.error'), 'error');
    }
  } finally {
    inviteModal.submitting = false;
  }
}

// ─── Edit Permissions Handlers ────────────────────────────────
function openEditPermissionsModal(member) {
  editPermsModal.member = member;
  editPermsModal.permissions = [...(member.permissions || [])];
  editPermsModal.open = true;
}

function closeEditPermsModal() {
  editPermsModal.open = false;
  editPermsModal.member = null;
  editPermsModal.permissions = [];
}

async function submitEditPermissions() {
  if (editPermsModal.permissions.length === 0) {
    showToast(t('access.messages.selectPermsWarning'), 'warning');
    return;
  }
  editPermsModal.submitting = true;
  try {
    await updateMemberPermissions(
      projectId.value,
      editPermsModal.member.user_id,
      editPermsModal.permissions
    );
    showToast(t('access.messages.permissionsUpdated'), 'success');
    closeEditPermsModal();
    await loadMembersList();
  } catch (err) {
    showToast(err.response?.data?.message || t('common.error'), 'error');
  } finally {
    editPermsModal.submitting = false;
  }
}

// ─── Member Removal & Invite Cancellation ─────────────────────
function promptRemoveMember(member) {
  confirmModal.title = t('access.modals.removeConfirmTitle');
  confirmModal.desc = `${member.user_name || member.user_email}: ${t('access.modals.removeConfirmDesc')}`;
  confirmModal.confirmText = t('access.actions.removeMember');
  confirmModal.isDanger = true;
  confirmModal.onConfirm = async () => {
    confirmModal.submitting = true;
    try {
      await removeMember(projectId.value, member.user_id);
      showToast(t('access.messages.memberRemoved'), 'success');
      closeConfirmModal();
      await loadMembersList();
    } catch (err) {
      showToast(err.response?.data?.message || t('common.error'), 'error');
    } finally {
      confirmModal.submitting = false;
    }
  };
  confirmModal.open = true;
}

async function handleCancelInvitation(invitationId) {
  actionInProgress.value = invitationId;
  try {
    await cancelInvitation(invitationId);
    showToast(t('access.messages.inviteCanceled'), 'success');
    await loadPendingList();
  } catch (err) {
    showToast(err.response?.data?.message || t('common.error'), 'error');
  } finally {
    actionInProgress.value = null;
  }
}

function closeConfirmModal() {
  confirmModal.open = false;
  confirmModal.onConfirm = null;
}

function formatDate(dateStr) {
  if (!dateStr) return '—';
  try {
    const d = new Date(dateStr);
    return d.toLocaleDateString(undefined, {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  } catch {
    return dateStr;
  }
}
</script>

<style scoped>
.tab-fade-in {
  animation: fadeIn 0.3s ease;
}
@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

.tab-content {
  padding: 32px;
  max-width: 1000px;
}

.page-header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  gap: 16px;
}

.page-title {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--text-main, #f0f6fc);
}

.page-subtitle {
  margin: 0;
  font-size: 14px;
  color: var(--text-muted, #8b949e);
}

.access-table-card {
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-md, 8px);
  overflow: hidden;
}

.empty-state {
  padding: 48px 24px;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}
.empty-state h3 {
  margin: 0;
  color: var(--text-main, #f0f6fc);
}
.empty-icon-wrap {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: var(--bg-tertiary, #21262d);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--primary, #58a6ff);
}

.state-container {
  padding: 60px 20px;
  text-align: center;
  color: var(--text-tertiary, #8b949e);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.table-wrapper {
  width: 100%;
  overflow-x: auto;
}

.yandex-games-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
}
.yandex-games-table th {
  color: var(--text-tertiary, #8b949e);
  font-size: 13px;
  font-weight: 500;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border, #30363d);
  background: var(--bg-secondary, #0d1117);
  white-space: nowrap;
}

.yandex-games-table th.col-user { width: 28%; }
.yandex-games-table th.col-perms { width: 34%; }
.yandex-games-table th.col-status { width: 14%; }
.yandex-games-table th.col-date { width: 14%; }
.yandex-games-table th.col-actions { width: 10%; text-align: right; padding-right: 16px; }

.table-row {
  border-bottom: 1px solid var(--border, #21262d);
  transition: background-color 0.15s ease;
}
.table-row:hover {
  background: var(--bg-secondary, #161b22);
}
.table-row td {
  padding: 12px 16px;
  vertical-align: middle;
}
.table-row td.col-actions {
  text-align: right;
  padding-right: 16px;
}

.pending-row {
  background: rgba(245, 176, 39, 0.02);
}

.user-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.user-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-main, #f0f6fc);
}
.user-email {
  font-size: 12px;
  color: var(--text-muted, #8b949e);
}

.date-text {
  font-size: 13px;
  font-weight: 400;
  color: var(--text-muted, #b0b8c4);
}

.perms-badges-wrap {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  align-items: center;
}

.badge-role-collab {
  display: inline-flex;
  align-items: center;
  font-size: 11px;
  font-weight: 500;
  padding: 2px 7px;
  border-radius: 6px;
  background: var(--bg-tertiary, #21262d);
  color: var(--primary, #58a6ff);
  border: 1px solid var(--border, #30363d);
  white-space: nowrap;
}

.badge-full-access {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 12px;
  font-weight: 600;
  padding: 3px 10px;
  border-radius: 6px;
  background: rgba(88, 166, 255, 0.12);
  color: #58a6ff;
  border: 1px solid rgba(88, 166, 255, 0.35);
  cursor: default;
  transition: all 0.15s ease;
  white-space: nowrap;
}
.badge-full-access:hover {
  background: rgba(88, 166, 255, 0.2);
  border-color: rgba(88, 166, 255, 0.55);
}

.status-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 26px;
  padding: 0 12px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  white-space: nowrap;
}
.status-pending {
  background: rgba(245, 176, 39, 0.12);
  border: 1px solid rgba(245, 176, 39, 0.35);
  color: #f5b027;
}
.status-active {
  background: rgba(46, 204, 113, 0.12);
  border: 1px solid rgba(46, 204, 113, 0.35);
  color: #2ecc71;
}

.row-actions {
  display: inline-flex;
  align-items: center;
  gap: 8px;
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

/* Modals */
.modal-overlay {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0, 0, 0, 0.65);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 20px;
  backdrop-filter: blur(2px);
}
.modal-card {
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: 8px;
  width: 100%;
  max-width: 500px;
  display: flex;
  flex-direction: column;
  box-shadow: 0 12px 36px rgba(0, 0, 0, 0.4);
}
.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border, #30363d);
}
.modal-header h3 { margin: 0; font-size: 16px; font-weight: 600; color: var(--text-main); }
.modal-body { padding: 20px; display: flex; flex-direction: column; gap: 16px; }
.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 20px;
  border-top: 1px solid var(--border, #30363d);
}

.form-group { display: flex; flex-direction: column; gap: 6px; }
.form-label { font-size: 13px; font-weight: 500; color: var(--text-main); }
.form-input {
  background: var(--bg-secondary, #0d1117);
  border: 1px solid var(--border, #30363d);
  color: var(--text-main, #c9d1d9);
  padding: 8px 12px;
  border-radius: 6px;
  font-size: 14px;
}
.req { color: var(--danger, #f85149); }

.member-header-summary {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 10px 14px;
  background: var(--bg-secondary, #0d1117);
  border: 1px solid var(--border, #30363d);
  border-radius: 6px;
}
.user-summary-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
}
.user-summary-email {
  font-size: 12px;
  color: var(--text-muted, #8b949e);
}

/* Buttons */
.btn-add-game-primary {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 18px;
  background: var(--primary, #58a6ff);
  color: #ffffff;
  border: none;
  border-radius: var(--radius-sm, 6px);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.15s;
}
.btn-add-game-primary:hover { background: var(--primary-hover, #79c0ff); }
.btn-add-game-primary:disabled { opacity: 0.6; cursor: not-allowed; }
.btn-add-game-primary.btn-danger { background: var(--danger, #f85149); }
.btn-add-game-primary.btn-danger:hover { background: #ff7b72; }

.btn-reset {
  padding: 6px 14px;
  background: var(--bg-secondary, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  color: var(--text-main, #f0f6fc);
  font-size: 13px;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}
.btn-reset:hover { background: var(--bg-tertiary); }

/* Search User */
.search-user-wrapper { position: relative; width: 100%; }
.search-input-box { position: relative; display: flex; align-items: center; width: 100%; }
.search-icon {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--text-muted, #8b949e);
  pointer-events: none;
}
.search-loader {
  position: absolute;
  right: 12px;
  top: 50%;
  transform: translateY(-50%);
  border-color: var(--border, #30363d);
  border-top-color: var(--primary, #58a6ff);
}
.has-search-icon {
  padding-left: 38px !important;
  width: 100%;
  box-sizing: border-box;
}
.user-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 0; right: 0;
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: 6px;
  max-height: 220px;
  overflow-y: auto;
  z-index: 50;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.5);
}
.user-dropdown-item {
  padding: 10px 14px;
  cursor: pointer;
  display: flex;
  flex-direction: column;
  gap: 2px;
  transition: background-color 0.15s ease;
}
.user-dropdown-item:hover,
.user-dropdown-item.active {
  background: var(--bg-tertiary, #21262d);
}
.user-dropdown-item.active .user-item-name {
  color: var(--primary, #58a6ff);
}
.user-item-name { font-size: 13px; font-weight: 500; }
.user-item-email { font-size: 12px; color: var(--text-muted); }
.selected-user-pill {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  border-radius: 6px;
  margin-top: 8px;
}
.pill-info { display: flex; flex-direction: column; }
.pill-info strong { font-size: 13px; }
.pill-email { font-size: 12px; color: var(--text-muted); }

/* Permissions Selection */
.permissions-header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.quick-perms-actions {
  display: flex;
  gap: 6px;
  align-items: center;
}
.btn-text {
  background: none;
  border: none;
  color: var(--primary, #58a6ff);
  font-size: 12px;
  cursor: pointer;
  padding: 0;
}
.btn-text:hover { text-decoration: underline; }
.text-separator { color: var(--text-muted, #8b949e); font-size: 12px; }

.permissions-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
  max-height: 240px;
  overflow-y: auto;
  border: 1px solid var(--border, #30363d);
  padding: 8px;
  border-radius: 6px;
  background: var(--bg-secondary, #0d1117);
}

.perm-item-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 12px;
  border-radius: 6px;
  cursor: pointer;
  background: var(--bg-card, #161b22);
  border: 1px solid transparent;
  transition: all 0.15s ease;
  user-select: none;
}
.perm-item-row:hover {
  background: var(--bg-tertiary, #21262d);
  border-color: var(--border, #30363d);
}
.perm-item-row.selected {
  background: rgba(88, 166, 255, 0.08);
  border-color: rgba(88, 166, 255, 0.35);
}

.custom-checkbox-box {
  width: 18px;
  height: 18px;
  border-radius: 4px;
  border: 1.5px solid var(--border, #484f58);
  background: var(--bg-secondary, #0d1117);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: all 0.15s ease;
}
.custom-checkbox-box.checked {
  background: var(--primary, #58a6ff);
  border-color: var(--primary, #58a6ff);
}

.custom-check-icon {
  width: 12px;
  height: 12px;
  color: #ffffff;
  stroke-width: 3;
}

.perm-label-wrap { display: flex; align-items: center; flex: 1; }
.perm-title { font-size: 13px; font-weight: 500; color: var(--text-main, #f0f6fc); }
.text-muted { color: var(--text-muted, #8b949e) !important; }

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
.spinner-md {
  width: 32px;
  height: 32px;
  border: 3px solid var(--border, #30363d);
  border-top-color: var(--primary, #58a6ff);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin: 0 auto;
  display: block;
  flex-shrink: 0;
}
@keyframes spin { to { transform: rotate(360deg); } }
</style>
