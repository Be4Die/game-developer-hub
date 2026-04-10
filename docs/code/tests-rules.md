# Правила написания тестов

## Структура тестов по папкам

tests/
  e2e/           — end-to-end, реальный gRPC через TCP, контейнер для каждого теста
  integration/   — интеграция компонентов, bufconn in-memory gRPC, memory storage
  internal/*/    — unit тесты рядом с тестируемым кодом (_test.go)

## Когда какой уровень использовать

Unit тесты — чистая логика без внешних зависимостей. Сервисы с моками, конвертеры, конфиги. Без Docker, без сети, без файлов.

Integration тесты — проверка что компоненты работают вместе. bufconn для in-memory gRPC, реальная memory storage, реальный Docker runtime. Каждый тест получает чистое окружение через setupIntegration(t).

E2E тесты — полный стек как в продакшене. Тест поднимает реальный контейнер game-server-node через testcontainers, подключается по TCP, проверяет поведение с точки зрения клиента.

## Именование тестов и файлов

Файлы группируем по домену: discovery_test.go, deployment_test.go, grpc_test.go и т.д. Common функции (setup, helpers) выносим в отдельные файлы: e2e_test.go и setup_test.go.

Имя теста: Test<Уровень>_<Компонент>_<ЧтоПроверяет>

Примеры:
  TestE2E_Discovery_ListInstances
  TestGRPC_Heartbeat
  TestDockerRuntime_ContainerLogs
  TestDeploymentService_FullLifecycle

## Изоляция тестов

Каждый тест должен быть независимым. Никакого общего состояния между тестами.

E2E: setupServerContainer(t) поднимает свежий контейнер для каждого теста. Контейнер уничтожается автоматически через t.Cleanup. Порт определяется динамически через MappedPort. Не подключайтесь к запущенному серверу вручную.

Integration: setupIntegration(t) создаёт новые storage, runtime, сервисы и gRPC сервер через bufconn. Ничего общего между тестами.

Если тест создаёт ресурс (инстанс, контейнер) — он должен быть удалён в том же тесте через t.Cleanup или defer.

## Контекст и таймауты

Каждый тест создаёт свой контекст с таймаутом:

  ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
  defer cancel()

Не используйте контекст теста в t.Cleanup — к моменту cleanup он может быть истёк. Внутри cleanup создавайте свой контекст:

  t.Cleanup(func() {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    // cleanup логика
  })

t.Cleanup предпочтительнее defer когда нужно гарантировать выполнение даже при t.Fatalf.

## Helpers

Общие вспомогательные функции (startTestInstance, loadTestImage) размещайте в workflow файлах e2e или в setup файлах integration. Они должны принимать t *testing.T первым параметром после ctx и вызывать t.Helper().

## Что не нужно делать

- Не смешивайте unit и интеграционную логику в одном файле
- Не полагайтесь на запущенный вручную контейнер в e2e тестах
- Не делайте cleanupInstances — если тест требует очистки состояния, значит тест не изолирован
- Не используйте глобальные переменные для хранения состояния между тестами
- Не пишите тесты которые зависят от порядка запуска
- Не используйте grpc.WithBlock и grpc.WithTimeout — они deprecated, таймауты через контекст

## Запуск

  go test ./...                          — unit тесты (быстро, всегда проходят)
  go test -tags=integration ./tests/integration/...  — интеграция (нужен Docker)
  go test -tags=e2e ./tests/e2e/...      — e2e (нужен Docker + собранный образ)

Для CI запускай все три. Для локальной разработки чаще всего хватает unit.
