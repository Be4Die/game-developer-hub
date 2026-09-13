FROM golang:1.25.3-alpine AS builder

WORKDIR /app/ingress-proxy

COPY ingress-proxy/go.mod ingress-proxy/go.sum ./
RUN go mod download

COPY ingress-proxy/ .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /ingress-proxy ./cmd/server

FROM alpine:3.21

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /ingress-proxy /app/ingress-proxy

ENV PORT=8085

EXPOSE 8085

ENTRYPOINT ["/app/ingress-proxy"]
