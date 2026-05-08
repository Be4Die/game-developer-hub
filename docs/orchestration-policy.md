# Политика оркестрации серверов проекта

> Документ фиксирует договорённости по внедрению автоматической оркестрации серверов на уровне проекта (игры).
> Основан на требованиях ТЗ (подсистема оркестрации серверов) и обсуждениях команды.

## Концепция

Вместо ручного управления каждым инстансом у проекта есть **единый набор правил** — политика оркестрации. Система на основе этих правил сама решает, когда запускать, останавливать или перезапускать серверные экземпляры.

Правила хранятся в подсистеме `orchestrator` и привязаны к `game_id` (bare int64, без отдельного сервиса проектов).

## Структура политики

| Поле | Тип | Описание |
|------|-----|----------|
| `game_id` | int64 | Идентификатор проекта (PK) |
| `mode` | enum | `disabled` / `keep_alive` / `scale_to_zero` |
| `target_instances` | int | Сколько инстансов держать запущенными (актуально для `keep_alive`) |
| `auto_restart` | bool | Перезапускать ли инстанс при статусе `crashed` |
| `scale_to_zero_timeout` | int | Минут без игроков до остановки (только `scale_to_zero`) |
| `default_build_version` | string | Версия билда для авто-старта (`latest` или конкретная) |
| `max_players_per_instance` | int | Лимит игроков на один инстанс. При достижении — триггер масштабирования |
| `max_instances_per_game` | int | Потолок инстансов для этого проекта (в рамках глобального лимита разработчика) |
| `scale_behavior` | enum | `spawn` — поднять новый инстанс при переполнении; `queue` — игроки ждут в очереди |
| `node_preference` | string | `auto` (любая нода) или `node-<id>` (конкретная авторизованная нода разработчика) |

## Режимы работы

### `disabled`
- Система не вмешивается в жизненный цикл инстансов.
- Все операции только вручную через UI.
- Используется для отладки и тестирования.

### `keep_alive`
- Система поддерживает запущенными ровно `target_instances` инстансов.
- Если инстанс падает (`crashed`) и `auto_restart = true` — перезапускает.
- Если разработчик остановил инстанс вручную — система **не** поднимает его обратно (чтобы не конфликтовать с пользователем).
- Discovery всегда отдаёт рабочие инстансы.

### `scale_to_zero`
- Экономичный режим: инстансы работают только когда есть спрос.
- При Discovery-запросе, если нет ни одного `Running` инстанса, система автоматически поднимает один (с `default_build_version`).
- Если на инстансе 0 игроков в течение `scale_to_zero_timeout` минут — система останавливает лишние, оставляя `target_instances`.
- При переполнении (`player_count >= max_players_per_instance`) и `scale_behavior = spawn` — поднимается дополнительный инстанс до `max_instances_per_game`.

## Поведение системы по событиям

| Событие | `disabled` | `keep_alive` | `scale_to_zero` |
|---------|------------|--------------|-----------------|
| Инстанс `crashed` | Ничего | Перезапустить (если `auto_restart`) | Перезапустить (если `auto_restart`) |
| Инстанс остановлен вручную | Ничего | Не трогать | Не трогать |
| Все инстансы остановлены | Ничего | Поднять `target_instances` | Ждать Discovery-запроса |
| Discovery, нет `Running` | Ошибка «нет серверов» | Поднять инстанс (если вдруг нет) | Поднять инстанс |
| Инстанс заполнен (`max_players`) | Ничего | Поднять ещё один (до `max_instances`) | Поднять ещё один (до `max_instances`) |
| 0 игроков на инстансе X мин | Ничего | Ничего | Остановить лишние (оставить `target_instances`) |

## UI / UX

Размещение: вкладка **Сервера → Обзор** (`ServersOverview.vue`).

Блок **«Политика оркестрации»** — **inline**, **collapsible** (свернуть/развернуть). По умолчанию свёрнут, если политика не менялась.

Элементы формы:
- Переключатель режима (`disabled` / `keep_alive` / `scale_to_zero`)
- `target_instances` — числовое поле (min 0)
- `auto_restart` — чекбокс
- `scale_to_zero_timeout` — числовое поле (активно только при `scale_to_zero`)
- `default_build_version` — выпадающий список из загруженных сборок проекта + опция `latest`
- `max_players_per_instance` — числовое поле
- `max_instances_per_game` — числовое поле
- `scale_behavior` — радио-кнопки `spawn` / `queue`
- `node_preference` — радио `auto` + выпадающий список авторизованных нод разработчика

Кнопка **«Сохранить политику»**.

## План реализации (краткий)

| # | Задача | Слой |
|---|--------|------|
| 1 | Таблица `game_policies` в миграциях `orchestrator` | БД |
| 2 | Доменная модель `GamePolicy` + репозиторий | `orchestrator/internal/domain/`, `storage/postgres/` |
| 3 | CRUD-сервис политики | `orchestrator/internal/service/` |
| 4 | gRPC/HTTP API для политики | `orchestrator/internal/transport/grpc/`, `protos/orchestrator/v1/` |
| 5 | Интеграция в Discovery: авто-старт при отсутствии `Running` | `orchestrator/internal/service/discovery.go` |
| 6 | Интеграция в Heartbeat: авто-рестарт / scale-to-zero после `reconcileInstances` | `orchestrator/internal/service/heartbeat.go` |
| 7 | UI: карточка политики в `ServersOverview.vue` | Frontend |
| 8 | API-клиент политики | `frontend/src/api/orchestrator.js` |

## Связь с ТЗ

- **ТЗ 4.2.4, функция №2** — выбор режима работы (ручной / автоматический). Реализуется через `mode`.
- **ТЗ 4.2.4, функция №3** — запуск в автоматическом режиме. Реализуется через `keep_alive` + `target_instances`.
- **ТЗ 4.4.4** — аварийный перезапуск. Реализуется через `auto_restart`.
- **Тест-кейс СЕРВЕР-002** — автоматический режим с перезапуском при сбое. Покрывается `keep_alive` или `scale_to_zero` + `auto_restart = true`.

## Примеры использования

### Сессионная игра (лобби)
```yaml
mode: scale_to_zero
target_instances: 0
auto_restart: true
scale_to_zero_timeout: 10
default_build_version: "latest"
max_players_per_instance: 100
max_instances_per_game: 2
scale_behavior: spawn
node_preference: auto
```

### Постоянный мир (MMO-like)
```yaml
mode: keep_alive
target_instances: 1
auto_restart: true
max_players_per_instance: 500
max_instances_per_game: 1
node_preference: "node-5"
```

### Отладка новой сборки
```yaml
mode: disabled
```
