package service

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/infrastructure/config"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
)

// ─── Mocks for HeartbeatService ─────────────────────────────────────────────

type hbMockNodeClient struct {
	heartbeatFn        func(ctx context.Context, address, apiKey string) (*domain.HeartbeatResult, error)
	getNodeInfoFn      func(ctx context.Context, address, apiKey string) (*domain.NodeInfo, error)
	startInstanceFn    func(ctx context.Context, address, apiKey string, req domain.StartInstanceRequest) (*domain.StartInstanceResult, error)
	stopInstanceFn     func(ctx context.Context, address, apiKey string, instanceID int64, timeoutSec uint32) error
	streamLogsFn       func(ctx context.Context, address, apiKey string, req domain.StreamLogsRequest) (domain.LogStream, error)
	buildImageFn       func(ctx context.Context, address, apiKey string, metadata domain.BuildImageMetadata, archive io.Reader) error
	loadImageFn        func(ctx context.Context, address, apiKey string, meta domain.ImageMetadata, reader io.Reader) (*domain.ImageLoadResult, error)
	listInstancesFn    func(ctx context.Context, address, apiKey string) ([]*domain.Instance, error)
	getInstanceFn      func(ctx context.Context, address, apiKey string, instanceID int64) (*domain.Instance, error)
	getInstanceUsageFn func(ctx context.Context, address, apiKey string, instanceID int64) (*domain.ResourceUsage, error)
	deleteInstanceFn   func(ctx context.Context, address, apiKey string, instanceID int64) error
}

