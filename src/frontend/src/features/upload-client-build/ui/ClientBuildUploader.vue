<template>
  <div class="client-build-uploader">
    <!-- Верхняя строка формы: Ввод версии + кнопка "Загрузить сборку" справа -->
    <div class="build-form-header-row">
      <div class="input-group version-input-group">
        <div class="input-header">
          <label class="input-label">{{ t('common.version') }} <span class="req">*</span></label>
        </div>
        <input
          v-model="newBuildVersion"
          type="text"
          class="input-control"
          placeholder="1.0.0"
          :disabled="buildStatus === 'uploading'"
        />
      </div>

      <div class="upload-btn-wrap">
        <label class="input-label-placeholder">&nbsp;</label>
        <button
          type="button"
          class="btn-upload-build-action"
          :disabled="!canUpload"
          @click="startUpload"
        >
          <Loader2 v-if="buildStatus === 'uploading'" class="icon-sm spin" />
          <UploadCloud v-else class="icon-sm" />
          <span>{{ buildStatus === 'uploading' ? t('common.loading') : 'Загрузить сборку' }}</span>
        </button>
      </div>
    </div>

    <!-- Область прикрепления файла: Прикреплен локально (Staged) -->
    <div v-if="buildStatus === 'idle' && stagedFile" class="staged-file-card">
      <div class="staged-file-left">
        <div class="file-icon-box">
          <FileArchive class="icon-md text-primary" />
        </div>
        <div class="file-info-text">
          <div class="file-name-row">
            <span class="file-name">{{ stagedFile.name }}</span>
            <span class="file-size">({{ formatFileSize(stagedFile.size) }})</span>
          </div>
          <span class="file-ready-badge">✓ Файл прикреплен и готов к отправке</span>
        </div>
      </div>

      <div class="staged-file-actions">
        <button type="button" class="btn-action-text" @click="$refs.fileZip.click()">
          {{ t('projectDraft.replaceFile') }}
        </button>
        <button type="button" class="btn-action-text text-danger" @click="removeStagedFile">
          {{ t('common.delete') }}
        </button>
      </div>
    </div>

    <!-- Область прикрепления файла: Dropzone (когда файл еще не выбран) -->
    <div
      v-else-if="buildStatus === 'idle'"
      class="dropzone"
      :class="{ 'is-dragging': isDragging }"
      @dragover.prevent="onDragOver"
      @dragleave.prevent="onDragLeave"
      @drop.prevent="onDrop"
      @click="$refs.fileZip.click()"
    >
      <div class="upload-prompt-content">
        <div class="upload-icon-circle">
          <UploadCloud class="icon-md" />
        </div>
        <p class="upload-prompt-text">
          {{ t('projectDraft.dropOrClick') }}
        </p>
        <span class="upload-prompt-sub">
          {{ t('projectDraft.uploadFormats') }}
        </span>
        <button type="button" class="btn-select-file" @click.stop="$refs.fileZip.click()">
          {{ t('projectDraft.selectFile') }}
        </button>
      </div>
    </div>

    <!-- Идет загрузка -->
    <div v-if="buildStatus === 'uploading'" class="upload-progress-box">
      <div class="prog-info">
        <span class="prog-label">
          <Loader2 class="icon-xs spin text-primary" />
          <span>Загрузка и распаковка сборки в Dev-окружение...</span>
        </span>
        <span class="prog-percent"> {{ buildProgress }}% </span>
      </div>
      <div class="prog-bg">
        <div class="prog-fill" :style="{ width: buildProgress + '%' }"></div>
      </div>
      <div v-if="stagedFile" class="prog-subtext">
        {{ stagedFile.name }} ({{ formatFileSize(stagedFile.size) }})
      </div>
    </div>

    <!-- Hidden file input -->
    <input
      ref="fileZip"
      type="file"
      accept=".zip,.tar.gz,.tgz"
      hidden
      @change="handleZipSelected"
    />
  </div>
</template>

