<template>
  <div class="ticket-detail" v-if="ticket">
    <button class="back-btn" @click="$router.push('/moderator/tickets')">
      <span class="back-arrow">←</span> Назад к тикетам
    </button>

    <div class="ticket-layout">
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
              <span class="meta-value">{{ ticket.priority }}</span>
            </div>
            <div class="meta-item">
              <span class="meta-label">Создан</span>
              <span class="meta-value">{{ ticket.created }}</span>
            </div>
          </div>

          <div class="description-block">
            <div class="description-label">Описание</div>
            <p class="description-text">{{ ticket.description }}</p>
          </div>

          <div v-if="ticket.status === 'rejected' && ticket.rejectionReason" class="rejection-block">
            <div class="description-label">Причина отказа</div>
            <p class="rejection-text">{{ ticket.rejectionReason }}</p>
          </div>
        </div>

        <div class="ticket-actions" v-if="ticket.status === 'pending'">
          <button class="btn-action btn-success" @click="handleApprove" :disabled="actionLoading">
            ✓ Одобрить игру
          </button>
          <button class="btn-action btn-danger" @click="showRejectModal = true" :disabled="actionLoading">
            ✕ Отклонить игру
          </button>
        </div>

        <div v-else-if="ticket.status === 'approved'" class="result-banner approved">
          ✓ Игра одобрена и может быть опубликована
        </div>
        <div v-else-if="ticket.status === 'rejected'" class="result-banner rejected">
          ✕ Игра отклонена
        </div>
      </div>

      <div class="chat-column">
        <div class="chat-header">
          <h3>Обсуждение с разработчиком</h3>
          <span class="messages-count">{{ messages.length }} сообщений</span>
        </div>

        <div class="chat-messages" ref="chatContainer">
          <div v-if="chatLoading" class="chat-empty">Загрузка чата...</div>
          <div v-else-if="!conversationId" class="chat-empty">Чат с разработчиком пока недоступен</div>
          <div v-else-if="messages.length === 0" class="chat-empty">Сообщений пока нет. Начните обсуждение.</div>
          <div
            v-else
            v-for="msg in messages"
            :key="msg.id"
            class="message-wrap"
            :class="{ own: isOwnMessage(msg) }"
          >
            <div class="message-bubble" :class="{ own: isOwnMessage(msg) }">
              <div class="message-meta">
                <span class="message-author">{{ formatSender(msg) }}</span>
                <span class="message-time">{{ formatTime(msg.created_at) }}</span>
              </div>
              <div class="message-text">{{ msg.content }}</div>
            </div>
          </div>
        </div>

        <div class="chat-input-area" v-if="conversationId">
          <textarea
            v-model="newMessage"
            placeholder="Напишите сообщение... Ctrl+Enter для отправки"
            rows="3"
            @keyup.ctrl.enter="sendMessage"
          ></textarea>
          <div class="input-footer">
            <span class="input-hint">Ctrl+Enter — отправить</span>
            <button class="send-btn" @click="sendMessage" :disabled="!newMessage.trim() || sending">
              {{ sending ? '...' : 'Отправить' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showRejectModal" class="modal-overlay" @click.self="showRejectModal = false">
      <div class="modal-card">
        <h3>Отклонить игру</h3>
        <p class="modal-hint">Укажите причину отказа — разработчик увидит её в карточке заявки.</p>
        <textarea
          v-model="rejectReason"
          placeholder="Например: не загружена иконка, описание слишком короткое..."
          rows="4"
          autofocus
        ></textarea>
        <div class="modal-actions">
          <button class="btn-cancel" @click="showRejectModal = false">Отмена</button>
          <button
            class="btn-confirm-reject"
            @click="handleReject"
            :disabled="!rejectReason.trim() || actionLoading"
          >
            Отклонить
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, nextTick, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { tickets, approveTicket, rejectTicket, showToast } from '../store'
import { moderationApi, moderationToTicket, ticketStatusText } from '../api/moderation'
import { chatApi } from '../api/chat'
import { useAuth } from '../store/auth'
import { useConversationPolling } from '../composables/useChatNotifications'

const route = useRoute()
const router = useRouter()
const ticketId = parseInt(route.params.id, 10)
const { state: authState } = useAuth()

const ticket = ref(null)
const messages = ref([])
const newMessage = ref('')
const chatContainer = ref(null)
const conversationId = ref(null)
const chatLoading = ref(false)
const sending = ref(false)
const actionLoading = ref(false)
const showRejectModal = ref(false)
const rejectReason = ref('')

const statusText = ticketStatusText
const currentUserId = computed(() => authState.user?.id)

useConversationPolling(
  () => conversationId.value,
  () => currentUserId.value,
  (msgs) => {
    messages.value = msgs
    scrollToBottom()
  }
)

async function loadTicket() {
  const local = tickets.find(t => t.id === ticketId)
  if (local) {
    ticket.value = { ...local }
    return
  }
  try {
    const data = await moderationApi.getStatus(ticketId)
    ticket.value = moderationToTicket(data.moderation)
  } catch {
    router.push('/moderator/tickets')
  }
}

async function findConversation() {
  if (!ticket.value?.developerId) return
  chatLoading.value = true
  try {
    const data = await chatApi.getConversations()
    const devId = ticket.value.developerId
    const conv = (data.conversations || []).find(c =>
      c.participant_id === devId || c.participantId === devId
    )
    if (conv) {
      conversationId.value = conv.id
      return
    }
    const created = await chatApi.createConversation(devId, 'Разработчик')
    conversationId.value = created.conversation?.id
  } catch (e) {
    console.error('Failed to init chat:', e)
  } finally {
    chatLoading.value = false
  }
}

function isOwnMessage(msg) {
  return (msg.sender_id || msg.senderId) === currentUserId.value
}

function formatSender(msg) {
  return isOwnMessage(msg) ? 'Вы' : (msg.sender_name || msg.senderName || 'Разработчик')
}

function formatTime(timestamp) {
  if (!timestamp) return ''
  return new Date(timestamp * 1000).toLocaleString('ru-RU', {
    day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit'
  })
}

function scrollToBottom() {
  nextTick(() => {
    if (chatContainer.value) chatContainer.value.scrollTop = chatContainer.value.scrollHeight
  })
}

watch(() => messages.value.length, scrollToBottom)

async function sendMessage() {
  if (!newMessage.value.trim() || !conversationId.value) return
  sending.value = true
  try {
    await chatApi.sendMessage(conversationId.value, newMessage.value)
    newMessage.value = ''
    showToast('Сообщение отправлено', 'success')
    await refreshMessages()
  } catch {
    showToast('Не удалось отправить сообщение', 'error')
  } finally {
    sending.value = false
  }
}

async function refreshMessages() {
  if (!conversationId.value) return
  const data = await chatApi.getMessages(conversationId.value)
  messages.value = (data.messages || []).reverse()
  await chatApi.markAsRead(conversationId.value)
  scrollToBottom()
}

async function handleApprove() {
  actionLoading.value = true
  try {
    ticket.value = await approveTicket(ticketId)
  } catch (e) {
    showToast(e.message || 'Ошибка одобрения', 'error')
  } finally {
    actionLoading.value = false
  }
}

async function handleReject() {
  if (!rejectReason.value.trim()) return
  actionLoading.value = true
  try {
    ticket.value = await rejectTicket(ticketId, rejectReason.value.trim())
    showRejectModal.value = false
    rejectReason.value = ''
  } catch (e) {
    showToast(e.message || 'Ошибка отклонения', 'error')
  } finally {
    actionLoading.value = false
  }
}

onMounted(async () => {
  await loadTicket()
  if (!ticket.value) return
  await findConversation()
  if (conversationId.value) await refreshMessages()
})
</script>

<style scoped>
.ticket-detail { padding: 28px 40px; max-width: 1300px; margin: 0 auto; width: 100%; box-sizing: border-box; }
.back-btn { display: inline-flex; align-items: center; gap: 6px; background: none; border: none; color: var(--text-muted); cursor: pointer; font-size: 0.9rem; margin-bottom: 20px; padding: 6px 0; font-weight: 500; }
.back-btn:hover { color: var(--text-main); }
.ticket-layout { display: grid; grid-template-columns: 320px 1fr; gap: 24px; align-items: start; }
.ticket-sidebar { display: flex; flex-direction: column; gap: 12px; }
.sidebar-card { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 24px; display: flex; flex-direction: column; gap: 16px; }
.ticket-id { font-size: 0.8rem; font-weight: 700; color: var(--text-muted); }
.ticket-title { margin: 0; font-size: 1.2rem; font-weight: 700; }
.status-badge { display: inline-block; padding: 4px 12px; border-radius: 20px; font-size: 0.78rem; font-weight: 700; width: fit-content; }
.status-pending { background: #FEF3C7; color: #D97706; }
.status-approved { background: #D1FAE5; color: #059669; }
.status-rejected { background: #FEE2E2; color: #DC2626; }
.meta-list { display: flex; flex-direction: column; gap: 10px; padding: 16px 0; border-top: 1px solid var(--border); border-bottom: 1px solid var(--border); }
.meta-item { display: flex; justify-content: space-between; font-size: 0.88rem; }
.meta-label { color: var(--text-muted); }
.meta-value { font-weight: 600; }
.description-label { font-size: 0.8rem; font-weight: 700; color: var(--text-muted); text-transform: uppercase; margin-bottom: 8px; }
.description-text { margin: 0; font-size: 0.9rem; line-height: 1.6; }
.rejection-block { padding: 12px; background: #FEF2F2; border-radius: var(--radius-md); border: 1px solid #FECACA; }
.rejection-text { margin: 0; font-size: 0.9rem; color: #B91C1C; line-height: 1.5; }
.ticket-actions { display: flex; flex-direction: column; gap: 8px; }
.btn-action { width: 100%; padding: 11px; border-radius: var(--radius-md); font-weight: 600; border: none; cursor: pointer; }
.btn-action:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-success { background: #10B981; color: white; }
.btn-danger { background: #EF4444; color: white; }
.result-banner { text-align: center; padding: 12px; font-weight: 600; border-radius: var(--radius-md); font-size: 0.9rem; }
.result-banner.approved { color: #059669; background: #D1FAE5; }
.result-banner.rejected { color: #DC2626; background: #FEE2E2; }
.chat-column { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); display: flex; flex-direction: column; min-height: 580px; max-height: calc(100vh - 160px); }
.chat-header { display: flex; justify-content: space-between; padding: 20px 24px 16px; border-bottom: 1px solid var(--border); }
.chat-header h3 { margin: 0; font-size: 1rem; }
.messages-count { font-size: 0.8rem; color: var(--text-muted); }
.chat-messages { flex: 1; overflow-y: auto; padding: 20px 24px; display: flex; flex-direction: column; gap: 14px; background: var(--bg-app); }
.chat-empty { margin: auto; color: var(--text-muted); font-size: 0.9rem; text-align: center; }
.message-wrap { display: flex; }
.message-wrap.own { justify-content: flex-end; }
.message-bubble { max-width: 68%; padding: 10px 14px; border-radius: 14px; font-size: 0.9rem; background: var(--bg-card); border: 1px solid var(--border); border-bottom-left-radius: 4px; }
.message-bubble.own { background: var(--primary); color: white; border: none; border-bottom-right-radius: 4px; border-bottom-left-radius: 14px; }
.message-meta { display: flex; justify-content: space-between; gap: 12px; margin-bottom: 5px; }
.message-author { font-size: 0.72rem; font-weight: 700; }
.message-bubble.own .message-author { color: rgba(255,255,255,0.75); }
.message-time { font-size: 0.7rem; opacity: 0.6; }
.chat-input-area { padding: 16px 24px; border-top: 1px solid var(--border); display: flex; flex-direction: column; gap: 10px; }
.chat-input-area textarea { width: 100%; padding: 12px; border: 1px solid var(--border); border-radius: var(--radius-md); font-family: inherit; font-size: 0.9rem; resize: none; background: var(--bg-app); box-sizing: border-box; }
.input-footer { display: flex; justify-content: space-between; align-items: center; }
.input-hint { font-size: 0.78rem; color: var(--text-muted); }
.send-btn { background: var(--primary); color: white; border: none; padding: 9px 22px; border-radius: var(--radius-md); cursor: pointer; font-weight: 600; }
.send-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.45); display: flex; align-items: center; justify-content: center; z-index: 1000; }
.modal-card { background: var(--bg-card); border-radius: var(--radius-lg); padding: 28px; width: 100%; max-width: 480px; display: flex; flex-direction: column; gap: 16px; margin: 20px; }
.modal-card h3 { margin: 0; }
.modal-hint { margin: 0; font-size: 0.88rem; color: var(--text-muted); }
.modal-card textarea { width: 100%; padding: 12px; border: 1px solid var(--border); border-radius: var(--radius-md); font-family: inherit; box-sizing: border-box; background: var(--bg-app); }
.modal-actions { display: flex; gap: 10px; justify-content: flex-end; }
.btn-cancel { padding: 9px 18px; border: 1px solid var(--border); border-radius: var(--radius-md); background: transparent; cursor: pointer; font-weight: 600; }
.btn-confirm-reject { padding: 9px 18px; border: none; border-radius: var(--radius-md); background: #EF4444; color: white; cursor: pointer; font-weight: 600; }
.btn-confirm-reject:disabled { opacity: 0.5; cursor: not-allowed; }
</style>
