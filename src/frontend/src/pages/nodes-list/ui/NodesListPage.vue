<template>
  <div class="nodes-page-container">
    <div class="main-content-wrap">
      <!-- Панель фильтров и действий -->
      <div class="filters-toolbar">
        <!-- Поиск по адресу или региону -->
        <div class="filter-field field-search">
          <label class="field-label">Адрес / Регион</label>
          <div class="input-wrapper">
            <input
              type="text"
              v-model="searchQuery"
              placeholder="Поиск по адресу или региону..."
              class="filter-input"
            />
            <button
              v-if="searchQuery"
              class="clear-input-btn"
              title="Очистить"
              @click="searchQuery = ''"
            >
              <X class="icon-xs" />
            </button>
          </div>
        </div>

        <!-- Фильтр по статусу -->
        <div class="filter-field field-status">
          <label class="field-label">Статус</label>
          <div class="select-wrapper">
            <select
              v-model="statusFilter"
              class="filter-select"
              @change="fetchNodes"
            >
              <option value="all">—</option>
              <option value="online">В сети</option>
              <option value="offline">Не в сети</option>
              <option value="unauthorized">Не авторизована</option>
              <option value="maintenance">Обслуживание</option>
            </select>
            <ChevronDown class="icon-xs select-arrow" />
          </div>
        </div>

        <!-- Кнопка сброса фильтров -->
        <button
          v-if="searchQuery || statusFilter !== 'all'"
          class="btn-reset-filters"
          @click="resetFilters"
          title="Сбросить фильтры"
        >
          <RotateCcw class="icon-xs" />
          <span>Сбросить</span>
        </button>

        <!-- Кнопка подключения ноды -->
        <button class="btn-add-node" @click="openRegisterModal">
          <Plus class="icon-sm" />
          <span>Подключить ноду</span>
        </button>
      </div>

      <!-- Ошибка -->
      <div v-if="error" class="error-banner">
        <AlertCircle class="icon-sm" /> {{ error }}
        <button class="btn-retry" @click="fetchNodes">
          Повторить
        </button>
      </div>

      <!-- Загрузка -->
      <div v-if="loading" class="state-container">
        <div class="spinner-md"></div>
        <p>Загрузка нод...</p>
      </div>

      <!-- Таблица нод -->
      <div v-else-if="filteredNodes.length" class="table-wrapper">
        <table class="nodes-table">
          <thead>
            <tr>
              <th class="col-addr">Адрес</th>
              <th class="col-region">Регион</th>
              <th class="col-status">Статус</th>
              <th class="col-cpu">CPU</th>
              <th class="col-ram">Память</th>
              <th class="col-disk">Диск</th>
              <th class="col-agent">Версия</th>
              <th class="col-ping">Heartbeat</th>
              <th class="col-actions"></th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="node in filteredNodes"
              :key="node.id"
              class="table-row"
              @click="$router.push(`/nodes/${node.id}`)"
            >
              <!-- Адрес -->
              <td class="col-addr">
                <span class="node-address">{{ node.address }}</span>
              </td>

              <!-- Регион -->
              <td class="col-region">
                <span class="cell-text">{{ node.region || '—' }}</span>
              </td>

              <!-- Статус -->
              <td class="col-status">
                <StatusBadge :status="node.status" type="node" />
              </td>

              <!-- CPU -->
              <td class="col-cpu">
                <span class="cell-text">
                  {{ node.cpu_cores ? node.cpu_cores + ' ядер' : '—' }}
                </span>
              </td>

              <!-- Память -->
              <td class="col-ram">
                <span class="cell-text">
                  {{
                    node.total_memory_bytes
                      ? formatBytes(node.total_memory_bytes)
                      : '—'
                  }}
                </span>
              </td>

              <!-- Диск -->
              <td class="col-disk">
                <span class="cell-text">
                  {{
                    node.total_disk_bytes
                      ? formatBytes(node.total_disk_bytes)
                      : '—'
                  }}
                </span>
              </td>

              <!-- Агент -->
              <td class="col-agent">
                <span class="cell-muted">{{ node.agent_version || '—' }}</span>
              </td>

              <!-- Heartbeat -->
              <td class="col-ping">
                <span class="cell-muted">{{ formatTime(node.last_ping_at) }}</span>
              </td>

              <!-- Действия -->
              <td class="col-actions" @click.stop>
                <button
                  class="btn-icon text-danger-hover"
                  @click="confirmDelete(node)"
                  title="Удалить"
                  :disabled="deletingId === node.id"
                >
                  <Trash2 class="icon-xs" />
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Пустое состояние -->
      <div v-else class="state-container empty-card">
        <div class="empty-icon-wrap">
          <Server class="icon-lg" />
        </div>
        <h3>Нет вычислительных нод</h3>
        <p>
          {{ statusFilter !== 'all' ? 'Нет нод с выбранным статусом' : 'Подключите свой сервер для оркестрации игровых инстансов' }}
        </p>
        <button class="btn-add-node" @click="openRegisterModal">
          <Plus class="icon-sm" />
          <span>Подключить ноду</span>
        </button>
      </div>
    </div>

    <!-- Модал подключения ноды -->
    <RegisterNodeModal
      v-if="showRegisterForm"
      :available-nodes="availableNodes"
      @registered="onNodeRegistered"
      @cancel="showRegisterForm = false"
    />

    <!-- Подтверждение удаления -->
    <div
      v-if="deleteTarget"
      class="modal-overlay"
      @click.self="deleteTarget = null"
    >
      <div class="modal card">
        <h3>Удалить ноду?</h3>
        <p>
          Нода <code>{{ deleteTarget.address }}</code> будет удалена из
          реестра.
        </p>
        <p class="text-danger">
          Все инстансы на этой ноде будут переведены в статус «Авария».
        </p>
        <div class="modal-actions">
          <button
            class="btn-danger"
            @click="doDelete"
            :disabled="deleting"
          >
            {{ deleting ? 'Удаление...' : 'Удалить' }}
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
import { ref, computed, onMounted, watch } from 'vue';
import { Plus, Trash2, AlertCircle, Server, X, ChevronDown, RotateCcw } from 'lucide-vue-next';
import { StatusBadge } from '@/shared/ui';
import { RegisterNodeModal } from '@/features/manage-nodes';
import { listNodes, deleteNode } from '@/entities/node';
import { formatBytes, formatTime, showToast } from '@/shared/lib';

