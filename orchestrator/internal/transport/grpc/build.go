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

	build, err := h.pipeline.UploadBuildFromBytes(ctx, params)
	if err != nil {
		return nil, domainError(err, "upload build")
	}

	return &pb.BuildServiceUploadResponse{Build: buildToProto(build)}, nil
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
