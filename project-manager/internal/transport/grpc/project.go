package grpc

import (
	"context"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
	"github.com/Be4Die/game-developer-hub/project-manager/internal/service"
	pb "github.com/Be4Die/game-developer-hub/protos/project_manager/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ProjectHandler gRPC-хендлер для ProjectService.
type ProjectHandler struct {
	pb.UnimplementedProjectServiceServer
	svc *service.ProjectService
}

// NewProjectHandler создаёт хендлер.
func NewProjectHandler(svc *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{svc: svc}
}

// Create создаёт проект.
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

// Get возвращает проект.
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

// List возвращает список проектов.
func (h *ProjectHandler) List(ctx context.Context, req *pb.ProjectListRequest) (*pb.ProjectListResponse, error) {
	ownerID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}
	projects, err := h.svc.ListProjects(ctx, ownerID)
	if err != nil {
		return nil, domainError(err, "list projects")
	}
	resp := &pb.ProjectListResponse{Projects: make([]*pb.Project, len(projects))}
	for i, p := range projects {
		resp.Projects[i] = projectToProto(p)
	}
	return resp, nil
}

// Update обновляет проект.
func (h *ProjectHandler) Update(ctx context.Context, req *pb.ProjectUpdateRequest) (*pb.ProjectUpdateResponse, error) {
	ownerID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}
	meta := service.ProjectMeta{
		TitleRu:            req.GetTitleRu(),
		TitleEn:            req.GetTitleEn(),
		SeoRu:              req.GetSeoRu(),
		SeoEn:              req.GetSeoEn(),
		About:              req.GetAbout(),
		ActiveBuildVersion: req.GetActiveBuildVersion(),
	}
	if err := h.svc.UpdateMeta(ctx, req.GetId(), ownerID, meta); err != nil {
		return nil, domainError(err, "update project")
	}
	p, err := h.svc.GetProject(ctx, req.GetId())
	if err != nil {
		return nil, domainError(err, "get updated project")
	}
	return &pb.ProjectUpdateResponse{Project: projectToProto(p)}, nil
}

// Delete удаляет проект.
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

// UploadBuild загружает билд.
func (h *ProjectHandler) UploadBuild(ctx context.Context, req *pb.ProjectUploadBuildRequest) (*pb.ProjectUploadBuildResponse, error) {
	ownerID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}
	if err := h.svc.UploadBuild(ctx, req.GetProjectId(), ownerID, req.GetVersion(), req.GetData()); err != nil {
		return nil, domainError(err, "upload build")
	}
	return &pb.ProjectUploadBuildResponse{Success: true}, nil
}

// ListBuilds возвращает билды.
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

// DeleteBuild удаляет билд.
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

// UploadMedia загружает промо-материал.
func (h *ProjectHandler) UploadMedia(ctx context.Context, req *pb.ProjectUploadMediaRequest) (*pb.ProjectUploadMediaResponse, error) {
	ownerID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}
	if err := h.svc.UploadMedia(ctx, req.GetProjectId(), ownerID, req.GetMediaType(), req.GetData()); err != nil {
		return nil, domainError(err, "upload media")
	}
	return &pb.ProjectUploadMediaResponse{Success: true}, nil
}

// ─── Конвертеры ──────────────────────────────────────────────

func projectToProto(p *domain.Project) *pb.Project {
	return &pb.Project{
		Id:        p.ID,
		OwnerId:   p.OwnerID,
		TitleRu:            p.TitleRu,
		TitleEn:            p.TitleEn,
		SeoRu:              p.SeoRu,
		SeoEn:              p.SeoEn,
		About:              p.About,
		Status:             pb.ProjectStatus(p.Status),
		IconPath:           p.IconPath,
		CoverPath:          p.CoverPath,
		VideoPath:          p.VideoPath,
		ActiveBuildVersion: p.ActiveBuildVersion,
		CreatedAt: p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func buildToProto(b *domain.ProjectBuild) *pb.ProjectBuild {
	return &pb.ProjectBuild{
		Id:        b.ID,
		ProjectId: b.ProjectID,
		Version:   b.Version,
		FilePath:  b.FilePath,
		FileSize:  b.FileSize,
		CreatedAt: b.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// domainError мапит доменные ошибки в gRPC статусы.
func domainError(err error, action string) error {
	switch {
	case err == domain.ErrNotFound:
		return status.Errorf(codes.NotFound, "%s: not found", action)
	case err == domain.ErrAlreadyExists:
		return status.Errorf(codes.AlreadyExists, "%s: already exists", action)
	case err == domain.ErrForbidden:
		return status.Errorf(codes.PermissionDenied, "%s: forbidden", action)
	case err == domain.ErrInvalidInput:
		return status.Errorf(codes.InvalidArgument, "%s: invalid input", action)
	default:
		return status.Errorf(codes.Internal, "%s: %v", action, err)
	}
}
