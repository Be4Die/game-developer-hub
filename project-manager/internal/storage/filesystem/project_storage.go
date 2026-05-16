// Package filesystem реализует файловое хранилище проектов.
package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// ProjectStorage файловое хранилище проектов.
type ProjectStorage struct {
	basePath string
}

// NewProjectStorage создаёт хранилище.
func NewProjectStorage(basePath string) *ProjectStorage {
	return &ProjectStorage{basePath: basePath}
}

func (s *ProjectStorage) projectDir(projectID int64) string {
	return filepath.Join(s.basePath, strconv.FormatInt(projectID, 10))
}

func (s *ProjectStorage) buildsDir(projectID int64) string {
	return filepath.Join(s.projectDir(projectID), "builds")
}

// SaveIcon сохраняет иконку проекта.
func (s *ProjectStorage) SaveIcon(projectID int64, data []byte) (string, error) {
	dir := s.projectDir(projectID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("fs.SaveIcon mkdir: %w", err)
	}
	path := filepath.Join(dir, "icon.png")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("fs.SaveIcon write: %w", err)
	}
	return path, nil
}

// SaveCover сохраняет обложку проекта.
func (s *ProjectStorage) SaveCover(projectID int64, data []byte) (string, error) {
	dir := s.projectDir(projectID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("fs.SaveCover mkdir: %w", err)
	}
	path := filepath.Join(dir, "cover.png")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("fs.SaveCover write: %w", err)
	}
	return path, nil
}

// SaveVideo сохраняет видео проекта.
func (s *ProjectStorage) SaveVideo(projectID int64, data []byte) (string, error) {
	dir := s.projectDir(projectID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("fs.SaveVideo mkdir: %w", err)
	}
	path := filepath.Join(dir, "video.mp4")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("fs.SaveVideo write: %w", err)
	}
	return path, nil
}

// SaveBuild сохраняет билд проекта.
func (s *ProjectStorage) SaveBuild(projectID int64, version string, data []byte) (string, error) {
	dir := s.buildsDir(projectID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("fs.SaveBuild mkdir: %w", err)
	}
	path := filepath.Join(dir, version+".zip")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("fs.SaveBuild write: %w", err)
	}
	return path, nil
}

// DeleteBuild удаляет билд.
func (s *ProjectStorage) DeleteBuild(projectID int64, version string) error {
	path := filepath.Join(s.buildsDir(projectID), version+".zip")
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("fs.DeleteBuild: %w", err)
	}
	return nil
}

// DeleteProjectDir удаляет директорию проекта целиком.
func (s *ProjectStorage) DeleteProjectDir(projectID int64) error {
	dir := s.projectDir(projectID)
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("fs.DeleteProjectDir: %w", err)
	}
	return nil
}

// BuildExists проверяет существование билда.
func (s *ProjectStorage) BuildExists(projectID int64, version string) bool {
	path := filepath.Join(s.buildsDir(projectID), version+".zip")
	_, err := os.Stat(path)
	return err == nil
}
