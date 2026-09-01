import { i18n } from '@/shared/lib';

export function normalizeProjectStatus(status) {
  if (typeof status === 'number') return status;
  if (!status) return 1;
  const s = String(status).toUpperCase();
  if (s.includes('PUBLISHED') || s === '3') return 3;
  if (s.includes('PENDING') || s.includes('MODERATION') || s === '2') return 2;
  if (s.includes('APPROVED') || s === '4') return 4;
  if (s.includes('REJECTED') || s === '5') return 5;
  return 1;
}

export function statusClass(status) {
  const norm = normalizeProjectStatus(status);
  if (norm === 3) return 'status-published';
  if (norm === 2) return 'status-pending';
  if (norm === 4) return 'status-approved';
  if (norm === 5) return 'status-rejected';
  return 'status-draft';
}

export function statusLabel(status) {
  const t = i18n.global.t;
  const norm = normalizeProjectStatus(status);
  const map = {
    1: t('projects.draft'),
    2: t('projects.moderation'),
    3: t('projects.published'),
    4: t('projects.approved'),
    5: t('projects.rejected'),
  };
  return map[norm] || t('projects.draft');
}

export function getMediaUrl(path) {
  if (!path) return '';
  if (
    path.startsWith('http://') ||
    path.startsWith('https://') ||
    path.startsWith('blob:') ||
    path.startsWith('data:')
  ) {
    return path;
  }
  if (path.startsWith('/api/')) {
    return path;
  }
  const clean = path
    .replace(/^(\.\/|\/)?(data\/projects\/|projects\/)?/, '')
    .replace(/^media\//, '');
  return `/api/v1/media/${clean}`;
}
