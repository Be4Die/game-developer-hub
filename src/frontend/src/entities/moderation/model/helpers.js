import { i18n } from '@/shared/lib';
import { REQUEST_STATUS, REQUEST_STATUS_LABELS, REQUEST_STATUS_BADGES } from './constants';

export function getStatusText(status) {
  const t = i18n.global.t;
  if (
    status === REQUEST_STATUS.PENDING ||
    status === 'REQUEST_STATUS_PENDING' ||
    status === 'pending'
  ) {
    return t('projects.moderation');
  }
  if (
    status === REQUEST_STATUS.IN_REVIEW ||
    status === 'REQUEST_STATUS_IN_REVIEW' ||
    status === 'in_review'
  ) {
    return t('projects.moderation');
  }
  if (
    status === REQUEST_STATUS.APPROVED ||
    status === 'REQUEST_STATUS_APPROVED' ||
    status === 'approved'
  ) {
    return t('projects.approved');
  }
  if (
    status === REQUEST_STATUS.REJECTED ||
    status === 'REQUEST_STATUS_REJECTED' ||
    status === 'rejected'
  ) {
    return t('projects.rejected');
  }
  return REQUEST_STATUS_LABELS[status] || t('common.unknown');
}

export function getStatusBadgeClass(status) {
  return REQUEST_STATUS_BADGES[status] || 'badge-neutral';
}

export function formatDateTime(isoOrTs) {
  if (!isoOrTs) return '—';
  if (typeof isoOrTs === 'object' && isoOrTs.seconds) {
    return new Date(isoOrTs.seconds * 1000).toLocaleString('ru-RU');
  }
  const date = new Date(isoOrTs);
  if (isNaN(date.getTime())) return String(isoOrTs);
  return date.toLocaleString('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

export function parseSenderRole(role) {
  if (role === undefined || role === null) return 1;
  if (typeof role === 'number') return role;
  const str = String(role).toUpperCase();
  if (
    str === 'SENDER_ROLE_DEVELOPER' ||
    str === 'DEVELOPER' ||
    str === 'USER_ROLE_USER' ||
    str === 'USER'
  )
    return 1;
  if (str === 'SENDER_ROLE_MODERATOR' || str === 'MODERATOR' || str === 'USER_ROLE_MODERATOR')
    return 2;
  if (str === 'SENDER_ROLE_SYSTEM' || str === 'SYSTEM') return 3;
  const num = parseInt(role, 10);
  return isNaN(num) ? 1 : num;
}

export function parseMessageType(type) {
  if (type === undefined || type === null) return 1;
  if (typeof type === 'number') return type;
  const str = String(type).toUpperCase();
  if (str === 'MESSAGE_TYPE_TEXT' || str === 'TEXT') return 1;
  if (str === 'MESSAGE_TYPE_SUBMITTED' || str === 'SUBMITTED') return 2;
  if (str === 'MESSAGE_TYPE_STATUS_CHANGED' || str === 'STATUS_CHANGED') return 3;
  if (str === 'MESSAGE_TYPE_APPROVED' || str === 'APPROVED') return 4;
  if (str === 'MESSAGE_TYPE_REJECTED' || str === 'REJECTED') return 5;
  const num = parseInt(type, 10);
  return isNaN(num) ? 1 : num;
}

export function determineDialogState(messagesOrLastMsg) {
  let lastMsg;
  if (Array.isArray(messagesOrLastMsg)) {
    if (messagesOrLastMsg.length === 0) {
      return 'none';
    }
    lastMsg = messagesOrLastMsg[messagesOrLastMsg.length - 1];
  } else {
    lastMsg = messagesOrLastMsg;
  }

  if (!lastMsg || (!lastMsg.content && !lastMsg.id)) {
    return 'none';
  }

  const role = parseSenderRole(lastMsg.sender_role ?? lastMsg.senderRole);
  const msgType = parseMessageType(lastMsg.message_type ?? lastMsg.messageType);
  const content = (lastMsg.content || '').toLowerCase();

  // Если вынесено решение или диалог явно закрыт
  if (
    msgType === 3 || // status_changed
    msgType === 4 || // approved
    msgType === 5 || // rejected
    content.includes('закрыл диалог') ||
    content.includes('вопрос решён') ||
    content.includes('все вопросы решены') ||
    content.includes('одобрен') ||
    content.includes('отклонен')
  ) {
    return 'resolved';
  }

  // Если последнее сообщение от разработчика -> ожидает ответа
  if (role === 1) {
    return 'unanswered';
  }

  // Если последнее сообщение от модератора -> в диалоге
  if (role === 2) {
    return 'in_dialog';
  }

  return 'resolved';
}

export function normalizeRequest(req) {
  if (!req) return null;
  const snapshot = req.snapshot || {};
  return {
    id: Number(req.id),
    projectId: Number(req.project_id || req.projectId),
    ownerId: req.owner_id || req.ownerId,
    moderatorId: req.moderator_id || req.moderatorId,
    status: req.status,
    rejectionReason: req.rejection_reason || req.rejectionReason || '',
    submittedAt: req.submitted_at || req.submittedAt,
    reviewedAt: req.reviewed_at || req.reviewedAt,
    snapshot: {
      projectId: Number(snapshot.project_id || snapshot.projectId || req.project_id),
      titleRu: snapshot.title_ru || snapshot.titleRu || '',
      titleEn: snapshot.title_en || snapshot.titleEn || '',
      seoRu: snapshot.seo_ru || snapshot.seoRu || '',
      seoEn: snapshot.seo_en || snapshot.seoEn || '',
      aboutRu: snapshot.about_ru || snapshot.aboutRu || snapshot.about || '',
      aboutEn: snapshot.about_en || snapshot.aboutEn || '',
      about: snapshot.about_ru || snapshot.aboutRu || snapshot.about || '',
      iconPath: snapshot.icon_path || snapshot.iconPath || '',
      coverPath: snapshot.cover_path || snapshot.coverPath || '',
      videoPath: snapshot.video_path || snapshot.videoPath || '',
      activeBuildVersion: snapshot.active_build_version || snapshot.activeBuildVersion || '',
      devUrl: snapshot.dev_url || snapshot.devUrl || '',
    },
  };
}
