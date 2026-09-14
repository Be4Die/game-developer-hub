<template>
  <span class="status-badge" :class="statusClass">{{ label }}</span>
</template>

<script setup>
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

const { t, te } = useI18n();

const props = defineProps({
  status: { type: [String, Number], required: true },
  type: {
    type: String,
    default: 'instance',
    validator: (v) => ['instance', 'node', 'role', 'service', 'platform_access'].includes(v),
  },
});

const platformAccessMap = {
  PLATFORM_ACCESS_STATUS_UNSPECIFIED: { label: 'Неизвестно', cls: 'muted' },
  PLATFORM_ACCESS_STATUS_PENDING: { label: 'На рассмотрении', cls: 'warning' },
  PLATFORM_ACCESS_STATUS_APPROVED: { label: 'Доступ открыт', cls: 'success' },
  PLATFORM_ACCESS_STATUS_REJECTED: { label: 'Отклонено', cls: 'danger' },
  REQUEST_STATUS_UNSPECIFIED: { label: 'Неизвестно', cls: 'muted' },
  REQUEST_STATUS_PENDING: { label: 'На рассмотрении', cls: 'warning' },
  REQUEST_STATUS_IN_REVIEW: { label: 'В проверке', cls: 'warning' },
  REQUEST_STATUS_APPROVED: { label: 'Доступ открыт', cls: 'success' },
  REQUEST_STATUS_REJECTED: { label: 'Отклонено', cls: 'danger' },
  pending: { label: 'На рассмотрении', cls: 'warning' },
  in_review: { label: 'В проверке', cls: 'warning' },
  approved: { label: 'Доступ открыт', cls: 'success' },
  rejected: { label: 'Отклонено', cls: 'danger' },
};

const instanceMap = {
  starting: { label: 'Запускается', cls: 'warning' },
  running: { label: 'Работает', cls: 'success' },
  stopping: { label: 'Останавливается', cls: 'warning' },
  stopped: { label: 'Остановлен', cls: 'muted' },
  crashed: { label: 'Авария', cls: 'danger' },
  // Proto-format statuses
  INSTANCE_STATUS_UNSPECIFIED: { label: 'Неизвестно', cls: 'muted' },
  INSTANCE_STATUS_STARTING: { label: 'Запускается', cls: 'warning' },
  INSTANCE_STATUS_RUNNING: { label: 'Работает', cls: 'success' },
  INSTANCE_STATUS_STOPPING: { label: 'Останавливается', cls: 'warning' },
  INSTANCE_STATUS_STOPPED: { label: 'Остановлен', cls: 'muted' },
  INSTANCE_STATUS_CRASHED: { label: 'Авария', cls: 'danger' },
};

const nodeMap = {
  unauthorized: { label: 'Не авторизована', cls: 'warning' },
  online: { label: 'В сети', cls: 'success' },
  offline: { label: 'Не в сети', cls: 'muted' },
  maintenance: { label: 'Обслуживание', cls: 'warning' },
  // Proto-format statuses
  NODE_STATUS_UNSPECIFIED: { label: 'Неизвестно', cls: 'muted' },
  NODE_STATUS_UNAUTHORIZED: { label: 'Не авторизована', cls: 'warning' },
  NODE_STATUS_ONLINE: { label: 'В сети', cls: 'success' },
  NODE_STATUS_OFFLINE: { label: 'Не в сети', cls: 'muted' },
  NODE_STATUS_MAINTENANCE: { label: 'Обслуживание', cls: 'warning' },
};

const roleLabels = computed(() => ({
  mixed: te('servers.nodeRoles.mixed') ? t('servers.nodeRoles.mixed') : 'Mixed',
  compute: te('servers.nodeRoles.compute') ? t('servers.nodeRoles.compute') : 'Compute',
  storage: te('servers.nodeRoles.storage') ? t('servers.nodeRoles.storage') : 'Storage',
}));

