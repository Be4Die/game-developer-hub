FROM golang:1.25.3-alpine AS builder

WORKDIR /src

# Копируем go.mod/go.sum основного модуля protos.
COPY protos/go.mod protos/go.sum ./protos/

# Копируем go.mod/go.sum gateway.
COPY gateway/go.mod gateway/go.sum ./gateway/

RUN cd gateway && go mod download

# Копируем весь проект.
COPY protos/ ./protos/
COPY gateway/ ./gateway/

WORKDIR /src/gateway
RUN CGO_ENABLED=0 GOOS=linux go build -o /gateway .

# Финальный образ.
FROM alpine:3.21

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /gateway /gateway
COPY --from=builder /src/gateway/config/ ./config/

EXPOSE 8080

ENTRYPOINT ["/gateway"]
