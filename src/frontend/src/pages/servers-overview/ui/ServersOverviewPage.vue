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
        <span class="summary-value">{{ loading ? '...' : builds.length }}</span>
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
        <span class="summary-value">{{ loading ? '...' : totalPlayers }}</span>
      </div>
      <div class="summary-card">
        <span class="summary-label">{{ t('servers.tabs.nodes') }}</span>
        <span class="summary-value"
          >{{ loading ? '...' : onlineNodes
          }}<span class="summary-sub"> / {{ nodes.length }}</span></span
        >
        <span v-if="platformNodesCount > 0" class="summary-extra">
          ⭐ {{ platformNodesCount }} платформенных
        </span>
      </div>
      <div class="summary-card">
        <span class="summary-label">{{ t('servers.inQueue') }}</span>
        <span class="summary-value">{{ loading ? '...' : queueCount }}</span>
      </div>
    </div>

    <!-- Блок доступа к серверам платформы -->
    <div class="platform-access-card" :class="platformCardClass">
      <div class="platform-access-header">
        <div class="platform-header-left">
          <div class="platform-icon-box">
            <Server class="icon-md" />
          </div>
          <div>
            <div class="platform-title-row">
              <h3>Серверы платформы GDH</h3>
              <StatusBadge
                v-if="platformAccess"
                :status="platformAccess.status"
                type="platform_access"
              />
              <span v-else class="platform-badge-available">Бесплатный пул</span>
            </div>
            <p v-if="!platformAccess" class="platform-desc">
              GDH предоставляет пул производительных серверов для поддержки начинающих проектов. Запросите доступ к серверам платформы, чтобы разворачивать сессии без собственного сервера.
            </p>
            <p v-else-if="isPlatformPending" class="platform-desc">
              Ваша заявка на доступ к серверам платформы находится на рассмотрении модераторами. После подтверждения ноды платформы автоматически появятся в вашем проекте.
            </p>
            <div v-else-if="isPlatformApproved" class="platform-approved-info">
              <p class="platform-desc">
                Проекту открыт доступ к серверам платформы. Платформенные ноды подключены к общему пулу вашего проекта и готовы к запуску сессий.
              </p>
              <div v-if="platformAccess.moderatorComment || platformAccess.moderator_comment" class="moderator-comment-pill">
                <MessageSquare class="icon-xxs" />
                <span>Комментарий модератора: «{{ platformAccess.moderatorComment || platformAccess.moderator_comment }}»</span>
              </div>
            </div>
            <p v-else-if="isPlatformRejected" class="platform-desc">
              Заявка на доступ к серверам платформы отклонена модератором.
              <span v-if="platformAccess.rejectionReason || platformAccess.rejection_reason" class="rejection-text">
                Причина: «{{ platformAccess.rejectionReason || platformAccess.rejection_reason }}»
              </span>
            </p>
          </div>
        </div>

        <div class="platform-header-actions">
          <button
            v-if="!platformAccess"
            class="btn-primary"
            @click="openRequestModal"
          >
            <Sparkles class="icon-sm" />
            <span>Запросить доступ к серверам</span>
          </button>
          <button
            v-else-if="isPlatformRejected"
            class="btn-outline"
            @click="openRequestModal"
          >
            <span>Подать заявку повторно</span>
          </button>
          <div v-else-if="isPlatformApproved" class="quota-badges-grid">
            <div class="quota-pill" title="Максимум одновременно работающих инстансов">
              <span class="quota-label">Инстансы:</span>
              <span class="quota-value">{{ platformAccess.maxInstances || platformAccess.max_instances || 2 }}</span>
            </div>
            <div class="quota-pill" title="Лимит CPU на весь проект">
              <span class="quota-label">Всего CPU:</span>
              <span class="quota-value">{{ formatCpu(platformAccess.maxTotalCpuMillis || platformAccess.max_total_cpu_millis) }}</span>
            </div>
            <div class="quota-pill" title="Лимит RAM на весь проект">
              <span class="quota-label">Всего RAM:</span>
              <span class="quota-value">{{ formatMemory(platformAccess.maxTotalMemoryMb || platformAccess.max_total_memory_mb) }}</span>
            </div>
            <div
              v-if="(platformAccess.maxInstanceCpuMillis || platformAccess.max_instance_cpu_millis) || (platformAccess.maxInstanceMemoryMb || platformAccess.max_instance_memory_mb)"
              class="quota-pill"
              title="Лимиты на один инстанс"
            >
              <span class="quota-label">На инстанс:</span>
              <span class="quota-value">
                {{ formatCpu(platformAccess.maxInstanceCpuMillis || platformAccess.max_instance_cpu_millis) }} / {{ formatMemory(platformAccess.maxInstanceMemoryMb || platformAccess.max_instance_memory_mb) }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Политика оркестрации -->
    <OrchestrationPolicyEditor :game-id="gameId" :builds="builds" :nodes="nodes" />

    <!-- Быстрые действия -->
    <div class="quick-actions">
      <router-link :to="`/projects/${gameId}/servers/builds`" class="action-card">
        <Upload class="action-icon" />
        <div>
          <strong>{{ t('servers.uploadBuild') }}</strong>
          <p>{{ t('servers.uploadBuildDesc') }}</p>
        </div>
      </router-link>
      <router-link :to="`/projects/${gameId}/servers/instances`" class="action-card">
        <Play class="action-icon" />
        <div>
          <strong>{{ t('servers.startInstance') }}</strong>
          <p>{{ t('servers.startInstanceDesc') }}</p>
        </div>
      </router-link>
    </div>

    <!-- Список доступных серверов проекта -->
    <div v-if="nodes.length" class="recent-section">
      <div class="section-header">
        <h2>Серверные ноды проекта</h2>
        <router-link to="/nodes" class="link">Все ноды платформы →</router-link>
      </div>
      <div class="recent-table-wrap">
        <table class="recent-table">
          <thead>
            <tr>
              <th>Адрес</th>
              <th>Тип сервера</th>
              <th>Регион</th>
              <th>Режим</th>
              <th>Статус</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="node in nodes"
              :key="node.id"
              class="clickable-row"
              @click="$router.push(`/nodes/${node.id}`)"
            >
              <td class="cell-name">
                {{ node.address }}
              </td>
              <td>
                <span
                  v-if="node.is_platform"
                  class="platform-node-badge"
                  title="Общедоступная нода платформы с открытым доступом"
                >
                  ⭐ Платформа • Доступ открыт
                </span>
                <span v-else class="custom-node-badge">
                  Собственная нода
                </span>
              </td>
              <td class="cell-muted">{{ node.region || '—' }}</td>
              <td><StatusBadge :status="node.role || 'mixed'" type="role" /></td>
              <td><StatusBadge :status="node.status" type="node" /></td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>


    <!-- Последние инстансы -->
    <div class="recent-section">
      <div class="section-header">
        <h2>{{ t('servers.recentInstances') }}</h2>
        <router-link :to="`/projects/${gameId}/servers/instances`" class="link"
          >{{ t('servers.allInstances') }} →</router-link
        >
      </div>
      <div class="recent-table-wrap">
        <table v-if="instances.length" class="recent-table">
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
              <td>{{ inst.player_count ?? 0 }} / {{ inst.max_players }}</td>
              <td class="cell-muted">
                {{ inst.started_at ? formatDate(inst.started_at) : '—' }}
              </td>
            </tr>
          </tbody>
        </table>
        <div v-else-if="!loading" class="empty-state">
          {{ t('servers.noInstances') }}
        </div>
      </div>
    </div>

    <!-- Модальное окно запроса доступа к платформенным нодам -->
    <div v-if="showRequestModal" class="modal-overlay" @click.self="showRequestModal = false">
      <div class="modal card quota-modal">
        <div class="modal-header">
          <h3>Запрос доступа к серверам платформы</h3>
          <p class="modal-sub">
            Укажите желаемую квоту ресурсов и расскажите, зачем вашему проекту бесплатные мощности GDH.
          </p>
        </div>

        <div class="modal-body">
          <div class="quota-fields-grid">
            <!-- Количество инстансов -->
            <div class="form-group mb-12">
              <label>Желаемое число инстансов *</label>
              <input
                v-model.number="requestMaxInstances"
                type="number"
                min="1"
                max="20"
                class="form-input"
              />
              <span class="field-hint">Максимум одновременно работающих игровых серверов проекта</span>
            </div>

            <!-- Суммарные ресурсы проекта -->
            <div class="quota-block">
              <h4 class="quota-block-title">Суммарный лимит на весь проект</h4>
              <div class="quota-inputs-row">
                <div class="form-group flex-1">
                  <div class="label-with-toggle">
                    <label>CPU (ядер)</label>
                    <label class="toggle-label">
                      <input type="checkbox" v-model="requestUnlimitedTotalCpu" />
                      <span>Без огр.</span>
                    </label>
                  </div>
                  <input
                    v-if="!requestUnlimitedTotalCpu"
                    v-model.number="requestTotalCpuCores"
                    type="number"
                    step="0.5"
                    min="0.5"
                    max="64"
                    class="form-input"
                    placeholder="Например, 2.0"
                  />
                  <div v-else class="unlimited-placeholder">∞ Без ограничений</div>
                </div>

                <div class="form-group flex-1">
                  <div class="label-with-toggle">
                    <label>RAM (МБ)</label>
                    <label class="toggle-label">
                      <input type="checkbox" v-model="requestUnlimitedTotalRam" />
                      <span>Без огр.</span>
                    </label>
                  </div>
                  <input
                    v-if="!requestUnlimitedTotalRam"
                    v-model.number="requestTotalRamMb"
                    type="number"
                    step="256"
                    min="256"
                    max="131072"
                    class="form-input"
                    placeholder="Например, 4096"
                  />
                  <div v-else class="unlimited-placeholder">∞ Без ограничений</div>
                </div>
              </div>
            </div>

            <!-- Ресурсы на 1 инстанс -->
            <div class="quota-block">
              <h4 class="quota-block-title">Лимит на 1 инстанс (сервер)</h4>
              <div class="quota-inputs-row">
                <div class="form-group flex-1">
                  <div class="label-with-toggle">
                    <label>CPU на инстанс</label>
                    <label class="toggle-label">
                      <input type="checkbox" v-model="requestUnlimitedInstanceCpu" />
                      <span>Без огр.</span>
                    </label>
                  </div>
                  <input
                    v-if="!requestUnlimitedInstanceCpu"
                    v-model.number="requestInstanceCpuCores"
                    type="number"
                    step="0.5"
                    min="0.5"
                    max="32"
                    class="form-input"
                    placeholder="Например, 1.0"
                  />
                  <div v-else class="unlimited-placeholder">∞ Без ограничений</div>
                </div>

                <div class="form-group flex-1">
                  <div class="label-with-toggle">
                    <label>RAM на инстанс (МБ)</label>
                    <label class="toggle-label">
                      <input type="checkbox" v-model="requestUnlimitedInstanceRam" />
                      <span>Без огр.</span>
                    </label>
                  </div>
                  <input
                    v-if="!requestUnlimitedInstanceRam"
                    v-model.number="requestInstanceRamMb"
                    type="number"
                    step="256"
                    min="256"
                    max="32768"
                    class="form-input"
                    placeholder="Например, 1024"
                  />
                  <div v-else class="unlimited-placeholder">∞ Без ограничений</div>
                </div>
              </div>
            </div>
          </div>

          <div class="form-group">
            <label>Сопроводительное письмо / Обоснование запроса *</label>
            <textarea
              v-model="requestReason"
              rows="3"
              class="form-textarea"
              placeholder="Расскажите о проекте (жанр, движок, ожидаемое число игроков онлайн, почему нужны серверы платформы)..."
            ></textarea>
          </div>
          <div v-if="requestError" class="modal-error">
            {{ requestError }}
          </div>
        </div>

        <div class="modal-actions">
          <button
            class="btn-primary"
            :disabled="submittingRequest || !requestReason.trim() || requestMaxInstances < 1"
            @click="submitPlatformRequest"
          >
            <Send class="icon-xs" />
            <span>{{ submittingRequest ? 'Отправка...' : 'Отправить заявку' }}</span>
          </button>
          <button class="btn-outline" :disabled="submittingRequest" @click="showRequestModal = false">
            Отмена
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { Upload, Play, AlertCircle, Server, Sparkles, Send, MessageSquare } from 'lucide-vue-next';
import { StatusBadge } from '@/shared/ui';
import { OrchestrationPolicyEditor } from '@/features/edit-orchestration-policy';
import { listServerBuilds } from '@/entities/build';
import { listInstances } from '@/entities/instance';
import { listNodes } from '@/entities/node';
import { moderationApi, formatCpu, formatMemory } from '@/entities/moderation';
import { getQueueCount } from '@/entities/policy';
import { formatDate, showToast } from '@/shared/lib';

const { t } = useI18n();

const props = defineProps({
  gameId: { type: [String, Number], required: true },
});

const builds = ref([]);
const instances = ref([]);
const nodes = ref([]);
const platformAccess = ref(null);
const loading = ref(true);
const error = ref(null);
const queueCount = ref(0);

const showRequestModal = ref(false);
const requestReason = ref('');
const requestMaxInstances = ref(2);
const requestUnlimitedTotalCpu = ref(true);
const requestTotalCpuCores = ref(2.0);
const requestUnlimitedTotalRam = ref(true);
const requestTotalRamMb = ref(4096);
const requestUnlimitedInstanceCpu = ref(true);
const requestInstanceCpuCores = ref(1.0);
const requestUnlimitedInstanceRam = ref(true);
const requestInstanceRamMb = ref(1024);

const submittingRequest = ref(false);
const requestError = ref(null);

const runningCount = computed(() => instances.value.filter((i) => i.status === 'running').length);
const totalPlayers = computed(() =>
  instances.value.reduce((sum, i) => sum + (i.player_count ?? 0), 0)
);
const onlineNodes = computed(
  () => nodes.value.filter((n) => n.status === 'online' || n.status === 'NODE_STATUS_ONLINE').length
);
const platformNodesCount = computed(
  () => nodes.value.filter((n) => n.is_platform).length
);

const isPlatformPending = computed(() => {
  const s = platformAccess.value?.status;
  return s === 'REQUEST_STATUS_PENDING' || s === 'PLATFORM_ACCESS_STATUS_PENDING' || s === 'pending' || s === 1 || s === '1';
});

const isPlatformApproved = computed(() => {
  const s = platformAccess.value?.status;
  return (
    s === 'REQUEST_STATUS_APPROVED' ||
    s === 'PLATFORM_ACCESS_STATUS_APPROVED' ||
    s === 'approved' ||
    s === 2 ||
    s === '2' ||
    s === 3 ||
    s === '3'
  );
});

const isPlatformRejected = computed(() => {
  const s = platformAccess.value?.status;
  return (
    s === 'REQUEST_STATUS_REJECTED' ||
    s === 'PLATFORM_ACCESS_STATUS_REJECTED' ||
    s === 'rejected' ||
    s === 4 ||
    s === '4'
  );
});

const platformCardClass = computed(() => {
  if (isPlatformApproved.value) return 'card-approved';
  if (isPlatformPending.value) return 'card-pending';
  if (isPlatformRejected.value) return 'card-rejected';
  return 'card-available';
});

function openRequestModal() {
  requestReason.value = '';
  requestError.value = null;
  requestMaxInstances.value = 2;
  requestUnlimitedTotalCpu.value = true;
  requestTotalCpuCores.value = 2.0;
  requestUnlimitedTotalRam.value = true;
  requestTotalRamMb.value = 4096;
  requestUnlimitedInstanceCpu.value = true;
  requestInstanceCpuCores.value = 1.0;
  requestUnlimitedInstanceRam.value = true;
  requestInstanceRamMb.value = 1024;
  showRequestModal.value = true;
}

async function submitPlatformRequest() {
  if (!requestReason.value.trim()) return;
  submittingRequest.value = true;
  requestError.value = null;
  try {
    const payload = {
      reason: requestReason.value.trim(),
      maxInstances: Number(requestMaxInstances.value) || 2,
      maxTotalCpuMillis: requestUnlimitedTotalCpu.value ? 0 : Math.round((Number(requestTotalCpuCores.value) || 0) * 1000),
      maxTotalMemoryMb: requestUnlimitedTotalRam.value ? 0 : Number(requestTotalRamMb.value) || 0,
      maxInstanceCpuMillis: requestUnlimitedInstanceCpu.value ? 0 : Math.round((Number(requestInstanceCpuCores.value) || 0) * 1000),
      maxInstanceMemoryMb: requestUnlimitedInstanceRam.value ? 0 : Number(requestInstanceRamMb.value) || 0,
    };
    const res = await moderationApi.submitServerAccess(props.gameId, payload);
    platformAccess.value = res.request;
    showRequestModal.value = false;
    showToast('Заявка на доступ к серверам платформы успешно отправлена', 'success');
    nodes.value = await listNodes(null, props.gameId).catch(() => nodes.value);
  } catch (e) {
    requestError.value = e.response?.data?.message ?? 'Ошибка отправки заявки';
  } finally {
    submittingRequest.value = false;
  }
}

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
    const [b, i, n, paRes] = await Promise.all([
      listServerBuilds(props.gameId).catch(() => []),
      listInstances(props.gameId).catch(() => []),
      listNodes(null, props.gameId).catch(() => []),
      moderationApi.getServerAccess(props.gameId).catch(() => null),
    ]);
    builds.value = b;
    instances.value = i;
    nodes.value = n;
    platformAccess.value = paRes?.request || null;
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

.summary-extra {
  font-size: 0.76rem;
  color: #eab308;
  font-weight: 600;
  margin-top: 2px;
}

.platform-access-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  padding: 20px;
  position: relative;
  overflow: hidden;
  transition: all 0.2s ease;
}

