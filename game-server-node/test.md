

# Отчёт о ручном тестировании микросервиса game-server-node

## 1. Общие сведения

| Параметр | Значение |
|---|---|
| Объект тестирования | Микросервис `game-server-node` (агент вычислительного узла) |
| Вид тестирования | Ручное функциональное тестирование |
| Инструмент | `grpcurl` (CLI-клиент для gRPC) |
| Транспортный протокол | gRPC (Protocol Buffers v3) |
| Адрес сервиса | `localhost:44044` |
| Окружение | Локальное (`env: local`) |
| Дата составления | Июль 2025 г. |

## 2. Предварительные условия

### 2.1. Установка grpcurl

**Linux / macOS:**

```bash
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

**Windows (PowerShell):**

```powershell
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

Альтернативно — загрузить бинарный релиз со страницы [github.com/fullstorydev/grpcurl/releases](https://github.com/fullstorydev/grpcurl/releases).

### 2.2. Сборка и запуск микросервиса

```bash
cd game-server-node
go build -o node ./cmd/node
./node --config=./config/local.yaml
```

Либо через Docker:

```bash
docker build -t game-server-node .
docker run -p 44044:44044 game-server-node
```

### 2.3. Подготовка Docker-образа для тестирования деплоя

Для тестирования RPC-методов `LoadImage` и `StartInstance` необходим тестовый Docker-образ. Создание минимального образа:

```bash
# Создать временную директорию
mkdir /tmp/test-game-server && cd /tmp/test-game-server

# Создать Dockerfile
cat > Dockerfile <<'EOF'
FROM alpine:3.21
EXPOSE 7777
CMD ["sh", "-c", "echo 'Game server started on port 7777' && sleep infinity"]
EOF

# Собрать образ
docker build -t test-game:v1 .

# Сохранить образ в tar-архив (для LoadImage)
docker save test-game:v1 -o test-game-v1.tar
```

### 2.4. Соглашения по вызовам

Все команды `grpcurl` используют флаг `-plaintext`, поскольку сервис в локальном окружении работает без TLS. Флаг `-import-path` и `-proto` указывают путь к proto-файлам для корректной сериализации.

Общий шаблон вызова:

```bash
grpcurl -plaintext \
  -import-path ./protos \
  -proto game_server_node/v1/<файл>.proto \
  -d '<JSON-тело запроса>' \
  localhost:44044 \
  game_server_node.v1.<Сервис>/<Метод>
```

Для методов без тела запроса параметр `-d` опускается либо передаётся `'{}'`.

---

## 3. Тестируемые gRPC-сервисы

Микросервис предоставляет два gRPC-сервиса:

| № | Сервис | Proto-файл | Назначение |
|---|---|---|---|
| 1 | `DiscoveryService` | `discovery.proto` | Обнаружение узла, получение информации об инстансах |
| 2 | `DeploymentService` | `deployment.proto` | Загрузка образов, управление жизненным циклом инстансов |

---

## 4. Тестовые сценарии

### 4.1. Сервис DiscoveryService

#### TC-01. Получение информации об узле (GetNodeInfo)

**Цель:** Проверить, что метод `GetNodeInfo` возвращает корректную информацию о вычислительном узле (регион, ресурсы, версию агента, время запуска).

**Предусловия:** Микросервис запущен и доступен на `localhost:44044`.

**Команда:**

```bash
grpcurl -plaintext \
  -import-path ./protos \
  -proto game_server_node/v1/discovery.proto \
  localhost:44044 \
  game_server_node.v1.DiscoveryService/GetNodeInfo
```

**Ожидаемое поведение:**
- Статус ответа: `OK` (код 0).
- Тело ответа содержит поля: `region`, `cpuCores`, `totalMemoryBytes`, `totalDiskBytes`, `networkBandwidthBytesPerSec`, `startedAt`, `agentVersion`.
- Значение `region` соответствует конфигурации (`"local"`).
- Значение `agentVersion` соответствует конфигурации (`"0.0.1"`).

---

#### TC-02. Получение heartbeat-данных (Heartbeat)

**Цель:** Проверить, что метод `Heartbeat` возвращает текущую утилизацию ресурсов узла и количество активных инстансов.

**Предусловия:** Микросервис запущен.

**Команда:**

```bash
grpcurl -plaintext \
  -import-path ./protos \
  -proto game_server_node/v1/discovery.proto \
  localhost:44044 \
  game_server_node.v1.DiscoveryService/Heartbeat
```

**Ожидаемое поведение:**
- Статус ответа: `OK`.
- Тело ответа содержит объект `usage` с полями: `cpuUsagePercent`, `memoryUsedBytes`, `diskUsedBytes`, `networkBytesPerSec`.
- Поле `activeInstanceCount` содержит целое число ≥ 0.

---

#### TC-03. Получение списка инстансов при пустом хранилище (ListInstances)

**Цель:** Проверить, что метод `ListInstances` корректно возвращает пустой список, когда ни один инстанс не запущен.

**Предусловия:** Микросервис только что запущен, инстансы не создавались.

**Команда:**

```bash
grpcurl -plaintext \
  -import-path ./protos \
  -proto game_server_node/v1/discovery.proto \
  localhost:44044 \
  game_server_node.v1.DiscoveryService/ListInstances
```

**Ожидаемое поведение:**
- Статус ответа: `OK`.
- Поле `instances` отсутствует либо содержит пустой массив.

---

#### TC-04. Получение инстанса по несуществующему ID (GetInstance)

**Цель:** Проверить, что метод `GetInstance` возвращает ошибку `NOT_FOUND` при обращении к несуществующему инстансу.

**Предусловия:** Микросервис запущен, инстанс с ID `999` не существует.

**Команда:**

```bash
grpcurl -plaintext \
  -import-path ./protos \
  -proto game_server_node/v1/discovery.proto \
  -d '{"instance_id": 999}' \
  localhost:44044 \
  game_server_node.v1.DiscoveryService/GetInstance
```

**Ожидаемое поведение:**
- Статус ответа: `NOT_FOUND` (код 5).
- Сообщение об ошибке содержит указание на отсутствие инстанса.

---

#### TC-05. Получение списка инстансов по ID игры при пустом хранилище (ListInstancesByGame)

**Цель:** Проверить, что метод `ListInstancesByGame` возвращает пустой список для игры, у которой нет запущенных инстансов.

**Предусловия:** Микросервис запущен, инстансы для `game_id = 42` не создавались.

**Команда:**

```bash
grpcurl -plaintext \
  -import-path ./protos \
  -proto game_server_node/v1/discovery.proto \
  -d '{"game_id": 42}' \
  localhost:44044 \
  game_server_node.v1.DiscoveryService/ListInstancesByGame
```

**Ожидаемое поведение:**
- Статус ответа: `OK`.
- Поле `instances` отсутствует либо содержит пустой массив.

---

#### TC-06. Получение утилизации ресурсов несуществующего инстанса (GetInstanceUsage)

**Цель:** Проверить, что метод `GetInstanceUsage` возвращает ошибку при запросе утилизации для несуществующего инстанса.

**Предусловия:** Микросервис запущен, инстанс с ID `999` не существует.

**Команда:**

```bash
grpcurl -plaintext \
  -import-path ./protos \
  -proto game_server_node/v1/discovery.proto \
  -d '{"instance_id": 999}' \
  localhost:44044 \
  game_server_node.v1.DiscoveryService/GetInstanceUsage
```

**Ожидаемое поведение:**
- Статус ответа: `NOT_FOUND` (код 5).

---

### 4.2. Сервис DeploymentService

#### TC-07. Загрузка Docker-образа (LoadImage)

**Цель:** Проверить, что метод `LoadImage` (client-streaming) корректно принимает tar-архив Docker-образа и регистрирует его в системе.

**Предусловия:**
- Микросервис запущен.
- Подготовлен файл `test-game-v1.tar` (см. п. 2.3).
- Docker daemon доступен из контейнера/хоста микросервиса.

**Примечание:** `grpcurl` поддерживает client-streaming через передачу нескольких JSON-объектов. Для передачи бинарных данных (chunks) необходимо закодировать содержимое tar-файла в Base64.

**Подготовка данных (bash):**

```bash
# Разбить tar-файл на Base64-чанки и сформировать JSON-поток
{
  echo '{"metadata": {"game_id": 1, "image_tag": "test-game:v1"}}'
  base64 test-game-v1.tar | fold -w 65536 | while IFS= read -r line; do
    echo "{\"chunk\": \"$line\"}"
  done
} > load_image_stream.json
```

**Подготовка данных (PowerShell):**

```powershell
$meta = '{"metadata": {"game_id": 1, "image_tag": "test-game:v1"}}'
$base64 = [Convert]::ToBase64String([IO.File]::ReadAllBytes("test-game-v1.tar"))
$chunks = $base64 -split "(.{65536})" | Where-Object { $_ -ne "" }
$lines = @($meta)
foreach ($chunk in $chunks) {
    $lines += "{`"chunk`": `"$chunk`"}"
}
$lines | Out-File -Encoding utf8 load_image_stream.json
```

**Команда:**

```bash
cat load_image_stream.json | grpcurl -plaintext \
  -import-path ./protos \
  -proto game_server_node/v1/deployment.proto \
  -d @ \
  localhost:44044 \
  game_server_node.v1.DeploymentService/LoadImage
