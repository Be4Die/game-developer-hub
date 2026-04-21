// Package orchestrator реализует gRPC-клиент для связи с оркестратором.
package orchestrator

import (
	"context"
	"fmt"
	"time"

	pb "github.com/Be4Die/game-developer-hub/protos/orchestrator/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client предоставляет методы для взаимодействия с оркестратором.
type Client struct {
	conn   *grpc.ClientConn
	nodePb pb.NodeServiceClient
}

// NewClient создаёт новый клиент для подключения к оркестратору.
// Возвращает ошибку если не удалось установить соединение.
func NewClient(_ context.Context, address string, _ time.Duration) (*Client, error) {
	conn, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("orchestrator.NewClient: create client for %s: %w", address, err)
	}

	return &Client{
		conn:   conn,
		nodePb: pb.NewNodeServiceClient(conn),
	}, nil
}

// Close закрывает соединение с оркестратором.
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// AnnounceNode отправляет запрос на анонсирование ноды в оркестраторе.
// Возвращает ID созданной ноды.
func (c *Client) AnnounceNode(ctx context.Context, req *AnnounceRequest) (*AnnounceResponse, error) {
	pbReq := &pb.NodeServiceAnnounceRequest{
		Address:          req.Address,
		AgentVersion:     req.AgentVersion,
		CpuCores:         req.CPUCores,
		TotalMemoryBytes: req.TotalMemoryBytes,
		TotalDiskBytes:   req.TotalDiskBytes,
		ApiKey:           req.APIKey,
	}

	if req.Region != "" {
		pbReq.Region = &req.Region
	}

	resp, err := c.nodePb.Announce(ctx, pbReq)
	if err != nil {
		return nil, fmt.Errorf("orchestrator.Client.AnnounceNode: %w", err)
	}

	return &AnnounceResponse{
		NodeID: resp.GetNodeId(),
	}, nil
}

// AnnounceRequest содержит данные для анонсирования ноды.
type AnnounceRequest struct {
	Address          string
	Region           string
	AgentVersion     string
	CPUCores         uint32
	TotalMemoryBytes uint64
	TotalDiskBytes   uint64
	APIKey           string // NODE_API_KEY — используется как токен авторизации
}

// AnnounceResponse содержит результат анонсирования ноды.
type AnnounceResponse struct {
	NodeID int64
}
