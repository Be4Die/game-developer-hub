package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
)

// NodeService управляет вычислительными нодами.
type NodeService struct {
	nodeRepo      domain.NodeRepo
	nodeState     domain.NodeStateStore
	instanceRepo  domain.InstanceRepo
	instanceState domain.InstanceStateStore
	nodeClient    domain.NodeClient
}

// NewNodeService создаёт сервис управления нодами.
func NewNodeService(
	nodeRepo domain.NodeRepo,
	nodeState domain.NodeStateStore,
	instanceRepo domain.InstanceRepo,
	instanceState domain.InstanceStateStore,
	nodeClient domain.NodeClient,
) *NodeService {
	return &NodeService{
		nodeRepo:      nodeRepo,
		nodeState:     nodeState,
		instanceRepo:  instanceRepo,
		instanceState: instanceState,
		nodeClient:    nodeClient,
	}
}

// RegisterNodeParams содержит параметры подключения ноды.
// Один из Address или NodeID должен быть заполнен.
type RegisterNodeParams struct {
	// Manual: адрес ноды + токен.
	Address string
	Token   string
	Region  string

	// Authorize: ID обнаруженной ноды + токен.
	NodeID *int64
}

// RegisterNode подключает ноду:
// 1. Manual — GetNodeInfo gRPC → создание записи в PG
// 2. Authorize — проверка токена → обновление записи в PG
func (s *NodeService) RegisterNode(ctx context.Context, params RegisterNodeParams) (*domain.Node, error) {
	if params.NodeID != nil {
		return s.authorizeNode(ctx, *params.NodeID, params.Token)
	}

	return s.registerNodeManual(ctx, params.Address, params.Token, params.Region)
}

func (s *NodeService) registerNodeManual(ctx context.Context, address, token, region string) (*domain.Node, error) {
	// Проверка — не существует ли уже нода с таким адресом.
	existing, err := s.nodeRepo.GetByAddress(ctx, address)
	if err == nil && existing.Status == domain.NodeStatusOnline {
		return nil, domain.ErrAlreadyExists
	}
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("NodeService.registerNodeManual: get by address: %w", err)
	}

	// Получаем информацию о ноде через gRPC.
	info, err := s.nodeClient.GetNodeInfo(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("NodeService.registerNodeManual: GetNodeInfo: %w", err)
	}

	now := time.Now()
	tokenHash := sha256.Sum256([]byte(token))

	node := &domain.Node{
		Address:      address,
		TokenHash:    tokenHash[:],
		Region:       region,
		Status:       domain.NodeStatusOnline,
		CPUCores:     info.CPUCores,
		TotalMemory:  info.TotalMemoryBytes,
		TotalDisk:    info.TotalDiskBytes,
		AgentVersion: info.AgentVersion,
		LastPingAt:   now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.nodeRepo.Create(ctx, node); err != nil {
		return nil, fmt.Errorf("NodeService.registerNodeManual: create: %w", err)
	}

	return node, nil
}

func (s *NodeService) authorizeNode(ctx context.Context, nodeID int64, token string) (*domain.Node, error) {
	node, err := s.nodeRepo.GetByID(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("NodeService.authorizeNode: get node: %w", err)
	}

	if node.Status != domain.NodeStatusUnauthorized {
		return nil, domain.ErrAlreadyExists
	}

	// Проверка токена.
	tokenHash := sha256.Sum256([]byte(token))
	if !constantTimeEqual(tokenHash[:], node.TokenHash) {
		return nil, domain.ErrInvalidToken
	}

	// Авторизуем ноду.
	now := time.Now()
	node.Status = domain.NodeStatusOnline
	node.LastPingAt = now
	node.UpdatedAt = now

	if err := s.nodeRepo.Update(ctx, node); err != nil {
		return nil, fmt.Errorf("NodeService.authorizeNode: update: %w", err)
	}

	return node, nil
}

