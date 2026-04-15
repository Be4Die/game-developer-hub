package service

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/config"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
)

// ─── Mocks for InstanceService ──────────────────────────────────────────────

type instMockInstanceRepo struct {
	createFn      func(ctx context.Context, instance *domain.Instance) error
	getByIDFn     func(ctx context.Context, id int64) (*domain.Instance, error)
	listByGameFn  func(ctx context.Context, gameID int64, status *domain.InstanceStatus) ([]*domain.Instance, error)
	listByNodeFn  func(ctx context.Context, nodeID int64) ([]*domain.Instance, error)
	updateFn      func(ctx context.Context, instance *domain.Instance) error
	deleteFn      func(ctx context.Context, id int64) error
	countByGameFn func(ctx context.Context, gameID int64) (int, error)
}

func (m *instMockInstanceRepo) Create(ctx context.Context, instance *domain.Instance) error {
	if m.createFn != nil {
		return m.createFn(ctx, instance)
	}
	return nil
}
func (m *instMockInstanceRepo) GetByID(ctx context.Context, id int64) (*domain.Instance, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}
func (m *instMockInstanceRepo) ListByGame(ctx context.Context, gameID int64, status *domain.InstanceStatus) ([]*domain.Instance, error) {
	if m.listByGameFn != nil {
		return m.listByGameFn(ctx, gameID, status)
	}
	return nil, nil
}
func (m *instMockInstanceRepo) ListByNode(ctx context.Context, nodeID int64) ([]*domain.Instance, error) {
	if m.listByNodeFn != nil {
		return m.listByNodeFn(ctx, nodeID)
	}
	return nil, nil
}
func (m *instMockInstanceRepo) Update(ctx context.Context, instance *domain.Instance) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, instance)
	}
	return nil
}
func (m *instMockInstanceRepo) Delete(ctx context.Context, id int64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}
func (m *instMockInstanceRepo) CountByGame(ctx context.Context, gameID int64) (int, error) {
	if m.countByGameFn != nil {
		return m.countByGameFn(ctx, gameID)
	}
	return 0, nil
}
func (m *instMockInstanceRepo) CountActiveInstancesByBuild(ctx context.Context, buildID int64) (int, error) {
	return 0, nil
}

type instMockInstanceState struct {
	setStatusFn      func(ctx context.Context, instanceID int64, status domain.InstanceStatus) error
	getStatusFn      func(ctx context.Context, instanceID int64) (domain.InstanceStatus, error)
	setPlayerCountFn func(ctx context.Context, instanceID int64, count uint32) error
	getPlayerCountFn func(ctx context.Context, instanceID int64) (uint32, error)
	setUsageFn       func(ctx context.Context, instanceID int64, usage *domain.ResourceUsage) error
	getUsageFn       func(ctx context.Context, instanceID int64) (*domain.ResourceUsage, error)
	deleteFn         func(ctx context.Context, instanceID int64) error
}

func (m *instMockInstanceState) SetStatus(ctx context.Context, instanceID int64, status domain.InstanceStatus) error {
	if m.setStatusFn != nil {
		return m.setStatusFn(ctx, instanceID, status)
	}
	return nil
}
func (m *instMockInstanceState) GetStatus(ctx context.Context, instanceID int64) (domain.InstanceStatus, error) {
	if m.getStatusFn != nil {
		return m.getStatusFn(ctx, instanceID)
	}
	return 0, nil
}
func (m *instMockInstanceState) SetPlayerCount(ctx context.Context, instanceID int64, count uint32) error {
	if m.setPlayerCountFn != nil {
		return m.setPlayerCountFn(ctx, instanceID, count)
	}
	return nil
}
func (m *instMockInstanceState) GetPlayerCount(ctx context.Context, instanceID int64) (uint32, error) {
	if m.getPlayerCountFn != nil {
		return m.getPlayerCountFn(ctx, instanceID)
	}
	return 0, nil
}
func (m *instMockInstanceState) SetUsage(ctx context.Context, instanceID int64, usage *domain.ResourceUsage) error {
	if m.setUsageFn != nil {
		return m.setUsageFn(ctx, instanceID, usage)
	}
	return nil
}
func (m *instMockInstanceState) GetUsage(ctx context.Context, instanceID int64) (*domain.ResourceUsage, error) {
	if m.getUsageFn != nil {
		return m.getUsageFn(ctx, instanceID)
	}
	return nil, nil
}
func (m *instMockInstanceState) Delete(ctx context.Context, instanceID int64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, instanceID)
	}
	return nil
}

