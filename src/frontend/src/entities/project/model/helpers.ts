import { i18n } from '@/shared/lib';
import type { Project } from '@/shared/types';

export function normalizeProjectStatus(status?: string | number | null): number {
  if (typeof status === 'number') return status;
  if (!status) return 1;
  const s = String(status).toUpperCase();
  if (s.includes('PUBLISHED') || s === '3') return 3;
  if (s.includes('PENDING') || s.includes('MODERATION') || s === '2') return 2;
  if (s.includes('APPROVED') || s === '4') return 4;
  if (s.includes('REJECTED') || s === '5') return 5;
  return 1;
}

export function statusClass(status?: string | number | null): string {
  const norm = normalizeProjectStatus(status);
  if (norm === 3) return 'status-published';
  if (norm === 2) return 'status-pending';
  if (norm === 4) return 'status-approved';
  if (norm === 5) return 'status-rejected';
  return 'status-draft';
}

export function statusLabel(status?: string | number | null): string {
  const t = (i18n.global as any).t;
  const norm = normalizeProjectStatus(status);
  const map: Record<number, string> = {
    1: t('projects.draft'),
    2: t('projects.moderation'),
    3: t('projects.published'),
    4: t('projects.approved'),
    5: t('projects.rejected'),
  };
  return map[norm] || t('projects.draft');
}

export function getMediaUrl(path?: string | null): string {
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
  if (path.startsWith('/media/')) {
    return path;
  }
  if (path.includes('projects/')) {
    const projKey = path.replace(/^(\.\/|\/)?(data\/)?/, '');
    return `/${projKey.startsWith('media/') ? projKey : 'media/' + projKey}`;
  }
  const clean = path
    .replace(/^(\.\/|\/)?(data\/projects\/|projects\/)?/, '')
    .replace(/^media\//, '');
  return `/media/projects/${clean}`;
}

export function hasPermission(project?: Partial<Project> | null, perm?: string): boolean {
  if (!project) return false;
  if (project.is_owner !== false) return true;
  if (!perm) return true;
  return (
    Array.isArray(project.current_user_permissions) &&
    project.current_user_permissions.includes(perm)
  );
}

export function permissionLabel(perm: string): string {
  const t = (i18n.global as any).t;
  const key = `access.permissions.${perm}`;
  const translated = t(key);
  return translated !== key ? translated : perm;
}
