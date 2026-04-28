<template>
  <div class="page-container">
    <h1>Управление ролями</h1>
    <div class="users-table">
      <table>
        <thead><tr><th>Пользователь</th><th>Email</th><th>Текущая роль</th><th>Действие</th></tr></thead>
        <tbody>
        <tr v-for="u in users" :key="u.id" class="user-row" @click="openUser(u)">
          <td><div class="user-cell"><div class="avatar" :class="u.role === 'Модератор' ? 'avatar-purple' : 'avatar-blue'">{{ u.name[0] }}</div><strong>{{ u.name }}</strong></div></td>
          <td class="text-muted">{{ u.email }}</td>
          <td><span class="role-badge" :class="u.role === 'Модератор' ? 'bg-purple' : 'bg-blue'">{{ u.role }}</span></td>
          <td @click.stop><select v-model="u.role" @change="changeRole(u)" class="role-select"><option value="Разработчик">Разработчик</option><option value="Модератор">Модератор</option><option value="Администратор">Администратор</option></select></td>
        </tr>
        </tbody>
      </table>
    </div>
    <transition name="modal-fade">
      <div v-if="selectedUser" class="modal-overlay" @click.self="selectedUser = null">
        <div class="modal-card">
          <button class="modal-close" @click="selectedUser = null">✕</button>
          <div class="modal-avatar" :class="selectedUser.role === 'Модератор' ? 'avatar-purple' : 'avatar-blue'">{{ selectedUser.name[0] }}</div>
          <h2 class="modal-name">{{ selectedUser.name }}</h2>
          <span class="role-badge" :class="selectedUser.role === 'Модератор' ? 'bg-purple' : 'bg-blue'">{{ selectedUser.role }}</span>
          <div class="modal-info">
            <div class="info-row"><span class="info-label">Email</span><span class="info-value">{{ selectedUser.email }}</span></div>
            <div class="info-row"><span class="info-label">ID</span><span class="info-value">#{{ selectedUser.id }}</span></div>
            <div class="info-row"><span class="info-label">Зарегистрирован</span><span class="info-value">{{ selectedUser.registered }}</span></div>
            <div class="info-row"><span class="info-label">Проектов</span><span class="info-value">{{ selectedUser.projects }}</span></div>
            <div class="info-row"><span class="info-label">Тикетов закрыто</span><span class="info-value">{{ selectedUser.ticketsClosed }}</span></div>
          </div>
          <div class="modal-actions"><div class="action-label">Сменить роль</div><div class="role-buttons"><button v-for="role in ['Разработчик', 'Модератор', 'Администратор']" :key="role" class="role-btn" :class="{ active: selectedUser.role === role }" @click="changeRoleFromModal(selectedUser, role)">{{ role }}</button></div></div>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { searchUsers } from '../api/sso'
import { showToast } from '../store'

const users = ref([])
const loading = ref(false)
const selectedUser = ref(null)

const ROLE_MAP = {
  0: 'Пользователь',
  1: 'Разработчик',
  2: 'Модератор',
  3: 'Администратор',
  'USER_ROLE_UNSPECIFIED': 'Пользователь',
  'USER_ROLE_DEVELOPER': 'Разработчик',
  'USER_ROLE_MODERATOR': 'Модератор',
  'USER_ROLE_ADMIN': 'Администратор',
}

const fetchUsers = async () => {
  loading.value = true
  try {
    const res = await searchUsers({ query: '' })
    users.value = (res.users || []).map(u => ({
      id: u.id,
      name: u.display_name || u.email,
      email: u.email,
      role: ROLE_MAP[u.role] || 'Пользователь',
      registered: u.created_at ? new Date(u.created_at).toLocaleDateString() : '—',
      projects: 0,
      ticketsClosed: 0
    }))
  } catch (err) {
    console.error('Failed to fetch users:', err)
    showToast('Ошибка загрузки пользователей', 'danger')
  } finally {
    loading.value = false
  }
}

onMounted(fetchUsers)

const openUser = (u) => { selectedUser.value = { ...u } }
const changeRole = (u) => { showToast(`Роль ${u.name} изменена на "${u.role}"`, 'success') }
const changeRoleFromModal = (u, role) => {
  if (u.role === role) return
  u.role = role
  const original = users.value.find(x => x.id === u.id)
  if (original) original.role = role
  showToast(`Роль ${u.name} изменена на "${role}"`, 'success')
}
</script>

<style scoped>
.page-container { padding: 32px 40px; max-width: 1000px; margin: 0 auto; }
h1 { margin-bottom: 24px; color: var(--text-main); }
.users-table { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); overflow: hidden; }
table { width: 100%; border-collapse: collapse; }
th, td { text-align: left; padding: 16px; border-bottom: 1px solid var(--border); }
th { background: var(--bg-app); font-weight: 600; color: var(--text-muted); }
.user-row { cursor: pointer; }
.user-row:hover td { background: var(--bg-hover); }
.user-cell { display: flex; align-items: center; gap: 10px; }
.avatar { width: 34px; height: 34px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 0.85rem; }
/* Тёмные аналоги светлых бейджей */
.avatar-blue { background: rgba(59, 130, 246, 0.2); color: #60A5FA; }
.avatar-purple { background: rgba(168, 85, 247, 0.2); color: #C084FC; }
.text-muted { color: var(--text-muted); }
.role-badge { padding: 4px 10px; border-radius: 20px; font-size: 0.78rem; font-weight: 700; }
.bg-blue { background: rgba(59, 130, 246, 0.2); color: #60A5FA; }
.bg-purple { background: rgba(168, 85, 247, 0.2); color: #C084FC; }
.role-select { padding: 6px 12px; border: 1px solid var(--border); border-radius: var(--radius-md); background: var(--bg-card); color: var(--text-main); font-weight: 500; cursor: pointer; }
.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.6); z-index: 200; display: flex; align-items: center; justify-content: center; backdrop-filter: blur(3px); }
.modal-card { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 32px 28px; width: 340px; position: relative; display: flex; flex-direction: column; align-items: center; gap: 12px; }
.modal-close { position: absolute; top: 14px; right: 14px; background: none; border: none; font-size: 1rem; color: var(--text-muted); cursor: pointer; }
.modal-avatar { width: 64px; height: 64px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-weight: 800; font-size: 1.4rem; }
.modal-name { margin: 0; font-size: 1.2rem; color: var(--text-main); }
.modal-info { width: 100%; border: 1px solid var(--border); border-radius: var(--radius-md); overflow: hidden; }
.info-row { display: flex; justify-content: space-between; padding: 10px 14px; border-bottom: 1px solid var(--border); font-size: 0.88rem; }
.info-row:last-child { border-bottom: none; }
.info-label { color: var(--text-muted); }
.info-value { font-weight: 600; color: var(--text-main); }
.modal-actions { width: 100%; margin-top: 4px; }
.action-label { font-size: 0.8rem; color: var(--text-muted); font-weight: 600; margin-bottom: 8px; }
.role-buttons { display: flex; gap: 8px; }
.role-btn { flex: 1; padding: 9px; border-radius: var(--radius-md); border: 1px solid var(--border); background: var(--bg-app); color: var(--text-muted); font-weight: 600; cursor: pointer; font-size: 0.88rem; }
.role-btn.active { background: rgba(59, 130, 246, 0.2); color: #60A5FA; border-color: var(--primary); }
.modal-fade-enter-active, .modal-fade-leave-active { transition: opacity 0.2s, transform 0.2s; }
.modal-fade-enter-from, .modal-fade-leave-to { opacity: 0; transform: scale(0.95); }
</style>