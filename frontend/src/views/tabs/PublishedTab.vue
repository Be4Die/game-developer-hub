<template>
  <div class="tab-fade-in">
    <div class="form-grid">
      <!-- ЗАГОЛОВОК + КНОПКИ -->
      <div class="form-toolbar">
        <div class="title-block">
          <h1 style="margin: 0 0 8px 0; font-size: 1.5rem;">Опубликованная версия</h1>
          <span class="status-badge bg-green">Доступна игрокам</span>
        </div>
        <div class="actions">
          <button class="btn-prod-link" @click="showToast('Открытие Prod-среды...', 'info')">Перейти к игре (Prod)</button>
          <button class="btn-outline" @click="loadProject">Обновить</button>
          <button class="btn btn-danger" @click="unpublish">Снять с публикации</button>
        </div>
      </div>

      <!-- БЛОК 1: МЕТАДАННЫЕ -->
      <div class="card form-section">
        <div class="section-head">
          <h3>Основная информация</h3>
          <p class="version-info">Версия: {{ activeBuildDisplay }}</p>
        </div>

        <div class="input-row">
          <div class="input-group">
            <label>Название игры на русском</label>
            <div class="readonly-field">{{ project?.title_ru || '—' }}</div>
          </div>
          <div class="input-group">
            <label>Название игры на английском</label>
            <div class="readonly-field">{{ project?.title_en || '—' }}</div>
          </div>
        </div>

        <div class="input-row">
          <div class="input-group">
            <label>SEO Описание (RU)</label>
            <div class="readonly-field">{{ project?.seo_ru || '—' }}</div>
          </div>
          <div class="input-group">
            <label>SEO Описание (EN)</label>
            <div class="readonly-field">{{ project?.seo_en || '—' }}</div>
          </div>
        </div>

        <div class="input-group" style="margin-top: 16px;">
          <label>Описание "Об Игре"</label>
          <div class="readonly-field multiline">{{ project?.about || '—' }}</div>
        </div>
      </div>

      <!-- БЛОК 2: ПРОМО -->
      <div class="card form-section">
        <div class="section-head"><h3>Промо-материалы</h3></div>

        <div class="media-list">
          <div class="media-item" :class="{ uploaded: !!project?.icon_path, empty: !project?.icon_path }">
            <template v-if="project?.icon_path">
              <CheckCircle class="icon-md text-green" />
              <span class="m-title">Загружено</span>
              <span class="m-req">512 x 512, png</span>
              <span class="upload-label">Иконка</span>
            </template>
            <template v-else>
              <ImageIcon class="icon-md" />
              <span class="m-title">Не загружено</span>
              <span class="m-req">512 x 512, png</span>
              <span class="upload-label">Иконка</span>
            </template>
          </div>

          <div class="media-item" :class="{ uploaded: !!project?.cover_path, empty: !project?.cover_path }">
            <template v-if="project?.cover_path">
              <CheckCircle class="icon-md text-green" />
              <span class="m-title">Загружено</span>
              <span class="m-req">800 x 470, png</span>
              <span class="upload-label">Обложка</span>
            </template>
            <template v-else>
              <ImageIcon class="icon-md" />
              <span class="m-title">Не загружено</span>
              <span class="m-req">800 x 470, png</span>
              <span class="upload-label">Обложка</span>
            </template>
          </div>

          <div class="media-item" :class="{ uploaded: !!project?.video_path, empty: !project?.video_path }">
            <template v-if="project?.video_path">
              <CheckCircle class="icon-md text-green" />
              <span class="m-title">Загружено</span>
              <span class="m-req">До 12 МБ</span>
              <span class="upload-label">Видео</span>
            </template>
            <template v-else>
              <Film class="icon-md" />
              <span class="m-title">Не загружено</span>
              <span class="m-req">До 12 МБ</span>
              <span class="upload-label">Видео</span>
            </template>
          </div>
        </div>
      </div>

      <!-- БЛОК 3: БИЛД -->
      <div class="card form-section">
        <div class="section-head"><h3>Билд</h3></div>

        <div v-if="activeBuildDisplay !== 'не выбрана'" class="build-info">
          <div class="build-success-box">
            <CheckCircle class="icon-md text-green" />
            <div>
              <span style="display:block; font-weight:600;">Билд {{ activeBuildDisplay }} активен</span>
              <span style="display:block; font-size:0.85rem; color:var(--success);">Версия прошла модерацию и доступна игрокам.</span>
            </div>
          </div>
        </div>
        <div v-else class="build-info">
          <div class="build-empty-box">
            <AlertCircle class="icon-md" style="color: var(--warning);" />
            <div>
              <span style="display:block; font-weight:600;">Активный билд не выбран</span>
              <span style="display:block; font-size:0.85rem; color:var(--text-muted);">Загрузите билд во вкладке "Черновик" и выберите активную версию.</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { CheckCircle, Image as ImageIcon, Film, AlertCircle } from 'lucide-vue-next'
