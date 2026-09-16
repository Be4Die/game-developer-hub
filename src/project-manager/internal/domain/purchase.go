package domain

import (
	"context"
	"time"
)

// GameItem представляет внутриигровой товар (IAP).
type GameItem struct {
	ID          int64     `json:"id"`
	ProjectID   int64     `json:"project_id"`
	GameItemID  string    `json:"game_item_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ImageURL    string    `json:"image_url"`
	PriceCoins  int64     `json:"price_coins"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PurchaseClient определяет контракт взаимодействия с каталогом товаров PurchaseService.
type PurchaseClient interface {
	// ListItems возвращает список товаров для игры.
	ListItems(ctx context.Context, gameID int64) ([]*GameItem, error)
	// GetItem возвращает товар по его внутриигровому идентификатору.
	GetItem(ctx context.Context, gameID int64, gameItemID string) (*GameItem, error)
	// CreateItem создает новый товар.
	CreateItem(ctx context.Context, item *GameItem) (*GameItem, error)
	// UpdateItem обновляет существующий товар.
	UpdateItem(ctx context.Context, item *GameItem) (*GameItem, error)
	// DeleteItem удаляет или деактивирует товар.
	DeleteItem(ctx context.Context, gameID int64, gameItemID string) error
}
