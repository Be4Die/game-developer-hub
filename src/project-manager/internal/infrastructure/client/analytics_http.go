package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
)

// HTTPAnalyticsClient реализует domain.AnalyticsClient поверх HTTP API шлюза площадки или сервиса PlayerGameAnalytics.
type HTTPAnalyticsClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	log        *slog.Logger
}

// NewHTTPAnalyticsClient создает экземпляр HTTPAnalyticsClient.
func NewHTTPAnalyticsClient(baseURL, apiKey string, timeout time.Duration, log *slog.Logger) *HTTPAnalyticsClient {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	if log == nil {
		log = slog.Default()
	}
	return &HTTPAnalyticsClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		log: log,
	}
}

// DTO-структуры ответов сервиса аналитики

type dtoActiveUsersItem struct {
	Date               string `json:"date"`
	UniquePlayersCount int64  `json:"uniquePlayersCount"`
}

type dtoActiveUsersResponse struct {
	Items []dtoActiveUsersItem `json:"items"`
}

type dtoRevenueItem struct {
	Date           string  `json:"date"`
	TotalRevenue   float64 `json:"totalRevenue"`
	PurchasesCount int64   `json:"purchasesCount"`
}

type dtoRevenueResponse struct {
	Items []dtoRevenueItem `json:"items"`
}

type dtoArpuArppuResponse struct {
	TotalRevenue              float64 `json:"totalRevenue"`
	Arpu                      float64 `json:"arpu"`
	Arppu                     float64 `json:"arppu"`
	PayingUsersPercent        float64 `json:"payingUsersPercent"`
	PayingUsersCount          int64   `json:"payingUsersCount"`
	TotalActiveUsers          int64   `json:"totalActiveUsers"`
	AvgOrderValue             float64 `json:"avgOrderValue"`
	AvgPurchasesPerUser       float64 `json:"avgPurchasesPerUser"`
	AvgPurchasesPerPayingUser float64 `json:"avgPurchasesPerPayingUser"`
	TotalPurchases            int64   `json:"totalPurchases"`
}

type dtoSessionDurationItem struct {
	Date                     string  `json:"date"`
	AvgDurationMilliseconds float64 `json:"avgDurationMilliseconds"`
	DurationSeconds          string  `json:"durationSeconds"`
}

type dtoSessionDurationResponse struct {
	Items []dtoSessionDurationItem `json:"items"`
}

type dtoGameSessionsCountItem struct {
	Date                string  `json:"date"`
	GameSessionsCount   int64   `json:"gameSessionsCount"`
	GameSessionsPerUser float64 `json:"gameSessionsPerUser"`
}

type dtoGameSessionsCountResponse struct {
	Items []dtoGameSessionsCountItem `json:"items"`
}

type dtoRetentionItem struct {
	AnalysisDate  string  `json:"analysisDate"`
	Date          string  `json:"date"`
	RetentionRate float64 `json:"retentionRate"`
}

type dtoRetentionResponse struct {
	Items []dtoRetentionItem `json:"items"`
}

type dtoChurnItem struct {
	Date      string  `json:"date"`
	ChurnRate float64 `json:"churnRate"`
	ChurnRatio float64 `json:"churnRatio"`
}

type dtoChurnResponse struct {
	Items []dtoChurnItem `json:"items"`
}

type dtoAdGameType struct {
	AdType           string `json:"adType"`
	ImpressionsCount int64  `json:"impressionsCount"`
}

type dtoAdGame struct {
	GameID  int64           `json:"gameId"`
	AdTypes []dtoAdGameType `json:"adTypes"`
}

type dtoAdDistributionDateItem struct {
	Date  string      `json:"date"`
	Games []dtoAdGame `json:"games"`
}

type dtoAdDistributionResponse struct {
	Dates []dtoAdDistributionDateItem `json:"dates"`
}

type dtoPromoCtrItem struct {
	Date        string  `json:"date"`
	Impressions int64   `json:"impressions"`
	Clicks      int64   `json:"clicks"`
	Ctr         float64 `json:"ctr"`
}

type dtoPromoPlacementItem struct {
	Placement     any     `json:"placement"`
	PlacementName string  `json:"placementName"`
	Impressions   int64   `json:"impressions"`
	Clicks        int64   `json:"clicks"`
	Ctr           float64 `json:"ctr"`
}

