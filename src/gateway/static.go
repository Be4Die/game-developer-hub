package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// handleProjectBuildDownload отдаёт zip-архив клиентской сборки для скачивания.
// GET /api/v1/projects/{project_id}/builds/{version}/download
func handleProjectBuildDownload(basePath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := r.PathValue("project_id")
		version := strings.TrimSuffix(r.PathValue("version"), ".zip")

		if projectID == "" || version == "" {
			http.NotFound(w, r)
			return
		}

		filePath := filepath.Join(basePath, "archives", projectID, version+".zip")
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			// Fallback пути для локальной разработки
			for _, alt := range []string{
				filepath.Join("./data/projects/archives", projectID, version+".zip"),
				filepath.Join("../project-manager/data/projects/archives", projectID, version+".zip"),
			} {
				if _, err := os.Stat(alt); err == nil {
					filePath = alt
					break
				}
			}
		}

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}

		downloadFilename := fmt.Sprintf("build_project_%s_v%s.zip", projectID, version)
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", downloadFilename))
		http.ServeFile(w, r, filePath)
	}
}

// handleProjectMediaServe отдаёт статические файлы промо-материалов (иконки, обложки, видео).
func handleProjectMediaServe(basePath string) http.HandlerFunc {
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

		filePath := filepath.Join(basePath, "media", cleanSub)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			// Fallback пути для локальной разработки
			for _, alt := range []string{
				filepath.Join("./data/projects/media", cleanSub),
				filepath.Join("../project-manager/data/projects/media", cleanSub),
			} {
				if _, err := os.Stat(alt); err == nil {
					filePath = alt
					break
				}
			}
		}

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Cache-Control", "public, max-age=3600")
		http.ServeFile(w, r, filePath)
	}
}
