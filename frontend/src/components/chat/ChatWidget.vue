<template>
  <div class="chat-widget">
    <button class="chat-toggle" @click="toggleOpen" :class="{ 'has-unread': unreadCount > 0 }">
      <MessageCircle class="icon-md" />
      <span v-if="unreadCount > 0" class="unread-badge">{{ unreadCount > 9 ? '9+' : unreadCount }}</span>
    </button>

    <transition name="chat-slide">
      <div v-if="isOpen" class="chat-panel">
        <div class="chat-header">
          <h3>Сообщения</h3>
          <button class="close-btn" @click="isOpen = false">
            <X class="icon-sm" />
          </button>
        </div>

        <div class="chat-tabs">
          <button
            class="tab-btn"
            :class="{ active: activeTab === 'conversations' }"
            @click="activeTab = 'conversations'"
          >
            Диалоги
          </button>
          <button
            class="tab-btn"
            :class="{ active: activeTab === 'unread' }"
            @click="activeTab = 'unread'"
          >
            Непрочитанные
            <span v-if="unreadCount > 0" class="tab-badge">{{ unreadCount }}</span>
          </button>
        </div>

        <div class="chat-content">
          <div v-if="activeTab === 'conversations'" class="conversations-list">
            <div v-if="loading" class="loading">Загрузка...</div>
            <div v-else-if="conversations.length === 0" class="empty">Нет диалогов</div>
            <div
              v-else
              v-for="conv in conversations"
              :key="conv.id"
              class="conversation-item"
              :class="{ active: selectedConversation === conv.id }"
              @click="selectConversation(conv.id)"
            >
              <div class="conv-avatar">{{ conv.participant_name?.[0]?.toUpperCase() || conv.participantName?.[0]?.toUpperCase() || '?' }}</div>
              <div class="conv-info">
                <div class="conv-name">{{ conv.participant_name || conv.participantName }}</div>
                <div class="conv-preview">{{ conv.last_message || conv.lastMessage || 'Нет сообщений' }}</div>
              </div>
              <div class="conv-meta">
                <span class="conv-time">{{ formatTime(conv.last_message_at || conv.lastMessageAt) }}</span>
                <span v-if="conv.unread_count > 0 || conv.unreadCount > 0" class="unread-count">{{ conv.unread_count || conv.unreadCount }}</span>
              </div>
            </div>
          </div>

          <div v-if="activeTab === 'unread'" class="unread-list">
            <div v-if="unreadMessages.length === 0" class="empty">Нет непрочитанных сообщений</div>
            <div
              v-else
              v-for="msg in unreadMessages"
              :key="msg.id"
              class="unread-item"
              @click="openConversation(msg.conversation_id)"
            >
              <div class="msg-author">{{ msg.sender_name }}</div>
              <div class="msg-text">{{ msg.content }}</div>
              <div class="msg-time">{{ formatTime(msg.created_at) }}</div>
            </div>
          </div>

          <div v-if="selectedConversation" class="messages-view">
            <div class="messages-header">
              <button class="back-btn" @click="selectedConversation = null">
                <ChevronLeft class="icon-sm" />
              </button>
              <span>{{ getConversationName(selectedConversation) }}</span>
            </div>
            <div class="messages-list" ref="messagesContainer">
              <div
                v-for="msg in messages"
                :key="msg.id"
                class="message"
                :class="{ own: msg.sender_id === currentUserId }"
              >
                <div class="msg-author" v-if="msg.sender_id !== currentUserId">{{ formatSender(msg) }}</div>
                <div class="msg-content">{{ msg.content }}</div>
                <div class="msg-time">{{ formatTime(msg.created_at) }}</div>
              </div>
            </div>
            <div class="message-input">
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
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, nextTick } from 'vue'
import { MessageCircle, X, ChevronLeft, Send } from 'lucide-vue-next'
import { chatApi } from '../../api/chat'
import { useAuth } from '../../store/auth'

const { state: authState } = useAuth()
const currentUserId = computed(() => authState.user?.id)

const isOpen = ref(false)
const activeTab = ref('conversations')
const loading = ref(false)
const conversations = ref([])
const messages = ref([])
const unreadMessages = ref([])
const selectedConversation = ref(null)
const newMessage = ref('')
const unreadCount = ref(0)
const messagesContainer = ref(null)

async function loadConversations() {
  loading.value = true
  try {
    const data = await chatApi.getConversations()
    conversations.value = data.conversations || []
  } catch (e) {
    console.error('Failed to load conversations:', e)
  } finally {
    loading.value = false
  }
}

async function loadUnreadMessages() {
  try {
    const allConvs = await chatApi.getConversations()
    const convs = allConvs.conversations || []
    const unread = []
    for (const conv of convs) {
      if ((conv.unread_count || conv.unreadCount) > 0) {
        const data = await chatApi.getMessages(conv.id, conv.unread_count || conv.unreadCount, 0)
        unread.push(...(data.messages || []).filter(m => !m.is_read && !m.isRead))
      }
    }
    unreadMessages.value = unread
  } catch (e) {
    console.error('Failed to load unread:', e)
  }
}

async function loadMessages(convId) {
  try {
    const data = await chatApi.getMessages(convId)
    messages.value = (data.messages || []).reverse()
    await chatApi.markAsRead(convId)
    await refreshUnreadCount()
    await loadConversations()
    await nextTick()
    scrollToBottom()
  } catch (e) {
    console.error('Failed to load messages:', e)
  }
}

async function sendMessage() {
  if (!newMessage.value.trim() || !selectedConversation.value) return
  try {
    await chatApi.sendMessage(selectedConversation.value, newMessage.value)
    newMessage.value = ''
    await loadMessages(selectedConversation.value)
  } catch (e) {
    console.error('Failed to send message:', e)
  }
}

