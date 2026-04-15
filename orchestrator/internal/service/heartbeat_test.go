package service

import (
	"context"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/config"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
)

// ─── Mocks for HeartbeatService ─────────────────────────────────────────────

type hbMockNodeClient struct {
	heartbeatFn        func(ctx context.Context, address string) (*domain.ResourceUsage, error)
	getNodeInfoFn      func(ctx context.Context, address string) (*domain.NodeInfo, error)
	startInstanceFn    func(ctx context.Context, address string, req domain.StartInstanceRequest) (*domain.StartInstanceResult, error)
	stopInstanceFn     func(ctx context.Context, address string, instanceID int64, timeoutSec uint32) error
	streamLogsFn       func(ctx context.Context, address string, req domain.StreamLogsRequest) (domain.LogStream, error)
	loadImageFn        func(ctx context.Context, address string, meta domain.ImageMetadata, reader io.Reader) (*domain.ImageLoadResult, error)
	listInstancesFn    func(ctx context.Context, address string) ([]*domain.Instance, error)
	getInstanceFn      func(ctx context.Context, address string, instanceID int64) (*domain.Instance, error)
	getInstanceUsageFn func(ctx context.Context, address string, instanceID int64) (*domain.ResourceUsage, error)
}

func (m *hbMockNodeClient) Heartbeat(ctx context.Context, address string) (*domain.ResourceUsage, error) {
	return m.heartbeatFn(ctx, address)
}
func (m *hbMockNodeClient) GetNodeInfo(ctx context.Context, address string) (*domain.NodeInfo, error) {
	return m.getNodeInfoFn(ctx, address)
}
func (m *hbMockNodeClient) StartInstance(ctx context.Context, address string, req domain.StartInstanceRequest) (*domain.StartInstanceResult, error) {
	return m.startInstanceFn(ctx, address, req)
}
func (m *hbMockNodeClient) StopInstance(ctx context.Context, address string, instanceID int64, timeoutSec uint32) error {
	return m.stopInstanceFn(ctx, address, instanceID, timeoutSec)
}
func (m *hbMockNodeClient) StreamLogs(ctx context.Context, address string, req domain.StreamLogsRequest) (domain.LogStream, error) {
	return m.streamLogsFn(ctx, address, req)
}
func (m *hbMockNodeClient) LoadImage(ctx context.Context, address string, meta domain.ImageMetadata, reader io.Reader) (*domain.ImageLoadResult, error) {
	return m.loadImageFn(ctx, address, meta, reader)
}
func (m *hbMockNodeClient) ListInstances(ctx context.Context, address string) ([]*domain.Instance, error) {
	return m.listInstancesFn(ctx, address)
}
func (m *hbMockNodeClient) GetInstance(ctx context.Context, address string, instanceID int64) (*domain.Instance, error) {
	return m.getInstanceFn(ctx, address, instanceID)
}
func (m *hbMockNodeClient) GetInstanceUsage(ctx context.Context, address string, instanceID int64) (*domain.ResourceUsage, error) {
	return m.getInstanceUsageFn(ctx, address, instanceID)
}

type hbMockNodeRepo struct {
	createFn         func(ctx context.Context, node *domain.Node) error
	updateFn         func(ctx context.Context, node *domain.Node) error
	getByIDFn        func(ctx context.Context, id int64) (*domain.Node, error)
	getByAddressFn   func(ctx context.Context, addr string) (*domain.Node, error)
	listFn           func(ctx context.Context, status *domain.NodeStatus) ([]*domain.Node, error)
	deleteFn         func(ctx context.Context, id int64) error
	updateLastPingFn func(ctx context.Context, id int64) error
}

func (m *hbMockNodeRepo) Create(ctx context.Context, node *domain.Node) error {
	if m.createFn != nil {
		return m.createFn(ctx, node)
	}
	return nil
}
func (m *hbMockNodeRepo) Update(ctx context.Context, node *domain.Node) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, node)
	}
	return nil
}
func (m *hbMockNodeRepo) GetByID(ctx context.Context, id int64) (*domain.Node, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}
func (m *hbMockNodeRepo) GetByAddress(ctx context.Context, addr string) (*domain.Node, error) {
	if m.getByAddressFn != nil {
		return m.getByAddressFn(ctx, addr)
	}
	return nil, nil
}
func (m *hbMockNodeRepo) List(ctx context.Context, status *domain.NodeStatus) ([]*domain.Node, error) {
	if m.listFn != nil {
		return m.listFn(ctx, status)
	}
	return nil, nil
}
func (m *hbMockNodeRepo) Delete(ctx context.Context, id int64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}
func (m *hbMockNodeRepo) UpdateLastPing(ctx context.Context, id int64) error {
	if m.updateLastPingFn != nil {
		return m.updateLastPingFn(ctx, id)
	}
	return nil
}