const nodes = ref([]);
const loading = ref(true);
const error = ref(null);
const searchQuery = ref('');
const statusFilter = ref('all');
const showRegisterForm = ref(false);
const deleteTarget = ref(null);
const deleting = ref(false);
const deletingId = ref(null);

const filteredNodes = computed(() => {
  let list = [...nodes.value];
  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase().trim();
    list = list.filter((n) => {
      const addr = (n.address || '').toLowerCase();
      const reg = (n.region || '').toLowerCase();
      return addr.includes(q) || reg.includes(q);
    });
  }
  if (statusFilter.value !== 'all') {
    const statusMap = {
      unauthorized: 'NODE_STATUS_UNAUTHORIZED',
      online: 'NODE_STATUS_ONLINE',
      offline: 'NODE_STATUS_OFFLINE',
      maintenance: 'NODE_STATUS_MAINTENANCE',
    };
    const targetStatus = statusMap[statusFilter.value] || statusFilter.value;
    list = list.filter((n) => n.status === targetStatus);
  }
  return list;
});

const availableNodes = computed(() =>
  nodes.value.filter(
    (n) =>
      n.status === 'NODE_STATUS_UNAUTHORIZED' &&
      (!n.owner_id || n.owner_id === '')
  )
);

function resetFilters() {
  searchQuery.value = '';
  statusFilter.value = 'all';
}

async function fetchNodes() {
  loading.value = true;
  error.value = null;
  try {
    nodes.value = await listNodes(
      statusFilter.value === 'all' ? undefined : statusFilter.value
    );
  } catch (e) {
    error.value = e.response?.data?.message ?? e.message;
  } finally {
    loading.value = false;
  }
}

function openRegisterModal() {
  showRegisterForm.value = true;
}

function onNodeRegistered() {
  showRegisterForm.value = false;
  fetchNodes();
}

function confirmDelete(node) {
  deleteTarget.value = node;
}

async function doDelete() {
  deleting.value = true;
  deletingId.value = deleteTarget.value.id;
  try {
    await deleteNode(deleteTarget.value.id);
    showToast('Нода успешно удалена', 'success');
    deleteTarget.value = null;
    await fetchNodes();
  } catch (e) {
    showToast(e.response?.data?.message ?? 'Ошибка удаления', 'danger');
  } finally {
    deleting.value = false;
    deletingId.value = null;
  }
}

watch(showRegisterForm, async (show) => {
  if (show) {
    try {
      nodes.value = await listNodes();
    } catch (e) {
      console.error('Failed to load nodes for modal:', e);
    }
  }
});

onMounted(fetchNodes);
</script>

<style scoped>
.nodes-page-container {
  width: 100%;
  max-width: 100%;
  min-height: calc(100vh - 60px);
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 24px 32px;
  box-sizing: border-box;
}

.main-content-wrap {
  width: 100%;
}

.filters-toolbar {
  display: flex;
  align-items: flex-end;
  gap: 16px;
  margin-bottom: 24px;
  width: 100%;
}

.filter-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.field-search {
  flex: 1;
  min-width: 220px;
}

.field-status {
  width: 180px;
  flex-shrink: 0;
}

.field-label {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-tertiary, #8b949e);
  letter-spacing: 0.1px;
}

.input-wrapper,
.select-wrapper {
  position: relative;
  display: flex;
  align-items: center;
  width: 100%;
}

