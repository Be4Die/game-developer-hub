<template>
    <div class="page-container">
        <h1>Администрирование</h1>

        <!-- Tabs -->
        <div class="tabs">
            <button
                class="tab-btn"
                :class="{ active: activeTab === 'users' }"
                @click="activeTab = 'users'"
            >
                Пользователи
            </button>
            <button
                class="tab-btn"
                :class="{ active: activeTab === 'moderators' }"
                @click="activeTab = 'moderators'"
            >
                Модераторы
            </button>
        </div>

        <!-- Users Tab -->
        <div v-if="activeTab === 'users'" class="tab-content">
            <div class="card">
                <div class="card-header">
                    <h2>Все пользователи</h2>
                    <input
                        v-model="searchQuery"
                        type="text"
                        placeholder="Поиск по имени или email..."
                        class="search-input"
                        @input="debouncedSearch"
                    />
                </div>
                <div v-if="loading" class="empty-state">Загрузка...</div>
                <div v-else-if="users.length === 0" class="empty-state">
                    Пользователи не найдены
                </div>
                <table v-else class="data-table">
                    <thead>
                        <tr>
                            <th>Имя</th>
                            <th>Email</th>
                            <th>Роль</th>
                            <th>Статус</th>
                            <th>Дата регистрации</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="u in users" :key="u.id" class="table-row">
                            <td>
                                <strong>{{ u.display_name }}</strong>
                            </td>
                            <td class="email-cell">{{ u.email }}</td>
                            <td>
                                <span
                                    class="role-badge"
                                    :class="roleClass(u.role)"
                                >
                                    {{ roleLabel(u.role) }}
                                </span>
                            </td>
                            <td>
                                <span
                                    class="status-badge"
                                    :class="statusClass(u.status)"
                                >
                                    {{ statusLabel(u.status) }}
                                </span>
                            </td>
                            <td>{{ formatDate(u.created_at) }}</td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>

        <!-- Moderators Tab -->
        <div v-if="activeTab === 'moderators'" class="tab-content">
            <!-- Create Moderator Form -->
            <div class="card create-form">
                <h2>Создать модератора</h2>
                <form @submit.prevent="handleCreate" class="form-grid">
                    <div class="form-group">
                        <label for="login">Логин</label>
                        <div class="email-input-group">
                            <input
                                id="login"
                                v-model="form.login"
                                type="text"
                                placeholder="username"
                                required
                            />
                            <span class="email-domain">@welwise.com</span>
                        </div>
                        <span class="form-hint"
                            >Email будет сформирован автоматически</span
                        >
                    </div>
                    <div class="form-group">
                        <label for="display_name">Имя</label>
                        <input
                            id="display_name"
                            v-model="form.display_name"
                            type="text"
                            placeholder="Имя модератора"
                            required
                        />
                    </div>
                    <div class="form-group">
                        <label for="password">Пароль</label>
                        <input
                            id="password"
                            v-model="form.password"
                            type="password"
                            placeholder="Минимум 6 символов"
                            required
                            minlength="6"
                        />
                    </div>
                    <div class="form-group form-actions">
                        <button
                            type="submit"
                            class="btn-primary"
                            :disabled="loading"
                        >
                            {{ loading ? "Создание..." : "Создать" }}
                        </button>
                    </div>
                </form>
                <div v-if="createdEmail" class="created-info">
                    <strong>Email для входа:</strong> {{ createdEmail }}
                </div>
            </div>

            <!-- Moderators List -->
            <div class="card">
                <h2>Список модераторов</h2>
                <div v-if="moderators.length === 0" class="empty-state">
                    Нет активных модераторов
                </div>
                <table v-else class="data-table">
                    <thead>
                        <tr>
                            <th>Имя</th>
                            <th>Email</th>
                            <th>Статус</th>
                            <th>Действия</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr
                            v-for="mod in moderators"
                            :key="mod.id"
                            class="table-row"
                        >
                            <td>
                                <strong>{{ mod.display_name }}</strong>
                            </td>
                            <td class="email-cell">{{ mod.email }}</td>
                            <td>
                                <span
                                    class="status-badge"
                                    :class="statusClass(mod.status)"
                                >
                                    {{ statusLabel(mod.status) }}
                                </span>
                            </td>
                            <td>
                                <button
                                    class="btn-danger btn-sm"
                                    @click="confirmDelete(mod)"
                                    :disabled="deleting"
                                >
                                    {{ deleting ? "Удаление..." : "Удалить" }}
                                </button>
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>

        <!-- Delete Confirmation Modal -->
        <transition name="modal-fade">
            <div
                v-if="deleteTarget"
                class="modal-overlay"
                @click.self="deleteTarget = null"
            >
                <div class="modal-card">
                    <button class="modal-close" @click="deleteTarget = null">
                        &#x2715;
                    </button>
                    <h3>Подтверждение удаления</h3>
                    <p>
                        Вы уверены, что хотите удалить модератора
                        <strong>{{ deleteTarget.display_name }}</strong
                        >?
                    </p>
                    <p class="warning-text">Это действие нельзя отменить.</p>
                    <div class="modal-actions">
                        <button
                            class="btn-secondary"
                            @click="deleteTarget = null"
                        >
                            Отмена
                        </button>
                        <button
                            class="btn-danger"
                            @click="handleDelete"
                            :disabled="deleting"
                        >
                            {{ deleting ? "Удаление..." : "Удалить" }}
                        </button>
                    </div>
                </div>
            </div>
        </transition>
    </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from "vue";
