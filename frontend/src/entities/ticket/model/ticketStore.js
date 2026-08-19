import { reactive } from 'vue';
import { moderationApi } from '../api/moderationApi';
import { moderationToTicket } from './helpers';
import { showToast } from '@/shared/lib';

export const tickets = reactive([]);

export async function loadTickets() {
  try {
    const data = await moderationApi.listPending();
    const items = (data.moderations || []).map(moderationToTicket);
    tickets.splice(0, tickets.length, ...items);
  } catch (e) {
    console.error('Failed to load moderation tickets:', e);
  }
}

export function upsertTicket(updated) {
  const idx = tickets.findIndex((t) => t.id === updated.id);
  if (idx >= 0) {
    Object.assign(tickets[idx], updated);
  } else {
    tickets.push(updated);
  }
}

export async function approveTicket(gameId, comment = 'Одобрено') {
  await moderationApi.approve(gameId, comment);
  await loadTickets();
  showToast('Игра одобрена и опубликована в Prod!', 'success');
}

export async function rejectTicket(gameId, reason) {
  await moderationApi.reject(gameId, reason);
  await loadTickets();
  showToast('Игра отклонена', 'info');
}
