package service

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/infrastructure/config"
)

// ─── Mocks for BuildPipeline ────────────────────────────────────────────────

type buildMockStorage struct {
	createFn               func(ctx context.Context, build *domain.ServerBuild) error
	getByIDFn              func(ctx context.Context, id int64) (*domain.ServerBuild, error)
	getByVersionFn         func(ctx context.Context, gameID int64, version string) (*domain.ServerBuild, error)
	listByGameFn           func(ctx context.Context, gameID int64, limit int) ([]*domain.ServerBuild, error)
	countByGameFn          func(ctx context.Context, gameID int64) (int, error)
	deleteFn               func(ctx context.Context, id int64) error
	countActiveInstancesFn func(ctx context.Context, buildID int64) (int, error)
}

func (m *buildMockStorage) Create(ctx context.Context, build *domain.ServerBuild) error {
	if m.createFn != nil {
		return m.createFn(ctx, build)
	}
	return nil
}
func (m *buildMockStorage) GetByID(ctx context.Context, id int64) (*domain.ServerBuild, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, domain.ErrNotFound
}
func (m *buildMockStorage) GetByVersion(ctx context.Context, gameID int64, version string) (*domain.ServerBuild, error) {
	if m.getByVersionFn != nil {
		return m.getByVersionFn(ctx, gameID, version)
	}
	return nil, domain.ErrNotFound
}
func (m *buildMockStorage) ListByGame(ctx context.Context, gameID int64, limit int) ([]*domain.ServerBuild, error) {
	if m.listByGameFn != nil {
		return m.listByGameFn(ctx, gameID, limit)
	}
	return nil, nil
}
func (m *buildMockStorage) CountByGame(ctx context.Context, gameID int64) (int, error) {
	if m.countByGameFn != nil {
		return m.countByGameFn(ctx, gameID)
	}
	return 0, nil
}
func (m *buildMockStorage) Delete(ctx context.Context, id int64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}
func (m *buildMockStorage) CountActiveInstancesByBuild(ctx context.Context, buildID int64) (int, error) {
	if m.countActiveInstancesFn != nil {
		return m.countActiveInstancesFn(ctx, buildID)
	}
	return 0, nil
}

type buildMockStorageFS struct {
	saveFn   func(gameID int64, version string, reader io.Reader, size int64) (string, error)
	getFn    func(gameID int64, version string) (io.ReadCloser, error)
	deleteFn func(gameID int64, version string) error
	existsFn func(gameID int64, version string) bool
}

func (m *buildMockStorageFS) Save(gameID int64, version string, reader io.Reader, size int64) (string, error) {
	if m.saveFn != nil {
		return m.saveFn(gameID, version, reader, size)
	}
	return "", nil
}
func (m *buildMockStorageFS) Get(gameID int64, version string) (io.ReadCloser, error) {
	if m.getFn != nil {
		return m.getFn(gameID, version)
	}
	return nil, domain.ErrNotFound
}
func (m *buildMockStorageFS) Delete(gameID int64, version string) error {
	if m.deleteFn != nil {
		return m.deleteFn(gameID, version)
	}
	return nil
}
func (m *buildMockStorageFS) Exists(gameID int64, version string) bool {
	if m.existsFn != nil {
		return m.existsFn(gameID, version)
	}
	return false
}

// ─── Tests: generateDockerfile ───────────────────────────────────────────────

// ─── Tests: extractZip ───────────────────────────────────────────────────────

func TestBuildPipeline_ExtractZip_Success(t *testing.T) {
	// Create a ZIP in memory.
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	// Add a file.
	fw, err := zw.Create("hello.txt")
	if err != nil {
		t.Fatalf("failed to create zip entry: %v", err)
	}
	if _, err := fw.Write([]byte("hello world")); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	// Add a file in a subdirectory.
	fw2, err := zw.Create("sub/file.txt")
	if err != nil {
		t.Fatalf("failed to create zip entry: %v", err)
	}
	if _, err := fw2.Write([]byte("nested content")); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	if err := zw.Close(); err != nil {
		t.Fatalf("failed to close zip: %v", err)
	}

	destDir := t.TempDir()
	pipeline := &BuildPipeline{}
	if err := pipeline.extractZip(buf.Bytes(), destDir); err != nil {
		t.Fatalf("extractZip failed: %v", err)
	}

	// Verify files exist.
	content, err := os.ReadFile(filepath.Join(destDir, "hello.txt")) //nolint:gosec // путь из t.TempDir()
	if err != nil {
		t.Fatalf("hello.txt not found: %v", err)
	}
	if string(content) != "hello world" {
		t.Errorf("expected 'hello world', got %q", string(content))
	}

	content2, err := os.ReadFile(filepath.Join(destDir, "sub", "file.txt")) //nolint:gosec // путь из t.TempDir()
	if err != nil {
		t.Fatalf("sub/file.txt not found: %v", err)
	}
	if string(content2) != "nested content" {
		t.Errorf("expected 'nested content', got %q", string(content2))
	}
}

