package client

import (
	"context"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
)

// DisabledPurchaseClient используется, когда интеграция с сервисом покупок выключена (driver: disabled).
// Чтение возвращает пустой каталог, изменение — domain.ErrIntegrationDisabled.
// В отличие от заглушки, не создаёт видимость работающего каталога товаров.
type DisabledPurchaseClient struct{}

// ListItems возвращает пустой список: товаров на этой установке нет.
func (DisabledPurchaseClient) ListItems(context.Context, int64) ([]*domain.GameItem, error) {
	return []*domain.GameItem{}, nil
}

// GetItem всегда возвращает domain.ErrNotFound.
func (DisabledPurchaseClient) GetItem(context.Context, int64, string) (*domain.GameItem, error) {
	return nil, domain.ErrNotFound
}

// CreateItem возвращает domain.ErrIntegrationDisabled.
func (DisabledPurchaseClient) CreateItem(context.Context, *domain.GameItem) (*domain.GameItem, error) {
	return nil, domain.ErrIntegrationDisabled
}

// UpdateItem возвращает domain.ErrIntegrationDisabled.
func (DisabledPurchaseClient) UpdateItem(context.Context, *domain.GameItem) (*domain.GameItem, error) {
	return nil, domain.ErrIntegrationDisabled
}

// DeleteItem возвращает domain.ErrIntegrationDisabled.
func (DisabledPurchaseClient) DeleteItem(context.Context, int64, string) error {
	return domain.ErrIntegrationDisabled
}

// DisabledAnalyticsClient используется, когда интеграция с сервисом аналитики выключена (driver: disabled).
// Не подменяет реальную статистику сгенерированной.
type DisabledAnalyticsClient struct{}

// GetGameAnalytics возвращает domain.ErrIntegrationDisabled.
func (DisabledAnalyticsClient) GetGameAnalytics(context.Context, int64, domain.AnalyticsFilter) (*domain.GameAnalytics, error) {
	return nil, domain.ErrIntegrationDisabled
}

var (
	_ domain.PurchaseClient  = DisabledPurchaseClient{}
	_ domain.AnalyticsClient = DisabledAnalyticsClient{}
)
