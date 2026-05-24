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

// --- DRAFT PROJECT (shared between DraftTab and PublishedTab) ---
export const draftProject = reactive({
    meta: {
        titleRu: '',
        titleEn: '',
        seoRu: '',
        seoEn: '',
        about: ''
    },
    media: {
        icon: null,      // { file, preview }
        coverMain: null, // { file, preview }
        video: null      // { file, preview }
    },
    builds: [],          // { version: string, date: string }
    activeBuildVersion: null
})

// --- TICKETS ---
export const tickets = reactive([])

export async function loadTickets() {
  const { moderationApi, moderationToTicket } = await import('../api/moderation.js')
  try {
    const data = await moderationApi.listPending()
    const items = (data.moderations || []).map(moderationToTicket)
    tickets.splice(0, tickets.length, ...items)
  } catch (e) {
    console.error('Failed to load moderation tickets:', e)
  }
}

function upsertTicket(updated) {
  const idx = tickets.findIndex(t => t.id === updated.id)
  if (idx >= 0) {
    Object.assign(tickets[idx], updated)
  } else {
    tickets.push(updated)
  }
}

export async function approveTicket(gameId) {
  const { moderationApi, moderationToTicket } = await import('../api/moderation.js')
  const data = await moderationApi.approve(gameId)
  const updated = moderationToTicket(data.moderation)
  upsertTicket(updated)
  showToast('Игра одобрена', 'success')
  return updated
}

export async function rejectTicket(gameId, reason) {
  const { moderationApi, moderationToTicket } = await import('../api/moderation.js')
  const data = await moderationApi.reject(gameId, reason)
  const updated = moderationToTicket(data.moderation)
  upsertTicket(updated)
  showToast('Игра отклонена', 'info')
  return updated
}