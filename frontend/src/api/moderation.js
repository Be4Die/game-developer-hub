import axios from "axios";

const http = axios.create({ baseURL: "/api/v1" });

http.interceptors.request.use((config) => {
  const token = localStorage.getItem("gdh_access_token");
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

export const moderationApi = {
  async submitForReview(gameId) {
    const res = await http.post(`/projects/${gameId}/moderation`);
    return { moderation: res.data.ticket };
  },

  async listPending(limit = 50, offset = 0) {
    const res = await http.get(`/moderation/tickets?status=1&limit=${limit}&offset=${offset}`);
    return {
      moderations: res.data.tickets ?? [],
    };
  },

  async getStatus(gameId) {
    try {
      const res = await http.get(`/moderation/tickets/by-project/${gameId}`);
      return { moderation: res.data.ticket };
    } catch {
      return null;
    }
  },

  async approve(gameId, comment = "Одобрено") {
    const res = await http.post(`/moderation/tickets/${gameId}/approve`, { comment });
    return { release: res.data.release };
  },

  async reject(gameId, reason) {
    const res = await http.post(`/moderation/tickets/${gameId}/reject`, { reason });
    return { ticket: res.data.ticket };
  },
};

export function moderationToTicket(m) {
  if (!m) return null;
  const statusMap = {
    MODERATION_STATUS_PENDING: "pending",
    MODERATION_STATUS_APPROVED: "approved",
    MODERATION_STATUS_REJECTED: "rejected",
    1: "pending",
    2: "approved",
    3: "rejected",
    pending: "pending",
    approved: "approved",
    rejected: "rejected",
  };
  const submittedAt = m.submitted_at || m.submittedAt;
  let created = "";
  if (submittedAt) {
    const ts = submittedAt.seconds
      ? submittedAt.seconds * 1000
      : typeof submittedAt === "number"
      ? (submittedAt < 10000000000 ? submittedAt * 1000 : submittedAt)
      : new Date(submittedAt).getTime();
    created = new Date(ts).toLocaleDateString("ru-RU");
  }
  return {
    id: Number(m.project_id ?? m.projectId ?? m.game_id ?? m.gameId ?? m.id),
    ticketId: Number(m.id),
    title: m.game_title || m.gameTitle || m.game_name || m.gameName || `Проект #${m.project_id ?? m.id}`,
    description: m.game_description || m.gameDescription || "",
    status: statusMap[m.status] || "pending",
    priority: "Средний",
    created,
    developerId: m.owner_id || m.ownerId || m.developer_id || m.developerId,
    rejectionReason: m.rejection_reason || m.rejectionReason || "",
    devUrl: m.dev_url || m.devUrl || "",
    activeBuildVersion: m.active_build_version || m.activeBuildVersion || "",
    moderationStatus: m.status,
  };
}

export function ticketStatusText(status) {
  const map = {
    pending: "На модерации",
    approved: "Одобрено",
    rejected: "Отклонено",
    new: "Новый",
    in_progress: "В работе",
    resolved: "Решён",
  };
  return map[status] || status;
}
