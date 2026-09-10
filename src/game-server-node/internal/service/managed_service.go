package service

import (
	"context"
	"encoding/json"
	"fmt"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/Be4Die/game-developer-hub/game-server-node/assets"
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
	AutoBackupEnabled bool             `json:"auto_backup_enabled"`
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
	state.Reconcile(context.Background())
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
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.saveLocked()
}

func (s *ManagedServiceState) saveLocked() error {
	if s.registryPath == "" {
		return nil
	}
	data, err := json.MarshalIndent(s.services, "", "  ")
	if err != nil {
		return err
	}
	_ = os.MkdirAll(filepath.Dir(s.registryPath), 0755)
	return os.WriteFile(s.registryPath, data, 0644)
}

// Reconcile синхронизирует состояние сервисов с Docker:
// - Запускает остановленные контейнеры
// - Обеспечивает политику рестарта unless-stopped
// - Актуализирует host port и connection URI
func (s *ManagedServiceState) Reconcile(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	hasChanges := false
	for name, svc := range s.services {
		if svc.ServiceType == domain.ServiceTypeVolume {
			continue
		}

		targetID := svc.ContainerID
		if targetID == "" {
			targetID = fmt.Sprintf("gdh-svc-%s", name)
		}

		details, err := s.runtime.InspectContainer(ctx, targetID)
		if err != nil {
			containerName := fmt.Sprintf("gdh-svc-%s", name)
			if targetID != containerName {
				if d, errName := s.runtime.InspectContainer(ctx, containerName); errName == nil {
					details = d
					targetID = containerName
					svc.ContainerID = d.ID
					hasChanges = true
					err = nil
				}
			}
		}

		if err != nil {
			s.log.Warn("managed service container not found in runtime",
				slog.String("service", name),
				slog.String("container_id", svc.ContainerID),
				slog.String("error", err.Error()),
			)
			if svc.Status != "stopped" && svc.Status != "error" {
				svc.Status = "stopped"
				hasChanges = true
			}
			s.services[name] = svc
			continue
		}

		// Убеждаемся, что установлена политика unless-stopped
		if details.RestartPolicy != "unless-stopped" && details.RestartPolicy != "always" {
			s.log.Info("updating container restart policy to unless-stopped",
				slog.String("service", name),
				slog.String("container_id", details.ID),
			)
			if err := s.runtime.UpdateRestartPolicy(ctx, details.ID, "unless-stopped"); err != nil {
				s.log.Warn("failed to update restart policy", slog.String("error", err.Error()))
			}
		}

		// Запускаем контейнер, если он остановлен (и не был намеренно остановлен пользователем)
		if !details.Running {
			if svc.Status == "stopped" {
				continue
			}
			s.log.Info("starting stopped managed service container",
				slog.String("service", name),
				slog.String("container_id", details.ID),
			)
			if err := s.runtime.StartContainer(ctx, details.ID); err != nil {
				s.log.Error("failed to start stopped service container",
					slog.String("service", name),
					slog.String("error", err.Error()),
				)
				svc.Status = "error"
				hasChanges = true
			} else {
				details.Running = true
				svc.Status = "running"
				hasChanges = true
			}
		} else {
			if svc.Status != "running" {
				svc.Status = "running"
				hasChanges = true
			}
		}

		// Проверяем актуальный хост-порт
		if details.Running && svc.InternalPort > 0 {
			actualPort, err := s.runtime.GetHostPort(ctx, details.ID, svc.InternalPort)
			if err == nil && actualPort > 0 {
				if svc.HostPort != actualPort {
					s.log.Info("reconciled service host port from docker",
						slog.String("service", name),
						slog.Uint64("old_port", uint64(svc.HostPort)),
						slog.Uint64("new_port", uint64(actualPort)),
					)
					svc.ConnectionURI = replacePortInURI(svc.ConnectionURI, actualPort)
					svc.HostPort = actualPort
					hasChanges = true
				}
			}
		}

		// Если это Adminer и он запущен, обеспечиваем наличие темы и плагинов
		if svc.ServiceType == domain.ServiceTypeAdminer && details.Running {
			if len(assets.AdminerCSS) > 0 {
				_ = s.runtime.CopyToContainer(ctx, details.ID, "/var/www/html", "adminer.css", assets.AdminerCSS)
			}
			if len(assets.AdminerPasswordPlugin) > 0 {
				_ = s.runtime.CopyToContainer(ctx, details.ID, "/var/www/html/plugins-enabled", "password.php", assets.AdminerPasswordPlugin)
			}
		}

		s.services[name] = svc
	}

	if hasChanges {
		_ = s.saveLocked()
	}
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

	// 3. Подгружаем образ: если образа нет локально, скачиваем из реестра
	if !s.runtime.ImageExists(ctx, imageTag) {
		s.log.Info("image not found locally, pulling from registry", slog.String("image", imageTag))
		if err := s.runtime.PullImage(ctx, imageTag); err != nil {
			return nil, fmt.Errorf("%s: image '%s' not found locally and failed to pull from registry: %w", op, imageTag, err)
		}
	}

	// 4. Создаем контейнер
	var binds []string
	if bindPath != "" {
		binds = append(binds, bindPath)
	}

	hostPort := req.Port
	if hostPort == 0 {
		if p, err := findFreePort(); err == nil {
			hostPort = p
		} else {
			s.log.Warn("failed to allocate free host port, falling back to dynamic port", slog.String("error", err.Error()))
		}
	}

	containerName := fmt.Sprintf("gdh-svc-%s", req.Name)
	opts := domain.ContainerOpts{
		ContainerName: containerName,
		ImageTag:      imageTag,
		InternalPort:  internalPort,
		HostPort:      hostPort,
		RestartPolicy: "unless-stopped",
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

	// 5. Для Adminer копируем файл кастомной темы GDH и плагины прямо в контейнер
	if req.ServiceType == domain.ServiceTypeAdminer {
		if len(assets.AdminerCSS) > 0 {
			if err := s.runtime.CopyToContainer(ctx, containerID, "/var/www/html", "adminer.css", assets.AdminerCSS); err != nil {
				s.log.Warn("failed to copy custom adminer theme into container", slog.String("err", err.Error()))
			}
		}
		if len(assets.AdminerPasswordPlugin) > 0 {
			if err := s.runtime.CopyToContainer(ctx, containerID, "/var/www/html/plugins-enabled", "password.php", assets.AdminerPasswordPlugin); err != nil {
				s.log.Warn("failed to copy adminer password plugin into container", slog.String("err", err.Error()))
			}
		}
	}

	// 6. Запускаем контейнер
	if err := s.runtime.StartContainer(ctx, containerID); err != nil {
		_ = s.runtime.RemoveContainer(ctx, containerID)
		return nil, fmt.Errorf("%s: start container: %w", op, err)
	}

	// 6. Получаем реальный хост-порт
	actualHostPort, err := s.runtime.GetHostPort(ctx, containerID, internalPort)
	if err != nil {
		_ = s.runtime.StopContainer(ctx, containerID, 5*time.Second)
		_ = s.runtime.RemoveContainer(ctx, containerID)
		return nil, fmt.Errorf("%s: get host port: %w", op, err)
	}
	hostPort = actualHostPort

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
	case domain.ServiceTypeAdminer, domain.ServiceTypePGAdmin:
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

// StopService останавливает контейнер сервиса без удаления тома и данных.
func (s *ManagedServiceState) StopService(ctx context.Context, name string) error {
	const op = "ManagedServiceState.StopService"

	s.mu.Lock()
	svc, ok := s.services[name]
	if !ok {
		s.mu.Unlock()
		return fmt.Errorf("%s: service '%s' not found", op, name)
	}
	svc.Status = "stopped"
	s.services[name] = svc
	s.mu.Unlock()
	_ = s.save()

	if svc.ContainerID != "" {
		_ = s.runtime.StopContainer(ctx, svc.ContainerID, 10*time.Second)
	}

	s.log.Info("managed service stopped", slog.String("name", name))
	return nil
}

// StartService запускает ранее остановленный контейнер сервиса.
func (s *ManagedServiceState) StartService(ctx context.Context, name string) (uint32, string, error) {
	const op = "ManagedServiceState.StartService"

	s.mu.Lock()
	svc, ok := s.services[name]
	if !ok {
		s.mu.Unlock()
		return 0, "", fmt.Errorf("%s: service '%s' not found", op, name)
	}
	s.mu.Unlock()

	if svc.ContainerID != "" {
		if err := s.runtime.StartContainer(ctx, svc.ContainerID); err != nil {
			return 0, "", fmt.Errorf("%s: start container: %w", op, err)
		}
		if svc.InternalPort > 0 {
			actualPort, err := s.runtime.GetHostPort(ctx, svc.ContainerID, svc.InternalPort)
			if err == nil && actualPort > 0 {
				svc.ConnectionURI = replacePortInURI(svc.ConnectionURI, actualPort)
				svc.HostPort = actualPort
			}
		}
	}

	s.mu.Lock()
	svc.Status = "running"
	s.services[name] = svc
	s.mu.Unlock()
	_ = s.save()

	s.log.Info("managed service started", slog.String("name", name), slog.Uint64("host_port", uint64(svc.HostPort)))
	return svc.HostPort, svc.ConnectionURI, nil
}

// ListServices возвращает список всех управляемых сервисов и размер их томов.
func (s *ManagedServiceState) ListServices(ctx context.Context) ([]domain.ServiceInfo, error) {
	s.Reconcile(ctx)

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
			AutoBackupEnabled: svc.AutoBackupEnabled,
			VolumeSizeBytes: size,
		})
	}
	return result, nil
}