type hbMockNodeStateStore struct {
	updateHeartbeatFn        func(ctx context.Context, nodeID int64, usage *domain.ResourceUsage) error
	getUsageFn               func(ctx context.Context, nodeID int64) (*domain.ResourceUsage, error)
	getActiveInstanceCountFn func(ctx context.Context, nodeID int64) (uint32, error)
	setActiveInstanceCountFn func(ctx context.Context, nodeID int64, count uint32) error
	deleteFn                 func(ctx context.Context, nodeID int64) error
}

func (m *hbMockNodeStateStore) UpdateHeartbeat(ctx context.Context, nodeID int64, usage *domain.ResourceUsage) error {
	if m.updateHeartbeatFn != nil {
		return m.updateHeartbeatFn(ctx, nodeID, usage)
	}
	return nil
}
func (m *hbMockNodeStateStore) GetUsage(ctx context.Context, nodeID int64) (*domain.ResourceUsage, error) {
	if m.getUsageFn != nil {
		return m.getUsageFn(ctx, nodeID)
	}
	return nil, nil
}
func (m *hbMockNodeStateStore) GetActiveInstanceCount(ctx context.Context, nodeID int64) (uint32, error) {
	if m.getActiveInstanceCountFn != nil {
		return m.getActiveInstanceCountFn(ctx, nodeID)
	}
	return 0, nil
}
func (m *hbMockNodeStateStore) SetActiveInstanceCount(ctx context.Context, nodeID int64, count uint32) error {
	if m.setActiveInstanceCountFn != nil {
		return m.setActiveInstanceCountFn(ctx, nodeID, count)
	}
	return nil
}
func (m *hbMockNodeStateStore) Delete(ctx context.Context, nodeID int64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, nodeID)
	}
	return nil
}

type hbMockInstanceRepo struct {
	createFn                    func(ctx context.Context, instance *domain.Instance) error
	getByIDFn                   func(ctx context.Context, id int64) (*domain.Instance, error)
	listByGameFn                func(ctx context.Context, gameID int64, status *domain.InstanceStatus) ([]*domain.Instance, error)
	listByNodeFn                func(ctx context.Context, nodeID int64) ([]*domain.Instance, error)
	updateFn                    func(ctx context.Context, instance *domain.Instance) error
	deleteFn                    func(ctx context.Context, id int64) error
	countByGameFn               func(ctx context.Context, gameID int64) (int, error)
	countActiveInstancesByBuild func(ctx context.Context, buildID int64) (int, error)
}

func (m *hbMockInstanceRepo) Create(ctx context.Context, instance *domain.Instance) error {
	if m.createFn != nil {
		return m.createFn(ctx, instance)
	}
	return nil
}
func (m *hbMockInstanceRepo) GetByID(ctx context.Context, id int64) (*domain.Instance, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}
func (m *hbMockInstanceRepo) ListByGame(ctx context.Context, gameID int64, status *domain.InstanceStatus) ([]*domain.Instance, error) {
	if m.listByGameFn != nil {
		return m.listByGameFn(ctx, gameID, status)
	}
	return nil, nil
}
func (m *hbMockInstanceRepo) ListByNode(ctx context.Context, nodeID int64) ([]*domain.Instance, error) {
	if m.listByNodeFn != nil {
		return m.listByNodeFn(ctx, nodeID)
	}
	return nil, nil
}
func (m *hbMockInstanceRepo) Update(ctx context.Context, instance *domain.Instance) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, instance)
	}
	return nil
}
func (m *hbMockInstanceRepo) Delete(ctx context.Context, id int64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}
func (m *hbMockInstanceRepo) CountByGame(ctx context.Context, gameID int64) (int, error) {
	if m.countByGameFn != nil {
		return m.countByGameFn(ctx, gameID)
	}
	return 0, nil
}
func (m *hbMockInstanceRepo) CountActiveInstancesByBuild(ctx context.Context, buildID int64) (int, error) {
	if m.countActiveInstancesByBuild != nil {
		return m.countActiveInstancesByBuild(ctx, buildID)
	}
	return 0, nil
}

type hbMockInstanceState struct {
	setStatusFn      func(ctx context.Context, instanceID int64, status domain.InstanceStatus) error
	getStatusFn      func(ctx context.Context, instanceID int64) (domain.InstanceStatus, error)
	setPlayerCountFn func(ctx context.Context, instanceID int64, count uint32) error
	getPlayerCountFn func(ctx context.Context, instanceID int64) (uint32, error)
	setUsageFn       func(ctx context.Context, instanceID int64, usage *domain.ResourceUsage) error
	getUsageFn       func(ctx context.Context, instanceID int64) (*domain.ResourceUsage, error)
	deleteFn         func(ctx context.Context, instanceID int64) error
}

