import { http } from '@/shared/api';

export interface GameItem {
  id?: string | number;
  project_id: number;
  game_item_id: string;
  name: string;
  description?: string;
  image_url?: string;
  price_coins: number;
  is_active?: boolean;
  created_at?: string;
  updated_at?: string;
  [key: string]: any;
}

export const listGameItems = (projectId: string | number): Promise<GameItem[]> =>
  http.get(`/projects/${projectId}/purchases/items`).then((r) => r.data.items ?? []);

export const getGameItem = (projectId: string | number, itemId: string): Promise<GameItem> =>
  http.get(`/projects/${projectId}/purchases/items/${itemId}`).then((r) => r.data.item);

export const createGameItem = (projectId: string | number, payload: any): Promise<GameItem> =>
  http.post(`/projects/${projectId}/purchases/items`, {
    project_id: Number(projectId),
    game_item_id: payload.game_item_id,
    name: payload.name,
    description: payload.description || '',
    image_url: payload.image_url || '',
    price_coins: Number(payload.price_coins),
    is_active: payload.is_active !== false,
  }).then((r) => r.data.item);

export const updateGameItem = (projectId: string | number, itemId: string, payload: any): Promise<GameItem> =>
  http.put(`/projects/${projectId}/purchases/items/${itemId}`, {
    project_id: Number(projectId),
    game_item_id: itemId,
    name: payload.name,
    description: payload.description || '',
    image_url: payload.image_url || '',
    price_coins: Number(payload.price_coins),
    is_active: payload.is_active !== false,
  }).then((r) => r.data.item);

export const deleteGameItem = (projectId: string | number, itemId: string): Promise<any> =>
  http.delete(`/projects/${projectId}/purchases/items/${itemId}`).then((r) => r.data);

export const uploadItemImage = (projectId: string | number, itemId: string, file: File): Promise<any> => {
  const form = new FormData();
  form.append('media_type', `item:${itemId}`);
  form.append('file', file);
  return http
    .post(`/projects/${projectId}/media`, form, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    .then((r) => r.data);
};
