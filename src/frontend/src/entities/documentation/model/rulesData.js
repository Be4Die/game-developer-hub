/**
 * Реестр правил платформы и регламента модерации Game Developer Hub.
 * Является единым источником истины (Single Source of Truth) для модерации и документации.
 */

export const RULE_CATEGORIES = {
  SECURITY: {
    id: 'security',
    title: 'Безопасность и Сеть',
    badgeClass: 'badge-security',
    icon: 'Shield',
  },
  SERVERS: {
    id: 'servers',
    title: 'Серверы и Оркестрация',
    badgeClass: 'badge-servers',
    icon: 'Server',
  },
  BUILDS: {
    id: 'builds',
    title: 'Стабильность и Качество билда',
    badgeClass: 'badge-builds',
    icon: 'Cpu',
  },
  CONTENT: {
    id: 'content',
    title: 'Контент и Политика платформы',
    badgeClass: 'badge-content',
    icon: 'FileText',
  },
};

export const PLATFORM_RULES = [
  // ─── 1. БЕЗОПАСНОСТЬ И СЕТЬ ─────────────────────────────────────────
  {
    code: 'SEC-01',
    category: RULE_CATEGORIES.SECURITY.id,
    categoryTitle: RULE_CATEGORIES.SECURITY.title,
    title: 'Вредоносный код и скрытые процессы',
    severity: 'critical',
    lastUpdated: '2026-08-15',
    summary: 'Запрещено внедрение троянов, кейлоггеров, криптомайнеров или скрытых фоновых служб.',
    description: `Исполняемые файлы клиента и сервера не должны содержать вредоносного ПО, обфусцированных дропперов, скрытых процессов или утилит удаленного управления без явного согласия пользователя. Все исполняемые бинарники сканируются антивирусными пайплайнами платформы.`,
    howToFix: `Проведите аудит зависимостей и сторонних плагинов движка. Убедитесь, что билд не запускает сторонние .exe/.dll/.so в фоновом режиме. Пересоберите проект из чистого репозитория.`,
  },
  {
    code: 'SEC-02',
    category: RULE_CATEGORIES.SECURITY.id,
    categoryTitle: RULE_CATEGORIES.SECURITY.title,
    title: 'Несанкционированный сбор данных и телеметрия',
    severity: 'critical',
    lastUpdated: '2026-08-15',
    summary: 'Сбор персональных данных без согласия пользователя или отправка незашифрованной телеметрии запрещены.',
    description: `Клиент игры не имеет права собирать конфиденциальные данные пользователя (файлы с диска, историю браузера, токены авторизации, системные пароли). Любая аналитика обязана соответствовать регламенту GDPR/152-ФЗ и требовать предварительного согласия игрока.`,
    howToFix: `Удалите сторонние несертифицированные трекеры. Добавьте в настройки игры переключатель согласия на сбор анонимной аналитики (Opt-In / Opt-Out).`,
  },
  {
    code: 'SEC-03',
    category: RULE_CATEGORIES.SECURITY.id,
    categoryTitle: RULE_CATEGORIES.SECURITY.title,
    title: 'Отсутствие шифрования трафика (TLS/HTTPS/WSS)',
    severity: 'warning',
    lastUpdated: '2026-09-01',
    summary: 'Все сетевые запросы клиента и веб-сокетов обязаны использовать защищенные протоколы (HTTPS/WSS).',
    description: `Любое сетевое взаимодействие игрового клиента со сторонними бэкендами или серверами авторизации должно выполняться по зашифрованным каналам (HTTPS, TLS 1.3, WSS, DTLS). Передача токенов сессий и паролей по незащищенному HTTP строго запрещена.`,
    howToFix: `Замените http:// на https:// и ws:// на wss:// во всех эндпоинтах сетевого кода вашей игры. Настройте доверенные SSL/TLS сертификаты на ваших внешних серверах.`,
  },
  {
    code: 'SEC-04',
    category: RULE_CATEGORIES.SECURITY.id,
    categoryTitle: RULE_CATEGORIES.SECURITY.title,
    title: 'Несанкционированные сетевые запросы',
    severity: 'warning',
    lastUpdated: '2026-08-20',
    summary: 'Клиент игры отправляет скрытые запросы к незадекларированным внешним хостам.',
    description: `При запуске и во время игрового процесса клиент не должен обращаться к подозрительным внешним хостам, IP-адресам или P2P-сетям, не указанным в проектной документации. Допустимы только запросы к GDH SDK API и задекларированным мастер-серверам разработчика.`,
    howToFix: `Задекларируйте внешние домены в манифесте проекта либо перенесите внешнюю бизнес-логику на Dedicated Server под управлением платформы.`,
  },

  // ─── 2. СЕРВЕРЫ И ОРКЕСТРАЦИЯ ───────────────────────────────────────
  {
    code: 'SRV-01',
    category: RULE_CATEGORIES.SERVERS.id,
    categoryTitle: RULE_CATEGORIES.SERVERS.title,
    title: 'Превышение выделенных квот CPU и оперативной памяти',
    severity: 'critical',
    lastUpdated: '2026-09-05',
    summary: 'Игровой сервер не укладывается в утвержденный лимит ресурсов ноды или допускает утечки памяти.',
    description: `Каждый инстанс игрового сервера выделяется с жесткими лимитами CPU (millis) и RAM (MB). Сервер, потребляющий память с постоянным ростом (Memory Leak) или перегружающий CPU на 100% в режиме ожидания, будет принудительно остановлен OOM Killer оркестратора.`,
    howToFix: `Проведите профилирование сервера с помощью Memory Profiler. Устраните утечки при завершении матча и уничтожении игровых комнат. При необходимости запросите расширение квот платформы.`,
  },
  {
    code: 'SRV-02',
    category: RULE_CATEGORIES.SERVERS.id,
    categoryTitle: RULE_CATEGORIES.SERVERS.title,
    title: 'Отсутствие обработки сигналов остановки (Graceful Shutdown)',
    severity: 'critical',
    lastUpdated: '2026-09-10',
    summary: 'Сервер обязан корректно перехватывать SIGTERM/SIGINT и корректно освобождать сетевой порт в течение 15 секунд.',
    description: `При масштабировании или плановом перезапуске ноды агент отправляет процессу сигнал SIGTERM. Сервер должен уведомить игроков о завершении матча, сохранить прогресс через GDH Cloud Storage API и завершить процесс с exit code 0.`,
    howToFix: `Добавьте перехватчик сигналов POSIX (SIGTERM, SIGINT) в главном цикле сервера. Интегрируйте метод GDH Server SDK: \`GDH.Server.OnShutdownRequested(() => { SaveAndClose(); })\`.`,
  },
  {
    code: 'SRV-03',
    category: RULE_CATEGORIES.SERVERS.id,
    categoryTitle: RULE_CATEGORIES.SERVERS.title,
    title: 'Хардкод сетевых портов вместо переменной окружения $PORT',
    severity: 'critical',
    lastUpdated: '2026-09-01',
    summary: 'Игровой сервер обязан считывать порт для входящих UDP/TCP соединений из переменной окружения PORT.',
    description: `Оркестратор динамически назначает порт для каждого инстанса через env-переменную PORT (или аргумент запуска --port). Если сервер пытается привязаться к фиксированному порту (например, 7777), запуск на ноде с несколькими инстансами упадет с ошибкой 'Address already in use'.`,
    howToFix: `Используйте: \`int port = int.Parse(Environment.GetEnvironmentVariable("PORT") ?? "7777");\` и передавайте его в сетевой сокет вашего движка (Mirror, FishNet, Unreal NetDriver, Photon).`,
  },
  {
    code: 'SRV-04',
    category: RULE_CATEGORIES.SERVERS.id,
    categoryTitle: RULE_CATEGORIES.SERVERS.title,
    title: 'Отсутствие Health Check и зависание тикрейта',
    severity: 'warning',
    lastUpdated: '2026-08-25',
    summary: 'Сервер обязан поддерживать стабильный Tick Rate и отвечать на внутренние пинги статуса оркестратора.',
    description: `Если главный поток игрового сервера блокируется тяжелыми синхронными операциями ввода-вывода или бесконечными циклами на время > 10 секунд, агент оркестратора признает инстанс 'UNHEALTHY' и инициирует его перезапуск.`,
    howToFix: `Выносите загрузку больших файлов, вызовы внешних баз данных и генерацию мира в асинхронные потоки (Async / Coroutines). Не блокируйте главный поток сервера (Main Thread).`,
  },
  {
    code: 'SRV-05',
    category: RULE_CATEGORIES.SERVERS.id,
    categoryTitle: RULE_CATEGORIES.SERVERS.title,
    title: 'Некорректная сборка Docker-образа / бинарника для Linux',
    severity: 'warning',
    lastUpdated: '2026-09-02',
    summary: 'Dedicated Server должен быть скомпилирован под Linux x86_64 в Headless-режиме (без GUI и графических библиотек).',
    description: `Ноды оркестрации работают под управлением Linux. Серверный билд не должен требовать X11, Display Server или графического драйвера GPU, если для игры не задекларирован GPU-рендеринг.`,
    howToFix: `Экспортируйте билд как 'Dedicated Server (Linux / Headless)'. В Unity включите 'Server Build / Dedicated Server', в Unreal Engine выберите Target Type: 'Server'.`,
  },

  // ─── 3. СТАБИЛЬНОСТЬ И БИЛДЫ ─────────────────────────────────────────
  {
    code: 'BLD-01',
    category: RULE_CATEGORIES.BUILDS.id,
    categoryTitle: RULE_CATEGORIES.BUILDS.title,
    title: 'Критическая ошибка запуска (Crash on Start)',
    severity: 'critical',
    lastUpdated: '2026-08-10',
    summary: 'Клиент или сервер игры аварийно завершается при запуске на целевой конфигурации.',
    description: `Билд игры падает с критической ошибкой (Segmentation Fault, NullReferenceException, Fatal Error) на этапе инициализации или загрузки стартового экрана.`,
    howToFix: `Проверьте логи инициализации. Убедитесь, что все требуемые зависимости (DirectX, Visual C++ Redistributables, OpenGL/Vulkan библиотеки) упакованы в билд или обрабатываются корректно.`,
  },
  {
    code: 'BLD-02',
    category: RULE_CATEGORIES.BUILDS.id,
    categoryTitle: RULE_CATEGORIES.BUILDS.title,
    title: 'Блокирующие баги прохождения (Game Breaking Bugs)',
    severity: 'critical',
    lastUpdated: '2026-08-10',
    summary: 'В игре обнаружены ошибки, препятствующие завершению основного сценария, обучения или сетевой сессии.',
    description: `Игрок попадает в тупиковое состояние без возможности продолжить игру (застревание геометрии, невозможность нажать интерактивный триггер, непрогружающийся уровень).`,
    howToFix: `Воспроизведите баг по шагам, указанным в отчете модератора, и исправьте логику квеста/коллизий. Загрузите исправленную версию билда.`,
  },
  {
    code: 'BLD-03',
    category: RULE_CATEGORIES.BUILDS.id,
    categoryTitle: RULE_CATEGORIES.BUILDS.title,
    title: 'Неработоспособное или перекрывающее управление',
    severity: 'warning',
    lastUpdated: '2026-08-15',
    summary: 'Элементы интерфейса перекрывают игровое поле, либо управление клавиатурой/мышью/геймпадом не реагирует.',
    description: `Интерфейс не адаптируется под стандартные разрешения экранов (1080p, 1440p, 4K), кликабельные зоны кнопок смещены, либо отсутствует базовая реакция на ввод.`,
    howToFix: `Проверьте работу Canvas Scaler / Anchor points в вашем UI. Добавьте в настройки возможность переназначения клавиш.`,
  },
  {
    code: 'BLD-04',
    category: RULE_CATEGORIES.BUILDS.id,
    categoryTitle: RULE_CATEGORIES.BUILDS.title,
    title: 'Поврежденные или отсутствующие ресурсы (Missing Assets)',
    severity: 'warning',
    lastUpdated: '2026-08-18',
    summary: 'В сборке отображаются розовые шейдеры (Magenta), отсутствуют текстуры или звуковые дорожки.',
    description: `Наличие плейсхолдеров, отсутствующих текстур ("Missing Material"), незагруженных 3D-моделей или ошибок файловой системы при чтении ассетов.`,
    howToFix: `Проверьте ассет-бандлы и включение всех используемых материалов в билд-настройки движка. Пересоберите ресурсы проекта.`,
  },

  // ─── 4. КОНТЕНТ И ПОЛИТИКА ──────────────────────────────────────────
  {
    code: 'CNT-01',
    category: RULE_CATEGORIES.CONTENT.id,
    categoryTitle: RULE_CATEGORIES.CONTENT.title,
    title: 'Несоответствие заявленному возрастному рейтингу',
    severity: 'warning',
    lastUpdated: '2026-09-05',
    summary: 'В игре обнаружены сцены насилия, нецензурная лексика или контент, превышающий выбранный возрастной ценз (0+, 6+, 12+, 16+, 18+).',
    description: `Материалы игры (диалоги, текстуры, механики) должны строго соответствовать рейтингу, указанному разработчиком на странице черновика проекта.`,
    howToFix: `Скорректируйте возрастной рейтинг в карточке проекта на вкладке 'О проекте' либо удалите недопустимые для текущей категории сцены.`,
  },
  {
    code: 'CNT-02',
    category: RULE_CATEGORIES.CONTENT.id,
    categoryTitle: RULE_CATEGORIES.CONTENT.title,
    title: 'Вводящие в заблуждение промо-материалы (Misleading Metadata)',
    severity: 'warning',
    lastUpdated: '2026-08-30',
    summary: 'Скриншоты, видеоролики или описание игры не соответствуют реальному геймплею.',
    description: `Скриншоты в карточке игры обязаны отражать актуальный игровой процесс. Запрещено использовать рендеры других игр или заявлять механики, которых нет в билде.`,
    howToFix: `Замените скриншоты и обложку на актуальные снимки из реального игрового процесса последней версии сборки.`,
  },
  {
    code: 'CNT-03',
    category: RULE_CATEGORIES.CONTENT.id,
    categoryTitle: RULE_CATEGORIES.CONTENT.title,
    title: 'Нарушение авторских прав и чужая интеллектуальная собственность',
    severity: 'critical',
    lastUpdated: '2026-08-10',
    summary: 'Запрещено использование чужих торговых марок, музыки, брендов или ассетов без лицензии правообладателя.',
    description: `Использование чужих логотипов, музыки, персонажей известных франшиз без подтвержденного права на коммерческое или некоммерческое использование.`,
    howToFix: `Замените спорные материалы на собственные ассеты или материалы с лицензией Creative Commons / Commercial Royalty-Free. Предоставьте лицензионное соглашение в чат модерации.`,
  },
  {
    code: 'CNT-04',
    category: RULE_CATEGORIES.CONTENT.id,
    categoryTitle: RULE_CATEGORIES.CONTENT.title,
    title: 'Запрещенные темы и разжигание вражды',
    severity: 'critical',
    lastUpdated: '2026-08-10',
    summary: 'Проект содержит дискриминационные материалы, пропаганду запрещенных веществ или экстремистские символы.',
    description: `Платформа категорически запрещает публикацию игр, пропагандирующих экстремизм, насилие над реальными группами людей, терроризм, употребление наркотических веществ или иные нарушения законодательства.`,
    howToFix: `Удалите запрещенные материалы из игры и всех описаний карточки проекта.`,
  },
];

/**
 * Получить правило по его коду (например, "SEC-04" или "sec-04").
 */
export function getRuleByCode(code) {
  if (!code) return null;
  const clean = String(code).trim().toUpperCase();
  return PLATFORM_RULES.find((r) => r.code.toUpperCase() === clean) || null;
}

/**
 * Поиск по правилам платформы (по коду, заголовку, категории или ключевым словам).
 */
export function searchPlatformRules(query) {
  if (!query || !query.trim()) return PLATFORM_RULES;
  const q = query.trim().toLowerCase();
  return PLATFORM_RULES.filter(
    (r) =>
      r.code.toLowerCase().includes(q) ||
      r.title.toLowerCase().includes(q) ||
      r.summary.toLowerCase().includes(q) ||
      r.categoryTitle.toLowerCase().includes(q) ||
      (r.howToFix && r.howToFix.toLowerCase().includes(q))
  );
}

/**
 * Получить правила сгруппированные по категориям.
 */
export function getRulesByCategory() {
  const categories = Object.values(RULE_CATEGORIES).map((cat) => ({
    ...cat,
    rules: PLATFORM_RULES.filter((r) => r.category === cat.id),
  }));
  return categories;
}
