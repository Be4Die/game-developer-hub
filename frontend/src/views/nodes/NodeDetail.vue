<template>
  <div class="node-detail">
    <!-- Шапка -->
    <div class="detail-header">
      <div class="header-left">
        <button class="btn-outline back-btn" @click="$router.push('/nodes')">
          <ArrowLeft class="icon-sm" /> Ноды
        </button>
        <h1>{{ node.address }}</h1>
        <StatusBadge v-if="node.status" :status="node.status" type="node" />
      </div>
      <button class="btn-delete-lg" @click="showDeleteConfirm = true">
        <Trash2 class="icon-sm" /> Удалить ноду
      </button>
    </div>

    <!-- Ошибка -->
    <div v-if="error" class="error-banner">
      <AlertCircle class="icon-sm" /> {{ error }}
      <button class="btn-outline btn-sm" @click="fetchNode">Повторить</button>
    </div>

    <template v-if="!error">
      <!-- Информация -->
      <div class="card info-card">
        <h3>Информация о ноде</h3>
        <div class="info-grid-inner">
          <div class="info-row"><span class="info-label">ID</span><span class="info-val">{{ node.id }}</span></div>
          <div class="info-row"><span class="info-label">Адрес</span><span class="info-val"><code>{{ node.address }}</code></span></div>
          <div class="info-row"><span class="info-label">Регион</span><span class="info-val">{{ node.region || '—' }}</span></div>
          <div class="info-row"><span class="info-label">Статус</span><span class="info-val"><StatusBadge :status="node.status" type="node" /></span></div>
          <div class="info-row"><span class="info-label">CPU</span><span class="info-val">{{ node.cpu_cores ? node.cpu_cores + ' ядер' : '—' }}</span></div>
          <div class="info-row"><span class="info-label">Память</span><span class="info-val">{{ node.total_memory_bytes ? formatBytes(node.total_memory_bytes) : '—' }}</span></div>
          <div class="info-row"><span class="info-label">Диск</span><span class="info-val">{{ node.total_disk_bytes ? formatBytes(node.total_disk_bytes) : '—' }}</span></div>
          <div class="info-row"><span class="info-label">Версия агента</span><span class="info-val">{{ node.agent_version || '—' }}</span></div>
          <div class="info-row"><span class="info-label">Последний heartbeat</span><span class="info-val">{{ formatDateTime(node.last_ping_at) }}</span></div>
          <div class="info-row"><span class="info-label">Создана</span><span class="info-val">{{ formatDateTime(node.created_at) }}</span></div>
        </div>
      </div>

      <!-- Ресурсы -->
      <div class="section-header"><h2>Потребление ресурсов</h2></div>
      <div class="resources-grid">
        <ResourceUsageCard label="CPU" :value="usage.cpu_usage_percent" type="percent" />
        <ResourceUsageCard label="Память" :value="usage.memory_used_bytes" :max="node.total_memory_bytes" type="bytes" />
        <ResourceUsageCard label="Диск" :value="usage.disk_used_bytes" :max="node.total_disk_bytes" type="bytes" />
        <ResourceUsageCard label="Сеть" :value="usage.network_bytes_per_sec" unit=" байт/с" type="raw" />
      </div>

      <!-- Активные инстансы -->
      <div class="section-header">
        <h2>Инстансы на ноде <span class="count">{{ activeCount }}</span></h2>
      </div>
      <div v-if="instancesLoading" class="loading-state">Загрузка...</div>
      <div class="table-wrap" v-else-if="nodeInstances.length">
        <table class="data-table">
          <thead>
            <tr>
              <th>Имя</th>
              <th>Игра</th>
              <th>Версия</th>
              <th>Статус</th>
              <th>Игроки</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="inst in nodeInstances" :key="inst.id"
                @click="$router.push(`/projects/${inst.game_id}/servers/instances/${inst.id}`)"
                class="clickable-row">
              <td class="cell-name">{{ inst.name || `#${inst.id}` }}</td>
              <td>Игра #{{ inst.game_id }}</td>
              <td><code>{{ inst.build_version }}</code></td>
              <td><StatusBadge :status="inst.status" type="instance" /></td>
              <td>{{ inst.player_count ?? 0 }} / {{ inst.max_players }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <div v-else class="empty-state">Нет инстансов на этой ноде</div>
    </template>

    <!-- Подтверждение удаления -->
    <div v-if="showDeleteConfirm" class="modal-overlay" @click.self="showDeleteConfirm = false">
      <div class="modal card">
        <h3>Удалить ноду?</h3>
        <p>Нода <code>{{ node.address }}</code> будет удалена из реестра.</p>
        <p class="text-danger">Все инстансы на этой ноде будут переведены в статус «Авария».</p>
        <div class="modal-actions">
          <button class="btn-primary" @click="doDelete" :disabled="deleting">Удалить</button>
          <button class="btn-outline" @click="showDeleteConfirm = false">Отмена</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ArrowLeft, Trash2, AlertCircle } from 'lucide-vue-next'
import StatusBadge from '../../components/orchestrator/StatusBadge.vue'
import ResourceUsageCard from '../../components/orchestrator/ResourceUsageCard.vue'
import { getNode, getNodeUsage, deleteNode, listInstances } from '../../api/orchestrator'
import { showToast } from '../../store'

const props = defineProps({ nodeId: { type: [String, Number], required: true } })
const router = useRouter()

