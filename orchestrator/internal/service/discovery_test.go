package service

import (
	"context"
	"testing"

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
func (m *discMockInstanceRepo) Update(ctx context.Context, instance *domain.Instance) error {
	return nil
}
func (m *discMockInstanceRepo) Delete(ctx context.Context, id int64) error { return nil }
func (m *discMockInstanceRepo) CountByGame(ctx context.Context, gameID int64) (int, error) {
	return 0, nil
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

	svc := NewDiscoveryService(instanceRepo, instanceState, nodeRepo)
	endpoints, err := svc.DiscoverServers(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(endpoints) != 2 {
		t.Fatalf("expected 2 endpoints, got %d", len(endpoints))
	}

	// Least-loaded first: instance 2 (2 players) before instance 1 (8 players).
	if endpoints[0].InstanceID != 2 {
		t.Errorf("expected first endpoint to be instance 2, got %d", endpoints[0].InstanceID)
	}
	if endpoints[1].InstanceID != 1 {
		t.Errorf("expected second endpoint to be instance 1, got %d", endpoints[1].InstanceID)
	}
}

func TestDiscoveryService_DiscoverServers_EmptyList(t *testing.T) {
	instanceRepo := &discMockInstanceRepo{
		listByGameFn: func(ctx context.Context, gameID int64, status *domain.InstanceStatus) ([]*domain.Instance, error) {
			return []*domain.Instance{}, nil
		},
	}
	svc := NewDiscoveryService(instanceRepo, &discMockInstanceState{}, &discMockNodeRepo{})
	endpoints, err := svc.DiscoverServers(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(endpoints) != 0 {
		t.Fatalf("expected 0 endpoints, got %d", len(endpoints))
	}
}

func TestDiscoveryService_DiscoverServers_ListError(t *testing.T) {
	instanceRepo := &discMockInstanceRepo{
		listByGameFn: func(ctx context.Context, gameID int64, status *domain.InstanceStatus) ([]*domain.Instance, error) {
			return nil, context.Canceled
		},
	}
	svc := NewDiscoveryService(instanceRepo, &discMockInstanceState{}, &discMockNodeRepo{})
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
			return instances, nil
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

	svc := NewDiscoveryService(instanceRepo, instanceState, nodeRepo)
	endpoints, err := svc.DiscoverServers(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(endpoints) != 0 {
		t.Fatalf("expected 0 endpoints (node not found), got %d", len(endpoints))
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

	svc := NewDiscoveryService(instanceRepo, instanceState, nodeRepo)
	endpoints, err := svc.DiscoverServers(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(endpoints) != 1 {
		t.Fatalf("expected 1 endpoint, got %d", len(endpoints))
	}
	if endpoints[0].PlayerCount == nil || *endpoints[0].PlayerCount != 0 {
		t.Errorf("expected player count 0 (default), got %v", endpoints[0].PlayerCount)
	}
}

func TestSortByPlayerCount(t *testing.T) {
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
