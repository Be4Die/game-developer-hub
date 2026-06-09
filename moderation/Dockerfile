FROM golang:1.25.3-alpine AS builder
COPY protos/ /protos/
COPY moderation/go.mod moderation/go.sum ./
RUN go mod download
COPY moderation/ .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server ./cmd/server

FROM alpine:3.21
RUN apk add --no-cache postgresql-client ca-certificates
COPY --from=builder /server /app/server
COPY moderation/config/ /app/config/
ENV CONFIG_PATH=/app/config/docker.yaml
EXPOSE 50053
ENTRYPOINT ["/app/server"]
