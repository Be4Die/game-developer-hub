#!/usr/bin/env bash
# wait-for — ожидание доступности TCP-хоста или HTTP-эндпоинта.
# Использование:
#   wait-for host:port [timeout]
#   wait-for http://host:port/health [timeout]
set -euo pipefail

TARGET="${1:?Usage: wait-for host:port|http://... [timeout]}"
TIMEOUT="${2:-60}"

if [[ "$TARGET" == http://* ]]; then
  echo "Ожидание HTTP: $TARGET (таймаут ${TIMEOUT}s)"
  for i in $(seq 1 "$TIMEOUT"); do
    if curl -sf "$TARGET" >/dev/null 2>&1; then
      echo "  -> доступен"
      exit 0
    fi
    sleep 1
  done
else
  HOST="${TARGET%%:*}"
  PORT="${TARGET##*:}"
  echo "Ожидание TCP: $HOST:$PORT (таймаут ${TIMEOUT}s)"
  for i in $(seq 1 "$TIMEOUT"); do
    if bash -c "echo >/dev/tcp/$HOST/$PORT" 2>/dev/null; then
      echo "  -> доступен"
      exit 0
    fi
    sleep 1
  done
fi

echo "  -> ТАЙМАУТ: $TARGET не ответил за ${TIMEOUT}s" >&2
exit 1
