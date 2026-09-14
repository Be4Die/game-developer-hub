package service

import (
	"context"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/infrastructure/config"
)

type mockPlatformAccessRepo struct {
	grants            map[int64]*domain.PlatformGrant
	platformInstances map[int64]int32
	allocatedCPU      map[int64]uint32
	allocatedMem      map[int64]uint64
}

func newMockPlatformAccessRepo() *mockPlatformAccessRepo {
	return &mockPlatformAccessRepo{
		grants:            make(map[int64]*domain.PlatformGrant),
		platformInstances: make(map[int64]int32),
		allocatedCPU:      make(map[int64]uint32),
		allocatedMem:      make(map[int64]uint64),
	}
}

func (m *mockPlatformAccessRepo) SaveGrant(ctx context.Context, g *domain.PlatformGrant) (*domain.PlatformGrant, error) {
	clone := *g
	clone.IsActive = true
	clone.CreatedAt = time.Now()
	clone.UpdatedAt = time.Now()
	m.grants[g.GameID] = &clone
	return &clone, nil
}

func (m *mockPlatformAccessRepo) RevokeGrant(ctx context.Context, gameID int64) error {
	if g, ok := m.grants[gameID]; ok {
		g.IsActive = false
	}
	return nil
}

func (m *mockPlatformAccessRepo) GetGrant(ctx context.Context, gameID int64) (*domain.PlatformGrant, error) {
	g, ok := m.grants[gameID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return g, nil
}

func (m *mockPlatformAccessRepo) HasApprovedAccess(ctx context.Context, gameID int64) (bool, *domain.PlatformGrant, error) {
	g, ok := m.grants[gameID]
	if !ok || !g.IsActive {
		return false, nil, nil
	}
	return true, g, nil
}

func (m *mockPlatformAccessRepo) CountPlatformInstances(ctx context.Context, gameID int64) (int32, error) {
	return m.platformInstances[gameID], nil
}

func (m *mockPlatformAccessRepo) GetActivePlatformUsage(ctx context.Context, gameID int64) (count int32, allocatedCPUMillis uint32, allocatedMemoryBytes uint64, err error) {
	return m.platformInstances[gameID], m.allocatedCPU[gameID], m.allocatedMem[gameID], nil
}

func TestPlatformAccessService_Workflow(t *testing.T) {
	repo := newMockPlatformAccessRepo()
	svc := NewPlatformAccessService(repo)
	ctx := context.Background()

	// 1. Grant access
	grant, err := svc.GrantAccess(ctx, 101, 3, 4000, 8192, 2000, 4096)
	if err != nil {
		t.Fatalf("GrantAccess failed: %v", err)
	}
	if !grant.IsActive || grant.MaxInstances != 3 || grant.MaxTotalCPUMillis != 4000 || grant.MaxTotalMemoryMB != 8192 {
		t.Errorf("grant = %+v, want active with 3 instances and limits", grant)
	}

	// 2. Get grant
	got, active, err := svc.GetGrant(ctx, 101)
	if err != nil || got == nil {
		t.Fatalf("GetGrant failed: %v", err)
	}
	if !got.IsActive || active != 0 {
		t.Errorf("got active = %d, IsActive = %v", active, got.IsActive)
	}

	// 3. Revoke access
	if err := svc.RevokeAccess(ctx, 101); err != nil {
		t.Fatalf("RevokeAccess failed: %v", err)
	}
	gotRevoked, _, err := svc.GetGrant(ctx, 101)
	if err != nil || gotRevoked.IsActive {
		t.Errorf("expected revoked grant, got: %+v", gotRevoked)
	}
}

func TestPlatformNodes_Scheduling_Enforcement(t *testing.T) {
	ctx := context.Background()
	platformRepo := newMockPlatformAccessRepo()

	nodes := []*domain.Node{
		{
			ID:         1,
			OwnerID:    "admin",
			Address:    "platform-node-1:44044",
			Status:     domain.NodeStatusOnline,
			Role:       domain.NodeRoleCompute,
			IsPlatform: true,
		},
		{
			ID:         2,
			OwnerID:    "dev-other",
			Address:    "other-node:44044",
			Status:     domain.NodeStatusOnline,
			Role:       domain.NodeRoleCompute,
			IsPlatform: false,
		},
	}

	nodeRepo := &instMockNodeRepo{
		listFn: func(ctx context.Context, status *domain.NodeStatus) ([]*domain.Node, error) {
			return nodes, nil
		},
		getByIDFn: func(ctx context.Context, id int64) (*domain.Node, error) {
			for _, n := range nodes {
				if n.ID == id {
					return n, nil
				}
			}
			return nil, domain.ErrNotFound
		},
	}
	nodeState := &instMockNodeStateStore{}

	instSvc := NewInstanceService(
		&instMockInstanceRepo{},
		&instMockInstanceState{},
		&instMockBuildStorage{},
		nodeRepo,
		nodeState,
		&instMockNodeClient{},
		config.LimitsConfig{MaxInstancesPerGame: 5},
	).WithPlatformAccessRepo(platformRepo)

	build := &domain.ServerBuild{
		ID:      1,
		OwnerID: "dev-1",
		GameID:  101,
		Version: "1.0.0",
	}

	// 1. Without access: attempt to schedule on platform node
	_, err := instSvc.selectNodeForInstance(ctx, build, "auto")
	if err != domain.ErrPlatformAccessRequired {
		t.Errorf("expected ErrPlatformAccessRequired, got: %v", err)
	}

	// 2. Approve access with quota = 2 and CPU limit = 2000
	_, _ = platformRepo.SaveGrant(ctx, &domain.PlatformGrant{
		GameID:               101,
		MaxInstances:         2,
		MaxTotalCPUMillis:    2000,
		MaxTotalMemoryMB:     4096,
		MaxInstanceCPUMillis: 1000,
		MaxInstanceMemoryMB:  2048,
		IsActive:             true,
	})

	// Now should select platform node
	node, err := instSvc.selectNodeForInstance(ctx, build, "auto")
	if err != nil {
		t.Fatalf("expected node selection, got err: %v", err)
	}
	if node.ID != 1 || !node.IsPlatform {
		t.Errorf("expected platform node 1, got node ID: %d", node.ID)
	}

	// 3. Quota exceeded: set running count to 2
	platformRepo.platformInstances[101] = 2
	_, err = instSvc.selectNodeForInstance(ctx, build, "auto")
	if err != domain.ErrPlatformAccessRequired && err != domain.ErrPlatformQuotaExceeded && err != domain.ErrNoAvailableNode {
		t.Errorf("expected quota or availability error, got: %v", err)
	}

	// Reset count, but exceed CPU
	platformRepo.platformInstances[101] = 1
	platformRepo.allocatedCPU[101] = 2000
	_, err = instSvc.selectNodeForInstance(ctx, build, "auto")
	if err != domain.ErrPlatformAccessRequired && err != domain.ErrPlatformQuotaExceeded && err != domain.ErrNoAvailableNode {
		t.Errorf("expected CPU quota error, got: %v", err)
	}
}
