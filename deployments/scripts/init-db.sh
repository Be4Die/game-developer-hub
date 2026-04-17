#!/usr/bin/env bash
# init-db — инициализация схемы БД оркестратора.
# Запускается после поднятия postgres.
# Использование: ./init-db [postgres-container] [sql-path]
set -euo pipefail

CONTAINER="${1:-gdh-postgres}"
SQL_PATH="${2:-orchestrator/migrations/init.sql}"

echo "Инициализация БД в контейнере $CONTAINER..."
docker exec -i "$CONTAINER" psql -U postgres -d orchestrator < "$SQL_PATH"
echo "  -> Схема инициализирована"
