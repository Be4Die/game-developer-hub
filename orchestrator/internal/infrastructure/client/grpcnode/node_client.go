// Package grpcnode реализует gRPC-клиент для подключения к game-server-node.
package grpcnode

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/infrastructure/config"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
	pb "github.com/Be4Die/game-developer-hub/protos/game_server_node/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
)

// Client реализует domain.NodeClient через gRPC к game-server-node.
type Client struct {
	cfg   config.GRPCClientConfig
	conns map[string]*grpc.ClientConn
	mu    sync.Mutex
}

// New создаёт gRPC-клиент для управления нодами.
func New(cfg config.GRPCClientConfig) *Client {
	return &Client{
		cfg:   cfg,
		conns: make(map[string]*grpc.ClientConn),
	}
}

// authContext добавляет auth metadata в контекст.
func authContext(ctx context.Context, apiKey string) context.Context {
	if apiKey == "" {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+apiKey)
}

// getConn возвращает gRPC-соединение к ноде (кеширует).
func (c *Client) getConn(_ context.Context, address string) (*grpc.ClientConn, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if conn, ok := c.conns[address]; ok {
		return conn, nil
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(c.cfg.MaxMessageSize)),
		grpc.WithConnectParams(grpc.ConnectParams{
			MinConnectTimeout: c.cfg.ConnectTimeout,
		}),
	}

	if c.cfg.EnableCompression {
		opts = append(opts, grpc.WithDefaultCallOptions(grpc.UseCompressor(gzip.Name)))
	}

	if c.cfg.KeepAliveTime > 0 {
		ka := keepalive.ClientParameters{
			Time:                c.cfg.KeepAliveTime,
			Timeout:             c.cfg.KeepAliveTimeout,
			PermitWithoutStream: true,
		}
		opts = append(opts, grpc.WithKeepaliveParams(ka))
	}

	conn, err := grpc.NewClient(address, opts...) //nolint:staticcheck
	if err != nil {
		return nil, fmt.Errorf("grpc.NewClient(%s): %w", address, err)
	}

	c.conns[address] = conn
	return conn, nil
}

