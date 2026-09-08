<template>
  <div class="projects-page-container">
    <div class="main-content-wrap">
      <!-- Header -->
      <div class="page-header">
        <div class="header-titles">
          <h1 class="page-title">{{ t('access.title') }}</h1>
          <p class="page-subtitle">{{ t('access.subtitle') }}</p>
        </div>
        <button class="btn-add-game-primary" @click="openInviteModal()">
          <UserPlus class="icon-sm" />
          <span>{{ t('access.actions.invite') }}</span>
        </button>
      </div>

      <!-- Navigation Tabs -->
      <div class="tabs-bar">
        <button
          class="tab-item"
          :class="{ active: activeTab === 'incoming' }"
          @click="activeTab = 'incoming'"
        >
          <span>{{ t('access.tabs.incoming') }}</span>
          <span v-if="incomingList.length > 0" class="tab-badge incoming-badge">{{
            incomingList.length
          }}</span>
        </button>

        <button
          class="tab-item"
          :class="{ active: activeTab === 'shared' }"
          @click="activeTab = 'shared'"
        >
          <span>{{ t('access.tabs.shared') }}</span>
          <span v-if="sharedProjectsList.length > 0" class="tab-badge">{{
            sharedProjectsList.length
          }}</span>
        </button>

        <button
          class="tab-item"
          :class="{ active: activeTab === 'myProjects' }"
          @click="activeTab = 'myProjects'"
        >
          <span>{{ t('access.tabs.myProjects') }}</span>
          <span v-if="myProjects.length > 0" class="tab-badge">{{ myProjects.length }}</span>
        </button>

        <button
          class="tab-item"
          :class="{ active: activeTab === 'outgoing' }"
          @click="activeTab = 'outgoing'"
        >
          <span>{{ t('access.tabs.outgoing') }}</span>
          <span v-if="outgoingList.length > 0" class="tab-badge">{{ outgoingList.length }}</span>
        </button>

        <button
          class="tab-item"
          :class="{ active: activeTab === 'blacklist' }"
          @click="activeTab = 'blacklist'"
        >
          <span>{{ t('access.tabs.blacklist') }}</span>
          <span v-if="blockedUsersList.length > 0" class="tab-badge text-danger-badge">{{
            blockedUsersList.length
          }}</span>
        </button>
      </div>

      <!-- Loading State -->
      <div v-if="loading" class="state-container">
        <div class="spinner-md"></div>
        <p>{{ t('common.loading') }}</p>
      </div>

      <div v-else class="tab-content">
        <!-- 1. ВХОДЯЩИЕ ЗАПРОСЫ -->
        <div v-if="activeTab === 'incoming'" class="tab-pane">
          <div v-if="incomingList.length === 0" class="state-container empty-card">
            <div class="empty-icon-wrap">
              <Mail class="icon-lg" />
            </div>
            <h3>{{ t('access.messages.emptyIncoming') }}</h3>
          </div>

          <div v-else class="table-wrapper">
            <table class="yandex-games-table">
              <thead>
                <tr>
                  <th class="col-game">Проект</th>
                  <th class="col-user">Отправитель</th>
                  <th class="col-date">Дата</th>
                  <th class="col-actions">Действия</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="inv in incomingList" :key="inv.id" class="table-row">
                  <td class="col-game">
                    <div class="game-cell">
                      <div class="game-icon-box">
                        <img
                          v-if="inv.project_icon"
                          :src="getMediaUrl(inv.project_icon)"
                          alt="Icon"
                          class="game-icon-img"
                        />
                        <div v-else class="game-icon-mock">
                          <span>Draft</span>
                        </div>
                      </div>
                      <div class="game-text">
                        <div class="game-title">
                          <span>{{ inv.project_title || '—' }}</span>
                        </div>
                        <div class="game-type-label">
                          <span v-for="perm in inv.permissions" :key="perm" class="badge-role-collab">
                            <Shield class="icon-xs" style="margin-right:2px;" />
                            {{ permissionLabel(perm) }}
                          </span>
                        </div>
                      </div>
                    </div>
                  </td>
                  <td class="col-user">
                    <div class="user-text">
                      <div class="user-name">{{ inv.inviter_name || inv.inviter_email }}</div>
                      <div v-if="inv.inviter_name && inv.inviter_email" class="user-email text-muted">{{ inv.inviter_email }}</div>
                    </div>
                  </td>
                  <td class="col-date">
                    <span class="date-text">
                      <Clock class="icon-xs" style="margin-right: 4px; vertical-align: middle;"/>
                      {{ formatDate(inv.created_at) }}
                    </span>
                  </td>
                  <td class="col-actions">
                    <div class="row-actions">
                      <button
                        class="btn-icon text-success-hover"
                        :disabled="actionInProgress === inv.id"
                        title="Принять"
                        @click="handleRespond(inv.id, true)"
                      >
                        <Check class="icon-xs" />
                      </button>
                      <button
                        class="btn-icon text-warning-hover"
                        :disabled="actionInProgress === inv.id"
                        title="Отклонить"
                        @click="handleRespond(inv.id, false)"
                      >
                        <X class="icon-xs" />
                      </button>
                      <button
                        class="btn-icon text-danger-hover"
                        :disabled="actionInProgress === inv.id"
                        title="Заблокировать"
                        @click="promptBlockUser(inv.inviter_id, inv.inviter_name || inv.inviter_email)"
                      >
                        <Ban class="icon-xs" />
                      </button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- 2. ПОДКЛЮЧЕННЫЕ ПРОЕКТЫ -->
        <div v-if="activeTab === 'shared'" class="tab-pane">
          <div v-if="sharedProjectsList.length === 0" class="state-container empty-card">
            <div class="empty-icon-wrap">
              <Users class="icon-lg" />
            </div>
            <h3>{{ t('access.messages.emptyShared') }}</h3>
          </div>

          <div v-else class="table-wrapper">
            <table class="yandex-games-table">
              <thead>
                <tr>
                  <th class="col-game">Проект</th>
                  <th class="col-user">Владелец</th>
                  <th class="col-date">Присоединился</th>
                  <th class="col-actions">Действия</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in sharedProjectsList" :key="item.project.id" class="table-row" @click="goToProject(item.project.id)">
                  <td class="col-game">
                    <div class="game-cell">
                      <div class="game-icon-box">
                        <img
                          v-if="item.project.icon_path"
                          :src="getMediaUrl(item.project.icon_path)"
                          alt="Icon"
                          class="game-icon-img"
                        />
                        <div v-else class="game-icon-mock">
                          <span>Draft</span>
                        </div>
                      </div>
                      <div class="game-text">
                        <div class="game-title">
                          <span>{{ item.project.title_ru || item.project.title_en || '—' }}</span>
                        </div>
                        <div class="game-type-label">
                          <span v-for="perm in item.permissions" :key="perm" class="badge-role-collab">
                            <ShieldCheck class="icon-xs" style="margin-right:2px;" />
                            {{ permissionLabel(perm) }}
                          </span>
                        </div>
                      </div>
                    </div>
                  </td>
                  <td class="col-user">
                    <div class="user-text">
                      <div class="user-name">{{ item.owner_name || item.owner_email }}</div>
                      <div v-if="item.owner_name && item.owner_email" class="user-email text-muted">{{ item.owner_email }}</div>
                    </div>
                  </td>
                  <td class="col-date">
                    <span class="date-text">
                      {{ formatDate(item.joined_at) }}
                    </span>
                  </td>
                  <td class="col-actions" @click.stop>
                    <div class="row-actions">
                      <button
                        class="btn-icon"
                        title="Открыть проект"
                        @click="goToProject(item.project.id)"
                      >
                        <ExternalLink class="icon-xs" />
                      </button>
                      <button
                        class="btn-icon text-warning-hover"
                        title="Покинуть проект"
                        @click="promptLeaveProject(item.project.id, item.project.title_ru || item.project.title_en)"
                      >
                        <LogOut class="icon-xs" />
                      </button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- 3. МОИ ПРОЕКТЫ И ДОСТУПЫ -->
        <div v-if="activeTab === 'myProjects'" class="tab-pane">
          <div v-if="myProjects.length === 0" class="state-container empty-card">
            <div class="empty-icon-wrap">
              <Shield class="icon-lg" />
            </div>
            <h3>{{ t('access.messages.emptyMyProjects') }}</h3>
          </div>

          <div v-else class="my-projects-list">
            <div v-for="proj in myProjects" :key="proj.id" class="project-block">
              <div class="project-block-header">
                <div class="game-cell">
                  <div class="game-icon-box">
                    <img
                      v-if="proj.icon_path"
                      :src="getMediaUrl(proj.icon_path)"
                      alt="Icon"
                      class="game-icon-img"
                    />
                    <div v-else class="game-icon-mock">
                      <span>Draft</span>
                    </div>
                  </div>
                  <div class="game-text">
                    <div class="game-title">
                      <span>{{ proj.title_ru || proj.title_en || '—' }}</span>
                    </div>
                    <div class="game-type-label">ID: {{ proj.id }}</div>
                  </div>
                </div>

                <div class="row-actions">
                  <button class="btn-secondary" @click="openInviteModal(proj.id)">
                    <UserPlus class="icon-xs" />
                    {{ t('access.actions.invite') }}
                  </button>
                  <button class="btn-reset" @click="goToProject(proj.id)">
                    <ExternalLink class="icon-xs" />
                    {{ t('access.actions.openProject') }}
                  </button>
                </div>
              </div>

              <!-- Members Table -->
              <div class="table-wrapper">
                <div v-if="!projectMembers[proj.id] || projectMembers[proj.id].length === 0" class="empty-members">
                  <span class="text-muted">{{ t('access.messages.noMembers') }}</span>
                </div>
                <table v-else class="yandex-games-table inner-table">
                  <thead>
                    <tr>
                      <th class="col-user">{{ t('common.name') }}</th>
                      <th class="col-perms">{{ t('access.modals.permissionsSelect') }}</th>
                      <th class="col-date">{{ t('common.created') }}</th>
                      <th class="col-actions"></th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="m in projectMembers[proj.id]" :key="m.id" class="table-row">
                      <td class="col-user">
                        <div class="user-text">
                          <div class="user-name">{{ m.user_name || m.user_email }}</div>
                          <div v-if="m.user_name" class="user-email text-muted">{{ m.user_email }}</div>
                        </div>
                      </td>
                      <td class="col-perms">
                        <span v-for="p in m.permissions" :key="p" class="badge-role-collab">
                          {{ permissionLabel(p) }}
                        </span>
                      </td>
                      <td class="col-date">
                        <span class="date-text">{{ formatDate(m.created_at) }}</span>
                      </td>
                      <td class="col-actions">
                        <div class="row-actions">
                          <button
                            class="btn-icon"
                            :title="t('access.actions.editPermissions')"
                            @click="openEditPermissionsModal(proj, m)"
                          >
                            <Sliders class="icon-xs" />
                          </button>
                          <button
                            class="btn-icon text-danger-hover"
                            :title="t('access.actions.removeMember')"
                            @click="promptRemoveMember(proj.id, m)"
                          >
                            <UserX class="icon-xs" />
                          </button>
                        </div>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </div>

        <!-- 4. ИСХОДЯЩИЕ ПРИГЛАШЕНИЯ -->
        <div v-if="activeTab === 'outgoing'" class="tab-pane">
          <div v-if="outgoingList.length === 0" class="state-container empty-card">
            <div class="empty-icon-wrap">
              <Mail class="icon-lg" />
            </div>
            <h3>{{ t('access.messages.emptyOutgoing') }}</h3>
          </div>

          <div v-else class="table-wrapper">
            <table class="yandex-games-table">
              <thead>
                <tr>
                  <th class="col-game">Проект</th>
                  <th class="col-user">Кому</th>
                  <th class="col-status">Статус</th>
                  <th class="col-date">Дата</th>
                  <th class="col-actions">Действия</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="inv in outgoingList" :key="inv.id" class="table-row">
                  <td class="col-game">
                    <div class="game-text">
                      <div class="game-title">
                        <span>{{ inv.project_title || '—' }}</span>
                      </div>
                      <div class="game-type-label">
                        <span v-for="perm in inv.permissions" :key="perm" class="badge-role-collab">
                          {{ permissionLabel(perm) }}
                        </span>
                      </div>
                    </div>
                  </td>
                  <td class="col-user">
                    <div class="user-text">
                      <div class="user-name">{{ inv.invitee_email || inv.invitee_id }}</div>
                    </div>
                  </td>
                  <td class="col-status">
                    <span class="status-pill status-pending">{{ t('access.statuses.pending') }}</span>
                  </td>
                  <td class="col-date">
                    <span class="date-text">
                      <Clock class="icon-xs" style="margin-right: 4px; vertical-align: middle;"/>
                      {{ formatDate(inv.created_at) }}
                    </span>
                  </td>
                  <td class="col-actions">
                    <div class="row-actions">
                      <button
                        class="btn-icon text-danger-hover"
                        :disabled="actionInProgress === inv.id"
                        title="Отменить"
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

        <!-- 5. ЧЕРНЫЙ СПИСОК -->
        <div v-if="activeTab === 'blacklist'" class="tab-pane">
          <div v-if="blockedUsersList.length === 0" class="state-container empty-card">
            <div class="empty-icon-wrap">
              <Ban class="icon-lg" />
            </div>
            <h3>{{ t('access.messages.emptyBlacklist') }}</h3>
          </div>

          <div v-else class="table-wrapper">
            <table class="yandex-games-table">
              <thead>
                <tr>
                  <th class="col-user">{{ t('common.name') }}</th>
                  <th class="col-user">{{ t('auth.email') }}</th>
                  <th class="col-date">{{ t('common.date') }}</th>
                  <th class="col-actions">Действия</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="b in blockedUsersList" :key="b.id" class="table-row">
                  <td class="col-user">
                    <div class="user-text">
                      <div class="user-name">{{ b.blocked_user_name || '—' }}</div>
                    </div>
                  </td>
                  <td class="col-user">
                    <div class="user-text">
                      <div class="user-email text-muted">{{ b.blocked_user_email || b.blocked_user_id }}</div>
                    </div>
                  </td>
                  <td class="col-date">
                    <span class="date-text">{{ formatDate(b.created_at) }}</span>
                  </td>
                  <td class="col-actions">
                    <div class="row-actions">
                      <button
                        class="btn-secondary"
                        :disabled="actionInProgress === b.id"
                        @click="handleUnblock(b.blocked_user_id)"
                      >
                        <Check class="icon-xs" />
                        {{ t('access.actions.unblock') }}
                      </button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>

    <!-- МОДАЛКИ (оформление как в ProjectsListPage/и других страницах) -->
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
            <label class="form-label">{{ t('access.modals.selectProject') }} <span class="req">*</span></label>
            <select v-model="inviteModal.projectId" class="form-input">
              <option :value="null" disabled>-- {{ t('access.modals.selectProject') }} --</option>
              <option v-for="p in myProjects" :key="p.id" :value="p.id">
                {{ p.title_ru || p.title_en || `#${p.id}` }}
              </option>
            </select>
          </div>

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
                />
                <div v-if="inviteModal.searching" class="spinner-sm search-loader"></div>
              </div>

              <div v-if="inviteModal.searchResults.length > 0" class="user-dropdown">
                <div
                  v-for="u in inviteModal.searchResults"
                  :key="u.id"
                  class="user-dropdown-item"
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
                <button class="btn-icon" @click="clearSelectedUser">
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

            <div class="permissions-checkbox-list">
              <label
                v-for="p in ALL_PERMISSIONS"
                :key="p"
                class="perm-checkbox-item"
                :class="{ checked: inviteModal.permissions.includes(p) }"
              >
                <input
                  v-model="inviteModal.permissions"
                  type="checkbox"
                  :value="p"
                  class="checkbox-input"
                />
                <div class="checkbox-label-wrap">
                  <span class="perm-title">{{ permissionLabel(p) }}</span>
                </div>
              </label>
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

            <div class="permissions-checkbox-list">
              <label
                v-for="p in ALL_PERMISSIONS"
                :key="p"
                class="perm-checkbox-item"
                :class="{ checked: editPermsModal.permissions.includes(p) }"
              >
                <input
                  v-model="editPermsModal.permissions"
                  type="checkbox"
                  :value="p"
                  class="checkbox-input"
                />
                <div class="checkbox-label-wrap">
                  <span class="perm-title">{{ permissionLabel(p) }}</span>
                </div>
              </label>
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
import { ref, reactive, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import {
  Users,
  UserPlus,
  UserX,
  Shield,
  ShieldCheck,
  Check,
  X,
  Ban,
  LogOut,
  ExternalLink,
  Loader2,
  Search,
  Sliders,
  Mail,
  Clock,
} from 'lucide-vue-next';
import {
  listProjects,
  listIncomingInvitations,
  listOutgoingInvitations,
  respondInvitation,
  cancelInvitation,
  listMembers,
  updateMemberPermissions,
  removeMember,
  leaveProject,
  listSharedProjects,
  blockUser,
  unblockUser,
  listBlockedUsers,
  sendInvitation,
  ALL_PERMISSIONS,
  getMediaUrl,
  permissionLabel,
} from '@/entities/project';
import { searchUsers } from '@/entities/user';
import { showToast } from '@/shared/lib';

const { t } = useI18n();
const router = useRouter();

const activeTab = ref('incoming');
const loading = ref(true);
const actionInProgress = ref(null);

const incomingList = ref([]);
const sharedProjectsList = ref([]);
const myProjects = ref([]);
const outgoingList = ref([]);
const blockedUsersList = ref([]);
const projectMembers = ref({}); // { [projectId]: membersList }

// ─── Modal States ─────────────────────────────────────────────
const inviteModal = reactive({
  open: false,
  projectId: null,
  query: '',
  selectedUser: null,
  searchResults: [],
  searching: false,
  permissions: [...ALL_PERMISSIONS],
  submitting: false,
});

const editPermsModal = reactive({
  open: false,
  projectId: null,
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
async function loadAllData() {
  loading.value = true;
  try {
    await Promise.all([
      loadIncoming(),
      loadShared(),
      loadMyProjects(),
      loadOutgoing(),
      loadBlacklist(),
    ]);
  } catch (err) {
    console.error('Failed to load shared access data', err);
  } finally {
    loading.value = false;
  }
}

async function loadIncoming() {
  try {
    incomingList.value = await listIncomingInvitations();
  } catch (err) {
    incomingList.value = [];
  }
}

async function loadShared() {
  try {
    sharedProjectsList.value = await listSharedProjects();
  } catch (err) {
    sharedProjectsList.value = [];
  }
}

async function loadMyProjects() {
  try {
    const res = await listProjects({ limit: 100 });
    myProjects.value = res.projects || [];
    // Load members for each project
    await Promise.all(
      myProjects.value.map(async (p) => {
        try {
          const members = await listMembers(p.id);
          projectMembers.value[p.id] = members;
        } catch {
          projectMembers.value[p.id] = [];
        }
      })
    );
  } catch (err) {
    myProjects.value = [];
  }
}

async function loadOutgoing() {
  try {
    outgoingList.value = await listOutgoingInvitations();
  } catch (err) {
    outgoingList.value = [];
  }
}

async function loadBlacklist() {
  try {
    blockedUsersList.value = await listBlockedUsers();
  } catch (err) {
    blockedUsersList.value = [];
  }
}

onMounted(() => {
  loadAllData();
});

// ─── Actions: Incoming ────────────────────────────────────────
async function handleRespond(invitationId, accept) {
  actionInProgress.value = invitationId;
  try {
    await respondInvitation(invitationId, accept);
    showToast(
      accept ? t('access.messages.inviteAccepted') : t('access.messages.inviteDeclined'),
      'success'
    );
    await Promise.all([loadIncoming(), loadShared()]);
  } catch (err) {
    showToast(err.response?.data?.message || t('common.error'), 'error');
  } finally {
    actionInProgress.value = null;
  }
}

function promptBlockUser(userId, userName) {
  confirmModal.title = t('access.modals.blockConfirmTitle');
  confirmModal.desc = `${userName}: ${t('access.modals.blockConfirmDesc')}`;
  confirmModal.confirmText = t('access.actions.block');
  confirmModal.isDanger = true;
  confirmModal.onConfirm = async () => {
    confirmModal.submitting = true;
    try {
      await blockUser(userId);
      showToast(t('access.messages.userBlocked'), 'success');
      closeConfirmModal();
      await Promise.all([loadIncoming(), loadBlacklist()]);
    } catch (err) {
      showToast(err.response?.data?.message || t('common.error'), 'error');
    } finally {
      confirmModal.submitting = false;
    }
  };
  confirmModal.open = true;
}

// ─── Actions: Shared Projects ────────────────────────────────
function goToProject(projectId) {
  router.push(`/projects/${projectId}`);
}

function promptLeaveProject(projectId, projectTitle) {
  confirmModal.title = t('access.modals.leaveConfirmTitle');
  confirmModal.desc = `«${projectTitle}»: ${t('access.modals.leaveConfirmDesc')}`;
  confirmModal.confirmText = t('access.actions.leaveProject');
  confirmModal.isDanger = true;
  confirmModal.onConfirm = async () => {
    confirmModal.submitting = true;
    try {
      await leaveProject(projectId);
      showToast(t('access.messages.leftProject'), 'success');
      closeConfirmModal();
      await loadShared();
    } catch (err) {
      showToast(err.response?.data?.message || t('common.error'), 'error');
    } finally {
      confirmModal.submitting = false;
    }
  };
  confirmModal.open = true;
}

// ─── Actions: My Projects & Collaborators ─────────────────────
function promptRemoveMember(projectId, member) {
  confirmModal.title = t('access.modals.removeConfirmTitle');
  confirmModal.desc = `${member.user_name || member.user_email}: ${t(
    'access.modals.removeConfirmDesc'
  )}`;
  confirmModal.confirmText = t('access.actions.removeMember');
  confirmModal.isDanger = true;
  confirmModal.onConfirm = async () => {
    confirmModal.submitting = true;
    try {
      await removeMember(projectId, member.user_id);
      showToast(t('access.messages.memberRemoved'), 'success');
      closeConfirmModal();
      projectMembers.value[projectId] = await listMembers(projectId);
    } catch (err) {
      showToast(err.response?.data?.message || t('common.error'), 'error');
    } finally {
      confirmModal.submitting = false;
    }
  };
  confirmModal.open = true;
}

function openEditPermissionsModal(project, member) {
  editPermsModal.projectId = project.id;
  editPermsModal.member = member;
  editPermsModal.permissions = [...(member.permissions || [])];
  editPermsModal.open = true;
}

function closeEditPermsModal() {
  editPermsModal.open = false;
  editPermsModal.member = null;
  editPermsModal.permissions = [];
}

function selectAllEditPerms() {
  editPermsModal.permissions = [...ALL_PERMISSIONS];
}

function clearAllEditPerms() {
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
      editPermsModal.projectId,
      editPermsModal.member.user_id,
      editPermsModal.permissions
    );
    showToast(t('access.messages.permissionsUpdated'), 'success');
    closeEditPermsModal();
    projectMembers.value[editPermsModal.projectId] = await listMembers(editPermsModal.projectId);
  } catch (err) {
    showToast(err.response?.data?.message || t('common.error'), 'error');
  } finally {
    editPermsModal.submitting = false;
  }
}

// ─── Actions: Outgoing ────────────────────────────────────────
async function handleCancelInvitation(invitationId) {
  actionInProgress.value = invitationId;
  try {
    await cancelInvitation(invitationId);
    showToast(t('access.messages.inviteCanceled'), 'success');
    await loadOutgoing();
  } catch (err) {
    showToast(err.response?.data?.message || t('common.error'), 'error');
  } finally {
    actionInProgress.value = null;
  }
}

// ─── Actions: Blacklist ───────────────────────────────────────
async function handleUnblock(userId) {
  actionInProgress.value = userId;
  try {
    await unblockUser(userId);
    showToast(t('access.messages.userUnblocked'), 'success');
    await loadBlacklist();
  } catch (err) {
    showToast(err.response?.data?.message || t('common.error'), 'error');
  } finally {
    actionInProgress.value = null;
  }
}

// ─── Invite Developer Modal ───────────────────────────────────
function openInviteModal(defaultProjectId = null) {
  inviteModal.projectId = defaultProjectId || (myProjects.value[0]?.id ?? null);
  inviteModal.query = '';
  inviteModal.selectedUser = null;
  inviteModal.searchResults = [];
  inviteModal.permissions = [...ALL_PERMISSIONS];
  inviteModal.open = true;
}

function closeInviteModal() {
  inviteModal.open = false;
}

function selectAllPerms() {
  inviteModal.permissions = [...ALL_PERMISSIONS];
}

function clearAllPerms() {
  inviteModal.permissions = [];
}

let searchTimer = null;
function onSearchUserInput() {
  clearTimeout(searchTimer);
  const q = inviteModal.query.trim();
  if (!q) {
    inviteModal.searchResults = [];
    return;
  }
  inviteModal.searching = true;
  searchTimer = setTimeout(async () => {
    try {
      const res = await searchUsers({ query: q, limit: 10 });
      inviteModal.searchResults = res.users || [];
    } catch {
      inviteModal.searchResults = [];
    } finally {
      inviteModal.searching = false;
    }
  }, 300);
}

function selectUserToInvite(user) {
  inviteModal.selectedUser = user;
  inviteModal.searchResults = [];
  inviteModal.query = '';
}

function clearSelectedUser() {
  inviteModal.selectedUser = null;
}

async function submitInvitation() {
  if (!inviteModal.projectId) {
    showToast(t('access.messages.selectProjectWarning'), 'warning');
    return;
  }
  const targetEmail = inviteModal.selectedUser?.email || inviteModal.query.trim();
  const targetId = inviteModal.selectedUser?.id || '';
  if (!targetEmail && !targetId) {
    showToast(t('access.messages.selectUserWarning'), 'warning');
    return;
  }
  if (inviteModal.permissions.length === 0) {
    showToast(t('access.messages.selectPermsWarning'), 'warning');
    return;
  }

  inviteModal.submitting = true;
  try {
    await sendInvitation(inviteModal.projectId, {
      invitee_id: targetId,
      invitee_email: targetEmail,
      permissions: inviteModal.permissions,
    });
    showToast(t('access.messages.inviteSent'), 'success');
    closeInviteModal();
    await loadOutgoing();
  } catch (err) {
    const msg = err.response?.data?.message || err.message || '';
    if (msg.includes('blocked') || msg.includes('user is blocked')) {
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

// ─── Modal Helpers ────────────────────────────────────────────
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
.projects-page-container {
  width: 100%;
  max-width: 100%;
  min-height: calc(100vh - 60px);
  display: flex;
  flex-direction: column;
  padding: 24px 32px;
  box-sizing: border-box;
}

.main-content-wrap {
  width: 100%;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
}

.page-title {
  margin: 0 0 6px 0;
  font-size: 26px;
  font-weight: 700;
  color: var(--text-main, #f0f6fc);
}

.page-subtitle {
  margin: 0;
  font-size: 14px;
  color: var(--text-muted, #8b949e);
}

/* Tabs */
.tabs-bar {
  display: flex;
  gap: 8px;
  border-bottom: 1px solid var(--border, #30363d);
  margin-bottom: 24px;
  overflow-x: auto;
  padding-bottom: 2px;
}

.tab-item {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  border: none;
  background: transparent;
  color: var(--text-muted, #8b949e);
  font-size: 14px;
  font-weight: 600;
  border-bottom: 2px solid transparent;
  cursor: pointer;
  transition: all 0.2s ease;
  white-space: nowrap;
}

.tab-item:hover {
  color: var(--text-main, #f0f6fc);
}

.tab-item.active {
  color: var(--primary, #58a6ff);
  border-bottom-color: var(--primary, #58a6ff);
}

.tab-badge {
  padding: 2px 7px;
  font-size: 11px;
  font-weight: 700;
  border-radius: 12px;
  background: var(--bg-tertiary, #21262d);
  color: var(--text-muted, #8b949e);
}
.incoming-badge {
  background: var(--primary, #58a6ff);
  color: #fff;
}
.text-danger-badge {
  background: var(--danger, #f85149);
  color: #fff;
}

/* Common Layouts */
.state-container {
  padding: 60px 20px;
  text-align: center;
  color: var(--text-tertiary, #8b949e);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.empty-card {
  background: var(--bg-card, #161b22);
  border: 1px dashed var(--border, #30363d);
  border-radius: var(--radius-md, 8px);
  padding: 48px 24px;
}
.empty-card h3 {
  margin: 0;
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
  margin-bottom: 6px;
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
  background: transparent;
  white-space: nowrap;
}

.yandex-games-table th.col-game { width: 35%; padding-left: 8px; }
.yandex-games-table th.col-user { width: 30%; }
.yandex-games-table th.col-date { width: 20%; }
.yandex-games-table th.col-status { width: 15%; }
.yandex-games-table th.col-perms { width: 30%; }
.yandex-games-table th.col-actions { width: 15%; text-align: right; padding-right: 12px; }

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
.table-row td.col-game { padding-left: 8px; }
.table-row td.col-actions { padding-right: 12px; text-align: right; }

.game-cell {
  display: flex;
  align-items: center;
  gap: 14px;
}

.game-icon-box {
  width: 40px;
  height: 40px;
  border-radius: var(--radius-sm, 8px);
  background: var(--bg-tertiary, #21262d);
  border: 1px solid var(--border, #30363d);
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  flex-shrink: 0;
}
.game-icon-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.game-icon-mock {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-tertiary, #21262d);
  color: var(--text-tertiary, #8b949e);
  font-size: 11px;
  font-weight: 600;
}

.game-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.game-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
  line-height: 1.3;
}
.game-type-label {
  font-size: 13px;
  font-weight: 400;
  color: var(--text-tertiary, #8b949e);
  line-height: 1.3;
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
  align-items: center;
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
  border-color: rgba(245, 176, 39, 0.35);
  color: #f5b027;
  border: 1px solid;
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
.text-warning-hover:hover { color: var(--warning, #d29922) !important; }
.text-success-hover:hover { color: #2ecc71 !important; }

/* My Projects Block */
.my-projects-list {
  display: flex;
  flex-direction: column;
  gap: 24px;
}
.project-block {
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-md, 8px);
  overflow: hidden;
}
.project-block-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  background: var(--bg-secondary, #0d1117);
  border-bottom: 1px solid var(--border, #30363d);
}
.empty-members {
  padding: 16px 20px;
  font-size: 13px;
}

/* Modals */
.modal-overlay {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0,0,0,0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 20px;
}
.modal-card {
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: 8px;
  width: 100%;
  max-width: 460px;
  display: flex;
  flex-direction: column;
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

.btn-secondary {
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
.btn-secondary:hover { background: var(--bg-tertiary); }

/* Search User */
.search-user-wrapper { position: relative; }
.search-input-box { position: relative; }
.search-icon { position: absolute; left: 10px; top: 10px; color: var(--text-muted); }
.search-loader { position: absolute; right: 10px; top: 10px; }
.has-search-icon { padding-left: 32px; width: 100%; box-sizing: border-box; }
.user-dropdown {
  position: absolute; top: 100%; left: 0; right: 0;
  background: var(--bg-card); border: 1px solid var(--border);
  border-radius: 6px; max-height: 200px; overflow-y: auto; z-index: 10;
  margin-top: 4px; box-shadow: 0 4px 12px rgba(0,0,0,0.5);
}
.user-dropdown-item {
  padding: 8px 12px; cursor: pointer; display: flex; flex-direction: column;
}
.user-dropdown-item:hover { background: var(--bg-secondary); }
.user-item-name { font-size: 13px; font-weight: 500; }
.user-item-email { font-size: 12px; color: var(--text-muted); }
.selected-user-pill {
  display: flex; justify-content: space-between; align-items: center;
  padding: 8px 12px; background: var(--bg-secondary); border: 1px solid var(--border);
  border-radius: 6px; margin-top: 8px;
}
.pill-info { display: flex; flex-direction: column; }
.pill-info strong { font-size: 13px; }
.pill-email { font-size: 12px; color: var(--text-muted); }

/* Permissions Checkboxes */
.permissions-header-row { display: flex; justify-content: space-between; align-items: center; }
.quick-perms-actions { display: flex; gap: 4px; align-items: center; }
.btn-text { background: none; border: none; color: var(--primary); font-size: 12px; cursor: pointer; padding: 0;}
.btn-text:hover { text-decoration: underline; }
.text-separator { color: var(--text-muted); font-size: 12px; }

.permissions-checkbox-list {
  display: flex; flex-direction: column; gap: 8px; max-height: 200px; overflow-y: auto;
  border: 1px solid var(--border); padding: 12px; border-radius: 6px; background: var(--bg-secondary);
}
.perm-checkbox-item {
  display: flex; align-items: center; gap: 8px; cursor: pointer;
}
.perm-title { font-size: 13px; color: var(--text-main); }
.checkbox-input { cursor: pointer; }
.text-muted { color: var(--text-muted) !important; }

.spinner-sm {
  width: 14px; height: 14px; border: 2px solid rgba(255,255,255,0.3);
  border-top-color: #fff; border-radius: 50%; animation: spin 0.8s linear infinite;
}
.spinner-md {
  width: 24px; height: 24px; border: 2px solid rgba(255,255,255,0.3);
  border-top-color: var(--primary); border-radius: 50%; animation: spin 0.8s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }
</style>
