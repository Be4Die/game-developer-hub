// Package grpc provides gRPC transport handlers and interceptors for project-manager.
package grpc

import (
	"context"
	"crypto/subtle"
	"fmt"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Ключи метаданных, которые формирует только сам интерсептор.
// Значения, присланные клиентом под этими ключами, отбрасываются до обработки запроса.
const (
	mdUserID       = "x-user-id"
	mdUserRole     = "x-user-role"
	mdUserEmail    = "x-user-email"
	mdUserName     = "x-user-name"
	mdInternalCall = "x-internal-call"
	// mdInternalToken — общий секрет межсервисных вызовов (INTERNAL_API_KEY).
	mdInternalToken = "x-internal-token"

	maxUserNameRunes = 100
)

// JWTAuth gRPC interceptor для JWT-аутентификации.
//
// Личность и роль пользователя берутся ТОЛЬКО из проверенного JWT. Заголовки x-user-id / x-user-role
// из входящего запроса не являются доверенными и перезаписываются значениями из токена.
type JWTAuth struct {
	secret        []byte
	issuer        string
	internalToken []byte
}

// NewJWTAuth создаёт JWT interceptor.
func NewJWTAuth(secret, issuer string) (*JWTAuth, error) {
	if secret == "" {
		return nil, fmt.Errorf("JWT secret is empty")
	}
	return &JWTAuth{secret: []byte(secret), issuer: issuer}, nil
}

// WithInternalToken задаёт общий секрет межсервисных вызовов.
// Запрос с совпадающим x-internal-token помечается как внутренний (см. IsInternalCall).
func (a *JWTAuth) WithInternalToken(token string) *JWTAuth {
	a.internalToken = []byte(token)
	return a
}

// identity — данные пользователя, извлечённые из проверенного токена.
type identity struct {
	userID   string
	role     float64
	email    string
	name     string
	internal bool
}

func (a *JWTAuth) authenticate(ctx context.Context) (identity, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return identity{}, fmt.Errorf("missing metadata")
	}

	vals := md.Get("authorization")
	if len(vals) == 0 {
		return identity{}, fmt.Errorf("missing authorization header")
	}
	tokenStr := strings.TrimPrefix(vals[0], "Bearer ")
	if tokenStr == vals[0] {
		return identity{}, fmt.Errorf("invalid authorization format")
	}

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return a.secret, nil
	}, jwt.WithIssuer(a.issuer), jwt.WithExpirationRequired())
	if err != nil {
		return identity{}, fmt.Errorf("parse token: %w", err)
	}
	if !token.Valid {
		return identity{}, fmt.Errorf("token invalid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return identity{}, fmt.Errorf("invalid claims")
	}
	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return identity{}, fmt.Errorf("missing sub claim")
	}

	id := identity{userID: sub}
	id.role, _ = claims["role"].(float64)
	id.email, _ = claims["email"].(string)

	// Отображаемое имя не влияет на права доступа: берётся из токена, а при его отсутствии — из запроса.
	if claimName, ok := claims["display_name"].(string); ok && claimName != "" {
		id.name = claimName
	} else if nameVals := md.Get(mdUserName); len(nameVals) > 0 {
		id.name = truncateRunes(strings.TrimSpace(nameVals[0]), maxUserNameRunes)
	}

	if len(a.internalToken) > 0 {
		if tokVals := md.Get(mdInternalToken); len(tokVals) > 0 &&
			subtle.ConstantTimeCompare([]byte(tokVals[0]), a.internalToken) == 1 {
			id.internal = true
		}
	}

	return id, nil
}

func truncateRunes(s string, limit int) string {
	r := []rune(s)
	if len(r) > limit {
		return string(r[:limit])
	}
	return s
}

func (a *JWTAuth) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		id, err := a.authenticate(ctx)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "unauthenticated: %v", err)
		}
		return handler(withUserContext(ctx, id), req)
	}
}

func (a *JWTAuth) Stream() grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		id, err := a.authenticate(stream.Context())
		if err != nil {
			return status.Errorf(codes.Unauthenticated, "unauthenticated: %v", err)
		}
		wrapped := &wrappedStream{ServerStream: stream, ctx: withUserContext(stream.Context(), id)}
		return handler(srv, wrapped)
	}
}

func withUserContext(ctx context.Context, id identity) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	} else {
		md = md.Copy()
	}
	// Сначала удаляем всё, что мог прислать клиент под служебными ключами.
	for _, key := range []string{mdUserID, mdUserRole, mdUserEmail, mdUserName, mdInternalCall} {
		md.Delete(key)
	}
	md.Set(mdUserID, id.userID)
	md.Set(mdUserRole, fmt.Sprintf("%v", id.role))
	if id.email != "" {
		md.Set(mdUserEmail, id.email)
	}
	if id.name != "" {
		md.Set(mdUserName, id.name)
	}
	if id.internal {
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

// UserEmailFromContext извлекает email пользователя.
func UserEmailFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	vals := md.Get("x-user-email")
	if len(vals) > 0 {
		return vals[0]
	}
	return ""
}

// UserNameFromContext извлекает display_name пользователя.
func UserNameFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	vals := md.Get("x-user-name")
	if len(vals) > 0 {
		return vals[0]
	}
	return ""
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
	r := strings.ToLower(vals[0])
	switch r {
	case "admin", "user_role_admin", "3":
		return 3, true
	case "moderator", "user_role_moderator", "2":
		return 2, true
	case "developer", "user_role_developer", "1":
		return 1, true
	default:
		role, err := strconv.Atoi(vals[0])
		if err != nil {
			return 0, false
		}
		return role, true
	}
}

type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}
