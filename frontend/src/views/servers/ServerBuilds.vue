<template>
  <div class="builds-page tab-fade-in">
    <div class="page-header">
      <h1>Серверные билды</h1>
      <button class="btn-primary" @click="showUploadForm = !showUploadForm">
        <Upload class="icon-sm" /> Загрузить билд
      </button>
    </div>

    <!-- Ошибка -->
    <div v-if="error" class="error-banner">
      <AlertCircle class="icon-sm" /> {{ error }}
      <button class="btn-outline btn-sm" @click="fetchBuilds">Повторить</button>
    </div>

    <!-- Форма загрузки -->
    <div v-if="showUploadForm" class="upload-form card">
      <h3>Новый серверный билд</h3>
      <div class="form-grid">
        <div class="form-group form-group-wide">
          <label>Файл образа (TAR, до 2 ГБ)</label>
          <div class="file-drop" @dragover.prevent @drop.prevent="onDrop">
            <input type="file" ref="fileInput" accept=".tar,.tar.gz" @change="onFileSelect" hidden />
            <button class="btn-outline" @click="$refs.fileInput.click()">Выбрать файл</button>
            <span class="file-name">{{ uploadForm.file?.name ?? 'или перетащите сюда' }}</span>
          </div>
        </div>
        <div class="form-group">
          <label>Версия билда *</label>
          <input type="text" v-model="uploadForm.build_version" placeholder="1.0.0" class="form-input" />
        </div>
        <div class="form-group">
          <label>Протокол</label>
          <select v-model="uploadForm.protocol" class="form-input">
            <option value="tcp">TCP</option>
            <option value="udp">UDP</option>
            <option value="websocket">WebSocket</option>
            <option value="webrtc">WebRTC</option>
          </select>
        </div>
        <div class="form-group">
          <label>Внутренний порт</label>
          <input type="number" v-model.number="uploadForm.internal_port" class="form-input" min="1" max="65535" />
        </div>
        <div class="form-group">
          <label>Макс. игроков</label>
          <input type="number" v-model.number="uploadForm.max_players" class="form-input" min="1" />
        </div>
      </div>
      <div class="form-actions">
        <button class="btn-primary" @click="submitBuild" :disabled="!uploadForm.build_version || !uploadForm.file || uploading">
          Загрузить
        </button>
        <button class="btn-outline" @click="showUploadForm = false">Отмена</button>
      </div>
      <div v-if="uploading" class="upload-progress">
        <div class="progress-info">
          <span>Загрузка билда...</span>
          <span>{{ uploadProgress }}%</span>
        </div>
        <div class="progress-bar-bg">
          <div class="progress-bar-fill" :style="{ width: uploadProgress + '%' }"></div>
        </div>
      </div>
    </div>

    <!-- Таблица билдов -->
    <div v-if="loading" class="loading-state">Загрузка...</div>
    <div v-else-if="error" class="empty-state"></div>
    <div class="table-wrap" v-else-if="builds.length">
      <table class="data-table">
        <thead>
          <tr>
            <th>Версия</th>
            <th>Образ</th>
            <th>Протокол</th>
            <th>Порт</th>
            <th>Макс. игроков</th>
            <th>Размер</th>
            <th>Дата</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="b in builds" :key="b.id">
            <td><code>{{ b.build_version }}</code></td>
            <td class="cell-muted">{{ b.image_tag }}</td>
            <td>{{ b.protocol }}</td>
            <td>{{ b.internal_port }}</td>
            <td>{{ b.max_players }}</td>
            <td>{{ formatBytes(b.file_size_bytes) }}</td>
            <td class="cell-muted">{{ formatDate(b.created_at) }}</td>
            <td>
              <button class="btn-icon" @click="confirmDelete(b)" title="Удалить" :disabled="deleting">
                <Trash2 class="icon-sm" />
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <div v-else class="empty-state">Нет загруженных билдов</div>

    <!-- Диалог подтверждения удаления -->
    <div v-if="deleteTarget" class="modal-overlay" @click.self="deleteTarget = null">
      <div class="modal card">
        <h3>Удалить билд?</h3>
        <p>Билд <code>{{ deleteTarget.build_version }}</code> будет удалён из хранилища и со всех нод.</p>
        <p v-if="deleteTarget._inUse" class="text-danger">Этот билд используется работающими инстансами и не может быть удалён.</p>
        <div class="modal-actions">
          <button class="btn-primary" @click="doDelete" :disabled="deleteTarget._inUse || deleting">Удалить</button>
          <button class="btn-outline" @click="deleteTarget = null">Отмена</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Upload, Trash2, AlertCircle } from 'lucide-vue-next'
import { listBuilds, uploadBuild, deleteBuild, listInstances } from '../../api/orchestrator'
import { showToast } from '../../store'

const props = defineProps({ gameId: { type: [String, Number], required: true } })

const builds = ref([])
const loading = ref(true)
const error = ref(null)
const showUploadForm = ref(false)
const uploading = ref(false)
const uploadProgress = ref(0)
const fileInput = ref(null)
const deleteTarget = ref(null)
const deleting = ref(false)

const uploadForm = ref({
  file: null,
  build_version: '',
  protocol: 'websocket',
  internal_port: 8080,
  max_players: 16,
})

function onFileSelect(e) {
  uploadForm.value.file = e.target.files[0] || null
}
function onDrop(e) {
  const file = e.dataTransfer.files[0]
  if (file) uploadForm.value.file = file
}

async function fetchBuilds() {
  loading.value = true
  error.value = null
  try {
    builds.value = await listBuilds(props.gameId)
  } catch (e) {
    error.value = e.response?.data?.message ?? e.message
  } finally {
    loading.value = false
  }
}