type dtoPromoCtrResponse struct {
	TotalImpressions int64                   `json:"totalImpressions"`
	TotalClicks      int64                   `json:"totalClicks"`
	OverallCtr       float64                 `json:"overallCtr"`
	Items            []dtoPromoCtrItem       `json:"items"`
	ByPlacement      []dtoPromoPlacementItem `json:"byPlacement"`
}

type dtoRepeatPurchaseSegmentItem struct {
	SegmentName string  `json:"segmentName"`
	UsersCount  int64   `json:"usersCount"`
	Percentage  float64 `json:"percentage"`
}

type dtoRepeatPurchaseRateResponse struct {
	ConversionRate1To2 float64                        `json:"conversionRate1To2"`
	Segments           []dtoRepeatPurchaseSegmentItem `json:"segments"`
}

type dtoUserTypeItem struct {
	Revenue    float64 `json:"revenue"`
	UsersCount int64   `json:"usersCount"`
	AvgRevenue float64 `json:"avgRevenue"`
}

type dtoUserTypeResponse struct {
	NewUsers       dtoUserTypeItem `json:"newUsers"`
	ReturningUsers dtoUserTypeItem `json:"returningUsers"`
}

type dtoLtvResponse struct {
	Ltv float64 `json:"ltv"`
}

type dtoLtvRangeItem struct {
	CohortDate   string  `json:"cohortDate"`
	Date         string  `json:"date"`
	CohortSize   int64   `json:"cohortSize"`
	TotalRevenue float64 `json:"totalRevenue"`
	Ltv          float64 `json:"ltv"`
}

type dtoLtvRangeResponse struct {
	Items []dtoLtvRangeItem `json:"items"`
}

type dtoTtfpResponse struct {
	AvgHoursToFirstPurchase float64 `json:"avgHoursToFirstPurchase"`
}

