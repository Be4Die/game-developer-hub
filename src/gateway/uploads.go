package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	gwpb "github.com/Be4Die/game-developer-hub/protos/orchestrator/v1"
	projpb "github.com/Be4Die/game-developer-hub/protos/project_manager/v1"
)

// streamFileChunks считывает файл чанками по 64 КБ и передаёт в gRPC стрим.
func streamFileChunks(r io.Reader, send func([]byte) error) error {
	buf := make([]byte, 64*1024)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			chunk := make([]byte, n)
			copy(chunk, buf[:n])
			if sendErr := send(chunk); sendErr != nil {
				return sendErr
			}
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

// writeProtoJSON сериализует protobuf-ответ в JSON с сохранением имен proto-полей.
func writeProtoJSON(w http.ResponseWriter, msg proto.Message) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	marshaler := &protojson.MarshalOptions{
		UseProtoNames:   true,
		EmitUnpopulated: true,
	}
	bytes, err := marshaler.Marshal(msg)
	if err != nil {
		http.Error(w, fmt.Sprintf("marshal response: %v", err), http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(bytes)
}

// authContext добавляет токены аутентификации из HTTP-запроса в исходящий gRPC-контекст.
func authContext(ctx context.Context, r *http.Request) context.Context {
	if token := r.Header.Get("Authorization"); token != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", token)
	}
	if apiKey := r.Header.Get("X-Api-Key"); apiKey != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "x-api-key", apiKey)
	}
	return ctx
}

// handleBuildUpload обрабатывает POST /api/v1/games/{game_id}/builds (серверные билды Orchestrator).
func handleBuildUpload(client gwpb.BuildServiceClient, fallback http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
			if fallback != nil {
				fallback.ServeHTTP(w, r)
			} else {
				http.Error(w, "multipart/form-data required", http.StatusBadRequest)
			}
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 2048<<20)
		if err := r.ParseMultipartForm(32 << 20); err != nil { //nolint:gosec // bounded by MaxBytesReader
			http.Error(w, fmt.Sprintf("parse multipart: %v", err), http.StatusBadRequest)
			return
		}

		gameID, err := strconv.ParseInt(r.PathValue("game_id"), 10, 64)
		if err != nil {
			http.Error(w, "invalid game_id", http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("image")
		if err != nil {
			http.Error(w, "missing 'image' file field", http.StatusBadRequest)
			return
		}
		defer func() { _ = file.Close() }()

		buildVersion := r.FormValue("build_version")
		if buildVersion == "" {
			http.Error(w, "missing 'build_version' field", http.StatusBadRequest)
			return
		}

		protocol := parseProtocol(r.FormValue("protocol"))
		internalPort := parseUint32(r.FormValue("internal_port"), 8080)
		maxPlayers := parseUint32(r.FormValue("max_players"), 16)

		cleanFilename := filepath.Base(filepath.Clean(header.Filename))
		log.Printf("upload build: game=%d version=%s file=%s size=%d", gameID, buildVersion, cleanFilename, header.Size) //nolint:gosec

		stream, err := client.UploadStream(authContext(r.Context(), r))
		if err != nil {
			http.Error(w, fmt.Sprintf("create upload stream: %v", err), http.StatusInternalServerError)
			return
		}

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

		if err := streamFileChunks(file, func(chunk []byte) error {
			return stream.Send(&gwpb.BuildServiceUploadStreamRequest{
				Payload: &gwpb.BuildServiceUploadStreamRequest_Chunk{Chunk: chunk},
			})
		}); err != nil {
			http.Error(w, fmt.Sprintf("stream file: %v", err), http.StatusInternalServerError)
			return
		}

		resp, err := stream.CloseAndRecv()
		if err != nil {
			http.Error(w, fmt.Sprintf("upload build error: %v", err), http.StatusInternalServerError)
			return
		}

		writeProtoJSON(w, resp)
	}
}

