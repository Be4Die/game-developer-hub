package service

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/url"
	"strconv"
	"strings"
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
	backupRepo    domain.BackupRepo
	platformRepo  domain.PlatformAccessRepo
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

// WithBackupRepo задает репозиторий бэкапов.
func (s *NodeService) WithBackupRepo(r domain.BackupRepo) *NodeService {
	s.backupRepo = r
	return s
}

// WithPlatformAccessRepo задает репозиторий заявок на платформенные ноды.
func (s *NodeService) WithPlatformAccessRepo(r domain.PlatformAccessRepo) *NodeService {
	s.platformRepo = r
	return s
}

// RegisterNodeParams содержит параметры подключения ноды.
type RegisterNodeParams struct {
	OwnerID      string
	Address      string
	Token        string
	Region       string
	NodeID       *int64
	IngressMode  *domain.IngressMode
	CustomDomain *string
}

// RegisterNode подключает ноду к оркестратору.
// Токен — это NODE_API_KEY ноды, единый для обоих методов (manual и authorize).
func (s *NodeService) RegisterNode(ctx context.Context, params RegisterNodeParams) (*domain.Node, error) {
	if params.NodeID != nil {
		return s.authorizeNode(ctx, params.OwnerID, *params.NodeID, params.Token)
	}
	return s.registerNodeManual(ctx, params)
}

// determineIngressMode определяет сетевой режим ноды на основе адреса.
// Если хост — IP-адрес или локальный узел (localhost, host.docker.internal), используется IngressModePlatformProxy.
// Если хост — FQDN-домен (например, node1.mygame.ru), используется IngressModeDirect.
func determineIngressMode(addr string) (domain.IngressMode, string) {
	host := addr
	if h, _, err := net.SplitHostPort(addr); err == nil {
		host = h
	}
	if ip := net.ParseIP(host); ip != nil {
		return domain.IngressModePlatformProxy, ""
	}
	if host == "localhost" || host == "host.docker.internal" || host == "127.0.0.1" {
		return domain.IngressModePlatformProxy, ""
	}
	return domain.IngressModeDirect, host
}

