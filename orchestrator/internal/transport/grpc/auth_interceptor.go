package grpc

import (
	"context"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// contextKey — ключ для хранения данных пользователя в контексте.
type contextKey string

const (
	// UserIDKey — ключ для user_id в контексте.
	UserIDKey contextKey = "user_id"
	// SessionIDKey — ключ для session_id в контексте.
	SessionIDKey contextKey = "session_id"
	// UserRoleKey — ключ для role в контексте.
	UserRoleKey contextKey = "role"
)

// JWTAuth извлекает и проверяет JWT-токен из метаданных.
type JWTAuth struct {
	secret []byte
	issuer string
}

// NewJWTAuth создаёт новый интерцептор авторизации по JWT.
func NewJWTAuth(secret, issuer string) (*JWTAuth, error) {
	if secret == "" {
		return nil, status.Error(codes.Internal, "jwt secret is required")
	}
	if issuer == "" {
		return nil, status.Error(codes.Internal, "jwt issuer is required")
	}
	return &JWTAuth{
		secret: []byte(secret),
		issuer: issuer,
	}, nil
}

// Unary возвращает unary interceptor для проверки JWT.
func (a *JWTAuth) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if !isPublicMethod(info.FullMethod) {
			var err error
			ctx, err = a.authenticate(ctx)
			if err != nil {
				return nil, err
			}
		}
		return handler(ctx, req)
	}
}

// Stream возвращает stream interceptor для проверки JWT.
func (a *JWTAuth) Stream() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if !isPublicStream(info.FullMethod) {
			ctx, err := a.authenticate(ss.Context())
			if err != nil {
				return err
			}
			return handler(srv, &wrappedServerStream{ServerStream: ss, ctx: ctx})
		}
		return handler(srv, ss)
	}
}

func (a *JWTAuth) authenticate(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	values := md.Get("authorization")
	if len(values) == 0 || values[0] == "" {
		return nil, status.Error(codes.Unauthenticated, "missing authorization header")
	}

	tokenStr := strings.TrimPrefix(values[0], "Bearer ")
	if tokenStr == values[0] {
		return nil, status.Error(codes.Unauthenticated, "invalid authorization format, expected 'Bearer <token>'")
	}

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, status.Errorf(codes.Unauthenticated, "unexpected signing method: %v", t.Header["alg"])
		}
		return a.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "invalid token claims")
	}

	if iss, _ := claims.GetIssuer(); iss != a.issuer {
		return nil, status.Error(codes.Unauthenticated, "invalid token issuer")
	}

	userID, _ := claims.GetSubject()
	sessionID, _ := claims["sid"].(string)
	roleFloat, _ := claims["role"].(float64)

	ctx = context.WithValue(ctx, UserIDKey, userID)
	ctx = context.WithValue(ctx, SessionIDKey, sessionID)
	ctx = context.WithValue(ctx, UserRoleKey, uint8(roleFloat))

	return ctx, nil
}

// isPublicMethod возвращает true для методов, не требующих аутентификации.
func isPublicMethod(method string) bool {
	// DiscoverServers — публичный метод для game client.
	return method == "/orchestrator.v1.DiscoveryService/Discover"
}

// isPublicStream возвращает true для streaming методов без аутентификации.
func isPublicStream(_ string) bool {
	return false
}

// wrappedServerStream оборачивает ServerStream с аутентифицированным контекстом.
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

// GetUserID извлекает user_id из контекста.
func GetUserID(ctx context.Context) (string, bool) {
	uid, ok := ctx.Value(UserIDKey).(string)
	return uid, ok
}

// GetSessionID извлекает session_id из контекста.
func GetSessionID(ctx context.Context) (string, bool) {
	sid, ok := ctx.Value(SessionIDKey).(string)
	return sid, ok
}

// GetUserRole извлекает role из контекста.
func GetUserRole(ctx context.Context) (uint8, bool) {
	role, ok := ctx.Value(UserRoleKey).(uint8)
	return role, ok
}
