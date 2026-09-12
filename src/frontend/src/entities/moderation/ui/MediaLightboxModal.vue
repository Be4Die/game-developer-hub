<template>
  <div class="lightbox-backdrop" @click.self="$emit('close')" @keydown.esc="$emit('close')">
    <div class="lightbox-toolbar">
      <span class="lightbox-title">{{ fileName || 'Изображение' }}</span>
      <div class="lightbox-actions">
        <a
          v-if="downloadUrl"
          :href="downloadUrl"
          download
          class="lightbox-btn"
          title="Скачать оригинал"
        >
          <Download class="icon-sm" />
          <span>Скачать</span>
        </a>
        <button class="lightbox-btn btn-close" title="Закрыть (Esc)" @click="$emit('close')">
          <X class="icon-sm" />
        </button>
      </div>
    </div>
    <div class="lightbox-body" @click.self="$emit('close')">
      <img :src="src" :alt="fileName || 'Full view'" class="lightbox-image" />
    </div>
  </div>
</template>

<script setup>
import { onMounted, onUnmounted } from 'vue';
import { Download, X } from 'lucide-vue-next';

defineProps({
  src: {
    type: String,
    required: true,
  },
  fileName: {
    type: String,
    default: '',
  },
  downloadUrl: {
    type: String,
    default: '',
  },
});

const emit = defineEmits(['close']);

function handleKeyDown(e) {
  if (e.key === 'Escape') {
    emit('close');
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleKeyDown);
  document.body.style.overflow = 'hidden';
});

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeyDown);
  document.body.style.overflow = '';
});
</script>

<style scoped>
.lightbox-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.85);
  backdrop-filter: blur(4px);
  z-index: 9999;
  display: flex;
  flex-direction: column;
  animation: fadeIn 0.15s ease-out;
}

.lightbox-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 20px;
  background: rgba(13, 17, 23, 0.9);
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  color: #f0f6fc;
}

.lightbox-title {
  font-size: 14px;
  font-weight: 500;
  max-width: 60%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.lightbox-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.lightbox-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.15);
  color: #f0f6fc;
  border-radius: 6px;
  padding: 6px 12px;
  font-size: 12px;
  cursor: pointer;
  text-decoration: none;
  transition: all 0.15s ease;
}

.lightbox-btn:hover {
  background: rgba(255, 255, 255, 0.2);
  border-color: rgba(255, 255, 255, 0.3);
}

.lightbox-btn.btn-close:hover {
  background: rgba(239, 68, 68, 0.3);
  border-color: #ef4444;
  color: #ef4444;
}

.lightbox-body {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  overflow: auto;
}

.lightbox-image {
  max-width: 95vw;
  max-height: 85vh;
  object-fit: contain;
  border-radius: 6px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.5);
  animation: zoomIn 0.15s ease-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

@keyframes zoomIn {
  from {
    transform: scale(0.95);
  }
  to {
    transform: scale(1);
  }
}
</style>