```

**Ожидаемое поведение:**
- Статус ответа: `OK`.
- Тело ответа содержит поле `imageTag` со значением `"test-game:v1"`.

---

#### TC-08. Запуск инстанса с фиксированным портом (StartInstance)

**Цель:** Проверить, что метод `StartInstance` создаёт и запускает контейнер с заданными параметрами, возвращает ID инстанса и назначенный хост-порт.

**Предусловия:**
- Микросервис запущен.
- Образ `test-game:v1` загружен (TC-07 выполнен успешно).

**Команда:**

```bash
grpcurl -plaintext \
  -import-path ./protos \
  -proto game_server_node/v1/deployment.proto \
  -d '{
    "game_id": 1,
    "name": "test-lobby-1",
    "protocol": "PROTOCOL_TCP",
    "internal_port": 7777,
    "port_allocation": {"exact": 27015},
    "max_players": 16,
    "developer_payload": {"map": "de_dust2", "mode": "competitive"},
    "env_vars": {"GAME_MODE": "competitive"},
    "args": []
  }' \
  localhost:44044 \
  game_server_node.v1.DeploymentService/StartInstance
```

**Ожидаемое поведение:**
- Статус ответа: `OK`.
- Поле `instanceId` содержит положительное целое число.
- Поле `hostPort` равно `27015`.

---

#### TC-09. Запуск инстанса с автоматическим выбором порта (StartInstance — Any)

**Цель:** Проверить стратегию `any` для выделения порта.

**Предусловия:** Образ `test-game:v1` загружен.

**Команда:**

```bash
grpcurl -plaintext \
  -import-path ./protos \
  -proto game_server_node/v1/deployment.proto \
  -d '{
    "game_id": 1,
    "name": "test-lobby-auto",
    "protocol": "PROTOCOL_UDP",
    "internal_port": 7777,
    "port_allocation": {"any": true},
    "max_players": 8
  }' \
  localhost:44044 \
  game_server_node.v1.DeploymentService/StartInstance
