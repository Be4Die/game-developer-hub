FROM golang:1.25-alpine AS builder
COPY protos/ /protos/
COPY chat/go.mod chat/go.sum ./
RUN go mod download
COPY chat/ .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /chat-server ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /chat-server /app/server
COPY chat/config/ /app/config/

EXPOSE 50053

CMD ["/app/server"]