func (m *hbMockNodeClient) Heartbeat(ctx context.Context, address, apiKey string) (*domain.HeartbeatResult, error) {
	return m.heartbeatFn(ctx, address, apiKey)
}
func (m *hbMockNodeClient) GetNodeInfo(ctx context.Context, address, apiKey string) (*domain.NodeInfo, error) {
	return m.getNodeInfoFn(ctx, address, apiKey)
}
func (m *hbMockNodeClient) StartInstance(ctx context.Context, address, apiKey string, req domain.StartInstanceRequest) (*domain.StartInstanceResult, error) {
	return m.startInstanceFn(ctx, address, apiKey, req)
}
func (m *hbMockNodeClient) StopInstance(ctx context.Context, address, apiKey string, instanceID int64, timeoutSec uint32) error {
	return m.stopInstanceFn(ctx, address, apiKey, instanceID, timeoutSec)
}
func (m *hbMockNodeClient) RestartInstance(ctx context.Context, address, apiKey string, instanceID int64, timeoutSec uint32) error {
	return nil
}
func (m *hbMockNodeClient) StartStoppedInstance(ctx context.Context, address, apiKey string, instanceID int64) error {
	return nil
}
func (m *hbMockNodeClient) StreamLogs(ctx context.Context, address, apiKey string, req domain.StreamLogsRequest) (domain.LogStream, error) {
	return m.streamLogsFn(ctx, address, apiKey, req)
}
func (m *hbMockNodeClient) BuildImage(ctx context.Context, address, apiKey string, metadata domain.BuildImageMetadata, archive io.Reader) error {
	if m.buildImageFn != nil {
		return m.buildImageFn(ctx, address, apiKey, metadata, archive)
	}
	return nil
}
func (m *hbMockNodeClient) LoadImage(ctx context.Context, address, apiKey string, meta domain.ImageMetadata, reader io.Reader) (*domain.ImageLoadResult, error) {
	return m.loadImageFn(ctx, address, apiKey, meta, reader)
}
func (m *hbMockNodeClient) ListInstances(ctx context.Context, address, apiKey string) ([]*domain.Instance, error) {
	if m.listInstancesFn != nil {
		return m.listInstancesFn(ctx, address, apiKey)
	}
	return []*domain.Instance{}, nil
}
func (m *hbMockNodeClient) GetInstance(ctx context.Context, address, apiKey string, instanceID int64) (*domain.Instance, error) {
	return m.getInstanceFn(ctx, address, apiKey, instanceID)
}
func (m *hbMockNodeClient) GetInstanceUsage(ctx context.Context, address, apiKey string, instanceID int64) (*domain.ResourceUsage, error) {
	return m.getInstanceUsageFn(ctx, address, apiKey, instanceID)
}
func (m *hbMockNodeClient) DeleteInstance(ctx context.Context, address, apiKey string, instanceID int64) error {
	if m.deleteInstanceFn != nil {
		return m.deleteInstanceFn(ctx, address, apiKey, instanceID)
	}
	return nil
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
func (m *hbMockInstanceRepo) List(ctx context.Context) ([]*domain.Instance, error) {
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
func (m *hbMockInstanceRepo) GetNextID(ctx context.Context) (int64, error) {
	return 1, nil
}

type hbMockBuildStorage struct {
	listByGameFn func(ctx context.Context, gameID int64, limit int) ([]*domain.ServerBuild, error)
}

func (m *hbMockBuildStorage) Create(ctx context.Context, build *domain.ServerBuild) error { return nil }
func (m *hbMockBuildStorage) GetByID(ctx context.Context, id int64) (*domain.ServerBuild, error) {
	return nil, nil
}
func (m *hbMockBuildStorage) GetByVersion(ctx context.Context, gameID int64, version string) (*domain.ServerBuild, error) {
	return nil, nil
}
func (m *hbMockBuildStorage) ListByGame(ctx context.Context, gameID int64, limit int) ([]*domain.ServerBuild, error) {
	if m.listByGameFn != nil {
		return m.listByGameFn(ctx, gameID, limit)
	}
	return nil, nil
}
func (m *hbMockBuildStorage) CountByGame(ctx context.Context, gameID int64) (int, error) { return 0, nil }
func (m *hbMockBuildStorage) Delete(ctx context.Context, id int64) error { return nil }
func (m *hbMockBuildStorage) CountActiveInstancesByBuild(ctx context.Context, buildID int64) (int, error) {
	return 0, nil
}

type hbMockGamePolicyRepo struct {
	getFn     func(ctx context.Context, gameID int64) (*domain.GamePolicy, error)
	setFn     func(ctx context.Context, policy *domain.GamePolicy) error
	deleteFn  func(ctx context.Context, gameID int64) error
	listAllFn func(ctx context.Context) ([]*domain.GamePolicy, error)
}

func (m *hbMockGamePolicyRepo) Get(ctx context.Context, gameID int64) (*domain.GamePolicy, error) {
	if m.getFn != nil {
		return m.getFn(ctx, gameID)
	}
	return nil, domain.ErrNotFound
}
func (m *hbMockGamePolicyRepo) Set(ctx context.Context, policy *domain.GamePolicy) error {
	if m.setFn != nil {
		return m.setFn(ctx, policy)
	}
	return nil
}
func (m *hbMockGamePolicyRepo) Delete(ctx context.Context, gameID int64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, gameID)
	}
	return nil
}
func (m *hbMockGamePolicyRepo) ListAll(ctx context.Context) ([]*domain.GamePolicy, error) {
	if m.listAllFn != nil {
		return m.listAllFn(ctx)
	}
	return nil, nil
}

