<template>
  <div class="tab-fade-in">
    <div class="form-grid">
      <!-- ЗАГОЛОВОК + КНОПКИ -->
      <div class="form-toolbar">
        <div class="title-block">
          <h1 style="margin: 0 0 8px 0; font-size: 1.5rem;">Настройка черновика</h1>
          <span class="status-badge bg-yellow" v-if="!moderationStatus">Заполнение данных</span>
          <span class="status-badge bg-yellow" v-else-if="moderationStatus === 'pending'">На модерации</span>
          <span class="status-badge bg-green" v-else-if="moderationStatus === 'approved'">Одобрено</span>
          <span class="status-badge bg-red" v-else-if="moderationStatus === 'rejected'">Отклонено</span>
        </div>
        <div class="actions">
          <button class="btn-dev-link" @click="showToast('Открытие Dev-среды...', 'info')">Перейти к игре (Dev)</button>
          <button class="btn-outline" @click="saveMeta">Сохранить</button>
          <button class="btn-primary" @click="submitForModeration" :disabled="submitting || moderationStatus === 'pending' || moderationStatus === 'approved'">
            {{ submitting ? 'Отправка...' : 'На модерацию' }}
          </button>
          <button class="btn-dev" @click="simulatePublish">[DEV] Симулировать публикацию</button>
        </div>
      </div>

      <div v-if="moderationStatus === 'rejected' && rejectionReason" class="card rejection-notice">
        <strong>Причина отказа:</strong> {{ rejectionReason }}
      </div>
      <div v-else-if="moderationStatus === 'approved'" class="card approval-notice">
        Игра одобрена модератором. Можно переходить к публикации.
      </div>

      <!-- БЛОК 1: МЕТАДАННЫЕ -->
      <div class="card form-section">
        <div class="section-head"><h3>Основная информация</h3></div>
        <div class="input-row">
          <div class="input-group">
            <label>Название игры на русском <span class="req">*</span></label>
            <input type="text" class="input-control" v-model="meta.title_ru" />
          </div>
          <div class="input-group">
            <label>Название игры на английском <span class="req">*</span></label>
            <input type="text" class="input-control" v-model="meta.title_en" />
          </div>
        </div>
        <div class="input-row">
          <div class="input-group">
            <label>SEO Описание (RU) <span class="req">*</span> <span class="char-count">{{ meta.seo_ru.length }}/180</span></label>
            <textarea class="input-control" rows="2" maxlength="180" v-model="meta.seo_ru"></textarea>
          </div>
          <div class="input-group">
            <label>SEO Описание (EN) <span class="req">*</span> <span class="char-count">{{ meta.seo_en.length }}/180</span></label>
            <textarea class="input-control" rows="2" maxlength="180" v-model="meta.seo_en"></textarea>
          </div>
        </div>
        <div class="input-group" style="margin-top: 16px;">
          <label>Описание "Об Игре" <span class="req">*</span> <span class="char-count">{{ meta.about.length }}/800</span></label>
          <textarea class="input-control" rows="4" maxlength="800" v-model="meta.about"></textarea>
        </div>
      </div>

      <!-- СКРЫТЫЕ ИНПУТЫ ДЛЯ МЕДИА И БИЛДОВ -->
      <input type="file" ref="fileIcon" accept="image/png" hidden @change="handleFile('icon', $event)" />
      <input type="file" ref="fileCoverMain" accept="image/png" hidden @change="handleFile('cover', $event)" />
      <input type="file" ref="fileVideo" accept="video/mp4" hidden @change="handleFile('video', $event)" />
      <input type="file" ref="fileZip" accept=".zip,.tar.gz" hidden @change="handleZipUpload" />

      <!-- БЛОК 2: ПРОМО -->
      <div class="card form-section">
        <div class="section-head"><h3>Промо-материалы</h3></div>
        <div class="media-list">
          <!-- Иконка -->
          <div class="media-item" :class="{ uploaded: media.icon }" @click="$refs.fileIcon.click()">
            <CheckCircle v-if="media.icon" class="icon-md text-green" />
            <ImageIcon v-else class="icon-md" />
            <span class="m-title">{{ media.icon ? 'Загружено' : 'Иконка' }}</span>
            <span class="m-req">512 x 512, png</span>
          </div>
          <!-- Обложка -->
          <div class="media-item" :class="{ uploaded: media.cover }" @click="$refs.fileCoverMain.click()">
            <CheckCircle v-if="media.cover" class="icon-md text-green" />
            <ImageIcon v-else class="icon-md" />
            <span class="m-title">{{ media.cover ? 'Загружено' : 'Обложка' }}</span>
            <span class="m-req">800 x 470, png</span>
          </div>
          <!-- Видео -->
          <div class="media-item" :class="{ uploaded: media.video }" @click="$refs.fileVideo.click()">
            <CheckCircle v-if="media.video" class="icon-md text-green" />
            <Film v-else class="icon-md" />
            <span class="m-title">{{ media.video ? 'Загружено' : 'Видео' }}</span>
            <span class="m-req">До 12 МБ</span>
          </div>
        </div>
      </div>

      <!-- БЛОК 3: БИЛД -->
      <div class="card form-section">
        <div class="section-head"><h3>Билд</h3></div>

        <div class="input-group" style="margin-bottom: 16px;">
          <label>Версия билда <span class="req">*</span></label>
          <input type="text" v-model="newBuildVersion" class="input-control" placeholder="1.0.0" :disabled="buildStatus !== 'idle'" />
        </div>

        <!-- Ожидание загрузки -->
        <div v-if="buildStatus === 'idle'" class="dropzone" @click="$refs.fileZip.click()">
          <UploadCloud style="width:32px; height:32px; color: var(--text-muted); margin-bottom:8px;" />
          <span style="display:block; font-weight:600;">Нажмите для загрузки .zip или .tar.gz архива</span>
        </div>

        <!-- Идет загрузка -->
        <div v-if="buildStatus === 'uploading'" class="upload-progress-box">
          <div class="prog-info">
            <span style="font-weight: 600; color: var(--text-main);">Загрузка и распаковка архива...</span>
            <span style="font-weight: 600; color: var(--primary);">{{ buildProgress }}%</span>
          </div>
          <div class="prog-bg">
            <div class="prog-fill" :style="{ width: buildProgress + '%' }"></div>
          </div>
        </div>

        <!-- Загрузка завершена -->
        <div v-if="buildStatus === 'done'" class="upload-success-box">
          <CheckCircle class="icon-md text-green" />
          <div>
            <span style="display:block; font-weight:600;">Билд v{{ uploadedVersion }} успешно загружен!</span>
            <button class="btn-text mt-8" @click="resetBuildUpload">Загрузить новую версию</button>
          </div>
        </div>

        <!-- Список версий -->
        <div v-if="recentBuilds.length" class="build-versions">
          <h4 class="versions-title">Версии (последние 5)</h4>
          <div
            v-for="b in recentBuilds"
            :key="b.version"
            class="build-row"
            :class="{ active: activeBuildVersion === b.version }"
          >
            <div class="build-info">
              <strong>{{ b.version }}</strong>
              <span class="build-date">{{ b.created_at }}</span>
            </div>
            <button
              v-if="activeBuildVersion !== b.version"
              class="btn-text"
              @click="setActiveBuild(b.version)"
            >Сделать активной</button>
            <span v-else class="active-label">Активная</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch, inject } from 'vue'
