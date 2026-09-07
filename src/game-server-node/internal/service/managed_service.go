package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Be4Die/game-developer-hub/game-server-node/internal/domain"
)

// ManagedServiceRecord хранит информацию о развернутом управляемом сервисе.
type ManagedServiceRecord struct {
	Name          string             `json:"name"`
	ServiceType   domain.ServiceType `json:"service_type"`
	ContainerID   string             `json:"container_id"`
	HostPort      uint32             `json:"host_port"`
	InternalPort  uint32             `json:"internal_port"`
	VolumePath    string             `json:"volume_path"`
	ConnectionURI string             `json:"connection_uri"`
	Status        string             `json:"status"`
	CreatedAt     time.Time          `json:"created_at"`
}

// ManagedServiceState управляет состоянием баз данных и хранилищ на ноде.
type ManagedServiceState struct {
	mu           sync.RWMutex
	services     map[string]ManagedServiceRecord
	registryPath string
	volumesDir   string
	log          *slog.Logger
	runtime      domain.ContainerRuntime
}

// NewManagedServiceState создает менеджер сервисов на ноде.
func NewManagedServiceState(log *slog.Logger, runtime domain.ContainerRuntime, registryPath, volumesDir string) *ManagedServiceState {
	if volumesDir == "" {
		volumesDir = os.Getenv("GDH_VOLUMES_DIR")
		if volumesDir == "" {
			volumesDir = filepath.Join(os.TempDir(), "gdh-volumes")
		}
	}

	state := &ManagedServiceState{
		services:     make(map[string]ManagedServiceRecord),
		registryPath: registryPath,
		volumesDir:   volumesDir,
		log:          log,
		runtime:      runtime,
	}

	_ = state.load()
	return state
}