type hbMockInstanceState struct {
	setStatusFn          func(ctx context.Context, instanceID int64, status domain.InstanceStatus) error
	getStatusFn          func(ctx context.Context, instanceID int64) (domain.InstanceStatus, error)
	setPlayerCountFn     func(ctx context.Context, instanceID int64, count uint32) error
	getPlayerCountFn     func(ctx context.Context, instanceID int64) (uint32, error)
	setUsageFn           func(ctx context.Context, instanceID int64, usage *domain.ResourceUsage) error
	getUsageFn           func(ctx context.Context, instanceID int64) (*domain.ResourceUsage, error)
	deleteFn             func(ctx context.Context, instanceID int64) error
	setZeroSinceFn       func(ctx context.Context, instanceID int64, t time.Time) error
	getZeroSinceFn       func(ctx context.Context, instanceID int64) (time.Time, error)
	deleteZeroSinceFn    func(ctx context.Context, instanceID int64) error
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
func (m *hbMockInstanceState) SetZeroPlayersSince(ctx context.Context, instanceID int64, t time.Time) error {
	if m.setZeroSinceFn != nil {
		return m.setZeroSinceFn(ctx, instanceID, t)
	}
	return nil
}
func (m *hbMockInstanceState) GetZeroPlayersSince(ctx context.Context, instanceID int64) (time.Time, error) {
	if m.getZeroSinceFn != nil {
		return m.getZeroSinceFn(ctx, instanceID)
	}
	return time.Time{}, domain.ErrNotFound
}
func (m *hbMockInstanceState) DeleteZeroPlayersSince(ctx context.Context, instanceID int64) error {
	if m.deleteZeroSinceFn != nil {
		return m.deleteZeroSinceFn(ctx, instanceID)
	}
	return nil
}

// hbMockInstanceOrchestrator мок для instanceOrchestrator.
type hbMockInstanceOrchestrator struct {
	startInstanceFn  func(ctx context.Context, params StartInstanceParams) (*domain.Instance, error)
	restartInstanceFn func(ctx context.Context, ownerID string, gameID, instanceID int64) (*domain.Instance, error)
	stopInstanceFn   func(ctx context.Context, ownerID string, gameID, instanceID int64, timeoutSec uint32) (*domain.Instance, error)
}

func (m *hbMockInstanceOrchestrator) StartInstance(ctx context.Context, params StartInstanceParams) (*domain.Instance, error) {
	if m.startInstanceFn != nil {
		return m.startInstanceFn(ctx, params)
	}
	return nil, nil
}
func (m *hbMockInstanceOrchestrator) RestartInstance(ctx context.Context, ownerID string, gameID, instanceID int64) (*domain.Instance, error) {
	if m.restartInstanceFn != nil {
		return m.restartInstanceFn(ctx, ownerID, gameID, instanceID)
	}
	return nil, nil
}
func (m *hbMockInstanceOrchestrator) StopInstance(ctx context.Context, ownerID string, gameID, instanceID int64, timeoutSec uint32) (*domain.Instance, error) {
	if m.stopInstanceFn != nil {
		return m.stopInstanceFn(ctx, ownerID, gameID, instanceID, timeoutSec)
	}
	return nil, nil
}

// newTestHeartbeatService создаёт HeartbeatService с дефолтными моками.
func newTestHeartbeatService(
	nodeRepo domain.NodeRepo,
	nodeState domain.NodeStateStore,
	instanceRepo domain.InstanceRepo,
	instanceState domain.InstanceStateStore,
	nodeClient domain.NodeClient,
	buildRepo domain.BuildStorage,
	policyRepo domain.GamePolicyRepo,
	instanceSvc instanceOrchestrator,
	hbCfg config.NodeHeartbeatCfg,
	log *slog.Logger,
) *HeartbeatService {
	if buildRepo == nil {
		buildRepo = &hbMockBuildStorage{}
	}
	if policyRepo == nil {
		policyRepo = &hbMockGamePolicyRepo{}
	}
	policyService := NewGamePolicyService(policyRepo)
	if instanceSvc == nil {
		instanceSvc = &hbMockInstanceOrchestrator{}
	}
	return NewHeartbeatService(
		nodeRepo,
		nodeState,
		instanceRepo,
		instanceState,
		nodeClient,
		buildRepo,
		policyService,
		instanceSvc,
		nil, // queueSvc
		hbCfg,
		log,
	)
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
		heartbeatFn: func(ctx context.Context, address, apiKey string) (*domain.HeartbeatResult, error) {
			return &domain.HeartbeatResult{Usage: usage, ActiveInstanceCount: 0}, nil
		},
	}

	svc := newTestHeartbeatService(nodeRepo, nodeState, &hbMockInstanceRepo{}, &hbMockInstanceState{}, nodeClient, nil, &hbMockGamePolicyRepo{}, nil, config.NodeHeartbeatCfg{
		CheckInterval:     15 * time.Second,
		InactivityTimeout: 60 * time.Second,
	}, log)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
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
		heartbeatFn: func(ctx context.Context, address, apiKey string) (*domain.HeartbeatResult, error) {
			return nil, context.DeadlineExceeded
		},
	}

	svc := newTestHeartbeatService(nodeRepo, nodeState, &hbMockInstanceRepo{}, &hbMockInstanceState{}, nodeClient, nil, &hbMockGamePolicyRepo{}, nil, config.NodeHeartbeatCfg{
		CheckInterval:     15 * time.Second,
		InactivityTimeout: 60 * time.Second,
	}, log)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	err := svc.checkNode(ctx, node)
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
		heartbeatFn: func(ctx context.Context, address, apiKey string) (*domain.HeartbeatResult, error) {
			return nil, context.DeadlineExceeded
		},
	}

	svc := newTestHeartbeatService(nodeRepo, nodeState, instanceRepo, instanceState, nodeClient, nil, &hbMockGamePolicyRepo{}, nil, config.NodeHeartbeatCfg{
		CheckInterval:     15 * time.Second,
		InactivityTimeout: 60 * time.Second,
	}, log)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	err := svc.checkNode(ctx, node)
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
		heartbeatFn: func(ctx context.Context, address, apiKey string) (*domain.HeartbeatResult, error) {
			return &domain.HeartbeatResult{Usage: &domain.ResourceUsage{CPUUsagePercent: 10.0}}, nil
		},
	}

	svc := newTestHeartbeatService(nodeRepo, nodeState, &hbMockInstanceRepo{}, &hbMockInstanceState{}, nodeClient, nil, &hbMockGamePolicyRepo{}, nil, config.NodeHeartbeatCfg{
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
		heartbeatFn: func(ctx context.Context, address, apiKey string) (*domain.HeartbeatResult, error) {
			heartbeatCalled = true
			return nil, nil
		},
	}

	svc := newTestHeartbeatService(nodeRepo, &hbMockNodeStateStore{}, &hbMockInstanceRepo{}, &hbMockInstanceState{}, nodeClient, nil, &hbMockGamePolicyRepo{}, nil, config.NodeHeartbeatCfg{
		CheckInterval:     15 * time.Second,
		InactivityTimeout: 60 * time.Second,
	}, log)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
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

	svc := newTestHeartbeatService(nodeRepo, &hbMockNodeStateStore{}, &hbMockInstanceRepo{}, &hbMockInstanceState{}, &hbMockNodeClient{}, nil, &hbMockGamePolicyRepo{}, nil, config.NodeHeartbeatCfg{
		CheckInterval:     15 * time.Second,
		InactivityTimeout: 60 * time.Second,
	}, log)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	// Should not panic.
	svc.checkAllNodes(ctx)
}

