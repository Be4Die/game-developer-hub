import axios from "axios";

const api = axios.create({
  baseURL: "/api/v1",
});

// Добавляем перехватчик для вставки токена авторизации
api.interceptors.request.use((config) => {
  const token = localStorage.getItem("gdh_access_token");
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

export const createChat = async (data) => {
  const response = await api.post("/chats", data);
  return response.data;
};

export const getChat = async (chatId) => {
  const response = await api.get(`/chats/${chatId}`);
  return response.data;
};

export const listChats = async (params) => {
  const response = await api.get("/chats", { params });
  return response.data;
};

export const sendMessage = async (chatId, data) => {
  const response = await api.post(`/chats/${chatId}/messages`, data);
  return response.data;
};

export const getMessages = async (chatId, params) => {
  const response = await api.get(`/chats/${chatId}/messages`, { params });
  return response.data;
};

export const addParticipant = async (chatId, data) => {
  const response = await api.post(`/chats/${chatId}/participants`, data);
  return response.data;
};

export const removeParticipant = async (chatId, userId) => {
  const response = await api.delete(`/chats/${chatId}/participants/${userId}`);
  return response.data;
};
