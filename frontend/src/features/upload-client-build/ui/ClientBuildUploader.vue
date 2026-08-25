<template>
  <div class="client-build-uploader">
    <div class="input-group" style="margin-bottom: 16px;">
      <label>{{ t('common.version') }} <span class="req">*</span></label>
      <input
        type="text"
        v-model="newBuildVersion"
        class="input-control"
        placeholder="1.0.0"
        :disabled="buildStatus !== 'idle'"
      />
    </div>

    <!-- Ожидание загрузки -->
    <div
      v-if="buildStatus === 'idle'"
      class="dropzone"
      @click="$refs.fileZip.click()"
    >
      <UploadCloud
        style="width: 32px; height: 32px; color: var(--text-muted); margin-bottom: 8px;"
      />
      <span style="display: block; font-weight: 600;">
        {{ t('projectDraft.uploadPrompt') }}
      </span>
      <span style="display: block; font-size: 0.8rem; color: var(--text-muted); margin-top: 4px;">
        {{ t('projectDraft.uploadFormats') }}
      </span>
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

async function handleZipUpload(event) {
  const file = event.target.files[0];
  if (!file) return;

  const version = newBuildVersion.value.trim();
  if (!version) {
    showToast('Укажите версию билда перед загрузкой', 'danger');
    event.target.value = '';
    return;
  }

  const name = file.name.toLowerCase();
  const isZip = name.endsWith('.zip');
  const isTarGz = name.endsWith('.tar.gz');

  if (!isZip && !isTarGz) {
    showToast('Допустимые форматы: .zip и .tar.gz', 'danger');
    event.target.value = '';
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
    event.target.value = '';
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
  }
}

function resetBuildUpload() {
  buildStatus.value = 'idle';
  buildProgress.value = 0;
  uploadedVersion.value = '';
}
</script>

<style scoped>
.input-group label {
  display: block;
  font-size: 0.85rem;
  font-weight: 600;
  margin-bottom: 8px;
}

.req {
  color: var(--danger);
}

.input-control {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  background: var(--bg-app);
  color: var(--text-main);
  outline: none;
  box-sizing: border-box;
}

.dropzone {
  border: 2px dashed var(--border);
  border-radius: var(--radius-md);
  padding: 32px;
  text-align: center;
  cursor: pointer;
  background: var(--bg-app);
  transition: 0.2s;
}

.dropzone:hover {
  border-color: var(--primary);
  background: var(--primary-light);
}

.upload-progress-box {
  padding: 20px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  background: var(--bg-app);
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
  border-radius: var(--radius-md);
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
