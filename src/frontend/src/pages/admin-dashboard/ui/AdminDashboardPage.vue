<template>
  <div class="page-container">
    <div class="header-row">
      <h1>{{ t('admin.title') }}</h1>
    </div>

    <!-- Tabs -->
    <div class="tabs-nav">
      <button
        class="tab-btn"
        :class="{ active: activeTab === 'users' }"
        @click="activeTab = 'users'"
      >
        {{ t('admin.usersTab') }}
      </button>
      <button
        class="tab-btn"
        :class="{ active: activeTab === 'moderators' }"
        @click="activeTab = 'moderators'"
      >
        {{ t('admin.moderatorsTab') }}
      </button>
    </div>

    <!-- Users Tab -->
    <div v-if="activeTab === 'users'" class="tab-content">
      <div class="card">
        <div class="card-header">
          <h2>{{ t('admin.usersTab') }}</h2>
          <input
            v-model="searchQuery"
            type="text"
            :placeholder="t('projects.searchPlaceholder')"
            class="search-input"
            @input="debouncedSearch"
          />
        </div>
        <div v-if="loading" class="state-container">
          <div class="spinner-md"></div>
          <p>{{ t('common.loading') }}</p>
        </div>
        <div v-else-if="users.length === 0" class="empty-state">
          {{ t('common.empty') }}
        </div>
        <div v-else class="table-container">
          <table>
            <thead>
              <tr>
                <th>{{ t('common.name') }}</th>
                <th>{{ t('auth.email') }}</th>
                <th>{{ t('profile.role') }}</th>
                <th>{{ t('common.status') }}</th>
                <th>{{ t('common.created') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="u in users" :key="u.id">
                <td>
                  <strong>{{ u.display_name }}</strong>
                </td>
                <td class="email-cell">{{ u.email }}</td>
                <td>
                  <span class="badge" :class="roleClass(u.role)">
                    {{ roleLabel(u.role) }}
                  </span>
                </td>
                <td>
                  <span class="badge" :class="statusBadgeClass(u.status)">
                    <span class="status-dot"></span>
                    {{ statusLabel(u.status) }}
                  </span>
                </td>
                <td>{{ formatDate(u.created_at) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- Moderators Tab -->
    <div v-if="activeTab === 'moderators'" class="tab-content">
      <!-- Create Moderator Form -->
      <CreateModeratorForm @created="loadUsers" />

      <!-- Moderators List -->
      <div class="card">
        <h2>Список модераторов</h2>
        <div v-if="moderators.length === 0" class="empty-state">Нет активных модераторов</div>
        <div v-else class="table-container">
          <table>
            <thead>
              <tr>
                <th>Имя</th>
                <th>Email</th>
                <th>Статус</th>
                <th>Действия</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="mod in moderators" :key="mod.id">
                <td>
                  <strong>{{ mod.display_name }}</strong>
                </td>
                <td class="email-cell">{{ mod.email }}</td>
                <td>
                  <span class="badge" :class="statusBadgeClass(mod.status)">
                    <span class="status-dot"></span>
                    {{ statusLabel(mod.status) }}
                  </span>
                </td>
                <td>
                  <button
                    class="btn btn-danger"
                    style="padding: 6px 14px; font-size: 0.8rem"
                    :disabled="deleting"
                    @click="confirmDelete(mod)"
                  >
                    {{ deleting ? 'Удаление...' : 'Удалить' }}
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- Delete Confirmation Modal -->
    <DeleteModeratorModal
      v-if="deleteTarget"
      :target="deleteTarget"
      :deleting="deleting"
      @confirm="handleDelete"
      @cancel="deleteTarget = null"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { searchUsers, deleteUser } from '@/entities/user';
import { CreateModeratorForm, DeleteModeratorModal } from '@/features/manage-moderators';
import { formatDate, showToast } from '@/shared/lib';

const { t } = useI18n();

const activeTab = ref('users');
const loading = ref(false);
const deleting = ref(false);
const deleteTarget = ref(null);
const searchQuery = ref('');
const allUsers = ref([]);
let searchTimeout = null;

const users = computed(() => {
  if (!searchQuery.value) return allUsers.value;
  const q = searchQuery.value.toLowerCase();
  return allUsers.value.filter(
    (u) =>
      (u.display_name && u.display_name.toLowerCase().includes(q)) ||
      (u.email && u.email.toLowerCase().includes(q))
  );
});

const moderators = computed(() =>
  allUsers.value.filter((u) => u.role === 'USER_ROLE_MODERATOR' || u.role === 'moderator')
);

function debouncedSearch() {
  clearTimeout(searchTimeout);
  searchTimeout = setTimeout(() => loadUsers(), 300);
}

async function loadUsers() {
  loading.value = true;
  try {
    const res = await searchUsers({ query: searchQuery.value, limit: 100 });
    allUsers.value = res.users || [];
  } catch {
    allUsers.value = [];
  } finally {
    loading.value = false;
  }
}

function roleClass(role) {
  switch (role) {
    case 'USER_ROLE_ADMIN':
    case 'admin':
      return 'badge-danger';
    case 'USER_ROLE_MODERATOR':
    case 'moderator':
      return 'badge-warning';
    default:
      return 'badge-success';
  }
}

function roleLabel(role) {
  switch (role) {
    case 'USER_ROLE_ADMIN':
    case 'admin':
    case 3:
      return t('profile.adminRole');
    case 'USER_ROLE_MODERATOR':
    case 'moderator':
    case 2:
      return t('profile.moderatorRole');
    default:
      return t('profile.developerRole');
  }
}

function statusBadgeClass(status) {
  switch (status) {
    case 'USER_STATUS_ACTIVE':
    case 'active':
      return 'badge-success';
    default:
      return 'badge-danger';
  }
}

function statusLabel(status) {
  switch (status) {
    case 'USER_STATUS_ACTIVE':
    case 'active':
      return 'Активен';
    case 'USER_STATUS_BANNED':
    case 'USER_STATUS_SUSPENDED':
    case 'banned':
    case 'suspended':
      return 'Заблокирован';
    default:
      return 'Неизвестно';
  }
}

function confirmDelete(mod) {
  deleteTarget.value = mod;
}

async function handleDelete() {
  if (!deleteTarget.value) return;
  deleting.value = true;
  try {
    await deleteUser(deleteTarget.value.id);
    showToast(`Модератор "${deleteTarget.value.display_name}" удалён`, 'success');
    deleteTarget.value = null;
    await loadUsers();
  } catch (err) {
    showToast(err.response?.data?.message || 'Не удалось удалить модератора', 'error');
  } finally {
    deleting.value = false;
  }
}

onMounted(() => {
  loadUsers();
});
</script>

<style scoped>
.tabs-nav {
  display: flex;
  gap: 4px;
  margin-bottom: 24px;
  border-bottom: 1px solid var(--border);
}

.tab-btn {
  padding: 10px 20px;
  border: none;
  background: transparent;
  color: var(--text-muted);
  font-weight: 600;
  font-size: 0.9rem;
  cursor: pointer;
  border-bottom: 2px solid transparent;
  transition: all 0.2s;
}

.tab-btn:hover {
  color: var(--text-main);
}

.tab-btn.active {
  color: var(--primary);
  border-bottom-color: var(--primary);
}

.tab-content {
  animation: fadeIn 0.2s ease;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(4px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.card {
  margin-bottom: 24px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  gap: 16px;
  flex-wrap: wrap;
}

.card-header h2 {
  margin: 0;
  font-size: 1.1rem;
}

.search-input {
  padding: 8px 14px;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  background: var(--bg-input);
  color: var(--text-main);
  font-size: 0.9rem;
  width: 280px;
  font-family: inherit;
  transition:
    border-color 0.2s,
    box-shadow 0.2s;
  outline: none;
}

.search-input:focus {
  border-color: var(--primary);
  box-shadow: 0 0 0 3px var(--primary-light);
}

.email-cell {
  font-family: monospace;
  font-size: 0.85rem;
  color: var(--text-muted);
}

.empty-state {
  text-align: center;
  padding: 32px;
  color: var(--text-muted);
}

@media (max-width: 768px) {
  .card-header {
    flex-direction: column;
    gap: 12px;
    align-items: flex-start;
  }

  .search-input {
    width: 100%;
  }
}
</style>