import { searchUsers, createModerator, deleteUser } from "../api/sso";
import { showToast } from "../store";

const activeTab = ref("users");
const loading = ref(false);
const deleting = ref(false);
const createdEmail = ref("");
const deleteTarget = ref(null);
const searchQuery = ref("");
const allUsers = ref([]);
let searchTimeout = null;

const users = computed(() => {
    if (!searchQuery.value) return allUsers.value;
    const q = searchQuery.value.toLowerCase();
    return allUsers.value.filter(
        (u) =>
            u.display_name.toLowerCase().includes(q) ||
            u.email.toLowerCase().includes(q),
    );
});

const moderators = computed(() =>
    allUsers.value.filter(
        (u) => u.role === "USER_ROLE_MODERATOR" || u.role === "moderator",
    ),
);

const form = reactive({
    login: "",
    display_name: "",
    password: "",
});

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
        case "USER_ROLE_ADMIN":
        case "admin":
            return "role-admin";
        case "USER_ROLE_MODERATOR":
        case "moderator":
            return "role-moderator";
        default:
            return "role-developer";
    }
}

function roleLabel(role) {
    switch (role) {
        case "USER_ROLE_ADMIN":
        case "admin":
            return "Администратор";
        case "USER_ROLE_MODERATOR":
        case "moderator":
            return "Модератор";
        default:
            return "Разработчик";
    }
}

function statusClass(status) {
    switch (status) {
        case "USER_STATUS_ACTIVE":
        case "active":
            return "status-active";
        default:
            return "status-inactive";
    }
}

function statusLabel(status) {
    switch (status) {
        case "USER_STATUS_ACTIVE":
        case "active":
            return "Активен";
        case "USER_STATUS_BANNED":
        case "USER_STATUS_SUSPENDED":
        case "banned":
        case "suspended":
            return "Заблокирован";
        default:
            return "Неизвестно";
    }
}

function formatDate(dateStr) {
    if (!dateStr) return "";
    return new Date(dateStr).toLocaleDateString("ru-RU");
}

async function handleCreate() {
    loading.value = true;
    try {
        const res = await createModerator({
            login: form.login,
            password: form.password,
            display_name: form.display_name,
        });
        createdEmail.value = res.user.email;
        showToast(
            `Модератор "${res.user.display_name}" успешно создан`,
            "success",
        );
        form.login = "";
        form.display_name = "";
        form.password = "";
        await loadUsers();
    } catch (err) {
        showToast(
            err.response?.data?.message || "Не удалось создать модератора",
            "error",
        );
    } finally {
        loading.value = false;
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
        showToast(
            `Модератор "${deleteTarget.value.display_name}" удалён`,
            "success",
        );
        deleteTarget.value = null;
        await loadUsers();
    } catch (err) {
        showToast(
            err.response?.data?.message || "Не удалось удалить модератора",
            "error",
        );
    } finally {
        deleting.value = false;
    }
}

onMounted(() => {
    loadUsers();
});
</script>

<style scoped>
.page-container {
    padding: 32px 40px;
    max-width: 1100px;
    margin: 0 auto;
}

h1 {
    margin-bottom: 24px;
    color: var(--text-main);
}

h2 {
    margin: 0 0 20px;
    font-size: 1.1rem;
    color: var(--text-main);
}

/* Tabs */
.tabs {
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

/* Card */
.card {
    background: var(--bg-card);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    padding: 24px;
    margin-bottom: 24px;
}

.card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
}

.card-header h2 {
    margin: 0;
}

/* Search */
.search-input {
    padding: 8px 14px;
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    background: var(--bg-input);
    color: var(--text-main);
    font-size: 0.9rem;
    width: 280px;
}

.search-input:focus {
    outline: none;
    border-color: var(--primary);
}

/* Table */
.data-table {
    width: 100%;
    border-collapse: collapse;
}

.data-table th,
.data-table td {
    text-align: left;
    padding: 12px 16px;
    border-bottom: 1px solid var(--border);
}

.data-table th {
    background: var(--bg-app);
    font-weight: 600;
    color: var(--text-muted);
    font-size: 0.85rem;
}

.email-cell {
    font-family: monospace;
    font-size: 0.85rem;
    color: var(--text-muted);
}

/* Badges */
.role-badge {
    display: inline-block;
    padding: 4px 10px;
    border-radius: 20px;
    font-size: 0.78rem;
    font-weight: 700;
}