import { useRoute } from 'vue-router'
import { UploadCloud, CheckCircle, Image as ImageIcon, Film } from 'lucide-vue-next'
import { showToast } from '../../store'
import { getProject, updateProject, uploadBuild, listBuilds, uploadMedia } from '../../api/projects'
import { moderationApi, moderationToTicket } from '../../api/moderation'
import JSZip from 'jszip'
import pako from 'pako'

const route = useRoute()
const projectId = computed(() => route.params.id)
const project = inject('project')

const meta = ref({ title_ru: '', title_en: '', seo_ru: '', seo_en: '', about: '' })
const media = ref({ icon: false, cover: false, video: false })
const activeBuildVersion = ref('')

const submitting = ref(false)
const moderationStatus = ref(null)
const rejectionReason = ref('')

const newBuildVersion = ref('')
const buildStatus = ref('idle')
const buildProgress = ref(0)
const uploadedVersion = ref('')
const recentBuilds = ref([])

let autoSaveTimeout = null
let skipAutoSave = false

async function loadModerationStatus() {
  const gameId = parseInt(projectId.value, 10)
  if (!gameId) return
  try {
    const data = await moderationApi.getStatus(gameId)
    if (!data) return
    const ticket = moderationToTicket(data.moderation)
    moderationStatus.value = ticket.status
    rejectionReason.value = ticket.rejectionReason
  } catch {
    // moderation service unavailable
  }
}

