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
export const tickets = reactive([
    {
        id: 1, title: 'Некорректное отображение иконки', description: 'Иконка игры не отображается в каталоге на мобильных устройствах.', status: 'resolved', priority: 'Высокий', created: '2026-03-30', closedAt: '2026-03-31', developerName: 'Михаил В.', messages: [
            { id: 1, author: 'Разработчик', text: 'Помогите разобраться с иконкой.', timestamp: '2026-03-30 10:00', role: 'developer' },
            { id: 2, author: 'Модератор', text: 'Принято в работу.', timestamp: '2026-03-30 10:30', role: 'moderator' }
        ]
    },
    {
        id: 2, title: 'Проблема с авторизацией через VK', description: 'Пользователи не могут войти через VK.', status: 'in_progress', priority: 'Средний', created: '2026-03-29', developerName: 'Михаил В.', messages: [
            { id: 1, author: 'Модератор', text: 'Жалобы на VK.', timestamp: '2026-03-29 09:00', role: 'moderator' },
            { id: 2, author: 'Разработчик', text: 'Проверим ключи.', timestamp: '2026-03-29 09:45', role: 'developer' }
        ]
    },
    {
        id: 3, title: 'Запрос на модерацию игры "RIVALS"', description: 'Новый билд, требуется проверка.', status: 'new', priority: 'Низкий', created: '2026-03-28', developerName: null, messages: []
    }
])

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