const roleMap = computed(() => ({
  mixed: { label: roleLabels.value.mixed, cls: 'primary' },
  compute: { label: roleLabels.value.compute, cls: 'neutral' },
  storage: { label: roleLabels.value.storage, cls: 'warning' },
  NODE_ROLE_UNSPECIFIED: { label: roleLabels.value.mixed, cls: 'primary' },
  NODE_ROLE_MIXED: { label: roleLabels.value.mixed, cls: 'primary' },
  NODE_ROLE_COMPUTE: { label: roleLabels.value.compute, cls: 'neutral' },
  NODE_ROLE_STORAGE: { label: roleLabels.value.storage, cls: 'warning' },
}));

const serviceMap = {
  running: { label: 'Работает', cls: 'success' },
  starting: { label: 'Запуск...', cls: 'warning' },
  stopped: { label: 'Остановлен', cls: 'muted' },
  failed: { label: 'Ошибка', cls: 'danger' },
  unknown: { label: 'Неизвестно', cls: 'muted' },
  SERVICE_STATUS_UNSPECIFIED: { label: 'Неизвестно', cls: 'muted' },
  SERVICE_STATUS_STARTING: { label: 'Запуск...', cls: 'warning' },
  SERVICE_STATUS_RUNNING: { label: 'Работает', cls: 'success' },
  SERVICE_STATUS_STOPPED: { label: 'Остановлен', cls: 'muted' },
  SERVICE_STATUS_FAILED: { label: 'Ошибка', cls: 'danger' },
};

const map = computed(() => {
  if (props.type === 'node') return nodeMap;
  if (props.type === 'role') return roleMap.value;
  if (props.type === 'service') return serviceMap;
  if (props.type === 'platform_access') return platformAccessMap;
  return instanceMap;
});

// Convert numeric status to proto enum string if needed
const statusKey = computed(() => {
  const status = props.status;
  if (typeof status === 'number' || /^\d+$/.test(String(status))) {
    const numStatus = Number(status);
    if (props.type === 'platform_access') {
      const numPlatformMap = {
        0: 'REQUEST_STATUS_UNSPECIFIED',
        1: 'REQUEST_STATUS_PENDING',
        2: 'REQUEST_STATUS_APPROVED',
        3: 'REQUEST_STATUS_APPROVED',
        4: 'REQUEST_STATUS_REJECTED',
      };
      return numPlatformMap[numStatus] || 'REQUEST_STATUS_UNSPECIFIED';
    }
    if (props.type === 'role') {
      const numRoleMap = {
        0: 'NODE_ROLE_UNSPECIFIED',
        1: 'NODE_ROLE_MIXED',
        2: 'NODE_ROLE_COMPUTE',
        3: 'NODE_ROLE_STORAGE',
      };
      return numRoleMap[numStatus] || 'NODE_ROLE_UNSPECIFIED';
    }
    if (props.type === 'service') {
      const numServiceMap = {
        0: 'SERVICE_STATUS_UNSPECIFIED',
        1: 'SERVICE_STATUS_STARTING',
        2: 'SERVICE_STATUS_RUNNING',
        3: 'SERVICE_STATUS_STOPPED',
        4: 'SERVICE_STATUS_FAILED',
      };
      return numServiceMap[numStatus] || 'SERVICE_STATUS_UNSPECIFIED';
    }
    if (props.type === 'node') {
      const nodeStatusMap = {
        0: 'NODE_STATUS_UNSPECIFIED',
        1: 'NODE_STATUS_UNAUTHORIZED',
        2: 'NODE_STATUS_ONLINE',
        3: 'NODE_STATUS_OFFLINE',
        4: 'NODE_STATUS_MAINTENANCE',
      };
      return nodeStatusMap[numStatus] || 'NODE_STATUS_UNSPECIFIED';
    }
    const instanceStatusMap = {
      0: 'INSTANCE_STATUS_UNSPECIFIED',
      1: 'INSTANCE_STATUS_STARTING',
      2: 'INSTANCE_STATUS_RUNNING',
      3: 'INSTANCE_STATUS_STOPPING',
      4: 'INSTANCE_STATUS_STOPPED',
      5: 'INSTANCE_STATUS_CRASHED',
    };
    return instanceStatusMap[numStatus] || 'INSTANCE_STATUS_UNSPECIFIED';
  }
  return status;
});

const label = computed(() => map.value[statusKey.value]?.label ?? String(props.status ?? '—'));
const statusClass = computed(() => `badge-${map.value[statusKey.value]?.cls ?? 'muted'}`);
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
</style>
