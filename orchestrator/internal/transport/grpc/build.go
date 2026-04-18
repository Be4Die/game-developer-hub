package grpc

import (
	"context"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/service"
	pb "github.com/Be4Die/game-developer-hub/protos/orchestrator/v1"
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

// Upload загружает серверный билд.
func (h *BuildHandler) Upload(ctx context.Context, req *pb.UploadBuildRequest) (*pb.UploadBuildResponse, error) {
	params := service.UploadBuildParams{
		GameID:       req.GetGameId(),
		Version:      req.GetBuildVersion(),
		Protocol:     protocolFromProto(req.GetProtocol()),
		InternalPort: req.GetInternalPort(),
		MaxPlayers:   req.GetMaxPlayers(),
		Archive:      nil, // Будет обработано отдельно — image_data в памяти.
		ArchiveSize:  int64(len(req.GetImageData())),
	}

	// Для gRPC передача данных идёт напрямую через bytes, используем bytesReader.
	params.ArchiveData = req.GetImageData()

	build, err := h.pipeline.UploadBuildFromBytes(ctx, params)
	if err != nil {
		return nil, domainError(err, "upload build")
	}

	return &pb.UploadBuildResponse{Build: buildToProto(build)}, nil
}

// List возвращает список билдов игры.
func (h *BuildHandler) List(ctx context.Context, req *pb.ListBuildsRequest) (*pb.ListBuildsResponse, error) {
	builds, err := h.pipeline.ListBuilds(ctx, req.GetGameId())
	if err != nil {
		return nil, domainError(err, "list builds")
	}

	resp := make([]*pb.ServerBuild, 0, len(builds))
	for _, b := range builds {
		resp = append(resp, buildToProto(b))
	}

	return &pb.ListBuildsResponse{Builds: resp}, nil
}

// Get возвращает информацию о билде.
func (h *BuildHandler) Get(ctx context.Context, req *pb.GetBuildRequest) (*pb.GetBuildResponse, error) {
	build, err := h.pipeline.GetBuild(ctx, req.GetGameId(), req.GetBuildVersion())
	if err != nil {
		return nil, domainError(err, "get build")
	}

	return &pb.GetBuildResponse{Build: buildToProto(build)}, nil
}

// Delete удаляет серверный билд.
func (h *BuildHandler) Delete(ctx context.Context, req *pb.DeleteBuildRequest) (*pb.DeleteBuildResponse, error) {
	err := h.pipeline.DeleteBuild(ctx, req.GetGameId(), req.GetBuildVersion())
	if err != nil {
		return nil, domainError(err, "delete build")
	}

	return &pb.DeleteBuildResponse{}, nil
}
