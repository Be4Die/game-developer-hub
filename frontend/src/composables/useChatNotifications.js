import { ref, onMounted, onUnmounted } from 'vue'
import { chatApi } from '../api/chat'
import { showToast } from '../store'
import { isAuthenticated } from '../store/auth'

const POLL_INTERVAL = 5000
let lastUnreadCount = null
let pollTimer = null
const isPolling = ref(false)

async function checkUnread() {
  if (!isAuthenticated()) return
  try {
    const count = await chatApi.getUnreadCount()
    if (lastUnreadCount !== null && count > lastUnreadCount) {
      const diff = count - lastUnreadCount
      showToast(
        diff === 1 ? 'Новое сообщение в чате' : `${diff} новых сообщений`,
        'info'
      )
    }
    lastUnreadCount = count
  } catch {
    // ignore polling errors
  }
}

export function startChatNotifications() {
  if (pollTimer) return
  isPolling.value = true
  checkUnread()
  pollTimer = setInterval(checkUnread, POLL_INTERVAL)
}

export function stopChatNotifications() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
  isPolling.value = false
  lastUnreadCount = null
}

export function resetUnreadBaseline() {
  lastUnreadCount = null
}

export function useChatNotifications() {
  onMounted(startChatNotifications)
  onUnmounted(stopChatNotifications)
  return { isPolling }
}

/**
 * Poll messages in an open conversation and notify about incoming ones.
 */
export function useConversationPolling(conversationId, currentUserId, onUpdate) {
  let timer = null
  let knownIds = new Set()
  let initialized = false

  async function poll() {
    const id = typeof conversationId === 'function' ? conversationId() : conversationId.value ?? conversationId
    if (!id) return
    try {
      const data = await chatApi.getMessages(id, 50)
      const msgs = (data.messages || []).reverse()
      const uid = typeof currentUserId === 'function' ? currentUserId() : currentUserId.value ?? currentUserId

      for (const msg of msgs) {
        if (!knownIds.has(msg.id)) {
          if (initialized && (msg.sender_id || msg.senderId) !== uid) {
            const name = msg.sender_name || msg.senderName || 'Собеседник'
            showToast(`Новое сообщение от ${name}`, 'info')
          }
          knownIds.add(msg.id)
        }
      }
      initialized = true

      await chatApi.markAsRead(id)
      onUpdate(msgs)
    } catch {
      // ignore
    }
  }

  onMounted(() => {
    poll()
    timer = setInterval(poll, 4000)
  })

  onUnmounted(() => {
    if (timer) clearInterval(timer)
  })

  return { refresh: poll }
}