func (m *hbMockInstanceState) SetStatus(ctx context.Context, instanceID int64, status domain.InstanceStatus) error {
	if m.setStatusFn != nil {
		return m.setStatusFn(ctx, instanceID, status)
	}
	return nil
}
func (m *hbMockInstanceState) GetStatus(ctx context.Context, instanceID int64) (domain.InstanceStatus, error) {
	if m.getStatusFn != nil {
		return m.getStatusFn(ctx, instanceID)
	}
	return 0, nil
}
func (m *hbMockInstanceState) SetPlayerCount(ctx context.Context, instanceID int64, count uint32) error {
	if m.setPlayerCountFn != nil {
		return m.setPlayerCountFn(ctx, instanceID, count)
	}
	return nil
}
func (m *hbMockInstanceState) GetPlayerCount(ctx context.Context, instanceID int64) (uint32, error) {
	if m.getPlayerCountFn != nil {
		return m.getPlayerCountFn(ctx, instanceID)
	}
	return 0, nil
}
func (m *hbMockInstanceState) SetUsage(ctx context.Context, instanceID int64, usage *domain.ResourceUsage) error {
	if m.setUsageFn != nil {
		return m.setUsageFn(ctx, instanceID, usage)
	}
	return nil
}
func (m *hbMockInstanceState) GetUsage(ctx context.Context, instanceID int64) (*domain.ResourceUsage, error) {
	if m.getUsageFn != nil {
		return m.getUsageFn(ctx, instanceID)
	}
	return nil, nil
}
func (m *hbMockInstanceState) Delete(ctx context.Context, instanceID int64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, instanceID)
	}
	return nil
}

// ─── Tests ───────────────────────────────────────────────────────────────────

func TestHeartbeatService_CheckNode_Success(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	now := time.Now().Add(-10 * time.Second)
	node := &domain.Node{
		ID:         1,
		Address:    "node1:44044",
		Status:     domain.NodeStatusOnline,
		LastPingAt: now,
	}

	usage := &domain.ResourceUsage{CPUUsagePercent: 45.0}
	updateLastPingCalled := false

	nodeRepo := &hbMockNodeRepo{
		listFn: func(ctx context.Context, status *domain.NodeStatus) ([]*domain.Node, error) {
			return []*domain.Node{node}, nil
		},
		updateLastPingFn: func(ctx context.Context, id int64) error {
			updateLastPingCalled = true
			return nil
		},
	}

	nodeState := &hbMockNodeStateStore{
		updateHeartbeatFn: func(ctx context.Context, nodeID int64, u *domain.ResourceUsage) error { return nil },
	}

	nodeClient := &hbMockNodeClient{
		heartbeatFn: func(ctx context.Context, address string) (*domain.ResourceUsage, error) {
			return usage, nil
		},
	}

	svc := NewHeartbeatService(nodeRepo, nodeState, &hbMockInstanceRepo{}, &hbMockInstanceState{}, nodeClient, config.NodeHeartbeatCfg{
		CheckInterval:     15 * time.Second,
		InactivityTimeout: 60 * time.Second,
	}, log)

	ctx := context.Background()
	err := svc.checkNode(ctx, node)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !updateLastPingCalled {
		t.Error("expected UpdateLastPing to be called")
	}
}

func TestHeartbeatService_CheckNode_HeartbeatError_WithinTimeout(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	node := &domain.Node{
		ID:         1,
		Address:    "node1:44044",
		Status:     domain.NodeStatusOnline,
		LastPingAt: time.Now(), // just pinged
	}

	nodeRepo := &hbMockNodeRepo{}
	nodeState := &hbMockNodeStateStore{}
	nodeClient := &hbMockNodeClient{
		heartbeatFn: func(ctx context.Context, address string) (*domain.ResourceUsage, error) {
			return nil, context.DeadlineExceeded
		},
	}

	svc := NewHeartbeatService(nodeRepo, nodeState, &hbMockInstanceRepo{}, &hbMockInstanceState{}, nodeClient, config.NodeHeartbeatCfg{
		CheckInterval:     15 * time.Second,
		InactivityTimeout: 60 * time.Second,
	}, log)

	err := svc.checkNode(context.Background(), node)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if node.Status == domain.NodeStatusOffline {
		t.Fatal("node should not be marked offline yet")
	}
}

