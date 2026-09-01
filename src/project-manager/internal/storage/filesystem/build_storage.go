package filesystem

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

// BuildStorage реализует domain.BuildStorage для хранения архивов сборок на локальной файловой системе.
type BuildStorage struct {
	basePath string
}

// NewBuildStorage создаёт экземпляр хранилища сборок.
func NewBuildStorage(basePath string) *BuildStorage {
	return &BuildStorage{basePath: basePath}
}

func (s *BuildStorage) projectDir(projectID int64) string {
	return filepath.Join(s.basePath, "archives", strconv.FormatInt(projectID, 10))
}

// SaveArchiveStream сохраняет архив сборки из потока io.Reader.
// Потоково записывает данные во временный файл и атомарно перемещает в целевой путь.
func (s *BuildStorage) SaveArchiveStream(_ context.Context, projectID int64, version string, src io.Reader) (string, int64, error) {
	dir := s.projectDir(projectID)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", 0, fmt.Errorf("mkdir: %w", err)
	}

	targetPath := filepath.Join(dir, version+".zip")
	tmpFile, err := os.CreateTemp(dir, "upload-*.tmp")
	if err != nil {
		return "", 0, fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	var writeErr error
	var written int64

	defer func() {
		_ = tmpFile.Close()
		if writeErr != nil {
			_ = os.Remove(tmpPath)
		}
	}()

	written, writeErr = io.Copy(tmpFile, src)
	if writeErr != nil {
		return "", 0, fmt.Errorf("stream copy: %w", writeErr)
	}

	if err := tmpFile.Close(); err != nil {
		writeErr = err
		return "", 0, fmt.Errorf("close temp file: %w", err)
	}

	if err := os.Rename(tmpPath, targetPath); err != nil {
		writeErr = err
		return "", 0, fmt.Errorf("rename temp to target: %w", err)
	}
	_ = os.Chmod(targetPath, 0o600)

	return targetPath, written, nil
}

// SaveArchive сохраняет архив сборки из массива байтов в хранилище.
func (s *BuildStorage) SaveArchive(ctx context.Context, projectID int64, version string, data []byte) (string, int64, error) {
	return s.SaveArchiveStream(ctx, projectID, version, bytes.NewReader(data))
}

// GetArchivePath возвращает абсолютный путь к архиву версии.
func (s *BuildStorage) GetArchivePath(projectID int64, version string) string {
	return filepath.Join(s.projectDir(projectID), version+".zip")
}

// DeleteBuild удаляет файл архива конкретной версии сборки.
func (s *BuildStorage) DeleteBuild(projectID int64, version string) error {
	path := s.GetArchivePath(projectID, version)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove archive: %w", err)
	}
	return nil
}

// DeleteProject удаляет всю директорию архивов проекта.
func (s *BuildStorage) DeleteProject(projectID int64) error {
	dir := s.projectDir(projectID)
	if err := os.RemoveAll(dir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove project dir: %w", err)
	}
	return nil
}
