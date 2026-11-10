// Package grpc provides gRPC transport and middleware for moderation service.
package grpc

import (
	"context"
	"crypto/subtle"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	roleDeveloper = "developer"
	roleModerator = "moderator"
	roleAdmin     = "admin"

	mdInternalCall = "x-internal-call"
	// mdInternalToken — общий секрет межсервисных вызовов (INTERNAL_API_KEY).
	mdInternalToken = "x-internal-token"
)

// JWTAuth gRPC interceptor для JWT-аутентификации.
type JWTAuth struct {
	secret        []byte
	issuer        string
	internalToken []byte
}

// WithInternalToken задаёт общий секрет межсервисных вызовов.
// Запрос с совпадающим x-internal-token помечается как внутренний (см. IsInternalCall).
func (a *JWTAuth) WithInternalToken(token string) *JWTAuth {
	a.internalToken = []byte(token)
	return a
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
		userID, role, internal, err := a.authenticate(ctx)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "unauthenticated: %v", err)
		}
		return handler(withUserInfo(ctx, userID, role, internal), req)
	}
}

// Stream возвращает grpc.StreamServerInterceptor.
func (a *JWTAuth) Stream() grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		userID, role, internal, err := a.authenticate(stream.Context())
		if err != nil {
			return status.Errorf(codes.Unauthenticated, "unauthenticated: %v", err)
		}
		wrapped := &wrappedStream{ServerStream: stream, ctx: withUserInfo(stream.Context(), userID, role, internal)}
		return handler(srv, wrapped)
	}
}

// authenticate проверяет подпись JWT и возвращает личность и роль пользователя из токена.
// Заголовки x-user-id / x-user-role входящего запроса доверенными не считаются.
func (a *JWTAuth) authenticate(ctx context.Context) (userID, role string, internal bool, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", "", false, fmt.Errorf("missing metadata")
	}

	vals := md.Get("authorization")
	if len(vals) == 0 {
		return "", "", false, fmt.Errorf("missing authorization header")
	}
	tokenStr := strings.TrimPrefix(vals[0], "Bearer ")
	if tokenStr == vals[0] {
		return "", "", false, fmt.Errorf("invalid authorization format")
	}

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return a.secret, nil
	}, jwt.WithIssuer(a.issuer), jwt.WithExpirationRequired())
	if err != nil {
		return "", "", false, fmt.Errorf("parse token: %w", err)
	}
	if !token.Valid {
		return "", "", false, fmt.Errorf("token invalid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", false, fmt.Errorf("invalid claims")
	}
	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return "", "", false, fmt.Errorf("missing sub claim")
	}

	role = roleDeveloper
	if v, ok := claims["role"].(float64); ok {
		switch v {
		case 2:
			role = roleModerator
		case 3:
			role = roleAdmin
		}
	}

	if len(a.internalToken) > 0 {
		if tokVals := md.Get(mdInternalToken); len(tokVals) > 0 &&
			subtle.ConstantTimeCompare([]byte(tokVals[0]), a.internalToken) == 1 {
			internal = true
		}
	}

	return sub, role, internal, nil
}

func withUserInfo(ctx context.Context, userID, role string, internal bool) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	} else {
		md = md.Copy()
	}
	// Сначала удаляем всё, что мог прислать клиент под служебными ключами.
	md.Delete(mdInternalCall)
	md.Set("x-user-id", userID)
	md.Set("x-user-role", role)
	if internal {
		md.Set(mdInternalCall, "1")
	}
	return metadata.NewIncomingContext(ctx, md)
}

// IsInternalCall сообщает, пришёл ли запрос от другого сервиса платформы (с верным x-internal-token).
func IsInternalCall(ctx context.Context) bool {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return false
	}
	vals := md.Get(mdInternalCall)
	return len(vals) > 0 && vals[0] == "1"
}

// IsStaff сообщает, является ли вызывающий модератором или администратором.
func IsStaff(ctx context.Context) bool {
	role := UserRoleFromContext(ctx)
	return role == roleModerator || role == roleAdmin
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
