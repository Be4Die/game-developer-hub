package filesystem

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
)

func TestUnit_Archive_ValidZip(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "game.zip")
	targetDir := filepath.Join(tmpDir, "unpacked")

	// Создаем тестовый zip-архив с index.html
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	f, err := zw.Create("index.html")
	if err != nil {
		t.Fatalf("create index.html: %v", err)
	}
	if _, err := f.Write([]byte("<html><body>Game</body></html>")); err != nil {
		t.Fatalf("write index.html: %v", err)
	}

	asset, err := zw.Create("js/game.js")
	if err != nil {
		t.Fatalf("create game.js: %v", err)
	}
	if _, err := asset.Write([]byte("console.log('started');")); err != nil {
		t.Fatalf("write game.js: %v", err)
	}

	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}

	if err := os.WriteFile(archivePath, buf.Bytes(), 0o600); err != nil {
		t.Fatalf("write archive file: %v", err)
	}

	// Распаковываем
	if err := ExtractArchive(archivePath, targetDir); err != nil {
		t.Fatalf("expected successful extraction, got: %v", err)
	}

	// Проверяем наличие распакованных файлов
	indexPath := filepath.Join(targetDir, "index.html")
	content, err := os.ReadFile(filepath.Clean(indexPath)) //nolint:gosec
	if err != nil {
		t.Fatalf("read unpacked index.html: %v", err)
	}
	if string(content) != "<html><body>Game</body></html>" {
		t.Errorf("unexpected content: %s", string(content))
	}

	jsPath := filepath.Join(targetDir, "js", "game.js")
	if _, err := os.Stat(jsPath); err != nil {
		t.Errorf("expected js/game.js to exist: %v", err)
	}
}

func TestUnit_Archive_MissingIndexHtml(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "invalid.zip")
	targetDir := filepath.Join(tmpDir, "unpacked")

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	f, err := zw.Create("readme.txt")
	if err != nil {
		t.Fatalf("create readme.txt: %v", err)
	}
	_, _ = f.Write([]byte("no index here"))
	_ = zw.Close()

	if err := os.WriteFile(archivePath, buf.Bytes(), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	err = ExtractArchive(archivePath, targetDir)
	if !errors.Is(err, domain.ErrNoIndexHTML) {
		t.Errorf("expected ErrNoIndexHTML, got: %v", err)
	}
}

func TestUnit_Archive_ZipSlipProtection(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "malicious.zip")
	targetDir := filepath.Join(tmpDir, "unpacked")

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	f1, _ := zw.Create("index.html")
	_, _ = f1.Write([]byte("valid"))

	f2, _ := zw.Create("../../malicious.txt")
	_, _ = f2.Write([]byte("escaped!"))
	_ = zw.Close()

	_ = os.WriteFile(archivePath, buf.Bytes(), 0o600)

	err := ExtractArchive(archivePath, targetDir)
	if !errors.Is(err, domain.ErrInvalidArchive) {
		t.Errorf("expected ErrInvalidArchive for zip slip, got: %v", err)
	}
}

func TestUnit_Archive_ValidTarGz(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "game.tar.gz")
	targetDir := filepath.Join(tmpDir, "unpacked")

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	content := []byte("<html><body>Tar Game</body></html>")
	hdr := &tar.Header{
		Name: "index.html",
		Mode: 0o644,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("write tar header: %v", err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatalf("write tar content: %v", err)
	}

	_ = tw.Close()
	_ = gw.Close()

	_ = os.WriteFile(archivePath, buf.Bytes(), 0o600)

	if err := ExtractArchive(archivePath, targetDir); err != nil {
		t.Fatalf("expected successful tar.gz extraction, got: %v", err)
	}

	indexPath := filepath.Join(targetDir, "index.html")
	readContent, err := os.ReadFile(filepath.Clean(indexPath)) //nolint:gosec
	if err != nil {
		t.Fatalf("read index.html: %v", err)
	}
	if string(readContent) != string(content) {
		t.Errorf("content mismatch: %s", string(readContent))
	}
}

func TestUnit_Archive_UnityWebGLBrotliAssets(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "unity_webgl.zip")
	targetDir := filepath.Join(tmpDir, "unpacked")

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	files := map[string][]byte{
		"index.html":                  []byte("<html><body>Unity</body></html>"),
		"style.css":                   []byte("body { margin: 0; }"),
		"Build/Build.data.br":         []byte("unity-data-brotli"),
		"Build/Build.framework.js.br": []byte("unity-framework-brotli"),
		"Build/Build.loader.js":       []byte("unity-loader-js"),
		"Build/Build.wasm.br":         []byte("unity-wasm-brotli"),
	}

	for name, content := range files {
		f, err := zw.Create(name)
		if err != nil {
			t.Fatalf("create entry %s: %v", name, err)
		}
		if _, err := f.Write(content); err != nil {
			t.Fatalf("write entry %s: %v", name, err)
		}
	}

	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}

	if err := os.WriteFile(archivePath, buf.Bytes(), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	if err := ExtractArchive(archivePath, targetDir); err != nil {
		t.Fatalf("expected successful Unity WebGL extraction, got: %v", err)
	}

	for name := range files {
		unpackedFile := filepath.Join(targetDir, filepath.FromSlash(name))
		if _, err := os.Stat(unpackedFile); err != nil {
			t.Errorf("expected %s to be extracted: %v", name, err)
		}
	}
}

func TestArchive_MacOSMetadataAndNested(t *testing.T) {
	// 1. ZIP with macOS metadata and single-root wrapper folder
	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "game.zip")
	targetDir := filepath.Join(tmpDir, "unpacked")

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	mw, _ := zw.Create("__MACOSX/._index.html")
	_, _ = mw.Write([]byte{0x01, 0x02})
	iw, _ := zw.Create("GameFolder/index.html")
	_, _ = iw.Write([]byte("<html>Game</html>"))
	bw, _ := zw.Create("GameFolder/game.wasm")
	_, _ = bw.Write([]byte("wasm"))
	_ = zw.Close()

	if err := os.WriteFile(archivePath, buf.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := ExtractArchive(archivePath, targetDir); err != nil {
		t.Fatalf("ExtractArchive failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(targetDir, "index.html")); err != nil {
		t.Fatalf("expected index.html at root of targetDir: %v", err)
	}
	if _, err := os.Stat(filepath.Join(targetDir, "game.wasm")); err != nil {
		t.Fatalf("expected game.wasm at root of targetDir: %v", err)
	}
}