async function loadProject() {
  skipAutoSave = true
  try {
    const project = await getProject(projectId.value)
    meta.value = {
      title_ru: project.title_ru || '',
      title_en: project.title_en || '',
      seo_ru: project.seo_ru || '',
      seo_en: project.seo_en || '',
      about: project.about || ''
    }
    media.value.icon = !!project.icon_path
    media.value.cover = !!project.cover_path
    media.value.video = !!project.video_path
    activeBuildVersion.value = project.active_build_version || ''

    const builds = await listBuilds(projectId.value)
    recentBuilds.value = builds
    if (!activeBuildVersion.value && builds.length > 0) {
      activeBuildVersion.value = builds[0].version
    }
    await loadModerationStatus()
  } catch (err) {
    showToast('Не удалось загрузить данные проекта', 'danger')
  }
  setTimeout(() => { skipAutoSave = false }, 1600)
}

onMounted(loadProject)

async function submitForModeration() {
  const gameId = parseInt(projectId.value, 10)
  if (!gameId) {
    showToast('Не удалось определить ID игры', 'danger')
    return
  }
  if (!meta.value.title_ru.trim()) {
    showToast('Укажите название игры', 'danger')
    return
  }
  submitting.value = true
  try {
    await saveMeta(true)
    await moderationApi.submitForReview(
      gameId,
      meta.value.title_ru.trim(),
      meta.value.about.trim() || meta.value.title_ru.trim()
    )
    moderationStatus.value = 'pending'
    rejectionReason.value = ''
    showToast('Отправлено модератору!', 'success')
  } catch (e) {
    showToast(e.message || 'Ошибка отправки на модерацию', 'danger')
  } finally {
    submitting.value = false
  }
}

async function saveMeta(silent = false) {
  try {
    const payload = {
      ...meta.value,
      active_build_version: activeBuildVersion.value
    }
    await updateProject(projectId.value, payload)
    if (!silent) showToast('Сохранено!', 'success')
  } catch (err) {
    if (!silent) showToast('Ошибка сохранения', 'danger')
  }
}

// Автосохранение метаданных при изменении (debounce 1.5с)
watch(
  () => ({ ...meta.value, active_build_version: activeBuildVersion.value }),
  () => {
    if (skipAutoSave) return
    if (autoSaveTimeout) clearTimeout(autoSaveTimeout)
    autoSaveTimeout = setTimeout(() => saveMeta(true), 1500)
  },
  { deep: true }
)

function validateImageDimensions(file, expectedWidth, expectedHeight) {
  return new Promise((resolve, reject) => {
    const img = new Image()
    img.onload = () => {
      URL.revokeObjectURL(img.src)
      if (img.width === expectedWidth && img.height === expectedHeight) {
        resolve(true)
      } else {
        reject(new Error(`Разрешение должно быть ${expectedWidth}x${expectedHeight}px (загружено ${img.width}x${img.height})`))
      }
    }
    img.onerror = () => {
      URL.revokeObjectURL(img.src)
      reject(new Error('Не удалось загрузить изображение'))
    }
    img.src = URL.createObjectURL(file)
  })
}

const handleFile = async (type, event) => {
  const file = event.target.files[0]
  if (!file) return
  try {
    if (type === 'icon') {
      await validateImageDimensions(file, 512, 512)
    } else if (type === 'cover') {
      await validateImageDimensions(file, 800, 470)
    }
    await uploadMedia(projectId.value, type, file)
    media.value[type] = true
    showToast('Медиафайл успешно сохранен', 'success')
  } catch (err) {
    showToast(err.message || 'Ошибка загрузки', 'danger')
  } finally {
    event.target.value = ''
  }
}