// GetGameAnalytics запрашивает аналитику по игре у внешнего сервиса площадки.
func (c *HTTPAnalyticsClient) GetGameAnalytics(ctx context.Context, gameID int64, filter domain.AnalyticsFilter) (*domain.GameAnalytics, error) {
	dateFrom, dateTo := parseDateRange(filter.DateFrom, filter.DateTo)
	fromStr := dateFrom.Format("2006-01-02")
	toStr := dateTo.Format("2006-01-02")
	gameIDStr := strconv.FormatInt(gameID, 10)

	var (
		wg          sync.WaitGroup
		dauResp     dtoActiveUsersResponse
		revResp     dtoRevenueResponse
		arpuResp    dtoArpuArppuResponse
		durResp     dtoSessionDurationResponse
		sessResp    dtoGameSessionsCountResponse
		d1Resp      dtoRetentionResponse
		d3Resp      dtoRetentionResponse
		d7Resp      dtoRetentionResponse
		d30Resp     dtoRetentionResponse
		churnResp   dtoChurnResponse
		adResp      dtoAdDistributionResponse
		promoResp   dtoPromoCtrResponse
		rprResp     dtoRepeatPurchaseRateResponse
		userRevResp dtoUserTypeResponse
		ltvResp     dtoLtvResponse
		ltvRangeResp dtoLtvRangeResponse
		ttfpResp    dtoTtfpResponse
	)

	fetchJSON := func(endpoint string, params url.Values, target any) {
		defer wg.Done()
		reqURL := fmt.Sprintf("%s/%s?%s", c.baseURL, endpoint, params.Encode())
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
		if err != nil {
			c.log.Warn("failed to create http request for analytics", slog.String("url", reqURL), slog.String("error", err.Error()))
			return
		}
		if c.apiKey != "" {
			req.Header.Set("X-API-Key", c.apiKey)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			c.log.Warn("http request failed for analytics", slog.String("url", reqURL), slog.String("error", err.Error()))
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			c.log.Warn("analytics endpoint non-200 status", slog.String("url", reqURL), slog.Int("status", resp.StatusCode), slog.String("body", string(body)))
			return
		}

		if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
			c.log.Warn("failed to decode analytics json", slog.String("url", reqURL), slog.String("error", err.Error()))
		}
	}

	baseParams := url.Values{}
	baseParams.Set("dateFrom", fromStr)
	baseParams.Set("dateTo", toStr)
	baseParams.Set("gameId", gameIDStr)

	// 1. DAU
	dauParams := url.Values{}
	for k, v := range baseParams {
		dauParams[k] = v
	}
	dauParams.Set("aggregationPeriod", "Day")
	wg.Add(1)
	go fetchJSON("player-activity/analytics/active-users", dauParams, &dauResp)

	// 2. Revenue range
	wg.Add(1)
	go fetchJSON("purchases/metrics/revenue-range", baseParams, &revResp)

	// 3. ARPU / ARPPU
	wg.Add(1)
	go fetchJSON("purchases/metrics/arpu-arppu", baseParams, &arpuResp)

	// 4. Session duration
	durParams := url.Values{}
	for k, v := range baseParams {
		durParams[k] = v
	}
	durParams.Set("aggregationPeriod", "Day")
	wg.Add(1)
	go fetchJSON("player-activity/analytics/session-avg-duration-range", durParams, &durResp)

	// 5. Game sessions count
	sessParams := url.Values{}
	for k, v := range baseParams {
		sessParams[k] = v
	}
	sessParams.Set("aggregationPeriod", "Day")
	wg.Add(1)
	go fetchJSON("player-activity/analytics/game-sessions-count-per-user-range", sessParams, &sessResp)

	// 6. Cohort Retention D1, D3, D7, D30
	makeRetParams := func(days int) url.Values {
		p := url.Values{}
		for k, v := range baseParams {
			p[k] = v
		}
		p.Set("returnAfterDays", strconv.Itoa(days))
		return p
	}
	wg.Add(4)
	go fetchJSON("player-activity/analytics/cohort-retention-range", makeRetParams(1), &d1Resp)
	go fetchJSON("player-activity/analytics/cohort-retention-range", makeRetParams(3), &d3Resp)
	go fetchJSON("player-activity/analytics/cohort-retention-range", makeRetParams(7), &d7Resp)
	go fetchJSON("player-activity/analytics/cohort-retention-range", makeRetParams(30), &d30Resp)

	// 7. Churn
	churnParams := url.Values{}
	for k, v := range baseParams {
		churnParams[k] = v
	}
	churnParams.Set("inactiveDaysThreshold", "30")
	wg.Add(1)
	go fetchJSON("player-activity/analytics/churn-range", churnParams, &churnResp)

	// 8. Ads distribution
	adsParams := url.Values{}
	adsParams.Set("from", fromStr)
	adsParams.Set("to", toStr)
	adsParams.Set("gameId", gameIDStr)
	wg.Add(1)
	go fetchJSON("ads/distribution", adsParams, &adResp)

	// 9. Promo CTR
	wg.Add(1)
	go fetchJSON("promo-ctr/ctr-range", baseParams, &promoResp)

	// 10. Repeat purchase rate
	wg.Add(1)
	go fetchJSON("purchases/metrics/repeat-purchase-rate", baseParams, &rprResp)

	// 11. User type revenue
	wg.Add(1)
	go fetchJSON("purchases/metrics/revenue-by-user-type", baseParams, &userRevResp)

	// 12. LTV
	wg.Add(2)
	go fetchJSON("purchases/metrics/ltv", baseParams, &ltvResp)
	go fetchJSON("purchases/metrics/ltv-range", baseParams, &ltvRangeResp)

	// 13. Time to first purchase
	wg.Add(1)
	go fetchJSON("purchases/metrics/time-to-first-purchase", baseParams, &ttfpResp)

	wg.Wait()

	// Сборка доменной сущности
	dauPoints := make([]domain.PlayerPoint, len(dauResp.Items))
	for i, it := range dauResp.Items {
		dauPoints[i] = domain.PlayerPoint{
			Date:               it.Date,
			UniquePlayersCount: it.UniquePlayersCount,
		}
	}

	revenuePoints := make([]domain.RevenuePoint, len(revResp.Items))
	for i, it := range revResp.Items {
		revenuePoints[i] = domain.RevenuePoint{
			Date:           it.Date,
			TotalRevenue:   it.TotalRevenue,
			PurchasesCount: it.PurchasesCount,
		}
	}

	// Сессии
	durMap := make(map[string]float64)
	for _, it := range durResp.Items {
		durMin := it.AvgDurationMilliseconds / 60000.0
		if durMin <= 0 && it.DurationSeconds != "" {
			durMin = parseTimeSpanMinutes(it.DurationSeconds)
		}
		durMap[it.Date] = math.Round(durMin*100) / 100
	}

	sessionPoints := make([]domain.SessionPoint, len(sessResp.Items))
	for i, it := range sessResp.Items {
		sessionPoints[i] = domain.SessionPoint{
			Date:               it.Date,
			TotalSessions:      it.GameSessionsCount,
			AvgDurationMinutes: durMap[it.Date],
			SessionsPerUser:    it.GameSessionsPerUser,
		}
	}

	// Когортный ретеншн
	retMapD1 := make(map[string]float64)
	for _, it := range d1Resp.Items {
		d := it.AnalysisDate
		if d == "" {
			d = it.Date
		}
		retMapD1[d] = it.RetentionRate
	}
	retMapD3 := make(map[string]float64)
	for _, it := range d3Resp.Items {
		d := it.AnalysisDate
		if d == "" {
			d = it.Date
		}
		retMapD3[d] = it.RetentionRate
	}
	retMapD7 := make(map[string]float64)
	for _, it := range d7Resp.Items {
		d := it.AnalysisDate
		if d == "" {
			d = it.Date
		}
		retMapD7[d] = it.RetentionRate
	}
	retMapD30 := make(map[string]float64)
	for _, it := range d30Resp.Items {
		d := it.AnalysisDate
		if d == "" {
			d = it.Date
		}
		retMapD30[d] = it.RetentionRate
	}

	dateList := generateDateList(dateFrom, dateTo)
	retentionPoints := make([]domain.CohortRetentionPoint, len(dateList))
	for i, d := range dateList {
		retentionPoints[i] = domain.CohortRetentionPoint{
			Date: d,
			D1:   retMapD1[d],
			D3:   retMapD3[d],
			D7:   retMapD7[d],
			D30:  retMapD30[d],
		}
	}

	// Отток
	churnPoints := make([]domain.ChurnPoint, len(churnResp.Items))
	for i, it := range churnResp.Items {
		rate := it.ChurnRate
		if rate == 0 && it.ChurnRatio > 0 {
			rate = it.ChurnRatio
		}
		churnPoints[i] = domain.ChurnPoint{
			Date:      it.Date,
			ChurnRate: rate,
		}
	}

	// Реклама
	adPoints := make([]domain.AdImpressionPoint, len(adResp.Dates))
	var totalAdImpressions int64
	for i, it := range adResp.Dates {
		var inter, rew, ban int64
		for _, g := range it.Games {
			if g.GameID == gameID || g.GameID == 0 {
				for _, t := range g.AdTypes {
					switch strings.ToLower(t.AdType) {
					case "interstitial":
						inter += t.ImpressionsCount
					case "rewarded":
						rew += t.ImpressionsCount
					case "banner":
						ban += t.ImpressionsCount
					}
				}
			}
		}
		tot := inter + rew + ban
		totalAdImpressions += tot
		adPoints[i] = domain.AdImpressionPoint{
			Date:         it.Date,
			Interstitial: inter,
			Rewarded:     rew,
			Banner:       ban,
			Total:        tot,
		}
	}

	// Промо CTR
	promoPoints := make([]domain.PromoPoint, len(promoResp.Items))
	for i, it := range promoResp.Items {
		promoPoints[i] = domain.PromoPoint{
			Date:        it.Date,
			Impressions: it.Impressions,
			Clicks:      it.Clicks,
			CTR:         it.Ctr,
		}
	}
	promoPlacements := make([]domain.PromoPlacementPoint, len(promoResp.ByPlacement))
	for i, it := range promoResp.ByPlacement {
		plStr := fmt.Sprintf("%v", it.Placement)
		plName := it.PlacementName
		if plName == "" {
			plName = plStr
		}
		promoPlacements[i] = domain.PromoPlacementPoint{
			Placement:     plStr,
			PlacementName: plName,
			Impressions:   it.Impressions,
			Clicks:        it.Clicks,
			CTR:           it.Ctr,
		}
	}

	// Сегменты повторных покупок
	rprSegments := make([]domain.RepeatPurchaseSegment, len(rprResp.Segments))
	for i, it := range rprResp.Segments {
		rprSegments[i] = domain.RepeatPurchaseSegment{
			SegmentName: it.SegmentName,
			UsersCount:  it.UsersCount,
			Percentage:  it.Percentage,
		}
	}

	// LTV
	ltvPoints := make([]domain.LtvPoint, len(ltvRangeResp.Items))
	for i, it := range ltvRangeResp.Items {
		d := it.CohortDate
		if d == "" {
			d = it.Date
		}
		ltvPoints[i] = domain.LtvPoint{
			Date:         d,
			CohortSize:   it.CohortSize,
			TotalRevenue: it.TotalRevenue,
			LTV:          it.Ltv,
		}
	}

	// Расчет средних сессий
	var sumSessions int64
	var sumMinutes float64
	for _, it := range sessionPoints {
		sumSessions += it.TotalSessions
		sumMinutes += float64(it.TotalSessions) * it.AvgDurationMinutes
	}
	avgSessionMin := 0.0
	if sumSessions > 0 {
		avgSessionMin = math.Round((sumMinutes/float64(sumSessions))*100) / 100
	}

	avgD1 := 0.0
	if len(retentionPoints) > 0 {
		var s float64
		for _, r := range retentionPoints {
			s += r.D1
		}
		avgD1 = math.Round((s/float64(len(retentionPoints)))*10) / 10
	}

	summary := domain.AnalyticsSummary{
		UniquePlayers:           arpuResp.TotalActiveUsers,
		TotalRevenue:            arpuResp.TotalRevenue,
		TotalSessions:           sumSessions,
		AvgSessionMinutes:       avgSessionMin,
		D1RetentionRate:         avgD1,
		OverallCTR:              promoResp.OverallCtr,
		ARPU:                    arpuResp.Arpu,
		ARPPU:                   arpuResp.Arppu,
		PayingUsersCount:        arpuResp.PayingUsersCount,
		PayingUsersPercent:      arpuResp.PayingUsersPercent,
		TotalPurchases:          arpuResp.TotalPurchases,
		AvgOrderValue:           arpuResp.AvgOrderValue,
		TotalAdImpressions:      totalAdImpressions,
		TotalPromoImpressions:   promoResp.TotalImpressions,
		TotalPromoClicks:        promoResp.TotalClicks,
		LTV:                     ltvResp.Ltv,
		AvgHoursToFirstPurchase: ttfpResp.AvgHoursToFirstPurchase,
		RPRConversionRate1To2:   rprResp.ConversionRate1To2,
	}

	return &domain.GameAnalytics{
		ProjectID:       gameID,
		DateFrom:        fromStr,
		DateTo:          toStr,
		Summary:         summary,
		RevenueItems:    revenuePoints,
		DAUItems:        dauPoints,
		SessionItems:    sessionPoints,
		RetentionItems:  retentionPoints,
		ChurnItems:      churnPoints,
		AdItems:         adPoints,
		PromoItems:      promoPoints,
		PromoPlacements: promoPlacements,
		RPRSegments:     rprSegments,
		UserTypeData: domain.UserTypeRevenue{
			NewUsersRevenue:       userRevResp.NewUsers.Revenue,
			NewUsersCount:         userRevResp.NewUsers.UsersCount,
			ReturningUsersRevenue: userRevResp.ReturningUsers.Revenue,
			ReturningUsersCount:   userRevResp.ReturningUsers.UsersCount,
		},
		LTVItems: ltvPoints,
	}, nil
}

func parseTimeSpanMinutes(ts string) float64 {
	parts := strings.Split(ts, ":")
	if len(parts) < 3 {
		return 0
	}
	h, _ := strconv.ParseFloat(parts[0], 64)
	m, _ := strconv.ParseFloat(parts[1], 64)
	s, _ := strconv.ParseFloat(parts[2], 64)
	return h*60.0 + m + s/60.0
}

var _ domain.AnalyticsClient = (*HTTPAnalyticsClient)(nil)
