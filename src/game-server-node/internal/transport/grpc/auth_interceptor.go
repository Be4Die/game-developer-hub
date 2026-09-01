// Package grpc реализует обработчики и middleware для gRPC-сервера.
package grpc

import (
	"context"
	"crypto/subtle"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const authMetadataKey = "authorization"

// APIKeyAuthInterceptor проверяет наличие валидного API-ключа в metadata запроса.
type APIKeyAuthInterceptor struct {
	apiKey string
}

// NewAPIKeyAuth создаёт interceptor для проверки API-ключа.
// Требует непустой ключ. Безопасен для конкурентного использования.
func NewAPIKeyAuth(apiKey string) *APIKeyAuthInterceptor {
	return &APIKeyAuthInterceptor{apiKey: apiKey}
}

// Unary возвращает gRPC-обработчик, который проверяет API-ключ в metadata.
// Ожидает ключ в формате "Bearer <key>" в поле "authorization".
// Возвращает Unauthenticated при отсутствии или невалидном ключе.
func (a *APIKeyAuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		_ *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if err := a.validate(ctx); err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

// Stream возвращает stream interceptor для проверки API-ключа.
func (a *APIKeyAuthInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		_ *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		if err := a.validate(ss.Context()); err != nil {
			return err
		}

		return handler(srv, ss)
	}
}

// validate извлекает и проверяет API-ключ из контекста.
func (a *APIKeyAuthInterceptor) validate(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "missing metadata")
	}

	values := md.Get(authMetadataKey)
	if len(values) == 0 {
		return status.Error(codes.Unauthenticated, "missing authorization header")
	}

	authHeader := values[0]
	expectedPrefix := "Bearer "
	if len(authHeader) < len(expectedPrefix) {
		return status.Error(codes.Unauthenticated, "invalid authorization format")
	}

	token := authHeader[len(expectedPrefix):]

	// ConstantTimeCompare защищает от timing-атаки.
	if subtle.ConstantTimeCompare([]byte(token), []byte(a.apiKey)) != 1 {
		return status.Error(codes.Unauthenticated, "invalid API key")
	}

	return nil
}
