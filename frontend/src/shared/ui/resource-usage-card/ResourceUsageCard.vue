<template>
  <div class="resource-card">
    <span class="resource-label">{{ label }}</span>
    <div class="resource-bar-bg" v-if="typeof percent === 'number'">
      <div class="resource-bar-fill" :style="{ width: clampPercent(percent) + '%' }" :class="barColor" />
    </div>
    <span class="resource-value">{{ displayValue }}</span>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  label: { type: String, required: true },
  value: { type: [Number, String], default: null },
  max:   { type: Number, default: null },
  unit:  { type: String, default: '' },
  type:  { type: String, default: 'percent', validator: v => ['percent', 'bytes', 'raw'].includes(v) },
})

const percent = computed(() => {
  if (props.type === 'percent' && typeof props.value === 'number') return props.value
  if (props.type === 'bytes' && props.max) return (props.value / props.max) * 100
  return null
})

function clampPercent(v) { return Math.min(100, Math.max(0, v)) }

const barColor = computed(() => {
  const p = percent.value
  if (p == null) return ''
  if (p >= 90) return 'bar-danger'
  if (p >= 70) return 'bar-warning'
  return 'bar-ok'
})

function formatBytes(b) {
  if (b == null) return '—'
  if (b < 1024) return b + ' B'
  if (b < 1024 * 1024) return (b / 1024).toFixed(1) + ' KB'
  if (b < 1024 * 1024 * 1024) return (b / (1024 * 1024)).toFixed(1) + ' MB'
  return (b / (1024 * 1024 * 1024)).toFixed(1) + ' GB'
}

const displayValue = computed(() => {
  if (props.value == null) return '—'
  if (props.type === 'percent') return props.value.toFixed(1) + '%'
  if (props.type === 'bytes') {
    const used = formatBytes(props.value)
    return props.max ? `${used} / ${formatBytes(props.max)}` : used
  }
  return props.value + props.unit
})
</script>

<style scoped>
.resource-card {
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  padding: 12px 16px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.resource-label { font-size: 0.8rem; color: var(--text-muted); font-weight: 500; }
.resource-value { font-size: 1rem; font-weight: 700; color: var(--text-main); }
.resource-bar-bg {
  width: 100%;
  height: 6px;
  background: var(--border);
  border-radius: 3px;
  overflow: hidden;
}
.resource-bar-fill {
  height: 100%;
  border-radius: 3px;
  transition: width 0.4s ease;
}
.bar-ok      { background: var(--success); }
.bar-warning { background: var(--warning); }
.bar-danger  { background: var(--danger); }
</style>
