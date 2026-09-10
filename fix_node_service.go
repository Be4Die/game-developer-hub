package service

import (
	"context"
	"fmt"
	"io"
	"time"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
)

func (s *NodeService) CreateServiceBackup(ctx context.Context, callerID string, nodeID int64, serviceName string) (*domain.ServiceBackup, error) {
	node, err := s.GetNode(ctx, callerID, nodeID)
	if err != nil {
		return nil, err
	}
	return s.nodeClient.CreateServiceBackup(ctx, node.Address, node.APIToken, serviceName)
}

func (s *NodeService) ListServiceBackups(ctx context.Context, callerID string, nodeID int64, serviceName string) ([]*domain.ServiceBackup, error) {
	node, err := s.GetNode(ctx, callerID, nodeID)
	if err != nil {
		return nil, err
	}
	return s.nodeClient.ListServiceBackups(ctx, node.Address, node.APIToken, serviceName)
}

func (s *NodeService) RestoreServiceBackup(ctx context.Context, callerID string, nodeID int64, serviceName string, backupID string) (bool, string, error) {
	node, err := s.GetNode(ctx, callerID, nodeID)
	if err != nil {
		return false, "", err
	}
	err = s.nodeClient.RestoreServiceBackup(ctx, node.Address, node.APIToken, serviceName, backupID)
	if err != nil {
		return false, err.Error(), err
	}
	return true, "Success", nil
}

func (s *NodeService) DeleteServiceBackup(ctx context.Context, callerID string, nodeID int64, serviceName string, backupID string) error {
	node, err := s.GetNode(ctx, callerID, nodeID)
	if err != nil {
		return err
	}
	return s.nodeClient.DeleteServiceBackup(ctx, node.Address, node.APIToken, serviceName, backupID)
}

func (s *NodeService) GetBackupTicket(ctx context.Context, callerID string, nodeID int64, serviceName string, backupID string) (string, int64, error) {
	return "ticket", 3600, nil // dummy for now
}

func (s *NodeService) DownloadServiceBackup(ctx context.Context, nodeID int64, serviceName string, backupID string) (io.ReadCloser, error) {
	node, err := s.nodeRepo.GetByID(ctx, nodeID)
	if err != nil {
		return nil, err
	}
	return s.nodeClient.DownloadServiceBackup(ctx, node.Address, node.APIToken, serviceName, backupID)
}

func (s *NodeService) UploadServiceBackup(ctx context.Context, nodeID int64, serviceName string, fileName string, restoreImmediately bool, r io.Reader) (*domain.ServiceBackup, error) {
	node, err := s.nodeRepo.GetByID(ctx, nodeID)
	if err != nil {
		return nil, err
	}
	return s.nodeClient.UploadServiceBackup(ctx, node.Address, node.APIToken, serviceName, fileName, restoreImmediately, r)
}
