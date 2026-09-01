<template>
  <div class="instances-page tab-fade-in">
    <div class="page-header">
      <h1>
        {{ t('servers.tabs.instances') }} <span class="counter">{{ instances.length }}/4</span>
      </h1>
      <div class="header-actions">
        <select v-model="statusFilter" class="filter-select" @change="fetchInstances">
          <option value="all">{{ t('projects.allStatuses') }}</option>
          <option value="starting">Starting</option>
          <option value="running">Running</option>
          <option value="stopping">Stopping</option>
          <option value="stopped">Stopped</option>
          <option value="crashed">Crashed</option>
        </select>
        <button class="btn-primary" @click="showStartForm = true">
          <Play class="icon-sm" /> {{ t('servers.startInstance') }}
        </button>
      </div>
    </div>

    <!-- Ошибка -->
    <div v-if="error" class="error-banner">
      <AlertCircle class="icon-sm" /> {{ error }}
      <button class="btn-outline btn-sm" @click="fetchInstances">
        {{ t('common.refresh') }}
      </button>
    </div>

    <!-- Таблица инстансов -->
    <div v-if="loading" class="loading-state">{{ t('common.loading') }}</div>
    <div v-else-if="filteredInstances.length" class="table-wrap">
      <table class="data-table">
        <thead>
          <tr>
            <th>{{ t('common.name') }}</th>
            <th>{{ t('common.version') }}</th>
            <th>{{ t('common.status') }}</th>
            <th>Node</th>
            <th>{{ t('stats.players') }}</th>
            <th>Address</th>
            <th>{{ t('common.created') }}</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="inst in filteredInstances"
            :key="inst.id"
            class="clickable-row"
            @click="$router.push(`/projects/${gameId}/servers/instances/${inst.id}`)"
          >
            <td class="cell-name">
              {{ inst.name || `Instance #${inst.id}` }}
            </td>
            <td>
              <code>{{ inst.build_version }}</code>
            </td>
            <td><StatusBadge :status="inst.status" type="instance" /></td>
            <td class="cell-muted">{{ inst.node_id }}</td>
            <td>{{ inst.player_count ?? 0 }} / {{ inst.max_players }}</td>
            <td class="cell-muted">{{ inst.server_address }}:{{ inst.host_port }}</td>
            <td class="cell-muted">
              {{ inst.started_at ? formatDate(inst.started_at) : '—' }}
            </td>
            <td class="cell-actions" @click.stop>
              <button
                v-if="inst.status === 'running'"
                class="btn-stop"
                :disabled="stoppingId === inst.id"
                title="Stop"
                @click="handleStop(inst)"
              >
                <Square class="icon-sm" />
              </button>
              <button
                v-if="inst.status === 'stopped' || inst.status === 'crashed'"
                class="btn-resume"
                :disabled="resumingId === inst.id"
                title="Start"
                @click="handleResume(inst)"
              >
                <Play class="icon-sm" />
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <div v-else class="empty-state">
      {{ t('servers.noInstances') }}
    </div>

    <!-- Модал запуска нового инстанса -->
    <StartInstanceModal
      v-if="showStartForm"
      :game-id="gameId"
      :available-builds="availableBuilds"
      @started="onInstanceStarted"
      @cancel="showStartForm = false"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { Play, Square, AlertCircle } from 'lucide-vue-next';
import { StatusBadge } from '@/shared/ui';
import { listInstances, stopInstance, resumeInstance } from '@/entities/instance';
import { listServerBuilds } from '@/entities/build';
import { StartInstanceModal } from '@/features/manage-instances';
import { formatDate, showToast } from '@/shared/lib';

const { t } = useI18n();

const props = defineProps({
  gameId: { type: [String, Number], required: true },
});

const instances = ref([]);
const availableBuilds = ref([]);
const loading = ref(true);
const error = ref(null);
const statusFilter = ref('all');
const showStartForm = ref(false);
const stoppingId = ref(null);
const resumingId = ref(null);

const filteredInstances = computed(() => {
  if (statusFilter.value === 'all') return instances.value;
  return instances.value.filter((i) => i.status === statusFilter.value);
});

async function fetchInstances() {
  loading.value = true;
  error.value = null;
  try {
    instances.value = await listInstances(
      props.gameId,
      statusFilter.value === 'all' ? undefined : statusFilter.value
    );
  } catch (e) {
    error.value = e.response?.data?.message ?? e.message;
  } finally {
    loading.value = false;
  }
}

async function fetchBuilds() {
  try {
    availableBuilds.value = await listServerBuilds(props.gameId);
  } catch {
    /* non-critical */
  }
}

function onInstanceStarted() {
  showStartForm.value = false;
  fetchInstances();
}

async function handleStop(inst) {
  stoppingId.value = inst.id;
  try {
    await stopInstance(props.gameId, inst.id);
    showToast(`Инстанс ${inst.name || inst.id} останавливается...`);
    await fetchInstances();
  } catch (e) {
    showToast(e.response?.data?.message ?? 'Ошибка остановки', 'error');
  } finally {
    stoppingId.value = null;
  }
}

async function handleResume(inst) {
  resumingId.value = inst.id;
  try {
    await resumeInstance(props.gameId, inst.id);
    showToast(`Инстанс ${inst.name || inst.id} запускается...`);
    await fetchInstances();
  } catch (e) {
    showToast(e.response?.data?.message ?? 'Ошибка запуска', 'error');
  } finally {
    resumingId.value = null;
  }
}

onMounted(() => {
  fetchInstances();
  fetchBuilds();
});
</script>

<style scoped>
.tab-fade-in {
  animation: fadeIn 0.3s ease;
}
@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
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
.counter {
  font-size: 0.9rem;
  font-weight: 400;
  color: var(--text-muted);
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
.cell-name {
  font-weight: 600;
}
.cell-muted {
  color: var(--text-muted);
}
.cell-actions {
  display: flex;
  gap: 4px;
}
code {
  background: var(--bg-secondary);
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 0.82rem;
}

.btn-stop {
  background: none;
  border: 1px solid var(--border);
  padding: 4px 8px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  display: flex;
  align-items: center;
  color: var(--text-muted);
}
.btn-stop:hover {
  color: var(--danger);
  border-color: var(--danger);
}
.btn-stop:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.btn-resume {
  background: none;
  border: 1px solid var(--border);
  padding: 4px 8px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  display: flex;
  align-items: center;
  color: var(--text-muted);
}
.btn-resume:hover {
  color: var(--success);
  border-color: var(--success);
}
.btn-resume:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.empty-state {
  padding: 40px;
  text-align: center;
  color: var(--text-muted);
}
</style>
