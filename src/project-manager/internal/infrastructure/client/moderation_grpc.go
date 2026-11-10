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
	client        modpb.ModerationServiceClient
	conn          *grpc.ClientConn
	internalToken string
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

// WithInternalToken задаёт общий секрет межсервисных вызовов (INTERNAL_API_KEY).
func (c *GRPCModerationClient) WithInternalToken(token string) *GRPCModerationClient {
	c.internalToken = token
	return c
}

// Close закрывает сетевое gRPC соединение с сервисом модерации.
func (c *GRPCModerationClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// forwardContext пробрасывает токен пользователя в сервис модерации и добавляет межсервисный токен.
// Личность пользователя модерация определяет сама по JWT: проброшенные x-user-* она перезаписывает
// значениями из токена.
func (c *GRPCModerationClient) forwardContext(ctx context.Context) context.Context {
	outgoingMD := metadata.MD{}
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		for _, key := range []string{"authorization", "x-user-id", "x-user-role", "x-user-name"} {
			if vals := md.Get(key); len(vals) > 0 {
				outgoingMD.Set(key, vals...)
			}
		}
	}
	if c.internalToken != "" {
		outgoingMD.Set("x-internal-token", c.internalToken)
	}
	return metadata.NewOutgoingContext(ctx, outgoingMD)
}

// SubmitDraft отправляет снимок черновика на модерацию через gRPC RPC.
func (c *GRPCModerationClient) SubmitDraft(ctx context.Context, snapshot *domain.ProjectSnapshot) (int64, error) {
	outCtx := c.forwardContext(ctx)

	var items []*modpb.GameItemSnapshot
	for _, itm := range snapshot.Items {
		if itm == nil {
			continue
		}
		items = append(items, &modpb.GameItemSnapshot{
			GameItemId:  itm.GameItemID,
			Name:        itm.Name,
			Description: itm.Description,
			PriceCoins:  itm.PriceCoins,
			ImageUrl:    itm.ImageURL,
			IsActive:    itm.IsActive,
		})
	}

	reqType := modpb.RequestType_REQUEST_TYPE_PROJECT_PUBLICATION
	if snapshot.IsUpdate {
		reqType = modpb.RequestType_REQUEST_TYPE_PROJECT_UPDATE
	}

	req := &modpb.SubmitDraftRequest{
		ProjectId: snapshot.ProjectID,
		OwnerId:   snapshot.OwnerID,
		Type:      reqType,
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
			IsOnline:           snapshot.IsOnline,
			Items:              items,
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
	outCtx := c.forwardContext(ctx)
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

	rawStatus := int32(req.GetStatus())
	var status int16
	if rawStatus >= -32768 && rawStatus <= 32767 {
		status = int16(rawStatus)
	}

	return &domain.ModerationRequestInfo{
		RequestID:       req.GetId(),
		ProjectID:       req.GetProjectId(),
		Status:          status,
		RejectionReason: req.GetRejectionReason(),
	}, nil
}
