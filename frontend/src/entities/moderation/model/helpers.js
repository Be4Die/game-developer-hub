import {
  REQUEST_STATUS,
  REQUEST_STATUS_LABELS,
  REQUEST_STATUS_BADGES,
} from './constants';

export function getStatusText(status) {
  return REQUEST_STATUS_LABELS[status] || 'Неизвестно';
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
      about: snapshot.about || '',
      iconPath: snapshot.icon_path || snapshot.iconPath || '',
      coverPath: snapshot.cover_path || snapshot.coverPath || '',
      videoPath: snapshot.video_path || snapshot.videoPath || '',
      activeBuildVersion: snapshot.active_build_version || snapshot.activeBuildVersion || '',
      devUrl: snapshot.dev_url || snapshot.devUrl || '',
    },
  };
}