// findFreePort находит свободный TCP-порт в системе.
func findFreePort() (uint32, error) {
	l, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return uint32(l.Addr().(*net.TCPAddr).Port), nil
}

// replacePortInURI заменяет порт в URI подключения, сохраняя схему, креды и параметры.
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

// GetService возвращает информацию о сервисе по его имени.
func (s *ManagedServiceState) GetService(name string) (ManagedServiceRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	svc, ok := s.services[name]
	return svc, ok
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

// StopService делегирует остановку сервиса в ManagedServiceState.
func (s *DeploymentService) StopService(ctx context.Context, name string) error {
	return s.managedState.StopService(ctx, name)
}

// StartService делегирует запуск сервиса в ManagedServiceState.
func (s *DeploymentService) StartService(ctx context.Context, name string) (uint32, string, error) {
	return s.managedState.StartService(ctx, name)
}

// ListServices делегирует вызов в ManagedServiceState.
func (s *DeploymentService) ListServices(ctx context.Context) ([]domain.ServiceInfo, error) {
	return s.managedState.ListServices(ctx)
}

// CreateBackup делегирует вызов в BackupManager.
func (s *DeploymentService) CreateBackup(ctx context.Context, serviceName string) (*domain.BackupInfo, error) {
	return s.backupMgr.CreateBackup(ctx, serviceName, domain.BackupTypeManual)
}

// ListBackups делегирует вызов в BackupManager.
func (s *DeploymentService) ListBackups(ctx context.Context, serviceName string) ([]domain.BackupInfo, error) {
	return s.backupMgr.ListBackups(ctx, serviceName)
}

// RestoreBackup делегирует вызов в BackupManager.
func (s *DeploymentService) RestoreBackup(ctx context.Context, serviceName, backupID string) error {
	return s.backupMgr.RestoreBackup(ctx, serviceName, backupID)
}

// DeleteBackup делегирует вызов в BackupManager.
func (s *DeploymentService) DeleteBackup(ctx context.Context, serviceName, backupID string) error {
	return s.backupMgr.DeleteBackup(ctx, serviceName, backupID)
}

// OpenBackup делегирует вызов в BackupManager.
func (s *DeploymentService) OpenBackup(ctx context.Context, serviceName, backupID string) (io.ReadCloser, int64, string, error) {
	return s.backupMgr.OpenBackup(ctx, serviceName, backupID)
}

// UploadBackup делегирует вызов в BackupManager.
func (s *DeploymentService) UploadBackup(ctx context.Context, serviceName, fileName string, restoreImmediately bool, r io.Reader) (*domain.BackupInfo, error) {
	return s.backupMgr.UploadBackup(ctx, serviceName, fileName, restoreImmediately, r)
}


func (s *ManagedServiceState) ToggleServiceAutoBackup(name string, enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.services[name]
	if !ok {
		return errors.New("service not found")
	}

	record.AutoBackupEnabled = enabled
	s.services[name] = record

	return s.saveLocked()
}
