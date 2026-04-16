<template>
  <div class="game-workspace">
    <!-- ЛЕВОЕ МЕНЮ ИГРЫ -->
    <aside class="game-sidebar">
      <div class="game-header">
        <button class="back-btn" @click="$router.push('/projects')"><ArrowLeft class="icon-sm" /> К списку</button>
        <h2 class="game-title-short">Проект #{{ id }}</h2>
      </div>
      <nav class="game-nav">
        <router-link :to="`/projects/${id}/stats`" class="nav-btn" active-class="active"><BarChart2 class="icon-sm" /> Статистика</router-link>
        <router-link :to="`/projects/${id}/draft`" class="nav-btn" active-class="active"><PenTool class="icon-sm" /> Черновик</router-link>
        <router-link :to="`/projects/${id}/published`" class="nav-btn" active-class="active"><CheckCircle class="icon-sm" /> Опубликовано</router-link>
        <router-link :to="`/projects/${id}/servers`" class="nav-btn" active-class="active"><Server class="icon-sm" /> Сервера</router-link>
      </nav>
    </aside>

    <!-- ЦЕНТР (Подгружает табы) -->
    <main class="content-area scrollable">
      <router-view />
    </main>

    <!-- ПРАВЫЙ ЧАТ (Виден только в черновике) -->
    <aside class="chat-sidebar" v-if="$route.name === 'draft'">
      <div class="chat-header"><MessageSquare class="icon-sm" /> <h3>Связь с модератором</h3></div>
      <div class="chat-messages scrollable">
        <div class="message system">Черновик создан (12:00)</div>
      </div>
      <div class="chat-input-area">
        <input type="text" placeholder="Написать..." class="chat-input" />
        <button class="send-btn"><Send class="icon-sm" /></button>
      </div>
    </aside>
  </div>
</template>

<script setup>
import { ArrowLeft, BarChart2, PenTool, CheckCircle, Server, MessageSquare, Send } from 'lucide-vue-next'
import { useRoute } from 'vue-router'
defineProps(['id'])
const route = useRoute()
</script>

<style scoped>
.game-workspace { display: flex; height: calc(100vh - 60px); overflow: hidden; background: var(--bg-app); }
.scrollable { overflow-y: auto; }

/* Левый сайдбар */
.game-sidebar { width: 260px; background: var(--bg-card); border-right: 1px solid var(--border); display: flex; flex-direction: column; flex-shrink: 0;}
.game-header { padding: 20px; border-bottom: 1px solid var(--border); }
.back-btn { background: none; border: none; color: var(--text-muted); display: flex; align-items: center; gap: 6px; cursor: pointer; padding: 0; font-size: 0.85rem; margin-bottom: 12px;}
.game-title-short { margin: 0; font-size: 1.2rem; font-weight: 700; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.game-nav { padding: 12px; display: flex; flex-direction: column; gap: 4px; }
.nav-btn { display: flex; align-items: center; gap: 10px; width: 100%; padding: 10px 12px; border: none; background: transparent; border-radius: var(--radius-md); font-size: 0.9rem; font-weight: 500; color: var(--text-muted); cursor: pointer; text-decoration: none;}
.nav-btn:hover { background: var(--bg-app); color: var(--text-main); }
.nav-btn.active { background: #EFF6FF; color: var(--primary); }

.content-area { flex: 1; padding: 32px 40px; }

/* Правый чат */
.chat-sidebar { width: 320px; background: var(--bg-card); border-left: 1px solid var(--border); display: flex; flex-direction: column; flex-shrink: 0;}
.chat-header { padding: 16px; border-bottom: 1px solid var(--border); display: flex; align-items: center; gap: 8px; }
.chat-header h3 { margin: 0; font-size: 1rem; }
.chat-messages { flex: 1; padding: 16px; display: flex; flex-direction: column; background: #F9FAFB; }
.message.system { align-self: center; color: var(--text-muted); font-size: 0.75rem; background: none; }
.chat-input-area { padding: 16px; border-top: 1px solid var(--border); display: flex; gap: 8px; background: white;}
.chat-input { flex: 1; padding: 10px; border: 1px solid var(--border); border-radius: 20px; outline: none; font-size: 0.9rem;}
.send-btn { background: var(--primary); color: white; border: none; border-radius: 50%; width: 38px; height: 38px; display: flex; justify-content: center; align-items: center; cursor: pointer; }
</style>