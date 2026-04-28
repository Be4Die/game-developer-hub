FROM golang:1.25.3-alpine AS builder

WORKDIR /app

# Копируем protos, так как они нужны для сборки (replace в go.mod)
COPY protos/ /protos/
COPY chat/go.mod chat/go.sum ./
RUN go mod download

COPY chat/ .

RUN CGO_ENABLED=0 GOOS=linux go build -o /chat-server ./cmd/server

FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /chat-server /app/chat-server

EXPOSE 9091

CMD ["/app/chat-server"]