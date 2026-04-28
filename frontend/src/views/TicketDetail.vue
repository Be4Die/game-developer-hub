<template>
  <div class="ticket-detail" v-if="ticket">
    <button class="back-btn" @click="$router.push('/moderator/tickets')">
      <span class="back-arrow">←</span> Назад к тикетам
    </button>

    <div class="ticket-layout">

      <!-- Левая колонка: инфо -->
      <div class="ticket-sidebar">
        <div class="sidebar-card">
          <div class="ticket-id">#{{ ticket.id }}</div>
          <h2 class="ticket-title">{{ ticket.title }}</h2>

          <span class="status-badge" :class="`status-${ticket.status}`">
            {{ statusText(ticket.status) }}
          </span>

          <div class="meta-list">
            <div class="meta-item">
              <span class="meta-label">Приоритет</span>
              <span class="meta-value" :class="`priority-${ticket.priority}`">{{ ticket.priority }}</span>
            </div>
            <div class="meta-item">
              <span class="meta-label">Создан</span>
              <span class="meta-value">{{ ticket.created }}</span>
            </div>
            <div class="meta-item" v-if="ticket.developerName">
              <span class="meta-label">Назначен</span>
              <span class="meta-value">{{ ticket.developerName }}</span>
            </div>
          </div>

          <div class="description-block">
            <div class="description-label">Описание</div>
            <p class="description-text">{{ ticket.description }}</p>
          </div>
        </div>

        <!-- Действия -->
        <div class="ticket-actions">
          <button v-if="ticket.status === 'new'" class="btn-action btn-primary" @click="takeTicket">
            Взять в работу
          </button>
          <button v-if="ticket.status === 'in_progress'" class="btn-action btn-success" @click="resolveTicket">
            ✓ Закрыть тикет
          </button>
          <template v-if="ticket.status === 'resolved'">
            <div class="ticket-resolved">✓ Тикет закрыт</div>
            <button class="btn-action btn-reopen" @click="handleReopen">
              ↩ Открыть повторно
            </button>
          </template>
        </div>

        <button class="history-link" @click="$router.push('/moderator/history')">
          📋 История тикетов
        </button>
      </div>

      <!-- Правая колонка: чат -->
      <div class="chat-column">
        <div class="chat-header">
          <h3>Обсуждение</h3>
          <span class="messages-count">{{ ticket.messages.length }} сообщений</span>
        </div>

        <div class="chat-messages" ref="chatContainer">
          <div v-if="ticket.messages.length === 0" class="chat-empty">
            Сообщений пока нет. Начните обсуждение.
          </div>
          <div v-for="msg in ticket.messages" :key="msg.id" class="message-wrap" :class="msg.role">
            <div class="message-bubble" :class="msg.role">
              <div class="message-meta">
                <span class="message-author">{{ msg.author }}</span>
                <span class="message-time">{{ msg.timestamp }}</span>
              </div>
              <div class="message-text">{{ msg.text }}</div>
            </div>
          </div>
        </div>

        <div class="chat-input-area">
          <textarea
              v-model="newMessage"
              placeholder="Напишите сообщение... Ctrl+Enter для отправки"
              rows="3"
              @keyup.ctrl.enter="sendMessage"
          ></textarea>
          <div class="input-footer">
            <span class="input-hint">Ctrl+Enter — отправить</span>
            <button class="send-btn" @click="sendMessage" :disabled="!newMessage.trim()">Отправить</button>
          </div>
        </div>
      </div>

    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, nextTick, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { tickets, addMessage, loadMessages, loadTickets, showToast } from '../store'
import { useAuth } from '../store/auth'

const route = useRoute()
const router = useRouter()
const { state: authState } = useAuth()
const ticketId = route.params.id
const ticket = computed(() => tickets.find(t => t.id === ticketId))

onMounted(async () => {
  if (tickets.length === 0) {
    await loadTickets()
  }
  if (ticket.value) {
    await loadMessages(ticketId, authState.user?.id)
  } else {
    router.push('/moderator/tickets')
  }
  scrollToBottom()
})

const newMessage = ref('')
const chatContainer = ref(null)

const statusText = (s) => ({ new: 'Новый', in_progress: 'В работе', resolved: 'Решён' }[s] || s)

const scrollToBottom = () => nextTick(() => {
  if (chatContainer.value) chatContainer.value.scrollTop = chatContainer.value.scrollHeight
})

watch(() => ticket.value?.messages.length, scrollToBottom)

const sendMessage = () => {
  if (!newMessage.value.trim()) return
  addMessage(ticketId, newMessage.value, authState.user?.id)
  newMessage.value = ''
}

const takeTicket = () => showToast('Функция взятия в работу в разработке', 'info')
const resolveTicket = () => showToast('Функция закрытия тикета в разработке', 'info')
const handleReopen = () => showToast('Функция переоткрытия в разработке', 'info')
</script>

<style scoped>
.ticket-detail { padding: 28px 40px; max-width: 1300px; margin: 0 auto; width: 100%; box-sizing: border-box; }
.back-btn { display: inline-flex; align-items: center; gap: 6px; background: none; border: none; color: var(--text-muted); cursor: pointer; font-size: 0.9rem; margin-bottom: 20px; padding: 6px 0; font-weight: 500; transition: color 0.2s; }
.back-btn:hover { color: var(--text-main); }
.back-arrow { font-size: 1.1rem; }
.ticket-layout { display: grid; grid-template-columns: 320px 1fr; gap: 24px; align-items: start; }

