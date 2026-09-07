package grpc

import (
	"time"

	"github.com/Be4Die/game-developer-hub/moderation/internal/domain"
	pb "github.com/Be4Die/game-developer-hub/protos/moderation/v1"
)

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

func requestToProto(r *domain.ModerationRequest) *pb.ModerationRequest {
	if r == nil {
		return nil
	}

	proto := &pb.ModerationRequest{
		Id:              r.ID,
		ProjectId:       r.ProjectID,
		OwnerId:         r.OwnerID,
		ModeratorId:     r.ModeratorID,
		Status:          pb.RequestStatus(r.Status),
		RejectionReason: r.RejectionReason,
		SubmittedAt:     formatTime(r.SubmittedAt),
		Snapshot: &pb.ProjectSnapshot{
			ProjectId:          r.Snapshot.ProjectID,
			TitleRu:            r.Snapshot.TitleRu,
			TitleEn:            r.Snapshot.TitleEn,
			SeoRu:              r.Snapshot.SeoRu,
			SeoEn:              r.Snapshot.SeoEn,
			AboutRu:            r.Snapshot.AboutRu,
			AboutEn:            r.Snapshot.AboutEn,
			IconPath:           r.Snapshot.IconPath,
			CoverPath:          r.Snapshot.CoverPath,
			VideoPath:          r.Snapshot.VideoPath,
			ActiveBuildVersion: r.Snapshot.ActiveBuildVersion,
			DevUrl:             r.Snapshot.DevURL,
		},
	}

	if r.StartedReviewAt != nil {
		proto.StartedReviewAt = formatTime(*r.StartedReviewAt)
	}
	if r.ResolvedAt != nil {
		proto.ResolvedAt = formatTime(*r.ResolvedAt)
	}

	return proto
}

func messageToProto(m *domain.ChatMessage) *pb.ChatMessage {
	if m == nil {
		return nil
	}

	proto := &pb.ChatMessage{
		Id:          m.ID,
		ProjectId:   m.ProjectID,
		SenderId:    m.SenderID,
		SenderRole:  pb.SenderRole(m.SenderRole),
		MessageType: pb.MessageType(m.MessageType),
		Content:     m.Content,
		CreatedAt:   formatTime(m.CreatedAt),
	}

	if m.RequestID != nil {
		proto.RequestId = *m.RequestID
	}

	return proto
}

func snapshotFromProto(p *pb.ProjectSnapshot) domain.ProjectSnapshot {
	if p == nil {
		return domain.ProjectSnapshot{}
	}
	return domain.ProjectSnapshot{
		ProjectID:          p.GetProjectId(),
		TitleRu:            p.GetTitleRu(),
		TitleEn:            p.GetTitleEn(),
		SeoRu:              p.GetSeoRu(),
		SeoEn:              p.GetSeoEn(),
		AboutRu:            p.GetAboutRu(),
		AboutEn:            p.GetAboutEn(),
		IconPath:           p.GetIconPath(),
		CoverPath:          p.GetCoverPath(),
		VideoPath:          p.GetVideoPath(),
		ActiveBuildVersion: p.GetActiveBuildVersion(),
		DevURL:             p.GetDevUrl(),
	}
}

func moderatorStatsToProto(s *domain.ModeratorStats) *pb.ModeratorStats {
	if s == nil {
		return nil
	}
	return &pb.ModeratorStats{
		ModeratorId:              s.ModeratorID,
		TotalAssigned:            s.TotalAssigned,
		InReviewCount:            s.InReviewCount,
		ApprovedCount:            s.ApprovedCount,
		RejectedCount:            s.RejectedCount,
		TotalResolved:            s.TotalResolved,
		ApprovalRate:             s.ApprovalRate,
		RejectionRate:            s.RejectionRate,
		AvgReviewDurationSeconds: s.AvgReviewDurationSeconds,
		MessagesSent:             s.MessagesSent,
		TodayResolved:            s.TodayResolved,
		WeekResolved:             s.WeekResolved,
		MonthResolved:            s.MonthResolved,
	}
}

