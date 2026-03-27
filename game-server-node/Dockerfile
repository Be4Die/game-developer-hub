FROM golang:1.25.3-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /node ./cmd/node


FROM alpine:3.21

WORKDIR /app

COPY --from=builder /node /app/node

COPY config/ /app/config/

ENV CONFIG_PATH=/app/config/local.yaml

EXPOSE 44044

ENTRYPOINT ["/app/node"]
