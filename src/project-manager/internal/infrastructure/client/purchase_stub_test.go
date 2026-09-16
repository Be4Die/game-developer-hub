package client_test

import (
	"context"
	"testing"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
	"github.com/Be4Die/game-developer-hub/project-manager/internal/infrastructure/client"
	"github.com/stretchr/testify/require"
)

func TestStubPurchaseClient_CRUD(t *testing.T) {
	ctx := context.Background()
	stub := client.NewStubPurchaseClient(nil)

	const gameID = int64(42)

	// 1. Сначала список пуст
	items, err := stub.ListItems(ctx, gameID)
	require.NoError(t, err)
	require.Empty(t, items)

	// 2. Создание товара
	created, err := stub.CreateItem(ctx, &domain.GameItem{
		ProjectID:   gameID,
		GameItemID:  "gold_100",
		Name:        "100 золотых монет",
		Description: "Базовый сундук",
		ImageURL:    "games/42/items/gold_100.png",
		PriceCoins:  50,
		IsActive:    true,
	})
	require.NoError(t, err)
	require.Equal(t, "gold_100", created.GameItemID)
	require.Equal(t, int64(50), created.PriceCoins)
	require.True(t, created.IsActive)

	// 3. Дубликат ID выдает ErrAlreadyExists
	_, err = stub.CreateItem(ctx, &domain.GameItem{
		ProjectID:  gameID,
		GameItemID: "gold_100",
		Name:       "Дубликат",
		PriceCoins: 100,
	})
	require.ErrorIs(t, err, domain.ErrAlreadyExists)

	// 4. Получение по ID
	item, err := stub.GetItem(ctx, gameID, "gold_100")
	require.NoError(t, err)
	require.Equal(t, "100 золотых монет", item.Name)

	// 5. Несуществующий товар
	_, err = stub.GetItem(ctx, gameID, "non_existent")
	require.ErrorIs(t, err, domain.ErrNotFound)

	// 6. Обновление товара
	item.Name = "100 золота (со скидкой)"
	item.PriceCoins = 40
	updated, err := stub.UpdateItem(ctx, item)
	require.NoError(t, err)
	require.Equal(t, "100 золота (со скидкой)", updated.Name)
	require.Equal(t, int64(40), updated.PriceCoins)

	// 7. Список теперь содержит 1 товар
	items, err = stub.ListItems(ctx, gameID)
	require.NoError(t, err)
	require.Len(t, items, 1)

	// 8. Удаление товара
	err = stub.DeleteItem(ctx, gameID, "gold_100")
	require.NoError(t, err)

	// 9. После удаления товар не найден
	_, err = stub.GetItem(ctx, gameID, "gold_100")
	require.ErrorIs(t, err, domain.ErrNotFound)
}
