package grpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
	"github.com/Be4Die/game-developer-hub/project-manager/internal/service"
	pb "github.com/Be4Die/game-developer-hub/protos/project_manager/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func clampInt32(v int) int32 {
	if v > math.MaxInt32 {
		return math.MaxInt32
	}
	if v < math.MinInt32 {
		return math.MinInt32
	}
	return int32(v)
}

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
	p, err := h.svc.CreateProject(ctx, ownerID, req.GetTitleRu(), req.GetTitleEn(), req.GetIsOnline())
	if err != nil {
		return nil, domainError(err, "create project")
	}
	return &pb.ProjectCreateResponse{Project: projectToProto(p)}, nil
}

// Get возвращает проект по ID с правами доступа текущего пользователя.
func (h *ProjectHandler) Get(ctx context.Context, req *pb.ProjectGetRequest) (*pb.ProjectGetResponse, error) {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}
	userRole, _ := UserRoleFromContext(ctx)
	isStaff := userRole == 2 || userRole == 3
	p, err := h.svc.GetProjectForUser(ctx, req.GetId(), userID, isStaff)
	if err != nil {
		return nil, domainError(err, "get project")
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
		Total:    clampInt32(total),
	}
	for i, p := range projects {
		resp.Projects[i] = projectToProto(p)
	}
	return resp, nil
}

// ListPublished возвращает список опубликованных проектов каталога.
func (h *ProjectHandler) ListPublished(ctx context.Context, req *pb.ProjectListPublishedRequest) (*pb.ProjectListPublishedResponse, error) {
	projects, total, err := h.svc.ListPublishedProjects(ctx, int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, domainError(err, "list published projects")
	}
	resp := &pb.ProjectListPublishedResponse{
		Projects: make([]*pb.Project, len(projects)),
		Total:    clampInt32(total),
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
		AboutRu:            req.GetAboutRu(),
		AboutEn:            req.GetAboutEn(),
		ActiveBuildVersion: req.GetActiveBuildVersion(),
		IsOnline:           req.IsOnline,
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
	if errors.Is(err, io.EOF) {
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
			if errors.Is(recvErr, io.EOF) {
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
	_ = pr.Close()
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
	if errors.Is(err, io.EOF) {
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
			if errors.Is(recvErr, io.EOF) {
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
	_ = pr.Close()
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
	requestID, err := h.svc.SubmitForModeration(ctx, req.GetProjectId(), ownerID)
	if err != nil {
		return nil, domainError(err, "submit for moderation")
	}
	return &pb.SubmitForModerationResponse{
		Success:   true,
		RequestId: requestID,
	}, nil
}

// PublishRelease публикует одобренную версию игры в продуктивное окружение.
func (h *ProjectHandler) PublishRelease(ctx context.Context, req *pb.ProjectPublishReleaseRequest) (*pb.ProjectPublishReleaseResponse, error) {
	rel, err := h.svc.PublishRelease(ctx, req.GetProjectId(), req.GetVersion(), req.GetPublishedBy())
	if err != nil {
		return nil, domainError(err, "publish release")
	}
	return &pb.ProjectPublishReleaseResponse{
		Success: true,
		Release: releaseToProto(rel),
	}, nil
}

// RejectDraft возвращает черновик на доработку при отклонении модератором.
func (h *ProjectHandler) RejectDraft(ctx context.Context, req *pb.ProjectRejectDraftRequest) (*pb.ProjectRejectDraftResponse, error) {
	if err := h.svc.RejectDraft(ctx, req.GetProjectId()); err != nil {
		return nil, domainError(err, "reject draft")
	}
	return &pb.ProjectRejectDraftResponse{Success: true}, nil
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
	userRole, _ := UserRoleFromContext(ctx)
	isStaff := userRole == 2 || userRole == 3
	if err := h.svc.Unpublish(ctx, req.GetId(), ownerID, isStaff); err != nil {
		return nil, domainError(err, "unpublish project")
	}
	return &pb.ProjectUnpublishResponse{Success: true}, nil
}

// ─── Приглашения и общий доступ ──────────────────────────────

// SendInvitation отправляет приглашение пользователю в команду проекта.
func (h *ProjectHandler) SendInvitation(ctx context.Context, req *pb.ProjectSendInvitationRequest) (*pb.ProjectSendInvitationResponse, error) {
	inviterID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}
	inviterEmail := UserEmailFromContext(ctx)
	inviterName := UserNameFromContext(ctx)

	inv, err := h.svc.SendInvitation(
		ctx,
		req.GetProjectId(),
		inviterID,
		inviterEmail,
		inviterName,
		req.GetInviteeId(),
		req.GetInviteeEmail(),
		req.GetPermissions(),
	)
	if err != nil {
		return nil, domainError(err, "send invitation")
	}
	return &pb.ProjectSendInvitationResponse{Invitation: invitationToProto(inv)}, nil
}

// ListIncomingInvitations возвращает входящие активные приглашения текущего пользователя.
func (h *ProjectHandler) ListIncomingInvitations(ctx context.Context, _ *pb.ProjectListIncomingInvitationsRequest) (*pb.ProjectListIncomingInvitationsResponse, error) {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}
	invitations, err := h.svc.ListIncomingInvitations(ctx, userID)
	if err != nil {
		return nil, domainError(err, "list incoming invitations")
	}
	resp := &pb.ProjectListIncomingInvitationsResponse{
		Invitations: make([]*pb.ProjectInvitation, len(invitations)),
	}
	for i, inv := range invitations {
		resp.Invitations[i] = invitationToProto(inv)
	}
	return resp, nil
}

// ListOutgoingInvitations возвращает исходящие приглашения пользователя по проектам.
func (h *ProjectHandler) ListOutgoingInvitations(ctx context.Context, req *pb.ProjectListOutgoingInvitationsRequest) (*pb.ProjectListOutgoingInvitationsResponse, error) {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}
	invitations, err := h.svc.ListOutgoingInvitations(ctx, userID, req.GetProjectId())
	if err != nil {
		return nil, domainError(err, "list outgoing invitations")
	}
	resp := &pb.ProjectListOutgoingInvitationsResponse{
		Invitations: make([]*pb.ProjectInvitation, len(invitations)),
	}
	for i, inv := range invitations {
		resp.Invitations[i] = invitationToProto(inv)
	}
	return resp, nil
}

// RespondInvitation принимает или отклоняет приглашение в проект.
func (h *ProjectHandler) RespondInvitation(ctx context.Context, req *pb.ProjectRespondInvitationRequest) (*pb.ProjectRespondInvitationResponse, error) {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}
	if err := h.svc.RespondInvitation(ctx, req.GetInvitationId(), userID, req.GetAccept()); err != nil {
		return nil, domainError(err, "respond invitation")
	}
	return &pb.ProjectRespondInvitationResponse{Success: true}, nil
}

// CancelInvitation отменяет отправленное приглашение.
func (h *ProjectHandler) CancelInvitation(ctx context.Context, req *pb.ProjectCancelInvitationRequest) (*pb.ProjectCancelInvitationResponse, error) {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}
	if err := h.svc.CancelInvitation(ctx, req.GetInvitationId(), userID); err != nil {
		return nil, domainError(err, "cancel invitation")
	}
	return &pb.ProjectCancelInvitationResponse{Success: true}, nil
}

