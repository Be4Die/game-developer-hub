<template>
    <div class="nodes-page">
        <div class="page-header">
            <h1>Вычислительные ноды</h1>
            <div class="header-actions">
                <select
                    v-model="statusFilter"
                    class="filter-select"
                    @change="fetchNodes"
                >
                    <option value="all">Все статусы</option>
                    <option value="unauthorized">Не авторизованы</option>
                    <option value="online">В сети</option>
                    <option value="offline">Не в сети</option>
                    <option value="maintenance">Обслуживание</option>
                </select>
                <button class="btn-primary" @click="openRegisterModal">
                    <Plus class="icon-sm" /> Подключить ноду
                </button>
            </div>
        </div>

        <!-- Ошибка -->
        <div v-if="error" class="error-banner">
            <AlertCircle class="icon-sm" /> {{ error }}
            <button class="btn-outline btn-sm" @click="fetchNodes">
                Повторить
            </button>
        </div>

        <!-- Таблица нод -->
        <div v-if="loading" class="loading-state">Загрузка...</div>
        <div class="table-wrap" v-else-if="filteredNodes.length">
            <table class="data-table">
                <thead>
                    <tr>
                        <th>ID</th>
                        <th>Адрес</th>
                        <th>Регион</th>
                        <th>Статус</th>
                        <th>CPU</th>
                        <th>Память</th>
                        <th>Диск</th>
                        <th>Агент</th>
                        <th>Heartbeat</th>
                        <th></th>
                    </tr>
                </thead>
                <tbody>
                    <tr
                        v-for="node in filteredNodes"
                        :key="node.id"
                        @click="$router.push(`/nodes/${node.id}`)"
                        class="clickable-row"
                    >
                        <td class="cell-id">{{ node.id }}</td>
                        <td class="cell-addr">
                            <code>{{ node.address }}</code>
                        </td>
                        <td>{{ node.region || "—" }}</td>
                        <td>
                            <StatusBadge :status="node.status" type="node" />
                        </td>
                        <td>
                            {{
                                node.cpu_cores ? node.cpu_cores + " ядер" : "—"
                            }}
                        </td>
                        <td>
                            {{
                                node.total_memory_bytes
                                    ? formatBytes(node.total_memory_bytes)
                                    : "—"
                            }}
                        </td>
                        <td>
                            {{
                                node.total_disk_bytes
                                    ? formatBytes(node.total_disk_bytes)
                                    : "—"
                            }}
                        </td>
                        <td class="cell-muted">
                            {{ node.agent_version || "—" }}
                        </td>
                        <td class="cell-muted">
                            {{ formatTime(node.last_ping_at) }}
                        </td>
                        <td class="cell-actions" @click.stop>
                            <button
                                class="btn-icon"
                                @click="confirmDelete(node)"
                                title="Удалить"
                                :disabled="deletingId === node.id"
                            >
                                <Trash2 class="icon-sm" />
                            </button>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>
        <div v-else class="empty-state">
            Нет нод{{ statusFilter !== "all" ? " с выбранным статусом" : "" }}
        </div>

        <!-- Модал подключения ноды -->
        <div
            v-if="showRegisterForm"
            class="modal-overlay"
            @click.self="showRegisterForm = false"
        >
            <div class="modal card">
                <h3>Подключить ноду</h3>
                <div class="tabs">
                    <button
                        class="tab-btn"
                        :class="{ active: registerTab === 'available' }"
                        @click="registerTab = 'available'"
                    >
                        Доступные ноды
                    </button>
                    <button
                        class="tab-btn"
                        :class="{ active: registerTab === 'manual' }"
                        @click="registerTab = 'manual'"
                    >
                        Ручной ввод
                    </button>
                </div>

                <!-- Доступные ноды (auto-discovery) -->
                <div v-if="registerTab === 'available'" class="form-content">
                    <div class="form-group">
                        <label>Доступная нода *</label>
                        <select
                            v-model="availableForm.node_id"
                            class="form-input"
                        >
                            <option value="" disabled>Выберите ноду</option>
                            <option
                                v-for="n in availableNodes"
                                :key="n.id"
                                :value="n.id"
                            >
                                {{ n.address }} (ID: {{ n.id }})
                            </option>
                        </select>
                        <p v-if="!availableNodes.length" class="hint">
                            Нет доступных нод для подключения. Убедитесь, что
                            нода запущена в режиме auto-discovery.
                        </p>
                    </div>
                    <div class="form-group">
                        <label>Ключ авторизации (API-ключ ноды) *</label>
                        <input
                            type="text"
                            v-model="availableForm.token"
                            class="form-input"
                            placeholder="dev-api-key-for-local-testing"
                        />
                    </div>
                    <p class="hint">
                        Ноды в этом списке самостоятельно анонсировали себя
                        оркестратору и ожидают авторизации. Введите API-ключ
                        ноды (NODE_API_KEY) для подключения.
                    </p>
                </div>

                <!-- Ручной ввод -->
                <div v-if="registerTab === 'manual'" class="form-content">
                    <div class="form-group">
                        <label>Адрес (host:port) *</label>
                        <input
                            type="text"
                            v-model="manualForm.address"
                            class="form-input"
                            placeholder="192.168.1.100:44044"
                        />
                    </div>
                    <div class="form-group">
                        <label>Ключ авторизации (API-ключ ноды) *</label>
                        <input
                            type="text"
                            v-model="manualForm.token"
                            class="form-input"
                            placeholder="dev-api-key-for-local-testing"
                        />
                    </div>
                    <div class="form-group">
                        <label>Регион (опционально)</label>
                        <input
                            type="text"
                            v-model="manualForm.region"
                            class="form-input"
                            placeholder="EU"
                        />
                    </div>
                    <p class="hint">
                        Введите адрес ноды и её API-ключ (NODE_API_KEY) для
                        прямого подключения.
                    </p>
                </div>

                <div v-if="registerError" class="start-error">
                    {{ registerError }}
                </div>
                <div class="modal-actions">
                    <button
                        class="btn-primary"
                        @click="submitRegister"
                        :disabled="!canRegister || registering"
                    >
                        {{ registering ? "Подключение..." : "Подключить" }}
                    </button>
                    <button
                        class="btn-outline"
                        @click="showRegisterForm = false"
                    >
                        Отмена
                    </button>
                </div>
            </div>
        </div>

        <!-- Подтверждение удаления -->
        <div
            v-if="deleteTarget"
            class="modal-overlay"
            @click.self="deleteTarget = null"
        >
            <div class="modal card">
                <h3>Удалить ноду?</h3>
                <p>
                    Нода <code>{{ deleteTarget.address }}</code> будет удалена
                    из реестра.
                </p>
                <p class="text-danger">
                    Все инстансы на этой ноде будут переведены в статус
                    «Авария».
                </p>
                <div class="modal-actions">
                    <button
                        class="btn-primary"
                        @click="doDelete"
                        :disabled="deleting"
                    >
                        Удалить
                    </button>
                    <button class="btn-outline" @click="deleteTarget = null">
                        Отмена
                    </button>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, computed, reactive, onMounted, watch } from "vue";