.filter-input {
  width: 100%;
  height: 36px;
  padding: 0 32px 0 12px;
  background: var(--bg-secondary, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  color: var(--text-main, #f0f6fc);
  font-size: 13px;
  outline: none;
  transition: border-color 0.15s;
}

.filter-input:focus {
  border-color: var(--primary, #58a6ff);
}

.filter-input::placeholder {
  color: var(--text-tertiary, #6e7681);
}

.clear-input-btn {
  position: absolute;
  right: 8px;
  background: transparent;
  border: none;
  color: var(--text-tertiary, #8b949e);
  cursor: pointer;
  padding: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.clear-input-btn:hover {
  color: var(--text-main, #f0f6fc);
}

.filter-select {
  width: 100%;
  height: 36px;
  padding: 0 30px 0 12px;
  background: var(--bg-secondary, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  color: var(--text-main, #f0f6fc);
  font-size: 13px;
  outline: none;
  cursor: pointer;
  appearance: none;
  -webkit-appearance: none;
  transition: border-color 0.15s;
}

.filter-select:focus {
  border-color: var(--primary, #58a6ff);
}

.select-arrow {
  position: absolute;
  right: 10px;
  pointer-events: none;
  color: var(--text-tertiary, #8b949e);
}

.btn-reset-filters {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 36px;
  padding: 0 12px;
  background: transparent;
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  color: var(--text-muted, #b0b8c4);
  font-size: 13px;
  cursor: pointer;
  flex-shrink: 0;
  transition: all 0.15s;
}

.btn-reset-filters:hover {
  background: var(--bg-tertiary, #21262d);
  color: var(--text-main, #f0f6fc);
  border-color: var(--border-secondary, #484f58);
}

.btn-add-node {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  height: 36px;
  padding: 0 18px;
  background: var(--primary, #58a6ff);
  color: #ffffff;
  border: none;
  border-radius: var(--radius-sm, 6px);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  flex-shrink: 0;
  transition: background-color 0.15s;
  white-space: nowrap;
}

.btn-add-node:hover {
  background: var(--primary-hover, #79c0ff);
}

.error-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 16px;
  background: var(--danger-light, rgba(248, 81, 73, 0.1));
  color: var(--danger, #f85149);
  border: 1px solid rgba(248, 81, 73, 0.25);
  border-radius: var(--radius-sm, 6px);
  margin-bottom: 20px;
  font-size: 13px;
}

.btn-retry {
  padding: 4px 10px;
  background: transparent;
  border: 1px solid currentColor;
  border-radius: 4px;
  color: inherit;
  font-size: 12px;
  cursor: pointer;
}

.state-container {
  padding: 60px 20px;
  text-align: center;
  color: var(--text-tertiary, #8b949e);
}

.spinner-md {
  width: 32px;
  height: 32px;
  border: 3px solid var(--border, #30363d);
  border-top-color: var(--primary, #58a6ff);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin: 0 auto 12px;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.empty-card {
  background: var(--bg-card, #161b22);
  border: 1px dashed var(--border, #30363d);
  border-radius: var(--radius-md, 8px);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 48px 24px;
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

.nodes-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
}

.nodes-table th {
  color: var(--text-tertiary, #8b949e);
  font-size: 13px;
  font-weight: 500;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border, #30363d);
  background: transparent;
  white-space: nowrap;
}

.nodes-table th.col-addr {
  width: 28%;
  padding-left: 8px;
}

.nodes-table th.col-region {
  width: 12%;
}

.nodes-table th.col-status {
  width: 14%;
}

.nodes-table th.col-cpu {
  width: 10%;
}

.nodes-table th.col-ram {
  width: 10%;
}

.nodes-table th.col-disk {
  width: 10%;
}

.nodes-table th.col-agent {
  width: 8%;
}

.nodes-table th.col-ping {
  width: 8%;
}

.nodes-table th.col-actions {
  width: 4%;
  text-align: right;
  padding-right: 12px;
}

.table-row {
  cursor: pointer;
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

.table-row td.col-addr {
  padding-left: 8px;
}

.table-row td.col-actions {
  padding-right: 12px;
  text-align: right;
}

.node-address {
  font-family: inherit;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
}

.cell-text {
  font-size: 13px;
  color: var(--text-muted, #b0b8c4);
}

.cell-muted {
  font-size: 13px;
  color: var(--text-tertiary, #8b949e);
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

.text-danger-hover:hover {
  color: var(--danger, #f85149) !important;
}

.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
}

.modal {
  max-width: 480px;
  width: 90%;
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-lg, 10px);
  padding: 24px;
}

.modal h3 {
  margin: 0 0 12px;
  font-size: 1.1rem;
  color: var(--text-main, #f0f6fc);
}

.modal p {
  margin: 8px 0;
  font-size: 0.9rem;
  color: var(--text-muted, #b0b8c4);
}

.modal-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
  margin-top: 20px;
}

.text-danger {
  color: var(--danger, #f85149);
  font-weight: 500;
}

code {
  background: var(--bg-tertiary, #21262d);
  color: var(--text-main, #f0f6fc);
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 0.82rem;
}

@media (max-width: 800px) {
  .filters-toolbar {
    flex-wrap: wrap;
  }
  .field-search {
    width: 100%;
    min-width: 100%;
  }
}
</style>
