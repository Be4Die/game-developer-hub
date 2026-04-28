package transport

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	ssopb "github.com/Be4Die/game-developer-hub/protos/sso/v1"
)

func AuthInterceptor(ssoClient ssopb.TokenServiceClient) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if isPublicMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
		}

		authHeader := md.Get("authorization")
		if len(authHeader) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "missing authorization header")
		}

		token := strings.TrimPrefix(authHeader[0], "Bearer ")
		if token == authHeader[0] {
			return nil, status.Errorf(codes.Unauthenticated, "invalid authorization format")
		}

		userID, err := validateToken(ctx, ssoClient, token)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}

		ctx = context.WithValue(ctx, userIDKey{}, userID)
		return handler(ctx, req)
	}
}

func isPublicMethod(method string) bool {
	// Можно добавить публичные методы, если они появятся
	return false
}

type userIDKey struct{}

func UserIDFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(userIDKey{})
	if v == nil {
		return "", false
	}
	return v.(string), true
}

func validateToken(ctx context.Context, ssoClient ssopb.TokenServiceClient, token string) (string, error) {
	resp, err := ssoClient.ValidateToken(ctx, &ssopb.ValidateTokenRequest{
		AccessToken: token,
	})
	if err != nil {
		return "", err
	}

	if !resp.Valid {
		return "", status.Error(codes.Unauthenticated, "token is invalid")
	}

	return resp.User.Id, nil
}
