<template>
  <Teleport to="body">
    <div class="modal-overlay" @click.self="$emit('close')">
      <div class="modal card backups-modal">
        <!-- Заголовок модального окна -->
        <div class="modal-header">
          <div class="modal-header-info">
            <div class="service-badge-icon" :style="{ backgroundColor: getServiceBgColor(service.service_type) }">
              <component :is="getServiceIcon(service.service_type)" class="icon-sm" :style="{ color: getServiceColor(service.service_type) }" />
            </div>
            <div>
              <div class="title-with-badge">
                <h3>Резервные копии (Бэкапы)</h3>
                <span class="service-name-tag">{{ service.name }}</span>
              </div>
              <p class="modal-subtitle">
                Управление снимками данных, скачивание дампов на ПК и восстановление сервиса
              </p>
            </div>
          </div>
          <button class="close-btn" @click="$emit('close')">&times;</button>
        </div>

        <!-- Полоса настройки авто-бэкапов сервиса -->
        <div class="auto-backup-strip">
          <div class="auto-backup-info">
            <Clock class="icon-xs auto-backup-clock" />
            <div>
              <div class="auto-backup-title">Автоматическое расписание бэкапов</div>
              <div class="auto-backup-desc">Ежедневный снимок в 03:00 UTC (глубина хранения — 7 последних копий)</div>
            </div>
          </div>
          <div class="auto-backup-toggle">
            <label class="switch-toggle" title="Включить / отключить авто-бэкапы">
              <input
                type="checkbox"
                v-model="autoBackupEnabled"
                @change="handleToggleAutoBackup"
              />
              <span class="switch-slider"></span>
            </label>
            <span class="switch-text" :class="autoBackupEnabled ? 'text-success' : 'text-muted'">
              {{ autoBackupEnabled ? 'Включено' : 'Отключено' }}
            </span>
          </div>
        </div>

      <!-- Сообщения об ошибке / успехе -->
      <div v-if="error" class="alert-banner alert-danger">
        <AlertTriangle class="icon-xs" />
        <span>{{ error }}</span>
        <button class="alert-close" @click="error = null">&times;</button>
      </div>
      <div v-if="successMessage" class="alert-banner alert-success">
        <Check class="icon-xs" />
        <span>{{ successMessage }}</span>
        <button class="alert-close" @click="successMessage = null">&times;</button>
      </div>

      <!-- Верхняя панель действий -->
      <div class="actions-bar">
        <div class="actions-left">
          <button
            class="btn-primary btn-sm"
            :disabled="creating || loading || !isServiceOnline"
            @click="handleCreateBackup"
          >
            <span v-if="creating" class="spinner-inline"></span>
            <Plus v-else class="icon-xs" />
            <span>{{ creating ? 'Создание снимка...' : 'Создать бэкап сейчас' }}</span>
          </button>

          <button
            class="btn-outline btn-sm"
            :disabled="uploading || loading || !isServiceOnline"
            @click="showUploadForm = !showUploadForm"
          >
            <UploadCloud class="icon-xs" />
            <span>{{ showUploadForm ? 'Скрыть загрузку' : 'Загрузить свой бэкап' }}</span>
          </button>
        </div>

        <button class="icon-btn-secondary" title="Обновить список" :disabled="loading" @click="loadBackups">
          <RotateCcw class="icon-xs" :class="{ 'spin-icon': loading }" />
        </button>
      </div>

      <!-- Сворачиваемая форма загрузки кастомного бэкапа -->
      <div v-if="showUploadForm" class="upload-section">
        <div class="upload-dropzone" :class="{ 'has-file': !!uploadFile }" @dragover.prevent @drop.prevent="handleDrop">
          <input
            ref="fileInputRef"
            type="file"
            class="file-input-hidden"
            accept=".sql,.sql.gz,.gz,.tar.gz,.rdb,.dump"
            @change="handleFileChange"
          />

          <div v-if="!uploadFile" class="dropzone-content" @click="$refs.fileInputRef.click()">
            <UploadCloud class="dropzone-icon" />
            <p class="dropzone-title">Перетащите файл бэкапа сюда или <span>выберите на компьютере</span></p>
            <p class="dropzone-hint">Поддерживаются: .sql, .sql.gz, .tar.gz, .rdb (до 2 ГБ)</p>
          </div>

          <div v-else class="dropzone-selected">
            <FileArchive class="file-icon" />
            <div class="selected-meta">
              <span class="selected-name">{{ uploadFile.name }}</span>
              <span class="selected-size">{{ formatBytes(uploadFile.size) }}</span>
            </div>
            <button class="remove-file-btn" :disabled="uploading" @click.stop="uploadFile = null">&times;</button>
          </div>
        </div>

        <div v-if="uploadFile" class="upload-options">
          <label class="checkbox-label">
            <input v-model="restoreImmediately" type="checkbox" :disabled="uploading" />
            <span>Автоматически восстановить базу сразу после загрузки</span>
          </label>

          <div v-if="uploading" class="progress-bar-wrap">
            <div class="progress-track">
              <div class="progress-fill" :style="{ width: `${uploadProgress}%` }"></div>
            </div>
            <span class="progress-text">Загрузка: {{ uploadProgress }}%</span>
          </div>

          <div class="upload-buttons">
            <button class="btn-primary btn-sm" :disabled="uploading" @click="submitUpload">
              {{ uploading ? 'Загрузка архива...' : 'Начать загрузку' }}
            </button>
            <button class="btn-ghost btn-sm" :disabled="uploading" @click="cancelUpload">Отмена</button>
          </div>
        </div>
      </div>

      <!-- Список бэкапов -->
      <div class="backups-content">
        <div v-if="loading && !backups.length" class="loading-state">
          <div class="spinner-sm"></div>
          <span>Загрузка списка резервных копий...</span>
        </div>

        <div v-else-if="!backups.length" class="empty-state">
          <HardDrive class="empty-icon" />
          <p class="empty-title">Резервных копий пока нет</p>
          <p class="empty-desc">
            Создайте первый снимок данных прямо сейчас или загрузите существующий дамп.
          </p>
        </div>

        <div v-else class="table-wrap">
          <table class="data-table">
            <thead>
              <tr>
                <th>Имя снимка / Файл</th>
                <th>Тип</th>
                <th>Размер</th>
                <th>Создан</th>
                <th>Статус</th>
                <th class="col-actions">Действия</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="b in backups" :key="b.backup_id">
                <!-- Файл -->
                <td class="cell-file">
                  <div class="file-name-wrap">
                    <FileArchive class="icon-xs text-muted" />
                    <div>
                      <span class="file-name" :title="b.file_name">{{ b.file_name || b.backup_id }}</span>
                      <span v-if="b.checksum" class="checksum-hint" :title="b.checksum">
                        SHA256: {{ b.checksum.substring(0, 10) }}...
                      </span>
                    </div>
                  </div>
                </td>

                <!-- Тип -->
                <td>
                  <span class="type-badge" :class="getTypeBadgeClass(b.backup_type)">
                    {{ formatBackupType(b.backup_type) }}
                  </span>
                </td>

                <!-- Размер -->
                <td class="cell-size">
                  {{ formatBytes(b.size_bytes) }}
                </td>

                <!-- Дата -->
                <td class="cell-date">
                  {{ formatDate(b.created_at) }}
                </td>

                <!-- Статус -->
                <td>
                  <span class="status-badge" :class="getStatusBadgeClass(b.status)">
                    <span class="status-dot"></span>
                    {{ formatStatus(b.status) }}
                  </span>
                </td>

                <!-- Действия -->
                <td class="col-actions">
                  <div class="action-buttons-cell">
                    <!-- Скачать на ПК -->
                    <button
                      class="btn-action-icon"
                      title="Скачать дамп на компьютер"
                      :disabled="downloadingBackupId === b.backup_id"
                      @click="handleDownload(b)"
                    >
                      <Download class="icon-xs" />
                    </button>

                    <!-- Восстановить -->
                    <button
                      class="btn-action-icon text-warning"
                      title="Восстановить сервис из этой копии"
                      :disabled="restoringBackupId === b.backup_id || !isServiceOnline"
                      @click="promptRestore(b)"
                    >
                      <RotateCcw class="icon-xs" />
                    </button>

                    <!-- Удалить -->
                    <button
                      class="btn-action-icon text-danger"
                      title="Удалить снимок с диска ноды"
                      :disabled="deletingBackupId === b.backup_id"
                      @click="handleDelete(b)"
                    >
                      <Trash2 class="icon-xs" />
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Модальное окно подтверждения восстановления -->
      <div v-if="showRestoreConfirm" class="confirm-overlay" @click.self="showRestoreConfirm = false">
        <div class="confirm-modal card">
          <div class="confirm-icon-box">
            <AlertTriangle class="confirm-icon text-warning" />
          </div>
          <h4>Восстановить данные из бэкапа?</h4>
          <p class="confirm-text">
            Вы собираетесь восстановить сервис <strong>{{ service.name }}</strong> из снимка
            <code>{{ targetRestoreBackup?.file_name || targetRestoreBackup?.backup_id }}</code>.
          </p>
          <div class="confirm-warning">
            ⚠️ <strong>Внимание:</strong> Текущие данные базы данных будут <u>полностью перезаписаны</u> данными из этого архива.
          </div>
          <div class="confirm-actions">
            <button
              class="btn-danger btn-sm"
              :disabled="restoringBackupId !== null"
              @click="confirmRestore"
            >
              <span v-if="restoringBackupId" class="spinner-inline"></span>
              <span>{{ restoringBackupId ? 'Восстановление...' : 'Да, перезаписать и восстановить' }}</span>
            </button>
            <button class="btn-outline btn-sm" :disabled="restoringBackupId !== null" @click="showRestoreConfirm = false">
              Отмена
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</Teleport>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue';
import {
  Database,
  Layers,
  HardDrive,
  Download,
  RotateCcw,
  Trash2,
  UploadCloud,
  Check,
  AlertTriangle,
  Plus,
  FileArchive,
  Clock,
} from 'lucide-vue-next';
import {
  listServiceBackups,
  createServiceBackup,
  restoreServiceBackup,
  deleteServiceBackup,
  downloadServiceBackup,
  uploadServiceBackup,
  toggleServiceAutoBackup,
} from '@/entities/node/api/nodeApi';

