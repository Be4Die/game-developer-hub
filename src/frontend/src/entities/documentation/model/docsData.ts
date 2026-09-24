/**
 * Структурированные статьи и разделы документации Game Developer Hub.
 * Поддерживает полноценную двухуровневую иерархию: Раздел (Root Section) -> Статья (Article).
 */

export interface DocCodeTab {
  label: string;
  language: string;
  code: string;
}

export interface DocArticle {
  id: string;
  title: string;
  summary: string;
  lastUpdated?: string;
  content?: string;
  codeLanguage?: string;
  codeSnippet?: string;
  codeTabs?: DocCodeTab[];
  isRulesCatalog?: boolean;
}

export interface DocSection {
  id: string;
  title: string;
  description: string;
  icon?: string;
  lastUpdated?: string;
  articles?: DocArticle[];
}

export const DOCS_SECTIONS: DocSection[] = [
  {
    id: 'getting-started',
    title: 'Быстрый старт',
    description: 'Основы работы с платформой, жизненный цикл проектов и первая публикация.',
    icon: 'Sparkles',
    lastUpdated: '2026-09-10',
    articles: [
      {
        id: 'overview',
        title: 'Обзор платформы GDH',
        summary: 'Назначение Game Developer Hub, архитектура и ключевые возможности.',
        lastUpdated: '2026-09-10',
        content: `**Game Developer Hub (GDH)** — это комплексная экосистема для разработчиков одиночных и мультиплеерных игр.

Платформа решает ключевые задачи оперирования играми:
- **Управление сборками и версионирование** — изолированное хранение черновиков и релизных версий клиентов и серверов.
- **Оркестрация Dedicated Servers** — динамическое выделение игровых серверов на пуле нод под нагрузкой.
- **Welwise Games SDK и Unity плагин** — готовые клиентские библиотеки авторизации, облачных сохранений, рекламы и покупок.
- **Интегрированная модерация** — прозрачный регламент проверки билдов с видео/фото-доказательствами в едином чате.`,
      },
      {
        id: 'lifecycle',
        title: 'Жизненный цикл проекта',
        summary: 'Этапы: Черновик, Песочница (Sandbox), Модерация и Каталог.',
        lastUpdated: '2026-09-10',
        content: `Каждый проект в GDH проходит через 4 последовательных этапа:

1. **Черновик (Draft)** — заполнение метаданных (название, описание, обложка, иконка, возрастной рейтинг), загрузка первичных файлов и настройка прав доступа команды.
2. **Песочница (Sandbox)** — изолированная тестовая среда разработчика. Здесь проверяется запуск клиента, эмулируются облачные сохранения и тестируется сетевой код.
3. **Модерация (Review)** — отправка билда на проверку. Модератор тестирует билд на стабильность, сетевую безопасность и соответствие регламенту. В случае замечаний формируется структурированный вердикт со ссылками на пункты правил.
4. **Релиз и Каталог (Catalog)** — игра становится доступна пользователям платформы, а серверные инстансы автоматически запускаются и масштабируются оркестратором.`,
      },
      {
        id: 'roles',
        title: 'Роли и права доступа',
        summary: 'Разработчики, модераторы, системные администраторы.',
        lastUpdated: '2026-09-08',
        content: `В платформе реализована ролевая модель доступа:

- **Разработчик (Developer)** — создание проектов, управление черновиками, загрузка сборок, управление серверами в рамках выделенных квот и просмотр аналитики.
- **Модератор (Moderator)** — рассмотрение заявок на публикацию и доступ к серверам, проведение проверок, ведение чата модерации и вынесение вердиктов со ссылками на регламент.
- **Администратор (Admin)** — управление физическими нодами оркестратора, квотами, каталогом пользователей и системными журналами.`,
      },
    ],
  },
  {
    id: 'orchestration',
    title: 'Оркестрация серверов',
    description: 'Управление выделенными серверами (Dedicated Servers), нодами и авто-масштабированием.',
    icon: 'Server',
    lastUpdated: '2026-09-15',
    articles: [
      {
        id: 'architecture',
        title: 'Архитектура оркестратора и нод',
        summary: 'Взаимодействие Orchestrator, Game Server Node и Game Deployment Agent.',
        lastUpdated: '2026-09-15',
        content: `Оркестратор GDH управляет пулом физических или виртуальных серверов (**Game Server Nodes**).

На каждой ноде развернут системный сервис **Game Deployment Agent**, выполняющий следующие функции:
- Получение команд на запуск/остановку инстансов от центрального оркестратора;
- Изоляция процессов и контроль лимитов ресурсов (CPU millis, RAM MB);
- Динамический маппинг UDP/TCP портов через Ingress Proxy;
- Сбор логов stdout/stderr и передача метрик нагрузки в реальном времени.`,
      },
      {
        id: 'server-requirements',
        title: 'Требования к Dedicated Server',
        summary: 'Linux Headless, динамический $PORT, потребление памяти и CPU.',
        lastUpdated: '2026-09-15',
        content: `Каждый инстанс сервера запускается в изолированном окружении. Порт выделяется оркестратором динамически. **Запрещено жестко прошивать порт в коде (hardcode)**.

Игровой сервер считывает конфигурацию через переменные окружения:`,
        codeLanguage: 'bash',
        codeSnippet: `# Переменные окружения, передаваемые оркестратором в игровой сервер
PORT=7778                  # Порт для входящих игровых UDP/TCP соединений
SERVER_ID=srv-prod-9481    # Уникальный идентификатор инстанса
GAME_ID=42                 # ID проекта в GDH
MAX_PLAYERS=16             # Максимальное количество игроков в сессии
TICK_RATE=60               # Целевой тикрейт сервера
GDH_API_URL=https://api.gdh.local
GDH_SERVER_SECRET=sec_...  # Токен авторизации инстанса`,
        codeTabs: [
          {
            label: 'C# (Unity / .NET)',
            language: 'csharp',
            code: `using System;

public class DedicatedServerBootstrap
{
    public static void StartServer()
    {
        // Считываем порт из переменной окружения PORT
        string portEnv = Environment.GetEnvironmentVariable("PORT");
        ushort port = !string.IsNullOrEmpty(portEnv) ? ushort.Parse(portEnv) : (ushort)7777;

        Console.WriteLine($"[GDH] Запуск игрового сервера на порту: {port}");
        
        // Передаем порт в сетевой транспорт (Mirror / FishNet / Netcode)
        NetworkManager.singleton.GetComponent<kcp2k.KcpTransport>().Port = port;
        NetworkManager.singleton.StartServer();
    }
}`,
          },
          {
            label: 'C++ (Unreal Engine)',
            language: 'cpp',
            code: `// В методе инициализации GameMode / Server Startup
FString PortStr = FPlatformMisc::GetEnvironmentVariable(TEXT("PORT"));
int32 Port = PortStr.IsEmpty() ? 7777 : FCString::Atoi(*PortStr);

UE_LOG(LogNet, Display, TEXT("[GDH] Запуск выделенного сервера UE на порту: %d"), Port);

// Запуск сокета с указанным портом
FURL URL;
URL.Port = Port;
GetWorld()->Listen(URL);`,
          },
          {
            label: 'Go (Headless)',
            language: 'go',
            code: `package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "7777"
	}

	addr, _ := net.ResolveUDPAddr("udp", ":"+port)
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	fmt.Printf("[GDH] Сервер успешно запущен на порту %s\n", port)
}`,
          },
        ],
      },
      {
        id: 'lifecycle-hooks',
        title: 'Graceful Shutdown и Health Checks',
        summary: 'Обработка сигналов SIGTERM, таймауты и протокол проверки доступности.',
        lastUpdated: '2026-09-14',
        content: `Когда матч завершается или нода отправляется на обслуживание, оркестратор посылает серверному процессу сигнал **SIGTERM**.

У сервера есть **15 секунд**, чтобы:
1. Уведомить подключенных клиентов о завершении сессии;
2. Отправить итоговый прогресс и результаты матча в GDH API;
3. Закрыть открытые сокеты и файлы;
4. Завершить процесс с кодом \`exit(0)\`.

Если процесс не завершается за 15 секунд, демон ноды отправляет **SIGKILL** (принудительное завершение). Нарушение этого правила фиксируется модерацией по коду **SRV-02**.`,
      },
      {
        id: 'packaging',
        title: 'Сборка и загрузка билда',
        summary: 'Упаковка в ZIP-архив и сборка Docker-контейнера для сервера.',
        lastUpdated: '2026-09-12',
        content: `Для контейнеризованных билдов используйте минимальный базовый образ Linux (Ubuntu / Debian-slim):`,
        codeLanguage: 'dockerfile',
        codeSnippet: `FROM ubuntu:22.04

# Установка минимальных зависимостей (CA-сертификаты, tzdata, glibc)
RUN apt-get update && apt-get install -y --no-install-recommends \\
    ca-certificates \\
    libstdc++6 \\
    libgcc-s1 \\
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Копируем скомпилированный билд сервера
COPY ./build/Server.x86_64 /app/Server.x86_64
COPY ./build/Server_Data /app/Server_Data

# Устанавливаем права на исполнение
RUN chmod +x /app/Server.x86_64

# Пользователь без root-прав для безопасности
USER 1000:1000

# Запуск в headless-режиме
ENTRYPOINT ["/app/Server.x86_64", "-batchmode", "-nographics"]`,
      },
    ],
  },
  {
    id: 'sdk',
    title: 'JavaScript SDK',
    description: 'Официальная клиентская библиотека для веб-игр: сохранения, метавселенная, реклама, платежи и время.',
    icon: 'Code2',
    lastUpdated: '2026-09-24',
    articles: [
      {
        id: "overview",
        title: "Обзор и установка JS SDK",
        summary: "Назначение библиотеки WelwiseGames SDK, скачивание и подключение в HTML5 проект.",
        lastUpdated: "2026-09-24",
        content: `**WelwiseGames SDK** предоставляет единый стандартизированный клиентский интерфейс для взаимодействия веб-игры с платформой Welwise Games.

### Возможности SDK:
- **Облачные сохранения** — надежное сохранение и синхронизация прогресса игрока на серверах платформы.
- **Данные метавселенной** — кросс-игровые сохранения, доступные между различными играми экосистемы Welwise.
- **Внутриигровые покупки** — каталог товаров и безопасная оплата за платформенную валюту Welwise Coins.
- **Реклама и монетизация** — вызовы полноэкранной межстраничной рекламы и видео с вознаграждением.
- **Параметры окружения** — получение идентификатора игрока (\`PlayerID\`), типа клиентского устройства (\`DeviceInfo\`) и языка (\`LanguageCode\`).
- **Синхронизация времени** — запрос точного серверного времени по стандарту ISO 8601 для защиты таймеров и кулдаунов.
- **Навигация** — перенаправление пользователя на другие игры платформы.

### Установка и подключение
1. Скачайте актуальный файл \`sdk.js\` с официального GitHub-репозитория: [https://github.com/Welwise-Games/JS_SDK](https://github.com/Welwise-Games/JS_SDK).
2. Разместите файл \`sdk.js\` в каталоге публичных статических ресурсов вашей веб-игры.
3. Подключите скрипт в секцию \`<head>\` файла \`index.html\`:

\`\`\`html
<script src="./sdk.js"></script>
\`\`\``,
      },
      {
        id: "initialization",
        title: "Инициализация SDK",
        summary: "Подключение к платформе с помощью метода WelwiseGames.init().",
        lastUpdated: "2026-09-24",
        content: `Для начала работы с SDK необходимо выполнить инициализацию с помощью метода \`init()\` глобального объекта \`WelwiseGames\`.

Метод устанавливает защищенное соединение с родительской платформой и возвращает рабочий экземпляр \`sdk\`, предоставляющий доступ ко всем методам платформенных сервисов.

### Использование с async/await:
\`\`\`javascript
try {
  const sdk = await WelwiseGames.init();
  console.log('SDK успешно инициализирован:', sdk);
  // Доступ к модулям: sdk.player, sdk.AdvManager, sdk.purchases...
} catch (err) {
  console.error('Ошибка инициализации SDK:', err);
}
\`\`\`

### Использование с Promise:
\`\`\`javascript
WelwiseGames.init()
  .then((sdk) => {
    console.log('SDK успешно инициализирован');
    // Дальнейший запуск игрового цикла
  })
  .catch((err) => {
    console.error('Ошибка инициализации SDK:', err);
  });
\`\`\``,
      },
      {
        id: "environment",
        title: "Переменные окружения и серверное время",
        summary: "Идентификатор игрока, тип устройства, язык интерфейса и синхронизация времени.",
        lastUpdated: "2026-09-24",
        content: `После успешной инициализации через свойство \`sdk.Environment\` становятся доступны параметры пользовательского окружения.

### Доступные свойства:
Все свойства возвращают значение типа \`string\`:

| Свойство | Описание | Пример значения |
| :--- | :--- | :--- |
| \`sdk.Environment.PlayerId\` | Уникальный строковый идентификатор игрока | \`d6bc2025-bed2-4d7b-9349-3f0cb56d53f0\` |
| \`sdk.Environment.DeviceType\` | Тип клиентского устройства: \`mobile\`, \`tablet\`, \`desktop\` | \`mobile\` |
| \`sdk.Environment.LanguageCode\` | Двухбуквенный код языка интерфейса | \`ru\`, \`en\` |

\`\`\`javascript
console.log('ID игрока:', sdk.Environment.PlayerId);
console.log('Устройство:', sdk.Environment.DeviceType);
console.log('Язык:', sdk.Environment.LanguageCode);
\`\`\`

### Серверное время
Метод \`sdk.serverTime()\` запрашивает и возвращает текущее точное время сервера в стандарте ISO 8601.

- **Параметры:** отсутствуют.
- **Возвращает:** \`Promise<string>\` — строку формата \`YYYY-MM-DDTHH:mm:ssZ\`.

\`\`\`javascript
// async/await
const serverTime = await sdk.serverTime();
console.log('Серверное время:', serverTime);

// then/catch
sdk.serverTime()
  .then((time) => console.log('Серверное время:', time))
  .catch((err) => console.error('Ошибка при получении времени:', err));
\`\`\``,
      },
      {
        id: "navigation",
        title: "Навигация по платформе",
        summary: "Перенаправление игрока на страницу другой игры в каталоге платформы.",
        lastUpdated: "2026-09-24",
        content: `Метод \`sdk.PlatformNavigation.goToGame(gameId)\` выполняет бесшовное перенаправление пользователя на страницу другой игры каталога Welwise Games с указанным идентификатором.

### Параметры:
| Параметр | Тип | Описание |
| :--- | :--- | :--- |
| \`gameId\` | \`number\` | Уникальный числовой идентификатор целевой игры |

- **Возвращает:** \`Promise<void>\` — пустой результат при успешной навигации.

### Примеры вызова:
\`\`\`javascript
// async/await
try {
  await sdk.PlatformNavigation.goToGame(2);
  console.log('Успешный переход на игру с ID = 2');
} catch (err) {
  console.error('Ошибка навигации:', err);
}

// then/catch
sdk.PlatformNavigation.goToGame(2)
  .then(() => console.log('Перешли на игру с ID = 2'))
  .catch((err) => console.error('Ошибка навигации:', err));
\`\`\``,
      },
      {
        id: "player-data",
        title: "Игровые данные",
        summary: "Получение и сохранение пользовательского прогресса конкретной игры на сервере.",
        lastUpdated: "2026-09-24",
        content: `Модуль \`sdk.player\` предназначен для управления облачными сохранениями пользователя в рамках текущей игры.

### Получение сохраненных данных
Метод \`sdk.player.getData()\` запрашивает с сервера сохраненные данные игрока.

- **Параметры:** отсутствуют.
- **Возвращает:** \`Promise<Object>\` следующей структуры:

\`\`\`typescript
{
  playerName: string;
  playerGameData: Array<{
    identifier?: string;
    value?: string;
    values?: string[];
  }>;
}
\`\`\`

> [!NOTE]
> Поля со знаком \`?\` являются необязательными и могут отсутствовать, если игра еще не сохраняла данные.

\`\`\`javascript
// Получение данных игрока
const playerData = await sdk.player.getData();
console.log('Имя игрока:', playerData.playerName);
console.log('Массив данных:', playerData.playerGameData);
\`\`\`

### Сохранение данных игрока
Метод \`sdk.player.setData(data)\` сохраняет или обновляет игровые данные на сервере.

- **Параметр data:**
\`\`\`typescript
{
  playerName?: string;
  playerGameData?: Array<{
    identifier?: string;
    value?: string;
    values?: string[];
  }>;
}
\`\`\`

- **Возвращает:** \`Promise<void>\` — пустой результат при успешной фиксации.

\`\`\`javascript
// Сохранение прогресса
await sdk.player.setData({
  playerName: 'Игрок123',
  playerGameData: [
    { identifier: 'level', value: '7' },
    { identifier: 'weapon', value: 'sword', values: ['sword', 'axe'] },
    { identifier: 'coins', value: '1500' }
  ]
});
console.log('Данные игрока успешно сохранены на сервере');
\`\`\``,
      },
      {
        id: "metaverse-data",
        title: "Данные метавселенной",
        summary: "Глобальные сохранения игрока на уровне метавселенной и комбинированные данные.",
        lastUpdated: "2026-09-24",
        content: `В платформе Welwise Games поддерживается двухуровневая модель данных:
1. **Локальные данные игры (\`playerGameData\`)** — прогресс конкретного проекта.
2. **Данные метавселенной (\`playerMetaverseData\`)** — сквозные данные профиля, общие для всех игр (ранг, репутация, глобальные титулы, кросс-игровой инвентарь).

### Работа с данными метавселенной
- \`sdk.metaversePlayer.getData()\` — возвращает глобальные метаданные:
\`\`\`typescript
{
  playerName: string;
  playerMetaverseData: Array<{
    identifier?: string;
    value?: string;
    values?: string[];
  }>;
}
\`\`\`

- \`sdk.metaversePlayer.setData(data)\` — сохраняет данные игрока на уровне метавселенной:
\`\`\`javascript
await sdk.metaversePlayer.setData({
  playerName: 'User',
  playerMetaverseData: [
    { identifier: 'reputation', value: 'gold' },
    { identifier: 'inventory', values: ['sword', 'shield'] }
  ]
});
\`\`\`

### Комбинированные методы
Для оптимизации сетевых запросов SDK позволяет запрашивать и сохранять оба слоя данных одновременно:

- **Получение комбинированных данных:**
\`\`\`javascript
const combined = await sdk.metaversePlayer.getGameData();
console.log('Данные игры и метавселенной:', combined);
\`\`\`

- **Сохранение комбинированных данных:**
\`\`\`javascript
await sdk.metaversePlayer.setGameData({
  playerName: 'VRUser',
  playerMetaverseData: [{ identifier: 'rank', value: 'bronze' }],
  playerGameData: [{ identifier: 'score', value: '2000' }]
});
\`\`\``,
      },
      {
        id: "advertisement",
        title: "Реклама",
        summary: "Показ межуровневой рекламы и видео с вознаграждением.",
        lastUpdated: "2026-09-24",
        content: `После инициализации SDK становится доступен модуль \`sdk.AdvManager\` для управления показом рекламных блоков.

### Межуровневая реклама
Метод \`sdk.AdvManager.showMidgame({ callbacks })\` показывает полноэкранную рекламу между раундами или уровнями:

\`\`\`javascript
sdk.AdvManager.showMidgame({
  callbacks: {
    onOpen: () => console.log('Реклама открыта'),
    onClose: () => console.log('Реклама закрыта, продолжаем игру'),
    onError: (e) => console.log('Ошибка показа рекламы:', e.message)
  }
});
\`\`\`
- **Обязательные коллбэки:**
  - \`onClose\` — вызывается при закрытии рекламы;
  - \`onError\` — вызывается при ошибке показа.

### Реклама с вознаграждением
Метод \`sdk.AdvManager.showRewarded({ callbacks })\` воспроизводит видеоролик с начислением бонуса игроку:

\`\`\`javascript
sdk.AdvManager.showRewarded({
  callbacks: {
    onOpen: () => console.log('Реклама с наградой открыта'),
    onRewarded: () => {
      console.log('Вознаграждение получено игроком');
      giveCoinsToPlayer(100);
    },
    onClose: () => console.log('Реклама закрыта'),
    onError: (e) => console.log('Ошибка показа:', e.message)
  }
});
\`\`\`
- **Обязательные коллбэки:**
  - \`onRewarded\` — вызывается при успешном завершении просмотра видео (начисляйте награду здесь!);
  - \`onClose\` — вызывается при закрытии окна рекламы;
  - \`onError\` — вызывается при ошибке показа.

### Особенности работы с коллбэками
> [!WARNING]
> - При отсутствии обязательных callback'ов метод выполнится, но в консоль браузера будет выведено предупреждение.
> - Награду в игре выдавайте СТРОГО в коллбэке \`onRewarded\`, а не в \`onClose\`.
> - Можно передавать любое количество callback'ов, лишние будут проигнорированы.
> - При дублировании callback'ов сработает только последний переданный:
> \`\`\`javascript
> callbacks: {
>   onRewarded: () => console.log('1'), // игнорируется
>   onRewarded: () => console.log('2'), // игнорируется
>   onRewarded: () => console.log('3')  // сработает этот!
> }
> \`\`\``,
      },
      {
        id: "purchases",
        title: "Внутриигровые покупки",
        summary: "Каталог товаров за Welwise Coins, покупка и подтверждение выдачи.",
        lastUpdated: "2026-09-24",
        content: `В данном разделе описаны методы для совершения внутриигровых покупок за платформенную валюту Welwise Coins (1 Coin = 1 ₽).

> [!IMPORTANT]
> Перед интеграцией покупок каждый товар должен быть предварительно зарегистрирован на платформе. Для этого обратитесь к менеджеру Welwise Games и передайте \`itemId\` и цену в Coins для каждого товара.

### Получение списка доступных товаров
Метод \`sdk.purchases.getAvailableItems()\` возвращает список товаров, доступных для покупки в текущей игре:

- **Возвращает:** \`Promise<Object[]>\` — массив объектов с полями:
\`\`\`typescript
{
  itemId: string;        // уникальный идентификатор товара
  name: string;          // отображаемое название
  priceCoins: number;    // цена в Welwise Coins
  description?: string;  // описание товара
  imageUrl?: string;     // URL изображения товара
}
\`\`\`

\`\`\`javascript
const items = await sdk.purchases.getAvailableItems();
console.log('Доступные товары:', items);
\`\`\`

### Покупка товара
Метод \`sdk.purchases.purchase(itemId, options)\` инициирует покупку товара. Результат возвращается через коллбэки:

\`\`\`javascript
sdk.purchases.purchase('energy_potion_small', {
  quantity: 1,
  callbacks: {
    onSuccess: (data) => {
      console.log('Покупка успешна, purchaseId:', data.purchaseId);
      // 1. Выдать купленный предмет игроку
      giveItemToPlayer(data.itemId, data.quantity);
      // 2. ОБЯЗАТЕЛЬНО подтвердить покупку!
      sdk.purchases.confirmPurchase(data.purchaseId);
    },
    onError: (reason, error, purchaseId) => {
      console.error('Ошибка покупки:', reason, error.message);
    }
  }
});
\`\`\`

- **Коллбэк onSuccess получает:**
\`\`\`typescript
{
  purchaseId: string;   // уникальный идентификатор покупки
  itemId: string;       // идентификатор купленного товара
  quantity: number;     // количество
  debitedCoins: number; // списанная сумма в Coins
}
\`\`\`

- **Коллбэк onError получает параметры:**
  - \`reason (string)\` — код причины ошибки: \`USER_CANCELLED\`, \`INSUFFICIENT_FUNDS\`, \`ITEM_NOT_FOUND\`, \`INTERNAL_ERROR\` и др.
  - \`error (Error)\` — объект с сообщением об ошибке.
  - \`purchaseId (string, необязателен)\` — идентификатор покупки, если оплата прошла, но выдача не состоялась.

### Подтверждение покупки
Метод \`sdk.purchases.confirmPurchase(purchaseId)\` подтверждает, что купленный предмет был начислен игроку.

> [!IMPORTANT]
> **Необходимо вызывать после каждой успешной покупки!** Если покупка не подтверждена, система платформы может расценить транзакцию как незавершенную и вернуть средства игроку.

\`\`\`javascript
await sdk.purchases.confirmPurchase('63d50da8-d8aa-4bdf-88ed-320cc678190e');
console.log('Покупка подтверждена');
\`\`\`

### Неподтвержденные покупки
Метод \`sdk.purchases.getPendingPurchases()\` возвращает список оплаченных покупок, для которых еще не был вызван \`confirmPurchase()\`. Рекомендуется вызывать при каждом старте игры для доначисления зависших покупок:

\`\`\`javascript
const pending = await sdk.purchases.getPendingPurchases();
for (const p of pending) {
  giveItemToPlayer(p.itemId, p.quantity);
  await sdk.purchases.confirmPurchase(p.purchaseId);
}
\`\`\``,
      },
    ],
  },
  {
    id: 'unity-plugin',
    title: 'Unity Welwise Plugin',
    description: 'Пакет для Unity, кросс-платформенная адаптация сторонних SDK, WebGL-шаблоны и монетизация.',
    icon: 'Gamepad2',
    lastUpdated: '2026-09-24',
    articles: [
      {
        id: "installation",
        title: "Установка плагина",
        summary: "Подключение пакета WelwiseGamesSDK через Unity Package Manager по Git URL.",
        lastUpdated: "2026-09-24",
        content: `Пакет **WelwiseGamesSDK** для Unity разработан для интеграции игр с платформой Welwise Games, а также предоставляет кросс-платформенный абстрактный слой для работы с различными SDK через единый C# интерфейс.

### Шаги установки через UPM:
1. В Unity откройте окно Package Manager через верхнее меню: \`Window -> Package Manager\`:

![Открытие Package Manager](/images/docs/unity/unity_window_package_manager.png)

2. В открывшемся окне нажмите на кнопку с плюсом \`(+)\` в левом верхнем углу и выберите пункт **Add package from git URL...**:

![Добавление пакета из Git URL](/images/docs/unity/unity_add_package_git.png)

3. Введите адрес репозитория плагина на GitHub:
\`\`\`text
https://github.com/Welwise-Games/WelwiseGamesSDK.git
\`\`\`

4. Нажмите кнопку **Add**. Unity автоматически загрузит пакет и зарегистрирует его в манифесте проекта (\`Packages/manifest.json\`).`,
      },
      {
        id: "editor-settings",
        title: "Окно настроек",
        summary: "Конфигурация параметров, шаблона сборки WebGL и симуляции модулей в Unity Editor.",
        lastUpdated: "2026-09-24",
        content: `Для вызова окна параметров плагина используйте пункт верхнего меню Unity: \`WelwiseGamesSDK -> SDK Settings\`.

![Путь до открытия окна настроек](/images/docs/unity/unity_menu_sdk_settings.png)

---

### Вкладка General
![Вкладка General](/images/docs/unity/unity_settings_general.png)

> [!NOTE]
> Внешний вид и количество настроек могут отличаться в зависимости от выбранного SDK! Для некоторых SDK требуются различные идентификаторы (например, ID игры или рекламной сети). Подробности о специфических полях можно узнать в документации конкретной площадки.

| Название поля | Описание | Допустимые значения |
| :--- | :--- | :--- |
| **SDK Type** | Целевой тип SDK для интеграции | Список доступных SDK |
| **[Динамические поля]** | Конфигурационные поля выбранного SDK | Зависят от конкретного SDK |
| **Mute Audio On Pause** | Автоматически отключать звук игры при паузе | \`true\` / \`false\` |
| **Auto Singleton** | Автоматическое создание синглтона SDK при запуске | \`true\` / \`false\` |
| **Initialization Time** | Время симуляции задержки инициализации в редакторе | \`0 - 10\` секунд |
| **Advertisement Module** | Включить модуль рекламы в редакторе | \`true\` / \`false\` |
| **Payments Module** | Включить модуль платежей в редакторе | \`true\` / \`false\` |
| **Analytics Module** | Включить модуль аналитики в редакторе | \`true\` / \`false\` |
| **Environment Module** | Включить модуль окружения в редакторе | \`true\` / \`false\` |
| **Platform Navigation Module** | Включить модуль навигации в редакторе | \`true\` / \`false\` |
| **Player Data Module** | Включить модуль данных игрока в редакторе | \`true\` / \`false\` |
| **Game Data Module** | Включить модуль данных игры в редакторе | \`true\` / \`false\` |
| **Metaverse Data Module** | Включить модуль данных метавселенной в редакторе | \`true\` / \`false\` |

---

### Вкладка Build
![Вкладка Build](/images/docs/unity/unity_settings_build.png)

> [!WARNING]
> Использование ThreeJS загрузчика увеличит время загрузки WebGL-страницы, а также может увеличить расход оперативной памяти клиентского устройства.

| Название поля | Описание | Допустимые значения |
| :--- | :--- | :--- |
| **Use ThreeJS Loader** | Использовать продвинутый ThreeJS 3D загрузчик | \`true\` / \`false\` |
| **Aspect Ratio Mode** | Режим адаптации соотношения сторон экрана | \`Default\`, \`Custom\`, \`Fill Screen\`, \`Fixed\` |
| **Background Image** | Фоновое изображение экрана запуска | \`Texture2D\` из папки \`Resources\` |

---

### Вкладка Advertisement
![Вкладка Advertisement](/images/docs/unity/unity_settings_ad.png)

Настройки для симуляции рекламы в Unity Editor:

| Название поля | Описание | Допустимые значения |
| :--- | :--- | :--- |
| **Ad Duration** | Длительность симуляции показа рекламы | \`0 - 10\` секунд |
| **Interstitial Result** | Результат симуляции межстраничной рекламы | \`Closed\`, \`Failed\`, \`Open\`, \`None\` |
| **Rewarded Result** | Результат симуляции рекламы с вознаграждением | \`Closed\`, \`Failed\`, \`Open\`, \`Rewarded\`, \`None\` |

---

### Вкладка Environment
![Вкладка Environment](/images/docs/unity/unity_settings_environment.png)

Настройки для отладки среды выполнения в Unity Editor:

| Название поля | Описание | Допустимые значения |
| :--- | :--- | :--- |
| **Player ID** | Идентификатор игрока для локальной отладки | Любая строка или GUID (есть кнопка генерации) |
| **Device Type** | Тип клиентского устройства для симуляции | \`Desktop\`, \`Mobile\`, \`Tablet\` |
| **Language** | Язык для симуляции | Двухбуквенный код, например \`ru\`, \`en\` |

---

### Вкладка Payments
![Вкладка Payments](/images/docs/unity/unity_settings_payments.png)

Настройки для тестирования магазина и покупок в Unity Editor:

| Название поля | Описание | Допустимые значения |
| :--- | :--- | :--- |
| **Simulation Duration** | Длительность симуляции проведения покупки | \`0.5 - 10\` секунд |
| **Products** | Список моковых товаров для тестирования каталога | Массив \`Product\` (\`ID\`, \`Title\`, \`Description\`, \`Image URI\`, \`Price\`, \`Currency\`, \`Price Value\`) |
| **Purchases** | Список совершенных покупок для проверки инвентаря | Массив \`Purchase\` (\`Product ID\`, \`Token\`, \`Payload\`, \`Signature\`) |

---

### Вкладка Player Data
![Вкладка Player Data](/images/docs/unity/unity_settings_player_data.png)

> [!TIP]
> Здесь можно просматривать, вручную редактировать и сбрасывать локальные сохранения плагина прямо в окне инспектора.

| Название поля | Описание | Допустимые значения |
| :--- | :--- | :--- |
| **Load Save On Initialize** | Автоматически загружать сохранения при инициализации | \`true\` / \`false\` |
| **Delete All Player Data** | Кнопка полной очистки всех локальных сохранений | Кнопка \`Delete All\` |
| **Player Name** | Тестовое имя игрока | Любая строка |
| **Container** | Контейнер для добавления нового поля | \`Game Data\`, \`Metaverse Data\` |
| **Key Name** | Имя ключа сохранения | Любая строка |
| **Value Type** | Тип данных сохраняемого поля | \`Integer\`, \`Float\`, \`Boolean\`, \`String\` |
| **Value** | Тестовое значение | Соответствует выбранному типу |

---

### Вкладка About
![Вкладка About](/images/docs/unity/unity_settings_about.png)

| Название поля | Описание | Значение |
| :--- | :--- | :--- |
| **Template Version** | Версия шаблона WebGL SDK | Только чтение (текущая: \`1.0.2\`) |
| **Manifest Version** | Версия манифеста пакета | Только чтение |
| **Documentation** | Кнопка перехода к документации | Открывает сайт документации |`,
      },
      {
        id: "initialization",
        title: "Создание и инициализация",
        summary: "Способы создания и синхронная/асинхронная инициализация.",
        lastUpdated: "2026-09-24",
        content: `### Создание объекта SDK
Существует 2 способа создать объект SDK:

1. **Как Singleton:** подходит для небольших проектов, в которых нет DI. Обращения к SDK будут происходить через статическое свойство \`WelwiseSDK.Instance\`. Если в настройках плагина включен пункт \`Auto Singleton\`, данный этап можно пропустить — синглтон будет создан автоматически:
\`\`\`csharp
WelwiseSDK.Construct().AsSingle();
\`\`\`

2. **Как отдельный объект:** подходит для проектов с использованием DI-контейнеров (VContainer, Zenject) или при проблемах с Domain Reloading:
\`\`\`csharp
var sdk = WelwiseSDK.Construct().AsTransient();
// Возвращает объект типа ISDK для регистрации в контейнере
\`\`\`

### Способы инициализации
Инициализацию можно производить двумя способами — синхронным и асинхронным.

- **Синхронная инициализация с событием \`Initialized\`:**
\`\`\`csharp
var sdk = WelwiseSDK.Construct().AsTransient();
sdk.Initialize();
sdk.Initialized += OnInitialized;

void OnInitialized()
{
    Debug.Log("SDK успешно инициализирован!");
}
\`\`\`

- **Асинхронная инициализация с async / await:**
\`\`\`csharp
var sdk = WelwiseSDK.Construct().AsTransient();
await sdk.InitializeAsync();
// Если SDK уже инициализирован, вернется Task.CompletedTask
\`\`\`

### Проверка состояния инициализации
Для проверки готовности SDK используется булево свойство \`IsInitialized\`:
\`\`\`csharp
if (sdk.IsInitialized)
{
    // SDK готов к выполнению вызовов
}
\`\`\``,
      },
      {
        id: "environment",
        title: "Окружение и серверное время",
        summary: "Интерфейс IEnvironment: PlayerId, DeviceType, Language и серверное время.",
        lastUpdated: "2026-09-24",
        content: `Данные, связанные с окружением, в котором запущен клиент игры, вынесены в отдельное свойство \`sdk.Environment\`.

### Интерфейс \`IEnvironment\`:
\`\`\`csharp
public interface IEnvironment
{
    // Уникальный идентификатор игрока
    public Guid PlayerId { get; }

    // Тип клиентского устройства игрока (Desktop, Mobile, Tablet)
    public DeviceType DeviceType { get; }

    // Код языка интерфейса игрока, например "ru", "en"
    public string Language { get; }

    // Запрос серверного времени с обратным вызовом
    public void RequestServerTime(Action<long> callback);
}
\`\`\`

### Методы-расширения
Для удобства ветвления логики и адаптации управления доступны расширения:
\`\`\`csharp
bool isDesktop = sdk.Environment.IsDesktop; // true если Desktop
bool isMobile = sdk.Environment.IsMobile;   // true если Mobile или Tablet
\`\`\`

### Пример использования:
\`\`\`csharp
var sdk = WelwiseSDK.Construct().AsTransient();

Debug.Log($"Игрок GUID: {sdk.Environment.PlayerId}");
Debug.Log($"Устройство: {sdk.Environment.DeviceType}, Мобильное: {sdk.Environment.IsMobile}");
Debug.Log($"Язык: {sdk.Environment.Language}");

// Запрос времени сервера для синхронизации ежедневных наград
sdk.Environment.RequestServerTime((serverTimestamp) =>
{
    Debug.Log($"Точное время сервера: {serverTimestamp}");
});
\`\`\``,
      },
      {
        id: "player-data",
        title: "Данные игрока и сохранения",
        summary: "Интерфейс IPlayerData и работа с хранилищем IData.",
        lastUpdated: "2026-09-24",
        content: `Для доступа к пользовательским сохранениям используется свойство \`sdk.PlayerData\`.

### Интерфейс \`IPlayerData\`:
\`\`\`csharp
public interface IPlayerData
{
    // Получение имени игрока
    public string GetPlayerName();

    // Установка имени игрока
    public void SetPlayerName(string name);

    // Хранилище локальных игровых сохранений текущей игры
    public IData GameData { get; }

    // Хранилище глобальных сохранений метавселенной
    public IData MetaverseData { get; }

    // Метод сохранения данных на сервер платформы
    public void Save();
}
\`\`\`

### Интерфейс \`IData\`:
\`\`\`csharp
namespace WelwiseGamesSDK.Shared
{
    public interface IData
    {
        public void SetString(string key, string value);
        public string GetString(string key, string defaultValue = "");

        public void SetInt(string key, int value);
        public int GetInt(string key, int defaultValue = 0);

        public void SetFloat(string key, float value);
        public float GetFloat(string key, float defaultValue = 0f);

        public void SetBool(string key, bool value);
        public bool GetBool(string key, bool defaultValue = false);

        // Проверка наличия ключа в хранилище
        public bool HasKey(string key);
    }
}
\`\`\`

> [!IMPORTANT]
> **Если не вызывать метод \`Save()\`, то измененные данные не будут зафиксированы на сервере платформы!**

> [!WARNING]
> Лучше не сохранять данные в игру и метавселенную под одинаковыми именами ключей во избежание логических конфликтов.

### Пример использования:
\`\`\`csharp
var sdk = WelwiseSDK.Construct().AsTransient();

// Чтение очков с дефолтным значением 100
int coins = sdk.PlayerData.GameData.GetInt("coins", 100);

// Изменение значения и отправка в облако
sdk.PlayerData.GameData.SetInt("coins", coins + 50);
sdk.PlayerData.Save();
\`\`\``,
      },
      {
        id: "analytics",
        title: "Метрики и аналитика",
        summary: "Отправка игровых событий и обязательный вызов Game Ready API.",
        lastUpdated: "2026-09-24",
        content: `Для отправки метрик и платформенных событий используется модуль \`sdk.Analytics\`.

> [!WARNING]
> Отправка метрик активна при сборке под Yandex Games.

### Интерфейс \`IAnalytics\`:
\`\`\`csharp
public interface IAnalytics
{
    // Отправка пользовательского события
    public void SendEvent(string eventName);

    // Отправка события с сопутствующими строковыми или JSON данными
    public void SendEvent(string eventName, string data);

    // Отправка события о полной готовности игры к взаимодействию
    // Вызывается когда игрок получил доступ к игре и нет экранов загрузки
    public void GameReady();

    // Вызывается когда игрок приступает к активному геймплею
    public void GameplayStart();

    // Вызывается когда игровой процесс приостановлен (катсцена, диалог, меню)
    public void GameplayEnd();
}
\`\`\`

> [!IMPORTANT]
> Площадки требуют обязательного вызова **Game Ready API**. Вызов метода \`GameReady()\` **строго обязателен** сразу после завершения загрузки ресурсов и исчезновения лоадера!

### Пример использования:
\`\`\`csharp
var sdk = WelwiseSDK.Construct().AsTransient();

// Игра загрузилась, скрываем стартовый экран
sdk.Analytics.GameReady();

// Начало сессии прохождения уровня
sdk.Analytics.GameplayStart();
sdk.Analytics.SendEvent("levelCompleted", "{"level": 1}");
\`\`\``,
      },
      {
        id: "navigation",
        title: "Навигация по платформе",
        summary: "Переход между играми каталога через IPlatformNavigation.GoToGame.",
        lastUpdated: "2026-09-24",
        content: `Данная возможность доступна при использовании SDK Welwise Games. Для перехода используется свойство \`sdk.PlatformNavigation\`.

### Интерфейс \`IPlatformNavigation\`:
\`\`\`csharp
public interface IPlatformNavigation
{
    // Переход к игре с заданным числовым идентификатором
    // При возникновении ошибки вызывается callback onError
    public void GoToGame(int id, Action<string> onError);
}
\`\`\`

### Пример использования:
\`\`\`csharp
var sdk = WelwiseSDK.Construct().AsTransient();

sdk.PlatformNavigation.GoToGame(1, (errorMsg) =>
{
    Debug.LogError($"Ошибка перехода к игре: {errorMsg}");
    // Возврат к нормальному игровому процессу
});
\`\`\``,
      },
      {
        id: "advertisement",
        title: "Реклама",
        summary: "Интерфейс IAdvertisement, обработка состояний InterstitialState и RewardedState.",
        lastUpdated: "2026-09-24",
        content: `Вызовы показа рекламы вынесены в модуль \`sdk.Advertisement\`.

### Интерфейс \`IAdvertisement\`:
\`\`\`csharp
public interface IAdvertisement
{
    // Показ межстраничной рекламы
    public void ShowInterstitial(Action<InterstitialState> callbackState = null);

    // Показ рекламы с вознаграждением
    public void ShowRewarded(Action<RewardedState> callbackState = null);
}
\`\`\`

### Возвращаемые состояния:
- **\`RewardedState\`:**
  - \`RewardedState.Error\` — ошибка показа;
  - \`RewardedState.Opened\` — рекламный оверлей открыт;
  - \`RewardedState.Closed\` — закрыта игроком досрочно, награду не выдавать;
  - \`RewardedState.Rewarded\` — ролик просмотрен до момента начисления награды.

- **\`InterstitialState\`:**
  - \`InterstitialState.Error\` — ошибка показа;
  - \`InterstitialState.Opened\` — реклама открылась;
  - \`InterstitialState.Closed\` — реклама закрыта.

> [!NOTE]
> В Unity Editor реклама отображается через интерфейс OnGUI, где можно интерактивно выбрать результат показа для симуляции игровых сценариев.

### Пример кода:
\`\`\`csharp
var sdk = WelwiseSDK.Construct().AsTransient();

// Показ Rewarded рекламы с начислением жизней
sdk.Advertisement.ShowRewarded(state =>
{
    Debug.Log($"Rewarded State: {state}");
    if (state == RewardedState.Rewarded)
    {
        AddExtraLife();
    }
});

// Показ межстраничной рекламы между раундами
sdk.Advertisement.ShowInterstitial(state =>
{
    Debug.Log($"Interstitial State: {state}");
});
\`\`\``,
      },
      {
        id: "purchases",
        title: "Внутриигровые покупки",
        summary: "Интерфейс IPayments, каталог Product, обработка Purchase и подтверждение выдачи.",
        lastUpdated: "2026-09-24",
        content: `Для работы с покупками используется модуль \`sdk.Payments\`.

### Интерфейс \`IPayments\`:
\`\`\`csharp
public interface IPayments : IModule, IInitializable
{
    // Основные методы
    public void Purchase(string productId, string developerPayload = null);
    public void Consume(string purchaseToken);

    // Методы загрузки данных
    public void LoadCatalog();
    public void LoadPurchases();

    // Свойства
    public IReadOnlyList<Product> Products { get; }
    public IReadOnlyList<Purchase> Purchases { get; }

    // События успешных операций
    public event Action<Purchase> PurchaseSuccess;
    public event Action<string> ConsumeSuccess;
    public event Action CatalogUpdated;
    public event Action PurchasesUpdated;

    // События ошибок
    public event Action<string, string> PurchaseFailed;
    public event Action<string, string> ConsumeFailed;
    public event Action<string> CatalogLoadFailed;
    public event Action<string> PurchasesLoadFailed;
}
\`\`\`

### Модели данных:
\`\`\`csharp
public class Product
{
    public string Id;                // Идентификатор товара
    public string Title;             // Отображаемое название товара
    public string Description;       // Описание товара
    public string ImageUri;          // URL изображения товара
    public string Price;             // Цена в формате строки
    public string PriceCurrencyCode; // Код валюты
    public string PriceValue;        // Числовое значение цены
}

public class Purchase
{
    public string ProductId;        // Идентификатор купленного товара
    public string PurchaseToken;    // Токен покупки для подтверждения
    public string DeveloperPayload; // Дополнительные метаданные
    public string Signature;        // Цифровая подпись
}
\`\`\`

### Важные правила работы:
1. **Инициализация:** модуль должен быть инициализирован перед использованием.
2. **Загрузка данных:** перед отображением товаров в UI вызовите \`LoadCatalog()\`.
3. **Обработка событий:** все операции покупок асинхронны, результаты приходят в события (\`PurchaseSuccess\`, \`PurchaseFailed\` и др.).
4. **Подтверждение потребления:** для расходуемых товаров **обязательно** вызывать \`Consume(purchaseToken)\` после выдачи награды игроку!
5. **Проверка подлинности:** подпись \`Signature\` и параметр \`developerPayload\` поддерживаются при работе через Yandex Games SDK.

### Полный пример интеграции:
\`\`\`csharp
var sdk = WelwiseSDK.Construct().AsTransient();

// Подписка на события
sdk.Payments.PurchaseSuccess += (purchase) =>
{
    Debug.Log($"Покупка успешна: {purchase.ProductId}");
    GiveReward(purchase.ProductId);

    // Подтверждаем расходование товара
    sdk.Payments.Consume(purchase.PurchaseToken);
};

sdk.Payments.PurchaseFailed += (productId, error) =>
{
    Debug.LogError($"Ошибка покупки {productId}: {error}");
};

sdk.Payments.CatalogUpdated += () =>
{
    Debug.Log($"Каталог загружен, доступно товаров: {sdk.Payments.Products.Count}");
};

// Загрузка каталога и покупок
sdk.Payments.LoadCatalog();
sdk.Payments.LoadPurchases();

// Инициация покупки
sdk.Payments.Purchase("premium_coins_100");
\`\`\``,
      },
      {
        id: "custom-sdk-adapters",
        title: "Интеграция новых SDK",
        summary: "Модульная архитектура, файлы конфигурации в Resources/SDKs и мост sdk-adapter.jslib.",
        lastUpdated: "2026-09-24",
        content: `### Архитектура плагина
Плагин построен по модульной архитектуре, где основное платформенное взаимодействие происходит через JavaScript-слой:

\`\`\`text
Unity Game → PluginRuntime (C#) → sdk-adapter.jslib → Custom SDK Adapter → Target SDK
\`\`\`

Плагин описывает свой унифицированный контракт в библиотеке \`sdk-adapter.jslib\`. Этот интерфейс реализует JavaScript-адаптер конкретной площадки.

Благодаря этому различные SDK приводятся к единому виду, плагин на стороне Unity остается неизменным, а адаптер может быть даже подменен в готовом WebGL билде без необходимости повторной сборки проекта!

---

### Файлы интеграции в \`Resources/SDKs\`
Файлы интеграции SDK должны иметь расширение \`.txt\`. Для каждого нового SDK минимально требуется два файла:

1. **Файл определения:** \`SDKName_defenition.txt\` — описывается в формате JSON согласно схеме:
\`\`\`json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "SDKDefinition",
  "type": "object",
  "properties": {
    "name": { "type": "string" },
    "adapterFile": { "type": "string" },
    "configFields": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "name": { "type": "string" },
          "type": { "type": "string", "enum": ["string", "bool", "float", "int"] },
          "displayName": { "type": "string" },
          "defaultValue": { "type": "string" },
          "tooltip": { "type": "string" }
        },
        "required": ["name", "type"]
      }
    },
    "postBuildScripts": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "injectPoint": { "type": "string", "enum": ["head", "before_body_end", "after_body_start"] },
          "file": { "type": "string" }
        },
        "required": ["injectPoint", "file"]
      }
    }
  },
  "required": ["name", "adapterFile"]
}
\`\`\`

2. **Файл адаптера:** \`SDKName_adapter.txt\` — JavaScript-код адаптера, реализующий методы плагина. Копируется в итоговый WebGL билд с расширением \`.js\` и подключается в \`index.html\`.

3. **Фрагменты инъекции:** \`SDKName_injection_PLACE.txt\` — опциональный HTML/JS код для вставки в \`head\`, \`before_body_end\`, \`after_body_start\`. В коде поддерживаются переменные \`{ConfigFieldName}\`, которые автоматически подменяются на значения из настроек в Unity Inspector.

---

### Объект \`window.__sdk_adapter\`
Адаптер должен определять глобальный объект \`window.__sdk_adapter\`:

\`\`\`javascript
window.__sdk_adapter = {
  // Список поддерживаемых модулей
  GetModules: function() {
    return ["Advertisement", "PlayerData", "Payments", "Environment"];
  },
  // Инициализация целевого SDK
  Init: function() {
    // Логика инициализации
  }
  // Остальные методы...
};
\`\`\`

### Полный список методов адаптера по категориям:
| Категория | Методы |
| :--- | :--- |
| **Системные** | \`GetModules\`, \`Init\` |
| **Данные игрока** | \`GetPlayerData\`, \`SetPlayerData\` |
| **Окружение** | \`GetPlayerId\`, \`GetDeviceType\`, \`GetLanguageCode\`, \`GetServerTime\` |
| **Реклама** | \`ShowInterstitial\`, \`ShowRewarded\` |
| **Платежи** | \`PaymentsInit\`, \`PaymentsGetCatalog\`, \`PaymentsGetPurchases\`, \`PaymentsPurchase\`, \`PaymentsConsume\` |
| **Навигация** | \`GoToGame\` |
| **Аналитика** | \`GameReady\`, \`GameplayStart\`, \`GameplayStop\` |
| **Метавселенная** | \`IsMetaverseSupported\`, \`GetMetaversePlayerData\`, \`SetMetaversePlayerData\` |
| **Комбинированные** | \`GetCombinedPlayerData\`, \`SetCombinedPlayerData\` |

---

### События обратного вызова в Unity
Адаптер отправляет следующие события в среду выполнения Unity:

- **Системные:** \`HandleInitSuccess\`, \`HandleInitError\`
- **Данные:** \`HandleGetDataSuccess\`, \`HandleGetDataError\`, \`HandleSetDataSuccess\`, \`HandleSetDataError\`
- **Реклама:** \`HandleInterstitialOpen\`, \`HandleInterstitialClose\`, \`HandleInterstitialError\`, \`HandleRewardedOpen\`, \`HandleRewardedRewarded\`, \`HandleRewardedClose\`, \`HandleRewardedError\`
- **Платежи:** \`HandlePaymentsInitSuccess\`, \`HandlePaymentsInitError\`, \`HandlePaymentsGetCatalogSuccess\`, \`HandlePaymentsPurchaseSuccess\``,
      },
    ],
  },
  {
    id: 'rules',
    title: 'Правила и Регламент',
    description: 'Официальные требования к безопасности, стабильности и контенту для прохождения модерации.',
    icon: 'ShieldCheck',
    lastUpdated: '2026-09-12',
    articles: [
      {
        id: 'rules-catalog',
        title: 'Каталог правил платформы',
        summary: 'Полный список требований с кодами (SEC, SRV, BLD, CNT) и советами по исправлению.',
        lastUpdated: '2026-09-12',
        isRulesCatalog: true,
      },
      {
        id: 'moderation-process',
        title: 'Процесс модерации и ревью',
        summary: 'Как проходит проверка проекта, чат с модератором и исправление замечаний.',
        lastUpdated: '2026-09-12',
        content: `Каждая игра и запрос на доступ к серверам проходят обязательную ручную и автоматизированную проверку командой модераторов платформы.

**Статусы заявки на модерацию:**
- **В очереди (Pending)** — заявка принята и ожидает назначения модератора.
- **На проверке (In Review)** — модератор тестирует сборку в изолированном стенде. Открывается чат проекта для оперативной связи между разработчиком и модератором.
- **Одобрено (Approved)** — проект публикуется в каталоге или получает квоту серверных мощностей.
- **Отклонено (Rejected)** — заявка возвращается разработчику с подробным списком нарушений, прикрепленными скриншотами/видеозаписями багов и кликабельными ссылками на конкретные пункты регламента.`,
      },
    ],
  },
];

/**
 * Получить данные раздела по его id.
 */
export function getSectionById(sectionId?: string | null): DocSection | null {
  if (!sectionId) return null;
  return DOCS_SECTIONS.find((s) => s.id === sectionId) || null;
}

/**
 * Получить данные конкретной статьи по id раздела и id статьи.
 */
export function getArticleById(sectionId?: string | null, articleId?: string | null): DocArticle | null {
  if (!sectionId || !articleId) return null;
  const section = getSectionById(sectionId);
  if (!section) return null;
  return section.articles?.find((a) => a.id === articleId) || null;
}
