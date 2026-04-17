<template>
  <div class="builds-page tab-fade-in">
    <div class="page-header">
      <h1>Серверные билды</h1>
      <button class="btn-primary" @click="showUploadForm = !showUploadForm">
        <Upload class="icon-sm" /> Загрузить билд
      </button>
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
        <button class="btn-primary" @click="submitBuild" :disabled="!uploadForm.build_version || !uploadForm.file">
          Загрузить
        </button>
        <button class="btn-outline" @click="showUploadForm = false">Отмена</button>
      </div>

      <!-- Прогресс -->
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
    <div class="table-wrap" v-if="builds.length">
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
              <button class="btn-icon" @click="confirmDelete(b)" title="Удалить">
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
          <button class="btn-primary" @click="doDelete" :disabled="deleteTarget._inUse">Удалить</button>
          <button class="btn-outline" @click="deleteTarget = null">Отмена</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { Upload, Trash2 } from 'lucide-vue-next'
import { mockBuilds } from '../../data/mock-orchestrator'
import { showToast } from '../../store'

defineProps({ gameId: { type: [String, Number], required: true } })

const builds = ref([...mockBuilds])
const showUploadForm = ref(false)
const uploading = ref(false)
const uploadProgress = ref(0)
const fileInput = ref(null)
const deleteTarget = ref(null)

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

function submitBuild() {
  uploading.value = true
  uploadProgress.value = 0
  const interval = setInterval(() => {
    uploadProgress.value += Math.random() * 15
    if (uploadProgress.value >= 100) {
      uploadProgress.value = 100
      clearInterval(interval)
      setTimeout(() => {
        uploading.value = false
        showUploadForm.value = false
        uploadForm.value = { file: null, build_version: '', protocol: 'websocket', internal_port: 8080, max_players: 16 }
        showToast('Билд успешно загружен')
      }, 400)
    }
  }, 300)
}

function confirmDelete(b) {
  // Моковая проверка: если инстансы используют билд
  const inUse = Math.random() > 0.7 // иногда показываем ошибку
  deleteTarget.value = { ...b, _inUse: inUse }
}

function doDelete() {
  builds.value = builds.value.filter(b => b.id !== deleteTarget.value.id)
  showToast(`Билд ${deleteTarget.value.build_version} удалён`)
  deleteTarget.value = null
}

function formatBytes(b) {
  if (b < 1024 * 1024) return (b / 1024).toFixed(0) + ' KB'
  if (b < 1024 * 1024 * 1024) return (b / (1024 * 1024)).toFixed(1) + ' MB'
  return (b / (1024 * 1024 * 1024)).toFixed(1) + ' GB'
}

function formatDate(ts) {
  return new Date(ts).toLocaleDateString('ru-RU', { day: 'numeric', month: 'short', year: 'numeric' })
}
</script>

<style scoped>
.tab-fade-in { animation: fadeIn 0.3s ease; }
@keyframes fadeIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }

.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.page-header h1 { margin: 0; }

.upload-form { margin-bottom: 24px; }
.upload-form h3 { margin: 0 0 16px; }
.form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
.form-group { display: flex; flex-direction: column; gap: 4px; }
.form-group-wide { grid-column: 1 / -1; }
.form-group label { font-size: 0.82rem; font-weight: 600; color: var(--text-muted); }
.form-input {
  padding: 8px 12px;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  background: var(--bg-input);
  color: var(--text-main);
  font-size: 0.88rem;
}
.file-drop {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  border: 2px dashed var(--border);
  border-radius: var(--radius-md);
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
.empty-state { padding: 40px; text-align: center; color: var(--text-muted); }

.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.4); z-index: 100; display: flex; align-items: center; justify-content: center; }
.modal { max-width: 420px; width: 90%; }
.modal h3 { margin: 0 0 8px; }
.modal p { margin: 8px 0; font-size: 0.9rem; color: var(--text-muted); }
.modal-actions { display: flex; gap: 8px; margin-top: 16px; }
.text-danger { color: var(--danger); font-weight: 600; }
</style>
