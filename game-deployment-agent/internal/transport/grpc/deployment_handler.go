package grpc

import (
	"context"
	"errors"
	"io"

	"github.com/Be4Die/game-developer-hub/game-deployment-agent/internal/domain"
	"github.com/Be4Die/game-developer-hub/game-deployment-agent/internal/service"
	pb "github.com/Be4Die/game-developer-hub/protos/game_deployment_agent/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DeploymentHandler реализует pb.WebGameDeploymentServiceServer.
type DeploymentHandler struct {
	pb.UnimplementedWebGameDeploymentServiceServer
	svc     *service.DeploymentService
	version string
}

// NewDeploymentHandler создает экземпляр gRPC хендлера для агента развертывания.
func NewDeploymentHandler(svc *service.DeploymentService, version string) *DeploymentHandler {
	if version == "" {
		version = "1.0.0"
	}
	return &DeploymentHandler{
		svc:     svc,
		version: version,
	}
}

// DeployDevStream принимает поток архива, распаковывает и переключает dev-симлинк.
func (h *DeploymentHandler) DeployDevStream(stream pb.WebGameDeploymentService_DeployDevStreamServer) error {
	// Первое сообщение обязательно содержит метаданные
	firstMsg, err := stream.Recv()
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "failed to receive metadata: %v", err)
	}

	meta := firstMsg.GetMetadata()
	if meta == nil || meta.ProjectId <= 0 || meta.Version == "" {
		return status.Error(codes.InvalidArgument, "first message must contain project_id and version metadata")
	}

	pr, pw := io.Pipe()

	errChan := make(chan error, 1)
	go func() {
		defer pw.Close()
		for {
			req, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				return
			}
			if err != nil {
				pw.CloseWithError(err)
				return
			}
			if chunk := req.GetChunk(); len(chunk) > 0 {
				if _, err := pw.Write(chunk); err != nil {
					return
				}
			}
		}
	}()

	var res *domain.DeploymentResult
	go func() {
		var deployErr error
		res, deployErr = h.svc.DeployDevStream(stream.Context(), meta.ProjectId, meta.Version, pr)
		errChan <- deployErr
	}()

	if err := <-errChan; err != nil {
		if errors.Is(err, domain.ErrInvalidArchive) || errors.Is(err, domain.ErrNoIndexHtml) ||
			errors.Is(err, domain.ErrZipSlip) || errors.Is(err, domain.ErrZipBomb) ||
			errors.Is(err, domain.ErrDisallowedFileType) {
			return status.Errorf(codes.InvalidArgument, "%v", err)
		}
		return status.Errorf(codes.Internal, "deployment failed: %v", err)
	}

	return stream.SendAndClose(&pb.DeployResponse{
		Success:      res.Success,
		Url:          res.URL,
		UnpackedPath: res.UnpackedPath,
	})
}

// DeployProd переключает боевой прод-симлинк на указанную версию.
func (h *DeploymentHandler) DeployProd(ctx context.Context, req *pb.DeployProdRequest) (*pb.DeployResponse, error) {
	if req.ProjectId <= 0 || req.Version == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id and version are required")
	}

	res, err := h.svc.DeployProd(ctx, req.ProjectId, req.Version)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "deploy prod failed: %v", err)
	}

	return &pb.DeployResponse{
		Success:      res.Success,
		Url:          res.URL,
		UnpackedPath: res.UnpackedPath,
	}, nil
}

// UndeployProd снимает игру с публикации.
func (h *DeploymentHandler) UndeployProd(ctx context.Context, req *pb.UndeployProdRequest) (*pb.UndeployResponse, error) {
	if req.ProjectId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}

	if err := h.svc.UndeployProd(ctx, req.ProjectId); err != nil {
		return nil, status.Errorf(codes.Internal, "undeploy prod failed: %v", err)
	}

	return &pb.UndeployResponse{Success: true}, nil
}

// DeleteVersion удаляет распакованную сборку.
func (h *DeploymentHandler) DeleteVersion(ctx context.Context, req *pb.DeleteVersionRequest) (*pb.DeleteVersionResponse, error) {
	if req.ProjectId <= 0 || req.Version == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id and version are required")
	}

	if err := h.svc.DeleteVersion(ctx, req.ProjectId, req.Version); err != nil {
		return nil, status.Errorf(codes.Internal, "delete version failed: %v", err)
	}

	return &pb.DeleteVersionResponse{Success: true}, nil
}

// DeleteProject удаляет все файлы проекта на агенте.
func (h *DeploymentHandler) DeleteProject(ctx context.Context, req *pb.DeleteProjectRequest) (*pb.DeleteProjectResponse, error) {
	if req.ProjectId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}

	if err := h.svc.DeleteProject(ctx, req.ProjectId); err != nil {
		return nil, status.Errorf(codes.Internal, "delete project failed: %v", err)
	}

	return &pb.DeleteProjectResponse{Success: true}, nil
}

// Health возвращает статус здоровья агента.
func (h *DeploymentHandler) Health(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	freeBytes := h.svc.GetFreeDiskBytes()
	return &pb.HealthResponse{
		Healthy:       true,
		Version:       h.version,
		FreeDiskBytes: freeBytes,
	}, nil
}
