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

	"github.com/Be4Die/game-developer-hub/game-deployment-agent/internal/domain"
)

func createTestZip(t *testing.T, files map[string]string) string {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "test-*.zip")
	if err != nil {
		t.Fatalf("failed to create temp zip: %v", err)
	}
	defer func() {
		_ = tmpFile.Close()
	}()

	w := zip.NewWriter(tmpFile)
	for name, content := range files {
		f, err := w.Create(name)
		if err != nil {
			t.Fatalf("failed to write entry %s: %v", name, err)
		}
		if _, err := f.Write([]byte(content)); err != nil {
			t.Fatalf("failed to write content %s: %v", name, err)
		}
	}

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}

	return tmpFile.Name()
}

func TestUnit_AgentArchive_ValidZip(t *testing.T) {
	zipPath := createTestZip(t, map[string]string{
		"index.html":       "<!DOCTYPE html><html><body><h1>Game</h1></body></html>",
		"game.js":          "console.log('game started');",
		"style.css":        "body { background: black; }",
		"assets/data.json": "{\"score\": 100}",
	})
	defer func() {
		_ = os.Remove(zipPath)
	}()

	targetDir, err := os.MkdirTemp("", "unpacked-*")
	if err != nil {
		t.Fatalf("failed to create target dir: %v", err)
	}
	defer func() {
		_ = os.RemoveAll(targetDir)
	}()

	err = ExtractArchive(zipPath, targetDir, 10*1024*1024, 100)
	if err != nil {
		t.Fatalf("expected valid zip extraction, got: %v", err)
	}

	indexContent, err := os.ReadFile(filepath.Join(targetDir, "index.html")) //nolint:gosec
	if err != nil {
		t.Fatalf("expected index.html to exist, err: %v", err)
	}
	if !bytes.Contains(indexContent, []byte("Game")) {
		t.Errorf("unexpected index content: %s", string(indexContent))
	}
}

func TestUnit_AgentArchive_DisallowedFileType(t *testing.T) {
	zipPath := createTestZip(t, map[string]string{
		"index.html": "<!DOCTYPE html><html></html>",
		"hack.php":   "<?php system($_GET['cmd']); ?>",
	})
	defer func() {
		_ = os.Remove(zipPath)
	}()

	targetDir, err := os.MkdirTemp("", "unpacked-*")
	if err != nil {
		t.Fatalf("failed to create target dir: %v", err)
	}
	defer func() {
		_ = os.RemoveAll(targetDir)
	}()

	err = ExtractArchive(zipPath, targetDir, 10*1024*1024, 100)
	if err == nil {
		t.Fatal("expected error for disallowed file type .php, got nil")
	}
	if !bytes.Contains([]byte(err.Error()), []byte("disallowed file type")) {
		t.Errorf("expected disallowed file type error, got: %v", err)
	}
}

func TestUnit_AgentArchive_MissingIndexHtml(t *testing.T) {
	zipPath := createTestZip(t, map[string]string{
		"main.js": "console.log('no index');",
	})
	defer func() {
		_ = os.Remove(zipPath)
	}()

	targetDir, err := os.MkdirTemp("", "unpacked-*")
	if err != nil {
		t.Fatalf("failed to create target dir: %v", err)
	}
	defer func() {
		_ = os.RemoveAll(targetDir)
	}()

	err = ExtractArchive(zipPath, targetDir, 10*1024*1024, 100)
	if err == nil {
		t.Fatal("expected error for missing index.html, got nil")
	}
	if !errors.Is(err, domain.ErrNoIndexHTML) {
		t.Errorf("expected ErrNoIndexHTML, got: %v", err)
	}
}

func TestUnit_AgentArchive_ZipSlipProtection(t *testing.T) {
	zipPath := createTestZip(t, map[string]string{
		"index.html":    "<!DOCTYPE html><html></html>",
		"../../evil.js": "evil()",
	})
	defer func() {
		_ = os.Remove(zipPath)
	}()

	targetDir, err := os.MkdirTemp("", "unpacked-*")
	if err != nil {
		t.Fatalf("failed to create target dir: %v", err)
	}
	defer func() {
		_ = os.RemoveAll(targetDir)
	}()

	err = ExtractArchive(zipPath, targetDir, 10*1024*1024, 100)
	if err == nil {
		t.Fatal("expected ZipSlip error, got nil")
	}
}

func TestArchiveNestedAndMacOSMetadata(t *testing.T) {
	// 1. Test ZIP with __MACOSX metadata and .tmp extension
	zipTmp, err := os.CreateTemp("", "test-*.tmp")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(zipTmp.Name())

	zw := zip.NewWriter(zipTmp)
	// Add __MACOSX entry
	mw, _ := zw.Create("__MACOSX/._index.html")
	_, _ = mw.Write([]byte{0x00, 0x05, 0x16})
	// Add nested game files
	iw, _ := zw.Create("MyGame/index.html")
	_, _ = iw.Write([]byte("<!DOCTYPE html><html></html>"))
	bw, _ := zw.Create("MyGame/game.wasm")
	_, _ = bw.Write([]byte("wasm-binary"))
	_ = zw.Close()
	_ = zipTmp.Close()

	destDir := t.TempDir()
	if err := ExtractArchive(zipTmp.Name(), destDir, 0, 0); err != nil {
		t.Fatalf("failed to extract nested zip with macos metadata: %v", err)
	}
	if _, err := os.Stat(filepath.Join(destDir, "index.html")); err != nil {
		t.Fatalf("expected index.html at root of destDir: %v", err)
	}
	if _, err := os.Stat(filepath.Join(destDir, "game.wasm")); err != nil {
		t.Fatalf("expected game.wasm at root of destDir: %v", err)
	}

	// 2. Test TAR.GZ with nested directory and .tmp extension
	tarTmp, err := os.CreateTemp("", "test-*.tmp")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tarTmp.Name())

	gw := gzip.NewWriter(tarTmp)
	tw := tar.NewWriter(gw)

	indexContent := []byte("<!DOCTYPE html><html></html>")
	_ = tw.WriteHeader(&tar.Header{
		Name: "Archive/index.html",
		Mode: 0o644,
		Size: int64(len(indexContent)),
	})
	_, _ = tw.Write(indexContent)

	_ = tw.Close()
	_ = gw.Close()
	_ = tarTmp.Close()

	destDir2 := t.TempDir()
	if err := ExtractArchive(tarTmp.Name(), destDir2, 0, 0); err != nil {
		t.Fatalf("failed to extract nested tar.gz: %v", err)
	}
	if _, err := os.Stat(filepath.Join(destDir2, "index.html")); err != nil {
		t.Fatalf("expected index.html at root of destDir2: %v", err)
	}
}

