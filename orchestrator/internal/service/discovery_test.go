package service

import (
	"context"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
)

// ─── Mocks for DiscoveryService ─────────────────────────────────────────────

type discMockInstanceRepo struct {
	listByGameFn func(ctx context.Context, gameID int64, status *domain.InstanceStatus) ([]*domain.Instance, error)
}

func (m *discMockInstanceRepo) Create(ctx context.Context, instance *domain.Instance) error {
	return nil
}
func (m *discMockInstanceRepo) GetByID(ctx context.Context, id int64) (*domain.Instance, error) {
	return nil, nil
}
func (m *discMockInstanceRepo) ListByGame(ctx context.Context, gameID int64, status *domain.InstanceStatus) ([]*domain.Instance, error) {
	return m.listByGameFn(ctx, gameID, status)
}
func (m *discMockInstanceRepo) ListByNode(ctx context.Context, nodeID int64) ([]*domain.Instance, error) {
	return nil, nil
}
func (m *discMockInstanceRepo) List(ctx context.Context) ([]*domain.Instance, error) {
	return nil, nil
}
func (m *discMockInstanceRepo) Update(ctx context.Context, instance *domain.Instance) error {
	return nil
}
func (m *discMockInstanceRepo) Delete(ctx context.Context, id int64) error { return nil }
func (m *discMockInstanceRepo) CountByGame(ctx context.Context, gameID int64) (int, error) {
	return 0, nil
}
func (m *discMockInstanceRepo) GetNextID(ctx context.Context) (int64, error) {
	return 1, nil
}

type discMockInstanceState struct {
	getPlayerCountFn func(ctx context.Context, instanceID int64) (uint32, error)
	getStatusFn      func(ctx context.Context, instanceID int64) (domain.InstanceStatus, error)
}

func (m *discMockInstanceState) SetStatus(ctx context.Context, instanceID int64, status domain.InstanceStatus) error {
	return nil
}
func (m *discMockInstanceState) GetStatus(ctx context.Context, instanceID int64) (domain.InstanceStatus, error) {
	if m.getStatusFn != nil {
		return m.getStatusFn(ctx, instanceID)
	}
	return 0, nil
}
func (m *discMockInstanceState) SetPlayerCount(ctx context.Context, instanceID int64, count uint32) error {
	return nil
}
func (m *discMockInstanceState) GetPlayerCount(ctx context.Context, instanceID int64) (uint32, error) {
	return m.getPlayerCountFn(ctx, instanceID)
}
func (m *discMockInstanceState) SetUsage(ctx context.Context, instanceID int64, usage *domain.ResourceUsage) error {
	return nil
}
func (m *discMockInstanceState) GetUsage(ctx context.Context, instanceID int64) (*domain.ResourceUsage, error) {
	return nil, nil
}
func (m *discMockInstanceState) Delete(ctx context.Context, instanceID int64) error { return nil }
func (m *discMockInstanceState) SetZeroPlayersSince(ctx context.Context, instanceID int64, t time.Time) error {
	return nil
}
func (m *discMockInstanceState) GetZeroPlayersSince(ctx context.Context, instanceID int64) (time.Time, error) {
	return time.Time{}, domain.ErrNotFound
}
func (m *discMockInstanceState) DeleteZeroPlayersSince(ctx context.Context, instanceID int64) error {
	return nil
}

type discMockBuildStorage struct {
	listByGameFn func(ctx context.Context, gameID int64, limit int) ([]*domain.ServerBuild, error)
}

