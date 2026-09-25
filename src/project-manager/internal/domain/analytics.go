package domain

import (
	"context"
)

// AnalyticsFilter параметры фильтрации для получения аналитики.
type AnalyticsFilter struct {
	DateFrom          string // YYYY-MM-DD
	DateTo            string // YYYY-MM-DD
	AggregationPeriod string // Day, Week, Month
}

// AnalyticsSummary сводные ключевые показатели эффективности (KPI) игры за период.
type AnalyticsSummary struct {
	UniquePlayers           int64   `json:"unique_players"`
	TotalRevenue            float64 `json:"total_revenue"`
	TotalSessions           int64   `json:"total_sessions"`
	AvgSessionMinutes       float64 `json:"avg_session_minutes"`
	D1RetentionRate         float64 `json:"d1_retention_rate"`
	OverallCTR              float64 `json:"overall_ctr"`
	ARPU                    float64 `json:"arpu"`
	ARPPU                   float64 `json:"arppu"`
	PayingUsersCount        int64   `json:"paying_users_count"`
	PayingUsersPercent      float64 `json:"paying_users_percent"`
	TotalPurchases          int64   `json:"total_purchases"`
	AvgOrderValue           float64 `json:"avg_order_value"`
	TotalAdImpressions      int64   `json:"total_ad_impressions"`
	TotalPromoImpressions   int64   `json:"total_promo_impressions"`
	TotalPromoClicks        int64   `json:"total_promo_clicks"`
	LTV                     float64 `json:"ltv"`
	AvgHoursToFirstPurchase float64 `json:"avg_hours_to_first_purchase"`
	RPRConversionRate1To2   float64 `json:"rpr_conversion_rate1_to2"`
}

// RevenuePoint точка графика выручки по дате.
type RevenuePoint struct {
	Date           string  `json:"date"`
	TotalRevenue   float64 `json:"total_revenue"`
	PurchasesCount int64   `json:"purchases_count"`
}

// PlayerPoint точка графика активных игроков (DAU) по дате.
type PlayerPoint struct {
	Date               string `json:"date"`
	UniquePlayersCount int64  `json:"unique_players_count"`
}

// SessionPoint точка графика игровых сессий и их длительности.
type SessionPoint struct {
	Date               string  `json:"date"`
	TotalSessions      int64   `json:"total_sessions"`
	AvgDurationMinutes float64 `json:"avg_duration_minutes"`
	SessionsPerUser    float64 `json:"sessions_per_user"`
}

// CohortRetentionPoint точка когортного удержания игроков.
type CohortRetentionPoint struct {
	Date string  `json:"date"`
	D1   float64 `json:"d1"`
	D3   float64 `json:"d3"`
	D7   float64 `json:"d7"`
	D30  float64 `json:"d30"`
}

// ChurnPoint точка оттока игроков по дате.
type ChurnPoint struct {
	Date      string  `json:"date"`
	ChurnRate float64 `json:"churn_rate"`
}

// AdImpressionPoint показы рекламы по типам за дату.
type AdImpressionPoint struct {
	Date         string `json:"date"`
	Interstitial int64  `json:"interstitial"`
	Rewarded     int64  `json:"rewarded"`
	Banner       int64  `json:"banner"`
	Total        int64  `json:"total"`
}

// PromoPoint показы, клики и CTR промо-материалов за дату.
type PromoPoint struct {
	Date        string  `json:"date"`
	Impressions int64   `json:"impressions"`
	Clicks      int64   `json:"clicks"`
	CTR         float64 `json:"ctr"`
}

// PromoPlacementPoint статистика промо-материалов по конкретному плейсменту.
type PromoPlacementPoint struct {
	Placement     string  `json:"placement"`
	PlacementName string  `json:"placement_name"`
	Impressions   int64   `json:"impressions"`
	Clicks        int64   `json:"clicks"`
	CTR           float64 `json:"ctr"`
}

// RepeatPurchaseSegment сегмент повторных покупок.
type RepeatPurchaseSegment struct {
	SegmentName string  `json:"segment_name"`
	UsersCount  int64   `json:"users_count"`
	Percentage  float64 `json:"percentage"`
}

// UserTypeRevenue выручка в разрезе новых и возвращающихся пользователей.
type UserTypeRevenue struct {
	NewUsersRevenue       float64 `json:"new_users_revenue"`
	NewUsersCount         int64   `json:"new_users_count"`
	ReturningUsersRevenue float64 `json:"returning_users_revenue"`
	ReturningUsersCount   int64   `json:"returning_users_count"`
}

// LtvPoint когортный LTV по дате.
type LtvPoint struct {
	Date         string  `json:"date"`
	CohortSize   int64   `json:"cohort_size"`
	TotalRevenue float64 `json:"total_revenue"`
	LTV          float64 `json:"ltv"`
}

// GameAnalytics полный набор аналитических данных по игре за период.
type GameAnalytics struct {
	ProjectID       int64                   `json:"project_id"`
	DateFrom        string                  `json:"date_from"`
	DateTo          string                  `json:"date_to"`
	Summary         AnalyticsSummary        `json:"summary"`
	RevenueItems    []RevenuePoint          `json:"revenue_items"`
	DAUItems        []PlayerPoint           `json:"dau_items"`
	SessionItems    []SessionPoint          `json:"session_items"`
	RetentionItems  []CohortRetentionPoint  `json:"retention_items"`
	ChurnItems      []ChurnPoint            `json:"churn_items"`
	AdItems         []AdImpressionPoint     `json:"ad_items"`
	PromoItems      []PromoPoint            `json:"promo_items"`
	PromoPlacements []PromoPlacementPoint   `json:"promo_placements"`
	RPRSegments     []RepeatPurchaseSegment `json:"rpr_segments"`
	UserTypeData    UserTypeRevenue         `json:"user_type_data"`
	LTVItems        []LtvPoint              `json:"ltv_items"`
}

// AnalyticsClient определяет контракт получения игровой аналитики от внешнего сервиса или мок-драйвера.
type AnalyticsClient interface {
	// GetGameAnalytics возвращает детальную аналитику по игре за указанный диапазон дат.
	GetGameAnalytics(ctx context.Context, gameID int64, filter AnalyticsFilter) (*GameAnalytics, error)
}
