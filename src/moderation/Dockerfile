# syntax=docker/dockerfile:1

FROM golang:1.25.3-alpine AS builder

WORKDIR /app/moderation

COPY protos/ /app/protos/
COPY moderation/go.mod moderation/go.sum ./
RUN go mod download

COPY moderation/ .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /moderation ./cmd/server

FROM alpine:3.21

WORKDIR /app

COPY --from=builder /moderation /app/moderation
COPY moderation/config/ /app/config/

ENV CONFIG_PATH=/app/config/docker.yaml

EXPOSE 50054

ENTRYPOINT ["/app/moderation"]
