package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
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
	// 2GB max message size for large build uploads
	const maxMsgSize = 2 * 1024 * 1024 * 1024 // 2GB
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(maxMsgSize),
			grpc.MaxCallSendMsgSize(maxMsgSize),
		),
	}

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

	// Custom handler для multipart upload — intercepts build upload.
	buildUploadHandler := newBuildUploadHandler(orchConn)

	// HTTP-сервер с CORS и custom routing.
	handler := corsMiddleware(buildUploadRouter(mux, buildUploadHandler))

	srv := &http.Server{
		Addr:         httpAddr,
		Handler:      handler,
		ReadTimeout:  600 * time.Second,
		WriteTimeout: 600 * time.Second,
		IdleTimeout:  120 * time.Second,
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

// buildUploadHandler принимает multipart/form-data и стримит в gRPC UploadStream.
type buildUploadHandler struct {
	client gwpb.BuildServiceClient
}

func newBuildUploadHandler(conn *grpc.ClientConn) *buildUploadHandler {
	return &buildUploadHandler{client: gwpb.NewBuildServiceClient(conn)}
}

// ServeHTTP обрабатывает POST /api/v1/games/{game_id}/builds с multipart/form-data.
func (h *buildUploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Проверяем что это POST с multipart.
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Парсим multipart.
	err := r.ParseMultipartForm(32 << 20) // 32MB буфер в памяти
	if err != nil {
		http.Error(w, fmt.Sprintf("parse multipart: %v", err), http.StatusBadRequest)
		return
	}

	// Извлекаем game_id из URL path: /api/v1/games/{id}/builds
	gameID, err := extractGameID(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Извлекаем файл.
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "missing 'image' file field", http.StatusBadRequest)
		return
	}
	defer func() { _ = file.Close() }()

	// Извлекаем метаданные из form fields.
	buildVersion := r.FormValue("build_version")
	if buildVersion == "" {
		http.Error(w, "missing 'build_version' field", http.StatusBadRequest)
		return
	}

	protocol := parseProtocol(r.FormValue("protocol"))
	internalPort := parseUint32(r.FormValue("internal_port"), 8080)
	maxPlayers := parseUint32(r.FormValue("max_players"), 16)

	log.Printf("upload build: game=%d version=%s file=%s size=%d", gameID, buildVersion, header.Filename, header.Size)

	// Открываем gRPC UploadStream.
	ctx := r.Context()
	// Пробрасываем auth заголовок.
	if token := r.Header.Get("Authorization"); token != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", token)
	}
	if apiKey := r.Header.Get("X-Api-Key"); apiKey != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "x-api-key", apiKey)
	}

	stream, err := h.client.UploadStream(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("create upload stream: %v", err), http.StatusInternalServerError)
		return
	}

	// Отправляем метаданные первым сообщением.
	if err := stream.Send(&gwpb.BuildServiceUploadStreamRequest{
		Payload: &gwpb.BuildServiceUploadStreamRequest_Metadata{
			Metadata: &gwpb.BuildUploadStreamMetadata{
				GameId:       gameID,
				BuildVersion: buildVersion,
				Protocol:     protocol,
				InternalPort: internalPort,
				MaxPlayers:   maxPlayers,
			},
		},
	}); err != nil {
		http.Error(w, fmt.Sprintf("send metadata: %v", err), http.StatusInternalServerError)
		return
	}

	// Стримим файл чанками.
	buf := make([]byte, 64*1024) // 64KB chunks
	var totalSent int64
	for {
		n, readErr := file.Read(buf)
		if n > 0 {
			chunk := make([]byte, n)
			copy(chunk, buf[:n])
			if sendErr := stream.Send(&gwpb.BuildServiceUploadStreamRequest{
				Payload: &gwpb.BuildServiceUploadStreamRequest_Chunk{
					Chunk: chunk,
				},
			}); sendErr != nil {
				http.Error(w, fmt.Sprintf("send chunk: %v", sendErr), http.StatusInternalServerError)
				return
			}
			totalSent += int64(n)
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			http.Error(w, fmt.Sprintf("read file: %v", readErr), http.StatusInternalServerError)
			return
		}
	}

	log.Printf("upload build: streamed %d bytes for game=%d", totalSent, gameID)

	// Закрываем стрим и получаем ответ.
	resp, err := stream.CloseAndRecv()
	if err != nil {
		http.Error(w, fmt.Sprintf("close stream: %v", err), http.StatusInternalServerError)
		return
	}

	// Возвращаем JSON ответ.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	marshaler := &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames:   true,
			EmitUnpopulated: true,
		},
	}
	jsonBytes, _ := marshaler.Marshal(resp)
	_, _ = w.Write(jsonBytes)
}

// buildUploadRouter маршрутизирует upload запросы к custom handler, остальные — к grpc-gateway.
func buildUploadRouter(gw http.Handler, uploadHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		isMultipart := len(ct) >= 19 && ct[:19] == "multipart/form-data"

		prefix := "/api/v1/games/"
		suffix := "builds"
		pathOK := len(r.URL.Path) >= 22 &&
			r.URL.Path[:len(prefix)] == prefix &&
			r.URL.Path[len(r.URL.Path)-len(suffix):] == suffix

		// Intercept POST /api/v1/games/{id}/builds with multipart/form-data.
		if r.Method == http.MethodPost && isMultipart && pathOK {
			uploadHandler.ServeHTTP(w, r)
			return
		}
		gw.ServeHTTP(w, r)
	})
}

func extractGameID(path string) (int64, error) {
	// /api/v1/games/{id}/builds
	const prefix = "/api/v1/games/"
	const suffix = "/builds"

	if len(path) < len(prefix)+len(suffix)+1 {
		return 0, fmt.Errorf("invalid upload path")
	}

	start := len(prefix)
	end := len(path) - len(suffix)
	idStr := path[start:end]

	gameID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid game_id: %s", idStr)
	}
	return gameID, nil
}

func parseProtocol(s string) gwpb.Protocol {
	switch s {
	case "tcp":
		return gwpb.Protocol_PROTOCOL_TCP
	case "udp":
		return gwpb.Protocol_PROTOCOL_UDP
	case "websocket":
		return gwpb.Protocol_PROTOCOL_WEBSOCKET
	case "webrtc":
		return gwpb.Protocol_PROTOCOL_WEBRTC
	default:
		return gwpb.Protocol_PROTOCOL_WEBSOCKET
	}
}

func parseUint32(s string, fallback uint32) uint32 {
	if s == "" {
		return fallback
	}
	v, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return fallback
	}
	return uint32(v)
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
