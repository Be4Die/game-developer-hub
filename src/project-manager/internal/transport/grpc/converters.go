package grpc

import (
	"time"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
	pb "github.com/Be4Die/game-developer-hub/protos/project_manager/v1"
)

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

func projectToProto(p *domain.Project) *pb.Project {
	if p == nil {
		return nil
	}

	proto := &pb.Project{
		Id:            p.ID,
		OwnerId:       p.OwnerID,
		Status:        pb.ProjectStatus(p.Status),
		CreatedAt:     formatTime(p.CreatedAt),
		UpdatedAt:     formatTime(p.UpdatedAt),
		IsOnline:      p.IsOnline,
		IsUnderReview: p.IsUnderReview,
	}

	if p.Draft != nil {
		proto.Draft = draftToProto(p.Draft)
		proto.TitleRu = p.Draft.TitleRu
		proto.TitleEn = p.Draft.TitleEn
		proto.SeoRu = p.Draft.SeoRu
		proto.SeoEn = p.Draft.SeoEn
		proto.AboutRu = p.Draft.AboutRu
		proto.AboutEn = p.Draft.AboutEn
		proto.IconPath = p.Draft.IconPath
		proto.CoverPath = p.Draft.CoverPath
		proto.VideoPath = p.Draft.VideoPath
		proto.ActiveBuildVersion = p.Draft.ActiveBuildVersion
		proto.DevUrl = p.Draft.DevURL
	}

	if p.Release != nil {
		proto.Release = releaseToProto(p.Release)
		proto.ProdUrl = p.Release.ProdURL
	}

	proto.CurrentUserPermissions = p.CurrentUserPermissions
	proto.IsOwner = p.IsOwner

	return proto
}

func draftToProto(d *domain.Draft) *pb.ProjectDraft {
	if d == nil {
		return nil
	}
	return &pb.ProjectDraft{
		ProjectId:          d.ProjectID,
		TitleRu:            d.TitleRu,
		TitleEn:            d.TitleEn,
		SeoRu:              d.SeoRu,
		SeoEn:              d.SeoEn,
		AboutRu:            d.AboutRu,
		AboutEn:            d.AboutEn,
		IconPath:           d.IconPath,
		CoverPath:          d.CoverPath,
		VideoPath:          d.VideoPath,
		ActiveBuildVersion: d.ActiveBuildVersion,
		DevUrl:             d.DevURL,
		IsOnline:           d.IsOnline,
		UpdatedAt:          formatTime(d.UpdatedAt),
	}
}

func buildToProto(b *domain.Build) *pb.ProjectBuild {
	if b == nil {
		return nil
	}
	return &pb.ProjectBuild{
		Id:         b.ID,
		ProjectId:  b.ProjectID,
		Version:    b.Version,
		FilePath:   b.FilePath,
		FileSize:   b.FileSize,
		IsUnpacked: b.IsUnpacked,
		CreatedAt:  formatTime(b.CreatedAt),
	}
}

func releaseToProto(r *domain.Release) *pb.ProjectRelease {
	if r == nil {
		return nil
	}
	return &pb.ProjectRelease{
		Id:          r.ID,
		ProjectId:   r.ProjectID,
		Version:     r.Version,
		TitleRu:     r.TitleRu,
		TitleEn:     r.TitleEn,
		SeoRu:       r.SeoRu,
		SeoEn:       r.SeoEn,
		AboutRu:     r.AboutRu,
		AboutEn:     r.AboutEn,
		IconPath:    r.IconPath,
		CoverPath:   r.CoverPath,
		VideoPath:   r.VideoPath,
		ProdUrl:     r.ProdURL,
		IsOnline:    r.IsOnline,
		PublishedAt: formatTime(r.PublishedAt),
	}
}

func memberToProto(m *domain.Member) *pb.ProjectMember {
	if m == nil {
		return nil
	}
	return &pb.ProjectMember{
		Id:          m.ID,
		ProjectId:   m.ProjectID,
		UserId:      m.UserID,
		UserEmail:   m.UserEmail,
		UserName:    m.UserName,
		Permissions: m.Permissions,
		CreatedAt:   formatTime(m.CreatedAt),
		UpdatedAt:   formatTime(m.UpdatedAt),
	}
}

