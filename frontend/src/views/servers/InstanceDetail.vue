<template>
  <div class="instance-detail tab-fade-in">
    <!-- Шапка -->
    <div class="detail-header">
      <div class="header-left">
        <button class="btn-outline back-btn" @click="$router.push(`/projects/${gameId}/servers/instances`)">
          <ArrowLeft class="icon-sm" /> Инстансы
        </button>
        <h1>{{ instance.name || `Инстанс #${instance.id}` }}</h1>
        <StatusBadge v-if="instance.status" :status="instance.status" type="instance" />
      </div>
      <div class="header-actions">
        <button v-if="instance.status === 'running'" class="btn-stop-lg" @click="handleStop" :disabled="stopping">
          <Square class="icon-sm" /> {{ stopping ? 'Остановка...' : 'Остановить' }}
        </button>
      </div>
    </div>

    <!-- Ошибка загрузки -->
    <div v-if="error" class="error-banner">
      <AlertCircle class="icon-sm" /> {{ error }}
      <button class="btn-outline btn-sm" @click="fetchInstance">Повторить</button>
    </div>

    <template v-if="!error">
      <!-- Информация -->
      <div class="info-grid">
        <div class="card info-card">
          <h3>Информация</h3>
          <div class="info-rows">
            <div class="info-row"><span class="info-label">ID</span><span class="info-val">{{ instance.id }}</span></div>
            <div class="info-row"><span class="info-label">Версия билда</span><span class="info-val"><code>{{ instance.build_version }}</code></span></div>
            <div class="info-row"><span class="info-label">Протокол</span><span class="info-val">{{ instance.protocol }}</span></div>
            <div class="info-row"><span class="info-label">Порты</span><span class="info-val">{{ instance.internal_port }} → {{ instance.host_port }}</span></div>
            <div class="info-row"><span class="info-label">Нода</span><span class="info-val">{{ instance.node_id }}</span></div>
            <div class="info-row"><span class="info-label">Адрес</span><span class="info-val">{{ instance.server_address }}:{{ instance.host_port }}</span></div>
            <div class="info-row"><span class="info-label">Игроки</span><span class="info-val">{{ instance.player_count ?? 0 }} / {{ instance.max_players }}</span></div>
            <div class="info-row"><span class="info-label">Создан</span><span class="info-val">{{ formatDateTime(instance.created_at) }}</span></div>
            <div class="info-row"><span class="info-label">Запущен</span><span class="info-val">{{ instance.started_at ? formatDateTime(instance.started_at) : '—' }}</span></div>
          </div>
        </div>

        <!-- developer_payload -->
        <div class="card info-card" v-if="Object.keys(instance.developer_payload || {}).length">
          <h3>Developer Payload</h3>
          <div class="info-rows">
            <div class="info-row" v-for="(v, k) in instance.developer_payload" :key="k">
              <span class="info-label">{{ k }}</span><span class="info-val">{{ v }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Ресурсы -->
      <div class="section-header">
        <h2>Потребление ресурсов</h2>
      </div>
      <div class="resources-grid">
        <ResourceUsageCard label="CPU" :value="usage.cpu_usage_percent" type="percent" />
        <ResourceUsageCard label="Память" :value="usage.memory_used_bytes" type="bytes" />
        <ResourceUsageCard label="Диск" :value="usage.disk_used_bytes" type="bytes" />
        <ResourceUsageCard label="Сеть" :value="usage.network_bytes_per_sec" unit=" байт/с" type="raw" />
      </div>

      <!-- Логи -->
      <div class="section-header">
        <h2>Журнал</h2>
      </div>
      <LogsViewer :game-id="gameId" :instance-id="instanceId" />
    </template>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ArrowLeft, Square, AlertCircle } from 'lucide-vue-next'
import StatusBadge from '../../components/orchestrator/StatusBadge.vue'
import ResourceUsageCard from '../../components/orchestrator/ResourceUsageCard.vue'
import LogsViewer from '../../components/orchestrator/LogsViewer.vue'
import { getInstance, getInstanceUsage, stopInstance } from '../../api/orchestrator'
import { showToast } from '../../store'

const props = defineProps({
  gameId: { type: [String, Number], required: true },
  instanceId: { type: [String, Number], required: true },
})
const router = useRouter()

const instance = ref({})
const usage = ref({ cpu_usage_percent: 0, memory_used_bytes: 0, disk_used_bytes: 0, network_bytes_per_sec: 0 })
const loading = ref(true)
const error = ref(null)
const stopping = ref(false)

let usageInterval = null

async function fetchInstance() {
  error.value = null
  try {
    instance.value = await getInstance(props.gameId, props.instanceId)
  } catch (e) {
    error.value = e.response?.data?.message ?? e.message
  } finally {
    loading.value = false
  }
}

async function fetchUsage() {
  try {
    const data = await getInstanceUsage(props.gameId, props.instanceId)
    usage.value = data
  } catch { /* не критично */ }
}

async function handleStop() {
  stopping.value = true
  try {
    await stopInstance(props.gameId, props.instanceId)
    showToast('Инстанс останавливается...')
    await fetchInstance()
  } catch (e) {
    showToast(e.response?.data?.message ?? 'Ошибка остановки', 'error')
  } finally {
    stopping.value = false
  }
}

function formatDateTime(ts) {
  if (!ts) return '—'
  return new Date(ts).toLocaleString('ru-RU', { day: 'numeric', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit' })
}

onMounted(async () => {
  await fetchInstance()
  await fetchUsage()
  usageInterval = setInterval(fetchUsage, 5000)
})

onUnmounted(() => {
  clearInterval(usageInterval)
})
</script>

<style scoped>
.tab-fade-in { animation: fadeIn 0.3s ease; }
@keyframes fadeIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }

.detail-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.header-left { display: flex; align-items: center; gap: 12px; }
.header-left h1 { margin: 0; font-size: 1.3rem; }
.back-btn { font-size: 0.82rem; padding: 6px 12px; }
.btn-stop-lg {
  display: flex; align-items: center; gap: 6px;
  background: none; border: 1px solid var(--danger); color: var(--danger);
  padding: 8px 16px; border-radius: var(--radius-md); font-weight: 600; cursor: pointer;
}
.btn-stop-lg:hover { background: var(--danger-light); }
.btn-stop-lg:disabled { opacity: 0.5; cursor: not-allowed; }

.error-banner {
  display: flex; align-items: center; gap: 8px;
  padding: 12px 16px; background: var(--danger-light); color: var(--danger);
  border-radius: var(--radius-md); margin-bottom: 16px; font-size: 0.88rem;
}
.btn-sm { padding: 4px 12px; font-size: 0.82rem; }

.info-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; margin-bottom: 24px; }
.info-card h3 { margin: 0 0 12px; font-size: 0.95rem; }
.info-rows { display: flex; flex-direction: column; gap: 8px; }
.info-row { display: flex; justify-content: space-between; gap: 16px; }
.info-label { font-size: 0.82rem; color: var(--text-muted); font-weight: 500; min-width: 120px; }
.info-val { font-size: 0.88rem; font-weight: 500; text-align: right; }
code { background: var(--bg-secondary); padding: 2px 6px; border-radius: 4px; font-size: 0.82rem; }

.section-header { margin-bottom: 12px; }
.section-header h2 { margin: 0; font-size: 1.05rem; }

.resources-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; margin-bottom: 24px; }

@media (max-width: 768px) {
  .info-grid { grid-template-columns: 1fr; }
  .resources-grid { grid-template-columns: repeat(2, 1fr); }
}
</style>
