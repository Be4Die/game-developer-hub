<template>
  <div class="overview tab-fade-in">
    <div class="overview-header">
      <h1>{{ t('servers.overviewTitle') }}</h1>
    </div>

    <!-- Ошибка загрузки -->
    <div v-if="error" class="error-banner">
      <AlertCircle class="icon-sm" /> {{ t('common.error') }}:
      {{ error }}
      <button class="btn-outline btn-sm" @click="fetchAll">{{ t('common.refresh') }}</button>
    </div>

    <!-- Карточки-сводки -->
    <div class="summary-grid">
      <div class="summary-card">
        <span class="summary-label">{{ t('servers.tabs.builds') }}</span>
        <span class="summary-value">{{
          loading ? '...' : builds.length
        }}</span>
      </div>
      <div class="summary-card">
        <span class="summary-label">{{ t('servers.runningInstances') }}</span>
        <span class="summary-value"
          >{{ loading ? '...' : runningCount
          }}<span class="summary-sub"> / {{ instances.length }}</span></span
        >
      </div>
      <div class="summary-card">
        <span class="summary-label">{{ t('servers.onlinePlayers') }}</span>
        <span class="summary-value">{{
          loading ? '...' : totalPlayers
        }}</span>
      </div>
      <div class="summary-card">
        <span class="summary-label">{{ t('servers.tabs.nodes') }}</span>
        <span class="summary-value"
          >{{ loading ? '...' : onlineNodes
          }}<span class="summary-sub"> / {{ nodes.length }}</span></span
        >
      </div>
      <div class="summary-card">
        <span class="summary-label">{{ t('servers.inQueue') }}</span>
        <span class="summary-value">{{ loading ? '...' : queueCount }}</span>
      </div>
    </div>

    <!-- Политика оркестрации -->
    <OrchestrationPolicyEditor
      :game-id="gameId"
      :builds="builds"
      :nodes="nodes"
    />

    <!-- Быстрые действия -->
    <div class="quick-actions">
      <router-link
        :to="`/projects/${gameId}/servers/builds`"
        class="action-card"
      >
        <Upload class="action-icon" />
        <div>
          <strong>{{ t('servers.uploadBuild') }}</strong>
          <p>{{ t('servers.uploadBuildDesc') }}</p>
        </div>
      </router-link>
      <router-link
        :to="`/projects/${gameId}/servers/instances`"
        class="action-card"
      >
        <Play class="action-icon" />
        <div>
          <strong>{{ t('servers.startInstance') }}</strong>
          <p>{{ t('servers.startInstanceDesc') }}</p>
        </div>
      </router-link>
    </div>

    <!-- Последние инстансы -->
    <div class="recent-section">
      <div class="section-header">
        <h2>{{ t('servers.recentInstances') }}</h2>
        <router-link
          :to="`/projects/${gameId}/servers/instances`"
          class="link"
          >{{ t('servers.allInstances') }} →</router-link
        >
      </div>
      <div class="recent-table-wrap">
        <table class="recent-table" v-if="instances.length">
          <thead>
            <tr>
              <th>{{ t('common.name') }}</th>
              <th>{{ t('common.version') }}</th>
              <th>{{ t('common.status') }}</th>
              <th>{{ t('servers.onlinePlayers') }}</th>
              <th>{{ t('common.created') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="inst in instances.slice(0, 5)"
              :key="inst.id"
              @click="
                $router.push(
                  `/projects/${gameId}/servers/instances/${inst.id}`
                )
              "
              class="clickable-row"
            >
              <td class="cell-name">
                {{ inst.name || `Instance #${inst.id}` }}
              </td>
              <td>
                <code>{{ inst.build_version }}</code>
              </td>
              <td><StatusBadge :status="inst.status" type="instance" /></td>
              <td>{{ inst.player_count ?? 0 }} / {{ inst.max_players }}</td>
              <td class="cell-muted">
                {{
                  inst.started_at ? formatDate(inst.started_at) : '—'
                }}
              </td>
            </tr>
          </tbody>
        </table>
        <div v-else-if="!loading" class="empty-state">
          {{ t('servers.noInstances') }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { Upload, Play, AlertCircle } from 'lucide-vue-next';
import { StatusBadge } from '@/shared/ui';
import { OrchestrationPolicyEditor } from '@/features/edit-orchestration-policy';
import { listServerBuilds } from '@/entities/build';
import { listInstances } from '@/entities/instance';
import { listNodes } from '@/entities/node';
import { getQueueCount } from '@/entities/policy';
import { formatDate } from '@/shared/lib';

const { t } = useI18n();

const props = defineProps({
  gameId: { type: [String, Number], required: true },
});


const builds = ref([]);
const instances = ref([]);
const nodes = ref([]);
const loading = ref(true);
const error = ref(null);
const queueCount = ref(0);

const runningCount = computed(
  () => instances.value.filter((i) => i.status === 'running').length
);
const totalPlayers = computed(() =>
  instances.value.reduce((sum, i) => sum + (i.player_count ?? 0), 0)
);
const onlineNodes = computed(
  () =>
    nodes.value.filter(
      (n) => n.status === 'online' || n.status === 'NODE_STATUS_ONLINE'
    ).length
);

async function fetchQueueCount() {
  try {
    const count = await getQueueCount(props.gameId);
    queueCount.value = Number(count);
  } catch (e) {
    queueCount.value = 0;
  }
}

async function fetchAll() {
  loading.value = true;
  error.value = null;
  try {
    const [b, i, n] = await Promise.all([
      listServerBuilds(props.gameId).catch(() => []),
      listInstances(props.gameId).catch(() => []),
      listNodes().catch(() => []),
    ]);
    builds.value = b;
    instances.value = i;
    nodes.value = n;
    await fetchQueueCount();
  } catch (e) {
    error.value = e.message;
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  fetchAll();
});
</script>

<style scoped>
.overview {
  display: flex;
  flex-direction: column;
  gap: 24px;
}
.overview-header h1 {
  margin: 0;
  font-size: 1.5rem;
}

.error-banner {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: var(--danger-light);
  color: var(--danger);
  border-radius: var(--radius-md);
  font-size: 0.85rem;
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: 16px;
}

.summary-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  padding: 18px 20px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.summary-label {
  font-size: 0.8rem;
  color: var(--text-muted);
  font-weight: 500;
}

.summary-value {
  font-size: 1.6rem;
  font-weight: 700;
  color: var(--text-main);
}

.summary-sub {
  font-size: 1rem;
  font-weight: 400;
  color: var(--text-muted);
}

.quick-actions {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 16px;
}

.action-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  padding: 18px 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  text-decoration: none;
  color: inherit;
  transition: all 0.2s ease;
}

.action-card:hover {
  border-color: var(--primary);
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.action-icon {
  width: 24px;
  height: 24px;
  color: var(--primary);
}

.action-card strong {
  display: block;
  font-size: 0.95rem;
  margin-bottom: 2px;
}

.action-card p {
  margin: 0;
  font-size: 0.8rem;
  color: var(--text-muted);
}

.recent-section {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  padding: 20px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.section-header h2 {
  margin: 0;
  font-size: 1.1rem;
}

.link {
  color: var(--primary);
  font-size: 0.85rem;
  text-decoration: none;
  font-weight: 500;
}

.recent-table-wrap {
  overflow-x: auto;
}

.recent-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.88rem;
}

.recent-table th {
  text-align: left;
  padding: 8px 12px;
  color: var(--text-muted);
  font-weight: 500;
  border-bottom: 1px solid var(--border);
  font-size: 0.8rem;
}

.recent-table td {
  padding: 12px;
  border-bottom: 1px solid var(--border);
}

.clickable-row {
  cursor: pointer;
  transition: background 0.15s;
}

.clickable-row:hover {
  background: var(--bg-hover);
}

.cell-name {
  font-weight: 600;
}

.cell-muted {
  color: var(--text-muted);
  font-size: 0.82rem;
}

.empty-state {
  padding: 32px 0;
  text-align: center;
  color: var(--text-muted);
  font-size: 0.9rem;
}
</style>