func TestHeartbeatService_CheckNode_HeartbeatError_TimeoutExceeded(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	node := &domain.Node{
		ID:         1,
		Address:    "node1:44044",
		Status:     domain.NodeStatusOnline,
		LastPingAt: time.Now().Add(-120 * time.Second), // 2 min ago, timeout = 60s
	}

	updateCalled := false
	instancesCrashed := false

	nodeRepo := &hbMockNodeRepo{
		updateFn: func(ctx context.Context, n *domain.Node) error {
			updateCalled = true
			if n.Status != domain.NodeStatusOffline {
				t.Error("expected node status to be set to Offline")
			}
			return nil
		},
	}

	instanceRepo := &hbMockInstanceRepo{
		listByNodeFn: func(ctx context.Context, nodeID int64) ([]*domain.Instance, error) {
			return []*domain.Instance{{ID: 10, Status: domain.InstanceStatusRunning}}, nil
		},
		updateFn: func(ctx context.Context, inst *domain.Instance) error {
			if inst.Status == domain.InstanceStatusCrashed {
				instancesCrashed = true
			}
			return nil
		},
	}

	instanceState := &hbMockInstanceState{
		setStatusFn: func(ctx context.Context, instanceID int64, status domain.InstanceStatus) error {
			return nil
		},
	}

	nodeState := &hbMockNodeStateStore{}
	nodeClient := &hbMockNodeClient{
		heartbeatFn: func(ctx context.Context, address string) (*domain.ResourceUsage, error) {
			return nil, context.DeadlineExceeded
		},
	}

	svc := NewHeartbeatService(nodeRepo, nodeState, instanceRepo, instanceState, nodeClient, config.NodeHeartbeatCfg{
		CheckInterval:     15 * time.Second,
		InactivityTimeout: 60 * time.Second,
	}, log)

	err := svc.checkNode(context.Background(), node)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !updateCalled {
		t.Error("expected node update to be called")
	}
	if !instancesCrashed {
		t.Error("expected instances to be marked as crashed")
	}
}

func TestHeartbeatService_CheckNode_RecoverFromOffline(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	node := &domain.Node{
		ID:      1,
		Address: "node1:44044",
		Status:  domain.NodeStatusOffline,
	}

	updateCalled := false
	nodeRepo := &hbMockNodeRepo{
		updateFn: func(ctx context.Context, n *domain.Node) error {
			updateCalled = true
			if n.Status != domain.NodeStatusOnline {
				t.Error("expected node status to be restored to Online")
			}
			return nil
		},
	}

	nodeState := &hbMockNodeStateStore{
		updateHeartbeatFn: func(ctx context.Context, nodeID int64, u *domain.ResourceUsage) error { return nil },
	}

	nodeClient := &hbMockNodeClient{
		heartbeatFn: func(ctx context.Context, address string) (*domain.ResourceUsage, error) {
			return &domain.ResourceUsage{CPUUsagePercent: 10.0}, nil
		},
	}

	svc := NewHeartbeatService(nodeRepo, nodeState, &hbMockInstanceRepo{}, &hbMockInstanceState{}, nodeClient, config.NodeHeartbeatCfg{
		CheckInterval:     15 * time.Second,
		InactivityTimeout: 60 * time.Second,
	}, log)

	err := svc.checkNode(context.Background(), node)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !updateCalled {
		t.Error("expected node update to restore status to Online")
	}
}

func TestHeartbeatService_CheckAllNodes_SkipMaintenance(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	node := &domain.Node{
		ID:     1,
		Status: domain.NodeStatusMaintenance,
	}

	nodeRepo := &hbMockNodeRepo{
		listFn: func(ctx context.Context, status *domain.NodeStatus) ([]*domain.Node, error) {
			return []*domain.Node{node}, nil
		},
	}

	heartbeatCalled := false
	nodeClient := &hbMockNodeClient{
		heartbeatFn: func(ctx context.Context, address string) (*domain.ResourceUsage, error) {
			heartbeatCalled = true
			return nil, nil
		},
	}

	svc := NewHeartbeatService(nodeRepo, &hbMockNodeStateStore{}, &hbMockInstanceRepo{}, &hbMockInstanceState{}, nodeClient, config.NodeHeartbeatCfg{
		CheckInterval:     15 * time.Second,
		InactivityTimeout: 60 * time.Second,
	}, log)

	ctx := context.Background()
	svc.checkAllNodes(ctx)

	if heartbeatCalled {
		t.Error("heartbeat should not be called for maintenance nodes")
	}
}

func TestHeartbeatService_CheckAllNodes_ListError(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	nodeRepo := &hbMockNodeRepo{
		listFn: func(ctx context.Context, status *domain.NodeStatus) ([]*domain.Node, error) {
			return nil, context.Canceled
		},
	}

	svc := NewHeartbeatService(nodeRepo, &hbMockNodeStateStore{}, &hbMockInstanceRepo{}, &hbMockInstanceState{}, &hbMockNodeClient{}, config.NodeHeartbeatCfg{
		CheckInterval:     15 * time.Second,
		InactivityTimeout: 60 * time.Second,
	}, log)

	ctx := context.Background()
	// Should not panic.
	svc.checkAllNodes(ctx)
}
