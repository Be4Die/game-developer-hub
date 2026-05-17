<template>
    <div class="page-container">
        <div class="header-row">
            <div>
                <div class="page-subtitle">Модерация</div>
                <h1>Панель модератора</h1>
            </div>
        </div>

        <div class="dashboard-grid">
            <div class="dashboard-section">
                <div class="section-header">
                    <h2>Непрочитанные сообщения</h2>
                    <span class="badge badge-warning" v-if="unreadCount > 0">{{ unreadCount }}</span>
                </div>
                <div class="feeds-list" v-if="conversations.length > 0">
                    <div
                        v-for="conv in conversations"
                        :key="conv.id"
                        class="feed-card card card-hover"
                        @click="openConversation(conv.id)"
                    >
                        <div class="feed-header">
                            <span class="feed-author">{{ conv.participant_name }}</span>
                            <span class="feed-unread" v-if="conv.unread_count > 0">
                                {{ conv.unread_count }} новых
                            </span>
                        </div>
                        <p class="feed-preview">{{ conv.last_message || 'Нет сообщений' }}</p>
                        <div class="feed-time">{{ formatTime(conv.last_message_at) }}</div>
                    </div>
                </div>
                <div class="empty-state" v-else>
                    <p>Нет непрочитанных сообщений</p>
                </div>
            </div>

            <div class="dashboard-section">
                <div class="section-header">
                    <h2>Заявки на модерацию</h2>
                    <span class="badge badge-info" v-if="tickets.length > 0">{{ tickets.length }}</span>
                </div>
                <div class="feeds-list" v-if="tickets.length > 0">
                    <div
                        v-for="ticket in tickets"
                        :key="ticket.id"
                        class="feed-card card card-hover"
                        @click="$router.push(`/moderator/tickets/${ticket.id}`)"
                    >
                        <div class="feed-header">
                            <span class="feed-id">#{{ ticket.id }}</span>
                            <span class="badge" :class="statusBadgeClass(ticket.status)">
                                {{ statusText(ticket.status) }}
                            </span>
                        </div>
                        <h3 class="feed-title">{{ ticket.title }}</h3>
                        <p class="feed-preview">{{ ticket.description }}</p>
                        <div class="feed-meta">
                            <span>Приоритет: {{ ticket.priority }}</span>
                            <span>{{ ticket.created }}</span>
                        </div>
                    </div>
                </div>
                <div class="empty-state" v-else>
                    <p>Нет заявок на модерацию</p>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { chatApi } from '../api/chat'
import { tickets } from '../store'

const router = useRouter()
const conversations = ref([])
const unreadCount = ref(0)

onMounted(async () => {
    try {
        const data = await chatApi.getConversations()
        conversations.value = (data.conversations || [])
        unreadCount.value = await chatApi.getUnreadCount()
    } catch (e) {
        console.error('Failed to load conversations:', e)
    }
})

function openConversation(id) {
    router.push(`/chat/${id}`)
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

function statusText(status) {
    const map = { new: 'Новый', in_progress: 'В работе', resolved: 'Решён' }
    return map[status] || status
}

function statusBadgeClass(status) {
    switch (status) {
        case 'new': return 'badge-warning'
        case 'in_progress': return 'badge-success'
        case 'resolved': return 'badge-neutral'
        default: return 'badge-neutral'
    }
}
</script>

<style scoped>
.page-subtitle {
    font-size: 12px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    color: var(--text-tertiary);
    margin-bottom: 4px;
}

.header-row h1 {
    font-size: 1.5rem;
    font-weight: 700;
    letter-spacing: -0.5px;
    margin: 0;
}

.dashboard-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 24px;
    margin-top: 24px;
}

.dashboard-section {
    display: flex;
    flex-direction: column;
    gap: 16px;
}

.section-header {
    display: flex;
    align-items: center;
    gap: 12px;
}

.section-header h2 {
    font-size: 1.1rem;
    font-weight: 600;
    margin: 0;
}

.feeds-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
}

.feed-card {
    cursor: pointer;
    padding: 16px;
    display: flex;
    flex-direction: column;
    gap: 8px;
}

.feed-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
}

.feed-author, .feed-id {
    font-weight: 600;
    color: var(--text-main);
    font-size: 0.9rem;
}

.feed-unread {
    font-size: 0.75rem;
    color: var(--primary);
    font-weight: 600;
}

.feed-title {
    margin: 0;
    font-size: 1rem;
    font-weight: 600;
    color: var(--text-main);
}

.feed-preview {
    color: var(--text-muted);
    margin: 0;
    font-size: 0.85rem;
    line-height: 1.4;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
}

.feed-time, .feed-meta {
    font-size: 0.75rem;
    color: var(--text-tertiary);
}

.feed-meta {
    display: flex;
    gap: 12px;
}

.empty-state {
    padding: 32px;
    text-align: center;
    color: var(--text-tertiary);
    background: var(--bg-secondary);
    border-radius: 8px;
}

.badge-warning {
    background: var(--warning);
    color: white;
    padding: 2px 8px;
    border-radius: 12px;
    font-size: 0.75rem;
    font-weight: 600;
}

.badge-info {
    background: var(--info);
    color: white;
    padding: 2px 8px;
    border-radius: 12px;
    font-size: 0.75rem;
    font-weight: 600;
}

.badge-success {
    background: var(--success);
    color: white;
    padding: 2px 8px;
    border-radius: 12px;
    font-size: 0.75rem;
}

.badge-neutral {
    background: var(--bg-tertiary);
    color: var(--text-tertiary);
    padding: 2px 8px;
    border-radius: 12px;
    font-size: 0.75rem;
}
</style>