func (s *ManagedServiceState) load() error {
	if s.registryPath == "" {
		return nil
	}
	data, err := os.ReadFile(s.registryPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return json.Unmarshal(data, &s.services)
}

func (s *ManagedServiceState) save() error {
	if s.registryPath == "" {
		return nil
	}
	s.mu.RLock()
	data, err := json.MarshalIndent(s.services, "", "  ")
	s.mu.RUnlock()
	if err != nil {
		return err
	}
	_ = os.MkdirAll(filepath.Dir(s.registryPath), 0755)
	return os.WriteFile(s.registryPath, data, 0644)
}

// DeployService разворачивает управляемый сервис с персистентным томом.
func (s *ManagedServiceState) DeployService(ctx context.Context, req domain.DeployServiceRequest, hostIP string) (*domain.DeployServiceResult, error) {
	const op = "ManagedServiceState.DeployService"

	s.mu.Lock()
	if _, exists := s.services[req.Name]; exists {
		s.mu.Unlock()
		return nil, fmt.Errorf("%s: service '%s' already exists", op, req.Name)
	}
	s.mu.Unlock()

	// 1. Создаем папку под том на хосте (для Raw Volume или fallback)
	volDir := filepath.Join(s.volumesDir, req.Name)
	_ = os.MkdirAll(volDir, 0777)

	// Имя именованного тома Docker (гарантирует кросс-платформенную изоляцию и работу в Docker Desktop/DinD)
	volumeName := fmt.Sprintf("gdh-vol-%s", req.Name)

	// Если запрошен чистый том хранения (Raw Volume), контейнер не требуется
	if req.ServiceType == domain.ServiceTypeVolume {
		record := ManagedServiceRecord{
			Name:        req.Name,
			ServiceType: req.ServiceType,
			Status:      "running",
			VolumePath:  volDir,
			CreatedAt:   time.Now(),
		}
		s.mu.Lock()
		s.services[req.Name] = record
		s.mu.Unlock()
		_ = s.save()

		return &domain.DeployServiceResult{
			Name:          req.Name,
			ContainerID:   "",
			HostPort:      0,
			ConnectionURI: volDir,
			VolumePath:    volDir,
		}, nil
	}

	// 2. Инициализируем bridge-сеть
	_ = s.runtime.EnsureNetwork(ctx, "gdh-network")

	var (
		imageTag     string
		internalPort uint32
		bindPath     string
		args         []string
		env          = make(map[string]string)
	)

	// Копируем переданные env vars
	for k, v := range req.EnvVars {
		env[k] = v
	}

	switch req.ServiceType {
	case domain.ServiceTypePostgres:
		imageTag = "postgres:17-alpine"
		internalPort = 5432
		bindPath = fmt.Sprintf("%s:/var/lib/postgresql/data", volumeName)
		if _, ok := env["POSTGRES_USER"]; !ok {
			env["POSTGRES_USER"] = "postgres"
		}
		if _, ok := env["POSTGRES_PASSWORD"]; !ok {
			env["POSTGRES_PASSWORD"] = "postgres"
		}
		if _, ok := env["POSTGRES_DB"]; !ok {
			env["POSTGRES_DB"] = "game_db"
		}

	case domain.ServiceTypeRedis:
		imageTag = "redis:7-alpine"
		internalPort = 6379
		bindPath = fmt.Sprintf("%s:/data", volumeName)
		args = []string{"redis-server", "--appendonly", "yes"}
		if pass, ok := env["REDIS_PASSWORD"]; ok && pass != "" {
			args = append(args, "--requirepass", pass)
		}

	case domain.ServiceTypeMySQL:
		imageTag = "mariadb:11"
		internalPort = 3306
		bindPath = fmt.Sprintf("%s:/var/lib/mysql", volumeName)
		if _, ok := env["MYSQL_ROOT_PASSWORD"]; !ok {
			env["MYSQL_ROOT_PASSWORD"] = "root"
		}
		if _, ok := env["MYSQL_DATABASE"]; !ok {
			env["MYSQL_DATABASE"] = "game_db"
		}

	case domain.ServiceTypeMinIO:
		imageTag = "minio/minio:latest"
		internalPort = 9000
		bindPath = fmt.Sprintf("%s:/data", volumeName)
		args = []string{"server", "/data", "--console-address", ":9001"}
		if _, ok := env["MINIO_ROOT_USER"]; !ok {
			env["MINIO_ROOT_USER"] = "minioadmin"
		}
		if _, ok := env["MINIO_ROOT_PASSWORD"]; !ok {
			env["MINIO_ROOT_PASSWORD"] = "minioadmin"
		}

	case domain.ServiceTypeAdminer:
		imageTag = "adminer:latest"
		internalPort = 8080

	case domain.ServiceTypePGAdmin:
		imageTag = "dpage/pgadmin4:latest"
		internalPort = 80
		bindPath = fmt.Sprintf("%s:/var/lib/pgadmin", volumeName)
		if _, ok := env["PGADMIN_DEFAULT_EMAIL"]; !ok {
			env["PGADMIN_DEFAULT_EMAIL"] = "admin@gdh.local"
		}
		if _, ok := env["PGADMIN_DEFAULT_PASSWORD"]; !ok {
			env["PGADMIN_DEFAULT_PASSWORD"] = "admin"
		}

	default:
		return nil, fmt.Errorf("%s: unsupported service type %d", op, req.ServiceType)
	}

	// 3. Подгружаем образ
	if err := s.runtime.PullImage(ctx, imageTag); err != nil {
		s.log.Warn("pull image failed, trying local cache", slog.String("image", imageTag), slog.String("err", err.Error()))
	}

	// 4. Создаем контейнер
	var binds []string
	if bindPath != "" {
		binds = append(binds, bindPath)
	}

	containerName := fmt.Sprintf("gdh-svc-%s", req.Name)
	opts := domain.ContainerOpts{
		ContainerName: containerName,
		ImageTag:      imageTag,
		InternalPort:  internalPort,
		HostPort:      req.Port,
		EnvVars:       env,
		Args:          args,
		Binds:         binds,
		Network:       "gdh-network",
		Labels: map[string]string{
			"managed_by":   "game-server-node",
			"service_type": fmt.Sprintf("%d", req.ServiceType),
			"service_name": req.Name,
		},
	}

	containerID, err := s.runtime.CreateContainer(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("%s: create container: %w", op, err)
	}

	// 5. Запускаем контейнер
	if err := s.runtime.StartContainer(ctx, containerID); err != nil {
		_ = s.runtime.RemoveContainer(ctx, containerID)
		return nil, fmt.Errorf("%s: start container: %w", op, err)
	}

	// 6. Получаем реальный хост-порт
	hostPort, err := s.runtime.GetHostPort(ctx, containerID, internalPort)
	if err != nil {
		_ = s.runtime.StopContainer(ctx, containerID, 5*time.Second)
		_ = s.runtime.RemoveContainer(ctx, containerID)
		return nil, fmt.Errorf("%s: get host port: %w", op, err)
	}

	if hostIP == "" {
		hostIP = "127.0.0.1"
	}

	// 7. Формируем строку подключения
	var connURI string
	switch req.ServiceType {
	case domain.ServiceTypePostgres:
		connURI = fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable",
			env["POSTGRES_USER"], env["POSTGRES_PASSWORD"], hostIP, hostPort, env["POSTGRES_DB"])
	case domain.ServiceTypeRedis:
		if pass, ok := env["REDIS_PASSWORD"]; ok && pass != "" {
			connURI = fmt.Sprintf("redis://:%s@%s:%d/0", pass, hostIP, hostPort)
		} else {
			connURI = fmt.Sprintf("redis://%s:%d/0", hostIP, hostPort)
		}
	case domain.ServiceTypeMySQL:
		user := env["MYSQL_USER"]
		if user == "" {
			user = "root"
		}
		pass := env["MYSQL_PASSWORD"]
		if pass == "" {
			pass = env["MYSQL_ROOT_PASSWORD"]
		}
		connURI = fmt.Sprintf("mysql://%s:%s@%s:%d/%s", user, pass, hostIP, hostPort, env["MYSQL_DATABASE"])
	case domain.ServiceTypeMinIO, domain.ServiceTypeAdminer, domain.ServiceTypePGAdmin:
		connURI = fmt.Sprintf("http://%s:%d", hostIP, hostPort)
	}

	record := ManagedServiceRecord{
		Name:          req.Name,
		ServiceType:   req.ServiceType,
		ContainerID:   containerID,
		HostPort:      hostPort,
		InternalPort:  internalPort,
		VolumePath:    volumeName,
		ConnectionURI: connURI,
		Status:        "running",
		CreatedAt:     time.Now(),
	}

	s.mu.Lock()
	s.services[req.Name] = record
	s.mu.Unlock()
	_ = s.save()

	s.log.Info("managed service deployed",
		slog.String("name", req.Name),
		slog.String("type", fmt.Sprintf("%d", req.ServiceType)),
		slog.Uint64("host_port", uint64(hostPort)),
		slog.String("volume", volumeName),
	)

	return &domain.DeployServiceResult{
		Name:          req.Name,
		ContainerID:   containerID,
		HostPort:      hostPort,
		ConnectionURI: connURI,
		VolumePath:    volumeName,
	}, nil
}

