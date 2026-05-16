# syntax=docker/dockerfile:1

FROM golang:1.25.3-alpine AS builder

WORKDIR /app/project-manager

COPY protos/ /app/protos/
COPY project-manager/go.mod project-manager/go.sum ./
RUN go mod download

COPY project-manager/ .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /project-manager ./cmd/server


FROM alpine:3.21

WORKDIR /app

RUN mkdir -p /data/projects

COPY --from=builder /project-manager /app/project-manager
COPY project-manager/config/ /app/config/

ENV CONFIG_PATH=/app/config/local.yaml

EXPOSE 50053

ENTRYPOINT ["/app/project-manager"]
