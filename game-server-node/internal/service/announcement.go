// Package service реализует бизнес-логику анонсирования ноды в оркестраторе.
package service

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/Be4Die/game-developer-hub/game-server-node/internal/infrastructure/client/orchestrator"
	"github.com/Be4Die/game-developer-hub/game-server-node/internal/infrastructure/config"
	"github.com/Be4Die/game-developer-hub/game-server-node/internal/infrastructure/sysinfo"
)

// AnnouncementService управляет процессом анонсирования ноды в оркестраторе.
type AnnouncementService struct {
	log         *slog.Logger
	cfg         *config.Config
	sysProvider sysinfo.Provider
}

// AnnouncementResult содержит результат успешного анонсирования.
type AnnouncementResult struct {
	NodeID int64
}

// NewAnnouncementService создаёт сервис анонсирования.
func NewAnnouncementService(log *slog.Logger, cfg *config.Config, sysProvider sysinfo.Provider) *AnnouncementService {
	return &AnnouncementService{
		log:         log,
		cfg:         cfg,
		sysProvider: sysProvider,
	}
}

// Announce выполняет анонсирование ноды в оркестраторе.
// Нода отправляет свой NODE_API_KEY, который оркестратор использует
// как токен авторизации — пользователь вводит тот же NODE_API_KEY
// для подключения ноды через дашборд.
func (s *AnnouncementService) Announce(ctx context.Context) (*AnnouncementResult, error) {
	if s.cfg.Orchestrator.Mode != "auto-discovery" {
		return nil, fmt.Errorf("AnnouncementService.Announce: mode is %s, expected auto-discovery", s.cfg.Orchestrator.Mode)
	}

	// Определяем внешний адрес ноды.
	address, err := s.determineExternalAddress()
	if err != nil {
		return nil, fmt.Errorf("AnnouncementService.Announce: determine address: %w", err)
	}

	// Получаем системную информацию.
	resources, err := s.sysProvider.GetMax()
	if err != nil {
		return nil, fmt.Errorf("AnnouncementService.Announce: get system info: %w", err)
	}

	// Создаём клиент для подключения к оркестратору.
	//nolint:contextcheck // контекст передаётся в метод
	client, err := orchestrator.NewClient(ctx, s.cfg.Orchestrator.Address, s.cfg.Orchestrator.AnnounceTimeout)
	if err != nil {
		return nil, fmt.Errorf("AnnouncementService.Announce: create client: %w", err)
	}
	defer func() {
		_ = client.Close()
	}()

	// Формируем и отправляем запрос. Передаём NODE_API_KEY как api_key.
	req := &orchestrator.AnnounceRequest{
		Address:          address,
		Region:           s.cfg.Node.Region,
		AgentVersion:     s.cfg.Node.Version,
		CPUCores:         resources.CPUCores,
		TotalMemoryBytes: resources.TotalMemorySize,
		TotalDiskBytes:   resources.TotalDiskSpace,
		APIKey:           s.cfg.APIKey,
	}

	announceCtx, cancel := context.WithTimeout(ctx, s.cfg.Orchestrator.AnnounceTimeout)
	defer cancel()

	resp, err := client.AnnounceNode(announceCtx, req)
	if err != nil {
		return nil, fmt.Errorf("AnnouncementService.Announce: announce: %w", err)
	}

	s.log.Info("node announced successfully",
		slog.Int64("node_id", resp.NodeID),
		slog.String("address", address),
	)

	return &AnnouncementResult{
		NodeID: resp.NodeID,
	}, nil
}

// AnnounceWithRetry выполняет анонсирование с повторными попытками.
// Продолжает попытки до успеха или отмены контекста.
func (s *AnnouncementService) AnnounceWithRetry(ctx context.Context) (*AnnouncementResult, error) {
	ticker := time.NewTicker(s.cfg.Orchestrator.AnnounceInterval)
	defer ticker.Stop()

	// Первая попытка сразу.
	result, err := s.Announce(ctx)
	if err == nil {
		return result, nil
	}

	s.log.Warn("initial announce failed, will retry", slog.String("error", err.Error()))

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("AnnouncementService.AnnounceWithRetry: %w", ctx.Err())
		case <-ticker.C:
			result, err := s.Announce(ctx)
			if err == nil {
				return result, nil
			}
			s.log.Warn("announce retry failed", slog.String("error", err.Error()))
		}
	}
}

// determineExternalAddress определяет внешний адрес ноды.
// Если задан external_address в конфиге — использует его.
// Иначе пытается определить автоматически.
func (s *AnnouncementService) determineExternalAddress() (string, error) {
	// Если задан явный адрес — используем его.
	if s.cfg.Orchestrator.ExternalAddress != "" {
		return s.cfg.Orchestrator.ExternalAddress, nil
	}

	// Иначе пытаемся определить автоматически.
	addr, err := s.getAutoAddress()
	if err != nil {
		return "", err
	}

	return addr, nil
}

// getAutoAddress определяет адрес автоматически.
// Использует имя интерфейса из конфига или выбирает первый подходящий.
func (s *AnnouncementService) getAutoAddress() (string, error) {
	// Получаем список всех интерфейсов.
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("getAutoAddress: list interfaces: %w", err)
	}

	var preferred net.Interface
	ethName := s.cfg.Node.EthName

	for _, iface := range ifaces {
		// Пропускаем down-интерфейсы и loopback.
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// Если задан конкретный интерфейс — ищем его.
		if ethName != "" {
			if iface.Name == ethName {
				preferred = iface
				break
			}
			continue
		}

		// Иначе берём первый подходящий.
		if preferred.Name == "" {
			preferred = iface
		}
	}

	if preferred.Name == "" {
		return "", fmt.Errorf("getAutoAddress: no suitable network interface found")
	}

	// Получаем адреса интерфейса.
	addrs, err := preferred.Addrs()
	if err != nil {
		return "", fmt.Errorf("getAutoAddress: get addresses for %s: %w", preferred.Name, err)
	}

	for _, addr := range addrs {
		// Ищем IPv4 адрес.
		if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
			return fmt.Sprintf("%s:%d", ipnet.IP.String(), s.cfg.GRPC.Port), nil
		}
	}

	return "", fmt.Errorf("getAutoAddress: no IPv4 address found on interface %s", preferred.Name)
}