func (m *discMockBuildStorage) Create(ctx context.Context, build *domain.ServerBuild) error { return nil }
func (m *discMockBuildStorage) GetByID(ctx context.Context, id int64) (*domain.ServerBuild, error) {
	return nil, nil
}
func (m *discMockBuildStorage) GetByVersion(ctx context.Context, gameID int64, version string) (*domain.ServerBuild, error) {
	return nil, nil
}
func (m *discMockBuildStorage) ListByGame(ctx context.Context, gameID int64, limit int) ([]*domain.ServerBuild, error) {
	if m.listByGameFn != nil {
		return m.listByGameFn(ctx, gameID, limit)
	}
	return nil, nil
}
func (m *discMockBuildStorage) CountByGame(ctx context.Context, gameID int64) (int, error) { return 0, nil }
func (m *discMockBuildStorage) Delete(ctx context.Context, id int64) error { return nil }
func (m *discMockBuildStorage) CountActiveInstancesByBuild(ctx context.Context, buildID int64) (int, error) {
	return 0, nil
}

type discMockGamePolicyRepo struct {
	getFn func(ctx context.Context, gameID int64) (*domain.GamePolicy, error)
}

func (m *discMockGamePolicyRepo) Get(ctx context.Context, gameID int64) (*domain.GamePolicy, error) {
	if m.getFn != nil {
		return m.getFn(ctx, gameID)
	}
	return nil, domain.ErrNotFound
}
func (m *discMockGamePolicyRepo) Set(ctx context.Context, policy *domain.GamePolicy) error { return nil }
func (m *discMockGamePolicyRepo) Delete(ctx context.Context, gameID int64) error { return nil }
func (m *discMockGamePolicyRepo) ListAll(ctx context.Context) ([]*domain.GamePolicy, error) { return nil, nil }

type discMockInstanceStarter struct {
	startInstanceFn func(ctx context.Context, params StartInstanceParams) (*domain.Instance, error)
}

func (m *discMockInstanceStarter) StartInstance(ctx context.Context, params StartInstanceParams) (*domain.Instance, error) {
	if m.startInstanceFn != nil {
		return m.startInstanceFn(ctx, params)
	}
	return nil, nil
}

type discMockNodeRepo struct {
	getByIDFn func(ctx context.Context, id int64) (*domain.Node, error)
}

func (m *discMockNodeRepo) Create(ctx context.Context, node *domain.Node) error { return nil }
func (m *discMockNodeRepo) Update(ctx context.Context, node *domain.Node) error { return nil }
func (m *discMockNodeRepo) GetByID(ctx context.Context, id int64) (*domain.Node, error) {
	return m.getByIDFn(ctx, id)
}
func (m *discMockNodeRepo) GetByAddress(ctx context.Context, addr string) (*domain.Node, error) {
	return nil, nil
}
func (m *discMockNodeRepo) List(ctx context.Context, status *domain.NodeStatus) ([]*domain.Node, error) {
	return nil, nil
}
func (m *discMockNodeRepo) Delete(ctx context.Context, id int64) error         { return nil }
func (m *discMockNodeRepo) UpdateLastPing(ctx context.Context, id int64) error { return nil }

func newTestDiscoveryService(
	instanceRepo domain.InstanceRepo,
	instanceState domain.InstanceStateStore,
	nodeRepo domain.NodeRepo,
	buildRepo domain.BuildStorage,
	policyRepo domain.GamePolicyRepo,
	instanceSvc instanceStarter,
) *DiscoveryService {
	if policyRepo == nil {
		policyRepo = &discMockGamePolicyRepo{}
	}
	policyService := NewGamePolicyService(policyRepo)
	if instanceSvc == nil {
		instanceSvc = &discMockInstanceStarter{}
	}
	return NewDiscoveryService(instanceRepo, instanceState, nodeRepo, buildRepo, policyService, instanceSvc)
}

// ─── Tests ───────────────────────────────────────────────────────────────────

