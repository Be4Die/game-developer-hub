package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
)

// HTTPPurchaseClient реализует domain.PurchaseClient поверх удаленного HTTP REST API PurchaseService.
type HTTPPurchaseClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewHTTPPurchaseClient создает экземпляр HTTPPurchaseClient.
func NewHTTPPurchaseClient(baseURL string, timeout time.Duration) *HTTPPurchaseClient {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &HTTPPurchaseClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// dtoItemResponse структура ответа PurchaseService на получение информации о предмете.
type dtoItemResponse struct {
	ID                int64     `json:"id"`
	GameID            int64     `json:"gameId"`
	GameItemID        string    `json:"gameItemId"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	ImageURL          string    `json:"imageUrl"`
	PricePerUnitCoins int64     `json:"pricePerUnitCoins"`
	IsActive          bool      `json:"isActive"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

func (dto *dtoItemResponse) toDomain() *domain.GameItem {
	return &domain.GameItem{
		ID:          dto.ID,
		ProjectID:   dto.GameID,
		GameItemID:  dto.GameItemID,
		Name:        dto.Name,
		Description: dto.Description,
		ImageURL:    dto.ImageURL,
		PriceCoins:  dto.PricePerUnitCoins,
		IsActive:    dto.IsActive,
		CreatedAt:   dto.CreatedAt,
		UpdatedAt:   dto.UpdatedAt,
	}
}

// ListItems запрашивает список товаров для игры.
func (c *HTTPPurchaseClient) ListItems(ctx context.Context, gameID int64) ([]*domain.GameItem, error) {
	url := fmt.Sprintf("%s/purchase-service/games/%d/items", c.baseURL, gameID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create list items request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute list items request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("purchase-service returned status %d: %s", resp.StatusCode, string(body))
	}

	var dtos []dtoItemResponse
	if err := json.NewDecoder(resp.Body).Decode(&dtos); err != nil {
		return nil, fmt.Errorf("decode items response: %w", err)
	}

	items := make([]*domain.GameItem, len(dtos))
	for i, d := range dtos {
		items[i] = d.toDomain()
	}
	return items, nil
}

// GetItem запрашивает конкретный товар.
func (c *HTTPPurchaseClient) GetItem(ctx context.Context, gameID int64, gameItemID string) (*domain.GameItem, error) {
	url := fmt.Sprintf("%s/purchase-service/games/%d/items/%s", c.baseURL, gameID, gameItemID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create get item request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute get item request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, domain.ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("purchase-service returned status %d: %s", resp.StatusCode, string(body))
	}

	var dto dtoItemResponse
	if err := json.NewDecoder(resp.Body).Decode(&dto); err != nil {
		return nil, fmt.Errorf("decode item response: %w", err)
	}
	return dto.toDomain(), nil
}

// CreateItem отправляет запрос на создание товара.
func (c *HTTPPurchaseClient) CreateItem(ctx context.Context, item *domain.GameItem) (*domain.GameItem, error) {
	url := fmt.Sprintf("%s/purchase-service/games/%d/items", c.baseURL, item.ProjectID)

	payload := map[string]any{
		"gameId":            item.ProjectID,
		"gameItemId":        item.GameItemID,
		"name":              item.Name,
		"description":       item.Description,
		"imageUrl":          item.ImageURL,
		"pricePerUnitCoins": item.PriceCoins,
		"isActive":          item.IsActive,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal item payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create item request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute create item request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return nil, domain.ErrAlreadyExists
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("purchase-service returned status %d: %s", resp.StatusCode, string(body))
	}

	var dto dtoItemResponse
	if err := json.NewDecoder(resp.Body).Decode(&dto); err != nil {
		return nil, fmt.Errorf("decode create item response: %w", err)
	}
	return dto.toDomain(), nil
}

// UpdateItem отправляет запрос на обновление товара.
func (c *HTTPPurchaseClient) UpdateItem(ctx context.Context, item *domain.GameItem) (*domain.GameItem, error) {
	url := fmt.Sprintf("%s/purchase-service/games/%d/items/%s", c.baseURL, item.ProjectID, item.GameItemID)

	payload := map[string]any{
		"name":              item.Name,
		"description":       item.Description,
		"imageUrl":          item.ImageURL,
		"pricePerUnitCoins": item.PriceCoins,
		"isActive":          item.IsActive,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal update item payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create update item request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute update item request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, domain.ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("purchase-service returned status %d: %s", resp.StatusCode, string(body))
	}

	var dto dtoItemResponse
	if err := json.NewDecoder(resp.Body).Decode(&dto); err != nil {
		return nil, fmt.Errorf("decode update item response: %w", err)
	}
	return dto.toDomain(), nil
}

// DeleteItem отправляет запрос на удаление/деактивацию товара.
func (c *HTTPPurchaseClient) DeleteItem(ctx context.Context, gameID int64, gameItemID string) error {
	url := fmt.Sprintf("%s/purchase-service/games/%d/items/%s", c.baseURL, gameID, gameItemID)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("create delete item request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute delete item request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return domain.ErrNotFound
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("purchase-service returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
