package client

import (
	"context"
	"log/slog"
	"math"
	"math/rand"
	"time"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
)

// StubAnalyticsClient реализует domain.AnalyticsClient для локальной разработки и тестирования,
// генерируя детерминированные реалистичные аналитические метрики на основе gameID и диапазона дат.
type StubAnalyticsClient struct {
	log *slog.Logger
}

// NewStubAnalyticsClient создает экземпляр StubAnalyticsClient.
func NewStubAnalyticsClient(log *slog.Logger) *StubAnalyticsClient {
	if log == nil {
		log = slog.Default()
	}
	return &StubAnalyticsClient{
		log: log,
	}
}

// GetGameAnalytics генерирует реалистичный набор аналитических метрик за указанный период.
func (s *StubAnalyticsClient) GetGameAnalytics(_ context.Context, gameID int64, filter domain.AnalyticsFilter) (*domain.GameAnalytics, error) {
	s.log.Debug("[StubAnalyticsClient] Generating mock analytics",
		slog.Int64("game_id", gameID),
		slog.String("date_from", filter.DateFrom),
		slog.String("date_to", filter.DateTo),
		slog.String("period", filter.AggregationPeriod),
	)

	dateFrom, dateTo := parseDateRange(filter.DateFrom, filter.DateTo)
	dates := generateDateList(dateFrom, dateTo)
	n := len(dates)
	if n == 0 {
		dates = []string{time.Now().UTC().Format("2006-01-02")}
		n = 1
	}

	// Сид генератора на основе gameID для воспроизводимости данных по игре
	seed := int64(gameID)*100003 + int64(dateFrom.Unix())
	rng := rand.New(rand.NewSource(seed))

	basePlayers := 80.0 + float64(gameID%10)*35.0 + rng.Float64()*120.0
	baseRevenue := basePlayers * (15.0 + rng.Float64()*35.0)

	dauItems := make([]domain.PlayerPoint, n)
	revenueItems := make([]domain.RevenuePoint, n)
	sessionItems := make([]domain.SessionPoint, n)
	retentionItems := make([]domain.CohortRetentionPoint, n)
	churnItems := make([]domain.ChurnPoint, n)
	adItems := make([]domain.AdImpressionPoint, n)
	promoItems := make([]domain.PromoPoint, n)
	ltvItems := make([]domain.LtvPoint, n)

	var (
		sumRevenue          float64
		sumPurchases        int64
		sumSessions         int64
		sumDurationWeighted float64
		sumAdImpressions    int64
		sumPromoImpressions int64
		sumPromoClicks      int64
		maxDAU              int64
	)

	for i, d := range dates {
		// Колебания дня недели (выходные чуть активнее)
		parsedD, _ := time.Parse("2006-01-02", d)
		dayFactor := 1.0
		if parsedD.Weekday() == time.Saturday || parsedD.Weekday() == time.Sunday {
			dayFactor = 1.25
		}

		// DAU
		dauNoise := (rng.Float64() - 0.48) * 20.0
		dau := int64(math.Max(15, (basePlayers+dauNoise)*dayFactor))
		dauItems[i] = domain.PlayerPoint{
			Date:               d,
			UniquePlayersCount: dau,
		}
		if dau > maxDAU {
			maxDAU = dau
		}

		// Покупки и выручка
		payingRatio := 0.04 + rng.Float64()*0.06
		payingUsers := int64(math.Round(float64(dau) * payingRatio))
		purchases := int64(math.Round(float64(payingUsers) * (1.1 + rng.Float64()*0.8)))
		if purchases < 0 {
			purchases = 0
		}
		avgCheck := 250.0 + rng.Float64()*300.0
		dayRevenue := math.Round(float64(purchases)*avgCheck*100) / 100
		revenueItems[i] = domain.RevenuePoint{
			Date:           d,
			TotalRevenue:   dayRevenue,
			PurchasesCount: purchases,
		}
		sumRevenue += dayRevenue
		sumPurchases += purchases

		// Сессии
		spu := math.Round((1.6+rng.Float64()*1.4)*100) / 100
		daySessions := int64(math.Round(float64(dau) * spu))
		avgDurMin := math.Round((7.0+rng.Float64()*12.0)*100) / 100
		sessionItems[i] = domain.SessionPoint{
			Date:               d,
			TotalSessions:      daySessions,
			AvgDurationMinutes: avgDurMin,
			SessionsPerUser:    spu,
		}
		sumSessions += daySessions
		sumDurationWeighted += float64(daySessions) * avgDurMin

		// Когортный ретеншн
		d1 := math.Round((28.0+rng.Float64()*22.0)*10) / 10
		d3 := math.Round((d1*(0.55+rng.Float64()*0.15))*10) / 10
		d7 := math.Round((d3*(0.50+rng.Float64()*0.20))*10) / 10
		d30 := math.Round((d7*(0.30+rng.Float64()*0.20))*10) / 10
		retentionItems[i] = domain.CohortRetentionPoint{
			Date: d,
			D1:   d1,
			D3:   d3,
			D7:   d7,
			D30:  d30,
		}

		// Отток
		churn := math.Round((4.5+rng.Float64()*7.5)*10) / 10
		churnItems[i] = domain.ChurnPoint{
			Date:      d,
			ChurnRate: churn,
		}

		// Реклама
		interstitial := int64(rng.Intn(300) + 100)
		rewarded := int64(rng.Intn(200) + 50)
		banner := int64(rng.Intn(500) + 150)
		totalAd := interstitial + rewarded + banner
		adItems[i] = domain.AdImpressionPoint{
			Date:         d,
			Interstitial: interstitial,
			Rewarded:     rewarded,
			Banner:       banner,
			Total:        totalAd,
		}
		sumAdImpressions += totalAd

		// Промо CTR
		promoImp := int64(rng.Intn(800) + 300)
		promoClk := int64(float64(promoImp) * (0.03 + rng.Float64()*0.07))
		promoCtr := 0.0
		if promoImp > 0 {
			promoCtr = math.Round((float64(promoClk)/float64(promoImp)*100)*100) / 100
		}
		promoItems[i] = domain.PromoPoint{
			Date:        d,
			Impressions: promoImp,
			Clicks:      promoClk,
			CTR:         promoCtr,
		}
		sumPromoImpressions += promoImp
		sumPromoClicks += promoClk

		// LTV когорты
		cohortSize := int64(rng.Intn(80) + 20)
		cohortRev := float64(cohortSize) * (180.0 + rng.Float64()*300.0)
		cohortLTV := math.Round((cohortRev/float64(cohortSize))*100) / 100
		ltvItems[i] = domain.LtvPoint{
			Date:         d,
			CohortSize:   cohortSize,
			TotalRevenue: math.Round(cohortRev*100) / 100,
			LTV:          cohortLTV,
		}
	}

	// Сводный расчет средних
	avgSessionMinutes := 0.0
	if sumSessions > 0 {
		avgSessionMinutes = math.Round((sumDurationWeighted/float64(sumSessions))*100) / 100
	}

	avgD1Retention := 0.0
	if n > 0 {
		sumD1 := 0.0
		for _, r := range retentionItems {
			sumD1 += r.D1
		}
		avgD1Retention = math.Round((sumD1/float64(n))*10) / 10
	}

	overallCTR := 0.0
	if sumPromoImpressions > 0 {
		overallCTR = math.Round((float64(sumPromoClicks)/float64(sumPromoImpressions)*100)*100) / 100
	}

	totalActiveUnique := int64(float64(maxDAU) * (1.8 + rng.Float64()*1.2))
	if totalActiveUnique < maxDAU {
		totalActiveUnique = maxDAU
	}

	payingUsersTotal := int64(math.Round(float64(totalActiveUnique) * (0.06 + rng.Float64()*0.06)))
	if payingUsersTotal == 0 {
		payingUsersTotal = 1
	}

	arpu := math.Round((sumRevenue/float64(totalActiveUnique))*100) / 100
	arppu := math.Round((sumRevenue/float64(payingUsersTotal))*100) / 100
	payingPct := math.Round((float64(payingUsersTotal)/float64(totalActiveUnique)*100)*10) / 10

	avgOrderVal := 0.0
	if sumPurchases > 0 {
		avgOrderVal = math.Round((sumRevenue/float64(sumPurchases))*100) / 100
	}

	// Разбивка по плейсментам промо
	placementsMeta := []struct {
		id    string
		name  string
		share float64
	}{
		{"MainPageCatalog", "Главная (каталог)", 0.52},
		{"InGameSidebar", "Блок внутри игры", 0.28},
		{"SearchResults", "Результаты поиска", 0.14},
		{"CategoryPage", "Страница категории", 0.06},
	}
	promoPlacements := make([]domain.PromoPlacementPoint, len(placementsMeta))
	for idx, pl := range placementsMeta {
		imp := int64(float64(sumPromoImpressions) * pl.share)
		clk := int64(float64(sumPromoClicks) * pl.share * (0.8 + rng.Float64()*0.4))
		ctr := 0.0
		if imp > 0 {
			ctr = math.Round((float64(clk)/float64(imp)*100)*100) / 100
		}
		promoPlacements[idx] = domain.PromoPlacementPoint{
			Placement:     pl.id,
			PlacementName: pl.name,
			Impressions:   imp,
			Clicks:        clk,
			CTR:           ctr,
		}
	}

	// Сегменты повторных покупок (RPR)
	seg1 := int64(float64(payingUsersTotal) * 0.55)
	seg2 := int64(float64(payingUsersTotal) * 0.25)
	seg3 := int64(float64(payingUsersTotal) * 0.14)
	seg4 := payingUsersTotal - seg1 - seg2 - seg3
	if seg4 < 0 {
		seg4 = 1
	}
	totSeg := float64(seg1 + seg2 + seg3 + seg4)

	rprSegments := []domain.RepeatPurchaseSegment{
		{SegmentName: "1 покупка", UsersCount: seg1, Percentage: math.Round((float64(seg1)/totSeg*100)*10) / 10},
		{SegmentName: "2–3 покупки", UsersCount: seg2, Percentage: math.Round((float64(seg2)/totSeg*100)*10) / 10},
		{SegmentName: "4–10 покупок", UsersCount: seg3, Percentage: math.Round((float64(seg3)/totSeg*100)*10) / 10},
		{SegmentName: "10+ покупок", UsersCount: seg4, Percentage: math.Round((float64(seg4)/totSeg*100)*10) / 10},
	}
	rprConv := 0.0
	if totSeg > 0 {
		rprConv = math.Round((float64(seg2+seg3+seg4)/totSeg)*10000) / 10000
	}

	// Новые vs вернувшиеся покупатели
	newShare := 0.35 + rng.Float64()*0.15
	newRev := math.Round(sumRevenue*newShare*100) / 100
	retRev := math.Round((sumRevenue-newRev)*100) / 100
	newU := int64(float64(payingUsersTotal) * newShare)
	retU := payingUsersTotal - newU

	userTypeData := domain.UserTypeRevenue{
		NewUsersRevenue:       newRev,
		NewUsersCount:         newU,
		ReturningUsersRevenue: retRev,
		ReturningUsersCount:   retU,
	}

	return &domain.GameAnalytics{
		ProjectID: gameID,
		DateFrom:  dates[0],
		DateTo:    dates[len(dates)-1],
		Summary: domain.AnalyticsSummary{
			UniquePlayers:           totalActiveUnique,
			TotalRevenue:            math.Round(sumRevenue*100) / 100,
			TotalSessions:           sumSessions,
			AvgSessionMinutes:       avgSessionMinutes,
			D1RetentionRate:         avgD1Retention,
			OverallCTR:              overallCTR,
			ARPU:                    arpu,
			ARPPU:                   arppu,
			PayingUsersCount:        payingUsersTotal,
			PayingUsersPercent:      payingPct,
			TotalPurchases:          sumPurchases,
			AvgOrderValue:           avgOrderVal,
			TotalAdImpressions:      sumAdImpressions,
			TotalPromoImpressions:   sumPromoImpressions,
			TotalPromoClicks:        sumPromoClicks,
			LTV:                     math.Round((baseRevenue/basePlayers*8.0)*100) / 100,
			AvgHoursToFirstPurchase: math.Round((24.0+rng.Float64()*48.0)*10) / 10,
			RPRConversionRate1To2:   rprConv,
		},
		RevenueItems:    revenueItems,
		DAUItems:        dauItems,
		SessionItems:    sessionItems,
		RetentionItems:  retentionItems,
		ChurnItems:      churnItems,
		AdItems:         adItems,
		PromoItems:      promoItems,
		PromoPlacements: promoPlacements,
		RPRSegments:     rprSegments,
		UserTypeData:    userTypeData,
		LTVItems:        ltvItems,
	}, nil
}

func parseDateRange(fromStr, toStr string) (time.Time, time.Time) {
	now := time.Now().UTC()
	to := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	from := to.AddDate(0, 0, -30)

	if toStr != "" {
		if t, err := time.Parse("2006-01-02", toStr); err == nil {
			to = t
		}
	}
	if fromStr != "" {
		if f, err := time.Parse("2006-01-02", fromStr); err == nil {
			from = f
		}
	}

	if from.After(to) {
		from = to.AddDate(0, 0, -30)
	}

	return from, to
}

func generateDateList(from, to time.Time) []string {
	var dates []string
	cur := from
	for !cur.After(to) {
		dates = append(dates, cur.Format("2006-01-02"))
		cur = cur.AddDate(0, 0, 1)
	}
	return dates
}

var _ domain.AnalyticsClient = (*StubAnalyticsClient)(nil)