async function checkZipForIndexHtml(arrayBuffer) {
  const zip = await JSZip.loadAsync(arrayBuffer)
  const entry = zip.file('index.html')
  if (!entry) throw new Error('В корне архива отсутствует index.html')
}

function checkTarGzForIndexHtml(arrayBuffer) {
  const inflated = pako.inflate(new Uint8Array(arrayBuffer))
  let offset = 0
  while (offset < inflated.length) {
    let name = ''
    for (let i = 0; i < 100; i++) {
      if (inflated[offset + i] === 0) break
      name += String.fromCharCode(inflated[offset + i])
    }
    if (name === 'index.html') return true
    if (name.length === 0) break
    let sizeStr = ''
    for (let i = 124; i < 136; i++) {
      sizeStr += String.fromCharCode(inflated[offset + i])
    }
    const size = parseInt(sizeStr.trim(), 8) || 0
    offset += 512 + Math.ceil(size / 512) * 512
  }
  throw new Error('В корне архива отсутствует index.html')
}

const handleZipUpload = async (event) => {
  const file = event.target.files[0]
  if (!file) return

  const version = newBuildVersion.value.trim()
  if (!version) {
    showToast('Укажите версию билда перед загрузкой', 'danger')
    event.target.value = ''
    return
  }

  const name = file.name.toLowerCase()
  const isZip = name.endsWith('.zip')
  const isTarGz = name.endsWith('.tar.gz')

  if (!isZip && !isTarGz) {
    showToast('Допустимые форматы: .zip и .tar.gz', 'danger')
    event.target.value = ''
    return
  }

  try {
    const buffer = await file.arrayBuffer()
    if (isZip) {
      await checkZipForIndexHtml(buffer)
    } else {
      checkTarGzForIndexHtml(buffer)
    }
  } catch (err) {
    showToast(err.message, 'danger')
    event.target.value = ''
    return
  }

  buildStatus.value = 'uploading'
  buildProgress.value = 0

  try {
    await uploadBuild(projectId.value, version, file, (p) => {
      buildProgress.value = p
    })
    buildStatus.value = 'done'
    uploadedVersion.value = version
    newBuildVersion.value = ''
    activeBuildVersion.value = version
    await loadProject()
    showToast('Билд развернут в Dev-среде!', 'success')
  } catch (err) {
    buildStatus.value = 'idle'
    showToast('Ошибка загрузки билда', 'danger')
  }
}

function resetBuildUpload() {
  buildStatus.value = 'idle'
  buildProgress.value = 0
  uploadedVersion.value = ''
}

function setActiveBuild(version) {
  activeBuildVersion.value = version
  showToast(`Активная версия изменена на ${version}`, 'success')
}

function simulatePublish() {
  if (project && project.value) {
    project.value.status = 3
    showToast('Статус изменён на "Опубликована" (dev)', 'info')
  }
}
</script>

<style scoped>
.tab-fade-in { animation: fadeIn 0.3s ease; }
@keyframes fadeIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }

