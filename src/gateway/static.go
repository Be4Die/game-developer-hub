package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// handleProjectBuildDownload отдаёт zip-архив клиентской сборки для скачивания.
// GET /api/v1/projects/{project_id}/builds/{version}/download
func handleProjectBuildDownload(basePath string, s3Clients ...*s3.Client) http.HandlerFunc {
	var s3Client *s3.Client
	if len(s3Clients) > 0 {
		s3Client = s3Clients[0]
	}

	buildsBucket := os.Getenv("S3_BUILDS_BUCKET")
	if buildsBucket == "" {
		buildsBucket = "builds"
	}

	return func(w http.ResponseWriter, r *http.Request) {
		projectID := r.PathValue("project_id")
		version := strings.TrimSuffix(r.PathValue("version"), ".zip")

		if projectID == "" || version == "" {
			http.NotFound(w, r)
			return
		}

		downloadFilename := fmt.Sprintf("build_project_%s_v%s.zip", projectID, version)

		// 1. Попытка отдать из S3 (SeaweedFS)
		if s3Client != nil {
			s3Key := fmt.Sprintf("projects/%s/%s.zip", projectID, version)
			out, err := s3Client.GetObject(r.Context(), &s3.GetObjectInput{
				Bucket: aws.String(buildsBucket),
				Key:    aws.String(s3Key),
			})
			if err == nil {
				defer func() { _ = out.Body.Close() }()
				w.Header().Set("Content-Type", "application/zip")
				w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", downloadFilename))
				if out.ContentLength != nil && *out.ContentLength > 0 {
					w.Header().Set("Content-Length", fmt.Sprintf("%d", *out.ContentLength))
				}
				_, _ = io.Copy(w, out.Body)
				return
			}
		}

		// 2. Фоллбэк на локальную файловую систему
		filePath := filepath.Join(basePath, "archives", projectID, version+".zip")
		if _, err := os.Stat(filePath); os.IsNotExist(err) { //nolint:gosec
			for _, alt := range []string{
				filepath.Join("./data/projects/archives", projectID, version+".zip"),
				filepath.Join("../project-manager/data/projects/archives", projectID, version+".zip"),
			} {
				if _, err := os.Stat(alt); err == nil { //nolint:gosec
					filePath = alt
					break
				}
			}
		}

		if _, err := os.Stat(filePath); os.IsNotExist(err) { //nolint:gosec
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", downloadFilename))
		http.ServeFile(w, r, filePath) //nolint:gosec
	}
}

// handleProjectMediaServe отдаёт статические файлы промо-материалов (иконки, обложки, видео).
func handleProjectMediaServe(basePath string, s3Clients ...*s3.Client) http.HandlerFunc {
	var s3Client *s3.Client
	if len(s3Clients) > 0 {
		s3Client = s3Clients[0]
	}

	mediaBucket := os.Getenv("S3_MEDIA_BUCKET")
	if mediaBucket == "" {
		mediaBucket = "media"
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var subPath string

		projectID := r.PathValue("project_id")
		mediaType := r.PathValue("type")
		if projectID != "" && mediaType != "" {
			switch mediaType {
			case "icon", "icon.png":
				subPath = filepath.Join(projectID, "icon.png")
			case "cover", "cover.png":
				subPath = filepath.Join(projectID, "cover.png")
			case "video", "video.mp4":
				subPath = filepath.Join(projectID, "video.mp4")
			default:
				subPath = filepath.Join(projectID, mediaType)
			}
		} else {
			rel := r.PathValue("path")
			rel = strings.TrimPrefix(rel, "data/projects/")
			rel = strings.TrimPrefix(rel, "projects/")
			rel = strings.TrimPrefix(rel, "media/")
			subPath = rel
		}

		cleanSub := filepath.Clean(subPath)
		if cleanSub == "." || strings.HasPrefix(cleanSub, "..") {
			http.NotFound(w, r)
			return
		}

		// 1. Попытка отдать из S3 (SeaweedFS)
		if s3Client != nil {
			s3Key := filepath.ToSlash(cleanSub)
			out, err := s3Client.GetObject(r.Context(), &s3.GetObjectInput{
				Bucket: aws.String(mediaBucket),
				Key:    aws.String(s3Key),
			})
			if err == nil {
				defer func() { _ = out.Body.Close() }()
				contentType := "application/octet-stream"
				if out.ContentType != nil && *out.ContentType != "" {
					contentType = *out.ContentType
				} else if ct := mime.TypeByExtension(filepath.Ext(cleanSub)); ct != "" {
					contentType = ct
				}
				w.Header().Set("Content-Type", contentType)
				w.Header().Set("Cache-Control", "public, max-age=3600")
				if out.ContentLength != nil && *out.ContentLength > 0 {
					w.Header().Set("Content-Length", fmt.Sprintf("%d", *out.ContentLength))
				}
				_, _ = io.Copy(w, out.Body)
				return
			}
		}

		// 2. Фоллбэк на локальную файловую систему
		filePath := filepath.Join(basePath, "media", cleanSub)
		if _, err := os.Stat(filePath); os.IsNotExist(err) { //nolint:gosec
			for _, alt := range []string{
				filepath.Join("./data/projects/media", cleanSub),
				filepath.Join("../project-manager/data/projects/media", cleanSub),
			} {
				if _, err := os.Stat(alt); err == nil { //nolint:gosec
					filePath = alt
					break
				}
			}
		}

		if _, err := os.Stat(filePath); os.IsNotExist(err) { //nolint:gosec
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Cache-Control", "public, max-age=3600")
		http.ServeFile(w, r, filePath) //nolint:gosec
	}
}
