# Развёртывание в продуктивной среде

Продуктивный стек описан в [`deployments/docker-compose.prod.yml`](../../deployments/docker-compose.prod.yml),
переменные — в [`deployments/.env.prod.example`](../../deployments/.env.prod.example).
Локальный стек (`docker-compose.backend.yml`) для продуктива не подходит: в нём открыты порты
баз данных, отладочные секреты и тестовые учётные записи.

## Схема

```
              ┌──────────────── nginx площадки (TLS, certbot) ────────────────┐
              │  devhub.example.com            games-devhub.example.com       │
              └───────────────┬─────────────────────────────┬─────────────────┘
                              │ сеть PLATFORM_NETWORK       │
                              ▼                             ▼
                        ┌──────────── gdh-nginx:80 ────────────┐
                        │ кабинет: SPA, /api, /media           │ игры: /games, /game-proxy
                        └───┬───────────────────────────┬──────┘
                            ▼                           ▼
                         gateway ── sso, project-manager, moderation, orchestrator
                                         │                         │
                                  postgres, valkey, seaweedfs (только внутренняя сеть)
```

* Игры — недоверенный код разработчиков, поэтому они отдаются **только** с отдельного домена
  (`GDH_GAMES_HOST`). project-manager не стартует в `prod`, если `DEPLOYMENT_URL_PREFIX`
  не является абсолютным `https://` адресом.
* Наружу из стека опубликованы только nginx (`GDH_HTTP_BIND`, по умолчанию localhost) и gRPC
  оркестратора для агентов нод (`ORCHESTRATOR_BIND`).
* Анонимно из S3 читаются только бакеты `games` и `media`. Сборки и вложения чата отдаёт шлюз
  после проверки токена и доступа к проекту.

## Режим `ENV=prod`

Во всех сервисах продуктивный режим работает по принципу fail-closed: сервис не запускается, если

* `JWT_SECRET`, `INTERNAL_API_KEY`, `SSO_API_KEY` короче 32 символов или похожи на отладочные (`dev-`, `local-testing`);
* пароль БД пустой или `postgres`, не заданы ключи S3;
* почта в режиме `stub` (коды подтверждения писались бы в лог) или SMTP без TLS;
* пароль администратора короче 12 символов;
* покупки или аналитика в режиме `stub` (допустимы `http` или `disabled`);
* не задан адрес project-manager у оркестратора (проверка владельца игры) или `INGRESS_PROXY_HOST` равен `localhost`;
* шлюз без `CORS_ALLOWED_ORIGINS`.

## Первый запуск

```bash
cd deployments
cp .env.prod.example .env.prod && chmod 600 .env.prod
# заполнить домены и секреты: openssl rand -hex 32
docker network create welwise-games-network   # если сети площадки ещё нет
docker compose -f docker-compose.prod.yml --env-file .env.prod up -d --build
docker compose -f docker-compose.prod.yml --env-file .env.prod ps
```

Затем в nginx площадки добавить два server-блока по образцу
[`deployments/prod/platform-nginx.conf.example`](../../deployments/prod/platform-nginx.conf.example),
выпустить сертификаты для обоих доменов и добавить их в HTTP→HTTPS редирект.

Проверка без внешнего прокси:

```bash
curl -H "Host: $GDH_APP_HOST" http://127.0.0.1:8088/            # SPA
curl -H "Host: $GDH_GAMES_HOST" http://127.0.0.1:8088/games/1/prod/
```

## Схема БД

Схема накатывается из `src/*/migrations/init.sql` только при первом старте с пустым томом
`pgdata`. Тестовые пользователи (`deployments/scripts/seed-dev-users.sql`) в продуктив не попадают,
администратор создаётся sso из `ADMIN_EMAIL`/`ADMIN_PASSWORD`.

Изменения схемы после запуска требуют версионированных миграций — до их внедрения любое
изменение `init.sql` нужно применять к рабочей БД вручную.

## Игровые ноды

`game-server-node` и `game-deployment-agent` разворачиваются отдельно на серверах нод
(`deployments/docker-compose.agents.yml`) с `ENV=prod` и уникальными `NODE_API_KEY`/`AGENT_API_KEY`.
Оркестратор должен быть доступен нодам по частной сети или VPN.

Игровой сервер отправляет отчёты об онлайне на `GAME_SERVER_NODE_REPORT_URL` с заголовком
`Authorization: Bearer $GAME_SERVER_NODE_REPORT_TOKEN`; обе переменные нода передаёт контейнеру
при запуске. Отчёты без токена своего инстанса отклоняются.

## Резервное копирование

Состояние хранится в томах `pgdata` (все сервисы), `seaweed_data` (сборки, игры, медиа, вложения)
и `valkeydata` (сессии и коды подтверждения; потеря не критична). Минимально:

```bash
docker compose -f docker-compose.prod.yml --env-file .env.prod exec -T postgres \
  pg_dump -U postgres -d orchestrator -Fc > gdh-$(date +%F).dump
```