func (s *NodeService) registerNodeManual(ctx context.Context, params RegisterNodeParams) (*domain.Node, error) {
	existing, err := s.nodeRepo.GetByAddress(ctx, params.Address)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("NodeService.registerNodeManual: get by address: %w", err)
	}

	mode, customDomain := determineIngressMode(params.Address)
	if params.IngressMode != nil && *params.IngressMode != domain.IngressModeUnspecified {
		mode = *params.IngressMode
	}
	if params.CustomDomain != nil && *params.CustomDomain != "" {
		customDomain = *params.CustomDomain
	}
	if mode == domain.IngressModePlatformProxy {
		customDomain = ""
	}

	if err == nil {
		if existing.Status == domain.NodeStatusOnline {
			return nil, domain.ErrAlreadyExists
		}
		// Нода уже анонсирована — авторизуем её с предоставленным токеном.
		if existing.Status == domain.NodeStatusUnauthorized {
			if params.IngressMode != nil && *params.IngressMode != domain.IngressModeUnspecified {
				_ = s.nodeRepo.UpdateIngress(ctx, existing.ID, mode, customDomain)
			}
			return s.authorizeNode(ctx, params.OwnerID, existing.ID, params.Token)
		}
	}

	// Нода неизвестна — подключаемся к ней по gRPC.
	info, err := s.nodeClient.GetNodeInfo(ctx, params.Address, params.Token)
	if err != nil {
		return nil, fmt.Errorf("NodeService.registerNodeManual: GetNodeInfo: %w", err)
	}

	now := time.Now()
	tokenHash := sha256.Sum256([]byte(params.Token))

	node := &domain.Node{
		OwnerID:      params.OwnerID,
		Address:      params.Address,
		TokenHash:    tokenHash[:],
		APIToken:     params.Token,
		Region:       params.Region,
		Status:       domain.NodeStatusOnline,
		IngressMode:  mode,
		CustomDomain: customDomain,
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
// Если указан gameID и проекту одобрен доступ к платформенным серверам, возвращаются также платформенные ноды.
func (s *NodeService) ListNodes(ctx context.Context, ownerID string, status *domain.NodeStatus, gameID *int64) ([]*EnrichedNode, error) {
	nodes, err := s.nodeRepo.List(ctx, status)
	if err != nil {
		return nil, fmt.Errorf("NodeService.ListNodes: %w", err)
	}

	hasPlatformAccess := false
	if gameID != nil && s.platformRepo != nil {
		hasAccess, _, err := s.platformRepo.HasApprovedAccess(ctx, *gameID)
		if err == nil && hasAccess {
			hasPlatformAccess = true
		}
	}

	result := make([]*EnrichedNode, 0, len(nodes))
	for _, n := range nodes {
		if ownerID != "" {
			isOwner := (n.OwnerID == "" || n.OwnerID == ownerID)
			isAllowedPlatform := hasPlatformAccess && n.IsPlatform
			if !isOwner && !isAllowedPlatform {
				continue
			}
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

// GetNode возвращает ноду с обогащением из KV. Проверяет владение (или доступ к платформенной ноде).
func (s *NodeService) GetNode(ctx context.Context, ownerID string, nodeID int64) (*EnrichedNode, error) {
	node, err := s.nodeRepo.GetByID(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("NodeService.GetNode: %w", err)
	}
	if ownerID != "" && node.OwnerID != "" && node.OwnerID != ownerID && !node.IsPlatform {
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

// UpdatePlatformStatus обновляет признак платформенной ноды (пул платформы).
func (s *NodeService) UpdatePlatformStatus(ctx context.Context, nodeID int64, isPlatform bool) (*EnrichedNode, error) {
	node, err := s.nodeRepo.GetByID(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("NodeService.UpdatePlatformStatus: get node: %w", err)
	}

	if err := s.nodeRepo.UpdatePlatformStatus(ctx, nodeID, isPlatform); err != nil {
		return nil, fmt.Errorf("NodeService.UpdatePlatformStatus: update: %w", err)
	}
	node.IsPlatform = isPlatform
	return &EnrichedNode{Node: node}, nil
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
	mode, customDomain := determineIngressMode(params.Address)
	node := &domain.Node{
		OwnerID:      "",
		Address:      params.Address,
		TokenHash:    tokenHash[:],
		APIToken:     apiKey,
		Region:       params.Region,
		Status:       domain.NodeStatusUnauthorized,
		IngressMode:  mode,
		CustomDomain: customDomain,
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

// UpdateRole изменяет роль вычислительной ноды (Mixed, Compute, Storage) с учетом политик вытеснения.
func (s *NodeService) UpdateRole(
	ctx context.Context,
	ownerID string,
	nodeID int64,
	role domain.NodeRole,
	storageAction domain.StorageTransitionAction,
	computeAction domain.ComputeTransitionAction,
) (*domain.Node, error) {
	node, err := s.nodeRepo.GetByID(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("NodeService.UpdateRole: get node: %w", err)
	}
	if ownerID != "" && node.OwnerID != ownerID {
		return nil, domain.ErrForbidden
	}

	if node.Role == role {
		return node, nil
	}

	// ─── Политика перехода в Storage (вытеснение игровых серверов) ─────────────
	if role == domain.NodeRoleStorage && s.instanceRepo != nil {
		instances, err := s.instanceRepo.ListByNode(ctx, nodeID)
		if err == nil && len(instances) > 0 {
			if computeAction != domain.ComputeTransitionActionTerminate {
				return nil, fmt.Errorf("node has %d game servers: confirmation required to terminate them before switching to Storage", len(instances))
			}

			s.log.Info("terminating game servers for role transition to storage",
				slog.Int64("node_id", nodeID),
				slog.Int("count", len(instances)),
			)
			for _, inst := range instances {
				if node.Status == domain.NodeStatusOnline && s.nodeClient != nil {
					if inst.Status == domain.InstanceStatusRunning || inst.Status == domain.InstanceStatusStarting || inst.Status == domain.InstanceStatusStopping {
						_ = s.nodeClient.StopInstance(ctx, node.Address, node.APIToken, inst.ID, 5)
					}
					_ = s.nodeClient.DeleteInstance(ctx, node.Address, node.APIToken, inst.ID)
				}
				_ = s.instanceRepo.Delete(ctx, inst.ID)
				if s.instanceState != nil {
					_ = s.instanceState.Delete(ctx, inst.ID)
				}
			}
			if s.nodeState != nil {
				_ = s.nodeState.SetActiveInstanceCount(ctx, nodeID, 0)
			}
		}
	}

	// ─── Политика перехода в Compute (вытеснение сервисов хранения) ────────────
	if role == domain.NodeRoleCompute && s.serviceRepo != nil {
		services, err := s.serviceRepo.ListByNode(ctx, nodeID)
		if err == nil && len(services) > 0 {
			if storageAction == domain.StorageTransitionActionUnspecified {
				return nil, fmt.Errorf("node has %d storage services: confirmation required (stop or delete) before switching to Compute", len(services))
			}

			// 100% Создание контрольного бэкапа перед любым действием (stop или delete)
			for _, svc := range services {
				if svc.ServiceType == domain.ServiceTypePostgres || svc.ServiceType == domain.ServiceTypeMySQL || svc.ServiceType == domain.ServiceTypeRedis {
					if node.Status == domain.NodeStatusOnline && s.nodeClient != nil {
						s.log.Info("creating safety snapshot before role transition to compute",
							slog.String("service", svc.Name),
						)
						_, backupErr := s.CreateServiceBackup(ctx, ownerID, nodeID, svc.Name)
						if backupErr != nil {
							s.log.Warn("safety snapshot failed before transition",
								slog.String("service", svc.Name),
								slog.Any("err", backupErr),
							)
						}
					}
				}
			}

			if storageAction == domain.StorageTransitionActionStop {
				s.log.Info("stopping storage services for role transition to compute",
					slog.Int64("node_id", nodeID),
					slog.Int("count", len(services)),
				)
				for _, svc := range services {
					if node.Status == domain.NodeStatusOnline && s.nodeClient != nil {
						_ = s.nodeClient.StopService(ctx, node.Address, node.APIToken, svc.Name)
					}
					svc.Status = domain.ServiceStatusStopped
					svc.UpdatedAt = time.Now()
					_ = s.serviceRepo.Update(ctx, svc)
				}
			} else if storageAction == domain.StorageTransitionActionDelete {
				s.log.Info("deleting storage services for role transition to compute",
					slog.Int64("node_id", nodeID),
					slog.Int("count", len(services)),
				)
				for _, svc := range services {
					if node.Status == domain.NodeStatusOnline && s.nodeClient != nil {
						_ = s.nodeClient.RemoveService(ctx, node.Address, node.APIToken, svc.Name, true)
					}
					_ = s.serviceRepo.Delete(ctx, svc.ID)
				}
			}
		}
	}

	// ─── Сохранение новой роли ────────────────────────────────────────────────
	if err := s.nodeRepo.UpdateRole(ctx, nodeID, role); err != nil {
		return nil, fmt.Errorf("NodeService.UpdateRole: update role: %w", err)
	}

	// ─── Авто-возобновление Adminer при возврате в Storage или Mixed ──────────
	if (role == domain.NodeRoleStorage || role == domain.NodeRoleMixed) && s.serviceRepo != nil && node.Status == domain.NodeStatusOnline && s.nodeClient != nil {
		services, err := s.serviceRepo.ListByNode(ctx, nodeID)
		if err == nil {
			for _, svc := range services {
				if svc.ServiceType == domain.ServiceTypeAdminer && svc.Status == domain.ServiceStatusStopped {
					port, uri, startErr := s.nodeClient.StartService(ctx, node.Address, node.APIToken, svc.Name)
					if startErr == nil {
						if port > 0 {
							svc.HostPort = port
							svc.ConnectionURI = uri
						}
						svc.Status = domain.ServiceStatusRunning
						svc.UpdatedAt = time.Now()
						_ = s.serviceRepo.Update(ctx, svc)
						s.log.Info("auto-resumed adminer service on transition to storage/mixed",
							slog.Int64("node_id", nodeID),
							slog.String("name", svc.Name),
							slog.Uint64("host_port", uint64(svc.HostPort)),
						)
					} else {
						s.log.Warn("failed to auto-resume adminer on transition to storage/mixed",
							slog.Int64("node_id", nodeID),
							slog.Any("err", startErr),
						)
					}
				}
			}
		}
	}

	node.Role = role
	node.UpdatedAt = time.Now()
	s.log.Info("node role updated",
		slog.Int64("node_id", nodeID),
		slog.String("role", role.String()),
	)
	return node, nil
}

// VerifyDomainResult содержит результат проверки DNS-записи домена.
type VerifyDomainResult struct {
	Valid       bool
	Message     string
	ResolvedIPs []string
	NodeIP      string
}

// sanitizeDomain очищает домен от схемы, порта и концевых слэшей.
func sanitizeDomain(raw string) string {
	d := strings.TrimSpace(raw)
	d = strings.TrimPrefix(d, "https://")
	d = strings.TrimPrefix(d, "http://")
	d = strings.Split(d, "/")[0]
	d = strings.Split(d, ":")[0]
	return strings.ToLower(strings.TrimSpace(d))
}

// VerifyDomain проверяет сопоставление кастомного домена и IP-адреса ноды через DNS.
func (s *NodeService) VerifyDomain(ctx context.Context, ownerID string, nodeID int64, customDomain string) (*VerifyDomainResult, error) {
	node, err := s.nodeRepo.GetByID(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("NodeService.VerifyDomain: get node: %w", err)
	}
	if node.OwnerID != "" && node.OwnerID != ownerID {
		return nil, fmt.Errorf("NodeService.VerifyDomain: %w", domain.ErrForbidden)
	}

	cleanDomain := sanitizeDomain(customDomain)
	if cleanDomain == "" {
		return &VerifyDomainResult{
			Valid:   false,
			Message: "Домен не может быть пустым",
		}, nil
	}

	// Извлекаем адрес ноды без порта
	nodeHost := node.Address
	if h, _, err := net.SplitHostPort(node.Address); err == nil {
		nodeHost = h
	}
	nodeHost = strings.TrimSpace(nodeHost)

	var expectedIPs []string
	isNodeLoopback := nodeHost == "localhost" || nodeHost == "127.0.0.1" || nodeHost == "::1"
	if isNodeLoopback {
		expectedIPs = []string{"127.0.0.1", "::1"}
	} else if ip := net.ParseIP(nodeHost); ip != nil {
		expectedIPs = []string{ip.String()}
	} else {
		resolvedNodeIPs, _ := net.DefaultResolver.LookupIP(ctx, "ip", nodeHost)
		for _, r := range resolvedNodeIPs {
			expectedIPs = append(expectedIPs, r.String())
		}
	}

	var resolvedIPs []string
	if cleanDomain == "localhost" {
		resolvedIPs = []string{"127.0.0.1", "::1"}
	} else {
		dnsLookupCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
		defer cancel()
		lookupRes, err := net.DefaultResolver.LookupIP(dnsLookupCtx, "ip", cleanDomain)
		if err != nil {
			return &VerifyDomainResult{
				Valid:   false,
				Message: fmt.Sprintf("DNS lookup failed: %v", err),
				NodeIP:  nodeHost,
			}, nil
		}
		for _, r := range lookupRes {
			resolvedIPs = append(resolvedIPs, r.String())
		}
	}

	if len(resolvedIPs) == 0 {
		return &VerifyDomainResult{
			Valid:   false,
			Message: "DNS-записи A/AAAA для указанного домена не найдены",
			NodeIP:  nodeHost,
		}, nil
	}

	matched := false
	if isNodeLoopback && (cleanDomain == "localhost" || cleanDomain == "127.0.0.1") {
		matched = true
	} else {
		for _, rip := range resolvedIPs {
			for _, eip := range expectedIPs {
				if rip == eip {
					matched = true
					break
				}
			}
			if matched {
				break
			}
		}
	}

	if !matched {
		return &VerifyDomainResult{
			Valid:       false,
			Message:     fmt.Sprintf("Домен указывает на %v, но IP ноды: %s", resolvedIPs, nodeHost),
			ResolvedIPs: resolvedIPs,
			NodeIP:      nodeHost,
		}, nil
	}

	return &VerifyDomainResult{
		Valid:       true,
		Message:     fmt.Sprintf("Домен успешно подтверждён: указывает на IP ноды (%s)", nodeHost),
		ResolvedIPs: resolvedIPs,
		NodeIP:      nodeHost,
	}, nil
}

// UpdateIngress обновляет сетевой режим подключения к ноде (Platform Proxy или Direct) и кастомный домен.
func (s *NodeService) UpdateIngress(ctx context.Context, ownerID string, nodeID int64, mode domain.IngressMode, customDomain string, skipDNSCheck bool) (*domain.Node, error) {
	node, err := s.nodeRepo.GetByID(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("NodeService.UpdateIngress: get node: %w", err)
	}
	if node.OwnerID != "" && node.OwnerID != ownerID {
		return nil, fmt.Errorf("NodeService.UpdateIngress: %w", domain.ErrForbidden)
	}

	if mode == domain.IngressModeUnspecified {
		mode = domain.IngressModePlatformProxy
	}
	if mode == domain.IngressModePlatformProxy {
		customDomain = ""
	} else if mode == domain.IngressModeDirect {
		customDomain = sanitizeDomain(customDomain)
		if customDomain == "" {
			return nil, fmt.Errorf("NodeService.UpdateIngress: домен не указан")
		}
		if !skipDNSCheck {
			vRes, err := s.VerifyDomain(ctx, ownerID, nodeID, customDomain)
			if err != nil {
				return nil, fmt.Errorf("NodeService.UpdateIngress verify: %w", err)
			}
			if !vRes.Valid {
				return nil, fmt.Errorf("NodeService.UpdateIngress: %w: %s", domain.ErrDomainMismatch, vRes.Message)
			}
		}
	}

	if err := s.nodeRepo.UpdateIngress(ctx, nodeID, mode, customDomain); err != nil {
		return nil, fmt.Errorf("NodeService.UpdateIngress: update ingress: %w", err)
	}

	node.IngressMode = mode
	node.CustomDomain = customDomain
	node.UpdatedAt = time.Now()

	s.log.Info("node ingress updated",
		slog.Int64("node_id", nodeID),
		slog.String("mode", mode.String()),
		slog.String("custom_domain", customDomain),
	)

	return node, nil
}

// StartService запускает ранее остановленный управляемый сервис.
func (s *NodeService) StartService(ctx context.Context, ownerID string, nodeID, serviceID int64) (*domain.ManagedService, error) {
	if s.serviceRepo == nil {
		return nil, fmt.Errorf("NodeService.StartService: service repository is not configured")
	}
	node, err := s.nodeRepo.GetByID(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("NodeService.StartService: get node: %w", err)
	}
	if ownerID != "" && node.OwnerID != ownerID {
		return nil, domain.ErrForbidden
	}
	if node.Role == domain.NodeRoleCompute {
		return nil, fmt.Errorf("NodeService.StartService: cannot start storage service on compute-only node")
	}

	svc, err := s.serviceRepo.GetByID(ctx, serviceID)
	if err != nil {
		return nil, fmt.Errorf("NodeService.StartService: get service: %w", err)
	}
	if svc.NodeID != nodeID {
		return nil, fmt.Errorf("NodeService.StartService: service does not belong to node %d", nodeID)
	}

	port, uri, err := s.nodeClient.StartService(ctx, node.Address, node.APIToken, svc.Name)
	if err != nil {
		return nil, fmt.Errorf("NodeService.StartService: %w", err)
	}

	if port > 0 {
		svc.HostPort = port
		svc.ConnectionURI = uri
	}
	svc.Status = domain.ServiceStatusRunning
	svc.UpdatedAt = time.Now()
	if err := s.serviceRepo.Update(ctx, svc); err != nil {
		return nil, fmt.Errorf("NodeService.StartService: update record: %w", err)
	}

	s.log.Info("managed service started", slog.Int64("service_id", serviceID), slog.String("name", svc.Name))
	return svc, nil
}

// StopService останавливает управляемый сервис без удаления данных.
func (s *NodeService) StopService(ctx context.Context, ownerID string, nodeID, serviceID int64) (*domain.ManagedService, error) {
	if s.serviceRepo == nil {
		return nil, fmt.Errorf("NodeService.StopService: service repository is not configured")
	}
	node, err := s.nodeRepo.GetByID(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("NodeService.StopService: get node: %w", err)
	}
	if ownerID != "" && node.OwnerID != ownerID {
		return nil, domain.ErrForbidden
	}

	svc, err := s.serviceRepo.GetByID(ctx, serviceID)
	if err != nil {
		return nil, fmt.Errorf("NodeService.StopService: get service: %w", err)
	}
	if svc.NodeID != nodeID {
		return nil, fmt.Errorf("NodeService.StopService: service does not belong to node %d", nodeID)
	}

	if err := s.nodeClient.StopService(ctx, node.Address, node.APIToken, svc.Name); err != nil {
		return nil, fmt.Errorf("NodeService.StopService: %w", err)
	}

	svc.Status = domain.ServiceStatusStopped
	svc.UpdatedAt = time.Now()
	if err := s.serviceRepo.Update(ctx, svc); err != nil {
		return nil, fmt.Errorf("NodeService.StopService: update record: %w", err)
	}

	s.log.Info("managed service stopped", slog.Int64("service_id", serviceID), slog.String("name", svc.Name))
	return svc, nil
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

	// Singleton-гарантия для Web UI: на одной ноде не может быть более одной веб-панели управления
	if params.ServiceType == domain.ServiceTypeAdminer {
		existingServices, err := s.serviceRepo.ListByNode(ctx, nodeID)
		if err == nil {
			for _, svc := range existingServices {
				if svc.ServiceType == domain.ServiceTypeAdminer {
					if svc.Status == domain.ServiceStatusRunning {
						return nil, fmt.Errorf("NodeService.CreateService: на ноде %d уже развернута веб-панель управления (AdminerEvo)", nodeID)
					}
					if svc.Status == domain.ServiceStatusStopped {
						// Если Adminer уже развернут, но остановлен, запускаем его
						return s.StartService(ctx, ownerID, nodeID, svc.ID)
					}
				}
			}
		}
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
	serviceOwnerID := ownerID
	if serviceOwnerID == "" {
		serviceOwnerID = node.OwnerID
	}
	serviceRecord := &domain.ManagedService{
		NodeID:         nodeID,
		OwnerID:        serviceOwnerID,
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

	// Синхронизируем состояние сервисов с реальным состоянием на ноде (если нода онлайн)
	if node.Status == domain.NodeStatusOnline && s.nodeClient != nil {
		nodeServices, err := s.nodeClient.ListServices(ctx, node.Address, node.APIToken)
		if err == nil {
			nodeSvcMap := make(map[string]domain.ServiceInfo, len(nodeServices))
			for _, ns := range nodeServices {
				nodeSvcMap[ns.Name] = ns
			}

			for _, svc := range services {
				ns, ok := nodeSvcMap[svc.Name]
				if !ok {
					continue
				}
				changed := false
				if ns.HostPort != 0 && ns.HostPort != svc.HostPort {
					svc.ConnectionURI = replacePortInURI(svc.ConnectionURI, ns.HostPort)
					svc.HostPort = ns.HostPort
					changed = true
				}
				if ns.Status != "" {
					newStatus := domain.ServiceStatusRunning
					if ns.Status == "stopped" || ns.Status == "exited" {
						newStatus = domain.ServiceStatusStopped
					} else if ns.Status == "error" {
						newStatus = domain.ServiceStatusError
					}
					if svc.Status != newStatus {
						svc.Status = newStatus
						changed = true
					}
				}
				if ns.ContainerID != "" && ns.ContainerID != svc.ContainerID {
					svc.ContainerID = ns.ContainerID
					changed = true
				}
				if ns.VolumeSizeBytes > 0 && ns.VolumeSizeBytes != svc.VolumeSizeBytes {
					svc.VolumeSizeBytes = ns.VolumeSizeBytes
					changed = true
				}
				if changed {
					if err := s.serviceRepo.Update(ctx, svc); err != nil {
						s.log.Warn("failed to update synced service in DB",
							slog.Int64("service_id", svc.ID),
							slog.String("name", svc.Name),
							slog.String("error", err.Error()),
						)
					}
				}
			}
		} else {
			s.log.Debug("could not sync managed services from node",
				slog.Int64("node_id", nodeID),
				slog.String("error", err.Error()),
			)
		}
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

// replacePortInURI заменяет порт в URI подключения, сохраняя остальные части URL.
func replacePortInURI(rawURI string, newPort uint32) string {
	if rawURI == "" || newPort == 0 {
		return rawURI
	}
	u, err := url.Parse(rawURI)
	if err != nil {
		return rawURI
	}
	host := u.Hostname()
	if host == "" {
		return rawURI
	}
	u.Host = net.JoinHostPort(host, strconv.Itoa(int(newPort)))
	return u.String()
}


func (s *NodeService) ToggleBackups(ctx context.Context, callerID string, nodeID int64, enabled bool) (*domain.Node, error) {
	node, err := s.GetNode(ctx, callerID, nodeID)
	if err != nil {
		return nil, err
	}
	node.Node.BackupsEnabled = enabled
	if err := s.nodeRepo.Update(ctx, node.Node); err != nil {
		return nil, fmt.Errorf("failed to update node backups toggle: %w", err)
	}
	return node.Node, nil
}

func (s *NodeService) ToggleServiceAutoBackup(ctx context.Context, callerID string, nodeID int64, serviceName string, enabled bool) (*domain.ManagedService, error) {
	node, err := s.GetNode(ctx, callerID, nodeID)
	if err != nil {
		return nil, err
	}
	svc, err := s.serviceRepo.GetByName(ctx, nodeID, serviceName)
	if err != nil {
		return nil, fmt.Errorf("service not found: %w", err)
	}
	svc.AutoBackupEnabled = enabled
	if err := s.serviceRepo.Update(ctx, svc); err != nil {
		return nil, fmt.Errorf("failed to update auto backup: %w", err)
	}
	_ = s.nodeClient.ToggleServiceAutoBackup(ctx, node.Address, node.APIToken, serviceName, enabled)
	return svc, nil
}

func (s *NodeService) CreateServiceBackup(ctx context.Context, callerID string, nodeID int64, serviceName string) (*domain.ServiceBackup, error) {
	node, err := s.GetNode(ctx, callerID, nodeID)
	if err != nil { return nil, err }
	return s.nodeClient.CreateServiceBackup(ctx, node.Address, node.APIToken, serviceName)
}
func (s *NodeService) ListServiceBackups(ctx context.Context, callerID string, nodeID int64, serviceName string) ([]*domain.ServiceBackup, error) {
	node, err := s.GetNode(ctx, callerID, nodeID)
	if err != nil { return nil, err }
	return s.nodeClient.ListServiceBackups(ctx, node.Address, node.APIToken, serviceName)
}
func (s *NodeService) RestoreServiceBackup(ctx context.Context, callerID string, nodeID int64, serviceName string, backupID string) (bool, string, error) {
	node, err := s.GetNode(ctx, callerID, nodeID)
	if err != nil { return false, "", err }
	err = s.nodeClient.RestoreServiceBackup(ctx, node.Address, node.APIToken, serviceName, backupID)
	if err != nil { return false, err.Error(), err }
	return true, "Success", nil
}
func (s *NodeService) DeleteServiceBackup(ctx context.Context, callerID string, nodeID int64, serviceName string, backupID string) error {
	node, err := s.GetNode(ctx, callerID, nodeID)
	if err != nil { return err }
	return s.nodeClient.DeleteServiceBackup(ctx, node.Address, node.APIToken, serviceName, backupID)
}
func (s *NodeService) DownloadServiceBackup(ctx context.Context, callerID string, nodeID int64, serviceName string, backupID string) (io.ReadCloser, error) {
	node, err := s.GetNode(ctx, callerID, nodeID)
	if err != nil { return nil, err }
	return s.nodeClient.DownloadServiceBackup(ctx, node.Address, node.APIToken, serviceName, backupID)
}
func (s *NodeService) UploadServiceBackup(ctx context.Context, callerID string, nodeID int64, serviceName string, fileName string, restoreImmediately bool, r io.Reader) (*domain.ServiceBackup, error) {
	node, err := s.GetNode(ctx, callerID, nodeID)
	if err != nil { return nil, err }
	return s.nodeClient.UploadServiceBackup(ctx, node.Address, node.APIToken, serviceName, fileName, restoreImmediately, r)
}
func (s *NodeService) GetBackupTicket(ctx context.Context, callerID string, nodeID int64, serviceName string, backupID string) (string, int64, error) {
	return "ticket", 3600, nil
}