// LoadImage загружает Docker-образ на ноду через стрим.
func (c *Client) LoadImage(ctx context.Context, nodeAddress, apiKey string, metadata domain.ImageMetadata, chunks io.Reader) (*domain.ImageLoadResult, error) {
	conn, err := c.getConn(ctx, nodeAddress)
	if err != nil {
		return nil, err
	}

	ctx = authContext(ctx, apiKey)
	depClient := pb.NewDeploymentServiceClient(conn)
	stream, err := depClient.LoadImage(ctx)
	if err != nil {
		return nil, fmt.Errorf("LoadImage stream: %w", err)
	}

	// Отправляем метаданные первым сообщением.
	if err := stream.Send(&pb.LoadImageRequest{
		Payload: &pb.LoadImageRequest_Metadata{
			Metadata: &pb.ImageMetadata{
				GameId:   metadata.GameID,
				ImageTag: metadata.ImageTag,
			},
		},
	}); err != nil {
		return nil, fmt.Errorf("LoadImage metadata: %w", err)
	}

	// Чанками отправляем данные.
	buf := make([]byte, 64*1024) // 64 KB chunks
	for {
		n, readErr := chunks.Read(buf)
		if n > 0 {
			chunk := make([]byte, n)
			copy(chunk, buf[:n])
			if err := stream.Send(&pb.LoadImageRequest{
				Payload: &pb.LoadImageRequest_Chunk{
					Chunk: chunk,
				},
			}); err != nil {
				return nil, fmt.Errorf("LoadImage chunk: %w", err)
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return nil, fmt.Errorf("LoadImage read: %w", readErr)
		}
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		return nil, fmt.Errorf("LoadImage close: %w", err)
	}

	return &domain.ImageLoadResult{
		ImageTag:  resp.ImageTag,
		SizeBytes: resp.SizeBytes,
	}, nil
}

// BuildImage отправляет исходный архив на ноду для сборки Docker-образа.
func (c *Client) BuildImage(ctx context.Context, nodeAddress, apiKey string, metadata domain.BuildImageMetadata, archive io.Reader) error {
	conn, err := c.getConn(ctx, nodeAddress)
	if err != nil {
		return err
	}

	ctx = authContext(ctx, apiKey)
	depClient := pb.NewDeploymentServiceClient(conn)
	stream, err := depClient.BuildImage(ctx)
	if err != nil {
		return fmt.Errorf("BuildImage stream: %w", err)
	}

	// Отправляем метаданные первым сообщением.
	if err := stream.Send(&pb.BuildImageRequest{
		Payload: &pb.BuildImageRequest_Metadata{
			Metadata: &pb.BuildImageMetadata{
				GameId:       metadata.GameID,
				ImageTag:     metadata.ImageTag,
				InternalPort: metadata.InternalPort,
			},
		},
	}); err != nil {
		return fmt.Errorf("BuildImage metadata: %w", err)
	}

	// Чанками отправляем архив.
	buf := make([]byte, 64*1024) // 64 KB chunks
	for {
		n, readErr := archive.Read(buf)
		if n > 0 {
			chunk := make([]byte, n)
			copy(chunk, buf[:n])
			if err := stream.Send(&pb.BuildImageRequest{
				Payload: &pb.BuildImageRequest_Chunk{
					Chunk: chunk,
				},
			}); err != nil {
				return fmt.Errorf("BuildImage chunk: %w", err)
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return fmt.Errorf("BuildImage read: %w", readErr)
		}
	}

	if _, err := stream.CloseAndRecv(); err != nil {
		return fmt.Errorf("BuildImage close: %w", err)
	}

	return nil
}

// StartInstance запускает экземпляр игрового сервера на ноде.
func (c *Client) StartInstance(ctx context.Context, nodeAddress, apiKey string, req domain.StartInstanceRequest) (*domain.StartInstanceResult, error) {
	conn, err := c.getConn(ctx, nodeAddress)
	if err != nil {
		return nil, err
	}

	ctx = authContext(ctx, apiKey)
	depClient := pb.NewDeploymentServiceClient(conn)

	pbReq := &pb.StartInstanceRequest{
		GameId:           req.GameID,
		Name:             req.Name,
		Protocol:         toPBProtocol(req.Protocol),
		InternalPort:     req.InternalPort,
		PortAllocation:   toPBPortAllocation(req.PortAllocation),
		MaxPlayers:       req.MaxPlayers,
		DeveloperPayload: req.DeveloperPayload,
		EnvVars:          req.EnvVars,
		Args:             req.Args,
	}

	if req.ResourceLimits != nil {
		pbReq.ResourceLimits = &pb.ResourceLimits{}
		if req.ResourceLimits.CPUMillis != nil {
			pbReq.ResourceLimits.CpuMillis = req.ResourceLimits.CPUMillis
		}
		if req.ResourceLimits.MemoryBytes != nil {
			pbReq.ResourceLimits.MemoryBytes = req.ResourceLimits.MemoryBytes
		}
	}

	resp, err := depClient.StartInstance(ctx, pbReq)
	if err != nil {
		return nil, fmt.Errorf("StartInstance: %w", err)
	}

	return &domain.StartInstanceResult{
		InstanceID: resp.InstanceId,
		HostPort:   resp.HostPort,
	}, nil
}

// StopInstance выполняет graceful остановку экземпляра.
func (c *Client) StopInstance(ctx context.Context, nodeAddress, apiKey string, instanceID int64, timeoutSec uint32) error {
	conn, err := c.getConn(ctx, nodeAddress)
	if err != nil {
		return err
	}

	ctx = authContext(ctx, apiKey)
	depClient := pb.NewDeploymentServiceClient(conn)
	_, err = depClient.StopInstance(ctx, &pb.StopInstanceRequest{
		InstanceId:     instanceID,
		TimeoutSeconds: timeoutSec,
	})
	if err != nil {
		return fmt.Errorf("StopInstance: %w", err)
	}

	return nil
}

// StreamLogs открывает поток журналов инстанса.
func (c *Client) StreamLogs(ctx context.Context, nodeAddress, apiKey string, req domain.StreamLogsRequest) (domain.LogStream, error) {
	conn, err := c.getConn(ctx, nodeAddress)
	if err != nil {
		return nil, err
	}

	ctx = authContext(ctx, apiKey)
	depClient := pb.NewDeploymentServiceClient(conn)
	pbReq := &pb.StreamLogsRequest{
		InstanceId:   req.InstanceID,
		FollowStdout: req.FollowStdout,
		FollowStderr: req.FollowStderr,
		Tail:         req.Tail,
	}

	stream, err := depClient.StreamLogs(ctx, pbReq)
	if err != nil {
		return nil, fmt.Errorf("StreamLogs: %w", err)
	}

	return &logStream{stream: stream}, nil
}

// GetNodeInfo запрашивает статические характеристики ноды.
func (c *Client) GetNodeInfo(ctx context.Context, nodeAddress, apiKey string) (*domain.NodeInfo, error) {
	conn, err := c.getConn(ctx, nodeAddress)
	if err != nil {
		return nil, err
	}

	ctx = authContext(ctx, apiKey)
	discClient := pb.NewDiscoveryServiceClient(conn)
	resp, err := discClient.GetNodeInfo(ctx, &pb.GetNodeInfoRequest{})
	if err != nil {
		return nil, fmt.Errorf("GetNodeInfo: %w", err)
	}

	return &domain.NodeInfo{
		Region:           resp.Region,
		CPUCores:         resp.CpuCores,
		TotalMemoryBytes: resp.TotalMemoryBytes,
		TotalDiskBytes:   resp.TotalDiskBytes,
		NetworkBandwidth: resp.NetworkBandwidthBytesPerSec,
		AgentVersion:     resp.AgentVersion,
	}, nil
}

// Heartbeat получает текущую загруженность ноды.
func (c *Client) Heartbeat(ctx context.Context, nodeAddress, apiKey string) (*domain.ResourceUsage, error) {
	conn, err := c.getConn(ctx, nodeAddress)
	if err != nil {
		return nil, err
	}

	ctx = authContext(ctx, apiKey)
	discClient := pb.NewDiscoveryServiceClient(conn)
	resp, err := discClient.Heartbeat(ctx, &pb.HeartbeatRequest{})
	if err != nil {
		return nil, fmt.Errorf("Heartbeat: %w", err)
	}

	return &domain.ResourceUsage{
		CPUUsagePercent:    resp.Usage.CpuUsagePercent,
		MemoryUsedBytes:    resp.Usage.MemoryUsedBytes,
		DiskUsedBytes:      resp.Usage.DiskUsedBytes,
		NetworkBytesPerSec: resp.Usage.NetworkBytesPerSec,
	}, nil
}

// ListInstances возвращает все экземпляры на ноде.
func (c *Client) ListInstances(ctx context.Context, nodeAddress, apiKey string) ([]*domain.Instance, error) {
	conn, err := c.getConn(ctx, nodeAddress)
	if err != nil {
		return nil, err
	}

	ctx = authContext(ctx, apiKey)
	discClient := pb.NewDiscoveryServiceClient(conn)
	resp, err := discClient.ListInstances(ctx, &pb.ListInstancesRequest{})
	if err != nil {
		return nil, fmt.Errorf("ListInstances: %w", err)
	}

	result := make([]*domain.Instance, 0, len(resp.Instances))
	for _, inst := range resp.Instances {
		result = append(result, fromPBInstance(inst))
	}

	return result, nil
}

// GetInstance возвращает экземпляр по идентификатору.
func (c *Client) GetInstance(ctx context.Context, nodeAddress, apiKey string, instanceID int64) (*domain.Instance, error) {
	conn, err := c.getConn(ctx, nodeAddress)
	if err != nil {
		return nil, err
	}

	ctx = authContext(ctx, apiKey)
	discClient := pb.NewDiscoveryServiceClient(conn)
	resp, err := discClient.GetInstance(ctx, &pb.GetInstanceRequest{
		InstanceId: instanceID,
	})
	if err != nil {
		return nil, fmt.Errorf("GetInstance: %w", err)
	}

	return fromPBInstance(resp.Instance), nil
}

// GetInstanceUsage возвращает потребление ресурсов экземпляра.
func (c *Client) GetInstanceUsage(ctx context.Context, nodeAddress, apiKey string, instanceID int64) (*domain.ResourceUsage, error) {
	conn, err := c.getConn(ctx, nodeAddress)
	if err != nil {
		return nil, err
	}

	ctx = authContext(ctx, apiKey)
	discClient := pb.NewDiscoveryServiceClient(conn)
	resp, err := discClient.GetInstanceUsage(ctx, &pb.GetInstanceUsageRequest{
		InstanceId: instanceID,
	})
	if err != nil {
		return nil, fmt.Errorf("GetInstanceUsage: %w", err)
	}

	return &domain.ResourceUsage{
		CPUUsagePercent:    resp.Usage.CpuUsagePercent,
		MemoryUsedBytes:    resp.Usage.MemoryUsedBytes,
		DiskUsedBytes:      resp.Usage.DiskUsedBytes,
		NetworkBytesPerSec: resp.Usage.NetworkBytesPerSec,
	}, nil
}

// Close закрывает все gRPC-соединения.
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for addr, conn := range c.conns {
		if err := conn.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "grpcnode: close conn %s: %v\n", addr, err)
		}
	}
	c.conns = make(map[string]*grpc.ClientConn)
}

