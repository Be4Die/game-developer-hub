package service

import (
	"context"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/game-server-node/internal/config"
	"github.com/Be4Die/game-developer-hub/game-server-node/internal/domain"
	"github.com/Be4Die/game-developer-hub/game-server-node/internal/storage/memory"
)

// stubSysProvider — заглушка для системных метрик.
// Позволяет нам возвращать фиксированные данные для тестов.
type stubSysProvider struct {
	maxToReturn   domain.ResourcesMax
	usageToReturn domain.ResourcesUsage
}

func (s *stubSysProvider) GetMax() (domain.ResourcesMax, error) {
	return s.maxToReturn, nil
}

func (s *stubSysProvider) GetUsage() (domain.ResourcesUsage, error) {
	return s.usageToReturn, nil
}

func TestDiscoveryService_Heartbeat(t *testing.T) {
	// 1. Arrange (Подготовка)
	ctx := context.Background()

	// Настраиваем in-memory хранилище и заполняем его тестовыми инстансами.
	// Наша цель — проверить, что Heartbeat посчитает ТОЛЬКО активные (id 1 и 2).
	storage := memory.NewMemoryInstanceStorage()
	_ = storage.RecordInstance(ctx, domain.Instance{ID: 1, Status: domain.InstanceStatusRunning})
	_ = storage.RecordInstance(ctx, domain.Instance{ID: 2, Status: domain.InstanceStatusStarting})
	_ = storage.RecordInstance(ctx, domain.Instance{ID: 3, Status: domain.InstanceStatusStopped})
	_ = storage.RecordInstance(ctx, domain.Instance{ID: 4, Status: domain.InstanceStatusCrashed})

	// Настраиваем заглушку для sysinfo (как будто процессор загружен на 42.5%)
	mockSys := &stubSysProvider{
		usageToReturn: domain.ResourcesUsage{
			CPU:     42.5,
			Memory:  1024 * 1024 * 500, // 500 MB
			Network: 1000,
		},
	}

	// Создаем фейковый конфиг
	cfg := &config.Config{
		Node: config.NodeConfig{
			Region:  "test-region",
			Version: "1.0.0",
		},
	}

	// Собираем сервис вручную, подменяя sysProvider (переопределяем после New)
	svc := NewDiscoveryService(storage, nil, cfg)
	svc.sysProvider = mockSys // Внедряем нашу заглушку

	// 2. Act (Действие)
	result, err := svc.Heartbeat(ctx)

	// 3. Assert (Проверка)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Проверяем бизнес-логику: должно быть ровно 2 активных инстанса (Running и Starting)
	if result.ActiveInstanceCount != 2 {
		t.Errorf("expected 2 active instances, got %d", result.ActiveInstanceCount)
	}

	// Проверяем, что метрики пробросились корректно
	if result.Usage.CPU != 42.5 {
		t.Errorf("expected CPU 42.5, got %f", result.Usage.CPU)
	}
}

func TestDiscoveryService_GetNode(t *testing.T) {
	// 1. Arrange
	mockSys := &stubSysProvider{
		maxToReturn: domain.ResourcesMax{
			CPUCores: 8,
		},
	}

	cfg := &config.Config{
		Node: config.NodeConfig{
			Region:  "eu-central",
			Version: "v1.2.3",
		},
	}

	svc := NewDiscoveryService(nil, nil, cfg)
	svc.sysProvider = mockSys
	// Фиксируем время для точности проверки
	testTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	svc.startedAt = testTime

	// 2. Act
	node, err := svc.GetNode()

	// 3. Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Проверяем склейку данных из конфига и из sysinfo
	if node.Region != "eu-central" {
		t.Errorf("expected region eu-central, got %s", node.Region)
	}

	if node.Version != "v1.2.3" {
		t.Errorf("expected version v1.2.3, got %s", node.Version)
	}

	if node.Resources.CPUCores != 8 {
		t.Errorf("expected 8 CPU cores, got %d", node.Resources.CPUCores)
	}

	if !node.StartedAt.Equal(testTime) {
		t.Errorf("expected time %v, got %v", testTime, node.StartedAt)
	}
}
