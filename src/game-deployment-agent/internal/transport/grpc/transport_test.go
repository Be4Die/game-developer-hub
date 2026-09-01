package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestAPIKeyInterceptor(t *testing.T) {
	expectedKey := "secret-agent-key"
	unary, _ := APIKeyInterceptor(expectedKey)

	t.Run("valid x-api-key header", func(t *testing.T) {
		md := metadata.Pairs("x-api-key", expectedKey)
		ctx := metadata.NewIncomingContext(context.Background(), md)

		called := false
		resp, err := unary(ctx, "req", &grpc.UnaryServerInfo{}, func(ctx context.Context, req any) (any, error) {
			called = true
			return "success", nil
		})

		require.NoError(t, err)
		assert.True(t, called)
		assert.Equal(t, "success", resp)
	})

	t.Run("valid authorization bearer header", func(t *testing.T) {
		md := metadata.Pairs("authorization", "Bearer "+expectedKey)
		ctx := metadata.NewIncomingContext(context.Background(), md)

		called := false
		_, err := unary(ctx, "req", &grpc.UnaryServerInfo{}, func(ctx context.Context, req any) (any, error) {
			called = true
			return "success", nil
		})

		require.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("missing header", func(t *testing.T) {
		ctx := context.Background()
		_, err := unary(ctx, "req", &grpc.UnaryServerInfo{}, func(ctx context.Context, req any) (any, error) {
			return nil, nil
		})

		require.Error(t, err)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("invalid key", func(t *testing.T) {
		md := metadata.Pairs("x-api-key", "wrong-key")
		ctx := metadata.NewIncomingContext(context.Background(), md)

		_, err := unary(ctx, "req", &grpc.UnaryServerInfo{}, func(ctx context.Context, req any) (any, error) {
			return nil, nil
		})

		require.Error(t, err)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))
	})
}