const props = defineProps({
  nodeId: {
    type: [Number, String],
    required: true,
  },
  service: {
    type: Object,
    required: true,
  },
});

defineEmits(['close']);

const autoBackupEnabled = ref(props.service.auto_backup_enabled || false);

async function handleToggleAutoBackup() {
  try {
    await toggleServiceAutoBackup(props.nodeId, props.service.name, autoBackupEnabled.value);
    if (autoBackupEnabled.value) {
      successMessage.value = "Автоматическое регулярное резервное копирование включено";
    } else {
      successMessage.value = "Автоматическое регулярное резервное копирование отключено";
    }
  } catch (err) {
    error.value = "Ошибка при переключении автобэкапов: " + (err.response?.data?.message || err.message);
    autoBackupEnabled.value = !autoBackupEnabled.value;
  }
}

async function loadBackups() {
  loading.value = true;
  error.value = null;
  try {
    const res = await listServiceBackups(props.nodeId, props.service.name);
    backups.value = res;
  } catch (err) {
    error.value = err.response?.data?.message || err.message || 'Ошибка загрузки бэкапов';
  } finally {
    loading.value = false;
  }
}

async function handleCreateBackup() {
  creating.value = true;
  error.value = null;
  successMessage.value = null;
  try {
    const backup = await createServiceBackup(props.nodeId, props.service.name);
    successMessage.value = `Бэкап успешно создан (${formatBytes(backup?.size_bytes || 0)})`;
    await loadBackups();
  } catch (err) {
    error.value = err.response?.data?.message || err.message || 'Ошибка создания бэкапа';
  } finally {
    creating.value = false;
  }
}

