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
		userID, role, err := a.extractUserID(ctx)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "unauthenticated: %v", err)
		}
		ctx = withUserContext(ctx, userID, role)
		return handler(ctx, req)
	}
}

// Stream возвращает grpc.StreamServerInterceptor.
func (a *JWTAuth) Stream() grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		userID, role, err := a.extractUserID(stream.Context())
		if err != nil {
			return status.Errorf(codes.Unauthenticated, "unauthenticated: %v", err)
		}
		ctx := withUserContext(stream.Context(), userID, role)
		wrapped := &wrappedStream{ServerStream: stream, ctx: ctx}
		return handler(srv, wrapped)
	}
}

func (a *JWTAuth) extractUserID(ctx context.Context) (string, float64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", 0, fmt.Errorf("missing metadata")
	}

	// 1. Проверяем явный заголовок x-user-id (проброшенный шлюзом)
	if vals := md.Get("x-user-id"); len(vals) > 0 && vals[0] != "" {
		var role float64
		if roleVals := md.Get("x-user-role"); len(roleVals) > 0 && roleVals[0] != "" {
			var roleInt int
			fmt.Sscanf(roleVals[0], "%d", &roleInt)
			role = float64(roleInt)
		}
		return vals[0], role, nil
	}

	// 2. Проверяем заголовок Authorization: Bearer <token>
	vals := md.Get("authorization")
	if len(vals) == 0 {
		return "", 0, fmt.Errorf("missing authorization header or x-user-id")
	}
	tokenStr := strings.TrimPrefix(vals[0], "Bearer ")
	if tokenStr == vals[0] {
		return "", 0, fmt.Errorf("invalid authorization format")
	}

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return a.secret, nil
	}, jwt.WithIssuer(a.issuer))
	if err != nil {
		return "", 0, fmt.Errorf("parse token: %w", err)
	}
	if !token.Valid {
		return "", 0, fmt.Errorf("token invalid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", 0, fmt.Errorf("invalid claims")
	}
	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return "", 0, fmt.Errorf("missing sub claim")
	}
	role, _ := claims["role"].(float64)
	return sub, role, nil
}

func withUserContext(ctx context.Context, userID string, role float64) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	} else {
		md = md.Copy()
	}
	md.Set("x-user-id", userID)
	md.Set("x-user-role", fmt.Sprintf("%v", role))
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

// UserRoleFromContext извлекает роль пользователя.
func UserRoleFromContext(ctx context.Context) (int, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, false
	}
	vals := md.Get("x-user-role")
	if len(vals) == 0 || vals[0] == "" {
		return 0, false
	}
	var role int
	fmt.Sscanf(vals[0], "%d", &role)
	return role, true
}

type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}