func invitationToProto(inv *domain.Invitation) *pb.ProjectInvitation {
	if inv == nil {
		return nil
	}
	return &pb.ProjectInvitation{
		Id:           inv.ID,
		ProjectId:    inv.ProjectID,
		ProjectTitle: inv.ProjectTitle,
		ProjectIcon:  inv.ProjectIcon,
		InviterId:    inv.InviterID,
		InviterEmail: inv.InviterEmail,
		InviterName:  inv.InviterName,
		InviteeId:    inv.InviteeID,
		InviteeEmail: inv.InviteeEmail,
		Permissions:  inv.Permissions,
		Status:       pb.InvitationStatus(inv.Status),
		CreatedAt:    formatTime(inv.CreatedAt),
		UpdatedAt:    formatTime(inv.UpdatedAt),
	}
}

func blockToProto(b *domain.UserBlock) *pb.UserAccessBlock {
	if b == nil {
		return nil
	}
	return &pb.UserAccessBlock{
		Id:               b.ID,
		UserId:           b.UserID,
		BlockedUserId:    b.BlockedUserID,
		BlockedUserEmail: b.BlockedUserEmail,
		BlockedUserName:  b.BlockedUserName,
		CreatedAt:        formatTime(b.CreatedAt),
	}
}

func sharedProjectToProto(sp *domain.SharedProject) *pb.SharedProjectItem {
	if sp == nil {
		return nil
	}
	return &pb.SharedProjectItem{
		Project:     projectToProto(sp.Project),
		Permissions: sp.Permissions,
		OwnerEmail:  sp.OwnerEmail,
		OwnerName:   sp.OwnerName,
		JoinedAt:    formatTime(sp.JoinedAt),
	}
}

func gameItemToProto(item *domain.GameItem) *pb.GameItem {
	if item == nil {
		return nil
	}
	return &pb.GameItem{
		Id:          item.ID,
		ProjectId:   item.ProjectID,
		GameItemId:  item.GameItemID,
		Name:        item.Name,
		Description: item.Description,
		ImageUrl:    item.ImageURL,
		PriceCoins:  item.PriceCoins,
		IsActive:    item.IsActive,
		CreatedAt:   formatTime(item.CreatedAt),
		UpdatedAt:   formatTime(item.UpdatedAt),
	}
}