// ListMembers возвращает участников проекта.
func (h *ProjectHandler) ListMembers(ctx context.Context, req *pb.ProjectListMembersRequest) (*pb.ProjectListMembersResponse, error) {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}
	members, ownerID, err := h.svc.ListProjectMembers(ctx, req.GetProjectId(), userID)
	if err != nil {
		return nil, domainError(err, "list members")
	}
	resp := &pb.ProjectListMembersResponse{
		Members: make([]*pb.ProjectMember, len(members)),
		OwnerId: ownerID,
	}
	for i, m := range members {
		resp.Members[i] = memberToProto(m)
	}
	return resp, nil
}

// UpdateMemberPermissions изменяет права участника проекта.
func (h *ProjectHandler) UpdateMemberPermissions(ctx context.Context, req *pb.ProjectUpdateMemberPermissionsRequest) (*pb.ProjectUpdateMemberPermissionsResponse, error) {
	ownerID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}
	member, err := h.svc.UpdateMemberPermissions(ctx, req.GetProjectId(), ownerID, req.GetUserId(), req.GetPermissions())
	if err != nil {
		return nil, domainError(err, "update member permissions")
	}
	return &pb.ProjectUpdateMemberPermissionsResponse{Member: memberToProto(member)}, nil
}

// RemoveMember удаляет участника из проекта.
func (h *ProjectHandler) RemoveMember(ctx context.Context, req *pb.ProjectRemoveMemberRequest) (*pb.ProjectRemoveMemberResponse, error) {
	ownerID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}
	if err := h.svc.RemoveMember(ctx, req.GetProjectId(), ownerID, req.GetUserId()); err != nil {
		return nil, domainError(err, "remove member")
	}
	return &pb.ProjectRemoveMemberResponse{Success: true}, nil
}

// LeaveProject позволяет участнику покинуть проект.
func (h *ProjectHandler) LeaveProject(ctx context.Context, req *pb.ProjectLeaveRequest) (*pb.ProjectLeaveResponse, error) {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}
	if err := h.svc.LeaveProject(ctx, req.GetProjectId(), userID); err != nil {
		return nil, domainError(err, "leave project")
	}
	return &pb.ProjectLeaveResponse{Success: true}, nil
}

// ListSharedProjects возвращает проекты, в которых текущий пользователь является участником.
func (h *ProjectHandler) ListSharedProjects(ctx context.Context, _ *pb.ProjectListSharedProjectsRequest) (*pb.ProjectListSharedProjectsResponse, error) {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}
	projects, err := h.svc.ListSharedProjects(ctx, userID)
	if err != nil {
		return nil, domainError(err, "list shared projects")
	}
	resp := &pb.ProjectListSharedProjectsResponse{
		Projects: make([]*pb.SharedProjectItem, len(projects)),
	}
	for i, sp := range projects {
		resp.Projects[i] = sharedProjectToProto(sp)
	}
	return resp, nil
}