// RemoveService останавливает контейнер сервиса и опционально удаляет персистентный том.
func (s *ManagedServiceState) RemoveService(ctx context.Context, name string, deleteVolume bool) error {
	const op = "ManagedServiceState.RemoveService"

	s.mu.Lock()
	svc, ok := s.services[name]
	if !ok {
		s.mu.Unlock()
		return fmt.Errorf("%s: service '%s' not found", op, name)
	}
	delete(s.services, name)
	s.mu.Unlock()
	_ = s.save()

	if svc.ContainerID != "" {
		_ = s.runtime.StopContainer(ctx, svc.ContainerID, 10*time.Second)
		_ = s.runtime.RemoveContainer(ctx, svc.ContainerID)
	}

	if deleteVolume {
		if svc.VolumePath != "" {
			_ = os.RemoveAll(svc.VolumePath)
		}
		_ = s.runtime.RemoveVolume(ctx, fmt.Sprintf("gdh-vol-%s", name))
		s.log.Info("deleted service volume", slog.String("name", name), slog.String("path", svc.VolumePath))
	}

	s.log.Info("managed service removed", slog.String("name", name))
	return nil
}

// ListServices возвращает список всех управляемых сервисов и размер их томов.
func (s *ManagedServiceState) ListServices(ctx context.Context) ([]domain.ServiceInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]domain.ServiceInfo, 0, len(s.services))
	for _, svc := range s.services {
		size := calculateDirSize(svc.VolumePath)
		result = append(result, domain.ServiceInfo{
			Name:            svc.Name,
			ServiceType:     svc.ServiceType,
			ContainerID:     svc.ContainerID,
			Status:          svc.Status,
			HostPort:        svc.HostPort,
			VolumePath:      svc.VolumePath,
			VolumeSizeBytes: size,
		})
	}
	return result, nil
}

// calculateDirSize рекурсивно подсчитывает размер директории в байтах.
func calculateDirSize(path string) uint64 {
	var size int64
	_ = filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err == nil && info != nil && !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	if size < 0 {
		return 0
	}
	return uint64(size)
}

// DeployService делегирует вызов в ManagedServiceState.
func (s *DeploymentService) DeployService(ctx context.Context, req domain.DeployServiceRequest) (*domain.DeployServiceResult, error) {
	hostIP := ""
	if host, _, err := net.SplitHostPort(s.nodeID); err == nil {
		hostIP = host
	}
	return s.managedState.DeployService(ctx, req, hostIP)
}

// RemoveService делегирует вызов в ManagedServiceState.
func (s *DeploymentService) RemoveService(ctx context.Context, name string, deleteVolume bool) error {
	return s.managedState.RemoveService(ctx, name, deleteVolume)
}

// ListServices делегирует вызов в ManagedServiceState.
func (s *DeploymentService) ListServices(ctx context.Context) ([]domain.ServiceInfo, error) {
	return s.managedState.ListServices(ctx)
}