func analyticsToProto(a *domain.GameAnalytics) *pb.GetProjectAnalyticsResponse {
	if a == nil {
		return &pb.GetProjectAnalyticsResponse{}
	}

	resp := &pb.GetProjectAnalyticsResponse{
		ProjectId: a.ProjectID,
		DateFrom:  a.DateFrom,
		DateTo:    a.DateTo,
		Summary: &pb.AnalyticsSummary{
			UniquePlayers:           a.Summary.UniquePlayers,
			TotalRevenue:            a.Summary.TotalRevenue,
			TotalSessions:           a.Summary.TotalSessions,
			AvgSessionMinutes:       a.Summary.AvgSessionMinutes,
			D1RetentionRate:         a.Summary.D1RetentionRate,
			OverallCtr:              a.Summary.OverallCTR,
			Arpu:                    a.Summary.ARPU,
			Arppu:                   a.Summary.ARPPU,
			PayingUsersCount:        a.Summary.PayingUsersCount,
			PayingUsersPercent:      a.Summary.PayingUsersPercent,
			TotalPurchases:          a.Summary.TotalPurchases,
			AvgOrderValue:           a.Summary.AvgOrderValue,
			TotalAdImpressions:      a.Summary.TotalAdImpressions,
			TotalPromoImpressions:   a.Summary.TotalPromoImpressions,
			TotalPromoClicks:        a.Summary.TotalPromoClicks,
			Ltv:                     a.Summary.LTV,
			AvgHoursToFirstPurchase: a.Summary.AvgHoursToFirstPurchase,
			RprConversionRate1To2:   a.Summary.RPRConversionRate1To2,
		},
		RevenueItems:    make([]*pb.AnalyticsRevenuePoint, len(a.RevenueItems)),
		DauItems:        make([]*pb.AnalyticsPlayerPoint, len(a.DAUItems)),
		SessionItems:    make([]*pb.AnalyticsSessionPoint, len(a.SessionItems)),
		RetentionItems:  make([]*pb.AnalyticsCohortRetentionPoint, len(a.RetentionItems)),
		ChurnItems:      make([]*pb.AnalyticsChurnPoint, len(a.ChurnItems)),
		AdItems:         make([]*pb.AnalyticsAdImpressionPoint, len(a.AdItems)),
		PromoItems:      make([]*pb.AnalyticsPromoPoint, len(a.PromoItems)),
		PromoPlacements: make([]*pb.AnalyticsPromoPlacementPoint, len(a.PromoPlacements)),
		RprSegments:     make([]*pb.AnalyticsRepeatPurchaseSegment, len(a.RPRSegments)),
		UserTypeData: &pb.AnalyticsUserTypeRevenue{
			NewUsersRevenue:       a.UserTypeData.NewUsersRevenue,
			NewUsersCount:         a.UserTypeData.NewUsersCount,
			ReturningUsersRevenue: a.UserTypeData.ReturningUsersRevenue,
			ReturningUsersCount:   a.UserTypeData.ReturningUsersCount,
		},
		LtvItems: make([]*pb.AnalyticsLtvPoint, len(a.LTVItems)),
	}

	for i, r := range a.RevenueItems {
		resp.RevenueItems[i] = &pb.AnalyticsRevenuePoint{
			Date:           r.Date,
			TotalRevenue:   r.TotalRevenue,
			PurchasesCount: r.PurchasesCount,
		}
	}
	for i, p := range a.DAUItems {
		resp.DauItems[i] = &pb.AnalyticsPlayerPoint{
			Date:               p.Date,
			UniquePlayersCount: p.UniquePlayersCount,
		}
	}
	for i, s := range a.SessionItems {
		resp.SessionItems[i] = &pb.AnalyticsSessionPoint{
			Date:               s.Date,
			TotalSessions:      s.TotalSessions,
			AvgDurationMinutes: s.AvgDurationMinutes,
			SessionsPerUser:    s.SessionsPerUser,
		}
	}
	for i, ret := range a.RetentionItems {
		resp.RetentionItems[i] = &pb.AnalyticsCohortRetentionPoint{
			Date: ret.Date,
			D1:   ret.D1,
			D3:   ret.D3,
			D7:   ret.D7,
			D30:  ret.D30,
		}
	}
	for i, c := range a.ChurnItems {
		resp.ChurnItems[i] = &pb.AnalyticsChurnPoint{
			Date:      c.Date,
			ChurnRate: c.ChurnRate,
		}
	}
	for i, ad := range a.AdItems {
		resp.AdItems[i] = &pb.AnalyticsAdImpressionPoint{
			Date:         ad.Date,
			Interstitial: ad.Interstitial,
			Rewarded:     ad.Rewarded,
			Banner:       ad.Banner,
			Total:        ad.Total,
		}
	}
	for i, pr := range a.PromoItems {
		resp.PromoItems[i] = &pb.AnalyticsPromoPoint{
			Date:        pr.Date,
			Impressions: pr.Impressions,
			Clicks:      pr.Clicks,
			Ctr:         pr.CTR,
		}
	}
	for i, pl := range a.PromoPlacements {
		resp.PromoPlacements[i] = &pb.AnalyticsPromoPlacementPoint{
			Placement:     pl.Placement,
			PlacementName: pl.PlacementName,
			Impressions:   pl.Impressions,
			Clicks:        pl.Clicks,
			Ctr:           pl.CTR,
		}
	}
	for i, rpr := range a.RPRSegments {
		resp.RprSegments[i] = &pb.AnalyticsRepeatPurchaseSegment{
			SegmentName: rpr.SegmentName,
			UsersCount:  rpr.UsersCount,
			Percentage:  rpr.Percentage,
		}
	}
	for i, ltv := range a.LTVItems {
		resp.LtvItems[i] = &pb.AnalyticsLtvPoint{
			Date:         ltv.Date,
			CohortSize:   ltv.CohortSize,
			TotalRevenue: ltv.TotalRevenue,
			Ltv:          ltv.LTV,
		}
	}

	return resp
}