```

**Ожидаемое поведение:**
- Статус ответа: `OK`.
- Поле `instanceId` содержит положительное целое число.
- Поле `hostPort` содержит значение порта (может быть `0`, если Docker назначает динамически).

---

#### TC-10. Запуск инстанса с диапазоном портов (StartInstance — Range)

**Цель:** Проверить стратегию `range` для выделения порта из заданного диапазона.

**Предусловия:** Образ `test-game:v1` загружен.

**Команда:**

```bash
grpcurl -plaintext \
  -import-path ./protos \
  -proto game_server_node/v1/deployment.proto \
  -d '{
    "game_id": 1,
    "name": "test-lobby-range",
    "protocol": "PROTOCOL_TCP",
    "internal_port": 7777,
    "port_allocation": {"range": {"min_port": 30000, "max_port": 30010}},
    "max_players": 32
  }' \
  localhost:44044 \
  game_server_node.v1.DeploymentService/StartInstance
```

**Ожидаемое поведение:**
- Статус ответа: `OK`.
- Поле `hostPort` содержит значение в диапазоне [30000, 30010].

---

#### TC-11. Запуск инстанса с ограничением ресурсов (StartInstance — ResourceLimits)

**Цель:** Проверить, что ограничения CPU и памяти корректно передаются в контейнер.

**Предусловия:** Образ `test-game:v1` загружен.

**Команда:**

```bash
grpcurl -plaintext \
  -import-path ./protos \
  -proto game_server_node/v1/deployment.proto \
  -d '{
    "game_id": 1,
    "name": "test-lobby-limited",
    "protocol": "PROTOCOL_TCP",
    "internal_port": 7777,
    "port_allocation": {"exact": 27020},
    "max_players": 4,
    "resource_limits": {
      "cpu_millis": 500,
      "memory_bytes": 134217728
    }
  }' \
  localhost:44044 \
  game_server_node.v1.DeploymentService/StartInstance
