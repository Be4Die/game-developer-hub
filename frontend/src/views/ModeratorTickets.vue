<template>
    <div class="page-container">
        <div class="header-row">
            <div>
                <div class="page-subtitle">Модерация</div>
                <h1>Очередь тикетов</h1>
            </div>
        </div>

        <div class="tickets-list">
            <div
                v-for="ticket in tickets"
                :key="ticket.id"
                class="card ticket-card card-hover"
                @click="$router.push(`/moderator/tickets/${ticket.id}`)"
            >
                <div class="ticket-header">
                    <span class="ticket-id">#{{ ticket.id }}</span>
                    <span class="badge" :class="statusBadgeClass(ticket.status)">
                        <span class="status-dot"></span>
                        {{ statusText(ticket.status) }}
                    </span>
                </div>
                <h3 class="ticket-title">{{ ticket.title }}</h3>
                <p class="ticket-desc">{{ ticket.description }}</p>
                <div class="ticket-meta">
                    <span>Приоритет: {{ ticket.priority }}</span>
                    <span>Создан: {{ ticket.created }}</span>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { ticketStatusText } from '../api/moderation'
import { tickets, loadTickets } from '../store'

onMounted(() => loadTickets())

function statusText(status) {
    return ticketStatusText(status)
}

function statusBadgeClass(status) {
    switch (status) {
        case 'pending': return 'badge-warning'
        case 'approved': return 'badge-success'
        case 'rejected': return 'badge-neutral'
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

.tickets-list {
    display: flex;
    flex-direction: column;
    gap: 16px;
}

.ticket-card {
    cursor: pointer;
    padding: 24px;
    display: flex;
    flex-direction: column;
    gap: 8px;
}

.ticket-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
}

.ticket-id {
    font-weight: 600;
    color: var(--text-tertiary);
    font-size: 0.85rem;
    font-family: monospace;
}

.ticket-title {
    margin: 0;
    font-size: 1.1rem;
    font-weight: 600;
    color: var(--text-main);
}

.ticket-desc {
    color: var(--text-muted);
    margin: 0;
    font-size: 0.9rem;
    line-height: 1.5;
}

.ticket-meta {
    display: flex;
    gap: 16px;
    font-size: 0.8rem;
    color: var(--text-tertiary);
    margin-top: 4px;
}

.status-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: currentColor;
}
</style>
