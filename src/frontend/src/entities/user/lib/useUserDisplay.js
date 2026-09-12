import { reactive } from 'vue';
import { getUser } from '../api/userApi';

// Реактивный глобальный кэш информации о пользователях
const userCache = reactive({});
const pendingRequests = new Set();

/**
 * Получить отображаемое имя пользователя по его UUID/ID.
 * Если данных нет в кэше, запускает асинхронную загрузку и возвращает запасной вариант (обрезанный ID или прочерк).
 * После успешной загрузки реактивное свойство обновится и компонент перерендерится с реальным именем.
 */
export function getUserDisplayName(userId, fallback = '') {
  if (!userId) return fallback || '—';
  
  // Если это системный аккаунт
  if (userId === 'system' || userId === 'SYSTEM') {
    return 'Система';
  }

  const cached = userCache[userId];
  if (cached) {
    return cached.display_name || cached.email || userId;
  }

  // Запуск фонового запроса, если еще не запрашивается
  if (!pendingRequests.has(userId) && userId.length >= 8) {
    pendingRequests.add(userId);
    getUser(userId)
      .then((res) => {
        if (res && res.user) {
          userCache[userId] = res.user;
        } else if (res) {
          userCache[userId] = res;
        }
      })
      .catch(() => {
        // Запоминаем заглушку, чтобы не долбить сервер повторно
        userCache[userId] = { display_name: userId };
      })
      .finally(() => {
        pendingRequests.delete(userId);
      });
  }

  return fallback || (userId.length > 12 ? `${userId.slice(0, 8)}…` : userId);
}

/**
 * Асинхронно разрешить пользователя (с кэшированием)
 */
export async function resolveUser(userId) {
  if (!userId) return null;
  if (userCache[userId]) return userCache[userId];

  try {
    const res = await getUser(userId);
    const u = res?.user || res;
    if (u) {
      userCache[userId] = u;
      return u;
    }
  } catch (err) {
    console.warn(`Failed to resolve user ${userId}:`, err);
  }
  return null;
}

export function useUserDisplay() {
  return {
    userCache,
    getUserDisplayName,
    resolveUser,
  };
}
