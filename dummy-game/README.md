# Dummy Game — заглушка для тестирования оркестрации

Минимальный игровой сервер и клиент для end-to-end тестирования оркестратора.

## 📁 Файлы

| Файл | Назначение |
|------|-----------|
| `server.go` | Dummy игровой сервер (TCP + HTTP admin API) |
| `client.go` | CLI клиент (discovery / queue / direct connect) |
| `config.yaml` | Пример конфигурации сервера |
| `Makefile` | Сборка, упаковка, запуск |

## 🚀 Быстрый старт

### Локальный запуск (без оркестратора)

```bash
# Сервер
make run-server

# Клиент (прямое подключение)
make run-client
```

### Загрузка в оркестратор

Оркестратор ожидает **архив с серверным билдом** (`tar.gz` или `zip`).  
`game-server-node` распакует архив, найдёт исполняемый файл, сгенерирует `Dockerfile` и соберёт Docker-образ.

```bash
make package-server
```

Создаст `dummy-server.tar.gz` (Linux бинарник, cross-compiled с `CGO_ENABLED=0`).

Загрузка через Gateway:
```bash
curl -X POST http://localhost:8080/api/v1/games/1/builds \
  -H "Authorization: Bearer <JWT>" \
  -F "image=@dummy-server.tar.gz" \
  -F "build_version=1.0.0" \
  -F "protocol=tcp" \
  -F "internal_port=7777" \
  -F "max_players=16"
```

Запуск инстанса:
```bash
curl -X POST http://localhost:8080/api/v1/games/1/instances \
  -H "Authorization: Bearer <JWT>" \
  -H "Content-Type: application/json" \
  -d '{"build_version":"1.0.0","max_players":16}'
```

## 🎮 Клиент через оркестратор

```bash
make run-client-orch
# или
./bin/dummy-client -gateway http://localhost:8080 -game-id 1 -player-id alice
```

Клиент выполняет полный flow:
1. **Discovery** — запрашивает доступные серверы
2. **Queue** — если серверы заняты, встаёт в очередь
3. **Connect** — подключается по TCP к игровому серверу
4. **Gameplay** — интерактивный CLI (`ping`, `list`, `chat`, `leave`)

## 🖥️ TCP Протокол игры

JSON over TCP, каждая строка — одно сообщение (newline-delimited):

```json
{"cmd":"join","name":"alice"}
{"cmd":"list"}
{"cmd":"chat","message":"hello"}
{"cmd":"ping"}
{"cmd":"leave"}
```

## 🔧 HTTP Admin API (сервер)

| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/health` | Health check |
| GET | `/status` | Статус сервера (player_count, max_players, uptime) |
| POST | `/admin/set-players` | Принудительно установить player_count |
| POST | `/admin/set-max-players` | Изменить max_players |

## 🛠️ Makefile команды

```bash
make build-server          # Собрать сервер под текущую платформу
make build-client          # Собрать клиент
make build-server-linux    # Cross-compile для Linux (для Docker)
make package-server        # Создать dummy-server.tar.gz для оркестратора
make package-zip           # Альтернатива — .zip
make all                   # build + package-server
make clean                 # Удалить артефакты
make help                  # Справка
```

## ⚙️ Переменные окружения сервера

| Переменная | Описание |
|------------|----------|
| `DUMMY_GAME_PORT` | TCP порт игры (default: 7777) |
| `DUMMY_HTTP_PORT` | HTTP admin порт (default: 7778) |
| `DUMMY_MAX_PLAYERS` | Макс. игроков (default: 16) |
| `DUMMY_GAME_ID` | ID игры (default: 1) |
| `DUMMY_ORCH_ENABLED` | Включить heartbeat к оркестратору (`true`/`1`) |
| `DUMMY_ORCH_GATEWAY` | URL gateway (default: http://localhost:8080) |
