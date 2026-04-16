<template>
  <div class="page-container">
    <h1>История тикетов</h1>

    <div class="tickets-list">
      <div
          v-for="ticket in closedTickets"
          :key="ticket.id"
          class="ticket-card"
          @click="$router.push(`/moderator/tickets/${ticket.id}`)"
      >
        <div class="ticket-header">
          <span class="ticket-id">#{{ ticket.id }}</span>
          <span class="ticket-status status-resolved">Решён</span>
        </div>
        <h3 class="ticket-title">{{ ticket.title }}</h3>
        <p class="ticket-desc">{{ ticket.description }}</p>
        <div class="ticket-meta">
          <span>Приоритет: {{ ticket.priority }}</span>
          <span>Закрыт: {{ ticket.closedAt || ticket.created }}</span>
        </div>
      </div>
    </div>

    <div v-if="closedTickets.length === 0" class="empty-state">
      Нет закрытых тикетов
    </div>
  </div>
</template>

<script setup>
import { tickets } from '../store'

const closedTickets = tickets.filter(t => t.status === 'resolved')
</script>

<style scoped>
.page-container {
  padding: 32px 40px;
  max-width: 1000px;
  margin: 0 auto;
}
h1 {
  margin-bottom: 24px;
}
.tickets-list {
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.ticket-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  padding: 20px;
  cursor: pointer;
  transition: 0.2s;
}
.ticket-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0,0,0,0.05);
}
.ticket-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}
.ticket-id {
  font-weight: 600;
  color: var(--text-muted);
  font-size: 0.85rem;
}
.ticket-status {
  font-size: 0.75rem;
  padding: 4px 10px;
  border-radius: 20px;
  font-weight: 600;
}
.status-resolved {
  background: #D1FAE5;
  color: #059669;
}
.ticket-title {
  margin: 0 0 8px 0;
  font-size: 1.1rem;
}
.ticket-desc {
  color: var(--text-muted);
  margin: 0 0 12px 0;
  font-size: 0.9rem;
}
.ticket-meta {
  display: flex;
  gap: 16px;
  font-size: 0.8rem;
  color: var(--text-muted);
}
.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: var(--text-muted);
  background: var(--bg-card);
  border-radius: var(--radius-lg);
  border: 1px dashed var(--border);
}
</style>