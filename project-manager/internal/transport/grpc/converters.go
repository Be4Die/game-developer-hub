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
		proto.About = p.Draft.About
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
		About:              d.About,
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
		About:       r.About,
		IconPath:    r.IconPath,
		CoverPath:   r.CoverPath,
		VideoPath:   r.VideoPath,
		ProdUrl:     r.ProdURL,
		PublishedAt: formatTime(r.PublishedAt),
	}
}
