// Mock Moderation API (in-memory & localStorage persistent)

const STORAGE_KEY = 'gdh_mock_moderation_v1'

function getStoredData() {
  const defaultData = {
    moderations: [
      {
        game_id: 1,
        gameId: 1,
        game_name: 'Space Runner',
        gameName: 'Space Runner',
        game_description: 'Космический раннер с динамическими препятствиями и таблицей рекордов.',
        gameDescription: 'Космический раннер с динамическими препятствиями и таблицей рекордов.',
        developer_id: '1',
        developerId: '1',
        status: 'MODERATION_STATUS_PENDING',
        rejection_reason: '',
        rejectionReason: '',
        submitted_at: Math.floor((Date.now() - 1000 * 60 * 60 * 3) / 1000),
      },
      {
        game_id: 2,
        gameId: 2,
        game_name: 'Puzzle Quest',
        gameName: 'Puzzle Quest',
        game_description: 'Логическая головоломка с разнообразными уровнями.',
        gameDescription: 'Логическая головоломка с разнообразными уровнями.',
        developer_id: '2',
        developerId: '2',
        status: 'MODERATION_STATUS_APPROVED',
        rejection_reason: '',
        rejectionReason: '',
        submitted_at: Math.floor((Date.now() - 1000 * 60 * 60 * 24) / 1000),
      },
      {
        game_id: 3,
        gameId: 3,
        game_name: 'Tower Defense',
        gameName: 'Tower Defense',
        game_description: 'Стратегия с защитой базы от наступающих волн противников.',
        gameDescription: 'Стратегия с защитой базы от наступающих волн противников.',
        developer_id: '3',
        developerId: '3',
        status: 'MODERATION_STATUS_REJECTED',
        rejection_reason: 'Не загружена обложка 800x470, отсутствует описание механик.',
        rejectionReason: 'Не загружена обложка 800x470, отсутствует описание механик.',
        submitted_at: Math.floor((Date.now() - 1000 * 60 * 60 * 48) / 1000),
      },
    ],
  }

  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (raw) return JSON.parse(raw)
  } catch {
    // fallback
  }
  return defaultData
}

function saveStoredData(data) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(data))
  } catch {
    // ignore
  }
}

export const moderationApi = {
  async submitForReview(gameId, gameName, gameDescription) {
    const data = getStoredData()
    const id = Number(gameId)
    let item = (data.moderations || []).find((m) => Number(m.game_id ?? m.gameId) === id)

    const now = Math.floor(Date.now() / 1000)
    if (!item) {
      item = {
        game_id: id,
        gameId: id,
        game_name: gameName,
        gameName: gameName,
        game_description: gameDescription,
        gameDescription: gameDescription,
        developer_id: '1',
        developerId: '1',
        status: 'MODERATION_STATUS_PENDING',
        rejection_reason: '',
        rejectionReason: '',
        submitted_at: now,
      }
      if (!data.moderations) data.moderations = []
      data.moderations.unshift(item)
    } else {
      item.game_name = gameName
      item.gameName = gameName
      item.game_description = gameDescription
      item.gameDescription = gameDescription
      item.status = 'MODERATION_STATUS_PENDING'
      item.rejection_reason = ''
      item.rejectionReason = ''
      item.submitted_at = now
    }

    saveStoredData(data)
    return { moderation: item }
  },

  async listPending(limit = 50, offset = 0) {
    const data = getStoredData()
    const list = (data.moderations || []).filter(
      (m) => m.status === 'MODERATION_STATUS_PENDING' || m.status === 'pending'
    )
    const paged = list.slice(offset, offset + limit)
    return {
      moderations: paged,
    }
  },

  async getStatus(gameId) {
    const data = getStoredData()
    const id = Number(gameId)
    const item = (data.moderations || []).find((m) => Number(m.game_id ?? m.gameId) === id)
    if (!item) return null
    return { moderation: item }
  },

  async approve(gameId) {
    const data = getStoredData()
    const id = Number(gameId)
    let item = (data.moderations || []).find((m) => Number(m.game_id ?? m.gameId) === id)
    if (!item) {
      item = {
        game_id: id,
        gameId: id,
        game_name: `Игра #${id}`,
        gameName: `Игра #${id}`,
        game_description: '',
        gameDescription: '',
        developer_id: '1',
        developerId: '1',
        status: 'MODERATION_STATUS_APPROVED',
        rejection_reason: '',
        rejectionReason: '',
        submitted_at: Math.floor(Date.now() / 1000),
      }
      if (!data.moderations) data.moderations = []
      data.moderations.push(item)
    } else {
      item.status = 'MODERATION_STATUS_APPROVED'
      item.rejection_reason = ''
      item.rejectionReason = ''
    }

    saveStoredData(data)
    return { moderation: item }
  },

  async reject(gameId, reason) {
    const data = getStoredData()
    const id = Number(gameId)
    let item = (data.moderations || []).find((m) => Number(m.game_id ?? m.gameId) === id)
    if (!item) {
      item = {
        game_id: id,
        gameId: id,
        game_name: `Игра #${id}`,
        gameName: `Игра #${id}`,
        game_description: '',
        gameDescription: '',
        developer_id: '1',
        developerId: '1',
        status: 'MODERATION_STATUS_REJECTED',
        rejection_reason: reason,
        rejectionReason: reason,
        submitted_at: Math.floor(Date.now() / 1000),
      }
      if (!data.moderations) data.moderations = []
      data.moderations.push(item)
    } else {
      item.status = 'MODERATION_STATUS_REJECTED'
      item.rejection_reason = reason
      item.rejectionReason = reason
    }

    saveStoredData(data)
    return { moderation: item }
  },
}

export function moderationToTicket(m) {
  if (!m) return null
  const statusMap = {
    MODERATION_STATUS_PENDING: 'pending',
    MODERATION_STATUS_APPROVED: 'approved',
    MODERATION_STATUS_REJECTED: 'rejected',
    pending: 'pending',
    approved: 'approved',
    rejected: 'rejected',
  }
  const submittedAt = m.submitted_at || m.submittedAt
  let created = ''
  if (submittedAt) {
    const ts = submittedAt.seconds ? submittedAt.seconds * 1000 : typeof submittedAt === 'number' ? (submittedAt < 10000000000 ? submittedAt * 1000 : submittedAt) : new Date(submittedAt).getTime()
    created = new Date(ts).toLocaleDateString('ru-RU')
  }
  return {
    id: Number(m.game_id ?? m.gameId),
    title: m.game_name || m.gameName || `Игра #${m.game_id || m.gameId}`,
    description: m.game_description || m.gameDescription || '',
    status: statusMap[m.status] || 'pending',
    priority: 'Средний',
    created,
    developerId: m.developer_id || m.developerId,
    rejectionReason: m.rejection_reason || m.rejectionReason || '',
    moderationStatus: m.status,
  }
}

export function ticketStatusText(status) {
  const map = {
    pending: 'На модерации',
    approved: 'Одобрено',
    rejected: 'Отклонено',
    new: 'Новый',
    in_progress: 'В работе',
    resolved: 'Решён',
  }
  return map[status] || status
}
