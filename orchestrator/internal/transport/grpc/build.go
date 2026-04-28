package grpc

import (
	"context"
	"fmt"
	"io"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/service"
	pb "github.com/Be4Die/game-developer-hub/protos/orchestrator/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// BuildHandler реализует BuildService.
type BuildHandler struct {
	pb.UnimplementedBuildServiceServer
	pipeline *service.BuildPipeline
}

// NewBuildHandler создаёт обработчик билдов.
func NewBuildHandler(pipeline *service.BuildPipeline) *BuildHandler {
	return &BuildHandler{pipeline: pipeline}
}

// Upload загружает серверный билд (unary, через grpc-gateway).
func (h *BuildHandler) Upload(ctx context.Context, req *pb.BuildServiceUploadRequest) (*pb.BuildServiceUploadResponse, error) {
	ownerID, _ := GetUserID(ctx)

	params := service.UploadBuildParams{
		GameID:       req.GetGameId(),
		OwnerID:      ownerID,
		Version:      req.GetBuildVersion(),
		Protocol:     protocolFromProto(req.GetProtocol()),
		InternalPort: req.GetInternalPort(),
		MaxPlayers:   req.GetMaxPlayers(),
		Archive:      nil,
		ArchiveData:  req.GetImageData(),
		ArchiveSize:  int64(len(req.GetImageData())),
	}

	build, err := h.pipeline.UploadBuild(ctx, params)
	if err != nil {
		return nil, domainError(err, "upload build")
	}

	return &pb.BuildServiceUploadResponse{Build: buildToProto(build)}, nil
}

// UploadStream загружает серверный билд через client streaming (internal, no HTTP).
// Первое сообщение содержит метаданные, последующие — чанки файла.
func (h *BuildHandler) UploadStream(stream pb.BuildService_UploadStreamServer) error {
	// Читаем первое сообщение с метаданными.
	metaReq, err := stream.Recv()
	if err == io.EOF {
		return status.Error(codes.InvalidArgument, "missing upload metadata")
	}
	if err != nil {
		return status.Errorf(codes.Internal, "read metadata: %v", err)
	}

	meta := metaReq.GetMetadata()
	if meta == nil {
		return status.Error(codes.InvalidArgument, "first message must contain metadata")
	}

	gameID := meta.GetGameId()
	ownerID, _ := GetUserID(stream.Context())

	params := service.UploadBuildParams{
		GameID:       gameID,
		OwnerID:      ownerID,
		Version:      meta.GetBuildVersion(),
		Protocol:     protocolFromProto(meta.GetProtocol()),
		InternalPort: meta.GetInternalPort(),
		MaxPlayers:   meta.GetMaxPlayers(),
	}

	// Создаём pipe для стриминга чанков в сервис.
	pr, pw := io.Pipe()
	params.Archive = pr

	// Канал для передачи ошибки из горутины чтения.
	done := make(chan error, 1)

	go func() {
		for {
			req, recvErr := stream.Recv()
			if recvErr == io.EOF {
				_ = pw.Close()
				done <- nil
				return
			}
			if recvErr != nil {
				_ = pw.CloseWithError(fmt.Errorf("stream recv: %w", recvErr))
				done <- recvErr
				return
			}

			chunk := req.GetChunk()
			if len(chunk) == 0 {
				continue
			}

			if _, writeErr := pw.Write(chunk); writeErr != nil {
				_ = pw.CloseWithError(fmt.Errorf("pipe write: %w", writeErr))
				done <- writeErr
				return
			}
		}
	}()

	build, err := h.pipeline.UploadBuild(stream.Context(), params)

	// Ждём завершения горутины чтения.
	streamErr := <-done
	if err == nil && streamErr != nil {
		err = streamErr
	}
	if err != nil {
		return domainError(err, "upload build")
	}

	return stream.SendAndClose(&pb.BuildServiceUploadResponse{Build: buildToProto(build)})
}

// List возвращает список билдов игры.
func (h *BuildHandler) List(ctx context.Context, req *pb.BuildServiceListRequest) (*pb.BuildServiceListResponse, error) {
	builds, err := h.pipeline.ListBuilds(ctx, req.GetGameId())
	if err != nil {
		return nil, domainError(err, "list builds")
	}

	resp := make([]*pb.ServerBuild, 0, len(builds))
	for _, b := range builds {
		resp = append(resp, buildToProto(b))
	}

	return &pb.BuildServiceListResponse{Builds: resp}, nil
}

// Get возвращает информацию о билде.
func (h *BuildHandler) Get(ctx context.Context, req *pb.BuildServiceGetRequest) (*pb.BuildServiceGetResponse, error) {
	build, err := h.pipeline.GetBuild(ctx, req.GetGameId(), req.GetBuildVersion())
	if err != nil {
		return nil, domainError(err, "get build")
	}

	return &pb.BuildServiceGetResponse{Build: buildToProto(build)}, nil
}

// Delete удаляет серверный билд.
func (h *BuildHandler) Delete(ctx context.Context, req *pb.BuildServiceDeleteRequest) (*pb.BuildServiceDeleteResponse, error) {
	err := h.pipeline.DeleteBuild(ctx, req.GetGameId(), req.GetBuildVersion())
	if err != nil {
		return nil, domainError(err, "delete build")
	}

	return &pb.BuildServiceDeleteResponse{}, nil
}
