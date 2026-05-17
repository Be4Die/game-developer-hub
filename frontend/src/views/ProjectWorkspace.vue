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
    <main class="content-area">
      <router-view />
    </main>

    <!-- ПРАВЫЙ ЧАТ (Виден только в черновике) -->
    <aside class="chat-sidebar" v-if="isDraftRoute">
      <div class="chat-header"><MessageSquare class="icon-sm" /> <h3>Связь с модератором</h3></div>
      <div class="chat-messages scrollable" ref="messagesContainer">
        <div v-if="loading" class="message system">Загрузка...</div>
        <div v-else-if="messages.length === 0" class="message system">Нет сообщений</div>
        <div
          v-else
          v-for="msg in messages"
          :key="msg.id"
          class="message"
          :class="{ own: msg.sender_id === currentUserId }"
        >
          <div v-if="msg.sender_id !== currentUserId" class="msg-author">{{ msg.sender_name || 'Модератор' }}</div>
          <div class="msg-content">{{ msg.content }}</div>
          <div class="msg-time">{{ formatTime(msg.created_at) }}</div>
        </div>
      </div>
      <div class="chat-input-area">
        <input
          v-model="newMessage"
          type="text"
          placeholder="Написать..."
          class="chat-input"
          @keyup.enter="sendMessage"
        />
        <button class="send-btn" @click="() => { console.log('Button clicked'); sendMessage(); }" :disabled="!newMessage.trim()"><Send class="icon-sm" /></button>
      </div>
    </aside>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick, computed, watch } from 'vue'
import { ArrowLeft, BarChart2, PenTool, CheckCircle, Server, MessageSquare, Send } from 'lucide-vue-next'
import { useRoute } from 'vue-router'
import { chatApi } from '../api/chat'
import { getModerators } from '../api/sso'
import { useAuth } from '../store/auth'

defineProps(['id'])
const route = useRoute()

const isDraftRoute = computed(() => {
  return route.name === 'draft' || route.path.includes('/draft')
})

const { state: authState } = useAuth()
const currentUserId = computed(() => authState.user?.id)

const messages = ref([])
const newMessage = ref('')
const loading = ref(false)
const messagesContainer = ref(null)
const conversationId = ref(null)

async function initChat() {
  loading.value = true
  try {
    const moderators = await getModerators()
    if (moderators.length === 0) {
      loading.value = false
      return
    }
    const moderator = moderators[0]
    const conv = await chatApi.createConversation(moderator.id, moderator.name || moderator.username || 'Модератор')
    conversationId.value = conv.conversation?.id
    if (conversationId.value) {
      await loadMessages()
    }
  } catch (e) {
    console.error('Failed to init chat:', e)
  } finally {
    loading.value = false
  }
}

async function loadMessages() {
  if (!conversationId.value) return
  try {
    const data = await chatApi.getMessages(conversationId.value)
    messages.value = (data.messages || []).reverse()
    await chatApi.markAsRead(conversationId.value)
    await nextTick()
    scrollToBottom()
  } catch (e) {
    console.error('Failed to load messages:', e)
  }
}

async function sendMessage() {
  if (!newMessage.value.trim() || !conversationId.value) return
  try {
    await chatApi.sendMessage(conversationId.value, newMessage.value)
    newMessage.value = ''
    await loadMessages()
  } catch (e) {
    console.error('Failed to send message:', e)
  }
}

function formatTime(timestamp) {
  if (!timestamp) return ''
  const date = new Date(timestamp * 1000)
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

function scrollToBottom() {
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
  }
}

onMounted(() => {
  initChat()
})

watch(() => route.name, (newName) => {
  if (newName === 'draft') {
    initChat()
  }
})
</script>

<style scoped>
    .game-workspace { display: flex; min-height: calc(100vh - 60px); background: var(--bg-app); }
.scrollable { overflow-y: auto; }

/* Левый сайдбар */
.game-sidebar { width: 260px; background: var(--bg-card); border-right: 1px solid var(--border); display: flex; flex-direction: column; flex-shrink: 0;}
.game-header { padding: 20px; border-bottom: 1px solid var(--border); }
.back-btn { background: none; border: none; color: var(--text-muted); display: flex; align-items: center; gap: 6px; cursor: pointer; padding: 0; font-size: 0.85rem; margin-bottom: 12px;}
.game-title-short { margin: 0; font-size: 1.2rem; font-weight: 700; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.game-nav { padding: 12px; display: flex; flex-direction: column; gap: 4px; }
.nav-btn { display: flex; align-items: center; gap: 10px; width: 100%; padding: 10px 12px; border: none; background: transparent; border-radius: var(--radius-md); font-size: 0.9rem; font-weight: 500; color: var(--text-muted); cursor: pointer; text-decoration: none;}
.nav-btn:hover { background: var(--bg-app); color: var(--text-main); }
.nav-btn.active { background: var(--primary-light); color: var(--primary); }

.content-area { flex: 1; padding: 32px 40px; }

/* Правый чат */
.chat-sidebar { width: 320px; background: var(--bg-card); border-left: 1px solid var(--border); display: flex; flex-direction: column; flex-shrink: 0;}
.chat-header { padding: 16px; border-bottom: 1px solid var(--border); display: flex; align-items: center; gap: 8px; }
.chat-header h3 { margin: 0; font-size: 1rem; }
.chat-messages { flex: 1; padding: 16px; display: flex; flex-direction: column; background: var(--bg-app); }
.message.system { align-self: center; color: var(--text-muted); font-size: 0.75rem; background: none; }
.message { max-width: 80%; padding: 8px 12px; border-radius: 12px; background: var(--bg-secondary); align-self: flex-start; }
.message.own { align-self: flex-end; background: var(--primary-light); color: var(--primary); }
.msg-author { font-size: 0.75rem; font-weight: 600; margin-bottom: 4px; color: var(--primary); }
.msg-content { font-size: 0.85rem; word-break: break-word; }
.msg-time { font-size: 0.7rem; margin-top: 4px; opacity: 0.7; }
.chat-input-area { padding: 16px; border-top: 1px solid var(--border); display: flex; gap: 8px; background: var(--bg-card);}
.chat-input { flex: 1; padding: 10px; border: 1px solid var(--border); border-radius: 20px; outline: none; font-size: 0.9rem;}
.send-btn { background: var(--primary); color: white; border: none; border-radius: 50%; width: 38px; height: 38px; display: flex; justify-content: center; align-items: center; cursor: pointer; }
</style>