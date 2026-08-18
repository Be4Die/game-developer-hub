# syntax=docker/dockerfile:1

FROM golang:1.25.3-alpine AS builder

WORKDIR /app/game-deployment-agent

COPY protos/ /app/protos/
COPY game-deployment-agent/go.mod game-deployment-agent/go.sum* ./
RUN go mod tidy

COPY game-deployment-agent/ .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /game-deployment-agent ./cmd/server

FROM alpine:3.21

WORKDIR /app

RUN mkdir -p /data/games

COPY --from=builder /game-deployment-agent /app/game-deployment-agent
COPY game-deployment-agent/config/ /app/config/

ENV CONFIG_PATH=/app/config/docker.yaml

EXPOSE 50055

ENTRYPOINT ["/app/game-deployment-agent"]