// ─── Маппинг доменных типов в proto ─────────────────────────────────────────

func toPBProtocol(p domain.Protocol) pb.Protocol {
	switch p {
	case domain.ProtocolTCP:
		return pb.Protocol_PROTOCOL_TCP
	case domain.ProtocolUDP:
		return pb.Protocol_PROTOCOL_UDP
	case domain.ProtocolWebSocket:
		return pb.Protocol_PROTOCOL_WEBSOCKET
	case domain.ProtocolWebRTC:
		return pb.Protocol_PROTOCOL_WEBRTC
	default:
		return pb.Protocol_PROTOCOL_UNSPECIFIED
	}
}

func fromPBProtocol(p pb.Protocol) domain.Protocol {
	switch p {
	case pb.Protocol_PROTOCOL_TCP:
		return domain.ProtocolTCP
	case pb.Protocol_PROTOCOL_UDP:
		return domain.ProtocolUDP
	case pb.Protocol_PROTOCOL_WEBSOCKET:
		return domain.ProtocolWebSocket
	case pb.Protocol_PROTOCOL_WEBRTC:
		return domain.ProtocolWebRTC
	default:
		return 0
	}
}

func toPBPortAllocation(pa domain.PortAllocation) *pb.PortAllocation {
	if pa.Any {
		return &pb.PortAllocation{
			Strategy: &pb.PortAllocation_Any{Any: true},
		}
	}
	if pa.Exact != 0 {
		return &pb.PortAllocation{
			Strategy: &pb.PortAllocation_Exact{Exact: pa.Exact},
		}
	}
	if pa.Range != nil {
		return &pb.PortAllocation{
			Strategy: &pb.PortAllocation_Range{
				Range: &pb.PortRange{
					MinPort: pa.Range.Min,
					MaxPort: pa.Range.Max,
				},
			},
		}
	}
	return &pb.PortAllocation{
		Strategy: &pb.PortAllocation_Any{Any: true},
	}
}

