package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
)

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

	svc := NewNodeService(nodeRepo, &hbMockNodeStateStore{}, &hbMockInstanceRepo{}, &hbMockInstanceState{}, nodeClient)

	ctx := context.Background()
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

	svc := NewNodeService(nodeRepo, &hbMockNodeStateStore{}, &hbMockInstanceRepo{}, &hbMockInstanceState{}, nodeClient)

	ctx := context.Background()
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

	svc := NewNodeService(nodeRepo, &hbMockNodeStateStore{}, &hbMockInstanceRepo{}, &hbMockInstanceState{}, &hbMockNodeClient{})

	ctx := context.Background()
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

	svc := NewNodeService(nodeRepo, &hbMockNodeStateStore{}, &hbMockInstanceRepo{}, &hbMockInstanceState{}, &hbMockNodeClient{})

	ctx := context.Background()
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

	svc := NewNodeService(nodeRepo, &hbMockNodeStateStore{}, &hbMockInstanceRepo{}, &hbMockInstanceState{}, &hbMockNodeClient{})

	ctx := context.Background()
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

	svc := NewNodeService(nodeRepo, nodeState, &hbMockInstanceRepo{}, &hbMockInstanceState{}, &hbMockNodeClient{})

	ctx := context.Background()
	result, err := svc.ListNodes(ctx, nil)

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

	svc := NewNodeService(nodeRepo, nodeState, &hbMockInstanceRepo{}, &hbMockInstanceState{}, &hbMockNodeClient{})

	ctx := context.Background()
	result, err := svc.GetNode(ctx, nodeID)

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

	svc := NewNodeService(nodeRepo, nodeState, instanceRepo, instanceState, &hbMockNodeClient{})

	ctx := context.Background()
	err := svc.DeleteNode(ctx, nodeID)

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

	svc := NewNodeService(nodeRepo, nodeState, &hbMockInstanceRepo{}, &hbMockInstanceState{}, &hbMockNodeClient{})

	ctx := context.Background()
	result, err := svc.GetNodeUsage(ctx, nodeID)

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