<script setup>
import { ref, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { UploadCloud, FileArchive, Loader2 } from 'lucide-vue-next';
import JSZip from 'jszip';
import pako from 'pako';
import { uploadClientBuild } from '@/entities/build';
import { showToast } from '@/shared/lib';

const { t } = useI18n();

const props = defineProps({
  projectId: { type: [String, Number], required: true },
});

const emit = defineEmits(['buildUploaded']);

const fileZip = ref(null);
const newBuildVersion = ref('');
const stagedFile = ref(null);
const buildStatus = ref('idle');
const buildProgress = ref(0);
const isDragging = ref(false);

const canUpload = computed(() => {
  return !!newBuildVersion.value.trim() && !!stagedFile.value && buildStatus.value !== 'uploading';
});

function formatFileSize(bytes) {
  if (!bytes) return '0 B';
  const k = 1024;
  const sizes = ['B', 'КБ', 'МБ', 'ГБ'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
}

function onDragOver(e) {
  e.preventDefault();
  isDragging.value = true;
}

function onDragLeave(e) {
  e.preventDefault();
  isDragging.value = false;
}

function onDrop(e) {
  e.preventDefault();
  isDragging.value = false;
  const files = e.dataTransfer?.files;
  if (files && files.length > 0) {
    stageFile(files[0]);
  }
}

function stageFile(file) {
  if (!file) return;

  const name = file.name.toLowerCase();
  const isZip = name.endsWith('.zip');
  const isTarGz = name.endsWith('.tar.gz') || name.endsWith('.tgz');

  if (!isZip && !isTarGz) {
    showToast('Допустимые форматы: .zip и .tar.gz', 'danger');
    if (fileZip.value) fileZip.value.value = '';
    return;
  }

  stagedFile.value = file;
  showToast(`Файл "${file.name}" выбран`, 'info');
}

function handleZipSelected(event) {
  const file = event.target.files[0];
  if (file) {
    stageFile(file);
  }
  event.target.value = '';
}

function removeStagedFile() {
  stagedFile.value = null;
  if (fileZip.value) fileZip.value.value = '';
}

async function checkZipForIndexHtml(arrayBuffer) {
  const zip = await JSZip.loadAsync(arrayBuffer);
  let hasIndex = false;
  zip.forEach((relativePath) => {
    const clean = relativePath.replace(/\\/g, '/');
    if (clean === 'index.html' || clean.endsWith('/index.html')) {
      hasIndex = true;
    }
  });
  if (!hasIndex) {
    throw new Error('В архиве отсутствует файл index.html');
  }
}

function checkTarGzForIndexHtml(arrayBuffer) {
  const inflated = pako.inflate(new Uint8Array(arrayBuffer));
  let offset = 0;
  while (offset < inflated.length) {
    let name = '';
    for (let i = 0; i < 100; i++) {
      if (inflated[offset + i] === 0) break;
      name += String.fromCharCode(inflated[offset + i]);
    }
    if (name === 'index.html' || name.endsWith('/index.html')) return true;
    if (name.length === 0) break;
    let sizeStr = '';
    for (let i = 124; i < 136; i++) {
      sizeStr += String.fromCharCode(inflated[offset + i]);
    }
    const size = parseInt(sizeStr.trim(), 8) || 0;
    offset += 512 + Math.ceil(size / 512) * 512;
  }
  throw new Error('В архиве отсутствует файл index.html');
}

async function startUpload() {
  if (!canUpload.value) {
    if (!newBuildVersion.value.trim()) {
      showToast('Укажите версию билда', 'danger');
      return;
    }
    if (!stagedFile.value) {
      showToast('Прикрепите архив со сборкой (.zip или .tar.gz)', 'danger');
      return;
    }
    return;
  }

  const file = stagedFile.value;
  const version = newBuildVersion.value.trim();
  const name = file.name.toLowerCase();
  const isZip = name.endsWith('.zip');
  const isTarGz = name.endsWith('.tar.gz') || name.endsWith('.tgz');

  if (!isZip && !isTarGz) {
    showToast('Допустимые форматы: .zip и .tar.gz', 'danger');
    return;
  }

  try {
    const buffer = await file.arrayBuffer();
    if (isZip) {
      await checkZipForIndexHtml(buffer);
    } else {
      checkTarGzForIndexHtml(buffer);
    }
  } catch (err) {
    showToast(err.message || 'Ошибка валидации архива', 'danger');
    return;
  }

  buildStatus.value = 'uploading';
  buildProgress.value = 0;

  try {
    await uploadClientBuild(props.projectId, version, file, (p) => {
      buildProgress.value = p;
    });
    buildStatus.value = 'idle';
    buildProgress.value = 0;
    newBuildVersion.value = '';
    stagedFile.value = null;
    if (fileZip.value) fileZip.value.value = '';
    showToast(`Версия v${version} успешно загружена и развернута в Dev!`, 'success');
    emit('buildUploaded', version);
  } catch (err) {
    buildStatus.value = 'idle';
    showToast(err.response?.data?.message || err.message || 'Ошибка загрузки билда', 'danger');
  }
}
</script>

<style scoped>
.client-build-uploader {
  margin-bottom: 20px;
}

.build-form-header-row {
  display: flex;
  align-items: flex-end;
  gap: 12px;
  margin-bottom: 16px;
}

.version-input-group {
  flex: 1;
  margin-bottom: 0;
}

.upload-btn-wrap {
  display: flex;
  flex-direction: column;
}

.input-label-placeholder {
  font-size: 0.85rem;
  margin-bottom: 8px;
  visibility: hidden;
}

.btn-upload-build-action {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  height: 42px;
  padding: 0 18px;
  background: var(--primary, #58a6ff);
  border: 1px solid var(--primary, #58a6ff);
  color: #ffffff;
  border-radius: var(--radius-md, 8px);
  font-size: 0.88rem;
  font-weight: 600;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.2s ease;
}

.btn-upload-build-action:hover:not(:disabled) {
  background: var(--primary-hover, #79c0ff);
  border-color: var(--primary-hover, #79c0ff);
}

.btn-upload-build-action:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.input-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.input-label {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.req {
  color: var(--danger, #f85149);
  font-weight: 700;
}

.input-control {
  width: 100%;
  height: 42px;
  padding: 0 12px;
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-md, 8px);
  background: var(--bg-input, #0d1117);
  font-family: inherit;
  box-sizing: border-box;
  color: var(--text-main, #f0f6fc);
  outline: none;
  transition:
    border-color 0.2s,
    background-color 0.2s;
}

.input-control:focus {
  border-color: var(--primary, #58a6ff);
  background: var(--bg-card, #161b22);
}

/* Карточка выбранного архива */
.staged-file-card {
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-md, 8px);
  padding: 16px 20px;
  background: var(--bg-secondary, #161b22);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  box-sizing: border-box;
}

.staged-file-left {
  display: flex;
  align-items: center;
  gap: 14px;
}

.file-icon-box {
  width: 42px;
  height: 42px;
  border-radius: var(--radius-sm, 6px);
  background: var(--bg-tertiary, #21262d);
  border: 1px solid var(--border, #30363d);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.file-info-text {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.file-name-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.file-name {
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
  word-break: break-all;
}

.file-size {
  font-size: 0.8rem;
  color: var(--text-tertiary, #8b949e);
}

.file-ready-badge {
  font-size: 0.78rem;
  font-weight: 500;
  color: #2ecc71;
}

.staged-file-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-shrink: 0;
}

.btn-action-text {
  background: none;
  border: none;
  color: var(--primary, #58a6ff);
  font-size: 0.82rem;
  font-weight: 500;
  cursor: pointer;
  padding: 4px 6px;
  border-radius: 4px;
  transition: all 0.15s;
}

.btn-action-text:hover {
  text-decoration: underline;
}

.btn-action-text.text-danger {
  color: var(--danger, #f85149);
}

/* Dropzone */
.dropzone {
  border: 2px dashed var(--border, #30363d);
  border-radius: var(--radius-md, 8px);
  padding: 24px 20px;
  background: var(--bg-card, #161b22);
  cursor: pointer;
  transition: all 0.2s ease;
  box-sizing: border-box;
}

.dropzone:hover {
  border-color: var(--primary, #58a6ff);
  background: var(--bg-hover, #21262d);
}

.dropzone.is-dragging {
  border-color: var(--primary, #58a6ff) !important;
  background: rgba(88, 166, 255, 0.08) !important;
  box-shadow: 0 0 0 3px rgba(88, 166, 255, 0.2);
}

.upload-prompt-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  text-align: center;
}

.upload-icon-circle {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  background: var(--bg-secondary, #0d1117);
  color: var(--primary, #58a6ff);
  display: flex;
  align-items: center;
  justify-content: center;
}

.icon-md {
  width: 20px;
  height: 20px;
}

.icon-sm {
  width: 16px;
  height: 16px;
}

.icon-xs {
  width: 14px;
  height: 14px;
}

.upload-prompt-text {
  margin: 0;
  font-size: 0.85rem;
  color: var(--text-muted, #b0b8c4);
}

.upload-prompt-sub {
  font-size: 0.75rem;
  color: var(--text-muted, #8b949e);
}

.btn-select-file {
  background: var(--bg-secondary, #21262d);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  color: var(--text-main, #f0f6fc);
  font-size: 0.82rem;
  font-weight: 600;
  padding: 6px 14px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-select-file:hover {
  border-color: var(--primary, #58a6ff);
  color: var(--primary, #58a6ff);
  background: var(--bg-card, #161b22);
}

.upload-progress-box {
  padding: 20px;
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-md, 8px);
  background: var(--bg-secondary, #161b22);
}

.prog-info {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
  font-size: 0.85rem;
}

.prog-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
}

.prog-percent {
  font-weight: 600;
  color: var(--primary, #58a6ff);
}

.prog-bg {
  width: 100%;
  height: 8px;
  background: var(--border, #30363d);
  border-radius: 4px;
  overflow: hidden;
}

.prog-fill {
  height: 100%;
  background: var(--primary, #58a6ff);
  transition: width 0.2s;
}

.prog-subtext {
  margin-top: 8px;
  font-size: 0.78rem;
  color: var(--text-tertiary, #8b949e);
}

.upload-success-box {
  padding: 18px 20px;
  border: 1px solid rgba(46, 204, 113, 0.35);
  border-radius: var(--radius-md, 8px);
  background: rgba(46, 204, 113, 0.08);
  display: flex;
  align-items: center;
  gap: 16px;
}

.success-text-block {
  display: flex;
  flex-direction: column;
}

.success-title {
  font-size: 0.88rem;
  color: var(--text-main, #f0f6fc);
}

.text-green {
  color: #2ecc71;
}

.btn-text {
  background: none;
  border: none;
  color: var(--primary, #58a6ff);
  cursor: pointer;
  padding: 0;
  font-size: 0.85rem;
  font-weight: 600;
}

.btn-text:hover {
  text-decoration: underline;
}

.mt-8 {
  margin-top: 8px;
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>