function promptRestore(b) {
  targetRestoreBackup.value = b;
  showRestoreConfirm.value = true;
}

async function confirmRestore() {
  if (!targetRestoreBackup.value) return;
  const b = targetRestoreBackup.value;
  restoringBackupId.value = b.backup_id;
  error.value = null;
  successMessage.value = null;
  try {
    await restoreServiceBackup(props.nodeId, props.service.name, b.backup_id);
    successMessage.value = `Данные сервиса ${props.service.name} успешно восстановлены из бэкапа!`;
    showRestoreConfirm.value = false;
  } catch (err) {
    error.value = err.response?.data?.message || err.message || 'Ошибка восстановления';
  } finally {
    restoringBackupId.value = null;
  }
}

async function handleDelete(b) {
  if (!confirm(`Удалить резервную копию ${b.file_name || b.backup_id}?`)) {
    return;
  }
  deletingBackupId.value = b.backup_id;
  error.value = null;
  try {
    await deleteServiceBackup(props.nodeId, props.service.name, b.backup_id);
    backups.value = backups.value.filter((item) => item.backup_id !== b.backup_id);
  } catch (err) {
    error.value = err.response?.data?.message || err.message || 'Ошибка удаления бэкапа';
  } finally {
    deletingBackupId.value = null;
  }
}

