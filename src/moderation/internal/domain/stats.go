package domain

import "time"

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

// ModeratorActivityItem представляет отдельную операцию модератора в журнале действий.
type ModeratorActivityItem struct {
	ID           int64     `json:"id"`
	ProjectID    int64     `json:"project_id"`
	ActionType   string    `json:"action_type"`
	ActionTitle  string    `json:"action_title"`
	ProjectTitle string    `json:"project_title"`
	ProjectIcon  string    `json:"project_icon"`
	BuildVersion string    `json:"build_version"`
	Details      string    `json:"details"`
	CreatedAt    time.Time `json:"created_at"`
}

