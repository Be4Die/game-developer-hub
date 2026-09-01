<template>
  <div class="upload-form card">
    <h3>{{ t('servers.uploadBuild') }}</h3>
    <div class="form-grid">
      <div class="form-group form-group-wide">
        <label>Image file (TAR, max 2GB)</label>
        <div class="file-drop" @dragover.prevent @drop.prevent="onDrop">
          <input
            type="file"
            ref="fileInput"
            accept=".tar,.tar.gz"
            @change="onFileSelect"
            hidden
          />
          <button class="btn-outline" @click="$refs.fileInput.click()">
            {{ t('common.upload') }}
          </button>
          <span class="file-name">{{
            uploadForm.file?.name ?? 'or drag and drop here'
          }}</span>
        </div>
      </div>
      <div class="form-group">
        <label>{{ t('common.version') }} *</label>
        <input
          type="text"
          v-model="uploadForm.build_version"
          placeholder="1.0.0"
          class="form-input"
        />
      </div>
      <div class="form-group">
        <label>Protocol</label>
        <select v-model="uploadForm.protocol" class="form-input">
          <option value="tcp">TCP</option>
          <option value="udp">UDP</option>
          <option value="websocket">WebSocket</option>
          <option value="webrtc">WebRTC</option>
        </select>
      </div>
      <div class="form-group">
        <label>Internal Port</label>
        <input
          type="number"
          v-model.number="uploadForm.internal_port"
          class="form-input"
          min="1"
          max="65535"
        />
      </div>
      <div class="form-group">
        <label>Max Players</label>
        <input
          type="number"
          v-model.number="uploadForm.max_players"
          class="form-input"
          min="1"
        />
      </div>
    </div>
    <div class="form-actions">
      <button
        class="btn-primary"
        @click="submitBuild"
        :disabled="
          !uploadForm.build_version || !uploadForm.file || uploading
        "
      >
        {{ t('common.upload') }}
      </button>
      <button class="btn-outline" @click="$emit('cancel')">{{ t('common.cancel') }}</button>
    </div>
    <div v-if="uploading" class="upload-progress">
      <div class="progress-info">
        <span>{{ t('common.loading') }}</span>
        <span>{{ uploadProgress }}%</span>
      </div>
      <div class="progress-bar-bg">
        <div
          class="progress-bar-fill"
          :style="{ width: uploadProgress + '%' }"
        ></div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { uploadServerBuild } from '@/entities/build';
import { showToast } from '@/shared/lib';

const { t } = useI18n();

const props = defineProps({

  gameId: { type: [String, Number], required: true },
});

const emit = defineEmits(['uploaded', 'cancel']);

const uploading = ref(false);
const uploadProgress = ref(0);
const fileInput = ref(null);

const uploadForm = ref({
  file: null,
  build_version: '',
  protocol: 'websocket',
  internal_port: 8080,
  max_players: 16,
});

function onFileSelect(e) {
  uploadForm.value.file = e.target.files[0] || null;
}

function onDrop(e) {
  const file = e.dataTransfer.files[0];
  if (file) uploadForm.value.file = file;
}

async function submitBuild() {
  uploading.value = true;
  uploadProgress.value = 0;
  const form = uploadForm.value;
  const fd = new FormData();
  fd.append('image', form.file);
  fd.append('build_version', form.build_version);
  fd.append('protocol', form.protocol);
  fd.append('internal_port', String(form.internal_port));
  fd.append('max_players', String(form.max_players));

  try {
    await uploadServerBuild(props.gameId, fd, (e) => {
      if (e.total) uploadProgress.value = Math.round((e.loaded / e.total) * 100);
    });
    showToast('Билд успешно загружен', 'success');
    uploadForm.value = {
      file: null,
      build_version: '',
      protocol: 'websocket',
      internal_port: 8080,
      max_players: 16,
    };
    emit('uploaded');
  } catch (e) {
    showToast(e.response?.data?.message ?? 'Ошибка загрузки билда', 'error');
  } finally {
    uploading.value = false;
  }
}
</script>

<style scoped>
.upload-form {
  margin-bottom: 24px;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  padding: 24px;
}

.upload-form h3 {
  margin: 0 0 16px 0;
  font-size: 1.1rem;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
  margin-bottom: 20px;
}

.form-group-wide {
  grid-column: 1 / -1;
}

.form-group label {
  display: block;
  font-size: 0.82rem;
  font-weight: 600;
  color: var(--text-muted);
  margin-bottom: 6px;
}

.form-input {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  background: var(--bg-app);
  color: var(--text-main);
  font-size: 0.9rem;
  outline: none;
  box-sizing: border-box;
}

.file-drop {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border: 1px dashed var(--border);
  border-radius: var(--radius-md);
  background: var(--bg-app);
}

.file-name {
  font-size: 0.85rem;
  color: var(--text-muted);
}

.form-actions {
  display: flex;
  gap: 12px;
}

.upload-progress {
  margin-top: 16px;
}

.progress-info {
  display: flex;
  justify-content: space-between;
  font-size: 0.82rem;
  color: var(--text-muted);
  margin-bottom: 4px;
}

.progress-bar-bg {
  width: 100%;
  height: 6px;
  background: var(--border);
  border-radius: 3px;
  overflow: hidden;
}

.progress-bar-fill {
  height: 100%;
  background: var(--primary);
  transition: width 0.2s ease;
}
</style>