.platform-access-card.card-available {
  border-color: rgba(88, 166, 255, 0.4);
  background: linear-gradient(180deg, rgba(88, 166, 255, 0.05) 0%, var(--bg-card) 100%);
}

.platform-access-card.card-pending {
  border-color: rgba(234, 179, 8, 0.4);
  background: linear-gradient(180deg, rgba(234, 179, 8, 0.05) 0%, var(--bg-card) 100%);
}

.platform-access-card.card-approved {
  border-color: rgba(63, 185, 80, 0.4);
  background: linear-gradient(180deg, rgba(63, 185, 80, 0.05) 0%, var(--bg-card) 100%);
}

.platform-access-card.card-rejected {
  border-color: rgba(248, 81, 73, 0.4);
  background: linear-gradient(180deg, rgba(248, 81, 73, 0.05) 0%, var(--bg-card) 100%);
}

.platform-access-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 20px;
  flex-wrap: wrap;
}

.platform-header-left {
  display: flex;
  align-items: flex-start;
  gap: 16px;
  max-width: 750px;
}

.platform-icon-box {
  width: 44px;
  height: 44px;
  border-radius: var(--radius-md);
  background: var(--bg-tertiary, #21262d);
  border: 1px solid var(--border);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--primary, #58a6ff);
  flex-shrink: 0;
}

.card-approved .platform-icon-box {
  color: #3fb950;
  border-color: rgba(63, 185, 80, 0.3);
}

.card-pending .platform-icon-box {
  color: #eab308;
  border-color: rgba(234, 179, 8, 0.3);
}

.card-rejected .platform-icon-box {
  color: #f85149;
  border-color: rgba(248, 81, 73, 0.3);
}

.platform-title-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 6px;
}

