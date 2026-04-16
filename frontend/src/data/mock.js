export const requests = [
    {
        id: '1',
        gameTitle: 'Space Runner',
        version: '1.0.0',
        developer: 'Иван',
        status: 'pending',
        priority: 'high',
        createdAt: '2025-04-01 10:15',
        comment: '',
    },
    {
        id: '2',
        gameTitle: 'Puzzle Quest',
        version: '1.2.1',
        developer: 'Анна',
        status: 'in_review',
        priority: 'medium',
        createdAt: '2025-04-01 11:30',
        comment: 'Проверяем метаданные',
    },
    {
        id: '3',
        gameTitle: 'Tower Defense',
        version: '0.9.5',
        developer: 'Максим',
        status: 'approved',
        priority: 'low',
        createdAt: '2025-03-31 17:20',
        comment: 'Игра соответствует требованиям',
    },
]

export const reports = [
    {
        id: '1',
        gameTitle: 'Space Runner',
        user: 'user123',
        reason: 'Неприемлемый контент',
        status: 'open',
        createdAt: '2025-04-01 12:00',
    },
    {
        id: '2',
        gameTitle: 'Puzzle Quest',
        user: 'player007',
        reason: 'Игра не запускается',
        status: 'resolved',
        createdAt: '2025-03-31 18:00',
    },
]

export const projects = [
    {
        id: '1',
        title: 'Space Runner',
        owner: 'Иван',
        version: '1.0.0',
        status: 'moderation',
    },
    {
        id: '2',
        title: 'Puzzle Quest',
        owner: 'Анна',
        version: '1.2.1',
        status: 'published',
    },
]

export const servers = [
    {
        id: 'srv-1',
        gameTitle: 'Space Runner',
        region: 'EU',
        status: 'running',
        cpu: 46,
        players: 123,
    },
    {
        id: 'srv-2',
        gameTitle: 'Puzzle Quest',
        region: 'US',
        status: 'stopped',
        cpu: 0,
        players: 0,
    },
]