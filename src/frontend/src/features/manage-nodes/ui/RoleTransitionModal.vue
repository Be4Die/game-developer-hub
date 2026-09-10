<template>
  <Teleport to="body">
    <div class="modal-overlay" @click.self="handleCancel">
      <div class="modal card role-transition-modal">
        <!-- Header -->
        <div class="modal-header">
          <div class="header-title-wrap">
            <div class="header-icon-box" :class="targetRole">
              <Server v-if="targetRole === 'compute'" class="icon-sm" />
              <Database v-else-if="targetRole === 'storage'" class="icon-sm" />
              <Layers v-else class="icon-sm" />
            </div>
            <div>
              <h3>{{ modalTitle }}</h3>
              <p class="modal-subtitle">{{ modalSubtitle }}</p>
            </div>
          </div>
          <button class="close-btn" :disabled="processing" @click="handleCancel">&times;</button>
        </div>

        <!-- Body: Transition to Storage (Game servers conflict) -->
        <div v-if="targetRole === 'storage'" class="modal-body">
          <div class="warning-callout">
            <AlertTriangle class="callout-icon text-warning" />
            <div class="callout-text">
              <strong>Режим Storage не поддерживает игровые серверы.</strong>
              На ноде обнаружено <strong>{{ instances.length }}</strong> активных игровых серверов. 
              Для переключения роли необходимо принудительно завершить их работу и освободить ноду.
            </div>
          </div>

          <!-- Список инстансов -->
          <div class="workload-list-section">
            <div class="workload-list-header">
              <span>Затронутые серверы ({{ instances.length }})</span>
            </div>
            <div class="workload-scroll-container">
              <div v-for="inst in instances" :key="inst.id" class="workload-item">
                <div class="item-left">
                  <span class="item-id">#{{ inst.id }}</span>
                  <span class="item-name">{{ getGameTitle(inst.game_id) }}</span>
                </div>
                <div class="item-right">
                  <span class="item-badge status-danger">Будет завершен</span>
                </div>
              </div>
            </div>
          </div>

          <div class="danger-note">
            <p>Внимание: Все текущие матчи на этих серверах будут немедленно прерваны, а контейнеры удалены.</p>
          </div>
        </div>

        <!-- Body: Transition to Compute (Storage databases conflict) -->
        <div v-else-if="targetRole === 'compute'" class="modal-body">
          <!-- 100% Data Guarantee Banner -->
          <div class="guarantee-banner">
            <ShieldCheck class="guarantee-icon text-success" />
            <div class="guarantee-content">
              <strong>Гарантия 100% сохранности данных:</strong>
              Перед выполнением любого действия система в обязательном порядке создаст полную резервную копию всех баз данных, независимо от статуса авторезервирования.
            </div>
          </div>

          <!-- Список баз данных / сервисов -->
          <div class="workload-list-section">
            <div class="workload-list-header">
              <span>Сервисы хранения на ноде ({{ services.length }})</span>
            </div>
            <div class="workload-scroll-container">
              <div v-for="svc in services" :key="svc.id" class="workload-item">
                <div class="item-left">
                  <Database class="item-svc-icon" />
                  <span class="item-name">{{ svc.name }}</span>
                  <span class="item-type-tag">{{ formatServiceType(svc.service_type) }}</span>
                </div>
                <div class="item-right">
                  <span class="item-size">{{ svc.volume_size_bytes ? formatBytes(svc.volume_size_bytes) : '0 Б' }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- Варианты действия (3 варианта) -->
          <div class="decision-options-group">
            <label class="decision-label">Выберите действие над базами данных:</label>
            
            <!-- Вариант 1: Остановить БД (Рекомендуется) -->
            <div
              class="decision-card"
              :class="{ selected: selectedStorageAction === 'stop' }"
              @click="selectAction('stop')"
            >
              <div class="decision-radio">
                <input
                  type="radio"
                  name="storage_action"
                  value="stop"
                  :checked="selectedStorageAction === 'stop'"
                  :disabled="processing"
                />
              </div>
              <div class="decision-info">
                <div class="decision-title-row">
                  <span class="decision-title">Остановить базы данных</span>
                  <span class="badge-recommended">Рекомендуется</span>
                </div>
                <p class="decision-desc">
                  Создается резервная копия, контейнеры СУБД останавливаются. 
                  Тома данных <strong>остаются сохранными на диске</strong>. При обратном включении Storage сервисы можно будет запустить в 1 клик.
                </p>
              </div>
            </div>

            <!-- Вариант 2: Удалить БД -->
            <div
              class="decision-card"
              :class="{ selected: selectedStorageAction === 'delete' }"
              @click="selectAction('delete')"
            >
              <div class="decision-radio">
                <input
                  type="radio"
                  name="storage_action"
                  value="delete"
                  :checked="selectedStorageAction === 'delete'"
                  :disabled="processing"
                />
              </div>
              <div class="decision-info">
                <div class="decision-title-row">
                  <span class="decision-title text-danger">Удалить базы данных и тома</span>
                  <span class="badge-danger">Очистка диска</span>
                </div>
                <p class="decision-desc">
                  Создается резервная копия, сервисы останавливаются, а контейнеры и тома данных <strong>полностью удаляются</strong> с диска. 
                  Длительный процесс.
                </p>
              </div>
            </div>
          </div>
        </div>

        <!-- Processing State Overlay / Indicator -->
        <div v-if="processing" class="processing-overlay">
          <div class="spinner"></div>
          <span class="processing-text">{{ processingMessage }}</span>
        </div>

        <!-- Footer -->
        <div class="modal-footer">
          <button
            class="btn-secondary"
            :disabled="processing"
            @click="handleCancel"
          >
            Отмена
          </button>

          <!-- Storage action button -->
          <template v-if="targetRole === 'storage'">
            <button
              class="btn-danger"
              :disabled="processing"
              @click="confirmStorageTransition"
            >
              <Trash2 class="icon-xs" />
              <span>Завершить серверы и перейти</span>
            </button>
          </template>

          <!-- Compute action button -->
          <template v-else-if="targetRole === 'compute'">
            <button
              v-if="selectedStorageAction === 'stop'"
              class="btn-primary"
              :disabled="processing"
              @click="confirmComputeTransition('stop')"
            >
              <PauseCircle class="icon-xs" />
              <span>Остановить БД и перейти в Compute</span>
            </button>
            <button
              v-else
              class="btn-danger"
              :disabled="processing"
              @click="confirmComputeTransition('delete')"
            >
              <Trash2 class="icon-xs" />
              <span>Удалить БД и перейти в Compute</span>
            </button>
          </template>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, computed } from 'vue';
