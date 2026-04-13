FROM golang:1.25.3-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY protos ../protos
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /orchestrator ./cmd/server


FROM alpine:3.21

RUN apk add --no-cache ca-certificates docker-cli

WORKDIR /app

COPY --from=builder /orchestrator /app/orchestrator
COPY config/ /app/config/

ENV CONFIG_PATH=/app/config/local.yaml

EXPOSE 8080

ENTRYPOINT ["/app/orchestrator"]