async function submitBuild() {
  uploading.value = true
  uploadProgress.value = 0
  const form = uploadForm.value
  const fd = new FormData()
  fd.append('image', form.file)
  fd.append('build_version', form.build_version)
  fd.append('protocol', form.protocol)
  fd.append('internal_port', String(form.internal_port))
  fd.append('max_players', String(form.max_players))

  try {
    await uploadBuild(props.gameId, fd, (e) => {
      if (e.total) uploadProgress.value = Math.round((e.loaded / e.total) * 100)
    })
    showToast('Билд успешно загружен')
    showUploadForm.value = false
    uploadForm.value = { file: null, build_version: '', protocol: 'websocket', internal_port: 8080, max_players: 16 }
    await fetchBuilds()
  } catch (e) {
    showToast(e.response?.data?.message ?? 'Ошибка загрузки билда', 'error')
  } finally {
    uploading.value = false
  }
}

async function confirmDelete(b) {
  // Проверяем, используется ли билд работающими инстансами
  try {
    const instances = await listInstances(props.gameId)
    const inUse = instances.some(i => i.build_version === b.build_version && i.status === 'running')
    deleteTarget.value = { ...b, _inUse: inUse }
  } catch {
    deleteTarget.value = { ...b, _inUse: false }
  }
}

async function doDelete() {
  deleting.value = true
  try {
    await deleteBuild(props.gameId, deleteTarget.value.build_version)
    showToast(`Билд ${deleteTarget.value.build_version} удалён`)
    deleteTarget.value = null
    await fetchBuilds()
  } catch (e) {
    if (e.response?.status === 409) {
      showToast('Билд используется работающими инстансами', 'error')
      deleteTarget.value._inUse = true
    } else {
      showToast(e.response?.data?.message ?? 'Ошибка удаления', 'error')
    }
  } finally {
    deleting.value = false
  }
}

function formatBytes(b) {
  if (!b) return '—'
  if (b < 1024 * 1024) return (b / 1024).toFixed(0) + ' KB'
  if (b < 1024 * 1024 * 1024) return (b / (1024 * 1024)).toFixed(1) + ' MB'
  return (b / (1024 * 1024 * 1024)).toFixed(1) + ' GB'
}

function formatDate(ts) {
  return new Date(ts).toLocaleDateString('ru-RU', { day: 'numeric', month: 'short', year: 'numeric' })
}

onMounted(fetchBuilds)
</script>

<style scoped>
.tab-fade-in { animation: fadeIn 0.3s ease; }
@keyframes fadeIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }

.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.page-header h1 { margin: 0; }

.error-banner {
  display: flex; align-items: center; gap: 8px;
  padding: 12px 16px; background: var(--danger-light); color: var(--danger);
  border-radius: var(--radius-md); margin-bottom: 16px; font-size: 0.88rem;
}
.btn-sm { padding: 4px 12px; font-size: 0.82rem; }
.loading-state { padding: 40px; text-align: center; color: var(--text-muted); }

.upload-form { margin-bottom: 24px; }
.upload-form h3 { margin: 0 0 16px; }
.form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
.form-group { display: flex; flex-direction: column; gap: 4px; }
.form-group-wide { grid-column: 1 / -1; }
.form-group label { font-size: 0.82rem; font-weight: 600; color: var(--text-muted); }
.form-input {
  padding: 8px 12px; border: 1px solid var(--border); border-radius: var(--radius-sm);
  background: var(--bg-input); color: var(--text-main); font-size: 0.88rem;
}
.file-drop {
  display: flex; align-items: center; gap: 12px;
  padding: 16px; border: 2px dashed var(--border); border-radius: var(--radius-md);
  background: var(--bg-secondary);
}
.file-name { color: var(--text-muted); font-size: 0.85rem; }
.form-actions { display: flex; gap: 8px; margin-top: 16px; }

.upload-progress { margin-top: 16px; }
.progress-info { display: flex; justify-content: space-between; font-weight: 600; margin-bottom: 6px; }
.progress-bar-bg { width: 100%; height: 8px; background: var(--border); border-radius: 4px; overflow: hidden; }
.progress-bar-fill { height: 100%; background: var(--success); transition: width 0.3s; }

.table-wrap { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); overflow: hidden; }
.data-table { width: 100%; border-collapse: collapse; }
.data-table th {
  text-align: left; padding: 12px 16px; font-size: 0.78rem; font-weight: 600;
  color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.03em;
  background: var(--bg-secondary); border-bottom: 1px solid var(--border);
}
.data-table td { padding: 12px 16px; font-size: 0.88rem; border-bottom: 1px solid var(--border); }
.data-table tr:last-child td { border-bottom: none; }
.cell-muted { color: var(--text-muted); }
code { background: var(--bg-secondary); padding: 2px 6px; border-radius: 4px; font-size: 0.82rem; }
.btn-icon {
  background: none; border: none; color: var(--text-muted); cursor: pointer;
  padding: 4px; border-radius: 4px; display: flex; align-items: center;
}
.btn-icon:hover { color: var(--danger); background: var(--danger-light); }
.btn-icon:disabled { opacity: 0.4; cursor: not-allowed; }
.empty-state { padding: 40px; text-align: center; color: var(--text-muted); }

.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.4); z-index: 100; display: flex; align-items: center; justify-content: center; }
.modal { max-width: 420px; width: 90%; }
.modal h3 { margin: 0 0 8px; }
.modal p { margin: 8px 0; font-size: 0.9rem; color: var(--text-muted); }
.modal-actions { display: flex; gap: 8px; margin-top: 16px; }
.text-danger { color: var(--danger); font-weight: 600; }
</style>