.platform-title-row h3 {
  margin: 0;
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--text-main);
}

.platform-badge-available {
  font-size: 0.72rem;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 10px;
  background: rgba(88, 166, 255, 0.15);
  color: #58a6ff;
  border: 1px solid rgba(88, 166, 255, 0.3);
}

.platform-desc {
  margin: 0;
  font-size: 0.86rem;
  color: var(--text-muted);
  line-height: 1.45;
}

.rejection-text {
  display: block;
  margin-top: 4px;
  color: #f85149;
  font-weight: 500;
}

.platform-header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.quota-pill {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 14px;
  background: rgba(63, 185, 80, 0.12);
  border: 1px solid rgba(63, 185, 80, 0.3);
  border-radius: var(--radius-md);
}

.quota-label {
  font-size: 0.82rem;
  color: var(--text-muted);
}

.quota-value {
  font-size: 1.1rem;
  font-weight: 700;
  color: #3fb950;
}

.platform-node-badge {
  display: inline-flex;
  align-items: center;
  font-size: 0.76rem;
  font-weight: 600;
  padding: 3px 8px;
  border-radius: 6px;
  background: rgba(234, 179, 8, 0.15);
  color: #eab308;
  border: 1px solid rgba(234, 179, 8, 0.35);
}

.custom-node-badge {
  display: inline-flex;
  align-items: center;
  font-size: 0.76rem;
  font-weight: 500;
  padding: 3px 8px;
  border-radius: 6px;
  background: var(--bg-tertiary);
  color: var(--text-muted);
  border: 1px solid var(--border);
}

