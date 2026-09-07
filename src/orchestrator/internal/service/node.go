package service

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
)

// NodeService управляет вычислительными нодами.
type NodeService struct {
	log           *slog.Logger
	nodeRepo      domain.NodeRepo
	nodeState     domain.NodeStateStore
	instanceRepo  domain.InstanceRepo
	instanceState domain.InstanceStateStore
	nodeClient    domain.NodeClient
	serviceRepo   domain.ManagedServiceRepo
}

// NewNodeService создаёт сервис управления нодами.
func NewNodeService(
	log *slog.Logger,
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
		log:           log,
	}
}

// WithServiceRepo задает репозиторий управляемых сервисов хранения.
func (s *NodeService) WithServiceRepo(r domain.ManagedServiceRepo) *NodeService {
	s.serviceRepo = r
	return s
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

	usage, _ := s.nodeState.GetUsage(ctx, nodeID)
	activeCount, _ := s.nodeState.GetActiveInstanceCount(ctx, nodeID)

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
	Address            string
	Region             string
	AgentVersion       string
	CPUCores           uint32
	TotalMemoryBytes   uint64
	TotalDiskBytes     uint64
	APIKey             string // NODE_API_KEY ноды
	ActiveContainerIDs []string
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
		// Нода с таким адресом уже существует — обновляем данные.
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

	if existing.Status == domain.NodeStatusOnline {
		// Если нода уже онлайн и токен совпадает — нода перезапустилась, обновляем характеристики
		if len(existing.TokenHash) > 0 && subtle.ConstantTimeCompare(existing.TokenHash, tokenHash[:]) == 1 {
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
		return nil, domain.ErrAlreadyExists
	}

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

// ListNodeInstances возвращает список инстансов на указанной ноде.
func (s *NodeService) ListNodeInstances(ctx context.Context, ownerID string, nodeID int64) ([]*EnrichedInstance, error) {
	node, err := s.nodeRepo.GetByID(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("NodeService.ListNodeInstances: get node: %w", err)
	}
	if ownerID != "" && node.OwnerID != "" && node.OwnerID != ownerID {
		return nil, domain.ErrForbidden
	}

	instances, err := s.instanceRepo.ListByNode(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("NodeService.ListNodeInstances: list instances: %w", err)
	}

	result := make([]*EnrichedInstance, 0, len(instances))
	for _, inst := range instances {
		enriched := &EnrichedInstance{Instance: inst, Status: inst.Status}

		st, err := s.instanceState.GetStatus(ctx, inst.ID)
		if err == nil {
			enriched.Status = st
		}

		pc, err := s.instanceState.GetPlayerCount(ctx, inst.ID)
		if err == nil {
			enriched.PlayerCount = &pc
		}

		result = append(result, enriched)
	}

	return result, nil
}

// SyncInstances синхронизирует статусы инстансов с активными контейнерами на ноде.
// Примечание: в доменной модели Instance отсутствует поле ContainerID, поэтому
// сопоставление выполняется по строковому представлению Instance.ID.
func (s *NodeService) SyncInstances(ctx context.Context, nodeID int64, activeContainerIDs []string) error {
	const op = "NodeService.SyncInstances"

	instances, err := s.instanceRepo.ListByNode(ctx, nodeID)
	if err != nil {
		return fmt.Errorf("%s: list instances: %w", op, err)
	}

	activeSet := make(map[string]struct{}, len(activeContainerIDs))
	for _, cid := range activeContainerIDs {
		activeSet[cid] = struct{}{}
	}

	now := time.Now()
	for _, inst := range instances {
		instIDStr := strconv.FormatInt(inst.ID, 10)
		if inst.Status == domain.InstanceStatusRunning || inst.Status == domain.InstanceStatusStarting {
			if _, ok := activeSet[instIDStr]; !ok {
				inst.Status = domain.InstanceStatusStopped
				inst.UpdatedAt = now
				if err := s.instanceRepo.Update(ctx, inst); err != nil {
					s.log.Warn("failed to update instance status during sync",
						slog.Int64("instance_id", inst.ID),
						slog.String("error", err.Error()),
					)
					continue
				}
				_ = s.instanceState.SetStatus(ctx, inst.ID, domain.InstanceStatusStopped)
				s.log.Info("instance status updated during sync",
					slog.Int64("instance_id", inst.ID),
					slog.String("new_status", "stopped"),
				)
			}
		}
	}

	return nil
}

// UpdateRole изменяет роль вычислительной ноды (Mixed, Compute, Storage).
func (s *NodeService) UpdateRole(ctx context.Context, ownerID string, nodeID int64, role domain.NodeRole) (*domain.Node, error) {
	node, err := s.nodeRepo.GetByID(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("NodeService.UpdateRole: get node: %w", err)
	}
	if ownerID != "" && node.OwnerID != ownerID {
		return nil, domain.ErrForbidden
	}

	if err := s.nodeRepo.UpdateRole(ctx, nodeID, role); err != nil {
		return nil, fmt.Errorf("NodeService.UpdateRole: update role: %w", err)
	}

	node.Role = role
	node.UpdatedAt = time.Now()
	s.log.Info("node role updated",
		slog.Int64("node_id", nodeID),
		slog.String("role", role.String()),
	)
	return node, nil
}

// CreateServiceParams задает параметры для создания сервиса данных на ноде.
type CreateServiceParams struct {
	ServiceType    domain.ServiceType
	Name           string
	AllowedGameIDs []int64
	Password       string
	DBName         string
	Port           uint32
}

// CreateService разворачивает управляемый сервис хранения (Postgres, Redis, MySQL, MinIO) на ноде.
func (s *NodeService) CreateService(ctx context.Context, ownerID string, nodeID int64, params CreateServiceParams) (*domain.ManagedService, error) {
	if s.serviceRepo == nil {
		return nil, fmt.Errorf("NodeService.CreateService: service repository is not configured")
	}

	node, err := s.nodeRepo.GetByID(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("NodeService.CreateService: get node: %w", err)
	}
	if ownerID != "" && node.OwnerID != ownerID {
		return nil, domain.ErrForbidden
	}
	if node.Status != domain.NodeStatusOnline {
		return nil, fmt.Errorf("NodeService.CreateService: node %d is not online", nodeID)
	}
	if node.Role == domain.NodeRoleCompute {
		return nil, fmt.Errorf("NodeService.CreateService: node %d is in compute-only role", nodeID)
	}

	envVars := make(map[string]string)
	credentials := make(map[string]string)

	switch params.ServiceType {
	case domain.ServiceTypePostgres:
		pass := params.Password
		if pass == "" {
			pass = "gdh_secret_" + strconv.FormatInt(time.Now().UnixNano()%100000, 10)
		}
		db := params.DBName
		if db == "" {
			db = "game_db"
		}
		envVars["POSTGRES_USER"] = "postgres"
		envVars["POSTGRES_PASSWORD"] = pass
		envVars["POSTGRES_DB"] = db
		credentials["username"] = "postgres"
		credentials["password"] = pass
		credentials["database"] = db

	case domain.ServiceTypeRedis:
		if params.Password != "" {
			envVars["REDIS_PASSWORD"] = params.Password
			credentials["password"] = params.Password
		}

	case domain.ServiceTypeMySQL:
		pass := params.Password
		if pass == "" {
			pass = "gdh_root_secret"
		}
		db := params.DBName
		if db == "" {
			db = "game_db"
		}
		envVars["MYSQL_ROOT_PASSWORD"] = pass
		envVars["MYSQL_DATABASE"] = db
		credentials["username"] = "root"
		credentials["password"] = pass
		credentials["database"] = db

	case domain.ServiceTypeMinIO:
		user := "minioadmin"
		pass := params.Password
		if pass == "" {
			pass = "minioadmin"
		}
		envVars["MINIO_ROOT_USER"] = user
		envVars["MINIO_ROOT_PASSWORD"] = pass
		credentials["access_key"] = user
		credentials["secret_key"] = pass

	case domain.ServiceTypeVolume:
		// Чистый персистентный том под данные (SQLite, RocksDB, кастомные сейвы)

	case domain.ServiceTypeAdminer:
		// Web-панель управления СУБД (PostgreSQL / MySQL)

	case domain.ServiceTypePGAdmin:
		credentials["username"] = "admin@gdh.local"
		credentials["password"] = "admin"
	}

	req := domain.DeployServiceRequest{
		ServiceType: params.ServiceType,
		Name:        params.Name,
		Port:        params.Port,
		EnvVars:     envVars,
		VolumeName:  params.Name,
	}

	deployResult, err := s.nodeClient.DeployService(ctx, node.Address, node.APIToken, req)
	if err != nil {
		return nil, fmt.Errorf("NodeService.CreateService: node deploy: %w", err)
	}

	now := time.Now()
	serviceRecord := &domain.ManagedService{
		NodeID:         nodeID,
		OwnerID:        ownerID,
		AllowedGameIDs: params.AllowedGameIDs,
		ServiceType:    params.ServiceType,
		Name:           params.Name,
		ContainerID:    deployResult.ContainerID,
		HostPort:       deployResult.HostPort,
		ConnectionURI:  deployResult.ConnectionURI,
		Credentials:    credentials,
		Status:         domain.ServiceStatusRunning,
		VolumePath:     deployResult.VolumePath,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.serviceRepo.Create(ctx, serviceRecord); err != nil {
		// Rollback on node
		_ = s.nodeClient.RemoveService(ctx, node.Address, node.APIToken, params.Name, true)
		return nil, fmt.Errorf("NodeService.CreateService: save record: %w", err)
	}

	s.log.Info("managed service created",
		slog.Int64("service_id", serviceRecord.ID),
		slog.String("name", params.Name),
		slog.Int64("node_id", nodeID),
	)
	return serviceRecord, nil
}

// ListServices возвращает сервисы хранения на ноде с опциональной фильтрацией по игре.
func (s *NodeService) ListServices(ctx context.Context, ownerID string, nodeID int64, gameID *int64) ([]*domain.ManagedService, error) {
	if s.serviceRepo == nil {
		return nil, fmt.Errorf("NodeService.ListServices: service repository is not configured")
	}

	node, err := s.nodeRepo.GetByID(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("NodeService.ListServices: get node: %w", err)
	}
	if ownerID != "" && node.OwnerID != ownerID {
		return nil, domain.ErrForbidden
	}

	services, err := s.serviceRepo.ListByNode(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("NodeService.ListServices: %w", err)
	}

	if gameID == nil {
		return services, nil
	}

	filtered := make([]*domain.ManagedService, 0, len(services))
	for _, svc := range services {
		if len(svc.AllowedGameIDs) == 0 {
			filtered = append(filtered, svc)
			continue
		}
		for _, gid := range svc.AllowedGameIDs {
			if gid == *gameID {
				filtered = append(filtered, svc)
				break
			}
		}
	}
	return filtered, nil
}

// DeleteService удаляет управляемый сервис с ноды и из базы данных.
func (s *NodeService) DeleteService(ctx context.Context, ownerID string, nodeID, serviceID int64, deleteVolume bool) error {
	if s.serviceRepo == nil {
		return fmt.Errorf("NodeService.DeleteService: service repository is not configured")
	}

	node, err := s.nodeRepo.GetByID(ctx, nodeID)
	if err != nil {
		return fmt.Errorf("NodeService.DeleteService: get node: %w", err)
	}
	if ownerID != "" && node.OwnerID != ownerID {
		return domain.ErrForbidden
	}

	svc, err := s.serviceRepo.GetByID(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("NodeService.DeleteService: get service: %w", err)
	}
	if svc.NodeID != nodeID {
		return fmt.Errorf("NodeService.DeleteService: service does not belong to node %d", nodeID)
	}

	// Удаляем контейнер и том на ноде
	if node.Status == domain.NodeStatusOnline {
		_ = s.nodeClient.RemoveService(ctx, node.Address, node.APIToken, svc.Name, deleteVolume)
	}

	if err := s.serviceRepo.Delete(ctx, serviceID); err != nil {
		return fmt.Errorf("NodeService.DeleteService: delete from DB: %w", err)
	}

	s.log.Info("managed service deleted",
		slog.Int64("service_id", serviceID),
		slog.String("name", svc.Name),
		slog.Bool("delete_volume", deleteVolume),
	)
	return nil
}

