package grpc

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// internalErrorMessage — текст, который клиент получает вместо деталей внутренней ошибки.
const internalErrorMessage = "internal error"

// ErrorSanitizer скрывает от клиентов детали внутренних ошибок (SQL, адреса, пути
// пакетов), сохраняя их в журнале. Ошибки с прикладными кодами (NotFound,
// PermissionDenied и т.д.) передаются без изменений.
type ErrorSanitizer struct {
	log *slog.Logger
}

// NewErrorSanitizer создаёт перехватчик, скрывающий детали внутренних ошибок.
func NewErrorSanitizer(log *slog.Logger) *ErrorSanitizer {
	return &ErrorSanitizer{log: log}
}

// Unary возвращает unary-перехватчик.
func (e *ErrorSanitizer) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		resp, err := handler(ctx, req)
		return resp, e.sanitize(info.FullMethod, err)
	}
}

// Stream возвращает stream-перехватчик.
func (e *ErrorSanitizer) Stream() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return e.sanitize(info.FullMethod, handler(srv, ss))
	}
}

func (e *ErrorSanitizer) sanitize(method string, err error) error {
	if err == nil {
		return nil
	}
	st, ok := status.FromError(err)
	if ok && st.Code() != codes.Internal && st.Code() != codes.Unknown {
		return err
	}
	e.log.Error("internal error", slog.String("method", method), slog.String("error", err.Error()))
	return status.Error(codes.Internal, internalErrorMessage)
}