func fromPBInstance(inst *pb.Instance) *domain.Instance {
	result := &domain.Instance{
		ID:               inst.InstanceId,
		Name:             inst.Name,
		GameID:           inst.GameId,
		BuildVersion:     inst.BuildVersion,
		HostPort:         inst.Port,
		Protocol:         fromPBProtocol(inst.Protocol),
		Status:           fromPBInstanceStatus(inst.Status),
		MaxPlayers:       inst.MaxPlayers,
		DeveloperPayload: inst.DeveloperPayload,
	}

	if inst.PlayerCount != nil {
		pc := inst.GetPlayerCount()
		result.PlayerCount = &pc
	}

	if inst.StartedAt != nil {
		result.StartedAt = inst.StartedAt.AsTime()
	}

	return result
}

func fromPBInstanceStatus(s pb.InstanceStatus) domain.InstanceStatus {
	switch s {
	case pb.InstanceStatus_INSTANCE_STATUS_STARTING:
		return domain.InstanceStatusStarting
	case pb.InstanceStatus_INSTANCE_STATUS_RUNNING:
		return domain.InstanceStatusRunning
	case pb.InstanceStatus_INSTANCE_STATUS_STOPPING:
		return domain.InstanceStatusStopping
	case pb.InstanceStatus_INSTANCE_STATUS_STOPPED:
		return domain.InstanceStatusStopped
	case pb.InstanceStatus_INSTANCE_STATUS_CRASHED:
		return domain.InstanceStatusCrashed
	default:
		return 0
	}
}

// ─── LogStream адаптер ──────────────────────────────────────────────────────

type logStream struct {
	stream pb.DeploymentService_StreamLogsClient
}

func (ls *logStream) Recv() (*domain.LogEntry, error) {
	resp, err := ls.stream.Recv()
	if err != nil {
		return nil, err
	}

	return &domain.LogEntry{
		Timestamp: resp.Timestamp.AsTime(),
		Source:    fromPBLogSource(resp.Source),
		Message:   resp.Message,
	}, nil
}

func (ls *logStream) Close() error {
	// Stream закрывается автоматически при отмене контекста.
	return nil
}

func fromPBLogSource(s pb.LogSource) domain.LogSource {
	switch s {
	case pb.LogSource_LOG_SOURCE_STDOUT:
		return domain.LogSourceStdout
	case pb.LogSource_LOG_SOURCE_STDERR:
		return domain.LogSourceStderr
	default:
		return 0
	}
}

var _ domain.NodeClient = (*Client)(nil)
