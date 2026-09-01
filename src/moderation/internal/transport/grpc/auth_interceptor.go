// Package grpc provides gRPC transport and middleware for moderation service.
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
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		userID, role, err := a.extractUserInfo(ctx)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "unauthenticated: %v", err)
		}
		ctx = withUserInfo(ctx, userID, role)
		return handler(ctx, req)
	}
}

// Stream возвращает grpc.StreamServerInterceptor.
func (a *JWTAuth) Stream() grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		userID, role, err := a.extractUserInfo(stream.Context())
		if err != nil {
			return status.Errorf(codes.Unauthenticated, "unauthenticated: %v", err)
		}
		ctx := withUserInfo(stream.Context(), userID, role)
		wrapped := &wrappedStream{ServerStream: stream, ctx: ctx}
		return handler(srv, wrapped)
	}
}

func (a *JWTAuth) extractUserInfo(ctx context.Context) (string, string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", "", fmt.Errorf("missing metadata")
	}

	// 1. Проверяем явные заголовки x-user-id и x-user-role (проброшенные Gateway)
	if vals := md.Get("x-user-id"); len(vals) > 0 && vals[0] != "" {
		role := "developer"
		if rVals := md.Get("x-user-role"); len(rVals) > 0 && rVals[0] != "" {
			role = rVals[0]
		}
		return vals[0], role, nil
	}

	// 2. Проверяем заголовок Authorization: Bearer <token>
	vals := md.Get("authorization")
	if len(vals) == 0 {
		return "", "", fmt.Errorf("missing authorization header or x-user-id")
	}
	tokenStr := strings.TrimPrefix(vals[0], "Bearer ")
	if tokenStr == vals[0] {
		return "", "", fmt.Errorf("invalid authorization format")
	}

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return a.secret, nil
	}, jwt.WithIssuer(a.issuer))
	if err != nil {
		return "", "", fmt.Errorf("parse token: %w", err)
	}
	if !token.Valid {
		return "", "", fmt.Errorf("token invalid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", fmt.Errorf("invalid claims")
	}
	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return "", "", fmt.Errorf("missing sub claim")
	}

	role := "developer"
	switch v := claims["role"].(type) {
	case string:
		role = v
	case float64:
		switch v {
		case 2:
			role = "moderator"
		case 3:
			role = "admin"
		}
	}

	return sub, role, nil
}

func withUserInfo(ctx context.Context, userID, role string) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	} else {
		md = md.Copy()
	}
	md.Set("x-user-id", userID)
	md.Set("x-user-role", role)
	return metadata.NewIncomingContext(ctx, md)
}

// UserIDFromContext извлекает user_id из контекста gRPC запроса.
func UserIDFromContext(ctx context.Context) (string, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", false
	}
	vals := md.Get("x-user-id")
	if len(vals) == 0 || vals[0] == "" {
		return "", false
	}
	return vals[0], true
}

// UserRoleFromContext извлекает user_role из контекста gRPC запроса.
func UserRoleFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "developer"
	}
	vals := md.Get("x-user-role")
	if len(vals) == 0 || vals[0] == "" {
		return "developer"
	}
	if vals[0] == "2" {
		return "moderator"
	}
	if vals[0] == "3" {
		return "admin"
	}
	return vals[0]
}

type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}
