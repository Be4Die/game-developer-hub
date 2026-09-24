# AGENTS.md — Правила и инструкции для AI-агентов

Данный документ содержит обязательные соглашения и правила разработки в репозитории `game-developer-hub`.

---

## 1. Управление схемой базы данных (Database Schema Management)

> [!IMPORTANT]
> **ПРАВИЛО ЕДИНОГО `init.sql`**:
> В текущей фазе активной разработки проект существует локально.
> В каждом сервисе в директории `migrations/` должен находиться **строго один** файл: `init.sql`.
> **Не больше одного SQL-файла на сервис!**

### Правила работы со схемой:
1. **Запрещено создавать инкрементальные/нумерованные миграции**:
   - Никаких `02_shared_access.sql`, `03_is_online.sql`, `04_*.sql` и т.д.
   - Любые изменения таблиц, добавление новых колонок, индексов, триггеров или комментариев вносятся **напрямую в `src/<service>/migrations/init.sql`**.
2. **Переход на версионированные миграции**:
   - Система версионированных миграций (golang-migrate, goose и др.) будет внедрена **только при выводе проекта в продакшен**, когда потребуется обеспечивать zero-downtime миграции рабочей БД. До этого момента миграции только через `init.sql`.
3. **Расположение файлов схемы**:
   - Orchestrator: `src/orchestrator/migrations/init.sql`
   - SSO: `src/sso/migrations/init.sql`
   - Project Manager: `src/project-manager/migrations/init.sql`
   - Moderation: `src/moderation/migrations/init.sql`
4. **Сброс и инициализация БД**:
   - При изменении схемы базы данных локальный стек пересоздается через `Taskfile`:
     ```bash
     task stack:db:reset   # Дропает БД, пересоздает и накладывает все init.sql
     task stack:seed       # Наполняет тестовыми данными
     ```
   - Запуск только накатывания схем:
     ```bash
     task stack:db:init    # Применяет init.sql всех сервисов
     ```

---

## 2. Архитектура проекта

- **Микросервисы (Go + gRPC + Protobuf)**:
  - `src/sso` (порт 50051) — аутентификация, JWT-токены, пользователи.
  - `src/orchestrator` (порт 50052) — управление игровыми нодами, серверными билдами и инстансами.
  - `src/project-manager` (порт 50053) — проекты, черновики, веб-сборки, релизы, командный доступ.
  - `src/moderation` (порт 50054) — заявки на модерацию (публикация, сервера, обновления), аудит-снимки, тикет-чат.
  - `src/game-deployment-agent` (порт 50055) — агент развертывания веб-игр.
  - `src/gateway` (порт 8080) — единый API-шлюз (grpc-gateway + reverse-proxy HTTP/REST в gRPC).
  - `src/ingress-proxy` (порт 8085) — маршрутизация игрового трафика.
  - `src/frontend` — Vue 3 + TypeScript SPA.
- **Инфраструктура**:
  - PostgreSQL 17 (`gdh-postgres`, порт 5432, БД `orchestrator`).
  - Valkey 8 (`gdh-valkey`, порт 6379) — распределенные блокировки (redsync) и кэш.
  - SeaweedFS (`gdh-seaweedfs`, S3 API на 8333, WebDAV/Filer на 8888, Master на 9333).
  - Nginx (`gdh-nginx`, порты 80, 443).

---

## 3. Сборка и тестирование

- Управление задачами выполняется через `task` (`Taskfile.yml`):
  - `task backend:up` — запуск инфраструктуры и бэкенд-сервисов в Docker.
  - `task backend:restart` — перезапуск всех бэкенд-контейнеров.
  - `task test` — запуск всех юнит-тестов.
- При изменении Protobuf-контрактов в `src/protos/`:
  - `task proto:gen` — генерация Go и TypeScript структур.
- Правила написания тестов и коммитов зафиксированы в:
  - `docs/code/commits-rules.md` (Conventional Commits).
  - `docs/code/tests-rules.md` (структура тестов, mock-генерация, `testify`).
  - `docs/code/comments-rules.md` (Go-doc стандарты).