async function handleDownload(b) {
  downloadingBackupId.value = b.backup_id;
  try {
    await downloadServiceBackup(props.nodeId, props.service.name, b.backup_id, b.file_name);
  } catch (err) {
    error.value = err.response?.data?.message || err.message || 'Ошибка скачивания';
  } finally {
    downloadingBackupId.value = null;
  }
}

function handleFileChange(e) {
  const file = e.target.files?.[0];
  if (file) {
    uploadFile.value = file;
  }
}

function handleDrop(e) {
  const file = e.dataTransfer?.files?.[0];
  if (file) {
    uploadFile.value = file;
  }
}

function cancelUpload() {
  uploadFile.value = null;
  showUploadForm.value = false;
  uploadProgress.value = 0;
}

async function submitUpload() {
  if (!uploadFile.value) return;
  uploading.value = true;
  uploadProgress.value = 0;
  error.value = null;
  successMessage.value = null;
  try {
    const backup = await uploadServiceBackup(
      props.nodeId,
      props.service.name,
      uploadFile.value,
      restoreImmediately.value,
      (p) => {
        uploadProgress.value = p;
      }
    );
    successMessage.value = restoreImmediately.value
      ? 'Бэкап успешно загружен и восстановлен в базе данных!'
      : 'Бэкап успешно сохранен на ноде!';
    cancelUpload();
    await loadBackups();
  } catch (err) {
    error.value = err.response?.data?.message || err.message || 'Ошибка загрузки файла';
  } finally {
    uploading.value = false;
  }
}

