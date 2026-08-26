<template>
  <div class="client-build-uploader">
    <div class="input-group">
      <div class="input-header">
        <label class="input-label">{{ t('common.version') }} <span class="req">*</span></label>
      </div>
      <input
        type="text"
        v-model="newBuildVersion"
        class="input-control"
        placeholder="1.0.0"
        :disabled="buildStatus !== 'idle'"
      />
    </div>

    <!-- Ожидание загрузки (Dropzone) -->
    <div
      v-if="buildStatus === 'idle'"
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
        <button
          type="button"
          class="btn-select-file"
          @click.stop="$refs.fileZip.click()"
        >
          {{ t('projectDraft.selectFile') }}
        </button>
      </div>
    </div>

    <!-- Идет загрузка -->
    <div v-if="buildStatus === 'uploading'" class="upload-progress-box">
      <div class="prog-info">
        <span style="font-weight: 600; color: var(--text-main);">
          {{ t('common.loading') }}
        </span>
        <span style="font-weight: 600; color: var(--primary);">
          {{ buildProgress }}%
        </span>
      </div>
      <div class="prog-bg">
        <div
          class="prog-fill"
          :style="{ width: buildProgress + '%' }"
        ></div>
      </div>
    </div>

    <!-- Загрузка завершена -->
    <div v-if="buildStatus === 'done'" class="upload-success-box">
      <CheckCircle class="icon-md text-green" />
      <div>
        <span style="display: block; font-weight: 600;">
          {{ t('common.version') }} v{{ uploadedVersion }} {{ t('common.saved') }}!
        </span>
        <button class="btn-text mt-8" @click="resetBuildUpload">
          {{ t('common.upload') }}
        </button>
      </div>
    </div>

    <!-- Hidden file input -->
    <input
      type="file"
      ref="fileZip"
      accept=".zip,.tar.gz"
      hidden
      @change="handleZipUpload"
    />
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { UploadCloud, CheckCircle } from 'lucide-vue-next';
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
const buildStatus = ref('idle');
const buildProgress = ref(0);
const uploadedVersion = ref('');
const isDragging = ref(false);

function onDragOver(e) {
  isDragging.value = true;
}

function onDragLeave(e) {
  isDragging.value = false;
}

function onDrop(e) {
  isDragging.value = false;
  const files = e.dataTransfer?.files;
  if (files && files.length > 0) {
    processFile(files[0]);
  }
}

async function checkZipForIndexHtml(arrayBuffer) {
  const zip = await JSZip.loadAsync(arrayBuffer);
  const entry = zip.file('index.html');
  if (!entry) throw new Error('В корне архива отсутствует index.html');
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
    if (name === 'index.html') return true;
    if (name.length === 0) break;
    let sizeStr = '';
    for (let i = 124; i < 136; i++) {
      sizeStr += String.fromCharCode(inflated[offset + i]);
    }
    const size = parseInt(sizeStr.trim(), 8) || 0;
    offset += 512 + Math.ceil(size / 512) * 512;
  }
  throw new Error('В корне архива отсутствует index.html');
}

async function processFile(file) {
  if (!file) return;

  const version = newBuildVersion.value.trim();
  if (!version) {
    showToast('Укажите версию билда перед загрузкой', 'danger');
    if (fileZip.value) fileZip.value.value = '';
    return;
  }

  const name = file.name.toLowerCase();
  const isZip = name.endsWith('.zip');
  const isTarGz = name.endsWith('.tar.gz');

  if (!isZip && !isTarGz) {
    showToast('Допустимые форматы: .zip и .tar.gz', 'danger');
    if (fileZip.value) fileZip.value.value = '';
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
    showToast(err.message, 'danger');
    if (fileZip.value) fileZip.value.value = '';
    return;
  }

  buildStatus.value = 'uploading';
  buildProgress.value = 0;

  try {
    await uploadClientBuild(props.projectId, version, file, (p) => {
      buildProgress.value = p;
    });
    buildStatus.value = 'done';
    uploadedVersion.value = version;
    newBuildVersion.value = '';
    showToast('Билд развернут в Dev-среде!', 'success');
    emit('buildUploaded', version);
  } catch (err) {
    buildStatus.value = 'idle';
    showToast('Ошибка загрузки билда', 'danger');
  } finally {
    if (fileZip.value) fileZip.value.value = '';
  }
}

async function handleZipUpload(event) {
  const file = event.target.files[0];
  if (file) {
    await processFile(file);
  }
}

function resetBuildUpload() {
  buildStatus.value = 'idle';
  buildProgress.value = 0;
  uploadedVersion.value = '';
}
</script>

<style scoped>
.client-build-uploader {
  margin-bottom: 20px;
}

.input-group {
  margin-bottom: 16px;
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
  color: var(--text-main);
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.req {
  color: var(--danger);
  font-weight: 700;
}

.input-control {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md, 8px);
  background: var(--bg-input);
  font-family: inherit;
  box-sizing: border-box;
  color: var(--text-main);
  outline: none;
  transition: border-color 0.2s, background-color 0.2s;
}

.input-control:focus {
  border-color: var(--primary);
  background: var(--bg-card);
}

.dropzone {
  border: 2px dashed var(--border);
  border-radius: var(--radius-md, 8px);
  padding: 24px 20px;
  background: var(--bg-card);
  cursor: pointer;
  transition: all 0.2s ease;
  box-sizing: border-box;
}

.dropzone:hover {
  border-color: var(--primary);
  background: var(--bg-hover);
}

.dropzone.is-dragging {
  border-color: var(--primary) !important;
  background: var(--primary-light, rgba(88, 166, 255, 0.08)) !important;
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
  background: var(--bg-secondary);
  color: var(--primary);
  display: flex;
  align-items: center;
  justify-content: center;
}

.icon-md {
  width: 20px;
  height: 20px;
}

.upload-prompt-text {
  margin: 0;
  font-size: 0.85rem;
  color: var(--text-muted);
}

.upload-prompt-sub {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.btn-select-file {
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm, 6px);
  color: var(--text-main);
  font-size: 0.82rem;
  font-weight: 600;
  padding: 6px 14px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-select-file:hover {
  border-color: var(--primary);
  color: var(--primary);
  background: var(--bg-card);
}

.upload-progress-box {
  padding: 20px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md, 8px);
  background: var(--bg-secondary);
}

.prog-info {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
  font-size: 0.85rem;
}

.prog-bg {
  width: 100%;
  height: 8px;
  background: var(--border);
  border-radius: 4px;
  overflow: hidden;
}

.prog-fill {
  height: 100%;
  background: var(--primary);
  transition: width 0.2s;
}

.upload-success-box {
  padding: 20px;
  border: 1px solid var(--success);
  border-radius: var(--radius-md, 8px);
  background: var(--success-light);
  display: flex;
  align-items: center;
  gap: 16px;
}

.text-green {
  color: var(--success);
}

.btn-text {
  background: none;
  border: none;
  color: var(--primary);
  cursor: pointer;
  padding: 0;
  font-size: 0.85rem;
  font-weight: 600;
}

.mt-8 {
  margin-top: 8px;
}
</style>