func TestHeartbeatService_ReconcileInstances(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	node := &domain.Node{
		ID:      1,
		Address: "node1:44044",
		Status:  domain.NodeStatusOnline,
	}

	// DB has 3 instances: 10 (Running), 11 (Starting), 12 (Stopped)
	// Node reports only instance 10.
	updatedInstances := make(map[int64]domain.InstanceStatus)
	instanceRepo := &hbMockInstanceRepo{
		listByNodeFn: func(ctx context.Context, nodeID int64) ([]*domain.Instance, error) {
			return []*domain.Instance{
				{ID: 10, NodeID: 1, Status: domain.InstanceStatusRunning},
				{ID: 11, NodeID: 1, Status: domain.InstanceStatusStarting},
				{ID: 12, NodeID: 1, Status: domain.InstanceStatusStopped},
			}, nil
		},
		updateFn: func(ctx context.Context, inst *domain.Instance) error {
			updatedInstances[inst.ID] = inst.Status
			return nil
		},
	}

	instanceState := &hbMockInstanceState{
		setStatusFn: func(ctx context.Context, instanceID int64, status domain.InstanceStatus) error {
			return nil
		},
	}

	nodeClient := &hbMockNodeClient{
		listInstancesFn: func(ctx context.Context, address, apiKey string) ([]*domain.Instance, error) {
			return []*domain.Instance{
				{ID: 10, Status: domain.InstanceStatusRunning},
			}, nil
		},
	}

	svc := newTestHeartbeatService(&hbMockNodeRepo{}, &hbMockNodeStateStore{}, instanceRepo, instanceState, nodeClient, nil, &hbMockGamePolicyRepo{}, nil, config.NodeHeartbeatCfg{
		CheckInterval:     15 * time.Second,
		InactivityTimeout: 60 * time.Second,
	}, log)

	ctx := context.Background()
	err := svc.reconcileInstances(ctx, node)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if updatedInstances[10] != 0 {
		t.Errorf("instance 10 should not be updated, got status %d", updatedInstances[10])
	}
	if updatedInstances[11] != domain.InstanceStatusCrashed {
		t.Errorf("instance 11 should be marked crashed, got status %d", updatedInstances[11])
	}
	if updatedInstances[12] != 0 {
		t.Errorf("instance 12 should not be updated (already stopped), got status %d", updatedInstances[12])
	}
}

