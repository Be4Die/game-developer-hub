package grpc_test

import (
	"context"
	"testing"

	grpcpkg "github.com/Be4Die/game-developer-hub/game-server-node/internal/transport/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const testAPIKey = "test-api-key-12345"

// mockHandler всегда возвращает успех для проверки прохождения авторизации.
func mockHandler(
	ctx context.Context,
	req interface{},
) (interface{}, error) {
	return &struct{}{}, nil
}

func TestUnaryInterceptor_ValidAPIKey(t *testing.T) {
	// Arrange
	interceptor := grpcpkg.NewAPIKeyAuth(testAPIKey)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		"authorization", "Bearer "+testAPIKey,
	))

	// Act
	_, err := interceptor.Unary()(ctx, nil, &grpc.UnaryServerInfo{}, mockHandler)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestUnaryInterceptor_MissingMetadata(t *testing.T) {
	// Arrange
	interceptor := grpcpkg.NewAPIKeyAuth(testAPIKey)
	ctx := context.Background() // Без metadata

	// Act
	_, err := interceptor.Unary()(ctx, nil, &grpc.UnaryServerInfo{}, mockHandler)

	// Assert
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}
	if st.Code() != codes.Unauthenticated {
		t.Errorf("expected code Unauthenticated, got %s", st.Code())
	}
}

func TestUnaryInterceptor_MissingAuthorizationHeader(t *testing.T) {
	// Arrange
	interceptor := grpcpkg.NewAPIKeyAuth(testAPIKey)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		"other-key", "some-value",
	))

	// Act
	_, err := interceptor.Unary()(ctx, nil, &grpc.UnaryServerInfo{}, mockHandler)

	// Assert
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}
	if st.Code() != codes.Unauthenticated {
		t.Errorf("expected code Unauthenticated, got %s", st.Code())
	}
}

func TestUnaryInterceptor_InvalidAPIKey(t *testing.T) {
	// Arrange
	interceptor := grpcpkg.NewAPIKeyAuth(testAPIKey)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		"authorization", "Bearer wrong-key",
	))

	// Act
	_, err := interceptor.Unary()(ctx, nil, &grpc.UnaryServerInfo{}, mockHandler)

	// Assert
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}
	if st.Code() != codes.Unauthenticated {
		t.Errorf("expected code Unauthenticated, got %s", st.Code())
	}
}

func TestUnaryInterceptor_InvalidFormat(t *testing.T) {
	// Arrange
	interceptor := grpcpkg.NewAPIKeyAuth(testAPIKey)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		"authorization", "invalid-format",
	))

	// Act
	_, err := interceptor.Unary()(ctx, nil, &grpc.UnaryServerInfo{}, mockHandler)

	// Assert
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}
	if st.Code() != codes.Unauthenticated {
		t.Errorf("expected code Unauthenticated, got %s", st.Code())
	}
}

func TestUnaryInterceptor_EmptyAPIKey(t *testing.T) {
	// Arrange — interceptor с пустым ключом (должен отклонять все запросы)
	interceptor := grpcpkg.NewAPIKeyAuth("")
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		"authorization", "Bearer "+testAPIKey,
	))

	// Act
	_, err := interceptor.Unary()(ctx, nil, &grpc.UnaryServerInfo{}, mockHandler)

	// Assert
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}
	if st.Code() != codes.Unauthenticated {
		t.Errorf("expected code Unauthenticated, got %s", st.Code())
	}
}