```

**Ожидаемое поведение:**
- Статус ответа: `OK`.
- Инстанс создан с ограничением 500 mCPU (0.5 ядра) и 128 МБ оперативной памяти.

---

#### TC-12. Запуск инстанса для незагруженного образа (StartInstance — ошибка)

**Цель:** Проверить, что попытка запуска инстанса для игры без предварительно загруженного образа приводит к ошибке.

**Предусловия:** Образ для `game_id = 9999` не загружался.

**Команда:**

```bash
grpcurl -plaintext \
  -import-path ./protos \
  -proto game_server_node/v1/deployment.proto \
  -d '{
    "game_id": 9999,
    "name": "no-image-test",
    "protocol": "PROTOCOL_TCP",
    "internal_port": 7777,
    "port_allocation": {"any": true},
    "max_players": 2
  }' \
  localhost:44044 \
  game_server_node.v1.DeploymentService/StartInstance
```

**Ожидаемое поведение:**
- Статус ответа: `INTERNAL` (код 13).
- Сообщение об ошибке указывает на отсутствие загруженного образа для данной игры.

---

#### TC-13. Проверка инстанса через Discovery после запуска (GetInstance)

**Цель:** Проверить, что созданный инстанс доступен через `DiscoveryService.GetInstance`.

**Предусловия:** TC-08 выполнен успешно, известен `instanceId` (например, `1`).

**Команда:**

```bash
grpcurl -plaintext \
  -import-path ./protos \
  -proto game_server_node/v1/discovery.proto \
  -d '{"instance_id": 1}' \
  localhost:44044 \
  game_server_node.v1.DiscoveryService/GetInstance
```

**Ожидаемое поведение:**
- Статус ответа: `OK`.
- Поле `instance.name` равно `"test-lobby-1"`.
- Поле `instance.status` равно `"INSTANCE_STATUS_RUNNING"`.
- Поле `instance.port` равно `27015`.
- Поле `instance.gameId` равно `"1"`.
- Поля `developerPayload` содержат переданные значения.

---

#### TC-14. Проверка списка инстансов после запуска (ListInstances)

**Цель:** Убедиться, что все запущенные инстансы отображаются в списке.

**Предусловия:** Один или более инстансов запущены (TC-08, TC-09, TC-10).

**Команда:**

```bash
grpcurl -plaintext \
  -import-path ./protos \
  -proto game_server_node/v1/discovery.proto \
  localhost:44044 \
  game_server_node.v1.DiscoveryService/ListInstances
```

**Ожидаемое поведение:**
- Статус ответа: `OK`.
- Массив `instances` содержит все ранее запущенные инстансы.

---

#### TC-15. Проверка списка инстансов по ID игры (ListInstancesByGame)

**Цель:** Проверить фильтрацию инстансов по `game_id`.

**Предусловия:** Запущены инстансы с `game_id = 1`.

**Команда:**

```bash
grpcurl -plaintext \
  -import-path ./protos \
  -proto game_server_node/v1/discovery.proto \
  -d '{"game_id": 1}' \
  localhost:44044 \
  game_server_node.v1.DiscoveryService/ListInstancesByGame
```

**Ожидаемое поведение:**
- Статус ответа: `OK`.
- Все возвращённые инстансы имеют `gameId` равный `"1"`.

---

#### TC-16. Получение утилизации ресурсов работающего инстанса (GetInstanceUsage)

**Цель:** Проверить, что метод возвращает данные об использовании ресурсов контейнером.

**Предусловия:** Инстанс с ID `1` запущен и находится в статусе `RUNNING`.

**Команда:**

```bash
grpcurl -plaintext \
  -import-path ./protos \
  -proto game_server_node/v1/discovery.proto \
  -d '{"instance_id": 1}' \
  localhost:44044 \
  game_server_node.v1.DiscoveryService/GetInstanceUsage
