import { i18n } from '@/shared/lib';

export function statusClass(status) {
  if (status === 3) return 'status-published';
  if (status === 2) return 'status-pending';
  if (status === 4) return 'status-rejected';
  return 'status-draft';
}

export function statusLabel(status) {
  const t = i18n.global.t;
  const map = {
    1: t('projects.created'),
    2: t('projects.moderation'),
    3: t('projects.published'),
    4: t('projects.rejected'),
  };
  return map[status] || t('projects.created');
}


export function getMediaUrl(path) {
  if (!path) return '';
  if (
    path.startsWith('http://') ||
    path.startsWith('https://') ||
    path.startsWith('/api/')
  ) {
    return path;
  }
  return `/api/v1/media/${path}`;
}
