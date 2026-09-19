export * from './model/rulesData';
export * from './model/docsData';

/**
 * Форматирование ISO-даты в человекочитаемый русский/английский формат.
 */
export function formatDocDate(isoDateStr?: string | null): string {
  if (!isoDateStr) return 'Актуально';
  try {
    const d = new Date(isoDateStr);
    if (isNaN(d.getTime())) return isoDateStr;
    return d.toLocaleDateString('ru-RU', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  } catch {
    return isoDateStr;
  }
}