const node = ref({})
const usage = ref({ cpu_usage_percent: 0, memory_used_bytes: 0, disk_used_bytes: 0, network_bytes_per_sec: 0, active_instance_count: 0 })
const nodeInstances = ref([])
const instancesLoading = ref(true)
const error = ref(null)
const showDeleteConfirm = ref(false)
const deleting = ref(false)

const activeCount = computed(() => usage.value.active_instance_count ?? 0)

let usageInterval = null

async function fetchNode() {
  error.value = null
  try {
    node.value = await getNode(props.nodeId)
  } catch (e) {
    error.value = e.response?.data?.message ?? e.message
  }
}

async function fetchUsage() {
  try {
    const data = await getNodeUsage(props.nodeId)
    usage.value = data
  } catch { /* не критично */ }
}

async function fetchInstances() {
  instancesLoading.value = true
  try {
    // Получаем все инстансы и фильтруем по ноде
    // TODO: когда API поддержит фильтрацию по node_id, убрать фильтрацию на клиенте
    const allInstances = await listInstances(0) // 0 = все игры? Или нужен перебор
    nodeInstances.value = allInstances.filter(i => i.node_id === Number(props.nodeId))
  } catch {
    nodeInstances.value = []
  } finally {
    instancesLoading.value = false
  }
}

async function doDelete() {
  deleting.value = true
  try {
    await deleteNode(props.nodeId)
    showToast('Нода удалена')
    router.push('/nodes')
  } catch (e) {
    showToast(e.response?.data?.message ?? 'Ошибка удаления', 'error')
  } finally {
    deleting.value = false
  }
}

function formatBytes(b) {
  if (!b) return '—'
  if (b < 1024 * 1024 * 1024) return (b / (1024 * 1024)).toFixed(0) + ' MB'
  return (b / (1024 * 1024 * 1024)).toFixed(1) + ' GB'
}

function formatDateTime(ts) {
  if (!ts) return '—'
  return new Date(ts).toLocaleString('ru-RU', { day: 'numeric', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit' })
}

onMounted(async () => {
  await fetchNode()
  await fetchUsage()
  // Инстансы загружаем только если нода найдена
  if (!error.value) {
    await fetchInstances()
    usageInterval = setInterval(fetchUsage, 5000)
  }
})

onUnmounted(() => {
  clearInterval(usageInterval)
})
</script>

<style scoped>
.node-detail { padding: 32px 40px; max-width: 1200px; margin: 0 auto; width: 100%; }

.detail-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.header-left { display: flex; align-items: center; gap: 12px; }
.header-left h1 { margin: 0; font-size: 1.2rem; }
.back-btn { font-size: 0.82rem; padding: 6px 12px; }
.btn-delete-lg {
  display: flex; align-items: center; gap: 6px;
  background: none; border: 1px solid var(--danger); color: var(--danger);
  padding: 8px 16px; border-radius: var(--radius-md); font-weight: 600; cursor: pointer;
}
.btn-delete-lg:hover { background: var(--danger-light); }

.error-banner {
  display: flex; align-items: center; gap: 8px;
  padding: 12px 16px; background: var(--danger-light); color: var(--danger);
  border-radius: var(--radius-md); margin-bottom: 16px; font-size: 0.88rem;
}
.btn-sm { padding: 4px 12px; font-size: 0.82rem; }

.info-card { margin-bottom: 24px; }
.info-card h3 { margin: 0 0 16px; font-size: 0.95rem; }
.info-grid-inner { display: grid; grid-template-columns: 1fr 1fr; gap: 8px 24px; }
.info-row { display: flex; justify-content: space-between; }
.info-label { font-size: 0.82rem; color: var(--text-muted); font-weight: 500; }
.info-val { font-size: 0.88rem; font-weight: 500; }
code { background: var(--bg-secondary); padding: 2px 6px; border-radius: 4px; font-size: 0.82rem; }

.section-header { margin-bottom: 12px; display: flex; align-items: center; gap: 8px; }
.section-header h2 { margin: 0; font-size: 1.05rem; }
.count { font-size: 0.82rem; font-weight: 400; color: var(--text-muted); background: var(--bg-secondary); padding: 2px 8px; border-radius: 10px; }

.resources-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; margin-bottom: 24px; }

.loading-state { padding: 40px; text-align: center; color: var(--text-muted); }
.table-wrap { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); overflow: hidden; }
.data-table { width: 100%; border-collapse: collapse; }
.data-table th {
  text-align: left; padding: 12px 16px; font-size: 0.78rem; font-weight: 600;
  color: var(--text-muted); text-transform: uppercase; background: var(--bg-secondary); border-bottom: 1px solid var(--border);
}
.data-table td { padding: 12px 16px; font-size: 0.88rem; border-bottom: 1px solid var(--border); }
.data-table tr:last-child td { border-bottom: none; }
.clickable-row { cursor: pointer; transition: 0.1s; }
.clickable-row:hover { background: var(--bg-hover); }
.cell-name { font-weight: 600; }
.empty-state { padding: 40px; text-align: center; color: var(--text-muted); }

.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.4); z-index: 100; display: flex; align-items: center; justify-content: center; }
.modal { max-width: 420px; width: 90%; }
.modal h3 { margin: 0 0 8px; }
.modal p { margin: 8px 0; font-size: 0.9rem; color: var(--text-muted); }
.modal-actions { display: flex; gap: 8px; margin-top: 16px; }
.text-danger { color: var(--danger); font-weight: 600; }

@media (max-width: 768px) {
  .resources-grid { grid-template-columns: repeat(2, 1fr); }
  .info-grid-inner { grid-template-columns: 1fr; }
}
</style>
