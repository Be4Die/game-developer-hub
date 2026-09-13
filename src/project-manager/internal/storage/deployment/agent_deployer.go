// Package deployment implements artifact deployment mechanisms.
package deployment

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
	pb "github.com/Be4Die/game-developer-hub/protos/game_deployment_agent/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// AgentDeployer реализует domain.Deployer для удаленного развертывания веб-сборок игр через gRPC агент.
type AgentDeployer struct {
	agentAddr     string
	apiKey        string
	client        pb.WebGameDeploymentServiceClient
	conn          *grpc.ClientConn
	archiveOpener func(ctx context.Context, projectID int64, version string, archivePath string) (io.ReadCloser, error)
}

// NewAgentDeployer создает клиент для взаимодействия с удаленным агентом развертывания.
func NewAgentDeployer(agentAddr, apiKey string) (*AgentDeployer, error) {
	conn, err := grpc.NewClient(agentAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("grpc dial agent %s: %w", agentAddr, err)
	}

	client := pb.NewWebGameDeploymentServiceClient(conn)

	return &AgentDeployer{
		agentAddr: agentAddr,
		apiKey:    apiKey,
		client:    client,
		conn:      conn,
	}, nil
}

// SetArchiveOpener задает кастомную функцию открытия архива (например, из S3).
func (d *AgentDeployer) SetArchiveOpener(fn func(ctx context.Context, projectID int64, version string, archivePath string) (io.ReadCloser, error)) {
	d.archiveOpener = fn
}

// Close закрывает сетевое gRPC соединение с агентом.
func (d *AgentDeployer) Close() error {
	if d.conn != nil {
		return d.conn.Close()
	}
	return nil
}

func (d *AgentDeployer) withAuth(ctx context.Context) context.Context {
	if d.apiKey == "" {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, "x-api-key", d.apiKey)
}

// DeployDev передает потоком архив сборки на удаленный агент для распаковки и переключения dev symlink.
func (d *AgentDeployer) DeployDev(ctx context.Context, projectID int64, version string, archivePath string) (*domain.DeploymentResult, error) {
	ctx = d.withAuth(ctx)

	stream, err := d.client.DeployDevStream(ctx)
	if err != nil {
		return nil, fmt.Errorf("agent_deployer: start dev stream: %w", err)
	}

	// 1. Отправляем метаданные
	err = stream.Send(&pb.DeployDevStreamRequest{
		Payload: &pb.DeployDevStreamRequest_Metadata{
			Metadata: &pb.DeployMetadata{
				ProjectId: projectID,
				Version:   version,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("agent_deployer: send metadata: %w", err)
	}

	// 2. Читаем архив из источника (диск или S3) и стримим чанками по 64 КБ
	var file io.ReadCloser
	if d.archiveOpener != nil {
		rc, openErr := d.archiveOpener(ctx, projectID, version, archivePath)
		if openErr != nil {
			return nil, fmt.Errorf("agent_deployer: open archive stream %s: %w", archivePath, openErr)
		}
		file = rc
	} else {
		f, openErr := os.Open(filepath.Clean(archivePath)) //nolint:gosec
		if openErr != nil {
			return nil, fmt.Errorf("agent_deployer: open archive %s: %w", archivePath, openErr)
		}
		file = f
	}
	defer func() {
		_ = file.Close()
	}()

	buf := make([]byte, 64*1024)
	for {
		n, readErr := file.Read(buf)
		if n > 0 {
			sendErr := stream.Send(&pb.DeployDevStreamRequest{
				Payload: &pb.DeployDevStreamRequest_Chunk{
					Chunk: buf[:n],
				},
			})
			if sendErr != nil {
				return nil, fmt.Errorf("agent_deployer: send chunk: %w", sendErr)
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return nil, fmt.Errorf("agent_deployer: read archive: %w", readErr)
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		return nil, fmt.Errorf("agent_deployer: close and recv: %w", err)
	}

	return &domain.DeploymentResult{
		URL:          res.Url,
		UnpackedPath: res.UnpackedPath,
		Success:      res.Success,
	}, nil
}

// DeployProd активирует версию в прод-окружении на агенте.
func (d *AgentDeployer) DeployProd(ctx context.Context, projectID int64, version string, _ string) (*domain.DeploymentResult, error) {
	ctx = d.withAuth(ctx)

	res, err := d.client.DeployProd(ctx, &pb.DeployProdRequest{
		ProjectId: projectID,
		Version:   version,
	})
	if err != nil {
		return nil, fmt.Errorf("agent_deployer: deploy prod: %w", err)
	}

	return &domain.DeploymentResult{
		URL:          res.Url,
		UnpackedPath: res.UnpackedPath,
		Success:      res.Success,
	}, nil
}

// UndeployProd снимает игру с публикации на агенте.
func (d *AgentDeployer) UndeployProd(ctx context.Context, projectID int64) error {
	ctx = d.withAuth(ctx)

	_, err := d.client.UndeployProd(ctx, &pb.UndeployProdRequest{
		ProjectId: projectID,
	})
	if err != nil {
		return fmt.Errorf("agent_deployer: undeploy prod: %w", err)
	}
	return nil
}

// DeleteVersion удаляет распакованные файлы версии на агенте (Garbage Collection).
func (d *AgentDeployer) DeleteVersion(ctx context.Context, projectID int64, version string) error {
	ctx = d.withAuth(ctx)

	_, err := d.client.DeleteVersion(ctx, &pb.DeleteVersionRequest{
		ProjectId: projectID,
		Version:   version,
	})
	if err != nil {
		return fmt.Errorf("agent_deployer: delete version: %w", err)
	}
	return nil
}

// DeleteProject удаляет все файлы проекта на агенте.
func (d *AgentDeployer) DeleteProject(ctx context.Context, projectID int64) error {
	ctx = d.withAuth(ctx)

	_, err := d.client.DeleteProject(ctx, &pb.DeleteProjectRequest{
		ProjectId: projectID,
	})
	if err != nil {
		return fmt.Errorf("agent_deployer: delete project: %w", err)
	}
	return nil
}

// UpdateCSP обновляет манифест csp.json на удаленном агенте развертывания.
func (d *AgentDeployer) UpdateCSP(ctx context.Context, projectID int64, env string, isOnline bool, allowedHosts []string) error {
	ctx = d.withAuth(ctx)

	_, err := d.client.UpdateCSP(ctx, &pb.UpdateCSPRequest{
		ProjectId:    projectID,
		Env:          env,
		IsOnline:     isOnline,
		AllowedHosts: allowedHosts,
	})
	if err != nil {
		return fmt.Errorf("agent_deployer: update csp: %w", err)
	}
	return nil
}
