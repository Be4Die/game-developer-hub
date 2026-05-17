import { reactive } from 'vue'

// --- USER & TOAST ---
export const user = reactive({
    name: 'Михаил В.',
    email: 'mikhail@welwise.com',
    role: 'Разработчик'
})

export const toast = reactive({ show: false, message: '', type: 'success' })
export const showToast = (message, type = 'success') => {
    toast.message = message
    toast.type = type
    toast.show = true
    setTimeout(() => { toast.show = false }, 3000)
}

// --- TICKETS ---
export const tickets = reactive([])

export const addMessage = (ticketId, text, role) => {
    const ticket = tickets.find(t => t.id === ticketId)
    if (ticket) {
        const author = role === 'moderator' ? 'Модератор' : 'Разработчик'
        ticket.messages.push({ id: Date.now(), author, text, timestamp: new Date().toLocaleString(), role })
        showToast(`Сообщение отправлено в тикет #${ticketId}`, 'success')
    }
}

export const updateTicketStatus = (ticketId, status) => {
    const ticket = tickets.find(t => t.id === ticketId)
    if (ticket) {
        ticket.status = status
        if (status === 'resolved') ticket.closedAt = new Date().toISOString().slice(0,10)
        showToast(`Статус тикета #${ticketId} изменён`, 'info')
    }
}

export const assignTicketToModerator = (ticketId, moderatorName) => {
    const ticket = tickets.find(t => t.id === ticketId)
    if (ticket && ticket.status === 'new') {
        ticket.status = 'in_progress'
        ticket.developerName = moderatorName
        showToast(`Тикет #${ticketId} взят в работу`, 'success')
    }
}

export const reopenTicket = (ticketId) => {
    const ticket = tickets.find(t => t.id === ticketId)
    if (ticket && ticket.status === 'resolved') {
        ticket.status = 'in_progress'
        ticket.closedAt = null
        showToast(`Тикет #${ticketId} открыт повторно`, 'info')
    }
}