func TestHeartbeatService_EnforcePolicies_KeepAlive_NoInstances(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	gameID := int64(42)
	startedCount := 0

	policyRepo := &hbMockGamePolicyRepo{
		listAllFn: func(ctx context.Context) ([]*domain.GamePolicy, error) {
			return []*domain.GamePolicy{
				{
					GameID:                gameID,
					OwnerID:               "test-user",
					Mode:                  domain.OrchestrationModeKeepAlive,
					TargetInstances:       3,
					AutoRestart:           true,
					DefaultBuildVersion:   "latest",
					MaxPlayersPerInstance: 100,
					MaxInstancesPerGame:   5,
				},
			}, nil
		},
	}

	instanceRepo := &hbMockInstanceRepo{
		listByGameFn: func(ctx context.Context, gameID int64, status *domain.InstanceStatus) ([]*domain.Instance, error) {
			return []*domain.Instance{}, nil // нет инстансов
		},
	}

	buildRepo := &hbMockBuildStorage{
		listByGameFn: func(ctx context.Context, gameID int64, limit int) ([]*domain.ServerBuild, error) {
			return []*domain.ServerBuild{{Version: "v1.0.0"}}, nil
		},
	}

	instanceSvc := &hbMockInstanceOrchestrator{
		startInstanceFn: func(ctx context.Context, params StartInstanceParams) (*domain.Instance, error) {
			startedCount++
			return nil, nil
		},
	}

	svc := newTestHeartbeatService(&hbMockNodeRepo{}, &hbMockNodeStateStore{}, instanceRepo, &hbMockInstanceState{}, &hbMockNodeClient{}, buildRepo, policyRepo, instanceSvc, config.NodeHeartbeatCfg{}, log)

	ctx := context.Background()
	svc.EnforcePolicies(ctx)

	// Даём горутинам время запуститься (в реальном коде они fire-and-forget).
	time.Sleep(100 * time.Millisecond)

	if startedCount != 3 {
		t.Errorf("expected 3 instances to be started, got %d", startedCount)
	}
}

func TestHeartbeatService_EnforcePolicies_KeepAlive_AfterManualStop(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	gameID := int64(42)
	startedCount := 0

	policyRepo := &hbMockGamePolicyRepo{
		listAllFn: func(ctx context.Context) ([]*domain.GamePolicy, error) {
			return []*domain.GamePolicy{
				{
					GameID:                gameID,
					OwnerID:               "test-user",
					Mode:                  domain.OrchestrationModeKeepAlive,
					TargetInstances:       1,
					AutoRestart:           true,
					DefaultBuildVersion:   "latest",
					MaxPlayersPerInstance: 100,
					MaxInstancesPerGame:   5,
				},
			}, nil
		},
	}

	// Один остановленный инстанс вручную — total=1 >= target=1, не поднимаем.
	instanceRepo := &hbMockInstanceRepo{
		listByGameFn: func(ctx context.Context, gameID int64, status *domain.InstanceStatus) ([]*domain.Instance, error) {
			return []*domain.Instance{
				{ID: 1, GameID: gameID, Status: domain.InstanceStatusStopped},
			}, nil
		},
	}

	instanceSvc := &hbMockInstanceOrchestrator{
		startInstanceFn: func(ctx context.Context, params StartInstanceParams) (*domain.Instance, error) {
			startedCount++
			return nil, nil
		},
	}

	svc := newTestHeartbeatService(&hbMockNodeRepo{}, &hbMockNodeStateStore{}, instanceRepo, &hbMockInstanceState{}, &hbMockNodeClient{}, nil, policyRepo, instanceSvc, config.NodeHeartbeatCfg{}, log)

	ctx := context.Background()
	svc.EnforcePolicies(ctx)

	time.Sleep(100 * time.Millisecond)

	if startedCount != 0 {
		t.Errorf("expected 0 starts after manual stop (total >= target), got %d", startedCount)
	}
}