// ListNodes возвращает все ноды с обогащением из KV.
func (s *NodeService) ListNodes(ctx context.Context, status *domain.NodeStatus) ([]*EnrichedNode, error) {
	nodes, err := s.nodeRepo.List(ctx, status)
	if err != nil {
		return nil, fmt.Errorf("NodeService.ListNodes: %w", err)
	}

	result := make([]*EnrichedNode, 0, len(nodes))
	for _, n := range nodes {
		enriched := &EnrichedNode{Node: n}

		usage, err := s.nodeState.GetUsage(ctx, n.ID)
		if err == nil {
			enriched.Usage = usage
		}

		count, err := s.nodeState.GetActiveInstanceCount(ctx, n.ID)
		if err == nil {
			enriched.ActiveInstanceCount = &count
		}

		result = append(result, enriched)
	}

	return result, nil
}

// GetNode возвращает ноду с обогащением из KV.
func (s *NodeService) GetNode(ctx context.Context, nodeID int64) (*EnrichedNode, error) {
	node, err := s.nodeRepo.GetByID(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("NodeService.GetNode: %w", err)
	}

	enriched := &EnrichedNode{Node: node}

	usage, err := s.nodeState.GetUsage(ctx, nodeID)
	if err == nil {
		enriched.Usage = usage
	}

	count, err := s.nodeState.GetActiveInstanceCount(ctx, nodeID)
	if err == nil {
		enriched.ActiveInstanceCount = &count
	}

	return enriched, nil
}

// DeleteNode удаляет ноду:
// 1. Инстансы на ноде → crashed (PG + KV)
// 2. Удаление из KV
// 3. Удаление из PG
func (s *NodeService) DeleteNode(ctx context.Context, nodeID int64) error {
	// Переводим все инстансы ноды в crashed.
	instances, err := s.instanceRepo.ListByNode(ctx, nodeID)
	if err != nil {
		return fmt.Errorf("NodeService.DeleteNode: list instances: %w", err)
	}

	for _, inst := range instances {
		inst.Status = domain.InstanceStatusCrashed
		inst.UpdatedAt = time.Now()
		_ = s.instanceRepo.Update(ctx, inst)
		_ = s.instanceState.SetStatus(ctx, inst.ID, domain.InstanceStatusCrashed)
	}

	// Удаляем состояние из KV.
	if err := s.nodeState.Delete(ctx, nodeID); err != nil {
		return fmt.Errorf("NodeService.DeleteNode: delete KV: %w", err)
	}

	// Удаляем из PG.
	if err := s.nodeRepo.Delete(ctx, nodeID); err != nil {
		return fmt.Errorf("NodeService.DeleteNode: delete repo: %w", err)
	}

	return nil
}

// GetNodeUsage возвращает метрики ноды.
func (s *NodeService) GetNodeUsage(ctx context.Context, nodeID int64) (*NodeUsageResult, error) {
	node, err := s.nodeRepo.GetByID(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("NodeService.GetNodeUsage: get node: %w", err)
	}

	usage, err := s.nodeState.GetUsage(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("NodeService.GetNodeUsage: get usage: %w", err)
	}

	activeCount, err := s.nodeState.GetActiveInstanceCount(ctx, nodeID)
	if err != nil {
		activeCount = 0
	}

	return &NodeUsageResult{
		NodeID:              node.ID,
		Usage:               usage,
		ActiveInstanceCount: activeCount,
	}, nil
}

// EnrichedNode — нода с данными из KV.
type EnrichedNode struct {
	*domain.Node
	Usage               *domain.ResourceUsage
	ActiveInstanceCount *uint32
}

// NodeUsageResult — результат запроса метрик ноды.
type NodeUsageResult struct {
	NodeID              int64
	Usage               *domain.ResourceUsage
	ActiveInstanceCount uint32
}

// constantTimeEqual сравнивает два хеша за постоянное время.
func constantTimeEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var diff byte
	for i := range a {
		diff |= a[i] ^ b[i]
	}
	return diff == 0
}
