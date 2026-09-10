package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
)

// testLogger возвращает логгер, который не пишет в вывод тестов.
func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestNodeService_RegisterNode_Manual_Success(t *testing.T) {
	token := "test-token"
	address := "node1:44044"
	tokenHash := sha256.Sum256([]byte(token))
	var createdNode *domain.Node
	var mu sync.Mutex

	nodeRepo := &hbMockNodeRepo{
		getByAddressFn: func(ctx context.Context, addr string) (*domain.Node, error) {
			return nil, domain.ErrNotFound
		},
		createFn: func(ctx context.Context, node *domain.Node) error {
			mu.Lock()
			defer mu.Unlock()
			createdNode = node
			return nil
		},
	}

	nodeClient := &hbMockNodeClient{
		getNodeInfoFn: func(ctx context.Context, addr, apiKey string) (*domain.NodeInfo, error) {
			return &domain.NodeInfo{
				CPUCores:         8,
				TotalMemoryBytes: 16 * 1024 * 1024 * 1024,
				TotalDiskBytes:   500 * 1024 * 1024 * 1024,
				AgentVersion:     "1.0.0",
			}, nil
		},
	}

	svc := NewNodeService(testLogger(), nodeRepo, &hbMockNodeStateStore{}, &hbMockInstanceRepo{}, &hbMockInstanceState{}, nodeClient)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	result, err := svc.RegisterNode(ctx, RegisterNodeParams{
		Address: address,
		Token:   token,
		Region:  "us-east",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result node, got nil")
	}

	mu.Lock()
	if createdNode == nil {
		t.Fatal("expected node to be created in repo")
	}
	mu.Unlock()

	if result.Address != address {
		t.Errorf("expected address %q, got %q", address, result.Address)
	}
	if result.Region != "us-east" {
		t.Errorf("expected region %q, got %q", "us-east", result.Region)
	}
	if result.Status != domain.NodeStatusOnline {
		t.Errorf("expected status Online, got %v", result.Status)
	}
	if !constantTimeEqual(result.TokenHash, tokenHash[:]) {
		t.Error("token hash mismatch")
	}
	if result.CPUCores != 8 {
		t.Errorf("expected 8 CPU cores, got %d", result.CPUCores)
	}
	if result.AgentVersion != "1.0.0" {
		t.Errorf("expected agent version %q, got %q", "1.0.0", result.AgentVersion)
	}
}

func TestNodeService_RegisterNode_Manual_AlreadyExists(t *testing.T) {
	address := "node1:44044"
	nodeRepo := &hbMockNodeRepo{
		getByAddressFn: func(ctx context.Context, addr string) (*domain.Node, error) {
			return &domain.Node{
				ID:      1,
				Address: address,
				Status:  domain.NodeStatusOnline,
			}, nil
		},
	}

	getNodeInfoCalled := false
	nodeClient := &hbMockNodeClient{
		getNodeInfoFn: func(ctx context.Context, addr, apiKey string) (*domain.NodeInfo, error) {
			getNodeInfoCalled = true
			return nil, nil
		},
	}

	svc := NewNodeService(testLogger(), nodeRepo, &hbMockNodeStateStore{}, &hbMockInstanceRepo{}, &hbMockInstanceState{}, nodeClient)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	_, err := svc.RegisterNode(ctx, RegisterNodeParams{
		Address: address,
		Token:   "some-token",
	})

	if !errors.Is(err, domain.ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists, got %v", err)
	}

	if getNodeInfoCalled {
		t.Error("GetNodeInfo should not be called when node already exists")
	}
}

func TestNodeService_RegisterNode_Authorize_Success(t *testing.T) {
	token := "correct-token"
	tokenHash := sha256.Sum256([]byte(token))
	nodeID := int64(42)
	var updatedNode *domain.Node
	var mu sync.Mutex

	nodeRepo := &hbMockNodeRepo{
		getByIDFn: func(ctx context.Context, id int64) (*domain.Node, error) {
			return &domain.Node{
				ID:        nodeID,
				Address:   "node1:44044",
				TokenHash: tokenHash[:],
				Status:    domain.NodeStatusUnauthorized,
			}, nil
		},
		updateFn: func(ctx context.Context, node *domain.Node) error {
			mu.Lock()
			defer mu.Unlock()
			updatedNode = node
			return nil
		},
	}

	svc := NewNodeService(testLogger(), nodeRepo, &hbMockNodeStateStore{}, &hbMockInstanceRepo{}, &hbMockInstanceState{}, &hbMockNodeClient{})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	result, err := svc.RegisterNode(ctx, RegisterNodeParams{
		NodeID: &nodeID,
		Token:  token,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result node, got nil")
	}

	if result.Status != domain.NodeStatusOnline {
		t.Errorf("expected status Online, got %v", result.Status)
	}

	mu.Lock()
	if updatedNode == nil {
		t.Fatal("expected node to be updated in repo")
	}
	if updatedNode.Status != domain.NodeStatusOnline {
		t.Errorf("expected updated status Online, got %v", updatedNode.Status)
	}
	mu.Unlock()
}

func TestNodeService_RegisterNode_Authorize_InvalidToken(t *testing.T) {
	correctToken := "correct-token"
	tokenHash := sha256.Sum256([]byte(correctToken))
	nodeID := int64(42)
	wrongToken := "wrong-token"
	nodeRepo := &hbMockNodeRepo{
		getByIDFn: func(ctx context.Context, id int64) (*domain.Node, error) {
			return &domain.Node{
				ID:        nodeID,
				Address:   "node1:44044",
				TokenHash: tokenHash[:],
				Status:    domain.NodeStatusUnauthorized,
			}, nil
		},
	}

	svc := NewNodeService(testLogger(), nodeRepo, &hbMockNodeStateStore{}, &hbMockInstanceRepo{}, &hbMockInstanceState{}, &hbMockNodeClient{})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	_, err := svc.RegisterNode(ctx, RegisterNodeParams{
		NodeID: &nodeID,
		Token:  wrongToken,
	})

	if !errors.Is(err, domain.ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestNodeService_RegisterNode_Authorize_NotUnauthorized(t *testing.T) {
	token := "some-token"
	tokenHash := sha256.Sum256([]byte(token))
	nodeID := int64(42)
	nodeRepo := &hbMockNodeRepo{
		getByIDFn: func(ctx context.Context, id int64) (*domain.Node, error) {
			return &domain.Node{
				ID:        nodeID,
				Address:   "node1:44044",
				TokenHash: tokenHash[:],
				Status:    domain.NodeStatusOnline, // already online, not unauthorized
			}, nil
		},
	}

	svc := NewNodeService(testLogger(), nodeRepo, &hbMockNodeStateStore{}, &hbMockInstanceRepo{}, &hbMockInstanceState{}, &hbMockNodeClient{})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	_, err := svc.RegisterNode(ctx, RegisterNodeParams{
		NodeID: &nodeID,
		Token:  token,
	})

	if !errors.Is(err, domain.ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestNodeService_ListNodes_Success(t *testing.T) {
	now := time.Now()
	nodes := []*domain.Node{
		{ID: 1, Address: "node1:44044", Status: domain.NodeStatusOnline, LastPingAt: now},
		{ID: 2, Address: "node2:44044", Status: domain.NodeStatusOffline, LastPingAt: now.Add(-120 * time.Second)},
	}
	nodeRepo := &hbMockNodeRepo{
		listFn: func(ctx context.Context, status *domain.NodeStatus) ([]*domain.Node, error) {
			return nodes, nil
		},
	}

	usage := &domain.ResourceUsage{CPUUsagePercent: 55.0}
	activeCount := uint32(3)

	nodeState := &hbMockNodeStateStore{
		getUsageFn: func(ctx context.Context, nodeID int64) (*domain.ResourceUsage, error) {
			return usage, nil
		},
		getActiveInstanceCountFn: func(ctx context.Context, nodeID int64) (uint32, error) {
			return activeCount, nil
		},
	}

	svc := NewNodeService(testLogger(), nodeRepo, nodeState, &hbMockInstanceRepo{}, &hbMockInstanceState{}, &hbMockNodeClient{})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	result, err := svc.ListNodes(ctx, "", nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != len(nodes) {
		t.Fatalf("expected %d nodes, got %d", len(nodes), len(result))
	}

	for i, r := range result {
		if r.ID != nodes[i].ID {
			t.Errorf("node[%d]: expected ID %d, got %d", i, nodes[i].ID, r.ID)
		}
		if r.Usage == nil {
			t.Errorf("node[%d]: expected Usage, got nil", i)
		} else if r.Usage.CPUUsagePercent != 55.0 {
			t.Errorf("node[%d]: expected CPUUsagePercent 55.0, got %f", i, r.Usage.CPUUsagePercent)
		}
		if r.ActiveInstanceCount == nil {
			t.Errorf("node[%d]: expected ActiveInstanceCount, got nil", i)
		} else if *r.ActiveInstanceCount != activeCount {
			t.Errorf("node[%d]: expected ActiveInstanceCount %d, got %d", i, activeCount, *r.ActiveInstanceCount)
		}
	}
}

func TestNodeService_GetNode_Success(t *testing.T) {
	nodeID := int64(1)
	now := time.Now()
	node := &domain.Node{
		ID:         nodeID,
		Address:    "node1:44044",
		Status:     domain.NodeStatusOnline,
		LastPingAt: now,
	}
	nodeRepo := &hbMockNodeRepo{
		getByIDFn: func(ctx context.Context, id int64) (*domain.Node, error) {
			return node, nil
		},
	}

	usage := &domain.ResourceUsage{CPUUsagePercent: 30.0}
	activeCount := uint32(5)

	nodeState := &hbMockNodeStateStore{
		getUsageFn: func(ctx context.Context, id int64) (*domain.ResourceUsage, error) {
			return usage, nil
		},
		getActiveInstanceCountFn: func(ctx context.Context, id int64) (uint32, error) {
			return activeCount, nil
		},
	}

	svc := NewNodeService(testLogger(), nodeRepo, nodeState, &hbMockInstanceRepo{}, &hbMockInstanceState{}, &hbMockNodeClient{})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	result, err := svc.GetNode(ctx, "", nodeID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != nodeID {
		t.Errorf("expected node ID %d, got %d", nodeID, result.ID)
	}
	if result.Usage == nil {
		t.Fatal("expected Usage, got nil")
	}
	if result.Usage.CPUUsagePercent != 30.0 {
		t.Errorf("expected CPUUsagePercent 30.0, got %f", result.Usage.CPUUsagePercent)
	}
	if result.ActiveInstanceCount == nil {
		t.Fatal("expected ActiveInstanceCount, got nil")
	}
	if *result.ActiveInstanceCount != activeCount {
		t.Errorf("expected ActiveInstanceCount %d, got %d", activeCount, *result.ActiveInstanceCount)
	}
}

func TestNodeService_DeleteNode_Success(t *testing.T) {
	nodeID := int64(1)
	instance1 := &domain.Instance{ID: 10, NodeID: nodeID, Status: domain.InstanceStatusRunning}
	instance2 := &domain.Instance{ID: 11, NodeID: nodeID, Status: domain.InstanceStatusRunning}
	var updatedInstances []*domain.Instance
	var mu sync.Mutex
	var kvDeleted bool
	var pgDeleted bool

	instanceRepo := &hbMockInstanceRepo{
		listByNodeFn: func(ctx context.Context, nid int64) ([]*domain.Instance, error) {
			return []*domain.Instance{instance1, instance2}, nil
		},
		updateFn: func(ctx context.Context, inst *domain.Instance) error {
			mu.Lock()
			defer mu.Unlock()
			updatedInstances = append(updatedInstances, inst)
			return nil
		},
	}

	instanceState := &hbMockInstanceState{
		setStatusFn: func(ctx context.Context, instanceID int64, status domain.InstanceStatus) error {
			return nil
		},
	}

	nodeState := &hbMockNodeStateStore{
		deleteFn: func(ctx context.Context, nid int64) error {
			kvDeleted = true
			return nil
		},
	}

	nodeRepo := &hbMockNodeRepo{
		deleteFn: func(ctx context.Context, nid int64) error {
			pgDeleted = true
			return nil
		},
	}

	svc := NewNodeService(testLogger(), nodeRepo, nodeState, instanceRepo, instanceState, &hbMockNodeClient{})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	err := svc.DeleteNode(ctx, "", nodeID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mu.Lock()
	if len(updatedInstances) != 2 {
		t.Fatalf("expected 2 instance updates, got %d", len(updatedInstances))
	}
	for _, inst := range updatedInstances {
		if inst.Status != domain.InstanceStatusCrashed {
			t.Errorf("expected instance %d status Crashed, got %v", inst.ID, inst.Status)
		}
	}
	mu.Unlock()

	if !kvDeleted {
		t.Error("expected KV state to be deleted")
	}
	if !pgDeleted {
		t.Error("expected PG record to be deleted")
	}
}

func TestNodeService_GetNodeUsage_Success(t *testing.T) {
	nodeID := int64(1)
	node := &domain.Node{
		ID:     nodeID,
		Status: domain.NodeStatusOnline,
	}
	nodeRepo := &hbMockNodeRepo{
		getByIDFn: func(ctx context.Context, id int64) (*domain.Node, error) {
			return node, nil
		},
	}

	expectedUsage := &domain.ResourceUsage{
		CPUUsagePercent: 72.5,
		MemoryUsedBytes: 8 * 1024 * 1024 * 1024,
		DiskUsedBytes:   200 * 1024 * 1024 * 1024,
	}
	expectedActiveCount := uint32(10)

	nodeState := &hbMockNodeStateStore{
		getUsageFn: func(ctx context.Context, id int64) (*domain.ResourceUsage, error) {
			return expectedUsage, nil
		},
		getActiveInstanceCountFn: func(ctx context.Context, id int64) (uint32, error) {
			return expectedActiveCount, nil
		},
	}

	svc := NewNodeService(testLogger(), nodeRepo, nodeState, &hbMockInstanceRepo{}, &hbMockInstanceState{}, &hbMockNodeClient{})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	result, err := svc.GetNodeUsage(ctx, "", nodeID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.NodeID != nodeID {
		t.Errorf("expected NodeID %d, got %d", nodeID, result.NodeID)
	}
	if result.Usage == nil {
		t.Fatal("expected Usage, got nil")
	}
	if result.Usage.CPUUsagePercent != expectedUsage.CPUUsagePercent {
		t.Errorf("expected CPUUsagePercent %f, got %f", expectedUsage.CPUUsagePercent, result.Usage.CPUUsagePercent)
	}
	if result.ActiveInstanceCount != expectedActiveCount {
		t.Errorf("expected ActiveInstanceCount %d, got %d", expectedActiveCount, result.ActiveInstanceCount)
	}
}

func TestNodeService_ConstantTimeEqual(t *testing.T) {
	hash1 := sha256.Sum256([]byte("hello"))
	hash2 := sha256.Sum256([]byte("hello"))
	hash3 := sha256.Sum256([]byte("world"))
	tests := []struct {
		name     string
		a        []byte
		b        []byte
		expected bool
	}{
		{
			name:     "equal hashes",
			a:        hash1[:],
			b:        hash2[:],
			expected: true,
		},
		{
			name:     "different hashes",
			a:        hash1[:],
			b:        hash3[:],
			expected: false,
		},
		{
			name:     "different lengths",
			a:        hash1[:],
			b:        hash3[:16],
			expected: false,
		},
		{
			name:     "nil slices",
			a:        nil,
			b:        nil,
			expected: true,
		},
		{
			name:     "one nil slice",
			a:        hash1[:],
			b:        nil,
			expected: false,
		},
		{
			name:     "empty slices",
			a:        []byte{},
			b:        []byte{},
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := constantTimeEqual(tc.a, tc.b)
			if result != tc.expected {
				t.Errorf("constantTimeEqual = %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestNodeService_AnnounceNode_Success(t *testing.T) {
	address := "192.168.1.100:44044"
	var createdNode *domain.Node
	var mu sync.Mutex
	nodeRepo := &hbMockNodeRepo{
		getByAddressFn: func(ctx context.Context, addr string) (*domain.Node, error) {
			return nil, domain.ErrNotFound
		},
		createFn: func(ctx context.Context, node *domain.Node) error {
			mu.Lock()
			defer mu.Unlock()
			createdNode = node
			node.ID = 42
			return nil
		},
	}

	svc := NewNodeService(testLogger(), nodeRepo, &hbMockNodeStateStore{}, &hbMockInstanceRepo{}, &hbMockInstanceState{}, &hbMockNodeClient{})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	result, err := svc.AnnounceNode(ctx, AnnounceNodeParams{
		Address:          address,
		Region:           "EU",
		AgentVersion:     "1.2.3",
		CPUCores:         8,
		TotalMemoryBytes: 16 * 1024 * 1024 * 1024,
		TotalDiskBytes:   500 * 1024 * 1024 * 1024,
		APIKey:           "test-api-key",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}

	if result.NodeID != 42 {
		t.Errorf("expected NodeID 42, got %d", result.NodeID)
	}

	mu.Lock()
	if createdNode == nil {
		t.Fatal("expected node to be created in repo")
	}
	if createdNode.Status != domain.NodeStatusUnauthorized {
		t.Errorf("expected status Unauthorized, got %v", createdNode.Status)
	}
	if createdNode.OwnerID != "" {
		t.Errorf("expected empty OwnerID, got %q", createdNode.OwnerID)
	}
	if createdNode.Address != address {
		t.Errorf("expected address %q, got %q", address, createdNode.Address)
	}
	mu.Unlock()
}

func TestNodeService_AnnounceNode_UpdateExistingUnauthorized(t *testing.T) {
	address := "192.168.1.100:44044"
	existingNode := &domain.Node{
		ID:        42,
		Address:   address,
		Status:    domain.NodeStatusUnauthorized,
		OwnerID:   "",
		TokenHash: []byte("old-hash"),
	}
	var updatedNode *domain.Node
	var mu sync.Mutex

	nodeRepo := &hbMockNodeRepo{
		getByAddressFn: func(ctx context.Context, addr string) (*domain.Node, error) {
			return existingNode, nil
		},
		updateFn: func(ctx context.Context, node *domain.Node) error {
			mu.Lock()
			defer mu.Unlock()
			updatedNode = node
			return nil
		},
	}

	svc := NewNodeService(testLogger(), nodeRepo, &hbMockNodeStateStore{}, &hbMockInstanceRepo{}, &hbMockInstanceState{}, &hbMockNodeClient{})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	result, err := svc.AnnounceNode(ctx, AnnounceNodeParams{
		Address:          address,
		Region:           "US",
		AgentVersion:     "2.0.0",
		CPUCores:         16,
		TotalMemoryBytes: 32 * 1024 * 1024 * 1024,
		TotalDiskBytes:   1000 * 1024 * 1024 * 1024,
		APIKey:           "test-api-key",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.NodeID != 42 {
		t.Errorf("expected NodeID 42, got %d", result.NodeID)
	}

	mu.Lock()
	if updatedNode == nil {
		t.Fatal("expected node to be updated in repo")
	}
	if updatedNode.Region != "US" {
		t.Errorf("expected region US, got %q", updatedNode.Region)
	}
	if updatedNode.AgentVersion != "2.0.0" {
		t.Errorf("expected agent version 2.0.0, got %q", updatedNode.AgentVersion)
	}
	mu.Unlock()
}

func TestNodeService_AnnounceNode_AlreadyOnline(t *testing.T) {
	address := "192.168.1.100:44044"
	existingNode := &domain.Node{
		ID:      42,
		Address: address,
		Status:  domain.NodeStatusOnline,
		OwnerID: "user-123",
	}
	nodeRepo := &hbMockNodeRepo{
		getByAddressFn: func(ctx context.Context, addr string) (*domain.Node, error) {
			return existingNode, nil
		},
	}

	svc := NewNodeService(testLogger(), nodeRepo, &hbMockNodeStateStore{}, &hbMockInstanceRepo{}, &hbMockInstanceState{}, &hbMockNodeClient{})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	_, err := svc.AnnounceNode(ctx, AnnounceNodeParams{
		Address:          address,
		AgentVersion:     "1.0.0",
		CPUCores:         8,
		TotalMemoryBytes: 16 * 1024 * 1024 * 1024,
		TotalDiskBytes:   500 * 1024 * 1024 * 1024,
		APIKey:           "test-api-key",
	})

	if !errors.Is(err, domain.ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists, got %v", err)
	}
}

type mockManagedServiceRepo struct {
	createFn      func(ctx context.Context, s *domain.ManagedService) error
	getByIDFn     func(ctx context.Context, id int64) (*domain.ManagedService, error)
	getByNameFn   func(ctx context.Context, nodeID int64, name string) (*domain.ManagedService, error)
	listByNodeFn  func(ctx context.Context, nodeID int64) ([]*domain.ManagedService, error)
	listByOwnerFn func(ctx context.Context, ownerID string) ([]*domain.ManagedService, error)
	listByGameFn  func(ctx context.Context, gameID int64) ([]*domain.ManagedService, error)
	updateFn      func(ctx context.Context, s *domain.ManagedService) error
	deleteFn      func(ctx context.Context, id int64) error
}

func (m *mockManagedServiceRepo) Create(ctx context.Context, s *domain.ManagedService) error {
	if m.createFn != nil {
		return m.createFn(ctx, s)
	}
	return nil
}
func (m *mockManagedServiceRepo) GetByID(ctx context.Context, id int64) (*domain.ManagedService, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}
func (m *mockManagedServiceRepo) GetByName(ctx context.Context, nodeID int64, name string) (*domain.ManagedService, error) {
	if m.getByNameFn != nil {
		return m.getByNameFn(ctx, nodeID, name)
	}
	return nil, nil
}
func (m *mockManagedServiceRepo) ListByNode(ctx context.Context, nodeID int64) ([]*domain.ManagedService, error) {
	if m.listByNodeFn != nil {
		return m.listByNodeFn(ctx, nodeID)
	}
	return nil, nil
}
func (m *mockManagedServiceRepo) ListByOwner(ctx context.Context, ownerID string) ([]*domain.ManagedService, error) {
	if m.listByOwnerFn != nil {
		return m.listByOwnerFn(ctx, ownerID)
	}
	return nil, nil
}
func (m *mockManagedServiceRepo) ListByGame(ctx context.Context, gameID int64) ([]*domain.ManagedService, error) {
	if m.listByGameFn != nil {
		return m.listByGameFn(ctx, gameID)
	}
	return nil, nil
}
func (m *mockManagedServiceRepo) Update(ctx context.Context, s *domain.ManagedService) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, s)
	}
	return nil
}
func (m *mockManagedServiceRepo) Delete(ctx context.Context, id int64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func TestNodeService_UpdateRole(t *testing.T) {
	node := &domain.Node{
		ID:      10,
		OwnerID: "user-1",
		Role:    domain.NodeRoleMixed,
	}
	var updatedRole domain.NodeRole
	nodeRepo := &hbMockNodeRepo{
		getByIDFn: func(ctx context.Context, id int64) (*domain.Node, error) {
			return node, nil
		},
		updateRoleFn: func(ctx context.Context, id int64, role domain.NodeRole) error {
			updatedRole = role
			return nil
		},
	}

	svc := NewNodeService(testLogger(), nodeRepo, &hbMockNodeStateStore{}, &hbMockInstanceRepo{}, &hbMockInstanceState{}, &hbMockNodeClient{})

	ctx := context.Background()
	updatedNode, err := svc.UpdateRole(ctx, "user-1", 10, domain.NodeRoleStorage, domain.StorageTransitionActionUnspecified, domain.ComputeTransitionActionUnspecified)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updatedNode.Role != domain.NodeRoleStorage || updatedRole != domain.NodeRoleStorage {
		t.Errorf("expected role %v, got %v (updatedNode=%v)", domain.NodeRoleStorage, updatedRole, updatedNode.Role)
	}

	// Test forbidden owner
	_, err = svc.UpdateRole(ctx, "wrong-user", 10, domain.NodeRoleCompute, domain.StorageTransitionActionUnspecified, domain.ComputeTransitionActionUnspecified)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestNodeService_CreateService(t *testing.T) {
	node := &domain.Node{
		ID:       10,
		OwnerID:  "user-1",
		Role:     domain.NodeRoleCompute,
		Address:  "127.0.0.1:44044",
		APIToken: "token",
		Status:   domain.NodeStatusOnline,
	}

	nodeRepo := &hbMockNodeRepo{
		getByIDFn: func(ctx context.Context, id int64) (*domain.Node, error) {
			return node, nil
		},
	}

	svcRepo := &mockManagedServiceRepo{}
	svc := NewNodeService(testLogger(), nodeRepo, &hbMockNodeStateStore{}, &hbMockInstanceRepo{}, &hbMockInstanceState{}, &hbMockNodeClient{}).WithServiceRepo(svcRepo)

	ctx := context.Background()
	// Should fail on Compute-only node
	_, err := svc.CreateService(ctx, "user-1", 10, CreateServiceParams{
		ServiceType: domain.ServiceTypePostgres,
		Name:        "my-pg",
	})
	if err == nil {
		t.Fatal("expected error on compute-only node, got nil")
	}

	// Change role to Mixed
	node.Role = domain.NodeRoleMixed
	var deployedReq domain.DeployServiceRequest
	nodeClient := &hbMockNodeClient{
		DeployServiceFn: func(ctx context.Context, nodeAddress, apiKey string, req domain.DeployServiceRequest) (*domain.DeployServiceResult, error) {
			deployedReq = req
			return &domain.DeployServiceResult{
				ContainerID:   "c-123",
				HostPort:      5432,
				ConnectionURI: "postgresql://postgres:pass@127.0.0.1:5432/game_db",
				VolumePath:    "/var/lib/gdh/volumes/my-pg",
			}, nil
		},
	}

	var createdService *domain.ManagedService
	svcRepo.createFn = func(ctx context.Context, s *domain.ManagedService) error {
		createdService = s
		s.ID = 100
		return nil
	}

	svc = NewNodeService(testLogger(), nodeRepo, &hbMockNodeStateStore{}, &hbMockInstanceRepo{}, &hbMockInstanceState{}, nodeClient).WithServiceRepo(svcRepo)

	result, err := svc.CreateService(ctx, "user-1", 10, CreateServiceParams{
		ServiceType:    domain.ServiceTypePostgres,
		Name:           "my-pg",
		Password:       "pass",
		DBName:         "game_db",
		AllowedGameIDs: []int64{1, 2},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != 100 {
		t.Errorf("expected ID 100, got %d", result.ID)
	}
	if deployedReq.Name != "my-pg" {
		t.Errorf("expected deployed name my-pg, got %s", deployedReq.Name)
	}
	if createdService.HostPort != 5432 {
		t.Errorf("expected port 5432, got %d", createdService.HostPort)
	}
}

func TestNodeService_ListServices_FilterByGame(t *testing.T) {
	node := &domain.Node{
		ID:      10,
		OwnerID: "user-1",
	}
	nodeRepo := &hbMockNodeRepo{
		getByIDFn: func(ctx context.Context, id int64) (*domain.Node, error) {
			return node, nil
		},
	}
	allServices := []*domain.ManagedService{
		{ID: 1, Name: "global-redis", AllowedGameIDs: nil},
		{ID: 2, Name: "game1-pg", AllowedGameIDs: []int64{1}},
		{ID: 3, Name: "game2-pg", AllowedGameIDs: []int64{2}},
	}
	svcRepo := &mockManagedServiceRepo{
		listByNodeFn: func(ctx context.Context, nodeID int64) ([]*domain.ManagedService, error) {
			return allServices, nil
		},
	}

	svc := NewNodeService(testLogger(), nodeRepo, &hbMockNodeStateStore{}, &hbMockInstanceRepo{}, &hbMockInstanceState{}, &hbMockNodeClient{}).WithServiceRepo(svcRepo)

	ctx := context.Background()
	// Without game filter: returns all 3
	res, err := svc.ListServices(ctx, "user-1", 10, nil)
	if err != nil || len(res) != 3 {
		t.Fatalf("expected 3 services, got %d (err: %v)", len(res), err)
	}

	// Filter by game 1: should return global-redis and game1-pg (total 2)
	g1 := int64(1)
	res, err = svc.ListServices(ctx, "user-1", 10, &g1)
	if err != nil || len(res) != 2 {
		t.Fatalf("expected 2 services for game 1, got %d (err: %v)", len(res), err)
	}
}

func TestNodeService_DeleteService(t *testing.T) {
	node := &domain.Node{
		ID:       10,
		OwnerID:  "user-1",
		Status:   domain.NodeStatusOnline,
		Address:  "127.0.0.1:44044",
		APIToken: "tok",
	}
	nodeRepo := &hbMockNodeRepo{
		getByIDFn: func(ctx context.Context, id int64) (*domain.Node, error) {
			return node, nil
		},
	}

	removedName := ""
	nodeClient := &hbMockNodeClient{
		RemoveServiceFn: func(ctx context.Context, nodeAddress, apiKey string, name string, deleteVolume bool) error {
			removedName = name
			return nil
		},
	}

	deletedID := int64(0)
	svcRepo := &mockManagedServiceRepo{
		getByIDFn: func(ctx context.Context, id int64) (*domain.ManagedService, error) {
			return &domain.ManagedService{
				ID:     id,
				NodeID: 10,
				Name:   "to-del",
			}, nil
		},
		deleteFn: func(ctx context.Context, id int64) error {
			deletedID = id
			return nil
		},
	}

	svc := NewNodeService(testLogger(), nodeRepo, &hbMockNodeStateStore{}, &hbMockInstanceRepo{}, &hbMockInstanceState{}, nodeClient).WithServiceRepo(svcRepo)

	ctx := context.Background()
	err := svc.DeleteService(ctx, "user-1", 10, 50, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if removedName != "to-del" {
		t.Errorf("expected nodeClient.RemoveService called with to-del, got %q", removedName)
	}
	if deletedID != 50 {
		t.Errorf("expected repo.Delete called with 50, got %d", deletedID)
	}
}

func TestNodeService_UpdateRole_TransitionToStorage(t *testing.T) {
	node := &domain.Node{
		ID:       10,
		OwnerID:  "user-1",
		Role:     domain.NodeRoleCompute,
		Address:  "127.0.0.1:44044",
		APIToken: "token",
		Status:   domain.NodeStatusOnline,
	}
	var updatedRole domain.NodeRole
	nodeRepo := &hbMockNodeRepo{
		getByIDFn: func(ctx context.Context, id int64) (*domain.Node, error) {
			return node, nil
		},
		updateRoleFn: func(ctx context.Context, id int64, role domain.NodeRole) error {
			updatedRole = role
			node.Role = role
			return nil
		},
	}

	deletedInstID := int64(0)
	instRepo := &hbMockInstanceRepo{
		listByNodeFn: func(ctx context.Context, nodeID int64) ([]*domain.Instance, error) {
			return []*domain.Instance{
				{ID: 99, NodeID: 10, GameID: 1, Status: domain.InstanceStatusRunning},
			}, nil
		},
		deleteFn: func(ctx context.Context, id int64) error {
			deletedInstID = id
			return nil
		},
	}
	nodeClient := &hbMockNodeClient{
		deleteInstanceFn: func(ctx context.Context, address, apiKey string, instanceID int64) error {
			return nil
		},
		stopInstanceFn: func(ctx context.Context, address, apiKey string, instanceID int64, timeoutSec uint32) error {
			return nil
		},
	}

	svc := NewNodeService(testLogger(), nodeRepo, &hbMockNodeStateStore{}, instRepo, &hbMockInstanceState{}, nodeClient)

	ctx := context.Background()
	// 1. Without confirmation -> error
	_, err := svc.UpdateRole(ctx, "user-1", 10, domain.NodeRoleStorage, domain.StorageTransitionActionUnspecified, domain.ComputeTransitionActionUnspecified)
	if err == nil {
		t.Fatal("expected error requiring confirmation when game servers exist, got nil")
	}

	// 2. With ComputeTransitionActionTerminate -> success, deletes instance, changes role
	resNode, err := svc.UpdateRole(ctx, "user-1", 10, domain.NodeRoleStorage, domain.StorageTransitionActionUnspecified, domain.ComputeTransitionActionTerminate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resNode.Role != domain.NodeRoleStorage || updatedRole != domain.NodeRoleStorage {
		t.Errorf("expected role Storage, got resNode=%v updatedRole=%v", resNode.Role, updatedRole)
	}
	if deletedInstID != 99 {
		t.Errorf("expected instance 99 to be deleted, got %d", deletedInstID)
	}
}

func TestNodeService_UpdateRole_TransitionToCompute(t *testing.T) {
	node := &domain.Node{
		ID:       10,
		OwnerID:  "user-1",
		Role:     domain.NodeRoleStorage,
		Address:  "127.0.0.1:44044",
		APIToken: "token",
		Status:   domain.NodeStatusOnline,
	}
	var updatedRole domain.NodeRole
	nodeRepo := &hbMockNodeRepo{
		getByIDFn: func(ctx context.Context, id int64) (*domain.Node, error) {
			return node, nil
		},
		updateRoleFn: func(ctx context.Context, id int64, role domain.NodeRole) error {
			updatedRole = role
			node.Role = role
			return nil
		},
	}

	stoppedSvc := ""
	nodeClient := &hbMockNodeClient{
		RemoveServiceFn: func(ctx context.Context, nodeAddress, apiKey string, name string, deleteVolume bool) error {
			return nil
		},
	}
	// override StopServiceFn by mocking behavior
	dbSvc := &domain.ManagedService{
		ID:     101,
		NodeID: 10,
		Name:   "postgres-test",
		ServiceType: domain.ServiceTypePostgres,
		Status: domain.ServiceStatusRunning,
	}

	var updatedSvcStatus domain.ServiceStatus
	svcRepo := &mockManagedServiceRepo{
		listByNodeFn: func(ctx context.Context, nodeID int64) ([]*domain.ManagedService, error) {
			return []*domain.ManagedService{dbSvc}, nil
		},
		updateFn: func(ctx context.Context, s *domain.ManagedService) error {
			updatedSvcStatus = s.Status
			return nil
		},
	}

	svc := NewNodeService(testLogger(), nodeRepo, &hbMockNodeStateStore{}, &hbMockInstanceRepo{}, &hbMockInstanceState{}, nodeClient).WithServiceRepo(svcRepo)

	ctx := context.Background()
	// 1. Without action -> error
	_, err := svc.UpdateRole(ctx, "user-1", 10, domain.NodeRoleCompute, domain.StorageTransitionActionUnspecified, domain.ComputeTransitionActionUnspecified)
	if err == nil {
		t.Fatal("expected error requiring confirmation when storage services exist, got nil")
	}

	// 2. With StorageTransitionActionStop -> marks stopped, changes role
	resNode, err := svc.UpdateRole(ctx, "user-1", 10, domain.NodeRoleCompute, domain.StorageTransitionActionStop, domain.ComputeTransitionActionUnspecified)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resNode.Role != domain.NodeRoleCompute || updatedRole != domain.NodeRoleCompute {
		t.Errorf("expected role Compute, got resNode=%v updatedRole=%v", resNode.Role, updatedRole)
	}
	if updatedSvcStatus != domain.ServiceStatusStopped {
		t.Errorf("expected service status Stopped, got %v", updatedSvcStatus)
	}
	_ = stoppedSvc
}

func TestNodeService_StartStopService(t *testing.T) {
	node := &domain.Node{
		ID:       10,
		OwnerID:  "user-1",
		Role:     domain.NodeRoleMixed,
		Address:  "127.0.0.1:44044",
		APIToken: "token",
		Status:   domain.NodeStatusOnline,
	}
	nodeRepo := &hbMockNodeRepo{
		getByIDFn: func(ctx context.Context, id int64) (*domain.Node, error) {
			return node, nil
		},
	}

	dbSvc := &domain.ManagedService{
		ID:     201,
		NodeID: 10,
		Name:   "redis-test",
		ServiceType: domain.ServiceTypeRedis,
		Status: domain.ServiceStatusRunning,
	}

	svcRepo := &mockManagedServiceRepo{
		getByIDFn: func(ctx context.Context, id int64) (*domain.ManagedService, error) {
			if id == 201 {
				return dbSvc, nil
			}
			return nil, nil
		},
		updateFn: func(ctx context.Context, s *domain.ManagedService) error {
			dbSvc.Status = s.Status
			return nil
		},
	}

	nodeClient := &hbMockNodeClient{}
	svc := NewNodeService(testLogger(), nodeRepo, &hbMockNodeStateStore{}, &hbMockInstanceRepo{}, &hbMockInstanceState{}, nodeClient).WithServiceRepo(svcRepo)

	ctx := context.Background()
	// Stop
	stopped, err := svc.StopService(ctx, "user-1", 10, 201)
	if err != nil {
		t.Fatalf("StopService failed: %v", err)
	}
	if stopped.Status != domain.ServiceStatusStopped {
		t.Errorf("expected Stopped, got %v", stopped.Status)
	}

	// Start
	started, err := svc.StartService(ctx, "user-1", 10, 201)
	if err != nil {
		t.Fatalf("StartService failed: %v", err)
	}
	if started.Status != domain.ServiceStatusRunning {
		t.Errorf("expected Running, got %v", started.Status)
	}
}


