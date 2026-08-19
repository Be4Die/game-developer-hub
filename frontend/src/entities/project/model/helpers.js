export function statusClass(status) {
  if (status === 3) return 'status-published';
  if (status === 2) return 'status-pending';
  if (status === 4) return 'status-rejected';
  return 'status-draft';
}

export function statusLabel(status) {
  const map = {
    1: 'Черновик',
    2: 'На модерации',
    3: 'Опубликована',
    4: 'Отклонена',
  };
  return map[status] || 'Черновик';
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
