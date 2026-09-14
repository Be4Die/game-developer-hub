package client

import (
	"context"
	"fmt"

	"github.com/Be4Die/game-developer-hub/moderation/internal/domain"
	orchpb "github.com/Be4Die/game-developer-hub/protos/orchestrator/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// OrchestratorGRPCClient реализует domain.OrchestratorClient через gRPC вызовы к сервису orchestrator.
type OrchestratorGRPCClient struct {
	client orchpb.NodeServiceClient
	conn   *grpc.ClientConn
}

// NewOrchestratorGRPCClient подключается к сервису orchestrator по gRPC.
func NewOrchestratorGRPCClient(addr string) (*OrchestratorGRPCClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("connect to orchestrator: %w", err)
	}

	return &OrchestratorGRPCClient{
		client: orchpb.NewNodeServiceClient(conn),
		conn:   conn,
	}, nil
}

// Close закрывает gRPC соединение.
func (c *OrchestratorGRPCClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// GrantPlatformAccess выдает квоту серверов платформы проекту.
func (c *OrchestratorGRPCClient) GrantPlatformAccess(
	ctx context.Context,
	projectID int64,
	maxInstances int32,
	maxTotalCPU uint32,
	maxTotalMemoryMB uint64,
	maxInstanceCPU uint32,
	maxInstanceMemoryMB uint64,
) error {
	outCtx := forwardContext(ctx, "system-moderation", "3")
	_, err := c.client.GrantPlatformAccess(outCtx, &orchpb.GrantPlatformAccessRequest{
		ProjectId:            projectID,
		MaxInstances:         maxInstances,
		MaxTotalCpuMillis:    maxTotalCPU,
		MaxTotalMemoryMb:     maxTotalMemoryMB,
		MaxInstanceCpuMillis: maxInstanceCPU,
		MaxInstanceMemoryMb:  maxInstanceMemoryMB,
	})
	if err != nil {
		return fmt.Errorf("OrchestratorGRPCClient.GrantPlatformAccess: %w", err)
	}
	return nil
}

// RevokePlatformAccess отзывает доступ к платформенным мощностям проекта.
func (c *OrchestratorGRPCClient) RevokePlatformAccess(ctx context.Context, projectID int64) error {
	outCtx := forwardContext(ctx, "system-moderation", "3")
	_, err := c.client.RevokePlatformAccess(outCtx, &orchpb.RevokePlatformAccessRequest{
		ProjectId: projectID,
	})
	if err != nil {
		return fmt.Errorf("OrchestratorGRPCClient.RevokePlatformAccess: %w", err)
	}
	return nil
}

var _ domain.OrchestratorClient = (*OrchestratorGRPCClient)(nil)