```

**Ожидаемое поведение:**
- Статус ответа: `OK`.
- Поле `instanceId` равно `"1"`.
- Объект `usage` содержит числовые значения утилизации (CPU, память, диск, сеть).

---

#### TC-17. Стриминг логов работающего инстанса (StreamLogs)

**Цель:** Проверить, что метод `StreamLogs` (server-streaming) возвращает поток лог-сообщений контейнера.

**Предусловия:** Инстанс с ID `1` запущен.

**Команда:**

```bash
grpcurl -plaintext \
  -import-path ./protos \
  -proto game_server_node/v1/deployment.proto \
  -d '{
    "instance_id": 1,
    "follow_stdout": true,
    "follow_stderr": false
  }' \
  localhost:44044 \
  game_server_node.v1.DeploymentService/StreamLogs
```

**Ожидаемое поведение:**
- Сервер начинает потоковую отправку сообщений типа `StreamLogsResponse`.
- Каждое сообщение содержит поля `timestamp`, `source` (`LOG_SOURCE_STDOUT`) и `message`.
- При нажатии `Ctrl+C` стрим корректно завершается.

---

#### TC-18. Стриминг логов без режима follow (StreamLogs — без follow)

**Цель:** Проверить получение уже существующих логов без ожидания новых.

**Предусловия:** Инстанс с ID `1` запущен, контейнер уже произвёл вывод в stdout.

**Команда:**

```bash
grpcurl -plaintext \
  -import-path ./protos \
  -proto game_server_node/v1/deployment.proto \
  -d '{
    "instance_id": 1,
    "follow_stdout": false,
    "follow_stderr": false
  }' \
  localhost:44044 \
  game_server_node.v1.DeploymentService/StreamLogs
```

**Ожидаемое поведение:**
- Сервер отправляет существующие логи и завершает стрим (EOF).
- Команда `grpcurl` завершается самостоятельно.

---

#### TC-19. Остановка инстанса (StopInstance)

**Цель:** Проверить, что метод `StopInstance` останавливает контейнер и обновляет статус инстанса.

**Предусловия:** Инстанс с ID `1` запущен и находится в статусе `RUNNING`.

**Команда:**

```bash
grpcurl -plaintext \
  -import-path ./protos \
  -proto game_server_node/v1/deployment.proto \
  -d '{
    "instance_id": 1,
    "timeout_seconds": 10
  }' \
  localhost:44044 \
  game_server_node.v1.DeploymentService/StopInstance
```

**Ожидаемое поведение:**
- Статус ответа: `OK`.
- Тело ответа — пустой объект `{}`.

---

#### TC-20. Проверка статуса инстанса после остановки (GetInstance)

**Цель:** Убедиться, что статус инстанса изменился на `STOPPED` после вызова `StopInstance`.

**Предусловия:** TC-19 выполнен успешно.

**Команда:**

```bash
grpcurl -plaintext \
  -import-path ./protos \
  -proto game_server_node/v1/discovery.proto \
  -d '{"instance_id": 1}' \
  localhost:44044 \
  game_server_node.v1.DiscoveryService/GetInstance
```

**Ожидаемое поведение:**
- Статус ответа: `OK`.
- Поле `instance.status` равно `"INSTANCE_STATUS_STOPPED"`.

---

#### TC-21. Остановка несуществующего инстанса (StopInstance — ошибка)

**Цель:** Проверить обработку ошибки при попытке остановить несуществующий инстанс.

**Предусловия:** Инстанс с ID `999` не существует.

**Команда:**

```bash
grpcurl -plaintext \
  -import-path ./protos \
  -proto game_server_node/v1/deployment.proto \
  -d '{
    "instance_id": 999,
    "timeout_seconds": 5
  }' \
  localhost:44044 \
  game_server_node.v1.DeploymentService/StopInstance
```

**Ожидаемое поведение:**
- Статус ответа: `NOT_FOUND` (код 5).

---

#### TC-22. Heartbeat после запуска инстансов

**Цель:** Проверить, что `activeInstanceCount` в heartbeat корректно отражает количество активных инстансов.

**Предусловия:** Запущен хотя бы один инстанс в статусе `RUNNING`.

**Команда:**

```bash
grpcurl -plaintext \
  -import-path ./protos \
  -proto game_server_node/v1/discovery.proto \
  localhost:44044 \
  game_server_node.v1.DiscoveryService/Heartbeat