func TestBuildPipeline_ExtractZip_PathTraversal(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	// Add a file with path traversal.
	fw, err := zw.Create("../evil.txt")
	if err != nil {
		t.Fatalf("failed to create zip entry: %v", err)
	}
	if _, err := fw.Write([]byte("evil")); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	if err := zw.Close(); err != nil {
		t.Fatalf("failed to close zip: %v", err)
	}

	destDir := t.TempDir()
	pipeline := &BuildPipeline{}
	err = pipeline.extractZip(buf.Bytes(), destDir)
	if err == nil {
		t.Fatal("expected error for path traversal, got nil")
	}
	if !strings.Contains(err.Error(), "invalid path") && !strings.Contains(err.Error(), "escapes") {
		t.Errorf("expected path traversal error, got: %v", err)
	}
}

// ─── Tests: extractTar ───────────────────────────────────────────────────────

func TestBuildPipeline_ExtractTar_Success(t *testing.T) {
	// Create a tar in memory.
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	// Add a directory.
	if err := tw.WriteHeader(&tar.Header{
		Name:     "data/",
		Typeflag: tar.TypeDir,
		Mode:     0o750,
	}); err != nil {
		t.Fatalf("failed to write tar header: %v", err)
	}

	// Add a file.
	if err := tw.WriteHeader(&tar.Header{
		Name: "data/file.txt",
		Mode: 0o644,
		Size: int64(len("tar content")),
	}); err != nil {
		t.Fatalf("failed to write tar header: %v", err)
	}
	if _, err := tw.Write([]byte("tar content")); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	// Add a root-level file.
	if err := tw.WriteHeader(&tar.Header{
		Name: "root.txt",
		Mode: 0o644,
		Size: int64(len("root")),
	}); err != nil {
		t.Fatalf("failed to write tar header: %v", err)
	}
	if _, err := tw.Write([]byte("root")); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	if err := tw.Close(); err != nil {
		t.Fatalf("failed to close tar: %v", err)
	}

	destDir := t.TempDir()
	pipeline := &BuildPipeline{}
	if err := pipeline.extractTar(bytes.NewReader(buf.Bytes()), destDir); err != nil {
		t.Fatalf("extractTar failed: %v", err)
	}

	// Verify files.
	content, err := os.ReadFile(filepath.Join(destDir, "data", "file.txt")) //nolint:gosec // путь из t.TempDir()
	if err != nil {
		t.Fatalf("data/file.txt not found: %v", err)
	}
	if string(content) != "tar content" {
		t.Errorf("expected 'tar content', got %q", string(content))
	}

	content2, err := os.ReadFile(filepath.Join(destDir, "root.txt")) //nolint:gosec // путь из t.TempDir()
	if err != nil {
		t.Fatalf("root.txt not found: %v", err)
	}
	if string(content2) != "root" {
		t.Errorf("expected 'root', got %q", string(content2))
	}
}

func TestBuildPipeline_ExtractTar_PathTraversal(t *testing.T) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	// Add a file with path traversal.
	if err := tw.WriteHeader(&tar.Header{
		Name:     "../evil.txt",
		Mode:     0o644,
		Size:     int64(len("evil")),
		Typeflag: tar.TypeReg,
	}); err != nil {
		t.Fatalf("failed to write tar header: %v", err)
	}
	if _, err := tw.Write([]byte("evil")); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	if err := tw.Close(); err != nil {
		t.Fatalf("failed to close tar: %v", err)
	}

	destDir := t.TempDir()
	pipeline := &BuildPipeline{}
	err := pipeline.extractTar(bytes.NewReader(buf.Bytes()), destDir)
	if err == nil {
		t.Fatal("expected error for path traversal, got nil")
	}
	if !strings.Contains(err.Error(), "invalid path") && !strings.Contains(err.Error(), "escapes") {
		t.Errorf("expected path traversal error, got: %v", err)
	}
}

// ─── Tests: DeleteBuild ──────────────────────────────────────────────────────

func TestBuildPipeline_DeleteBuild_Success(t *testing.T) {
	build := &domain.ServerBuild{
		ID:      100,
		GameID:  1,
		Version: "v1",
	}

	storage := &buildMockStorage{
		getByVersionFn: func(ctx context.Context, gameID int64, version string) (*domain.ServerBuild, error) {
			return build, nil
		},
		countActiveInstancesFn: func(ctx context.Context, buildID int64) (int, error) {
			return 0, nil // no active instances
		},
		deleteFn: func(ctx context.Context, id int64) error {
			if id != 100 {
				t.Errorf("expected delete build ID 100, got %d", id)
			}
			return nil
		},
	}

	fsDeleted := false
	fs := &buildMockStorageFS{
		deleteFn: func(gameID int64, version string) error {
			if gameID != 1 || version != "v1" {
				t.Errorf("expected delete gameID=1, version=v1, got %d, %s", gameID, version)
			}
			fsDeleted = true
			return nil
		},
	}

	pipeline := NewBuildPipeline(storage, fs, nil, nil, nil, config.LimitsConfig{})
	ctx := context.Background()

	err := pipeline.DeleteBuild(ctx, 1, "v1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !fsDeleted {
		t.Error("expected FS delete to be called")
	}
}

