package grpc

import (
	"context"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/service"
	pb "github.com/Be4Die/game-developer-hub/protos/orchestrator/v1"
)

// InstanceHandler реализует InstanceService.
type InstanceHandler struct {
	pb.UnimplementedInstanceServiceServer
	instanceService *service.InstanceService
	maxLogTailLines uint32
}

// NewInstanceHandler создаёт обработчик инстансов.
func NewInstanceHandler(svc *service.InstanceService, maxLogTailLines uint32) *InstanceHandler {
	return &InstanceHandler{instanceService: svc, maxLogTailLines: maxLogTailLines}
}

// Start запускает инстанс сервера.
func (h *InstanceHandler) Start(ctx context.Context, req *pb.StartInstanceRequest) (*pb.StartInstanceResponse, error) {
	params := service.StartInstanceParams{
		GameID:           req.GetGameId(),
		BuildVersion:     req.GetBuildVersion(),
		Name:             req.GetName(),
		EnvVars:          req.GetEnvVars(),
		Args:             req.GetArgs(),
		DeveloperPayload: req.GetDeveloperPayload(),
	}

	if req.MaxPlayers != nil {
		mp := req.GetMaxPlayers()
		params.MaxPlayers = &mp
	}

	if req.PortAllocation != nil {
		params.PortAllocation = portAllocationFromProto(req.GetPortAllocation())
	}
	if req.ResourceLimits != nil {
		params.ResourceLimits = resourceLimitsFromProto(req.GetResourceLimits())
	}

	instance, err := h.instanceService.StartInstance(ctx, params)
	if err != nil {
		return nil, domainError(err, "start instance")
	}

	return &pb.StartInstanceResponse{Instance: instanceToProto(instance)}, nil
}

// List возвращает список инстансов игры.
func (h *InstanceHandler) List(ctx context.Context, req *pb.ListInstancesRequest) (*pb.ListInstancesResponse, error) {
	var statusFilter *domain.InstanceStatus
	if req.Status != nil {
		s := instanceStatusFromProto(req.GetStatus())
		statusFilter = &s
	}

	instances, err := h.instanceService.ListInstances(ctx, req.GetGameId(), statusFilter)
	if err != nil {
		return nil, domainError(err, "list instances")
	}

	resp := make([]*pb.Instance, 0, len(instances))
	for _, inst := range instances {
		resp = append(resp, enrichedInstanceToProto(inst))
	}

	return &pb.ListInstancesResponse{Instances: resp}, nil
}

// Get возвращает информацию об инстансе.
func (h *InstanceHandler) Get(ctx context.Context, req *pb.GetInstanceRequest) (*pb.GetInstanceResponse, error) {
	instance, err := h.instanceService.GetInstance(ctx, req.GetGameId(), req.GetInstanceId())
	if err != nil {
		return nil, domainError(err, "get instance")
	}

	return &pb.GetInstanceResponse{Instance: enrichedInstanceToProto(instance)}, nil
}

// Stop останавливает инстанс.
func (h *InstanceHandler) Stop(ctx context.Context, req *pb.StopInstanceRequest) (*pb.StopInstanceResponse, error) {
	timeout := req.GetTimeout()
	if timeout == 0 {
		timeout = 30
	}

	instance, err := h.instanceService.StopInstance(ctx, req.GetGameId(), req.GetInstanceId(), uint32(timeout))
	if err != nil {
		return nil, domainError(err, "stop instance")
	}

	return &pb.StopInstanceResponse{Instance: instanceToProto(instance)}, nil
}

// StreamLogs стримит логи инстанса.
func (h *InstanceHandler) StreamLogs(req *pb.StreamLogsRequest, stream pb.InstanceService_StreamLogsServer) error {
	ctx := stream.Context()

	follow := req.GetFollow()
	tail := uint32(req.GetTail()) //nolint:gosec // tail всегда положительный из proto
	if tail <= 0 {
		tail = 100
	}
	if tail > h.maxLogTailLines {
		tail = h.maxLogTailLines
	}

	var source *domain.LogSource
	if req.Source != nil {
		s := logSourceFromProto(req.GetSource())
		source = &s
	}

	logStream, err := h.instanceService.StreamInstanceLogs(ctx, req.GetGameId(), req.GetInstanceId(), domain.StreamLogsRequest{
		InstanceID:   req.GetInstanceId(),
		FollowStdout: follow,
		FollowStderr: follow,
		Tail:         tail,
	})
	if err != nil {
		return domainError(err, "stream logs")
	}
	defer func() { _ = logStream.Close() }()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		entry, err := logStream.Recv()
		if err != nil {
			return nil // Конец потока.
		}

		if source != nil && entry.Source != *source {
			continue
		}

		if err := stream.Send(logEntryToProto(entry)); err != nil {
			return err
		}
	}
}

// GetUsage возвращает потребление ресурсов инстанса.
func (h *InstanceHandler) GetUsage(ctx context.Context, req *pb.GetInstanceUsageRequest) (*pb.GetInstanceUsageResponse, error) {
	usage, err := h.instanceService.GetInstanceUsage(ctx, req.GetGameId(), req.GetInstanceId())
	if err != nil {
		return nil, domainError(err, "get instance usage")
	}

	return &pb.GetInstanceUsageResponse{
		InstanceId: req.GetInstanceId(),
		Usage:      resourceUsageToProto(usage),
	}, nil
}

// ─── Helper converters ──────────────────────────────────────────────────────

func portAllocationFromProto(pa *pb.PortAllocation) domain.PortAllocation {
	var result domain.PortAllocation
	if pa == nil {
		return result
	}

	switch v := pa.Strategy.(type) {
	case *pb.PortAllocation_Any:
		result.Any = true
	case *pb.PortAllocation_Exact:
		result.Exact = v.Exact.GetPort()
	case *pb.PortAllocation_Range:
		result.Range = &domain.PortRange{
			Min: v.Range.GetMinPort(),
			Max: v.Range.GetMaxPort(),
		}
	}
	return result
}

func resourceLimitsFromProto(rl *pb.ResourceLimits) *domain.ResourceLimits {
	if rl == nil {
		return nil
	}
	var cpuMillis *uint32
	if rl.GetCpuMillis() > 0 {
		v := uint32(rl.GetCpuMillis()) //nolint:gosec // cpuMillis всегда положительный
		cpuMillis = &v
	}
	var memoryBytes *uint64
	if rl.GetMemoryBytes() > 0 {
		v := rl.GetMemoryBytes()
		memoryBytes = &v
	}
	return &domain.ResourceLimits{
		CPUMillis:   cpuMillis,
		MemoryBytes: memoryBytes,
	}
}
