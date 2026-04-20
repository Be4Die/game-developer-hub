// Package grpc предоставляет gRPC обработчики для SSO-сервиса.
package grpc

import (
	"context"
	"crypto/subtle"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// APIKeyAuthInterceptor — interceptor для проверки API ключа.
type APIKeyAuthInterceptor struct {
	apiKey string
}

// NewAPIKeyAuth создаёт interceptor для аутентификации по API ключу.
func NewAPIKeyAuth(apiKey string) *APIKeyAuthInterceptor {
	return &APIKeyAuthInterceptor{apiKey: apiKey}
}

// Unary возвращает unary interceptor.
func (a *APIKeyAuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Публичные методы не требуют API ключ.
		if isPublicMethod(info.FullMethod) {
			return handler(ctx, req)
		}
		if err := a.validate(ctx); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

// isPublicMethod возвращает true для методов, не требующих API ключа.
func isPublicMethod(method string) bool {
	// Auth endpoints — публичные методы (не требуют API key).
	return method == "/sso.v1.AuthService/Register" ||
		method == "/sso.v1.AuthService/Login" ||
		method == "/sso.v1.AuthService/RefreshToken" ||
		method == "/sso.v1.AuthService/Logout" ||
		method == "/sso.v1.AuthService/VerifyEmail" ||
		method == "/sso.v1.AuthService/ResendVerificationEmail" ||
		method == "/sso.v1.AuthService/RequestPasswordReset" ||
		method == "/sso.v1.AuthService/ResetPassword"
}

func (a *APIKeyAuthInterceptor) validate(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.Unauthenticated, "missing metadata")
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return status.Errorf(codes.Unauthenticated, "missing authorization header")
	}

	authHeader := values[0]
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return status.Errorf(codes.Unauthenticated, "invalid authorization format")
	}

	token := authHeader[7:]

	// Если токен похож на JWT (содержит 3 части, разделенные точками),
	// пропускаем проверку API key — JWT interceptor проверит его позже.
	if isJWT(token) {
		return nil
	}

	if subtle.ConstantTimeCompare([]byte(token), []byte(a.apiKey)) != 1 {
		return status.Errorf(codes.Unauthenticated, "invalid API key")
	}

	return nil
}

// isJWT проверяет, является ли токен JWT (имеет 3 части, разделенные точками).
func isJWT(token string) bool {
	parts := strings.Split(token, ".")
	return len(parts) == 3
}