.ticket-sidebar { display: flex; flex-direction: column; gap: 12px; }
.sidebar-card { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 24px; display: flex; flex-direction: column; gap: 16px; }
.ticket-id { font-size: 0.8rem; font-weight: 700; color: var(--text-muted); letter-spacing: 0.05em; }
.ticket-title { margin: 0; font-size: 1.2rem; font-weight: 700; color: var(--text-main); line-height: 1.4; }
.status-badge { display: inline-block; padding: 4px 12px; border-radius: 20px; font-size: 0.78rem; font-weight: 700; width: fit-content; }
.status-new { background: #FEF3C7; color: #D97706; }
.status-in_progress { background: #EFF6FF; color: var(--primary); }
.status-resolved { background: #D1FAE5; color: #059669; }
.meta-list { display: flex; flex-direction: column; gap: 10px; padding: 16px 0; border-top: 1px solid var(--border); border-bottom: 1px solid var(--border); }
.meta-item { display: flex; justify-content: space-between; align-items: center; font-size: 0.88rem; }
.meta-label { color: var(--text-muted); }
.meta-value { font-weight: 600; color: var(--text-main); }
.priority-Высокий { color: #EF4444; }
.priority-Средний { color: #F59E0B; }
.priority-Низкий { color: #10B981; }
.description-label { font-size: 0.8rem; font-weight: 700; color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 8px; }
.description-text { margin: 0; font-size: 0.9rem; color: var(--text-main); line-height: 1.6; }

.ticket-actions { display: flex; flex-direction: column; gap: 8px; }
.btn-action { width: 100%; padding: 11px; border-radius: var(--radius-md); font-weight: 600; font-size: 0.95rem; border: none; cursor: pointer; transition: opacity 0.2s; }
.btn-action:hover { opacity: 0.85; }
.btn-primary { background: var(--primary); color: white; }
.btn-success { background: #10B981; color: white; }
.btn-reopen { background: var(--bg-app); color: var(--text-main); border: 1px solid var(--border) !important; }
.ticket-resolved { text-align: center; padding: 10px; font-size: 0.9rem; font-weight: 600; color: #059669; background: #D1FAE5; border-radius: var(--radius-md); }

.history-link { width: 100%; padding: 10px; border-radius: var(--radius-md); background: none; border: 1px dashed var(--border); color: var(--text-muted); font-size: 0.88rem; font-weight: 500; cursor: pointer; transition: 0.2s; text-align: center; box-sizing: border-box; }
.history-link:hover { border-color: var(--primary); color: var(--primary); background: var(--bg-app); }

.chat-column { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); display: flex; flex-direction: column; min-height: 580px; max-height: calc(100vh - 160px); }
.chat-header { display: flex; justify-content: space-between; align-items: center; padding: 20px 24px 16px; border-bottom: 1px solid var(--border); }
.chat-header h3 { margin: 0; font-size: 1rem; color: var(--text-main); }
.messages-count { font-size: 0.8rem; color: var(--text-muted); font-weight: 500; }
.chat-messages { flex: 1; overflow-y: auto; padding: 20px 24px; display: flex; flex-direction: column; gap: 14px; background: var(--bg-app); }
.chat-empty { margin: auto; color: var(--text-muted); font-size: 0.9rem; text-align: center; }
.message-wrap { display: flex; }
.message-wrap.moderator { justify-content: flex-end; }
.message-wrap.developer { justify-content: flex-start; }
.message-bubble { max-width: 68%; padding: 10px 14px; border-radius: 14px; font-size: 0.9rem; line-height: 1.5; }
.message-bubble.moderator { background: var(--primary); color: white; border-bottom-right-radius: 4px; }
.message-bubble.developer { background: var(--bg-card); border: 1px solid var(--border); color: var(--text-main); border-bottom-left-radius: 4px; }
.message-meta { display: flex; justify-content: space-between; gap: 12px; margin-bottom: 5px; }
.message-author { font-size: 0.72rem; font-weight: 700; }
.message-bubble.moderator .message-author { color: rgba(255,255,255,0.75); }
.message-bubble.developer .message-author { color: var(--text-muted); }
.message-time { font-size: 0.7rem; opacity: 0.6; }
.chat-input-area { padding: 16px 24px; border-top: 1px solid var(--border); display: flex; flex-direction: column; gap: 10px; }
.chat-input-area textarea { width: 100%; padding: 12px 14px; border: 1px solid var(--border); border-radius: var(--radius-md); font-family: inherit; font-size: 0.9rem; resize: none; background: var(--bg-app); color: var(--text-main); box-sizing: border-box; outline: none; transition: border-color 0.2s; }
.chat-input-area textarea:focus { border-color: var(--primary); }
.input-footer { display: flex; justify-content: space-between; align-items: center; }
.input-hint { font-size: 0.78rem; color: var(--text-muted); }
.send-btn { background: var(--primary); color: white; border: none; padding: 9px 22px; border-radius: var(--radius-md); cursor: pointer; font-weight: 600; font-size: 0.9rem; transition: background 0.2s, opacity 0.2s; }
.send-btn:hover:not(:disabled) { background: var(--primary-hover); }
.send-btn:disabled { opacity: 0.4; cursor: not-allowed; }
</style>