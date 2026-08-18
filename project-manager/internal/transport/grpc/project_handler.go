package grpc

import (
	"context"
	"fmt"
	"io"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
	"github.com/Be4Die/game-developer-hub/project-manager/internal/service"
	pb "github.com/Be4Die/game-developer-hub/protos/project_manager/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ProjectHandler реализует gRPC-сервис ProjectServiceServer.
type ProjectHandler struct {
	pb.UnimplementedProjectServiceServer
	svc *service.ProjectService
}

// NewProjectHandler создаёт новый экземпляр ProjectHandler.
func NewProjectHandler(svc *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{svc: svc}
}

// Create создаёт новый проект.
func (h *ProjectHandler) Create(ctx context.Context, req *pb.ProjectCreateRequest) (*pb.ProjectCreateResponse, error) {
	ownerID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}
	p, err := h.svc.CreateProject(ctx, ownerID, req.GetTitleRu(), req.GetTitleEn())
	if err != nil {
		return nil, domainError(err, "create project")
	}
	return &pb.ProjectCreateResponse{Project: projectToProto(p)}, nil
}

// Get возвращает проект по ID.
func (h *ProjectHandler) Get(ctx context.Context, req *pb.ProjectGetRequest) (*pb.ProjectGetResponse, error) {
	ownerID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}
	p, err := h.svc.GetProject(ctx, req.GetId())
	if err != nil {
		return nil, domainError(err, "get project")
	}
	if p.OwnerID != ownerID {
		return nil, status.Error(codes.PermissionDenied, "access denied")
	}
	return &pb.ProjectGetResponse{Project: projectToProto(p)}, nil
}

// List возвращает список проектов пользователя.
func (h *ProjectHandler) List(ctx context.Context, req *pb.ProjectListRequest) (*pb.ProjectListResponse, error) {
	ownerID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}
	projects, total, err := h.svc.ListProjects(ctx, ownerID, int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, domainError(err, "list projects")
	}
	resp := &pb.ProjectListResponse{
		Projects: make([]*pb.Project, len(projects)),
		Total:    int32(total),
	}
	for i, p := range projects {
		resp.Projects[i] = projectToProto(p)
	}
	return resp, nil
}

// Update обновляет метаданные черновика проекта.
func (h *ProjectHandler) Update(ctx context.Context, req *pb.ProjectUpdateRequest) (*pb.ProjectUpdateResponse, error) {
	ownerID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}
	meta := domain.DraftMeta{
		TitleRu:            req.GetTitleRu(),
		TitleEn:            req.GetTitleEn(),
		SeoRu:              req.GetSeoRu(),
		SeoEn:              req.GetSeoEn(),
		About:              req.GetAbout(),
		ActiveBuildVersion: req.GetActiveBuildVersion(),
	}
	if err := h.svc.UpdateDraft(ctx, req.GetId(), ownerID, meta); err != nil {
		return nil, domainError(err, "update project draft")
	}
	p, err := h.svc.GetProject(ctx, req.GetId())
	if err != nil {
		return nil, domainError(err, "get updated project")
	}
	return &pb.ProjectUpdateResponse{Project: projectToProto(p)}, nil
}

// Delete удаляет проект со всеми файлами.
func (h *ProjectHandler) Delete(ctx context.Context, req *pb.ProjectDeleteRequest) (*pb.ProjectDeleteResponse, error) {
	ownerID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}
	if err := h.svc.DeleteProject(ctx, req.GetId(), ownerID); err != nil {
		return nil, domainError(err, "delete project")
	}
	return &pb.ProjectDeleteResponse{Success: true}, nil
}

// UploadBuild загружает сборку через Unary RPC (fallback).
func (h *ProjectHandler) UploadBuild(ctx context.Context, req *pb.ProjectUploadBuildRequest) (*pb.ProjectUploadBuildResponse, error) {
	ownerID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}
	b, devURL, err := h.svc.UploadBuild(ctx, req.GetProjectId(), ownerID, req.GetVersion(), req.GetData())
	if err != nil {
		return nil, domainError(err, "upload build")
	}
	return &pb.ProjectUploadBuildResponse{
		Success: true,
		Build:   buildToProto(b),
		DevUrl:  devURL,
	}, nil
}

// UploadBuildStream загружает сборку веб-игры потоково (чанки по 64 КБ).
func (h *ProjectHandler) UploadBuildStream(stream pb.ProjectService_UploadBuildStreamServer) error {
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

	projectID := meta.GetProjectId()
	version := meta.GetVersion()
	ownerID, ok := UserIDFromContext(stream.Context())
	if !ok {
		return status.Error(codes.Unauthenticated, "missing user id")
	}

	pr, pw := io.Pipe()
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

	b, devURL, svcErr := h.svc.UploadBuildStream(stream.Context(), projectID, ownerID, version, pr)
	streamErr := <-done

	if svcErr == nil && streamErr != nil {
		svcErr = streamErr
	}
	if svcErr != nil {
		return domainError(svcErr, "upload build stream")
	}

	return stream.SendAndClose(&pb.ProjectUploadBuildResponse{
		Success: true,
		Build:   buildToProto(b),
		DevUrl:  devURL,
	})
}