```

**Ожидаемое поведение:**
- Статус ответа: `OK`.
- Значение `activeInstanceCount` больше нуля и соответствует количеству запущенных инстансов.

---

#### TC-23. Heartbeat после остановки всех инстансов

**Цель:** Проверить, что после остановки всех инстансов `activeInstanceCount` равно нулю.

**Предусловия:** Все ранее запущенные инстансы остановлены.

**Команда:**

```bash
grpcurl -plaintext \
  -import-path ./protos \
  -proto game_server_node/v1/discovery.proto \
  localhost:44044 \
  game_server_node.v1.DiscoveryService/Heartbeat
```

**Ожидаемое поведение:**
- Статус ответа: `OK`.
- Значение `activeInstanceCount` равно `0`.

---

## 5. Порядок выполнения тестов

Тестовые сценарии должны выполняться в следующей последовательности для соблюдения зависимостей между ними:

```
Фаза 1. Базовая проверка (чистое состояние):
  TC-01 → TC-02 → TC-03 → TC-04 → TC-05 → TC-06

Фаза 2. Загрузка образа:
  TC-07

Фаза 3. Запуск инстансов:
  TC-08 → TC-09 → TC-10 → TC-11 → TC-12

Фаза 4. Проверка состояния после запуска:
  TC-13 → TC-14 → TC-15 → TC-16 → TC-17 → TC-18 → TC-22

Фаза 5. Остановка и финальная проверка:
  TC-19 → TC-20 → TC-21 → TC-23
```

---

## 6. Критерии прохождения

Тест считается **пройденным**, если:

- gRPC-ответ содержит ожидаемый код статуса;
- структура ответа соответствует proto-определению;
- значения полей соответствуют ожидаемым (при наличии детерминированных данных).

Тест считается **не пройденным**, если:

- возвращён неожиданный код статуса;
- структура ответа не соответствует proto-определению;
- сервис вернул ошибку при корректных входных данных;
- сервис не вернул ошибку при некорректных входных данных.

---

## 7. Сводная таблица тестовых сценариев

| № | Метод | Описание | Тип проверки |
|---|---|---|---|
| TC-01 | `GetNodeInfo` | Информация об узле | Позитивный |
| TC-02 | `Heartbeat` | Утилизация ресурсов узла | Позитивный |
| TC-03 | `ListInstances` | Пустой список инстансов | Позитивный (граничный) |
| TC-04 | `GetInstance` | Несуществующий инстанс | Негативный |
| TC-05 | `ListInstancesByGame` | Пустой список по game_id | Позитивный (граничный) |
| TC-06 | `GetInstanceUsage` | Утилизация несуществующего инстанса | Негативный |
| TC-07 | `LoadImage` | Загрузка Docker-образа (streaming) | Позитивный |
| TC-08 | `StartInstance` | Запуск с фиксированным портом | Позитивный |
| TC-09 | `StartInstance` | Запуск с автовыбором порта | Позитивный |
| TC-10 | `StartInstance` | Запуск с диапазоном портов | Позитивный |
| TC-11 | `StartInstance` | Запуск с ограничением ресурсов | Позитивный |
| TC-12 | `StartInstance` | Запуск без загруженного образа | Негативный |
| TC-13 | `GetInstance` | Проверка созданного инстанса | Позитивный |
| TC-14 | `ListInstances` | Список после запуска | Позитивный |
| TC-15 | `ListInstancesByGame` | Фильтрация по game_id | Позитивный |
| TC-16 | `GetInstanceUsage` | Утилизация работающего инстанса | Позитивный |
| TC-17 | `StreamLogs` | Потоковые логи (follow) | Позитивный |
| TC-18 | `StreamLogs` | Логи без follow | Позитивный |
| TC-19 | `StopInstance` | Остановка инстанса | Позитивный |
| TC-20 | `GetInstance` | Проверка статуса после остановки | Позитивный |
| TC-21 | `StopInstance` | Остановка несуществующего инстанса | Негативный |
| TC-22 | `Heartbeat` | Счётчик активных после запуска | Позитивный |
| TC-23 | `Heartbeat` | Счётчик активных после остановки всех | Позитивный (граничный) |
