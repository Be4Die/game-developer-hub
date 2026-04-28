import { reactive } from 'vue'
import * as chatApi from '../api/chat'

// --- TOAST ---
export const toast = reactive({ show: false, message: '', type: 'success' })
export const showToast = (message, type = 'success') => {
    toast.message = message
    toast.type = type
    toast.show = true
    setTimeout(() => { toast.show = false }, 3000)
}

// --- TICKETS ---
export const tickets = reactive([])

export const loadTickets = async () => {
    try {
        const res = await chatApi.listChats({ type: 2 }) // CHAT_TYPE_TICKET
        tickets.splice(0, tickets.length, ...(res.chats || []).map(c => ({
            id: c.id,
            title: c.title,
            status: 'in_progress',
            priority: 'Средний',
            created: c.created_at,
            messages: []
        })))
    } catch (err) {
        console.error('Failed to load tickets:', err)
        showToast('Ошибка загрузки тикетов', 'danger')
    }
}

export const loadMessages = async (ticketId, currentUserId) => {
    try {
        const res = await chatApi.getMessages(ticketId)
        const ticket = tickets.find(t => t.id === ticketId)
        if (ticket) {
            ticket.messages = (res.messages || []).map(m => ({
                id: m.id,
                author: m.author_id === currentUserId ? 'Вы' : 'Разработчик',
                text: m.content,
                timestamp: m.created_at,
                role: m.author_id === currentUserId ? 'moderator' : 'developer'
            }))
        }
    } catch (err) {
        console.error('Failed to load messages:', err)
    }
}

export const addMessage = async (ticketId, text, currentUserId) => {
    try {
        await chatApi.sendMessage(ticketId, { content: text })
        await loadMessages(ticketId, currentUserId)
        showToast(`Сообщение отправлено`, 'success')
    } catch (err) {
        console.error('Failed to send message:', err)
        showToast('Ошибка отправки сообщения', 'danger')
    }
}