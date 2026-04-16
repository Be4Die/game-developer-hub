<template>
  <div class="uploader-box">
    <div v-if="!isUploading" class="upload-idle">
      <input type="file" accept=".zip,.tar.gz" @change="handleFileSelect" id="file-input" hidden />
      <label for="file-input" class="upload-label">
        Выбрать .zip архив (до 100 МБ)
      </label>
      <p class="hint">или перетащите файл сюда</p>
    </div>

    <div v-else class="upload-progress">
      <div class="progress-info">
        <span>Загрузка билда...</span>
        <span>{{ progress }}%</span>
      </div>
      <div class="progress-bar-bg">
        <div class="progress-bar-fill" :style="{ width: progress + '%' }"></div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const emit = defineProps(['upload-success'])
const progress = ref(0)
const isUploading = ref(false)

const handleFileSelect = (event) => {
  const file = event.target.files[0]
  if (!file) return

  // Здесь в будущем будет валидация на 100МБ:
  // if (file.size > 100 * 1024 * 1024) return alert('Файл слишком большой!')

  isUploading.value = true
  progress.value = 0

  // Имитация загрузки на Go-бэкенд с помощью Axios onUploadProgress
  const interval = setInterval(() => {
    progress.value += 15
    if (progress.value >= 100) {
      progress.value = 100
      clearInterval(interval)
      setTimeout(() => {
        isUploading.value = false
        // Вызываем событие, чтобы родитель (ProjectEdit) обновил таблицу
        document.querySelector('#file-input').value = ''
      }, 500)
    }
  }, 300)
}
</script>

<style scoped>
.uploader-box { border: 2px dashed #90CAF9; border-radius: 8px; padding: 40px; text-align: center; background: #FAFAFA; transition: background 0.3s; }
.uploader-box:hover { background: #E3F2FD; }
.upload-label { display: inline-block; background: white; border: 1px solid #1565C0; color: #1565C0; padding: 10px 20px; border-radius: 6px; cursor: pointer; font-weight: bold; margin-bottom: 10px; }
.hint { color: #757575; font-size: 0.9rem; margin: 0; }

.upload-progress { text-align: left; }
.progress-info { display: flex; justify-content: space-between; font-weight: bold; margin-bottom: 10px; color: #333; }
.progress-bar-bg { width: 100%; height: 12px; background: #E0E0E0; border-radius: 6px; overflow: hidden; }
.progress-bar-fill { height: 100%; background: #4CAF50; transition: width 0.3s ease; }
</style>