import {
  Server,
  Database,
  Layers,
  AlertTriangle,
  ShieldCheck,
  Trash2,
  PauseCircle,
} from 'lucide-vue-next';
import {
  STORAGE_TRANSITION_STOP,
  STORAGE_TRANSITION_DELETE,
  COMPUTE_TRANSITION_TERMINATE,
} from '@/entities/node';

const props = defineProps({
  node: {
    type: Object,
    required: true,
  },
  targetRole: {
    type: String,
    required: true,
  },
  instances: {
    type: Array,
    default: () => [],
  },
  services: {
    type: Array,
    default: () => [],
  },
  projectsMap: {
    type: Object,
    default: () => ({}),
  },
  processing: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(['close', 'confirm']);

const selectedStorageAction = ref('stop');

const modalTitle = computed(() => {
  if (props.targetRole === 'storage') {
    return 'Переключение ноды в режим Storage';
  }
  if (props.targetRole === 'compute') {
    return 'Переключение ноды в режим Compute';
  }
  return 'Смена режима ноды';
});

const modalSubtitle = computed(() => {
  if (props.targetRole === 'storage') {
    return 'На ноде обнаружены работающие игровые сервера';
  }
  if (props.targetRole === 'compute') {
    return 'На ноде развернуты сервисы баз данных и хранилища';
  }
  return '';
});

const processingMessage = computed(() => {
  if (props.targetRole === 'storage') {
    return 'Остановка серверов и переключение режима ноды...';
  }
  if (selectedStorageAction.value === 'delete') {
    return 'Создание резервных копий, удаление сервисов и томов данных... Пожалуйста, подождите.';
  }
  return 'Создание резервных копий и остановка сервисов... Пожалуйста, подождите.';
});

function selectAction(action) {
  if (props.processing) return;
  selectedStorageAction.value = action;
}

function getGameTitle(gameId) {
  if (!gameId) return 'Неизвестная игра';
  if (props.projectsMap && props.projectsMap[gameId]) {
    return props.projectsMap[gameId];
  }
  return 'Игра #' + gameId;
}

function formatServiceType(t) {
  const map = {
    postgres: 'PostgreSQL',
    redis: 'Redis',
    mysql: 'MySQL',
    volume: 'Том данных',
    adminer: 'Adminer',
    pgadmin: 'pgAdmin',
  };
  return map[t] || t || 'Сервис';
}

function formatBytes(bytes) {
  if (!bytes || bytes === 0) return '0 Б';
  const k = 1024;
  const sizes = ['Б', 'КБ', 'МБ', 'ГБ', 'ТБ'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
}

function handleCancel() {
  if (props.processing) return;
  emit('close');
}

function confirmStorageTransition() {
  emit('confirm', {
    compute_action: COMPUTE_TRANSITION_TERMINATE,
  });
}

function confirmComputeTransition(action) {
  emit('confirm', {
    storage_action: action === 'delete' ? STORAGE_TRANSITION_DELETE : STORAGE_TRANSITION_STOP,
  });
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background-color: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.role-transition-modal {
  position: relative;
  width: 100%;
  max-width: 580px;
  background-color: var(--color-surface, #1e2024);
  border: 1px solid var(--color-border, #2d3139);
  border-radius: 12px;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.5);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid var(--color-border, #2d3139);
}

.header-title-wrap {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.header-icon-box {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(99, 102, 241, 0.15);
  color: #818cf8;
}

.header-icon-box.storage {
  background: rgba(16, 185, 129, 0.15);
  color: #34d399;
}

.header-icon-box.compute {
  background: rgba(59, 130, 246, 0.15);
  color: #60a5fa;
}

.modal-header h3 {
  margin: 0;
  font-size: 1.15rem;
  font-weight: 600;
  color: var(--color-text, #f3f4f6);
}

.modal-subtitle {
  margin: 0.2rem 0 0;
  font-size: 0.8rem;
  color: var(--color-text-muted, #9ca3af);
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.5rem;
  line-height: 1;
  color: var(--color-text-muted, #9ca3af);
  cursor: pointer;
  padding: 0.25rem;
  border-radius: 6px;
  transition: color 0.15s;
}

.close-btn:hover:not(:disabled) {
  color: var(--color-text, #fff);
}

.modal-body {
  padding: 1.25rem 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  max-height: 70vh;
  overflow-y: auto;
}

/* Callouts */
.warning-callout {
  display: flex;
  gap: 0.75rem;
  background: rgba(245, 158, 11, 0.1);
  border: 1px solid rgba(245, 158, 11, 0.3);
  padding: 0.85rem 1rem;
  border-radius: 8px;
  font-size: 0.85rem;
  line-height: 1.4;
  color: #f3f4f6;
}

.callout-icon {
  flex-shrink: 0;
  margin-top: 2px;
}

.guarantee-banner {
  display: flex;
  gap: 0.75rem;
  background: rgba(16, 185, 129, 0.1);
  border: 1px solid rgba(16, 185, 129, 0.3);
  padding: 0.85rem 1rem;
  border-radius: 8px;
  font-size: 0.85rem;
  line-height: 1.4;
  color: #f3f4f6;
}

.guarantee-icon {
  flex-shrink: 0;
  margin-top: 2px;
}

/* Workload items */
.workload-list-section {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.workload-list-header {
  font-size: 0.8rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--color-text-muted, #9ca3af);
}

.workload-scroll-container {
  max-height: 140px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  background: rgba(0, 0, 0, 0.2);
  border: 1px solid var(--color-border, #2d3139);
  border-radius: 8px;
  padding: 0.5rem;
}

.workload-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.4rem 0.6rem;
  background: rgba(255, 255, 255, 0.02);
  border-radius: 6px;
  font-size: 0.85rem;
}

.item-left {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.item-id {
  font-family: monospace;
  font-size: 0.8rem;
  color: var(--color-text-muted, #9ca3af);
}

.item-svc-icon {
  width: 14px;
  height: 14px;
  color: #60a5fa;
}

.item-type-tag {
  font-size: 0.72rem;
  padding: 0.1rem 0.4rem;
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.06);
  color: var(--color-text-muted, #9ca3af);
}

.status-danger {
  color: #f87171;
  font-size: 0.75rem;
  font-weight: 500;
}

.danger-note {
  font-size: 0.8rem;
  color: #f87171;
}

/* Decision Cards */
.decision-options-group {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.decision-label {
  font-size: 0.82rem;
  font-weight: 600;
  color: var(--color-text, #f3f4f6);
}

.decision-card {
  display: flex;
  gap: 0.85rem;
  padding: 0.9rem 1rem;
  border: 1px solid var(--color-border, #2d3139);
  background: rgba(255, 255, 255, 0.02);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.decision-card:hover {
  background: rgba(255, 255, 255, 0.04);
  border-color: rgba(99, 102, 241, 0.4);
}

.decision-card.selected {
  border-color: #6366f1;
  background: rgba(99, 102, 241, 0.08);
}

.decision-radio {
  padding-top: 2px;
}

.decision-info {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.decision-title-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.decision-title {
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--color-text, #f3f4f6);
}

.badge-recommended {
  font-size: 0.7rem;
  font-weight: 600;
  padding: 0.1rem 0.45rem;
  border-radius: 4px;
  background: rgba(16, 185, 129, 0.18);
  color: #34d399;
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.badge-danger {
  font-size: 0.7rem;
  font-weight: 600;
  padding: 0.1rem 0.45rem;
  border-radius: 4px;
  background: rgba(239, 68, 68, 0.18);
  color: #f87171;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.decision-desc {
  margin: 0;
  font-size: 0.8rem;
  color: var(--color-text-muted, #9ca3af);
  line-height: 1.35;
}

/* Modal Footer */
.modal-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.75rem;
  padding: 1rem 1.5rem;
  border-top: 1px solid var(--color-border, #2d3139);
  background: rgba(0, 0, 0, 0.15);
}

.processing-overlay {
  position: absolute;
  inset: 0;
  background: rgba(20, 22, 26, 0.88);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  z-index: 10;
  backdrop-filter: blur(2px);
}

.spinner {
  width: 36px;
  height: 36px;
  border: 3px solid rgba(255, 255, 255, 0.1);
  border-top-color: #6366f1;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.processing-text {
  font-size: 0.9rem;
  font-weight: 500;
  color: #f3f4f6;
  text-align: center;
  max-width: 80%;
  line-height: 1.4;
}
</style>