/* Модальное окно */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
}

.modal {
  width: 100%;
  max-width: 520px;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  padding: 24px;
}

.modal.quota-modal {
  max-width: 580px;
}

.modal-header h3 {
  margin: 0 0 6px;
  font-size: 1.15rem;
  color: var(--text-main);
}

.modal-sub {
  margin: 0 0 16px;
  font-size: 0.85rem;
  color: var(--text-muted);
}

.quota-badges-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.quota-block {
  background: var(--bg-tertiary);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  padding: 12px 14px;
  margin-bottom: 12px;
}

.quota-block-title {
  margin: 0 0 10px;
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--text-main);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.quota-inputs-row {
  display: flex;
  gap: 12px;
}

.flex-1 {
  flex: 1;
}

.mb-12 {
  margin-bottom: 12px;
}

.label-with-toggle {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
}

.label-with-toggle label {
  margin-bottom: 0 !important;
}

.toggle-label {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 0.75rem;
  color: var(--text-muted);
  cursor: pointer;
  user-select: none;
}

.toggle-label input {
  cursor: pointer;
}

.unlimited-placeholder {
  padding: 8px 12px;
  background: var(--bg-app);
  border: 1px dashed var(--border);
  border-radius: var(--radius-md);
  font-size: 0.82rem;
  color: var(--text-muted);
  text-align: center;
}

