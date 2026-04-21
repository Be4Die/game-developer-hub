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
type RegisterNodeParams struct {
	OwnerID string
	Address string
	Token   string
	Region  string
	NodeID  *int64
}

// RegisterNode подключает ноду к оркестратору.
// Токен — это NODE_API_KEY ноды, единый для обоих методов (manual и authorize).
func (s *NodeService) RegisterNode(ctx context.Context, params RegisterNodeParams) (*domain.Node, error) {
	if params.NodeID != nil {
		return s.authorizeNode(ctx, params.OwnerID, *params.NodeID, params.Token)
	}
	return s.registerNodeManual(ctx, params.OwnerID, params.Address, params.Token, params.Region)
}

func (s *NodeService) registerNodeManual(ctx context.Context, ownerID, address, token, region string) (*domain.Node, error) {
	existing, err := s.nodeRepo.GetByAddress(ctx, address)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("NodeService.registerNodeManual: get by address: %w", err)
	}

	if err == nil {
		if existing.Status == domain.NodeStatusOnline {
			return nil, domain.ErrAlreadyExists
		}
		// Нода уже анонсирована — авторизуем её с предоставленным токеном.
		if existing.Status == domain.NodeStatusUnauthorized {
			return s.authorizeNode(ctx, ownerID, existing.ID, token)
		}
	}

	// Нода неизвестна — подключаемся к ней по gRPC.
	info, err := s.nodeClient.GetNodeInfo(ctx, address, token)
	if err != nil {
		return nil, fmt.Errorf("NodeService.registerNodeManual: GetNodeInfo: %w", err)
	}

	now := time.Now()
	tokenHash := sha256.Sum256([]byte(token))

	node := &domain.Node{
		OwnerID:      ownerID,
		Address:      address,
		TokenHash:    tokenHash[:],
		APIToken:     token,
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

// authorizeNode проверяет токен и переводит ноду в статус online.
// Токен — это NODE_API_KEY ноды, тот же самый что и при анонсе.
func (s *NodeService) authorizeNode(ctx context.Context, ownerID string, nodeID int64, token string) (*domain.Node, error) {
	node, err := s.nodeRepo.GetByID(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("NodeService.authorizeNode: get node: %w", err)
	}

	if node.OwnerID != "" && node.OwnerID != ownerID {
		return nil, fmt.Errorf("NodeService.authorizeNode: %w", domain.ErrForbidden)
	}

	if node.Status != domain.NodeStatusUnauthorized {
		return nil, domain.ErrAlreadyExists
	}

	// Проверка токена (NODE_API_KEY).
	tokenHash := sha256.Sum256([]byte(token))
	if !constantTimeEqual(tokenHash[:], node.TokenHash) {
		return nil, domain.ErrInvalidToken
	}

	// Авторизуем ноду.
	now := time.Now()
	node.OwnerID = ownerID
	node.Status = domain.NodeStatusOnline
	node.LastPingAt = now
	node.UpdatedAt = now

	if err := s.nodeRepo.Update(ctx, node); err != nil {
		return nil, fmt.Errorf("NodeService.authorizeNode: update: %w", err)
	}

	return node, nil
}

// ListNodes возвращает ноды пользователя с обогащением из KV.
func (s *NodeService) ListNodes(ctx context.Context, ownerID string, status *domain.NodeStatus) ([]*EnrichedNode, error) {
	nodes, err := s.nodeRepo.List(ctx, status)
	if err != nil {
		return nil, fmt.Errorf("NodeService.ListNodes: %w", err)
	}

	result := make([]*EnrichedNode, 0, len(nodes))
	for _, n := range nodes {
		if ownerID != "" && n.OwnerID != "" && n.OwnerID != ownerID {
			continue
		}

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

// GetNode возвращает ноду с обогащением из KV. Проверяет владение.
func (s *NodeService) GetNode(ctx context.Context, ownerID string, nodeID int64) (*EnrichedNode, error) {
	node, err := s.nodeRepo.GetByID(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("NodeService.GetNode: %w", err)
	}

	if ownerID != "" && node.OwnerID != "" && node.OwnerID != ownerID {
		return nil, domain.ErrForbidden
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

// DeleteNode удаляет ноду из оркестратора. Проверяет владение.
func (s *NodeService) DeleteNode(ctx context.Context, ownerID string, nodeID int64) error {
	node, err := s.nodeRepo.GetByID(ctx, nodeID)
	if err != nil {
		return fmt.Errorf("NodeService.DeleteNode: get node: %w", err)
	}

	if ownerID != "" && node.OwnerID != "" && node.OwnerID != ownerID {
		return domain.ErrForbidden
	}

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

	if err := s.nodeState.Delete(ctx, nodeID); err != nil {
		return fmt.Errorf("NodeService.DeleteNode: delete KV: %w", err)
	}

	if err := s.nodeRepo.Delete(ctx, nodeID); err != nil {
		return fmt.Errorf("NodeService.DeleteNode: delete repo: %w", err)
	}

	return nil
}

// GetNodeUsage возвращает метрики ноды. Проверяет владение.
func (s *NodeService) GetNodeUsage(ctx context.Context, ownerID string, nodeID int64) (*NodeUsageResult, error) {
	node, err := s.nodeRepo.GetByID(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("NodeService.GetNodeUsage: get node: %w", err)
	}

	if ownerID != "" && node.OwnerID != "" && node.OwnerID != ownerID {
		return nil, domain.ErrForbidden
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

// AnnounceNodeParams содержит параметры анонсирования ноды.
type AnnounceNodeParams struct {
	Address          string
	Region           string
	AgentVersion     string
	CPUCores         uint32
	TotalMemoryBytes uint64
	TotalDiskBytes   uint64
	APIKey           string // NODE_API_KEY ноды
}

// AnnounceNodeResult содержит результат анонсирования ноды.
type AnnounceNodeResult struct {
	NodeID int64
}

// AnnounceNode обрабатывает анонсирование ноды от самой ноды.
// Нода передаёт свой NODE_API_KEY как api_key — этот ключ становится
// токеном авторизации. Пользователь вводит тот же NODE_API_KEY для подключения.
func (s *NodeService) AnnounceNode(ctx context.Context, params AnnounceNodeParams) (*AnnounceNodeResult, error) {
	existing, err := s.nodeRepo.GetByAddress(ctx, params.Address)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("NodeService.AnnounceNode: get by address: %w", err)
	}

	if err == nil {
		if existing.Status == domain.NodeStatusOnline {
			// Нода уже авторизована — обновляем системную информацию,
			// но не меняем токен и статус.
			existing.CPUCores = params.CPUCores
			existing.TotalMemory = params.TotalMemoryBytes
			existing.TotalDisk = params.TotalDiskBytes
			existing.AgentVersion = params.AgentVersion
			existing.LastPingAt = time.Now()
			existing.UpdatedAt = time.Now()
			if params.Region != "" {
				existing.Region = params.Region
			}
			if err := s.nodeRepo.Update(ctx, existing); err != nil {
				return nil, fmt.Errorf("NodeService.AnnounceNode: update online: %w", err)
			}
			return &AnnounceNodeResult{NodeID: existing.ID}, nil
		}
		// Обновляем существующую неавторизованную ноду.
		return s.updateAnnouncedNode(ctx, existing, params)
	}

	return s.createAnnouncedNode(ctx, params)
}

func (s *NodeService) createAnnouncedNode(ctx context.Context, params AnnounceNodeParams) (*AnnounceNodeResult, error) {
	apiKey := params.APIKey
	tokenHash := sha256.Sum256([]byte(apiKey))
	now := time.Now()

	node := &domain.Node{
		OwnerID:      "",
		Address:      params.Address,
		TokenHash:    tokenHash[:],
		APIToken:     apiKey,
		Region:       params.Region,
		Status:       domain.NodeStatusUnauthorized,
		CPUCores:     params.CPUCores,
		TotalMemory:  params.TotalMemoryBytes,
		TotalDisk:    params.TotalDiskBytes,
		AgentVersion: params.AgentVersion,
		LastPingAt:   now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.nodeRepo.Create(ctx, node); err != nil {
		return nil, fmt.Errorf("NodeService.createAnnouncedNode: create: %w", err)
	}

	return &AnnounceNodeResult{
		NodeID: node.ID,
	}, nil
}

func (s *NodeService) updateAnnouncedNode(ctx context.Context, existing *domain.Node, params AnnounceNodeParams) (*AnnounceNodeResult, error) {
	apiKey := params.APIKey
	tokenHash := sha256.Sum256([]byte(apiKey))
	now := time.Now()

	existing.TokenHash = tokenHash[:]
	existing.APIToken = apiKey
	existing.Region = params.Region
	existing.CPUCores = params.CPUCores
	existing.TotalMemory = params.TotalMemoryBytes
	existing.TotalDisk = params.TotalDiskBytes
	existing.AgentVersion = params.AgentVersion
	existing.LastPingAt = now
	existing.UpdatedAt = now

	if err := s.nodeRepo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("NodeService.updateAnnouncedNode: update: %w", err)
	}

	return &AnnounceNodeResult{
		NodeID: existing.ID,
	}, nil
}