type instMockBuildStorage struct {
	createFn                    func(ctx context.Context, build *domain.ServerBuild) error
	getByIDFn                   func(ctx context.Context, id int64) (*domain.ServerBuild, error)
	getByVersionFn              func(ctx context.Context, gameID int64, version string) (*domain.ServerBuild, error)
	listByGameFn                func(ctx context.Context, gameID int64, limit int) ([]*domain.ServerBuild, error)
	countByGameFn               func(ctx context.Context, gameID int64) (int, error)
	deleteFn                    func(ctx context.Context, id int64) error
	countActiveInstancesByBuild func(ctx context.Context, buildID int64) (int, error)
}

func (m *instMockBuildStorage) Create(ctx context.Context, build *domain.ServerBuild) error {
	if m.createFn != nil {
		return m.createFn(ctx, build)
	}
	return nil
}
func (m *instMockBuildStorage) GetByID(ctx context.Context, id int64) (*domain.ServerBuild, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}
func (m *instMockBuildStorage) GetByVersion(ctx context.Context, gameID int64, version string) (*domain.ServerBuild, error) {
	if m.getByVersionFn != nil {
		return m.getByVersionFn(ctx, gameID, version)
	}
	return nil, nil
}
func (m *instMockBuildStorage) ListByGame(ctx context.Context, gameID int64, limit int) ([]*domain.ServerBuild, error) {
	if m.listByGameFn != nil {
		return m.listByGameFn(ctx, gameID, limit)
	}
	return nil, nil
}
func (m *instMockBuildStorage) CountByGame(ctx context.Context, gameID int64) (int, error) {
	if m.countByGameFn != nil {
		return m.countByGameFn(ctx, gameID)
	}
	return 0, nil
}
func (m *instMockBuildStorage) Delete(ctx context.Context, id int64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}
func (m *instMockBuildStorage) CountActiveInstancesByBuild(ctx context.Context, buildID int64) (int, error) {
	if m.countActiveInstancesByBuild != nil {
		return m.countActiveInstancesByBuild(ctx, buildID)
	}
	return 0, nil
}

type instMockNodeRepo struct {
	createFn         func(ctx context.Context, node *domain.Node) error
	updateFn         func(ctx context.Context, node *domain.Node) error
	getByIDFn        func(ctx context.Context, id int64) (*domain.Node, error)
	getByAddressFn   func(ctx context.Context, addr string) (*domain.Node, error)
	listFn           func(ctx context.Context, status *domain.NodeStatus) ([]*domain.Node, error)
	deleteFn         func(ctx context.Context, id int64) error
	updateLastPingFn func(ctx context.Context, id int64) error
}

func (m *instMockNodeRepo) Create(ctx context.Context, node *domain.Node) error {
	if m.createFn != nil {
		return m.createFn(ctx, node)
	}
	return nil
}
func (m *instMockNodeRepo) Update(ctx context.Context, node *domain.Node) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, node)
	}
	return nil
}
func (m *instMockNodeRepo) GetByID(ctx context.Context, id int64) (*domain.Node, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}
func (m *instMockNodeRepo) GetByAddress(ctx context.Context, addr string) (*domain.Node, error) {
	if m.getByAddressFn != nil {
		return m.getByAddressFn(ctx, addr)
	}
	return nil, nil
}
func (m *instMockNodeRepo) List(ctx context.Context, status *domain.NodeStatus) ([]*domain.Node, error) {
	if m.listFn != nil {
		return m.listFn(ctx, status)
	}
	return nil, nil
}
func (m *instMockNodeRepo) Delete(ctx context.Context, id int64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}
func (m *instMockNodeRepo) UpdateLastPing(ctx context.Context, id int64) error {
	if m.updateLastPingFn != nil {
		return m.updateLastPingFn(ctx, id)
	}
	return nil
}

