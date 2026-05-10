<template>
  <span class="tooltip-wrapper">
    <HelpCircle class="tooltip-icon" @mouseenter="show = true" @mouseleave="show = false" @click="toggle" />
    <Transition name="tooltip">
      <div v-if="show" class="tooltip-box" :class="position">
        <div class="tooltip-arrow"></div>
        <div class="tooltip-content">
          <slot />
        </div>
      </div>
    </Transition>
  </span>
</template>

<script setup>
import { ref } from 'vue'
import { HelpCircle } from 'lucide-vue-next'

const props = defineProps({
  position: { type: String, default: 'top' },
})

const show = ref(false)

function toggle() {
  show.value = !show.value
}
</script>

<style scoped>
.tooltip-wrapper {
  display: inline-flex;
  position: relative;
  vertical-align: middle;
  margin-left: 6px;
}

.tooltip-icon {
  width: 16px;
  height: 16px;
  color: var(--text-muted);
  cursor: pointer;
  transition: color 0.15s;
  flex-shrink: 0;
}
.tooltip-icon:hover {
  color: var(--primary);
}

.tooltip-box {
  position: absolute;
  z-index: 100;
  width: 340px;
  pointer-events: none;
}

.tooltip-box.top {
  bottom: calc(100% + 8px);
  left: 50%;
  transform: translateX(-50%);
}
.tooltip-box.bottom {
  top: calc(100% + 8px);
  left: 50%;
  transform: translateX(-50%);
}
.tooltip-box.right {
  left: calc(100% + 8px);
  top: 50%;
  transform: translateY(-50%);
}
.tooltip-box.left {
  right: calc(100% + 8px);
  top: 50%;
  transform: translateY(-50%);
}

.tooltip-arrow {
  position: absolute;
  width: 0;
  height: 0;
  border-style: solid;
}

.tooltip-box.top .tooltip-arrow {
  top: 100%;
  left: 50%;
  transform: translateX(-50%);
  border-width: 6px 6px 0;
  border-color: var(--bg-card) transparent transparent transparent;
}
.tooltip-box.bottom .tooltip-arrow {
  bottom: 100%;
  left: 50%;
  transform: translateX(-50%);
  border-width: 0 6px 6px;
  border-color: transparent transparent var(--bg-card) transparent;
}
.tooltip-box.right .tooltip-arrow {
  right: 100%;
  top: 50%;
  transform: translateY(-50%);
  border-width: 6px 6px 6px 0;
  border-color: transparent var(--bg-card) transparent transparent;
}
.tooltip-box.left .tooltip-arrow {
  left: 100%;
  top: 50%;
  transform: translateY(-50%);
  border-width: 6px 0 6px 6px;
  border-color: transparent transparent transparent var(--bg-card);
}

.tooltip-content {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  padding: 12px 14px;
  font-size: 0.82rem;
  line-height: 1.5;
  color: var(--text-main);
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
  white-space: pre-line;
}

.tooltip-enter-active,
.tooltip-leave-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}
.tooltip-enter-from,
.tooltip-leave-to {
  opacity: 0;
  transform: translateX(-50%) translateY(4px);
}
.tooltip-box.right.tooltip-enter-from,
.tooltip-box.right.tooltip-leave-to,
.tooltip-box.left.tooltip-enter-from,
.tooltip-box.left.tooltip-leave-to {
  transform: translateY(-50%) translateX(4px);
}
</style>