func TestBuildPipeline_DeleteBuild_InUse(t *testing.T) {
	build := &domain.ServerBuild{
		ID:      100,
		GameID:  1,
		Version: "v1",
	}

	storage := &buildMockStorage{
		getByVersionFn: func(ctx context.Context, gameID int64, version string) (*domain.ServerBuild, error) {
			return build, nil
		},
		countActiveInstancesFn: func(ctx context.Context, buildID int64) (int, error) {
			return 3, nil // active instances exist
		},
	}

	fsCalled := false
	fs := &buildMockStorageFS{
		deleteFn: func(gameID int64, version string) error {
			fsCalled = true
			return nil
		},
	}

	pipeline := NewBuildPipeline(storage, fs, nil, nil, nil, config.LimitsConfig{})
	ctx := context.Background()

	err := pipeline.DeleteBuild(ctx, 1, "v1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrBuildInUse) {
		t.Errorf("expected ErrBuildInUse, got %v", err)
	}
	if fsCalled {
		t.Error("expected FS delete NOT to be called when build is in use")
	}
}

func TestBuildPipeline_DeleteBuild_NotFound(t *testing.T) {
	storage := &buildMockStorage{
		getByVersionFn: func(ctx context.Context, gameID int64, version string) (*domain.ServerBuild, error) {
			return nil, domain.ErrNotFound
		},
	}

	fs := &buildMockStorageFS{}

	pipeline := NewBuildPipeline(storage, fs, nil, nil, nil, config.LimitsConfig{})
	ctx := context.Background()

	err := pipeline.DeleteBuild(ctx, 1, "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "get build") {
		t.Errorf("expected 'get build' in error, got: %v", err)
	}
}

// ─── Tests: ListBuilds ───────────────────────────────────────────────────────

func TestBuildPipeline_ListBuilds(t *testing.T) {
	builds := []*domain.ServerBuild{
		{ID: 1, GameID: 42, Version: "v2"},
		{ID: 2, GameID: 42, Version: "v1"},
	}

	storage := &buildMockStorage{
		listByGameFn: func(ctx context.Context, gameID int64, limit int) ([]*domain.ServerBuild, error) {
			if gameID != 42 {
				t.Errorf("expected gameID 42, got %d", gameID)
			}
			return builds, nil
		},
	}

	pipeline := NewBuildPipeline(storage, nil, nil, nil, nil, config.LimitsConfig{})
	ctx := context.Background()

	result, err := pipeline.ListBuilds(ctx, 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 builds, got %d", len(result))
	}
	if result[0].Version != "v2" {
		t.Errorf("expected first version 'v2', got %q", result[0].Version)
	}
	if result[1].Version != "v1" {
		t.Errorf("expected second version 'v1', got %q", result[1].Version)
	}
}

// ─── Tests: GetBuild ─────────────────────────────────────────────────────────

func TestBuildPipeline_GetBuild_Success(t *testing.T) {
	expected := &domain.ServerBuild{
		ID:       100,
		GameID:   42,
		Version:  "v1.0",
		ImageTag: "welwise/game-42:v1.0",
	}

	storage := &buildMockStorage{
		getByVersionFn: func(ctx context.Context, gameID int64, version string) (*domain.ServerBuild, error) {
			if gameID != 42 || version != "v1.0" {
				t.Errorf("expected gameID=42, version=v1.0, got %d, %s", gameID, version)
			}
			return expected, nil
		},
	}

	pipeline := NewBuildPipeline(storage, nil, nil, nil, nil, config.LimitsConfig{})
	ctx := context.Background()

	result, err := pipeline.GetBuild(ctx, 42, "v1.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != expected.ID {
		t.Errorf("expected ID %d, got %d", expected.ID, result.ID)
	}
	if result.Version != expected.Version {
		t.Errorf("expected version %q, got %q", expected.Version, result.Version)
	}
}

func TestBuildPipeline_GetBuild_NotFound(t *testing.T) {
	storage := &buildMockStorage{
		getByVersionFn: func(ctx context.Context, gameID int64, version string) (*domain.ServerBuild, error) {
			return nil, domain.ErrNotFound
		},
	}

	pipeline := NewBuildPipeline(storage, nil, nil, nil, nil, config.LimitsConfig{})
	ctx := context.Background()

	_, err := pipeline.GetBuild(ctx, 42, "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
