package client

import (
	"context"
	"fmt"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
	modpb "github.com/Be4Die/game-developer-hub/protos/moderation/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// GRPCModerationClient реализует domain.ModerationClient поверх gRPC вызовов к сервису moderation.
// Потокобезопасен для конкурентного использования.
type GRPCModerationClient struct {
	client modpb.ModerationServiceClient
	conn   *grpc.ClientConn
}

// NewGRPCModerationClient подключается к gRPC-сервису модерации по указанному адресу.
func NewGRPCModerationClient(addr string) (*GRPCModerationClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("connect to moderation service: %w", err)
	}

	return &GRPCModerationClient{
		client: modpb.NewModerationServiceClient(conn),
		conn:   conn,
	}, nil
}

// Close закрывает сетевое gRPC соединение с сервисом модерации.
func (c *GRPCModerationClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func forwardContext(ctx context.Context, fallbackUserID, fallbackRole string) context.Context {
	outgoingMD := metadata.MD{}
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		for _, key := range []string{"authorization", "x-user-id", "x-user-role", "x-user-name"} {
			if vals := md.Get(key); len(vals) > 0 {
				outgoingMD.Set(key, vals...)
			}
		}
	}
	if len(outgoingMD.Get("x-user-id")) == 0 && fallbackUserID != "" {
		outgoingMD.Set("x-user-id", fallbackUserID)
	}
	if len(outgoingMD.Get("x-user-role")) == 0 {
		if fallbackRole != "" {
			outgoingMD.Set("x-user-role", fallbackRole)
		} else {
			outgoingMD.Set("x-user-role", "developer")
		}
	}
	return metadata.NewOutgoingContext(ctx, outgoingMD)
}

// SubmitDraft отправляет снимок черновика на модерацию через gRPC RPC.
func (c *GRPCModerationClient) SubmitDraft(ctx context.Context, snapshot *domain.ProjectSnapshot) (int64, error) {
	outCtx := forwardContext(ctx, snapshot.OwnerID, "developer")
	req := &modpb.SubmitDraftRequest{
		ProjectId: snapshot.ProjectID,
		OwnerId:   snapshot.OwnerID,
		Snapshot: &modpb.ProjectSnapshot{
			ProjectId:          snapshot.ProjectID,
			TitleRu:            snapshot.TitleRu,
			TitleEn:            snapshot.TitleEn,
			SeoRu:              snapshot.SeoRu,
			SeoEn:              snapshot.SeoEn,
			AboutRu:            snapshot.AboutRu,
			AboutEn:            snapshot.AboutEn,
			IconPath:           snapshot.IconPath,
			CoverPath:          snapshot.CoverPath,
			VideoPath:          snapshot.VideoPath,
			ActiveBuildVersion: snapshot.ActiveBuildVersion,
			DevUrl:             snapshot.DevURL,
		},
	}

	resp, err := c.client.SubmitDraft(outCtx, req)
	if err != nil {
		return 0, fmt.Errorf("GRPCModerationClient.SubmitDraft: %w", err)
	}

	return resp.GetRequest().GetId(), nil
}

// GetLatestRequest запрашивает статус последней заявки проекта через gRPC RPC.
func (c *GRPCModerationClient) GetLatestRequest(ctx context.Context, projectID int64) (*domain.ModerationRequestInfo, error) {
	outCtx := forwardContext(ctx, "system", "moderator")
	resp, err := c.client.GetLatestRequestByProject(outCtx, &modpb.GetLatestRequestByProjectRequest{
		ProjectId: projectID,
	})
	if err != nil {
		return nil, fmt.Errorf("GRPCModerationClient.GetLatestRequest: %w", err)
	}

	req := resp.GetRequest()
	if req == nil {
		return nil, domain.ErrNotFound
	}

	return &domain.ModerationRequestInfo{
		RequestID:       req.GetId(),
		ProjectID:       req.GetProjectId(),
		Status:          int16(req.GetStatus()),
		RejectionReason: req.GetRejectionReason(),
	}, nil
}