.moderator-comment-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin-top: 8px;
  padding: 4px 10px;
  background: rgba(88, 166, 255, 0.1);
  border: 1px solid rgba(88, 166, 255, 0.25);
  border-radius: var(--radius-sm);
  font-size: 0.8rem;
  color: #58a6ff;
}

.field-hint {
  display: block;
  margin-top: 4px;
  font-size: 0.74rem;
  color: var(--text-muted);
}

.form-input {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  background: var(--bg-app);
  color: var(--text-main);
  font-size: 0.88rem;
  font-family: inherit;
  outline: none;
  box-sizing: border-box;
}

.form-input:focus {
  border-color: var(--primary);
}

.form-textarea {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  background: var(--bg-app);
  color: var(--text-main);
  font-size: 0.88rem;
  font-family: inherit;
  outline: none;
  box-sizing: border-box;
  resize: vertical;
}

.form-textarea:focus {
  border-color: var(--primary);
}

.form-group label {
  display: block;
  font-size: 0.82rem;
  font-weight: 600;
  color: var(--text-muted);
  margin-bottom: 6px;
}

.modal-error {
  margin-top: 10px;
  padding: 8px 12px;
  background: var(--danger-light);
  color: var(--danger);
  border-radius: var(--radius-md);
  font-size: 0.82rem;
}

.modal-actions {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
  margin-top: 20px;
}
</style>

