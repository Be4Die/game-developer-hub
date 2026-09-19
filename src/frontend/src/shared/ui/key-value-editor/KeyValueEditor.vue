<template>
  <div class="kv-editor">
    <div v-for="(pair, i) in pairs" :key="i" class="kv-row">
      <input
        v-model="pair.key"
        type="text"
        class="kv-input kv-key"
        placeholder="Ключ"
        @input="emitUpdate"
      />
      <span class="kv-sep">=</span>
      <input
        v-model="pair.value"
        type="text"
        class="kv-input kv-val"
        placeholder="Значение"
        @input="emitUpdate"
      />
      <button class="kv-remove" title="Удалить" @click="remove(i)">&times;</button>
    </div>
    <button class="kv-add" @click="add">+ Добавить</button>
  </div>
</template>

<script setup lang="ts">
import { reactive, watch } from 'vue';

interface Pair {
  key: string;
  value: string;
}

const props = withDefaults(
  defineProps<{
    modelValue?: Record<string, any>;
  }>(),
  {
    modelValue: () => ({}),
  }
);

const emit = defineEmits<{
  (e: 'update:modelValue', value: Record<string, string>): void;
}>();

const pairs = reactive<Pair[]>(
  Object.entries(props.modelValue || {}).map(([key, value]) => ({ key, value: String(value ?? '') }))
);

watch(
  () => props.modelValue,
  (val) => {
    pairs.length = 0;
    if (val) {
      Object.entries(val).forEach(([key, value]) => pairs.push({ key, value: String(value ?? '') }));
    }
  },
  { deep: true }
);

function add(): void {
  pairs.push({ key: '', value: '' });
}

function remove(index: number): void {
  pairs.splice(index, 1);
  emitUpdate();
}

function emitUpdate(): void {
  const obj: Record<string, string> = {};
  for (const p of pairs) {
    if (p.key) obj[p.key] = p.value;
  }
  emit('update:modelValue', obj);
}
</script>

<style scoped>
.kv-editor {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.kv-row {
  display: flex;
  align-items: center;
  gap: 8px;
}
.kv-input {
  padding: 6px 10px;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  background: var(--bg-input);
  color: var(--text-main);
  font-size: 0.85rem;
}
.kv-key {
  width: 140px;
}
.kv-val {
  flex: 1;
}
.kv-sep {
  color: var(--text-muted);
  font-weight: 600;
}
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
.kv-add:hover {
  background: var(--bg-hover);
}
</style>
