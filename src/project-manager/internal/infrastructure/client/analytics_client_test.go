package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
	"github.com/Be4Die/game-developer-hub/project-manager/internal/infrastructure/client"
	"github.com/stretchr/testify/require"
)

func TestStubAnalyticsClient_GetGameAnalytics(t *testing.T) {
	ctx := context.Background()
	stub := client.NewStubAnalyticsClient(nil)

	const gameID = int64(42)

	// 1. Запрос аналитики за 7 дней
	filter := domain.AnalyticsFilter{
		DateFrom:          "2026-06-01",
		DateTo:            "2026-06-07",
		AggregationPeriod: "Day",
	}

	analytics, err := stub.GetGameAnalytics(ctx, gameID, filter)
	require.NoError(t, err)
	require.NotNil(t, analytics)
	require.Equal(t, gameID, analytics.ProjectID)
	require.Equal(t, "2026-06-01", analytics.DateFrom)
	require.Equal(t, "2026-06-07", analytics.DateTo)

	// Проверяем заполнение графиков (7 дней)
	require.Len(t, analytics.RevenueItems, 7)
	require.Len(t, analytics.DAUItems, 7)
	require.Len(t, analytics.SessionItems, 7)
	require.Len(t, analytics.RetentionItems, 7)
	require.Len(t, analytics.ChurnItems, 7)
	require.Len(t, analytics.AdItems, 7)
	require.Len(t, analytics.PromoItems, 7)
	require.Len(t, analytics.LTVItems, 7)

	// Проверяем сводные показатели
	require.Greater(t, analytics.Summary.UniquePlayers, int64(0))
	require.Greater(t, analytics.Summary.TotalRevenue, 0.0)
	require.Greater(t, analytics.Summary.TotalSessions, int64(0))
	require.Greater(t, analytics.Summary.AvgSessionMinutes, 0.0)
	require.Greater(t, analytics.Summary.D1RetentionRate, 0.0)
	require.NotEmpty(t, analytics.PromoPlacements)
	require.NotEmpty(t, analytics.RPRSegments)

	// 2. Детерминированность: повторный вызов с теми же параметрами возвращает те же данные
	repeat, err := stub.GetGameAnalytics(ctx, gameID, filter)
	require.NoError(t, err)
	require.Equal(t, analytics.Summary.TotalRevenue, repeat.Summary.TotalRevenue)
	require.Equal(t, analytics.Summary.UniquePlayers, repeat.Summary.UniquePlayers)
}

func TestHTTPAnalyticsClient_GetGameAnalytics(t *testing.T) {
	ctx := context.Background()

	// Создаем тестовый HTTP-сервер, эмулирующий ответы PlayerGameAnalytics
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/player-activity/analytics/active-users":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"items": []map[string]any{
					{"date": "2026-06-01", "uniquePlayersCount": 150},
					{"date": "2026-06-02", "uniquePlayersCount": 160},
				},
			})
		case "/purchases/metrics/revenue-range":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"items": []map[string]any{
					{"date": "2026-06-01", "totalRevenue": 15000.0, "purchasesCount": 15},
					{"date": "2026-06-02", "totalRevenue": 20000.0, "purchasesCount": 20},
				},
			})
		case "/purchases/metrics/arpu-arppu":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"totalRevenue":       35000.0,
				"arpu":               233.33,
				"arppu":              1000.0,
				"payingUsersPercent": 23.3,
				"payingUsersCount":   35,
				"totalActiveUsers":   150,
				"avgOrderValue":      1000.0,
				"totalPurchases":     35,
			})
		case "/player-activity/analytics/session-avg-duration-range":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"items": []map[string]any{
					{"date": "2026-06-01", "avgDurationMilliseconds": 720000.0},
					{"date": "2026-06-02", "avgDurationMilliseconds": 840000.0},
				},
			})
		case "/player-activity/analytics/game-sessions-count-per-user-range":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"items": []map[string]any{
					{"date": "2026-06-01", "gameSessionsCount": 300, "gameSessionsPerUser": 2.0},
					{"date": "2026-06-02", "gameSessionsCount": 320, "gameSessionsPerUser": 2.0},
				},
			})
		case "/player-activity/analytics/cohort-retention-range":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"items": []map[string]any{
					{"analysisDate": "2026-06-01", "retentionRate": 35.5},
					{"analysisDate": "2026-06-02", "retentionRate": 40.0},
				},
			})
		case "/player-activity/analytics/churn-range":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"items": []map[string]any{
					{"date": "2026-06-01", "churnRate": 5.2},
					{"date": "2026-06-02", "churnRate": 4.8},
				},
			})
		case "/ads/distribution":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"dates": []map[string]any{
					{
						"date": "2026-06-01",
						"games": []map[string]any{
							{
								"gameId": 42,
								"adTypes": []map[string]any{
									{"adType": "Interstitial", "impressionsCount": 100},
									{"adType": "Rewarded", "impressionsCount": 50},
								},
							},
						},
					},
				},
			})
		case "/promo-ctr/ctr-range":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"totalImpressions": 1000,
				"totalClicks":      80,
				"overallCtr":       8.0,
				"items": []map[string]any{
					{"date": "2026-06-01", "impressions": 500, "clicks": 40, "ctr": 8.0},
					{"date": "2026-06-02", "impressions": 500, "clicks": 40, "ctr": 8.0},
				},
				"byPlacement": []map[string]any{
					{"placement": "MainPageCatalog", "placementName": "Каталог", "impressions": 800, "clicks": 64, "ctr": 8.0},
				},
			})
		default:
			_ = json.NewEncoder(w).Encode(map[string]any{})
		}
	}))
	defer server.Close()

	httpClient := client.NewHTTPAnalyticsClient(server.URL, "test-api-key", 2*time.Second, nil)
	analytics, err := httpClient.GetGameAnalytics(ctx, 42, domain.AnalyticsFilter{
		DateFrom: "2026-06-01",
		DateTo:   "2026-06-02",
	})

	require.NoError(t, err)
	require.NotNil(t, analytics)
	require.Equal(t, int64(42), analytics.ProjectID)
	require.Equal(t, 35000.0, analytics.Summary.TotalRevenue)
	require.Equal(t, int64(150), analytics.Summary.UniquePlayers)
	require.Equal(t, 8.0, analytics.Summary.OverallCTR)
	require.Len(t, analytics.RevenueItems, 2)
	require.Len(t, analytics.DAUItems, 2)
	require.Len(t, analytics.SessionItems, 2)
	require.Equal(t, 12.0, analytics.SessionItems[0].AvgDurationMinutes) // 720000 ms = 12 min
	require.NotEmpty(t, analytics.PromoPlacements)
	require.Equal(t, "MainPageCatalog", analytics.PromoPlacements[0].Placement)
}
