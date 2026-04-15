#!/usr/bin/env bash
# loadtest-orchestrator.sh — полный стек для нагрузочного тестирования orchestrator.
#
# Поднимает: PostgreSQL, Valkey, orchestrator, game-server-node.
# Ждёт готовности всех сервисов, сидит данные и запускает тест.
#
# Использование:
#   ./scripts/loadtest-orchestrator.sh            # стандартный (1000 rps, 30s)
#   ./scripts/loadtest-orchestrator.sh heavy      # усиленный (10000 rps, 60s)
#   ./scripts/loadtest-orchestrator.sh custom 500 10s  # кастомный (rps, duration)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
ORCH_DIR="$PROJECT_DIR/orchestrator"
COMPOSE_FILE="$ORCH_DIR/docker-compose.yaml"
ORCH_ADDR="${ORCH_ADDR:-http://localhost:8080}"
API_KEY="${API_KEY:-dev-api-key-for-local-testing}"

# Цвета
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

log()  { echo -e "${GREEN}[INFO]${NC} $*"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $*"; }
err()  { echo -e "${RED}[ERR]${NC} $*" >&2; }

# ─── Cleanup ────────────────────────────────────────────────
cleanup() {
    log "Остановка контейнеров..."
    cd "$ORCH_DIR" && docker compose down --remove-orphans 2>/dev/null || true
    docker stop game-server-node 2>/dev/null && docker rm game-server-node 2>/dev/null || true
}

# ─── Trap for cleanup on exit ──────────────────────────────
trap cleanup EXIT

# ─── Build images ──────────────────────────────────────────
log "Сборка game-server-node..."
cd "$PROJECT_DIR"
docker build -q -t game-server-node:latest -f game-server-node/Dockerfile . >/dev/null

log "Сборка orchestrator..."
docker build -q -t orchestrator:latest -f orchestrator/Dockerfile . >/dev/null

# ─── Start full stack via docker compose ───────────────────
log "Запуск PostgreSQL, Valkey, Orchestrator..."
cd "$ORCH_DIR"
docker compose up -d --force-recreate

# ─── Wait for PostgreSQL ───────────────────────────────────
log "Ожидание PostgreSQL..."
for i in $(seq 1 20); do
    if docker exec orchestrator-postgres pg_isready -U postgres >/dev/null 2>&1; then
        log "PostgreSQL готов"
        break
    fi
    [ "$i" -eq 20 ] && { err "PostgreSQL не запустился за 20 сек"; exit 1; }
    sleep 1
done

# ─── Wait for Valkey ───────────────────────────────────────
log "Ожидание Valkey..."
for i in $(seq 1 20); do
    if docker exec orchestrator-valkey valkey-cli ping 2>/dev/null | grep -q PONG; then
        log "Valkey готов"
        break
    fi
    [ "$i" -eq 20 ] && { err "Valkey не запустился за 20 сек"; exit 1; }
    sleep 1
done

# ─── Wait for Orchestrator ─────────────────────────────────
log "Ожидание Orchestrator..."
for i in $(seq 1 30); do
    if curl -sf "$ORCH_ADDR/health" >/dev/null 2>&1; then
        log "Orchestrator готов"
        break
    fi
    [ "$i" -eq 30 ] && { err "Orchestrator не запустился за 30 сек"; exit 1; }
    sleep 1
done

# ─── Start game-server-node ────────────────────────────────
log "Запуск game-server-node..."
docker run -d --name game-server-node \
    -p 44044:44044 \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -e CONFIG_PATH=/app/config/local.yaml \
    -e NODE_API_KEY="$API_KEY" \
    game-server-node:latest >/dev/null

log "Ожидание game-server-node..."
for i in $(seq 1 15); do
    if bash -c "echo >/dev/tcp/localhost/44044" 2>/dev/null; then
        log "game-server-node готов (port 44044)"
        break
    fi
    [ "$i" -eq 15 ] && { err "game-server-node не запустился за 15 сек"; docker logs game-server-node 2>&1 | tail -5; exit 1; }
    sleep 1
done

# ─── Seed data ─────────────────────────────────────────────
log "Подготовка тестовых данных..."

# Создаём таблицы (оркестратор не делает миграции автоматически)
docker exec orchestrator-postgres psql -U postgres -d orchestrator -c "
CREATE TABLE IF NOT EXISTS nodes (
    id            BIGSERIAL PRIMARY KEY,
    address       TEXT NOT NULL UNIQUE,
    token_hash    BYTEA NOT NULL,
    api_token     TEXT NOT NULL DEFAULT '',
    region        TEXT,
    status        SMALLINT NOT NULL DEFAULT 1,
    cpu_cores     INTEGER NOT NULL DEFAULT 0,
    total_memory  BIGINT NOT NULL DEFAULT 0,
    total_disk    BIGINT NOT NULL DEFAULT 0,
    agent_version TEXT NOT NULL DEFAULT '',
    last_ping_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at    TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS server_builds (
    id             BIGSERIAL PRIMARY KEY,
    game_id        BIGINT NOT NULL,
    uploaded_by    BIGINT NOT NULL DEFAULT 0,
    version        TEXT NOT NULL,
    image_tag      TEXT NOT NULL,
    protocol       SMALLINT NOT NULL,
    internal_port  INTEGER NOT NULL,
    max_players    INTEGER NOT NULL,
    file_url       TEXT NOT NULL,
    file_size      BIGINT NOT NULL,
    created_at     TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (game_id, version)
);

CREATE TABLE IF NOT EXISTS instances (
    id               BIGSERIAL PRIMARY KEY,
    node_id          BIGINT NOT NULL REFERENCES nodes(id),
    server_build_id  BIGINT NOT NULL REFERENCES server_builds(id),
    game_id          BIGINT NOT NULL,
    name             TEXT NOT NULL,
    build_version    TEXT NOT NULL,
    protocol         SMALLINT NOT NULL,
    host_port        INTEGER NOT NULL,
    internal_port    INTEGER NOT NULL,
    status           SMALLINT NOT NULL DEFAULT 1,
    max_players      INTEGER NOT NULL,
    developer_payload JSONB,
    server_address   TEXT NOT NULL,
    started_at       TIMESTAMP NOT NULL,
    created_at       TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMP NOT NULL DEFAULT NOW()
);
" 2>&1 || { err "Не удалось создать таблицы"; exit 1; }

# Вставляем данные
docker exec orchestrator-postgres psql -U postgres -d orchestrator -c "
-- Очистка
TRUNCATE instances, server_builds, nodes RESTART IDENTITY CASCADE;

-- 3 ноды (для discovery gRPC не нужен, только БД+KV)
INSERT INTO nodes (address, token_hash, api_token, region, status, cpu_cores, total_memory, total_disk, agent_version, last_ping_at)
VALUES
    ('loadtest-node-1:44044', E'\\\\x00', 'loadtest', 'loadtest', 1, 8, 17179869184, 536870912000, '0.0.1', NOW()),
    ('loadtest-node-2:44044', E'\\\\x00', 'loadtest', 'loadtest', 1, 4,  8589934592, 268435456000, '0.0.1', NOW()),
    ('loadtest-node-3:44044', E'\\\\x00', 'loadtest', 'loadtest', 1, 16,34359738368, 1073741824000,'0.0.1', NOW());

-- Серверные билды
INSERT INTO server_builds (game_id, version, image_tag, protocol, internal_port, max_players, file_url, file_size)
VALUES
    (1, '1.0.0', 'welwise/game-1:1.0.0', 1, 8080, 100, '/builds/game1.tar', 1024),
    (2, '1.0.0', 'welwise/game-2:1.0.0', 1, 8080, 50,  '/builds/game2.tar', 1024);

-- 5 инстансов для game=1 (status=2 = running)
INSERT INTO instances (node_id, server_build_id, game_id, name, build_version, protocol, host_port, internal_port, status, max_players, server_address, started_at)
SELECT
    n.id, sb.id, 1, 'inst-1-' || g, '1.0.0', 1, 7000 + g, 8080, 2, 100,
    n.address, NOW()
FROM nodes n, (SELECT id FROM server_builds WHERE game_id = 1 LIMIT 1) sb, generate_series(1, 5) g
WHERE n.id = 1
LIMIT 5;

-- 3 инстанса для game=2
INSERT INTO instances (node_id, server_build_id, game_id, name, build_version, protocol, host_port, internal_port, status, max_players, server_address, started_at)
SELECT
    n.id, sb.id, 2, 'inst-2-' || g, '1.0.0', 1, 8000 + g, 8080, 2, 50,
    n.address, NOW()
FROM nodes n, (SELECT id FROM server_builds WHERE game_id = 2 LIMIT 1) sb, generate_series(1, 3) g
WHERE n.id = 1
LIMIT 3;
" 2>&1 || { err "Не удалось подготовить данные"; exit 1; }

# Seed KV: player count (ключи: inst:pc:<instanceID>, instanceID = 1..8 из SERIAL)
docker exec orchestrator-valkey valkey-cli MSET \
    "inst:pc:1" "45" "inst:pc:2" "78" "inst:pc:3" "12" \
    "inst:pc:4" "90" "inst:pc:5" "33" \
    "inst:pc:6" "20" "inst:pc:7" "35" "inst:pc:8" "5" \
    2>/dev/null || warn "KV seed failed"

log "Данные подготовлены: 3 ноды, 2 билда, 8 инстансов"

# ─── Run load test ─────────────────────────────────────────
echo ""
echo -e "${CYAN}═══════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}  НАГРУЗОЧНОЕ ТЕСТИРОВАНИЕ${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════${NC}"
echo ""

case "${1:-}" in
    heavy)
        log "Запуск: 10000 rps, 60 секунд"
        go run ./cmd/load-test --heavy --addr "$ORCH_ADDR"
        ;;
    custom)
        rate="${2:-1000}"
        duration="${3:-30s}"
        log "Запуск: $rate rps, $duration"
        go run ./cmd/load-test --rate "$rate" --duration "$duration" --addr "$ORCH_ADDR"
        ;;
    *)
        log "Запуск: 1000 rps, 30 секунд"
        go run ./cmd/load-test --rate 1000 --duration 30s --addr "$ORCH_ADDR"
        ;;
esac

echo ""
log "Тест завершён"
