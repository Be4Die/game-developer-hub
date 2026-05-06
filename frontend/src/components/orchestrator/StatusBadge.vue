<template>
  <span class="status-badge" :class="statusClass">{{ label }}</span>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  status: { type: String, required: true },
  type: { type: String, default: 'instance', validator: v => ['instance', 'node'].includes(v) },
})

const instanceMap = {
  starting: { label: 'Запускается', cls: 'warning' },
  running:  { label: 'Работает',    cls: 'success' },
  stopping: { label: 'Останавливается', cls: 'warning' },
  stopped:  { label: 'Остановлен',  cls: 'muted' },
  crashed:  { label: 'Авария',      cls: 'danger' },
  // Proto-format statuses
  INSTANCE_STATUS_UNSPECIFIED: { label: 'Неизвестно', cls: 'muted' },
  INSTANCE_STATUS_STARTING: { label: 'Запускается', cls: 'warning' },
  INSTANCE_STATUS_RUNNING:  { label: 'Работает',    cls: 'success' },
  INSTANCE_STATUS_STOPPING: { label: 'Останавливается', cls: 'warning' },
  INSTANCE_STATUS_STOPPED:  { label: 'Остановлен',  cls: 'muted' },
  INSTANCE_STATUS_CRASHED:  { label: 'Авария',      cls: 'danger' },
}

const nodeMap = {
  unauthorized: { label: 'Не авторизована', cls: 'warning' },
  online:       { label: 'В сети',          cls: 'success' },
  offline:      { label: 'Не в сети',       cls: 'muted' },
  maintenance:  { label: 'Обслуживание',    cls: 'warning' },
  // Proto-format statuses (from gRPC-gateway JSON)
  NODE_STATUS_UNSPECIFIED: { label: 'Неизвестно', cls: 'muted' },
  NODE_STATUS_UNAUTHORIZED: { label: 'Не авторизована', cls: 'warning' },
  NODE_STATUS_ONLINE:       { label: 'В сети',          cls: 'success' },
  NODE_STATUS_OFFLINE:      { label: 'Не в сети',       cls: 'muted' },
  NODE_STATUS_MAINTENANCE:  { label: 'Обслуживание',    cls: 'warning' },
}

const map = computed(() => props.type === 'node' ? nodeMap : instanceMap)

// Convert numeric status to proto enum string if needed
const statusKey = computed(() => {
  const status = props.status
  // Handle numeric statuses from API (0 -> INSTANCE_STATUS_UNSPECIFIED, etc.)
  if (typeof status === 'number' || /^\d+$/.test(String(status))) {
    const numStatus = Number(status)
    const instanceStatusMap = {
      0: 'INSTANCE_STATUS_UNSPECIFIED',
      1: 'INSTANCE_STATUS_STARTING',
      2: 'INSTANCE_STATUS_RUNNING',
      3: 'INSTANCE_STATUS_STOPPING',
      4: 'INSTANCE_STATUS_STOPPED',
      5: 'INSTANCE_STATUS_CRASHED',
    }
    const nodeStatusMap = {
      0: 'NODE_STATUS_UNSPECIFIED',
      1: 'NODE_STATUS_UNAUTHORIZED',
      2: 'NODE_STATUS_ONLINE',
      3: 'NODE_STATUS_OFFLINE',
      4: 'NODE_STATUS_MAINTENANCE',
    }
    if (props.type === 'node') {
      return nodeStatusMap[numStatus] || 'NODE_STATUS_UNSPECIFIED'
    }
    return instanceStatusMap[numStatus] || 'INSTANCE_STATUS_UNSPECIFIED'
  }
  return status
})

const label = computed(() => map.value[statusKey.value]?.label ?? props.status ?? '—')
const statusClass = computed(() => `badge-${map.value[statusKey.value]?.cls ?? 'muted'}`)
</script>

<style scoped>
.status-badge {
  display: inline-block;
  padding: 3px 10px;
  border-radius: 12px;
  font-size: 0.78rem;
  font-weight: 600;
  white-space: nowrap;
}
.badge-success { background: var(--success-light, #D1FAE5); color: #065F46; }
.badge-warning { background: var(--warning-light, #FEF3C7); color: #92400E; }
.badge-danger  { background: var(--danger-light, #FEF2F2); color: #991B1B; }
.badge-muted   { background: var(--bg-hover, #F3F4F6); color: var(--text-muted); }
</style>