// handleProjectBuildUpload обрабатывает POST /api/v1/projects/{project_id}/builds (клиентские билды Project Manager).
func handleProjectBuildUpload(client projpb.ProjectServiceClient, fallback http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
			if fallback != nil {
				fallback.ServeHTTP(w, r)
			} else {
				http.Error(w, "multipart/form-data required", http.StatusBadRequest)
			}
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 2048<<20)
		if err := r.ParseMultipartForm(32 << 20); err != nil { //nolint:gosec // bounded by MaxBytesReader
			http.Error(w, fmt.Sprintf("parse multipart: %v", err), http.StatusBadRequest)
			return
		}

		projectID, err := strconv.ParseInt(r.PathValue("project_id"), 10, 64)
		if err != nil {
			http.Error(w, "invalid project_id", http.StatusBadRequest)
			return
		}

		version := r.FormValue("version")
		if version == "" {
			http.Error(w, "missing 'version' field", http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "missing 'file' field", http.StatusBadRequest)
			return
		}
		defer func() { _ = file.Close() }()

		cleanFilename := filepath.Base(filepath.Clean(header.Filename))
		log.Printf("upload project build: project=%d version=%s file=%s size=%d", projectID, version, cleanFilename, header.Size) //nolint:gosec

		stream, err := client.UploadBuildStream(authContext(r.Context(), r))
		if err != nil {
			http.Error(w, fmt.Sprintf("create upload stream: %v", err), http.StatusInternalServerError)
			return
		}

		if err := stream.Send(&projpb.ProjectUploadBuildStreamRequest{
			Payload: &projpb.ProjectUploadBuildStreamRequest_Metadata{
				Metadata: &projpb.BuildUploadStreamMetadata{
					ProjectId: projectID,
					Version:   version,
				},
			},
		}); err != nil {
			http.Error(w, fmt.Sprintf("send metadata: %v", err), http.StatusInternalServerError)
			return
		}

		if err := streamFileChunks(file, func(chunk []byte) error {
			return stream.Send(&projpb.ProjectUploadBuildStreamRequest{
				Payload: &projpb.ProjectUploadBuildStreamRequest_Chunk{Chunk: chunk},
			})
		}); err != nil {
			http.Error(w, fmt.Sprintf("stream file: %v", err), http.StatusInternalServerError)
			return
		}

		resp, err := stream.CloseAndRecv()
		if err != nil {
			http.Error(w, fmt.Sprintf("upload build stream error: %v", err), http.StatusInternalServerError)
			return
		}

		writeProtoJSON(w, resp)
	}
}

// handleProjectMediaUpload обрабатывает POST /api/v1/projects/{project_id}/media (медиа Project Manager).
func handleProjectMediaUpload(client projpb.ProjectServiceClient, fallback http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
			if fallback != nil {
				fallback.ServeHTTP(w, r)
			} else {
				http.Error(w, "multipart/form-data required", http.StatusBadRequest)
			}
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 500<<20)
		if err := r.ParseMultipartForm(32 << 20); err != nil { //nolint:gosec // bounded by MaxBytesReader
			http.Error(w, fmt.Sprintf("parse multipart: %v", err), http.StatusBadRequest)
			return
		}

		projectID, err := strconv.ParseInt(r.PathValue("project_id"), 10, 64)
		if err != nil {
			http.Error(w, "invalid project_id", http.StatusBadRequest)
			return
		}

		mediaType := r.FormValue("media_type")
		if mediaType == "" {
			http.Error(w, "missing 'media_type' field", http.StatusBadRequest)
			return
		}

		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "missing 'file' field", http.StatusBadRequest)
			return
		}
		defer func() { _ = file.Close() }()

		stream, err := client.UploadMediaStream(authContext(r.Context(), r))
		if err != nil {
			http.Error(w, fmt.Sprintf("create upload stream: %v", err), http.StatusInternalServerError)
			return
		}

		if err := stream.Send(&projpb.ProjectUploadMediaStreamRequest{
			Payload: &projpb.ProjectUploadMediaStreamRequest_Metadata{
				Metadata: &projpb.MediaUploadStreamMetadata{
					ProjectId: projectID,
					MediaType: mediaType,
				},
			},
		}); err != nil {
			http.Error(w, fmt.Sprintf("send metadata: %v", err), http.StatusInternalServerError)
			return
		}

		if err := streamFileChunks(file, func(chunk []byte) error {
			return stream.Send(&projpb.ProjectUploadMediaStreamRequest{
				Payload: &projpb.ProjectUploadMediaStreamRequest_Chunk{Chunk: chunk},
			})
		}); err != nil {
			http.Error(w, fmt.Sprintf("stream file: %v", err), http.StatusInternalServerError)
			return
		}

		resp, err := stream.CloseAndRecv()
		if err != nil {
			http.Error(w, fmt.Sprintf("upload media stream error: %v", err), http.StatusInternalServerError)
			return
		}

		writeProtoJSON(w, resp)
	}
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
