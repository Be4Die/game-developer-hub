<template>
  <div class="kv-editor">
    <div class="kv-row" v-for="(pair, i) in pairs" :key="i">
      <input
        type="text"
        class="kv-input kv-key"
        v-model="pair.key"
        placeholder="Ключ"
        @input="emitUpdate"
      />
      <span class="kv-sep">=</span>
      <input
        type="text"
        class="kv-input kv-val"
        v-model="pair.value"
        placeholder="Значение"
        @input="emitUpdate"
      />
      <button class="kv-remove" @click="remove(i)" title="Удалить">&times;</button>
    </div>
    <button class="kv-add" @click="add">+ Добавить</button>
  </div>
</template>

<script setup>
import { reactive, watch } from 'vue'

const props = defineProps({
  modelValue: { type: Object, default: () => ({}) },
})
const emit = defineEmits(['update:modelValue'])

const pairs = reactive(
  Object.entries(props.modelValue).map(([key, value]) => ({ key, value }))
)

watch(() => props.modelValue, (val) => {
  pairs.length = 0
  Object.entries(val).forEach(([key, value]) => pairs.push({ key, value }))
}, { deep: true })

function add() {
  pairs.push({ key: '', value: '' })
}

function remove(index) {
  pairs.splice(index, 1)
  emitUpdate()
}

function emitUpdate() {
  const obj = {}
  for (const p of pairs) {
    if (p.key) obj[p.key] = p.value
  }
  emit('update:modelValue', obj)
}
</script>

<style scoped>
.kv-editor { display: flex; flex-direction: column; gap: 8px; }
.kv-row { display: flex; align-items: center; gap: 8px; }
.kv-input {
  padding: 6px 10px;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  background: var(--bg-input);
  color: var(--text-main);
  font-size: 0.85rem;
}
.kv-key { width: 140px; }
.kv-val { flex: 1; }
.kv-sep { color: var(--text-muted); font-weight: 600; }
.kv-remove {
  background: none;
  border: none;
  color: var(--danger);
  font-size: 1.2rem;
  cursor: pointer;
  padding: 0 4px;
  line-height: 1;
}
.kv-add {
  align-self: flex-start;
  background: none;
  border: 1px dashed var(--border);
  color: var(--primary);
  padding: 4px 12px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-size: 0.85rem;
  font-weight: 500;
}
.kv-add:hover { background: var(--bg-hover); }
</style>
