#!/usr/bin/env bash
# seed — заполнить оркестратор тестовыми данными.
# Использование: ./seed.sh [orchestrator-addr] [nodes] [builds] [instances]
set -euo pipefail

ADDR="${1:-http://localhost:8080}"
NODES="${2:-2}"
BUILDS="${3:-2}"
INSTANCES="${4:-4}"

echo "Заполнение тестовыми данными: $ADDR"
echo "  nodes=$NODES builds=$BUILDS instances=$INSTANCES"

cd "$(dirname "$0")/../orchestrator"

SEED_NODES="$NODES" SEED_BUILDS="$BUILDS" SEED_INSTANCES="$INSTANCES" \
  go run ./cmd/seed

echo "  -> Данные загружены"
