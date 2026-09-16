package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
)

// StubPurchaseClient реализует in-memory заглушку для работы с каталогом товаров
// без необходимости подключения к PurchaseService в локальной среде разработки.
// Формирует и логирует исходящие HTTP-запросы без их сетевой отправки.
type StubPurchaseClient struct {
	mu     sync.RWMutex
	items  map[int64]map[string]*domain.GameItem // gameID -> gameItemID -> GameItem
	nextID int64
	log    *slog.Logger
}

// NewStubPurchaseClient создаёт новый экземпляр StubPurchaseClient.
func NewStubPurchaseClient(log *slog.Logger) *StubPurchaseClient {
	if log == nil {
		log = slog.Default()
	}
	return &StubPurchaseClient{
		items: make(map[int64]map[string]*domain.GameItem),
		log:   log,
	}
}

// ListItems возвращает список товаров для игры.
func (s *StubPurchaseClient) ListItems(_ context.Context, gameID int64) ([]*domain.GameItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	s.log.Debug("[StubPurchaseClient] Simulated request",
		slog.String("method", "GET"),
		slog.String("url", fmt.Sprintf("/purchase-service/games/%d/items", gameID)),
	)

	gameMap, ok := s.items[gameID]
	if !ok {
		return []*domain.GameItem{}, nil
	}

	result := make([]*domain.GameItem, 0, len(gameMap))
	for _, item := range gameMap {
		copied := *item
		result = append(result, &copied)
	}
	return result, nil
}

// GetItem возвращает товар по его внутриигровому идентификатору.
func (s *StubPurchaseClient) GetItem(_ context.Context, gameID int64, gameItemID string) (*domain.GameItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	s.log.Debug("[StubPurchaseClient] Simulated request",
		slog.String("method", "GET"),
		slog.String("url", fmt.Sprintf("/purchase-service/games/%d/items/%s", gameID, gameItemID)),
	)

	gameMap, ok := s.items[gameID]
	if !ok {
		return nil, domain.ErrNotFound
	}

	item, ok := gameMap[gameItemID]
	if !ok {
		return nil, domain.ErrNotFound
	}

	copied := *item
	return &copied, nil
}

// CreateItem создает новый товар.
func (s *StubPurchaseClient) CreateItem(_ context.Context, item *domain.GameItem) (*domain.GameItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Формируем payload, идентичный схеме item_prices в PurchaseService
	payload := map[string]any{
		"game_id":             item.ProjectID,
		"game_item_id":        item.GameItemID,
		"name":                item.Name,
		"description":         item.Description,
		"image_url":           item.ImageURL,
		"price_per_unit_coins": item.PriceCoins,
		"is_active":           item.IsActive,
	}
	bodyJSON, _ := json.Marshal(payload)

	s.log.Info("[StubPurchaseClient] Formed simulated HTTP request (not sent to network)",
		slog.String("method", "POST"),
		slog.String("url", fmt.Sprintf("/purchase-service/games/%d/items", item.ProjectID)),
		slog.String("payload", string(bodyJSON)),
	)

	gameMap, ok := s.items[item.ProjectID]
	if !ok {
		gameMap = make(map[string]*domain.GameItem)
		s.items[item.ProjectID] = gameMap
	}

	if _, exists := gameMap[item.GameItemID]; exists {
		return nil, domain.ErrAlreadyExists
	}

	id := atomic.AddInt64(&s.nextID, 1)
	now := time.Now().UTC()

	newItem := &domain.GameItem{
		ID:          id,
		ProjectID:   item.ProjectID,
		GameItemID:  item.GameItemID,
		Name:        item.Name,
		Description: item.Description,
		ImageURL:    item.ImageURL,
		PriceCoins:  item.PriceCoins,
		IsActive:    item.IsActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	gameMap[item.GameItemID] = newItem
	copied := *newItem
	return &copied, nil
}

// UpdateItem обновляет существующий товар.
func (s *StubPurchaseClient) UpdateItem(_ context.Context, item *domain.GameItem) (*domain.GameItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	payload := map[string]any{
		"name":                item.Name,
		"description":         item.Description,
		"image_url":           item.ImageURL,
		"price_per_unit_coins": item.PriceCoins,
		"is_active":           item.IsActive,
	}
	bodyJSON, _ := json.Marshal(payload)

	s.log.Info("[StubPurchaseClient] Formed simulated HTTP request (not sent to network)",
		slog.String("method", "PUT"),
		slog.String("url", fmt.Sprintf("/purchase-service/games/%d/items/%s", item.ProjectID, item.GameItemID)),
		slog.String("payload", string(bodyJSON)),
	)

	gameMap, ok := s.items[item.ProjectID]
	if !ok {
		return nil, domain.ErrNotFound
	}

	existing, exists := gameMap[item.GameItemID]
	if !exists {
		return nil, domain.ErrNotFound
	}

	existing.Name = item.Name
	existing.Description = item.Description
	if item.ImageURL != "" {
		existing.ImageURL = item.ImageURL
	}
	existing.PriceCoins = item.PriceCoins
	existing.IsActive = item.IsActive
	existing.UpdatedAt = time.Now().UTC()

	copied := *existing
	return &copied, nil
}

// DeleteItem удаляет или деактивирует товар.
func (s *StubPurchaseClient) DeleteItem(_ context.Context, gameID int64, gameItemID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.log.Info("[StubPurchaseClient] Formed simulated HTTP request (not sent to network)",
		slog.String("method", "DELETE"),
		slog.String("url", fmt.Sprintf("/purchase-service/games/%d/items/%s", gameID, gameItemID)),
	)

	gameMap, ok := s.items[gameID]
	if !ok {
		return domain.ErrNotFound
	}

	if _, exists := gameMap[gameItemID]; !exists {
		return domain.ErrNotFound
	}

	delete(gameMap, gameItemID)
	return nil
}
