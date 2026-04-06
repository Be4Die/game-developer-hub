package grpc

import (
	"bufio"
	"context"
	"io"
	"time"

	"github.com/Be4Die/game-developer-hub/game-server-node/internal/service"
	pb "github.com/Be4Die/game-developer-hub/protos/game_server_node/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type DeploymentHandler struct {
	pb.UnimplementedDeploymentServiceServer
	svc *service.DeploymentService
}

func NewDeploymentHandler(svc *service.DeploymentService) *DeploymentHandler {
	return &DeploymentHandler{svc: svc}
}

func (h *DeploymentHandler) LoadImage(stream pb.DeploymentService_LoadImageServer) error {
	first, err := stream.Recv()
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "failed to receive metadata: %v", err)
	}

	meta := first.GetMetadata()
	if meta == nil {
		return status.Errorf(codes.InvalidArgument, "first message must contain metadata")
	}

	gameID := meta.GetGameId()
	imageTag := meta.GetImageTag()

	pr, pw := io.Pipe()
	errCh := make(chan error, 1)

	go func() {
		defer pw.Close()

		for {
			req, err := stream.Recv()
			if err == io.EOF {
				errCh <- nil
				return
			}
			if err != nil {
				pw.CloseWithError(err)
				errCh <- err
				return
			}

			chunk := req.GetChunk()
			if len(chunk) == 0 {
				continue
			}

			if _, err := pw.Write(chunk); err != nil {
				errCh <- err
				return
			}
		}
	}()

	if err := h.svc.LoadImage(stream.Context(), gameID, imageTag, pr); err != nil {
		return status.Errorf(codes.Internal, "failed to load image: %v", err)
	}

	if err := <-errCh; err != nil {
		return status.Errorf(codes.Internal, "failed reading chunks: %v", err)
	}

	return stream.SendAndClose(&pb.LoadImageResponse{
		ImageTag: imageTag,
	})
}

func (h *DeploymentHandler) StartInstance(
	ctx context.Context,
	req *pb.StartInstanceRequest,
) (*pb.StartInstanceResponse, error) {

	opts := service.StartInstanceOpts{
		GameID:           req.GetGameId(),
		Name:             req.GetName(),
		Protocol:         protoToProtocol(req.GetProtocol()),
		InternalPort:     req.GetInternalPort(),
		PortStrategy:     protoToPortStrategy(req.GetPortAllocation()),
		MaxPlayers:       req.GetMaxPlayers(),
		DeveloperPayload: req.GetDeveloperPayload(),
		EnvVars:          req.GetEnvVars(),
		Args:             req.GetArgs(),
	}

	if rl := req.GetResourceLimits(); rl != nil {
		if rl.CpuMillis != nil {
			v := rl.GetCpuMillis()
			opts.CPUMillis = &v
		}
		if rl.MemoryBytes != nil {
			v := rl.GetMemoryBytes()
			opts.MemoryBytes = &v
		}
	}

	id, hostPort, err := h.svc.StartInstance(ctx, opts)
	if err != nil {
		return nil, domainErrToStatus(err)
	}

	return &pb.StartInstanceResponse{
		InstanceId: id,
		HostPort:   hostPort,
	}, nil
}

func (h *DeploymentHandler) StopInstance(
	ctx context.Context,
	req *pb.StopInstanceRequest,
) (*pb.StopInstanceResponse, error) {
	// time.Duration base unit is nanoseconds.
	// Multiply by time.Second to convert "30" → 30 seconds.
	timeout := time.Duration(req.GetTimeoutSeconds()) * time.Second

	if err := h.svc.StopInstance(ctx, req.GetInstanceId(), timeout); err != nil {
		return nil, domainErrToStatus(err)
	}

	return &pb.StopInstanceResponse{}, nil
}

func (h *DeploymentHandler) StreamLogs(
	req *pb.StreamLogsRequest,
	stream pb.DeploymentService_StreamLogsServer,
) error {
	// Determine if we need to follow (stream continuously).
	follow := req.GetFollowStdout() || req.GetFollowStderr()

	// Get log stream from service.
	rc, err := h.svc.StreamLogs(stream.Context(), req.GetInstanceId(), follow)
	if err != nil {
		return domainErrToStatus(err)
	}
	defer rc.Close() // Always close the reader!

	// Read line by line — don't split lines in the middle.
	scanner := bufio.NewScanner(rc)

	for scanner.Scan() {
		// Check if client disconnected.
		if err := stream.Context().Err(); err != nil {
			return nil // Client gone — stop gracefully.
		}

		line := scanner.Text()

		resp := &pb.StreamLogsResponse{
			Timestamp: timestamppb.Now(),
			Source:    pb.LogSource_LOG_SOURCE_STDOUT,
			Message:   line,
		}

		if err := stream.Send(resp); err != nil {
			return err
		}
	}

	// Scanner error (not EOF — EOF is normal).
	if err := scanner.Err(); err != nil {
		// Context cancelled = client disconnected, not an error.
		if stream.Context().Err() != nil {
			return nil
		}
		return status.Errorf(codes.Internal, "reading logs: %v", err)
	}

	return nil
}