type instMockNodeStateStore struct {
	updateHeartbeatFn        func(ctx context.Context, nodeID int64, usage *domain.ResourceUsage) error
	getUsageFn               func(ctx context.Context, nodeID int64) (*domain.ResourceUsage, error)
	getActiveInstanceCountFn func(ctx context.Context, nodeID int64) (uint32, error)
	setActiveInstanceCountFn func(ctx context.Context, nodeID int64, count uint32) error
	deleteFn                 func(ctx context.Context, nodeID int64) error
}

func (m *instMockNodeStateStore) UpdateHeartbeat(ctx context.Context, nodeID int64, usage *domain.ResourceUsage) error {
	if m.updateHeartbeatFn != nil {
		return m.updateHeartbeatFn(ctx, nodeID, usage)
	}
	return nil
}
func (m *instMockNodeStateStore) GetUsage(ctx context.Context, nodeID int64) (*domain.ResourceUsage, error) {
	if m.getUsageFn != nil {
		return m.getUsageFn(ctx, nodeID)
	}
	return nil, nil
}
func (m *instMockNodeStateStore) GetActiveInstanceCount(ctx context.Context, nodeID int64) (uint32, error) {
	if m.getActiveInstanceCountFn != nil {
		return m.getActiveInstanceCountFn(ctx, nodeID)
	}
	return 0, nil
}
func (m *instMockNodeStateStore) SetActiveInstanceCount(ctx context.Context, nodeID int64, count uint32) error {
	if m.setActiveInstanceCountFn != nil {
		return m.setActiveInstanceCountFn(ctx, nodeID, count)
	}
	return nil
}
func (m *instMockNodeStateStore) Delete(ctx context.Context, nodeID int64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, nodeID)
	}
	return nil
}

type instMockNodeClient struct {
	loadImageFn        func(ctx context.Context, address string, meta domain.ImageMetadata, reader io.Reader) (*domain.ImageLoadResult, error)
	startInstanceFn    func(ctx context.Context, address string, req domain.StartInstanceRequest) (*domain.StartInstanceResult, error)
	stopInstanceFn     func(ctx context.Context, address string, instanceID int64, timeoutSec uint32) error
	streamLogsFn       func(ctx context.Context, address string, req domain.StreamLogsRequest) (domain.LogStream, error)
	getNodeInfoFn      func(ctx context.Context, address string) (*domain.NodeInfo, error)
	heartbeatFn        func(ctx context.Context, address string) (*domain.ResourceUsage, error)
	listInstancesFn    func(ctx context.Context, address string) ([]*domain.Instance, error)
	getInstanceFn      func(ctx context.Context, address string, instanceID int64) (*domain.Instance, error)
	getInstanceUsageFn func(ctx context.Context, address string, instanceID int64) (*domain.ResourceUsage, error)
}