import { Plus, Trash2, AlertCircle } from "lucide-vue-next";
import StatusBadge from "../../components/orchestrator/StatusBadge.vue";
import { listNodes, registerNode, deleteNode } from "../../api/orchestrator";
import { showToast } from "../../store";

const nodes = ref([]);
const loading = ref(true);
const error = ref(null);
const statusFilter = ref("all");
const showRegisterForm = ref(false);
const registerTab = ref("available");
const registerError = ref(null);
const registering = ref(false);
const deleteTarget = ref(null);
const deleting = ref(false);
const deletingId = ref(null);

const manualForm = reactive({ address: "", token: "", region: "" });
const availableForm = reactive({ node_id: "", token: "" });

const filteredNodes = computed(() => {
    if (statusFilter.value === "all") return nodes.value;
    const statusMap = {
        unauthorized: "NODE_STATUS_UNAUTHORIZED",
        online: "NODE_STATUS_ONLINE",
        offline: "NODE_STATUS_OFFLINE",
        maintenance: "NODE_STATUS_MAINTENANCE",
    };
    const targetStatus = statusMap[statusFilter.value] || statusFilter.value;
    return nodes.value.filter((n) => n.status === targetStatus);
});

const availableNodes = computed(() =>
    nodes.value.filter(
        (n) =>
            n.status === "NODE_STATUS_UNAUTHORIZED" &&
            (!n.owner_id || n.owner_id === ""),
    ),
);

