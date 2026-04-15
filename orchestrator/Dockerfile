# syntax=docker/dockerfile:1

FROM golang:1.25.3-alpine AS builder

WORKDIR /app/orchestrator

COPY protos/ /app/protos/
COPY orchestrator/go.mod orchestrator/go.sum ./
RUN go mod download

COPY orchestrator/ .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /orchestrator ./cmd/server


FROM alpine:3.21

RUN apk add --no-cache ca-certificates docker-cli

WORKDIR /app

COPY --from=builder /orchestrator /app/orchestrator
COPY orchestrator/config/ /app/config/

ENV CONFIG_PATH=/app/config/local.yaml

EXPOSE 8080

ENTRYPOINT ["/app/orchestrator"]
