# syntax=docker/dockerfile:1

# ─── Stage 1: Build ───
FROM golang:1.25.3-alpine AS builder

WORKDIR /build

# Копируем proto-контракты (зависимость go.mod replace).
COPY protos/ /protos/

# Копируем модуль game-server-node.
COPY game-server-node/go.mod game-server-node/go.sum ./
RUN go mod download

COPY game-server-node/ .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server ./cmd/server


# ─── Stage 2: Runtime ───
FROM alpine:3.21

RUN apk add --no-cache ca-certificates docker-cli

WORKDIR /app

COPY --from=builder /server /app/server
COPY game-server-node/config/ /app/config/

ENV CONFIG_PATH=/app/config/local.yaml

EXPOSE 44044

ENTRYPOINT ["/app/server"]
