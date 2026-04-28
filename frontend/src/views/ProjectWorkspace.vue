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
      <div class="chat-messages scrollable" ref="chatContainer">
        <div v-if="!activeChat" class="message system">Черновик создан ({{ new Date().toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'}) }})</div>
        <div v-for="msg in messages" :key="msg.id" class="message-wrap" :class="msg.role">
          <div class="message-bubble">
            <div class="message-meta">
              <span class="message-author">{{ msg.author }}</span>
              <span class="message-time">{{ formatTime(msg.timestamp) }}</span>
            </div>
            <div class="message-text">{{ msg.text }}</div>
          </div>
        </div>
      </div>
      <div class="chat-input-area">
        <input 
          type="text" 
          placeholder="Написать..." 
          class="chat-input" 
          v-model="newMessage" 
          @keyup.enter="handleSend"
          :disabled="sending"
        />
        <button class="send-btn" @click="handleSend" :disabled="sending || !newMessage.trim()">
          <Send class="icon-sm" />
        </button>
      </div>
    </aside>
  </div>
</template>

<script setup>
import { ref, onMounted, watch, nextTick } from 'vue'
import { ArrowLeft, BarChart2, PenTool, CheckCircle, Server, MessageSquare, Send } from 'lucide-vue-next'
import { useRoute } from 'vue-router'
import * as chatApi from '../api/chat'
import { useAuth } from '../store/auth'
import { showToast } from '../store'

const props = defineProps(['id'])
const route = useRoute()
const { state: authState } = useAuth()

const activeChat = ref(null)
const messages = ref([])
const newMessage = ref('')
const sending = ref(false)
const chatContainer = ref(null)

const formatTime = (ts) => {
  if (!ts) return ''
  const date = new Date(ts)
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

const scrollToBottom = () => {
  nextTick(() => {
    if (chatContainer.value) {
      chatContainer.value.scrollTop = chatContainer.value.scrollHeight
    }
  })
}

const loadChatAndMessages = async () => {
  try {
    // Ищем существующий чат-тикет для этого проекта
    // В реальности мы бы искали по какому-то признаку связи с проектом
    // Для демо ищем чат с типом TICKET и названием, содержащим ID проекта
    const res = await chatApi.listChats({ type: 2 })
    const chat = (res.chats || []).find(c => c.title.includes(props.id))
    
    if (chat) {
      activeChat.value = chat
      const msgRes = await chatApi.getMessages(chat.id)
      messages.value = (msgRes.messages || []).map(m => ({
        id: m.id,
        author: m.author_id === authState.user?.id ? 'Вы' : 'Модератор',
        text: m.content,
        timestamp: m.created_at,
        role: m.author_id === authState.user?.id ? 'developer' : 'moderator'
      }))
      scrollToBottom()
    }
  } catch (err) {
    console.error('Failed to load chat:', err)
  }
}

const handleSend = async () => {
  if (!newMessage.value.trim() || sending.value) return
  
  sending.value = true
  try {
    let chatId = activeChat.value?.id
    
    // Если чата еще нет, создаем его при первом сообщении
    if (!chatId) {
      const createRes = await chatApi.createChat({
        type: 2, // TICKET
        title: `Тикет по проекту #${props.id}`,
        participant_ids: [authState.user?.id, '35031c7b-0038-48f2-8e4f-3fc4a2b7bb6b']
      })
      activeChat.value = createRes.chat
      chatId = createRes.chat.id
    }
    
    await chatApi.sendMessage(chatId, { content: newMessage.value })
    newMessage.value = ''
    await loadChatAndMessages()
  } catch (err) {
    console.error('Failed to send message:', err)
    showToast('Ошибка отправки сообщения', 'danger')
  } finally {
    sending.value = false
  }
}

onMounted(() => {
  if (route.name === 'draft') {
    loadChatAndMessages()
  }
})

watch(() => route.name, (newName) => {
  if (newName === 'draft') {
    loadChatAndMessages()
  }
})
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
.chat-messages { flex: 1; padding: 16px; display: flex; flex-direction: column; gap: 12px; background: #F9FAFB; }
.message.system { align-self: center; color: var(--text-muted); font-size: 0.75rem; background: none; margin-bottom: 8px; }

.message-wrap { display: flex; width: 100%; }
.message-wrap.developer { justify-content: flex-end; }
.message-wrap.moderator { justify-content: flex-start; }

.message-bubble { max-width: 85%; padding: 8px 12px; border-radius: 12px; font-size: 0.85rem; line-height: 1.4; }
.message-wrap.developer .message-bubble { background: var(--primary); color: white; border-bottom-right-radius: 2px; }
.message-wrap.moderator .message-bubble { background: white; border: 1px solid var(--border); color: var(--text-main); border-bottom-left-radius: 2px; }

.message-meta { display: flex; justify-content: space-between; gap: 8px; margin-bottom: 4px; font-size: 0.7rem; opacity: 0.8; }
.message-author { font-weight: 700; }
.message-time { white-space: nowrap; }

.chat-input-area { padding: 16px; border-top: 1px solid var(--border); display: flex; gap: 8px; background: white;}
.chat-input { flex: 1; padding: 10px 14px; border: 1px solid var(--border); border-radius: 20px; outline: none; font-size: 0.9rem; background: var(--bg-app); color: var(--text-main); }
.chat-input:focus { border-color: var(--primary); }
.send-btn { background: var(--primary); color: white; border: none; border-radius: 50%; width: 38px; height: 38px; display: flex; justify-content: center; align-items: center; cursor: pointer; transition: 0.2s; flex-shrink: 0; }
.send-btn:hover:not(:disabled) { background: var(--primary-hover); }
.send-btn:disabled { opacity: 0.5; cursor: not-allowed; }
</style>