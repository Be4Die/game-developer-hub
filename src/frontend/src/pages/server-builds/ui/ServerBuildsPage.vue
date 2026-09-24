<template>
  <div class="builds-page tab-fade-in">
    <div class="page-header">
      <h1>{{ t('servers.buildsTitle') }}</h1>
      <button class="btn-primary" @click="showUploadForm = !showUploadForm">
        <Upload class="icon-sm" /> {{ t('servers.uploadBuild') }}
      </button>
    </div>

    <!-- Уведомление о модерации -->
    <div v-if="isUnderReview" class="moderation-notice-banner">
      <AlertCircle class="icon-sm text-warning" />
      <span>{{ t('servers.underReviewNotice') || 'Проект находится на проверке у модератора. Серверное окружение используется для тестирования игры модератором.' }}</span>
    </div>

    <!-- Ошибка -->
    <div v-if="error" class="error-banner">
      <AlertCircle class="icon-sm" /> {{ error }}
      <button class="btn-outline btn-sm" @click="fetchBuilds">
        {{ t('common.refresh') }}
      </button>
    </div>

    <!-- Форма загрузки -->
    <ServerBuildUploadModal
      v-if="showUploadForm"
      :game-id="gameId"
      @uploaded="onBuildUploaded"
      @cancel="showUploadForm = false"
    />

    <!-- Таблица билдов -->
    <div v-if="loading" class="state-container">
      <div class="spinner-md"></div>
      <p>{{ t('common.loading') }}</p>
    </div>
    <div v-else-if="error" class="empty-state"></div>
    <div v-else-if="builds.length" class="table-wrap">
      <table class="data-table">
        <thead>
          <tr>
            <th>{{ t('common.version') }}</th>
            <th>Image</th>
            <th>Protocol</th>
            <th>Port</th>
            <th>Max Players</th>
            <th>Size</th>
            <th>{{ t('common.created') }}</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="b in builds" :key="b.id">
            <td>
              <code>{{ b.build_version }}</code>
            </td>
            <td class="cell-muted">{{ b.image_tag }}</td>
            <td>{{ b.protocol }}</td>
            <td>{{ b.internal_port }}</td>
            <td>{{ b.max_players }}</td>
            <td>{{ formatBytes(b.file_size_bytes) }}</td>
            <td class="cell-muted">{{ formatDate(b.created_at) }}</td>
            <td>
              <button
                class="btn-icon"
                :title="t('common.delete')"
                :disabled="deleting"
                @click="confirmDelete(b)"
              >
                <Trash2 class="icon-sm" />
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <EmptyState
      v-else
      :icon="Layers"
      :title="t('servers.noBuilds')"
    />

    <!-- Диалог подтверждения удаления -->
    <div v-if="deleteTarget" class="modal-overlay" @click.self="deleteTarget = null">
      <div class="modal card">
        <h3>{{ t('common.delete') }}?</h3>
        <p>
          <code>{{ deleteTarget.build_version }}</code>
        </p>
        <p v-if="deleteTarget._inUse" class="text-danger">In use</p>
        <div class="modal-actions">
          <button class="btn-primary" :disabled="deleteTarget._inUse || deleting" @click="doDelete">
            {{ t('common.delete') }}
          </button>
          <button class="btn-outline" @click="deleteTarget = null">
            {{ t('common.cancel') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, inject, type Ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { Upload, Trash2, AlertCircle, Layers } from 'lucide-vue-next';
import { listServerBuilds, deleteServerBuild } from '@/entities/build';
import { listInstances } from '@/entities/instance';
import { ServerBuildUploadModal } from '@/features/upload-server-build';
import { formatBytes, formatDate, showToast } from '@/shared/lib';
import { EmptyState } from '@/shared/ui';

import type { Build, Project } from '@/shared/types';

interface DeleteTarget extends Build {
  _inUse?: boolean;
}

const { t } = useI18n();

const props = defineProps<{
  gameId: string | number;
}>();

const sharedProject = inject<Ref<Project | null> | null>('project', null);
const isUnderReview = computed<boolean>(() => sharedProject?.value?.is_under_review === true);

const builds = ref<Build[]>([]);
const loading = ref<boolean>(true);
const error = ref<string | null>(null);
const showUploadForm = ref<boolean>(false);
const deleteTarget = ref<DeleteTarget | null>(null);
const deleting = ref<boolean>(false);

async function fetchBuilds(): Promise<void> {
  loading.value = true;
  error.value = null;
  try {
    builds.value = await listServerBuilds(props.gameId);
  } catch (e: any) {
    error.value = e.response?.data?.message ?? e.message;
  } finally {
    loading.value = false;
  }
}

function onBuildUploaded(): void {
  showUploadForm.value = false;
  fetchBuilds();
}

async function confirmDelete(b: Build): Promise<void> {
  try {
    const instances = await listInstances(props.gameId);
    const inUse = instances.some(
      (i) => i.build_version === b.build_version && i.status === 'running'
    );
    deleteTarget.value = { ...b, _inUse: inUse };
  } catch {
    deleteTarget.value = { ...b, _inUse: false };
  }
}

async function doDelete(): Promise<void> {
  if (!deleteTarget.value || !deleteTarget.value.build_version) return;
  deleting.value = true;
  try {
    await deleteServerBuild(props.gameId, deleteTarget.value.build_version);
    showToast(`Билд ${deleteTarget.value.build_version} удалён`);
    deleteTarget.value = null;
    await fetchBuilds();
  } catch (e: any) {
    if (e.response?.status === 409) {
      showToast('Билд используется работающими инстансами', 'error');
      if (deleteTarget.value) {
        deleteTarget.value._inUse = true;
      }
    } else {
      showToast(e.response?.data?.message ?? 'Ошибка удаления', 'error');
    }
  } finally {
    deleting.value = false;
  }
}

onMounted(fetchBuilds);
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
.cell-muted {
  color: var(--text-muted);
}
code {
  background: var(--bg-secondary);
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 0.82rem;
}
.btn-icon {
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  display: flex;
  align-items: center;
}
.btn-icon:hover {
  color: var(--danger);
  background: var(--danger-light);
}
.btn-icon:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}
.empty-state {
  padding: 40px;
  text-align: center;
  color: var(--text-muted);
}

.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: center;
}
.modal {
  max-width: 420px;
  width: 90%;
}
.modal h3 {
  margin: 0 0 8px;
}
.modal p {
  margin: 8px 0;
  font-size: 0.9rem;
  color: var(--text-muted);
}
.modal-actions {
  display: flex;
  gap: 8px;
  margin-top: 16px;
}
.text-danger {
  color: var(--danger);
  font-weight: 600;
}

.moderation-notice-banner {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  margin-bottom: 16px;
  background: rgba(210, 153, 34, 0.12);
  border: 1px solid rgba(210, 153, 34, 0.35);
  border-radius: var(--radius-sm, 6px);
  color: var(--text-main, #f0f6fc);
  font-size: 0.88rem;
}
</style>