func TestDiscoveryService_DiscoverServers_Success(t *testing.T) {
	instances := []*domain.Instance{
		{ID: 1, NodeID: 10, HostPort: 7001, Protocol: domain.ProtocolTCP, MaxPlayers: 10},
		{ID: 2, NodeID: 11, HostPort: 7002, Protocol: domain.ProtocolUDP, MaxPlayers: 20},
	}

	instanceRepo := &discMockInstanceRepo{
		listByGameFn: func(ctx context.Context, gameID int64, status *domain.InstanceStatus) ([]*domain.Instance, error) {
			return instances, nil
		},
	}

	instanceState := &discMockInstanceState{
		getPlayerCountFn: func(ctx context.Context, instanceID int64) (uint32, error) {
			if instanceID == 1 {
				return 8, nil
			}
			return 2, nil
		},
	}

	nodeRepo := &discMockNodeRepo{
		getByIDFn: func(ctx context.Context, id int64) (*domain.Node, error) {
			return &domain.Node{ID: id, Address: "node.example.com:44044"}, nil
		},
	}

	svc := newTestDiscoveryService(instanceRepo, instanceState, nodeRepo, nil, nil, nil)
	result, err := svc.DiscoverServers(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != domain.DiscoveryStatusReady {
		t.Fatalf("expected status READY, got %v", result.Status)
	}
	if len(result.Servers) != 2 {
		t.Fatalf("expected 2 endpoints, got %d", len(result.Servers))
	}

	// Least-loaded first: instance 2 (2 players) before instance 1 (8 players).
	if result.Servers[0].InstanceID != 2 {
		t.Errorf("expected first endpoint to be instance 2, got %d", result.Servers[0].InstanceID)
	}
	if result.Servers[1].InstanceID != 1 {
		t.Errorf("expected second endpoint to be instance 1, got %d", result.Servers[1].InstanceID)
	}
}

func TestDiscoveryService_DiscoverServers_EmptyList(t *testing.T) {
	instanceRepo := &discMockInstanceRepo{
		listByGameFn: func(ctx context.Context, gameID int64, status *domain.InstanceStatus) ([]*domain.Instance, error) {
			return []*domain.Instance{}, nil
		},
	}
	svc := newTestDiscoveryService(instanceRepo, &discMockInstanceState{}, &discMockNodeRepo{}, nil, nil, nil)
	result, err := svc.DiscoverServers(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != domain.DiscoveryStatusUnavailable {
		t.Fatalf("expected status UNAVAILABLE (no policy), got %v", result.Status)
	}
}

func TestDiscoveryService_DiscoverServers_ListError(t *testing.T) {
	instanceRepo := &discMockInstanceRepo{
		listByGameFn: func(ctx context.Context, gameID int64, status *domain.InstanceStatus) ([]*domain.Instance, error) {
			return nil, context.Canceled
		},
	}
	svc := newTestDiscoveryService(instanceRepo, &discMockInstanceState{}, &discMockNodeRepo{}, nil, nil, nil)
	_, err := svc.DiscoverServers(context.Background(), 42)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDiscoveryService_DiscoverServers_NodeNotFound(t *testing.T) {
	instances := []*domain.Instance{
		{ID: 1, NodeID: 99, HostPort: 7001, Protocol: domain.ProtocolTCP, MaxPlayers: 10},
	}

	instanceRepo := &discMockInstanceRepo{
		listByGameFn: func(ctx context.Context, gameID int64, status *domain.InstanceStatus) ([]*domain.Instance, error) {
			if status != nil && *status == domain.InstanceStatusRunning {
				return instances, nil
			}
			return []*domain.Instance{}, nil
		},
	}
	instanceState := &discMockInstanceState{
		getPlayerCountFn: func(ctx context.Context, instanceID int64) (uint32, error) {
			return 5, nil
		},
	}
	nodeRepo := &discMockNodeRepo{
		getByIDFn: func(ctx context.Context, id int64) (*domain.Node, error) {
			return nil, domain.ErrNotFound
		},
	}

	svc := newTestDiscoveryService(instanceRepo, instanceState, nodeRepo, nil, nil, nil)
	result, err := svc.DiscoverServers(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != domain.DiscoveryStatusUnavailable {
		t.Fatalf("expected status UNAVAILABLE (no running and no policy), got %v", result.Status)
	}
	if len(result.Servers) != 0 {
		t.Fatalf("expected 0 endpoints (node not found), got %d", len(result.Servers))
	}
}

func TestDiscoveryService_DiscoverServerError_GetPlayerCount(t *testing.T) {
	instances := []*domain.Instance{
		{ID: 1, NodeID: 10, HostPort: 7001, Protocol: domain.ProtocolTCP, MaxPlayers: 10},
	}

	instanceRepo := &discMockInstanceRepo{
		listByGameFn: func(ctx context.Context, gameID int64, status *domain.InstanceStatus) ([]*domain.Instance, error) {
			return instances, nil
		},
	}
	instanceState := &discMockInstanceState{
		getPlayerCountFn: func(ctx context.Context, instanceID int64) (uint32, error) {
			return 0, context.DeadlineExceeded // KV error — should be ignored
		},
	}
	nodeRepo := &discMockNodeRepo{
		getByIDFn: func(ctx context.Context, id int64) (*domain.Node, error) {
			return &domain.Node{ID: id, Address: "node:44044"}, nil
		},
	}

	svc := newTestDiscoveryService(instanceRepo, instanceState, nodeRepo, nil, nil, nil)
	result, err := svc.DiscoverServers(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != domain.DiscoveryStatusReady {
		t.Fatalf("expected status READY, got %v", result.Status)
	}
	if len(result.Servers) != 1 {
		t.Fatalf("expected 1 endpoint, got %d", len(result.Servers))
	}
	if result.Servers[0].PlayerCount == nil || *result.Servers[0].PlayerCount != 0 {
		t.Errorf("expected player count 0 (default), got %v", result.Servers[0].PlayerCount)
	}
}

func TestDiscoveryService_SortByPlayerCount(t *testing.T) {
	a := uint32(5)
	b := uint32(1)
	c := uint32(10)
	endpoints := []domain.ServerEndpoint{
		{InstanceID: 1, PlayerCount: &a},
		{InstanceID: 2, PlayerCount: &b},
		{InstanceID: 3, PlayerCount: &c},
	}

	sortByPlayerCount(endpoints)

	if endpoints[0].InstanceID != 2 {
		t.Errorf("expected first = 2, got %d", endpoints[0].InstanceID)
	}
	if endpoints[1].InstanceID != 1 {
		t.Errorf("expected second = 1, got %d", endpoints[1].InstanceID)
	}
	if endpoints[2].InstanceID != 3 {
		t.Errorf("expected third = 3, got %d", endpoints[2].InstanceID)
	}
}

func TestSortByPlayerCount_NilValues(t *testing.T) {
	a := uint32(5)
	endpoints := []domain.ServerEndpoint{
		{InstanceID: 1, PlayerCount: &a},
		{InstanceID: 2, PlayerCount: nil},
	}

	sortByPlayerCount(endpoints)

	if endpoints[0].InstanceID != 2 {
		t.Errorf("expected nil player count first, got %d", endpoints[0].InstanceID)
	}
}

func TestDiscoveryService_DiscoverServers_AutoStart(t *testing.T) {
	instanceRepo := &discMockInstanceRepo{
		listByGameFn: func(ctx context.Context, gameID int64, status *domain.InstanceStatus) ([]*domain.Instance, error) {
			return []*domain.Instance{}, nil
		},
	}

	started := false
	instanceSvc := &discMockInstanceStarter{
		startInstanceFn: func(ctx context.Context, params StartInstanceParams) (*domain.Instance, error) {
			started = true
			return &domain.Instance{ID: 99}, nil
		},
	}

	policyRepo := &discMockGamePolicyRepo{
		getFn: func(ctx context.Context, gameID int64) (*domain.GamePolicy, error) {
			return &domain.GamePolicy{
				GameID:              gameID,
				OwnerID:             "test-user",
				Mode:                domain.OrchestrationModeKeepAlive,
				TargetInstances:     1,
				DefaultBuildVersion: "latest",
				MaxInstancesPerGame: 5,
				ScaleBehavior:       domain.ScaleBehaviorSpawn,
			}, nil
		},
	}

	buildRepo := &discMockBuildStorage{
		listByGameFn: func(ctx context.Context, gameID int64, limit int) ([]*domain.ServerBuild, error) {
			return []*domain.ServerBuild{{Version: "v1.0.0"}}, nil
		},
	}

	svc := newTestDiscoveryService(instanceRepo, &discMockInstanceState{}, &discMockNodeRepo{}, buildRepo, policyRepo, instanceSvc)

	result, err := svc.DiscoverServers(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != domain.DiscoveryStatusStarting {
		t.Fatalf("expected status STARTING, got %v", result.Status)
	}

	// autoStartInstance запускается в отдельной горутине.
	time.Sleep(100 * time.Millisecond)

	if !started {
		t.Error("expected auto-start to be triggered when no running instances exist")
	}
}

func TestDiscoveryService_DiscoverServers_FullAndQueue(t *testing.T) {
	instances := []*domain.Instance{
		{ID: 1, NodeID: 10, HostPort: 7001, Protocol: domain.ProtocolTCP, MaxPlayers: 10},
	}

	instanceRepo := &discMockInstanceRepo{
		listByGameFn: func(ctx context.Context, gameID int64, status *domain.InstanceStatus) ([]*domain.Instance, error) {
			if status != nil && *status == domain.InstanceStatusRunning {
				return instances, nil
			}
			return []*domain.Instance{}, nil
		},
	}
	instanceState := &discMockInstanceState{
		getPlayerCountFn: func(ctx context.Context, instanceID int64) (uint32, error) {
			return 10, nil // Full
		},
	}
	nodeRepo := &discMockNodeRepo{
		getByIDFn: func(ctx context.Context, id int64) (*domain.Node, error) {
			return &domain.Node{ID: id, Address: "node:44044"}, nil
		},
	}
	policyRepo := &discMockGamePolicyRepo{
		getFn: func(ctx context.Context, gameID int64) (*domain.GamePolicy, error) {
			return &domain.GamePolicy{
				GameID:              gameID,
				Mode:                domain.OrchestrationModeKeepAlive,
				MaxInstancesPerGame: 1,
				ScaleBehavior:       domain.ScaleBehaviorQueue,
			}, nil
		},
	}

	buildRepo := &discMockBuildStorage{
		listByGameFn: func(ctx context.Context, gameID int64, limit int) ([]*domain.ServerBuild, error) {
			return []*domain.ServerBuild{{Version: "v1.0.0"}}, nil
		},
	}

	svc := newTestDiscoveryService(instanceRepo, instanceState, nodeRepo, buildRepo, policyRepo, nil)
	result, err := svc.DiscoverServers(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != domain.DiscoveryStatusCapacityReached {
		t.Fatalf("expected status CAPACITY_REACHED, got %v", result.Status)
	}
}

func TestDiscoveryService_DiscoverServers_StartingInstances(t *testing.T) {
	instanceRepo := &discMockInstanceRepo{
		listByGameFn: func(ctx context.Context, gameID int64, status *domain.InstanceStatus) ([]*domain.Instance, error) {
			if status != nil && *status == domain.InstanceStatusStarting {
				return []*domain.Instance{
					{ID: 5, NodeID: 10, HostPort: 7005, Protocol: domain.ProtocolTCP, MaxPlayers: 10},
				}, nil
			}
			return []*domain.Instance{}, nil
		},
	}
	nodeRepo := &discMockNodeRepo{
		getByIDFn: func(ctx context.Context, id int64) (*domain.Node, error) {
			return &domain.Node{ID: id, Address: "node:44044"}, nil
		},
	}

	svc := newTestDiscoveryService(instanceRepo, &discMockInstanceState{}, nodeRepo, nil, nil, nil)
	result, err := svc.DiscoverServers(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != domain.DiscoveryStatusStarting {
		t.Fatalf("expected status STARTING, got %v", result.Status)
	}
}
