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
		Id:        p.ID,
		OwnerId:   p.OwnerID,
		Status:    pb.ProjectStatus(p.Status),
		CreatedAt: formatTime(p.CreatedAt),
		UpdatedAt: formatTime(p.UpdatedAt),
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
