package grpc

import (
	"bufio"
	"context"
	"errors"
	"io"
	"time"

	"github.com/Be4Die/game-developer-hub/game-server-node/internal/domain"
	"github.com/Be4Die/game-developer-hub/game-server-node/internal/service"
	pb "github.com/Be4Die/game-developer-hub/protos/game_server_node/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// DeploymentHandler обрабатывает gRPC-запросы к сервису развёртывания.
type DeploymentHandler struct {
	pb.UnimplementedDeploymentServiceServer
	svc *service.DeploymentService
}

// NewDeploymentHandler создаёт обработчик для сервиса развёртывания.
func NewDeploymentHandler(svc *service.DeploymentService) *DeploymentHandler {
	return &DeploymentHandler{svc: svc}
}

// BuildImage собирает Docker-образ из исходного архива на стороне ноды.
func (h *DeploymentHandler) BuildImage(stream pb.DeploymentService_BuildImageServer) error {
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
	internalPort := meta.GetInternalPort()

	pr, pw := io.Pipe()
	errCh := make(chan error, 1)

	go func() {
		defer func() { _ = pw.Close() }()

		for {
			req, err := stream.Recv()
			if errors.Is(err, io.EOF) {
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

	if err := h.svc.BuildImage(stream.Context(), gameID, imageTag, internalPort, pr); err != nil {
		return status.Errorf(codes.Internal, "failed to build image: %v", err)
	}

	if err := <-errCh; err != nil {
		return status.Errorf(codes.Internal, "failed reading chunks: %v", err)
	}

	return stream.SendAndClose(&pb.BuildImageResponse{
		ImageTag: imageTag,
	})
}

// LoadImage загружает образ через потоковую передачу чанков.
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
		defer func() { _ = pw.Close() }()

		for {
			req, err := stream.Recv()
			if errors.Is(err, io.EOF) {
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

// StartInstance создаёт и запускает новый игровой инстанс.
func (h *DeploymentHandler) StartInstance(
	ctx context.Context,
	req *pb.StartInstanceRequest,
) (*pb.StartInstanceResponse, error) {

	opts := service.StartInstanceOpts{
		GameID:           req.GetGameId(),
		InstanceID:       req.GetInstanceId(), // Может быть 0 — тогда нода генерирует сама
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

// StopInstance останавливает инстанс и удаляет его контейнер.
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

// RestartInstance перезапускает работающий инстанс.
func (h *DeploymentHandler) RestartInstance(
	ctx context.Context,
	req *pb.RestartInstanceRequest,
) (*pb.RestartInstanceResponse, error) {
	timeout := time.Duration(req.GetTimeoutSeconds()) * time.Second

	if err := h.svc.RestartInstance(ctx, req.GetInstanceId(), timeout); err != nil {
		return nil, domainErrToStatus(err)
	}

	return &pb.RestartInstanceResponse{}, nil
}

// StartStoppedInstance запускает остановленный инстанс.
func (h *DeploymentHandler) StartStoppedInstance(
	ctx context.Context,
	req *pb.StartStoppedInstanceRequest,
) (*pb.StartStoppedInstanceResponse, error) {
	if err := h.svc.StartStoppedInstance(ctx, req.GetInstanceId()); err != nil {
		return nil, domainErrToStatus(err)
	}

	return &pb.StartStoppedInstanceResponse{}, nil
}

// DeleteInstance удаляет инстанс и его контейнер.
func (h *DeploymentHandler) DeleteInstance(
	ctx context.Context,
	req *pb.DeleteInstanceRequest,
) (*pb.DeleteInstanceResponse, error) {
	if err := h.svc.DeleteInstance(ctx, req.GetInstanceId()); err != nil {
		return nil, domainErrToStatus(err)
	}

	return &pb.DeleteInstanceResponse{}, nil
}

// StreamLogs возвращает поток логов контейнера.
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
	defer func() { _ = rc.Close() }() // Always close the reader!

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

// DeployService разворачивает управляемый сервис на ноде.
func (h *DeploymentHandler) DeployService(ctx context.Context, req *pb.DeployServiceRequest) (*pb.DeployServiceResponse, error) {
	result, err := h.svc.DeployService(ctx, domain.DeployServiceRequest{
		ServiceType: protoToDomainServiceType(req.GetServiceType()),
		Name:        req.GetName(),
		Port:        req.GetPort(),
		EnvVars:     req.GetEnvVars(),
		VolumeName:  req.GetVolumeName(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "deploy service: %v", err)
	}

	return &pb.DeployServiceResponse{
		Name:          result.Name,
		ContainerId:   result.ContainerID,
		HostPort:      result.HostPort,
		ConnectionUri: result.ConnectionURI,
		VolumePath:    result.VolumePath,
	}, nil
}

// RemoveService останавливает и удаляет управляемый сервис.
func (h *DeploymentHandler) RemoveService(ctx context.Context, req *pb.RemoveServiceRequest) (*pb.RemoveServiceResponse, error) {
	if err := h.svc.RemoveService(ctx, req.GetName(), req.GetDeleteVolume()); err != nil {
		return nil, status.Errorf(codes.Internal, "remove service: %v", err)
	}
	return &pb.RemoveServiceResponse{}, nil
}

// StopService останавливает контейнер сервиса без удаления тома.
func (h *DeploymentHandler) StopService(ctx context.Context, req *pb.StopServiceRequest) (*pb.StopServiceResponse, error) {
	if err := h.svc.StopService(ctx, req.GetName()); err != nil {
		return nil, status.Errorf(codes.Internal, "stop service: %v", err)
	}
	return &pb.StopServiceResponse{}, nil
}

// StartService запускает ранее остановленный сервис.
func (h *DeploymentHandler) StartService(ctx context.Context, req *pb.StartServiceRequest) (*pb.StartServiceResponse, error) {
	port, uri, err := h.svc.StartService(ctx, req.GetName())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "start service: %v", err)
	}
	return &pb.StartServiceResponse{
		HostPort:      port,
		ConnectionUri: uri,
	}, nil
}

// ListServices возвращает список развернутых управляемых сервисов.
func (h *DeploymentHandler) ListServices(ctx context.Context, _ *pb.ListServicesRequest) (*pb.ListServicesResponse, error) {
	services, err := h.svc.ListServices(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list services: %v", err)
	}

	pbServices := make([]*pb.ServiceInfo, 0, len(services))
	for _, s := range services {
		pbServices = append(pbServices, &pb.ServiceInfo{
			Name:            s.Name,
			ServiceType:     domainToProtoServiceType(s.ServiceType),
			ContainerId:     s.ContainerID,
			Status:          s.Status,
			HostPort:        s.HostPort,
			VolumePath:      s.VolumePath,
			VolumeSizeBytes: s.VolumeSizeBytes,
			AutoBackupEnabled: s.AutoBackupEnabled,
		})
	}

	return &pb.ListServicesResponse{Services: pbServices}, nil
}

func protoToDomainServiceType(t pb.ServiceType) domain.ServiceType {
	switch t {
	case pb.ServiceType_SERVICE_TYPE_POSTGRES:
		return domain.ServiceTypePostgres
	case pb.ServiceType_SERVICE_TYPE_REDIS:
		return domain.ServiceTypeRedis
	case pb.ServiceType_SERVICE_TYPE_MYSQL:
		return domain.ServiceTypeMySQL
	case pb.ServiceType_SERVICE_TYPE_VOLUME:
		return domain.ServiceTypeVolume
	case pb.ServiceType_SERVICE_TYPE_ADMINER:
		return domain.ServiceTypeAdminer
	case pb.ServiceType_SERVICE_TYPE_PGADMIN:
		return domain.ServiceTypePGAdmin
	default:
		return domain.ServiceTypeUnspecified
	}
}

func domainToProtoServiceType(t domain.ServiceType) pb.ServiceType {
	switch t {
	case domain.ServiceTypePostgres:
		return pb.ServiceType_SERVICE_TYPE_POSTGRES
	case domain.ServiceTypeRedis:
		return pb.ServiceType_SERVICE_TYPE_REDIS
	case domain.ServiceTypeMySQL:
		return pb.ServiceType_SERVICE_TYPE_MYSQL
	case domain.ServiceTypeVolume:
		return pb.ServiceType_SERVICE_TYPE_VOLUME
	case domain.ServiceTypeAdminer:
		return pb.ServiceType_SERVICE_TYPE_ADMINER
	case domain.ServiceTypePGAdmin:
		return pb.ServiceType_SERVICE_TYPE_PGADMIN
	default:
		return pb.ServiceType_SERVICE_TYPE_UNSPECIFIED
	}
}

// CreateBackup создает снимок данных сервиса.
func (h *DeploymentHandler) CreateBackup(ctx context.Context, req *pb.CreateBackupRequest) (*pb.CreateBackupResponse, error) {
	backup, err := h.svc.CreateBackup(ctx, req.GetServiceName())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create backup: %v", err)
	}

	return &pb.CreateBackupResponse{
		Backup: backupToProto(backup),
	}, nil
}

// ListBackups возвращает список бэкапов сервиса.
func (h *DeploymentHandler) ListBackups(ctx context.Context, req *pb.ListBackupsRequest) (*pb.ListBackupsResponse, error) {
	backups, err := h.svc.ListBackups(ctx, req.GetServiceName())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list backups: %v", err)
	}

	pbBackups := make([]*pb.BackupInfo, 0, len(backups))
	for _, b := range backups {
		pbBackups = append(pbBackups, backupToProto(&b))
	}

	return &pb.ListBackupsResponse{Backups: pbBackups}, nil
}

// RestoreBackup восстанавливает состояние сервиса из бэкапа.
func (h *DeploymentHandler) RestoreBackup(ctx context.Context, req *pb.RestoreBackupRequest) (*pb.RestoreBackupResponse, error) {
	if err := h.svc.RestoreBackup(ctx, req.GetServiceName(), req.GetBackupId()); err != nil {
		return nil, status.Errorf(codes.Internal, "restore backup: %v", err)
	}

	return &pb.RestoreBackupResponse{
		Success: true,
		Message: "Backup restored successfully",
	}, nil
}

// DeleteBackup удаляет файл бэкапа.
func (h *DeploymentHandler) DeleteBackup(ctx context.Context, req *pb.DeleteBackupRequest) (*pb.DeleteBackupResponse, error) {
	if err := h.svc.DeleteBackup(ctx, req.GetServiceName(), req.GetBackupId()); err != nil {
		return nil, status.Errorf(codes.Internal, "delete backup: %v", err)
	}

	return &pb.DeleteBackupResponse{}, nil
}

// DownloadBackup стримит файл бэкапа чанками по 64 КБ.
func (h *DeploymentHandler) DownloadBackup(req *pb.DownloadBackupRequest, stream pb.DeploymentService_DownloadBackupServer) error {
	r, _, _, err := h.svc.OpenBackup(stream.Context(), req.GetServiceName(), req.GetBackupId())
	if err != nil {
		return status.Errorf(codes.NotFound, "open backup: %v", err)
	}
	defer r.Close()

	buf := make([]byte, 64*1024)
	for {
		n, readErr := r.Read(buf)
		if n > 0 {
			if err := stream.Send(&pb.BackupChunk{Chunk: buf[:n]}); err != nil {
				return status.Errorf(codes.Internal, "stream backup chunk: %v", err)
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return status.Errorf(codes.Internal, "read backup file: %v", readErr)
		}
	}

	return nil
}

// UploadBackup принимает файл бэкапа стримом чанков.
func (h *DeploymentHandler) UploadBackup(stream pb.DeploymentService_UploadBackupServer) error {
	first, err := stream.Recv()
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "receive metadata: %v", err)
	}

	meta := first.GetMetadata()
	if meta == nil {
		return status.Errorf(codes.InvalidArgument, "first chunk must contain metadata")
	}

	pipeReader, pipeWriter := io.Pipe()
	go func() {
		defer pipeWriter.Close()
		for {
			chunk, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				_ = pipeWriter.CloseWithError(err)
				return
			}
			if data := chunk.GetChunk(); len(data) > 0 {
				if _, err := pipeWriter.Write(data); err != nil {
					return
				}
			}
		}
	}()

	info, err := h.svc.UploadBackup(stream.Context(), meta.GetServiceName(), meta.GetFileName(), meta.GetRestoreImmediately(), pipeReader)
	if err != nil {
		return status.Errorf(codes.Internal, "upload backup: %v", err)
	}

	return stream.SendAndClose(&pb.UploadBackupResponse{
		Backup:   backupToProto(info),
		Restored: meta.GetRestoreImmediately(),
	})
}

func backupToProto(b *domain.BackupInfo) *pb.BackupInfo {
	if b == nil {
		return nil
	}
	return &pb.BackupInfo{
		BackupId:    b.BackupID,
		ServiceName: b.ServiceName,
		FileName:    b.FileName,
		SizeBytes:   b.SizeBytes,
		BackupType:  pb.BackupType(b.BackupType),
		Status:      pb.BackupStatus(b.Status),
		Checksum:    b.Checksum,
		CreatedAt:   timestamppb.New(b.CreatedAt),
	}
}



// ToggleServiceAutoBackup включает или выключает авторасписание для сервиса.
func (h *DeploymentHandler) ToggleServiceAutoBackup(ctx context.Context, req *pb.ToggleServiceAutoBackupRequest) (*pb.ToggleServiceAutoBackupResponse, error) {
	if req.ServiceName == "" {
		return nil, status.Error(codes.InvalidArgument, "service_name required")
	}

	err := h.svc.ToggleServiceAutoBackup(req.ServiceName, req.Enabled)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "toggle auto backup: %v", err)
	}

	return &pb.ToggleServiceAutoBackupResponse{Success: true}, nil
}