func TestHeartbeatService_EnforcePolicies_Disabled_DoesNothing(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	startedCount := 0

	policyRepo := &hbMockGamePolicyRepo{
		listAllFn: func(ctx context.Context) ([]*domain.GamePolicy, error) {
			return []*domain.GamePolicy{
				{
					GameID:          1,
					OwnerID:         "test-user",
					Mode:            domain.OrchestrationModeDisabled,
					TargetInstances: 3,
				},
			}, nil
		},
	}

	instanceRepo := &hbMockInstanceRepo{
		listByGameFn: func(ctx context.Context, gameID int64, status *domain.InstanceStatus) ([]*domain.Instance, error) {
			return []*domain.Instance{}, nil
		},
	}

	instanceSvc := &hbMockInstanceOrchestrator{
		startInstanceFn: func(ctx context.Context, params StartInstanceParams) (*domain.Instance, error) {
			startedCount++
			return nil, nil
		},
	}

	svc := newTestHeartbeatService(&hbMockNodeRepo{}, &hbMockNodeStateStore{}, instanceRepo, &hbMockInstanceState{}, &hbMockNodeClient{}, nil, policyRepo, instanceSvc, config.NodeHeartbeatCfg{}, log)

	ctx := context.Background()
	svc.EnforcePolicies(ctx)

	time.Sleep(100 * time.Millisecond)

	if startedCount != 0 {
		t.Errorf("expected 0 starts for disabled policy, got %d", startedCount)
	}
}

func TestHeartbeatService_EnforcePolicies_MaxInstancesReached(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	gameID := int64(42)
	startedCount := 0

	policyRepo := &hbMockGamePolicyRepo{
		listAllFn: func(ctx context.Context) ([]*domain.GamePolicy, error) {
			return []*domain.GamePolicy{
				{
					GameID:              gameID,
					Mode:                domain.OrchestrationModeKeepAlive,
					TargetInstances:     5,
					AutoRestart:         true,
					DefaultBuildVersion: "latest",
					MaxInstancesPerGame: 2, // лимит ниже target
				},
			}, nil
		},
	}

	instanceRepo := &hbMockInstanceRepo{
		listByGameFn: func(ctx context.Context, gameID int64, status *domain.InstanceStatus) ([]*domain.Instance, error) {
			return []*domain.Instance{
				{ID: 1, GameID: gameID, Status: domain.InstanceStatusRunning},
				{ID: 2, GameID: gameID, Status: domain.InstanceStatusRunning},
			}, nil // уже на лимите
		},
	}

	instanceSvc := &hbMockInstanceOrchestrator{
		startInstanceFn: func(ctx context.Context, params StartInstanceParams) (*domain.Instance, error) {
			startedCount++
			return nil, nil
		},
	}

	svc := newTestHeartbeatService(&hbMockNodeRepo{}, &hbMockNodeStateStore{}, instanceRepo, &hbMockInstanceState{}, &hbMockNodeClient{}, nil, policyRepo, instanceSvc, config.NodeHeartbeatCfg{}, log)

	ctx := context.Background()
	svc.EnforcePolicies(ctx)

	time.Sleep(100 * time.Millisecond)

	if startedCount != 0 {
		t.Errorf("expected 0 starts when max_instances reached, got %d", startedCount)
	}
}

func TestHeartbeatService_MaybeAutoRestart_RestartFails_StartsNew(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	gameID := int64(42)
	instanceID := int64(99)
	startedNew := false

	policyRepo := &hbMockGamePolicyRepo{
		getFn: func(ctx context.Context, id int64) (*domain.GamePolicy, error) {
			return &domain.GamePolicy{
				GameID:              gameID,
				Mode:                domain.OrchestrationModeKeepAlive,
				AutoRestart:         true,
				DefaultBuildVersion: "latest",
				MaxInstancesPerGame: 5,
			}, nil
		},
	}

	buildRepo := &hbMockBuildStorage{
		listByGameFn: func(ctx context.Context, gameID int64, limit int) ([]*domain.ServerBuild, error) {
			return []*domain.ServerBuild{{Version: "v1.0.0"}}, nil
		},
	}

	instanceSvc := &hbMockInstanceOrchestrator{
		restartInstanceFn: func(ctx context.Context, ownerID string, gameID, instID int64) (*domain.Instance, error) {
			return nil, fmt.Errorf("container not found")
		},
		startInstanceFn: func(ctx context.Context, params StartInstanceParams) (*domain.Instance, error) {
			startedNew = true
			return nil, nil
		},
	}

	svc := newTestHeartbeatService(&hbMockNodeRepo{}, &hbMockNodeStateStore{}, &hbMockInstanceRepo{}, &hbMockInstanceState{}, &hbMockNodeClient{}, buildRepo, policyRepo, instanceSvc, config.NodeHeartbeatCfg{}, log)

	inst := &domain.Instance{ID: instanceID, GameID: gameID, Status: domain.InstanceStatusCrashed}
	svc.maybeAutoRestart(context.Background(), inst)

	time.Sleep(100 * time.Millisecond)

	if !startedNew {
		t.Error("expected new instance to be started when RestartInstance fails")
	}
}

