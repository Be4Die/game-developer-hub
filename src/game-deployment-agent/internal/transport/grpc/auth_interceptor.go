// Package grpc provides gRPC handlers and interceptors for game-deployment-agent.
package grpc

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// APIKeyInterceptor проверяет заголовок x-api-key для всех входящих RPC вызовов.
func APIKeyInterceptor(expectedKey string) (grpc.UnaryServerInterceptor, grpc.StreamServerInterceptor) {
	unary := func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if err := validateKey(ctx, expectedKey); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}

	stream := func(srv any, ss grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if err := validateKey(ss.Context(), expectedKey); err != nil {
			return err
		}
		return handler(srv, ss)
	}

	return unary, stream
}

func validateKey(ctx context.Context, expectedKey string) error {
	if expectedKey == "" {
		return nil
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "metadata is not provided")
	}

	keys := md.Get("x-api-key")
	if len(keys) == 0 {
		// Fallback to authorization header Bearer <key>
		auths := md.Get("authorization")
		if len(auths) > 0 {
			token := strings.TrimPrefix(auths[0], "Bearer ")
			if token == expectedKey {
				return nil
			}
		}
		return status.Error(codes.Unauthenticated, "x-api-key header missing")
	}

	if keys[0] != expectedKey {
		return status.Error(codes.PermissionDenied, "invalid api key")
	}

	return nil
}
