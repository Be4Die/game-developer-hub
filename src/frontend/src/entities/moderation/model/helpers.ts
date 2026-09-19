import { i18n } from '@/shared/lib';
import type { ModerationRequest, ChatMessage } from '@/shared/types';
import {
  REQUEST_STATUS,
  REQUEST_STATUS_LABELS,
  REQUEST_STATUS_BADGES,
  REQUEST_TYPE,
} from './constants';

export function getStatusText(status?: string | number | null): string {
  const t = (i18n.global as any).t;
  if (!status) return t('common.unknown');
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

export function getStatusBadgeClass(status?: string | number | null): string {
  if (!status) return 'badge-neutral';
  return REQUEST_STATUS_BADGES[status] || 'badge-neutral';
}

export function formatDateTime(isoOrTs?: any): string {
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

export function parseSenderRole(role?: string | number | null): number {
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

export function parseMessageType(type?: string | number | null): number {
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

export function determineDialogState(messagesOrLastMsg: any): 'none' | 'resolved' | 'unanswered' | 'in_dialog' {
  let lastMsg: any;
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

export function normalizeRequest(req?: any): ModerationRequest | null {
  if (!req) return null;
  const snapshot = req.snapshot || {};
  return {
    id: Number(req.id),
    projectId: Number(req.project_id || req.projectId),
    project_id: Number(req.project_id || req.projectId),
    ownerId: req.owner_id || req.ownerId,
    owner_id: req.owner_id || req.ownerId,
    moderatorId: req.moderator_id || req.moderatorId,
    moderator_id: req.moderator_id || req.moderatorId,
    status: req.status,
    type: req.type ?? 1,
    reason: req.reason || '',
    maxInstances: req.max_instances ?? req.maxInstances ?? 0,
    max_instances: req.max_instances ?? req.maxInstances ?? 0,
    maxTotalCpuMillis: req.max_total_cpu_millis ?? req.maxTotalCpuMillis ?? 0,
    max_total_cpu_millis: req.max_total_cpu_millis ?? req.maxTotalCpuMillis ?? 0,
    maxTotalMemoryMb: req.max_total_memory_mb ?? req.maxTotalMemoryMb ?? 0,
    max_total_memory_mb: req.max_total_memory_mb ?? req.maxTotalMemoryMb ?? 0,
    maxInstanceCpuMillis: req.max_instance_cpu_millis ?? req.maxInstanceCpuMillis ?? 0,
    max_instance_cpu_millis: req.max_instance_cpu_millis ?? req.maxInstanceCpuMillis ?? 0,
    maxInstanceMemoryMb: req.max_instance_memory_mb ?? req.maxInstanceMemoryMb ?? 0,
    max_instance_memory_mb: req.max_instance_memory_mb ?? req.maxInstanceMemoryMb ?? 0,
    moderatorComment: req.moderator_comment || req.moderatorComment || '',
    moderator_comment: req.moderator_comment || req.moderatorComment || '',
    rejectionReason: req.rejection_reason || req.rejectionReason || '',
    rejection_reason: req.rejection_reason || req.rejectionReason || '',
    submittedAt: req.submitted_at || req.submittedAt,
    submitted_at: req.submitted_at || req.submittedAt,
    reviewedAt: req.reviewed_at || req.reviewedAt,
    startedReviewAt: req.started_review_at || req.startedReviewAt,
    resolvedAt: req.resolved_at || req.resolvedAt,
    resolved_at: req.resolved_at || req.resolvedAt,
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
      isOnline: Boolean(snapshot.is_online ?? snapshot.isOnline),
    },
  };
}

export function formatCpu(millis?: number | null, compact = false): string {
  if (!millis || millis <= 0) return compact ? 'не огр.' : 'Без ограничений';
  const cores = millis / 1000;
  return `${cores % 1 === 0 ? cores : cores.toFixed(1)} CPU`;
}

export function formatMemory(mb?: number | null, compact = false): string {
  if (!mb || mb <= 0) return compact ? 'не огр.' : 'Без ограничений';
  if (mb >= 1024 && mb % 1024 === 0) {
    return `${mb / 1024} ГБ RAM`;
  }
  return `${mb} МБ RAM`;
}

export function getTypeText(type?: string | number | null): string {
  if (type === REQUEST_TYPE.SERVER_ACCESS || type === 'REQUEST_TYPE_SERVER_ACCESS' || type === 2) {
    return 'Серверы';
  }
  return 'Публикация';
}

export function getTypeBadgeClass(type?: string | number | null): string {
  if (type === REQUEST_TYPE.SERVER_ACCESS || type === 'REQUEST_TYPE_SERVER_ACCESS' || type === 2) {
    return 'badge-warning';
  }
  return 'badge-primary';
}

export function formatDurationSeconds(sec?: number | null): string {
  if (!sec || sec <= 0) return '0 мин';
  const totalMin = Math.round(sec / 60);
  if (totalMin < 1) return '< 1 мин';
  if (totalMin < 60) return `${totalMin} мин`;
  const hours = Math.floor(totalMin / 60);
  const remMinutes = totalMin % 60;
  if (remMinutes === 0) return `${hours} ч`;
  return `${hours} ч ${remMinutes} мин`;
}

export function getRequestDuration(req?: Partial<ModerationRequest> | null): string {
  if (!req) return '—';
  const started = req.started_review_at || req.startedReviewAt;
  const resolved = req.resolved_at || req.resolvedAt;
  if (!started || !resolved) return '—';
  const diffSec = Math.round((new Date(resolved).getTime() - new Date(started).getTime()) / 1000);
  if (diffSec <= 0) return '< 1 мин';
  return formatDurationSeconds(diffSec);
}
