<template>
  <div class="resource-card">
    <div class="resource-header">
      <div class="resource-title-group">
        <component :is="resolvedIcon" v-if="resolvedIcon" class="resource-icon" />
        <span class="resource-label">{{ label }}</span>
      </div>
      <span class="resource-value">{{ displayValue }}</span>
    </div>

    <div class="resource-bar-bg" :class="{ 'bar-placeholder': percent == null }">
      <div
        v-if="typeof percent === 'number'"
        class="resource-bar-fill"
        :style="{ width: clampPercent(percent) + '%' }"
        :class="barColor"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, type Component } from 'vue';
import { Cpu, Layers, HardDrive, Network, Activity } from 'lucide-vue-next';

interface Props {
  label: string;
  value?: number | string | null;
  max?: number | null;
  unit?: string;
  type?: 'percent' | 'bytes' | 'raw' | string;
  icon?: Component | object | null;
}

const props = withDefaults(defineProps<Props>(), {
  value: null,
  max: null,
  unit: '',
  type: 'percent',
  icon: null,
});

const resolvedIcon = computed(() => {
  if (props.icon) return props.icon;
  const l = (props.label || '').toLowerCase();
  if (l.includes('cpu') || l.includes('процессор')) return Cpu;
  if (l.includes('пам') || l.includes('ram') || l.includes('mem')) return Layers;
  if (l.includes('диск') || l.includes('disk')) return HardDrive;
  if (l.includes('сеть') || l.includes('net')) return Network;
  return Activity;
});

const percent = computed<number | null>(() => {
  if (props.type === 'percent' && typeof props.value === 'number') return props.value;
  if (props.type === 'bytes' && props.max && typeof props.value === 'number') {
    return (props.value / props.max) * 100;
  }
  return null;
});

function clampPercent(v: number): number {
  return Math.min(100, Math.max(0, v));
}

const barColor = computed<string>(() => {
  const p = percent.value;
  if (p == null) return '';
  if (p >= 90) return 'bar-danger';
  if (p >= 70) return 'bar-warning';
  return 'bar-ok';
});

function formatBytes(b: number | null): string {
  if (b == null) return '—';
  if (b < 1024) return b + ' B';
  if (b < 1024 * 1024) return (b / 1024).toFixed(1) + ' KB';
  if (b < 1024 * 1024 * 1024) return (b / (1024 * 1024)).toFixed(1) + ' MB';
  return (b / (1024 * 1024 * 1024)).toFixed(1) + ' GB';
}

const displayValue = computed<string>(() => {
  if (props.value == null) return '—';
  if (props.type === 'percent') return Number(props.value).toFixed(1) + '%';
  if (props.type === 'bytes') {
    const numVal = typeof props.value === 'number' ? props.value : Number(props.value);
    const used = formatBytes(isNaN(numVal) ? null : numVal);
    return props.max ? `${used} / ${formatBytes(props.max)}` : used;
  }
  return String(props.value) + props.unit;
});
</script>

<style scoped>
.resource-card {
  background: rgba(255, 255, 255, 0.025);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-md, 8px);
  padding: 12px 14px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 10px;
  box-sizing: border-box;
  transition: border-color 0.15s, background 0.15s;
}

.resource-card:hover {
  border-color: rgba(255, 255, 255, 0.2);
  background: rgba(255, 255, 255, 0.04);
}

.resource-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.resource-title-group {
  display: flex;
  align-items: center;
  gap: 7px;
}

.resource-icon {
  width: 15px;
  height: 15px;
  color: var(--text-muted, #8b949e);
  flex-shrink: 0;
  transition: color 0.15s;
}

.resource-card:hover .resource-icon {
  color: var(--text-main, #f0f6fc);
}

.resource-label {
  font-size: 0.82rem;
  color: var(--text-muted, #8b949e);
  font-weight: 500;
}

.resource-value {
  font-size: 0.88rem;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}

.resource-bar-bg {
  width: 100%;
  height: 5px;
  background: rgba(255, 255, 255, 0.08);
  border-radius: 3px;
  overflow: hidden;
}

.bar-placeholder {
  opacity: 0.15;
}

.resource-bar-fill {
  height: 100%;
  border-radius: 3px;
  transition: width 0.4s ease;
}

.bar-ok {
  background: #3fb950;
}

.bar-warning {
  background: #d29922;
}

.bar-danger {
  background: #f85149;
}
</style>
