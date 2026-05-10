# Queue Service — Архитектура

> Упрощённая FIFO client-side очередь для оркестратора игровых серверов.
> Без matchmaking, без game-server-side queue.

## Концепция

Когда все инстансы игры заполнены и `scale_behavior = queue`, клиент не получает ошибку `capacity_reached`, а встаёт в очередь. Система выдаёт слот при освобождении места (disconnect другого игрока) или при масштабировании.

## Принципы

| Аспект | Решение |
|--------|---------|
| Хранение активных очередей | **Valkey** — SortedSet (`score = join_time_ms`) + Hash для мета-данных |
| Персистентность | **PostgreSQL** — audit-лог (`queue_events`) |
| Транспорт клиента | **HTTP REST** — polling каждые 3–5 сек через `Heartbeat` |
| Идентификация игрока | `player_id` из query-param / header (доверие клиенту) |
| Авто-очистка | **Heartbeat TTL** — если клиент не пинговал > `queue_heartbeat_timeout`, его выкидывает из очереди |
| Reservation | Слот резервируется на `queue_reservation_seconds` (default 30s) |

## Данные (Valkey)

```
queue:{game_id}                         # SortedSet   score=timestamp_ms  member=player_id
queue_meta:{game_id}:{player_id}        # Hash        join_time, last_heartbeat, reserved_instance_id, reserved_until
reservation:{game_id}:{player_id}       # String      JSON: {instance_id, endpoint, expires_at}
```

## Состояния очереди

```
waiting   →  reserved  →  connected
   ↓            ↓
expired    expired (не подключился за reservation_timeout)
```

## Flow

```
1. Discovery → capacity_reached + scale_behavior=queue
2. Client → POST /queue/join (player_id)
   → добавлен в queue:{game_id}, queue_meta создан
3. Client polling: POST /queue/heartbeat каждые 3–5 сек
   → обновляет last_heartbeat в queue_meta
4. Освобождается слот (player disconnect / scale up / scale-to-zero отмена):
   a. HeartbeatService или DiscoveryService вызывает QueueService.ProcessQueue(game_id)
   b. Берём player_id с минимым score (FIFO)
   c. Резервируем: заполняем reservation + queue_meta.reserved_instance_id
   d. При следующем heartbeat → status=reserved + endpoint
5. Client подключается к endpoint
6. Client → Discovery с player_id → получает reserved endpoint
7. При подключении / отключении — игровой сервер может ReportPlayerStatus
8. Если heartbeat не приходил > TTL → авто-удаление из очереди
```

## API

### Client HTTP REST

```
POST   /api/v1/games/{game_id}/queue/join       player_id, mode (opt)
POST   /api/v1/games/{game_id}/queue/heartbeat  player_id
DELETE /api/v1/games/{game_id}/queue/leave      player_id  (опционально)
GET    /api/v1/games/{game_id}/queue/status     player_id  (read-only)
```

### Discovery интеграция

```
GET /api/v1/games/{game_id}/discover?player_id=xxx

Если у player_id есть active reservation:
  → status=reserved, endpoint в servers[0]
Иначе если capacity_reached + scale_behavior=queue:
  → status=queue (новый статус)
Иначе:
  → текущая логика
```

## Параметры политики (game_policies)

```sql
queue_reservation_seconds INTEGER DEFAULT 30   -- резервация слота
queue_max_wait_seconds    INTEGER DEFAULT 300  -- авто-отмена после 5 мин
queue_heartbeat_timeout   INTEGER DEFAULT 15   -- выкидывание при отсутствии heartbeat
```

## Отличия от полной версии

- ❌ Нет matchmaking / skill-based
- ❌ Нет game-server-side API
- ❌ Нет party/groups
- ✅ Heartbeat вместо обязательного leave
- ✅ Простая FIFO очередь
