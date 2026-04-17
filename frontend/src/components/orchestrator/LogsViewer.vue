<template>
  <div class="logs-viewer">
    <div class="logs-toolbar">
      <div class="toolbar-left">
        <label class="toolbar-item">
          <input type="checkbox" v-model="follow" /> Стриминг
        </label>
        <label class="toolbar-item">
          Хвост
          <input type="number" v-model.number="tail" min="0" max="1000" class="toolbar-input" />
        </label>
        <label class="toolbar-item">
          Источник
          <select v-model="sourceFilter" class="toolbar-select">
            <option value="all">Все</option>
            <option value="stdout">stdout</option>
            <option value="stderr">stderr</option>
          </select>
        </label>
      </div>
      <div class="toolbar-right">
        <button class="toolbar-btn" @click="clearLogs" title="Очистить">Очистить</button>
        <button class="toolbar-btn" @click="copyLogs" title="Копировать">Копировать</button>
      </div>
    </div>
    <div class="logs-terminal" ref="terminal" @scroll="onScroll">
      <div v-if="filteredEntries.length === 0" class="logs-empty">Нет записей</div>
      <div
        v-for="(entry, i) in filteredEntries"
        :key="i"
        class="log-line"
        :class="'log-' + entry.source"
      >
        <span class="log-ts">{{ formatTime(entry.timestamp) }}</span>
        <span class="log-src">[{{ entry.source }}]</span>
        <span class="log-msg">{{ entry.message }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { generateMockLogEntries } from '../../data/mock-orchestrator'

const props = defineProps({
  instanceId: { type: [Number, String], default: null },
})

const terminal = ref(null)
const follow = ref(true)
const tail = ref(100)
const sourceFilter = ref('all')
const entries = ref([])

// Начальная загрузка моковых логов
onMounted(() => {
  entries.value = generateMockLogEntries(20)
  scrollToBottom()
})

// Имитация стриминга новых логов
let streamInterval = null
watch(follow, (val) => {
  if (val) {
    streamInterval = setInterval(() => {
      const newEntries = generateMockLogEntries(1)
      entries.value.push(...newEntries)
      if (entries.value.length > 500) entries.value = entries.value.slice(-300)
      nextTick(scrollToBottom)
    }, 2000)
  } else {
    clearInterval(streamInterval)
    streamInterval = null
  }
}, { immediate: true })

onUnmounted(() => {
  clearInterval(streamInterval)
})

const filteredEntries = computed(() => {
  if (sourceFilter.value === 'all') return entries.value
  return entries.value.filter(e => e.source === sourceFilter.value)
})

function formatTime(ts) {
  const d = new Date(ts)
  return d.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

function scrollToBottom() {
  if (follow.value && terminal.value) {
    terminal.value.scrollTop = terminal.value.scrollHeight
  }
}

function onScroll() {
  if (!terminal.value) return
  const { scrollTop, scrollHeight, clientHeight } = terminal.value
  if (scrollHeight - scrollTop - clientHeight > 50) {
    follow.value = false
  }
}

function clearLogs() {
  entries.value = []
}

function copyLogs() {
  const text = filteredEntries.value
    .map(e => `${formatTime(e.timestamp)} [${e.source}] ${e.message}`)
    .join('\n')
  navigator.clipboard.writeText(text).catch(() => {})
}
</script>

<style scoped>
.logs-viewer {
  display: flex;
  flex-direction: column;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  overflow: hidden;
  background: #1e1e2e;
}
.logs-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: #2a2a3c;
  border-bottom: 1px solid #3a3a4c;
  flex-wrap: wrap;
  gap: 8px;
}
.toolbar-left, .toolbar-right { display: flex; align-items: center; gap: 12px; }
.toolbar-item {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 0.78rem;
  color: #cdd6f4;
  cursor: pointer;
}
.toolbar-input {
  width: 60px;
  padding: 2px 6px;
  border: 1px solid #3a3a4c;
  border-radius: 4px;
  background: #1e1e2e;
  color: #cdd6f4;
  font-size: 0.78rem;
}
.toolbar-select {
  padding: 2px 6px;
  border: 1px solid #3a3a4c;
  border-radius: 4px;
  background: #1e1e2e;
  color: #cdd6f4;
  font-size: 0.78rem;
}
.toolbar-btn {
  padding: 4px 10px;
  border: 1px solid #3a3a4c;
  border-radius: 4px;
  background: #2a2a3c;
  color: #cdd6f4;
  cursor: pointer;
  font-size: 0.78rem;
}
.toolbar-btn:hover { background: #3a3a4c; }
.logs-terminal {
  height: 340px;
  overflow-y: auto;
  padding: 12px;
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  font-size: 0.8rem;
  line-height: 1.6;
}
.logs-empty { color: #6c7086; text-align: center; padding: 40px 0; }
.log-line { white-space: pre-wrap; word-break: break-all; }
.log-ts { color: #6c7086; margin-right: 8px; }
.log-src { color: #89b4fa; margin-right: 8px; }
.log-stdout .log-msg { color: #cdd6f4; }
.log-stderr .log-msg { color: #f38ba8; }
</style>
