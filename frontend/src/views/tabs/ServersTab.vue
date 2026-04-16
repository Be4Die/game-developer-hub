<template>
  <div class="tab-content tab-fade-in">
    <div class="header">
      <h1 style="margin: 0;">Игровые сервера</h1>
      <button class="btn-primary" @click="addServer">Развернуть новый</button>
    </div>
    <div class="servers-list">
      <div v-for="srv in servers" :key="srv.id" class="server-card">
        <div class="info">
          <h3 style="margin: 0 0 4px 0; color: var(--text-main);">Инстанс #{{ srv.id }}</h3>
          <span class="status text-green">Работает (CPU: 12%, RAM: 256MB)</span>
        </div>
        <button class="btn-outline" @click="goToConsole(srv.id)">Перейти в консоль сервера →</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { showToast } from '../../store'
const servers = ref([{ id: 'srv-8821' }, { id: 'srv-9910' }])
const addServer = () => showToast('Запрос на выделение сервера отправлен')
const goToConsole = (id) => alert(`Редирект в Подсистему Оркестрации для сервера ${id}`)
</script>

<style scoped>
.tab-fade-in { animation: fadeIn 0.3s ease; }
@keyframes fadeIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }
.tab-content { padding: 32px; max-width: 800px; }
.header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px;}
.servers-list { display: flex; flex-direction: column; gap: 16px; }
.server-card { display: flex; justify-content: space-between; align-items: center; background: var(--bg-card); padding: 20px; border-radius: 12px; border: 1px solid var(--border);}
.text-green { color: var(--success); font-size: 0.85rem; font-weight: 600;}
</style>