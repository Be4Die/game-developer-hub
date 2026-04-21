package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"

	gwpb "github.com/Be4Die/game-developer-hub/protos/orchestrator/v1"
	ssopb "github.com/Be4Die/game-developer-hub/protos/sso/v1"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Адреса gRPC-сервисов из переменных окружения.
	orchestratorAddr := envOr("ORCHESTRATOR_GRPC_ADDR", "orchestrator:9090")
	ssoAddr := envOr("SSO_GRPC_ADDR", "sso:9090")
	httpAddr := envOr("HTTP_ADDR", ":8080")

	// Создаём mux с настройками JSON.
	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames:   true,
				EmitUnpopulated: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
	)

	// Опции для gRPC-соединений.
	dialOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	// Подключаемся к Orchestrator.
	orchConn, err := grpc.NewClient(orchestratorAddr, dialOpts...)
	if err != nil {
		return err
	}
	defer func() { _ = orchConn.Close() }()

	// Подключаемся к SSO.
	ssoConn, err := grpc.NewClient(ssoAddr, dialOpts...)
	if err != nil {
		return err
	}
	defer func() { _ = ssoConn.Close() }()

	// Регистрируем Orchestrator handlers.
	if err := gwpb.RegisterBuildServiceHandler(ctx, mux, orchConn); err != nil {
		return err
	}
	if err := gwpb.RegisterInstanceServiceHandler(ctx, mux, orchConn); err != nil {
		return err
	}
	if err := gwpb.RegisterNodeServiceHandler(ctx, mux, orchConn); err != nil {
		return err
	}
	if err := gwpb.RegisterHealthServiceHandler(ctx, mux, orchConn); err != nil {
		return err
	}
	if err := gwpb.RegisterDiscoveryServiceHandler(ctx, mux, orchConn); err != nil {
		return err
	}

	// Регистрируем SSO handlers.
	if err := ssopb.RegisterAuthServiceHandler(ctx, mux, ssoConn); err != nil {
		return err
	}
	if err := ssopb.RegisterUserServiceHandler(ctx, mux, ssoConn); err != nil {
		return err
	}

	// HTTP-сервер с CORS.
	handler := corsMiddleware(mux)

	srv := &http.Server{
		Addr:    httpAddr,
		Handler: handler,
	}

	// Graceful shutdown.
	go func() {
		<-signalChan()
		log.Println("shutting down HTTP server...")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	log.Printf("HTTP gateway listening on %s", httpAddr)
	log.Printf("  Orchestrator gRPC: %s", orchestratorAddr)
	log.Printf("  SSO gRPC: %s", ssoAddr)

	return srv.ListenAndServe()
}

func signalChan() <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		close(ch)
	}()
	return ch
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization,x-api-key")
			w.Header().Set("Access-Control-Max-Age", "86400")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
