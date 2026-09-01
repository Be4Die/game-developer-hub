package main

import (
	"context"
	"errors"
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

	modpb "github.com/Be4Die/game-developer-hub/protos/moderation/v1"
	gwpb "github.com/Be4Die/game-developer-hub/protos/orchestrator/v1"
	projpb "github.com/Be4Die/game-developer-hub/protos/project_manager/v1"
	ssopb "github.com/Be4Die/game-developer-hub/protos/sso/v1"
)

const maxMsgSize = 2 * 1024 * 1024 * 1024 // 2GB

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("gateway error: %v", err)
	}
}

func run(ctx context.Context) error {
	orchAddr := envOr("ORCHESTRATOR_GRPC_ADDR", "orchestrator:9090")
	ssoAddr := envOr("SSO_GRPC_ADDR", "sso:9090")
	projAddr := envOr("PROJECT_MANAGER_GRPC_ADDR", "project-manager:50053")
	modAddr := envOr("MODERATION_GRPC_ADDR", "moderation:50054")
	httpAddr := envOr("HTTP_ADDR", ":8080")
	projectsBasePath := envOr("PROJECTS_DATA_PATH", "/data/projects")

	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(maxMsgSize),
			grpc.MaxCallSendMsgSize(maxMsgSize),
		),
	}

	orchConn, err := grpc.NewClient(orchAddr, dialOpts...)
	if err != nil {
		return err
	}
	defer func() { _ = orchConn.Close() }()

	ssoConn, err := grpc.NewClient(ssoAddr, dialOpts...)
	if err != nil {
		return err
	}
	defer func() { _ = ssoConn.Close() }()

	projConn, err := grpc.NewClient(projAddr, dialOpts...)
	if err != nil {
		return err
	}
	defer func() { _ = projConn.Close() }()

	modConn, err := grpc.NewClient(modAddr, dialOpts...)
	if err != nil {
		return err
	}
	defer func() { _ = modConn.Close() }()

	gwMux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames:   true,
				EmitUnpopulated: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
		runtime.WithMetadata(jwtMetadataAnnotator),
	)

	// Регистрация gRPC-Gateway сервисов
	for _, reg := range []func(context.Context, *runtime.ServeMux, *grpc.ClientConn) error{
		gwpb.RegisterBuildServiceHandler,
		gwpb.RegisterInstanceServiceHandler,
		gwpb.RegisterNodeServiceHandler,
		gwpb.RegisterHealthServiceHandler,
		gwpb.RegisterDiscoveryServiceHandler,
		gwpb.RegisterGamePolicyServiceHandler,
	} {
		if err := reg(ctx, gwMux, orchConn); err != nil {
			return err
		}
	}
	if err := ssopb.RegisterAuthServiceHandler(ctx, gwMux, ssoConn); err != nil {
		return err
	}
	if err := ssopb.RegisterUserServiceHandler(ctx, gwMux, ssoConn); err != nil {
		return err
	}
	if err := projpb.RegisterProjectServiceHandler(ctx, gwMux, projConn); err != nil {
		return err
	}
	if err := modpb.RegisterModerationServiceHandler(ctx, gwMux, modConn); err != nil {
		return err
	}

	// Корневой маршрутизатор
	mux := http.NewServeMux()

	// Кастомные хэндлеры: стриминг загрузок, SSE логи, статика
	mux.HandleFunc("POST /api/v1/games/{game_id}/builds", handleBuildUpload(gwpb.NewBuildServiceClient(orchConn), gwMux))
	mux.HandleFunc("POST /api/v1/projects/{project_id}/builds", handleProjectBuildUpload(projpb.NewProjectServiceClient(projConn), gwMux))
	mux.HandleFunc("POST /api/v1/projects/{project_id}/media", handleProjectMediaUpload(projpb.NewProjectServiceClient(projConn), gwMux))
	mux.HandleFunc("GET /api/v1/games/{game_id}/instances/{instance_id}/logs", handleInstanceLogsStream(gwpb.NewInstanceServiceClient(orchConn)))
	mux.HandleFunc("GET /api/v1/projects/{project_id}/builds/{version}/download", handleProjectBuildDownload(projectsBasePath))
	mux.HandleFunc("GET /api/v1/projects/{project_id}/media/{type}", handleProjectMediaServe(projectsBasePath))
	mux.HandleFunc("GET /api/v1/media/{path...}", handleProjectMediaServe(projectsBasePath))
	mux.HandleFunc("GET /media/{path...}", handleProjectMediaServe(projectsBasePath))

	// Все остальные запросы проксируются в gRPC-Gateway
	mux.Handle("/", gwMux)

	server := &http.Server{
		Addr:         httpAddr,
		Handler:      corsMiddleware(mux),
		ReadTimeout:  600 * time.Second,
		WriteTimeout: 600 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		<-ctx.Done()
		log.Println("shutting down HTTP server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	log.Printf("HTTP gateway listening on %s", httpAddr)
	log.Printf("  Orchestrator gRPC: %s", orchAddr)
	log.Printf("  SSO gRPC: %s", ssoAddr)
	log.Printf("  Project Manager gRPC: %s", projAddr)
	log.Printf("  Moderation gRPC: %s", modAddr)

	return server.ListenAndServe()
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