func TestHeartbeatService_EnforceScaleToZero_StopsIdleInstance(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	gameID := int64(42)
	instanceID := int64(10)
	stopped := false

	policyRepo := &hbMockGamePolicyRepo{
		listAllFn: func(ctx context.Context) ([]*domain.GamePolicy, error) {
			return []*domain.GamePolicy{
				{
					GameID:              gameID,
					Mode:                domain.OrchestrationModeScaleToZero,
					TargetInstances:     0,
					ScaleToZeroTimeout:  1, // 1 минута
					DefaultBuildVersion: "latest",
					MaxInstancesPerGame: 5,
				},
			}, nil
		},
	}

	instanceRepo := &hbMockInstanceRepo{
		listByGameFn: func(ctx context.Context, gameID int64, status *domain.InstanceStatus) ([]*domain.Instance, error) {
			return []*domain.Instance{
				{ID: instanceID, GameID: gameID, Status: domain.InstanceStatusRunning},
			}, nil
		},
	}

	instanceState := &hbMockInstanceState{
		getPlayerCountFn: func(ctx context.Context, id int64) (uint32, error) {
			return 0, nil
		},
		getZeroSinceFn: func(ctx context.Context, id int64) (time.Time, error) {
			return time.Now().Add(-2 * time.Minute), nil // простаивает 2 минуты
		},
		deleteZeroSinceFn: func(ctx context.Context, id int64) error {
			return nil
		},
	}

	instanceSvc := &hbMockInstanceOrchestrator{
		stopInstanceFn: func(ctx context.Context, ownerID string, gameID, instID int64, timeoutSec uint32) (*domain.Instance, error) {
			stopped = true
			return nil, nil
		},
	}

	svc := newTestHeartbeatService(&hbMockNodeRepo{}, &hbMockNodeStateStore{}, instanceRepo, instanceState, &hbMockNodeClient{}, nil, policyRepo, instanceSvc, config.NodeHeartbeatCfg{}, log)

	ctx := context.Background()
	svc.enforceScaleToZero(ctx)

	if !stopped {
		t.Error("expected idle instance to be stopped")
	}
}

func TestHeartbeatService_EnforceScaleToZero_KeepsBusyInstance(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	gameID := int64(42)
	stopped := false

	policyRepo := &hbMockGamePolicyRepo{
		listAllFn: func(ctx context.Context) ([]*domain.GamePolicy, error) {
			return []*domain.GamePolicy{
				{
					GameID:              gameID,
					Mode:                domain.OrchestrationModeScaleToZero,
					ScaleToZeroTimeout:  1,
					MaxInstancesPerGame: 5,
				},
			}, nil
		},
	}

	instanceRepo := &hbMockInstanceRepo{
		listByGameFn: func(ctx context.Context, gameID int64, status *domain.InstanceStatus) ([]*domain.Instance, error) {
			return []*domain.Instance{
				{ID: 1, GameID: gameID, Status: domain.InstanceStatusRunning},
			}, nil
		},
	}

	instanceState := &hbMockInstanceState{
		getPlayerCountFn: func(ctx context.Context, id int64) (uint32, error) {
			return 5, nil // есть игроки
		},
		deleteZeroSinceFn: func(ctx context.Context, id int64) error {
			return nil
		},
	}

	instanceSvc := &hbMockInstanceOrchestrator{
		stopInstanceFn: func(ctx context.Context, ownerID string, gameID, instID int64, timeoutSec uint32) (*domain.Instance, error) {
			stopped = true
			return nil, nil
		},
	}

	svc := newTestHeartbeatService(&hbMockNodeRepo{}, &hbMockNodeStateStore{}, instanceRepo, instanceState, &hbMockNodeClient{}, nil, policyRepo, instanceSvc, config.NodeHeartbeatCfg{}, log)

	ctx := context.Background()
	svc.enforceScaleToZero(ctx)

	if stopped {
		t.Error("expected busy instance NOT to be stopped")
	}
}

