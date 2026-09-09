<template>
  <div class="node-detail-page">
    <!-- Шапка страницы: адрес, статус, переключатель режимов и кнопка удалить в одну линию -->
    <div class="page-header">
      <div class="header-main">
        <h1 class="node-title">{{ node.address || 'Нода' }}</h1>
        <StatusBadge v-if="node.status" :status="node.status" type="node" />
      </div>

      <div v-if="!isUnauthorized" class="header-actions">
        <!-- Переключатель режима ноды -->
        <div class="role-selector-container">
          <span class="role-selector-label">Режим:</span>
          <div class="role-segmented-control">
            <button
              class="role-pill-btn"
              :class="{ active: currentRole === 'mixed' }"
              :disabled="updatingRole || isUnauthorized"
              title="Compute + Storage (Игровые комнаты и базы данных)"
              @click="setRole('mixed')"
            >
              Mixed
            </button>
            <button
              class="role-pill-btn"
              :class="{ active: currentRole === 'compute' }"
              :disabled="updatingRole || isUnauthorized"
              title="Только игровые комнаты (вычисления)"
              @click="setRole('compute')"
            >
              Compute
            </button>
            <button
              class="role-pill-btn"
              :class="{ active: currentRole === 'storage' }"
              :disabled="updatingRole || isUnauthorized"
              title="Выделенное хранилище (БД, кэш, тома)"
              @click="setRole('storage')"
            >
              Storage
            </button>
          </div>
        </div>

        <button
          v-if="!isUnauthorized"
          class="btn-delete-node"
          title="Удалить ноду из кластера"
          @click="showDeleteConfirm = true"
        >
          <Trash2 class="icon-sm" />
          <span>Удалить ноду</span>
        </button>
      </div>
    </div>

    <!-- Ошибка загрузки -->
    <div v-if="error" class="error-banner">
      <AlertCircle class="icon-sm" />
      <span>{{ error }}</span>
      <button class="btn-retry" @click="fetchNode">Повторить</button>
    </div>

    <template v-if="!error && node.id">
      <!-- Верхний обзорный блок: Информация о ноде (слева) и Потребление ресурсов (справа, 2x2) -->
      <div class="overview-grid" :class="{ 'is-unauthorized': isUnauthorized }">
        <!-- Блок информации о ноде -->
        <div class="card specs-card">
          <div class="card-header-simple">
            <h3>Информация о ноде</h3>
            <div class="specs-header-ping">
              <span class="ping-label">Последний пинг:</span>
              <span class="ping-value">{{ node.last_ping_at ? formatDateTime(node.last_ping_at) : '—' }}</span>
            </div>
          </div>
          <div class="specs-grid-2col">
            <div class="spec-col">
              <div class="spec-row">
                <span class="spec-label">ID ноды</span>
                <span class="spec-value"><code>#{{ node.id }}</code></span>
              </div>
              <div class="spec-row">
                <span class="spec-label">Адрес агента</span>
                <span class="spec-value"><code>{{ node.address }}</code></span>
              </div>
              <div class="spec-row">
                <span class="spec-label">Регион</span>
                <span class="spec-value">{{ node.region || 'local' }}</span>
              </div>
              <div class="spec-row">
                <span class="spec-label">Добавлена</span>
                <span class="spec-value">{{ formatDateTime(node.created_at) }}</span>
              </div>
            </div>

            <div class="spec-col">
              <div class="spec-row">
                <span class="spec-label">Процессор</span>
                <span class="spec-value">{{ node.cpu_cores ? node.cpu_cores + ' vCPU' : '—' }}</span>
              </div>
              <div class="spec-row">
                <span class="spec-label">Общая память</span>
                <span class="spec-value">{{ node.total_memory_bytes ? formatBytes(node.total_memory_bytes) : '—' }}</span>
              </div>
              <div class="spec-row">
                <span class="spec-label">Общий диск</span>
                <span class="spec-value">{{ node.total_disk_bytes ? formatBytes(node.total_disk_bytes) : '—' }}</span>
              </div>
              <div class="spec-row">
                <span class="spec-label">Версия агента</span>
                <span class="spec-value">{{ node.agent_version || '—' }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Текущее потребление ресурсов (2x2) -->
        <div v-if="!isUnauthorized" class="card resources-card">
          <div class="card-header-simple">
            <h3>Текущее потребление ресурсов</h3>
          </div>
          <div class="resources-grid-2x2">
            <ResourceUsageCard label="CPU" :value="usage.cpu_usage_percent" type="percent" />
            <ResourceUsageCard
              label="Память"
              :value="usage.memory_used_bytes"
              :max="node.total_memory_bytes"
              type="bytes"
            />
            <ResourceUsageCard
              label="Диск"
              :value="usage.disk_used_bytes"
              :max="node.total_disk_bytes"
              type="bytes"
            />
            <ResourceUsageCard
              label="Сеть"
              :value="usage.network_bytes_per_sec"
              unit=" байт/с"
              type="raw"
            />
          </div>
        </div>
      </div>

      <!-- Форма авторизации для неавторизованных нод -->
      <div v-if="isUnauthorized" class="card auth-card">
        <div class="auth-form">
          <div class="form-group">
            <label>API-ключ ноды *</label>
            <input
              v-model="authToken"
              type="text"
              class="form-input"
              placeholder="dev-api-key-for-local-testing"
              @keyup.enter="submitAuthorize"
            />
          </div>
          <div v-if="authError" class="auth-error">
            {{ authError }}
          </div>
          <button
            class="btn-primary"
            :disabled="!authToken || authorizing"
            @click="submitAuthorize"
          >
            {{ authorizing ? 'Авторизация...' : 'Авторизовать ноду' }}
          </button>
        </div>
      </div>

      <!-- Секции авторизованной ноды -->
      <template v-if="!isUnauthorized">
        <!-- Управляемые сервисы хранения и тома данных (Хранение данных) -->
        <div v-if="currentRole !== 'compute'" class="section-block">
          <div class="section-title-wrap">
            <div>
              <h2>
                Хранение данных
                <span class="count-badge">{{ storageServices.length }}</span>
              </h2>
            </div>
            <button
              class="btn-primary btn-sm"
              @click="showCreateServiceModal = true"
            >
              <Plus class="icon-xs" />
              <span>Развернуть базу данных / том</span>
            </button>
          </div>

          <!-- Веб-панель управления БД (AdminerEvo): компактная строка -->
          <div class="web-ui-row-container">
            <div class="web-ui-row">
              <div class="web-ui-cell-title">
                <div class="service-type-icon-box" style="background-color: rgba(88, 166, 255, 0.15)">
                  <LayoutDashboard class="service-icon" style="color: #58a6ff" />
                </div>
                <div class="service-name-text">
                  <span class="service-name">Панель управления — AdminerEvo</span>
                  <span class="service-subtext">Веб-консоль управления БД</span>
                </div>
              </div>

              <div class="web-ui-cell-status">
                <span v-if="adminerService" class="status-badge success">
                  <span class="status-dot"></span>
                  {{ adminerService.status === 'running' ? 'Работает' : adminerService.status }}
                </span>
                <span v-else class="status-badge muted">
                  <span class="status-dot"></span>
                  Отключена
                </span>
              </div>

              <div class="web-ui-cell-link">
                <a
                  v-if="adminerService && adminerService.status === 'running'"
                  :href="getAdminerUrl(adminerService)"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="adminer-link"
                >
                  <span>Перейти</span>
                  <ExternalLink class="icon-xs" />
                </a>
                <span v-else class="text-muted-link">Недоступно</span>
              </div>

              <div class="web-ui-cell-action">
                <button
                  v-if="adminerService"
                  class="btn-outline btn-sm btn-danger-outline"
                  :disabled="togglingAdminer"
                  @click="toggleAdminer"
                >
                  {{ togglingAdminer ? 'Отключение...' : 'Отключить' }}
                </button>
                <button
                  v-else
                  class="btn-primary btn-sm"
                  :disabled="togglingAdminer"
                  @click="toggleAdminer"
                >
                  {{ togglingAdminer ? 'Подключение...' : 'Подключить' }}
                </button>
              </div>
            </div>
          </div>

          <div v-if="servicesLoading" class="loading-state">
            <div class="spinner-sm"></div>
            <span>Загрузка сервисов...</span>
          </div>

          <div v-else-if="storageServices.length" class="table-container">
            <table class="data-table">
              <thead>
                <tr>
                  <th>Сервис / Том</th>
                  <th>Тип</th>
                  <th>Статус</th>
                  <th>Подключение / Путь</th>
                  <th>Объем на диске</th>
                  <th>Доступ к играм</th>
                  <th v-if="adminerService && adminerService.status === 'running'">Веб-панель</th>
                  <th class="col-actions"></th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="svc in storageServices" :key="svc.id">
                  <!-- Имя сервиса -->
                  <td class="cell-service-name">
                    <div class="service-identity">
                      <div
                        class="service-type-icon-box"
                        :style="{ backgroundColor: getServiceBgColor(svc.service_type) }"
                      >
                        <component
                          :is="getServiceIcon(svc.service_type)"
                          class="service-icon"
                          :style="{ color: getServiceColor(svc.service_type) }"
                        />
                      </div>
                      <div class="service-name-text">
                        <span class="service-name">{{ svc.name }}</span>
                        <span v-if="svc.service_type === 'volume'" class="service-subtext">Том на хосте</span>
                        <span v-else class="service-subtext">Docker-контейнер</span>
                      </div>
                    </div>
                  </td>

                  <!-- Тип сервиса -->
                  <td>
                    <span class="service-type-badge">{{ formatServiceType(svc.service_type) }}</span>
                  </td>

                  <!-- Статус -->
                  <td>
                    <StatusBadge :status="svc.status" type="service" />
                  </td>

                  <!-- Подключение / Путь -->
                  <td>
                    <!-- Для томов данных -->
                    <div v-if="svc.service_type === 'volume'" class="connection-cell">
                      <code class="path-code" :title="svc.volume_path || `/var/lib/gdh/volumes/${svc.name}`">
                        {{ svc.volume_path || `/var/lib/gdh/volumes/${svc.name}` }}
                      </code>
                      <button
                        class="icon-btn"
                        :title="copiedServiceId === svc.id ? 'Скопировано!' : 'Копировать путь к тому'"
                        @click="copyUri(svc.id, svc.volume_path || `/var/lib/gdh/volumes/${svc.name}`)"
                      >
                        <Check v-if="copiedServiceId === svc.id" class="icon-xs text-success" />
                        <Copy v-else class="icon-xs" />
                      </button>
                    </div>

                    <!-- Для баз данных и кэшей -->
                    <div v-else class="connection-cell">
                      <code
                        class="uri-code"
                        :title="svc.connection_uri || `Порт: ${svc.host_port}`"
                      >
                        {{ svc.connection_uri || `Порт: ${svc.host_port}` }}
                      </code>
                      <button
                        v-if="svc.connection_uri"
                        class="icon-btn"
                        :title="copiedServiceId === svc.id ? 'Скопировано!' : 'Копировать строку подключения'"
                        @click="copyUri(svc.id, svc.connection_uri)"
                      >
                        <Check v-if="copiedServiceId === svc.id" class="icon-xs text-success" />
                        <Copy v-else class="icon-xs" />
                      </button>
                    </div>
                  </td>

                  <!-- Объем на диске -->
                  <td>
                    <div class="volume-stat">
                      <span class="volume-size-num">
                        {{ svc.volume_size_bytes ? formatBytes(svc.volume_size_bytes) : '0 Б' }}
                      </span>
                      <span
                        v-if="svc.volume_path"
                        class="volume-path-snippet"
                        :title="svc.volume_path"
                      >
                        {{ svc.volume_path }}
                      </span>
                    </div>
                  </td>

                  <!-- Доступ к играм -->
                  <td>
                    <span v-if="!svc.allowed_game_ids || !svc.allowed_game_ids.length" class="access-badge-all">
                      <Globe class="icon-xs text-muted" />
                      <span>Все игры</span>
                    </span>
                    <div v-else class="game-badges-list">
                      <span
                        v-for="gid in svc.allowed_game_ids"
                        :key="gid"
                        class="game-id-tag"
                      >
                        {{ getGameTitle(gid) }}
                      </span>
                    </div>
                  </td>

                  <!-- Веб-панель (Adminer) -->
                  <td v-if="adminerService && adminerService.status === 'running'">
                    <a
                      v-if="isAdminerSupported(svc) && svc.status === 'running'"
                      :href="getAdminerDbUrl(svc)"
                      target="_blank"
                      rel="noopener noreferrer"
                      class="btn-adminer-open"
                      title="Открыть базу данных в панели управления AdminerEvo"
                    >
                      <ExternalLink class="icon-xs" />
                      <span>Открыть в панели</span>
                    </a>
                    <span v-else class="text-muted">—</span>
                  </td>

                  <!-- Действия -->
                  <td class="col-actions">
                    <div class="actions-cell">
                      <button
                        class="btn-icon-danger"
                        title="Удалить сервис"
                        @click="confirmDeleteService(svc)"
                      >
                        <Trash2 class="icon-xs" />
                      </button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <div v-else class="empty-state-card">
            <FolderPlus class="empty-icon" />
            <p class="empty-title">На этой ноде нет развернутых баз данных или томов</p>
            <p class="empty-desc">
              Вы можете развернуть PostgreSQL, Redis, MySQL или персистентный том для файлов игры.
            </p>
            <button class="btn-primary btn-sm" @click="showCreateServiceModal = true">
              <Plus class="icon-xs" />
              <span>Развернуть базу данных</span>
            </button>
          </div>
        </div>

        <!-- Активные сервера на ноде (Игровые сервера) -->
        <div v-if="currentRole !== 'storage'" class="section-block">
          <div class="section-title-wrap">
            <div>
              <h2>
                Игровые сервера
                <span class="count-badge">{{ activeCount }}</span>
              </h2>
            </div>
          </div>

          <div v-if="instancesLoading" class="loading-state">
            <div class="spinner-sm"></div>
            <span>Загрузка серверов...</span>
          </div>

          <div v-else-if="nodeInstances.length" class="table-container">
            <table class="data-table">
              <thead>
                <tr>
                  <th>Имя инстанса</th>
                  <th>Игра</th>
                  <th>Версия сборки</th>
                  <th>Статус</th>
                  <th>Игроки</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="inst in nodeInstances"
                  :key="inst.id"
                  class="clickable-row"
                  @click="$router.push(`/projects/${inst.game_id}/servers/instances/${inst.id}`)"
                >
                  <td class="cell-name">
                    {{ inst.name || `#${inst.id}` }}
                  </td>
                  <td>
                    <span class="game-tag">{{ getGameTitle(inst.game_id) }}</span>
                  </td>
                  <td>
                    <code>{{ inst.build_version }}</code>
                  </td>
                  <td>
                    <StatusBadge :status="inst.status" type="instance" />
                  </td>
                  <td>
                    <strong>{{ inst.player_count ?? 0 }}</strong> / {{ inst.max_players }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <div v-else class="empty-state-card">
            <Server class="empty-icon" />
            <p class="empty-title">Нет активных серверов на этой ноде</p>
            <p class="empty-desc">Когда игроки создают комнаты в играх, инстансы будут отображаться здесь.</p>
          </div>
        </div>
      </template>
    </template>

    <!-- Подтверждение удаления сервиса -->
    <div
      v-if="showDeleteServiceConfirm"
      class="modal-overlay"
      @click.self="showDeleteServiceConfirm = false"
    >
      <div class="modal card delete-modal">
        <h3>Удалить {{ serviceToDelete?.service_type === 'volume' ? 'том' : 'сервис' }} «{{ serviceToDelete?.name }}»?</h3>
        <p class="delete-hint">
          <template v-if="serviceToDelete?.service_type === 'volume'">
            Персистентный том будет отмонтирован от игровых инстансов.
          </template>
          <template v-else>
            Контейнер сервиса будет остановлен и удален из сети Docker на ноде.
          </template>
        </p>
        <div class="checkbox-group">
          <label class="checkbox-label">
            <input v-model="deleteVolumeOnService" type="checkbox" />
            <span>
              Удалить данные на диске хоста
              <template v-if="serviceToDelete?.volume_path">
                (<code>{{ serviceToDelete.volume_path }}</code>)
              </template>
            </span>
          </label>
        </div>
        <div class="modal-actions">
          <button class="btn-outline" @click="showDeleteServiceConfirm = false">Отмена</button>
          <button class="btn-danger" :disabled="deletingService" @click="doDeleteService">
            {{ deletingService ? 'Удаление...' : 'Удалить' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Модальное окно развертывания сервиса -->
    <CreateServiceModal
      v-if="showCreateServiceModal"
      :node-id="node.id"
      @created="onServiceCreated"
      @cancel="showCreateServiceModal = false"
    />

    <!-- Подтверждение удаления ноды -->
    <div v-if="showDeleteConfirm" class="modal-overlay" @click.self="showDeleteConfirm = false">
      <div class="modal card delete-modal">
        <h3>Удалить ноду?</h3>
        <p>
          Нода <code>{{ node.address }}</code> будет удалена из реестра кластера.
        </p>
        <p class="text-danger">Все активные инстансы на этой ноде будут переведены в статус «Авария».</p>
        <div class="modal-actions">
          <button class="btn-outline" @click="showDeleteConfirm = false">Отмена</button>
          <button class="btn-danger" :disabled="deleting" @click="doDelete">
            {{ deleting ? 'Удаление...' : 'Удалить ноду' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue';
import { useRouter } from 'vue-router';
import {
  Trash2,
  AlertCircle,
  Plus,
  Copy,
  Check,
  Database,
  Layers,
  Server,
  FolderPlus,
  LayoutDashboard,
  ExternalLink,
  Globe,
} from 'lucide-vue-next';
import { StatusBadge, ResourceUsageCard } from '@/shared/ui';
import {
  getNode,
  getNodeUsage,
  deleteNode,
  registerNode,
  listNodeInstances,
  updateNodeRole,
  listNodeServices,
  deleteNodeService,
  createNodeService,
} from '@/entities/node';
import { listProjects } from '@/entities/project';
import { CreateServiceModal } from '@/features/manage-nodes';
import { formatBytes, formatDateTime, showToast } from '@/shared/lib';

const props = defineProps({
  nodeId: { type: [String, Number], required: true },
});
const router = useRouter();

const node = ref({});
const projectsMap = ref({});
const usage = ref({
  cpu_usage_percent: 0,
  memory_used_bytes: 0,
  disk_used_bytes: 0,
  network_bytes_per_sec: 0,
  active_instance_count: 0,
});
const nodeInstances = ref([]);
const instancesLoading = ref(true);
const services = ref([]);
const servicesLoading = ref(false);
const error = ref(null);
const showDeleteConfirm = ref(false);
const deleting = ref(false);
const authToken = ref('');
const authError = ref(null);
const authorizing = ref(false);

const updatingRole = ref(false);
const showCreateServiceModal = ref(false);
const togglingAdminer = ref(false);
const showDeleteServiceConfirm = ref(false);
const serviceToDelete = ref(null);
const deleteVolumeOnService = ref(true);
const deletingService = ref(false);
const copiedServiceId = ref(null);

const isUnauthorized = computed(
  () => node.value.status === 'NODE_STATUS_UNAUTHORIZED' || node.value.status === 'unauthorized',
);
const currentRole = computed(() => {
  const r = String(node.value.role || 'mixed').toLowerCase();
  if (r.includes('compute')) return 'compute';
  if (r.includes('storage')) return 'storage';
  return 'mixed';
});
const activeCount = computed(() => usage.value.active_instance_count ?? 0);

// Отделяем сервисы баз данных и томов от веб-панели управления
const isAdminer = (s) =>
  s.service_type === 'adminer' || s.type === 'adminer' || s.name === 'adminer';

const storageServices = computed(() =>
  services.value.filter((s) => !isAdminer(s)),
);
const adminerService = computed(() =>
  services.value.find((s) => isAdminer(s)),
);

let usageInterval = null;

async function fetchNode() {
  error.value = null;
  try {
    const resp = await getNode(props.nodeId);
    node.value = resp?.node || resp || {};
  } catch (e) {
    error.value = e.response?.data?.message ?? e.message;
  }
}

async function fetchUsage() {
  try {
    const data = await getNodeUsage(props.nodeId);
    usage.value = data;
  } catch {
    /* non-critical */
  }
}

async function fetchInstances() {
  instancesLoading.value = true;
  try {
    const instances = await listNodeInstances(props.nodeId);
    nodeInstances.value = instances;
  } catch {
    nodeInstances.value = [];
  } finally {
    instancesLoading.value = false;
  }
}

async function fetchServices() {
  servicesLoading.value = true;
  try {
    services.value = await listNodeServices(props.nodeId);
  } catch {
    services.value = [];
  } finally {
    servicesLoading.value = false;
  }
}

async function loadProjects() {
  try {
    const res = await listProjects();
    const map = {};
    (res.projects || []).forEach((p) => {
      map[p.id] = p.title || p.name;
    });
    projectsMap.value = map;
  } catch {
    /* non-critical */
  }
}

function getGameTitle(gid) {
  if (projectsMap.value && projectsMap.value[gid]) {
    return projectsMap.value[gid];
  }
  return `Игра #${gid}`;
}

async function setRole(newRole) {
  if (currentRole.value === newRole || updatingRole.value) return;
  updatingRole.value = true;
  try {
    const updated = await updateNodeRole(props.nodeId, newRole);
    node.value.role = updated.role || newRole;
    showToast(`Режим ноды переключен на ${newRole.toUpperCase()}`, 'success');
  } catch (e) {
    showToast(e.response?.data?.message || e.message || 'Ошибка обновления режима', 'error');
  } finally {
    updatingRole.value = false;
  }
}

function onServiceCreated(newService) {
  showCreateServiceModal.value = false;
  if (newService) {
    services.value.push(newService);
  } else {
    fetchServices();
  }
}

async function toggleAdminer() {
  if (togglingAdminer.value) return;
  togglingAdminer.value = true;

  if (adminerService.value) {
    // Отключение
    try {
      const targetId = adminerService.value.id;
      await deleteNodeService(props.nodeId, targetId, false);
      showToast('Веб-панель AdminerEvo отключена', 'success');
      services.value = services.value.filter((s) => s.id !== targetId);
      await fetchServices();
    } catch (e) {
      showToast(e.response?.data?.message || e.message || 'Ошибка отключения веб-панели', 'error');
    } finally {
      togglingAdminer.value = false;
    }
  } else {
    // Подключение
    try {
      const payload = {
        service_type: 'adminer',
        name: 'adminer',
        port: 0,
        allowed_game_ids: [],
      };
      await createNodeService(props.nodeId, payload);
      showToast('Веб-панель AdminerEvo успешно подключена', 'success');
      await fetchServices();
    } catch (e) {
      showToast(e.response?.data?.message || e.message || 'Ошибка подключения веб-панели', 'error');
    } finally {
      togglingAdminer.value = false;
    }
  }
}

function confirmDeleteService(svc) {
  serviceToDelete.value = svc;
  deleteVolumeOnService.value = true;
  showDeleteServiceConfirm.value = true;
}

async function doDeleteService() {
  if (!serviceToDelete.value) return;
  deletingService.value = true;
  try {
    await deleteNodeService(props.nodeId, serviceToDelete.value.id, deleteVolumeOnService.value);
    showToast('Сервис удален', 'success');
    services.value = services.value.filter((s) => s.id !== serviceToDelete.value.id);
    showDeleteServiceConfirm.value = false;
  } catch (e) {
    showToast(e.response?.data?.message || e.message || 'Ошибка удаления сервиса', 'error');
  } finally {
    deletingService.value = false;
  }
}

function copyUri(serviceId, uri) {
  if (!uri) return;
  navigator.clipboard.writeText(uri).then(() => {
    copiedServiceId.value = serviceId;
    setTimeout(() => {
      if (copiedServiceId.value === serviceId) {
        copiedServiceId.value = null;
      }
    }, 2000);
    showToast('Скопировано в буфер обмена', 'info');
  });
}

function getServiceIcon(type) {
  switch (type) {
    case 'postgres':
      return Database;
    case 'redis':
      return Layers;
    case 'mysql':
      return Server;
    case 'volume':
      return FolderPlus;
    case 'adminer':
      return LayoutDashboard;
    default:
      return Database;
  }
}

function getServiceColor(type) {
  switch (type) {
    case 'postgres':
      return '#336791';
    case 'redis':
      return '#dc382d';
    case 'mysql':
      return '#00758f';
    case 'volume':
      return '#10b981';
    case 'adminer':
      return '#58a6ff';
    default:
      return '#58a6ff';
  }
}

function getServiceBgColor(type) {
  switch (type) {
    case 'postgres':
      return 'rgba(51, 103, 145, 0.15)';
    case 'redis':
      return 'rgba(220, 56, 45, 0.15)';
    case 'mysql':
      return 'rgba(0, 117, 143, 0.15)';
    case 'volume':
      return 'rgba(16, 185, 129, 0.15)';
    case 'adminer':
      return 'rgba(88, 166, 255, 0.15)';
    default:
      return 'rgba(88, 166, 255, 0.15)';
  }
}

function formatServiceType(type) {
  switch (type) {
    case 'postgres':
      return 'PostgreSQL 16';
    case 'redis':
      return 'Redis 7';
    case 'mysql':
      return 'MySQL 8.0';
    case 'volume':
      return 'Персистентный том';
    case 'adminer':
      return 'AdminerEvo (Web UI)';
    default:
      return type || 'Service';
  }
}

function getAdminerUrl(svc) {
  if (!svc) return '#';
  let host = window.location.hostname || 'localhost';
  if (node.value && node.value.address) {
    const nodeHost = node.value.address.split(':')[0];
    if (
      nodeHost &&
      nodeHost !== '0.0.0.0' &&
      nodeHost !== '127.0.0.1' &&
      nodeHost !== 'localhost' &&
      nodeHost !== 'host.docker.internal'
    ) {
      host = nodeHost;
    }
  }

  let port = svc.host_port;
  if (!port && svc.connection_uri) {
    try {
      const parsed = new URL(svc.connection_uri);
      if (parsed.port) port = Number(parsed.port);
    } catch {}
  }
  if (!port) port = 8080;

  return `http://${host}:${port}`;
}

function isAdminerSupported(svc) {
  if (!svc) return false;
  const type = String(svc.service_type || svc.type || '').toLowerCase();
  return type === 'postgres' || type === 'postgresql' || type === 'mysql' || type === 'mariadb';
}

function parseDbCredentials(svc) {
  if (!svc) return { username: '', password: '', database: '' };
  const uri = svc.connection_uri || '';
  let username = '';
  let password = '';
  let database = '';

  if (uri) {
    try {
      const normalized = uri.replace(/^[a-zA-Z0-9+.-]+:\/\//, 'http://');
      const u = new URL(normalized);
      if (u.username) username = decodeURIComponent(u.username);
      if (u.password) password = decodeURIComponent(u.password);
      if (u.pathname && u.pathname.length > 1) {
        database = decodeURIComponent(u.pathname.slice(1));
      }
    } catch {
      // fallback
    }
  }

  const type = String(svc.service_type || svc.type || '').toLowerCase();
  if (!username) {
    if (type.includes('postgres')) username = 'postgres';
    else if (type.includes('mysql')) username = 'root';
  }
  if (!database) {
    database = 'game_db';
  }

  return { username, password, database };
}

function getAdminerDbUrl(svc) {
  if (!adminerService.value) return '#';
  const baseUrl = getAdminerUrl(adminerService.value);
  if (!baseUrl || baseUrl === '#') return '#';

  const creds = parseDbCredentials(svc);
  const type = String(svc.service_type || svc.type || '').toLowerCase();
  const params = new URLSearchParams();

  // Внутри Docker-сети gdh-network имя контейнера базы данных формируется как gdh-svc-<name>
  const serverHost = `gdh-svc-${svc.name}`;

  if (type.includes('postgres')) {
    params.set('pgsql', serverHost);
  } else {
    params.set('server', serverHost);
  }

  if (creds.username) params.set('username', creds.username);
  if (creds.database) params.set('db', creds.database);
  if (creds.password) params.set('password', creds.password);

  return `${baseUrl}/?${params.toString()}`;
}

async function submitAuthorize() {
  if (!authToken.value) return;
  authorizing.value = true;
  authError.value = null;
  try {
    await registerNode({
      node_id: Number(props.nodeId),
      token: authToken.value,
    });
    showToast('Нода авторизована', 'success');
    await fetchNode();
  } catch (e) {
    if (e.response?.status === 401) {
      authError.value = 'Неверный API-ключ ноды';
    } else if (e.response?.status === 409) {
      authError.value = 'Нода уже авторизована';
    } else {
      authError.value = e.response?.data?.message ?? 'Ошибка авторизации ноды';
    }
  } finally {
    authorizing.value = false;
  }
}

async function doDelete() {
  deleting.value = true;
  try {
    await deleteNode(props.nodeId);
    showToast('Нода удалена', 'success');
    router.push('/nodes');
  } catch (e) {
    showToast(e.response?.data?.message ?? 'Ошибка удаления', 'error');
  } finally {
    deleting.value = false;
  }
}

watch(
  () => node.value.status,
  async (status) => {
    if (status !== 'NODE_STATUS_UNAUTHORIZED' && status !== 'unauthorized' && !error.value) {
      await fetchUsage();
      await fetchInstances();
      await fetchServices();
      if (usageInterval) clearInterval(usageInterval);
      usageInterval = setInterval(fetchUsage, 5000);
    }
  },
);

onMounted(async () => {
  loadProjects();
  await fetchNode();
  if (!error.value && !isUnauthorized.value) {
    await fetchUsage();
    await fetchInstances();
    await fetchServices();
    usageInterval = setInterval(fetchUsage, 5000);
  }
});

onUnmounted(() => {
  if (usageInterval) clearInterval(usageInterval);
});
</script>

<style scoped>
.node-detail-page {
  width: 100%;
  max-width: 100%;
  padding: 24px 32px;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

/* Шапка */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
}

.header-main {
  display: flex;
  align-items: center;
  gap: 12px;
}

.node-title {
  margin: 0;
  font-size: 1.4rem;
  font-weight: 700;
  color: var(--text-main, #f0f6fc);
  letter-spacing: -0.2px;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
}

/* Переключатель роли ноды */
.role-selector-container {
  display: flex;
  align-items: center;
  gap: 10px;
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border, #30363d);
  padding: 4px 8px 4px 12px;
  border-radius: var(--radius-md, 8px);
}

.role-selector-label {
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--text-muted, #8b949e);
}

.role-segmented-control {
  display: inline-flex;
  gap: 3px;
  background: var(--bg-app, #0d1117);
  padding: 3px;
  border-radius: var(--radius-sm, 6px);
  border: 1px solid var(--border, #30363d);
}

.role-pill-btn {
  background: transparent;
  border: none;
  padding: 5px 12px;
  font-size: 0.78rem;
  font-weight: 600;
  color: var(--text-muted, #8b949e);
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.role-pill-btn:hover:not(:disabled) {
  color: var(--text-main, #f0f6fc);
}

.role-pill-btn.active {
  background: var(--primary, #58a6ff);
  color: #ffffff;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.25);
}

.role-pill-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-delete-node {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: transparent;
  border: 1px solid rgba(248, 81, 73, 0.35);
  color: #f85149;
  padding: 7px 14px;
  border-radius: var(--radius-md, 6px);
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s;
}

.btn-delete-node:hover {
  background: rgba(248, 81, 73, 0.12);
  border-color: #f85149;
}

/* Ошибка */
.error-banner {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  background: rgba(248, 81, 73, 0.12);
  color: #f85149;
  border: 1px solid rgba(248, 81, 73, 0.3);
  border-radius: var(--radius-md, 8px);
  font-size: 0.88rem;
}

.btn-retry {
  margin-left: auto;
  background: transparent;
  border: 1px solid #f85149;
  color: #f85149;
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 0.8rem;
  cursor: pointer;
}

/* Авторизация */
.auth-card {
  border: 1px solid #d29922;
  background: rgba(210, 153, 34, 0.05);
  border-radius: var(--radius-lg, 10px);
  padding: 24px;
  width: 100%;
  max-width: 480px;
  margin: 0 auto;
  align-self: center;
  box-sizing: border-box;
}

.auth-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
  width: 100%;
}

.auth-form .form-group {
  margin-bottom: 0;
}

.auth-error {
  color: #f85149;
  font-size: 0.85rem;
}

/* Верхний обзорный блок: Информация о ноде и Потребление ресурсов */
.overview-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  align-items: stretch;
}

.overview-grid.is-unauthorized {
  grid-template-columns: 1fr;
}

@media (max-width: 992px) {
  .overview-grid {
    grid-template-columns: 1fr;
  }
}

.card-header-simple {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 26px;
  margin-bottom: 16px;
}

.specs-header-ping {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 0.82rem;
}

.ping-label {
  color: var(--text-muted, #8b949e);
  font-weight: 500;
}

.ping-value {
  color: var(--text-main, #f0f6fc);
  font-weight: 500;
}

.card-header-simple h3 {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
  line-height: 1.4;
  color: var(--text-main, #f0f6fc);
}

.specs-card {
  padding: 20px;
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-lg, 10px);
  display: flex;
  flex-direction: column;
}

.specs-grid-2col {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px 28px;
  flex: 1;
}

@media (max-width: 600px) {
  .specs-grid-2col {
    grid-template-columns: 1fr;
  }
}

.spec-col {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 8px;
}

.spec-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  padding: 6px 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
}

.spec-row:last-child {
  border-bottom: none;
}

.spec-label {
  font-size: 0.8rem;
  color: var(--text-muted, #8b949e);
  font-weight: 500;
}

.spec-value {
  font-size: 0.85rem;
  font-weight: 500;
  color: var(--text-main, #f0f6fc);
  text-align: right;
  word-break: break-all;
}

.resources-card {
  padding: 20px;
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-lg, 10px);
  display: flex;
  flex-direction: column;
}

.resources-grid-2x2 {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  grid-template-rows: repeat(2, 1fr);
  gap: 12px;
  flex: 1;
}

/* Секции */
.section-block {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.section-title-wrap {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

.section-title-wrap h2 {
  margin: 0;
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
  display: flex;
  align-items: center;
  gap: 8px;
}

.count-badge {
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--text-muted, #8b949e);
  background: var(--bg-secondary, #21262d);
  padding: 2px 8px;
  border-radius: 10px;
}

/* Таблицы */
.table-container {
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-lg, 10px);
  overflow: hidden;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
}

.data-table th {
  text-align: left;
  padding: 12px 16px;
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--text-muted, #8b949e);
  text-transform: uppercase;
  letter-spacing: 0.4px;
  background: var(--bg-app, #0d1117);
  border-bottom: 1px solid var(--border, #30363d);
}

.data-table td {
  padding: 12px 16px;
  font-size: 0.88rem;
  border-bottom: 1px solid var(--border, #30363d);
  color: var(--text-main, #f0f6fc);
  vertical-align: middle;
}

.data-table tr:last-child td {
  border-bottom: none;
}

.clickable-row {
  cursor: pointer;
  transition: background 0.15s;
}

.clickable-row:hover {
  background: var(--bg-hover, #21262d);
}

/* Сервис ячейки */
.cell-service-name {
  min-width: 180px;
}

.service-identity {
  display: flex;
  align-items: center;
  gap: 10px;
}

.service-type-icon-box {
  width: 32px;
  height: 32px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.service-icon {
  width: 17px;
  height: 17px;
}

.service-name-text {
  display: flex;
  flex-direction: column;
  gap: 1px;
}

.service-name {
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
}

.service-subtext {
  font-size: 0.72rem;
  color: var(--text-muted, #8b949e);
}

.service-type-badge {
  font-size: 0.8rem;
  color: var(--text-muted, #8b949e);
  font-weight: 500;
}

.connection-cell {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.path-code,
.uri-code {
  max-width: 240px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-family: monospace;
  font-size: 0.8rem;
  background: rgba(255, 255, 255, 0.05);
  padding: 2px 6px;
  border-radius: 4px;
}

/* ─── Web UI Row (AdminerEvo Singleton) ─────────────────────────────────── */
.web-ui-row-container {
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-md, 8px);
  margin-bottom: 20px;
  overflow: hidden;
}

.web-ui-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 20px;
  gap: 16px;
  background: var(--bg-card, #161b22);
  transition: background 0.15s ease;
}

.web-ui-row:hover {
  background: var(--bg-secondary, #1c2128);
}

.web-ui-cell-title {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 2;
  min-width: 250px;
}

.web-ui-cell-status {
  flex: 1;
  display: flex;
  align-items: center;
}

.web-ui-cell-link {
  flex: 1;
  display: flex;
  align-items: center;
}

.text-muted-link {
  font-size: 0.85rem;
  color: var(--text-tertiary, #6e7681);
}

.web-ui-cell-action {
  display: flex;
  align-items: center;
  justify-content: flex-end;
}

.btn-danger-outline {
  border-color: rgba(248, 81, 73, 0.4);
  color: var(--danger, #f85149);
}

.btn-danger-outline:hover {
  background: rgba(248, 81, 73, 0.15);
  border-color: var(--danger, #f85149);
}

.adminer-link {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  color: var(--primary, #58a6ff);
  font-size: 0.85rem;
  font-weight: 600;
  text-decoration: none;
}

.adminer-link:hover {
  text-decoration: underline;
}

.btn-adminer-open {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  background: rgba(88, 166, 255, 0.1);
  border: 1px solid rgba(88, 166, 255, 0.3);
  border-radius: 6px;
  color: var(--primary, #58a6ff);
  font-size: 0.8rem;
  font-weight: 500;
  text-decoration: none;
  transition: all 0.15s ease;
  white-space: nowrap;
}

.btn-adminer-open:hover {
  background: rgba(88, 166, 255, 0.2);
  border-color: var(--primary, #58a6ff);
  color: #fff;
}

.host-port-hint {
  font-size: 0.75rem;
  color: var(--text-muted, #8b949e);
  font-family: monospace;
}

.icon-btn {
  background: transparent;
  border: none;
  color: var(--text-muted, #8b949e);
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.1s;
}

.icon-btn:hover {
  color: var(--text-main, #f0f6fc);
  background: var(--bg-hover, #21262d);
}

.text-success {
  color: #3fb950;
}

.volume-stat {
  display: flex;
  flex-direction: column;
  gap: 1px;
}

.volume-size-num {
  font-weight: 600;
  font-size: 0.84rem;
  color: var(--text-main, #f0f6fc);
}

.volume-path-snippet {
  font-size: 0.72rem;
  color: var(--text-muted, #8b949e);
  max-width: 140px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.access-badge-all {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 0.88rem;
  color: var(--text-main, #f0f6fc);
}

.game-badges-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
}

.game-id-tag {
  display: inline-flex;
  align-items: center;
  font-size: 0.82rem;
  font-weight: 500;
  padding: 3px 8px;
  border-radius: 5px;
  background: rgba(88, 166, 255, 0.1);
  border: 1px solid rgba(88, 166, 255, 0.25);
  color: #79c0ff;
  white-space: nowrap;
}

.col-actions {
  width: 72px;
  text-align: center;
  padding-left: 8px !important;
  padding-right: 8px !important;
}

.actions-cell {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  width: 100%;
}

.btn-icon-action {
  background: transparent;
  border: none;
  color: var(--primary, #58a6ff);
  cursor: pointer;
  padding: 6px;
  border-radius: 4px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: background 0.15s;
}

.btn-icon-action:hover {
  background: rgba(88, 166, 255, 0.12);
}

.btn-icon-danger {
  background: transparent;
  border: none;
  color: #f85149;
  cursor: pointer;
  padding: 6px;
  border-radius: 4px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: background 0.15s;
}

.btn-icon-danger:hover {
  background: rgba(248, 81, 73, 0.12);
}

.game-tag {
  display: inline-block;
  font-size: 0.78rem;
  font-weight: 500;
  padding: 2px 8px;
  border-radius: 4px;
  background: var(--bg-app, #0d1117);
  border: 1px solid var(--border, #30363d);
}

/* Пустые состояния */
.empty-state-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: 36px 20px;
  background: var(--bg-card, #161b22);
  border: 1px dashed var(--border, #30363d);
  border-radius: var(--radius-lg, 10px);
  gap: 10px;
}

.empty-icon {
  width: 36px;
  height: 36px;
  color: var(--text-muted, #8b949e);
  opacity: 0.6;
}

.empty-title {
  margin: 0;
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
}

.empty-desc {
  margin: 0;
  font-size: 0.82rem;
  color: var(--text-muted, #8b949e);
  max-width: 500px;
}

.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 36px;
  color: var(--text-muted, #8b949e);
  font-size: 0.88rem;
}

.spinner-sm {
  width: 16px;
  height: 16px;
  border: 2px solid var(--border, #30363d);
  border-top-color: var(--primary, #58a6ff);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  display: inline-block;
  flex-shrink: 0;
  vertical-align: middle;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

/* Модалки удаления */
.delete-modal {
  max-width: 460px;
  width: 90%;
  padding: 24px;
}

.delete-modal h3 {
  margin: 0 0 10px;
  font-size: 1.15rem;
  color: var(--text-main, #f0f6fc);
}

.delete-hint {
  font-size: 0.85rem;
  color: var(--text-muted, #8b949e);
  margin: 0 0 16px;
  line-height: 1.45;
}

.checkbox-group {
  margin: 16px 0;
  padding: 12px;
  background: var(--bg-app, #0d1117);
  border-radius: var(--radius-sm, 6px);
  border: 1px solid var(--border, #30363d);
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 0.84rem;
  color: var(--text-main, #f0f6fc);
  cursor: pointer;
}

.checkbox-label input {
  accent-color: #f85149;
  width: 16px;
  height: 16px;
}

.text-danger {
  color: #f85149;
  font-size: 0.85rem;
  font-weight: 500;
  margin-top: 8px;
}

.btn-danger {
  background: #da3633;
  color: #ffffff;
  border: 1px solid rgba(248, 81, 73, 0.4);
  padding: 8px 16px;
  border-radius: var(--radius-md, 6px);
  font-weight: 600;
  font-size: 0.85rem;
  cursor: pointer;
  transition: background 0.15s;
}

.btn-danger:hover:not(:disabled) {
  background: #f85149;
}

.btn-danger:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.text-right {
  text-align: right;
}

.icon-xs {
  width: 14px;
  height: 14px;
}

.icon-sm {
  width: 16px;
  height: 16px;
}

.icon-md {
  width: 24px;
  height: 24px;
}

@media (max-width: 1024px) {
  .specs-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  .resources-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .node-detail-page {
    padding: 16px;
  }
  .specs-grid {
    grid-template-columns: 1fr;
  }
  .resources-grid {
    grid-template-columns: 1fr;
  }
  .header-actions {
    width: 100%;
    justify-content: space-between;
  }
}
</style>
