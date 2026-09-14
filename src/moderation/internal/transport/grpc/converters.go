package grpc

import (
	"encoding/json"
	"fmt"
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
		Type:                 pb.RequestType(r.Type),
		Reason:               r.Reason,
		MaxInstances:         r.MaxInstances,
		MaxTotalCpuMillis:    r.MaxTotalCPUMillis,
		MaxTotalMemoryMb:     r.MaxTotalMemoryMB,
		MaxInstanceCpuMillis: r.MaxInstanceCPUMillis,
		MaxInstanceMemoryMb:  r.MaxInstanceMemoryMB,
		ModeratorComment:     r.ModeratorComment,
		RejectionReason:      r.RejectionReason,
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
			IsOnline:           r.Snapshot.IsOnline,
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

func attachmentToProto(a *domain.Attachment) *pb.AttachmentInfo {
	if a == nil {
		return nil
	}
	url := fmt.Sprintf("/api/v1/projects/%d/chat/attachments/%s", a.ProjectID, a.ID)
	downloadURL := fmt.Sprintf("/api/v1/projects/%d/chat/attachments/%s/download", a.ProjectID, a.ID)
	return &pb.AttachmentInfo{
		Id:          a.ID,
		ProjectId:   a.ProjectID,
		FileName:    a.FileName,
		FileSize:    a.FileSize,
		MimeType:    a.MimeType,
		Url:         url,
		DownloadUrl: downloadURL,
		IsPurged:    a.IsPurged,
		CreatedAt:   formatTime(a.CreatedAt),
	}
}

func violationItemFromProto(v *pb.ViolationItem) *domain.ViolationItem {
	if v == nil {
		return nil
	}
	return &domain.ViolationItem{
		RuleCode:      v.GetRuleCode(),
		RuleTitle:     v.GetRuleTitle(),
		Description:   v.GetDescription(),
		AttachmentIDs: v.GetAttachmentIds(),
	}
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

	if len(m.Attachments) > 0 {
		proto.Attachments = make([]*pb.AttachmentInfo, len(m.Attachments))
		for i, a := range m.Attachments {
			proto.Attachments[i] = attachmentToProto(a)
		}
	}

	if len(m.Payload) > 0 {
		if bytes, err := json.Marshal(m.Payload); err == nil {
			proto.PayloadJson = string(bytes)
		}
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
		IsOnline:           p.GetIsOnline(),
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

func moderatorActivityItemToProto(item *domain.ModeratorActivityItem) *pb.ModeratorActivityItem {
	if item == nil {
		return nil
	}
	return &pb.ModeratorActivityItem{
		Id:           item.ID,
		ProjectId:    item.ProjectID,
		ActionType:   item.ActionType,
		ActionTitle:  item.ActionTitle,
		ProjectTitle: item.ProjectTitle,
		ProjectIcon:  item.ProjectIcon,
		BuildVersion: item.BuildVersion,
		Details:      item.Details,
		CreatedAt:    formatTime(item.CreatedAt),
	}
}


