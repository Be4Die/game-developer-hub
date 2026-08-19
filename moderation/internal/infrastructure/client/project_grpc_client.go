// Package client содержит реализации клиентов к внешним сервисам.
package client

import (
	"context"
	"fmt"

	pmpb "github.com/Be4Die/game-developer-hub/protos/project_manager/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ProjectGRPCClient реализует domain.ProjectClient через gRPC вызовы к сервису project-manager.
type ProjectGRPCClient struct {
	client pmpb.ProjectServiceClient
	conn   *grpc.ClientConn
}

// NewProjectGRPCClient подключается к сервису project-manager по gRPC.
func NewProjectGRPCClient(addr string) (*ProjectGRPCClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("connect to project-manager: %w", err)
	}

	return &ProjectGRPCClient{
		client: pmpb.NewProjectServiceClient(conn),
		conn:   conn,
	}, nil
}

// Close закрывает gRPC соединение.
func (c *ProjectGRPCClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// PublishRelease инициирует публикацию одобренного релиза в продуктивное окружение.
func (c *ProjectGRPCClient) PublishRelease(ctx context.Context, projectID int64, version, approvedBy, comment string) (string, error) {
	resp, err := c.client.PublishRelease(ctx, &pmpb.ProjectPublishReleaseRequest{
		ProjectId:   projectID,
		Version:     version,
		PublishedBy: approvedBy,
		Comment:     comment,
	})
	if err != nil {
		return "", fmt.Errorf("ProjectGRPCClient.PublishRelease: %w", err)
	}

	if resp.GetRelease() != nil {
		return resp.GetRelease().GetProdUrl(), nil
	}
	return "", nil
}

// RejectDraft возвращает черновик проекта на доработку.
func (c *ProjectGRPCClient) RejectDraft(ctx context.Context, projectID int64, reason string) error {
	_, err := c.client.RejectDraft(ctx, &pmpb.ProjectRejectDraftRequest{
		ProjectId: projectID,
		Reason:    reason,
	})
	if err != nil {
		return fmt.Errorf("ProjectGRPCClient.RejectDraft: %w", err)
	}
	return nil
}
