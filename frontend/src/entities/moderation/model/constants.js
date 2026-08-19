export const REQUEST_STATUS = {
  UNSPECIFIED: 0,
  PENDING: 1,
  IN_REVIEW: 2,
  APPROVED: 3,
  REJECTED: 4,
  CANCELLED: 5,
};

export const REQUEST_STATUS_LABELS = {
  [REQUEST_STATUS.UNSPECIFIED]: 'Не указан',
  [REQUEST_STATUS.PENDING]: 'Ожидает проверки',
  [REQUEST_STATUS.IN_REVIEW]: 'В проверке',
  [REQUEST_STATUS.APPROVED]: 'Одобрено',
  [REQUEST_STATUS.REJECTED]: 'Отклонено',
  [REQUEST_STATUS.CANCELLED]: 'Отозвано',
  REQUEST_STATUS_PENDING: 'Ожидает проверки',
  REQUEST_STATUS_IN_REVIEW: 'В проверке',
  REQUEST_STATUS_APPROVED: 'Одобрено',
  REQUEST_STATUS_REJECTED: 'Отклонено',
  REQUEST_STATUS_CANCELLED: 'Отозвано',
};

export const REQUEST_STATUS_BADGES = {
  [REQUEST_STATUS.UNSPECIFIED]: 'badge-neutral',
  [REQUEST_STATUS.PENDING]: 'badge-warning',
  [REQUEST_STATUS.IN_REVIEW]: 'badge-info',
  [REQUEST_STATUS.APPROVED]: 'badge-success',
  [REQUEST_STATUS.REJECTED]: 'badge-danger',
  [REQUEST_STATUS.CANCELLED]: 'badge-neutral',
  REQUEST_STATUS_PENDING: 'badge-warning',
  REQUEST_STATUS_IN_REVIEW: 'badge-info',
  REQUEST_STATUS_APPROVED: 'badge-success',
  REQUEST_STATUS_REJECTED: 'badge-danger',
  REQUEST_STATUS_CANCELLED: 'badge-neutral',
};
