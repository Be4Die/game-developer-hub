<template>
  <div class="chat-view">
    <div class="chat-header-bar">
      <button class="back-btn" @click="$router.push('/moderator')">
        <ArrowLeft class="icon-sm" /> Назад
      </button>
      <h2>{{ conversationName }}</h2>
    </div>
    <div class="messages-container" ref="messagesContainer">
      <div v-if="loading" class="loading">Загрузка...</div>
      <div v-else-if="messages.length === 0" class="empty">Нет сообщений</div>
      <div
        v-else
        v-for="msg in messages"
        :key="msg.id"
        class="message"
        :class="{ own: msg.sender_id === currentUserId }"
      >
        <div class="message-author">{{ formatSender(msg) }}</div>
        <div class="message-content">{{ msg.content }}</div>
        <div class="message-time">{{ formatTime(msg.created_at) }}</div>
      </div>
    </div>
    <div class="input-area">
      <input
        v-model="newMessage"
        type="text"
        placeholder="Введите сообщение..."
        @keyup.enter="sendMessage"
      />
      <button @click="sendMessage" :disabled="!newMessage.trim()">
        <Send class="icon-sm" />
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { ArrowLeft, Send } from 'lucide-vue-next'
import { chatApi } from '../api/chat'
import { useAuth } from '../store/auth'
import { showToast } from '../store'
import { useConversationPolling } from '../composables/useChatNotifications'

const route = useRoute()
const { state: authState } = useAuth()
const currentUserId = ref(authState.user?.id)

const messages = ref([])
const newMessage = ref('')
const loading = ref(false)
const conversationId = ref(route.params.id)
const conversationName = ref('Загрузка...')
const messagesContainer = ref(null)

useConversationPolling(
  () => conversationId.value,
  () => currentUserId.value,
  (msgs) => {
    messages.value = msgs
    scrollToBottom()
  }
)

async function loadConversationInfo() {
  try {
    const data = await chatApi.getConversations()
    const conv = data.conversations?.find(c => c.id === conversationId.value)
    if (conv) {
      conversationName.value = conv.participantName || 'Диалог'
    }
  } catch (e) {
    console.error('Failed to load conversation info:', e)
  }
}

async function loadMessages() {
  loading.value = true
  try {
    const data = await chatApi.getMessages(conversationId.value)
    messages.value = (data.messages || []).reverse()
    await chatApi.markAsRead(conversationId.value)
    await nextTick()
    scrollToBottom()
  } catch (e) {
    console.error('Failed to load messages:', e)
  } finally {
    loading.value = false
  }
}

async function sendMessage() {
  if (!newMessage.value.trim()) return
  try {
    await chatApi.sendMessage(conversationId.value, newMessage.value)
    newMessage.value = ''
    showToast('Сообщение отправлено', 'success')
    await loadMessages()
  } catch (e) {
    console.error('Failed to send message:', e)
    showToast('Не удалось отправить сообщение', 'error')
  }
}

function formatSender(msg) {
  return msg.sender_name || msg.senderName || 'Пользователь'
}

function formatTime(timestamp) {
  if (!timestamp) return ''
  const date = new Date(timestamp * 1000)
  return date.toLocaleString('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

function scrollToBottom() {
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
  }
}

onMounted(() => {
  loadConversationInfo()
  loadMessages()
})
</script>

<style scoped>
.chat-view {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 60px);
  background: var(--bg-card);
  max-width: 1200px;
  width: 100%;
  margin: 0 auto;
  overflow: hidden;
}

.chat-header-bar {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px 24px;
  border-bottom: 1px solid var(--border);
}

.back-btn {
  background: none;
  border: none;
  color: var(--text-muted);
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
}

.chat-header-bar h2 {
  margin: 0;
  font-size: 1.2rem;
}

.messages-container {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.loading, .empty {
  text-align: center;
  color: var(--text-muted);
  padding: 40px;
}

.message {
  max-width: 60%;
  padding: 12px 16px;
  border-radius: 12px;
  background: var(--bg-secondary);
  align-self: flex-start;
}

.message.own {
  align-self: flex-end;
  background: var(--primary);
  color: white;
  margin-left: auto;
}

.message-author {
  font-size: 12px;
  font-weight: 600;
  margin-bottom: 4px;
  color: var(--text-muted);
}

.message.own .message-author {
  color: var(--primary);
}

.message-content {
  font-size: 14px;
  word-break: break-word;
}

.message-time {
  font-size: 11px;
  margin-top: 4px;
  opacity: 0.7;
}

.input-area {
  display: flex;
  gap: 12px;
  padding: 16px 24px;
  border-top: 1px solid var(--border);
}

.input-area input {
  flex: 1;
  padding: 12px 16px;
  border: 1px solid var(--border);
  border-radius: 24px;
  outline: none;
  font-size: 14px;
}

.input-area input:focus {
  border-color: var(--primary);
}

.input-area button {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  background: var(--primary);
  color: white;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

.input-area button:disabled {
  background: var(--border);
  cursor: not-allowed;
}
</style>
