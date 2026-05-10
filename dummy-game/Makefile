# Dummy Game — Makefile
# Быстрая сборка клиента, сервера и упаковка сервера в архив для загрузки на площадку.
#
# ВАЖНО: оркестратор ожидает архив (tar.gz или zip) с серверным билдом.
# game-server-node распакует архив, найдёт исполняемый файл, сгенерирует Dockerfile
# и соберёт Docker-образ командой `docker build`.
#
# Поэтому `make package-server` делает cross-compile для Linux (CGO_ENABLED=0)
# и упаковывает бинарник с executable bit в tar.gz.

BINARY_DIR       := bin
SERVER_BIN       := $(BINARY_DIR)/dummy-server
SERVER_BIN_LINUX := $(BINARY_DIR)/dummy-server-linux
CLIENT_BIN       := $(BINARY_DIR)/dummy-client
PACKAGE_NAME     := dummy-server.tar.gz
PACKAGE_ZIP      := dummy-server.zip

GO          := go
GOFLAGS     := -ldflags="-s -w"

# ═══════════════════════════════════════════════════════════════════════════════
# Build targets
# ═══════════════════════════════════════════════════════════════════════════════

.PHONY: all build build-server build-client build-server-linux package-server package-zip clean test run-server run-client help

## Собрать всё: клиент, сервер и упаковать сервер в tar.gz
all: build package-server

## Собрать бинарники под текущую платформу (клиент + сервер)
build: build-server build-client

## Собрать dummy-server под текущую платформу
build-server:
	@echo "==> Building dummy-server (current platform)..."
	@mkdir -p $(BINARY_DIR)
	$(GO) build $(GOFLAGS) -o $(SERVER_BIN) server.go
	@echo "    $(SERVER_BIN)"

## Собрать dummy-client
build-client:
	@echo "==> Building dummy-client..."
	@mkdir -p $(BINARY_DIR)
	$(GO) build $(GOFLAGS) -o $(CLIENT_BIN) client.go
	@echo "    $(CLIENT_BIN)"

## Собрать dummy-server для Linux (для Docker-контейнера, CGO_ENABLED=0)
build-server-linux:
	@echo "==> Building dummy-server for Linux (amd64, CGO_ENABLED=0)..."
	@mkdir -p $(BINARY_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO) build $(GOFLAGS) -o $(SERVER_BIN_LINUX) server.go
	@chmod +x $(SERVER_BIN_LINUX)
	@echo "    $(SERVER_BIN_LINUX)"

# ═══════════════════════════════════════════════════════════════════════════════
# Package — архивы для загрузки в оркестратор
# ═══════════════════════════════════════════════════════════════════════════════

## Упаковать сервер в tar.gz для оркестратора (Linux бинарник → server, + конфиг)
package-server: build-server-linux
	@echo "==> Packaging $(PACKAGE_NAME) for orchestrator..."
	@mkdir -p $(BINARY_DIR)/package
	@cp $(SERVER_BIN_LINUX) $(BINARY_DIR)/package/server
	@chmod +x $(BINARY_DIR)/package/server
	@cp config.yaml $(BINARY_DIR)/package/
	@echo "# Dummy Game Server Build Archive" > $(BINARY_DIR)/package/README.txt
	@echo "" >> $(BINARY_DIR)/package/README.txt
	@echo "This archive is uploaded to the orchestrator as a server build." >> $(BINARY_DIR)/package/README.txt
	@echo "game-server-node will extract it, find the executable, generate a Dockerfile," >> $(BINARY_DIR)/package/README.txt
	@echo "and build a Docker image automatically." >> $(BINARY_DIR)/package/README.txt
	@echo "" >> $(BINARY_DIR)/package/README.txt
	@echo "Upload via Gateway:" >> $(BINARY_DIR)/package/README.txt
	@echo "  POST /api/v1/games/{game_id}/builds" >> $(BINARY_DIR)/package/README.txt
	@echo "    - image: @$(PACKAGE_NAME)" >> $(BINARY_DIR)/package/README.txt
	@echo "    - build_version: 1.0.0" >> $(BINARY_DIR)/package/README.txt
	@echo "    - protocol: tcp" >> $(BINARY_DIR)/package/README.txt
	@echo "    - internal_port: 7777" >> $(BINARY_DIR)/package/README.txt
	@echo "    - max_players: 16" >> $(BINARY_DIR)/package/README.txt
	@echo "" >> $(BINARY_DIR)/package/README.txt
	@echo "Container environment variables:" >> $(BINARY_DIR)/package/README.txt
	@echo "  DUMMY_GAME_PORT, DUMMY_HTTP_PORT, DUMMY_MAX_PLAYERS, DUMMY_GAME_ID" >> $(BINARY_DIR)/package/README.txt
	@echo "  DUMMY_ORCH_ENABLED=true, DUMMY_ORCH_GATEWAY=http://gateway:8080" >> $(BINARY_DIR)/package/README.txt
	@# Создаём tar.gz с сохранением executable bit
	@tar -czf $(PACKAGE_NAME) -C $(BINARY_DIR)/package .
	@echo "    $(PACKAGE_NAME) (`ls -lh $(PACKAGE_NAME) | awk '{print \$$5}'`)"
	@rm -rf $(BINARY_DIR)/package

## Упаковать сервер в zip (альтернатива tar.gz, бинарник → server)
package-zip: build-server-linux
	@echo "==> Packaging $(PACKAGE_ZIP) for orchestrator..."
	@mkdir -p $(BINARY_DIR)/package
	@cp $(SERVER_BIN_LINUX) $(BINARY_DIR)/package/server
	@chmod +x $(BINARY_DIR)/package/server
	@cp config.yaml $(BINARY_DIR)/package/
	@cd $(BINARY_DIR)/package && zip -r ../../$(PACKAGE_ZIP) server config.yaml
	@echo "    $(PACKAGE_ZIP) ($$(ls -lh $(PACKAGE_ZIP) | awk '{print \$$5}'))"
	@rm -rf $(BINARY_DIR)/package

# ═══════════════════════════════════════════════════════════════════════════════
# Run
# ═══════════════════════════════════════════════════════════════════════════════

## Запустить сервер локально (go run)
run-server:
	$(GO) run server.go

## Запустить клиент с прямым подключением (без оркестратора)
run-client:
	$(GO) run client.go -direct localhost:7777 -player-id test-player

## Запустить клиент через оркестратор (discovery + queue)
run-client-orch:
	$(GO) run client.go -gateway http://localhost:8080 -game-id 1 -player-id test-player

# ═══════════════════════════════════════════════════════════════════════════════
# Test / Dev
# ═══════════════════════════════════════════════════════════════════════════════

## Проверить сборку и код
vet:
	$(GO) vet ./...

## Скачать зависимости
deps:
	$(GO) mod tidy
	$(GO) mod download

# ═══════════════════════════════════════════════════════════════════════════════
# Clean
# ═══════════════════════════════════════════════════════════════════════════════

## Удалить собранные артефакты
clean:
	@echo "==> Cleaning..."
	@rm -rf $(BINARY_DIR) $(PACKAGE_NAME) $(PACKAGE_ZIP)

# ═══════════════════════════════════════════════════════════════════════════════
# Help
# ═══════════════════════════════════════════════════════════════════════════════

## Показать эту справку
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'
