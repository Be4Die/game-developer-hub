// Package grpc содержит gRPC-хендлеры и middleware.
package grpc

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// JWTAuth gRPC interceptor для JWT-аутентификации.
type JWTAuth struct {
	secret []byte
	issuer string
}

// NewJWTAuth создаёт JWT interceptor.
func NewJWTAuth(secret, issuer string) (*JWTAuth, error) {
	if secret == "" {
		return nil, fmt.Errorf("JWT secret is empty")
	}
	return &JWTAuth{secret: []byte(secret), issuer: issuer}, nil
}

// Unary возвращает grpc.UnaryServerInterceptor.
func (a *JWTAuth) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		userID, err := a.extractUserID(ctx)
		if err != nil {
			// Публичные методы можно добавить сюда при необходимости
			return nil, status.Errorf(codes.Unauthenticated, "unauthenticated: %v", err)
		}
		ctx = withUserID(ctx, userID)
		return handler(ctx, req)
	}
}

// Stream возвращает grpc.StreamServerInterceptor.
func (a *JWTAuth) Stream() grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		userID, err := a.extractUserID(stream.Context())
		if err != nil {
			return status.Errorf(codes.Unauthenticated, "unauthenticated: %v", err)
		}
		ctx := withUserID(stream.Context(), userID)
		wrapped := &wrappedStream{ServerStream: stream, ctx: ctx}
		return handler(srv, wrapped)
	}
}

func (a *JWTAuth) extractUserID(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("missing metadata")
	}
	vals := md.Get("authorization")
	if len(vals) == 0 {
		return "", fmt.Errorf("missing authorization header")
	}
	tokenStr := strings.TrimPrefix(vals[0], "Bearer ")
	if tokenStr == vals[0] {
		return "", fmt.Errorf("invalid authorization format")
	}

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return a.secret, nil
	}, jwt.WithIssuer(a.issuer))
	if err != nil {
		return "", fmt.Errorf("parse token: %w", err)
	}
	if !token.Valid {
		return "", fmt.Errorf("token invalid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid claims")
	}
	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return "", fmt.Errorf("missing sub claim")
	}
	return sub, nil
}

func withUserID(ctx context.Context, userID string) context.Context {
	return metadata.NewIncomingContext(ctx, metadata.MD{"x-user-id": {userID}})
}

// UserIDFromContext извлекает user_id из контекста.
func UserIDFromContext(ctx context.Context) (string, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", false
	}
	vals := md.Get("x-user-id")
	if len(vals) == 0 {
		return "", false
	}
	return vals[0], true
}

type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}