.role-admin {
    background: rgba(239, 68, 68, 0.15);
    color: #ef4444;
}

.role-moderator {
    background: rgba(168, 85, 247, 0.15);
    color: #a855f7;
}

.role-developer {
    background: rgba(59, 130, 246, 0.15);
    color: #3b82f6;
}

.status-badge {
    display: inline-block;
    padding: 4px 10px;
    border-radius: 20px;
    font-size: 0.78rem;
    font-weight: 700;
}

.status-active {
    background: rgba(34, 197, 94, 0.15);
    color: #22c55e;
}

.status-inactive {
    background: rgba(239, 68, 68, 0.15);
    color: #ef4444;
}

/* Form */
.create-form {
    margin-bottom: 24px;
}

.form-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 16px;
}

.form-group {
    display: flex;
    flex-direction: column;
}

.form-group label {
    margin-bottom: 6px;
    font-size: 0.85rem;
    font-weight: 600;
    color: var(--text-muted);
}

.form-group input {
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

.email-input-group {
    display: flex;
    align-items: center;
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    background: var(--bg-input);
    overflow: hidden;
    transition: border-color 0.2s;
}

.email-input-group:focus-within {
    border-color: var(--primary);
}

.email-input-group input {
    border: none;
    flex: 1;
    min-width: 0;
}

.email-input-group input:focus {
    border: none;
}

.email-domain {
    padding: 10px 12px;
    color: var(--text-muted);
    font-size: 0.95rem;
    background: var(--bg-app);
    border-left: 1px solid var(--border);
    white-space: nowrap;
    user-select: none;
}

.form-hint {
    margin-top: 4px;
    font-size: 0.75rem;
    color: var(--text-muted);
}

.form-actions {
    justify-content: flex-end;
    padding-top: 8px;
}

.created-info {
    margin-top: 16px;
    padding: 12px;
    background: var(--success-light);
    color: var(--success);
    border-radius: var(--radius-sm);
    font-size: 0.9rem;
    text-align: center;
}

/* Empty State */
.empty-state {
    text-align: center;
    padding: 32px;
    color: var(--text-muted);
}

/* Buttons */
.btn-primary {
    padding: 10px 24px;
    border: none;
    border-radius: var(--radius-md);
    background: var(--primary);
    color: white;
    font-weight: 600;
    font-size: 0.9rem;
    cursor: pointer;
    transition: opacity 0.2s;
}

.btn-primary:disabled {
    opacity: 0.6;
    cursor: not-allowed;
}

.btn-danger {
    padding: 6px 14px;
    border: none;
    border-radius: var(--radius-sm);
    background: #ef4444;
    color: white;
    font-weight: 600;
    font-size: 0.8rem;
    cursor: pointer;
    transition: opacity 0.2s;
}

.btn-danger:disabled {
    opacity: 0.6;
    cursor: not-allowed;
}

.btn-sm {
    padding: 6px 14px;
    font-size: 0.8rem;
}

.btn-secondary {
    padding: 10px 24px;
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    background: var(--bg-app);
    color: var(--text-muted);
    font-weight: 600;
    font-size: 0.9rem;
    cursor: pointer;
    transition: background 0.2s;
}

.btn-secondary:hover {
    background: var(--bg-hover);
}

/* Modal */
.modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.6);
    z-index: 200;
    display: flex;
    align-items: center;
    justify-content: center;
    backdrop-filter: blur(3px);
}

.modal-card {
    background: var(--bg-card);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    padding: 32px 28px;
    width: 420px;
    position: relative;
    display: flex;
    flex-direction: column;
    gap: 12px;
}

.modal-close {
    position: absolute;
    top: 14px;
    right: 14px;
    background: none;
    border: none;
    font-size: 1rem;
    color: var(--text-muted);
    cursor: pointer;
}

.modal-card h3 {
    margin: 0;
    font-size: 1.1rem;
    color: var(--text-main);
}

.modal-card p {
    margin: 0;
    color: var(--text-muted);
    font-size: 0.9rem;
    line-height: 1.5;
}

.warning-text {
    color: #ef4444 !important;
    font-weight: 600;
}

.modal-actions {
    display: flex;
    gap: 12px;
    justify-content: flex-end;
    margin-top: 8px;
}

/* Animations */
.toast-fade-enter-active,
.toast-fade-leave-active {
    transition: all 0.3s;
}

.toast-fade-enter-from,
.toast-fade-leave-to {
    opacity: 0;
    transform: translateY(-10px);
}

.modal-fade-enter-active,
.modal-fade-leave-active {
    transition:
        opacity 0.2s,
        transform 0.2s;
}

.modal-fade-enter-from,
.modal-fade-leave-to {
    opacity: 0;
    transform: scale(0.95);
}

@media (max-width: 768px) {
    .form-grid {
        grid-template-columns: 1fr;
    }

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