func TestHeartbeatService_EnforceScaleUp_SpawnsWhenFull(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	gameID := int64(42)
	startedCount := 0

	policyRepo := &hbMockGamePolicyRepo{
		listAllFn: func(ctx context.Context) ([]*domain.GamePolicy, error) {
			return []*domain.GamePolicy{
				{
					GameID:                gameID,
					OwnerID:               "test-user",
					Mode:                  domain.OrchestrationModeKeepAlive,
					TargetInstances:       1,
					MaxPlayersPerInstance: 10,
					MaxInstancesPerGame:   5,
					ScaleBehavior:         domain.ScaleBehaviorSpawn,
					DefaultBuildVersion:   "latest",
				},
			}, nil
		},
	}

	instanceRepo := &hbMockInstanceRepo{
		listByGameFn: func(ctx context.Context, gameID int64, status *domain.InstanceStatus) ([]*domain.Instance, error) {
			if status != nil && *status == domain.InstanceStatusRunning {
				return []*domain.Instance{
					{ID: 1, GameID: gameID, Status: domain.InstanceStatusRunning},
				}, nil
			}
			return []*domain.Instance{{ID: 1, GameID: gameID}}, nil // total=1 < max=5
		},
	}

	instanceState := &hbMockInstanceState{
		getPlayerCountFn: func(ctx context.Context, id int64) (uint32, error) {
			return 10, nil // полностью заполнен
		},
	}

	buildRepo := &hbMockBuildStorage{
		listByGameFn: func(ctx context.Context, gameID int64, limit int) ([]*domain.ServerBuild, error) {
			return []*domain.ServerBuild{{Version: "v1.0.0"}}, nil
		},
	}

	instanceSvc := &hbMockInstanceOrchestrator{
		startInstanceFn: func(ctx context.Context, params StartInstanceParams) (*domain.Instance, error) {
			startedCount++
			return nil, nil
		},
	}

	svc := newTestHeartbeatService(&hbMockNodeRepo{}, &hbMockNodeStateStore{}, instanceRepo, instanceState, &hbMockNodeClient{}, buildRepo, policyRepo, instanceSvc, config.NodeHeartbeatCfg{}, log)

	ctx := context.Background()
	svc.enforceScaleUp(ctx)

	time.Sleep(100 * time.Millisecond)

	if startedCount != 1 {
		t.Errorf("expected 1 new instance when full, got %d", startedCount)
	}
}

func TestHeartbeatService_EnforceScaleUp_DoesNothingWhenQueue(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	gameID := int64(42)
	startedCount := 0

	policyRepo := &hbMockGamePolicyRepo{
		listAllFn: func(ctx context.Context) ([]*domain.GamePolicy, error) {
			return []*domain.GamePolicy{
				{
					GameID:                gameID,
					OwnerID:               "test-user",
					Mode:                  domain.OrchestrationModeKeepAlive,
					MaxPlayersPerInstance: 10,
					ScaleBehavior:         domain.ScaleBehaviorQueue,
					DefaultBuildVersion:   "latest",
				},
			}, nil
		},
	}

	instanceRepo := &hbMockInstanceRepo{
		listByGameFn: func(ctx context.Context, gameID int64, status *domain.InstanceStatus) ([]*domain.Instance, error) {
			return []*domain.Instance{
				{ID: 1, GameID: gameID, Status: domain.InstanceStatusRunning},
			}, nil
		},
	}

	instanceState := &hbMockInstanceState{
		getPlayerCountFn: func(ctx context.Context, id int64) (uint32, error) {
			return 10, nil // полностью заполнен
		},
	}

	instanceSvc := &hbMockInstanceOrchestrator{
		startInstanceFn: func(ctx context.Context, params StartInstanceParams) (*domain.Instance, error) {
			startedCount++
			return nil, nil
		},
	}

	svc := newTestHeartbeatService(&hbMockNodeRepo{}, &hbMockNodeStateStore{}, instanceRepo, instanceState, &hbMockNodeClient{}, nil, policyRepo, instanceSvc, config.NodeHeartbeatCfg{}, log)

	ctx := context.Background()
	svc.enforceScaleUp(ctx)

	time.Sleep(100 * time.Millisecond)

	if startedCount != 0 {
		t.Errorf("expected 0 starts for queue behavior, got %d", startedCount)
	}
}