const canRegister = computed(() => {
    if (registerTab.value === "manual")
        return manualForm.address && manualForm.token;
    return availableForm.node_id && availableForm.token;
});

async function fetchNodes() {
    loading.value = true;
    error.value = null;
    try {
        nodes.value = await listNodes(
            statusFilter.value === "all" ? undefined : statusFilter.value,
        );
    } catch (e) {
        error.value = e.response?.data?.message ?? e.message;
    } finally {
        loading.value = false;
    }
}

function openRegisterModal() {
    registerTab.value = "available";
    registerError.value = null;
    showRegisterForm.value = true;
}

async function submitRegister() {
    registering.value = true;
    registerError.value = null;
    let payload;
    if (registerTab.value === "manual") {
        payload = { address: manualForm.address, token: manualForm.token };
        if (manualForm.region) payload.region = manualForm.region;
    } else {
        payload = {
            node_id: Number(availableForm.node_id),
            token: availableForm.token,
        };
    }

    try {
        await registerNode(payload);
        showToast("Нода подключена и авторизована");
        showRegisterForm.value = false;
        Object.assign(manualForm, { address: "", token: "", region: "" });
        Object.assign(availableForm, { node_id: "", token: "" });
        await fetchNodes();
    } catch (e) {
        if (e.response?.status === 401) {
            registerError.value = "Неверный ключ авторизации";
        } else if (e.response?.status === 409) {
            registerError.value = "Нода уже авторизована или уже существует";
        } else {
            registerError.value =
                e.response?.data?.message ?? "Ошибка подключения ноды";
        }
    } finally {
        registering.value = false;
    }
}

function confirmDelete(node) {
    deleteTarget.value = node;
}

async function doDelete() {
    deleting.value = true;
    deletingId.value = deleteTarget.value.id;
    try {
        await deleteNode(deleteTarget.value.id);
        showToast("Нода удалена");
        deleteTarget.value = null;
        await fetchNodes();
    } catch (e) {
        showToast(e.response?.data?.message ?? "Ошибка удаления", "error");
    } finally {
        deleting.value = false;
        deletingId.value = null;
    }
}

function formatBytes(b) {
    if (b < 1024 * 1024 * 1024) return (b / (1024 * 1024)).toFixed(0) + " MB";
    return (b / (1024 * 1024 * 1024)).toFixed(1) + " GB";
}

function formatTime(ts) {
    if (!ts) return "—";
    const d = new Date(ts);
    const now = Date.now();
    const diff = Math.floor((now - d.getTime()) / 1000);
    if (diff < 60) return "только что";
    if (diff < 3600) return Math.floor(diff / 60) + " мин. назад";
    if (diff < 86400) return Math.floor(diff / 3600) + " ч. назад";
    return d.toLocaleDateString("ru-RU", { day: "numeric", month: "short" });
}

watch(showRegisterForm, async (show) => {
    if (show) {
        try {
            nodes.value = await listNodes();
        } catch (e) {
            console.error("Failed to load nodes for modal:", e);
        }
    }
});

onMounted(fetchNodes);
</script>

