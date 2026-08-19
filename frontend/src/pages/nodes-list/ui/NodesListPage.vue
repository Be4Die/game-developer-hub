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
            <td>{{ node.region || '—' }}</td>
            <td>
              <StatusBadge :status="node.status" type="node" />
            </td>
            <td>
              {{ node.cpu_cores ? node.cpu_cores + ' ядер' : '—' }}
            </td>
            <td>
              {{
                node.total_memory_bytes
                  ? formatBytes(node.total_memory_bytes)
                  : '—'
              }}
            </td>
            <td>
              {{
                node.total_disk_bytes
                  ? formatBytes(node.total_disk_bytes)
                  : '—'
              }}
            </td>
            <td class="cell-muted">
              {{ node.agent_version || '—' }}
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
      Нет нод{{ statusFilter !== 'all' ? ' с выбранным статусом' : '' }}
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
import { ref, computed, onMounted, watch } from 'vue';
import { Plus, Trash2, AlertCircle } from 'lucide-vue-next';
import { StatusBadge } from '@/shared/ui';
import { RegisterNodeModal } from '@/features/manage-nodes';
import { listNodes, deleteNode } from '@/entities/node';
import { formatBytes, formatTime, showToast } from '@/shared/lib';

const nodes = ref([]);
const loading = ref(true);
const error = ref(null);
const statusFilter = ref('all');
const showRegisterForm = ref(false);
const deleteTarget = ref(null);
const deleting = ref(false);
const deletingId = ref(null);

const filteredNodes = computed(() => {
  if (statusFilter.value === 'all') return nodes.value;
  const statusMap = {
    unauthorized: 'NODE_STATUS_UNAUTHORIZED',
    online: 'NODE_STATUS_ONLINE',
    offline: 'NODE_STATUS_OFFLINE',
    maintenance: 'NODE_STATUS_MAINTENANCE',
  };
  const targetStatus = statusMap[statusFilter.value] || statusFilter.value;
  return nodes.value.filter((n) => n.status === targetStatus);
});

const availableNodes = computed(() =>
  nodes.value.filter(
    (n) =>
      n.status === 'NODE_STATUS_UNAUTHORIZED' &&
      (!n.owner_id || n.owner_id === '')
  )
);

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
    showToast('Нода удалена');
    deleteTarget.value = null;
    await fetchNodes();
  } catch (e) {
    showToast(e.response?.data?.message ?? 'Ошибка удаления', 'error');
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
</style>