func (m *instMockNodeClient) LoadImage(ctx context.Context, address string, meta domain.ImageMetadata, reader io.Reader) (*domain.ImageLoadResult, error) {
	if m.loadImageFn != nil {
		return m.loadImageFn(ctx, address, meta, reader)
	}
	return nil, nil
}
func (m *instMockNodeClient) StartInstance(ctx context.Context, address string, req domain.StartInstanceRequest) (*domain.StartInstanceResult, error) {
	if m.startInstanceFn != nil {
		return m.startInstanceFn(ctx, address, req)
	}
	return nil, nil
}
func (m *instMockNodeClient) StopInstance(ctx context.Context, address string, instanceID int64, timeoutSec uint32) error {
	if m.stopInstanceFn != nil {
		return m.stopInstanceFn(ctx, address, instanceID, timeoutSec)
	}
	return nil
}
func (m *instMockNodeClient) StreamLogs(ctx context.Context, address string, req domain.StreamLogsRequest) (domain.LogStream, error) {
	if m.streamLogsFn != nil {
		return m.streamLogsFn(ctx, address, req)
	}
	return nil, nil
}
func (m *instMockNodeClient) GetNodeInfo(ctx context.Context, address string) (*domain.NodeInfo, error) {
	if m.getNodeInfoFn != nil {
		return m.getNodeInfoFn(ctx, address)
	}
	return nil, nil
}
func (m *instMockNodeClient) Heartbeat(ctx context.Context, address string) (*domain.ResourceUsage, error) {
	if m.heartbeatFn != nil {
		return m.heartbeatFn(ctx, address)
	}
	return nil, nil
}
func (m *instMockNodeClient) ListInstances(ctx context.Context, address string) ([]*domain.Instance, error) {
	if m.listInstancesFn != nil {
		return m.listInstancesFn(ctx, address)
	}
	return nil, nil
}
func (m *instMockNodeClient) GetInstance(ctx context.Context, address string, instanceID int64) (*domain.Instance, error) {
	if m.getInstanceFn != nil {
		return m.getInstanceFn(ctx, address, instanceID)
	}
	return nil, nil
}
func (m *instMockNodeClient) GetInstanceUsage(ctx context.Context, address string, instanceID int64) (*domain.ResourceUsage, error) {
	if m.getInstanceUsageFn != nil {
		return m.getInstanceUsageFn(ctx, address, instanceID)
	}
	return nil, nil
}

// ─── Tests ───────────────────────────────────────────────────────────────────

