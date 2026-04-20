// Package grpc предоставляет gRPC обработчики для SSO-сервиса.
package grpc

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/Be4Die/game-developer-hub/sso/internal/domain"
)

// JWTAuthInterceptor — interceptor для проверки JWT токена.
type JWTAuthInterceptor struct {
	tokenManager domain.TokenManager
}

// NewJWTAuth создаёт interceptor для аутентификации по JWT.
func NewJWTAuth(tokenManager domain.TokenManager) *JWTAuthInterceptor {
	return &JWTAuthInterceptor{tokenManager: tokenManager}
}

// Unary возвращает unary interceptor.
func (j *JWTAuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Публичные методы не требуют JWT.
		if isPublicMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		// Извлекаем user_id из JWT и добавляем в контекст.
		newCtx, err := j.authenticate(ctx)
		if err != nil {
			return nil, err
		}

		return handler(newCtx, req)
	}
}

func (j *JWTAuthInterceptor) authenticate(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx, status.Errorf(codes.Unauthenticated, "missing metadata")
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return ctx, status.Errorf(codes.Unauthenticated, "missing authorization header")
	}

	authHeader := values[0]
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return ctx, status.Errorf(codes.Unauthenticated, "invalid authorization format")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	// Парсим JWT токен.
	claims, err := j.tokenManager.ParseAccessToken(ctx, token)
	if err != nil {
		return ctx, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	// Добавляем user_id в контекст.
	newMd := metadata.New(map[string]string{
		"x-user-id": claims.UserID,
	})
	return metadata.NewIncomingContext(ctx, newMd), nil
}
