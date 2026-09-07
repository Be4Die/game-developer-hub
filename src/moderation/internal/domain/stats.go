// Package domain содержит бизнес-сущности и правила предметной области модерации.
package domain

// ModeratorStats представляет агрегированные статистические показатели работы модератора.
type ModeratorStats struct {
	ModeratorID              string  `json:"moderator_id"`
	TotalAssigned            int32   `json:"total_assigned"`
	InReviewCount            int32   `json:"in_review_count"`
	ApprovedCount            int32   `json:"approved_count"`
	RejectedCount            int32   `json:"rejected_count"`
	TotalResolved            int32   `json:"total_resolved"`
	ApprovalRate             float64 `json:"approval_rate"`
	RejectionRate            float64 `json:"rejection_rate"`
	AvgReviewDurationSeconds int64   `json:"avg_review_duration_seconds"`
	MessagesSent             int32   `json:"messages_sent"`
	TodayResolved            int32   `json:"today_resolved"`
	WeekResolved             int32   `json:"week_resolved"`
	MonthResolved            int32   `json:"month_resolved"`
}