.form-grid { display: flex; flex-direction: column; gap: 24px; max-width: 800px; padding-bottom: 60px; }
.form-toolbar { display: flex; justify-content: space-between; align-items: flex-end; }
.actions { display: flex; gap: 12px; }
.btn-dev-link { display: flex; align-items: center; gap: 6px; padding: 8px 16px; border: 1px solid var(--text-muted); border-radius: var(--radius-md); background: transparent; color: var(--text-muted); font-weight: 600; font-size: 0.85rem; cursor: pointer; transition: 0.2s; }
.btn-dev-link:hover { border-color: var(--primary); color: var(--primary); background: var(--bg-hover); }
.btn-dev { display: flex; align-items: center; gap: 6px; padding: 8px 16px; border: 1px dashed var(--danger); border-radius: var(--radius-md); background: var(--danger-light); color: var(--danger); font-weight: 600; font-size: 0.85rem; cursor: pointer; transition: 0.2s; }
.btn-dev:hover { background: var(--danger); color: white; }
.status-badge { padding: 4px 10px; border-radius: 12px; font-size: 0.8rem; font-weight: 600; display: inline-block;}
.bg-yellow { background: var(--warning-light); color: var(--warning); }
.bg-green { background: var(--success-light); color: var(--success); }
.bg-red { background: #FEE2E2; color: #DC2626; }
.rejection-notice { padding: 16px; background: #FEF2F2; border: 1px solid #FECACA; color: #B91C1C; font-size: 0.9rem; line-height: 1.5; }
.approval-notice { padding: 16px; background: var(--success-light); border: 1px solid var(--success); color: var(--success); font-size: 0.9rem; }
.section-head { margin-bottom: 20px; border-bottom: 1px solid var(--border); padding-bottom: 12px; }
.section-head h3 { margin: 0; font-size: 1.1rem; }

.input-row { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; margin-bottom: 16px; }
.input-group label { display: block; font-size: 0.85rem; font-weight: 600; margin-bottom: 8px; display: flex; justify-content: space-between;}
.req { color: var(--danger); }
.char-count { font-weight: 400; color: var(--text-muted); }
.input-control { width: 100%; padding: 10px 12px; border: 1px solid var(--border); border-radius: var(--radius-md); background: var(--bg-input); font-family: inherit; box-sizing: border-box; resize: vertical; color: var(--text-main); }
.input-control:focus { outline: none; border-color: var(--primary); background: var(--bg-card); }

.media-list { display: flex; flex-direction: column; gap: 12px; align-items: center; }
.media-item { border: 1px dashed var(--border); border-radius: var(--radius-md); background: var(--bg-secondary); display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 6px; color: var(--text-muted); text-align: center; cursor: pointer; padding: 14px; transition: 0.2s; width: 510px; height: 110px; }
.media-item:hover { border-color: var(--primary); background: var(--bg-hover); color: var(--primary); }
.media-item.uploaded { border: 1px solid var(--success); background: var(--success-light); color: var(--success); }
.media-item .icon-md { width: 16px; height: 16px; }
.text-green { color: var(--success); }
.m-title { font-size: 0.8rem; font-weight: 600; color: var(--text-main); }
.media-item.uploaded .m-title { color: var(--success); }
.m-req { font-size: 0.65rem; }

.dropzone { border: 2px dashed var(--border); border-radius: var(--radius-md); padding: 32px; text-align: center; cursor: pointer; background: var(--bg-secondary); transition: 0.2s;}
.dropzone:hover { border-color: var(--primary); background: var(--bg-hover);}
.upload-progress-box, .upload-success-box { padding: 24px; border: 1px solid var(--border); border-radius: var(--radius-md); background: var(--bg-secondary); }
.prog-info { display: flex; justify-content: space-between; margin-bottom: 12px; font-size: 0.95rem; }
.prog-bg { background: var(--border); height: 8px; border-radius: 4px; overflow: hidden; }
.prog-fill { background: var(--primary); height: 100%; transition: width 0.3s ease; }
.upload-success-box { display: flex; align-items: center; gap: 16px; border-color: var(--success); background: var(--success-light); }
.btn-text { background: none; border: none; color: var(--primary); font-weight: 600; cursor: pointer; text-decoration: underline; padding: 0; }
.mt-8 { margin-top: 8px; }

.build-versions { margin-top: 16px; }
.versions-title { margin: 16px 0 8px; font-size: 0.9rem; color: var(--text-main); }
.build-row { display: flex; justify-content: space-between; align-items: center; padding: 10px 12px; border: 1px solid var(--border); border-radius: var(--radius-md); background: var(--bg-secondary); margin-bottom: 8px; transition: 0.2s; }
.build-row.active { border-color: var(--success); background: var(--success-light); }
.build-info { display: flex; align-items: center; gap: 8px; }
.build-date { font-size: 0.75rem; color: var(--text-muted); }
.active-label { color: var(--success); font-size: 0.8rem; font-weight: 600; }
</style>