func TestInstanceService_StartInstance(t *testing.T) {
	defaultLimits := config.LimitsConfig{MaxInstancesPerGame: 5}

	tests := []struct {
		name        string
		gameID      int64
		params      StartInstanceParams
		setupMocks  func() (*instMockInstanceRepo, *instMockInstanceState, *instMockBuildStorage, *instMockNodeRepo, *instMockNodeStateStore, *instMockNodeClient)
		wantErr     bool
		errContains string
	}{
		{
			name:   "success",
			gameID: 1,
			params: StartInstanceParams{
				GameID:       1,
				BuildVersion: "v1.0.0",
				Name:         "test-instance",
			},
			setupMocks: func() (*instMockInstanceRepo, *instMockInstanceState, *instMockBuildStorage, *instMockNodeRepo, *instMockNodeStateStore, *instMockNodeClient) {
				build := &domain.ServerBuild{
					ID:           100,
					GameID:       1,
					Version:      "v1.0.0",
					Protocol:     domain.ProtocolTCP,
					InternalPort: 7777,
					MaxPlayers:   10,
				}
				node := &domain.Node{
					ID:      10,
					Address: "node1:44044",
					Status:  domain.NodeStatusOnline,
				}

				instanceRepo := &instMockInstanceRepo{
					countByGameFn: func(ctx context.Context, gameID int64) (int, error) {
						return 2, nil
					},
					createFn: func(ctx context.Context, instance *domain.Instance) error {
						return nil
					},
					updateFn: func(ctx context.Context, instance *domain.Instance) error {
						return nil
					},
				}
				instanceState := &instMockInstanceState{
					setStatusFn: func(ctx context.Context, instanceID int64, status domain.InstanceStatus) error {
						return nil
					},
				}
				buildRepo := &instMockBuildStorage{
					getByVersionFn: func(ctx context.Context, gameID int64, version string) (*domain.ServerBuild, error) {
						return build, nil
					},
				}
				nodeRepo := &instMockNodeRepo{
					listFn: func(ctx context.Context, status *domain.NodeStatus) ([]*domain.Node, error) {
						return []*domain.Node{node}, nil
					},
				}
				nodeState := &instMockNodeStateStore{
					getActiveInstanceCountFn: func(ctx context.Context, nodeID int64) (uint32, error) {
						return 0, nil
					},
					setActiveInstanceCountFn: func(ctx context.Context, nodeID int64, count uint32) error {
						return nil
					},
				}
				nodeClient := &instMockNodeClient{
					startInstanceFn: func(ctx context.Context, address string, req domain.StartInstanceRequest) (*domain.StartInstanceResult, error) {
						return &domain.StartInstanceResult{
							InstanceID: 42,
							HostPort:   8001,
						}, nil
					},
				}
				return instanceRepo, instanceState, buildRepo, nodeRepo, nodeState, nodeClient
			},
			wantErr: false,
		},
		{
			name:   "limit reached",
			gameID: 1,
			params: StartInstanceParams{
				GameID:       1,
				BuildVersion: "v1.0.0",
				Name:         "test-instance",
			},
			setupMocks: func() (*instMockInstanceRepo, *instMockInstanceState, *instMockBuildStorage, *instMockNodeRepo, *instMockNodeStateStore, *instMockNodeClient) {
				instanceRepo := &instMockInstanceRepo{
					countByGameFn: func(ctx context.Context, gameID int64) (int, error) {
						return 5, nil // equals MaxInstancesPerGame
					},
				}
				return instanceRepo, &instMockInstanceState{}, &instMockBuildStorage{}, &instMockNodeRepo{}, &instMockNodeStateStore{}, &instMockNodeClient{}
			},
			wantErr:     true,
			errContains: "max instances limit reached",
		},
		{
			name:   "build not found",
			gameID: 1,
			params: StartInstanceParams{
				GameID:       1,
				BuildVersion: "v1.0.0",
				Name:         "test-instance",
			},
			setupMocks: func() (*instMockInstanceRepo, *instMockInstanceState, *instMockBuildStorage, *instMockNodeRepo, *instMockNodeStateStore, *instMockNodeClient) {
				instanceRepo := &instMockInstanceRepo{
					countByGameFn: func(ctx context.Context, gameID int64) (int, error) {
						return 2, nil
					},
				}
				buildRepo := &instMockBuildStorage{
					getByVersionFn: func(ctx context.Context, gameID int64, version string) (*domain.ServerBuild, error) {
						return nil, domain.ErrNotFound
					},
				}
				return instanceRepo, &instMockInstanceState{}, buildRepo, &instMockNodeRepo{}, &instMockNodeStateStore{}, &instMockNodeClient{}
			},
			wantErr:     true,
			errContains: "get build",
		},
		{
			name:   "no available node",
			gameID: 1,
			params: StartInstanceParams{
				GameID:       1,
				BuildVersion: "v1.0.0",
				Name:         "test-instance",
			},
			setupMocks: func() (*instMockInstanceRepo, *instMockInstanceState, *instMockBuildStorage, *instMockNodeRepo, *instMockNodeStateStore, *instMockNodeClient) {
				build := &domain.ServerBuild{
					ID:       100,
					GameID:   1,
					Version:  "v1.0.0",
					Protocol: domain.ProtocolTCP,
				}
				instanceRepo := &instMockInstanceRepo{
					countByGameFn: func(ctx context.Context, gameID int64) (int, error) {
						return 2, nil
					},
				}
				buildRepo := &instMockBuildStorage{
					getByVersionFn: func(ctx context.Context, gameID int64, version string) (*domain.ServerBuild, error) {
						return build, nil
					},
				}
				// No online nodes.
				nodeRepo := &instMockNodeRepo{
					listFn: func(ctx context.Context, status *domain.NodeStatus) ([]*domain.Node, error) {
						return []*domain.Node{
							{ID: 1, Address: "node1:44044", Status: domain.NodeStatusOffline},
						}, nil
					},
				}
				return instanceRepo, &instMockInstanceState{}, buildRepo, nodeRepo, &instMockNodeStateStore{}, &instMockNodeClient{}
			},
			wantErr:     true,
			errContains: "select node",
		},
		{
			name:   "node StartInstance error",
			gameID: 1,
			params: StartInstanceParams{
				GameID:       1,
				BuildVersion: "v1.0.0",
				Name:         "test-instance",
			},
			setupMocks: func() (*instMockInstanceRepo, *instMockInstanceState, *instMockBuildStorage, *instMockNodeRepo, *instMockNodeStateStore, *instMockNodeClient) {
				build := &domain.ServerBuild{
					ID:       100,
					GameID:   1,
					Version:  "v1.0.0",
					Protocol: domain.ProtocolTCP,
				}
				node := &domain.Node{
					ID:      10,
					Address: "node1:44044",
					Status:  domain.NodeStatusOnline,
				}
				instanceRepo := &instMockInstanceRepo{
					countByGameFn: func(ctx context.Context, gameID int64) (int, error) {
						return 2, nil
					},
				}
				buildRepo := &instMockBuildStorage{
					getByVersionFn: func(ctx context.Context, gameID int64, version string) (*domain.ServerBuild, error) {
						return build, nil
					},
				}
				nodeRepo := &instMockNodeRepo{
					listFn: func(ctx context.Context, status *domain.NodeStatus) ([]*domain.Node, error) {
						return []*domain.Node{node}, nil
					},
				}
				nodeState := &instMockNodeStateStore{
					getActiveInstanceCountFn: func(ctx context.Context, nodeID int64) (uint32, error) {
						return 0, nil
					},
				}
				nodeClient := &instMockNodeClient{
					startInstanceFn: func(ctx context.Context, address string, req domain.StartInstanceRequest) (*domain.StartInstanceResult, error) {
						return nil, errors.New("grpc error")
					},
				}
				return instanceRepo, &instMockInstanceState{}, buildRepo, nodeRepo, nodeState, nodeClient
			},
			wantErr:     true,
			errContains: "node StartInstance",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			instanceRepo, instanceState, buildRepo, nodeRepo, nodeState, nodeClient := tc.setupMocks()

			svc := NewInstanceService(
				instanceRepo,
				instanceState,
				buildRepo,
				nodeRepo,
				nodeState,
				nodeClient,
				defaultLimits,
			)

			ctx := context.Background()
			result, err := svc.StartInstance(ctx, tc.params)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tc.errContains != "" && !containsStr(err.Error(), tc.errContains) {
					t.Errorf("expected error containing %q, got %q", tc.errContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result == nil {
				t.Fatal("expected result, got nil")
			}
			if result.ID != 42 {
				t.Errorf("expected instance ID 42, got %d", result.ID)
			}
			if result.Status != domain.InstanceStatusRunning {
				t.Errorf("expected status running, got %s", result.Status)
			}
			if result.GameID != tc.gameID {
				t.Errorf("expected game ID %d, got %d", tc.gameID, result.GameID)
			}
		})
	}
}

func TestInstanceService_StopInstance(t *testing.T) {
	tests := []struct {
		name        string
		gameID      int64
		instanceID  int64
		timeoutSec  uint32
		setupMocks  func() (*instMockInstanceRepo, *instMockInstanceState, *instMockBuildStorage, *instMockNodeRepo, *instMockNodeStateStore, *instMockNodeClient)
		wantErr     bool
		errContains string
	}{
		{
			name:       "success",
			gameID:     1,
			instanceID: 42,
			timeoutSec: 10,
			setupMocks: func() (*instMockInstanceRepo, *instMockInstanceState, *instMockBuildStorage, *instMockNodeRepo, *instMockNodeStateStore, *instMockNodeClient) {
				instance := &domain.Instance{
					ID:     42,
					NodeID: 10,
					GameID: 1,
					Status: domain.InstanceStatusRunning,
				}
				node := &domain.Node{
					ID:      10,
					Address: "node1:44044",
					Status:  domain.NodeStatusOnline,
				}

				instanceRepo := &instMockInstanceRepo{
					getByIDFn: func(ctx context.Context, id int64) (*domain.Instance, error) {
						return instance, nil
					},
					updateFn: func(ctx context.Context, inst *domain.Instance) error {
						return nil
					},
				}
				instanceState := &instMockInstanceState{
					deleteFn: func(ctx context.Context, instanceID int64) error {
						return nil
					},
				}
				nodeRepo := &instMockNodeRepo{
					getByIDFn: func(ctx context.Context, id int64) (*domain.Node, error) {
						return node, nil
					},
				}
				nodeState := &instMockNodeStateStore{
					getActiveInstanceCountFn: func(ctx context.Context, nodeID int64) (uint32, error) {
						return 1, nil
					},
					setActiveInstanceCountFn: func(ctx context.Context, nodeID int64, count uint32) error {
						return nil
					},
				}
				stopCalled := false
				nodeClient := &instMockNodeClient{
					stopInstanceFn: func(ctx context.Context, address string, instanceID int64, timeoutSec uint32) error {
						stopCalled = true
						return nil
					},
				}
				_ = stopCalled
				return instanceRepo, instanceState, &instMockBuildStorage{}, nodeRepo, nodeState, nodeClient
			},
			wantErr: false,
		},
		{
			name:       "wrong game ID",
			gameID:     1,
			instanceID: 42,
			timeoutSec: 10,
			setupMocks: func() (*instMockInstanceRepo, *instMockInstanceState, *instMockBuildStorage, *instMockNodeRepo, *instMockNodeStateStore, *instMockNodeClient) {
				// Instance belongs to game 2, but we pass gameID=1.
				instance := &domain.Instance{
					ID:     42,
					NodeID: 10,
					GameID: 2,
					Status: domain.InstanceStatusRunning,
				}
				instanceRepo := &instMockInstanceRepo{
					getByIDFn: func(ctx context.Context, id int64) (*domain.Instance, error) {
						return instance, nil
					},
				}
				return instanceRepo, &instMockInstanceState{}, &instMockBuildStorage{}, &instMockNodeRepo{}, &instMockNodeStateStore{}, &instMockNodeClient{}
			},
			wantErr:     true,
			errContains: "does not belong to game",
		},
		{
			name:       "instance not found",
			gameID:     1,
			instanceID: 42,
			timeoutSec: 10,
			setupMocks: func() (*instMockInstanceRepo, *instMockInstanceState, *instMockBuildStorage, *instMockNodeRepo, *instMockNodeStateStore, *instMockNodeClient) {
				instanceRepo := &instMockInstanceRepo{
					getByIDFn: func(ctx context.Context, id int64) (*domain.Instance, error) {
						return nil, domain.ErrNotFound
					},
				}
				return instanceRepo, &instMockInstanceState{}, &instMockBuildStorage{}, &instMockNodeRepo{}, &instMockNodeStateStore{}, &instMockNodeClient{}
			},
			wantErr:     true,
			errContains: "get instance",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			instanceRepo, instanceState, buildRepo, nodeRepo, nodeState, nodeClient := tc.setupMocks()

			svc := NewInstanceService(
				instanceRepo,
				instanceState,
				buildRepo,
				nodeRepo,
				nodeState,
				nodeClient,
				config.LimitsConfig{MaxInstancesPerGame: 5},
			)

			ctx := context.Background()
			result, err := svc.StopInstance(ctx, tc.gameID, tc.instanceID, tc.timeoutSec)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tc.errContains != "" && !containsStr(err.Error(), tc.errContains) {
					t.Errorf("expected error containing %q, got %q", tc.errContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result == nil {
				t.Fatal("expected result, got nil")
			}
			if result.Status != domain.InstanceStatusStopped {
				t.Errorf("expected status stopped, got %s", result.Status)
			}
		})
	}
}

func TestInstanceService_ListInstances_Success(t *testing.T) {
	instances := []*domain.Instance{
		{ID: 1, NodeID: 10, GameID: 1, HostPort: 7001, Protocol: domain.ProtocolTCP, Status: domain.InstanceStatusRunning},
		{ID: 2, NodeID: 10, GameID: 1, HostPort: 7002, Protocol: domain.ProtocolUDP, Status: domain.InstanceStatusRunning},
	}

	instanceRepo := &instMockInstanceRepo{
		listByGameFn: func(ctx context.Context, gameID int64, status *domain.InstanceStatus) ([]*domain.Instance, error) {
			return instances, nil
		},
	}

	instanceState := &instMockInstanceState{
		getStatusFn: func(ctx context.Context, instanceID int64) (domain.InstanceStatus, error) {
			if instanceID == 1 {
				return domain.InstanceStatusRunning, nil
			}
			return domain.InstanceStatusStopping, nil
		},
		getPlayerCountFn: func(ctx context.Context, instanceID int64) (uint32, error) {
			if instanceID == 1 {
				return 5, nil
			}
			return 0, nil
		},
	}

	svc := NewInstanceService(
		instanceRepo,
		instanceState,
		&instMockBuildStorage{},
		&instMockNodeRepo{},
		&instMockNodeStateStore{},
		&instMockNodeClient{},
		config.LimitsConfig{MaxInstancesPerGame: 5},
	)

	ctx := context.Background()
	result, err := svc.ListInstances(ctx, 1, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 instances, got %d", len(result))
	}

	// Check KV enrichment: instance 1 should have status from KV.
	if result[0].Status != domain.InstanceStatusRunning {
		t.Errorf("expected instance 1 status from KV to be running, got %s", result[0].Status)
	}
	pc := result[0].PlayerCount
	if pc == nil || *pc != 5 {
		t.Errorf("expected instance 1 player count 5, got %v", pc)
	}
}

func TestInstanceService_GetInstance_KVFallback(t *testing.T) {
	instance := &domain.Instance{
		ID:     42,
		NodeID: 10,
		GameID: 1,
		Status: domain.InstanceStatusStarting,
	}

	instanceRepo := &instMockInstanceRepo{
		getByIDFn: func(ctx context.Context, id int64) (*domain.Instance, error) {
			return instance, nil
		},
	}

	// KV returns errors — enrichment should gracefully fall back.
	instanceState := &instMockInstanceState{
		getStatusFn: func(ctx context.Context, instanceID int64) (domain.InstanceStatus, error) {
			return 0, domain.ErrNotFound
		},
		getPlayerCountFn: func(ctx context.Context, instanceID int64) (uint32, error) {
			return 0, domain.ErrNotFound
		},
	}

	svc := NewInstanceService(
		instanceRepo,
		instanceState,
		&instMockBuildStorage{},
		&instMockNodeRepo{},
		&instMockNodeStateStore{},
		&instMockNodeClient{},
		config.LimitsConfig{MaxInstancesPerGame: 5},
	)

	ctx := context.Background()
	result, err := svc.GetInstance(ctx, 1, 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	// Status should remain the PG value since KV failed.
	if result.Instance.Status != domain.InstanceStatusStarting {
		t.Errorf("expected instance status from PG to be starting, got %s", result.Instance.Status)
	}
	// PlayerCount should be nil since KV failed.
	if result.PlayerCount != nil {
		t.Errorf("expected player count to be nil when KV fails, got %v", result.PlayerCount)
	}
}

// containsStr checks if s contains substr.
func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && searchStr(s, substr)
}

func searchStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
