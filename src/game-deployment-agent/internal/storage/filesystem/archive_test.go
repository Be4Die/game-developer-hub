package filesystem

import (
	"archive/zip"
	"bytes"
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
	defer tmpFile.Close()

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
	defer os.Remove(zipPath)

	targetDir, err := os.MkdirTemp("", "unpacked-*")
	if err != nil {
		t.Fatalf("failed to create target dir: %v", err)
	}
	defer os.RemoveAll(targetDir)

	err = ExtractArchive(zipPath, targetDir, 10*1024*1024, 100)
	if err != nil {
		t.Fatalf("expected valid zip extraction, got: %v", err)
	}

	indexContent, err := os.ReadFile(filepath.Join(targetDir, "index.html"))
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
	defer os.Remove(zipPath)

	targetDir, err := os.MkdirTemp("", "unpacked-*")
	if err != nil {
		t.Fatalf("failed to create target dir: %v", err)
	}
	defer os.RemoveAll(targetDir)

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
	defer os.Remove(zipPath)

	targetDir, err := os.MkdirTemp("", "unpacked-*")
	if err != nil {
		t.Fatalf("failed to create target dir: %v", err)
	}
	defer os.RemoveAll(targetDir)

	err = ExtractArchive(zipPath, targetDir, 10*1024*1024, 100)
	if err == nil {
		t.Fatal("expected error for missing index.html, got nil")
	}
	if err != domain.ErrNoIndexHtml {
		t.Errorf("expected ErrNoIndexHtml, got: %v", err)
	}
}

func TestUnit_AgentArchive_ZipSlipProtection(t *testing.T) {
	zipPath := createTestZip(t, map[string]string{
		"index.html":       "<!DOCTYPE html><html></html>",
		"../../evil.js":    "evil()",
	})
	defer os.Remove(zipPath)

	targetDir, err := os.MkdirTemp("", "unpacked-*")
	if err != nil {
		t.Fatalf("failed to create target dir: %v", err)
	}
	defer os.RemoveAll(targetDir)

	err = ExtractArchive(zipPath, targetDir, 10*1024*1024, 100)
	if err == nil {
		t.Fatal("expected ZipSlip error, got nil")
	}
}
