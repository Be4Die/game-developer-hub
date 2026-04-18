package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// APIKeyAuth извлекает и проверяет API-ключ из метаданных.
type APIKeyAuth struct {
	apiKey string
}

// NewAPIKeyAuth создаёт новый интерцептор авторизации по API-ключу.
func NewAPIKeyAuth(apiKey string) *APIKeyAuth {
	return &APIKeyAuth{apiKey: apiKey}
}

// Unary возвращает unary интерцептор для проверки API-ключа.
func (a *APIKeyAuth) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if err := a.authenticate(ctx); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

func (a *APIKeyAuth) authenticate(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "missing metadata")
	}

	values := md.Get("x-api-key")
	if len(values) == 0 || values[0] == "" {
		return status.Error(codes.Unauthenticated, "missing api key")
	}

	if values[0] != a.apiKey {
		return status.Error(codes.Unauthenticated, "invalid api key")
	}

	return nil
}