async function refreshUnreadCount() {
  try {
    unreadCount.value = await chatApi.getUnreadCount()
  } catch (e) {
    unreadCount.value = 0
  }
}

function selectConversation(convId) {
  selectedConversation.value = convId
  loadMessages(convId)
}

function openConversation(convId) {
  activeTab.value = 'conversations'
  selectConversation(convId)
}

function getConversationName(convId) {
  const conv = conversations.value.find(c => c.id === convId)
  return conv?.participantName || 'Диалог'
}

function formatSender(msg) {
  return msg.sender_name || msg.senderName || 'Пользователь'
}

function formatTime(timestamp) {
  if (!timestamp) return ''
  const date = new Date(timestamp * 1000)
  const now = new Date()
  const diff = now - date
  if (diff < 60000) return 'сейчас'
  if (diff < 3600000) return `${Math.floor(diff / 60000)}м`
  if (diff < 86400000) return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  return date.toLocaleDateString([], { day: '2-digit', month: '2-digit' })
}

function scrollToBottom() {
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
  }
}

function toggleOpen() {
  isOpen.value = !isOpen.value
  if (isOpen.value) {
    loadConversations()
    loadUnreadMessages()
    refreshUnreadCount()
  }
}

watch(activeTab, (tab) => {
  if (tab === 'unread') {
    loadUnreadMessages()
  }
})

onMounted(() => {
  refreshUnreadCount()
  setInterval(refreshUnreadCount, 30000)
})
</script>

<style scoped>
.chat-widget {
  position: fixed;
  bottom: 24px;
  right: 24px;
  z-index: 1000;
}

.chat-toggle {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: var(--primary);
  color: white;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: var(--shadow-lg);
  transition: all 0.2s;
  position: relative;
}

.chat-toggle:hover {
  transform: scale(1.05);
}

.chat-toggle.has-unread {
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%, 100% { box-shadow: var(--shadow-lg); }
  50% { box-shadow: 0 0 0 8px rgba(99, 102, 241, 0.2); }
}

.unread-badge {
  position: absolute;
  top: -4px;
  right: -4px;
  background: var(--danger);
  color: white;
  font-size: 11px;
  font-weight: 600;
  min-width: 20px;
  height: 20px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 6px;
}

.chat-panel {
  position: absolute;
  bottom: 72px;
  right: 0;
  width: 450px;
  height: 520px;
  background: var(--bg-card);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-xl);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.chat-slide-enter-active,
.chat-slide-leave-active {
  transition: all 0.3s ease;
}

.chat-slide-enter-from,
.chat-slide-leave-to {
  opacity: 0;
  transform: translateY(20px) scale(0.95);
}

.chat-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  border-bottom: 1px solid var(--border);
}

.chat-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
}

.close-btn {
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 4px;
}

.chat-tabs {
  display: flex;
  border-bottom: 1px solid var(--border);
}

.tab-btn {
  flex: 1;
  padding: 12px;
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  font-weight: 500;
  position: relative;
}

.tab-btn.active {
  color: var(--primary);
}

.tab-btn.active::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 2px;
  background: var(--primary);
}

.tab-badge {
  background: var(--danger);
  color: white;
  font-size: 10px;
  padding: 2px 6px;
  border-radius: 8px;
  margin-left: 4px;
}

.chat-content {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.conversations-list,
.unread-list {
  flex: 1;
  overflow-y: auto;
}

.loading,
.empty {
  padding: 32px;
  text-align: center;
  color: var(--text-muted);
}

.conversation-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  cursor: pointer;
  border-bottom: 1px solid var(--border);
  transition: background 0.15s;
}

.conversation-item:hover,
.conversation-item.active {
  background: var(--bg-secondary);
}

.conv-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: var(--primary-light);
  color: var(--primary);
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
}

.conv-info {
  flex: 1;
  min-width: 0;
}

.conv-name {
  font-weight: 500;
  font-size: 14px;
}

.conv-preview {
  font-size: 12px;
  color: var(--text-muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.conv-meta {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 4px;
}

.conv-time {
  font-size: 11px;
  color: var(--text-tertiary);
}

.unread-count {
  background: var(--primary);
  color: white;
  font-size: 10px;
  min-width: 18px;
  height: 18px;
  border-radius: 9px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 4px;
}

.unread-item {
  padding: 12px 16px;
  border-bottom: 1px solid var(--border);
  cursor: pointer;
}

.unread-item:hover {
  background: var(--bg-secondary);
}

.msg-author {
  font-weight: 600;
  font-size: 13px;
  margin-bottom: 4px;
}

.msg-text {
  font-size: 13px;
  color: var(--text-muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.msg-time {
  font-size: 11px;
  color: var(--text-tertiary);
  margin-top: 4px;
}

.messages-view {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.messages-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border);
  font-weight: 500;
}

.back-btn {
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 4px;
}

.messages-list {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.message {
  max-width: 75%;
  padding: 8px 12px;
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

.msg-content {
  font-size: 13px;
  word-break: break-word;
}

.message .msg-time {
  font-size: 10px;
  margin-top: 4px;
  opacity: 0.7;
}

.message-input {
  display: flex;
  gap: 8px;
  padding: 12px;
  border-top: 1px solid var(--border);
}

.message-input input {
  flex: 1;
  padding: 8px 12px;
  border: 1px solid var(--border);
  border-radius: 20px;
  outline: none;
  font-size: 13px;
}

.message-input input:focus {
  border-color: var(--primary);
}

.message-input button {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: var(--primary);
  color: white;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

.message-input button:disabled {
  background: var(--border);
  cursor: not-allowed;
}
</style>