// Форматирование
function formatBytes(bytes) {
  if (!bytes || bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

function formatDate(val) {
  if (!val) return '—';
  const d = new Date(val);
  if (isNaN(d.getTime())) return val;
  return d.toLocaleString('ru-RU', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

function formatBackupType(type) {
  if (type === 'BACKUP_TYPE_UPLOADED' || type === 2) return 'Загруженный';
  if (type === 'BACKUP_TYPE_SCHEDULED' || type === 3) return 'Авто';
  return 'Ручной';
}

function getTypeBadgeClass(type) {
  if (type === 'BACKUP_TYPE_UPLOADED' || type === 2) return 'badge-info';
  return 'badge-manual';
}

function formatStatus(status) {
  if (status === 'BACKUP_STATUS_READY' || status === 2) return 'Готов';
  if (status === 'BACKUP_STATUS_CREATING' || status === 1) return 'Создание...';
  if (status === 'BACKUP_STATUS_RESTORING' || status === 4) return 'Восстановление...';
  if (status === 'BACKUP_STATUS_FAILED' || status === 3) return 'Ошибка';
  return 'Готов';
}

function getStatusBadgeClass(status) {
  if (status === 'BACKUP_STATUS_READY' || status === 2) return 'success';
  if (status === 'BACKUP_STATUS_CREATING' || status === 1 || status === 'BACKUP_STATUS_RESTORING' || status === 4) return 'warning';
  if (status === 'BACKUP_STATUS_FAILED' || status === 3) return 'danger';
  return 'success';
}

function getServiceIcon(type) {
  switch (type) {
    case 'postgres':
    case 'mysql':
      return Database;
    case 'redis':
      return Layers;
    case 'volume':
      return HardDrive;
    default:
      return Database;
  }
}

function getServiceBgColor(type) {
  switch (type) {
    case 'postgres':
      return 'rgba(51, 103, 145, 0.2)';
    case 'redis':
      return 'rgba(216, 44, 32, 0.2)';
    case 'mysql':
      return 'rgba(242, 145, 17, 0.2)';
    case 'volume':
      return 'rgba(46, 160, 67, 0.2)';
    default:
      return 'rgba(88, 166, 255, 0.2)';
  }
}

function getServiceColor(type) {
  switch (type) {
    case 'postgres':
      return '#336791';
    case 'redis':
      return '#ea4335';
    case 'mysql':
      return '#f29111';
    case 'volume':
      return '#2ea043';
    default:
      return '#58a6ff';
  }
}

onMounted(() => {
  loadBackups();
});
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  width: 100vw;
  height: 100vh;
  background: rgba(0, 0, 0, 0.75);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  padding: 16px;
  box-sizing: border-box;
}

.backups-modal {
  width: 100%;
  max-width: 860px;
  max-height: 88vh;
  display: flex;
  flex-direction: column;
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border-color, #30363d);
  border-radius: 12px;
  box-shadow: 0 24px 48px rgba(0, 0, 0, 0.5);
  padding: 0;
  overflow: hidden;
  position: relative;
  animation: modal-appear 0.2s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes modal-appear {
  from {
    opacity: 0;
    transform: scale(0.96) translateY(-10px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

/* ─── Auto Backup Strip ─────────────────────────────────── */
.auto-backup-strip {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 24px;
  background: rgba(88, 166, 255, 0.05);
  border-bottom: 1px solid var(--border-color, #30363d);
  gap: 16px;
}

.auto-backup-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.auto-backup-clock {
  color: var(--primary, #58a6ff);
  flex-shrink: 0;
}

.auto-backup-title {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-color, #c9d1d9);
}

.auto-backup-desc {
  font-size: 0.75rem;
  color: var(--text-muted, #8b949e);
  margin-top: 1px;
}

.auto-backup-toggle {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-shrink: 0;
}

.switch-toggle {
  position: relative;
  display: inline-block;
  width: 38px;
  height: 20px;
  cursor: pointer;
}

.switch-toggle input {
  opacity: 0;
  width: 0;
  height: 0;
}

.switch-slider {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: var(--border-color, #30363d);
  transition: 0.2s ease;
  border-radius: 20px;
}

.switch-slider:before {
  position: absolute;
  content: "";
  height: 14px;
  width: 14px;
  left: 3px;
  bottom: 3px;
  background-color: #fff;
  transition: 0.2s ease;
  border-radius: 50%;
}

.switch-toggle input:checked + .switch-slider {
  background-color: #3fb950;
}

.switch-toggle input:checked + .switch-slider:before {
  transform: translateX(18px);
}

.switch-text {
  font-size: 0.8rem;
  font-weight: 600;
  min-width: 68px;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 18px 24px;
  border-bottom: 1px solid var(--border-color, #30363d);
}

.modal-header-info {
  display: flex;
  align-items: center;
  gap: 14px;
}

.service-badge-icon {
  width: 42px;
  height: 42px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.title-with-badge {
  display: flex;
  align-items: center;
  gap: 10px;
}

.title-with-badge h3 {
  margin: 0;
  font-size: 1.15rem;
  font-weight: 600;
  color: var(--text-color, #c9d1d9);
}

.service-name-tag {
  font-family: monospace;
  font-size: 0.8rem;
  background-color: var(--bg-tertiary, #21262d);
  border: 1px solid var(--border-color, #30363d);
  padding: 2px 8px;
  border-radius: 6px;
  color: #58a6ff;
}

.modal-subtitle {
  margin: 4px 0 0 0;
  font-size: 0.82rem;
  color: var(--text-muted, #8b949e);
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.5rem;
  color: var(--text-muted, #8b949e);
  cursor: pointer;
  padding: 0 4px;
  line-height: 1;
}

.close-btn:hover {
  color: var(--text-color, #c9d1d9);
}

/* Оповещения */
.alert-banner {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 16px;
  font-size: 0.85rem;
  margin: 12px 24px 0 24px;
  border-radius: 6px;
}

.alert-danger {
  background-color: rgba(248, 81, 73, 0.15);
  border: 1px solid rgba(248, 81, 73, 0.4);
  color: #f85149;
}

.alert-success {
  background-color: rgba(46, 160, 67, 0.15);
  border: 1px solid rgba(46, 160, 67, 0.4);
  color: #3fb950;
}

.alert-close {
  margin-left: auto;
  background: none;
  border: none;
  font-size: 1.1rem;
  color: inherit;
  cursor: pointer;
}

/* Верхняя панель */
.actions-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 24px;
  background-color: var(--bg-secondary, #161b22);
  border-bottom: 1px solid var(--border-color, #30363d);
}

.actions-left {
  display: flex;
  gap: 10px;
}

.icon-btn-secondary {
  background: var(--bg-tertiary, #21262d);
  border: 1px solid var(--border-color, #30363d);
  border-radius: 6px;
  color: var(--text-muted, #8b949e);
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.icon-btn-secondary:hover:not(:disabled) {
  color: var(--text-color, #c9d1d9);
  background-color: #30363d;
}

.spin-icon {
  animation: spin 1s linear infinite;
}

/* Секция загрузки */
.upload-section {
  padding: 16px 24px;
  background-color: rgba(88, 166, 255, 0.04);
  border-bottom: 1px solid var(--border-color, #30363d);
}

.upload-dropzone {
  border: 2px dashed var(--border-color, #30363d);
  border-radius: 8px;
  padding: 18px;
  text-align: center;
  background-color: var(--bg-secondary, #161b22);
  cursor: pointer;
  transition: all 0.2s ease;
}

.upload-dropzone:hover {
  border-color: #58a6ff;
}

.upload-dropzone.has-file {
  border-style: solid;
  border-color: #3fb950;
  cursor: default;
}

.file-input-hidden {
  display: none;
}

.dropzone-icon {
  width: 32px;
  height: 32px;
  color: #58a6ff;
  margin-bottom: 8px;
}

.dropzone-title {
  margin: 0;
  font-size: 0.88rem;
  color: var(--text-color, #c9d1d9);
}

.dropzone-title span {
  color: #58a6ff;
  text-decoration: underline;
}

.dropzone-hint {
  margin: 4px 0 0 0;
  font-size: 0.75rem;
  color: var(--text-muted, #8b949e);
}

.dropzone-selected {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 6px 10px;
}

.file-icon {
  width: 28px;
  height: 28px;
  color: #3fb950;
}

.selected-meta {
  display: flex;
  flex-direction: column;
  text-align: left;
  flex: 1;
}

.selected-name {
  font-size: 0.88rem;
  font-weight: 500;
  color: var(--text-color, #c9d1d9);
}

.selected-size {
  font-size: 0.75rem;
  color: var(--text-muted, #8b949e);
}

.remove-file-btn {
  background: none;
  border: none;
  font-size: 1.25rem;
  color: #f85149;
  cursor: pointer;
}

.upload-options {
  margin-top: 12px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.82rem;
  color: var(--text-color, #c9d1d9);
  cursor: pointer;
}

.progress-bar-wrap {
  display: flex;
  align-items: center;
  gap: 10px;
}

.progress-track {
  flex: 1;
  height: 6px;
  background-color: var(--bg-tertiary, #21262d);
  border-radius: 3px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background-color: #58a6ff;
  transition: width 0.2s ease;
}

.progress-text {
  font-size: 0.75rem;
  color: var(--text-muted, #8b949e);
  min-width: 90px;
}

.upload-buttons {
  display: flex;
  gap: 8px;
}

/* Контент бэкапов */
.backups-content {
  flex: 1;
  overflow-y: auto;
  min-height: 220px;
}

.loading-state,
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px 24px;
  gap: 12px;
  color: var(--text-muted, #8b949e);
}

.empty-icon {
  width: 40px;
  height: 40px;
  color: var(--border-color, #30363d);
}

.empty-title {
  font-size: 0.95rem;
  font-weight: 500;
  color: var(--text-color, #c9d1d9);
  margin: 0;
}

.empty-desc {
  font-size: 0.82rem;
  color: var(--text-muted, #8b949e);
  margin: 0;
  max-width: 380px;
  text-align: center;
}

.table-wrap {
  overflow-x: auto;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.85rem;
}

.data-table th {
  padding: 10px 16px;
  text-align: left;
  color: var(--text-muted, #8b949e);
  font-weight: 500;
  font-size: 0.78rem;
  border-bottom: 1px solid var(--border-color, #30363d);
  background-color: var(--bg-secondary, #161b22);
}

.data-table td {
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-color, #30363d);
  color: var(--text-color, #c9d1d9);
}

.cell-file {
  max-width: 260px;
}

.file-name-wrap {
  display: flex;
  align-items: center;
  gap: 10px;
}

.file-name {
  display: block;
  font-family: monospace;
  font-size: 0.82rem;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 220px;
}

.checksum-hint {
  display: block;
  font-size: 0.72rem;
  color: var(--text-muted, #8b949e);
  font-family: monospace;
}

.cell-size,
.cell-date {
  font-size: 0.8rem;
  color: var(--text-muted, #8b949e);
  white-space: nowrap;
}

/* Бейджи */
.type-badge {
  font-size: 0.72rem;
  padding: 2px 7px;
  border-radius: 12px;
  font-weight: 500;
}

.badge-manual {
  background-color: rgba(88, 166, 255, 0.15);
  color: #58a6ff;
  border: 1px solid rgba(88, 166, 255, 0.3);
}

.badge-info {
  background-color: rgba(163, 113, 247, 0.15);
  color: #d2a8ff;
  border: 1px solid rgba(163, 113, 247, 0.3);
}

.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 0.75rem;
  padding: 2px 8px;
  border-radius: 12px;
  font-weight: 500;
}

.status-badge.success {
  background-color: rgba(46, 160, 67, 0.15);
  color: #3fb950;
}

.status-badge.warning {
  background-color: rgba(210, 153, 34, 0.15);
  color: #d29922;
}

.status-badge.danger {
  background-color: rgba(248, 81, 73, 0.15);
  color: #f85149;
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background-color: currentColor;
}

/* Кнопки действий */
.col-actions {
  text-align: right;
  white-space: nowrap;
}

.action-buttons-cell {
  display: flex;
  justify-content: flex-end;
  gap: 6px;
}

.btn-action-icon {
  background: var(--bg-tertiary, #21262d);
  border: 1px solid var(--border-color, #30363d);
  border-radius: 6px;
  color: var(--text-muted, #8b949e);
  width: 30px;
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-action-icon:hover:not(:disabled) {
  color: var(--text-color, #c9d1d9);
  background-color: #30363d;
  border-color: #8b949e;
}

.btn-action-icon.text-warning:hover:not(:disabled) {
  color: #d29922;
  border-color: #d29922;
}

.btn-action-icon.text-danger:hover:not(:disabled) {
  color: #f85149;
  border-color: #f85149;
}

/* Модальное окно подтверждения */
.confirm-overlay {
  position: absolute;
  inset: 0;
  background-color: rgba(13, 17, 23, 0.85);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10;
  padding: 20px;
}

.confirm-modal {
  width: 100%;
  max-width: 440px;
  padding: 24px;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.confirm-icon-box {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background-color: rgba(210, 153, 34, 0.15);
  display: flex;
  align-items: center;
  justify-content: center;
}

.confirm-icon {
  width: 24px;
  height: 24px;
}

.confirm-modal h4 {
  margin: 0;
  font-size: 1.1rem;
  color: var(--text-color, #c9d1d9);
}

.confirm-text {
  margin: 0;
  font-size: 0.85rem;
  color: var(--text-muted, #8b949e);
  line-height: 1.4;
}

.confirm-warning {
  background-color: rgba(248, 81, 73, 0.1);
  border: 1px solid rgba(248, 81, 73, 0.3);
  padding: 10px 14px;
  border-radius: 6px;
  font-size: 0.8rem;
  color: #f85149;
  text-align: left;
}

.confirm-actions {
  display: flex;
  gap: 10px;
  margin-top: 6px;
  width: 100%;
  justify-content: center;
}

.spinner-inline {
  display: inline-block;
  width: 12px;
  height: 12px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
