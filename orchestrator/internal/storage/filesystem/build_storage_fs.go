// Package filesystem реализует файловое хранилище серверных билдов.
package filesystem

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
)

// BuildStorageFS реализует domain.BuildStorageFS на локальной файловой системе.
// Файлы хранятся по пути: basePath/{gameID}/{version}.tar
// Безопасен для конкурентного использования при условии уникальных версий.
type BuildStorageFS struct {
	basePath string
}

// NewBuildStorageFS создаёт файловое хранилище. basePath — корневая директория.
func NewBuildStorageFS(basePath string) *BuildStorageFS {
	return &BuildStorageFS{basePath: basePath}
}

// Save сохраняет tar-архив билда в директорию {basePath}/{gameID}/{version}.tar.
// Возвращает абсолютный путь к файлу. При ошибке частично записанный файл удаляется.
// Не безопасен при конкурентной записи одного gameID/version.
func (fs *BuildStorageFS) Save(gameID int64, version string, reader io.Reader, size int64) (string, error) {
	dir := filepath.Join(fs.basePath, fmt.Sprintf("%d", gameID))
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", fmt.Errorf("filesystem.BuildStorageFS.Save: mkdir: %w", err)
	}

	filePath := filepath.Join(dir, version+".tar")

	file, err := os.Create(filePath) //nolint:gosec // путь формируется из gameID и version, контролируемых приложением
	if err != nil {
		return "", fmt.Errorf("filesystem.BuildStorageFS.Save: create: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	written, err := io.Copy(file, reader)
	if err != nil {
		_ = os.Remove(filePath)
		return "", fmt.Errorf("filesystem.BuildStorageFS.Save: write: %w", err)
	}

	if size > 0 && written != size {
		_ = os.Remove(filePath)
		return "", fmt.Errorf("filesystem.BuildStorageFS.Save: size mismatch: expected %d, got %d", size, written)
	}

	return filePath, nil
}

// Get открывает reader для чтения файла билда. Caller обязан закрыть reader.
// Возвращает ErrNotFound при отсутствии файла.
func (fs *BuildStorageFS) Get(gameID int64, version string) (io.ReadCloser, error) {
	filePath := filepath.Join(fs.basePath, fmt.Sprintf("%d", gameID), version+".tar")

	file, err := os.Open(filePath) //nolint:gosec // путь формируется из gameID и version, контролируемых приложением
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("filesystem.BuildStorageFS.Get: open: %w", err)
	}

	return file, nil
}

// Delete удаляет файл билда. Возвращает ErrNotFound при отсутствии.
func (fs *BuildStorageFS) Delete(gameID int64, version string) error {
	filePath := filepath.Join(fs.basePath, fmt.Sprintf("%d", gameID), version+".tar")

	err := os.Remove(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("filesystem.BuildStorageFS.Delete: %w", err)
	}

	return nil
}

// Exists проверяет существование билда в файловом хранилище.
func (fs *BuildStorageFS) Exists(gameID int64, version string) bool {
	filePath := filepath.Join(fs.basePath, fmt.Sprintf("%d", gameID), version+".tar")

	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}

	return !info.IsDir()
}
