/**
 * Структурированные статьи и разделы документации Game Developer Hub.
 * Поддерживает полноценную двухуровневую иерархию: Раздел (Root Section) -> Статья (Article).
 */

export const DOCS_SECTIONS = [
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
- **GDH SDK** — готовые клиентские и серверные библиотеки авторизации, облачных сохранений и покупок.
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
    title: 'GDH SDK Reference',
    description: 'Клиентские и серверные библиотеки для Unity, Unreal Engine, Go и веб-клиентов.',
    icon: 'Code2',
    lastUpdated: '2026-09-18',
    articles: [
      {
        id: 'init-auth',
        title: 'Инициализация и Авторизация',
        summary: 'Подключение SDK в Unity/Unreal, получение Player Session Token.',
        lastUpdated: '2026-09-18',
        content: `При запуске игры через лаунчер GDH клиент автоматически получает одноразовый \`GDH_AUTH_TOKEN\` через аргументы командной строки или переменные окружения.`,
        codeTabs: [
          {
            label: 'C# (Unity)',
            language: 'csharp',
            code: `using GDH.Client;
using UnityEngine;

public class GameInitializer : MonoBehaviour
{
    async void Start()
    {
        // Инициализируем SDK
        bool ok = await GDHClient.InitializeAsync(new GDHConfig
        {
            AppId = "my-game-id",
            Environment = GDHEnvironment.Production
        });

        if (ok)
        {
            GDHUser user = GDHClient.CurrentUser;
            Debug.Log($"Авторизован игрок: {user.DisplayName} (ID: {user.Id})");
        }
    }
}`,
          },
          {
            label: 'C++ (Unreal Engine)',
            language: 'cpp',
            code: `#include "GDHClientSubsystem.h"

void AMyGameMode::BeginPlay()
{
    Super::BeginPlay();

    UGDHClientSubsystem* GDH = GetGameInstance()->GetSubsystem<UGDHClientSubsystem>();
    if (GDH)
    {
        GDH->Initialize(TEXT("my-game-id"), FOnGDHInitComplete::CreateLambda([](bool bSuccess) {
            if (bSuccess) {
                UE_LOG(LogTemp, Display, TEXT("GDH SDK успешно подключен!"));
            }
        }));
    }
}`,
          },
        ],
      },
      {
        id: 'cloud-saves',
        title: 'Облачные сохранения (Cloud Saves)',
        summary: 'Сохранение прогресса, синхронизация стейта и разрешение конфликтов.',
        lastUpdated: '2026-09-18',
        content: `Сохранение пользовательского прогресса в формате JSON или Binary (до 5 МБ на слот):`,
        codeLanguage: 'csharp',
        codeSnippet: `// Сохранение игрового прогресса
var playerSave = new PlayerSaveData
{
    Level = 42,
    Coins = 1500,
    UnlockedSkins = new List<string> { "cyber_ninja", "gold_armor" }
};

string jsonData = JsonUtility.ToJson(playerSave);
await GDHClient.CloudStorage.SaveSlotAsync("slot_auto_01", jsonData);

// Загрузка слота
string loadedJson = await GDHClient.CloudStorage.LoadSlotAsync("slot_auto_01");
PlayerSaveData loaded = JsonUtility.FromJson<PlayerSaveData>(loadedJson);`,
      },
      {
        id: 'in-game-purchases',
        title: 'Внутриигровые покупки и транзакции',
        summary: 'Каталог товаров, покупка за виртуальную валюту и верификация чеков.',
        lastUpdated: '2026-09-17',
        content: `Интеграция каталога товаров платформы и проведение безопасных покупок:`,
        codeLanguage: 'csharp',
        codeSnippet: `// Получение списка доступных внутриигровых товаров
var products = await GDHClient.Purchases.GetCatalogAsync();

// Инициация покупки товара
var order = await GDHClient.Purchases.PurchaseItemAsync("skin_dragon_sword");
if (order.IsSuccess)
{
    Debug.Log($"Покупка успешна: {order.TransactionId}");
    GrantItemToPlayer(order.ItemId);
}`,
      },
      {
        id: 'server-sdk',
        title: 'Серверный SDK (Server-Side)',
        summary: 'Регистрация комнат, обновление количества игроков, отправка логов.',
        lastUpdated: '2026-09-18',
        content: `Методы для взаимодействия Dedicated Server с оркестратором:`,
        codeLanguage: 'csharp',
        codeSnippet: `// Уведомление оркестратора о готовности принимать подключения игроков
await GDHServer.ReportReadyAsync(new ServerReadyPayload
{
    CurrentPlayers = 0,
    MaxPlayers = 16,
    MapName = "CyberCity_Arena"
});

// Обновление состояния матча в реальном времени
await GDHServer.UpdatePlayerCountAsync(currentPlayers: 12);

// Оповещение об окончании сессии и освобождении инстанса
await GDHServer.ReportMatchEndedAsync();`,
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
export function getSectionById(sectionId) {
  return DOCS_SECTIONS.find((s) => s.id === sectionId) || null;
}

/**
 * Получить данные конкретной статьи по id раздела и id статьи.
 */
export function getArticleById(sectionId, articleId) {
  const section = getSectionById(sectionId);
  if (!section) return null;
  return section.articles?.find((a) => a.id === articleId) || null;
}