<style scoped>
.nodes-page {
    padding: 32px 40px;
    max-width: 1400px;
    margin: 0 auto;
    width: 100%;
}
.page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 24px;
}
.page-header h1 {
    margin: 0;
}
.header-actions {
    display: flex;
    gap: 12px;
    align-items: center;
}
.filter-select {
    padding: 8px 12px;
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    background: var(--bg-input);
    color: var(--text-main);
    font-size: 0.88rem;
}

.error-banner {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 12px 16px;
    background: var(--danger-light);
    color: var(--danger);
    border-radius: var(--radius-md);
    margin-bottom: 16px;
    font-size: 0.88rem;
}
.btn-sm {
    padding: 4px 12px;
    font-size: 0.82rem;
}
.loading-state {
    padding: 40px;
    text-align: center;
    color: var(--text-muted);
}

.table-wrap {
    background: var(--bg-card);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    overflow: hidden;
}
.data-table {
    width: 100%;
    border-collapse: collapse;
}
.data-table th {
    text-align: left;
    padding: 12px 16px;
    font-size: 0.78rem;
    font-weight: 600;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.03em;
    background: var(--bg-secondary);
    border-bottom: 1px solid var(--border);
}
.data-table td {
    padding: 12px 16px;
    font-size: 0.88rem;
    border-bottom: 1px solid var(--border);
}
.data-table tr:last-child td {
    border-bottom: none;
}
.clickable-row {
    cursor: pointer;
    transition: 0.1s;
}
.clickable-row:hover {
    background: var(--bg-hover);
}
.cell-id {
    font-weight: 600;
}
.cell-addr code {
    background: var(--bg-secondary);
    padding: 2px 6px;
    border-radius: 4px;
    font-size: 0.82rem;
}
.cell-muted {
    color: var(--text-muted);
}
.cell-actions {
    display: flex;
    gap: 4px;
}
.btn-icon {
    background: none;
    border: none;
    color: var(--text-muted);
    cursor: pointer;
    padding: 4px;
    border-radius: 4px;
    display: flex;
    align-items: center;
}
.btn-icon:hover {
    color: var(--danger);
    background: var(--danger-light);
}
.btn-icon:disabled {
    opacity: 0.4;
    cursor: not-allowed;
}
.empty-state {
    padding: 40px;
    text-align: center;
    color: var(--text-muted);
}

.modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.4);
    z-index: 100;
    display: flex;
    align-items: center;
    justify-content: center;
}
.modal {
    max-width: 480px;
    width: 90%;
}
.modal h3 {
    margin: 0 0 16px;
}
.modal p {
    margin: 8px 0;
    font-size: 0.9rem;
    color: var(--text-muted);
}
.modal-actions {
    display: flex;
    gap: 8px;
    margin-top: 16px;
}
.text-danger {
    color: var(--danger);
    font-weight: 600;
}
code {
    background: var(--bg-secondary);
    padding: 2px 6px;
    border-radius: 4px;
    font-size: 0.82rem;
}

.tabs {
    display: flex;
    gap: 4px;
    margin-bottom: 20px;
    border-bottom: 1px solid var(--border);
}
.tab-btn {
    padding: 8px 16px;
    background: none;
    border: none;
    border-bottom: 2px solid transparent;
    font-size: 0.88rem;
    font-weight: 500;
    color: var(--text-muted);
    cursor: pointer;
}
.tab-btn.active {
    color: var(--primary);
    border-bottom-color: var(--primary);
}

.form-content {
    display: flex;
    flex-direction: column;
    gap: 16px;
}
.form-group {
    display: flex;
    flex-direction: column;
    gap: 4px;
}
.form-group label {
    font-size: 0.82rem;
    font-weight: 600;
    color: var(--text-muted);
}
.form-input {
    padding: 8px 12px;
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    background: var(--bg-input);
    color: var(--text-main);
    font-size: 0.88rem;
}
.hint {
    font-size: 0.82rem;
    color: var(--text-muted);
    margin: 4px 0 0;
}
.start-error {
    color: var(--danger);
    font-size: 0.85rem;
    margin-top: 8px;
    font-weight: 500;
}
</style>
