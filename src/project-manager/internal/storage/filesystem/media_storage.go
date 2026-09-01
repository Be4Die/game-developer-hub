package filesystem

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
)

// MediaStorage реализует domain.MediaStorage для хранения промо-материалов на файловой системе.
type MediaStorage struct {
	basePath string
}

// NewMediaStorage создаёт экземпляр хранилища медиа-файлов.
func NewMediaStorage(basePath string) *MediaStorage {
	return &MediaStorage{basePath: basePath}
}

func (s *MediaStorage) projectDir(projectID int64) string {
	return filepath.Join(s.basePath, "media", strconv.FormatInt(projectID, 10))
}

func (s *MediaStorage) getFileName(mediaType string) (string, error) {
	switch mediaType {
	case "icon":
		return "icon.png", nil
	case "cover":
		return "cover.png", nil
	case "video":
		return "video.mp4", nil
	default:
		return "", domain.ErrInvalidInput
	}
}

func validateMediaMime(path string, mediaType string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return err
	}
	if n == 0 {
		return domain.ErrInvalidInput
	}

	contentType := http.DetectContentType(buf[:n])
	switch mediaType {
	case "icon", "cover":
		if !strings.HasPrefix(contentType, "image/") {
			return fmt.Errorf("%w: expected image, detected %s", domain.ErrInvalidInput, contentType)
		}
	case "video":
		if !strings.HasPrefix(contentType, "video/") && contentType != "application/octet-stream" {
			return fmt.Errorf("%w: expected video, detected %s", domain.ErrInvalidInput, contentType)
		}
	}
	return nil
}

// SaveMediaStream сохраняет промо-файл из потока io.Reader с валидацией MIME-типа.
func (s *MediaStorage) SaveMediaStream(ctx context.Context, projectID int64, mediaType string, src io.Reader) (string, error) {
	fileName, err := s.getFileName(mediaType)
	if err != nil {
		return "", err
	}

	dir := s.projectDir(projectID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir: %w", err)
	}

	targetPath := filepath.Join(dir, fileName)
	tmpFile, err := os.CreateTemp(dir, "media-*.tmp")
	if err != nil {
		return "", fmt.Errorf("create temp: %w", err)
	}
	tmpPath := tmpFile.Name()

	var writeErr error
	defer func() {
		_ = tmpFile.Close()
		if writeErr != nil {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, writeErr = io.Copy(tmpFile, src); writeErr != nil {
		return "", fmt.Errorf("write stream: %w", writeErr)
	}

	if err := tmpFile.Close(); err != nil {
		writeErr = err
		return "", fmt.Errorf("close temp: %w", err)
	}

	// Валидируем реальный MIME-тип
	if err := validateMediaMime(tmpPath, mediaType); err != nil {
		writeErr = err
		return "", err
	}

	if err := os.Rename(tmpPath, targetPath); err != nil {
		writeErr = err
		return "", fmt.Errorf("rename: %w", err)
	}

	_ = os.Chmod(targetPath, 0o644)

	return targetPath, nil
}

// SaveMedia сохраняет промо-файл из среза байтов.
func (s *MediaStorage) SaveMedia(ctx context.Context, projectID int64, mediaType string, data []byte) (string, error) {
	return s.SaveMediaStream(ctx, projectID, mediaType, bytes.NewReader(data))
}

// DeleteMedia удаляет промо-файл указанного типа.
func (s *MediaStorage) DeleteMedia(projectID int64, mediaType string) error {
	fileName, err := s.getFileName(mediaType)
	if err != nil {
		return err
	}
	path := filepath.Join(s.projectDir(projectID), fileName)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove media: %w", err)
	}
	return nil
}

// SnapshotMediaForRelease создает неизменяемую копию медиафайла для конкретного релиза.
func (s *MediaStorage) SnapshotMediaForRelease(ctx context.Context, projectID int64, version string, srcPath string, mediaType string) (string, error) {
	if srcPath == "" {
		return "", nil
	}
	fileName, err := s.getFileName(mediaType)
	if err != nil {
		return "", err
	}

	releaseDir := filepath.Join(s.basePath, "media", strconv.FormatInt(projectID, 10), "releases", version)
	if err := os.MkdirAll(releaseDir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir release media: %w", err)
	}

	destPath := filepath.Join(releaseDir, fileName)

	// Открываем исходный файл
	srcFile, err := os.Open(srcPath)
	if err != nil {
		// Резервная попытка: проверить в директории проекта
		altPath := filepath.Join(s.projectDir(projectID), fileName)
		srcFile, err = os.Open(altPath)
		if err != nil {
			// Если исходного файла нет на диске, сохраняем исходный путь
			return srcPath, nil
		}
	}
	defer srcFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("create release media file: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return "", fmt.Errorf("copy media to release: %w", err)
	}
	_ = os.Chmod(destPath, 0o644)

	return destPath, nil
}