// ListBuilds возвращает список сборок проекта.
func (h *ProjectHandler) ListBuilds(ctx context.Context, req *pb.ProjectListBuildsRequest) (*pb.ProjectListBuildsResponse, error) {
	ownerID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}
	builds, err := h.svc.ListBuilds(ctx, req.GetProjectId(), ownerID)
	if err != nil {
		return nil, domainError(err, "list builds")
	}
	resp := &pb.ProjectListBuildsResponse{Builds: make([]*pb.ProjectBuild, len(builds))}
	for i, b := range builds {
		resp.Builds[i] = buildToProto(b)
	}
	return resp, nil
}

// DeleteBuild удаляет сборку проекта.
func (h *ProjectHandler) DeleteBuild(ctx context.Context, req *pb.ProjectDeleteBuildRequest) (*pb.ProjectDeleteBuildResponse, error) {
	ownerID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}
	if err := h.svc.DeleteBuild(ctx, req.GetProjectId(), ownerID, req.GetVersion()); err != nil {
		return nil, domainError(err, "delete build")
	}
	return &pb.ProjectDeleteBuildResponse{Success: true}, nil
}

// UploadMedia загружает промо-материал через Unary RPC.
func (h *ProjectHandler) UploadMedia(ctx context.Context, req *pb.ProjectUploadMediaRequest) (*pb.ProjectUploadMediaResponse, error) {
	ownerID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}
	path, err := h.svc.UploadMedia(ctx, req.GetProjectId(), ownerID, req.GetMediaType(), req.GetData())
	if err != nil {
		return nil, domainError(err, "upload media")
	}
	return &pb.ProjectUploadMediaResponse{Success: true, FilePath: path}, nil
}

// UploadMediaStream загружает промо-материал потоково.
func (h *ProjectHandler) UploadMediaStream(stream pb.ProjectService_UploadMediaStreamServer) error {
	metaReq, err := stream.Recv()
	if err == io.EOF {
		return status.Error(codes.InvalidArgument, "missing metadata")
	}
	if err != nil {
		return status.Errorf(codes.Internal, "read metadata: %v", err)
	}

	meta := metaReq.GetMetadata()
	if meta == nil {
		return status.Error(codes.InvalidArgument, "first message must contain metadata")
	}

	ownerID, ok := UserIDFromContext(stream.Context())
	if !ok {
		return status.Error(codes.Unauthenticated, "missing user id")
	}

	pr, pw := io.Pipe()
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

	path, svcErr := h.svc.UploadMediaStream(stream.Context(), meta.GetProjectId(), ownerID, meta.GetMediaType(), pr)
	streamErr := <-done

	if svcErr == nil && streamErr != nil {
		svcErr = streamErr
	}
	if svcErr != nil {
		return domainError(svcErr, "upload media stream")
	}

	return stream.SendAndClose(&pb.ProjectUploadMediaResponse{
		Success:  true,
		FilePath: path,
	})
}

// SubmitForModeration отправляет проект на модерацию.
func (h *ProjectHandler) SubmitForModeration(ctx context.Context, req *pb.SubmitForModerationRequest) (*pb.SubmitForModerationResponse, error) {
	ownerID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}
	ticket, err := h.svc.SubmitForModeration(ctx, req.GetProjectId(), ownerID)
	if err != nil {
		return nil, domainError(err, "submit for moderation")
	}
	return &pb.SubmitForModerationResponse{
		Success: true,
		Ticket:  ticketToProto(ticket),
	}, nil
}

// GetPublished возвращает опубликованную версию игры.
func (h *ProjectHandler) GetPublished(ctx context.Context, req *pb.ProjectGetPublishedRequest) (*pb.ProjectGetPublishedResponse, error) {
	rel, err := h.svc.GetPublished(ctx, req.GetId())
	if err != nil {
		return nil, domainError(err, "get published release")
	}
	return &pb.ProjectGetPublishedResponse{Release: releaseToProto(rel)}, nil
}

// Unpublish снимает игру с публикации в прод-окружении.
func (h *ProjectHandler) Unpublish(ctx context.Context, req *pb.ProjectUnpublishRequest) (*pb.ProjectUnpublishResponse, error) {
	ownerID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}
	if err := h.svc.Unpublish(ctx, req.GetId(), ownerID); err != nil {
		return nil, domainError(err, "unpublish project")
	}
	return &pb.ProjectUnpublishResponse{Success: true}, nil
}