import { showToast } from '../../store'
import { getProject, listBuilds } from '../../api/projects'

const route = useRoute()
const projectId = computed(() => route.params.id)
const project = ref(null)
const builds = ref([])

async function loadProject() {
  try {
    project.value = await getProject(projectId.value)
    builds.value = await listBuilds(projectId.value)
  } catch (err) {
    showToast('Не удалось загрузить данные проекта', 'danger')
  }
}
const unpublish = () => showToast('Игра снята с публикации', 'info')

onMounted(loadProject)
</script>

<style scoped>
.tab-fade-in { animation: fadeIn 0.3s ease; }
@keyframes fadeIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }

.form-grid { display: flex; flex-direction: column; gap: 24px; max-width: 800px; padding-bottom: 60px; }
.form-toolbar { display: flex; justify-content: space-between; align-items: flex-end; }
.actions { display: flex; gap: 12px; }
.btn-prod-link { display: flex; align-items: center; gap: 6px; padding: 8px 16px; border: 1px solid var(--success); border-radius: var(--radius-md); background: transparent; color: var(--success); font-weight: 600; font-size: 0.85rem; cursor: pointer; transition: 0.2s; }
.btn-prod-link:hover { background: var(--success-light); }

.status-badge { padding: 4px 10px; border-radius: 12px; font-size: 0.8rem; font-weight: 600; display: inline-block;}
.bg-green { background: var(--success-light); color: var(--success); }
.section-head { margin-bottom: 20px; border-bottom: 1px solid var(--border); padding-bottom: 12px; }
.section-head h3 { margin: 0; font-size: 1.1rem; }
.version-info { font-size: 0.85rem; color: var(--text-muted); margin-top: 4px; margin-bottom: 0; }

.input-row { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; margin-bottom: 16px; }
.input-group label { display: block; font-size: 0.85rem; font-weight: 600; margin-bottom: 8px; }
.readonly-field { padding: 10px 12px; background: var(--bg-secondary); border-radius: var(--radius-md); color: var(--text-main); font-size: 0.9rem; border: 1px solid var(--border); }
.readonly-field.multiline { line-height: 1.5; }

.media-list { display: flex; flex-direction: column; gap: 12px; align-items: center; }
.media-item { border-radius: var(--radius-md); display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 6px; text-align: center; padding: 14px; width: 510px; height: 110px; }
.media-item.uploaded { border: 1px solid var(--success); background: var(--success-light); color: var(--success); }
.media-item.empty { border: 1px dashed var(--border); background: var(--bg-secondary); color: var(--text-muted); }
.media-item .icon-md { width: 16px; height: 16px; }
.text-green { color: var(--success); }
.m-title { font-size: 0.8rem; font-weight: 600; }
.media-item.uploaded .m-title { color: var(--success); }
.m-req { font-size: 0.65rem; }
.upload-label { font-size: 0.8rem; font-weight: 600; color: var(--text-main); margin-top: 4px; }

.build-info { margin-top: 24px; }
.build-success-box { padding: 24px; border: 1px solid var(--success); border-radius: var(--radius-md); background: var(--success-light); display: flex; align-items: center; gap: 16px; }
.build-empty-box { padding: 24px; border: 1px solid var(--warning); border-radius: var(--radius-md); background: var(--warning-light); display: flex; align-items: center; gap: 16px; }
</style>
