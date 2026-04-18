FROM golang:1.25.3-alpine AS builder
COPY protos/ /protos/
COPY sso/go.mod sso/go.sum ./
RUN go mod download
COPY sso/ .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server ./cmd/server

FROM alpine:3.21
RUN apk add --no-cache postgresql-client
COPY --from=builder /server /app/server
COPY sso/config/ /app/config/
COPY sso/migrations/ /app/migrations/
ENV CONFIG_PATH=/app/config/local.yaml
EXPOSE 50051
ENTRYPOINT ["/app/server"]
