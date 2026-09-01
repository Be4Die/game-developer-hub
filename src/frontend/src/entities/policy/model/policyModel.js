export const modeLabels = {
  ORCHESTRATION_MODE_UNSPECIFIED: 'Не задан',
  ORCHESTRATION_MODE_DISABLED: 'Только ручное управление',
  ORCHESTRATION_MODE_KEEP_ALIVE: 'Держать запущенным',
  ORCHESTRATION_MODE_SCALE_TO_ZERO: 'Экономичный (scale-to-zero)',
};

export const behaviorLabels = {
  SCALE_BEHAVIOR_UNSPECIFIED: 'Не задано',
  SCALE_BEHAVIOR_SPAWN: 'Запускать новый инстанс',
  SCALE_BEHAVIOR_QUEUE: 'Очередь игроков',
};

export const queueLocationLabels = {
  QUEUE_LOCATION_UNSPECIFIED: 'Не задано',
  QUEUE_LOCATION_CLIENT: 'На стороне клиента',
  QUEUE_LOCATION_SERVER: 'На стороне сервера',
};

export function nodePreferenceLabel(pref) {
  if (!pref || pref === 'auto' || pref === '') return 'Авто (наименее загруженная)';
  return `Конкретная нода: ${pref}`;
}
