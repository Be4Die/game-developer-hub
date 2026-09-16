import { http } from '@/shared/api';

/**
 * Получить список всех внутриигровых товаров проекта
 * @param {string|number} projectId
 * @returns {Promise<Array>}
 */
export const listGameItems = (projectId) =>
  http.get(`/projects/${projectId}/purchases/items`).then((r) => r.data.items ?? []);

/**
 * Получить конкретный товар проекта
 * @param {string|number} projectId
 * @param {string} itemId
 * @returns {Promise<Object>}
 */
export const getGameItem = (projectId, itemId) =>
  http.get(`/projects/${projectId}/purchases/items/${itemId}`).then((r) => r.data.item);

/**
 * Создать внутриигровой товар
 * @param {string|number} projectId
 * @param {Object} payload
 * @returns {Promise<Object>}
 */
export const createGameItem = (projectId, payload) =>
  http.post(`/projects/${projectId}/purchases/items`, {
    project_id: Number(projectId),
    game_item_id: payload.game_item_id,
    name: payload.name,
    description: payload.description || '',
    image_url: payload.image_url || '',
    price_coins: Number(payload.price_coins),
    is_active: payload.is_active !== false,
  }).then((r) => r.data.item);

/**
 * Обновить существующий товар
 * @param {string|number} projectId
 * @param {string} itemId
 * @param {Object} payload
 * @returns {Promise<Object>}
 */
export const updateGameItem = (projectId, itemId, payload) =>
  http.put(`/projects/${projectId}/purchases/items/${itemId}`, {
    project_id: Number(projectId),
    game_item_id: itemId,
    name: payload.name,
    description: payload.description || '',
    image_url: payload.image_url || '',
    price_coins: Number(payload.price_coins),
    is_active: payload.is_active !== false,
  }).then((r) => r.data.item);

/**
 * Удалить/деактивировать товар
 * @param {string|number} projectId
 * @param {string} itemId
 * @returns {Promise<Object>}
 */
export const deleteGameItem = (projectId, itemId) =>
  http.delete(`/projects/${projectId}/purchases/items/${itemId}`).then((r) => r.data);

/**
 * Загрузить иконку для товара (сохраняется в games/{projectId}/items/{itemId}.png)
 * @param {string|number} projectId
 * @param {string} itemId
 * @param {File} file
 * @returns {Promise<Object>}
 */
export const uploadItemImage = (projectId, itemId, file) => {
  const form = new FormData();
  form.append('media_type', `item:${itemId}`);
  form.append('file', file);
  return http
    .post(`/projects/${projectId}/media`, form, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    .then((r) => r.data);
};