// BlockUser добавляет пользователя в черный список для блокировки спама.
func (h *ProjectHandler) BlockUser(ctx context.Context, req *pb.ProjectBlockUserRequest) (*pb.ProjectBlockUserResponse, error) {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}
	if err := h.svc.BlockUser(ctx, userID, req.GetBlockedUserId(), req.GetBlockedUserEmail(), ""); err != nil {
		return nil, domainError(err, "block user")
	}
	return &pb.ProjectBlockUserResponse{Success: true}, nil
}

// UnblockUser удаляет пользователя из черного списка.
func (h *ProjectHandler) UnblockUser(ctx context.Context, req *pb.ProjectUnblockUserRequest) (*pb.ProjectUnblockUserResponse, error) {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}
	if err := h.svc.UnblockUser(ctx, userID, req.GetBlockedUserId()); err != nil {
		return nil, domainError(err, "unblock user")
	}
	return &pb.ProjectUnblockUserResponse{Success: true}, nil
}

// ListBlockedUsers возвращает список заблокированных пользователей.
func (h *ProjectHandler) ListBlockedUsers(ctx context.Context, _ *pb.ProjectListBlockedUsersRequest) (*pb.ProjectListBlockedUsersResponse, error) {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}
	blocks, err := h.svc.ListBlockedUsers(ctx, userID)
	if err != nil {
		return nil, domainError(err, "list blocked users")
	}
	resp := &pb.ProjectListBlockedUsersResponse{
		Blocks: make([]*pb.UserAccessBlock, len(blocks)),
	}
	for i, b := range blocks {
		resp.Blocks[i] = blockToProto(b)
	}
	return resp, nil
}

// ─── Внутриигровые покупки (IAP) ───────────────────────────────

// ListGameItems возвращает список товаров для игры.
func (h *ProjectHandler) ListGameItems(ctx context.Context, req *pb.ProjectListGameItemsRequest) (*pb.ProjectListGameItemsResponse, error) {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}
	items, err := h.svc.ListGameItems(ctx, req.GetProjectId(), userID)
	if err != nil {
		return nil, domainError(err, "list game items")
	}
	resp := &pb.ProjectListGameItemsResponse{
		Items: make([]*pb.GameItem, len(items)),
	}
	for i, it := range items {
		resp.Items[i] = gameItemToProto(it)
	}
	return resp, nil
}

// GetGameItem возвращает конкретный товар.
func (h *ProjectHandler) GetGameItem(ctx context.Context, req *pb.ProjectGetGameItemRequest) (*pb.ProjectGetGameItemResponse, error) {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}
	item, err := h.svc.GetGameItem(ctx, req.GetProjectId(), req.GetGameItemId(), userID)
	if err != nil {
		return nil, domainError(err, "get game item")
	}
	return &pb.ProjectGetGameItemResponse{Item: gameItemToProto(item)}, nil
}

// CreateGameItem создает новый товар для игры.
func (h *ProjectHandler) CreateGameItem(ctx context.Context, req *pb.ProjectCreateGameItemRequest) (*pb.ProjectCreateGameItemResponse, error) {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}

	item := &domain.GameItem{
		ProjectID:   req.GetProjectId(),
		GameItemID:  req.GetGameItemId(),
		Name:        req.GetName(),
		Description: req.GetDescription(),
		ImageURL:    req.GetImageUrl(),
		PriceCoins:  req.GetPriceCoins(),
		IsActive:    req.GetIsActive(),
	}

	created, err := h.svc.CreateGameItem(ctx, item, userID)
	if err != nil {
		return nil, domainError(err, "create game item")
	}
	return &pb.ProjectCreateGameItemResponse{Item: gameItemToProto(created)}, nil
}

// UpdateGameItem обновляет существующий товар.
func (h *ProjectHandler) UpdateGameItem(ctx context.Context, req *pb.ProjectUpdateGameItemRequest) (*pb.ProjectUpdateGameItemResponse, error) {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}

	item := &domain.GameItem{
		ProjectID:   req.GetProjectId(),
		GameItemID:  req.GetGameItemId(),
		Name:        req.GetName(),
		Description: req.GetDescription(),
		ImageURL:    req.GetImageUrl(),
		PriceCoins:  req.GetPriceCoins(),
		IsActive:    req.GetIsActive(),
	}

	updated, err := h.svc.UpdateGameItem(ctx, item, userID)
	if err != nil {
		return nil, domainError(err, "update game item")
	}
	return &pb.ProjectUpdateGameItemResponse{Item: gameItemToProto(updated)}, nil
}

// DeleteGameItem удаляет или деактивирует товар.
func (h *ProjectHandler) DeleteGameItem(ctx context.Context, req *pb.ProjectDeleteGameItemRequest) (*pb.ProjectDeleteGameItemResponse, error) {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}
	if err := h.svc.DeleteGameItem(ctx, req.GetProjectId(), req.GetGameItemId(), userID); err != nil {
		return nil, domainError(err, "delete game item")
	}
	return &pb.ProjectDeleteGameItemResponse{Success: true}, nil
}
